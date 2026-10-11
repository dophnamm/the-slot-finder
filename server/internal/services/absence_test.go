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
	return ProvideAvailabilityService(repo, people, ProvideAuthzService(people))
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
			if tt.wantErr == 0 {
				repo.On("Create", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool {
					return a.Status == tt.wantStatus && *a.PersonID == tt.in.PersonID && a.Type == tt.in.Type
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

func TestAvailabilityService_CreateHoliday(t *testing.T) {
	day := time.Date(2026, time.December, 25, 0, 0, 0, 0, time.UTC)

	t.Run("planner adds a company-wide holiday", func(t *testing.T) {
		repo := new(MockAbsenceRepository)
		repo.On("Create", mock.Anything, mock.MatchedBy(func(a *models.Absence) bool {
			return a.PersonID == nil && a.Type == constants.AbsenceTypeHoliday && a.Status == constants.AbsenceStatusApproved
		})).Return(nil)
		_, err := newAvailability(repo).CreateHoliday(asRole("kc-planner", constants.RolePlanner),
			CreateHolidayInput{FromDate: day, ToDate: day, HoursPerDay: 8})
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("lead cannot", func(t *testing.T) {
		_, err := newAvailability(new(MockAbsenceRepository)).CreateHoliday(asRole("kc-lead", constants.RoleLead),
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
