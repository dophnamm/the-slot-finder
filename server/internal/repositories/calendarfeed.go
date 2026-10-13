package repositories

import (
	"context"
	"time"

	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/models"

	"gorm.io/gorm/clause"
)

// CalendarFeedRepository stores people's calendar links.
type CalendarFeedRepository interface {
	AbstractRepository[models.CalendarFeed]
	// FindByPerson returns the person's calendar, or gorm.ErrRecordNotFound.
	FindByPerson(ctx context.Context, personID string) (*models.CalendarFeed, error)
	// ClaimDue locks up to limit calendars due at now (FOR UPDATE SKIP LOCKED) and moves their
	// next sync to next, so another server skips them. Call it inside a transaction.
	ClaimDue(ctx context.Context, now, next time.Time, limit int) ([]models.CalendarFeed, error)
	// RecordSync stores how a sync went: on success the time and the events found, on failure
	// only the error code, so the last good sync stays visible.
	RecordSync(ctx context.Context, id string, at time.Time, eventsFound int, errorCode *string) error
	// DeleteByPerson removes the person's calendar for good: its link is a secret.
	DeleteByPerson(ctx context.Context, personID string) error
}

type calendarFeedRepository struct {
	abstractRepository[models.CalendarFeed]
}

// ProvideCalendarFeedRepository creates the calendar feed repository.
func ProvideCalendarFeedRepository(db *db.PostgresDB) CalendarFeedRepository {
	return &calendarFeedRepository{abstractRepository: abstractRepository[models.CalendarFeed]{db: db}}
}

func (r *calendarFeedRepository) FindByPerson(ctx context.Context, personID string) (*models.CalendarFeed, error) {
	var feed models.CalendarFeed
	if err := r.db.From(ctx).Where("person_id = ?", personID).First(&feed).Error; err != nil {
		return nil, err
	}
	return &feed, nil
}

func (r *calendarFeedRepository) ClaimDue(ctx context.Context, now, next time.Time, limit int) ([]models.CalendarFeed, error) {
	var feeds []models.CalendarFeed
	err := r.db.From(ctx).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate, Options: clause.LockingOptionsSkipLocked}).
		Where("next_sync_at <= ?", now).
		Order("next_sync_at ASC").Order("id ASC").
		Limit(limit).
		Find(&feeds).Error
	if err != nil || len(feeds) == 0 {
		return feeds, err
	}
	ids := make([]string, len(feeds))
	for i, f := range feeds {
		ids[i] = f.ID
	}
	return feeds, r.db.From(ctx).Model(&models.CalendarFeed{}).Where("id IN ?", ids).Update("next_sync_at", next).Error
}

func (r *calendarFeedRepository) RecordSync(ctx context.Context, id string, at time.Time, eventsFound int, errorCode *string) error {
	updates := map[string]any{"last_error": errorCode}
	if errorCode == nil {
		updates["last_synced_at"] = at
		updates["events_found"] = eventsFound
	}
	return r.db.From(ctx).Model(&models.CalendarFeed{}).Where("id = ?", id).Updates(updates).Error
}

func (r *calendarFeedRepository) DeleteByPerson(ctx context.Context, personID string) error {
	return r.db.From(ctx).Unscoped().Where("person_id = ?", personID).Delete(&models.CalendarFeed{}).Error
}
