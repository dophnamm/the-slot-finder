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
	opHeatmap          = "heatmap"
	opCapacityForecast = "capacity_forecast"
	opProgramOverview  = "program_overview"
	opProgramMembers   = "program_members"
	opProjectMetrics   = "project_metrics"
	opRisks            = "risks"
	resourceHeatmap    = "heatmap"
	resourceMetrics    = "metrics"
)

// HeatmapInput asks for the heatmap of a week range, optionally for one expertise or for the
// team of one program. A nil week takes its default: the current week, and the
// HEATMAP_DEFAULT_WEEKS that follow it.
type HeatmapInput struct {
	From        *isoweek.Week
	To          *isoweek.Week
	ExpertiseID string
	ProgramID   string
	// team is the people of ProgramID, set by Heatmap once it has checked the caller may read it.
	team map[string]bool
}

// InsightsService computes read-only views over the baseline, such as the team heatmap.
type InsightsService interface {
	// Heatmap is every active person's load, capacity, utilization band and hours per
	// expertise per week. With an expertise, it keeps people who hold that skill or work on
	// it in the range. With a program, it keeps the program's team, for whoever may read it.
	Heatmap(ctx context.Context, in HeatmapInput) (*models.Heatmap, error)
	// CapacityForecast is, per expertise and week, free capacity of the people whose primary skill it is, the
	// confirmed demand the baseline does not cover, and incoming projects' demand weighted by confidence. A
	// manager gets their teams' capacity and the demand of the programs they may read. Only From and To are used.
	CapacityForecast(ctx context.Context, in HeatmapInput) (*models.CapacityForecast, error)
	// ProgramOverview is a program's progress and cost figures summed over its projects that
	// are not cancelled, with staffing by expertise.
	ProgramOverview(ctx context.Context, programID string) (*models.ProgramOverview, error)
	// ProgramMembers is everyone with baseline hours on the program's projects, most hours
	// first: their hours per expertise, the weeks they work and the projects.
	ProgramMembers(ctx context.Context, programID string) ([]models.ProgramMember, error)
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
	teams    repositories.ProgramTeamRepository
	actuals  repositories.ActualHoursRepository
	authz    AuthzService
	clock    clock.Clock
	cfg      *config.Config
}

// ProvideInsightsService creates the insights service.
func ProvideInsightsService(plans repositories.PlanRepository, people repositories.PersonRepository,
	absences repositories.AbsenceRepository, programs repositories.ProgramRepository, projects repositories.ProjectRepository,
	teams repositories.ProgramTeamRepository, actuals repositories.ActualHoursRepository, authz AuthzService, clk clock.Clock,
	cfg *config.Config,
) InsightsService {
	return &insightsService{plans: plans, people: people, absences: absences, programs: programs, projects: projects,
		teams: teams, actuals: actuals, authz: authz, clock: clk, cfg: cfg}
}

func (s *insightsService) Heatmap(ctx context.Context, in HeatmapInput) (*models.Heatmap, error) {
	if _, err := s.authz.Require(ctx, CanViewInsights); err != nil {
		return nil, err
	}
	scope, err := s.authz.VisiblePeople(ctx)
	if err != nil {
		return nil, err
	}
	if in.ProgramID != "" {
		if in.team, err = s.programTeam(ctx, in.ProgramID); err != nil {
			return nil, err
		}
	}
	return s.heatmap(ctx, in, scope)
}

// programTeam is the people of a program's team, for a caller who may read the program.
func (s *insightsService) programTeam(ctx context.Context, programID string) (map[string]bool, error) {
	program, err := s.programs.FindWithLead(ctx, programID)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.ProgramNotFound, constants.MsgProgramNotFound, resourceProgram, opHeatmap)
	}
	if err := requireProgramReader(ctx, s.authz, s.projects, program, opHeatmap); err != nil {
		return nil, err
	}
	byProgram, err := s.teams.PersonIDsByProgram(ctx, []string{programID})
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opHeatmap).WithResource(resourceHeatmap)
	}
	team := make(map[string]bool, len(byProgram[programID]))
	for _, id := range byProgram[programID] {
		team[id] = true
	}
	return team, nil
}

// heatmap builds the heatmap of the people in scope, without the permission check, for
// Heatmap and Risks. A manager's scope is the teams of the programs they lead.
func (s *insightsService) heatmap(ctx context.Context, in HeatmapInput, scope PeopleScope) (*models.Heatmap, error) {
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
	actuals, err := s.actuals.ListInRange(ctx, from, to)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(opHeatmap).WithResource(resourceHeatmap)
	}

	weeks := isoweek.Range(from, to)
	load := loadByPersonWeek(withoutFreedHours(baseline.Assignments, isoweek.Of(s.clock.Now())))
	worked := workedByPersonWeek(actuals)
	rows := make([]models.HeatmapRow, 0, len(people))
	for _, p := range people {
		if !scope.Allows(p.ID) || (in.team != nil && !in.team[p.ID]) {
			continue
		}
		row := s.heatmapRow(p, weeks, absences, load[p.ID], worked[p.ID])
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

func (s *insightsService) heatmapRow(p models.Person, weeks []isoweek.Week, absences []models.Absence, load expertiseLoad,
	worked map[isoweek.Week]float64,
) models.HeatmapRow {
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
		if hours, ok := worked[week]; ok {
			cell.ActualHours = &hours
		}
		row.Cells = append(row.Cells, cell)
	}
	overloaded := make([]bool, len(row.Cells))
	for i, cell := range row.Cells {
		overloaded[i] = cell.Band == constants.BandOverloaded
	}
	if start, length, ok := metrics.SustainedOverload(overloaded, s.cfg.SustainedOverloadWeeks); ok {
		row.Sustained = &models.OverloadRun{From: row.Cells[start].Week, Weeks: length}
	}
	return row
}

// workedByPersonWeek sums imported timesheet rows per person and ISO week. A week with rows is
// present even when they sum to 0, so the heatmap can tell "worked 0 h" from "nothing imported".
func workedByPersonWeek(rows []models.ActualHours) map[string]map[isoweek.Week]float64 {
	out := map[string]map[isoweek.Week]float64{}
	for _, row := range rows {
		if out[row.PersonID] == nil {
			out[row.PersonID] = map[isoweek.Week]float64{}
		}
		out[row.PersonID][row.Week] += row.Hours
	}
	return out
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
	// upcoming are the allocations after the current week, for the next weeks' planned hours.
	upcoming []metrics.WeekHours
}

func plannedByDemand(assignments []models.Assignment, current isoweek.Week) map[string]plannedHours {
	out := map[string]plannedHours{}
	for _, a := range assignments {
		hours := out[a.DemandID]
		for _, alloc := range a.Allocations {
			hours.total += alloc.Hours
			if !alloc.Week.After(current) {
				hours.toDate += alloc.Hours
			} else {
				hours.upcoming = append(hours.upcoming,
					metrics.WeekHours{WeeksFromNow: isoweek.WeeksBetween(current, alloc.Week), Hours: alloc.Hours})
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

// projectInput is a project's metric input from its demand, the baseline and its actual hours.
func (s *insightsService) projectInput(p models.Project, planned map[string]plannedHours, actual map[string]float64, today time.Time) metrics.Input {
	current := isoweek.Of(today)
	in := metrics.Input{Start: p.StartEarliest, End: p.Deadline, Today: today, ScopePoints: p.ScopePoints, DonePoints: p.DonePoints,
		WeeksToDeadline: isoweek.WeeksBetween(current, p.DeadlineWeek()), LastPhaseLate: p.AtRisk(),
		Closed: p.Status == constants.ProjectStatusDone || p.Status == constants.ProjectStatusCancelled}
	for _, d := range p.Demands {
		in.BudgetHours += d.TotalHours
		in.PlannedHours += planned[d.ID].total
		in.PlannedToDateHours += planned[d.ID].toDate
		in.Upcoming = append(in.Upcoming, planned[d.ID].upcoming...)
	}
	if hours, ok := actual[p.ID]; ok {
		in.ActualHours = &hours
	}
	return in
}

// metricsData loads the projects (one program, or all), the baseline's planned hours and the
// actual hours of each project that has any.
func (s *insightsService) metricsData(ctx context.Context, programID, op string) ([]models.Project, map[string]plannedHours, map[string]float64, time.Time, error) {
	projects, err := s.projects.ListForMetrics(ctx, programID)
	if err != nil {
		return nil, nil, nil, time.Time{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceMetrics)
	}
	baseline, err := findBaseline(ctx, s.plans, op, true)
	if err != nil {
		return nil, nil, nil, time.Time{}, err
	}
	ids := make([]string, 0, len(projects))
	for _, p := range projects {
		ids = append(ids, p.ID)
	}
	actual, err := s.actuals.SumByProject(ctx, ids)
	if err != nil {
		return nil, nil, nil, time.Time{}, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceMetrics)
	}
	today := s.today()
	return projects, plannedByDemand(baseline.Assignments, isoweek.Of(today)), actual, today, nil
}

func (s *insightsService) ProgramOverview(ctx context.Context, programID string) (*models.ProgramOverview, error) {
	if _, err := s.authz.Require(ctx, CanUsePrograms); err != nil {
		return nil, err
	}
	program, err := s.programs.FindWithLead(ctx, programID)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.ProgramNotFound, constants.MsgProgramNotFound, resourceProgram, opProgramOverview)
	}
	if err := requireProgramReader(ctx, s.authz, s.projects, program, opProgramOverview); err != nil {
		return nil, err
	}
	projects, planned, actual, today, err := s.metricsData(ctx, programID, opProgramOverview)
	if err != nil {
		return nil, err
	}
	in := metrics.Input{Start: program.StartDate, End: program.EndDate, Today: today,
		WeeksToDeadline: isoweek.WeeksBetween(isoweek.Of(today), isoweek.Of(program.EndDate)),
		Closed:          program.Status == constants.ProgramStatusClosed || program.Status == constants.ProgramStatusCancelled}
	byExpertise := map[string]*models.ExpertiseStaffing{}
	var order []string
	for _, p := range projects {
		pin := s.projectInput(p, planned, actual, today)
		in.ScopePoints += pin.ScopePoints
		in.DonePoints += pin.DonePoints
		in.BudgetHours += pin.BudgetHours
		in.PlannedHours += pin.PlannedHours
		in.PlannedToDateHours += pin.PlannedToDateHours
		in.Upcoming = append(in.Upcoming, pin.Upcoming...)
		in.ActualHours = metrics.SumOptional(in.ActualHours, pin.ActualHours)
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

func (s *insightsService) ProgramMembers(ctx context.Context, programID string) ([]models.ProgramMember, error) {
	if _, err := s.authz.Require(ctx, CanUsePrograms); err != nil {
		return nil, err
	}
	program, err := s.programs.FindWithLead(ctx, programID)
	if err != nil {
		return nil, notFoundOrDatabase(err, constants.ProgramNotFound, constants.MsgProgramNotFound, resourceProgram, opProgramMembers)
	}
	if err := requireProgramReader(ctx, s.authz, s.projects, program, opProgramMembers); err != nil {
		return nil, err
	}
	baseline, err := findBaseline(ctx, s.plans, opProgramMembers, true)
	if err != nil {
		return nil, err
	}
	return programMembers(baseline.Assignments, programID), nil
}

// programMembers sums the baseline assignments on one program's projects per person.
func programMembers(assignments []models.Assignment, programID string) []models.ProgramMember {
	byPerson := map[string]*models.ProgramMember{}
	type expertiseKey struct{ expertiseID, kind string }
	expertises := map[string]map[expertiseKey]*models.MemberExpertise{}
	projects := map[string]map[string]models.Project{}
	for _, a := range assignments {
		if a.Person == nil || a.Demand == nil || a.Demand.Project == nil || a.Demand.Project.ProgramID != programID {
			continue
		}
		hours, first, last := allocationSpan(a.Allocations)
		if hours == 0 {
			continue
		}
		member, ok := byPerson[a.PersonID]
		if !ok {
			member = &models.ProgramMember{Person: *a.Person, FromWeek: first, ToWeek: last}
			byPerson[a.PersonID] = member
			expertises[a.PersonID] = map[expertiseKey]*models.MemberExpertise{}
			projects[a.PersonID] = map[string]models.Project{}
		}
		member.TotalHours += hours
		if first.Before(member.FromWeek) {
			member.FromWeek = first
		}
		if last.After(member.ToWeek) {
			member.ToWeek = last
		}
		key := expertiseKey{a.Demand.ExpertiseID, a.Kind}
		if row, ok := expertises[a.PersonID][key]; ok {
			row.Hours += hours
		} else {
			expertises[a.PersonID][key] = &models.MemberExpertise{Expertise: a.Demand.Expertise, Kind: a.Kind, Hours: hours}
		}
		projects[a.PersonID][a.Demand.ProjectID] = *a.Demand.Project
	}
	out := make([]models.ProgramMember, 0, len(byPerson))
	for id, member := range byPerson {
		for _, row := range expertises[id] {
			member.Expertises = append(member.Expertises, *row)
		}
		sort.Slice(member.Expertises, func(i, j int) bool {
			if member.Expertises[i].Hours != member.Expertises[j].Hours {
				return member.Expertises[i].Hours > member.Expertises[j].Hours
			}
			return member.Expertises[i].Expertise.Name < member.Expertises[j].Expertise.Name
		})
		for _, p := range projects[id] {
			member.Projects = append(member.Projects, p)
		}
		sort.Slice(member.Projects, func(i, j int) bool { return member.Projects[i].Name < member.Projects[j].Name })
		out = append(out, *member)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].TotalHours != out[j].TotalHours {
			return out[i].TotalHours > out[j].TotalHours
		}
		return out[i].Person.Name < out[j].Person.Name
	})
	return out
}

// allocationSpan is the hours of an assignment and its first and last weeks with hours.
func allocationSpan(allocations []models.Allocation) (hours float64, first, last isoweek.Week) {
	for _, alloc := range allocations {
		if alloc.Hours <= 0 {
			continue
		}
		hours += alloc.Hours
		if first.IsZero() || alloc.Week.Before(first) {
			first = alloc.Week
		}
		if alloc.Week.After(last) {
			last = alloc.Week
		}
	}
	return hours, first, last
}

func (s *insightsService) ProjectMetrics(ctx context.Context, programID string) (map[string]metrics.Progress, error) {
	if _, err := s.authz.Require(ctx, CanUsePrograms); err != nil {
		return nil, err
	}
	projects, planned, actual, today, err := s.metricsData(ctx, programID, opProjectMetrics)
	if err != nil {
		return nil, err
	}
	out := make(map[string]metrics.Progress, len(projects))
	for _, p := range projects {
		out[p.ID] = metrics.Compute(s.projectInput(p, planned, actual, today), s.settings())
	}
	return out, nil
}

func (s *insightsService) Risks(ctx context.Context) (*models.Risks, error) {
	if _, err := s.authz.Require(ctx, CanViewInsights); err != nil {
		return nil, err
	}
	reader, err := s.authz.ProgramReader(ctx)
	if err != nil {
		return nil, err
	}
	projects, planned, actual, today, err := s.metricsData(ctx, "", opRisks)
	if err != nil {
		return nil, err
	}
	readable := readablePrograms(projects, reader)
	risks := &models.Risks{Projects: []models.ProjectRisk{}, Sustained: []models.SustainedOverload{}, People: []models.PersonOverload{}}
	for _, p := range projects {
		if p.Status == constants.ProjectStatusDone || !readable(p.ProgramID) {
			continue
		}
		progress := metrics.Compute(s.projectInput(p, planned, actual, today), s.settings())
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

	scope, err := s.authz.VisiblePeople(ctx)
	if err != nil {
		return nil, err
	}
	heatmap, err := s.heatmap(ctx, HeatmapInput{}, scope)
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
		if row.Sustained != nil {
			risks.Sustained = append(risks.Sustained, models.SustainedOverload{Person: row.Person, From: row.Sustained.From, Weeks: row.Sustained.Weeks})
		}
		if len(weeks) > 0 {
			risks.People = append(risks.People, models.PersonOverload{Person: row.Person, Weeks: weeks})
		}
	}
	return risks, nil
}

func (s *insightsService) CapacityForecast(ctx context.Context, in HeatmapInput) (*models.CapacityForecast, error) {
	if _, err := s.authz.Require(ctx, CanViewInsights); err != nil {
		return nil, err
	}
	reader, err := s.authz.ProgramReader(ctx)
	if err != nil {
		return nil, err
	}
	scope, err := s.authz.VisiblePeople(ctx)
	if err != nil {
		return nil, err
	}
	in.ExpertiseID, in.ProgramID, in.team = "", "", nil
	heatmap, err := s.heatmap(ctx, in, scope)
	if err != nil {
		return nil, err
	}
	projects, planned, _, today, err := s.metricsData(ctx, "", opCapacityForecast)
	if err != nil {
		return nil, err
	}

	series := map[string]*models.ExpertiseForecast{}
	inputs := map[string][]metrics.ForecastInput{}
	people := map[string][][]metrics.ForecastPerson{}
	expertise := func(e models.Expertise) []metrics.ForecastInput {
		if inputs[e.ID] == nil {
			inputs[e.ID] = make([]metrics.ForecastInput, len(heatmap.Weeks))
			people[e.ID] = make([][]metrics.ForecastPerson, len(heatmap.Weeks))
			series[e.ID] = &models.ExpertiseForecast{Expertise: e}
		}
		return inputs[e.ID]
	}
	for _, row := range heatmap.Rows {
		if row.PrimaryExpertise.ID == "" {
			continue
		}
		weeks := expertise(row.PrimaryExpertise)
		for i, cell := range row.Cells {
			weeks[i].Capacity += cell.Capacity
			weeks[i].Load += cell.Load
			weeks[i].FreeHours += max(0, cell.Capacity-cell.Load)
			people[row.PrimaryExpertise.ID][i] = append(people[row.PrimaryExpertise.ID][i], metrics.ForecastPerson{
				PersonID: row.Person.ID, Name: row.Person.Name, Capacity: cell.Capacity, Load: cell.Load,
				Utilization: cell.Utilization, Band: cell.Band})
		}
	}

	addPipelineDemand(projects, planned, readablePrograms(projects, reader), isoweek.Of(today), heatmap.Weeks[0], expertise)

	out := &models.CapacityForecast{Weeks: heatmap.Weeks, Expertises: make([]models.ExpertiseForecast, 0, len(series))}
	for id, e := range series {
		weeks, start, length, sustained := metrics.CapacityForecast(inputs[id], s.cfg.PipelineConfidencePercent, s.cfg.SustainedOverloadWeeks)
		for i := range weeks {
			weeks[i].People = metrics.RankForecastPeople(people[id][i])
		}
		e.Weeks = weeks
		if sustained {
			e.Sustained = &models.OverloadRun{From: heatmap.Weeks[start], Weeks: length}
		}
		out.Expertises = append(out.Expertises, *e)
	}
	sort.SliceStable(out.Expertises, func(i, j int) bool {
		a, b := out.Expertises[i].Expertise, out.Expertises[j].Expertise
		if a.Sequence != b.Sequence {
			return a.Sequence < b.Sequence
		}
		return a.Name < b.Name
	})
	return out, nil
}

// addPipelineDemand spreads each readable project's unplanned demand hours over its remaining weeks.
func addPipelineDemand(projects []models.Project, planned map[string]plannedHours, readable func(string) bool, current, first isoweek.Week, expertise func(models.Expertise) []metrics.ForecastInput) {
	for _, p := range projects {
		incoming := p.Status == constants.ProjectStatusIncoming
		if !readable(p.ProgramID) || (!incoming && p.Status != constants.ProjectStatusPlanned && p.Status != constants.ProjectStatusActive) {
			continue
		}
		for _, d := range p.Demands {
			left := d.TotalHours - planned[d.ID].total
			from, to := p.DemandWindow(d)
			if current.After(from) {
				from = current
			}
			if left <= hoursEpsilon || to.Before(from) {
				continue
			}
			// Flat spread over the window's remaining weeks (the only profile built).
			each := left / float64(isoweek.WeeksBetween(from, to)+1)
			weeks := expertise(d.Expertise)
			for _, week := range isoweek.Range(from, to) {
				i := isoweek.WeeksBetween(first, week)
				if i < 0 || i >= len(weeks) {
					continue
				}
				if incoming {
					weeks[i].PipelineHours += each
				} else {
					weeks[i].UnmetHours += each
				}
			}
		}
	}
}
