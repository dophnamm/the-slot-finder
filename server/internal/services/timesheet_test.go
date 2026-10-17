package services

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockActualHoursRepository is a testify mock of repositories.ActualHoursRepository.
type MockActualHoursRepository struct {
	mock.Mock
}

func (m *MockActualHoursRepository) Upsert(ctx context.Context, rows []models.ActualHours) (int, int, error) {
	args := m.Called(ctx, rows)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockActualHoursRepository) SumByProject(ctx context.Context, projectIDs []string) (map[string]float64, error) {
	args := m.Called(ctx, projectIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MockActualHoursRepository) ByPersonWeeks(ctx context.Context, personID string, from, to isoweek.Week) (map[isoweek.Week]float64, error) {
	args := m.Called(ctx, personID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[isoweek.Week]float64), args.Error(1)
}

func (m *MockActualHoursRepository) ListInRange(ctx context.Context, from, to isoweek.Week) ([]models.ActualHours, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).([]models.ActualHours), args.Error(1)
}

const (
	tsProjectID = "0192f3a4-0000-7000-8000-00000000e101"
	tsHeader    = "person_email,person_name,program,project,project_id,week,planned_hours,actual_hours\n"
)

// tsLine is one template line for Ada on the test project.
func tsLine(week, actual string) string {
	return "ada@example.test,Ada,Acme,Checkout," + tsProjectID + "," + week + ",20," + actual + "\n"
}

// tsLimits allow weeks up to 2026-W42, 80 hours a row and 3 rows.
func tsLimits(t *testing.T) timesheetLimits {
	t.Helper()
	return timesheetLimits{Current: mustWeek(t, "2026-W42"), MaxHours: 80, MaxRows: 3}
}

func hoursOf(v float64) *float64 { return &v }

func TestParseTimesheet_Rows(t *testing.T) {
	w40, w41, w42 := mustWeek(t, "2026-W40"), mustWeek(t, "2026-W41"), mustWeek(t, "2026-W42")
	tests := []struct {
		name string
		file string
		want []timesheetRow
	}{
		{
			name: "a filled-in row, an empty one and zero hours",
			file: tsHeader + tsLine("2026-W40", "18.5") + tsLine("2026-W41", "") + tsLine("2026-W42", "0"),
			want: []timesheetRow{
				{Line: 2, Email: "ada@example.test", ProjectID: tsProjectID, Week: w40, Hours: hoursOf(18.5)},
				{Line: 3, Email: "ada@example.test", ProjectID: tsProjectID, Week: w41},
				{Line: 4, Email: "ada@example.test", ProjectID: tsProjectID, Week: w42, Hours: hoursOf(0)},
			},
		},
		{
			name: "saved by Excel: byte order mark, CRLF, other column order, case and spaces",
			file: "\ufeffWeek, Actual_Hours ,PROJECT_ID,person_email\r\n2026-W41,7.5, " + strings.ToUpper(tsProjectID) +
				" , Ada@Example.TEST \r\n",
			want: []timesheetRow{{Line: 2, Email: "ada@example.test", ProjectID: tsProjectID, Week: w41, Hours: hoursOf(7.5)}},
		},
		{
			name: "a row shorter than the header has an empty actual_hours",
			file: tsHeader + "ada@example.test,Ada,Acme,Checkout," + tsProjectID + ",2026-W41,20\n",
			want: []timesheetRow{{Line: 2, Email: "ada@example.test", ProjectID: tsProjectID, Week: w41}},
		},
		{name: "a header only", file: tsHeader},
		{
			name: "blank rows a spreadsheet leaves are skipped and do not count toward the row limit",
			file: tsHeader + tsLine("2026-W40", "1") + ",,,,,,,\r\n" + tsLine("2026-W41", "2") + " , ,,\n" +
				tsLine("2026-W42", "3") + ",,,,,,,\n",
			want: []timesheetRow{
				{Line: 2, Email: "ada@example.test", ProjectID: tsProjectID, Week: w40, Hours: hoursOf(1)},
				{Line: 4, Email: "ada@example.test", ProjectID: tsProjectID, Week: w41, Hours: hoursOf(2)},
				{Line: 6, Email: "ada@example.test", ProjectID: tsProjectID, Week: w42, Hours: hoursOf(3)},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rows, problems, err := parseTimesheet([]byte(tt.file), tsLimits(t))
			require.NoError(t, err)
			assert.Empty(t, problems)
			assert.Equal(t, tt.want, rows)
		})
	}
}

func TestParseTimesheet_Problems(t *testing.T) {
	missing := func(columns ...string) []RowProblem {
		out := make([]RowProblem, 0, len(columns))
		for _, column := range columns {
			out = append(out, RowProblem{Line: 1, Column: column, Code: constants.TimesheetColumnMissing})
		}
		return out
	}
	hours := func(line int) RowProblem {
		return RowProblem{Line: line, Column: constants.TimesheetColumnActualHours, Code: constants.TimesheetHoursInvalid}
	}
	tests := []struct {
		name string
		file string
		want []RowProblem
	}{
		{name: "an empty file", file: "", want: missing(constants.TimesheetRequiredColumns...)},
		{
			name: "a semicolon-separated file",
			file: "person_email;project_id;week;actual_hours\n",
			want: missing(constants.TimesheetRequiredColumns...),
		},
		{
			name: "a header without actual_hours",
			file: "person_email,project_id,week\n",
			want: missing(constants.TimesheetColumnActualHours),
		},
		{
			name: "every check of one row",
			file: tsHeader + ",Ada,Acme,Checkout,not-a-uuid,2026-W99,20,abc\n",
			want: []RowProblem{
				{Line: 2, Column: constants.TimesheetColumnPersonEmail, Code: constants.TimesheetPersonUnknown},
				{Line: 2, Column: constants.TimesheetColumnProjectID, Code: constants.TimesheetProjectUnknown},
				{Line: 2, Column: constants.TimesheetColumnWeek, Code: constants.TimesheetWeekInvalid},
				hours(2),
			},
		},
		{
			name: "a future week, too many hours and a comma decimal",
			file: tsHeader + tsLine("2026-W43", "8") + tsLine("2026-W41", "80.5") + tsLine("2026-W40", `"7,5"`),
			want: []RowProblem{
				{Line: 2, Column: constants.TimesheetColumnWeek, Code: constants.TimesheetWeekFuture},
				hours(3),
				hours(4),
			},
		},
		{
			name: "negative and not-a-number hours",
			file: tsHeader + tsLine("2026-W41", "-1") + tsLine("2026-W40", "NaN"),
			want: []RowProblem{hours(2), hours(3)},
		},
		{
			name: "the same person, project and week twice, in another case",
			file: tsHeader + tsLine("2026-W41", "8") + strings.Replace(tsLine("2026-W41", ""), "ada@", "ADA@", 1),
			want: []RowProblem{{Line: 3, Code: constants.TimesheetRowDuplicate}},
		},
		{
			name: "a broken quote",
			file: tsHeader + "\"ada@example.test,Ada\n",
			want: []RowProblem{{Line: 2, Code: constants.TimesheetCSVMalformed}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, problems, err := parseTimesheet([]byte(tt.file), tsLimits(t))
			require.NoError(t, err)
			assert.Equal(t, tt.want, problems)
		})
	}
}

func TestParseTimesheet_TooManyRows(t *testing.T) {
	file := tsHeader + tsLine("2026-W38", "1") + tsLine("2026-W39", "1") + tsLine("2026-W40", "1") + tsLine("2026-W41", "1")
	_, _, err := parseTimesheet([]byte(file), tsLimits(t))
	assert.ErrorIs(t, err, errTooManyRows)
}

const tsPersonID = "0192f3a4-0000-7000-8000-00000000e201"

// tsNow is a Wednesday in 2026-W42.
var tsNow = time.Date(2026, time.October, 14, 9, 0, 0, 0, time.UTC)

func timesheetConfig() *config.Config {
	return &config.Config{TimesheetTemplateWeeks: 4, TimesheetMaxWeeks: 12, TimesheetMaxWeekHours: 80,
		TimesheetMaxBytes: 1 << 10, TimesheetMaxRows: 3}
}

func newTimesheetService(actuals *MockActualHoursRepository, people *MockPersonRepository, projects *MockProjectRepository) TimesheetService {
	return ProvideTimesheetService(actuals, people, projects, nil, ProvideAuthzService(leadPeople{}), passthroughTx{},
		clock.NewFixed(tsNow), timesheetConfig())
}

func importFile(file string) ImportInput {
	return ImportInput{File: strings.NewReader(file), Size: int64(len(file))}
}

func TestTimesheetService_Import(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	ada := models.Person{BaseModel: models.BaseModel{ID: tsPersonID}, Email: "ada@example.test"}
	project := models.Project{BaseModel: models.BaseModel{ID: tsProjectID}}
	known := func() (*MockActualHoursRepository, *MockPersonRepository, *MockProjectRepository) {
		actuals, people, projects := new(MockActualHoursRepository), new(MockPersonRepository), new(MockProjectRepository)
		people.On("ListByEmails", mock.Anything, []string{"ada@example.test"}).Return([]models.Person{ada}, nil)
		projects.On("ListByIDs", mock.Anything, []string{tsProjectID}).Return([]models.Project{project}, nil)
		return actuals, people, projects
	}

	t.Run("saves filled-in rows and skips empty ones", func(t *testing.T) {
		actuals, people, projects := known()
		actuals.On("Upsert", mock.Anything, []models.ActualHours{{PersonID: tsPersonID, ProjectID: tsProjectID,
			Week: mustWeek(t, "2026-W41"), Hours: 18.5, UpdatedBy: "kc-planner"}}).Return(1, 0, nil)

		result, err := newTimesheetService(actuals, people, projects).Import(planner,
			importFile(tsHeader+tsLine("2026-W41", "18.5")+tsLine("2026-W42", "")))
		require.NoError(t, err)
		assert.Equal(t, models.ImportResult{Created: 1, Skipped: 1}, *result)
		actuals.AssertExpectations(t)
	})

	t.Run("an email in another case and with spaces finds the person", func(t *testing.T) {
		actuals, people, projects := known()
		actuals.On("Upsert", mock.Anything, mock.Anything).Return(0, 1, nil)
		file := tsHeader + strings.Replace(tsLine("2026-W41", "8"), "ada@example.test", " Ada@Example.TEST ", 1)

		result, err := newTimesheetService(actuals, people, projects).Import(planner, importFile(file))
		require.NoError(t, err)
		assert.Equal(t, models.ImportResult{Updated: 1}, *result)
		people.AssertExpectations(t)
	})

	t.Run("an unknown person or project saves nothing", func(t *testing.T) {
		actuals, people, projects := new(MockActualHoursRepository), new(MockPersonRepository), new(MockProjectRepository)
		people.On("ListByEmails", mock.Anything, mock.Anything).Return([]models.Person{}, nil)
		projects.On("ListByIDs", mock.Anything, mock.Anything).Return([]models.Project{}, nil)

		_, err := newTimesheetService(actuals, people, projects).Import(planner, importFile(tsHeader+tsLine("2026-W41", "8")))
		requireAppError(t, err, constants.TimesheetInvalid, http.StatusBadRequest)
		assert.Equal(t, []RowProblem{
			{Line: 2, Column: constants.TimesheetColumnPersonEmail, Code: constants.TimesheetPersonUnknown},
			{Line: 2, Column: constants.TimesheetColumnProjectID, Code: constants.TimesheetProjectUnknown},
		}, errors.GetAppError(err).Context[constants.TimesheetProblemsKey])
		actuals.AssertNotCalled(t, "Upsert", mock.Anything, mock.Anything)
	})

	t.Run("problems from the file and the lookups come in line order", func(t *testing.T) {
		actuals, people, projects := new(MockActualHoursRepository), new(MockPersonRepository), new(MockProjectRepository)
		people.On("ListByEmails", mock.Anything, mock.Anything).Return([]models.Person{}, nil)
		projects.On("ListByIDs", mock.Anything, mock.Anything).Return([]models.Project{project}, nil)
		file := tsHeader + tsLine("2026-W43", "8") + strings.Replace(tsLine("2026-W41", "8"), "ada@", "bo@", 1)

		_, err := newTimesheetService(actuals, people, projects).Import(planner, importFile(file))
		requireAppError(t, err, constants.TimesheetInvalid, http.StatusBadRequest)
		assert.Equal(t, []RowProblem{
			{Line: 2, Column: constants.TimesheetColumnWeek, Code: constants.TimesheetWeekFuture},
			{Line: 2, Column: constants.TimesheetColumnPersonEmail, Code: constants.TimesheetPersonUnknown},
			{Line: 3, Column: constants.TimesheetColumnPersonEmail, Code: constants.TimesheetPersonUnknown},
		}, errors.GetAppError(err).Context[constants.TimesheetProblemsKey])
	})

	t.Run("a file over the size limit, declared or not", func(t *testing.T) {
		svc := newTimesheetService(nil, nil, nil)
		big := tsHeader + strings.Repeat(" ", 1<<10)
		_, err := svc.Import(planner, importFile(big))
		requireAppError(t, err, constants.TimesheetTooLarge, http.StatusRequestEntityTooLarge)
		_, err = svc.Import(planner, ImportInput{File: strings.NewReader(big)})
		requireAppError(t, err, constants.TimesheetTooLarge, http.StatusRequestEntityTooLarge)
	})

	t.Run("more rows than allowed", func(t *testing.T) {
		file := tsHeader + tsLine("2026-W38", "1") + tsLine("2026-W39", "1") + tsLine("2026-W40", "1") + tsLine("2026-W41", "1")
		_, err := newTimesheetService(nil, nil, nil).Import(planner, importFile(file))
		requireAppError(t, err, constants.TimesheetTooLarge, http.StatusRequestEntityTooLarge)
	})

	t.Run("a database failure", func(t *testing.T) {
		actuals, people, projects := known()
		actuals.On("Upsert", mock.Anything, mock.Anything).Return(0, 0, assert.AnError)
		_, err := newTimesheetService(actuals, people, projects).Import(planner, importFile(tsHeader+tsLine("2026-W41", "8")))
		requireAppError(t, err, constants.DatabaseError, http.StatusInternalServerError)
	})

	t.Run("only planners import", func(t *testing.T) {
		_, err := newTimesheetService(nil, nil, nil).Import(asRole("kc-manager", constants.RoleManager), importFile(tsHeader))
		requireStatus(t, err, http.StatusForbidden)
	})
}

const (
	tsProgramID      = "0192f3a4-0000-7000-8000-00000000e301"
	tsOtherProjectID = "0192f3a4-0000-7000-8000-00000000e102"
)

func newTemplateService(plans *MockPlanRepository, actuals *MockActualHoursRepository) TimesheetService {
	return ProvideTimesheetService(actuals, nil, nil, plans, ProvideAuthzService(leadPeople{}), passthroughTx{},
		clock.NewFixed(tsNow), timesheetConfig())
}

func TestTimesheetService_Template(t *testing.T) {
	planner := asRole("kc-planner", constants.RolePlanner)
	program := &models.Program{BaseModel: models.BaseModel{ID: tsProgramID}, Name: "Acme"}
	checkout := &models.Project{BaseModel: models.BaseModel{ID: tsProjectID}, Name: "Checkout", Program: program}
	loyalty := &models.Project{BaseModel: models.BaseModel{ID: tsOtherProjectID}, Name: "Loyalty", Program: program}
	ada := &models.Person{BaseModel: models.BaseModel{ID: tsPersonID}, Name: "Ada", Email: "ada@example.test"}
	baseline := publishedBaseline(1)
	baseline.Assignments = []models.Assignment{{PersonID: tsPersonID, Person: ada,
		Demand: &models.Demand{ProjectID: tsProjectID, Project: checkout},
		Allocations: []models.Allocation{{Week: mustWeek(t, "2026-W38"), Hours: 40}, {Week: mustWeek(t, "2026-W41"), Hours: 20},
			{Week: mustWeek(t, "2026-W42"), Hours: 16}}}}
	recorded := []models.ActualHours{
		{PersonID: tsPersonID, ProjectID: tsProjectID, Week: mustWeek(t, "2026-W41"), Hours: 18, Person: ada, Project: checkout},
		{PersonID: tsPersonID, ProjectID: tsOtherProjectID, Week: mustWeek(t, "2026-W42"), Hours: 4.5, Person: ada, Project: loyalty},
	}
	fixture := func(from, to string) (*MockPlanRepository, *MockActualHoursRepository) {
		plans, actuals := new(MockPlanRepository), new(MockActualHoursRepository)
		plans.On("FindBaseline", mock.Anything).Return(baseline, nil)
		actuals.On("ListInRange", mock.Anything, mustWeek(t, from), mustWeek(t, to)).Return(recorded, nil)
		return plans, actuals
	}

	t.Run("the last weeks up to this one, from the baseline and the recorded hours", func(t *testing.T) {
		plans, actuals := fixture("2026-W39", "2026-W42")
		data, err := newTemplateService(plans, actuals).Template(planner, TemplateInput{})
		require.NoError(t, err)
		assert.Equal(t, tsHeader+
			"ada@example.test,Ada,Acme,Checkout,"+tsProjectID+",2026-W41,20,18\n"+
			"ada@example.test,Ada,Acme,Checkout,"+tsProjectID+",2026-W42,16,\n"+
			"ada@example.test,Ada,Acme,Loyalty,"+tsOtherProjectID+",2026-W42,0,4.5\n", string(data),
			"W38 is before the range; Loyalty has actual hours without baseline hours")
		actuals.AssertExpectations(t)
	})

	t.Run("a chosen range", func(t *testing.T) {
		plans, actuals := fixture("2026-W41", "2026-W41")
		data, err := newTemplateService(plans, actuals).Template(planner,
			TemplateInput{From: mustWeek(t, "2026-W41"), To: mustWeek(t, "2026-W41")})
		require.NoError(t, err)
		assert.Equal(t, tsHeader+"ada@example.test,Ada,Acme,Checkout,"+tsProjectID+",2026-W41,20,18\n", string(data))
	})

	ranges := []struct {
		name string
		in   TemplateInput
		code string
	}{
		{name: "a week after the current one", in: TemplateInput{To: mustWeek(t, "2026-W43")}, code: constants.TimesheetWeekInFuture},
		{name: "more weeks than allowed", in: TemplateInput{From: mustWeek(t, "2026-W20"), To: mustWeek(t, "2026-W42")},
			code: constants.WeekRangeTooLong},
		{name: "to before from", in: TemplateInput{From: mustWeek(t, "2026-W42"), To: mustWeek(t, "2026-W41")},
			code: constants.WeekRangeInvalid},
	}
	for _, tt := range ranges {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newTemplateService(nil, nil).Template(planner, tt.in)
			requireAppError(t, err, tt.code, http.StatusBadRequest)
		})
	}

	t.Run("only planners", func(t *testing.T) {
		_, err := newTemplateService(nil, nil).Template(asRole("kc-manager", constants.RoleManager), TemplateInput{})
		requireStatus(t, err, http.StatusForbidden)
	})
}

func TestWriteTemplate_FormulaCells(t *testing.T) {
	data, err := writeTemplate([]templateRow{{Email: "ada@example.test", Name: "+Ada", Program: "@Acme",
		Project: `=HYPERLINK("http://example.test")`, ProjectID: tsProjectID, Week: mustWeek(t, "2026-W41"), Planned: 20}})
	require.NoError(t, err)
	assert.Equal(t, tsHeader+`ada@example.test,'+Ada,'@Acme,"'=HYPERLINK(""http://example.test"")",`+tsProjectID+
		",2026-W41,20,\n", string(data))
}

func TestSafeCell(t *testing.T) {
	tests := map[string]string{"=SUM(A1)": "'=SUM(A1)", "+1": "'+1", "-2": "'-2", "@x": "'@x", "\tx": "'\tx", "Acme": "Acme", "": ""}
	for in, want := range tests {
		assert.Equal(t, want, safeCell(in), in)
	}
}
