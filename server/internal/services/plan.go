package services

import (
	"context"
	stderrors "errors"
	"net/http"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"gorm.io/gorm"
)

const (
	opEnsureBaseline    = "ensure_baseline"
	opGetPlan           = "get_plan"
	opPersonAssignments = "person_assignments"
	resourcePlan        = "plan"
)

// PersonAssignmentsInput asks for a person's baseline assignments in a week range.
type PersonAssignmentsInput struct {
	PersonID string
	From     isoweek.Week
	To       isoweek.Week
}

// PlanService reads and changes plans: the baseline in force and staffing drafts.
type PlanService interface {
	// EnsureBaseline creates the empty published baseline (version 0) when none exists.
	// It runs at startup and is safe to repeat.
	EnsureBaseline(ctx context.Context) error
	// Current returns the published baseline with its assignments.
	Current(ctx context.Context) (*models.Plan, error)
	// PersonAssignments returns a person's baseline assignments with their allocations in
	// the range; assignments without hours in the range are left out.
	PersonAssignments(ctx context.Context, in PersonAssignmentsInput) ([]models.Assignment, error)
}

type planService struct {
	plans  repositories.PlanRepository
	people repositories.PersonRepository
	authz  AuthzService
}

// ProvidePlanService creates the plan service.
func ProvidePlanService(plans repositories.PlanRepository, people repositories.PersonRepository, authz AuthzService) PlanService {
	return &planService{plans: plans, people: people, authz: authz}
}

func (s *planService) EnsureBaseline(ctx context.Context) error {
	if err := s.plans.EnsureBaseline(ctx); err != nil {
		return errors.DatabaseError(constants.MsgPlanLoadFailed, err).WithOperation(opEnsureBaseline).WithResource(resourcePlan)
	}
	return nil
}

func (s *planService) Current(ctx context.Context) (*models.Plan, error) {
	if _, err := s.authz.Require(ctx, CanViewPlans); err != nil {
		return nil, err
	}
	return findBaseline(ctx, s.plans, opGetPlan, true)
}

// findBaseline loads the published baseline, creating the empty one first if it is missing.
// With assignments it loads them too; without, only the plan record.
func findBaseline(ctx context.Context, plans repositories.PlanRepository, op string, withAssignments bool) (*models.Plan, error) {
	find := plans.FindBaselineSummary
	if withAssignments {
		find = plans.FindBaseline
	}
	plan, err := find(ctx)
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		if err := plans.EnsureBaseline(ctx); err != nil {
			return nil, errors.DatabaseError(constants.MsgPlanLoadFailed, err).WithOperation(op).WithResource(resourcePlan)
		}
		plan, err = find(ctx)
	}
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgPlanLoadFailed, err).WithOperation(op).WithResource(resourcePlan)
	}
	return plan, nil
}

func (s *planService) PersonAssignments(ctx context.Context, in PersonAssignmentsInput) ([]models.Assignment, error) {
	if _, err := s.authz.RequireVisiblePerson(ctx, in.PersonID); err != nil {
		return nil, err
	}
	if in.To.Before(in.From) {
		return nil, errors.NewAppError(constants.WeekRangeInvalid, constants.MsgWeekRangeInvalid, errors.ErrorTypeValidation,
			http.StatusBadRequest).WithOperation(opPersonAssignments).WithResource(resourcePlan)
	}
	if _, err := s.people.FindOneByID(ctx, in.PersonID); err != nil {
		return nil, notFoundOrDatabase(err, constants.PersonNotFound, constants.MsgPersonNotFound, resourcePerson, opPersonAssignments)
	}
	plan, err := findBaseline(ctx, s.plans, opPersonAssignments, false)
	if err != nil {
		return nil, err
	}
	assignments, err := s.plans.ListPersonAssignments(ctx, plan.ID, in.PersonID, in.From, in.To)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgPlanLoadFailed, err).WithOperation(opPersonAssignments).WithResource(resourcePlan)
	}
	out := make([]models.Assignment, 0, len(assignments))
	for _, a := range assignments {
		if len(a.Allocations) > 0 {
			out = append(out, a)
		}
	}
	return out, nil
}
