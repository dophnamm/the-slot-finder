package services

import (
	"context"
	stderrors "errors"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/request"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type MockProgramRepository struct {
	mock.Mock
}

func (m *MockProgramRepository) FindAll(ctx context.Context, pr *dtos.PageableRequest, preloads ...string) (*dtos.DataResponse[models.Program], error) {
	args := m.Called(ctx, pr, preloads)
	return args.Get(0).(*dtos.DataResponse[models.Program]), args.Error(1)
}

func (m *MockProgramRepository) FindOneByID(ctx context.Context, id string) (*models.Program, error) {
	args := m.Called(ctx, id)
	return programOrNil(args.Get(0)), args.Error(1)
}

func (m *MockProgramRepository) Create(ctx context.Context, entity *models.Program) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockProgramRepository) Save(ctx context.Context, entity *models.Program) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockProgramRepository) SaveAll(ctx context.Context, entities []models.Program) error {
	return m.Called(ctx, entities).Error(0)
}

func (m *MockProgramRepository) Delete(ctx context.Context, entity *models.Program) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockProgramRepository) Updates(ctx context.Context, entity *models.Program) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockProgramRepository) List(ctx context.Context, page *dtos.PageableRequest, readerID *string) (*dtos.DataResponse[models.Program], error) {
	args := m.Called(ctx, page, readerID)
	return args.Get(0).(*dtos.DataResponse[models.Program]), args.Error(1)
}

func (m *MockProgramRepository) FindWithLead(ctx context.Context, id string) (*models.Program, error) {
	args := m.Called(ctx, id)
	return programOrNil(args.Get(0)), args.Error(1)
}

func programOrNil(v any) *models.Program {
	if v == nil {
		return nil
	}
	return v.(*models.Program)
}

const (
	programID      = "0192f3a4-0000-7000-8000-0000000000d1"
	leadPersonID   = "0192f3a4-0000-7000-8000-0000000000d2"
	otherPersonID2 = "0192f3a4-0000-7000-8000-0000000000d3"
)

// leadPeople links the subject kc-manager to leadPersonID and knows every person id.
type leadPeople struct {
	PeopleService
}

func (leadPeople) FindForPrincipal(_ context.Context, principal request.Principal) (*models.Person, error) {
	if principal.Subject == "kc-manager" {
		return &models.Person{BaseModel: models.BaseModel{ID: leadPersonID}}, nil
	}
	return nil, nil
}

func (leadPeople) Get(_ context.Context, id string) (*models.Person, error) {
	if id == otherPersonID2 || id == leadPersonID {
		return &models.Person{BaseModel: models.BaseModel{ID: id}}, nil
	}
	return nil, notFoundOrDatabase(gorm.ErrRecordNotFound, constants.PersonNotFound, constants.MsgPersonNotFound, resourcePerson, opGetPerson)
}

func newProgramService(repo *MockProgramRepository) ProgramService {
	people := leadPeople{}
	return ProvideProgramService(repo, nil, people, ProvideAuthzService(people), passthroughTx{}, &recordedEvents{})
}

func program(status string, leadID *string) *models.Program {
	return &models.Program{BaseModel: models.BaseModel{ID: programID}, Status: status, LeadID: leadID,
		StartDate: time.Date(2026, time.September, 28, 0, 0, 0, 0, time.UTC), EndDate: time.Date(2026, time.December, 18, 0, 0, 0, 0, time.UTC)}
}

func TestProgramService_ApplyAction(t *testing.T) {
	ownLead := leadPersonID
	otherLead := otherPersonID2
	tests := []struct {
		name       string
		ctx        context.Context
		program    *models.Program
		action     string
		wantStatus string
		wantErr    int
	}{
		{name: "planner holds an open program", ctx: asRole("kc-planner", constants.RolePlanner), program: program(constants.ProgramStatusOpen, nil),
			action: constants.ProgramActionHold, wantStatus: constants.ProgramStatusPending},
		{name: "manager resumes own program", ctx: asRole("kc-manager", constants.RoleManager), program: program(constants.ProgramStatusPending, &ownLead),
			action: constants.ProgramActionResume, wantStatus: constants.ProgramStatusOpen},
		{name: "manager cannot touch another manager's program", ctx: asRole("kc-manager", constants.RoleManager), program: program(constants.ProgramStatusOpen, &otherLead),
			action: constants.ProgramActionHold, wantErr: http.StatusForbidden},
		{name: "close needs closing first", ctx: asRole("kc-planner", constants.RolePlanner), program: program(constants.ProgramStatusOpen, nil),
			action: constants.ProgramActionClose, wantErr: http.StatusConflict},
		{name: "cancelled is final", ctx: asRole("kc-planner", constants.RolePlanner), program: program(constants.ProgramStatusCancelled, nil),
			action: constants.ProgramActionResume, wantErr: http.StatusConflict},
		{name: "unknown action", ctx: asRole("kc-planner", constants.RolePlanner), program: program(constants.ProgramStatusOpen, nil),
			action: "archive", wantErr: http.StatusBadRequest},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockProgramRepository)
			repo.On("FindWithLead", mock.Anything, programID).Return(tt.program, nil)
			if tt.wantErr == 0 {
				repo.On("Save", mock.Anything, mock.MatchedBy(func(p *models.Program) bool { return p.Status == tt.wantStatus })).Return(nil)
			}
			_, err := newProgramService(repo).ApplyAction(tt.ctx, programID, tt.action)
			if tt.wantErr != 0 {
				requireStatus(t, err, tt.wantErr)
				repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
				return
			}
			require.NoError(t, err)
			repo.AssertExpectations(t)
		})
	}
}

func TestProgramService_Create(t *testing.T) {
	start := time.Date(2026, time.September, 28, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.December, 18, 0, 0, 0, 0, time.UTC)
	someone := otherPersonID2
	unknown := "0192f3a4-0000-7000-8000-0000000000ff"

	t.Run("a manager becomes the lead of the program they create", func(t *testing.T) {
		repo := new(MockProgramRepository)
		repo.On("Create", mock.Anything, mock.MatchedBy(func(p *models.Program) bool {
			return p.LeadID != nil && *p.LeadID == leadPersonID && p.Status == constants.ProgramStatusOpen && p.Objectives != nil
		})).Return(nil)
		repo.On("FindWithLead", mock.Anything, mock.Anything).Return(program(constants.ProgramStatusOpen, nil), nil)
		_, err := newProgramService(repo).Create(asRole("kc-manager", constants.RoleManager),
			CreateProgramInput{Name: "P", Client: "C", StartDate: start, EndDate: end, LeadID: &someone})
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("a planner picks the lead", func(t *testing.T) {
		repo := new(MockProgramRepository)
		repo.On("Create", mock.Anything, mock.MatchedBy(func(p *models.Program) bool { return *p.LeadID == someone })).Return(nil)
		repo.On("FindWithLead", mock.Anything, mock.Anything).Return(program(constants.ProgramStatusOpen, nil), nil)
		_, err := newProgramService(repo).Create(asRole("kc-planner", constants.RolePlanner),
			CreateProgramInput{Name: "P", Client: "C", StartDate: start, EndDate: end, LeadID: &someone})
		require.NoError(t, err)
	})

	t.Run("unknown lead", func(t *testing.T) {
		_, err := newProgramService(new(MockProgramRepository)).Create(asRole("kc-planner", constants.RolePlanner),
			CreateProgramInput{Name: "P", Client: "C", StartDate: start, EndDate: end, LeadID: &unknown})
		requireAppError(t, err, constants.LeadNotFound, http.StatusBadRequest)
	})

	t.Run("end before start", func(t *testing.T) {
		_, err := newProgramService(new(MockProgramRepository)).Create(asRole("kc-planner", constants.RolePlanner),
			CreateProgramInput{Name: "P", Client: "C", StartDate: end, EndDate: start})
		requireAppError(t, err, constants.ProgramDatesInvalid, http.StatusBadRequest)
	})
}

func TestProgramService_Update_LeadChange(t *testing.T) {
	ownLead := leadPersonID
	someone := otherPersonID2

	t.Run("a manager edits own program but cannot hand it over", func(t *testing.T) {
		repo := new(MockProgramRepository)
		repo.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, &ownLead), nil)
		_, err := newProgramService(repo).Update(asRole("kc-manager", constants.RoleManager), programID,
			UpdateProgramInput{LeadID: OptionalString{Set: true, Value: &someone}})
		requireStatus(t, err, http.StatusForbidden)
	})

	t.Run("a planner clears the lead", func(t *testing.T) {
		repo := new(MockProgramRepository)
		repo.On("FindWithLead", mock.Anything, programID).Return(program(constants.ProgramStatusOpen, &ownLead), nil)
		repo.On("Save", mock.Anything, mock.MatchedBy(func(p *models.Program) bool { return p.LeadID == nil && p.Name == "Renamed" })).Return(nil)
		_, err := newProgramService(repo).Update(asRole("kc-planner", constants.RolePlanner), programID,
			UpdateProgramInput{Name: new("Renamed"), LeadID: OptionalString{Set: true}})
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})
}

func TestProgramService_RecordsEndDateChange(t *testing.T) {
	day := func(m time.Month, d int) *time.Time { v := time.Date(2026, m, d, 0, 0, 0, 0, time.UTC); return &v }
	tests := []struct {
		name      string
		in        UpdateProgramInput
		failEvent bool
		want      *ProgramEndChange
		wantErr   bool
	}{
		{name: "end moved later", in: UpdateProgramInput{EndDate: day(time.November, 13)},
			want: &ProgramEndChange{ProgramID: programID, Program: "Checkout Revamp", OldEndDate: "2026-10-30", NewEndDate: "2026-11-13"}},
		{name: "end moved earlier", in: UpdateProgramInput{EndDate: day(time.October, 23)},
			want: &ProgramEndChange{ProgramID: programID, Program: "Checkout Revamp", OldEndDate: "2026-10-30", NewEndDate: "2026-10-23"}},
		{name: "same end date", in: UpdateProgramInput{EndDate: day(time.October, 30)}},
		{name: "only the name changes", in: UpdateProgramInput{Name: new("Renamed")}},
		{name: "only the start date changes", in: UpdateProgramInput{StartDate: day(time.October, 5)}},
		{name: "a failed event fails the update", in: UpdateProgramInput{EndDate: day(time.November, 13)}, failEvent: true, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, events := new(MockProgramRepository), &recordedEvents{}
			if tt.failEvent {
				events.fail = stderrors.New("outbox down")
			}
			p := program(constants.ProgramStatusOpen, nil)
			p.Name, p.EndDate = "Checkout Revamp", *day(time.October, 30)
			repo.On("FindWithLead", mock.Anything, programID).Return(p, nil)
			repo.On("Save", mock.Anything, mock.Anything).Return(nil)
			people := leadPeople{}
			service := ProvideProgramService(repo, nil, people, ProvideAuthzService(people), passthroughTx{}, events)
			_, err := service.Update(asRole("kc-planner", constants.RolePlanner), programID, tt.in)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.want == nil {
				require.Empty(t, events.events)
				return
			}
			require.Len(t, events.events, 1)
			require.Equal(t, constants.EventProgramSlipped, events.events[0].Type)
			require.Equal(t, *tt.want, events.events[0].Payload)
			require.Empty(t, events.events[0].ProjectIDs)
		})
	}
}
