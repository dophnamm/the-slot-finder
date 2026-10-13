package services

import (
	"context"
	"strings"
	"testing"

	"the-slot-finder/server/internal/constants"
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
