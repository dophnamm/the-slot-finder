package services

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"

	"the-slot-finder/server/internal/constants"
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
	// A gap of up to two characters keeps "+84 (0) 912" together.
	phonePattern = regexp.MustCompile(`\d{4}-\d{2}-\d{2}|[+(]?\d(?:[ \t\n().-]{0,2}\d)*`)
	trailingDate = regexp.MustCompile(`[ \t\n().-]\d{4}-\d{2}-\d{2}$`)
	datePattern  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

const minPhoneDigits = 7

// maskBrief hides emails, known customers, typed terms and phones, in that order.
// Customer and term matches are case-insensitive and whole word. Empty terms are ignored.
// The text and terms are NFC-normalised and stripped of invisible characters first, so a hidden name cannot slip by.
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

// normalizeText applies NFC, removes zero-width characters and the soft hyphen, and turns a no-break space into a space.
func normalizeText(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '\u00AD', '\u200B', '\u200C', '\u200D', '\uFEFF':
			return -1
		case '\u00A0':
			return ' '
		}
		return r
	}, norm.NFC.String(s))
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
			// A term equal to a placeholder word would hit the placeholders already written.
			if key := strings.ToLower(t); t != "" && !seen[key] && !placeholderWords[key] {
				seen[key] = true
				terms = append(terms, t)
			}
		}
	}
	sort.SliceStable(terms, func(i, j int) bool { return utf8.RuneCountInString(terms[i]) > utf8.RuneCountInString(terms[j]) })
	return terms
}

func isWordRune(r rune) bool { return r == '_' || unicode.IsLetter(r) || unicode.IsNumber(r) }

// maskTerm replaces every whole-word, case-insensitive match of term. Go's \b only knows ASCII, so the edges are
// checked on runes; an edge that is not a word character (as in "C++") needs no boundary.
func maskTerm(text, term string) (string, int) {
	re := regexp.MustCompile("(?i)" + regexp.QuoteMeta(term))
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
