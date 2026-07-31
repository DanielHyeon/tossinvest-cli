package console

import (
	"net/http"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

func TestRemoteSettingsCSPRemainsScriptFree(t *testing.T) {
	rig := newRemoteTestRig(t)
	rig.console.opts.Settings = &fakeSettings{block: config.Adoption{DefaultStopPct: 0.075}}
	cookie := remoteLogin(t, rig.console, "10.8.0.14:44000", "mobile")
	request := remoteRequest(http.MethodGet, "/settings", "10.8.0.14:44001", "mobile", nil)
	request.AddCookie(cookie)
	response := serveRemote(rig.console, request)
	if response.Code != http.StatusOK {
		t.Fatalf("remote settings status = %d, body=%s", response.Code, response.Body.String())
	}
	const want = "default-src 'none'; style-src 'unsafe-inline'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'"
	if got := response.Header().Get("Content-Security-Policy"); got != want {
		t.Errorf("Content-Security-Policy = %q, want %q", got, want)
	}
	body := response.Body.String()
	for _, required := range []string{`name="default_stop_percent"`, `value="7.5"`} {
		if !strings.Contains(body, required) {
			t.Errorf("remote settings body missing %q", required)
		}
	}
	for _, forbidden := range []string{"oninput=", "<script"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("remote adoption control contains %q under deny-by-default CSP", forbidden)
		}
	}
}
