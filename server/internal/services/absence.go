package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"slices"
	"time"

	"the-slot-finder/server/internal/capacity"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"gorm.io/gorm"
)

const (
	opListAbsences       = "list_absences"
	opCreateAbsence      = "create_absence"
	opDeleteAbsence      = "delete_absence"
	opApproveAbsence     = "approve_absence"
	opApplyCalendarLeave = "apply_calendar_leave"
	opListHolidays       = "list_holidays"
	opCreateHoliday      = "create_holiday"
	resourceAbsence      = "absence"
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

// CalendarLeave is one out-of-office event as leave: dates inclusive, hours per day resolved.
type CalendarLeave struct {
	UID         string
	FromDate    time.Time
	ToDate      time.Time
	HoursPerDay float64
}

// ApplyCalendarLeaveInput is the leave read from one person's calendar for the dates From to To.
type ApplyCalendarLeaveInput struct {
	PersonID string
	From     time.Time
	To       time.Time
	Leave    []CalendarLeave
}

// CalendarLeaveResult counts what one sync changed.
type CalendarLeaveResult struct {
	Created int
	Updated int
	Deleted int
}

// AvailabilityService manages absences and holidays. A person may manage their own
// absences; planners may manage anyone's, approve them, and manage holidays.
type AvailabilityService interface {
	ListForPerson(ctx context.Context, personID string) ([]models.Absence, error)
	Create(ctx context.Context, in CreateAbsenceInput) (*models.Absence, error)
	Delete(ctx context.Context, absenceID string) error
	// Approve turns a person's planned absence into an approved one and records the change.
	// Only planners may; an approved absence or a holiday gets 409.
	Approve(ctx context.Context, absenceID string) (*models.Absence, error)
	ListHolidays(ctx context.Context) ([]models.Absence, error)
	CreateHoliday(ctx context.Context, in CreateHolidayInput) (*models.Absence, error)
	// ApplyCalendarLeave makes the person's calendar absences match in.Leave, in one transaction:
	// a new event creates planned leave, a changed one moves it and sets it back to planned, and
	// one that is gone deletes it when it touches in.From to in.To. Each change records its
	// events. Absences entered by hand are never touched. It checks no access: the calendar
	// feed service does, and its runner acts for the system.
	ApplyCalendarLeave(ctx context.Context, in ApplyCalendarLeaveInput) (CalendarLeaveResult, error)
}

type availabilityService struct {
	repo   repositories.AbsenceRepository
	people PeopleService
	authz  AuthzService
	tx     db.TxManager
	events EventRecorder
}

// ProvideAvailabilityService creates the availability service.
func ProvideAvailabilityService(repo repositories.AbsenceRepository, people PeopleService, authz AuthzService, tx db.TxManager,
	events EventRecorder,
) AvailabilityService {
	return &availabilityService{repo: repo, people: people, authz: authz, tx: tx, events: events}
}

// absenceChange is the event payload for an absence; name is empty for a holiday.
func absenceChange(a *models.Absence, name string) AbsenceChange {
	return AbsenceChange{AbsenceID: a.ID, PersonID: a.PersonID, PersonName: name, FromDate: a.FromDate.Format(constants.DateLayout),
		ToDate: a.ToDate.Format(constants.DateLayout), Type: a.Type, Status: a.Status}
}

// saveWithEvent runs the write and records its change event in one transaction.
func (s *availabilityService) saveWithEvent(ctx context.Context, write func(ctx context.Context) error, event ChangeInput) error {
	return s.tx.Do(ctx, func(ctx context.Context) error {
		if err := write(ctx); err != nil {
			return err
		}
		return s.events.Record(ctx, event)
	})
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
	person, err := s.people.Get(ctx, in.PersonID)
	if err != nil {
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
		Source:      constants.AbsenceSourceManual,
	}
	create := func(ctx context.Context) error { return s.repo.Create(ctx, absence) }
	event := ChangeInput{Type: constants.EventAbsenceAdded, Payload: absenceChange(absence, person.Name)}
	if err := s.saveWithEvent(ctx, create, event); err != nil {
		return nil, asAppError(err, constants.MsgAbsenceCreateFailed, opCreateAbsence, resourceAbsence)
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

	name := ""
	if absence.PersonID != nil {
		if person, err := s.people.Get(ctx, *absence.PersonID); err == nil {
			name = person.Name
		}
	}
	remove := func(ctx context.Context) error { return s.repo.Delete(ctx, absence) }
	event := ChangeInput{Type: constants.EventAbsenceRemoved, Payload: absenceChange(absence, name)}
	if err := s.saveWithEvent(ctx, remove, event); err != nil {
		return asAppError(err, constants.MsgAbsenceDeleteFailed, opDeleteAbsence, resourceAbsence)
	}
	return nil
}

func (s *availabilityService) Approve(ctx context.Context, absenceID string) (*models.Absence, error) {
	if _, err := s.authz.Require(ctx, CanManageAbsences); err != nil {
		return nil, err
	}
	absence, err := s.repo.FindOneByID(ctx, absenceID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WrapError(err, constants.AbsenceNotFound, constants.MsgAbsenceNotFound,
				errors.ErrorTypeNotFound, http.StatusNotFound).WithOperation(opApproveAbsence).WithResource(resourceAbsence)
		}
		return nil, errors.DatabaseError(constants.MsgAbsenceApproveFailed, err).
			WithOperation(opApproveAbsence).WithResource(resourceAbsence)
	}
	if absence.PersonID == nil || absence.Status != constants.AbsenceStatusPlanned {
		return nil, absenceNotPlanned()
	}

	name := ""
	if person, err := s.people.Get(ctx, *absence.PersonID); err == nil {
		name = person.Name
	}
	absence.Status = constants.AbsenceStatusApproved
	// The write only approves a still-planned absence, so of two approvals that both read
	// "planned", one wins and the other records nothing.
	update := func(ctx context.Context) error {
		changed, err := s.repo.ApprovePlanned(ctx, absence.ID)
		if err == nil && !changed {
			return errAbsenceNoLongerPlanned
		}
		return err
	}
	// The constants have only AbsenceAdded and AbsenceRemoved; an approval is the absence
	// again with its new status, so the re-plan runner sees it like any absence change.
	event := ChangeInput{Type: constants.EventAbsenceAdded, Payload: absenceChange(absence, name)}
	if err := s.saveWithEvent(ctx, update, event); err != nil {
		if stderrors.Is(err, errAbsenceNoLongerPlanned) {
			return nil, absenceNotPlanned()
		}
		return nil, asAppError(err, constants.MsgAbsenceApproveFailed, opApproveAbsence, resourceAbsence)
	}
	return absence, nil
}

// errAbsenceNoLongerPlanned means another request approved the absence first.
var errAbsenceNoLongerPlanned = stderrors.New("absence is no longer planned")

// absenceNotPlanned is the 409 for approving an absence that is approved already or a holiday.
func absenceNotPlanned() error {
	return errors.NewAppError(constants.AbsenceNotPlanned, constants.MsgAbsenceNotPlanned, errors.ErrorTypeConflict,
		http.StatusConflict).WithOperation(opApproveAbsence).WithResource(resourceAbsence)
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
		Source:      constants.AbsenceSourceManual,
	}
	create := func(ctx context.Context) error { return s.repo.Create(ctx, holiday) }
	event := ChangeInput{Type: constants.EventAbsenceAdded, Payload: absenceChange(holiday, "")}
	if err := s.saveWithEvent(ctx, create, event); err != nil {
		return nil, asAppError(err, constants.MsgAbsenceCreateFailed, opCreateHoliday, resourceAbsence)
	}
	return holiday, nil
}

func (s *availabilityService) ApplyCalendarLeave(ctx context.Context, in ApplyCalendarLeaveInput) (CalendarLeaveResult, error) {
	person, err := s.people.Get(ctx, in.PersonID)
	if err != nil {
		return CalendarLeaveResult{}, err
	}
	var result CalendarLeaveResult
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		result = CalendarLeaveResult{}
		existing, err := s.repo.ListCalendarLeave(ctx, in.PersonID)
		if err != nil {
			return err
		}
		byUID := make(map[string]models.Absence, len(existing))
		for _, a := range existing {
			if a.ExternalUID != nil {
				byUID[*a.ExternalUID] = a
			}
		}
		for _, leave := range in.Leave {
			old, known := byUID[leave.UID]
			delete(byUID, leave.UID)
			switch {
			case !known:
				err = s.createCalendarLeave(ctx, in.PersonID, person.Name, leave)
				result.Created++
			case !sameLeave(old, leave):
				err = s.moveCalendarLeave(ctx, old, person.Name, leave)
				result.Updated++
			}
			if err != nil {
				return err
			}
		}
		for _, gone := range goneInRange(byUID, in.From, in.To) {
			remove := func(ctx context.Context) error { return s.repo.Delete(ctx, &gone) }
			event := ChangeInput{Type: constants.EventAbsenceRemoved, Payload: absenceChange(&gone, person.Name)}
			if err := s.saveWithEvent(ctx, remove, event); err != nil {
				return err
			}
			result.Deleted++
		}
		return nil
	})
	if err != nil {
		return CalendarLeaveResult{}, asAppError(err, constants.MsgAbsenceCreateFailed, opApplyCalendarLeave, resourceAbsence)
	}
	return result, nil
}

// createCalendarLeave saves a new calendar event as planned leave, with its event.
func (s *availabilityService) createCalendarLeave(ctx context.Context, personID, name string, leave CalendarLeave) error {
	uid := leave.UID
	absence := &models.Absence{BaseModel: models.NewBaseModel(), PersonID: &personID, FromDate: leave.FromDate,
		ToDate: leave.ToDate, HoursPerDay: leave.HoursPerDay, Type: constants.AbsenceTypeLeave,
		Status: constants.AbsenceStatusPlanned, Source: constants.AbsenceSourceCalendar, ExternalUID: &uid}
	create := func(ctx context.Context) error { return s.repo.Create(ctx, absence) }
	return s.saveWithEvent(ctx, create, ChangeInput{Type: constants.EventAbsenceAdded, Payload: absenceChange(absence, name)})
}

// moveCalendarLeave moves calendar leave to the event's new dates and hours, back to planned.
// It records the old dates as removed and the new ones as added, so the re-plan sees both weeks.
func (s *availabilityService) moveCalendarLeave(ctx context.Context, old models.Absence, name string, leave CalendarLeave) error {
	if err := s.events.Record(ctx, ChangeInput{Type: constants.EventAbsenceRemoved, Payload: absenceChange(&old, name)}); err != nil {
		return err
	}
	moved := old
	moved.FromDate, moved.ToDate, moved.HoursPerDay = leave.FromDate, leave.ToDate, leave.HoursPerDay
	moved.Status = constants.AbsenceStatusPlanned
	save := func(ctx context.Context) error { return s.repo.Save(ctx, &moved) }
	return s.saveWithEvent(ctx, save, ChangeInput{Type: constants.EventAbsenceAdded, Payload: absenceChange(&moved, name)})
}

// sameLeave reports calendar leave that already matches the event (status is not compared).
func sameLeave(a models.Absence, leave CalendarLeave) bool {
	return sameDate(a.FromDate, leave.FromDate) && sameDate(a.ToDate, leave.ToDate) &&
		a.HoursPerDay-leave.HoursPerDay < hoursEpsilon && leave.HoursPerDay-a.HoursPerDay < hoursEpsilon
}

func sameDate(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month() && a.Day() == b.Day()
}

// goneInRange is the calendar leave whose event is gone and that touches from to to, oldest first.
func goneInRange(gone map[string]models.Absence, from, to time.Time) []models.Absence {
	var out []models.Absence
	for _, a := range gone {
		if !a.ToDate.Before(from) && !a.FromDate.After(to) {
			out = append(out, a)
		}
	}
	slices.SortFunc(out, func(a, b models.Absence) int { return a.FromDate.Compare(b.FromDate) })
	return out
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
