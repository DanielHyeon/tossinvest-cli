package execgw_test

// a094 §2 (R1) — 브로커가 이름을 준 거절은 code 가 분류한다.
//
// 사건(2026-08-07): 브로커가 openapi 의 422 대신 409 로 `opposite-pending-order-exists` 를 돌려줬고, 상태 코드 표가 409 를
// 모호로 분류해 attempt 가 IN_DOUBT 로 얼었음. 계약과 실물이 일치한 필드는 본문의 code 하나뿐이었음.
//
// 이 파일은 게이트웨이 끝에서 끝까지(공식 클라이언트 + httptest) 잰다 — 분류기만 부르면 classifyMutation 이 그것을
// 부르는 자리(순서)를 못 잰다.

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"net/http"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a094ProductionBody 는 원장에 남은 프로덕션 409 본문의 모양임(requestId 는 시험용 값).
const a094ProductionBody = `{"error":{"requestId":"a094-test","code":"opposite-pending-order-exists","message":"반대 포지션 미체결 주문이 존재합니다."}}`

// a094Place 는 status · 본문 하나로 발주를 한 번 돌리고 결과와 원장의 미종결 attempt 수를 돌려줌.
func a094Place(t *testing.T, status int, body string) (execgw.Outcome, int, int) {
	t.Helper()
	gw, j, clk, posts := officialGateway(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
	out, _ := gw.Place(context.Background(), placeRequest(t, j, clk))
	pending, err := j.PendingAttempts(context.Background())
	if err != nil {
		t.Fatalf("PendingAttempts: %v", err)
	}
	return out, len(pending), *posts
}

// 2.1 · 2.2 · 2.5a · 2.5b · 2.5c · 2.3 · 2.4 · 2.5 · 2.5d · 2.5e · 2.5f · 2.5g — 세 결과의 표(확정 거절 · 모호 강제 · 판정 없음).
func TestA094RefusalCodeClassifiesTheAttempt(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   journal.AttemptState
		reason execgw.ReasonCode // 빈 값이면 사유는 단언하지 않음
	}{
		// 확정 거절 — 목록 안 · 모순 없음.
		{"2.1 production 409 with error.code", 409, a094ProductionBody, journal.StateFailedConfirmed, execgw.ReasonOppositePendingOrder},
		{"2.2 contract 422 with the same code", 422, a094ProductionBody, journal.StateFailedConfirmed, execgw.ReasonOppositePendingOrder},
		{"2.5b code at the top level", 409, `{"code":"opposite-pending-order-exists","message":"x"}`, journal.StateFailedConfirmed, execgw.ReasonOppositePendingOrder},
		{"2.5c upper case is the same code", 409, `{"error":{"code":"OPPOSITE-PENDING-ORDER-EXISTS"}}`, journal.StateFailedConfirmed, execgw.ReasonOppositePendingOrder},
		{"both places carry the same code", 409, `{"code":"opposite-pending-order-exists","error":{"code":"Opposite-Pending-Order-Exists"}}`, journal.StateFailedConfirmed, execgw.ReasonOppositePendingOrder},

		// 판정 없음 — 종전 경로(409 는 모호, 422 는 확정 거절).
		{"2.3 409 request-in-progress stays ambiguous", 409, `{"error":{"code":"request-in-progress"}}`, journal.StateInDoubt, ""},
		{"2.4 409 without a code stays ambiguous", 409, `{"error":{"message":"conflict"}}`, journal.StateInDoubt, ""},
		{"2.5 the phrase only in message is not a code", 409, `{"error":{"code":"something-else","message":"opposite-pending-order-exists"}}`, journal.StateInDoubt, ""},
		{"2.5c a suffixed value is not the code", 409, `{"error":{"code":"opposite-pending-order-exists-v2"}}`, journal.StateInDoubt, ""},
		{"2.5c a padded value is not the code", 409, `{"error":{"code":" opposite-pending-order-exists"}}`, journal.StateInDoubt, ""},
		{"2.5d not JSON", 409, `opposite-pending-order-exists`, journal.StateInDoubt, ""},
		{"2.5d empty code", 409, `{"error":{"code":""}}`, journal.StateInDoubt, ""},
		{"2.5d null code", 409, `{"code":null,"error":{"code":null}}`, journal.StateInDoubt, ""},
		{"2.5d a code that is not a string", 409, `{"code":7,"error":{"code":"opposite-pending-order-exists"}}`, journal.StateInDoubt, ""},
		{"2.5g no verdict on a 422 keeps the definitive status", 422, `{"error":{"code":"something-else"}}`, journal.StateFailedConfirmed, ""},

		// 모호 강제 — 두 자리의 값이 다름. 422 여도 확정 거절이 아님(2.5f).
		{"2.5e contradictory places on a 409", 409, `{"code":"opposite-pending-order-exists","error":{"code":"request-in-progress"}}`, journal.StateInDoubt, ""},
		{"2.5f contradictory places on a 422", 422, `{"code":"opposite-pending-order-exists","error":{"code":"request-in-progress"}}`, journal.StateInDoubt, ""},
		{"2.5f contradictory, neither listed, on a 422", 422, `{"code":"a","error":{"code":"b"}}`, journal.StateInDoubt, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, pending, posts := a094Place(t, tc.status, tc.body)
			if out.State != tc.want {
				t.Fatalf("state = %s, want %s (reason %s, detail %s)", out.State, tc.want, out.Reason, out.Detail)
			}
			if tc.reason != "" && out.Reason != tc.reason {
				t.Errorf("reason = %s, want %s", out.Reason, tc.reason)
			}
			wantPending := 0
			if !tc.want.IsTerminal() {
				wantPending = 1
			}
			if pending != wantPending {
				t.Errorf("unsettled attempts = %d, want %d — a definitive refusal must leave the symbol free", pending, wantPending)
			}
			if posts != 1 {
				t.Errorf("mutation POSTs = %d, want exactly 1", posts)
			}
		})
	}
}

// 2.7 — 기존 세 code 의 분류 무변화: 최상위 code 모양의 기존 fixture 는 이 분류기의 목록 밖이라 판정 없음으로 가고,
// 종전의 문구 분류(ClassifyBrokerRefusal)가 그대로 잡음. status 는 400 — 공식 클라이언트는 401·403 을 본문 없는 인증
// sentinel 로 바꾸므로(official/errors.go classifyStatus) 본문이 분류기에 닿는 것은 그 밖의 4xx 뿐임.
func TestA094ExistingRefusalCodesAreUnchanged(t *testing.T) {
	for _, tc := range []struct {
		name   string
		status int
		body   string
		reason execgw.ReasonCode
	}{
		{"trade auth", 400, `{"code":"TRADE_AUTH_REQUIRED","message":"거래 인증이 필요합니다"}`, execgw.ReasonInteractiveAuthRequired},
		{"fx consent", 400, `{"code":"FX_CONSENT_REQUIRED","message":"환전 동의"}`, execgw.ReasonFXConsentRequired},
		{"funding", 400, `{"code":"FUNDING_REQUIRED","message":"입금"}`, execgw.ReasonFundingRequired},
	} {
		t.Run(tc.name, func(t *testing.T) {
			out, _, _ := a094Place(t, tc.status, tc.body)
			if out.State != journal.StateFailedConfirmed || out.Reason != tc.reason {
				t.Errorf("state/reason = %s/%s, want FAILED_CONFIRMED/%s", out.State, out.Reason, tc.reason)
			}
		})
	}
}

// 2.6 — 상태 코드 확정 거절 목록 무변화(409 는 여전히 목록 밖).
func TestA094TheStatusTableIsUnchanged(t *testing.T) {
	for status, want := range map[int]journal.DispatchClass{
		400: journal.DispatchRejected, 401: journal.DispatchRejected, 403: journal.DispatchRejected,
		404: journal.DispatchRejected, 405: journal.DispatchRejected, 415: journal.DispatchRejected,
		422: journal.DispatchRejected,
		409: journal.DispatchAmbiguous, 408: journal.DispatchAmbiguous, 429: journal.DispatchAmbiguous,
		500: journal.DispatchAmbiguous, 503: journal.DispatchAmbiguous,
	} {
		got := journal.ClassifyHTTPMutation(journal.SendComplete, status, nil)
		if got.Class != want {
			t.Errorf("status %d → %s, want %s", status, got.Class, want)
		}
	}
}

// 2.11 — 재생 경계는 구조로 고정함: classifyReplay 는 code 분류기도 dispatch 분류기도 부르지 않음.
// 재생에서 같은 code 는 재생 요청의 거절일 뿐 원 주문에 대해 아무것도 증명하지 않음(정본 SHALL NOT).
func TestA094TheReplayClassifierNeverReadsTheRefusalCode(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "replay.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Name.Name != "classifyReplay" {
			continue
		}
		found = true
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := ""
			switch fn := call.Fun.(type) {
			case *ast.Ident:
				name = fn.Name
			case *ast.SelectorExpr:
				name = fn.Sel.Name
			}
			switch name {
			case "classifyRefusalCode", "classifyMutation", "ClassifyBrokerRefusal", "ClassifyHTTPMutation":
				t.Errorf("classifyReplay calls %s at %s — a replay answer must not be classified by the dispatch "+
					"classifiers (order-execution: 이 분류를 재생 응답에 적용해서는 안 된다)", name, fset.Position(call.Pos()))
			}
			return true
		})
	}
	if !found {
		t.Fatal("control: classifyReplay not found in replay.go — the pin is blind")
	}
}
