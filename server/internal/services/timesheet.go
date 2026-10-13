package services

import (
	"bytes"
	"encoding/csv"
	stderrors "errors"
	"io"
	"math"
	"strconv"
	"strings"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/isoweek"

	"github.com/google/uuid"
)

// utf8BOM starts a CSV file that Excel saves as UTF-8; it is dropped before reading.
const utf8BOM = "\ufeff"

// headerLine is the file line of the CSV header.
const headerLine = 1

// errTooManyRows stops reading a timesheet with more than TIMESHEET_MAX_ROWS rows.
var errTooManyRows = stderrors.New("timesheet has too many rows")

// RowProblem is one problem in an imported timesheet: the file line, the column (empty for
// the whole row) and a code such as PERSON_UNKNOWN. TIMESHEET_INVALID lists them in its data.
type RowProblem struct {
	Line   int    `json:"line"`
	Column string `json:"column"`
	Code   string `json:"code"`
}

// timesheetRow is one line of an imported timesheet: the email in lower case, the project id
// as a canonical UUID, and Hours nil when actual_hours is empty.
type timesheetRow struct {
	Line      int
	Email     string
	ProjectID string
	Week      isoweek.Week
	Hours     *float64
}

// timesheetKey is the person, project and week a row is about; a file may name each once.
type timesheetKey struct {
	email, projectID string
	week             isoweek.Week
}

// timesheetLimits are the checks besides the format: no week after Current, at most
// MaxHours in one row, and at most MaxRows rows.
type timesheetLimits struct {
	Current  isoweek.Week
	MaxHours float64
	MaxRows  int
}

// parseTimesheet reads a timesheet CSV and checks its header and the format and limits of
// each row. It returns every row it read and every problem it found; the caller saves nothing
// when there is any. Excel's byte order mark is dropped, header names match in any order and
// case, and the cells a short row lacks count as empty. errTooManyRows means the file holds
// more than limits.MaxRows rows.
func parseTimesheet(data []byte, limits timesheetLimits) ([]timesheetRow, []RowProblem, error) {
	reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(data, []byte(utf8BOM))))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if stderrors.Is(err, io.EOF) {
		return nil, missingColumns(map[string]int{}), nil
	}
	if err != nil {
		return nil, []RowProblem{malformedProblem(err)}, nil
	}
	columns := headerIndex(header)
	if problems := missingColumns(columns); len(problems) > 0 {
		return nil, problems, nil
	}
	var rows []timesheetRow
	var problems []RowProblem
	seen := map[timesheetKey]bool{}
	for {
		record, err := reader.Read()
		if stderrors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			// The reader cannot tell where a broken record ends, so the rest is not read.
			problems = append(problems, malformedProblem(err))
			break
		}
		if len(rows) == limits.MaxRows {
			return nil, nil, errTooManyRows
		}
		line, _ := reader.FieldPos(0)
		row, rowProblems := parseRow(record, columns, line, limits)
		if len(rowProblems) == 0 {
			key := timesheetKey{email: row.Email, projectID: row.ProjectID, week: row.Week}
			if seen[key] {
				rowProblems = append(rowProblems, RowProblem{Line: line, Code: constants.TimesheetRowDuplicate})
			}
			seen[key] = true
		}
		rows = append(rows, row)
		problems = append(problems, rowProblems...)
	}
	return rows, problems, nil
}

// parseRow reads one record by column name and checks its format and limits.
func parseRow(record []string, columns map[string]int, line int, limits timesheetLimits) (timesheetRow, []RowProblem) {
	cell := func(name string) string {
		if i := columns[name]; i < len(record) {
			return strings.TrimSpace(record[i])
		}
		return ""
	}
	var problems []RowProblem
	problem := func(column, code string) {
		problems = append(problems, RowProblem{Line: line, Column: column, Code: code})
	}
	row := timesheetRow{Line: line, Email: strings.ToLower(cell(constants.TimesheetColumnPersonEmail))}
	if row.Email == "" {
		problem(constants.TimesheetColumnPersonEmail, constants.TimesheetPersonUnknown)
	}
	if projectID, err := uuid.Parse(cell(constants.TimesheetColumnProjectID)); err != nil {
		problem(constants.TimesheetColumnProjectID, constants.TimesheetProjectUnknown)
	} else {
		row.ProjectID = projectID.String()
	}
	week, err := isoweek.Parse(cell(constants.TimesheetColumnWeek))
	switch {
	case err != nil:
		problem(constants.TimesheetColumnWeek, constants.TimesheetWeekInvalid)
	case week.After(limits.Current):
		problem(constants.TimesheetColumnWeek, constants.TimesheetWeekFuture)
	default:
		row.Week = week
	}
	if text := cell(constants.TimesheetColumnActualHours); text != "" {
		hours, err := strconv.ParseFloat(text, 64)
		if err != nil || math.IsNaN(hours) || hours < 0 || hours > limits.MaxHours {
			problem(constants.TimesheetColumnActualHours, constants.TimesheetHoursInvalid)
		} else {
			row.Hours = &hours
		}
	}
	return row, problems
}

// headerIndex maps each column name, trimmed and in lower case, to its position; the first of
// two equal names wins.
func headerIndex(header []string) map[string]int {
	out := make(map[string]int, len(header))
	for i, name := range header {
		key := strings.ToLower(strings.TrimSpace(name))
		if _, ok := out[key]; !ok {
			out[key] = i
		}
	}
	return out
}

// missingColumns reports each required column the header lacks.
func missingColumns(columns map[string]int) []RowProblem {
	var problems []RowProblem
	for _, name := range constants.TimesheetRequiredColumns {
		if _, ok := columns[name]; !ok {
			problems = append(problems, RowProblem{Line: headerLine, Column: name, Code: constants.TimesheetColumnMissing})
		}
	}
	return problems
}

// malformedProblem is a CSV syntax error at the line its record starts on.
func malformedProblem(err error) RowProblem {
	problem := RowProblem{Code: constants.TimesheetCSVMalformed}
	var parseErr *csv.ParseError
	if stderrors.As(err, &parseErr) {
		problem.Line = parseErr.StartLine
	}
	return problem
}
