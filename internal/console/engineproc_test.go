package console

// engineproc_test.go is the engine's half of the dashboard (openspec change
// add-engine-runtime, task 2.1).
//
// The claim under test is not "the button works". It is that the console cannot
// get past the engine's own startup interlock: it asks a seam to start a process,
// the process decides, and whatever the process decided is what the operator
// reads. A console that could report success over a refusal — or that answered
// the interlock question itself — would be the bypass the operator-console spec
// exists to forbid.

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/enginelock"
)

var engineNow = time.Date(2026, 7, 27, 12, 0, 0, 0, time.UTC)

// The status markup, matched exactly. The section's explanatory prose mentions
// both words, so a substring match on "실행 중" would pass whatever the status
// said — which is the assertion this file exists to make.
const (
	engineRunningMark = `<span class="ok">실행 중</span>`
	engineStoppedMark = `<span class="muted">정지</span>`
)

// engineHarness is newHarness with the engine seams wired and a marker path in
// the same temporary directory.
type engineHarness struct {
	*harness
	marker  string
	starts  int
	stops   int
	startFn func() (string, error)
	stopFn  func() (string, error)
}

func newEngineHarness(t *testing.T, tweak ...func(*Options)) *engineHarness {
	t.Helper()
	dir := t.TempDir()
	eh := &engineHarness{marker: enginelock.MarkerPath(dir)}
	eh.startFn = func() (string, error) { return "엔진을 시작했다 — pid 4242", nil }
	eh.stopFn = func() (string, error) { return "pid 4242에 SIGTERM을 보내 정상 종료시켰다", nil }

	opts := []func(*Options){func(o *Options) {
		o.EngineMarker = eh.marker
		o.StartEngine = func() (string, error) { eh.starts++; return eh.startFn() }
		o.StopEngine = func() (string, error) { eh.stops++; return eh.stopFn() }
		o.Now = func() time.Time { return engineNow }
	}}
	eh.harness = newHarness(t, append(opts, tweak...)...)
	return eh
}

// holdEngineMarker writes a marker as a live engine would.
func holdEngineMarker(t *testing.T, path string, at time.Time) {
	t.Helper()
	release, err := enginelock.Hold(context.Background(), path, at)
	if err != nil {
		t.Fatalf("enginelock.Hold: %v", err)
	}
	t.Cleanup(release)
}

// --- the status line ------------------------------------------------------------

// TestTheDashboardShowsWhetherTheEngineIsRunning is the SHALL: 엔진 상태(실행
// 여부·기동 거부 사유)는 대시보드에 표시한다.
func TestTheDashboardShowsWhetherTheEngineIsRunning(t *testing.T) {
	h := newEngineHarness(t)
	h.authenticate(t)

	page := body(t, h.get(t, "/"))
	if !strings.Contains(page, "엔진 런타임") {
		t.Fatal("the dashboard has no engine section")
	}
	if !strings.Contains(page, engineStoppedMark) {
		t.Errorf("with no marker the dashboard does not say the engine is stopped:\n%s", page)
	}

	holdEngineMarker(t, h.marker, engineNow)
	page = body(t, h.get(t, "/"))
	if !strings.Contains(page, engineRunningMark) {
		t.Errorf("a fresh marker is not reported as a running engine:\n%s", page)
	}
	if !strings.Contains(page, "pid "+strconv.Itoa(os.Getpid())) {
		t.Errorf("the running engine's pid is not shown:\n%s", page)
	}
}

// TestAStaleMarkerReadsAsStopped. The marker is advisory and its whole failure
// mode is a crashed engine: after the staleness window the dashboard stops
// claiming one is alive.
func TestAStaleMarkerReadsAsStopped(t *testing.T) {
	h := newEngineHarness(t)
	h.authenticate(t)
	holdEngineMarker(t, h.marker, engineNow.Add(-2*enginelock.StaleAfter))

	page := body(t, h.get(t, "/"))
	if strings.Contains(page, engineRunningMark) {
		t.Errorf("a stale marker still reads as a running engine:\n%s", page)
	}
}

// TestTheEngineSectionSaysWhenItIsUnwired rather than reporting "stopped", which
// would be a claim the console cannot substantiate.
func TestTheEngineSectionSaysWhenItIsUnwired(t *testing.T) {
	h := newHarness(t) // no EngineMarker, no seams
	h.authenticate(t)

	page := body(t, h.get(t, "/"))
	if !strings.Contains(page, "엔진 상태 배선이 없다") {
		t.Errorf("an unwired console does not say so:\n%s", page)
	}
	if strings.Contains(page, `action="/engine/start"`) {
		t.Error("the start button is drawn with no seam behind it")
	}
}

// --- the interlock cannot be bypassed --------------------------------------------

// TestARefusedStartShowsTheEnginesOwnReason is the scenario "인터록 미충족
// 상태의 엔진 시작 버튼": the engine process refuses and the clauses it
// enumerated are what the dashboard shows.
//
// The reason is the engine's, verbatim. The console does not evaluate the
// interlock, does not know what a clause is, and cannot report a start that did
// not happen as one that did.
func TestARefusedStartShowsTheEnginesOwnReason(t *testing.T) {
	h := newEngineHarness(t)
	h.authenticate(t)
	h.startFn = func() (string, error) {
		return "  - 9. 브로커측 보호 실행(ProtectionReady) — 이 빌드에는 없다 [2c 소관]",
			errors.New("기동 인터록 미충족: 루프를 하나도 시작하지 않았다")
	}

	resp := h.post(t, "/engine/start", url.Values{"csrf": {h.csrf}})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /engine/start returned %d", resp.StatusCode)
	}
	if h.starts != 1 {
		t.Fatalf("the start seam was called %d times", h.starts)
	}

	page := body(t, h.get(t, "/"))
	if !strings.Contains(page, "기동 인터록 미충족") {
		t.Errorf("the dashboard does not show the engine's refusal:\n%s", page)
	}
	if !strings.Contains(page, "ProtectionReady") {
		t.Errorf("the dashboard drops the unmet clause the engine enumerated:\n%s", page)
	}
	// And it did not claim a running engine.
	if strings.Contains(page, engineRunningMark) {
		t.Error("a refused start was reported as a running engine")
	}
}

// TestTheConsoleDecidesNothingAboutTheGate. The start seam is the only path, so
// there is no branch here that could conclude "the gate is fine, start it
// anyway": whatever the seam answers is the answer.
func TestTheConsoleDecidesNothingAboutTheGate(t *testing.T) {
	src := packageFiles(t)
	for name, file := range src {
		code := strings.Join(nonCommentLines(file), "\n")
		for _, banned := range []string{
			"AutomationGate", "Interlock", "ProtectionReady", "automation_gate",
		} {
			if strings.Contains(code, banned) {
				t.Errorf("%s names %q; the console asks the engine process and displays its answer, "+
					"it does not evaluate the gate", name, banned)
			}
		}
	}
}

// --- the buttons --------------------------------------------------------------------

// TestTheEngineButtonsNeedTheSessionAndTheCSRFToken. Both gates, from both sides.
func TestTheEngineButtonsNeedTheSessionAndTheCSRFToken(t *testing.T) {
	h := newEngineHarness(t)

	// No session: refused before anything is called.
	for _, path := range []string{"/engine/start", "/engine/stop"} {
		resp := h.post(t, path, url.Values{"csrf": {h.csrf}})
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("POST %s without a session returned %d, want 403", path, resp.StatusCode)
		}
	}
	if h.starts != 0 || h.stops != 0 {
		t.Fatalf("a request with no session reached the seams (%d start, %d stop)", h.starts, h.stops)
	}

	h.authenticate(t)

	// Session but no CSRF token.
	for _, path := range []string{"/engine/start", "/engine/stop"} {
		resp := h.post(t, path, url.Values{"csrf": {"wrong"}})
		if resp.StatusCode != http.StatusForbidden {
			t.Errorf("POST %s with a bad CSRF token returned %d, want 403", path, resp.StatusCode)
		}
	}
	if h.starts != 0 || h.stops != 0 {
		t.Fatalf("a request with a bad CSRF token reached the seams (%d start, %d stop)", h.starts, h.stops)
	}

	// GET on a state-changing route.
	for _, path := range []string{"/engine/start", "/engine/stop"} {
		resp := h.get(t, path)
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("GET %s returned %d, want 405", path, resp.StatusCode)
		}
	}
	if h.starts != 0 || h.stops != 0 {
		t.Fatalf("a GET reached the seams (%d start, %d stop)", h.starts, h.stops)
	}
}

// TestStoppingGoesThroughTheSignalDiscipline. The console holds no process API;
// what it can assert is that the stop it asks for is the one cmd/tossctl
// implements, and that its answer reaches the page.
func TestStoppingGoesThroughTheSignalDiscipline(t *testing.T) {
	h := newEngineHarness(t)
	h.authenticate(t)

	if resp := h.post(t, "/engine/stop", url.Values{"csrf": {h.csrf}}); resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /engine/stop returned %d", resp.StatusCode)
	}
	if h.stops != 1 {
		t.Fatalf("the stop seam was called %d times", h.stops)
	}
	page := body(t, h.get(t, "/"))
	if !strings.Contains(page, "SIGTERM") {
		t.Errorf("the stop's answer did not reach the dashboard:\n%s", page)
	}
}

// TestAFailedStopIsReportedAndNothingIsClaimed.
func TestAFailedStopIsReportedAndNothingIsClaimed(t *testing.T) {
	h := newEngineHarness(t)
	h.authenticate(t)
	h.stopFn = func() (string, error) {
		return "", errors.New("pid 4242이 30s 안에 종료되지 않았다")
	}

	if resp := h.post(t, "/engine/stop", url.Values{"csrf": {h.csrf}}); resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /engine/stop returned %d", resp.StatusCode)
	}
	page := body(t, h.get(t, "/"))
	if !strings.Contains(page, "엔진 정지 실패") {
		t.Errorf("a failed stop was not reported:\n%s", page)
	}
}

// TestAnUnwiredButtonSaysSoRatherThanPretending.
func TestAnUnwiredButtonSaysSoRatherThanPretending(t *testing.T) {
	h := newHarness(t, func(o *Options) {
		o.EngineMarker = filepath.Join(t.TempDir(), enginelock.MarkerFileName)
	})
	h.authenticate(t)

	resp := h.post(t, "/engine/start", url.Values{"csrf": {h.csrf}})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /engine/start returned %d", resp.StatusCode)
	}
	page := body(t, h.get(t, "/"))
	if !strings.Contains(page, "엔진 기동/정지 배선이 없다") {
		t.Errorf("an unwired start did not say so:\n%s", page)
	}
}

// TestTheStaleEngineBinaryIsWarnedAbout — task 2.1's third status item, answered
// from the engine's own statement about itself in the marker.
func TestTheStaleEngineBinaryIsWarnedAbout(t *testing.T) {
	h := newEngineHarness(t, func(o *Options) {
		// The "installed" binary is a file this process was certainly not loaded
		// from, so the marker's own stamp cannot match it.
		o.Binary = func() (binstamp.Stamp, error) {
			return binstamp.Stamp{Path: "/usr/local/bin/tossctl", Size: 1, ModTime: engineNow}, nil
		}
	})
	h.authenticate(t)
	holdEngineMarker(t, h.marker, engineNow)

	page := body(t, h.get(t, "/"))
	if !strings.Contains(page, "실행 중인 엔진은 설치된 바이너리보다 오래되었다") {
		t.Errorf("a running engine on a different build is not warned about:\n%s", page)
	}
}
