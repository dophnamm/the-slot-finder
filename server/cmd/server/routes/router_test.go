package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/handlers"
	"the-slot-finder/server/internal/logger"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	logger.Log = zap.NewNop()
	logger.Sugar = logger.Log.Sugar()
	os.Exit(m.Run())
}

func newTestRouter(cfg *config.Config) *echo.Echo {
	return Router(handlers.ProvideHealthHandler(cfg, nil), nil, cfg)
}

func baseConfig() *config.Config {
	return &config.Config{
		AppEnv:            config.EnvironmentProduction, // no Swagger route in tests
		AppName:           "slot-finder-test",
		AppBodyLimitBytes: 1 << 20,
		LogBodyMaxBytes:   256,
		RateLimit:         100,
		RateLimitDuration: time.Second,
		BasicAuthUsername: "ops",
		BasicAuthPassword: "secret",
	}
}

func get(e *echo.Echo, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.RemoteAddr = "192.0.2.10:5000"
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	return rec
}

func TestRouter_UnknownPathIsNotFound(t *testing.T) {
	e := newTestRouter(baseConfig())
	rec := get(e, constants.PathAPIBase+"/does-not-exist")
	assert.Equal(t, http.StatusNotFound, rec.Code, rec.Body.String())
}

func TestRouter_MetricsNeedBasicAuth(t *testing.T) {
	e := newTestRouter(baseConfig())
	rec := get(e, constants.PathAPIBase+constants.PathHealthMetrics)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRouter_RateLimit(t *testing.T) {
	cfg := baseConfig()
	cfg.RateLimit = 2
	cfg.RateLimitDuration = time.Minute
	e := newTestRouter(cfg)

	var last *httptest.ResponseRecorder
	for i := 0; i < cfg.RateLimit+1; i++ {
		last = get(e, constants.PathAPIBase+constants.PathHealth)
	}
	require.Equal(t, http.StatusTooManyRequests, last.Code)
	var body dtos.BaseResponse[dtos.RateLimitResponse]
	require.NoError(t, json.Unmarshal(last.Body.Bytes(), &body))
	assert.Equal(t, constants.RateLimitExceeded, body.Meta.ErrorCode)
}
