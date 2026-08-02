package console

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"testing"
)

func credentialCheck(state OpenAPICredentialState, message string) OpenAPICredentialCheck {
	return OpenAPICredentialCheck{State: state, Message: message}
}

func TestReadyCredentialsRestartSoakInOneClick(t *testing.T) {
	checks := 0
	restarts := 0
	h := newTLSHarness(t, func(o *Options) {
		o.CheckOpenAPI = func(context.Context) OpenAPICredentialCheck {
			checks++
			return credentialCheck(OpenAPICredentialsReady, "")
		}
		o.RestartSoak = func() (string, error) {
			restarts++
			return "soak ready", nil
		}
	})
	h.authenticate(t)

	resp := h.post(t, "/soak/restart", url.Values{"csrf": {h.csrf}})
	page := body(t, resp)
	if checks != 1 || restarts != 1 {
		t.Fatalf("checks=%d restarts=%d, want one each", checks, restarts)
	}
	if !strings.Contains(page, "soak ready") {
		t.Fatalf("restart result missing:\n%s", page)
	}
}

func TestMissingCredentialsOpenSetupWithoutRestart(t *testing.T) {
	restarts := 0
	h := newTLSHarness(t, func(o *Options) {
		o.CheckOpenAPI = func(context.Context) OpenAPICredentialCheck {
			return credentialCheck(OpenAPICredentialsMissing, "자격증명이 필요하다.")
		}
		o.SaveOpenAPI = func(context.Context, string, string) OpenAPICredentialCheck {
			t.Fatal("GET/setup redirect called the save seam")
			return OpenAPICredentialCheck{}
		}
		o.RestartSoak = func() (string, error) {
			restarts++
			return "", nil
		}
	})
	h.authenticate(t)

	resp := h.post(t, "/soak/restart", url.Values{"csrf": {h.csrf}})
	page := body(t, resp)
	if restarts != 0 {
		t.Fatalf("restart called %d time(s)", restarts)
	}
	if resp.Request.URL.Path != "/openapi/login" {
		t.Fatalf("final path = %q, want /openapi/login", resp.Request.URL.Path)
	}
	for _, want := range []string{"Open API", `type="password"`, `name="key"`, `name="secret"`} {
		if !strings.Contains(page, want) {
			t.Fatalf("setup page missing %q:\n%s", want, page)
		}
	}
}

func TestValidSetupSavesThenStartsSoakWithoutASecondClick(t *testing.T) {
	const key = "submitted-api-key"
	const secret = "submitted-super-secret"
	var order []string
	h := newTLSHarness(t, func(o *Options) {
		o.CheckOpenAPI = func(context.Context) OpenAPICredentialCheck {
			return credentialCheck(OpenAPICredentialsMissing, "")
		}
		o.SaveOpenAPI = func(_ context.Context, gotKey, gotSecret string) OpenAPICredentialCheck {
			if gotKey != key || gotSecret != secret {
				t.Fatalf("submitted credentials changed")
			}
			order = append(order, "save")
			return credentialCheck(OpenAPICredentialsReady, "")
		}
		o.RestartSoak = func() (string, error) {
			order = append(order, "restart")
			return "soak started after setup", nil
		}
	})
	h.authenticate(t)

	resp := h.post(t, "/openapi/login/save", url.Values{
		"csrf":   {h.csrf},
		"key":    {key},
		"secret": {secret},
	})
	page := body(t, resp)
	if got := strings.Join(order, ","); got != "save,restart" {
		t.Fatalf("order = %q, want save,restart", got)
	}
	if !strings.Contains(page, "soak started after setup") {
		t.Fatalf("automatic restart result missing:\n%s", page)
	}
	for _, forbidden := range []string{key, secret} {
		if strings.Contains(page, forbidden) || strings.Contains(resp.Request.URL.String(), forbidden) {
			t.Fatal("credential leaked into response or URL")
		}
	}
}

func TestRejectedSetupEchoesNoCredentialAndStartsNothing(t *testing.T) {
	const key = "rejected-api-key"
	const secret = "rejected-super-secret"
	restarts := 0
	h := newTLSHarness(t, func(o *Options) {
		o.SaveOpenAPI = func(context.Context, string, string) OpenAPICredentialCheck {
			return credentialCheck(OpenAPICredentialsRejected, "키 또는 시크릿이 거부되었다.")
		}
		o.RestartSoak = func() (string, error) {
			restarts++
			return "", nil
		}
	})
	h.authenticate(t)

	resp := h.post(t, "/openapi/login/save", url.Values{
		"csrf":   {h.csrf},
		"key":    {key},
		"secret": {secret},
	})
	page := body(t, resp)
	if restarts != 0 {
		t.Fatalf("restart called %d time(s)", restarts)
	}
	for _, forbidden := range []string{key, secret} {
		if strings.Contains(page, forbidden) || strings.Contains(resp.Request.URL.String(), forbidden) {
			t.Fatal("rejected credential leaked")
		}
	}
	if !strings.Contains(page, "키 또는 시크릿이 거부되었다") {
		t.Fatalf("fixed rejection guidance missing:\n%s", page)
	}
}

func TestSavedButNotStartedIsExplicitAndDoesNotRestart(t *testing.T) {
	restarts := 0
	h := newTLSHarness(t, func(o *Options) {
		o.SaveOpenAPI = func(context.Context, string, string) OpenAPICredentialCheck {
			return credentialCheck(OpenAPICredentialsSavedNotStarted,
				"자격증명은 저장됨, soak은 시작되지 않음 — 감사 기록 실패")
		}
		o.RestartSoak = func() (string, error) {
			restarts++
			return "", nil
		}
	})
	h.authenticate(t)

	resp := h.post(t, "/openapi/login/save", url.Values{
		"csrf":   {h.csrf},
		"key":    {"safe-key"},
		"secret": {"safe-secret"},
	})
	page := body(t, resp)
	if restarts != 0 {
		t.Fatalf("restart called %d time(s)", restarts)
	}
	if !strings.Contains(page, "자격증명은 저장됨, soak은 시작되지 않음") {
		t.Fatalf("partial state is ambiguous:\n%s", page)
	}
}

func TestPendingGenerationRestartReopensSetupWithoutRestart(t *testing.T) {
	restarts := 0
	h := newTLSHarness(t, func(o *Options) {
		o.CheckOpenAPI = func(context.Context) OpenAPICredentialCheck {
			return credentialCheck(OpenAPICredentialsSavedNotStarted,
				"자격증명은 저장됨, soak은 시작되지 않음 — 설정 화면에서 다시 저장하라.")
		}
		o.RestartSoak = func() (string, error) {
			restarts++
			return "", nil
		}
	})
	h.authenticate(t)

	resp := h.post(t, "/soak/restart", url.Values{"csrf": {h.csrf}})
	page := body(t, resp)
	if restarts != 0 {
		t.Fatalf("restart called %d time(s)", restarts)
	}
	if resp.Request.URL.Path != "/openapi/login" {
		t.Fatalf("pending generation opened %q, want setup", resp.Request.URL.Path)
	}
	if !strings.Contains(page, "설정을 완료하지 못했다") {
		t.Fatalf("pending recovery guidance missing:\n%s", page)
	}
}

func TestRejectedEnvironmentCredentialsStayOffTheSetupScreen(t *testing.T) {
	restarts := 0
	h := newTLSHarness(t, func(o *Options) {
		o.CheckOpenAPI = func(context.Context) OpenAPICredentialCheck {
			return credentialCheck(OpenAPICredentialsEnvironmentRejected,
				"TOSSCTL_OPENAPI_KEY와 TOSSCTL_OPENAPI_SECRET을 갱신한 뒤 컨테이너를 다시 생성하라.")
		}
		o.RestartSoak = func() (string, error) {
			restarts++
			return "", nil
		}
	})
	h.authenticate(t)

	resp := h.post(t, "/soak/restart", url.Values{"csrf": {h.csrf}})
	page := body(t, resp)
	if restarts != 0 {
		t.Fatalf("restart called %d time(s)", restarts)
	}
	if resp.Request.URL.Path != pathVerifyConsole {
		t.Fatalf("environment-managed rejection opened %q", resp.Request.URL.Path)
	}
	if !strings.Contains(page, "TOSSCTL_OPENAPI_KEY") || !strings.Contains(page, "컨테이너를 다시 생성") {
		t.Fatalf("container guidance missing:\n%s", page)
	}
}

func TestOpenAPISetupPreservesSessionAndCSRFGates(t *testing.T) {
	calls := 0
	h := newTLSHarness(t, func(o *Options) {
		o.SaveOpenAPI = func(context.Context, string, string) OpenAPICredentialCheck {
			calls++
			return credentialCheck(OpenAPICredentialsReady, "")
		}
	})

	resp := h.post(t, "/openapi/login/save", url.Values{
		"csrf":   {h.csrf},
		"key":    {"key"},
		"secret": {"secret"},
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("without session status=%d, want 403", resp.StatusCode)
	}
	h.authenticate(t)
	resp = h.post(t, "/openapi/login/save", url.Values{
		"key":    {"key"},
		"secret": {"secret"},
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("without CSRF status=%d, want 403", resp.StatusCode)
	}
	if calls != 0 {
		t.Fatalf("save seam called %d time(s)", calls)
	}
}

func TestOpenAPISetupPreservesRemotePeerHostOriginSessionAndCSRFGates(t *testing.T) {
	cert, key := writeRemoteTestCertificate(t, "console.vpn.test")
	calls := 0
	c, err := New(Options{
		StartVerify: noopRemoteStarter,
		Remote: RemoteAccess{
			Bind:         "0.0.0.0",
			AllowedCIDRs: []string{"10.8.0.0/24"},
			PublicURL:    "https://" + testRemoteHost,
			TLSCertFile:  cert,
			TLSKeyFile:   key,
			AccessToken:  testRemoteToken,
			RecordAccess: func(RemoteAccessEvent) error {
				return nil
			},
		},
		SaveOpenAPI: func(context.Context, string, string) OpenAPICredentialCheck {
			calls++
			return credentialCheck(OpenAPICredentialsReady, "")
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	cookie := remoteLogin(t, c, "10.8.0.12:44000", "mobile")
	form := url.Values{
		"csrf":   {c.csrf},
		"key":    {"key"},
		"secret": {"secret"},
	}

	outside := remoteRequest(http.MethodPost, "/openapi/login/save",
		"203.0.113.9:44001", "mobile", form)
	outside.Header.Set("Origin", "https://"+testRemoteHost)
	outside.AddCookie(cookie)
	if got := serveRemote(c, outside).Code; got != http.StatusForbidden {
		t.Fatalf("outside peer status=%d, want 403", got)
	}

	badHost := remoteRequest(http.MethodPost, "/openapi/login/save",
		"10.8.0.12:44002", "mobile", form)
	badHost.Host = "attacker.invalid"
	badHost.Header.Set("Origin", "https://"+testRemoteHost)
	badHost.AddCookie(cookie)
	if got := serveRemote(c, badHost).Code; got != http.StatusForbidden {
		t.Fatalf("bad Host status=%d, want 403", got)
	}

	badOrigin := remoteRequest(http.MethodPost, "/openapi/login/save",
		"10.8.0.12:44003", "mobile", form)
	badOrigin.Header.Set("Origin", "https://attacker.invalid")
	badOrigin.AddCookie(cookie)
	if got := serveRemote(c, badOrigin).Code; got != http.StatusForbidden {
		t.Fatalf("bad Origin status=%d, want 403", got)
	}

	noSession := remoteRequest(http.MethodPost, "/openapi/login/save",
		"10.8.0.12:44004", "mobile", form)
	noSession.Header.Set("Origin", "https://"+testRemoteHost)
	if got := serveRemote(c, noSession).Code; got != http.StatusSeeOther {
		t.Fatalf("missing session status=%d, want 303", got)
	}

	noCSRF := remoteRequest(http.MethodPost, "/openapi/login/save",
		"10.8.0.12:44005", "mobile", url.Values{"key": {"key"}, "secret": {"secret"}})
	noCSRF.Header.Set("Origin", "https://"+testRemoteHost)
	noCSRF.AddCookie(cookie)
	if got := serveRemote(c, noCSRF).Code; got != http.StatusForbidden {
		t.Fatalf("missing CSRF status=%d, want 403", got)
	}

	valid := remoteRequest(http.MethodPost, "/openapi/login/save",
		"10.8.0.12:44006", "mobile", form)
	valid.Header.Set("Origin", "https://"+testRemoteHost)
	valid.AddCookie(cookie)
	if got := serveRemote(c, valid).Code; got != http.StatusSeeOther {
		t.Fatalf("valid setup status=%d, want 303", got)
	}
	if calls != 1 {
		t.Fatalf("save seam calls=%d, want 1", calls)
	}
}

func TestOpenAPISetupUsesTrustedNetworkAccessWithoutApplicationLogin(t *testing.T) {
	cert, key := writeRemoteTestCertificate(t, "console.vpn.test")
	calls := 0
	c, err := New(Options{
		StartVerify: noopRemoteStarter,
		Remote: RemoteAccess{
			Bind:           "0.0.0.0",
			AllowedCIDRs:   []string{"10.8.0.0/24"},
			PublicURL:      "https://" + testRemoteHost,
			TLSCertFile:    cert,
			TLSKeyFile:     key,
			TrustedNetwork: true,
			RecordAccess: func(RemoteAccessEvent) error {
				return nil
			},
		},
		SaveOpenAPI: func(context.Context, string, string) OpenAPICredentialCheck {
			calls++
			return credentialCheck(OpenAPICredentialsReady, "")
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	form := url.Values{
		"csrf":   {c.csrf},
		"key":    {"key"},
		"secret": {"secret"},
	}

	outside := remoteRequest(http.MethodPost, "/openapi/login/save",
		"203.0.113.9:44101", "mobile", form)
	outside.Header.Set("Origin", "https://"+testRemoteHost)
	if got := serveRemote(c, outside).Code; got != http.StatusForbidden {
		t.Fatalf("outside peer status=%d, want 403", got)
	}

	badOrigin := remoteRequest(http.MethodPost, "/openapi/login/save",
		"10.8.0.12:44102", "mobile", form)
	badOrigin.Header.Set("Origin", "https://attacker.invalid")
	if got := serveRemote(c, badOrigin).Code; got != http.StatusForbidden {
		t.Fatalf("bad Origin status=%d, want 403", got)
	}

	noCSRF := remoteRequest(http.MethodPost, "/openapi/login/save",
		"10.8.0.12:44103", "mobile", url.Values{"key": {"key"}, "secret": {"secret"}})
	noCSRF.Header.Set("Origin", "https://"+testRemoteHost)
	if got := serveRemote(c, noCSRF).Code; got != http.StatusForbidden {
		t.Fatalf("missing CSRF status=%d, want 403", got)
	}

	valid := remoteRequest(http.MethodPost, "/openapi/login/save",
		"10.8.0.12:44104", "mobile", form)
	valid.Header.Set("Origin", "https://"+testRemoteHost)
	if got := serveRemote(c, valid).Code; got != http.StatusSeeOther {
		t.Fatalf("trusted-network setup status=%d, want 303", got)
	}
	if calls != 1 {
		t.Fatalf("save seam calls=%d, want 1", calls)
	}
}

func TestOpenAPISetupRejectsOversizeBeforeSeams(t *testing.T) {
	calls := 0
	h := newTLSHarness(t, func(o *Options) {
		o.SaveOpenAPI = func(context.Context, string, string) OpenAPICredentialCheck {
			calls++
			return credentialCheck(OpenAPICredentialsReady, "")
		}
	})
	h.authenticate(t)

	form := "csrf=" + url.QueryEscape(h.csrf) + "&key=" + strings.Repeat("k", openAPISetupMaxBody)
	req, err := http.NewRequest(http.MethodPost, h.srv.URL+"/openapi/login/save", strings.NewReader(form))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := h.client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Fatalf("status=%d, want 413", resp.StatusCode)
	}
	if calls != 0 {
		t.Fatalf("save seam called %d time(s)", calls)
	}
}

func TestOpenAPISetupRejectsPlaintextEvenWithSessionAndCSRF(t *testing.T) {
	calls := 0
	h := newHarness(t, func(o *Options) {
		o.SaveOpenAPI = func(context.Context, string, string) OpenAPICredentialCheck {
			calls++
			return credentialCheck(OpenAPICredentialsReady, "")
		}
	})
	h.authenticate(t)

	resp := h.post(t, "/openapi/login/save", url.Values{
		"csrf":   {h.csrf},
		"key":    {"key"},
		"secret": {"secret"},
	})
	page := body(t, resp)
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("plaintext status=%d, want 403", resp.StatusCode)
	}
	if calls != 0 {
		t.Fatalf("plaintext reached save seam %d time(s)", calls)
	}
	if !strings.Contains(page, "HTTPS") {
		t.Fatalf("plaintext refusal lacks HTTPS guidance:\n%s", page)
	}
}

func TestOpenAPIRequestsAreSerializedThroughRestart(t *testing.T) {
	firstEntered := make(chan struct{})
	releaseFirst := make(chan struct{})
	var mu sync.Mutex
	var firstOnce sync.Once
	active := 0
	maxActive := 0
	h := newTLSHarness(t, func(o *Options) {
		o.CheckOpenAPI = func(context.Context) OpenAPICredentialCheck {
			mu.Lock()
			active++
			if active > maxActive {
				maxActive = active
			}
			first := active == 1 && maxActive == 1
			mu.Unlock()
			if first {
				firstOnce.Do(func() {
					close(firstEntered)
					<-releaseFirst
				})
			}
			mu.Lock()
			active--
			mu.Unlock()
			return credentialCheck(OpenAPICredentialsReady, "")
		}
		o.RestartSoak = func() (string, error) { return "ok", nil }
	})
	h.authenticate(t)

	done := make(chan struct{}, 2)
	post := func() {
		resp, err := h.client.PostForm(h.srv.URL+"/soak/restart", url.Values{"csrf": {h.csrf}})
		if err == nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			resp.Body.Close()
		}
		done <- struct{}{}
	}
	go post()
	<-firstEntered
	go post()
	close(releaseFirst)
	<-done
	<-done
	if maxActive != 1 {
		t.Fatalf("credential transactions overlapped: max active=%d", maxActive)
	}
}
