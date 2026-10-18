package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"go.uber.org/zap"
	"golang.org/x/text/unicode/norm"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/integration/llm"
	"the-slot-finder/server/internal/logger"
	"the-slot-finder/server/internal/models"
)

// maskCounts is returned to the client as "masked".
type maskCounts struct {
	Emails    int `json:"emails"`
	Customers int `json:"customers"`
	Phones    int `json:"phones"`
}

var (
	emailPattern = regexp.MustCompile(`[\p{L}\p{N}._%+\-']+@[\p{L}\p{N}\-]+(?:\.[\p{L}\p{N}\-]+)+`)
	// The date alternative comes first, so an ISO date is matched whole and kept.
	// A gap of up to three characters keeps "+84 (0) 912" and "0912 - 345" together.
	phonePattern = regexp.MustCompile(`\d{4}-\d{2}-\d{2}|[+(]?\d(?:[ \t\n().-]{0,3}\d)*`)
	trailingDate = regexp.MustCompile(`[ \t\n().-]{1,3}\d{4}-\d{2}-\d{2}$`)
	datePattern  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

const minPhoneDigits = 7

// maskBrief hides emails, known customers, typed terms and phones, in that order.
// Customer and term matches are case-insensitive and whole word. Empty terms are ignored.
// The text and terms are NFKC-normalised and stripped of invisible characters first, so a hidden name cannot slip by.
func maskBrief(text string, customers []string, extra []string) (masked string, counts maskCounts) {
	text = normalizeText(text)
	masked = emailPattern.ReplaceAllStringFunc(text, func(string) string {
		counts.Emails++
		return constants.MaskEmail
	})

	for _, term := range maskTerms(customers, extra) {
		var n int
		masked, n = maskTerm(masked, term)
		counts.Customers += n
	}

	masked = phonePattern.ReplaceAllStringFunc(masked, func(m string) string {
		if datePattern.MatchString(m) {
			return m
		}
		// A date glued to the phone by one separator is not part of it.
		tail := ""
		if loc := trailingDate.FindStringIndex(m); loc != nil {
			m, tail = m[:loc[0]], m[loc[0]:]
		}
		if len(strings.Map(digitsOnly, m)) < minPhoneDigits {
			return m + tail
		}
		counts.Phones++
		return constants.MaskPhone + tail
	})
	return masked, counts
}

var placeholderWords = map[string]bool{"email": true, "customer": true, "phone": true}

// normalizeText applies NFKC, removes zero-width characters and the soft hyphen, and turns a no-break space into a space.
func normalizeText(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\u00AD', '\u200B', '\u200C', '\u200D', '\u2060', '\uFEFF':
			return -1
		case '\u00A0':
			return ' '
		}
		return r
	}, norm.NFKC.String(s))
}

func digitsOnly(r rune) rune {
	if r >= '0' && r <= '9' {
		return r
	}
	return -1
}

// maskTerms trims, drops empty and repeated terms, and puts the longest first so "Acme Corp" wins over "Acme".
func maskTerms(lists ...[]string) []string {
	seen := map[string]bool{}
	var terms []string
	for _, list := range lists {
		for _, t := range list {
			t = strings.TrimSpace(normalizeText(t))
			// A term with no letter or digit (like "-") would mangle dates. A term equal to a placeholder, with or
			// without its brackets, would hit the placeholders already written.
			key := strings.ToLower(t)
			if !strings.ContainsFunc(t, isWordRune) || isPlaceholder(key) {
				continue
			}
			if !seen[key] {
				seen[key] = true
				terms = append(terms, t)
			}
		}
	}
	sort.SliceStable(terms, func(i, j int) bool { return utf8.RuneCountInString(terms[i]) > utf8.RuneCountInString(terms[j]) })
	return terms
}

func isPlaceholder(key string) bool {
	// Trimming the brackets also covers the Mask* constants, which are "[" + word + "]".
	return placeholderWords[strings.Trim(key, "[]")]
}

func isWordRune(r rune) bool { return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r) }

// maskTerm replaces every whole-word, case-insensitive match of term. Go's \b only knows ASCII, so the edges are
// checked on runes; an edge that is not a word character (as in "C++") needs no boundary.
func maskTerm(text, term string) (string, int) {
	// Any run of whitespace (newline, double or thin space) joins the words of a name.
	words := strings.Fields(term)
	for i, w := range words {
		words[i] = regexp.QuoteMeta(w)
	}
	re := regexp.MustCompile(`(?i)` + strings.Join(words, `[\s\p{Z}]+`))
	first, _ := utf8.DecodeRuneInString(term)
	last, _ := utf8.DecodeLastRuneInString(term)
	var b strings.Builder
	pos, count := 0, 0
	for _, loc := range re.FindAllStringIndex(text, -1) {
		before, _ := utf8.DecodeLastRuneInString(text[:loc[0]])
		after, _ := utf8.DecodeRuneInString(text[loc[1]:])
		if (isWordRune(first) && loc[0] > 0 && isWordRune(before)) ||
			(isWordRune(last) && loc[1] < len(text) && isWordRune(after)) {
			continue
		}
		b.WriteString(text[pos:loc[0]])
		b.WriteString(constants.MaskCustomer)
		pos = loc[1]
		count++
	}
	b.WriteString(text[pos:])
	return b.String(), count
}

// rawSuggestion mirrors the suggest_project arguments. It is untrusted model output.
type rawSuggestion struct {
	Name          string      `json:"name"`
	StartEarliest *string     `json:"start_earliest"`
	Deadline      *string     `json:"deadline"`
	ScopePoints   *float64    `json:"scope_points"`
	Demands       []rawDemand `json:"demands"`
}

type rawDemand struct {
	ExpertiseName string  `json:"expertise_name"`
	TotalHours    float64 `json:"total_hours"`
	MinLevel      *int    `json:"min_level"`
	Quote         string  `json:"quote"`
}

// Suggestion is a validated project suggestion.
type Suggestion struct {
	Name          string
	StartEarliest *string
	Deadline      *string
	ScopePoints   *float64
	Demands       []SuggestedDemand
}

// SuggestedDemand is a validated demand, tied to a catalogue expertise.
type SuggestedDemand struct {
	ExpertiseID   string
	ExpertiseName string
	TotalHours    float64
	MinLevel      *int
	Quote         string
}

// normalizeForQuote makes two texts comparable: normalised, lower case, one space between words.
func normalizeForQuote(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(normalizeText(s))), " ")
}

// validateSuggestion keeps only what the masked brief supports and the catalogue knows; the rest is dropped or nulled,
// each with a warning. It never creates an expertise.
func validateSuggestion(raw rawSuggestion, maskedText string, catalogue []models.Expertise) (Suggestion, []string) {
	var warnings []string
	out := Suggestion{Name: strings.TrimSpace(normalizeText(raw.Name))}
	if r := []rune(out.Name); len(r) > constants.IntakeMaxNameChars {
		out.Name = strings.TrimSpace(string(r[:constants.IntakeMaxNameChars]))
	}
	for _, p := range []string{constants.MaskEmail, constants.MaskCustomer, constants.MaskPhone} {
		if strings.Contains(out.Name, p) {
			warnings = append(warnings, constants.WarnPlaceholderInName)
			break
		}
	}

	start, deadline := parseIntakeDate(raw.StartEarliest), parseIntakeDate(raw.Deadline)
	if start != nil && deadline != nil && deadline.Before(*start) {
		start, deadline = nil, nil
		warnings = append(warnings, constants.WarnDatesOrder)
	}
	out.StartEarliest, out.Deadline = formatIntakeDate(start), formatIntakeDate(deadline)
	if raw.ScopePoints != nil && *raw.ScopePoints >= 0 && !math.IsInf(*raw.ScopePoints, 0) {
		out.ScopePoints = raw.ScopePoints
	}

	byName := make(map[string]models.Expertise, len(catalogue))
	for _, e := range catalogue {
		byName[strings.ToLower(strings.TrimSpace(normalizeText(e.Name)))] = e
	}
	haystack := normalizeForQuote(maskedText)
	seen := map[string]bool{}
	for _, d := range raw.Demands {
		quote := normalizeForQuote(d.Quote)
		if quote == "" || !strings.Contains(haystack, quote) {
			warnings = append(warnings, constants.WarnDemandNoQuote)
			continue
		}
		e, ok := byName[strings.ToLower(strings.TrimSpace(normalizeText(d.ExpertiseName)))]
		if !ok {
			warnings = append(warnings, constants.WarnDemandUnknownExpertise)
			continue
		}
		if !(d.TotalHours > 0 && d.TotalHours <= constants.IntakeMaxHours) {
			warnings = append(warnings, constants.WarnDemandHours)
			continue
		}
		if seen[e.ID] {
			warnings = append(warnings, constants.WarnDemandDuplicateExpertise)
		}
		seen[e.ID] = true
		demand := SuggestedDemand{ExpertiseID: e.ID, ExpertiseName: e.Name, TotalHours: d.TotalHours, Quote: strings.TrimSpace(d.Quote)}
		if d.MinLevel != nil && *d.MinLevel >= constants.IntakeMinLevel && *d.MinLevel <= constants.IntakeMaxLevel {
			demand.MinLevel = d.MinLevel
		}
		out.Demands = append(out.Demands, demand)
	}
	return out, warnings
}

func formatIntakeDate(t *time.Time) *string {
	if t == nil {
		return nil
	}
	s := t.Format(constants.DateLayout)
	return &s
}

// parseIntakeDate returns the date, or nil when it is absent or not a real YYYY-MM-DD date.
func parseIntakeDate(s *string) *time.Time {
	if s == nil {
		return nil
	}
	t, err := time.Parse(constants.DateLayout, strings.TrimSpace(*s))
	if err != nil {
		return nil
	}
	return &t
}

// IntakeInput is a brief and the extra names the user asked to hide.
type IntakeInput struct {
	Text      string
	MaskTerms []string
}

// IntakeResult is a validated suggestion, what validation changed, and what masking hid.
type IntakeResult struct {
	Suggestion Suggestion
	Warnings   []string
	Masked     maskCounts
}

// intakeAttempts is how many times the model may be asked for a usable suggest_project call.
const intakeAttempts = 2

func (s *aiService) ExtractIntake(ctx context.Context, in IntakeInput) (*IntakeResult, error) {
	// Creating a project needs a planner, or a manager who leads a program; the program is chosen later.
	if _, err := s.authz.Require(ctx, CanUsePrograms); err != nil {
		return nil, err
	}
	reader, err := s.authz.ProgramReader(ctx)
	if err != nil {
		return nil, err
	}
	programs, err := s.programs.List(ctx, ListProgramsInput{PageSize: constants.NoLimit})
	if err != nil {
		return nil, err
	}
	customers := make([]string, 0, len(programs.Data))
	leads := false
	for _, p := range programs.Data {
		customers = append(customers, p.Client)
		leads = leads || (p.LeadID != nil && *p.LeadID == reader.LeadID)
	}
	if !reader.All && !leads {
		return nil, errors.ForbiddenError(constants.MsgInsufficientPermissions, nil).WithOperation(opIntakeAI).WithResource(resourceAI)
	}
	catalogue, err := s.people.ListExpertises(ctx)
	if err != nil {
		return nil, err
	}
	if s.cfg.LLMAPIKey == "" {
		return nil, errors.NewAppError(constants.AINotConfigured, constants.MsgAINotConfigured, errors.ErrorTypeExternal,
			http.StatusServiceUnavailable).WithOperation(opIntakeAI).WithResource(resourceAI)
	}
	// Only the masked text goes into the request; the original is never sent or logged.
	masked, counts := maskBrief(in.Text, customers, in.MaskTerms)
	var names strings.Builder
	for _, e := range catalogue {
		fmt.Fprintf(&names, "- %s\n", e.Name)
	}
	req := llm.Request{
		SystemInstruction: &llm.Content{Parts: []llm.Part{{Text: aiIntakePrompt + names.String()}}},
		Contents:          []llm.Content{{Role: constants.LLMRoleUser, Parts: []llm.Part{{Text: masked}}}},
		Tools:             []llm.Tool{{FunctionDeclarations: []llm.FunctionDeclaration{suggestProjectDeclaration()}}},
	}
	for attempt := 0; attempt < intakeAttempts; attempt++ {
		resp, err := s.llm.Generate(ctx, req)
		if err != nil {
			logger.Log.Warn("Intake extraction model call failed", zap.Error(err))
			return nil, upstreamError(err).WithOperation(opIntakeAI).WithResource(resourceAI)
		}
		if raw, ok := parseSuggestion(resp.Content); ok {
			suggestion, warnings := validateSuggestion(raw, masked, catalogue)
			return &IntakeResult{Suggestion: suggestion, Warnings: warnings, Masked: counts}, nil
		}
		logger.Log.Warn("Intake extraction answer was not a suggestion", zap.Int("attempt", attempt+1))
	}
	return nil, errors.NewAppError(constants.AIUpstreamFailed, constants.MsgAIIntakeBadAnswer, errors.ErrorTypeExternal,
		http.StatusBadGateway).WithOperation(opIntakeAI).WithResource(resourceAI)
}

// parseSuggestion reads the arguments of the suggest_project call; false when there is none or they do not fit.
func parseSuggestion(c llm.Content) (rawSuggestion, bool) {
	for _, part := range c.Parts {
		if part.FunctionCall == nil || part.FunctionCall.Name != constants.AIToolSuggestProject {
			continue
		}
		data, err := json.Marshal(part.FunctionCall.Args)
		if err != nil {
			return rawSuggestion{}, false
		}
		var raw rawSuggestion
		if err := json.Unmarshal(data, &raw); err != nil {
			return rawSuggestion{}, false
		}
		return raw, true
	}
	return rawSuggestion{}, false
}

// suggestProjectDeclaration is the output function: it runs nothing, its arguments are the suggestion.
func suggestProjectDeclaration() llm.FunctionDeclaration {
	field := func(kind, description string) map[string]any {
		return map[string]any{"type": kind, "description": description, "nullable": true}
	}
	demand := map[string]any{
		"type":     "OBJECT",
		"required": []string{"expertise_name", "total_hours", "quote"},
		"properties": map[string]any{
			"expertise_name": map[string]any{"type": "STRING", "description": "One name from the expertise list"},
			"total_hours":    map[string]any{"type": "NUMBER", "description": "Budget hours the text states"},
			"min_level":      field("INTEGER", "Minimum level 1 to 5, only if the text says it"),
			"quote":          map[string]any{"type": "STRING", "description": "The exact span of the text that supports this demand"},
		},
	}
	return llm.FunctionDeclaration{
		Name:        constants.AIToolSuggestProject,
		Description: "Suggest one project from the brief. Leave a field null when the text does not say it.",
		Parameters: map[string]any{
			"type": "OBJECT",
			"properties": map[string]any{
				"name":           field("STRING", "Project name"),
				"start_earliest": field("STRING", "Earliest start, YYYY-MM-DD"),
				"deadline":       field("STRING", "Deadline, YYYY-MM-DD"),
				"scope_points":   field("NUMBER", "Scope points, only if the text says it"),
				"demands":        map[string]any{"type": "ARRAY", "items": demand},
			},
		},
	}
}
