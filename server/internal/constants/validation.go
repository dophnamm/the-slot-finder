package constants

// Custom validation tags, registered in utils.RegisterValidations.
const (
	// ValidationTagRole accepts only the values in Roles.
	ValidationTagRole = "role"
	// ValidationTagDate accepts a calendar date in DateLayout, e.g. 2026-10-12.
	ValidationTagDate = "isodate"
	// ValidationTagAbsenceType accepts only the values in PersonalAbsenceTypes.
	ValidationTagAbsenceType = "absence_type"
	// ValidationTagAbsenceStatus accepts only the values in AbsenceStatuses.
	ValidationTagAbsenceStatus = "absence_status"
	// ValidationTagISOWeek accepts an existing ISO week such as 2026-W42.
	ValidationTagISOWeek = "isoweek"
	// ValidationTagDemandProfile accepts only the values in SupportedDemandProfiles.
	ValidationTagDemandProfile = "demand_profile"
	// ValidationTagProjectStatus accepts only the values in ProjectStatuses.
	ValidationTagProjectStatus = "project_status"
	// ValidationTagAssumptionType accepts only the values in AssumptionTypes.
	ValidationTagAssumptionType = "assumption_type"
	// ValidationTagMustBeEmpty is the validator's built-in isdefault: the field must not be sent.
	ValidationTagMustBeEmpty = "isdefault"
)
