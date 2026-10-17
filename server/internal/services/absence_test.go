package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/request"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type MockAbsenceRepository struct {
	mock.Mock
}

func (m *MockAbsenceRepository) FindAll(ctx context.Context, pr *dtos.PageableRequest, preloads ...string) (*dtos.DataResponse[models.Absence], error) {
	args := m.Called(ctx, pr, preloads)
	return args.Get(0).(*dtos.DataResponse[models.Absence]), args.Error(1)
}

func (m *MockAbsenceRepository) FindOneByID(ctx context.Context, id string) (*models.Absence, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Absence), args.Error(1)
}

func (m *MockAbsenceRepository) Create(ctx context.Context, entity *models.Absence) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockAbsenceRepository) Save(ctx context.Context, entity *models.Absence) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockAbsenceRepository) SaveAll(ctx context.Context, entities []models.Absence) error {
	return m.Called(ctx, entities).Error(0)
}

func (m *MockAbsenceRepository) Delete(ctx context.Context, entity *models.Absence) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockAbsenceRepository) Updates(ctx context.Context, entity *models.Absence) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockAbsenceRepository) ListForPerson(ctx context.Context, personID string) ([]models.Absence, error) {
	args := m.Called(ctx, personID)
	return args.Get(0).([]models.Absence), args.Error(1)
}

func (m *MockAbsenceRepository) ListHolidays(ctx context.Context) ([]models.Absence, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Absence), args.Error(1)
}

func (m *MockAbsenceRepository) ListOverlapping(ctx context.Context, from, to time.Time) ([]models.Absence, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).([]models.Absence), args.Error(1)
}

func (m *MockAbsenceRepository) HasOverlapForPerson(ctx context.Context, personID string, from, to time.Time) (bool, error) {
	args := m.Called(ctx, personID, from, to)
	return args.Bool(0), args.Error(1)
}

func (m *MockAbsenceRepository) ListCalendarLeave(ctx context.Context, personID string) ([]models.Absence, error) {
	args := m.Called(ctx, personID)
	return args.Get(0).([]models.Absence), args.Error(1)
}

// stubPeople knows one person, linked to the signed-in specialist.
type stubPeople struct {
	PeopleService
	self *models.Person
}

func (s stubPeople) Get(_ context.Context, id string) (*models.Person, error) {
	return &models.Person{BaseModel: models.BaseModel{ID: id}}, nil
}

func (s stubPeople) FindForPrincipal(_ context.Context, principal request.Principal) (*models.Person, error) {
	if principal.Subject == "kc-specialist" {
		return s.self, nil
	}
	return nil, nil
}

const (
	selfPersonID  = "0192f3a4-0000-7000-8000-0000000000b1"
	otherPersonID = "0192f3a4-0000-7000-8000-0000000000b2"
	absenceID     = "0192f3a4-0000-7000-8000-0000000000c1"
)

func asRole(subject, role string) context.Context {
	return request.NewPrincipalContext(context.Background(), request.Principal{Subject: subject, Role: role, Roles: []string{role}})
}

func newAvailability(repo *MockAbsenceRepository) AvailabilityService {
	people := stubPeople{self: &models.Person{BaseModel: models.BaseModel{ID: selfPersonID}}}
	return ProvideAvailabilityService(repo, people, ProvideAuthzService(people), passthroughTx{}, &recordedEvents{})
}

func TestAvailabilityService_Create(t *testing.T) {
	from := time.Date(2026, time.October, 19, 0, 0, 0, 0, time.UTC)
	to := from.AddDate(0, 0, 4)
	specialist := asRole("kc-specialist", constants.RoleSpecialist)
	planner := asRole("kc-planner", constants.RolePlanner)

	tests := []struct {
		name       string
		ctx        context.Context
		in         CreateAbsenceInput
		wantStatus string
		wantErr    int
	}{
		{name: "specialist reports own leave as planned, even when asking for approved", ctx: specialist,
			in:         CreateAbsenceInput{PersonID: selfPersonID, FromDate: from, ToDate: to, HoursPerDay: 8, Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusApproved},
			wantStatus: constants.AbsenceStatusPlanned},
		{name: "planner approves someone's leave", ctx: planner,
			in:         CreateAbsenceInput{PersonID: otherPersonID, FromDate: from, ToDate: to, HoursPerDay: 8, Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusApproved},
			wantStatus: constants.AbsenceStatusApproved},
		{name: "planner without a status gives planned", ctx: planner,
			in:         CreateAbsenceInput{PersonID: otherPersonID, FromDate: from, ToDate: to, HoursPerDay: 8, Type: constants.AbsenceTypeTraining},
			wantStatus: constants.AbsenceStatusPlanned},
		{name: "specialist cannot add leave for someone else", ctx: specialist,
			in:      CreateAbsenceInput{PersonID: otherPersonID, FromDate: from, ToDate: to, HoursPerDay: 8, Type: constants.AbsenceTypeLeave},
			wantErr: http.StatusForbidden},
		{name: "to before from", ctx: planner,
			in:      CreateAbsenceInput{PersonID: otherPersonID, FromDate: to, ToDate: from, HoursPerDay: 8, Type: constants.AbsenceTypeLeave},
			wantErr: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAbsenceRepository)
			repo.On("HasOverlapForPerson", mock.Anything, tt.in.PersonID, tt.in.FromDate, tt.in.ToDate).Return(false, nil).Maybe()
			if tt.wantErr == 0 {
				repo.On("Create", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool {
					return a.Status == tt.wantStatus && *a.PersonID == tt.in.PersonID && a.Type == tt.in.Type &&
						a.Source == constants.AbsenceSourceManual
				})).Return(nil)
			}
			absence, err := newAvailability(repo).Create(tt.ctx, tt.in)
			if tt.wantErr != 0 {
				requireStatus(t, err, tt.wantErr)
				repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantStatus, absence.Status)
			repo.AssertExpectations(t)
		})
	}
}

func TestAvailabilityService_Create_Overlap(t *testing.T) {
	from := time.Date(2026, time.October, 19, 0, 0, 0, 0, time.UTC)
	in := CreateAbsenceInput{PersonID: otherPersonID, FromDate: from, ToDate: from.AddDate(0, 0, 4), HoursPerDay: 8, Type: constants.AbsenceTypeLeave}
	tests := []struct {
		name     string
		overlaps bool
		wantCode string // "" means success
	}{
		{name: "free dates", overlaps: false},
		{name: "overlaps an existing absence", overlaps: true, wantCode: constants.AbsenceOverlaps},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAbsenceRepository)
			repo.On("HasOverlapForPerson", mock.Anything, in.PersonID, in.FromDate, in.ToDate).Return(tt.overlaps, nil)
			if tt.wantCode == "" {
				repo.On("Create", mock.Anything, mock.Anything).Return(nil)
			}
			_, err := newAvailability(repo).Create(asRole("kc-planner", constants.RolePlanner), in)
			if tt.wantCode == "" {
				require.NoError(t, err)
				return
			}
			requireAppError(t, err, tt.wantCode, http.StatusConflict)
			repo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
		})
	}
}

func TestAvailabilityService_Delete(t *testing.T) {
	self := selfPersonID
	other := otherPersonID
	tests := []struct {
		name    string
		ctx     context.Context
		absence *models.Absence
		findErr error
		wantErr int
	}{
		{name: "own absence", ctx: asRole("kc-specialist", constants.RoleSpecialist), absence: &models.Absence{PersonID: &self}},
		{name: "someone else's absence", ctx: asRole("kc-specialist", constants.RoleSpecialist), absence: &models.Absence{PersonID: &other}, wantErr: http.StatusForbidden},
		{name: "holiday by a specialist", ctx: asRole("kc-specialist", constants.RoleSpecialist), absence: &models.Absence{}, wantErr: http.StatusForbidden},
		{name: "holiday by a planner", ctx: asRole("kc-planner", constants.RolePlanner), absence: &models.Absence{}},
		{name: "not found", ctx: asRole("kc-planner", constants.RolePlanner), findErr: gorm.ErrRecordNotFound, wantErr: http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAbsenceRepository)
			if tt.findErr != nil {
				repo.On("FindOneByID", mock.Anything, absenceID).Return(nil, tt.findErr)
			} else {
				repo.On("FindOneByID", mock.Anything, absenceID).Return(tt.absence, nil)
			}
			if tt.wantErr == 0 {
				repo.On("Delete", mock.Anything, tt.absence).Return(nil)
			}
			err := newAvailability(repo).Delete(tt.ctx, absenceID)
			if tt.wantErr != 0 {
				requireStatus(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestAvailabilityService_Delete_CalendarLeaveIsDismissed(t *testing.T) {
	self, uid := selfPersonID, "event-1"
	absence := &models.Absence{PersonID: &self, Source: constants.AbsenceSourceCalendar, ExternalUID: &uid}
	repo := new(MockAbsenceRepository)
	repo.On("FindOneByID", mock.Anything, absenceID).Return(absence, nil)
	repo.On("DismissCalendarLeave", mock.Anything, absence).Return(nil)

	require.NoError(t, newAvailability(repo).Delete(asRole("kc-specialist", constants.RoleSpecialist), absenceID))

	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestAvailabilityService_CreateHoliday(t *testing.T) {
	day := time.Date(2026, time.December, 25, 0, 0, 0, 0, time.UTC)

	t.Run("planner adds a company-wide holiday", func(t *testing.T) {
		repo := new(MockAbsenceRepository)
		repo.On("Create", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool {
			return a.PersonID == nil && a.Type == constants.AbsenceTypeHoliday && a.Status == constants.AbsenceStatusApproved &&
				a.Name == "Christmas Day"
		})).Return(nil)
		_, err := newAvailability(repo).CreateHoliday(asRole("kc-planner", constants.RolePlanner),
			CreateHolidayInput{Name: "Christmas Day", FromDate: day, ToDate: day, HoursPerDay: 8})
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("manager cannot", func(t *testing.T) {
		_, err := newAvailability(new(MockAbsenceRepository)).CreateHoliday(asRole("kc-manager", constants.RoleManager),
			CreateHolidayInput{FromDate: day, ToDate: day, HoursPerDay: 8})
		requireStatus(t, err, http.StatusForbidden)
	})
}

func TestAvailabilityService_ListForPerson(t *testing.T) {
	repo := new(MockAbsenceRepository)
	repo.On("ListForPerson", mock.Anything, selfPersonID).Return([]models.Absence{{Type: constants.AbsenceTypeLeave}}, nil)
	absences, err := newAvailability(repo).ListForPerson(asRole("kc-specialist", constants.RoleSpecialist), selfPersonID)
	require.NoError(t, err)
	assert.Len(t, absences, 1)

	_, err = newAvailability(new(MockAbsenceRepository)).ListForPerson(asRole("kc-specialist", constants.RoleSpecialist), otherPersonID)
	requireStatus(t, err, http.StatusForbidden)
}

func TestAvailabilityService_Approve(t *testing.T) {
	self := selfPersonID
	planned := func() *models.Absence {
		return &models.Absence{BaseModel: models.BaseModel{ID: absenceID}, PersonID: &self, HoursPerDay: 8,
			FromDate: time.Date(2026, time.October, 22, 0, 0, 0, 0, time.UTC), ToDate: time.Date(2026, time.October, 23, 0, 0, 0, 0, time.UTC),
			Type: constants.AbsenceTypeLeave, Status: constants.AbsenceStatusPlanned}
	}
	approved := planned()
	approved.Status = constants.AbsenceStatusApproved
	// A holiday is never a person's; keep it planned so only the holiday rule can refuse it.
	holiday := &models.Absence{BaseModel: models.BaseModel{ID: absenceID}, Type: constants.AbsenceTypeHoliday,
		Status: constants.AbsenceStatusPlanned}
	planner := asRole("kc-planner", constants.RolePlanner)

	tests := []struct {
		name       string
		ctx        context.Context
		absence    *models.Absence
		findErr    error
		lostRace   bool // another planner approved it between the read and the write
		wantStatus int  // 0 means approved
		wantCode   string
	}{
		{name: "a planner approves planned leave", ctx: planner, absence: planned()},
		{name: "approved by someone else meanwhile", ctx: planner, absence: planned(), lostRace: true,
			wantStatus: http.StatusConflict, wantCode: constants.AbsenceNotPlanned},
		{name: "already approved", ctx: planner, absence: approved, wantStatus: http.StatusConflict, wantCode: constants.AbsenceNotPlanned},
		{name: "a holiday", ctx: planner, absence: holiday, wantStatus: http.StatusConflict, wantCode: constants.AbsenceNotPlanned},
		{name: "not found", ctx: planner, findErr: gorm.ErrRecordNotFound, wantStatus: http.StatusNotFound, wantCode: constants.AbsenceNotFound},
		{name: "a manager cannot approve", ctx: asRole("kc-manager", constants.RoleManager), absence: planned(), wantStatus: http.StatusForbidden},
		{name: "a specialist cannot approve their own", ctx: asRole("kc-specialist", constants.RoleSpecialist), absence: planned(),
			wantStatus: http.StatusForbidden},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockAbsenceRepository)
			if tt.findErr != nil {
				repo.On("FindOneByID", mock.Anything, absenceID).Return(nil, tt.findErr)
			} else {
				repo.On("FindOneByID", mock.Anything, absenceID).Return(tt.absence, nil)
			}
			repo.On("ApprovePlanned", mock.Anything, absenceID).Return(!tt.lostRace, nil)
			events := &recordedEvents{}
			people := stubPeople{self: &models.Person{BaseModel: models.BaseModel{ID: selfPersonID}}}
			service := ProvideAvailabilityService(repo, people, ProvideAuthzService(people), passthroughTx{}, events)

			absence, err := service.Approve(tt.ctx, absenceID)
			if tt.wantStatus != 0 {
				if tt.wantCode != "" {
					requireAppError(t, err, tt.wantCode, tt.wantStatus)
				} else {
					requireStatus(t, err, tt.wantStatus)
				}
				if !tt.lostRace {
					repo.AssertNotCalled(t, "ApprovePlanned", mock.Anything, mock.Anything)
				}
				assert.Empty(t, events.events, "a refused approve records no change event")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, constants.AbsenceStatusApproved, absence.Status)
			require.Len(t, events.events, 1)
			assert.Equal(t, constants.EventAbsenceAdded, events.events[0].Type)
			change, ok := events.events[0].Payload.(AbsenceChange)
			require.True(t, ok)
			assert.Equal(t, constants.AbsenceStatusApproved, change.Status, "the event carries the new status")
			assert.Equal(t, &self, change.PersonID)
		})
	}
}

func (m *MockAbsenceRepository) DetachCalendarLeave(ctx context.Context, personID string) error {
	return m.Called(ctx, personID).Error(0)
}

func (m *MockAbsenceRepository) DismissCalendarLeave(ctx context.Context, absence *models.Absence) error {
	return m.Called(ctx, absence).Error(0)
}

func (m *MockAbsenceRepository) ListDismissedCalendarUIDs(ctx context.Context, personID string) ([]string, error) {
	args := m.Called(ctx, personID)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockAbsenceRepository) ApprovePlanned(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}
