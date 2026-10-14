package services

import (
	"context"
	"io"
	"net/http"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/integration/calendar"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"
)

const (
	opImportHolidays = "import_holidays"
	resourceHoliday  = "holiday"
)

// ImportHolidaysInput is a holiday calendar to import: an ICS file or a link, not both, and the
// dates to keep. Every holiday takes HoursPerDay off everyone's capacity.
type ImportHolidaysInput struct {
	File        io.Reader
	URL         string
	From        time.Time
	To          time.Time
	HoursPerDay float64
}

// HolidayImportResult counts the holidays added and the ones that were there already.
type HolidayImportResult struct {
	Added   int
	Skipped int
}

// HolidayImportService turns a holiday calendar into company holidays.
type HolidayImportService interface {
	// Import adds each event of the calendar in the range as a holiday named after it. An event
	// imported before (same UID) is skipped, so importing the same calendar again adds nothing.
	Import(ctx context.Context, in ImportHolidaysInput) (*HolidayImportResult, error)
}

type holidayImportService struct {
	repo    repositories.AbsenceRepository
	authz   AuthzService
	tx      db.TxManager
	events  EventRecorder
	fetcher calendar.Fetcher
	clock   clock.Clock
	cfg     *config.Config
}

// ProvideHolidayImportService creates the holiday import service.
func ProvideHolidayImportService(repo repositories.AbsenceRepository, authz AuthzService, tx db.TxManager, events EventRecorder,
	fetcher calendar.Fetcher, clk clock.Clock, cfg *config.Config,
) HolidayImportService {
	return &holidayImportService{repo: repo, authz: authz, tx: tx, events: events, fetcher: fetcher, clock: clk, cfg: cfg}
}

func (s *holidayImportService) Import(ctx context.Context, in ImportHolidaysInput) (*HolidayImportResult, error) {
	if _, err := s.authz.Require(ctx, CanManageHolidays); err != nil {
		return nil, err
	}
	if err := checkAbsenceDates(in.From, in.To, opImportHolidays); err != nil {
		return nil, err
	}
	data, err := s.read(ctx, in)
	if err != nil {
		return nil, err
	}
	events, err := calendar.ParseHolidays(data, in.From, in.To, s.clock.Location())
	if err != nil {
		return nil, holidaySourceError(err)
	}
	existing, err := s.repo.ListHolidays(ctx)
	if err != nil {
		return nil, errors.DatabaseError(constants.MsgAbsenceListFailed, err).WithOperation(opImportHolidays).WithResource(resourceHoliday)
	}
	imported := make(map[string]bool, len(existing))
	for _, h := range existing {
		if h.ExternalUID != nil {
			imported[*h.ExternalUID] = true
		}
	}

	result := &HolidayImportResult{}
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		*result = HolidayImportResult{}
		for _, event := range events {
			if imported[event.UID] {
				result.Skipped++
				continue
			}
			uid := event.UID
			holiday := &models.Absence{BaseModel: models.NewBaseModel(), Name: event.Name, FromDate: event.From, ToDate: event.To,
				HoursPerDay: in.HoursPerDay, Type: constants.AbsenceTypeHoliday, Status: constants.AbsenceStatusApproved,
				Source: constants.AbsenceSourceCalendar, ExternalUID: &uid}
			if err := s.repo.Create(ctx, holiday); err != nil {
				return err
			}
			if err := s.events.Record(ctx, ChangeInput{Type: constants.EventAbsenceAdded, Payload: absenceChange(holiday, "")}); err != nil {
				return err
			}
			result.Added++
		}
		return nil
	})
	if err != nil {
		return nil, asAppError(err, constants.MsgAbsenceCreateFailed, opImportHolidays, resourceHoliday)
	}
	return result, nil
}

// read is the calendar's bytes from the file or the link, at most CALENDAR_MAX_BYTES. The link is
// never logged or returned.
func (s *holidayImportService) read(ctx context.Context, in ImportHolidaysInput) ([]byte, error) {
	hasFile, hasURL := in.File != nil, in.URL != ""
	if hasFile == hasURL {
		return nil, errors.NewAppError(constants.HolidaySourceInvalid, constants.MsgHolidaySourceInvalid, errors.ErrorTypeValidation,
			http.StatusBadRequest).WithOperation(opImportHolidays).WithResource(resourceHoliday)
	}
	if hasURL {
		if _, err := calendar.CheckURL(in.URL, s.cfg.CalendarAllowedHosts); err != nil {
			return nil, holidaySourceError(err)
		}
		data, err := s.fetcher.Fetch(ctx, in.URL)
		if err != nil {
			return nil, holidaySourceError(err)
		}
		return data, nil
	}
	data, err := io.ReadAll(io.LimitReader(in.File, s.cfg.CalendarMaxBytes+1))
	if err != nil {
		return nil, errors.ValidationError(constants.MsgHolidayFileUnreadable, err).WithOperation(opImportHolidays).WithResource(resourceHoliday)
	}
	if int64(len(data)) > s.cfg.CalendarMaxBytes {
		return nil, holidaySourceError(calendar.ErrTooLarge)
	}
	return data, nil
}

// holidaySourceError is the 400 for a calendar that cannot be fetched or read, with the codes
// calendar sync uses.
func holidaySourceError(err error) error {
	code, msg := feedErrorCode(err)
	return errors.NewAppError(code, msg, errors.ErrorTypeValidation, http.StatusBadRequest).
		WithOperation(opImportHolidays).WithResource(resourceHoliday)
}
