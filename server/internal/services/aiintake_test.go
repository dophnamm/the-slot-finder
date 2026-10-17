package services

import (
	"math"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/models"
)

func TestMaskBrief(t *testing.T) {
	tests := []struct {
		name      string
		text      string
		customers []string
		extra     []string
		want      string
		counts    maskCounts
	}{
		{name: "email replaced", text: "ana.lee@acme.com", want: constants.MaskEmail, counts: maskCounts{Emails: 1}},
		{name: "email upper case and plus address", text: "Write Ana.Lee+brief@ACME.co.uk now",
			want: "Write " + constants.MaskEmail + " now", counts: maskCounts{Emails: 1}},
		{name: "two emails", text: "a@b.io, c@d.io", want: constants.MaskEmail + ", " + constants.MaskEmail,
			counts: maskCounts{Emails: 2}},
		{name: "known customer case-insensitive", customers: []string{"Acme"}, text: "ACME and acme",
			want: constants.MaskCustomer + " and " + constants.MaskCustomer, counts: maskCounts{Customers: 2}},
		{name: "whole word only", customers: []string{"Acme"}, text: "Acmeville", want: "Acmeville"},
		{name: "word glued to a prefix kept", customers: []string{"Acme"}, text: "preAcme", want: "preAcme"},
		{name: "name inside a URL", customers: []string{"acme"}, text: "https://acme.example/brief",
			want: "https://" + constants.MaskCustomer + ".example/brief", counts: maskCounts{Customers: 1}},
		{name: "longer name wins over its prefix", customers: []string{"Acme", "Acme Corp"}, text: "Acme Corp asks",
			want: constants.MaskCustomer + " asks", counts: maskCounts{Customers: 1}},
		{name: "adjacent matches", customers: []string{"Acme"}, text: "acme acme",
			want: constants.MaskCustomer + " " + constants.MaskCustomer, counts: maskCounts{Customers: 2}},
		{name: "typed term", extra: []string{"Jane Roe"}, text: "Contact jane roe",
			want: "Contact " + constants.MaskCustomer, counts: maskCounts{Customers: 1}},
		{name: "empty terms ignored", customers: []string{" "}, extra: []string{""}, text: "plain text", want: "plain text"},
		{name: "regex metacharacters are literal", customers: []string{"A.c(me"}, extra: []string{".*"},
			text: "Abc(me and A.c(me and x.*y", want: "Abc(me and " + constants.MaskCustomer + " and x" + constants.MaskCustomer + "y",
			counts: maskCounts{Customers: 2}},
		{name: "term with non-word edge", customers: []string{"C++"}, text: "we use C++ daily",
			want: "we use " + constants.MaskCustomer + " daily", counts: maskCounts{Customers: 1}},
		{name: "phone", text: "+84 912 345 678", want: constants.MaskPhone, counts: maskCounts{Phones: 1}},
		{name: "phone with brackets and dashes", text: "call (555) 123-4567 today",
			want: "call " + constants.MaskPhone + " today", counts: maskCounts{Phones: 1}},
		{name: "short number kept", text: "about 120 hours, 12345", want: "about 120 hours, 12345"},
		{name: "ISO date kept", text: "2026-11-02", want: "2026-11-02"},
		{name: "date range kept", text: "from 2026-11-02 - 2026-12-18", want: "from 2026-11-02 - 2026-12-18"},
		{name: "date next to a phone", text: "2026-11-02 call 0912345678",
			want: "2026-11-02 call " + constants.MaskPhone, counts: maskCounts{Phones: 1}},
		{name: "non-ASCII name masked", customers: []string{"Đức Minh"}, text: "Đức Minh gửi brief",
			want: constants.MaskCustomer + " gửi brief", counts: maskCounts{Customers: 1}},
		{name: "no change when nothing to mask", customers: []string{"Acme"}, text: "Build a checkout in 2026-11-02",
			want: "Build a checkout in 2026-11-02"},
		{name: "NFD text with NFC customer", customers: []string{"Jos\u00e9 L\u00f3pez"}, text: "Jose\u0301 Lo\u0301pez asks",
			want: constants.MaskCustomer + " asks", counts: maskCounts{Customers: 1}},
		{name: "NFC text with NFD customer", customers: []string{"Jose\u0301 Lo\u0301pez"}, text: "Jos\u00e9 L\u00f3pez asks",
			want: constants.MaskCustomer + " asks", counts: maskCounts{Customers: 1}},
		{name: "zero-width space inside a name", customers: []string{"Acme"}, text: "Ac\u200bme", want: constants.MaskCustomer,
			counts: maskCounts{Customers: 1}},
		{name: "soft hyphen inside a name", customers: []string{"Acme"}, text: "Ac\u00adme", want: constants.MaskCustomer,
			counts: maskCounts{Customers: 1}},
		{name: "joiners and BOM inside a name", customers: []string{"Acme"}, text: "A\u200cc\u200dm\ufeffe", want: constants.MaskCustomer,
			counts: maskCounts{Customers: 1}},
		{name: "no-break space in a multi-word name", customers: []string{"Jane Roe"}, text: "Jane\u00a0Roe", want: constants.MaskCustomer,
			counts: maskCounts{Customers: 1}},
		{name: "invisible character in a term", extra: []string{"Ja\u200bne"}, text: "Jane", want: constants.MaskCustomer,
			counts: maskCounts{Customers: 1}},
		{name: "placeholder words are not terms", customers: []string{"email", "Customer"}, extra: []string{"PHONE"},
			text: "a@b.io", want: constants.MaskEmail, counts: maskCounts{Emails: 1}},
		{name: "apostrophe in email", text: "o'neil@acme.io", want: constants.MaskEmail, counts: maskCounts{Emails: 1}},
		{name: "newline inside a phone", text: "0912\n345 678", want: constants.MaskPhone, counts: maskCounts{Phones: 1}},
		{name: "trailing date kept after a phone", text: "0912345678 2026-11-02", want: constants.MaskPhone + " 2026-11-02",
			counts: maskCounts{Phones: 1}},
		{name: "two spaces in a name", customers: []string{"Acme Corp"}, text: "Acme  Corp", want: constants.MaskCustomer,
			counts: maskCounts{Customers: 1}},
		{name: "line wrap in a name", customers: []string{"Acme Corp"}, text: "Acme\nCorp", want: constants.MaskCustomer,
			counts: maskCounts{Customers: 1}},
		{name: "thin space in a name", customers: []string{"Acme Corp"}, text: "Acme\u2009Corp", want: constants.MaskCustomer,
			counts: maskCounts{Customers: 1}},
		{name: "spaces and newline in a name", customers: []string{"Acme Corp"}, text: "Acme \n Corp", want: constants.MaskCustomer,
			counts: maskCounts{Customers: 1}},
		{name: "word joiner inside a name", customers: []string{"Acme"}, text: "A\u2060cme", want: constants.MaskCustomer,
			counts: maskCounts{Customers: 1}},
		{name: "multi-word name keeps word edges", customers: []string{"Acme Corp"}, text: "Acme Corporation", want: "Acme Corporation"},
		{name: "two names split by a newline", customers: []string{"Acme", "Beta"}, text: "Acme\nBeta",
			want: constants.MaskCustomer + "\n" + constants.MaskCustomer, counts: maskCounts{Customers: 2}},
		{name: "all rules together", customers: []string{"Acme"}, extra: []string{"Jane Roe"},
			text:   "Acme: jane roe <jane@acme.com> +84 912 345 678",
			want:   constants.MaskCustomer + ": " + constants.MaskCustomer + " <" + constants.MaskEmail + "> " + constants.MaskPhone,
			counts: maskCounts{Emails: 1, Customers: 2, Phones: 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, counts := maskBrief(tt.text, tt.customers, tt.extra)
			assert.Equal(t, tt.want, got)
			assert.Equal(t, tt.counts, counts)
		})
	}
}

func TestMaskBrief_originalNeverInOutput(t *testing.T) {
	got, _ := maskBrief("Acme asks. Mail ana.lee@acme.com, Jane Roe, 0912 345 678.", []string{"Acme"}, []string{"Jane Roe"})
	for _, secret := range []string{"Acme", "acme", "ana.lee", "Jane", "Roe", "0912", "345"} {
		assert.NotContains(t, got, secret)
	}
}

func TestNormalizeForQuote(t *testing.T) {
	assert.Equal(t, "a b c", normalizeForQuote("  A \n\tB  c "))
	assert.Equal(t, "ab", normalizeForQuote("a\u200bb"))
	assert.Equal(t, "", normalizeForQuote(" \n"))
}

func TestValidateSuggestion(t *testing.T) {
	catalogue := []models.Expertise{
		{BaseModel: models.BaseModel{ID: "e1"}, Name: "Engineer"},
		{BaseModel: models.BaseModel{ID: "e2"}, Name: "QA"},
	}
	const text = "We need 400 hours of engineering for [customer].\nQA: 100h at level 3."
	str := func(s string) *string { return &s }
	num := func(f float64) *float64 { return &f }
	level := func(i int) *int { return &i }
	eng := func(mod func(*rawDemand)) rawSuggestion {
		d := rawDemand{ExpertiseName: "Engineer", TotalHours: 400, Quote: "400 hours of engineering"}
		if mod != nil {
			mod(&d)
		}
		return rawSuggestion{Name: "Portal", Demands: []rawDemand{d}}
	}
	tests := []struct {
		name     string
		raw      rawSuggestion
		check    func(t *testing.T, s Suggestion)
		warnings []string
	}{
		{name: "valid demand kept", raw: eng(nil), check: func(t *testing.T, s Suggestion) {
			require.Len(t, s.Demands, 1)
			assert.Equal(t, "e1", s.Demands[0].ExpertiseID)
			assert.Equal(t, "Engineer", s.Demands[0].ExpertiseName)
			assert.Equal(t, 400.0, s.Demands[0].TotalHours)
		}},
		{name: "quote missing dropped with warning", raw: eng(func(d *rawDemand) { d.Quote = "900 hours" }),
			warnings: []string{constants.WarnDemandNoQuote}, check: noDemands},
		{name: "empty quote dropped", raw: eng(func(d *rawDemand) { d.Quote = "  " }),
			warnings: []string{constants.WarnDemandNoQuote}, check: noDemands},
		{name: "quote with a placeholder kept when in masked text",
			raw: eng(func(d *rawDemand) { d.Quote = "engineering for [customer]" }), check: oneDemand},
		{name: "quote differs only in case and spacing: kept",
			raw: eng(func(d *rawDemand) { d.Quote = "400  HOURS\nof Engineering" }), check: oneDemand},
		{name: "unknown expertise dropped with warning, no expertise created",
			raw:      eng(func(d *rawDemand) { d.ExpertiseName = "Designer" }),
			warnings: []string{constants.WarnDemandUnknownExpertise}, check: noDemands},
		{name: "expertise name matched case-insensitively",
			raw: eng(func(d *rawDemand) { d.ExpertiseName = " qa "; d.Quote = "QA: 100h" }), check: func(t *testing.T, s Suggestion) {
				require.Len(t, s.Demands, 1)
				assert.Equal(t, "e2", s.Demands[0].ExpertiseID)
				assert.Equal(t, "QA", s.Demands[0].ExpertiseName)
			}},
		{name: "hours zero dropped", raw: eng(func(d *rawDemand) { d.TotalHours = 0 }),
			warnings: []string{constants.WarnDemandHours}, check: noDemands},
		{name: "hours negative dropped", raw: eng(func(d *rawDemand) { d.TotalHours = -5 }),
			warnings: []string{constants.WarnDemandHours}, check: noDemands},
		{name: "hours above 20000 dropped", raw: eng(func(d *rawDemand) { d.TotalHours = 20001 }),
			warnings: []string{constants.WarnDemandHours}, check: noDemands},
		{name: "hours NaN dropped", raw: eng(func(d *rawDemand) { d.TotalHours = math.NaN() }),
			warnings: []string{constants.WarnDemandHours}, check: noDemands},
		{name: "min level 0 becomes null", raw: eng(func(d *rawDemand) { d.MinLevel = level(0) }),
			check: func(t *testing.T, s Suggestion) { require.Len(t, s.Demands, 1); assert.Nil(t, s.Demands[0].MinLevel) }},
		{name: "min level 6 becomes null", raw: eng(func(d *rawDemand) { d.MinLevel = level(6) }),
			check: func(t *testing.T, s Suggestion) { require.Len(t, s.Demands, 1); assert.Nil(t, s.Demands[0].MinLevel) }},
		{name: "min level 3 kept", raw: eng(func(d *rawDemand) { d.MinLevel = level(3) }),
			check: func(t *testing.T, s Suggestion) {
				require.Len(t, s.Demands, 1)
				assert.Equal(t, 3, *s.Demands[0].MinLevel)
			}},
		{name: "deadline before start nulls both, warns",
			raw:      rawSuggestion{Name: "P", StartEarliest: str("2026-11-10"), Deadline: str("2026-11-01")},
			warnings: []string{constants.WarnDatesOrder}, check: func(t *testing.T, s Suggestion) {
				assert.Nil(t, s.StartEarliest)
				assert.Nil(t, s.Deadline)
			}},
		{name: "dates in order kept", raw: rawSuggestion{Name: "P", StartEarliest: str("2026-11-01"), Deadline: str("2026-11-01")},
			check: func(t *testing.T, s Suggestion) {
				assert.Equal(t, "2026-11-01", *s.StartEarliest)
				assert.Equal(t, "2026-11-01", *s.Deadline)
			}},
		{name: "bad date format nulls the date",
			raw: rawSuggestion{Name: "P", StartEarliest: str("11/01/2026"), Deadline: str("2026-02-30")},
			check: func(t *testing.T, s Suggestion) {
				assert.Nil(t, s.StartEarliest)
				assert.Nil(t, s.Deadline)
			}},
		{name: "scope points negative becomes null", raw: rawSuggestion{Name: "P", ScopePoints: num(-1)},
			check: func(t *testing.T, s Suggestion) { assert.Nil(t, s.ScopePoints) }},
		{name: "scope points zero kept", raw: rawSuggestion{Name: "P", ScopePoints: num(0)},
			check: func(t *testing.T, s Suggestion) { require.NotNil(t, s.ScopePoints) }},
		{name: "name over 200 characters trimmed", raw: rawSuggestion{Name: "  " + strings.Repeat("é", 300)},
			check: func(t *testing.T, s Suggestion) {
				assert.Equal(t, constants.IntakeMaxNameChars, utf8.RuneCountInString(s.Name))
			}},
		{name: "placeholder in name keeps it and warns", raw: rawSuggestion{Name: "[customer] portal"},
			warnings: []string{constants.WarnPlaceholderInName},
			check:    func(t *testing.T, s Suggestion) { assert.Equal(t, "[customer] portal", s.Name) }},
		{name: "duplicate expertise kept with a warning", raw: rawSuggestion{Name: "P", Demands: []rawDemand{
			{ExpertiseName: "Engineer", TotalHours: 400, Quote: "400 hours"},
			{ExpertiseName: "engineer", TotalHours: 100, Quote: "engineering"}}},
			warnings: []string{constants.WarnDemandDuplicateExpertise},
			check:    func(t *testing.T, s Suggestion) { assert.Len(t, s.Demands, 2) }},
		{name: "empty input is total", raw: rawSuggestion{}, check: noDemands},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, warnings := validateSuggestion(tt.raw, text, catalogue)
			assert.ElementsMatch(t, tt.warnings, warnings)
			tt.check(t, got)
		})
	}
	t.Run("nil catalogue does not panic", func(t *testing.T) {
		got, warnings := validateSuggestion(eng(nil), text, nil)
		assert.Empty(t, got.Demands)
		assert.Equal(t, []string{constants.WarnDemandUnknownExpertise}, warnings)
	})
}

func noDemands(t *testing.T, s Suggestion) { assert.Empty(t, s.Demands) }
func oneDemand(t *testing.T, s Suggestion) { assert.Len(t, s.Demands, 1) }
