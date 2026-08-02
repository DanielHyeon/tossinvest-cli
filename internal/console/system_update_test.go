package console

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/localupdate"
	"github.com/JungHoonGhae/tossinvest-cli/internal/releaseupdate"
)

type fakeSystemUpdater struct {
	mu       sync.Mutex
	view     localupdate.Inspection
	result   localupdate.Result
	err      error
	installs int
	reviewed string
	stages   int
	staged   []byte
	revision string
	stageErr error
	entered  chan<- struct{}
	release  <-chan struct{}
}

func (f *fakeSystemUpdater) Inspect() localupdate.Inspection { return f.view }

func (f *fakeSystemUpdater) Install(reviewed string, guard func() error) (localupdate.Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.installs++
	f.reviewed = reviewed
	if guard != nil {
		if err := guard(); err != nil {
			return localupdate.Result{}, err
		}
	}
	if f.entered != nil {
		f.entered <- struct{}{}
	}
	if f.release != nil {
		<-f.release
	}
	return f.result, f.err
}

func (f *fakeSystemUpdater) StageCandidate(reader io.Reader, expectedRevision string) (localupdate.StageResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stages++
	f.revision = expectedRevision
	f.staged, _ = io.ReadAll(reader)
	if f.stageErr != nil {
		return localupdate.StageResult{}, f.stageErr
	}
	return localupdate.StageResult{Metadata: localupdate.Metadata{
		Path: "/home/daniel/.local/bin/tossctl.candidate", SHA256: "staged-sha",
	}}, nil
}

type fakeReleaseDownloader struct {
	mu            sync.Mutex
	release       releaseupdate.Release
	releases      []releaseupdate.Release
	err           error
	calls         int
	active        int
	maxActive     int
	entered       chan<- struct{}
	continueFetch <-chan struct{}
}

func (f *fakeReleaseDownloader) Fetch(context.Context) (releaseupdate.Release, error) {
	f.mu.Lock()
	release := f.release
	if f.calls < len(f.releases) {
		release = f.releases[f.calls]
	}
	f.calls++
	f.active++
	if f.active > f.maxActive {
		f.maxActive = f.active
	}
	f.mu.Unlock()
	if f.entered != nil {
		f.entered <- struct{}{}
	}
	if f.continueFetch != nil {
		<-f.continueFetch
	}
	f.mu.Lock()
	f.active--
	f.mu.Unlock()
	return release, f.err
}

func validSignedRelease() releaseupdate.Release {
	return releaseupdate.Release{
		Tag:              "v1.2.3",
		AssetName:        "tossctl-linux-amd64.tar.gz",
		ArchiveSHA256:    "archive-sha",
		WorkflowIdentity: "https://github.com/JungHoonGhae/tossinvest-cli/.github/workflows/release.yml@refs/tags/v1.2.3",
		SourceCommit:     strings.Repeat("a", 40),
		Binary:           []byte("verified-binary"),
	}
}

func validUpdateView() localupdate.Inspection {
	return localupdate.Inspection{
		Current: localupdate.Metadata{
			Path: "/home/daniel/.local/bin/tossctl", Size: 10, SHA256: "old",
			ModTime: time.Date(2026, 7, 30, 1, 0, 0, 0, time.UTC),
		},
		Candidate: localupdate.Metadata{
			Path: "/home/daniel/.local/bin/tossctl.candidate", Size: 11, SHA256: "new-reviewed",
			ModTime:   time.Date(2026, 7, 30, 2, 0, 0, 0, time.UTC),
			GoVersion: "go1.26.5", ModulePath: "github.com/JungHoonGhae/tossinvest-cli",
			ModuleVersion: "(devel)",
			CommandPath:   "github.com/JungHoonGhae/tossinvest-cli/cmd/tossctl",
			GOOS:          "linux", GOARCH: "amd64",
		},
		CandidatePath: "/home/daniel/.local/bin/tossctl.candidate",
		Installable:   true,
	}
}

func updateHarness(t *testing.T, updater *fakeSystemUpdater, tweak ...func(*Options)) *harness {
	t.Helper()
	return newHarness(t, func(o *Options) {
		o.SystemUpdater = updater
		o.AcquireUpdateEngineLock = func() (func(), error) { return func() {}, nil }
		o.CheckUpdateVerifyActivity = func() error { return nil }
		o.Relaunch = func(int) error { return nil }
		for _, f := range tweak {
			f(o)
		}
	})
}

func TestSettingsRendersTheFixedReviewedSystemUpdate(t *testing.T) {
	updater := &fakeSystemUpdater{view: validUpdateView()}
	h := updateHarness(t, updater)
	h.authenticate(t)
	page := body(t, h.get(t, pathSettingsTools))
	for _, want := range []string{
		"시스템 업데이트",
		"/home/daniel/.local/bin/tossctl",
		"/home/daniel/.local/bin/tossctl.candidate",
		"new-reviewed",
		"github.com/JungHoonGhae/tossinvest-cli",
		`name="reviewed_sha256" value="new-reviewed"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("settings page does not contain %q", want)
		}
	}
	if strings.Contains(page, `name="path"`) || strings.Contains(page, `name="command"`) {
		t.Fatal("system update form accepts a path or command")
	}
}

func TestSystemUpdateInstallRequiresSessionPostAndCSRF(t *testing.T) {
	updater := &fakeSystemUpdater{view: validUpdateView()}
	h := updateHarness(t, updater)
	h.pretendListening("127.0.0.1:45678")

	if got := h.post(t, "/settings/system-update/install",
		url.Values{"csrf": {h.csrf}, "reviewed_sha256": {"new-reviewed"}}).StatusCode; got != http.StatusForbidden {
		t.Fatalf("without session = %d, want 403", got)
	}
	h.authenticate(t)
	if got := h.get(t, "/settings/system-update/install").StatusCode; got != http.StatusMethodNotAllowed {
		t.Fatalf("GET = %d, want 405", got)
	}
	if got := h.post(t, "/settings/system-update/install",
		url.Values{"reviewed_sha256": {"new-reviewed"}}).StatusCode; got != http.StatusForbidden {
		t.Fatalf("without CSRF = %d, want 403", got)
	}
	if updater.installs != 0 {
		t.Fatalf("refused requests reached installer %d time(s)", updater.installs)
	}
}

func TestSystemUpdateInstallIgnoresPathAndRequestsSamePortRelaunch(t *testing.T) {
	updater := &fakeSystemUpdater{
		view: validUpdateView(),
		result: localupdate.Result{
			OldSHA256: "old", NewSHA256: "new-reviewed",
			RollbackPath: "/home/daniel/.local/bin/tossctl.rollback",
		},
	}
	h := updateHarness(t, updater)
	h.pretendListening("127.0.0.1:45678")
	h.authenticate(t)

	resp := h.post(t, "/settings/system-update/install", url.Values{
		"csrf":            {h.csrf},
		"reviewed_sha256": {"new-reviewed"},
		"path":            {"/tmp/attacker"},
		"command":         {"rm -rf /"},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("install = %d", resp.StatusCode)
	}
	if updater.installs != 1 || updater.reviewed != "new-reviewed" {
		t.Fatalf("installer calls=%d reviewed=%q", updater.installs, updater.reviewed)
	}
	if page := body(t, resp); !strings.Contains(page, "업데이트") ||
		!strings.Contains(page, "new-reviewed") {
		t.Fatalf("interstitial does not describe update:\n%s", page)
	}
	if got := h.awaitRelaunch(t); got != 45678 {
		t.Fatalf("relaunch port = %d, want 45678", got)
	}
}

func TestSystemUpdateRefusalsDoNotInstallOrRelaunch(t *testing.T) {
	cases := []struct {
		name  string
		tweak func(*Options)
	}{
		{"engine lock", func(o *Options) {
			o.AcquireUpdateEngineLock = func() (func(), error) {
				return nil, errors.New("engine already running")
			}
		}},
		{"verification evidence", func(o *Options) {
			o.CheckUpdateVerifyActivity = func() error { return errors.New("verification active") }
		}},
		{"relaunch unwired", func(o *Options) { o.Relaunch = nil }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			updater := &fakeSystemUpdater{view: validUpdateView()}
			h := updateHarness(t, updater, tc.tweak)
			h.pretendListening("127.0.0.1:45678")
			h.authenticate(t)
			resp := h.post(t, "/settings/system-update/install", url.Values{
				"csrf": {h.csrf}, "reviewed_sha256": {"new-reviewed"},
			})
			if resp.StatusCode != http.StatusOK {
				t.Fatalf("refusal response = %d", resp.StatusCode)
			}
			if updater.installs != 0 {
				t.Fatalf("refusal reached installer %d time(s)", updater.installs)
			}
			h.noRelaunch(t)
		})
	}
}

func TestSystemUpdateRefusesAnInProcessVerificationAndInstallFailureDoesNotRelaunch(t *testing.T) {
	t.Run("verification running", func(t *testing.T) {
		updater := &fakeSystemUpdater{view: validUpdateView()}
		h := updateHarness(t, updater)
		h.pretendListening("127.0.0.1:45678")
		h.startAndWait(t)

		resp := h.post(t, "/settings/system-update/install", url.Values{
			"csrf": {h.csrf}, "reviewed_sha256": {"new-reviewed"},
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("refusal response = %d", resp.StatusCode)
		}
		if updater.installs != 0 {
			t.Fatalf("running verification reached installer %d time(s)", updater.installs)
		}
		h.noRelaunch(t)
	})

	t.Run("installer failure", func(t *testing.T) {
		updater := &fakeSystemUpdater{
			view: validUpdateView(),
			err:  errors.New("injected replacement failure"),
		}
		h := updateHarness(t, updater)
		h.pretendListening("127.0.0.1:45678")
		h.authenticate(t)

		resp := h.post(t, "/settings/system-update/install", url.Values{
			"csrf": {h.csrf}, "reviewed_sha256": {"new-reviewed"},
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("failure response = %d", resp.StatusCode)
		}
		if updater.installs != 1 {
			t.Fatalf("installer calls = %d, want 1", updater.installs)
		}
		h.noRelaunch(t)
	})
}

func TestSuccessfulUpdateBlocksNewEngineAndVerificationStartsUntilOldConsoleExits(t *testing.T) {
	updater := &fakeSystemUpdater{view: validUpdateView(), result: localupdate.Result{
		OldSHA256: "old", NewSHA256: "new-reviewed", RollbackPath: "/tmp/tossctl.rollback",
	}}
	engineStarts := 0
	h := updateHarness(t, updater, func(o *Options) {
		o.StartEngine = func() (string, error) { engineStarts++; return "started", nil }
	})
	h.pretendListening("127.0.0.1:45678")
	h.authenticate(t)
	_ = h.post(t, "/settings/system-update/install", url.Values{
		"csrf": {h.csrf}, "reviewed_sha256": {"new-reviewed"},
	})
	if got := h.post(t, "/engine/start", url.Values{"csrf": {h.csrf}}).StatusCode; got != http.StatusOK {
		t.Fatalf("engine start refusal response = %d", got)
	}
	if engineStarts != 0 {
		t.Fatal("engine started after update commit")
	}
	if got := h.post(t, "/verify/start", url.Values{"csrf": {h.csrf}}).StatusCode; got != http.StatusOK {
		t.Fatalf("verify start refusal response = %d", got)
	}
	if run := h.currentRun(); run != nil {
		t.Fatal("verification started after update commit")
	}
}

func TestSystemUpdateSerializesAConcurrentEngineStartThroughCommit(t *testing.T) {
	entered := make(chan struct{}, 1)
	release := make(chan struct{})
	updater := &fakeSystemUpdater{
		view:    validUpdateView(),
		result:  localupdate.Result{OldSHA256: "old", NewSHA256: "new-reviewed"},
		entered: entered,
		release: release,
	}
	engineStarts := 0
	h := updateHarness(t, updater, func(o *Options) {
		o.StartEngine = func() (string, error) { engineStarts++; return "started", nil }
	})
	h.pretendListening("127.0.0.1:45678")
	h.authenticate(t)

	type response struct {
		status int
		err    error
	}
	post := func(path string, form url.Values, done chan<- response) {
		resp, err := h.client.PostForm(h.srv.URL+path, form)
		if err != nil {
			done <- response{err: err}
			return
		}
		_ = resp.Body.Close()
		done <- response{status: resp.StatusCode}
	}

	updateDone := make(chan response, 1)
	go post("/settings/system-update/install", url.Values{
		"csrf": {h.csrf}, "reviewed_sha256": {"new-reviewed"},
	}, updateDone)
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("installer did not reach its commit interval")
	}

	engineDone := make(chan response, 1)
	go post("/engine/start", url.Values{"csrf": {h.csrf}}, engineDone)
	select {
	case got := <-engineDone:
		t.Fatalf("engine start escaped update serialization before commit: %+v", got)
	case <-time.After(100 * time.Millisecond):
	}

	close(release)
	for name, done := range map[string]<-chan response{
		"update": updateDone,
		"engine": engineDone,
	} {
		select {
		case got := <-done:
			if got.err != nil || got.status != http.StatusOK {
				t.Fatalf("%s response = %+v", name, got)
			}
		case <-time.After(5 * time.Second):
			t.Fatalf("%s response did not finish", name)
		}
	}
	if engineStarts != 0 {
		t.Fatal("engine started across the update commit")
	}
}
