package console

// restart_test.go covers task 1.8 ①②: the console restarting itself, the browser
// following it across, and the soak being restarted from the same page.
//
// Nothing here forks a process. Relaunch and RestartSoak are functions the console
// is handed, and every test hands it one that records the call — which is the whole
// reason they are seams. The handoff, by contrast, is the real internal/handoff
// store on a real temporary file, because "single use" is a property of a file and
// a fake would only assert that the fake is single use.

import (
	"errors"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/binstamp"
	"github.com/JungHoonGhae/tossinvest-cli/internal/handoff"
)

// --- helpers -------------------------------------------------------------------

// relaunchSpy is the re-exec seam, recording instead of executing.
type relaunchSpy struct {
	calls []int
	err   error
}

func (s *relaunchSpy) fn() Relaunch {
	return func(port int) error {
		s.calls = append(s.calls, port)
		return s.err
	}
}

// pretendListening gives the console an address without opening a socket, which is
// what the handler reads the port off.
func (h *harness) pretendListening(addr string) {
	h.mu.Lock()
	h.addr = addr
	h.mu.Unlock()
}

// awaitRelaunch reports the port a restart asked Serve for.
func (h *harness) awaitRelaunch(t *testing.T) int {
	t.Helper()
	select {
	case port := <-h.relaunch:
		return port
	case <-time.After(2 * time.Second):
		t.Fatal("no relaunch was requested")
		return 0
	}
}

func (h *harness) noRelaunch(t *testing.T) {
	t.Helper()
	select {
	case port := <-h.relaunch:
		t.Fatalf("a relaunch was requested (port %d); nothing should have been", port)
	case <-time.After(50 * time.Millisecond):
	}
}

// --- the gates ------------------------------------------------------------------

// TestTheRestartRoutesAreBehindBothGates.
//
// A restart is not dangerous the way an order is, but it is an act: a page that
// could restart this console by embedding an image would be able to interrupt a
// verification the operator is in the middle of reading.
func TestTheRestartRoutesAreBehindBothGates(t *testing.T) {
	for _, path := range []string{"/restart", "/soak/restart"} {
		t.Run(path, func(t *testing.T) {
			spy := &relaunchSpy{}
			soakCalls := 0
			h := newHarness(t, func(o *Options) {
				o.Relaunch = spy.fn()
				o.RestartSoak = func() (string, error) { soakCalls++; return "restarted", nil }
			})
			h.pretendListening("127.0.0.1:45678")

			// No session at all.
			resp := h.post(t, path, url.Values{"csrf": {h.csrf}})
			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("without a session: %d, want 403", resp.StatusCode)
			}

			h.authenticate(t)

			// A session but no CSRF token.
			resp = h.post(t, path, url.Values{})
			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("without the CSRF token: %d, want 403", resp.StatusCode)
			}
			// A session and the wrong CSRF token.
			resp = h.post(t, path, url.Values{"csrf": {"nope"}})
			if resp.StatusCode != http.StatusForbidden {
				t.Errorf("with a wrong CSRF token: %d, want 403", resp.StatusCode)
			}
			// A GET, which is how a prefetch or an <img> would arrive.
			resp = h.get(t, path)
			if resp.StatusCode != http.StatusMethodNotAllowed {
				t.Errorf("GET: %d, want 405", resp.StatusCode)
			}

			if len(spy.calls) != 0 || soakCalls != 0 {
				t.Fatalf("a refused request still reached a seam: relaunch=%v soak=%d", spy.calls, soakCalls)
			}
			h.noRelaunch(t)
		})
	}
}

// --- the console restart -----------------------------------------------------------

// TestARestartAsksForTheSamePortAndMintsOneHandoff.
func TestARestartAsksForTheSamePortAndMintsOneHandoff(t *testing.T) {
	dir := t.TempDir()
	store := handoff.New(filepath.Join(dir, handoff.FileName))
	spy := &relaunchSpy{}

	h := newHarness(t, func(o *Options) {
		o.Relaunch = spy.fn()
		o.Handoff = store
	})
	h.pretendListening("127.0.0.1:45678")
	h.authenticate(t)

	resp := h.post(t, "/restart", url.Values{"csrf": {h.csrf}})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /restart: %d, want 200", resp.StatusCode)
	}
	page := body(t, resp)
	if !strings.Contains(page, "다시 시작") {
		t.Errorf("the interstitial does not say what is happening:\n%s", page)
	}
	if !strings.Contains(page, "handoff=") {
		t.Errorf("the interstitial carries no handoff token:\n%s", page)
	}

	if got := h.awaitRelaunch(t); got != 45678 {
		t.Errorf("the relaunch was asked for port %d, want the port the console is on (45678) — a restart "+
			"that came back somewhere else is a restart the operator has to chase", got)
	}

	if _, err := os.Stat(store.Path()); err != nil {
		t.Fatalf("no handoff token was written: %v", err)
	}
	// The seam itself has not run: Serve calls it, and Serve is not running here.
	if len(spy.calls) != 0 {
		t.Errorf("the handler executed the relaunch itself (%v); only Serve may, because only Serve can "+
			"release the port the new process has to bind", spy.calls)
	}
}

// TestARestartIsRefusedWhileAVerificationIsWalking.
//
// This is the only refusal in the restart path that is about safety rather than
// about gates: a run in flight may have a live order resting, and throwing the
// process away would leave it there with nobody watching.
func TestARestartIsRefusedWhileAVerificationIsWalking(t *testing.T) {
	spy := &relaunchSpy{}
	h := newHarness(t, func(o *Options) { o.Relaunch = spy.fn() })
	h.pretendListening("127.0.0.1:45678")

	h.startAndWait(t) // parked on the approval, so the run is not finished

	resp := h.post(t, "/restart", url.Values{"csrf": {h.csrf}})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /restart: %d", resp.StatusCode)
	}
	if page := body(t, resp); !strings.Contains(page, "검증이 진행 중") {
		t.Errorf("the refusal does not say why:\n%s", page)
	}
	h.noRelaunch(t)
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("%d mutating call(s) were made by a refused restart", n)
	}
}

// TestARestartWithNoWiringSaysSoInsteadOfPretending.
func TestARestartWithNoWiringSaysSoInsteadOfPretending(t *testing.T) {
	h := newHarness(t)
	h.pretendListening("127.0.0.1:45678")
	h.authenticate(t)

	resp := h.post(t, "/restart", url.Values{"csrf": {h.csrf}})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /restart: %d", resp.StatusCode)
	}
	if page := body(t, resp); !strings.Contains(page, "재시작 배선이 없다") {
		t.Errorf("an unwired restart did not explain itself:\n%s", page)
	}
	h.noRelaunch(t)

	// And the dashboard does not draw a button for something this build cannot do.
	if page := body(t, h.get(t, "/")); strings.Contains(page, `action="/restart"`) {
		t.Error("the dashboard offers a restart button with no restart wiring behind it")
	}
}

// TestServeIsWhatExecutesTheRelaunchAndItReleasesThePortFirst.
//
// The handler cannot re-exec: it is holding the connection it just answered on and
// the listener is still open, so the successor could not bind. This drives the real
// Serve loop on a real loopback socket and asserts the order.
func TestServeIsWhatExecutesTheRelaunchAndItReleasesThePortFirst(t *testing.T) {
	dir := t.TempDir()
	var boundDuringRelaunch bool

	ln, err := Listen(0)
	if err != nil {
		t.Fatalf("Listen: %v", err)
	}
	port := ln.Addr().(*net.TCPAddr).Port

	c, err := New(Options{
		StartVerify:  realStarter(newFakeBroker(), filepath.Join(dir, "verify.jsonl")),
		VerifyRecord: filepath.Join(dir, "verify.jsonl"),
		Out:          io.Discard,
		Relaunch: func(p int) error {
			// The successor's job: bind the port this process was on.
			l, err := Listen(p)
			if err == nil {
				boundDuringRelaunch = true
				_ = l.Close()
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	served := make(chan error, 1)
	go func() { served <- c.Serve(t.Context(), ln) }()

	waitFor(t, func() bool { return c.Addr() != "" })

	client := newJarClient(t)
	base := "http://" + c.Addr()
	if _, err := client.Get(base + "/?session=" + c.SessionToken()); err != nil {
		t.Fatalf("authenticating: %v", err)
	}
	resp, err := client.PostForm(base+"/restart", url.Values{"csrf": {c.csrf}})
	if err != nil {
		t.Fatalf("POST /restart: %v", err)
	}
	resp.Body.Close()

	select {
	case err := <-served:
		if err != nil {
			t.Fatalf("Serve returned %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Serve did not return after a restart was requested")
	}
	if !boundDuringRelaunch {
		t.Errorf("the successor could not bind port %d; Serve executed the relaunch before releasing the "+
			"socket", port)
	}
}

// --- the browser coming back ------------------------------------------------------

// TestTheHandoffIsMintedByOneConsoleAndSpentByTheNext is the browser-continuity
// claim end to end, with two Console values and one real token file between them.
func TestTheHandoffIsMintedByOneConsoleAndSpentByTheNext(t *testing.T) {
	dir := t.TempDir()
	store := handoff.New(filepath.Join(dir, handoff.FileName))

	old := newHarness(t, func(o *Options) {
		o.Relaunch = (&relaunchSpy{}).fn()
		o.Handoff = store
	})
	old.pretendListening("127.0.0.1:45678")
	old.authenticate(t)

	resp := old.post(t, "/restart", url.Values{"csrf": {old.csrf}})
	token := handoffTokenFrom(t, body(t, resp))
	old.awaitRelaunch(t)

	// The successor: a different Console, a different session token, the same file.
	successor := newHarness(t, func(o *Options) { o.Handoff = store })
	if successor.SessionToken() == old.SessionToken() {
		t.Fatal("the two consoles minted the same session token; they are not independent processes")
	}

	got := successor.get(t, "/?handoff="+token)
	if got.StatusCode != http.StatusOK {
		t.Fatalf("the handoff was refused with %d: %s", got.StatusCode, body(t, got))
	}
	if page := body(t, got); !strings.Contains(page, "대시보드") {
		t.Errorf("the handoff did not land on the dashboard:\n%s", page)
	}
	// The session really is granted: a later request works without the token.
	if second := successor.get(t, "/verify"); second.StatusCode != http.StatusOK {
		t.Errorf("the granted session does not work: %d", second.StatusCode)
	}

	// And it is spent. A third console — a second restart's worth of browser — is
	// refused the same string.
	third := newHarness(t, func(o *Options) { o.Handoff = store })
	replay := third.get(t, "/?handoff="+token)
	if replay.StatusCode != http.StatusForbidden {
		t.Fatalf("replaying the handoff gave %d, want 403", replay.StatusCode)
	}
	if page := body(t, replay); !strings.Contains(page, "1회용") {
		t.Errorf("the replay refusal does not explain itself:\n%s", page)
	}
}

// TestAWrongOrExpiredHandoffIsRefusedAndGrantsNothing.
func TestAWrongOrExpiredHandoffIsRefusedAndGrantsNothing(t *testing.T) {
	dir := t.TempDir()

	t.Run("a token nobody minted", func(t *testing.T) {
		store := handoff.New(filepath.Join(dir, "a.json"))
		h := newHarness(t, func(o *Options) { o.Handoff = store })
		resp := h.get(t, "/?handoff=MADE-UP")
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("status %d, want 403", resp.StatusCode)
		}
		// Nothing was granted: the dashboard is still shut.
		if follow := h.get(t, "/"); follow.StatusCode != http.StatusForbidden {
			t.Errorf("a refused handoff still opened the console: %d", follow.StatusCode)
		}
	})

	t.Run("a token that sat too long", func(t *testing.T) {
		store := handoff.New(filepath.Join(dir, "b.json")).WithTTL(time.Second)
		clock := time.Date(2026, 7, 27, 10, 0, 0, 0, time.UTC)
		h := newHarness(t, func(o *Options) {
			o.Handoff = store
			o.Now = func() time.Time { return clock }
		})
		token, err := store.Mint(clock)
		if err != nil {
			t.Fatalf("Mint: %v", err)
		}
		clock = clock.Add(time.Minute)

		resp := h.get(t, "/?handoff="+token)
		if resp.StatusCode != http.StatusForbidden {
			t.Fatalf("an expired handoff gave %d, want 403", resp.StatusCode)
		}
	})

	t.Run("a wrong guess spends the real one", func(t *testing.T) {
		store := handoff.New(filepath.Join(dir, "c.json"))
		h := newHarness(t, func(o *Options) { o.Handoff = store })
		token, err := store.Mint(time.Now().UTC())
		if err != nil {
			t.Fatalf("Mint: %v", err)
		}
		if resp := h.get(t, "/?handoff=WRONG"); resp.StatusCode != http.StatusForbidden {
			t.Fatalf("a wrong guess gave %d, want 403", resp.StatusCode)
		}
		if resp := h.get(t, "/?handoff=" + token); resp.StatusCode != http.StatusForbidden {
			t.Fatalf("the real token survived a wrong guess (%d); the guess must spend it", resp.StatusCode)
		}
	})
}

// TestAConsoleWithNoHandoffStillWorksFromTheTerminal — the pre-1.8 path, which
// remains the root of trust.
func TestAConsoleWithNoHandoffStillWorksFromTheTerminal(t *testing.T) {
	h := newHarness(t)
	if resp := h.get(t, "/?handoff=anything"); resp.StatusCode != http.StatusForbidden {
		t.Errorf("a handoff on a console with no handoff store gave %d, want 403", resp.StatusCode)
	}
	h.authenticate(t) // and the session URL still works
}

// TestTheHandoffTokenLeavesTheAddressBar. It is a credential; it should not sit in
// history, in a referrer or in a screenshot.
func TestTheHandoffTokenLeavesTheAddressBar(t *testing.T) {
	store := handoff.New(filepath.Join(t.TempDir(), handoff.FileName))
	h := newHarness(t, func(o *Options) { o.Handoff = store })
	token, err := store.Mint(time.Now().UTC())
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	resp := h.get(t, "/?handoff="+token)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status %d", resp.StatusCode)
	}
	if final := resp.Request.URL.String(); strings.Contains(final, "handoff") {
		t.Errorf("the browser was left on %s with the token still in it", final)
	}
}

// --- the soak restart ---------------------------------------------------------------

// TestTheSoakRestartButtonCallsTheSeamAndReportsWhatItDid.
func TestTheSoakRestartButtonCallsTheSeamAndReportsWhatItDid(t *testing.T) {
	calls := 0
	h := newHarness(t, func(o *Options) {
		o.RestartSoak = func() (string, error) {
			calls++
			return "pid 4242에 SIGINT 후 재기동", nil
		}
	})
	h.authenticate(t)

	resp := h.post(t, "/soak/restart", url.Values{"csrf": {h.csrf}})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("POST /soak/restart: %d", resp.StatusCode)
	}
	if calls != 1 {
		t.Fatalf("the seam was called %d time(s), want 1", calls)
	}
	if page := body(t, resp); !strings.Contains(page, "pid 4242") {
		t.Errorf("the dashboard does not report what the restart did:\n%s", page)
	}
	if n := h.broker.mutationCount(); n != 0 {
		t.Errorf("restarting the soak made %d mutating call(s)", n)
	}
}

// TestAFailedSoakRestartIsReportedAndNotSwallowed.
func TestAFailedSoakRestartIsReportedAndNotSwallowed(t *testing.T) {
	h := newHarness(t, func(o *Options) {
		o.RestartSoak = func() (string, error) { return "", errors.New("no soak process was found") }
	})
	h.authenticate(t)

	resp := h.post(t, "/soak/restart", url.Values{"csrf": {h.csrf}})
	page := body(t, resp)
	if !strings.Contains(page, "soak 재기동 실패") || !strings.Contains(page, "no soak process was found") {
		t.Errorf("the failure is not on the page:\n%s", page)
	}
}

// --- the stale-binary warning ---------------------------------------------------------

// TestTheDashboardWarnsWhenAProcessIsOlderThanTheInstalledBinary.
func TestTheDashboardWarnsWhenAProcessIsOlderThanTheInstalledBinary(t *testing.T) {
	at := func(min int) binstamp.Stamp {
		return binstamp.Stamp{
			Path:    "/opt/tossctl",
			Size:    int64(100 + min),
			ModTime: time.Date(2026, 7, 27, 9, min, 0, 0, time.UTC),
		}
	}

	t.Run("quiet when the console is current", func(t *testing.T) {
		h := newHarness(t, func(o *Options) {
			o.Binary = func() (binstamp.Stamp, error) { return at(0), nil }
		})
		h.authenticate(t)
		if page := body(t, h.get(t, "/")); strings.Contains(page, "설치된 바이너리보다 오래되었다") {
			t.Errorf("a current console warned about itself:\n%s", page)
		}
	})

	t.Run("loud when the console is behind", func(t *testing.T) {
		reading := at(0)
		h := newHarness(t, func(o *Options) {
			o.Binary = func() (binstamp.Stamp, error) { return reading, nil }
		})
		reading = at(5) // reinstalled while the console was running
		h.authenticate(t)

		page := body(t, h.get(t, "/"))
		if !strings.Contains(page, "이 콘솔은 설치된 바이너리보다 오래되었다") {
			t.Errorf("a stale console did not say so:\n%s", page)
		}
	})

	t.Run("quiet when nothing could be observed", func(t *testing.T) {
		h := newHarness(t, func(o *Options) {
			o.Binary = func() (binstamp.Stamp, error) { return binstamp.Stamp{}, errors.New("no /proc") }
		})
		h.authenticate(t)
		if page := body(t, h.get(t, "/")); strings.Contains(page, "오래되었다") {
			t.Errorf("an unanswerable question became a warning:\n%s", page)
		}
	})
}

// --- small shared helpers --------------------------------------------------------

// waitFor blocks until cond holds, or fails the test.
func waitFor(t *testing.T, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("the condition never held")
}

// newJarClient is an HTTP client that keeps the session cookie, as a browser does.
func newJarClient(t *testing.T) *http.Client {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("cookiejar: %v", err)
	}
	return &http.Client{Jar: jar, Timeout: 10 * time.Second}
}

// handoffTokenFrom pulls the token out of the interstitial's own link, so the test
// reads what a browser would rather than what the store happens to hold.
func handoffTokenFrom(t *testing.T, page string) string {
	t.Helper()
	const marker = "handoff="
	idx := strings.Index(page, marker)
	if idx < 0 {
		t.Fatalf("no handoff token on the page:\n%s", page)
	}
	rest := page[idx+len(marker):]
	end := strings.IndexAny(rest, `"'<& `)
	if end < 0 {
		t.Fatalf("the handoff token is not delimited:\n%s", page)
	}
	token, err := url.QueryUnescape(rest[:end])
	if err != nil {
		t.Fatalf("unescaping the token: %v", err)
	}
	return token
}
