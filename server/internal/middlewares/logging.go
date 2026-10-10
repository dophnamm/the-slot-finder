package middlewares

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"strings"
	"time"

	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/request"

	"the-slot-finder/server/internal/logger"

	"github.com/getsentry/sentry-go"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/zap"
)

// RequestLogging provides structured logs to Sentry and zap
// and mirrors the rich request logging previously configured in the router.
func RequestLogging(cfg *config.Config) echo.MiddlewareFunc {
	return middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:       true,
		LogStatus:    true,
		LogLatency:   true,
		LogRequestID: true,
		LogUserAgent: true,
		LogMethod:    true,
		LogRemoteIP:  true,
		LogValuesFunc: func(c *echo.Context, values middleware.RequestLoggerValues) error {
			// Create Sentry logger with request context
			sentryLogger := sentry.NewLogger(c.Request().Context())
			stdLogger := log.New(sentryLogger, "", log.LstdFlags)

			// The signed-in user, if any. The token itself is never logged.
			principal, _ := request.PrincipalFromContext(c.Request().Context())
			userID := principal.Subject
			organizationID := ""
			if org, ok := c.Get("organization_id").(string); ok {
				organizationID = org
			}

			// Request body, captured and truncated by LogBody
			body, _ := c.Get(LogBodyKey).(string)

			// Get query parameters
			query := map[string]any{}
			_ = echo.BindQueryParams(c, &query)
			jsonQueryStr, _ := json.Marshal(query)

			// Get path parameters
			param := map[string]any{}
			_ = echo.BindPathValues(c, &param)
			jsonParamStr, _ := json.Marshal(param)

			// Get request headers (excluding sensitive ones)
			headers := make(map[string]string)
			for k, v := range c.Request().Header {
				// Skip sensitive headers
				if strings.EqualFold(k, "Authorization") ||
					strings.EqualFold(k, "Cookie") ||
					strings.EqualFold(k, "Set-Cookie") {
					continue
				}
				if len(v) > 0 {
					headers[k] = v[0]
				}
			}
			jsonHeadersStr, _ := json.Marshal(headers)

			ctx := c.Request().Context()
			correlationID, _ := request.CorrelationIDFromContext(ctx)
			languageCode, _ := request.LanguageCodeFromContext(ctx)

			// Format log message with all fields
			logMsg := fmt.Sprintf("Request: %s %s (status=%d, latency=%v, request_id=%s, correlation_id=%s, language=%s, remote_ip=%s, user_agent=%s, user_id=%s, org_id=%s, payload=%s, query=%s, path_params=%s, headers=%s, environment=%s, service=%s, version=%s, timestamp=%d, hostname=%s, protocol=%s)",
				values.Method,
				values.URI,
				values.Status,
				values.Latency,
				values.RequestID,
				correlationID,
				languageCode,
				values.RemoteIP,
				values.UserAgent,
				userID,
				organizationID,
				body,
				string(jsonQueryStr),
				string(jsonParamStr),
				string(jsonHeadersStr),
				cfg.AppEnv.String(),
				cfg.AppName,
				cfg.AppVersion,
				time.Now().UnixMilli(),
				c.Request().Host,
				c.Request().Proto,
			)

			// Log to Sentry
			stdLogger.Println(logMsg)

			// Also log to zap for structured logging
			logger.Log.Info("Request: "+values.Method+" "+values.URI,
				zap.String("uri", values.URI),
				zap.String("method", values.Method),
				zap.Int("status", values.Status),
				zap.Duration("latency", values.Latency),
				zap.String("request_id", values.RequestID),
				zap.String("remote_ip", values.RemoteIP),
				zap.String("correlation_id", correlationID),
				zap.String("language_code", languageCode),
				zap.String("user_agent", values.UserAgent),
				zap.String("user_id", userID),
				zap.String("org_id", organizationID),
				zap.String("payload", body),
				zap.String("query", string(jsonQueryStr)),
				zap.String("path_params", string(jsonParamStr)),
				zap.String("headers", string(jsonHeadersStr)),
				zap.String("environment", cfg.AppEnv.String()),
				zap.String("service", cfg.AppName),
				zap.String("version", cfg.AppVersion),
				zap.Int64("timestamp", time.Now().UnixMilli()),
				zap.String("hostname", c.Request().Host),
				zap.String("protocol", c.Request().Proto),
			)

			return nil
		},
	})
}

// LogBodyKey is the echo.Context key that holds the request body for logging.
const LogBodyKey = "log_body"

// truncatedSuffix marks a logged body that was cut at LOG_BODY_MAX_BYTES.
const truncatedSuffix = "…[truncated]"

// multipartMediaTypePrefix matches multipart/form-data, multipart/mixed and the like.
const multipartMediaTypePrefix = "multipart/"

// BodyLimit rejects bodies larger than APP_BODY_LIMIT_BYTES with 413. It runs before LogBody,
// so LogBody never reads more than that into memory.
func BodyLimit(cfg *config.Config) echo.MiddlewareFunc {
	return middleware.BodyLimit(cfg.AppBodyLimitBytes)
}

// LogBody keeps a copy of the request body for the request log, cut to LOG_BODY_MAX_BYTES.
// Multipart bodies (file uploads) are not read.
func LogBody(cfg *config.Config) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			req := c.Request()
			if req.Body == nil || req.Body == http.NoBody || isMultipart(req) {
				return next(c)
			}
			data, err := io.ReadAll(req.Body)
			if err != nil {
				return err
			}
			req.Body = io.NopCloser(bytes.NewReader(data))
			c.Set(LogBodyKey, truncateBody(data, cfg.LogBodyMaxBytes))
			return next(c)
		}
	}
}

func isMultipart(req *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(req.Header.Get(echo.HeaderContentType))
	return err == nil && strings.HasPrefix(mediaType, multipartMediaTypePrefix)
}

// truncateBody cuts data to maxBytes without splitting a UTF-8 character.
func truncateBody(data []byte, maxBytes int) string {
	if maxBytes <= 0 || len(data) <= maxBytes {
		return string(data)
	}
	return strings.ToValidUTF8(string(data[:maxBytes]), "") + truncatedSuffix
}
