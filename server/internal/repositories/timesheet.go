package repositories

import (
	"context"

	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"

	"gorm.io/gorm/clause"
)

// actualHoursBatchSize keeps one upsert statement far below Postgres' 65,535 parameters
// (9 columns per row).
const actualHoursBatchSize = 1000

// Relations of an actual hours row: the person, and the project with its program.
const (
	actualHoursPersonRel      = "Person"
	actualHoursProjectRel     = "Project"
	actualHoursProgramPreload = "Project.Program"
)

// ActualHoursRepository stores the hours people really worked, per person, project and ISO
// week. It belongs to the timesheets module.
type ActualHoursRepository interface {
	// Upsert saves each row, replacing the hours of a person, project and week that is
	// already stored. It returns how many rows were new and how many replaced a value. The
	// rows must not repeat a person, project and week.
	Upsert(ctx context.Context, rows []models.ActualHours) (created, updated int, err error)
	// SumByProject returns the actual hours of each project that has any; a project without
	// rows is missing from the map.
	SumByProject(ctx context.Context, projectIDs []string) (map[string]float64, error)
	// ByPersonWeeks returns one person's actual hours per week over all projects, from
	// through to; a week without rows is missing from the map.
	ByPersonWeeks(ctx context.Context, personID string, from, to isoweek.Week) (map[isoweek.Week]float64, error)
	// ListInRange returns every row from through to, with the person and the project and its
	// program.
	ListInRange(ctx context.Context, from, to isoweek.Week) ([]models.ActualHours, error)
}

type actualHoursRepository struct {
	db *db.PostgresDB
}

// ProvideActualHoursRepository creates the actual hours repository.
func ProvideActualHoursRepository(db *db.PostgresDB) ActualHoursRepository {
	return &actualHoursRepository{db: db}
}

func (r *actualHoursRepository) Upsert(ctx context.Context, rows []models.ActualHours) (int, int, error) {
	if len(rows) == 0 {
		return 0, 0, nil
	}
	keys := make([][]any, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, []any{row.PersonID, row.ProjectID, row.Week})
	}
	var existing int64
	if err := r.db.From(ctx).Model(&models.ActualHours{}).
		Where("(person_id, project_id, week_start) IN ?", keys).Count(&existing).Error; err != nil {
		return 0, 0, err
	}
	err := r.db.From(ctx).Omit(actualHoursPersonRel, actualHoursProjectRel).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "person_id"}, {Name: "project_id"}, {Name: "week_start"}},
		DoUpdates: clause.AssignmentColumns([]string{"hours", "updated_by", "updated_at"}),
	}).CreateInBatches(&rows, actualHoursBatchSize).Error
	if err != nil {
		return 0, 0, err
	}
	return len(rows) - int(existing), int(existing), nil
}

// projectHours is one row of SumByProject.
type projectHours struct {
	ProjectID string
	Hours     float64
}

func (r *actualHoursRepository) SumByProject(ctx context.Context, projectIDs []string) (map[string]float64, error) {
	out := make(map[string]float64, len(projectIDs))
	if len(projectIDs) == 0 {
		return out, nil
	}
	var sums []projectHours
	err := r.db.From(ctx).Model(&models.ActualHours{}).Select("project_id, SUM(hours) AS hours").
		Where("project_id IN ?", projectIDs).Group("project_id").Scan(&sums).Error
	if err != nil {
		return nil, err
	}
	for _, s := range sums {
		out[s.ProjectID] = s.Hours
	}
	return out, nil
}

// weekHours is one row of ByPersonWeeks.
type weekHours struct {
	Week  isoweek.Week `gorm:"column:week_start"`
	Hours float64
}

func (r *actualHoursRepository) ByPersonWeeks(ctx context.Context, personID string, from, to isoweek.Week) (map[isoweek.Week]float64, error) {
	var sums []weekHours
	err := r.db.From(ctx).Model(&models.ActualHours{}).Select("week_start, SUM(hours) AS hours").
		Where("person_id = ? AND week_start BETWEEN ? AND ?", personID, from, to).Group("week_start").Scan(&sums).Error
	if err != nil {
		return nil, err
	}
	out := make(map[isoweek.Week]float64, len(sums))
	for _, s := range sums {
		out[s.Week] = s.Hours
	}
	return out, nil
}

func (r *actualHoursRepository) ListInRange(ctx context.Context, from, to isoweek.Week) ([]models.ActualHours, error) {
	var rows []models.ActualHours
	err := r.db.From(ctx).Preload(actualHoursPersonRel).Preload(actualHoursProgramPreload).
		Where("week_start BETWEEN ? AND ?", from, to).
		Order("week_start ASC").Order("person_id ASC").Order("project_id ASC").Find(&rows).Error
	return rows, err
}
