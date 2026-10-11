package models

import "time"

// ChangeEvent is a change that may break the plan (leave, a slipped phase or deadline). It is
// inserted in the same transaction as the change (an outbox) and processed by the re-plan
// runner, which claims pending rows with FOR UPDATE SKIP LOCKED.
type ChangeEvent struct {
	BaseModel
	Type       string         `gorm:"column:type;not null;index"`
	Payload    map[string]any `gorm:"column:payload;serializer:json;type:jsonb;not null;default:'{}'"`
	OccurredAt time.Time      `gorm:"column:occurred_at;type:timestamptz;not null"`
	Status     string         `gorm:"column:status;not null;index:idx_change_events_status_available,priority:1"`
	// AvailableAt is when the runner may claim the event; a failure moves it later.
	AvailableAt time.Time `gorm:"column:available_at;type:timestamptz;not null;index:idx_change_events_status_available,priority:2"`
	Attempts    int       `gorm:"column:attempts;not null;default:0"`
	LastError   *string   `gorm:"column:last_error"`
	// DraftID is the re-plan draft the event led to; it follows the draft's latest version.
	DraftID *string `gorm:"column:draft_id;type:uuid;index"`
	// ProjectIDs are the projects the change touches: known when a phase or deadline moves,
	// found by the impact assessment for absence.
	ProjectIDs []string `gorm:"column:project_ids;serializer:json;type:jsonb;not null;default:'[]'"`

	Draft *Plan `gorm:"foreignKey:DraftID;constraint:OnDelete:SET NULL"`
}
