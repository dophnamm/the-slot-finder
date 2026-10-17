package services

import (
	"strings"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const peopleHeader = "name,email,role,weekly_hours,non_project_hours,start_date,end_date,primary_expertise,primary_level,other_skills\n"

var peopleExpertises = map[string]string{"qa": "qa-id", "engineer": "eng-id", "analyst": "an-id"}

func parsePeople(t *testing.T, csv string) ([]peopleRow, []RowProblem) {
	t.Helper()
	rows, problems, err := parsePeopleCSV([]byte(csv), peopleImportLimits{MaxRows: 100}, peopleExpertises)
	require.NoError(t, err)
	return rows, problems
}

func TestParsePeopleCSV_Rows(t *testing.T) {
	role, nonProject := "manager", 4.0
	start := time.Date(2026, 11, 2, 0, 0, 0, 0, time.UTC)
	end := time.Date(2027, 6, 30, 0, 0, 0, 0, time.UTC)
	minimal := peopleRow{Line: 2, Name: "Quinn", Email: "quinn@x.io", WeeklyHours: 36,
		Skills: []SkillInput{{ExpertiseID: "qa-id", Level: 3, Primary: true}}}
	tests := []struct {
		name string
		csv  string
		want peopleRow
	}{
		{name: "full row", csv: peopleHeader + "Quinn,QUINN@x.io,manager,36,4,2026-11-02,2027-06-30,qa,3,Engineer:2;Analyst:1\n",
			want: peopleRow{Line: 2, Name: "Quinn", Email: "quinn@x.io", Role: &role, WeeklyHours: 36, NonProjectHours: &nonProject,
				StartDate: &start, EndDate: &end, Skills: []SkillInput{{"qa-id", 3, true}, {"eng-id", 2, false}, {"an-id", 1, false}}}},
		{name: "required cells only", csv: "name,email,weekly_hours,primary_expertise,primary_level\nQuinn,quinn@x.io,36,QA,3\n", want: minimal},
		{name: "header case and spaces", csv: "Name, EMAIL ,Weekly_Hours,Primary_Expertise,PRIMARY_LEVEL\nQuinn,quinn@x.io,36,qa,3\n", want: minimal},
		{name: "unknown column ignored", csv: "name,team,email,weekly_hours,primary_expertise,primary_level\nQuinn,core,quinn@x.io,36,qa,3\n", want: minimal},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, problems := parsePeople(t, tt.csv)
			require.Empty(t, problems)
			require.Len(t, rows, 1)
			assert.Equal(t, tt.want, rows[0])
		})
	}
}

func TestParsePeopleCSV_Problems(t *testing.T) {
	// cellsOf builds a data row from the default one with one column replaced.
	cellsOf := func(column, value string) string {
		cells := strings.Split("Quinn,quinn@x.io,,36,,,,qa,3,", ",")
		for i, name := range constants.PeopleImportColumns {
			if name == column {
				cells[i] = value
			}
		}
		return strings.Join(cells, ",") + "\n"
	}
	tests := []struct {
		name       string
		csv        string
		wantLine   int
		wantColumn string
		wantCode   string
	}{
		{"missing email column", "name,weekly_hours,primary_expertise,primary_level\nQuinn,36,qa,3\n", 1,
			constants.PeopleImportColumnEmail, constants.PeopleImportColumnMissing},
		{"empty name", peopleHeader + cellsOf("name", ""), 2, constants.PeopleImportColumnName, constants.PeopleImportNameInvalid},
		{"long name", peopleHeader + cellsOf("name", strings.Repeat("a", 201)), 2, constants.PeopleImportColumnName, constants.PeopleImportNameInvalid},
		{"bad email", peopleHeader + cellsOf("email", "not-an-email"), 2, constants.PeopleImportColumnEmail, constants.PeopleImportEmailInvalid},
		{"duplicate email", peopleHeader + cellsOf("name", "A") + cellsOf("name", "B"), 3, constants.PeopleImportColumnEmail, constants.PeopleImportRowDuplicate},
		{"bad role", peopleHeader + cellsOf("role", "boss"), 2, constants.PeopleImportColumnRole, constants.PeopleImportRoleInvalid},
		{"zero hours", peopleHeader + cellsOf("weekly_hours", "0"), 2, constants.PeopleImportColumnWeeklyHours, constants.PeopleImportHoursInvalid},
		{"81 hours", peopleHeader + cellsOf("weekly_hours", "81"), 2, constants.PeopleImportColumnWeeklyHours, constants.PeopleImportHoursInvalid},
		{"text hours", peopleHeader + cellsOf("weekly_hours", "abc"), 2, constants.PeopleImportColumnWeeklyHours, constants.PeopleImportHoursInvalid},
		{"negative non-project", peopleHeader + cellsOf("non_project_hours", "-1"), 2, constants.PeopleImportColumnNonProjectHours, constants.PeopleImportHoursInvalid},
		{"bad start date", peopleHeader + cellsOf("start_date", "2026-13-01"), 2, constants.PeopleImportColumnStartDate, constants.PeopleImportDateInvalid},
		{"end before start", peopleHeader + "Quinn,quinn@x.io,,36,,2026-11-02,2026-10-01,qa,3,\n", 2,
			constants.PeopleImportColumnEndDate, constants.PeopleImportDateInvalid},
		{"unknown primary expertise", peopleHeader + cellsOf("primary_expertise", "Pilot"), 2, constants.PeopleImportColumnPrimaryExpertise, constants.PeopleImportExpertiseUnknown},
		{"bad primary level", peopleHeader + cellsOf("primary_level", "6"), 2, constants.PeopleImportColumnPrimaryLevel, constants.PeopleImportLevelInvalid},
		{"skill without level", peopleHeader + cellsOf("other_skills", "QA"), 2, constants.PeopleImportColumnOtherSkills, constants.PeopleImportSkillsInvalid},
		{"skill bad level", peopleHeader + cellsOf("other_skills", "QA:9"), 2, constants.PeopleImportColumnOtherSkills, constants.PeopleImportLevelInvalid},
		{"skill unknown", peopleHeader + cellsOf("other_skills", "Pilot:2"), 2, constants.PeopleImportColumnOtherSkills, constants.PeopleImportExpertiseUnknown},
		{"skill repeats primary", peopleHeader + cellsOf("other_skills", "QA:2"), 2, constants.PeopleImportColumnOtherSkills, constants.PeopleImportSkillDuplicate},
		{"malformed line", peopleHeader + "\"Quinn,quinn@x.io\n", 2, "", constants.PeopleImportCSVMalformed},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, problems := parsePeople(t, tt.csv)
			assert.Equal(t, []RowProblem{{Line: tt.wantLine, Column: tt.wantColumn, Code: tt.wantCode}}, problems)
		})
	}
}

func TestParsePeopleCSV_TooManyRows(t *testing.T) {
	csv := peopleHeader + "A,a@x.io,,36,,,,qa,3,\nB,b@x.io,,36,,,,qa,3,\nC,c@x.io,,36,,,,qa,3,\n"
	_, _, err := parsePeopleCSV([]byte(csv), peopleImportLimits{MaxRows: 2}, peopleExpertises)
	assert.ErrorIs(t, err, errTooManyPeopleRows)
}
