package services

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"the-slot-finder/server/internal/constants"
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
