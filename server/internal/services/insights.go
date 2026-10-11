package services

import (
	"context"
	"fmt"
	"net/http"
	"sort"

	"the-slot-finder/server/internal/capacity"
	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
)

const (
	opHeatmap       = "heatmap"
	resourceHeatmap = "heatmap"
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
}

type insightsService struct {
	plans    repositories.PlanRepository
	people   repositories.PersonRepository
	absences repositories.AbsenceRepository
	authz    AuthzService
	clock    clock.Clock
	cfg      *config.Config
}

// ProvideInsightsService creates the insights service.
func ProvideInsightsService(plans repositories.PlanRepository, people repositories.PersonRepository,
	absences repositories.AbsenceRepository, authz AuthzService, clk clock.Clock, cfg *config.Config,
) InsightsService {
	return &insightsService{plans: plans, people: people, absences: absences, authz: authz, clock: clk, cfg: cfg}
}

func (s *insightsService) Heatmap(ctx context.Context, in HeatmapInput) (*models.Heatmap, error) {
	if _, err := s.authz.Require(ctx, CanViewInsights); err != nil {
		return nil, err
	}
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
