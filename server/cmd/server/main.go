package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"the-slot-finder/server/docs"
	"the-slot-finder/server/internal/cache"
	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/handlers"
	"the-slot-finder/server/internal/httpclient"
	"the-slot-finder/server/internal/integration/auth"
	"the-slot-finder/server/internal/integration/email"
	"the-slot-finder/server/internal/logger"
	"the-slot-finder/server/internal/monitoring"
	"the-slot-finder/server/internal/repositories"
	"the-slot-finder/server/internal/services"
	"the-slot-finder/server/internal/utils"

	"the-slot-finder/server/cmd/server/routes"

	"the-slot-finder/server/internal/db"

	"github.com/go-playground/validator/v10"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func NewHTTPServer(lc fx.Lifecycle,
	healthHandler *handlers.HealthHandler,
	identityHandler *handlers.IdentityHandler,
	personHandler *handlers.PersonHandler,
	absenceHandler *handlers.AbsenceHandler,
	programHandler *handlers.ProgramHandler,
	projectHandler *handlers.ProjectHandler,
	planHandler *handlers.PlanHandler,
	staffingHandler *handlers.StaffingHandler,
	insightsHandler *handlers.InsightsHandler,
	eventHandler *handlers.EventHandler,
	verifier auth.TokenVerifier,
	cfg *config.Config,
	db *db.PostgresDB,
) *http.Server {
	handler := routes.Router(healthHandler, identityHandler, personHandler, absenceHandler, programHandler, projectHandler, planHandler, staffingHandler, insightsHandler, eventHandler, verifier, cfg)

	srv := &http.Server{
		Addr:              cfg.AppHTTPServer,
		Handler:           handler,
		ReadHeaderTimeout: time.Duration(cfg.AppRequestTimeout) * time.Second,
	}

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			ln, err := net.Listen("tcp", srv.Addr)
			if err != nil {
				return err
			}
			logger.Sugar.Infof("Starting HTTP server at %s", srv.Addr)
			go func() {
				err := srv.Serve(ln)
				if err != nil && !errors.Is(err, http.ErrServerClosed) {
					logger.Sugar.Panicf("HTTP server error: %v", err)
				}
			}()

			return nil
		},
		OnStop: func(ctx context.Context) error {
			logger.Sugar.Info("Shutting down HTTP server...")

			// Graceful shutdown with timeout
			shutdownCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
			defer cancel()

			// Shutdown HTTP server
			if err := srv.Shutdown(shutdownCtx); err != nil {
				logger.Sugar.Errorf("HTTP server shutdown error: %v", err)
				return err
			}

			// Close database connections
			if err := db.Close(); err != nil {
				logger.Sugar.Errorf("Database shutdown error: %v", err)
				return err
			}

			logger.Sugar.Info("Server shutdown completed")
			return nil
		},
	})

	return srv
}

// @title The Slot Finder API
// @version 1.0
// @description Staffing and capacity planning API for The Slot Finder
// @BasePath /api/v1
// @schemes http https
// @securityDefinitions.basic  BasicAuth
// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization
// @description Bearer Token Authentication. Use "Bearer {token}" as the value.
func main() {
	// Ensure Swagger spec is registered and optionally override fields at runtime
	docs.SwaggerInfo.BasePath = "/api/v1"
	cfg, err := config.Load()
	if err != nil {
		// Use standard log for fatal errors before logger is initialized
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}
	// Initialize global logger before any middleware uses it
	logger.Init(cfg.LogLevel, cfg.AppEnv.String())
	monitoring.InitSentry(*cfg)

	otelProvider, err := monitoring.InitOpenTelemetry(*cfg)
	if err != nil {
		logger.Sugar.Fatalf("Failed to initialize OpenTelemetry: %v", err)
	}
	defer func() {
		if otelProvider == nil {
			return
		}
		if err := otelProvider.Shutdown(context.Background()); err != nil {
			logger.Sugar.Errorf("OpenTelemetry shutdown error: %v", err)
		}
	}()

	if cfg.OTelLogsEnabled {
		monitoring.AttachOTelZapLogger(otelProvider.LoggerProvider(), []zapcore.Level{
			zapcore.ErrorLevel,
			zapcore.FatalLevel,
			zapcore.PanicLevel,
		})
	}

	// Add Sentry core to zap logger
	if logger.Log != nil {
		logger.Log = logger.Log.WithOptions(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewTee(core, monitoring.NewSentryCore(context.Background(), []zapcore.Level{
				zapcore.ErrorLevel,
				zapcore.FatalLevel,
				zapcore.PanicLevel,
			}))
		}))
		logger.Sugar = logger.Log.Sugar()
	}

	// Ensure all events are flushed before the program exits
	defer monitoring.FlushSentry()

	// The organization time zone lives in the Clock (TIMEZONE); time.Local is left alone.
	fx.New(appOptions(cfg)...).Run()
}

// appOptions lists every FX provider, so a test can validate the dependency graph.
func appOptions(cfg *config.Config) []fx.Option {
	return []fx.Option{
		fx.Supply(cfg),
		fx.Provide(
			NewHTTPServer,
			ProvideGormPostgres,
			db.ProvideTxManager,
			ProvideValidator,
			clock.ProvideClock,
			httpclient.ProvideRestClient,
			auth.ProvideAuth,
			auth.ProvideTokenVerifier,
			cache.ProvideCache,
			email.ProvideEmailSender,
			services.ProvideEmailService,
			services.ProvideAuthService,
			services.ProvideAuthzService,
			repositories.ProvidePersonRepository,
			repositories.ProvideAbsenceRepository,
			repositories.ProvideProgramRepository,
			repositories.ProvideProjectRepository,
			repositories.ProvidePlanRepository,
			repositories.ProvideAuditRepository,
			repositories.ProvideChangeEventRepository,
			services.ProvidePeopleService,
			services.ProvidePersonFinder,
			services.ProvideAvailabilityService,
			services.ProvideProgramService,
			services.ProvideProjectService,
			services.ProvidePlanService,
			services.ProvideStaffingService,
			services.ProvideInsightsService,
			services.ProvideEventService,
			services.ProvideEventRecorder,
			handlers.ProvideHealthHandler,
			handlers.ProvideIdentityHandler,
			handlers.ProvidePersonHandler,
			handlers.ProvideAbsenceHandler,
			handlers.ProvideProgramHandler,
			handlers.ProvideProjectHandler,
			handlers.ProvidePlanHandler,
			handlers.ProvideStaffingHandler,
			handlers.ProvideInsightsHandler,
			handlers.ProvideEventHandler,
		),
		fx.Invoke(func(*http.Server) {}),
		fx.Invoke(registerBaselineHook),
	}
}

func ProvideValidator() (*validator.Validate, error) {
	v := validator.New()
	if err := utils.RegisterValidations(v); err != nil {
		return nil, err
	}
	return v, nil
}

func ProvideGormPostgres(cfg *config.Config) *db.PostgresDB {
	appDB := &db.PostgresDB{}
	err := appDB.NewPostgresDB(cfg)
	if err != nil {
		logger.Sugar.Fatalf("Connecting to Database: %v", err)
	}
	return appDB
}

// registerBaselineHook creates the empty published baseline at startup when none exists,
// so the first accepted staffing draft follows the same path as every later one.
func registerBaselineHook(lc fx.Lifecycle, plans services.PlanService) {
	lc.Append(fx.Hook{OnStart: plans.EnsureBaseline})
}
