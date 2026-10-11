package models

// AuditEvent records who changed what, with the values before and after.
type AuditEvent struct {
	BaseModel
	// ActorID is the person record of the user, when they have one.
	ActorID *string `gorm:"column:actor_id;type:uuid;index"`
	// ActorSubject is the user's Keycloak subject.
	ActorSubject string         `gorm:"column:actor_subject;not null"`
	Action       string         `gorm:"column:action;not null;index"`
	Resource     string         `gorm:"column:resource;not null"`
	ResourceID   string         `gorm:"column:resource_id;type:uuid;not null;index"`
	Before       map[string]any `gorm:"column:before;serializer:json;type:jsonb"`
	After        map[string]any `gorm:"column:after;serializer:json;type:jsonb"`
}
