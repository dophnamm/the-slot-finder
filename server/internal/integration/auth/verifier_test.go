package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"the-slot-finder/server/internal/clock"
	"the-slot-finder/server/internal/config"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testRealm     = "slot-finder"
	testAPIClient = "slot-finder-api"
	testSPAClient = "slot-finder-web"
	testPublicURL = "http://localhost:8080/auth"
)

var testNow = time.Date(2026, time.October, 12, 9, 0, 0, 0, time.UTC)

// jwksServer serves a JWK Set that tests can rotate.
type jwksServer struct {
	mu       sync.Mutex
	keys     map[string]*rsa.PrivateKey
	requests atomic.Int32
	srv      *httptest.Server
}

func newJWKSServer(t *testing.T) *jwksServer {
	t.Helper()
	s := &jwksServer{keys: map[string]*rsa.PrivateKey{}}
	s.srv = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *jwksServer) addKey(t *testing.T, kid string) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.keys[kid] = key
	return key
}

func (s *jwksServer) serve(w http.ResponseWriter, _ *http.Request) {
	s.requests.Add(1)
	s.mu.Lock()
	defer s.mu.Unlock()
	keys := make([]map[string]string, 0, len(s.keys))
	for kid, key := range s.keys {
		keys = append(keys, map[string]string{
			"kty": "RSA",
			"kid": kid,
			"use": "sig",
			"alg": jwt.SigningMethodRS256.Alg(),
			"n":   base64.RawURLEncoding.EncodeToString(key.N.Bytes()),
			"e":   base64.RawURLEncoding.EncodeToString(big.NewInt(int64(key.E)).Bytes()),
		})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"keys": keys})
}

func testConfig(jwksBase string) *config.Config {
	return &config.Config{
		KeycloakURL:                jwksBase,
		KeycloakPublicURL:          testPublicURL,
		KeycloakRealm:              testRealm,
		KeycloakClientID:           testAPIClient,
		KeycloakSPAClientID:        testSPAClient,
		HTTPClientTimeout:          time.Second,
		AuthJWKSRefreshInterval:    time.Hour,
		AuthJWKSUnknownKIDInterval: time.Millisecond,
		AuthJWKSRefreshWaitMax:     time.Second,
		AuthClockSkew:              30 * time.Second,
	}
}

func validClaims() jwt.MapClaims {
	return jwt.MapClaims{
		"iss":                testPublicURL + "/realms/" + testRealm,
		"sub":                "user-1",
		"aud":                []string{testAPIClient, "account"},
		"azp":                testSPAClient,
		"exp":                testNow.Add(5 * time.Minute).Unix(),
		"iat":                testNow.Add(-time.Minute).Unix(),
		"preferred_username": "planner",
		"email":              "planner@example.test",
		"realm_access":       map[string]any{"roles": []string{"planner"}},
	}
}

func sign(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = kid
	raw, err := token.SignedString(key)
	require.NoError(t, err)
	return raw
}

func newVerifier(t *testing.T, srv *jwksServer) TokenVerifier {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	// The JWKS URL is built from the base URL, realm and certs path; the test server answers any path.
	v, err := NewTokenVerifier(ctx, testConfig(srv.srv.URL), clock.NewFixed(testNow))
	require.NoError(t, err)
	return v
}

func TestJWKSVerifier_Verify(t *testing.T) {
	srv := newJWKSServer(t)
	key := srv.addKey(t, "k1")
	otherKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	v := newVerifier(t, srv)

	with := func(change func(jwt.MapClaims)) jwt.MapClaims {
		c := validClaims()
		change(c)
		return c
	}

	tests := []struct {
		name    string
		token   string
		wantErr error
	}{
		{name: "valid", token: sign(t, key, "k1", validClaims())},
		{name: "azp only, no API audience", token: sign(t, key, "k1", with(func(c jwt.MapClaims) { c["aud"] = "account" }))},
		{name: "expired", token: sign(t, key, "k1", with(func(c jwt.MapClaims) { c["exp"] = testNow.Add(-time.Minute).Unix() })), wantErr: ErrTokenInvalid},
		{name: "expired within skew", token: sign(t, key, "k1", with(func(c jwt.MapClaims) { c["exp"] = testNow.Add(-10 * time.Second).Unix() }))},
		{name: "no expiry", token: sign(t, key, "k1", with(func(c jwt.MapClaims) { delete(c, "exp") })), wantErr: ErrTokenInvalid},
		{name: "wrong issuer", token: sign(t, key, "k1", with(func(c jwt.MapClaims) { c["iss"] = "http://evil.test/realms/" + testRealm })), wantErr: ErrTokenInvalid},
		{name: "wrong audience and azp", token: sign(t, key, "k1", with(func(c jwt.MapClaims) { c["aud"] = "account"; c["azp"] = "other-client" })), wantErr: ErrTokenAudience},
		{name: "bad signature", token: sign(t, otherKey, "k1", validClaims()), wantErr: ErrTokenInvalid},
		{name: "not a JWT", token: "not-a-jwt", wantErr: ErrTokenInvalid},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := v.Verify(context.Background(), tt.token)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, claims)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "user-1", claims.Subject)
			assert.Equal(t, "planner", claims.PreferredUsername)
			assert.Equal(t, []string{"planner"}, claims.Roles(testAPIClient))
		})
	}
}

func TestJWKSVerifier_RejectsHMAC(t *testing.T) {
	srv := newJWKSServer(t)
	srv.addKey(t, "k1")
	v := newVerifier(t, srv)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, validClaims())
	token.Header["kid"] = "k1"
	raw, err := token.SignedString([]byte("shared-secret"))
	require.NoError(t, err)

	_, err = v.Verify(context.Background(), raw)
	require.ErrorIs(t, err, ErrTokenInvalid)
}

func TestJWKSVerifier_RefreshesOnUnknownKID(t *testing.T) {
	srv := newJWKSServer(t)
	key := srv.addKey(t, "k1")
	v := newVerifier(t, srv)

	_, err := v.Verify(context.Background(), sign(t, key, "k1", validClaims()))
	require.NoError(t, err)
	fetchesBefore := srv.requests.Load()

	// Verifying with a known key again does not fetch the JWKS.
	_, err = v.Verify(context.Background(), sign(t, key, "k1", validClaims()))
	require.NoError(t, err)
	assert.Equal(t, fetchesBefore, srv.requests.Load())

	// Keycloak rotates its key: the new kid triggers one refresh.
	rotated := srv.addKey(t, "k2")
	_, err = v.Verify(context.Background(), sign(t, rotated, "k2", validClaims()))
	require.NoError(t, err)
	assert.Greater(t, srv.requests.Load(), fetchesBefore)
}

func TestTokenClaims_Roles(t *testing.T) {
	var nilClaims *TokenClaims
	assert.Nil(t, nilClaims.Roles(testAPIClient))

	claims := &TokenClaims{
		RealmAccess:    RoleClaim{Roles: []string{"manager"}},
		ResourceAccess: map[string]RoleClaim{testAPIClient: {Roles: []string{"planner"}}, "other": {Roles: []string{"x"}}},
	}
	assert.Equal(t, []string{"manager", "planner"}, claims.Roles(testAPIClient))
}
