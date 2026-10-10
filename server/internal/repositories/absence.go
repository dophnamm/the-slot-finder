package repositories

import (
	"context"

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
	err := r.db.WithContext(ctx).
		Where("person_id = ?", personID).
		Order("from_date ASC").Order("id ASC").
		Find(&absences).Error
	return absences, err
}

func (r *absenceRepository) ListHolidays(ctx context.Context) ([]models.Absence, error) {
	var holidays []models.Absence
	err := r.db.WithContext(ctx).
		Where("person_id IS NULL").
		Order("from_date ASC").Order("id ASC").
		Find(&holidays).Error
	return holidays, err
}
