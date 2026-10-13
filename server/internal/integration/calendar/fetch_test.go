package calendar

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	loopbackHost = "127.0.0.1"
	secretPath   = "/calendar/ical/secret-token-123/basic.ics"
	icsBody      = "BEGIN:VCALENDAR\r\nEND:VCALENDAR\r\n"
	testMaxBytes = 64
)

func allowAll(netip.Addr) bool { return true }

// testFetcher talks to a local TLS test server: it trusts its certificate and allows
// loopback, which the real fetcher refuses.
func testFetcher(srv *httptest.Server, allowAddr func(netip.Addr) bool, hosts ...string) *httpFetcher {
	policy := FetchPolicy{AllowedHosts: hosts, Timeout: 2 * time.Second, MaxBytes: testMaxBytes}
	return newHTTPFetcher(policy, allowAddr, srv.Client().Transport.(*http.Transport).TLSClientConfig.RootCAs)
}

func TestHTTPFetcher_Fetch(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case secretPath:
			_, _ = w.Write([]byte(icsBody))
		case "/big":
			_, _ = w.Write([]byte(strings.Repeat("x", testMaxBytes+1)))
		case "/away":
			http.Redirect(w, r, "https://accounts.google.com/ServiceLogin", http.StatusFound)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	tests := []struct {
		name    string
		fetcher *httpFetcher
		url     string
		want    string
		wantErr error
	}{
		{name: "reads an allowed calendar", fetcher: testFetcher(srv, allowAll, loopbackHost), url: srv.URL + secretPath, want: icsBody},
		{name: "http is refused before any request", fetcher: testFetcher(srv, allowAll, loopbackHost),
			url: strings.Replace(srv.URL, "https://", "http://", 1) + secretPath, wantErr: ErrHostNotAllowed},
		{name: "a host outside the list is refused", fetcher: testFetcher(srv, allowAll, "calendar.google.com"),
			url: srv.URL + secretPath, wantErr: ErrHostNotAllowed},
		{name: "an allowed name that resolves to loopback is refused when dialling", fetcher: testFetcher(srv, publicAddr, loopbackHost),
			url: srv.URL + secretPath, wantErr: ErrHostNotAllowed},
		// A link the organization blocks redirects to a sign-in page: the link was right, so the
		// person is told the calendar could not be fetched, not to paste another link.
		{name: "a redirect to another host, such as a sign-in page, is refused as unreachable",
			fetcher: testFetcher(srv, allowAll, loopbackHost), url: srv.URL + "/away", wantErr: ErrUnreachable},
		{name: "a body over the limit", fetcher: testFetcher(srv, allowAll, loopbackHost), url: srv.URL + "/big", wantErr: ErrTooLarge},
		{name: "a missing calendar", fetcher: testFetcher(srv, allowAll, loopbackHost), url: srv.URL + "/gone", wantErr: ErrUnreachable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.fetcher.Fetch(context.Background(), tt.url)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				assert.NotContains(t, err.Error(), "secret-token", "an error must never hold the link")
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestCheckURL(t *testing.T) {
	hosts := []string{"calendar.google.com", "outlook.office365.com"}
	tests := []struct {
		name     string
		url      string
		wantHost string
	}{
		{name: "google secret address", url: "https://calendar.google.com/calendar/ical/x%40y/private-abc/basic.ics", wantHost: "calendar.google.com"},
		{name: "host matched ignoring case", url: "https://Outlook.Office365.com/owa/calendar/a/b/calendar.ics", wantHost: "outlook.office365.com"},
		{name: "http", url: "http://calendar.google.com/x.ics"},
		{name: "credentials in the link", url: "https://user:pw@calendar.google.com/x.ics"},
		{name: "look-alike host", url: "https://calendar.google.com.evil.example/x.ics"},
		{name: "not a link", url: "calendar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			host, err := CheckURL(tt.url, hosts)
			if tt.wantHost == "" {
				assert.ErrorIs(t, err, ErrHostNotAllowed)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantHost, host)
		})
	}
}

func TestPublicAddr(t *testing.T) {
	for addr, want := range map[string]bool{
		"8.8.8.8": true, "2001:4860:4860::8888": true,
		"127.0.0.1": false, "10.1.2.3": false, "172.16.0.1": false, "192.168.1.1": false, "169.254.169.254": false,
		"100.64.0.1": false, "0.0.0.0": false, "::1": false, "fd00::1": false, "fe80::1": false,
	} {
		assert.Equal(t, want, publicAddr(netip.MustParseAddr(addr)), addr)
	}
}
