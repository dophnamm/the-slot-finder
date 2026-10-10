package auth

import (
	"context"
	stderrors "errors"
	"fmt"
	"slices"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"
	"the-slot-finder/server/internal/logger"

	"github.com/MicahParks/keyfunc/v3"
	jwt "github.com/golang-jwt/jwt/v5"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// Keycloak signs access tokens with RS256 unless the realm is changed on purpose.
var allowedSigningMethods = []string{jwt.SigningMethodRS256.Alg()}

var (
	// ErrTokenInvalid means the signature, issuer, expiry or format is wrong.
	ErrTokenInvalid = stderrors.New("invalid access token")
	// ErrTokenAudience means the token was issued for another client.
	ErrTokenAudience = stderrors.New("access token is not meant for this API")
)

// TokenVerifier checks access tokens locally, without calling Keycloak per request.
type TokenVerifier interface {
	Verify(ctx context.Context, rawToken string) (*TokenClaims, error)
}

type jwksVerifier struct {
	keys            keyfunc.Keyfunc
	parser          *jwt.Parser
	audience        string
	authorizedParty string
}

// ProvideTokenVerifier verifies tokens against the realm's JWKS. The keys are fetched once,
// refreshed every AUTH_JWKS_REFRESH_INTERVAL, and refetched (rate-limited) when a token
// carries an unknown key id. The refresh goroutine stops with the application.
func ProvideTokenVerifier(lc fx.Lifecycle, cfg *config.Config, clk clock.Clock) (TokenVerifier, error) {
	ctx, cancel := context.WithCancel(context.Background())
	verifier, err := NewTokenVerifier(ctx, cfg, clk)
	if err != nil {
		cancel()
		return nil, err
	}
	lc.Append(fx.Hook{OnStop: func(context.Context) error {
		cancel()
		return nil
	}})
	return verifier, nil
}

// NewTokenVerifier builds a verifier whose key refresh runs until ctx ends.
func NewTokenVerifier(ctx context.Context, cfg *config.Config, clk clock.Clock) (TokenVerifier, error) {
	jwksURL := cfg.KeycloakJWKSURL()
	keys, err := keyfunc.NewDefaultOverrideCtx(ctx, []string{jwksURL}, keyfunc.Override{
		HTTPTimeout:             cfg.HTTPClientTimeout,
		RefreshInterval:         cfg.AuthJWKSRefreshInterval,
		RefreshUnknownKID:       rate.NewLimiter(rate.Every(cfg.AuthJWKSUnknownKIDInterval), 1),
		RateLimitWaitMax:        cfg.AuthJWKSRefreshWaitMax,
		RefreshErrorHandlerFunc: logRefreshError,
	})
	if err != nil {
		return nil, fmt.Errorf("load JWKS from %s: %w", jwksURL, err)
	}

	return &jwksVerifier{
		keys: keys,
		parser: jwt.NewParser(
			jwt.WithValidMethods(allowedSigningMethods),
			jwt.WithIssuer(cfg.KeycloakIssuer()),
			jwt.WithExpirationRequired(),
			jwt.WithIssuedAt(),
			jwt.WithLeeway(cfg.AuthClockSkew),
			jwt.WithTimeFunc(clk.Now),
		),
		audience:        cfg.KeycloakClientID,
		authorizedParty: cfg.KeycloakSPAClientID,
	}, nil
}

// Verify checks the signature, issuer, expiry and audience. The token must list this API's
// client in aud, or have been issued to the web app's client (azp).
func (v *jwksVerifier) Verify(ctx context.Context, rawToken string) (*TokenClaims, error) {
	claims := &TokenClaims{}
	if _, err := v.parser.ParseWithClaims(rawToken, claims, v.keys.KeyfuncCtx(ctx)); err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTokenInvalid, err)
	}
	if !v.audienceAllowed(claims) {
		return nil, ErrTokenAudience
	}
	return claims, nil
}

func (v *jwksVerifier) audienceAllowed(claims *TokenClaims) bool {
	if v.audience != "" && slices.Contains(claims.Audience, v.audience) {
		return true
	}
	return v.authorizedParty != "" && claims.AuthorizedParty == v.authorizedParty
}

func logRefreshError(url string) func(ctx context.Context, err error) {
	return func(_ context.Context, err error) {
		if logger.Log == nil {
			return
		}
		logger.Log.Error("Failed to refresh the JWKS", zap.String("url", url), zap.Error(err))
	}
}
