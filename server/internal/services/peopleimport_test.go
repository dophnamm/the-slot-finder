package services

import (
	"context"
	"encoding/csv"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	piQA       = "0192f3a4-0000-7000-8000-00000000f101"
	piEngineer = "0192f3a4-0000-7000-8000-00000000f102"
	piQuinnID  = "0192f3a4-0000-7000-8000-00000000f201"
	piHeader   = "name,email,role,weekly_hours,non_project_hours,start_date,end_date,primary_expertise,primary_level,other_skills\n"
)

// recordingPeople records the saves the import asks for; every other method panics (nil embed).
type recordingPeople struct {
	PeopleService
	expertises []models.Expertise
	created    []CreatePersonInput
	updated    map[string]UpdatePersonInput
	replaced   map[string][]SkillInput
	updateErr  error
}

func newRecordingPeople() *recordingPeople {
	return &recordingPeople{
		expertises: []models.Expertise{{BaseModel: models.BaseModel{ID: piQA}, Name: "QA"},
			{BaseModel: models.BaseModel{ID: piEngineer}, Name: "Engineer"}},
		updated:  map[string]UpdatePersonInput{},
		replaced: map[string][]SkillInput{},
	}
}

func (p *recordingPeople) ListExpertises(context.Context) ([]models.Expertise, error) {
	return p.expertises, nil
}

func (p *recordingPeople) Create(_ context.Context, in CreatePersonInput) (*models.Person, error) {
	p.created = append(p.created, in)
	return &models.Person{}, nil
}

func (p *recordingPeople) Update(_ context.Context, id string, in UpdatePersonInput) (*models.Person, error) {
	p.updated[id] = in
	return &models.Person{}, p.updateErr
}

func (p *recordingPeople) ReplaceSkills(_ context.Context, id string, skills []SkillInput) (*models.Person, error) {
	p.replaced[id] = skills
	return &models.Person{}, nil
}

func (p *recordingPeople) saves() int { return len(p.created) + len(p.updated) + len(p.replaced) }

func peopleImportConfig() *config.Config {
	return &config.Config{PeopleImportMaxBytes: 1 << 10, PeopleImportMaxRows: 3, DefaultNonProjectHours: 4}
}

func quinn() models.Person {
	start := time.Date(2026, time.November, 2, 0, 0, 0, 0, time.UTC)
	return models.Person{BaseModel: models.BaseModel{ID: piQuinnID}, Name: "Quinn", Email: "quinn@example.test",
		Role: constants.RoleSpecialist, WeeklyHours: 40, NonProjectHours: 4, StartDate: &start,
		Skills: []models.PersonSkill{{PersonID: piQuinnID, ExpertiseID: piQA, Level: 3, IsPrimary: true},
			{PersonID: piQuinnID, ExpertiseID: piEngineer, Level: 2}}}
}

// quinnRow is Quinn's stored values as a file row, with replacements for any cell.
func quinnRow(weekly, end, level, other string) string {
	return "Quinn,quinn@example.test,specialist," + weekly + ",4,2026-11-02," + end + ",QA," + level + "," + other + "\n"
}

func TestPeopleImportService_Import(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	run := func(ctx context.Context, in ImportInput, stored ...models.Person) (*recordingPeople, *models.PeopleImportResult, error) {
		people, repo := newRecordingPeople(), new(MockPersonRepository)
		repo.On("ListByEmails", mock.Anything, mock.Anything).Return(stored, nil)
		for _, p := range stored {
			repo.On("FindWithSkills", mock.Anything, p.ID).Return(&p, nil)
		}
		result, err := ProvidePeopleImportService(people, repo, ProvideAuthzService(leadPeople{}), passthroughTx{},
			peopleImportConfig()).Import(ctx, in)
		return people, result, err
	}
	wantProblem := func(t *testing.T, err error, want RowProblem) {
		t.Helper()
		var appErr *errors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, constants.PeopleImportInvalid, appErr.Code)
		assert.Contains(t, appErr.Context[constants.PeopleImportProblemsKey], want)
	}
	wantCode := func(t *testing.T, err error, code string, status int) {
		t.Helper()
		var appErr *errors.AppError
		require.ErrorAs(t, err, &appErr)
		assert.Equal(t, code, appErr.Code)
		assert.Equal(t, status, appErr.HTTPStatus)
	}
	other := "Engineer:2"

	t.Run("creates a new person", func(t *testing.T) {
		people, result, err := run(planner, importFile(piHeader+"Ada,ADA@example.test,,40,,,,QA,3,Engineer:2\n"))
		require.NoError(t, err)
		assert.Equal(t, models.PeopleImportResult{Created: 1}, *result)
		require.Len(t, people.created, 1)
		got := people.created[0]
		assert.Equal(t, constants.RoleSpecialist, got.Role)
		assert.Nil(t, got.NonProjectHours)
		assert.Equal(t, "ada@example.test", got.Email)
		assert.Equal(t, []SkillInput{{ExpertiseID: piQA, Level: 3, Primary: true}, {ExpertiseID: piEngineer, Level: 2}}, got.Skills)
	})
	t.Run("updates a known person", func(t *testing.T) {
		people, result, err := run(planner, importFile(piHeader+"Quinn,quinn@example.test,,36,,,,QA,3,"+other+"\n"), quinn())
		require.NoError(t, err)
		assert.Equal(t, models.PeopleImportResult{Updated: 1}, *result)
		assert.Equal(t, UpdatePersonInput{WeeklyHours: hoursOf(36)}, people.updated[piQuinnID])
		assert.Empty(t, people.replaced)
	})
	t.Run("a changed level replaces the skills", func(t *testing.T) {
		people, result, err := run(planner, importFile(piHeader+quinnRow("40", "", "4", other)), quinn())
		require.NoError(t, err)
		assert.Equal(t, models.PeopleImportResult{Updated: 1}, *result)
		assert.Empty(t, people.updated)
		assert.Equal(t, []SkillInput{{ExpertiseID: piQA, Level: 4, Primary: true}, {ExpertiseID: piEngineer, Level: 2}},
			people.replaced[piQuinnID])
	})
	t.Run("empty other_skills removes secondary skills", func(t *testing.T) {
		people, _, err := run(planner, importFile(piHeader+quinnRow("40", "", "3", "")), quinn())
		require.NoError(t, err)
		assert.Equal(t, []SkillInput{{ExpertiseID: piQA, Level: 3, Primary: true}}, people.replaced[piQuinnID])
	})
	t.Run("same file twice", func(t *testing.T) {
		people, result, err := run(planner, importFile(piHeader+quinnRow("40", "", "3", other)), quinn())
		require.NoError(t, err)
		assert.Equal(t, models.PeopleImportResult{Unchanged: 1}, *result)
		assert.Zero(t, people.saves())
	})
	t.Run("empty end_date keeps the stored end date", func(t *testing.T) {
		people, _, err := run(planner, importFile(piHeader+quinnRow("36", "", "3", other)), quinn())
		require.NoError(t, err)
		assert.False(t, people.updated[piQuinnID].EndDate.Set)
	})
	t.Run("a filled end_date is set", func(t *testing.T) {
		people, _, err := run(planner, importFile(piHeader+quinnRow("40", "2026-12-18", "3", other)), quinn())
		require.NoError(t, err)
		end := time.Date(2026, time.December, 18, 0, 0, 0, 0, time.UTC)
		assert.Equal(t, DateChange{Set: true, Value: &end}, people.updated[piQuinnID].EndDate)
	})
	t.Run("end before the stored start", func(t *testing.T) {
		people, _, err := run(planner, importFile(piHeader+quinnRow("40", "2026-10-01", "3", other)), quinn())
		wantProblem(t, err, RowProblem{Line: 2, Column: constants.PeopleImportColumnEndDate, Code: constants.PeopleImportDateInvalid})
		assert.Zero(t, people.saves())
	})
	t.Run("non-project hours above the row's weekly hours", func(t *testing.T) {
		people, _, err := run(planner, importFile(piHeader+"Quinn,quinn@example.test,,3,,,,QA,3,"+other+"\n"), quinn())
		wantProblem(t, err, RowProblem{Line: 2, Column: constants.PeopleImportColumnNonProjectHours, Code: constants.PeopleImportHoursInvalid})
		assert.Zero(t, people.saves())
	})
	t.Run("one bad row saves nothing", func(t *testing.T) {
		people, _, err := run(planner, importFile(piHeader+"Ada,ada@example.test,,40,,,,QA,3,\nBo,bo@example.test,,40,,,,QA,3,\nCy,cy@example.test,,40,,,,Nope,3,\n"))
		wantProblem(t, err, RowProblem{Line: 4, Column: constants.PeopleImportColumnPrimaryExpertise, Code: constants.PeopleImportExpertiseUnknown})
		assert.Zero(t, people.saves())
	})
	t.Run("a file over the size limit", func(t *testing.T) {
		_, _, err := run(planner, ImportInput{File: strings.NewReader(""), Size: 1 << 20})
		wantCode(t, err, constants.PeopleImportTooLarge, http.StatusRequestEntityTooLarge)
		_, _, err = run(planner, ImportInput{File: strings.NewReader(strings.Repeat("x", 2<<10)), Size: 1})
		wantCode(t, err, constants.PeopleImportTooLarge, http.StatusRequestEntityTooLarge)
	})
	t.Run("too many rows", func(t *testing.T) {
		row := "Ada,ada@example.test,,40,,,,QA,3,\n"
		_, _, err := run(planner, importFile(piHeader+strings.Repeat(row, 4)))
		wantCode(t, err, constants.PeopleImportTooLarge, http.StatusRequestEntityTooLarge)
	})
	t.Run("a specialist is refused", func(t *testing.T) {
		in := importFile(piHeader)
		_, _, err := run(asRole("kc-specialist", constants.RoleSpecialist), in)
		wantCode(t, err, constants.Forbidden, http.StatusForbidden)
	})
	t.Run("a save error rolls back", func(t *testing.T) {
		people, repo := newRecordingPeople(), new(MockPersonRepository)
		people.updateErr = assert.AnError
		stored := quinn()
		repo.On("ListByEmails", mock.Anything, mock.Anything).Return([]models.Person{stored}, nil)
		repo.On("FindWithSkills", mock.Anything, piQuinnID).Return(&stored, nil)
		_, err := ProvidePeopleImportService(people, repo, ProvideAuthzService(leadPeople{}), passthroughTx{},
			peopleImportConfig()).Import(planner, importFile(piHeader+quinnRow("36", "", "3", other)))
		require.ErrorIs(t, err, assert.AnError)
	})
}

func TestPeopleImportService_Template(t *testing.T) {
	people := newRecordingPeople()
	svc := ProvidePeopleImportService(people, new(MockPersonRepository), ProvideAuthzService(leadPeople{}), passthroughTx{},
		peopleImportConfig())
	data, err := svc.Template(asRole("kc-planner", constants.RolePlanner))
	require.NoError(t, err)
	records, err := csv.NewReader(strings.NewReader(string(data))).ReadAll()
	require.NoError(t, err)
	require.Len(t, records, 2)
	assert.Equal(t, constants.PeopleImportColumns, records[0])
	assert.Equal(t, "QA", records[1][slices.Index(constants.PeopleImportColumns, constants.PeopleImportColumnPrimaryExpertise)])
	rows, problems, err := parsePeopleCSV(data, peopleImportLimits{MaxRows: 3}, map[string]string{"qa": piQA})
	require.NoError(t, err)
	assert.Empty(t, problems)
	assert.Len(t, rows, 1)

	_, err = svc.Template(asRole("kc-specialist", constants.RoleSpecialist))
	require.Error(t, err)
}

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
