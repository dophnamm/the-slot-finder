package services

import (
	"context"

	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockActualHoursRepository is a testify mock of repositories.ActualHoursRepository.
type MockActualHoursRepository struct {
	mock.Mock
}

func (m *MockActualHoursRepository) Upsert(ctx context.Context, rows []models.ActualHours) (int, int, error) {
	args := m.Called(ctx, rows)
	return args.Int(0), args.Int(1), args.Error(2)
}

func (m *MockActualHoursRepository) SumByProject(ctx context.Context, projectIDs []string) (map[string]float64, error) {
	args := m.Called(ctx, projectIDs)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[string]float64), args.Error(1)
}

func (m *MockActualHoursRepository) ByPersonWeeks(ctx context.Context, personID string, from, to isoweek.Week) (map[isoweek.Week]float64, error) {
	args := m.Called(ctx, personID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(map[isoweek.Week]float64), args.Error(1)
}

func (m *MockActualHoursRepository) ListInRange(ctx context.Context, from, to isoweek.Week) ([]models.ActualHours, error) {
	args := m.Called(ctx, from, to)
	return args.Get(0).([]models.ActualHours), args.Error(1)
}
