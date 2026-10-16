package constants

// Absence types. Holidays are created through /holidays and have no person.
const (
	AbsenceTypeLeave    = "leave"
	AbsenceTypeTraining = "training"
	AbsenceTypeHoliday  = "holiday"
)

// Absence statuses: approved always lowers capacity, planned only with COUNT_PLANNED_ABSENCE.
const (
	AbsenceStatusPlanned  = "planned"
	AbsenceStatusApproved = "approved"
)

// Absence sources: entered by hand, or read from the person's calendar feed.
const (
	AbsenceSourceManual   = "manual"
	AbsenceSourceCalendar = "calendar"
	// AbsenceSourceDismissed marks calendar leave a person deleted by hand. The row stays soft-deleted
	// with its event UID, so the next sync does not bring the event back.
	AbsenceSourceDismissed = "calendar_dismissed"
)

// PersonalAbsenceTypes are the types a person's own absence may have.
var PersonalAbsenceTypes = []string{AbsenceTypeLeave, AbsenceTypeTraining}

// AbsenceStatuses lists every absence status.
var AbsenceStatuses = []string{AbsenceStatusPlanned, AbsenceStatusApproved}
