package models

import "time"

// CalendarFeed is a person's secret calendar link (ICS), stored encrypted, with how its last
// sync went. Leave read from it is stored as absences with source calendar.
type CalendarFeed struct {
	BaseModel
	PersonID      string     `gorm:"column:person_id;type:uuid;not null;uniqueIndex"`
	URLCiphertext []byte     `gorm:"column:url_ciphertext;type:bytea;not null"`
	Host          string     `gorm:"column:host;not null"`
	LastSyncedAt  *time.Time `gorm:"column:last_synced_at;type:timestamptz"`
	LastError     *string    `gorm:"column:last_error"`
	EventsFound   int        `gorm:"column:events_found;not null;default:0"`
	// NextSyncAt is when the runner syncs it next; claiming a feed moves it one interval ahead.
	NextSyncAt time.Time `gorm:"column:next_sync_at;type:timestamptz;not null;index"`

	Person *Person `gorm:"foreignKey:PersonID;constraint:OnDelete:CASCADE"`
}
