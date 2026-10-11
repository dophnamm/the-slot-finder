package repositories

import (
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/models"
)

// AuditRepository stores audit events.
type AuditRepository interface {
	AbstractRepository[models.AuditEvent]
}

type auditRepository struct {
	abstractRepository[models.AuditEvent]
}

// ProvideAuditRepository creates the audit repository.
func ProvideAuditRepository(db *db.PostgresDB) AuditRepository {
	return &auditRepository{abstractRepository: abstractRepository[models.AuditEvent]{db: db}}
}
