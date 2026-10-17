package services

import (
	"fmt"
	"math"
	"regexp"
	"strings"
)

// numberPattern matches an integer or decimal, including the digits inside a week label.
var numberPattern = regexp.MustCompile(`\d+(?:\.\d+)?`)

// draftFacts is the only input the LLM sees. No names, emails or customer names.
type draftFacts struct {
	People       int         `json:"people"`
	Assignments  int         `json:"assignments"`
	Projects     int         `json:"projects"`
	PlannedHours float64     `json:"planned_hours"`
	HoursAdded   float64     `json:"hours_added"`
	HoursRemoved float64     `json:"hours_removed"`
	IsReplan     bool        `json:"is_replan"`
	Unmet        []unmetFact `json:"unmet"`
}

// unmetFact is one demand with unmet hours.
type unmetFact struct {
	Expertise string  `json:"expertise"`
	Project   string  `json:"project"`
	Hours     float64 `json:"hours"`
	FirstWeek string  `json:"first_week"`
	LastWeek  string  `json:"last_week"`
}

// buildDraftFacts reduces a draft to counts, hours and unmet demand. Unmet hours are the demand's
// hours minus what the draft assigns now, for every demand left short, by the engine or by hand.
func buildDraftFacts(view *DraftView) draftFacts {
	f := draftFacts{Projects: len(view.Projects), IsReplan: len(view.Events) > 0, Unmet: []unmetFact{}}
	people := map[string]bool{}
	assigned := map[string]float64{}
	for _, a := range view.Plan.Assignments {
		people[a.PersonID] = true
		assigned[a.DemandID] += a.TotalHours()
		f.PlannedHours += a.TotalHours()
	}
	f.People, f.Assignments = len(people), len(view.Plan.Assignments)
	for _, c := range view.Changes {
		if d := c.After - c.Before; d > 0 {
			f.HoursAdded += d
		} else {
			f.HoursRemoved -= d
		}
	}
	for _, p := range view.Projects {
		for _, d := range p.Demands {
			hours := d.TotalHours - assigned[d.ID]
			if hours <= hoursEpsilon {
				continue
			}
			from, to := p.DemandWindow(d)
			f.Unmet = append(f.Unmet, unmetFact{Expertise: d.Expertise.Name, Project: p.Name, Hours: round2(hours),
				FirstWeek: from.String(), LastWeek: to.String()})
		}
	}
	f.PlannedHours, f.HoursAdded, f.HoursRemoved = round2(f.PlannedHours), round2(f.HoursAdded), round2(f.HoursRemoved)
	return f
}

// templateSummary is the plain-English fallback written from the facts alone.
func templateSummary(f draftFacts) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%s on %s across %s, %s planned hours.", count(f.People, "person", "people"),
		count(f.Assignments, "assignment", "assignments"), count(f.Projects, "project", "projects"),
		formatHours(f.PlannedHours))
	if len(f.Unmet) == 0 {
		b.WriteString(" All demand is covered.")
	}
	for _, u := range f.Unmet {
		fmt.Fprintf(&b, " %s h of %s is still uncovered on %s (%s to %s).", formatHours(u.Hours), u.Expertise, u.Project,
			u.FirstWeek, u.LastWeek)
	}
	return b.String()
}

// numbersIn lists the numbers in text, in order.
func numbersIn(text string) []string { return numberPattern.FindAllString(text, -1) }

// numbersKnown reports whether every number in text also appears as a number in the facts JSON.
func numbersKnown(text string, factsJSON string) bool {
	known := map[string]bool{}
	for _, n := range numbersIn(factsJSON) {
		known[n] = true
	}
	for _, n := range numbersIn(text) {
		if !known[n] {
			return false
		}
	}
	return true
}

// round2 rounds hours to hundredths, as formatHours shows them.
func round2(v float64) float64 { return math.Round(v*hourPrecision) / hourPrecision }

// count writes n with the singular noun for 1 and the plural otherwise.
func count(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}
