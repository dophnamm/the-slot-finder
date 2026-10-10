package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"strings"
	"time"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
	"the-slot-finder/server/internal/request"

	"gorm.io/gorm"
)

const (
	opListPeople       = "list_people"
	opGetPerson        = "get_person"
	opCreatePerson     = "create_person"
	opUpdatePerson     = "update_person"
	opReplaceSkills    = "replace_person_skills"
	opListExpertises   = "list_expertises"
	opFindForPrincipal = "find_person_for_principal"
	resourcePerson     = "person"
	resourceExpertise  = "expertise"
)

// SkillInput is one expertise a person holds.
type SkillInput struct {
	ExpertiseID string
	Level       int
	Primary     bool
}

// ListPeopleInput pages the people list.
type ListPeopleInput struct {
	Page     int
	PageSize int
}

// CreatePersonInput holds a new person. NonProjectHours nil means DEFAULT_NON_PROJECT_HOURS.
type CreatePersonInput struct {
	Name            string
	Email           string
	Role            string
	WeeklyHours     float64
	NonProjectHours *float64
	StartDate       *time.Time
	EndDate         *time.Time
	Skills          []SkillInput
}

// DateChange is an optional date edit: Set false leaves the date alone, Value nil clears it.
type DateChange struct {
	Set   bool
	Value *time.Time
}

// UpdatePersonInput holds the fields a planner changed; nil fields stay as they are.
type UpdatePersonInput struct {
	Name            *string
	Role            *string
	WeeklyHours     *float64
	NonProjectHours *float64
	StartDate       DateChange
	EndDate         DateChange
	Active          *bool
}

// PersonFinder finds the person record of a signed-in user.
type PersonFinder interface {
	// FindForPrincipal matches by Keycloak id, then by email. It returns nil when nobody matches.
	FindForPrincipal(ctx context.Context, principal request.Principal) (*models.Person, error)
}

// PeopleService manages people, their skills and the expertise catalogue.
type PeopleService interface {
	PersonFinder
	List(ctx context.Context, in ListPeopleInput) (*dtos.DataResponse[models.Person], error)
	Get(ctx context.Context, id string) (*models.Person, error)
	Create(ctx context.Context, in CreatePersonInput) (*models.Person, error)
	Update(ctx context.Context, id string, in UpdatePersonInput) (*models.Person, error)
	ReplaceSkills(ctx context.Context, id string, skills []SkillInput) (*models.Person, error)
	ListExpertises(ctx context.Context) ([]models.Expertise, error)
	// LinkPrincipal finds the signed-in user's person and stores their Keycloak id on the
	// first match by email. It returns nil when nobody matches.
	LinkPrincipal(ctx context.Context, principal request.Principal) (*models.Person, error)
}

type peopleService struct {
	repo repositories.PersonRepository
	cfg  *config.Config
}

// ProvidePeopleService creates the people service.
func ProvidePeopleService(repo repositories.PersonRepository, cfg *config.Config) PeopleService {
	return &peopleService{repo: repo, cfg: cfg}
}

// ProvidePersonFinder exposes the people service to services that only need to find a person.
func ProvidePersonFinder(people PeopleService) PersonFinder {
	return people
}

func (s *peopleService) List(ctx context.Context, in ListPeopleInput) (*dtos.DataResponse[models.Person], error) {
	page, err := s.repo.List(ctx, &dtos.PageableRequest{Page: in.Page, PageSize: in.PageSize})
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgPersonListFailed, err).
			WithOperation(opListPeople).WithResource(resourcePerson)
	}
	return page, nil
}

func (s *peopleService) Get(ctx context.Context, id string) (*models.Person, error) {
	person, err := s.repo.FindWithSkills(ctx, id)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.PersonNotFound, constants.MsgPersonNotFound, resourcePerson, opGetPerson)
	}
	return person, nil
}

func (s *peopleService) Create(ctx context.Context, in CreatePersonInput) (*models.Person, error) {
	if err := checkDates(in.StartDate, in.EndDate, opCreatePerson); err != nil {
		return nil, err
	}
	if err := s.checkSkills(ctx, in.Skills, opCreatePerson); err != nil {
		return nil, err
	}

	nonProject := s.cfg.DefaultNonProjectHours
	if in.NonProjectHours != nil {
		nonProject = *in.NonProjectHours
	}
	person := &models.Person{
		BaseModel:       models.NewBaseModel(),
		Name:            in.Name,
		Email:           strings.ToLower(in.Email),
		Role:            in.Role,
		WeeklyHours:     in.WeeklyHours,
		NonProjectHours: nonProject,
		StartDate:       in.StartDate,
		EndDate:         in.EndDate,
		Active:          true,
	}
	person.Skills = toSkillModels(person.ID, in.Skills)

	if err := s.repo.Create(ctx, person); err != nil {
		if stderrors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, emailTaken(err, opCreatePerson)
		}
		return nil, errors.DatabaseError(constants.MsgPersonCreateFailed, err).
			WithOperation(opCreatePerson).WithResource(resourcePerson)
	}
	return s.Get(ctx, person.ID)
}

func (s *peopleService) Update(ctx context.Context, id string, in UpdatePersonInput) (*models.Person, error) {
	person, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if in.Name != nil {
		person.Name = *in.Name
	}
	if in.Role != nil {
		person.Role = *in.Role
	}
	if in.WeeklyHours != nil {
		person.WeeklyHours = *in.WeeklyHours
	}
	if in.NonProjectHours != nil {
		person.NonProjectHours = *in.NonProjectHours
	}
	if in.StartDate.Set {
		person.StartDate = in.StartDate.Value
	}
	if in.EndDate.Set {
		person.EndDate = in.EndDate.Value
	}
	if in.Active != nil {
		person.Active = *in.Active
	}
	if err := checkDates(person.StartDate, person.EndDate, opUpdatePerson); err != nil {
		return nil, err
	}

	// Skills change only through ReplaceSkills; Save must not touch them.
	person.Skills = nil
	if err := s.repo.Save(ctx, person); err != nil {
		return nil, errors.DatabaseError(constants.MsgPersonUpdateFailed, err).
			WithOperation(opUpdatePerson).WithResource(resourcePerson)
	}
	return s.Get(ctx, id)
}

func (s *peopleService) ReplaceSkills(ctx context.Context, id string, skills []SkillInput) (*models.Person, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}
	if err := s.checkSkills(ctx, skills, opReplaceSkills); err != nil {
		return nil, err
	}
	if err := s.repo.ReplaceSkills(ctx, id, toSkillModels(id, skills)); err != nil {
		return nil, errors.DatabaseError(constants.MsgPersonUpdateFailed, err).
			WithOperation(opReplaceSkills).WithResource(resourcePerson)
	}
	return s.Get(ctx, id)
}

func (s *peopleService) ListExpertises(ctx context.Context) ([]models.Expertise, error) {
	expertises, err := s.repo.ListExpertises(ctx)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgExpertiseListFailed, err).
			WithOperation(opListExpertises).WithResource(resourceExpertise)
	}
	return expertises, nil
}

func (s *peopleService) FindForPrincipal(ctx context.Context, principal request.Principal) (*models.Person, error) {
	person, _, err := s.findForPrincipal(ctx, principal)
	return person, err
}

func (s *peopleService) LinkPrincipal(ctx context.Context, principal request.Principal) (*models.Person, error) {
	person, byEmail, err := s.findForPrincipal(ctx, principal)
	if err != nil || person == nil || !byEmail || principal.Subject == "" {
		return person, err
	}
	person.KeycloakID = &principal.Subject
	if err := s.repo.Updates(ctx, &models.Person{BaseModel: person.BaseModel, KeycloakID: person.KeycloakID}); err != nil {
		return nil, errors.DatabaseError(constants.MsgPersonUpdateFailed, err).
			WithOperation(opFindForPrincipal).WithResource(resourcePerson)
	}
	return person, nil
}

// findForPrincipal reports whether the match came from the email, i.e. the link is new.
func (s *peopleService) findForPrincipal(ctx context.Context, principal request.Principal) (*models.Person, bool, error) {
	if principal.Subject != "" {
		person, err := s.repo.FindByKeycloakID(ctx, principal.Subject)
		if err == nil {
			return person, false, nil
		}
		if !stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, errors.DatabaseError(constants.MsgPersonListFailed, err).
				WithOperation(opFindForPrincipal).WithResource(resourcePerson)
		}
	}
	if principal.Email == "" {
		return nil, false, nil
	}
	person, err := s.repo.FindByEmail(ctx, principal.Email)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, errors.DatabaseError(constants.MsgPersonListFailed, err).
			WithOperation(opFindForPrincipal).WithResource(resourcePerson)
	}
	return person, true, nil
}

// checkSkills enforces exactly one primary skill, no repeated expertise, and known expertises.
func (s *peopleService) checkSkills(ctx context.Context, skills []SkillInput, op string) error {
	primaries := 0
	seen := make(map[string]bool, len(skills))
	ids := make([]string, 0, len(skills))
	for _, skill := range skills {
		if skill.Primary {
			primaries++
		}
		if seen[skill.ExpertiseID] {
			return skillsInvalid(constants.MsgSkillsDuplicate, op)
		}
		seen[skill.ExpertiseID] = true
		ids = append(ids, skill.ExpertiseID)
	}
	if primaries != 1 {
		return skillsInvalid(constants.MsgSkillsPrimaryRequired, op)
	}
	count, err := s.repo.CountExpertises(ctx, ids)
	if err != nil {
		return errors.DatabaseError(constants.MsgExpertiseListFailed, err).WithOperation(op).WithResource(resourceExpertise)
	}
	if count != int64(len(ids)) {
		return skillsInvalid(constants.MsgSkillsUnknown, op)
	}
	return nil
}

func checkDates(start, end *time.Time, op string) error {
	if start != nil && end != nil && end.Before(*start) {
		return errors.NewAppError(constants.PersonDatesInvalid, constants.MsgPersonDatesInvalid,
			errors.ErrorTypeValidation, http.StatusBadRequest).WithOperation(op).WithResource(resourcePerson)
	}
	return nil
}

func toSkillModels(personID string, skills []SkillInput) []models.PersonSkill {
	out := make([]models.PersonSkill, 0, len(skills))
	for _, skill := range skills {
		out = append(out, models.PersonSkill{
			PersonID:    personID,
			ExpertiseID: skill.ExpertiseID,
			Level:       skill.Level,
			IsPrimary:   skill.Primary,
		})
	}
	return out
}

func skillsInvalid(message, op string) error {
	return errors.NewAppError(constants.PersonSkillsInvalid, message, errors.ErrorTypeValidation, http.StatusBadRequest).
		WithOperation(op).WithResource(resourcePerson)
}

func emailTaken(err error, op string) error {
	return errors.WrapError(err, constants.PersonEmailAlreadyExists, constants.MsgPersonEmailTaken,
		errors.ErrorTypeConflict, http.StatusConflict).WithOperation(op).WithResource(resourcePerson)
}

// notFoundOrDatabase maps a repository read error to 404 with code and message, or a database error.
func notFoundOrDatabase(err error, code, message, resource, op string) error {
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.WrapError(err, code, message, errors.ErrorTypeNotFound, http.StatusNotFound).
			WithOperation(op).WithResource(resource)
	}
	return errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resource)
}
