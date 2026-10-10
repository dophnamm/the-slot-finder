// Command seed loads the demo data into the database. Running it again adds nothing new:
// every row is matched by its natural key first. Rows leave their id empty so that
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
	logger.Sugar.Infof("Seed data loaded: %d expertises, %d people, %d absences, %d holidays",
		len(expertises), len(people), len(absences), len(holidays))
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
	return seedHolidays(tx)
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
		err := tx.Where("person_id IS NULL AND from_date = ? AND to_date = ?", h.from, h.to).
			Attrs(row).FirstOrCreate(&row).Error
		if err != nil {
			return fmt.Errorf("holiday %s: %w", h.from.Format(constants.DateLayout), err)
		}
	}
	return nil
}
