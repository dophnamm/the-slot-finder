package services

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"time"

	"the-slot-finder/server/internal/capacity"
	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/metrics"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
)

const (
	opHeatmap         = "heatmap"
	opProgramOverview = "program_overview"
	opProjectMetrics  = "project_metrics"
	opRisks           = "risks"
	resourceHeatmap   = "heatmap"
	resourceMetrics   = "metrics"
)

// HeatmapInput asks for the heatmap of a week range, optionally for one expertise. A nil
// week takes its default: the current week, and the HEATMAP_DEFAULT_WEEKS that follow it.
type HeatmapInput struct {
	From        *isoweek.Week
	To          *isoweek.Week
	ExpertiseID string
}

// InsightsService computes read-only views over the baseline, such as the team heatmap.
type InsightsService interface {
	// Heatmap is every active person's load, capacity, utilization band and hours per
	// expertise per week. With an expertise, it keeps people who hold that skill or work on
	// it in the range.
	Heatmap(ctx context.Context, in HeatmapInput) (*models.Heatmap, error)
	// ProgramOverview is a program's progress and cost figures summed over its projects that
	// are not cancelled, with staffing by expertise.
	ProgramOverview(ctx context.Context, programID string) (*models.ProgramOverview, error)
	// ProjectMetrics is the progress of each project of a program, by project id.
	ProjectMetrics(ctx context.Context, programID string) (map[string]metrics.Progress, error)
	// Risks lists open projects that are understaffed, at risk or behind on delivery, and
	// people over 100 % in the next HEATMAP_DEFAULT_WEEKS weeks.
	Risks(ctx context.Context) (*models.Risks, error)
}

type insightsService struct {
	plans    repositories.PlanRepository
	people   repositories.PersonRepository
	absences repositories.AbsenceRepository
	programs repositories.ProgramRepository
	projects repositories.ProjectRepository
	authz    AuthzService
	clock    clock.Clock
	cfg      *config.Config
}

// ProvideInsightsService creates the insights service.
func ProvideInsightsService(plans repositories.PlanRepository, people repositories.PersonRepository,
	absences repositories.AbsenceRepository, programs repositories.ProgramRepository, projects repositories.ProjectRepository,
	authz AuthzService, clk clock.Clock, cfg *config.Config,
) InsightsService {
	return &insightsService{plans: plans, people: people, absences: absences, programs: programs, projects: projects,
		authz: authz, clock: clk, cfg: cfg}
}

func (s *insightsService) Heatmap(ctx context.Context, in HeatmapInput) (*models.Heatmap, error) {
	if _, err := s.authz.Require(ctx, CanViewInsights); err != nil {
		return nil, err
	}
	return s.heatmap(ctx, in)
}

// heatmap builds the heatmap without the permission check, for Heatmap and Risks.
func (s *insightsService) heatmap(ctx context.Context, in HeatmapInput) (*models.Heatmap, error) {
	from, to, err := s.weekRange(in)
	if err != nil {
		return nil, err
	}
	people, err := s.people.ListActive(ctx)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opHeatmap).WithResource(resourceHeatmap)
	}
	baseline, err := findBaseline(ctx, s.plans, opHeatmap, true)
	if err != nil {
		return nil, err
	}
	absences, err := s.absences.ListOverlapping(ctx, from.Monday(), to.Monday().AddDate(0, 0, daysInWeek-1))
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opHeatmap).WithResource(resourceHeatmap)
	}

	weeks := isoweek.Range(from, to)
	load := loadByPersonWeek(baseline.Assignments)
	rows := make([]models.HeatmapRow, 0, len(people))
	for _, p := range people {
		row := s.heatmapRow(p, weeks, absences, load[p.ID])
		if in.ExpertiseID == "" || keepsExpertise(row, in.ExpertiseID) {
			rows = append(rows, row)
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		a, b := rows[i], rows[j]
		switch {
		case a.PrimaryExpertise.Sequence != b.PrimaryExpertise.Sequence:
			return a.PrimaryExpertise.Sequence < b.PrimaryExpertise.Sequence
		case a.PrimaryExpertise.Name != b.PrimaryExpertise.Name:
			return a.PrimaryExpertise.Name < b.PrimaryExpertise.Name
		case a.Person.Name != b.Person.Name:
			return a.Person.Name < b.Person.Name
		default:
			return a.Person.ID < b.Person.ID
		}
	})
	return &models.Heatmap{Weeks: weeks, Rows: rows}, nil
}

// weekRange applies the defaults and checks the range is in order and not too long.
func (s *insightsService) weekRange(in HeatmapInput) (from, to isoweek.Week, err error) {
	from = isoweek.Of(s.clock.Now())
	if in.From != nil {
		from = *in.From
	}
	to = from.AddWeeks(s.cfg.HeatmapDefaultWeeks - 1)
	if in.To != nil {
		to = *in.To
	}
	if to.Before(from) {
		return from, to, errors.NewAppError(constants.WeekRangeInvalid, constants.MsgWeekRangeInvalid, errors.ErrorTypeValidation,
			http.StatusBadRequest).WithOperation(opHeatmap).WithResource(resourceHeatmap)
	}
	if isoweek.WeeksBetween(from, to)+1 > s.cfg.HeatmapMaxWeeks {
		return from, to, errors.NewAppError(constants.WeekRangeTooLong, fmt.Sprintf(constants.MsgWeekRangeTooLongFormat, s.cfg.HeatmapMaxWeeks),
			errors.ErrorTypeValidation, http.StatusBadRequest).WithOperation(opHeatmap).WithResource(resourceHeatmap)
	}
	return from, to, nil
}

// expertiseLoad is a person's hours on one expertise in one week.
type expertiseLoad map[isoweek.Week]map[string]*models.ExpertiseHours

// loadByPersonWeek sums the baseline's allocations per person, week and expertise.
func loadByPersonWeek(assignments []models.Assignment) map[string]expertiseLoad {
	out := map[string]expertiseLoad{}
	for _, a := range assignments {
		if a.Demand == nil {
			continue
		}
		if out[a.PersonID] == nil {
			out[a.PersonID] = expertiseLoad{}
		}
		for _, alloc := range a.Allocations {
			week := out[a.PersonID][alloc.Week]
			if week == nil {
				week = map[string]*models.ExpertiseHours{}
				out[a.PersonID][alloc.Week] = week
			}
			hours := week[a.Demand.ExpertiseID]
			if hours == nil {
				hours = &models.ExpertiseHours{Expertise: a.Demand.Expertise}
				week[a.Demand.ExpertiseID] = hours
			}
			hours.Hours += alloc.Hours
		}
	}
	return out
}

func (s *insightsService) heatmapRow(p models.Person, weeks []isoweek.Week, absences []models.Absence, load expertiseLoad) models.HeatmapRow {
	person, own := capacityInputs(p, absences)
	policy := capacityPolicy(s.cfg)
	row := models.HeatmapRow{Person: p, Cells: make([]models.HeatmapCell, 0, len(weeks))}
	for _, skill := range p.Skills {
		if skill.IsPrimary {
			row.PrimaryExpertise = skill.Expertise
		}
	}
	for _, week := range weeks {
		cell := models.HeatmapCell{Week: week, Capacity: capacity.Weekly(person, week, own, policy),
			Tentative: capacity.Tentative(person, week, own, policy)}
		for _, hours := range load[week] {
			entry := *hours
			entry.Primary = entry.Expertise.ID == row.PrimaryExpertise.ID
			cell.Load += entry.Hours
			cell.ByExpertise = append(cell.ByExpertise, entry)
		}
		sort.SliceStable(cell.ByExpertise, func(i, j int) bool {
			a, b := cell.ByExpertise[i], cell.ByExpertise[j]
			if a.Primary != b.Primary {
				return a.Primary
			}
			return a.Expertise.Name < b.Expertise.Name
		})
		cell.Utilization, cell.Band = utilizationBand(cell.Load, cell.Capacity, s.cfg.UtilizationUnderusedBelow, s.cfg.UtilizationNearFullFrom)
		row.Cells = append(row.Cells, cell)
	}
	return row
}

// utilizationBand is load ÷ capacity and its band. With no capacity the cell is unavailable,
// or overloaded when it still has load; utilization is then nil. Overloaded is over 100 %.
func utilizationBand(load, capacity, underusedBelow, nearFullFrom float64) (*float64, string) {
	if capacity <= hoursEpsilon {
		if load <= hoursEpsilon {
			return nil, constants.BandUnavailable
		}
		return nil, constants.BandOverloaded
	}
	utilization := load / capacity
	switch {
	case utilization < underusedBelow:
		return &utilization, constants.BandUnderused
	case utilization < nearFullFrom:
		return &utilization, constants.BandHealthy
	case utilization <= 1+hoursEpsilon:
		return &utilization, constants.BandNearFull
	default:
		return &utilization, constants.BandOverloaded
	}
}

// keepsExpertise is true when the person holds the expertise or works on it in the range.
func keepsExpertise(row models.HeatmapRow, expertiseID string) bool {
	for _, skill := range row.Person.Skills {
		if skill.ExpertiseID == expertiseID {
			return true
		}
	}
	for _, cell := range row.Cells {
		for _, hours := range cell.ByExpertise {
			if hours.Expertise.ID == expertiseID {
				return true
			}
		}
	}
	return false
}

// plannedHours is a demand's baseline hours in total and up to and including the current week.
type plannedHours struct {
	total  float64
	toDate float64
}

func plannedByDemand(assignments []models.Assignment, current isoweek.Week) map[string]plannedHours {
	out := map[string]plannedHours{}
	for _, a := range assignments {
		hours := out[a.DemandID]
		for _, alloc := range a.Allocations {
			hours.total += alloc.Hours
			if !alloc.Week.After(current) {
				hours.toDate += alloc.Hours
			}
		}
		out[a.DemandID] = hours
	}
	return out
}

// today is the organization's current date, as a UTC date like the stored dates.
func (s *insightsService) today() time.Time {
	now := s.clock.Now().In(s.clock.Location())
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *insightsService) settings() metrics.Settings {
	return metrics.Settings{ScopeLagThreshold: s.cfg.ScopeLagThreshold, AlertHorizonWeeks: s.cfg.AlertHorizonWeeks}
}

// projectInput is a project's metric input from its demand and the baseline.
func (s *insightsService) projectInput(p models.Project, planned map[string]plannedHours, today time.Time) metrics.Input {
	current := isoweek.Of(today)
	in := metrics.Input{Start: p.StartEarliest, End: p.Deadline, Today: today, ScopePoints: p.ScopePoints, DonePoints: p.DonePoints,
		WeeksToDeadline: isoweek.WeeksBetween(current, p.DeadlineWeek()), LastPhaseLate: p.AtRisk(),
		Closed: p.Status == constants.ProjectStatusDone || p.Status == constants.ProjectStatusCancelled}
	for _, d := range p.Demands {
		in.BudgetHours += d.TotalHours
		in.PlannedHours += planned[d.ID].total
		in.PlannedToDateHours += planned[d.ID].toDate
	}
	return in
}

// metricsData loads the projects (one program, or all) and the baseline's planned hours.
func (s *insightsService) metricsData(ctx context.Context, programID, op string) ([]models.Project, map[string]plannedHours, time.Time, error) {
	projects, err := s.projects.ListForMetrics(ctx, programID)
	if err != nil {
		return nil, nil, time.Time{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceMetrics)
	}
	baseline, err := findBaseline(ctx, s.plans, op, true)
	if err != nil {
		return nil, nil, time.Time{}, err
	}
	today := s.today()
	return projects, plannedByDemand(baseline.Assignments, isoweek.Of(today)), today, nil
}

func (s *insightsService) ProgramOverview(ctx context.Context, programID string) (*models.ProgramOverview, error) {
	if _, err := s.authz.Require(ctx, CanUsePrograms); err != nil {
		return nil, err
	}
	program, err := s.programs.FindWithLead(ctx, programID)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.ProgramNotFound, constants.MsgProgramNotFound, resourceProgram, opProgramOverview)
	}
	projects, planned, today, err := s.metricsData(ctx, programID, opProgramOverview)
	if err != nil {
		return nil, err
	}
	in := metrics.Input{Start: program.StartDate, End: program.EndDate, Today: today,
		WeeksToDeadline: isoweek.WeeksBetween(isoweek.Of(today), isoweek.Of(program.EndDate)),
		Closed:          program.Status == constants.ProgramStatusClosed || program.Status == constants.ProgramStatusCancelled}
	byExpertise := map[string]*models.ExpertiseStaffing{}
	var order []string
	for _, p := range projects {
		pin := s.projectInput(p, planned, today)
		in.ScopePoints += pin.ScopePoints
		in.DonePoints += pin.DonePoints
		in.BudgetHours += pin.BudgetHours
		in.PlannedHours += pin.PlannedHours
		in.PlannedToDateHours += pin.PlannedToDateHours
		in.LastPhaseLate = in.LastPhaseLate || metrics.Compute(pin, s.settings()).AtRisk
		for _, d := range p.Demands {
			row, ok := byExpertise[d.ExpertiseID]
			if !ok {
				row = &models.ExpertiseStaffing{Expertise: d.Expertise}
				byExpertise[d.ExpertiseID] = row
				order = append(order, d.ExpertiseID)
			}
			row.BudgetHours += d.TotalHours
			row.PlannedHours += planned[d.ID].total
		}
	}
	rows := make([]models.ExpertiseStaffing, 0, len(order))
	for _, id := range order {
		row := byExpertise[id]
		progress := metrics.Compute(metrics.Input{BudgetHours: row.BudgetHours, PlannedHours: row.PlannedHours}, s.settings())
		row.GapHours, row.StaffingPercent, row.Understaffed = progress.GapHours, progress.StaffingPercent, progress.Understaffed
		rows = append(rows, *row)
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].Expertise.Sequence != rows[j].Expertise.Sequence {
			return rows[i].Expertise.Sequence < rows[j].Expertise.Sequence
		}
		return rows[i].Expertise.Name < rows[j].Expertise.Name
	})
	return &models.ProgramOverview{Program: *program, Metrics: metrics.Compute(in, s.settings()), ByExpertise: rows, Projects: len(projects)}, nil
}

func (s *insightsService) ProjectMetrics(ctx context.Context, programID string) (map[string]metrics.Progress, error) {
	if _, err := s.authz.Require(ctx, CanUsePrograms); err != nil {
		return nil, err
	}
	projects, planned, today, err := s.metricsData(ctx, programID, opProjectMetrics)
	if err != nil {
		return nil, err
	}
	out := make(map[string]metrics.Progress, len(projects))
	for _, p := range projects {
		out[p.ID] = metrics.Compute(s.projectInput(p, planned, today), s.settings())
	}
	return out, nil
}

func (s *insightsService) Risks(ctx context.Context) (*models.Risks, error) {
	if _, err := s.authz.Require(ctx, CanViewInsights); err != nil {
		return nil, err
	}
	projects, planned, today, err := s.metricsData(ctx, "", opRisks)
	if err != nil {
		return nil, err
	}
	risks := &models.Risks{Projects: []models.ProjectRisk{}, People: []models.PersonOverload{}}
	for _, p := range projects {
		if p.Status == constants.ProjectStatusDone {
			continue
		}
		progress := metrics.Compute(s.projectInput(p, planned, today), s.settings())
		if progress.AtRisk || progress.Understaffed || progress.BehindOnDelivery {
			risks.Projects = append(risks.Projects, models.ProjectRisk{Project: p, Metrics: progress})
		}
	}
	sort.SliceStable(risks.Projects, func(i, j int) bool {
		a, b := risks.Projects[i].Metrics, risks.Projects[j].Metrics
		if a.AtRisk != b.AtRisk {
			return a.AtRisk
		}
		return a.Understaffed && !b.Understaffed
	})

	heatmap, err := s.heatmap(ctx, HeatmapInput{})
	if err != nil {
		return nil, err
	}
	for _, row := range heatmap.Rows {
		var weeks []models.HeatmapCell
		for _, cell := range row.Cells {
			if cell.Band == constants.BandOverloaded {
				weeks = append(weeks, cell)
			}
		}
		if len(weeks) > 0 {
			risks.People = append(risks.People, models.PersonOverload{Person: row.Person, Weeks: weeks})
		}
	}
	return risks, nil
}
