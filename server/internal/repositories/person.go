package repositories

import (
	"context"
	"strings"

	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/models"
)

// personSkillsPreload loads each person's skills with their expertise.
const personSkillsPreload = "Skills.Expertise"

// PersonRepository stores people, the expertise catalogue and people's skills.
type PersonRepository interface {
	AbstractRepository[models.Person]
	// List returns people ordered by name, with their skills.
	List(ctx context.Context, page *dtos.PageableRequest) (*dtos.DataResponse[models.Person], error)
	// FindWithSkills returns one person with their skills.
	FindWithSkills(ctx context.Context, id string) (*models.Person, error)
	FindByEmail(ctx context.Context, email string) (*models.Person, error)
	FindByKeycloakID(ctx context.Context, keycloakID string) (*models.Person, error)
	// ReplaceSkills deletes a person's skills and inserts the new ones. Run it inside
	// TxManager.Do so both steps commit together.
	ReplaceSkills(ctx context.Context, personID string, skills []models.PersonSkill) error
	ListExpertises(ctx context.Context) ([]models.Expertise, error)
	CountExpertises(ctx context.Context, ids []string) (int64, error)
}

type personRepository struct {
	abstractRepository[models.Person]
}

// ProvidePersonRepository creates the person repository.
func ProvidePersonRepository(db *db.PostgresDB) PersonRepository {
	return &personRepository{abstractRepository: abstractRepository[models.Person]{db: db}}
}

func (r *personRepository) List(ctx context.Context, page *dtos.PageableRequest) (*dtos.DataResponse[models.Person], error) {
	query := r.db.From(ctx).
		Preload(personSkillsPreload).
		Order("name ASC").Order("id ASC")
	return r.find(query, page)
}

func (r *personRepository) FindWithSkills(ctx context.Context, id string) (*models.Person, error) {
	var person models.Person
	if err := r.db.From(ctx).Preload(personSkillsPreload).First(&person, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &person, nil
}

func (r *personRepository) FindByEmail(ctx context.Context, email string) (*models.Person, error) {
	var person models.Person
	if err := r.db.From(ctx).First(&person, "email = ?", strings.ToLower(email)).Error; err != nil {
		return nil, err
	}
	return &person, nil
}

func (r *personRepository) FindByKeycloakID(ctx context.Context, keycloakID string) (*models.Person, error) {
	var person models.Person
	if err := r.db.From(ctx).First(&person, "keycloak_id = ?", keycloakID).Error; err != nil {
		return nil, err
	}
	return &person, nil
}

func (r *personRepository) ReplaceSkills(ctx context.Context, personID string, skills []models.PersonSkill) error {
	tx := r.db.From(ctx)
	if err := tx.Where("person_id = ?", personID).Delete(&models.PersonSkill{}).Error; err != nil {
		return err
	}
	if len(skills) == 0 {
		return nil
	}
	return tx.Omit("Expertise").Create(&skills).Error
}

func (r *personRepository) ListExpertises(ctx context.Context) ([]models.Expertise, error) {
	var expertises []models.Expertise
	err := r.db.From(ctx).Order("sequence ASC").Order("name ASC").Order("id ASC").Find(&expertises).Error
	return expertises, err
}

func (r *personRepository) CountExpertises(ctx context.Context, ids []string) (int64, error) {
	var count int64
	err := r.db.From(ctx).Model(&models.Expertise{}).Where("id IN ?", ids).Count(&count).Error
	return count, err
}
