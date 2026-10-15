package config

import (
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Environment represents the application environment.
type Environment string

// Keycloak URL paths, below the base URL.
const (
	keycloakRealmsPath = "/realms/"
	keycloakCertsPath  = "/protocol/openid-connect/certs"
)

const (
	EnvironmentDevelopment Environment = "development"
	EnvironmentStaging     Environment = "staging"
	EnvironmentProduction  Environment = "production"
	EnvironmentTest        Environment = "test"
)

func (e Environment) String() string {
	return string(e)
}

func (e Environment) IsDevelopment() bool {
	return e == EnvironmentDevelopment || e == EnvironmentTest
}

func (e Environment) IsProduction() bool {
	return e == EnvironmentProduction
}

// Config holds all configuration for the application
type Config struct {
	// Server configuration
	AppEnv        Environment
	AppName       string
	AppVersion    string
	Timezone      string
	AppHTTPServer string
	// AppRequestTimeout is the HTTP server read header timeout in seconds.
	AppRequestTimeout int
	// Largest request body accepted, in bytes; larger requests get 413.
	AppBodyLimitBytes int64
	// Trust X-Forwarded-For from private and loopback proxies when extracting the client IP.
	HTTPTrustProxy bool

	// Database configuration
	DatabaseHost        string
	DatabasePort        string
	DatabaseUsername    string
	DatabasePassword    string
	DatabaseName        string
	DatabaseEnableDebug bool

	// Database connection management
	DatabaseMaxOpenConns    int
	DatabaseMaxIdleConns    int
	DatabaseConnMaxLifetime time.Duration
	DatabaseConnMaxIdleTime time.Duration
	DatabaseConnectTimeout  time.Duration
	DatabaseQueryTimeout    time.Duration
	DatabaseHealthTimeout   time.Duration
	DatabaseRetryAttempts   int
	DatabaseRetryDelay      time.Duration
	DatabaseSSLMode         string
	DatabaseTimezone        string

	// Cache configuration
	CacheProvider   string
	RedisHost       string
	RedisPort       string
	RedisPassword   string
	RedisDB         int
	PoolSize        int
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	PoolTimeout     time.Duration
	MaxRetries      int
	MinRetryBackoff time.Duration
	MaxRetryBackoff time.Duration

	// Logging configuration
	LogLevel string
	// Longest request body written to the request log, in bytes; the rest is cut.
	LogBodyMaxBytes int

	// Authentication configuration
	AuthProvider        string
	KeycloakURL         string
	KeycloakRealm       string
	KeycloakClientID    string
	KeycloakSecret      string
	KeycloakRedirectURI string
	// Keycloak base URL as the browser sees it; tokens carry it in their issuer.
	KeycloakPublicURL string
	// Public client of the web app; tokens it obtains carry it in azp.
	KeycloakSPAClientID string

	// Local JWT verification against the realm's JWKS.
	AuthJWKSRefreshInterval    time.Duration
	AuthJWKSUnknownKIDInterval time.Duration
	AuthJWKSRefreshWaitMax     time.Duration
	AuthClockSkew              time.Duration

	// Email configuration
	EmailProvider   string
	AWSSESRegion    string
	AWSSESAccessKey string
	AWSSESSecretKey string

	// Rate limiting configuration
	RateLimit              int
	RateLimitDuration      time.Duration
	RateLimitHeavy         int
	RateLimitHeavyDuration time.Duration

	// Sentry configuration
	SentryDSN string

	// OpenTelemetry configuration
	OTelServiceName      string
	OTelExporterEndpoint string
	OTelExporterProtocol string
	OTelExporterInsecure bool
	OTelTracesEnabled    bool
	OTelMetricsEnabled   bool
	OTelLogsEnabled      bool

	// Basic Auth configuration
	BasicAuthUsername string
	BasicAuthPassword string

	// HTTP Client configuration
	HTTPClientTimeout            time.Duration
	HTTPClientRetryCount         int
	HTTPClientRetryWaitMin       time.Duration
	HTTPClientRetryWaitMax       time.Duration
	HTTPClientDebug              bool
	HTTPClientTLSInsecureSkipTLS bool

	// Capacity: days that count, hours kept for non-project work, and whether planned
	// (not yet approved) absence lowers capacity.
	WorkingWeekdays        []time.Weekday
	DefaultNonProjectHours float64
	CountPlannedAbsence    bool

	// Planning policy, copied into each engine request (server/README §8).
	PlanningTargetUtilization          float64
	PlanningMaxUtilization             float64
	PlanningMaxExpertisesPerPersonWeek int
	PlanningMinHoursPerWeek            float64
	PlanningHourStep                   float64
	PlanningWeightDeadline             float64
	PlanningWeightPrimaryFit           float64
	PlanningWeightBalance              float64
	PlanningWeightStability            float64
	PlanningWeightContinuity           float64

	// Heatmap range and utilization bands (server/README §8). Overloaded is always over 100 %.
	HeatmapDefaultWeeks       int
	HeatmapMaxWeeks           int
	UtilizationUnderusedBelow float64
	UtilizationNearFullFrom   float64
	// ScheduleMaxWeeks is the widest range of one person's schedule (a month grid needs 6).
	ScheduleMaxWeeks int

	// Re-plan runner (server/README §6): debounce, polling, batch size, retries, and the raised
	// stability weight that keeps a re-plan close to the baseline.
	ReplanDebounce                time.Duration
	ReplanPollInterval            time.Duration
	ReplanBatchSize               int
	ReplanMaxAttempts             int
	ReplanRetryBackoff            time.Duration
	PlanningReplanStabilityWeight float64
	ScenarioMaxShiftWeeks         int
	ScenarioMaxAssumptions        int

	// Timesheet import (docs/features/timesheet-import.md): the template's default and largest
	// range in weeks, the most actual hours one row may hold, and the largest file.
	TimesheetTemplateWeeks int
	TimesheetMaxWeeks      int
	TimesheetMaxWeekHours  float64
	TimesheetMaxBytes      int64
	TimesheetMaxRows       int

	// Calendar sync (docs/features/calendar-sync.md): the key that encrypts calendar links (empty
	// turns the feature off), where a link may point, how often and how far ahead the runner
	// reads, which titles mean out of office, and the limits on one fetch.
	CalendarFeedKey       string
	CalendarAllowedHosts  []string
	CalendarSyncInterval  time.Duration
	CalendarSyncBatchSize int
	CalendarSyncWeeks     int
	CalendarLeaveKeywords []string
	CalendarFetchTimeout  time.Duration
	CalendarMaxBytes      int64

	// Program metrics (root README §7): how far Scope % may trail Time %, in percentage
	// points, and how close a deadline turns a gap into at risk.
	ScopeLagThreshold float64
	AlertHorizonWeeks int
	// SustainedOverloadWeeks is how many weeks in a row over 100 % make a sustained overload.
	SustainedOverloadWeeks int
	// DraftAlternativesLimit is how many people a staffing draft offers to swap in for an assignment.
	DraftAlternativesLimit int
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (optional; ignore errors)
	_ = godotenv.Load()

	cfg := &Config{
		AppEnv:                       Environment(getEnv("APP_ENV", "development")),
		AppName:                      getEnv("APP_NAME", ""),
		AppVersion:                   getEnv("APP_VERSION", "1.0.0"),
		Timezone:                     getEnv("TIMEZONE", "UTC"),
		AppHTTPServer:                getEnv("APP_HTTP_SERVER", ":3000"),
		AppRequestTimeout:            getEnvAsInt("APP_REQUEST_TIMEOUT", 30),
		AppBodyLimitBytes:            int64(getEnvAsInt("APP_BODY_LIMIT_BYTES", 2<<20)),
		HTTPTrustProxy:               getEnvAsBool("HTTP_TRUST_PROXY", false),
		DatabaseHost:                 getEnv("POSTGRES_HOST", "localhost"),
		DatabasePort:                 getEnv("POSTGRES_PORT", "5432"),
		DatabaseUsername:             getEnv("POSTGRES_USER", "postgres"),
		DatabasePassword:             getEnv("POSTGRES_PASSWORD", ""),
		DatabaseName:                 getEnv("POSTGRES_DB", ""),
		DatabaseEnableDebug:          getEnvAsBool("DATABASE_DEBUG", false),
		DatabaseMaxOpenConns:         getEnvAsInt("DATABASE_MAX_OPEN_CONNS", 25),
		DatabaseMaxIdleConns:         getEnvAsInt("DATABASE_MAX_IDLE_CONNS", 5),
		DatabaseConnMaxLifetime:      getEnvAsDuration("DATABASE_CONN_MAX_LIFETIME", 5*time.Minute),
		DatabaseConnMaxIdleTime:      getEnvAsDuration("DATABASE_CONN_MAX_IDLE_TIME", 1*time.Minute),
		DatabaseConnectTimeout:       getEnvAsDuration("DATABASE_CONNECT_TIMEOUT", 30*time.Second),
		DatabaseQueryTimeout:         getEnvAsDuration("DATABASE_QUERY_TIMEOUT", 30*time.Second),
		DatabaseHealthTimeout:        getEnvAsDuration("DATABASE_HEALTH_TIMEOUT", 5*time.Second),
		DatabaseRetryAttempts:        getEnvAsInt("DATABASE_RETRY_ATTEMPTS", 3),
		DatabaseRetryDelay:           getEnvAsDuration("DATABASE_RETRY_DELAY", 1*time.Second),
		DatabaseSSLMode:              getEnv("DATABASE_SSL_MODE", "disable"),
		DatabaseTimezone:             getEnv("DATABASE_TIMEZONE", "UTC"),
		CacheProvider:                getEnv("CACHE_PROVIDER", "redis"),
		RedisHost:                    getEnv("REDIS_HOST", "localhost"),
		RedisPort:                    getEnv("REDIS_PORT", "6379"),
		RedisPassword:                getEnv("REDIS_PASSWORD", ""),
		RedisDB:                      getEnvAsInt("REDIS_DB", 0),
		PoolSize:                     getEnvAsInt("REDIS_POOL_SIZE", 10),
		DialTimeout:                  getEnvAsDuration("REDIS_DIAL_TIMEOUT", 5*time.Second),
		ReadTimeout:                  getEnvAsDuration("REDIS_READ_TIMEOUT", 5*time.Second),
		WriteTimeout:                 getEnvAsDuration("REDIS_WRITE_TIMEOUT", 5*time.Second),
		PoolTimeout:                  getEnvAsDuration("REDIS_POOL_TIMEOUT", 5*time.Second),
		MaxRetries:                   getEnvAsInt("REDIS_MAX_RETRIES", 3),
		MinRetryBackoff:              getEnvAsDuration("REDIS_MIN_RETRY_BACKOFF", 1*time.Second),
		MaxRetryBackoff:              getEnvAsDuration("REDIS_MAX_RETRY_BACKOFF", 5*time.Second),
		LogLevel:                     getEnv("LOG_LEVEL", "info"),
		LogBodyMaxBytes:              getEnvAsInt("LOG_BODY_MAX_BYTES", 2048),
		AuthProvider:                 getEnv("AUTH_PROVIDER", "keycloak"),
		KeycloakURL:                  getEnv("KEYCLOAK_URL", ""),
		KeycloakRealm:                getEnv("KEYCLOAK_REALM", ""),
		KeycloakClientID:             getEnv("KEYCLOAK_CLIENT_ID", ""),
		KeycloakSecret:               getEnv("KEYCLOAK_CLIENT_SECRET", ""),
		KeycloakRedirectURI:          getEnv("KEYCLOAK_REDIRECT_URI", ""),
		KeycloakPublicURL:            getEnv("KEYCLOAK_PUBLIC_URL", getEnv("KEYCLOAK_URL", "")),
		KeycloakSPAClientID:          getEnv("KEYCLOAK_SPA_CLIENT_ID", ""),
		AuthJWKSRefreshInterval:      getEnvAsDuration("AUTH_JWKS_REFRESH_INTERVAL", time.Hour),
		AuthJWKSUnknownKIDInterval:   getEnvAsDuration("AUTH_JWKS_UNKNOWN_KID_INTERVAL", time.Minute),
		AuthJWKSRefreshWaitMax:       getEnvAsDuration("AUTH_JWKS_REFRESH_WAIT_MAX", time.Second),
		AuthClockSkew:                getEnvAsDuration("AUTH_CLOCK_SKEW", 30*time.Second),
		EmailProvider:                getEnv("EMAIL_PROVIDER", "ses"),
		AWSSESRegion:                 getEnv("AWS_SES_REGION", ""),
		AWSSESAccessKey:              getEnv("AWS_SES_ACCESS_KEY", ""),
		AWSSESSecretKey:              getEnv("AWS_SES_SECRET_KEY", ""),
		RateLimit:                    getEnvAsInt("RATE_LIMIT", 20),
		RateLimitDuration:            getEnvAsDuration("RATE_LIMIT_DURATION", 1*time.Second),
		RateLimitHeavy:               getEnvAsInt("RATE_LIMIT_HEAVY", 20),
		RateLimitHeavyDuration:       getEnvAsDuration("RATE_LIMIT_HEAVY_DURATION", 30*time.Second),
		SentryDSN:                    getEnv("SENTRY_DSN", ""),
		OTelServiceName:              getEnv("OTEL_SERVICE_NAME", getEnv("APP_NAME", "slot-finder")),
		OTelExporterEndpoint:         getEnv("OTEL_EXPORTER_OTLP_ENDPOINT", ""),
		OTelExporterProtocol:         getEnv("OTEL_EXPORTER_OTLP_PROTOCOL", "grpc"),
		OTelExporterInsecure:         getEnvAsBool("OTEL_EXPORTER_OTLP_INSECURE", false),
		OTelTracesEnabled:            getEnvAsBool("OTEL_TRACES_ENABLED", true),
		OTelMetricsEnabled:           getEnvAsBool("OTEL_METRICS_ENABLED", true),
		OTelLogsEnabled:              getEnvAsBool("OTEL_LOGS_ENABLED", true),
		BasicAuthUsername:            getEnv("BASIC_AUTH_USER", ""),
		BasicAuthPassword:            getEnv("BASIC_AUTH_SECRET", ""),
		HTTPClientTimeout:            getEnvAsDuration("HTTP_CLIENT_TIMEOUT", 30*time.Second),
		HTTPClientRetryCount:         getEnvAsInt("HTTP_CLIENT_RETRY_COUNT", 2),
		HTTPClientRetryWaitMin:       getEnvAsDuration("HTTP_CLIENT_RETRY_WAIT_MIN", 250*time.Millisecond),
		HTTPClientRetryWaitMax:       getEnvAsDuration("HTTP_CLIENT_RETRY_WAIT_MAX", 2*time.Second),
		HTTPClientDebug:              getEnvAsBool("HTTP_CLIENT_DEBUG", false),
		HTTPClientTLSInsecureSkipTLS: getEnvAsBool("HTTP_CLIENT_TLS_INSECURE_SKIP_TLS", false),
		DefaultNonProjectHours:       getEnvAsFloat("DEFAULT_NON_PROJECT_HOURS", 4),
		CountPlannedAbsence:          getEnvAsBool("COUNT_PLANNED_ABSENCE", true),

		PlanningTargetUtilization:          getEnvAsFloat("PLANNING_TARGET_UTILIZATION", 0.8),
		PlanningMaxUtilization:             getEnvAsFloat("PLANNING_MAX_UTILIZATION", 1.0),
		PlanningMaxExpertisesPerPersonWeek: getEnvAsInt("PLANNING_MAX_EXPERTISES_PER_PERSON_WEEK", 2),
		PlanningMinHoursPerWeek:            getEnvAsFloat("PLANNING_MIN_HOURS_PER_WEEK", 8),
		PlanningHourStep:                   getEnvAsFloat("PLANNING_HOUR_STEP", 1),
		PlanningWeightDeadline:             getEnvAsFloat("PLANNING_WEIGHT_DEADLINE", 1.0),
		PlanningWeightPrimaryFit:           getEnvAsFloat("PLANNING_WEIGHT_PRIMARY_FIT", 0.8),
		PlanningWeightBalance:              getEnvAsFloat("PLANNING_WEIGHT_BALANCE", 0.6),
		PlanningWeightStability:            getEnvAsFloat("PLANNING_WEIGHT_STABILITY", 0.5),
		PlanningWeightContinuity:           getEnvAsFloat("PLANNING_WEIGHT_CONTINUITY", 0.3),

		HeatmapDefaultWeeks:       getEnvAsInt("HEATMAP_DEFAULT_WEEKS", 12),
		HeatmapMaxWeeks:           getEnvAsInt("HEATMAP_MAX_WEEKS", 52),
		UtilizationUnderusedBelow: getEnvAsFloat("UTILIZATION_UNDERUSED_BELOW", 0.5),
		UtilizationNearFullFrom:   getEnvAsFloat("UTILIZATION_NEAR_FULL_FROM", 0.85),
		ScheduleMaxWeeks:          getEnvAsInt("SCHEDULE_MAX_WEEKS", 8),

		ReplanDebounce:                getEnvAsDuration("REPLAN_DEBOUNCE", 5*time.Minute),
		ReplanPollInterval:            getEnvAsDuration("REPLAN_POLL_INTERVAL", 10*time.Second),
		ReplanBatchSize:               getEnvAsInt("REPLAN_BATCH_SIZE", 50),
		ReplanMaxAttempts:             getEnvAsInt("REPLAN_MAX_ATTEMPTS", 5),
		ReplanRetryBackoff:            getEnvAsDuration("REPLAN_RETRY_BACKOFF", 30*time.Second),
		PlanningReplanStabilityWeight: getEnvAsFloat("PLANNING_REPLAN_STABILITY_WEIGHT", 2.0),
		ScenarioMaxShiftWeeks:         getEnvAsInt("SCENARIO_MAX_SHIFT_WEEKS", 12),
		ScenarioMaxAssumptions:        getEnvAsInt("SCENARIO_MAX_ASSUMPTIONS", 10),

		TimesheetTemplateWeeks: getEnvAsInt("TIMESHEET_TEMPLATE_WEEKS", 4),
		TimesheetMaxWeeks:      getEnvAsInt("TIMESHEET_MAX_WEEKS", 12),
		TimesheetMaxWeekHours:  getEnvAsFloat("TIMESHEET_MAX_WEEK_HOURS", 80),
		TimesheetMaxBytes:      int64(getEnvAsInt("TIMESHEET_MAX_BYTES", 1<<20)),
		TimesheetMaxRows:       getEnvAsInt("TIMESHEET_MAX_ROWS", 5000),

		CalendarFeedKey:       getEnv("CALENDAR_FEED_KEY", ""),
		CalendarAllowedHosts:  getEnvAsList("CALENDAR_ALLOWED_HOSTS", defaultCalendarAllowedHosts),
		CalendarSyncInterval:  getEnvAsDuration("CALENDAR_SYNC_INTERVAL", 15*time.Minute),
		CalendarSyncBatchSize: getEnvAsInt("CALENDAR_SYNC_BATCH_SIZE", 20),
		CalendarSyncWeeks:     getEnvAsInt("CALENDAR_SYNC_WEEKS", 26),
		CalendarLeaveKeywords: getEnvAsList("CALENDAR_LEAVE_KEYWORDS", defaultCalendarLeaveKeywords),
		CalendarFetchTimeout:  getEnvAsDuration("CALENDAR_FETCH_TIMEOUT", 10*time.Second),
		CalendarMaxBytes:      getEnvAsInt64("CALENDAR_MAX_BYTES", 5<<20),

		ScopeLagThreshold: getEnvAsFloat("SCOPE_LAG_THRESHOLD", 20),
		AlertHorizonWeeks: getEnvAsInt("ALERT_HORIZON_WEEKS", 4),

		SustainedOverloadWeeks: getEnvAsInt("SUSTAINED_OVERLOAD_WEEKS", 3),
		DraftAlternativesLimit: getEnvAsInt("DRAFT_ALTERNATIVES_LIMIT", 5),
	}

	weekdays, err := parseWeekdays(getEnv("WORKING_WEEKDAYS", defaultWorkingWeekdays))
	if err != nil {
		return nil, fmt.Errorf("WORKING_WEEKDAYS: %w", err)
	}
	cfg.WorkingWeekdays = weekdays

	return cfg, nil
}

// defaultCalendarAllowedHosts are the hosts that serve Google and Outlook calendar links.
const defaultCalendarAllowedHosts = "calendar.google.com,outlook.office365.com,outlook.live.com"

// defaultCalendarLeaveKeywords are title words that mark an event as out of office.
const defaultCalendarLeaveKeywords = "OOO,Out of office,Vacation,Annual leave,Nghỉ phép"

// defaultWorkingWeekdays is Monday to Friday, in the WORKING_WEEKDAYS format.
const defaultWorkingWeekdays = "Mon,Tue,Wed,Thu,Fri"

// weekdayNames maps the three-letter names used in WORKING_WEEKDAYS.
var weekdayNames = map[string]time.Weekday{
	"mon": time.Monday, "tue": time.Tuesday, "wed": time.Wednesday, "thu": time.Thursday,
	"fri": time.Friday, "sat": time.Saturday, "sun": time.Sunday,
}

// parseWeekdays reads a comma-separated list such as "Mon,Tue,Wed,Thu,Fri".
func parseWeekdays(value string) ([]time.Weekday, error) {
	var days []time.Weekday
	for _, name := range strings.Split(value, ",") {
		day, ok := weekdayNames[strings.ToLower(strings.TrimSpace(name))]
		if !ok {
			return nil, fmt.Errorf("unknown weekday %q", name)
		}
		if !slices.Contains(days, day) {
			days = append(days, day)
		}
	}
	return days, nil
}

// getEnv gets an environment variable with a fallback value
// envValue reads an environment variable without the quotes around it. The Makefile exports
// cmd/server/.env as it is, so a line like TIMEZONE="UTC" reaches the process with its quotes;
// godotenv strips them only for variables it sets itself.
func envValue(key string) string {
	value := os.Getenv(key)
	if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
		return value[1 : len(value)-1]
	}
	return value
}

func getEnv(key, fallback string) string {
	if value := envValue(key); value != "" {
		return value
	}
	return fallback
}

// getEnvAsInt gets an environment variable as integer with a fallback value
func getEnvAsInt(key string, fallback int) int {
	if value := envValue(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return fallback
}

// getEnvAsInt64 gets an environment variable as a 64-bit integer with a fallback value.
func getEnvAsInt64(key string, fallback int64) int64 {
	if value := envValue(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
			return intValue
		}
	}
	return fallback
}

// getEnvAsList reads a comma-separated list, trimming each item and dropping empty ones.
func getEnvAsList(key, fallback string) []string {
	var items []string
	for _, item := range strings.Split(getEnv(key, fallback), ",") {
		if item = strings.TrimSpace(item); item != "" {
			items = append(items, item)
		}
	}
	return items
}

// getEnvAsBool gets an environment variable as boolean with a fallback value
func getEnvAsFloat(key string, fallback float64) float64 {
	if value := envValue(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return fallback
}

func getEnvAsBool(key string, fallback bool) bool {
	if value := envValue(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return fallback
}

// getEnvAsDuration gets an environment variable as duration with a fallback value
func getEnvAsDuration(key string, fallback time.Duration) time.Duration {
	if value := envValue(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return fallback
}

func (c *Config) ConnectionString() string {
	return fmt.Sprintf(
		"host=%v port=%v user=%v password=%v dbname=%v sslmode=%v timezone=%v connect_timeout=%d",
		c.DatabaseHost,
		c.DatabasePort,
		c.DatabaseUsername,
		c.DatabasePassword,
		c.DatabaseName,
		c.DatabaseSSLMode,
		c.DatabaseTimezone,
		int(c.DatabaseConnectTimeout.Seconds()),
	)
}

// KeycloakIssuer is the issuer that access tokens of the realm carry.
func (c *Config) KeycloakIssuer() string {
	return strings.TrimRight(c.KeycloakPublicURL, "/") + keycloakRealmsPath + c.KeycloakRealm
}

// KeycloakJWKSURL is where the server fetches the realm's signing keys.
func (c *Config) KeycloakJWKSURL() string {
	return strings.TrimRight(c.KeycloakURL, "/") + keycloakRealmsPath + c.KeycloakRealm + keycloakCertsPath
}

func (c *Config) IsDebugMode() bool {
	return c.DatabaseEnableDebug
}

// PopulateFromJSON reads a service account JSON and fills email and private key
func (c *Config) PopulateFromJSON(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return c.PopulateFromJSONBytes(data)
}

// PopulateFromJSONBytes parses service account JSON bytes (same validation as PopulateFromJSON).
func (c *Config) PopulateFromJSONBytes(data []byte) error {
	var payload struct {
		ClientEmail string `json:"client_email"`
		PrivateKey  string `json:"private_key"`
		ProjectID   string `json:"project_id"`
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		return err
	}

	return nil
}
