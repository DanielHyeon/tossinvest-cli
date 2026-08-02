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

	page := body(t, h.get(t, pathVerifyConsole))
	if !strings.Contains(page, "엔진 런타임") {
		t.Fatal("the dashboard has no engine section")
	}
	if !strings.Contains(page, engineStoppedMark) {
		t.Errorf("with no marker the dashboard does not say the engine is stopped:\n%s", page)
	}

	holdEngineMarker(t, h.marker, engineNow)
	page = body(t, h.get(t, pathVerifyConsole))
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

	page := body(t, h.get(t, pathVerifyConsole))
	if strings.Contains(page, engineRunningMark) {
		t.Errorf("a stale marker still reads as a running engine:\n%s", page)
	}
}

// TestTheEngineSectionSaysWhenItIsUnwired rather than reporting "stopped", which
// would be a claim the console cannot substantiate.
func TestTheEngineSectionSaysWhenItIsUnwired(t *testing.T) {
	h := newHarness(t) // no EngineMarker, no seams
	h.authenticate(t)

	page := body(t, h.get(t, pathVerifyConsole))
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

	page := body(t, h.get(t, pathVerifyConsole))
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
//
// The two lists below say different things, and the split is the point.
//
// "Interlock" and "ProtectionReady" name the engine's startup DECISION and stay
// banned in every file. Nothing in this package may compute whether the engine
// is allowed to run; it asks and prints the answer.
//
// "AutomationGate" and "automation_gate" name the config BLOCK. Since
// console-sets-guardian-limits the settings screen edits the five ceilings
// inside it, and since console-owns-the-operating-toggles it also writes the
// switch — so those spellings are permitted in the files that do the editing and
// nowhere else.
//
// The switch moving into the console does not weaken the ban that matters. The
// two words still banned everywhere are the engine's *decision*: nothing here
// may compute whether the engine is allowed to run. The gate section renders a
// pre-flight, and that pre-flight reads config values (`Limits().Validate()`,
// the four trading toggles) rather than reproducing the interlock — it says so
// on the screen, names the clauses it cannot judge, and refuses to claim a start
// is guaranteed. Saving ON with a failing pre-flight is allowed for exactly this
// reason: the console records a choice, the engine judges it.
func TestTheConsoleDecidesNothingAboutTheGate(t *testing.T) {
	// The editors and the page struct that carries the block to the template.
	// Byte-for-byte file names: a prefix or suffix rule here would let a
	// settings_limits_helper.go inherit the exemption without arguing for it.
	//
	// settings_operating.go is NOT here and that is worth noticing: the file that
	// writes the switch never spells the key. Its seam takes a bool and the key
	// lives in internal/config's closed member list, which is the same separation
	// the limit path has — the console names a capability, not a config byte.
	mayNameTheBlock := map[string]bool{
		"settings.go":           true,
		"settings_limits.go":    true,
		"templates_settings.go": true,
	}
	src := packageFiles(t)
	for name, file := range src {
		code := strings.Join(nonCommentLines(file), "\n")
		banned := []string{"Interlock", "ProtectionReady"}
		if !mayNameTheBlock[name] {
			banned = append(banned, "AutomationGate", "automation_gate")
		}
		for _, word := range banned {
			if strings.Contains(code, word) {
				t.Errorf("%s names %q; the console asks the engine process and displays its answer, "+
					"it does not evaluate the gate", name, word)
			}
		}
	}
}

// TestTheGateEditingExemptionIsNotIdle guards the guard: an exemption for a file
// that no longer needs it is an exemption nobody will notice being used.
func TestTheGateEditingExemptionIsNotIdle(t *testing.T) {
	src := packageFiles(t)
	// Each exempt file, with the spelling that earns it. settings_operating.go
	// and templates_settings.go earn theirs with the snake-case key rather than
	// the Go type: one writes `engine.automation_gate.enabled` through its seam,
	// the other renders the section that does.
	for name, spelling := range map[string]string{
		"settings.go":           "AutomationGate",
		"settings_limits.go":    "AutomationGate",
		"templates_settings.go": "automation_gate",
	} {
		file, ok := src[name]
		if !ok {
			t.Errorf("%s is exempt from the gate-naming ban but is not in the package", name)
			continue
		}
		if !strings.Contains(strings.Join(nonCommentLines(file), "\n"), spelling) {
			t.Errorf("%s no longer names %q; drop its exemption", name, spelling)
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
	page := body(t, h.get(t, pathVerifyConsole))
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
	page := body(t, h.get(t, pathVerifyConsole))
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
	page := body(t, h.get(t, pathVerifyConsole))
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

	page := body(t, h.get(t, pathVerifyConsole))
	if !strings.Contains(page, "실행 중인 엔진은 설치된 바이너리보다 오래되었다") {
		t.Errorf("a running engine on a different build is not warned about:\n%s", page)
	}
}
