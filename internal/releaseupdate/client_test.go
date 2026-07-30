package releaseupdate

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
)

type fakeVerifier struct {
	mu       sync.Mutex
	calls    int
	wantTag  string
	wantName string
	wantSHA  string
	result   Provenance
	err      error
}

func (f *fakeVerifier) Verify(_ context.Context, bundle []byte, digest, tag, asset string) (Provenance, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	if string(bundle) != `{"bundle":"ok"}` {
		return Provenance{}, errors.New("unexpected bundle")
	}
	if digest != f.wantSHA || tag != f.wantTag || asset != f.wantName {
		return Provenance{}, errors.New("verification inputs were not fixed to release")
	}
	return f.result, f.err
}

func releaseArchive(t *testing.T, binary []byte) []byte {
	t.Helper()
	var compressed bytes.Buffer
	gz := gzip.NewWriter(&compressed)
	tw := tar.NewWriter(gz)
	entries := []struct {
		name string
		mode int64
		body []byte
		kind byte
	}{
		{name: "tossctl", mode: 0o755, body: binary, kind: tar.TypeReg},
		{name: "auth-helper/", mode: 0o755, kind: tar.TypeDir},
		{name: "auth-helper/manifest.json", mode: 0o644, body: []byte("{}"), kind: tar.TypeReg},
	}
	for _, entry := range entries {
		if err := tw.WriteHeader(&tar.Header{
			Name: entry.name, Mode: entry.mode, Size: int64(len(entry.body)), Typeflag: entry.kind,
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write(entry.body); err != nil {
			t.Fatal(err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gz.Close(); err != nil {
		t.Fatal(err)
	}
	return compressed.Bytes()
}

func newReleaseFixture(t *testing.T, currentVersion, tag string, binary []byte) (*Client, *fakeVerifier) {
	t.Helper()
	archive := releaseArchive(t, binary)
	sum := sha256.Sum256(archive)
	sha := hex.EncodeToString(sum[:])
	verifier := &fakeVerifier{
		wantTag:  tag,
		wantName: "tossctl-linux-amd64.tar.gz",
		wantSHA:  sha,
		result: Provenance{
			WorkflowIdentity: "https://github.com/JungHoonGhae/tossinvest-cli/.github/workflows/release.yml@refs/tags/" + tag,
			SourceCommit:     strings.Repeat("a", 40),
		},
	}
	var server *httptest.Server
	server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/repos/JungHoonGhae/tossinvest-cli/releases/latest":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"tag_name": tag, "draft": false, "prerelease": false,
				"assets": []map[string]any{{
					"name": "tossctl-linux-amd64.tar.gz", "size": len(archive),
					"browser_download_url": server.URL + "/asset",
				}},
			})
		case r.URL.Path == "/asset":
			_, _ = w.Write(archive)
		case strings.HasPrefix(r.URL.Path, "/repos/JungHoonGhae/tossinvest-cli/attestations/sha256:"):
			if got := r.URL.Query().Get("predicate_type"); got != "provenance" {
				t.Errorf("predicate_type = %q", got)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"attestations": []map[string]any{{"bundle_url": server.URL + "/attestations/1"}},
			})
		case r.URL.Path == "/attestations/1":
			_, _ = io.WriteString(w, `{"bundle":"ok"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)
	base, err := url.Parse(server.URL)
	if err != nil {
		t.Fatal(err)
	}
	client, err := newClient(clientConfig{
		apiBase:        server.URL,
		httpClient:     server.Client(),
		goos:           "linux",
		goarch:         "amd64",
		currentVersion: currentVersion,
		verifier:       verifier,
		allowURL: func(u *url.URL, kind requestKind) bool {
			return u.Scheme == "https" && u.Host == base.Host &&
				(kind != requestBundle || strings.HasPrefix(u.Path, "/attestations/"))
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return client, verifier
}

func TestFetchSelectsFixedLatestAssetAndVerifiesBeforeExtracting(t *testing.T) {
	client, verifier := newReleaseFixture(t, "1.0.0", "v1.1.0", []byte("binary"))
	got, err := client.Fetch(context.Background())
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if got.Tag != "v1.1.0" || got.AssetName != "tossctl-linux-amd64.tar.gz" ||
		string(got.Binary) != "binary" || got.ArchiveSHA256 == "" ||
		got.SourceCommit != strings.Repeat("a", 40) || got.Bootstrap {
		t.Fatalf("release = %+v", got)
	}
	if verifier.calls != 1 {
		t.Fatalf("verifier calls = %d", verifier.calls)
	}
}

func TestFetchRejectsOlderEqualDraftPrereleaseAndUnsupportedPlatform(t *testing.T) {
	for _, tc := range []struct {
		name, current, tag, goos string
	}{
		{name: "equal", current: "1.1.0", tag: "v1.1.0", goos: "linux"},
		{name: "older", current: "1.2.0", tag: "v1.1.0", goos: "linux"},
		{name: "bad tag", current: "1.0.0", tag: "release-1.1.0", goos: "linux"},
		{name: "windows", current: "1.0.0", tag: "v1.1.0", goos: "windows"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			client, verifier := newReleaseFixture(t, tc.current, tc.tag, []byte("binary"))
			client.goos = tc.goos
			if _, err := client.Fetch(context.Background()); err == nil {
				t.Fatal("Fetch succeeded")
			}
			if verifier.calls != 0 {
				t.Fatal("refused release reached verifier")
			}
		})
	}
}

func TestFetchAllowsExplicitDevelopmentBootstrap(t *testing.T) {
	client, _ := newReleaseFixture(t, "dev", "v1.1.0", []byte("binary"))
	got, err := client.Fetch(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !got.Bootstrap {
		t.Fatalf("release = %+v, want bootstrap", got)
	}
}

func TestFetchRejectsEscapingRedirectAndOversizedChunkedResponse(t *testing.T) {
	t.Run("redirect", func(t *testing.T) {
		client, _ := newReleaseFixture(t, "1.0.0", "v1.1.0", []byte("binary"))
		var evil *httptest.Server
		evil = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "https://example.com/escape", http.StatusFound)
		}))
		t.Cleanup(evil.Close)
		client.httpClient = evil.Client()
		client.apiBase = evil.URL
		allowed, _ := url.Parse(evil.URL)
		client.allowURL = func(u *url.URL, _ requestKind) bool {
			return u.Scheme == "https" && u.Host == allowed.Host
		}
		if _, err := client.Fetch(context.Background()); err == nil {
			t.Fatal("redirect escape succeeded")
		}
	})

	t.Run("chunked limit", func(t *testing.T) {
		client, verifier := newReleaseFixture(t, "1.0.0", "v1.1.0", []byte("binary"))
		client.limits.metadata = 8
		if _, err := client.Fetch(context.Background()); err == nil {
			t.Fatal("oversized metadata succeeded")
		}
		if verifier.calls != 0 {
			t.Fatal("oversized metadata reached verifier")
		}
	})
}

func TestProductionBundleURLIsExactAndNeverRedirected(t *testing.T) {
	for _, raw := range []string{
		"https://evil.blob.core.windows.net/attestations/1",
		"https://x.tmaproduction.blob.core.windows.net/attestations/1",
		"https://tmaproduction.blob.core.windows.net/not-attestations/1",
		"https://tmaproduction.blob.core.windows.net:444/attestations/1",
		"https://user@tmaproduction.blob.core.windows.net/attestations/1",
	} {
		u, err := url.Parse(raw)
		if err != nil {
			t.Fatal(err)
		}
		if productionURLAllowed(u, requestBundle) {
			t.Errorf("allowed bundle URL %s", raw)
		}
	}
	u, _ := url.Parse("https://tmaproduction.blob.core.windows.net/attestations/1")
	if !productionURLAllowed(u, requestBundle) {
		t.Fatal("exact production attestation URL was refused")
	}
}
