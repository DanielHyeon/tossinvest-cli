package official

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestA112MBUSCursorBytesBodyBoundsAndContentEncodingHold(t *testing.T) {
	tests := []struct {
		name    string
		before  []byte
		headers http.Header
		body    func() []byte
	}{
		{
			name:    "cursor bytes percent encoded once",
			before:  []byte("\x00 space 🍏%"),
			headers: a112MBUSGoodRateHeaders(),
			body:    func() []byte { return []byte(`{"result":{"nextBefore":null}}`) },
		},
		{
			name:    "unexpected gzip",
			headers: http.Header{"Content-Encoding": {"gzip"}, "X-RateLimit-Limit": {"10"}, "X-RateLimit-Remaining": {"9"}, "X-RateLimit-Reset": {"1"}},
			body:    func() []byte { return []byte(`{"result":{"nextBefore":null}}`) },
		},
		{
			name:    "unexpected br",
			headers: http.Header{"Content-Encoding": {"br"}, "X-RateLimit-Limit": {"10"}, "X-RateLimit-Remaining": {"9"}, "X-RateLimit-Reset": {"1"}},
			body:    func() []byte { return []byte(`{"result":{"nextBefore":null}}`) },
		},
		{
			name:    "chunked limit plus one",
			headers: a112MBUSGoodRateHeaders(),
			body: func() []byte {
				return append([]byte(`{"result":{"nextBefore":null},"padding":"`), append(make([]byte, a112MBUSBodyLimit), []byte(`"}`)...)...)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls.Add(1)
				if tt.before != nil && r.URL.RawQuery != "adjusted=false&before=%00+space+%F0%9F%8D%8F%25&count=200&interval=1m&symbol=AAPL" {
					t.Fatalf("cursor query = %q", r.URL.RawQuery)
				}
				for key, values := range tt.headers {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
				_, _ = w.Write(tt.body())
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			got, err := A112MBUSCandle(ctx, a112MBUSTestClient(t, server), tt.before)
			if tt.name == "cursor bytes percent encoded once" {
				if err != nil || string(got.CursorJSON()) != "null" {
					t.Fatalf("exact cursor request = %+v, %v", got, err)
				}
			} else if err == nil {
				t.Fatal("unsafe response minted evidence")
			}
			if calls.Load() != 1 {
				t.Fatalf("calls = %d, want 1", calls.Load())
			}
		})
	}
}

func TestA112MBUSRedirectAndUnsafeCacheHoldBeforeRetry(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Redirect(w, r, "/redirected", http.StatusFound)
	}))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := A112MBUSCandle(ctx, a112MBUSTestClient(t, server), nil); err == nil {
		t.Fatal("redirect minted evidence")
	}
	if calls.Load() != 1 {
		t.Fatalf("redirect requests = %d, want 1", calls.Load())
	}

	noDataCalls := atomic.Int32{}
	noData := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { noDataCalls.Add(1) }))
	defer noData.Close()
	client := a112MBUSTestClient(t, noData)
	if err := os.Chmod(client.tm.cacheFile, 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := A112MBUSCandle(ctx, client, nil); err == nil {
		t.Fatal("non-0600 cache accepted")
	}
	if noDataCalls.Load() != 0 {
		t.Fatalf("unsafe cache caused data calls=%d", noDataCalls.Load())
	}
}

func TestA112MBUSCacheMissExpiryAndSymlinksHoldBeforeDataGET(t *testing.T) {
	serverCalls := atomic.Int32{}
	server := httptest.NewTLSServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { serverCalls.Add(1) }))
	defer server.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, setup := range []struct {
		name string
		edit func(t *testing.T, client *Client)
	}{
		{"cache miss", func(_ *testing.T, client *Client) { _ = os.Remove(client.tm.cacheFile) }},
		{"expired", func(t *testing.T, client *Client) {
			data, err := json.Marshal(cachedToken{AccessToken: "expired", ExpiresAt: time.Now().Add(-time.Hour)})
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(client.tm.cacheFile, data, 0o600); err != nil {
				t.Fatal(err)
			}
		}},
		{"symlink leaf", func(t *testing.T, client *Client) {
			real := filepath.Join(t.TempDir(), "real-token.json")
			if err := os.Rename(client.tm.cacheFile, real); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(real, client.tm.cacheFile); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
		}},
		{"symlink ancestor", func(t *testing.T, client *Client) {
			cacheDir := filepath.Dir(client.tm.cacheFile)
			real := cacheDir + "-real"
			if err := os.Rename(cacheDir, real); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(real, cacheDir); err != nil {
				t.Skipf("symlink unavailable: %v", err)
			}
		}},
	} {
		t.Run(setup.name, func(t *testing.T) {
			client := a112MBUSTestClient(t, server)
			setup.edit(t, client)
			if _, err := A112MBUSCandle(ctx, client, nil); err == nil {
				t.Fatal("unsafe cache minted evidence")
			}
		})
	}
	if serverCalls.Load() != 0 {
		t.Fatalf("cache failures sent data requests=%d", serverCalls.Load())
	}
}

func TestA112MBUSStaticNoProductCallerOrForbiddenPath(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if violations, err := a112MBUSRepositoryForbiddenRefs(root); err != nil {
		t.Fatal(err)
	} else if len(violations) != 0 {
		t.Fatalf("forbidden resolved A112 M-B0 references: %#v", violations)
	}
	for _, file := range []string{"a112_mbus_read.go", "a112_mbus_read_unix.go"} {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		for _, forbidden := range []string{".token(", ".refresh(", ".exchange(", ".saveCache(", "AuthHeaders(", ".get(", ".send("} {
			if strings.Contains(string(data), forbidden) {
				t.Fatalf("%s contains forbidden ordinary path %q", file, forbidden)
			}
		}
	}
}

func TestA112MBUSResolvedASTGuardCatchesProductSelectorAndTypeReference(t *testing.T) {
	source := `package product
import off "github.com/JungHoonGhae/tossinvest-cli/internal/official"
var _ off.A112MBUSResult
func read() { _, _ = off.A112MBUSCandle(nil, nil, nil) }
`
	if got := a112MBUSFindForbiddenRefs("internal/product/fixture.go", source); len(got) != 2 {
		t.Fatalf("forbidden A112 M-B0 product references = %#v, want selector and type violations", got)
	}
}

func TestA112MBUSResolvedASTGuardCatchesSamePackageUnqualifiedReference(t *testing.T) {
	source := `package official
var _ A112MBUSResult
var _ a112MBUSDescriptor
func read() {
	_, _ = A112MBUSCandle(nil, nil, nil)
	_ = a112MBUSRead
	_ = a112MBUSReadAt
	_ = a112MBUSCachedToken
}
`
	if got := a112MBUSFindForbiddenRefs("internal/official/ordinary_reader.go", source); len(got) != 6 {
		t.Fatalf("same-package unqualified M-B0 references = %#v, want exported and private capability violations", got)
	}
}

func TestA112MBUSResolvedASTGuardIgnoresUnrelatedOfficialPackageLocal(t *testing.T) {
	source := `package official
func fixture() { a112MBUSLocal := 1; _ = a112MBUSLocal }
`
	if got := a112MBUSFindForbiddenRefs("internal/other/official_fixture.go", source); len(got) != 0 {
		t.Fatalf("unrelated local identifier produced violations: %#v", got)
	}
}

func TestA112MBUSClonedTransportNeverCallsConfiguredProxy(t *testing.T) {
	var proxyCalls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		for key, values := range a112MBUSGoodRateHeaders() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		_, _ = w.Write([]byte(`{"result":{"nextBefore":null}}`))
	}))
	defer server.Close()
	client := a112MBUSTestClient(t, server)
	client.authorityTransport.Proxy = func(*http.Request) (*url.URL, error) {
		proxyCalls.Add(1)
		return &url.URL{Scheme: "http", Host: "proxy.invalid"}, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := A112MBUSCandle(ctx, client, nil); err != nil {
		t.Fatal(err)
	}
	if proxyCalls.Load() != 0 {
		t.Fatalf("configured proxy called %d times", proxyCalls.Load())
	}
}

func TestA112MBUSContentLengthBoundaryAndRateHeaderMatrix(t *testing.T) {
	for _, size := range []int{a112MBUSBodyLimit, a112MBUSBodyLimit + 1} {
		t.Run(strconv.Itoa(size), func(t *testing.T) {
			prefix, suffix := `{"result":{"nextBefore":null},"padding":"`, `"}`
			body := []byte(prefix + strings.Repeat("x", size-len(prefix)-len(suffix)) + suffix)
			server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				for key, values := range a112MBUSGoodRateHeaders() {
					for _, value := range values {
						w.Header().Add(key, value)
					}
				}
				w.Header().Set("Content-Length", strconv.Itoa(len(body)))
				_, _ = w.Write(body)
			}))
			defer server.Close()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, err := A112MBUSCandle(ctx, a112MBUSTestClient(t, server), nil)
			if size == a112MBUSBodyLimit && err != nil {
				t.Fatalf("exact 2 MiB rejected: %v", err)
			}
			if size > a112MBUSBodyLimit && err == nil {
				t.Fatal("2 MiB plus one minted evidence")
			}
		})
	}

	for _, test := range []struct {
		name string
		head http.Header
		ok   bool
	}{
		{"valid", a112MBUSGoodRateHeaders(), true},
		{"missing limit", http.Header{"X-RateLimit-Remaining": {"1"}, "X-RateLimit-Reset": {"1"}}, false},
		{"missing remaining", http.Header{"X-RateLimit-Limit": {"1"}, "X-RateLimit-Reset": {"1"}}, false},
		{"missing reset", http.Header{"X-RateLimit-Limit": {"1"}, "X-RateLimit-Remaining": {"1"}}, false},
		{"malformed", http.Header{"X-RateLimit-Limit": {"one"}, "X-RateLimit-Remaining": {"1"}, "X-RateLimit-Reset": {"1"}}, false},
		{"negative", http.Header{"X-RateLimit-Limit": {"1"}, "X-RateLimit-Remaining": {"-1"}, "X-RateLimit-Reset": {"1"}}, false},
		{"zero limit", http.Header{"X-RateLimit-Limit": {"0"}, "X-RateLimit-Remaining": {"0"}, "X-RateLimit-Reset": {"0"}}, false},
		{"success retry-after", http.Header{"X-RateLimit-Limit": {"1"}, "X-RateLimit-Remaining": {"1"}, "X-RateLimit-Reset": {"1"}, "Retry-After": {"1"}}, false},
	} {
		t.Run("rate "+test.name, func(t *testing.T) {
			_, err := a112MBUSValidateSuccessRateHeaders(a112MBUSCanonicalHeader(test.head))
			if (err == nil) != test.ok {
				t.Fatalf("success rate headers err=%v want ok=%t", err, test.ok)
			}
		})
	}
	for _, header := range []http.Header{
		{}, {"Retry-After": {"1", "2"}}, {"Retry-After": {"later"}}, {"Retry-After": {"-1"}}, {"Retry-After": {"1"}},
	} {
		header = a112MBUSCanonicalHeader(header)
		err := a112MBUSValidate429(header)
		if len(header.Values("Retry-After")) == 1 && header.Values("Retry-After")[0] == "1" {
			if err != nil {
				t.Fatalf("valid 429 Retry-After rejected: %v", err)
			}
		} else if err == nil {
			t.Fatalf("invalid 429 header accepted: %#v", header)
		}
	}
}

func a112MBUSCanonicalHeader(input http.Header) http.Header {
	out := make(http.Header, len(input))
	for key, values := range input {
		out[http.CanonicalHeaderKey(key)] = append([]string(nil), values...)
	}
	return out
}

func TestA112MBUSClockSkewAndDeadlineBoundariesArePrivate(t *testing.T) {
	fixed := time.Date(2026, time.August, 15, 9, 0, 0, 0, time.UTC)
	cache := filepath.Join(t.TempDir(), "token.json")
	write := func(expiry time.Time) {
		data, err := json.Marshal(cachedToken{AccessToken: "cached", ExpiresAt: expiry})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(cache, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	write(fixed.Add(60*time.Second + time.Nanosecond))
	if _, err := a112MBUSCachedToken(cache, fixed); err != nil {
		t.Fatalf("token just beyond skew rejected: %v", err)
	}
	write(fixed.Add(60 * time.Second))
	if _, err := a112MBUSCachedToken(cache, fixed); err == nil {
		t.Fatal("token exactly at skew accepted")
	}
	for _, duration := range []time.Duration{time.Nanosecond, 15 * time.Second, 15*time.Second + time.Nanosecond, 0, -time.Nanosecond} {
		if got, want := a112MBUSDeadlineValid(fixed.Add(duration), fixed), duration > 0 && duration <= 15*time.Second; got != want {
			t.Fatalf("deadline %s valid=%t want=%t", duration, got, want)
		}
	}
	data, err := os.ReadFile("a112_mbus_read.go")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "nowValue:") || strings.Contains(string(data), "evaluated_at") {
		t.Fatal("private clock was added to serialized result evidence")
	}
}

func TestA112MBUSReadAtUsesPrivateClockOnlyForDeadlineAndTokenValidation(t *testing.T) {
	fixed := time.Now()
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		for key, values := range a112MBUSGoodRateHeaders() {
			for _, value := range values {
				w.Header().Add(key, value)
			}
		}
		_, _ = w.Write([]byte(`{"result":{"nextBefore":null}}`))
	}))
	defer server.Close()
	client := a112MBUSTestClient(t, server)
	ctx := a112MBUSFixedDeadlineContext{Context: context.Background(), deadline: fixed.Add(15 * time.Second)}
	descriptor := a112MBUSDescriptor{path: "/api/v1/candles", query: url.Values{"symbol": {"AAPL"}, "interval": {"1m"}, "count": {"200"}, "adjusted": {"false"}}, requireCursor: true}
	result, err := a112MBUSReadAt(ctx, client, descriptor, func() time.Time { return fixed })
	if err != nil || string(result.Body()) != `{"result":{"nextBefore":null}}` {
		t.Fatalf("private clock read result=%+v err=%v", result, err)
	}
}

type a112MBUSFixedDeadlineContext struct {
	context.Context
	deadline time.Time
}

func (c a112MBUSFixedDeadlineContext) Deadline() (time.Time, bool) { return c.deadline, true }

var a112MBUSExportedSurface = map[string]struct{}{
	"A112MBUSResult": {}, "A112MBUSHoldError": {}, "A112MBUSCandle": {},
	"A112MBUSOrderbook": {}, "A112MBUSCalendar": {},
}

const a112MBUSOfficialImport = "github.com/JungHoonGhae/tossinvest-cli/internal/official"

func a112MBUSRepositoryForbiddenRefs(root string) ([]string, error) {
	var violations []string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() && (entry.Name() == ".git" || entry.Name() == "vendor") {
			return filepath.SkipDir
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, violation := range a112MBUSFindForbiddenRefs(rel, string(data)) {
			violations = append(violations, rel+": "+violation)
		}
		return nil
	})
	return violations, err
}

func a112MBUSFindForbiddenRefs(path, source string) []string {
	if a112MBUSReferenceAllowed(path) {
		return nil
	}
	file, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
	if err != nil {
		return []string{"parse error: " + err.Error()}
	}
	aliases := make(map[string]struct{})
	dotImport := false
	for _, spec := range file.Imports {
		importPath, err := strconv.Unquote(spec.Path.Value)
		if err != nil || importPath != a112MBUSOfficialImport {
			continue
		}
		name := "official"
		if spec.Name != nil {
			name = spec.Name.Name
		}
		switch name {
		case "_":
		case ".":
			dotImport = true
		default:
			aliases[name] = struct{}{}
		}
	}
	sameOfficialPackage := strings.HasPrefix(filepath.ToSlash(path), "internal/official/") && file.Name.Name == "official"
	if len(aliases) == 0 && !dotImport && !sameOfficialPackage {
		return nil
	}
	var violations []string
	selectorNames := make(map[token.Pos]struct{})
	ast.Inspect(file, func(node ast.Node) bool {
		if selector, ok := node.(*ast.SelectorExpr); ok {
			selectorNames[selector.Sel.Pos()] = struct{}{}
		}
		return true
	})
	ast.Inspect(file, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.SelectorExpr:
			ident, ok := typed.X.(*ast.Ident)
			if !ok {
				return true
			}
			if _, imported := aliases[ident.Name]; imported && a112MBUSSurfaceName(typed.Sel.Name) {
				violations = append(violations, "selector "+ident.Name+"."+typed.Sel.Name)
			}
		case *ast.Ident:
			if _, selected := selectorNames[typed.Pos()]; selected {
				return true
			}
			if sameOfficialPackage && (a112MBUSSurfaceName(typed.Name) || strings.HasPrefix(typed.Name, "a112MBUS")) {
				violations = append(violations, "same-package reference "+typed.Name)
				return true
			}
			if dotImport && a112MBUSSurfaceName(typed.Name) {
				violations = append(violations, "dot-import reference "+typed.Name)
			}
		}
		return true
	})
	return violations
}

func a112MBUSReferenceAllowed(path string) bool {
	path = filepath.ToSlash(path)
	if strings.HasPrefix(path, "tools/a112-mb-us-source/") {
		return true
	}
	if strings.HasPrefix(path, "internal/official/a112_mbus_") && strings.HasSuffix(path, "_test.go") {
		return true
	}
	return path == "internal/official/a112_mbus_read.go" || path == "internal/official/a112_mbus_read_unix.go" || path == "internal/official/a112_mbus_read_unsupported.go"
}

func a112MBUSSurfaceName(name string) bool {
	_, ok := a112MBUSExportedSurface[name]
	return ok
}
