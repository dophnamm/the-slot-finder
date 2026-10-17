package constants

// Columns of the people CSV (docs/features/people-import.md). The template writes them in this
// order; an import reads them by name, in any order, and ignores the others.
const (
	PeopleImportColumnName             = "name"
	PeopleImportColumnEmail            = "email"
	PeopleImportColumnRole             = "role"
	PeopleImportColumnWeeklyHours      = "weekly_hours"
	PeopleImportColumnNonProjectHours  = "non_project_hours"
	PeopleImportColumnStartDate        = "start_date"
	PeopleImportColumnEndDate          = "end_date"
	PeopleImportColumnPrimaryExpertise = "primary_expertise"
	PeopleImportColumnPrimaryLevel     = "primary_level"
	PeopleImportColumnOtherSkills      = "other_skills"
)

// PeopleImportColumns is the template's column order.
var PeopleImportColumns = []string{PeopleImportColumnName, PeopleImportColumnEmail, PeopleImportColumnRole,
	PeopleImportColumnWeeklyHours, PeopleImportColumnNonProjectHours, PeopleImportColumnStartDate,
	PeopleImportColumnEndDate, PeopleImportColumnPrimaryExpertise, PeopleImportColumnPrimaryLevel,
	PeopleImportColumnOtherSkills}

// PeopleImportRequiredColumns must be in the header of an imported file.
var PeopleImportRequiredColumns = []string{PeopleImportColumnName, PeopleImportColumnEmail,
	PeopleImportColumnWeeklyHours, PeopleImportColumnPrimaryExpertise, PeopleImportColumnPrimaryLevel}

// Separators of the other_skills cell, e.g. "Engineer:2;Analyst:1".
const (
	PeopleImportSkillSeparator = ";"
	PeopleImportLevelSeparator = ":"
)

// Row problem codes of a people import, each reported as { line, column, code }.
const (
	PeopleImportColumnMissing    = "COLUMN_MISSING"
	PeopleImportCSVMalformed     = "CSV_MALFORMED"
	PeopleImportRowDuplicate     = "ROW_DUPLICATE"
	PeopleImportNameInvalid      = "NAME_INVALID"
	PeopleImportEmailInvalid     = "EMAIL_INVALID"
	PeopleImportRoleInvalid      = "ROLE_INVALID"
	PeopleImportHoursInvalid     = "HOURS_INVALID"
	PeopleImportDateInvalid      = "DATE_INVALID"
	PeopleImportExpertiseUnknown = "EXPERTISE_UNKNOWN"
	PeopleImportLevelInvalid     = "LEVEL_INVALID"
	PeopleImportSkillsInvalid    = "SKILLS_INVALID"
	PeopleImportSkillDuplicate   = "SKILL_DUPLICATE"
)

// The people file: the key of the problems in a PEOPLE_IMPORT_INVALID response, the multipart
// field of an upload, the template's name, and the type of a CSV download.
const (
	PeopleImportProblemsKey = "problems"
	PeopleImportFormField   = "file"
	PeopleImportFileName    = "people.csv"
	CSVContentType          = "text/csv; charset=utf-8"
)
