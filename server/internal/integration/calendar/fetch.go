package calendar

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	stderrors "errors"
	"io"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strings"
	"syscall"
	"time"
)

// schemeHTTPS is the only scheme a calendar link may use.
const schemeHTTPS = "https"

// maxRedirects is how many redirects one fetch follows, each to an allowed host.
const maxRedirects = 3

// blockedPrefixes are address ranges that are not on the public internet, beyond the ones
// net/netip already names (loopback, private, link-local, multicast, unspecified).
var blockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("0.0.0.0/8"),
	netip.MustParsePrefix("100.64.0.0/10"),
	netip.MustParsePrefix("192.0.0.0/24"),
	netip.MustParsePrefix("198.18.0.0/15"),
	netip.MustParsePrefix("240.0.0.0/4"),
}

// Fetcher downloads a calendar. Tests pass a fake, so no test reaches the network.
type Fetcher interface {
	Fetch(ctx context.Context, rawURL string) ([]byte, error)
}

// FetchPolicy limits what a fetch may reach and read.
type FetchPolicy struct {
	AllowedHosts []string
	Timeout      time.Duration
	MaxBytes     int64
}

type httpFetcher struct {
	policy FetchPolicy
	client *http.Client
}

// NewHTTPFetcher fetches over https from allowed hosts only, and refuses to connect to an
// address that is not public, whatever the host name resolves to (DNS rebinding included).
func NewHTTPFetcher(policy FetchPolicy) Fetcher {
	return newHTTPFetcher(policy, publicAddr, nil)
}

// newHTTPFetcher lets tests allow the loopback test server and trust its certificate.
func newHTTPFetcher(policy FetchPolicy, allowAddr func(netip.Addr) bool, roots *x509.CertPool) *httpFetcher {
	dialer := &net.Dialer{
		Timeout: policy.Timeout,
		// Control runs after DNS, on the address actually dialled.
		Control: func(_, address string, _ syscall.RawConn) error {
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return ErrHostNotAllowed
			}
			addr, err := netip.ParseAddr(host)
			if err != nil || !allowAddr(addr.Unmap()) {
				return ErrHostNotAllowed
			}
			return nil
		},
	}
	transport := &http.Transport{
		Proxy:                 nil,
		DialContext:           dialer.DialContext,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12, RootCAs: roots},
		TLSHandshakeTimeout:   policy.Timeout,
		ResponseHeaderTimeout: policy.Timeout,
	}
	f := &httpFetcher{policy: policy}
	f.client = &http.Client{Transport: transport, Timeout: policy.Timeout, CheckRedirect: f.checkRedirect}
	return f
}

func (f *httpFetcher) checkRedirect(req *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return ErrUnreachable
	}
	// A link the organization blocks redirects to a sign-in page. The link itself was right, so
	// this is "unreachable", whose message names the block, not "host not allowed".
	if _, err := CheckURL(req.URL.String(), f.policy.AllowedHosts); err != nil {
		return ErrUnreachable
	}
	return nil
}

func (f *httpFetcher) Fetch(ctx context.Context, rawURL string) ([]byte, error) {
	if _, err := CheckURL(rawURL, f.policy.AllowedHosts); err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, ErrUnreachable
	}
	resp, err := f.client.Do(req)
	if err != nil {
		// The error text holds the link, so only the kind of failure goes back.
		if stderrors.Is(err, ErrHostNotAllowed) {
			return nil, ErrHostNotAllowed
		}
		return nil, ErrUnreachable
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, ErrUnreachable
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, f.policy.MaxBytes+1))
	if err != nil {
		return nil, ErrUnreachable
	}
	if int64(len(data)) > f.policy.MaxBytes {
		return nil, ErrTooLarge
	}
	return data, nil
}

// CheckURL accepts an https link without credentials to an allowed host, and returns the host.
func CheckURL(rawURL string, allowedHosts []string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || u.Scheme != schemeHTTPS || u.User != nil || u.Hostname() == "" {
		return "", ErrHostNotAllowed
	}
	host := strings.ToLower(u.Hostname())
	if !slices.ContainsFunc(allowedHosts, func(h string) bool { return strings.EqualFold(strings.TrimSpace(h), host) }) {
		return "", ErrHostNotAllowed
	}
	return host, nil
}

// publicAddr reports an address on the public internet.
func publicAddr(addr netip.Addr) bool {
	if !addr.IsValid() || addr.IsLoopback() || addr.IsPrivate() || addr.IsLinkLocalUnicast() || addr.IsLinkLocalMulticast() ||
		addr.IsInterfaceLocalMulticast() || addr.IsMulticast() || addr.IsUnspecified() {
		return false
	}
	return !slices.ContainsFunc(blockedPrefixes, func(p netip.Prefix) bool { return p.Contains(addr) })
}
