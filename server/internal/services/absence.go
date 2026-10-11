package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"time"

	"the-slot-finder/server/internal/capacity"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"gorm.io/gorm"
)

const (
	opListAbsences  = "list_absences"
	opCreateAbsence = "create_absence"
	opDeleteAbsence = "delete_absence"
	opListHolidays  = "list_holidays"
	opCreateHoliday = "create_holiday"
	resourceAbsence = "absence"
)

// CreateAbsenceInput is a person's leave or training. An empty Status means planned.
type CreateAbsenceInput struct {
	PersonID    string
	FromDate    time.Time
	ToDate      time.Time
	HoursPerDay float64
	Type        string
	Status      string
}

// CreateHolidayInput is a company-wide holiday, from FromDate to ToDate inclusive.
type CreateHolidayInput struct {
	FromDate    time.Time
	ToDate      time.Time
	HoursPerDay float64
}

// AvailabilityService manages absences and holidays. A person may manage their own
// absences; planners may manage anyone's, approve them, and manage holidays.
type AvailabilityService interface {
	ListForPerson(ctx context.Context, personID string) ([]models.Absence, error)
	Create(ctx context.Context, in CreateAbsenceInput) (*models.Absence, error)
	Delete(ctx context.Context, absenceID string) error
	ListHolidays(ctx context.Context) ([]models.Absence, error)
	CreateHoliday(ctx context.Context, in CreateHolidayInput) (*models.Absence, error)
}

type availabilityService struct {
	repo   repositories.AbsenceRepository
	people PeopleService
	authz  AuthzService
}

// ProvideAvailabilityService creates the availability service.
func ProvideAvailabilityService(repo repositories.AbsenceRepository, people PeopleService, authz AuthzService) AvailabilityService {
	return &availabilityService{repo: repo, people: people, authz: authz}
}

func (s *availabilityService) ListForPerson(ctx context.Context, personID string) ([]models.Absence, error) {
	if _, err := s.authz.RequireSelfOr(ctx, personID, CanManageAbsences); err != nil {
		return nil, err
	}
	if _, err := s.people.Get(ctx, personID); err != nil {
		return nil, err
	}
	absences, err := s.repo.ListForPerson(ctx, personID)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgAbsenceListFailed, err).
			WithOperation(opListAbsences).WithResource(resourceAbsence)
	}
	return absences, nil
}

func (s *availabilityService) Create(ctx context.Context, in CreateAbsenceInput) (*models.Absence, error) {
	principal, err := s.authz.RequireSelfOr(ctx, in.PersonID, CanManageAbsences)
	if err != nil {
		return nil, err
	}
	if _, err := s.people.Get(ctx, in.PersonID); err != nil {
		return nil, err
	}
	if err := checkAbsenceDates(in.FromDate, in.ToDate, opCreateAbsence); err != nil {
		return nil, err
	}

	// Only planners approve; anyone else reports planned absence.
	status := in.Status
	if status == "" || !CanManageAbsences(principal) {
		status = constants.AbsenceStatusPlanned
	}
	absence := &models.Absence{
		BaseModel:   models.NewBaseModel(),
		PersonID:    &in.PersonID,
		FromDate:    in.FromDate,
		ToDate:      in.ToDate,
		HoursPerDay: in.HoursPerDay,
		Type:        in.Type,
		Status:      status,
	}
	if err := s.repo.Create(ctx, absence); err != nil {
		return nil, errors.DatabaseError(constants.MsgAbsenceCreateFailed, err).
			WithOperation(opCreateAbsence).WithResource(resourceAbsence)
	}
	return absence, nil
}

func (s *availabilityService) Delete(ctx context.Context, absenceID string) error {
	absence, err := s.repo.FindOneByID(ctx, absenceID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.WrapError(err, constants.AbsenceNotFound, constants.MsgAbsenceNotFound,
				errors.ErrorTypeNotFound, http.StatusNotFound).WithOperation(opDeleteAbsence).WithResource(resourceAbsence)
		}
		return errors.DatabaseError(constants.MsgAbsenceDeleteFailed, err).
			WithOperation(opDeleteAbsence).WithResource(resourceAbsence)
	}

	if absence.PersonID == nil {
		_, err = s.authz.Require(ctx, CanManageHolidays)
	} else {
		_, err = s.authz.RequireSelfOr(ctx, *absence.PersonID, CanManageAbsences)
	}
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, absence); err != nil {
		return errors.DatabaseError(constants.MsgAbsenceDeleteFailed, err).
			WithOperation(opDeleteAbsence).WithResource(resourceAbsence)
	}
	return nil
}

func (s *availabilityService) ListHolidays(ctx context.Context) ([]models.Absence, error) {
	holidays, err := s.repo.ListHolidays(ctx)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgAbsenceListFailed, err).
			WithOperation(opListHolidays).WithResource(resourceAbsence)
	}
	return holidays, nil
}

func (s *availabilityService) CreateHoliday(ctx context.Context, in CreateHolidayInput) (*models.Absence, error) {
	if _, err := s.authz.Require(ctx, CanManageHolidays); err != nil {
		return nil, err
	}
	if err := checkAbsenceDates(in.FromDate, in.ToDate, opCreateHoliday); err != nil {
		return nil, err
	}
	holiday := &models.Absence{
		BaseModel:   models.NewBaseModel(),
		FromDate:    in.FromDate,
		ToDate:      in.ToDate,
		HoursPerDay: in.HoursPerDay,
		Type:        constants.AbsenceTypeHoliday,
		Status:      constants.AbsenceStatusApproved,
	}
	if err := s.repo.Create(ctx, holiday); err != nil {
		return nil, errors.DatabaseError(constants.MsgAbsenceCreateFailed, err).
			WithOperation(opCreateHoliday).WithResource(resourceAbsence)
	}
	return holiday, nil
}

func checkAbsenceDates(from, to time.Time, op string) error {
	if to.Before(from) {
		return errors.NewAppError(constants.AbsenceDatesInvalid, constants.MsgAbsenceDatesInvalid,
			errors.ErrorTypeValidation, http.StatusBadRequest).WithOperation(op).WithResource(resourceAbsence)
	}
	return nil
}

// capacityPolicy is the capacity settings from config.
func capacityPolicy(cfg *config.Config) capacity.Policy {
	return capacity.Policy{WorkingDays: cfg.WorkingWeekdays, CountPlannedAbsence: cfg.CountPlannedAbsence}
}

// capacityInputs is a person's capacity record and the absences that touch them: their
// own and the company-wide holidays.
func capacityInputs(p models.Person, absences []models.Absence) (capacity.Person, []capacity.Absence) {
	var own []capacity.Absence
	for _, a := range absences {
		if a.PersonID == nil || *a.PersonID == p.ID {
			own = append(own, capacity.Absence{From: a.FromDate, To: a.ToDate, HoursPerDay: a.HoursPerDay,
				Planned: a.Status == constants.AbsenceStatusPlanned})
		}
	}
	return capacity.Person{WeeklyHours: p.WeeklyHours, NonProjectHours: p.NonProjectHours, StartDate: p.StartDate, EndDate: p.EndDate}, own
}
