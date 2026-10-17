package services

import (
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"the-slot-finder/server/internal/constants"
)

// maskCounts is returned to the client as "masked".
type maskCounts struct {
	Emails    int `json:"emails"`
	Customers int `json:"customers"`
	Phones    int `json:"phones"`
}

var (
	emailPattern = regexp.MustCompile(`[\p{L}\p{N}._%+\-]+@[\p{L}\p{N}\-]+(?:\.[\p{L}\p{N}\-]+)+`)
	// The date alternative comes first, so an ISO date is matched whole and kept.
	// Separators stay on one line, and a gap of up to two characters keeps "+84 (0) 912" together.
	phonePattern = regexp.MustCompile(`\d{4}-\d{2}-\d{2}|[+(]?\d(?:[ \t().-]{0,2}\d)*`)
	datePattern  = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
)

const minPhoneDigits = 7

// maskBrief hides emails, known customers, typed terms and phones, in that order.
// Customer and term matches are case-insensitive and whole word. Empty terms are ignored.
// shortcut: a phone followed by a date after one space or dash is masked as one phone, upgrade to a tokenizer if that bites.
func maskBrief(text string, customers []string, extra []string) (masked string, counts maskCounts) {
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
		if datePattern.MatchString(m) || len(strings.Map(digitsOnly, m)) < minPhoneDigits {
			return m
		}
		counts.Phones++
		return constants.MaskPhone
	})
	return masked, counts
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
			t = strings.TrimSpace(t)
			if key := strings.ToLower(t); t != "" && !seen[key] {
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
