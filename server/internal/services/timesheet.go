package services

import (
	"bytes"
	"cmp"
	"context"
	"encoding/csv"
	stderrors "errors"
	"fmt"
	"io"
	"maps"
	"math"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"github.com/google/uuid"
)

const (
	opImportTimesheet = "import_timesheet"
	resourceTimesheet = "timesheet"
)

// ImportInput is an uploaded timesheet: its content and the size the client declared.
type ImportInput struct {
	File io.Reader
	Size int64
}

// TimesheetService builds the timesheet template and imports actual hours.
type TimesheetService interface {
	// Import checks every row of a filled-in template and saves the actual hours in one
	// transaction. Any problem saves nothing and returns TIMESHEET_INVALID with the problems;
	// a file over TIMESHEET_MAX_BYTES or TIMESHEET_MAX_ROWS returns TIMESHEET_TOO_LARGE.
	Import(ctx context.Context, in ImportInput) (*models.ImportResult, error)
}

type timesheetService struct {
	actuals  repositories.ActualHoursRepository
	people   repositories.PersonRepository
	projects repositories.ProjectRepository
	authz    AuthzService
	tx       db.TxManager
	clock    clock.Clock
	cfg      *config.Config
}

// ProvideTimesheetService creates the timesheet service.
func ProvideTimesheetService(actuals repositories.ActualHoursRepository, people repositories.PersonRepository,
	projects repositories.ProjectRepository, authz AuthzService, tx db.TxManager, clk clock.Clock, cfg *config.Config,
) TimesheetService {
	return &timesheetService{actuals: actuals, people: people, projects: projects, authz: authz, tx: tx, clock: clk, cfg: cfg}
}

func (s *timesheetService) Import(ctx context.Context, in ImportInput) (*models.ImportResult, error) {
	principal, err := s.authz.Require(ctx, CanImportTimesheets)
	if err != nil {
		return nil, err
	}
	data, err := s.readFile(in)
	if err != nil {
		return nil, err
	}
	rows, problems, err := parseTimesheet(data, timesheetLimits{Current: s.currentWeek(),
		MaxHours: s.cfg.TimesheetMaxWeekHours, MaxRows: s.cfg.TimesheetMaxRows})
	if err != nil {
		return nil, s.tooLarge(opImportTimesheet)
	}
	personIDs, projects, err := s.lookups(ctx, rows)
	if err != nil {
		return nil, err
	}
	problems = append(problems, unknownReferences(rows, personIDs, projects)...)
	if len(problems) > 0 {
		slices.SortStableFunc(problems, func(a, b RowProblem) int { return cmp.Compare(a.Line, b.Line) })
		return nil, errors.NewAppError(constants.TimesheetInvalid, constants.MsgTimesheetInvalid, errors.ErrorTypeValidation,
			http.StatusBadRequest).WithContext(constants.TimesheetProblemsKey, problems).
			WithOperation(opImportTimesheet).WithResource(resourceTimesheet)
	}
	saved := make([]models.ActualHours, 0, len(rows))
	for _, row := range rows {
		if row.Hours != nil {
			saved = append(saved, models.ActualHours{PersonID: personIDs[row.Email], ProjectID: row.ProjectID, Week: row.Week,
				Hours: *row.Hours, UpdatedBy: principal.Subject})
		}
	}
	result := &models.ImportResult{Skipped: len(rows) - len(saved)}
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		created, updated, err := s.actuals.Upsert(ctx, saved)
		result.Created, result.Updated = created, updated
		return err
	})
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgTimesheetSaveFailed, err).WithOperation(opImportTimesheet).WithResource(resourceTimesheet)
	}
	return result, nil
}

// readFile reads the upload, refusing more than TIMESHEET_MAX_BYTES whatever size the client
// declared.
func (s *timesheetService) readFile(in ImportInput) ([]byte, error) {
	if in.Size > s.cfg.TimesheetMaxBytes {
		return nil, s.tooLarge(opImportTimesheet)
	}
	data, err := io.ReadAll(io.LimitReader(in.File, s.cfg.TimesheetMaxBytes+1))
	if err != nil {
		return nil, errors.ValidationError(constants.MsgTimesheetUnreadable, err).WithOperation(opImportTimesheet).WithResource(resourceTimesheet)
	}
	if int64(len(data)) > s.cfg.TimesheetMaxBytes {
		return nil, s.tooLarge(opImportTimesheet)
	}
	return data, nil
}

// tooLarge is the 413 for a file over TIMESHEET_MAX_BYTES or TIMESHEET_MAX_ROWS.
func (s *timesheetService) tooLarge(op string) error {
	return errors.NewAppError(constants.TimesheetTooLarge,
		fmt.Sprintf(constants.MsgTimesheetTooLargeFormat, s.cfg.TimesheetMaxBytes, s.cfg.TimesheetMaxRows),
		errors.ErrorTypeValidation, http.StatusRequestEntityTooLarge).WithOperation(op).WithResource(resourceTimesheet)
}

// currentWeek is the organization's current ISO week.
func (s *timesheetService) currentWeek() isoweek.Week {
	return isoweek.Of(s.clock.Now().In(s.clock.Location()))
}

// lookups loads the people the rows name by email and the projects they name by id. It returns
// person ids by lower-case email, and the ids of the projects that exist.
func (s *timesheetService) lookups(ctx context.Context, rows []timesheetRow) (map[string]string, map[string]bool, error) {
	emails, projectIDs := map[string]bool{}, map[string]bool{}
	for _, row := range rows {
		if row.Email != "" {
			emails[row.Email] = true
		}
		if row.ProjectID != "" {
			projectIDs[row.ProjectID] = true
		}
	}
	people, err := s.people.ListByEmails(ctx, slices.Sorted(maps.Keys(emails)))
	if err != nil {
		return nil, nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opImportTimesheet).WithResource(resourceTimesheet)
	}
	projects, err := s.projects.ListByIDs(ctx, slices.Sorted(maps.Keys(projectIDs)))
	if err != nil {
		return nil, nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opImportTimesheet).WithResource(resourceTimesheet)
	}
	personIDs := make(map[string]string, len(people))
	for _, p := range people {
		personIDs[strings.ToLower(p.Email)] = p.ID
	}
	known := make(map[string]bool, len(projects))
	for _, p := range projects {
		known[p.ID] = true
	}
	return personIDs, known, nil
}

// unknownReferences reports rows whose email or project id matches nothing. Rows whose cell
// was already empty or malformed are reported by parseTimesheet instead.
func unknownReferences(rows []timesheetRow, personIDs map[string]string, projects map[string]bool) []RowProblem {
	var problems []RowProblem
	for _, row := range rows {
		if _, ok := personIDs[row.Email]; row.Email != "" && !ok {
			problems = append(problems, RowProblem{Line: row.Line, Column: constants.TimesheetColumnPersonEmail,
				Code: constants.TimesheetPersonUnknown})
		}
		if row.ProjectID != "" && !projects[row.ProjectID] {
			problems = append(problems, RowProblem{Line: row.Line, Column: constants.TimesheetColumnProjectID,
				Code: constants.TimesheetProjectUnknown})
		}
	}
	return problems
}

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
