package services

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sort"
	"time"

	"the-slot-finder/server/internal/capacity"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
)

const (
	opPersonSchedule = "person_schedule"
	resourceSchedule = "schedule"
)

// The schedule types are models (computed, never stored), so the dtos package can map them.
type (
	// PersonSchedule is one person's weeks and days in a range.
	PersonSchedule = models.PersonSchedule
	// ScheduleWeek is the person's load against capacity in one week.
	ScheduleWeek = models.ScheduleWeek
	// ScheduleProject is the person's hours on one project and expertise in a week.
	ScheduleProject = models.ScheduleProject
	// ScheduleDay is one working day with absences or holidays.
	ScheduleDay = models.ScheduleDay
	// ScheduleAbsence is one of the person's absences on a day.
	ScheduleAbsence = models.ScheduleAbsence
)

// ScheduleInput asks for one person's schedule from one week through another.
type ScheduleInput struct {
	PersonID string
	From, To isoweek.Week
}

// ScheduleService builds one person's calendar from the baseline and their availability.
type ScheduleService interface {
	// Get returns the person's weeks (capacity, load, band, hours per project) and the days
	// with absences or holidays. Planners and leads may read anyone; others only themselves.
	Get(ctx context.Context, in ScheduleInput) (*PersonSchedule, error)
}

type scheduleService struct {
	plans    repositories.PlanRepository
	people   repositories.PersonRepository
	absences repositories.AbsenceRepository
	actuals  repositories.ActualHoursRepository
	authz    AuthzService
	cfg      *config.Config
}

// ProvideScheduleService creates the schedule service.
func ProvideScheduleService(plans repositories.PlanRepository, people repositories.PersonRepository,
	absences repositories.AbsenceRepository, actuals repositories.ActualHoursRepository, authz AuthzService, cfg *config.Config,
) ScheduleService {
	return &scheduleService{plans: plans, people: people, absences: absences, actuals: actuals, authz: authz, cfg: cfg}
}

func (s *scheduleService) Get(ctx context.Context, in ScheduleInput) (*PersonSchedule, error) {
	// Everyone sees their own; a planner everyone's; a manager the teams of the programs they lead.
	if _, err := s.authz.RequireVisiblePerson(ctx, in.PersonID); err != nil {
		return nil, err
	}
	if err := s.checkRange(in.From, in.To); err != nil {
		return nil, err
	}
	person, err := s.people.FindOneByID(ctx, in.PersonID)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.PersonNotFound, constants.MsgPersonNotFound, resourcePerson, opPersonSchedule)
	}
	plan, err := findBaseline(ctx, s.plans, opPersonSchedule, false)
	if err != nil {
		return nil, err
	}
	assignments, err := s.plans.ListPersonAssignments(ctx, plan.ID, in.PersonID, in.From, in.To)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgPlanLoadFailed, err).WithOperation(opPersonSchedule).WithResource(resourceSchedule)
	}
	absences, err := s.absences.ListOverlapping(ctx, in.From.Monday(), in.To.Monday().AddDate(0, 0, daysInWeek-1))
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opPersonSchedule).WithResource(resourceSchedule)
	}

	worked, err := s.actuals.ByPersonWeeks(ctx, in.PersonID, in.From, in.To)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opPersonSchedule).WithResource(resourceSchedule)
	}
	weeks := isoweek.Range(in.From, in.To)
	scheduled := s.scheduleWeeks(*person, weeks, absences, assignments)
	withActualHours(scheduled, worked)
	return &PersonSchedule{
		WorkingWeekdays: s.cfg.WorkingWeekdays,
		Weeks:           scheduled,
		Days:            s.scheduleDays(person.ID, weeks, absences),
	}, nil
}

// withActualHours sets each week's actual hours from worked; the weeks it lacks stay nil.
func withActualHours(weeks []ScheduleWeek, worked map[isoweek.Week]float64) {
	for i := range weeks {
		if hours, ok := worked[weeks[i].Week]; ok {
			weeks[i].ActualHours = &hours
		}
	}
}

// checkRange refuses a range that runs backwards or is longer than SCHEDULE_MAX_WEEKS.
func (s *scheduleService) checkRange(from, to isoweek.Week) error {
	if to.Before(from) {
		return errors.NewAppError(constants.WeekRangeInvalid, constants.MsgWeekRangeInvalid, errors.ErrorTypeValidation,
			http.StatusBadRequest).WithOperation(opPersonSchedule).WithResource(resourceSchedule)
	}
	if isoweek.WeeksBetween(from, to)+1 > s.cfg.ScheduleMaxWeeks {
		return errors.NewAppError(constants.WeekRangeTooLong, fmt.Sprintf(constants.MsgWeekRangeTooLongFormat, s.cfg.ScheduleMaxWeeks),
			errors.ErrorTypeValidation, http.StatusBadRequest).WithOperation(opPersonSchedule).WithResource(resourceSchedule)
	}
	return nil
}

// scheduleWeeks is capacity, load and band per week, computed as the heatmap computes them,
// with the hours per project and expertise behind the load.
func (s *scheduleService) scheduleWeeks(p models.Person, weeks []isoweek.Week, absences []models.Absence,
	assignments []models.Assignment,
) []ScheduleWeek {
	person, own := capacityInputs(p, absences)
	policy := capacityPolicy(s.cfg)
	projects := projectHoursByWeek(assignments)
	out := make([]ScheduleWeek, 0, len(weeks))
	for _, week := range weeks {
		row := ScheduleWeek{Week: week, Capacity: capacity.Weekly(person, week, own, policy),
			Tentative: capacity.Tentative(person, week, own, policy), Projects: projects[week]}
		for _, project := range row.Projects {
			row.Load += project.Hours
		}
		row.Utilization, row.Band = utilizationBand(row.Load, row.Capacity, s.cfg.UtilizationUnderusedBelow, s.cfg.UtilizationNearFullFrom)
		out = append(out, row)
	}
	return out
}

// projectHoursByWeek sums the allocations per week, project and expertise, most hours first.
// Like the heatmap, it skips assignments without their demand.
func projectHoursByWeek(assignments []models.Assignment) map[isoweek.Week][]ScheduleProject {
	type key struct {
		week                   isoweek.Week
		projectID, expertiseID string
	}
	sums := map[key]*ScheduleProject{}
	for _, a := range assignments {
		if a.Demand == nil {
			continue
		}
		entry := ScheduleProject{Expertise: a.Demand.Expertise.Name}
		if project := a.Demand.Project; project != nil {
			entry.ProjectID, entry.ProjectName, entry.ProgramID = project.ID, project.Name, project.ProgramID
			if project.Program != nil {
				entry.ProgramName = project.Program.Name
			}
		}
		for _, alloc := range a.Allocations {
			k := key{week: alloc.Week, projectID: entry.ProjectID, expertiseID: a.Demand.ExpertiseID}
			if sums[k] == nil {
				copied := entry
				sums[k] = &copied
			}
			sums[k].Hours += alloc.Hours
		}
	}
	out := map[isoweek.Week][]ScheduleProject{}
	for k, entry := range sums {
		out[k.week] = append(out[k.week], *entry)
	}
	for _, projects := range out {
		sort.Slice(projects, func(i, j int) bool {
			a, b := projects[i], projects[j]
			switch {
			case a.Hours != b.Hours:
				return a.Hours > b.Hours
			case a.ProjectName != b.ProjectName:
				return a.ProjectName < b.ProjectName
			default:
				return a.Expertise < b.Expertise
			}
		})
	}
	return out
}

// scheduleDays lists the working days in the weeks that hold one of the person's absences or
// a company-wide holiday. Other people's absences are left out.
func (s *scheduleService) scheduleDays(personID string, weeks []isoweek.Week, absences []models.Absence) []ScheduleDay {
	var days []ScheduleDay
	for _, week := range weeks {
		for _, date := range week.Days() {
			if !slices.Contains(s.cfg.WorkingWeekdays, date.Weekday()) {
				continue
			}
			day := ScheduleDay{Date: date}
			for _, a := range absences {
				if !coversDay(a, date) {
					continue
				}
				switch {
				case a.PersonID == nil:
					day.Holidays = append(day.Holidays, a.ID)
				case *a.PersonID == personID:
					day.Absences = append(day.Absences, ScheduleAbsence{ID: a.ID, Type: a.Type, Status: a.Status, HoursPerDay: a.HoursPerDay, Source: a.Source})
				}
			}
			if len(day.Absences) > 0 || len(day.Holidays) > 0 {
				days = append(days, day)
			}
		}
	}
	return days
}

// coversDay is true when the absence runs over the date. Stored dates are whole days.
func coversDay(a models.Absence, date time.Time) bool {
	return !date.Before(calendarDay(a.FromDate)) && !date.After(calendarDay(a.ToDate))
}

// calendarDay drops the time of day, so dates from the database compare with week days.
func calendarDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}
