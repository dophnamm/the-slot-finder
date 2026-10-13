package repositories

import (
	"context"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/models"
)

// AbsenceRepository stores people's absences and company-wide holidays.
type AbsenceRepository interface {
	AbstractRepository[models.Absence]
	// ListForPerson returns a person's own absences, oldest first.
	ListForPerson(ctx context.Context, personID string) ([]models.Absence, error)
	// ListHolidays returns the company-wide holidays (no person), oldest first.
	ListHolidays(ctx context.Context) ([]models.Absence, error)
	// ListOverlapping returns every absence, personal or holiday, that touches the dates
	// from through to.
	ListOverlapping(ctx context.Context, from, to time.Time) ([]models.Absence, error)
	// ListCalendarLeave returns a person's absences read from their calendar, oldest first.
	ListCalendarLeave(ctx context.Context, personID string) ([]models.Absence, error)
	// DetachCalendarLeave turns a person's calendar leave into ordinary absences: source manual,
	// no event UID.
	DetachCalendarLeave(ctx context.Context, personID string) error
	// ApprovePlanned approves a person's absence only while it is still planned, in one
	// statement, so two approvals at the same time cannot both succeed. It reports whether the
	// absence changed.
	ApprovePlanned(ctx context.Context, id string) (bool, error)
}

type absenceRepository struct {
	abstractRepository[models.Absence]
}

// ProvideAbsenceRepository creates the absence repository.
func ProvideAbsenceRepository(db *db.PostgresDB) AbsenceRepository {
	return &absenceRepository{abstractRepository: abstractRepository[models.Absence]{db: db}}
}

func (r *absenceRepository) ListForPerson(ctx context.Context, personID string) ([]models.Absence, error) {
	var absences []models.Absence
	err := r.db.From(ctx).
		Where("person_id = ?", personID).
		Order("from_date ASC").Order("id ASC").
		Find(&absences).Error
	return absences, err
}

func (r *absenceRepository) ListHolidays(ctx context.Context) ([]models.Absence, error) {
	var holidays []models.Absence
	err := r.db.From(ctx).
		Where("person_id IS NULL").
		Order("from_date ASC").Order("id ASC").
		Find(&holidays).Error
	return holidays, err
}

func (r *absenceRepository) ListOverlapping(ctx context.Context, from, to time.Time) ([]models.Absence, error) {
	var absences []models.Absence
	err := r.db.From(ctx).
		Where("from_date <= ? AND to_date >= ?", to, from).
		Order("from_date ASC").Order("id ASC").
		Find(&absences).Error
	return absences, err
}

func (r *absenceRepository) ApprovePlanned(ctx context.Context, id string) (bool, error) {
	result := r.db.From(ctx).Model(&models.Absence{}).
		Where("id = ? AND status = ? AND person_id IS NOT NULL", id, constants.AbsenceStatusPlanned).
		Update("status", constants.AbsenceStatusApproved)
	return result.RowsAffected > 0, result.Error
}

func (r *absenceRepository) DetachCalendarLeave(ctx context.Context, personID string) error {
	return r.db.From(ctx).Model(&models.Absence{}).
		Where("person_id = ? AND source = ?", personID, constants.AbsenceSourceCalendar).
		Updates(map[string]any{"source": constants.AbsenceSourceManual, "external_uid": nil}).Error
}

func (r *absenceRepository) ListCalendarLeave(ctx context.Context, personID string) ([]models.Absence, error) {
	var absences []models.Absence
	err := r.db.From(ctx).
		Where("person_id = ? AND source = ?", personID, constants.AbsenceSourceCalendar).
		Order("from_date ASC").Order("id ASC").
		Find(&absences).Error
	return absences, err
}
