package repositories

import (
	"context"
	"encoding/json"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// eventDraftPreload loads the draft an event led to, for its status.
const eventDraftPreload = "Draft"

// ChangeEventRepository stores change events and lets the re-plan runner claim them.
type ChangeEventRepository interface {
	AbstractRepository[models.ChangeEvent]
	// List returns events newest first with their draft; with projectIDs, only events that
	// touch one of them.
	List(ctx context.Context, projectIDs []string, page *dtos.PageableRequest) (*dtos.DataResponse[models.ChangeEvent], error)
	// ClaimPending locks up to limit pending events available at now that occurred at or
	// before cutoff, oldest first. Rows another transaction holds are skipped. Run it inside
	// TxManager.Do so the locks last until the batch is done.
	ClaimPending(ctx context.Context, now, cutoff time.Time, limit int) ([]models.ChangeEvent, error)
	// Finish sets an event's status, its draft and the projects it touches.
	Finish(ctx context.Context, id, status string, draftID *string, projectIDs []string) error
	// RelinkDraft points every event of one draft at another, e.g. a draft's newer version.
	RelinkDraft(ctx context.Context, fromDraftID, toDraftID string) error
	// ListByDraft returns the events linked to a draft, oldest first.
	ListByDraft(ctx context.Context, draftID string) ([]models.ChangeEvent, error)
	// RecordFailure counts a failed attempt on pending events: it keeps the error, makes them
	// available again after backoff × attempts, and fails those that reach maxAttempts.
	RecordFailure(ctx context.Context, ids []string, message string, now time.Time, backoff time.Duration, maxAttempts int) error
	// LockReplanning takes a transaction-wide lock so only one runner, on any server, builds
	// re-plan drafts at a time; it keeps one open auto draft per baseline.
	LockReplanning(ctx context.Context) error
}

// replanLockKey names the advisory lock that serializes re-plan batches.
const replanLockKey = "slot-finder.replan"

type changeEventRepository struct {
	abstractRepository[models.ChangeEvent]
}

// ProvideChangeEventRepository creates the change event repository.
func ProvideChangeEventRepository(db *db.PostgresDB) ChangeEventRepository {
	return &changeEventRepository{abstractRepository: abstractRepository[models.ChangeEvent]{db: db}}
}

func (r *changeEventRepository) List(ctx context.Context, projectIDs []string, page *dtos.PageableRequest) (*dtos.DataResponse[models.ChangeEvent], error) {
	query := r.db.From(ctx).Preload(eventDraftPreload).Order("occurred_at DESC").Order("id DESC")
	if projectIDs != nil {
		query = query.Where("EXISTS (SELECT 1 FROM jsonb_array_elements_text(project_ids) AS touched(id) WHERE touched.id IN ?)", projectIDs)
	}
	return r.find(query, page)
}

func (r *changeEventRepository) ClaimPending(ctx context.Context, now, cutoff time.Time, limit int) ([]models.ChangeEvent, error) {
	var events []models.ChangeEvent
	err := r.db.From(ctx).
		Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate, Options: clause.LockingOptionsSkipLocked}).
		Where("status = ? AND available_at <= ? AND occurred_at <= ?", constants.EventStatusPending, now, cutoff).
		Order("occurred_at ASC").Order("id ASC").
		Limit(limit).
		Find(&events).Error
	return events, err
}

func (r *changeEventRepository) Finish(ctx context.Context, id, status string, draftID *string, projectIDs []string) error {
	return r.db.From(ctx).Model(&models.ChangeEvent{}).Where("id = ?", id).Updates(map[string]any{
		"status": status, "draft_id": draftID, "project_ids": gorm.Expr("?::jsonb", jsonArray(projectIDs)), "last_error": nil,
	}).Error
}

func (r *changeEventRepository) RelinkDraft(ctx context.Context, fromDraftID, toDraftID string) error {
	return r.db.From(ctx).Model(&models.ChangeEvent{}).Where("draft_id = ?", fromDraftID).Update("draft_id", toDraftID).Error
}

func (r *changeEventRepository) ListByDraft(ctx context.Context, draftID string) ([]models.ChangeEvent, error) {
	var events []models.ChangeEvent
	err := r.db.From(ctx).Where("draft_id = ?", draftID).Order("occurred_at ASC").Order("id ASC").Find(&events).Error
	return events, err
}

func (r *changeEventRepository) LockReplanning(ctx context.Context) error {
	return r.db.From(ctx).Exec("SELECT pg_advisory_xact_lock(hashtext(?))", replanLockKey).Error
}

func (r *changeEventRepository) RecordFailure(ctx context.Context, ids []string, message string, now time.Time, backoff time.Duration, maxAttempts int) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.From(ctx).Model(&models.ChangeEvent{}).
		Where("id IN ? AND status = ?", ids, constants.EventStatusPending).
		Updates(map[string]any{
			"attempts":     gorm.Expr("attempts + 1"),
			"last_error":   message,
			"available_at": gorm.Expr("?::timestamptz + (attempts + 1) * make_interval(secs => ?)", now, backoff.Seconds()),
			"status":       gorm.Expr("CASE WHEN attempts + 1 >= ? THEN ? ELSE status END", maxAttempts, constants.EventStatusFailed),
		}).Error
}

// jsonArray writes ids as a JSON array for a jsonb column; nil becomes [].
func jsonArray(ids []string) string {
	if len(ids) == 0 {
		return "[]"
	}
	out, err := json.Marshal(ids)
	if err != nil {
		return "[]"
	}
	return string(out)
}
