package services

import (
	"context"
	"net/http"
	"testing"
	"time"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/request"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

type MockPersonRepository struct {
	mock.Mock
}

func (m *MockPersonRepository) FindAll(ctx context.Context, pr *dtos.PageableRequest, preloads ...string) (*dtos.DataResponse[models.Person], error) {
	args := m.Called(ctx, pr, preloads)
	return args.Get(0).(*dtos.DataResponse[models.Person]), args.Error(1)
}

func (m *MockPersonRepository) FindOneByID(ctx context.Context, id string) (*models.Person, error) {
	args := m.Called(ctx, id)
	return personOrNil(args.Get(0)), args.Error(1)
}

func (m *MockPersonRepository) Create(ctx context.Context, entity *models.Person) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockPersonRepository) Save(ctx context.Context, entity *models.Person) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockPersonRepository) SaveAll(ctx context.Context, entities []models.Person) error {
	return m.Called(ctx, entities).Error(0)
}

func (m *MockPersonRepository) Delete(ctx context.Context, entity *models.Person) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockPersonRepository) Updates(ctx context.Context, entity *models.Person) error {
	return m.Called(ctx, entity).Error(0)
}

func (m *MockPersonRepository) List(ctx context.Context, page *dtos.PageableRequest) (*dtos.DataResponse[models.Person], error) {
	args := m.Called(ctx, page)
	return args.Get(0).(*dtos.DataResponse[models.Person]), args.Error(1)
}

func (m *MockPersonRepository) ListAmong(ctx context.Context, page *dtos.PageableRequest, ids []string) (*dtos.DataResponse[models.Person], error) {
	args := m.Called(ctx, page, ids)
	return args.Get(0).(*dtos.DataResponse[models.Person]), args.Error(1)
}

func (m *MockPersonRepository) FindWithSkills(ctx context.Context, id string) (*models.Person, error) {
	args := m.Called(ctx, id)
	return personOrNil(args.Get(0)), args.Error(1)
}

func (m *MockPersonRepository) FindByEmail(ctx context.Context, email string) (*models.Person, error) {
	args := m.Called(ctx, email)
	return personOrNil(args.Get(0)), args.Error(1)
}

func (m *MockPersonRepository) FindByKeycloakID(ctx context.Context, keycloakID string) (*models.Person, error) {
	args := m.Called(ctx, keycloakID)
	return personOrNil(args.Get(0)), args.Error(1)
}

func (m *MockPersonRepository) ReplaceSkills(ctx context.Context, personID string, skills []models.PersonSkill) error {
	return m.Called(ctx, personID, skills).Error(0)
}

func (m *MockPersonRepository) ListActive(ctx context.Context) ([]models.Person, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Person), args.Error(1)
}

func (m *MockPersonRepository) ListExpertises(ctx context.Context) ([]models.Expertise, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.Expertise), args.Error(1)
}

func (m *MockPersonRepository) CountExpertises(ctx context.Context, ids []string) (int64, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).(int64), args.Error(1)
}

func personOrNil(v any) *models.Person {
	if v == nil {
		return nil
	}
	return v.(*models.Person)
}

const (
	expEngineer = "0192f3a4-0000-7000-8000-000000000001"
	expQA       = "0192f3a4-0000-7000-8000-000000000002"
	personID    = "0192f3a4-0000-7000-8000-0000000000aa"
)

// passthroughTx runs the work without a real transaction, for unit tests.
type passthroughTx struct{}

func (passthroughTx) Do(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func testPeopleConfig() *config.Config {
	return &config.Config{DefaultNonProjectHours: 4}
}

func validCreateInput() CreatePersonInput {
	return CreatePersonInput{
		Name: "Alex", Email: "Alex@Example.test", Role: constants.RoleSpecialist, WeeklyHours: 40,
		Skills: []SkillInput{{ExpertiseID: expEngineer, Level: 4, Primary: true}, {ExpertiseID: expQA, Level: 3}},
	}
}

func requireStatus(t *testing.T, err error, status int) {
	t.Helper()
	var appErr *errors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, status, appErr.HTTPStatus)
}

func requireAppError(t *testing.T, err error, code string, status int) {
	t.Helper()
	var appErr *errors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, code, appErr.Code)
	assert.Equal(t, status, appErr.HTTPStatus)
}

func TestPeopleService_Create(t *testing.T) {
	start := time.Date(2026, time.October, 1, 0, 0, 0, 0, time.UTC)
	before := start.AddDate(0, 0, -1)

	tests := []struct {
		name       string
		input      func(in *CreatePersonInput)
		setup      func(repo *MockPersonRepository)
		wantCode   string
		wantStatus int
	}{
		{
			name: "creates with the default non-project hours",
			setup: func(repo *MockPersonRepository) {
				repo.On("CountExpertises", mock.Anything, []string{expEngineer, expQA}).Return(int64(2), nil)
				repo.On("Create", mock.Anything, mock.MatchedBy(func(p *models.Person) bool {
					return p.Email == "alex@example.test" && p.NonProjectHours == 4 && p.Active && len(p.Skills) == 2 &&
						p.Skills[0].PersonID == p.ID
				})).Return(nil)
				repo.On("FindWithSkills", mock.Anything, mock.Anything).Return(&models.Person{Name: "Alex"}, nil)
			},
		},
		{
			name: "duplicate email",
			setup: func(repo *MockPersonRepository) {
				repo.On("CountExpertises", mock.Anything, mock.Anything).Return(int64(2), nil)
				repo.On("Create", mock.Anything, mock.Anything).Return(gorm.ErrDuplicatedKey)
			},
			wantCode: constants.PersonEmailAlreadyExists, wantStatus: http.StatusConflict,
		},
		{
			name:     "two primary skills",
			input:    func(in *CreatePersonInput) { in.Skills[1].Primary = true },
			wantCode: constants.PersonSkillsInvalid, wantStatus: http.StatusBadRequest,
		},
		{
			name:     "no primary skill",
			input:    func(in *CreatePersonInput) { in.Skills[0].Primary = false },
			wantCode: constants.PersonSkillsInvalid, wantStatus: http.StatusBadRequest,
		},
		{
			name:     "same expertise twice",
			input:    func(in *CreatePersonInput) { in.Skills[1].ExpertiseID = expEngineer },
			wantCode: constants.PersonSkillsInvalid, wantStatus: http.StatusBadRequest,
		},
		{
			name: "unknown expertise",
			setup: func(repo *MockPersonRepository) {
				repo.On("CountExpertises", mock.Anything, mock.Anything).Return(int64(1), nil)
			},
			wantCode: constants.PersonSkillsInvalid, wantStatus: http.StatusBadRequest,
		},
		{
			name:     "end before start",
			input:    func(in *CreatePersonInput) { in.StartDate = &start; in.EndDate = &before },
			wantCode: constants.PersonDatesInvalid, wantStatus: http.StatusBadRequest,
		},
		{
			name:  "non-project hours above weekly hours",
			input: func(in *CreatePersonInput) { in.NonProjectHours = new(41.0) },
			setup: func(repo *MockPersonRepository) {
				repo.On("CountExpertises", mock.Anything, mock.Anything).Return(int64(2), nil)
			},
			wantCode: constants.PersonNonProjectHoursInvalid, wantStatus: http.StatusBadRequest,
		},
		{
			name:  "default non-project hours above a short week",
			input: func(in *CreatePersonInput) { in.WeeklyHours = 3 },
			setup: func(repo *MockPersonRepository) {
				repo.On("CountExpertises", mock.Anything, mock.Anything).Return(int64(2), nil)
			},
			wantCode: constants.PersonNonProjectHoursInvalid, wantStatus: http.StatusBadRequest,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := new(MockPersonRepository)
			if tt.setup != nil {
				tt.setup(repo)
			}
			in := validCreateInput()
			if tt.input != nil {
				tt.input(&in)
			}
			_, err := ProvidePeopleService(repo, passthroughTx{}, testPeopleConfig(), &recordedEvents{}).Create(context.Background(), in)
			if tt.wantCode == "" {
				require.NoError(t, err)
			} else {
				requireAppError(t, err, tt.wantCode, tt.wantStatus)
			}
			repo.AssertExpectations(t)
		})
	}
}

func TestPeopleService_Update(t *testing.T) {
	end := time.Date(2027, time.March, 31, 0, 0, 0, 0, time.UTC)

	t.Run("changes only the given fields and clears a date", func(t *testing.T) {
		repo := new(MockPersonRepository)
		existing := &models.Person{BaseModel: models.BaseModel{ID: personID}, Name: "Alex", WeeklyHours: 40, EndDate: &end,
			Skills: []models.PersonSkill{{ExpertiseID: expEngineer, IsPrimary: true}}}
		repo.On("FindWithSkills", mock.Anything, personID).Return(existing, nil)
		repo.On("Save", mock.Anything, mock.MatchedBy(func(p *models.Person) bool {
			return p.Name == "Alex" && p.WeeklyHours == 32 && p.EndDate == nil && p.Skills == nil
		})).Return(nil)

		_, err := ProvidePeopleService(repo, passthroughTx{}, testPeopleConfig(), &recordedEvents{}).Update(context.Background(), personID, UpdatePersonInput{
			WeeklyHours: new(32.0),
			EndDate:     DateChange{Set: true},
		})
		require.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("weekly hours cut below the stored non-project hours", func(t *testing.T) {
		repo := new(MockPersonRepository)
		existing := &models.Person{BaseModel: models.BaseModel{ID: personID}, WeeklyHours: 40, NonProjectHours: 8}
		repo.On("FindWithSkills", mock.Anything, personID).Return(existing, nil)
		_, err := ProvidePeopleService(repo, passthroughTx{}, testPeopleConfig(), &recordedEvents{}).Update(context.Background(), personID,
			UpdatePersonInput{WeeklyHours: new(6.0)})
		requireAppError(t, err, constants.PersonNonProjectHoursInvalid, http.StatusBadRequest)
		repo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything)
	})

	t.Run("not found", func(t *testing.T) {
		repo := new(MockPersonRepository)
		repo.On("FindWithSkills", mock.Anything, personID).Return(nil, gorm.ErrRecordNotFound)
		_, err := ProvidePeopleService(repo, passthroughTx{}, testPeopleConfig(), &recordedEvents{}).Update(context.Background(), personID, UpdatePersonInput{})
		requireAppError(t, err, constants.PersonNotFound, http.StatusNotFound)
	})
}

func TestPeopleService_LinkPrincipal(t *testing.T) {
	principal := request.Principal{Subject: "kc-1", Email: "alex@example.test", EmailVerified: true}

	t.Run("already linked", func(t *testing.T) {
		repo := new(MockPersonRepository)
		repo.On("FindByKeycloakID", mock.Anything, "kc-1").Return(&models.Person{Name: "Alex"}, nil)
		person, err := ProvidePeopleService(repo, passthroughTx{}, testPeopleConfig(), &recordedEvents{}).LinkPrincipal(context.Background(), principal)
		require.NoError(t, err)
		assert.Equal(t, "Alex", person.Name)
		repo.AssertNotCalled(t, "Updates", mock.Anything, mock.Anything)
	})

	t.Run("first sign-in links by email", func(t *testing.T) {
		repo := new(MockPersonRepository)
		repo.On("FindByKeycloakID", mock.Anything, "kc-1").Return(nil, gorm.ErrRecordNotFound)
		repo.On("FindByEmail", mock.Anything, "alex@example.test").
			Return(&models.Person{BaseModel: models.BaseModel{ID: personID}}, nil)
		repo.On("Updates", mock.Anything, mock.MatchedBy(func(p *models.Person) bool {
			return p.ID == personID && p.KeycloakID != nil && *p.KeycloakID == "kc-1"
		})).Return(nil)
		person, err := ProvidePeopleService(repo, passthroughTx{}, testPeopleConfig(), &recordedEvents{}).LinkPrincipal(context.Background(), principal)
		require.NoError(t, err)
		require.NotNil(t, person.KeycloakID)
		repo.AssertExpectations(t)
	})

	t.Run("unverified email does not link", func(t *testing.T) {
		repo := new(MockPersonRepository)
		repo.On("FindByKeycloakID", mock.Anything, "kc-1").Return(nil, gorm.ErrRecordNotFound)
		unverified := principal
		unverified.EmailVerified = false
		svc := ProvidePeopleService(repo, passthroughTx{}, testPeopleConfig(), &recordedEvents{})
		person, err := svc.LinkPrincipal(context.Background(), unverified)
		require.NoError(t, err)
		assert.Nil(t, person)
		person, err = svc.FindForPrincipal(context.Background(), unverified)
		require.NoError(t, err)
		assert.Nil(t, person)
		repo.AssertNotCalled(t, "FindByEmail", mock.Anything, mock.Anything)
		repo.AssertNotCalled(t, "Updates", mock.Anything, mock.Anything)
	})

	t.Run("row linked to another subject is not relinked", func(t *testing.T) {
		repo := new(MockPersonRepository)
		repo.On("FindByKeycloakID", mock.Anything, "kc-1").Return(nil, gorm.ErrRecordNotFound)
		repo.On("FindByEmail", mock.Anything, "alex@example.test").
			Return(&models.Person{BaseModel: models.BaseModel{ID: personID}, KeycloakID: new("kc-other")}, nil)
		person, err := ProvidePeopleService(repo, passthroughTx{}, testPeopleConfig(), &recordedEvents{}).LinkPrincipal(context.Background(), principal)
		require.NoError(t, err)
		assert.Nil(t, person)
		repo.AssertNotCalled(t, "Updates", mock.Anything, mock.Anything)
	})

	t.Run("no person for this user", func(t *testing.T) {
		repo := new(MockPersonRepository)
		repo.On("FindByKeycloakID", mock.Anything, "kc-1").Return(nil, gorm.ErrRecordNotFound)
		repo.On("FindByEmail", mock.Anything, "alex@example.test").Return(nil, gorm.ErrRecordNotFound)
		person, err := ProvidePeopleService(repo, passthroughTx{}, testPeopleConfig(), &recordedEvents{}).LinkPrincipal(context.Background(), principal)
		require.NoError(t, err)
		assert.Nil(t, person)
	})
}

func TestPeopleService_ReplaceSkills(t *testing.T) {
	repo := new(MockPersonRepository)
	repo.On("FindWithSkills", mock.Anything, personID).Return(&models.Person{BaseModel: models.BaseModel{ID: personID}}, nil)
	repo.On("CountExpertises", mock.Anything, []string{expQA}).Return(int64(1), nil)
	repo.On("ReplaceSkills", mock.Anything, personID, []models.PersonSkill{
		{PersonID: personID, ExpertiseID: expQA, Level: 2, IsPrimary: true},
	}).Return(nil)

	_, err := ProvidePeopleService(repo, passthroughTx{}, testPeopleConfig(), &recordedEvents{}).ReplaceSkills(context.Background(), personID,
		[]SkillInput{{ExpertiseID: expQA, Level: 2, Primary: true}})
	require.NoError(t, err)
	repo.AssertExpectations(t)
}

func (m *MockPersonRepository) ListByEmails(ctx context.Context, emails []string) ([]models.Person, error) {
	args := m.Called(ctx, emails)
	return args.Get(0).([]models.Person), args.Error(1)
}
