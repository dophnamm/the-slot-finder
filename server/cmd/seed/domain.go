package main

import (
	"fmt"
	"strings"

	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/logger"
	"the-slot-finder/server/internal/models"

	"gorm.io/gorm"
)

// emailOf builds a demo email from its local part and the configured domain.
func emailOf(local, domain string) string { return local + "@" + domain }

type emailRename struct{ from, to string }

// planRenames picks the legacy-domain emails to move to the target domain. An email whose target
// is already taken (by a person or by an earlier rename) is skipped, because people.email is unique.
func planRenames(emails []string, legacy, target string) (renames []emailRename, skipped int) {
	if target == legacy {
		return nil, 0
	}
	taken := make(map[string]bool, len(emails))
	for _, e := range emails {
		taken[e] = true
	}
	suffix := "@" + legacy
	for _, e := range emails {
		local, ok := strings.CutSuffix(e, suffix)
		if !ok {
			continue
		}
		to := emailOf(local, target)
		if taken[to] {
			skipped++
			continue
		}
		taken[to] = true
		renames = append(renames, emailRename{from: e, to: to})
	}
	return renames, skipped
}

// renameLegacyPeople moves people from the legacy demo domain to the configured one, so seeding
// a new domain updates the existing rows instead of adding duplicates.
func renameLegacyPeople(tx *gorm.DB, domain string) error {
	var emails []string
	if err := tx.Model(&models.Person{}).Order("email ASC").Pluck("email", &emails).Error; err != nil {
		return fmt.Errorf("list people emails: %w", err)
	}
	renames, skipped := planRenames(emails, constants.LegacySeedEmailDomain, domain)
	for _, r := range renames {
		if err := tx.Model(&models.Person{}).Where("email = ?", r.from).Update("email", r.to).Error; err != nil {
			return fmt.Errorf("rename %s: %w", r.from, err)
		}
	}
	if len(renames) > 0 || skipped > 0 {
		logger.Sugar.Infof("Moved people from @%s to @%s: %d renamed, %d skipped (target email exists)",
			constants.LegacySeedEmailDomain, domain, len(renames), skipped)
	}
	return nil
}
