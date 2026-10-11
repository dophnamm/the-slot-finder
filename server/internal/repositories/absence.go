package repositories

import (
	"context"
	"time"

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
