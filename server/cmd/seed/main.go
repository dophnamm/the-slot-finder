// Command seed loads the demo data into the database, then a demo baseline and what-if
// scenarios (plans.go). Running it again adds nothing new: every row is matched by its natural
// key first. Rows leave their id empty so that
// FirstOrCreate matches on the key only; BaseModel.BeforeCreate fills the id on insert.
//
//	make seed   (runs from cmd/server, so cmd/server/.env is read)
package main

import (
	"fmt"
	"os"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/isoweek"
	"the-slot-finder/server/internal/logger"
	"the-slot-finder/server/internal/models"

	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load configuration: %v\n", err)
		os.Exit(1)
	}
	logger.Init(cfg.LogLevel, cfg.AppEnv.String())

	database := &db.PostgresDB{}
	if err := database.NewPostgresDB(cfg); err != nil {
		logger.Sugar.Fatalf("connect to the database: %v", err)
	}
	defer func() { _ = database.Close() }()

	if err := database.Transaction(seed); err != nil {
		logger.Sugar.Fatalf("seed: %v", err)
	}
	logger.Sugar.Infof("Seed data loaded: %d expertises, %d people, %d absences, %d holidays, %d programs",
		len(expertises), len(people), len(absences), len(holidays), len(programs))

	plan, err := seedPlans(database, cfg)
	if err != nil {
		logger.Sugar.Fatalf("seed plans: %v", err)
	}
	if plan.staffed {
		logger.Sugar.Infof("Baseline staffed: %v", staffedProjects)
	}
	logger.Sugar.Infof("Scenarios added: %d", plan.scenarios)

	actuals, err := seedActuals(database, cfg)
	if err != nil {
		logger.Sugar.Fatalf("seed actual hours: %v", err)
	}
	logger.Sugar.Infof("Actual hours added: %d rows", actuals)
}

func seed(tx *gorm.DB) error {
	expertiseIDs, err := seedExpertises(tx)
	if err != nil {
		return err
	}
	personIDs, err := seedPeople(tx, expertiseIDs)
	if err != nil {
		return err
	}
	if err := seedAbsences(tx, personIDs); err != nil {
		return err
	}
	if err := seedHolidays(tx); err != nil {
		return err
	}
	return seedPrograms(tx, personIDs, expertiseIDs)
}

func seedExpertises(tx *gorm.DB) (map[string]string, error) {
	ids := make(map[string]string, len(expertises))
	for _, e := range expertises {
		row := models.Expertise{Name: e.name, Sequence: e.sequence}
		if err := tx.Where(models.Expertise{Name: e.name}).Attrs(row).FirstOrCreate(&row).Error; err != nil {
			return nil, fmt.Errorf("expertise %s: %w", e.name, err)
		}
		ids[e.name] = row.ID
	}
	return ids, nil
}

func seedPeople(tx *gorm.DB, expertiseIDs map[string]string) (map[string]string, error) {
	ids := make(map[string]string, len(people))
	for _, p := range people {
		row := models.Person{
			Name: p.name, Email: p.email, Role: p.role, WeeklyHours: p.weeklyHours,
			NonProjectHours: p.nonProjectHours, StartDate: p.startDate, Active: true,
		}
		if err := tx.Where(models.Person{Email: p.email}).Attrs(row).FirstOrCreate(&row).Error; err != nil {
			return nil, fmt.Errorf("person %s: %w", p.email, err)
		}
		ids[p.email] = row.ID

		var existing int64
		if err := tx.Model(&models.PersonSkill{}).Where("person_id = ?", row.ID).Count(&existing).Error; err != nil {
			return nil, err
		}
		if existing > 0 {
			continue
		}
		skills := make([]models.PersonSkill, 0, len(p.skills))
		for _, s := range p.skills {
			skills = append(skills, models.PersonSkill{PersonID: row.ID, ExpertiseID: expertiseIDs[s.expertise], Level: s.level, IsPrimary: s.primary})
		}
		if err := tx.Omit("Expertise").Create(&skills).Error; err != nil {
			return nil, fmt.Errorf("skills of %s: %w", p.email, err)
		}
	}
	return ids, nil
}

func seedAbsences(tx *gorm.DB, personIDs map[string]string) error {
	for _, a := range absences {
		personID := personIDs[a.email]
		row := models.Absence{
			PersonID: &personID, FromDate: a.from, ToDate: a.to,
			HoursPerDay: a.hoursPerDay, Type: a.kind, Status: a.status,
		}
		err := tx.Where("person_id = ? AND from_date = ? AND to_date = ? AND type = ?", personID, a.from, a.to, a.kind).
			Attrs(row).FirstOrCreate(&row).Error
		if err != nil {
			return fmt.Errorf("absence of %s: %w", a.email, err)
		}
	}
	return nil
}

func seedHolidays(tx *gorm.DB) error {
	for _, h := range holidays {
		row := models.Absence{
			FromDate: h.from, ToDate: h.to, HoursPerDay: h.hoursPerDay,
			Type: constants.AbsenceTypeHoliday, Status: constants.AbsenceStatusApproved,
		}
		// Assign keeps the name current on a holiday seeded before holidays had names.
		err := tx.Where("person_id IS NULL AND from_date = ? AND to_date = ?", h.from, h.to).
			Attrs(row).Assign(models.Absence{Name: h.name}).FirstOrCreate(&row).Error
		if err != nil {
			return fmt.Errorf("holiday %s: %w", h.from.Format(constants.DateLayout), err)
		}
	}
	return nil
}

func seedPrograms(tx *gorm.DB, personIDs, expertiseIDs map[string]string) error {
	for _, p := range programs {
		lead := personIDs[p.leadEmail]
		row := models.Program{
			Name: p.name, Client: p.client, Status: constants.ProgramStatusOpen, StartDate: p.start, EndDate: p.end,
			Goal: p.goal, Objectives: p.objectives, LeadID: &lead,
		}
		if err := tx.Where(models.Program{Name: p.name}).Attrs(row).FirstOrCreate(&row).Error; err != nil {
			return fmt.Errorf("program %s: %w", p.name, err)
		}
		for _, email := range p.team {
			member := models.ProgramTeamMember{ProgramID: row.ID, PersonID: personIDs[email]}
			if err := tx.Where(member).FirstOrCreate(&member).Error; err != nil {
				return fmt.Errorf("team of %s: %w", p.name, err)
			}
		}
		for _, project := range p.projects {
			if err := seedOneProject(tx, row.ID, project, personIDs, expertiseIDs); err != nil {
				return fmt.Errorf("project %s: %w", project.name, err)
			}
		}
	}
	return nil
}

func seedOneProject(tx *gorm.DB, programID string, p seedProject, personIDs, expertiseIDs map[string]string) error {
	lead := personIDs[p.leadEmail]
	row := models.Project{
		ProgramID: programID, Name: p.name, Priority: p.priority, Status: constants.ProjectStatusIncoming,
		StartEarliest: p.start, Deadline: p.deadline, ScopePoints: p.scopePoints, DonePoints: p.donePts, LeadID: &lead,
	}
	if err := tx.Where(models.Project{ProgramID: programID, Name: p.name}).Attrs(row).FirstOrCreate(&row).Error; err != nil {
		return err
	}

	phaseIDs := make(map[string]string, len(p.phases))
	for i, ph := range p.phases {
		start, err := isoweek.Parse(ph.start)
		if err != nil {
			return err
		}
		end, err := isoweek.Parse(ph.end)
		if err != nil {
			return err
		}
		phase := models.ProjectPhase{ProjectID: row.ID, Name: ph.name, Sequence: i + 1, StartWeek: start, EndWeek: end}
		if err := tx.Where(models.ProjectPhase{ProjectID: row.ID, Name: ph.name}).Attrs(phase).FirstOrCreate(&phase).Error; err != nil {
			return err
		}
		phaseIDs[ph.name] = phase.ID
	}

	for _, d := range p.demands {
		demand := models.Demand{
			ProjectID: row.ID, ExpertiseID: expertiseIDs[d.expertise], MinLevel: d.minLevel, Headcount: d.headcount,
			TotalHours: d.hours, Profile: constants.DemandProfileFlat,
		}
		query := tx.Where("project_id = ? AND expertise_id = ?", row.ID, demand.ExpertiseID)
		if d.phase == "" {
			query = query.Where("phase_id IS NULL")
		} else {
			phaseID := phaseIDs[d.phase]
			demand.PhaseID = &phaseID
			query = query.Where("phase_id = ?", phaseID)
		}
		if err := query.Attrs(demand).FirstOrCreate(&demand).Error; err != nil {
			return err
		}
	}
	return nil
}
