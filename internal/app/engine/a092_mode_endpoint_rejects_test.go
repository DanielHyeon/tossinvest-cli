//go:build unix

package engine

// a092 26라운드 보이스 C #3: mutating 표면(모드 완화 엔드포인트)의 거절 — 토큰(C12 「다른 힘이면 다른 토큰」) · POST 한정 ·
// 모르는 필드 거절. 거절된 요청은 원장에 닿지 않음(모드 · audit 불변). 양성 대조: 같은 요청이 올바른 모양이면 풀림.
// B#2: 엔진 고장(500)의 본문에 원문 오류(계좌를 담은 원장 문구)를 싣지 않음.

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

const (
	// 두 토큰은 같은 길이 — 길이 검사만으로 거절되면 내용 비교가 없어도 시험이 통과함(26라운드 보이스 C 재확인 N2). 운영 토큰도
	// 둘 다 같은 길이(43자)라 내용으로만 갈림.
	a092ModeToken  = "mode-control-token-0123456789abcdef"
	a092AlertToken = "alrt-control-token-0123456789abcdef"
	a092GoodBody   = `{"to":"normal","operator":"박지훈","approval":"OPS-1","reason":"checked"}`
)

func init() {
	if len(a092ModeToken) != len(a092AlertToken) || a092ModeToken == a092AlertToken {
		panic("a092 endpoint fixture: the two tokens must differ only in content")
	}
}

func a092ModeRequest(method, token, body string) *http.Request {
	r := httptest.NewRequest(method, ModeControlReleasePath, strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	if token != "" {
		r.Header.Set("Authorization", "Bearer "+token)
	}
	return r
}

func a092Serve(fx *a092ReleaseFixture, r *http.Request) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	modeControlRoutes(a092ModeToken, fx.ops).ServeHTTP(w, r)
	return w
}

func a092ModeNow(t *testing.T, fx *a092ReleaseFixture) string {
	t.Helper()
	cur, err := fx.j.CurrentOperatingMode(context.Background(), a092Account)
	if err != nil {
		t.Fatal(err)
	}
	return cur.Mode
}

func TestA092TheModeEndpointRejectsWhatItMustNotRun(t *testing.T) {
	cases := []struct {
		name   string
		req    *http.Request
		status int
	}{
		{"no-token", a092ModeRequest(http.MethodPost, "", a092GoodBody), http.StatusUnauthorized},
		{"alert-control-token", a092ModeRequest(http.MethodPost, a092AlertToken, a092GoodBody), http.StatusUnauthorized},
		{"get", a092ModeRequest(http.MethodGet, a092ModeToken, a092GoodBody), http.StatusMethodNotAllowed},
		{"unknown-field", a092ModeRequest(http.MethodPost, a092ModeToken,
			`{"to":"normal","operator":"박지훈","approval":"OPS-1","reason":"checked","force":true}`), http.StatusBadRequest},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fx := a092ReleaseEngine(t)
			w := a092Serve(fx, c.req)
			if w.Code != c.status {
				t.Errorf("status = %d, want %d (body %s)", w.Code, c.status, w.Body.String())
			}
			if mode := a092ModeNow(t, fx); mode != journal.ModeEntryBlocked {
				t.Errorf("mode = %s — a rejected request reached the ledger", mode)
			}
			if len(fx.auditor.calls) != 0 {
				t.Errorf("audit calls = %v — a rejected request reached the ledger", fx.auditor.calls)
			}
		})
	}
	t.Run("control", func(t *testing.T) {
		fx := a092ReleaseEngine(t)
		w := a092Serve(fx, a092ModeRequest(http.MethodPost, a092ModeToken, a092GoodBody))
		if w.Code != http.StatusOK || a092ModeNow(t, fx) != journal.ModeNormal {
			t.Fatalf("status = %d mode = %s — the well-formed request must release (otherwise the rejections prove nothing)",
				w.Code, a092ModeNow(t, fx))
		}
	})
}

func TestA092AnEngineFailureBodyCarriesNoLedgerText(t *testing.T) {
	fx := a092ReleaseEngine(t)
	_ = fx.j.Close() // 원장 고장 — 전이 자체가 실패
	w := a092Serve(fx, a092ModeRequest(http.MethodPost, a092ModeToken, a092GoodBody))
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500 (body %s)", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if strings.Contains(body, a092Account) || strings.Contains(body, "journal:") || strings.Contains(body, "sql:") {
		t.Errorf("the 500 body carries raw ledger text: %s", body)
	}
}
