// Package testenv is the mechanical half of TossOS's real-account protection.
//
// The rule it implements is in WORKFLOW's 불변 규칙: "실계좌 보호는 기계적으로 …
// 문구가 아니라 테스트 인프라가 막는다". A comment saying "tests must not touch the
// real account" protects nothing; a transport that refuses to dial Toss, and an
// environment where every state file lands in a temp directory, protects the
// account whether or not anybody read the comment.
//
// Two things live here, and each closes a different hole:
//
//	Isolate      every path TossOS resolves from the environment (config, cache,
//	             data, credentials) points inside t.TempDir(). A test cannot read
//	             the developer's session or write over their journal.
//	Guard        an http.RoundTripper that fails any request to a real Toss
//	             hostname, and fails a *mutation* to one loudly enough that the
//	             message says what nearly happened.
//
// # Why this is a normal package and not a _test.go file
//
// It has to be importable from other packages' tests, and Go does not export
// test-only code across package boundaries. The cost is that it ships in the
// binary; the content is a few hundred bytes of round-tripper and some
// environment names, and the alternative is copying the guard into every package
// that needs it, which is how one copy ends up missing.
package testenv

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// realHostSuffixes are the domains that reach Toss.
//
// Suffix matching rather than an exact list: the web session client derives host
// names at runtime (wts-api, wts-info-api, wts-cert-api, sse-message, www …) and
// an exact list would silently stop covering a host the day somebody adds one.
var realHostSuffixes = []string{
	"tossinvest.com",
	"toss.im",
	"tossbank.com",
}

// RealHostSuffixes returns the guarded domains.
func RealHostSuffixes() []string {
	return append([]string(nil), realHostSuffixes...)
}

// IsRealHost reports whether a host belongs to Toss.
func IsRealHost(host string) bool {
	h := strings.ToLower(strings.TrimSpace(host))
	// Strip a port; url.Host carries one and the suffix check must not miss it.
	if i := strings.LastIndex(h, ":"); i > 0 && !strings.Contains(h[i:], "]") {
		h = h[:i]
	}
	h = strings.TrimSuffix(h, ".")
	for _, suffix := range realHostSuffixes {
		if h == suffix || strings.HasSuffix(h, "."+suffix) {
			return true
		}
	}
	return false
}

// mutatingMethods are the ones that can change a live account.
func isMutation(method string) bool {
	switch strings.ToUpper(method) {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

// ErrRealHost is what the guard returns. It is a distinct type so a test can
// assert on the interception rather than on a message.
type ErrRealHost struct {
	Method   string
	URL      string
	Mutation bool
}

func (e *ErrRealHost) Error() string {
	if e.Mutation {
		return fmt.Sprintf(
			"testenv: BLOCKED a %s to the real Toss host %s — a test tried to mutate a live account. "+
				"Use httptest and official.WithBaseURL/WithHTTPClient.", e.Method, e.URL)
	}
	return fmt.Sprintf(
		"testenv: blocked a %s to the real Toss host %s — tests must use httptest, never the live API",
		e.Method, e.URL)
}

// Guard wraps a round tripper and refuses anything aimed at Toss.
//
// Reads are refused as well as writes. A read is not dangerous to the account,
// but a test that reaches the live API is a test whose result depends on
// somebody's portfolio, and that is its own kind of broken. The mutation case
// gets its own message because it is the one that would have done damage.
type Guard struct {
	// Base is the transport used for everything that is not a real Toss host.
	// Nil uses http.DefaultTransport.
	Base http.RoundTripper
	// OnBlock is called with the error before it is returned. Tests use it to
	// fail the enclosing test rather than only the request.
	OnBlock func(err *ErrRealHost)
}

// RoundTrip implements http.RoundTripper.
func (g *Guard) RoundTrip(req *http.Request) (*http.Response, error) {
	if req != nil && req.URL != nil && IsRealHost(req.URL.Host) {
		err := &ErrRealHost{
			Method:   req.Method,
			URL:      req.URL.Scheme + "://" + req.URL.Host + req.URL.Path,
			Mutation: isMutation(req.Method),
		}
		if g.OnBlock != nil {
			g.OnBlock(err)
		}
		return nil, err
	}
	base := g.Base
	if base == nil {
		base = http.DefaultTransport
	}
	return base.RoundTrip(req)
}

// GuardedClient returns an http.Client whose transport refuses Toss, and which
// fails the test if anything reaches for it.
func GuardedClient(t *testing.T, base http.RoundTripper) *http.Client {
	t.Helper()
	return &http.Client{Transport: &Guard{
		Base: base,
		OnBlock: func(err *ErrRealHost) {
			t.Errorf("%v", err)
		},
	}}
}

// InstallGuard replaces http.DefaultTransport for the duration of the test.
//
// It catches the case the per-client guards cannot: a code path that builds its
// own http.Client, or uses http.Get, and therefore never sees an injected
// transport. Call it from TestMain to cover a whole package.
func InstallGuard(t *testing.T) {
	t.Helper()
	previous := http.DefaultTransport
	http.DefaultTransport = &Guard{Base: previous, OnBlock: func(err *ErrRealHost) {
		t.Errorf("%v", err)
	}}
	t.Cleanup(func() { http.DefaultTransport = previous })
}

// Env names the environment variables that decide where TossOS reads and writes.
//
// Every one of them has to be redirected, because a test that isolates three of
// four still writes the fourth into the developer's home directory — which is
// how task 4.2 nearly put an audit log in ~/.local/share/tossos (issues.md).
var Env = []string{
	"XDG_CONFIG_HOME",
	"XDG_CACHE_HOME",
	"XDG_DATA_HOME",
	"TOSSOS_DATA_DIR",
	"TOSSCTL_OPENAPI_KEY",
	"TOSSCTL_OPENAPI_SECRET",
	"TOSSCTL_LANG",
}

// Isolate points every TossOS path at a fresh temporary directory and returns
// the config directory to pass as --config-dir.
//
// It uses t.Setenv, which fails the test if it is called from a parallel test —
// that restriction is a feature here: a parallel test sharing process-wide
// environment is a test that can leak one suite's paths into another's.
func Isolate(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	t.Setenv("XDG_CONFIG_HOME", filepath.Join(root, "config"))
	t.Setenv("XDG_CACHE_HOME", filepath.Join(root, "cache"))
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "data"))
	t.Setenv("TOSSOS_DATA_DIR", filepath.Join(root, "tossos"))
	// Credentials come from the environment before they come from a file, so an
	// unset here is what stops a developer's real key reaching a test.
	t.Setenv("TOSSCTL_OPENAPI_KEY", "")
	t.Setenv("TOSSCTL_OPENAPI_SECRET", "")

	configDir := filepath.Join(root, "cfg")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatalf("testenv: creating the isolated config directory: %v", err)
	}
	return configDir
}

// DataDir returns the isolated data directory Isolate established.
func DataDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("TOSSOS_DATA_DIR")
	if dir == "" {
		t.Fatal("testenv: no isolated data directory; call Isolate first")
	}
	return dir
}
