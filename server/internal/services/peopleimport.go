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
	"net/mail"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
)

const (
	opImportPeople       = "import_people"
	opPeopleImportTpl    = "people_import_template"
	resourcePeopleImport = "people_import"
)

// PeopleImportService imports people from a CSV and builds the CSV template.
type PeopleImportService interface {
	// Import checks every row of a people file and saves it in one transaction: a new email
	// creates a person, a known one is updated. Any problem saves nothing and returns
	// PEOPLE_IMPORT_INVALID with the problems; a file over PEOPLE_IMPORT_MAX_BYTES or
	// PEOPLE_IMPORT_MAX_ROWS returns PEOPLE_IMPORT_TOO_LARGE.
	Import(ctx context.Context, in ImportInput) (*models.PeopleImportResult, error)
	// Template is a CSV with the header and one example row.
	Template(ctx context.Context) ([]byte, error)
}

type peopleImportService struct {
	people PeopleService
	repo   repositories.PersonRepository
	authz  AuthzService
	tx     db.TxManager
	cfg    *config.Config
}

// ProvidePeopleImportService creates the people import service.
func ProvidePeopleImportService(people PeopleService, repo repositories.PersonRepository, authz AuthzService,
	tx db.TxManager, cfg *config.Config,
) PeopleImportService {
	return &peopleImportService{people: people, repo: repo, authz: authz, tx: tx, cfg: cfg}
}

func (s *peopleImportService) Template(ctx context.Context) ([]byte, error) {
	if _, err := s.authz.Require(ctx, CanManagePeople); err != nil {
		return nil, err
	}
	expertises, err := s.people.ListExpertises(ctx)
	if err != nil {
		return nil, err
	}
	primary := ""
	if len(expertises) > 0 {
		primary = expertises[0].Name
	}
	example := map[string]string{
		constants.PeopleImportColumnName:             constants.PeopleImportExampleName,
		constants.PeopleImportColumnEmail:            constants.PeopleImportExampleEmail,
		constants.PeopleImportColumnRole:             constants.RoleSpecialist,
		constants.PeopleImportColumnWeeklyHours:      constants.PeopleImportExampleWeeklyHours,
		constants.PeopleImportColumnNonProjectHours:  constants.PeopleImportExampleNonProjectHours,
		constants.PeopleImportColumnPrimaryExpertise: primary,
		constants.PeopleImportColumnPrimaryLevel:     constants.PeopleImportExamplePrimaryLevel,
	}
	record := make([]string, len(constants.PeopleImportColumns))
	for i, column := range constants.PeopleImportColumns {
		record[i] = example[column]
	}
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	_ = w.WriteAll([][]string{constants.PeopleImportColumns, record})
	if err := w.Error(); err != nil {
		return nil, errors.InternalError(constants.MsgPeopleImportTemplateFailed, err).
			WithOperation(opPeopleImportTpl).WithResource(resourcePeopleImport)
	}
	return buf.Bytes(), nil
}

func (s *peopleImportService) Import(ctx context.Context, in ImportInput) (*models.PeopleImportResult, error) {
	if _, err := s.authz.Require(ctx, CanManagePeople); err != nil {
		return nil, err
	}
	data, err := s.readFile(in)
	if err != nil {
		return nil, err
	}
	expertises, err := s.people.ListExpertises(ctx)
	if err != nil {
		return nil, err
	}
	byName := make(map[string]string, len(expertises))
	for _, e := range expertises {
		byName[strings.ToLower(e.Name)] = e.ID
	}
	rows, problems, err := parsePeopleCSV(data, peopleImportLimits{MaxRows: s.cfg.PeopleImportMaxRows}, byName)
	if err != nil {
		return nil, s.tooLarge()
	}
	stored, err := s.stored(ctx, rows)
	if err != nil {
		return nil, err
	}
	// A row with a cell problem is not checked further: its merged values mean nothing.
	bad := map[int]bool{}
	for _, p := range problems {
		bad[p.Line] = true
	}
	for _, row := range rows {
		if !bad[row.Line] {
			problems = append(problems, s.checkRow(row, stored[row.Email])...)
		}
	}
	if len(problems) > 0 {
		slices.SortStableFunc(problems, func(a, b RowProblem) int { return cmp.Compare(a.Line, b.Line) })
		return nil, errors.NewAppError(constants.PeopleImportInvalid, constants.MsgPeopleImportInvalid, errors.ErrorTypeValidation,
			http.StatusBadRequest).WithContext(constants.PeopleImportProblemsKey, problems).
			WithOperation(opImportPeople).WithResource(resourcePeopleImport)
	}
	result := &models.PeopleImportResult{}
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		for _, row := range rows {
			if err := s.save(ctx, row, stored[row.Email], result); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		var appErr *errors.AppError
		if stderrors.As(err, &appErr) {
			return nil, err
		}
		return nil, errors.DatabaseError(constants.MsgPeopleImportSaveFailed, err).
			WithOperation(opImportPeople).WithResource(resourcePeopleImport)
	}
	return result, nil
}

// stored loads the people the rows name by email, with their skills, keyed by lower-case email.
func (s *peopleImportService) stored(ctx context.Context, rows []peopleRow) (map[string]*models.Person, error) {
	emails := map[string]bool{}
	for _, row := range rows {
		if row.Email != "" {
			emails[row.Email] = true
		}
	}
	found, err := s.repo.ListByEmails(ctx, slices.Sorted(maps.Keys(emails)))
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opImportPeople).WithResource(resourcePeopleImport)
	}
	out := make(map[string]*models.Person, len(found))
	for i, p := range found {
		out[strings.ToLower(p.Email)] = &found[i]
	}
	return out, nil
}

// checkRow applies the rules of Create and Update to the values the save would store, so a
// failing rule is a row problem and not a save error.
func (s *peopleImportService) checkRow(row peopleRow, person *models.Person) []RowProblem {
	start, end, nonProject := row.StartDate, row.EndDate, s.cfg.DefaultNonProjectHours
	if person != nil {
		nonProject = person.NonProjectHours
		start, end = cmp.Or(start, person.StartDate), cmp.Or(end, person.EndDate)
	}
	if row.NonProjectHours != nil {
		nonProject = *row.NonProjectHours
	}
	var problems []RowProblem
	if checkDates(start, end, opImportPeople) != nil {
		problems = append(problems, RowProblem{Line: row.Line, Column: constants.PeopleImportColumnEndDate, Code: constants.PeopleImportDateInvalid})
	}
	if checkNonProjectHours(nonProject, row.WeeklyHours, opImportPeople) != nil {
		problems = append(problems, RowProblem{Line: row.Line, Column: constants.PeopleImportColumnNonProjectHours, Code: constants.PeopleImportHoursInvalid})
	}
	return problems
}

// save creates the person or applies only the cells that differ from the stored one.
func (s *peopleImportService) save(ctx context.Context, row peopleRow, person *models.Person, result *models.PeopleImportResult) error {
	if person == nil {
		_, err := s.people.Create(ctx, CreatePersonInput{Name: row.Name, Email: row.Email, Role: cmp.Or(deref(row.Role), constants.RoleSpecialist),
			WeeklyHours: row.WeeklyHours, NonProjectHours: row.NonProjectHours, StartDate: row.StartDate, EndDate: row.EndDate, Skills: row.Skills})
		result.Created++
		return err
	}
	var in UpdatePersonInput
	changed := false
	if row.Name != person.Name {
		in.Name, changed = &row.Name, true
	}
	if row.Role != nil && *row.Role != person.Role {
		in.Role, changed = row.Role, true
	}
	if row.WeeklyHours != person.WeeklyHours {
		in.WeeklyHours, changed = &row.WeeklyHours, true
	}
	if row.NonProjectHours != nil && *row.NonProjectHours != person.NonProjectHours {
		in.NonProjectHours, changed = row.NonProjectHours, true
	}
	if row.StartDate != nil && !samePersonDate(row.StartDate, person.StartDate) {
		in.StartDate, changed = DateChange{Set: true, Value: row.StartDate}, true
	}
	if row.EndDate != nil && !samePersonDate(row.EndDate, person.EndDate) {
		in.EndDate, changed = DateChange{Set: true, Value: row.EndDate}, true
	}
	if changed {
		if _, err := s.people.Update(ctx, person.ID, in); err != nil {
			return err
		}
	}
	skills := !sameSkills(person.Skills, toSkillModels(person.ID, row.Skills))
	if skills {
		if _, err := s.people.ReplaceSkills(ctx, person.ID, row.Skills); err != nil {
			return err
		}
	}
	if changed || skills {
		result.Updated++
	} else {
		result.Unchanged++
	}
	return nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// readFile reads the upload, refusing more than PEOPLE_IMPORT_MAX_BYTES whatever size the client declared.
func (s *peopleImportService) readFile(in ImportInput) ([]byte, error) {
	if in.Size > s.cfg.PeopleImportMaxBytes {
		return nil, s.tooLarge()
	}
	data, err := io.ReadAll(io.LimitReader(in.File, s.cfg.PeopleImportMaxBytes+1))
	if err != nil {
		return nil, errors.ValidationError(constants.MsgPeopleImportUnreadable, err).WithOperation(opImportPeople).WithResource(resourcePeopleImport)
	}
	if int64(len(data)) > s.cfg.PeopleImportMaxBytes {
		return nil, s.tooLarge()
	}
	return data, nil
}

// tooLarge is the 413 for a file over PEOPLE_IMPORT_MAX_BYTES or PEOPLE_IMPORT_MAX_ROWS.
func (s *peopleImportService) tooLarge() error {
	return errors.NewAppError(constants.PeopleImportTooLarge,
		fmt.Sprintf(constants.MsgPeopleImportTooLargeFormat, s.cfg.PeopleImportMaxBytes, s.cfg.PeopleImportMaxRows),
		errors.ErrorTypeValidation, http.StatusRequestEntityTooLarge).WithOperation(opImportPeople).WithResource(resourcePeopleImport)
}

// Limits of a person's cells, the same as the person DTO's validate tags.
const (
	peopleMaxNameLength  = 200
	peopleMaxEmailLength = 320
	peopleMaxHours       = 80
	peopleMinLevel       = 1
	peopleMaxLevel       = 5
)

// errTooManyPeopleRows stops reading a people file with more than PEOPLE_IMPORT_MAX_ROWS rows.
var errTooManyPeopleRows = stderrors.New("people file has too many rows")

// peopleRow is one checked CSV row. Optional cells that were empty stay nil.
type peopleRow struct {
	Line            int
	Name            string
	Email           string // lower-cased
	Role            *string
	WeeklyHours     float64
	NonProjectHours *float64
	StartDate       *time.Time
	EndDate         *time.Time
	Skills          []SkillInput // primary first
}

type peopleImportLimits struct{ MaxRows int }

// parsePeopleCSV reads the header by name (trimmed, case-insensitive; unknown columns ignored) and checks every row
// on its own. expertises maps a lower-cased expertise name to its id. It returns errTooManyPeopleRows past MaxRows.
func parsePeopleCSV(data []byte, limits peopleImportLimits, expertises map[string]string) ([]peopleRow, []RowProblem, error) {
	reader := csv.NewReader(bytes.NewReader(bytes.TrimPrefix(data, []byte(utf8BOM))))
	reader.FieldsPerRecord = -1
	reader.TrimLeadingSpace = true
	header, err := reader.Read()
	if err != nil && !stderrors.Is(err, io.EOF) {
		return nil, []RowProblem{peopleMalformed(err)}, nil
	}
	columns := headerIndex(header)
	var problems []RowProblem
	for _, name := range constants.PeopleImportRequiredColumns {
		if _, ok := columns[name]; !ok {
			problems = append(problems, RowProblem{Line: headerLine, Column: name, Code: constants.PeopleImportColumnMissing})
		}
	}
	if len(problems) > 0 {
		return nil, problems, nil
	}
	var rows []peopleRow
	seen := map[string]bool{}
	for {
		record, err := reader.Read()
		if stderrors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			// The reader cannot tell where a broken record ends, so the rest is not read.
			problems = append(problems, peopleMalformed(err))
			break
		}
		if blankRecord(record) {
			continue
		}
		if len(rows) == limits.MaxRows {
			return nil, nil, errTooManyPeopleRows
		}
		line, _ := reader.FieldPos(0)
		row, rowProblems := parsePeopleRow(record, columns, line, expertises)
		if row.Email != "" && !slices.ContainsFunc(rowProblems, func(p RowProblem) bool {
			return p.Column == constants.PeopleImportColumnEmail
		}) {
			if seen[row.Email] {
				rowProblems = append(rowProblems, RowProblem{Line: line, Column: constants.PeopleImportColumnEmail,
					Code: constants.PeopleImportRowDuplicate})
			}
			seen[row.Email] = true
		}
		rows = append(rows, row)
		problems = append(problems, rowProblems...)
	}
	return rows, problems, nil
}

// peopleMalformed is a CSV syntax error with the people code.
func peopleMalformed(err error) RowProblem {
	problem := malformedProblem(err)
	problem.Code = constants.PeopleImportCSVMalformed
	return problem
}

// parsePeopleRow reads one record by column name and checks every cell.
func parsePeopleRow(record []string, columns map[string]int, line int, expertises map[string]string) (peopleRow, []RowProblem) {
	cell := func(name string) string {
		if i, ok := columns[name]; ok && i < len(record) {
			return strings.TrimSpace(record[i])
		}
		return ""
	}
	var problems []RowProblem
	problem := func(column, code string) {
		problems = append(problems, RowProblem{Line: line, Column: column, Code: code})
	}
	row := peopleRow{Line: line, Name: cell(constants.PeopleImportColumnName),
		Email: strings.ToLower(cell(constants.PeopleImportColumnEmail))}
	if n := utf8.RuneCountInString(row.Name); n == 0 || n > peopleMaxNameLength {
		problem(constants.PeopleImportColumnName, constants.PeopleImportNameInvalid)
	}
	if addr, err := mail.ParseAddress(row.Email); err != nil || addr.Address != row.Email || len(row.Email) > peopleMaxEmailLength {
		problem(constants.PeopleImportColumnEmail, constants.PeopleImportEmailInvalid)
	}
	if role := strings.ToLower(cell(constants.PeopleImportColumnRole)); role != "" {
		if !slices.Contains(constants.Roles, role) {
			problem(constants.PeopleImportColumnRole, constants.PeopleImportRoleInvalid)
		} else {
			row.Role = &role
		}
	}
	if hours, ok := parseHours(cell(constants.PeopleImportColumnWeeklyHours)); !ok || hours == 0 {
		problem(constants.PeopleImportColumnWeeklyHours, constants.PeopleImportHoursInvalid)
	} else {
		row.WeeklyHours = hours
	}
	if text := cell(constants.PeopleImportColumnNonProjectHours); text != "" {
		if hours, ok := parseHours(text); !ok {
			problem(constants.PeopleImportColumnNonProjectHours, constants.PeopleImportHoursInvalid)
		} else {
			row.NonProjectHours = &hours
		}
	}
	row.StartDate = parseDateCell(cell(constants.PeopleImportColumnStartDate), constants.PeopleImportColumnStartDate, problem)
	row.EndDate = parseDateCell(cell(constants.PeopleImportColumnEndDate), constants.PeopleImportColumnEndDate, problem)
	if row.StartDate != nil && row.EndDate != nil && row.EndDate.Before(*row.StartDate) {
		problem(constants.PeopleImportColumnEndDate, constants.PeopleImportDateInvalid)
	}

	used := map[string]bool{}
	expertiseID, ok := expertises[strings.ToLower(cell(constants.PeopleImportColumnPrimaryExpertise))]
	if !ok {
		problem(constants.PeopleImportColumnPrimaryExpertise, constants.PeopleImportExpertiseUnknown)
	}
	level, levelOK := parseLevel(cell(constants.PeopleImportColumnPrimaryLevel))
	if !levelOK {
		problem(constants.PeopleImportColumnPrimaryLevel, constants.PeopleImportLevelInvalid)
	}
	if ok && levelOK {
		used[expertiseID] = true
		row.Skills = append(row.Skills, SkillInput{ExpertiseID: expertiseID, Level: level, Primary: true})
	}
	for _, part := range strings.Split(cell(constants.PeopleImportColumnOtherSkills), constants.PeopleImportSkillSeparator) {
		if strings.TrimSpace(part) == "" {
			continue
		}
		name, levelText, found := strings.Cut(part, constants.PeopleImportLevelSeparator)
		if !found {
			problem(constants.PeopleImportColumnOtherSkills, constants.PeopleImportSkillsInvalid)
			continue
		}
		id, known := expertises[strings.ToLower(strings.TrimSpace(name))]
		lvl, lvlOK := parseLevel(strings.TrimSpace(levelText))
		switch {
		case !known:
			problem(constants.PeopleImportColumnOtherSkills, constants.PeopleImportExpertiseUnknown)
		case !lvlOK:
			problem(constants.PeopleImportColumnOtherSkills, constants.PeopleImportLevelInvalid)
		case used[id]:
			problem(constants.PeopleImportColumnOtherSkills, constants.PeopleImportSkillDuplicate)
		default:
			used[id] = true
			row.Skills = append(row.Skills, SkillInput{ExpertiseID: id, Level: lvl})
		}
	}
	return row, problems
}

// parseHours reads a number of hours from 0 to peopleMaxHours.
func parseHours(text string) (float64, bool) {
	hours, err := strconv.ParseFloat(text, 64)
	return hours, err == nil && !math.IsNaN(hours) && hours >= 0 && hours <= peopleMaxHours
}

// parseLevel reads a skill level from peopleMinLevel to peopleMaxLevel.
func parseLevel(text string) (int, bool) {
	level, err := strconv.Atoi(text)
	return level, err == nil && level >= peopleMinLevel && level <= peopleMaxLevel
}

// parseDateCell reads an optional date cell in constants.DateLayout; a bad one is reported.
func parseDateCell(text, column string, problem func(column, code string)) *time.Time {
	if text == "" {
		return nil
	}
	date, err := time.Parse(constants.DateLayout, text)
	if err != nil {
		problem(column, constants.PeopleImportDateInvalid)
		return nil
	}
	return &date
}
