package repositories

import (
	"context"

	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/dtos"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// AbstractRepository holds the CRUD operations every repository shares. Every method takes
// the request context first so that cancellation and tracing reach the database.
type AbstractRepository[T any] interface {
	FindAll(ctx context.Context, pageableRequest *dtos.PageableRequest, preloads ...string) (*dtos.DataResponse[T], error)
	FindOneByID(ctx context.Context, id string) (*T, error)
	Create(ctx context.Context, entity *T) error
	Save(ctx context.Context, entity *T) error
	SaveAll(ctx context.Context, entities []T) error
	Delete(ctx context.Context, entity *T) error
	Updates(ctx context.Context, entity *T) error
}

// Compile-time check that the shared base implements the shared interface.
var _ AbstractRepository[struct{}] = (*abstractRepository[struct{}])(nil)

type abstractRepository[T any] struct {
	db *db.PostgresDB
}

func (r *abstractRepository[T]) FindAll(ctx context.Context, pr *dtos.PageableRequest, preloads ...string) (*dtos.DataResponse[T], error) {
	query := r.db.From(ctx)

	// Apply preloading if set
	for _, preload := range preloads {
		query = query.Preload(preload)
	}

	return r.find(query, pr)
}

func (r *abstractRepository[T]) FindOneByID(ctx context.Context, id string) (*T, error) {
	var entity T

	// Parse the UUID string to ensure it's valid
	uuid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	res := r.db.From(ctx).First(&entity, "id = ?", uuid)
	if res.Error != nil {
		return nil, res.Error
	}
	return &entity, nil
}

func (r *abstractRepository[T]) Create(ctx context.Context, entity *T) error {
	res := r.db.From(ctx).Create(entity)
	return res.Error
}

func (r *abstractRepository[T]) Save(ctx context.Context, entity *T) error {
	res := r.db.From(ctx).Save(entity)
	return res.Error
}

func (r *abstractRepository[T]) SaveAll(ctx context.Context, entities []T) error {
	res := r.db.From(ctx).Save(&entities)
	return res.Error
}

func (r *abstractRepository[T]) Delete(ctx context.Context, entity *T) error {
	res := r.db.From(ctx).Delete(entity)
	return res.Error
}

func (r *abstractRepository[T]) Updates(ctx context.Context, entity *T) error {
	res := r.db.From(ctx).Updates(entity)
	return res.Error
}

func (r *abstractRepository[T]) find(tx *gorm.DB, pr *dtos.PageableRequest) (*dtos.DataResponse[T], error) {
	var entities []T
	var total int64

	// Ensure model/table is set for the query to avoid "Table not set" issues
	var model T
	tx = tx.Model(&model)

	// Only count when pagination is requested; strip ORDER/LIMIT/OFFSET for faster COUNT
	if pr.ShouldPaginate() {
		countQuery := tx.Session(&gorm.Session{}).
			Clauses(clause.OrderBy{}).
			Limit(-1).Offset(-1)
		if err := countQuery.Count(&total).Error; err != nil {
			return nil, err
		}
		// Early return if nothing to fetch
		if total == 0 {
			return &dtos.DataResponse[T]{
				Data: entities,
				Pageable: &dtos.Pageable{
					Page:     pr.Page,
					PageSize: pr.PageSize,
					Total:    0,
				},
			}, nil
		}
	}

	// Apply pagination if needed
	if pr.ShouldPaginate() {
		tx = tx.Limit(pr.GetLimit()).Offset(pr.GetOffset())
	}

	// Execute the query
	res := tx.Find(&entities)
	if res.Error != nil {
		return nil, res.Error
	}

	// Include pageable metadata only when paginating
	if pr.ShouldPaginate() {
		pageable := &dtos.Pageable{
			Page:     pr.Page,
			PageSize: pr.PageSize,
			Total:    total,
		}
		return &dtos.DataResponse[T]{
			Data:     entities,
			Pageable: pageable,
		}, nil
	}

	return &dtos.DataResponse[T]{
		Data: entities,
	}, nil
}
