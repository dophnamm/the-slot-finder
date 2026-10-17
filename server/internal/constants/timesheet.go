package constants

// Columns of the timesheet CSV (docs/features/timesheet-import.md). The template writes them in
// this order; an import reads them by name, in any order, and ignores the others.
const (
	TimesheetColumnPersonEmail  = "person_email"
	TimesheetColumnPersonName   = "person_name"
	TimesheetColumnProgram      = "program"
	TimesheetColumnProject      = "project"
	TimesheetColumnProjectID    = "project_id"
	TimesheetColumnWeek         = "week"
	TimesheetColumnPlannedHours = "planned_hours"
	TimesheetColumnActualHours  = "actual_hours"
)

// TimesheetColumns is the template's header, in order.
var TimesheetColumns = []string{TimesheetColumnPersonEmail, TimesheetColumnPersonName, TimesheetColumnProgram,
	TimesheetColumnProject, TimesheetColumnProjectID, TimesheetColumnWeek, TimesheetColumnPlannedHours,
	TimesheetColumnActualHours}

// TimesheetRequiredColumns must be in the header of an imported file.
var TimesheetRequiredColumns = []string{TimesheetColumnPersonEmail, TimesheetColumnProjectID, TimesheetColumnWeek,
	TimesheetColumnActualHours}

// Problem codes of a timesheet import, each reported as { line, column, code }.
const (
	TimesheetColumnMissing  = "COLUMN_MISSING"
	TimesheetCSVMalformed   = "CSV_MALFORMED"
	TimesheetPersonUnknown  = "PERSON_UNKNOWN"
	TimesheetProjectUnknown = "PROJECT_UNKNOWN"
	TimesheetWeekInvalid    = "WEEK_INVALID"
	TimesheetWeekFuture     = "WEEK_IN_FUTURE"
	TimesheetHoursInvalid   = "HOURS_INVALID"
	TimesheetRowDuplicate   = "ROW_DUPLICATE"
)

// TimesheetProblemsKey holds the problems in the data of a TIMESHEET_INVALID response.
const TimesheetProblemsKey = "problems"

// The timesheet file: the multipart field of an upload, and the download's type and name.
const (
	TimesheetFormField                 = "file"
	TimesheetContentType               = CSVContentType
	TimesheetFileName                  = "timesheet.csv"
	ContentDispositionAttachmentFormat = "attachment; filename=%q"
)

// A template cell that starts with one of TimesheetFormulaPrefixes would run as a formula in a
// spreadsheet; it is written with TimesheetTextPrefix in front so it shows as text.
const (
	TimesheetFormulaPrefixes = "=+-@\t\r"
	TimesheetTextPrefix      = "'"
)
