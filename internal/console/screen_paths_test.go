package console

// screen_paths_test.go pins where the screens live and what they are called
// (change a054-console-status-shell, §3).
//
// The defect being fixed is that one screen had three names. The root path was
// "검증 콘솔" in the navigation and "대시보드" in its own title, while /dashboard
// was "개요" — so the word "dashboard" pointed at two different screens and the
// screen it was the URL of was called something else entirely.
//
// The moves are cheap and the mistakes are not. A blanket path substitution
// would narrow the session cookie's scope from / to a screen, and the session
// would then vanish the moment the operator opened a different one; the handoff
// token is consumed by the session middleware BEFORE any handler sees it, so a
// redirect that tried to carry it would be carrying something already spent.
// Both have a test here.

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/handoff"
)

// noRedirect is a client that stops at the first 3xx, so a test can read the
// redirect itself rather than the page at the end of the chain.
func (h *harness) noRedirect(t *testing.T, path string) *http.Response {
	t.Helper()
	client := *h.client
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.Get(h.srv.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

// TestTheRootPathAnswersWithTheOverview (task 3.1).
//
// A redirect and not a 404: bookmarks and already-printed session links exist
// outside this process, and the terminal prints http://127.0.0.1:PORT/?session=…
// on every start.
func TestTheRootPathAnswersWithTheOverview(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)

	resp := h.noRedirect(t, "/")
	if resp.StatusCode != http.StatusSeeOther {
		t.Fatalf("GET / = %d, want 303", resp.StatusCode)
	}
	if got := resp.Header.Get("Location"); got != pathOverview {
		t.Errorf("GET / redirects to %q, want the overview at %q", got, pathOverview)
	}

	// And following it lands on a rendered screen, not another redirect.
	page := body(t, h.get(t, "/"))
	if !strings.Contains(page, "<h1>개요</h1>") {
		t.Error("following the root path does not land on the overview")
	}
}

// TestTheVerificationConsoleHasItsOwnPathAndKeepsItsControls (tasks 3.6, 3.9).
func TestTheVerificationConsoleHasItsOwnPathAndKeepsItsControls(t *testing.T) {
	h := newEngineHarness(t, func(o *Options) {
		o.Relaunch = func(int) error { return nil }
		o.RestartSoak = func() (string, error) { return "", nil }
	})
	h.authenticate(t)

	resp := h.get(t, pathVerifyConsole)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("GET %s = %d, want 200", pathVerifyConsole, resp.StatusCode)
	}
	page := body(t, resp)
	for _, want := range []string{
		"<h1>검증 콘솔</h1>",
		`action="/engine/start"`, `action="/engine/stop"`,
		`action="/restart"`, `action="/soak/restart"`,
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the verification console lost %q when it moved off the root path", want)
		}
	}
}

// TestTheRestartHandoffSurvivesTheMove (task 3.2).
//
// The contract is the OUTCOME, not the mechanism. It is tempting to write
// "the redirect preserves ?handoff=", and that would be wrong twice: session0
// runs before any handler and acceptHandoff consumes the token there, and
// grantSession then deletes the parameter on purpose so a spent token stops
// living in the address bar.
func TestTheRestartHandoffSurvivesTheMove(t *testing.T) {
	dir := t.TempDir()
	store := handoff.New(dir + "/handoff.json")

	minting := newHarness(t, func(o *Options) { o.Handoff = store })
	token, note := minting.mintHandoff()
	if token == "" {
		t.Fatalf("no handoff token was minted: %s", note)
	}

	successor := newHarness(t, func(o *Options) { o.Handoff = store })
	resp := successor.get(t, "/?handoff="+url.QueryEscape(token))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("the handoff landed on %d, want a rendered screen", resp.StatusCode)
	}
	if !strings.Contains(body(t, resp), "<h1>") {
		t.Error("the handoff did not land on a rendered screen")
	}
	if got := resp.Request.URL.RawQuery; strings.Contains(got, "handoff") {
		t.Errorf("a spent handoff token is still in the address bar: %q", got)
	}

	// Task 3.2, second half: single use is not weakened by the move.
	third := newHarness(t, func(o *Options) { o.Handoff = store })
	if again := third.noRedirect(t, "/?handoff="+url.QueryEscape(token)); again.StatusCode == http.StatusSeeOther {
		t.Errorf("a spent handoff token was accepted again (%d)", again.StatusCode)
	}
}

// TestARestartNoticeComesBackToTheScreenThatStartedIt (task 3.3).
//
// The restart and soak-restart buttons are on the verification console. Sending
// their result to the overview would make an operator read the outcome of a
// button on a screen that does not have that button.
func TestARestartNoticeComesBackToTheScreenThatStartedIt(t *testing.T) {
	if got := restartTarget(""); got != pathVerifyConsole {
		t.Errorf("restartTarget() = %q, want the screen the restart button is on (%q)",
			got, pathVerifyConsole)
	}
	if got := restartTarget("TOKEN"); !strings.HasPrefix(got, pathVerifyConsole+"?handoff=") {
		t.Errorf("restartTarget(token) = %q, want %s?handoff=…", got, pathVerifyConsole)
	}

	h := newHarness(t, func(o *Options) {
		o.RestartSoak = func() (string, error) { return "soak을 다시 시작했다.", nil }
	})
	h.authenticate(t)
	resp := h.postNoRedirect(t, "/soak/restart", url.Values{"csrf": {h.csrf}})
	location := resp.Header.Get("Location")
	if !strings.HasPrefix(location, pathVerifyConsole) {
		t.Errorf("a soak restart lands on %q, want the screen its button is on", location)
	}
	if !strings.Contains(location, "notice=") {
		t.Errorf("the restart result was dropped on the way back: %q", location)
	}
}

// postNoRedirect posts and stops at the redirect.
func (h *harness) postNoRedirect(t *testing.T, path string, form url.Values) *http.Response {
	t.Helper()
	client := *h.client
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return http.ErrUseLastResponse
	}
	resp, err := client.PostForm(h.srv.URL+path, form)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	return resp
}

// TestTheSessionCookieStaysScopedToTheWholeConsole (task 3.5).
//
// The easiest accident in this change: `Path: "/"` on a cookie looks exactly
// like a route and is not one. Narrowing it to a screen logs the operator out of
// every other screen.
func TestTheSessionCookieStaysScopedToTheWholeConsole(t *testing.T) {
	h := newHarness(t)
	resp := h.noRedirect(t, "/?session="+h.SessionToken())
	var found bool
	for _, cookie := range resp.Cookies() {
		if cookie.Name != sessionCookie {
			continue
		}
		found = true
		if cookie.Path != "/" {
			t.Errorf("the session cookie is scoped to %q; every screen outside it loses the session",
				cookie.Path)
		}
	}
	if !found {
		t.Fatal("no session cookie was issued")
	}

	// The behavioural half: a session taken at the root reaches a screen that is
	// not under the screen the root now redirects to.
	for _, path := range []string{"/positions", "/orders", pathVerifyConsole} {
		if got := h.get(t, path).StatusCode; got != http.StatusOK {
			t.Errorf("GET %s = %d after authenticating at the root; the session did not travel",
				path, got)
		}
	}
}

// TestEveryScreenIsCalledOneThing (task 3.7).
//
// The comparison is between two RENDERED strings — the navigation item carrying
// aria-current, and the <h1> — because the three names live in Go string
// templates and handler fields and are not a static-parse target. A qualifier is
// allowed, in a muted span, and is not part of the name: "검증" and "검증 (실계좌
// · KR 시장)" are the same screen, "검증 콘솔" and "대시보드" are not.
func TestEveryScreenIsCalledOneThing(t *testing.T) {
	h := newHarness(t)
	h.authenticate(t)
	for _, screen := range consoleScreens {
		page := body(t, h.get(t, screen.path))
		nav, ok := currentNavLabel(page)
		if !ok {
			continue // a screen with no navigation entry has no name to disagree with
		}
		title, ok := documentTitle(page)
		if !ok {
			t.Errorf("%s renders no <h1>", screen.path)
			continue
		}
		if nav != title {
			t.Errorf("%s is called %q in the navigation and %q in its title", screen.path, nav, title)
		}
	}
}

// currentNavLabel is the text of the navigation item marked aria-current.
//
// It reads the top navigation by its aria-label rather than taking the first
// aria-current in the document. The settings tabs are also a navigation and also
// carry aria-current="page" — correctly, they ARE the current page — and a
// document-order rule would have compared a tab label against the screen title
// the first time the two bars were reordered (change a055).
//
// The one-line description is stripped for the reason documentTitle strips the
// muted qualifier: the description is what the screen answers, not what it is
// called. "개요" and "개요 · 지금 무엇이 참인가" are the same screen.
func currentNavLabel(page string) (string, bool) {
	bar := page
	if i := strings.Index(page, `<nav aria-label="주요 화면">`); i >= 0 {
		bar = page[i:]
		if j := strings.Index(bar, "</nav>"); j >= 0 {
			bar = bar[:j]
		}
	}
	i := strings.Index(bar, `aria-current="page"`)
	if i < 0 {
		return "", false
	}
	rest := bar[i:]
	open := strings.Index(rest, ">")
	closing := strings.Index(rest, "</a>")
	if open < 0 || closing < 0 || open > closing {
		return "", false
	}
	label := rest[open+1 : closing]
	if k := strings.Index(label, "<small>"); k >= 0 {
		label = label[:k]
	}
	return strings.TrimSpace(label), true
}

// documentTitle is the <h1>'s name, with a muted qualifier removed.
func documentTitle(page string) (string, bool) {
	i := strings.Index(page, "<h1>")
	if i < 0 {
		return "", false
	}
	rest := page[i+len("<h1>"):]
	j := strings.Index(rest, "</h1>")
	if j < 0 {
		return "", false
	}
	title := rest[:j]
	if k := strings.Index(title, `<span class="muted">`); k >= 0 {
		title = title[:k]
	}
	return strings.TrimSpace(title), true
}

// TestTheShellsPathConstantsAreRegisteredRoutes (task 3.9).
//
// The route table is read out of the source, so a route has to be registered
// with a literal — which means the constants the links use are a second
// spelling. This is what keeps the two from drifting.
func TestTheShellsPathConstantsAreRegisteredRoutes(t *testing.T) {
	registered := map[string]bool{}
	for _, r := range registeredRoutes(t) {
		registered[r.Path] = true
	}
	for _, path := range []string{pathOverview, pathVerifyConsole} {
		if !registered[path] {
			t.Errorf("%q is what the shell links to and no route answers it", path)
		}
	}

	// And the navigation links to the same place, from the template's own literal.
	h := newHarness(t)
	h.authenticate(t)
	page := body(t, h.get(t, pathVerifyConsole))
	if !strings.Contains(page, `href="`+pathOverview+`"`) {
		t.Errorf("the navigation does not link href=%q", pathOverview)
	}
	if strings.Contains(page, `<a href="/" `) || strings.Contains(page, `<a href="/">`) {
		t.Error("the navigation still points a link at the root redirect")
	}

	// The verification console left the top navigation when it went to six items
	// (change a055) and it is not thereby hidden: the 도구 tab links to it, which
	// is more than it had before — it was reachable only as the root path.
	tools := body(t, h.get(t, pathSettingsTools))
	if !strings.Contains(tools, `href="`+pathVerifyConsole+`"`) {
		t.Errorf("nothing links %q any more; the screen is unreachable from the navigation",
			pathVerifyConsole)
	}
}
