package console

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/releaseupdate"
)

func signedReleaseHarness(
	t *testing.T,
	updater *fakeSystemUpdater,
	downloader *fakeReleaseDownloader,
) *harness {
	t.Helper()
	return updateHarness(t, updater, func(o *Options) {
		o.ReleaseDownloader = downloader
		o.ReleaseCandidateStager = updater
	})
}

func TestSettingsRendersSignedReleaseActionWithoutFetching(t *testing.T) {
	updater := &fakeSystemUpdater{view: validUpdateView()}
	downloader := &fakeReleaseDownloader{release: validSignedRelease()}
	h := signedReleaseHarness(t, updater, downloader)
	h.authenticate(t)
	page := body(t, h.get(t, "/settings"))
	for _, want := range []string{
		`action="/settings/system-update/download"`,
		"서명된 최신 릴리스 확인·다운로드",
		"출처 확인 안 됨",
	} {
		if !strings.Contains(page, want) {
			t.Errorf("settings page missing %q", want)
		}
	}
	if downloader.calls != 0 {
		t.Fatalf("GET settings fetched a release %d time(s)", downloader.calls)
	}
}

func TestSignedReleaseDownloadRequiresSessionPostAndCSRF(t *testing.T) {
	updater := &fakeSystemUpdater{view: validUpdateView()}
	downloader := &fakeReleaseDownloader{release: validSignedRelease()}
	h := signedReleaseHarness(t, updater, downloader)

	form := url.Values{"csrf": {h.csrf}}
	if got := h.post(t, "/settings/system-update/download", form).StatusCode; got != http.StatusForbidden {
		t.Fatalf("without session = %d", got)
	}
	h.authenticate(t)
	if got := h.get(t, "/settings/system-update/download").StatusCode; got != http.StatusMethodNotAllowed {
		t.Fatalf("GET = %d", got)
	}
	if got := h.post(t, "/settings/system-update/download", url.Values{}).StatusCode; got != http.StatusForbidden {
		t.Fatalf("without CSRF = %d", got)
	}
	if downloader.calls != 0 || updater.stages != 0 {
		t.Fatalf("refusal reached downloader=%d stager=%d", downloader.calls, updater.stages)
	}
}

func TestSignedReleaseDownloadIgnoresSelectorsStagesButNeverInstallsOrRelaunches(t *testing.T) {
	updater := &fakeSystemUpdater{view: validUpdateView()}
	downloader := &fakeReleaseDownloader{release: validSignedRelease()}
	h := signedReleaseHarness(t, updater, downloader)
	h.authenticate(t)

	resp := h.post(t, "/settings/system-update/download", url.Values{
		"csrf":        {h.csrf},
		"repository":  {"attacker/repo"},
		"url":         {"https://attacker.invalid/payload"},
		"tag":         {"v9.9.9"},
		"path":        {"/tmp/attacker"},
		"destination": {"/usr/local/bin/tossctl"},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("download response = %d", resp.StatusCode)
	}
	if downloader.calls != 1 || updater.stages != 1 {
		t.Fatalf("calls downloader=%d stage=%d", downloader.calls, updater.stages)
	}
	if string(updater.staged) != "verified-binary" ||
		updater.revision != strings.Repeat("a", 40) {
		t.Fatalf("staged=%q revision=%q", updater.staged, updater.revision)
	}
	if updater.installs != 0 {
		t.Fatalf("download installed %d time(s)", updater.installs)
	}
	h.noRelaunch(t)
	page := body(t, resp)
	for _, want := range []string{
		"v1.2.3", "archive-sha", "staged-sha", "release.yml", strings.Repeat("a", 40),
	} {
		if !strings.Contains(page, want) {
			t.Errorf("result page missing %q", want)
		}
	}
}

func TestSignedReleaseFailuresNeverInstallOrRelaunch(t *testing.T) {
	t.Run("verification", func(t *testing.T) {
		updater := &fakeSystemUpdater{view: validUpdateView()}
		downloader := &fakeReleaseDownloader{err: errors.New("verification refused")}
		h := signedReleaseHarness(t, updater, downloader)
		h.authenticate(t)

		resp := h.post(t, "/settings/system-update/download",
			url.Values{"csrf": {h.csrf}})
		if updater.stages != 0 || updater.installs != 0 {
			t.Fatalf("verification failure reached stages=%d installs=%d",
				updater.stages, updater.installs)
		}
		h.noRelaunch(t)
		if page := body(t, resp); !strings.Contains(page, "기존 candidate는 유지됐다") {
			t.Fatalf("verification refusal did not explain preservation: %s", page)
		}
	})

	t.Run("staging", func(t *testing.T) {
		updater := &fakeSystemUpdater{
			view: validUpdateView(), stageErr: errors.New("publish refused"),
		}
		downloader := &fakeReleaseDownloader{release: validSignedRelease()}
		h := signedReleaseHarness(t, updater, downloader)
		h.authenticate(t)

		resp := h.post(t, "/settings/system-update/download",
			url.Values{"csrf": {h.csrf}})
		if updater.stages != 1 || updater.installs != 0 {
			t.Fatalf("staging failure calls stages=%d installs=%d",
				updater.stages, updater.installs)
		}
		h.noRelaunch(t)
		if page := body(t, resp); !strings.Contains(page, "기존 candidate는 유지됐다") {
			t.Fatalf("staging refusal did not explain preservation: %s", page)
		}
	})
}

func TestSignedReleaseDownloadRefusesPublishAfterInstallCommit(t *testing.T) {
	entered := make(chan struct{}, 1)
	continueFetch := make(chan struct{})
	updater := &fakeSystemUpdater{view: validUpdateView()}
	downloader := &fakeReleaseDownloader{
		release: validSignedRelease(), entered: entered, continueFetch: continueFetch,
	}
	h := signedReleaseHarness(t, updater, downloader)
	h.authenticate(t)

	done := make(chan *http.Response, 1)
	go func() {
		resp, _ := h.client.PostForm(h.srv.URL+"/settings/system-update/download",
			url.Values{"csrf": {h.csrf}})
		done <- resp
	}()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("download did not enter verification")
	}
	h.activityMu.Lock()
	h.updateCommitted = true
	h.activityMu.Unlock()
	close(continueFetch)
	select {
	case resp := <-done:
		if resp == nil {
			t.Fatal("request failed")
		}
		_ = resp.Body.Close()
	case <-time.After(5 * time.Second):
		t.Fatal("download did not finish")
	}
	if updater.stages != 0 {
		t.Fatal("old console published after install commit")
	}
}

func TestSignedReleaseChecksAreWholeOperationSingleFlight(t *testing.T) {
	entered := make(chan struct{}, 2)
	continueFetch := make(chan struct{})
	updater := &fakeSystemUpdater{view: validUpdateView()}
	first := validSignedRelease()
	first.Binary = []byte("first-tag")
	second := validSignedRelease()
	second.Tag = "v1.2.4"
	second.Binary = []byte("second-tag")
	downloader := &fakeReleaseDownloader{
		releases: []releaseupdate.Release{first, second},
		entered:  entered, continueFetch: continueFetch,
	}
	h := signedReleaseHarness(t, updater, downloader)
	h.authenticate(t)

	done := make(chan struct{}, 2)
	post := func() {
		resp, _ := h.client.PostForm(h.srv.URL+"/settings/system-update/download",
			url.Values{"csrf": {h.csrf}})
		if resp != nil {
			_ = resp.Body.Close()
		}
		done <- struct{}{}
	}
	go post()
	select {
	case <-entered:
	case <-time.After(5 * time.Second):
		t.Fatal("first fetch did not start")
	}
	go post()
	select {
	case <-entered:
		t.Fatal("second fetch entered while first owned release single-flight")
	case <-time.After(100 * time.Millisecond):
	}
	close(continueFetch)
	for range 2 {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("request did not finish")
		}
	}
	downloader.mu.Lock()
	maxActive := downloader.maxActive
	downloader.mu.Unlock()
	if maxActive != 1 || updater.stages != 2 {
		t.Fatalf("max active=%d stages=%d", maxActive, updater.stages)
	}
	if string(updater.staged) != "second-tag" {
		t.Fatalf("final staged bytes = %q; serialized newer request did not win",
			updater.staged)
	}
}
