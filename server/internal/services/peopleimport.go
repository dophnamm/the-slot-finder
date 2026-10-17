package services

import (
	"bytes"
	"encoding/csv"
	stderrors "errors"
	"io"
	"math"
	"net/mail"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"the-slot-finder/server/internal/constants"
)

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
	if role := cell(constants.PeopleImportColumnRole); role != "" {
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
