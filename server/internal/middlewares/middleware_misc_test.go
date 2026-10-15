package middlewares

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/constants"
	"the-slot-finder/server/internal/dtos"
	"the-slot-finder/server/internal/request"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasicAuthMiddleware(t *testing.T) {
	cfg := config.Config{BasicAuthUsername: "admin", BasicAuthPassword: "secret"}
	e := echo.New()
	e.Use(BasicAuthMiddleware(cfg))
	e.GET("/docs", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	tests := []struct {
		name       string
		user       string
		pass       string
		wantStatus int
	}{
		{name: "valid", user: "admin", pass: "secret", wantStatus: http.StatusOK},
		{name: "invalid password", user: "admin", pass: "wrong", wantStatus: http.StatusUnauthorized},
		{name: "missing", wantStatus: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/docs", nil)
			if tt.user != "" {
				req.SetBasicAuth(tt.user, tt.pass)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestBasicAuthMiddleware_NoConfiguredCredentials(t *testing.T) {
	e := echo.New()
	e.Use(BasicAuthMiddleware(config.Config{}))
	e.GET("/docs", func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/docs", nil)
	req.SetBasicAuth("", "")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCORS(t *testing.T) {
	e := echo.New()
	e.Use(CORS())
	e.GET("/ping", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodOptions, "/ping", nil)
	req.Header.Set(echo.HeaderOrigin, "http://example.com")
	req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodGet)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, "*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	assert.Contains(t, rec.Header().Get(echo.HeaderAccessControlAllowHeaders), "X-CSRF-Token")
}

func TestCSRFAndExposeToken(t *testing.T) {
	tests := []struct {
		name   string
		env    config.Environment
		secure bool
	}{
		{name: "development", env: config.EnvironmentDevelopment, secure: false},
		{name: "production", env: config.EnvironmentProduction, secure: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{AppEnv: tt.env}
			e := echo.New()
			e.Use(CSRF(cfg))
			e.Use(ExposeCSRFToken())
			e.GET("/form", func(c *echo.Context) error {
				return c.NoContent(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/form", nil)
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.NotEmpty(t, rec.Header().Get("X-CSRF-Token"))

			cookies := rec.Result().Cookies()
			require.NotEmpty(t, cookies)
			var csrfCookie *http.Cookie
			for _, c := range cookies {
				if c.Name == "csrf_token" {
					csrfCookie = c
					break
				}
			}
			require.NotNil(t, csrfCookie)
			assert.Equal(t, tt.secure, csrfCookie.Secure)
		})
	}

	t.Run("expose without token skips header", func(t *testing.T) {
		e := echo.New()
		e.Use(ExposeCSRFToken())
		e.GET("/x", func(c *echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Empty(t, rec.Header().Get("X-CSRF-Token"))
	})

	t.Run("expose empty token skips header", func(t *testing.T) {
		e := echo.New()
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				c.Set(echomw.DefaultCSRFConfig.ContextKey, "")
				return next(c)
			}
		})
		e.Use(ExposeCSRFToken())
		e.GET("/x", func(c *echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Empty(t, rec.Header().Get("X-CSRF-Token"))
	})
}

func TestSecurity(t *testing.T) {
	e := echo.New()
	e.Use(Security())
	e.GET("/", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	// HSTS is only written for TLS / forwarded HTTPS requests.
	req.Header.Set(echo.HeaderXForwardedProto, "https")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, "1; mode=block", rec.Header().Get(echo.HeaderXXSSProtection))
	assert.Equal(t, "nosniff", rec.Header().Get(echo.HeaderXContentTypeOptions))
	assert.Equal(t, "DENY", rec.Header().Get(echo.HeaderXFrameOptions))
	assert.Contains(t, rec.Header().Get(echo.HeaderStrictTransportSecurity), "max-age=31536000")
	assert.Equal(t, "no-referrer", rec.Header().Get(echo.HeaderReferrerPolicy))
}

func TestRateLimit(t *testing.T) {
	newServer := func(cfg *config.Config) *echo.Echo {
		e := echo.New()
		e.IPExtractor = echo.ExtractIPDirect()
		e.Use(RateLimit(cfg))
		e.GET("/limited", func(c *echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})
		return e
	}
	call := func(e *echo.Echo, ip string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, "/limited", nil)
		req.RemoteAddr = ip + ":9999"
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return rec
	}

	t.Run("under the limit", func(t *testing.T) {
		e := newServer(&config.Config{RateLimit: 2, RateLimitDuration: time.Second})
		assert.Equal(t, http.StatusNoContent, call(e, "127.0.0.1").Code)
	})

	t.Run("deny when exceeded, in the envelope", func(t *testing.T) {
		e := newServer(&config.Config{RateLimit: 1, RateLimitDuration: time.Minute})
		var last *httptest.ResponseRecorder
		for i := 0; i < 3; i++ {
			last = call(e, "10.0.0.1")
		}
		require.Equal(t, http.StatusTooManyRequests, last.Code)

		var body dtos.BaseResponse[dtos.RateLimitResponse]
		require.NoError(t, json.Unmarshal(last.Body.Bytes(), &body))
		assert.Equal(t, constants.RateLimitExceeded, body.Meta.ErrorCode)
		assert.Equal(t, http.StatusTooManyRequests, body.Meta.Code)
		assert.Equal(t, dtos.RateLimitResponse{Limit: 1, Window: time.Minute.String(), RetryAfter: 60}, body.Data)
		assert.Equal(t, "60", last.Header().Get(echo.HeaderRetryAfter))

		// Another client is not affected.
		assert.Equal(t, http.StatusNoContent, call(e, "10.0.0.2").Code)
	})

	t.Run("disabled with a zero limit", func(t *testing.T) {
		e := newServer(&config.Config{RateLimit: 0, RateLimitDuration: time.Second})
		for i := 0; i < 5; i++ {
			assert.Equal(t, http.StatusNoContent, call(e, "10.0.0.3").Code)
		}
	})

	t.Run("a short window still asks for a whole second", func(t *testing.T) {
		e := newServer(&config.Config{RateLimit: 20, RateLimitDuration: time.Second})
		var last *httptest.ResponseRecorder
		for i := 0; i < 25; i++ {
			last = call(e, "10.0.0.4")
		}
		require.Equal(t, http.StatusTooManyRequests, last.Code)
		assert.Equal(t, "1", last.Header().Get(echo.HeaderRetryAfter))
	})
}

func TestRequestContext(t *testing.T) {
	tests := []struct {
		name             string
		headers          map[string]string
		wantLang         string
		wantCorrPrefix   string
		wantExistingCorr string
	}{
		{
			name:           "generates correlation id and default language",
			wantLang:       "en",
			wantCorrPrefix: "svc-",
		},
		{
			name: "uses provided correlation and language",
			headers: map[string]string{
				CorrelationIDHeaderKey: "corr-fixed",
				LanguageCodeHeaderKey:  "fr-FR,en;q=0.8",
			},
			wantLang:         "fr-FR",
			wantExistingCorr: "corr-fixed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			e.Use(RequestContext("svc"))
			var gotCorr, gotLang string
			e.GET("/ctx", func(c *echo.Context) error {
				gotCorr, _ = request.CorrelationIDFromContext(c.Request().Context())
				gotLang, _ = request.LanguageCodeFromContext(c.Request().Context())
				return c.NoContent(http.StatusNoContent)
			})

			req := httptest.NewRequest(http.MethodGet, "/ctx?x=1", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			rec := httptest.NewRecorder()
			e.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNoContent, rec.Code)
			assert.Equal(t, tt.wantLang, gotLang)
			if tt.wantExistingCorr != "" {
				assert.Equal(t, tt.wantExistingCorr, gotCorr)
				assert.Equal(t, tt.wantExistingCorr, rec.Header().Get(CorrelationIDHeaderKey))
			} else {
				assert.True(t, strings.HasPrefix(gotCorr, tt.wantCorrPrefix))
				assert.Equal(t, gotCorr, rec.Header().Get(CorrelationIDHeaderKey))
			}
		})
	}
}

func TestParseLanguageCode(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", parseLanguageCode(""))
	assert.Equal(t, "en", parseLanguageCode("en"))
	assert.Equal(t, "vi", parseLanguageCode("vi,en;q=0.9"))
}

func TestFormatRandomSuffix(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "005", formatRandomSuffix(5))
	assert.Equal(t, "042", formatRandomSuffix(42))
	assert.Equal(t, "999", formatRandomSuffix(999))
	assert.Equal(t, "007", formatRandomSuffix(-7))
	assert.Equal(t, "000", formatRandomSuffix(0))
}

func TestLogBody(t *testing.T) {
	cfg := &config.Config{LogBodyMaxBytes: 8, AppBodyLimitBytes: 16}

	capture := func(t *testing.T, contentType, body string) (any, string, int) {
		t.Helper()
		e := echo.New()
		var logged any
		var seen string
		e.Use(BodyLimit(cfg), LogBody(cfg))
		e.POST("/body", func(c *echo.Context) error {
			logged = c.Get(LogBodyKey)
			data, err := io.ReadAll(c.Request().Body)
			require.NoError(t, err)
			seen = string(data)
			return c.NoContent(http.StatusNoContent)
		})
		req := httptest.NewRequest(http.MethodPost, "/body", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, contentType)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		return logged, seen, rec.Code
	}

	t.Run("short body is logged whole", func(t *testing.T) {
		logged, seen, code := capture(t, echo.MIMEApplicationJSON, `{"a":1}`)
		assert.Equal(t, http.StatusNoContent, code)
		assert.Equal(t, `{"a":1}`, logged)
		assert.Equal(t, `{"a":1}`, seen)
	})

	t.Run("long body is truncated in the log only", func(t *testing.T) {
		logged, seen, code := capture(t, echo.MIMEApplicationJSON, `{"a":12345}`)
		assert.Equal(t, http.StatusNoContent, code)
		assert.Equal(t, `{"a":123`+truncatedSuffix, logged)
		assert.Equal(t, `{"a":12345}`, seen)
	})

	t.Run("multipart body is not read", func(t *testing.T) {
		logged, seen, code := capture(t, echo.MIMEMultipartForm+"; boundary=x", "--x--")
		assert.Equal(t, http.StatusNoContent, code)
		assert.Nil(t, logged)
		assert.Equal(t, "--x--", seen)
	})

	t.Run("a body that holds a secret is never kept", func(t *testing.T) {
		e := echo.New()
		var logged any
		var seen string
		e.Use(LogBody(&config.Config{LogBodyMaxBytes: 256}))
		e.PUT(constants.PathAPIBase+constants.PathPersonCalendarFeed, func(c *echo.Context) error {
			logged = c.Get(LogBodyKey)
			data, err := io.ReadAll(c.Request().Body)
			require.NoError(t, err)
			seen = string(data)
			return c.NoContent(http.StatusNoContent)
		})
		body := `{"url":"https://calendar.google.com/calendar/ical/a/private-x/basic.ics"}`
		req := httptest.NewRequest(http.MethodPut, constants.PathAPIBase+"/people/p1/calendar-feed", strings.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		e.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, redactedBody, logged)
		assert.Equal(t, body, seen)
	})

	t.Run("body over the limit gets 413", func(t *testing.T) {
		_, _, code := capture(t, echo.MIMEApplicationJSON, strings.Repeat("x", 32))
		assert.Equal(t, http.StatusRequestEntityTooLarge, code)
	})

	t.Run("read error", func(t *testing.T) {
		handler := LogBody(cfg)(func(c *echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})
		req := httptest.NewRequest(http.MethodPost, "/body", errReader{})
		c := echo.New().NewContext(req, httptest.NewRecorder())
		require.Error(t, handler(c))
	})
}

func TestTruncateBody(t *testing.T) {
	assert.Equal(t, "abc", truncateBody([]byte("abc"), 0))
	assert.Equal(t, "abc", truncateBody([]byte("abc"), 3))
	// "é" is two bytes; cutting inside it drops the partial character.
	assert.Equal(t, "a"+truncatedSuffix, truncateBody([]byte("aé"), 2))
}

type errReader struct{}

func (errReader) Read(p []byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestRequestLogging(t *testing.T) {
	cfg := &config.Config{
		AppEnv:     config.EnvironmentTest,
		AppName:    "slot-finder-test",
		AppVersion: "1.2.3",
	}

	e := echo.New()
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			c.Set("organization_id", "org-1")
			c.Set(LogBodyKey, `{"hello":"world"}`)
			ctx := request.NewPrincipalContext(c.Request().Context(), request.Principal{Subject: "user-42"})
			ctx = request.NewCorrelationIDContext(ctx, "corr-1")
			ctx = request.NewLanguageCodeContext(ctx, "en")
			c.SetRequest(c.Request().WithContext(ctx))
			return next(c)
		}
	})
	e.Use(RequestLogging(cfg))
	e.GET("/logged/:id", func(c *echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/logged/abc?q=1", nil)
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Custom", "value")
	req.Header.Set("Cookie", "secret=1")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Cover branch where body/claims/org are absent and auth header is not JWT-shaped.
	e2 := echo.New()
	e2.Use(RequestLogging(cfg))
	e2.GET("/plain", func(c *echo.Context) error {
		return c.NoContent(http.StatusNoContent)
	})
	req2 := httptest.NewRequest(http.MethodGet, "/plain", nil)
	req2.Header.Set("Authorization", "Bearer not-a-jwt")
	rec2 := httptest.NewRecorder()
	e2.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusNoContent, rec2.Code)
}

var _ io.Reader = errReader{}
