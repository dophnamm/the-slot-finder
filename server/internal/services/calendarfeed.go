package services

import (
	"context"
	stderrors "errors"
	"math"
	"net/http"
	"strings"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/db"
	"the-slot-finder/server/internal/errors"
	"the-slot-finder/server/internal/integration/calendar"
	"the-slot-finder/server/internal/logger"
	"the-slot-finder/server/internal/models"
	"the-slot-finder/server/internal/repositories"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	opGetCalendarFeed        = "get_calendar_feed"
	opConnectCalendarFeed    = "connect_calendar_feed"
	opDisconnectCalendarFeed = "disconnect_calendar_feed"
	opSyncCalendarFeed       = "sync_calendar_feed"
	resourceCalendarFeed     = "calendar_feed"
)

// hoursPerDayScale rounds leave hours to two decimals, as absences store them (numeric(4,2)),
// so an unchanged event never looks moved after a round trip through the database.
const hoursPerDayScale = 100

// daysPerWeek turns CALENDAR_SYNC_WEEKS into days.
const daysPerWeek = 7

// ConnectFeedInput is a person's secret calendar link.
type ConnectFeedInput struct {
	PersonID string
	URL      string
}

// CalendarFeedService connects people's calendars and turns their out-of-office events into
// planned leave. A person may manage their own calendar; planners anyone's.
type CalendarFeedService interface {
	Get(ctx context.Context, personID string) (*models.CalendarFeed, error)
	// Connect checks the link, syncs it once, and stores it encrypted, replacing an earlier one.
	Connect(ctx context.Context, in ConnectFeedInput) (*models.CalendarFeed, error)
	Disconnect(ctx context.Context, personID string) error
	// SyncNow syncs the person's calendar at once and returns its status.
	SyncNow(ctx context.Context, personID string) (*models.CalendarFeed, error)
	// RunOnce claims the calendars that are due and syncs them; it returns how many it claimed.
	RunOnce(ctx context.Context) (int, error)
	// Run calls RunOnce every CALENDAR_SYNC_INTERVAL until ctx ends. It returns at once while
	// calendar sync is off.
	Run(ctx context.Context)
}

type calendarFeedService struct {
	feeds        repositories.CalendarFeedRepository
	people       PeopleService
	availability AvailabilityService
	authz        AuthzService
	tx           db.TxManager
	clock        clock.Clock
	cfg          *config.Config
	fetcher      calendar.Fetcher
	// sealer is nil while CALENDAR_FEED_KEY is empty: the feature is off.
	sealer *calendar.Sealer
}

// ProvideCalendarFeedService creates the calendar feed service. An empty CALENDAR_FEED_KEY turns
// the feature off; a malformed one stops the server at startup.
func ProvideCalendarFeedService(feeds repositories.CalendarFeedRepository, people PeopleService,
	availability AvailabilityService, authz AuthzService, tx db.TxManager, clk clock.Clock, cfg *config.Config,
	fetcher calendar.Fetcher,
) (CalendarFeedService, error) {
	s := &calendarFeedService{feeds: feeds, people: people, availability: availability, authz: authz, tx: tx,
		clock: clk, cfg: cfg, fetcher: fetcher}
	if cfg.CalendarFeedKey == "" {
		return s, nil
	}
	sealer, err := calendar.NewSealer(cfg.CalendarFeedKey)
	if err != nil {
		return nil, err
	}
	s.sealer = sealer
	return s, nil
}

func (s *calendarFeedService) Get(ctx context.Context, personID string) (*models.CalendarFeed, error) {
	if err := s.check(ctx, personID, opGetCalendarFeed); err != nil {
		return nil, err
	}
	return s.find(ctx, personID, opGetCalendarFeed)
}

func (s *calendarFeedService) Connect(ctx context.Context, in ConnectFeedInput) (*models.CalendarFeed, error) {
	if err := s.check(ctx, in.PersonID, opConnectCalendarFeed); err != nil {
		return nil, err
	}
	person, err := s.people.Get(ctx, in.PersonID)
	if err != nil {
		return nil, err
	}
	link := strings.TrimSpace(in.URL)
	host, err := calendar.CheckURL(link, s.cfg.CalendarAllowedHosts)
	if err != nil {
		return nil, feedError(err, opConnectCalendarFeed)
	}
	// Fetch before the transaction: a slow calendar must not hold database locks.
	leave, found, err := s.read(ctx, link, person)
	if err != nil {
		return nil, feedError(err, opConnectCalendarFeed)
	}
	sealed, err := s.sealer.Seal(link)
	if err != nil {
		return nil, errors.InternalError(constants.MsgCalendarFeedSaveFailed, err).
			WithOperation(opConnectCalendarFeed).WithResource(resourceCalendarFeed)
	}
	now := s.clock.Now().UTC()
	err = s.tx.Do(ctx, func(ctx context.Context) error {
		feed, err := s.feeds.FindByPerson(ctx, in.PersonID)
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			feed, err = &models.CalendarFeed{BaseModel: models.NewBaseModel(), PersonID: in.PersonID}, nil
		}
		if err != nil {
			return err
		}
		feed.URLCiphertext, feed.Host, feed.NextSyncAt = sealed, host, now.Add(s.cfg.CalendarSyncInterval)
		if err := s.feeds.Save(ctx, feed); err != nil {
			return err
		}
		if err := s.apply(ctx, person.ID, leave); err != nil {
			return err
		}
		return s.feeds.RecordSync(ctx, feed.ID, now, found, nil)
	})
	if err != nil {
		return nil, asAppError(err, constants.MsgCalendarFeedSaveFailed, opConnectCalendarFeed, resourceCalendarFeed)
	}
	return s.find(ctx, in.PersonID, opConnectCalendarFeed)
}

func (s *calendarFeedService) Disconnect(ctx context.Context, personID string) error {
	if err := s.check(ctx, personID, opDisconnectCalendarFeed); err != nil {
		return err
	}
	if _, err := s.find(ctx, personID, opDisconnectCalendarFeed); err != nil {
		return err
	}
	if err := s.feeds.DeleteByPerson(ctx, personID); err != nil {
		return errors.DatabaseError(constants.MsgCalendarFeedSaveFailed, err).
			WithOperation(opDisconnectCalendarFeed).WithResource(resourceCalendarFeed)
	}
	return nil
}

func (s *calendarFeedService) SyncNow(ctx context.Context, personID string) (*models.CalendarFeed, error) {
	if err := s.check(ctx, personID, opSyncCalendarFeed); err != nil {
		return nil, err
	}
	feed, err := s.find(ctx, personID, opSyncCalendarFeed)
	if err != nil {
		return nil, err
	}
	if err := s.sync(ctx, feed); err != nil {
		return nil, asAppError(err, constants.MsgCalendarFeedSaveFailed, opSyncCalendarFeed, resourceCalendarFeed)
	}
	return s.find(ctx, personID, opSyncCalendarFeed)
}

func (s *calendarFeedService) RunOnce(ctx context.Context) (int, error) {
	if s.sealer == nil {
		return 0, nil
	}
	now := s.clock.Now().UTC()
	var claimed []models.CalendarFeed
	err := s.tx.Do(ctx, func(ctx context.Context) error {
		feeds, err := s.feeds.ClaimDue(ctx, now, now.Add(s.cfg.CalendarSyncInterval), s.cfg.CalendarSyncBatchSize)
		claimed = feeds
		return err
	})
	if err != nil {
		return 0, err
	}
	var errs []error
	for i := range claimed {
		if err := s.sync(ctx, &claimed[i]); err != nil {
			errs = append(errs, err)
		}
	}
	return len(claimed), stderrors.Join(errs...)
}

func (s *calendarFeedService) Run(ctx context.Context) {
	if s.sealer == nil {
		return
	}
	ticker := time.NewTicker(s.cfg.CalendarSyncInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Errors never hold a calendar link, so they are safe to log.
			if n, err := s.RunOnce(ctx); err != nil {
				logger.Log.Warn("Calendar sync failed", zap.Int("calendars", n), zap.Error(err))
			}
		}
	}
}

// check answers 503 while the feature is off, then lets the person or a planner through.
func (s *calendarFeedService) check(ctx context.Context, personID, op string) error {
	if s.sealer == nil {
		return errors.NewAppError(constants.CalendarSyncDisabled, constants.MsgCalendarSyncDisabled, errors.ErrorTypeExternal,
			http.StatusServiceUnavailable).WithOperation(op).WithResource(resourceCalendarFeed)
	}
	_, err := s.authz.RequireSelfOr(ctx, personID, CanManageAbsences)
	return err
}

func (s *calendarFeedService) find(ctx context.Context, personID, op string) (*models.CalendarFeed, error) {
	feed, err := s.feeds.FindByPerson(ctx, personID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.WrapError(err, constants.CalendarFeedNotFound, constants.MsgCalendarFeedNotFound,
				errors.ErrorTypeNotFound, http.StatusNotFound).WithOperation(op).WithResource(resourceCalendarFeed)
		}
		return nil, errors.DatabaseError(constants.MsgDatabaseReadFailed, err).WithOperation(op).WithResource(resourceCalendarFeed)
	}
	return feed, nil
}

// sync reads one stored calendar and applies its leave. A calendar that cannot be read keeps
// its absences as they are and stores the error code; only database errors are returned.
func (s *calendarFeedService) sync(ctx context.Context, feed *models.CalendarFeed) error {
	person, err := s.people.Get(ctx, feed.PersonID)
	if err != nil {
		return err
	}
	now := s.clock.Now().UTC()
	link, err := s.sealer.Open(feed.URLCiphertext)
	if err != nil {
		return s.recordFailure(ctx, feed.ID, now, calendar.ErrUnreachable)
	}
	leave, found, err := s.read(ctx, link, person)
	if err != nil {
		return s.recordFailure(ctx, feed.ID, now, err)
	}
	return s.tx.Do(ctx, func(ctx context.Context) error {
		if err := s.apply(ctx, person.ID, leave); err != nil {
			return err
		}
		return s.feeds.RecordSync(ctx, feed.ID, now, found, nil)
	})
}

func (s *calendarFeedService) recordFailure(ctx context.Context, feedID string, at time.Time, cause error) error {
	code, _ := feedErrorCode(cause)
	return s.feeds.RecordSync(ctx, feedID, at, 0, &code)
}

// read fetches a calendar and returns its out-of-office events from today to
// CALENDAR_SYNC_WEEKS ahead as leave, and how many events it found.
func (s *calendarFeedService) read(ctx context.Context, link string, person *models.Person) ([]CalendarLeave, int, error) {
	data, err := s.fetcher.Fetch(ctx, link)
	if err != nil {
		return nil, 0, err
	}
	from, to := s.window()
	events, err := calendar.Parse(data, calendar.LeaveRules{Keywords: s.cfg.CalendarLeaveKeywords, From: from, To: to},
		s.clock.Location())
	if err != nil {
		return nil, 0, err
	}
	daily := dailyHours(person.WeeklyHours, s.cfg.WorkingWeekdays)
	leave := make([]CalendarLeave, 0, len(events))
	for _, e := range events {
		hours := daily
		if e.HoursPerDay != nil {
			hours = min(*e.HoursPerDay, daily)
		}
		if hours = math.Round(hours*hoursPerDayScale) / hoursPerDayScale; hours <= 0 {
			continue
		}
		leave = append(leave, CalendarLeave{UID: e.UID, FromDate: e.From, ToDate: e.To, HoursPerDay: hours})
	}
	return leave, len(events), nil
}

func (s *calendarFeedService) apply(ctx context.Context, personID string, leave []CalendarLeave) error {
	from, to := s.window()
	_, err := s.availability.ApplyCalendarLeave(ctx, ApplyCalendarLeaveInput{PersonID: personID, From: from, To: to, Leave: leave})
	return err
}

// window is today, in the organization's time zone, to CALENDAR_SYNC_WEEKS ahead, as UTC dates.
func (s *calendarFeedService) window() (time.Time, time.Time) {
	now := s.clock.Now().In(s.clock.Location())
	from := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return from, from.AddDate(0, 0, s.cfg.CalendarSyncWeeks*daysPerWeek)
}

// dailyHours is a person's working day: weekly hours spread over the working weekdays.
func dailyHours(weeklyHours float64, workingDays []time.Weekday) float64 {
	if len(workingDays) == 0 {
		return 0
	}
	return weeklyHours / float64(len(workingDays))
}

// feedError turns a fetch or parse failure into a 400 with its code. Its message never holds the link.
func feedError(err error, op string) error {
	code, msg := feedErrorCode(err)
	return errors.NewAppError(code, msg, errors.ErrorTypeValidation, http.StatusBadRequest).
		WithOperation(op).WithResource(resourceCalendarFeed)
}

func feedErrorCode(err error) (string, string) {
	switch {
	case stderrors.Is(err, calendar.ErrHostNotAllowed):
		return constants.CalendarFeedHostNotAllowed, constants.MsgCalendarFeedHostNotAllowed
	case stderrors.Is(err, calendar.ErrTooLarge):
		return constants.CalendarFeedTooLarge, constants.MsgCalendarFeedTooLarge
	case stderrors.Is(err, calendar.ErrNotICS):
		return constants.CalendarFeedNotICS, constants.MsgCalendarFeedNotICS
	default:
		return constants.CalendarFeedUnreachable, constants.MsgCalendarFeedUnreachable
	}
}
