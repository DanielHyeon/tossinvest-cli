package console

import (
	"context"
	"net/http"
	"strings"
)

const openAPISetupMaxBody = 8 << 10

// OpenAPICredentialState is a controlled, secret-free result from the CLI-owned
// Open API credential seam.
type OpenAPICredentialState string

const (
	OpenAPICredentialsReady               OpenAPICredentialState = "ready"
	OpenAPICredentialsMissing             OpenAPICredentialState = "missing"
	OpenAPICredentialsRejected            OpenAPICredentialState = "rejected"
	OpenAPICredentialsEnvironmentRejected OpenAPICredentialState = "environment_rejected"
	OpenAPICredentialsUnavailable         OpenAPICredentialState = "unavailable"
	OpenAPICredentialsSavedNotStarted     OpenAPICredentialState = "saved_not_started"
)

// OpenAPICredentialCheck contains no credential material. Message must be fixed
// operator guidance produced by cmd/tossctl, never a broker response body.
type OpenAPICredentialCheck struct {
	State   OpenAPICredentialState
	Message string
}

// CheckOpenAPICredentials validates the effective saved credential pair with
// one official read-only probe.
type CheckOpenAPICredentials func(context.Context) OpenAPICredentialCheck

// SaveOpenAPICredentials validates and persists a submitted pair. Its result is
// secret-free; Ready is the only state that permits automatic soak restart.
type SaveOpenAPICredentials func(context.Context, string, string) OpenAPICredentialCheck

type openAPISetupPage struct {
	Nav     string
	CSRF    string
	Message string
}

func (openAPISetupPage) Refresh() bool { return false }

// credentialHTTPS is deliberately narrower than remote.security: the legacy
// loopback console may remain HTTP for non-secret operations, but key/secret
// ingress is never accepted without direct TLS. Remote mode still applies its
// peer, Host, and exact-origin checks outside this wrapper.
func (c *Console) credentialHTTPS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil {
			c.refuse(w, http.StatusForbidden, "HTTPS 연결이 필요하다",
				"Open API 자격증명은 설정된 HTTPS 콘솔에서만 입력할 수 있다. 아무것도 저장하거나 시작하지 않았다.")
			return
		}
		next(w, r)
	}
}

func (c *Console) handleOpenAPILogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		c.refuse(w, http.StatusMethodNotAllowed, "GET 전용",
			"자격증명 입력 화면은 읽기 요청으로만 연다. 아무것도 저장하거나 시작하지 않았다.")
		return
	}
	message := "Open API 키와 시크릿을 입력하면 검증·보호 저장 후 soak을 바로 시작한다."
	switch r.URL.Query().Get("reason") {
	case "missing":
		message = "저장된 Open API 자격증명이 없다. 한 번 입력하면 이후 재사용한다."
	case "rejected":
		message = "저장된 Open API 자격증명이 거부되었다. 새 키와 시크릿으로 교체하라."
	case "pending":
		message = "이전 Open API 자격증명 설정을 완료하지 못했다. 키와 시크릿을 다시 저장하면 검증·감사 후 soak을 바로 시작한다."
	}
	c.renderOpenAPISetup(w, message)
}

func (c *Console) handleOpenAPILoginSave(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.PostFormValue("key"))
	secret := strings.TrimSpace(r.PostFormValue("secret"))
	if key == "" || secret == "" {
		c.renderOpenAPISetup(w, "Open API 키와 시크릿을 모두 입력하라.")
		return
	}
	if c.opts.SaveOpenAPI == nil {
		c.renderOpenAPISetup(w, "이 콘솔에는 Open API 자격증명 저장 기능이 연결되지 않았다.")
		return
	}

	c.openAPIMu.Lock()
	defer c.openAPIMu.Unlock()
	result := c.opts.SaveOpenAPI(r.Context(), key, secret)
	if result.State != OpenAPICredentialsReady {
		c.renderOpenAPISetup(w, credentialMessage(result,
			"Open API 자격증명을 저장하지 못했다. 잠시 후 다시 시도하라."))
		return
	}
	c.restartSoakLocked(w, r)
}

func (c *Console) renderOpenAPISetup(w http.ResponseWriter, message string) {
	c.render(w, "openapi-login", openAPISetupPage{
		Nav:     "verify-console",
		CSRF:    c.csrf,
		Message: message,
	})
}

func credentialMessage(result OpenAPICredentialCheck, fallback string) string {
	if message := strings.TrimSpace(result.Message); message != "" {
		return message
	}
	return fallback
}

const openAPIOnboardingTemplates = `
{{define "openapi-login"}}
{{template "head" .}}
<h1>Open API 자격증명 설정</h1>
<section>
  <p class="notice">{{.Message}}</p>
  <p>토스증권 개발자 포털에서 발급한 Open API 키와 시크릿을 입력하라.
  저장 후에는 접근 토큰이 만료될 때 공식 클라이언트가 자동으로 갱신한다.</p>
  <form method="post" action="/openapi/login/save" autocomplete="off">
    <input type="hidden" name="csrf" value="{{.CSRF}}">
    <p><label>Open API 키<br><input type="password" name="key" autocomplete="off" required></label></p>
    <p><label>Open API 시크릿<br><input type="password" name="secret" autocomplete="off" required></label></p>
    <button type="submit">저장하고 soak 시작</button>
  </form>
  <p class="muted">입력값은 이 화면·주소·로그·감사 기록에 다시 표시하지 않는다.</p>
</section>
{{template "foot" .}}
{{end}}
`
