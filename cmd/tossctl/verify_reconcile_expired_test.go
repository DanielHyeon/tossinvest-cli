package main

// verify_reconcile_expired_test.go — a164 tasks 2.5(cmd HOLD)·2.6(by-id 신원 재확인)·2.7(명령 표면·승인 질문).
// 전 경로 시험은 verify_reconcile_expired_seams_test.go(tossos_testseams 태그)에 있다.

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
)

// expiredDetailBody 는 0.6 영수증 ① 값을 계약 모양(ConditionalOrderDetailResponse)에 넣은 by-id 응답이다.
func expiredDetailBody(id, symbol string) string {
	return `{"result":{"conditionalOrderId":"` + id + `","type":"SINGLE","status":"EXPIRED","symbol":"` + symbol + `",` +
		`"market":"KR","quantity":"1","orderType":"MARKET","expireDate":"2026-08-07",` +
		`"first":{"type":"STOP","status":"EXPIRED","orderSide":"SELL","triggerPrice":"3200","targetProfitRate":null,"orderPrice":null,"triggeredOrderId":null},` +
		`"second":null,"createdAt":"2026-07-31T18:49:05+09:00"}}`
}

// TestVerifyReconcileExpiredIsAFlagOfTheSameMutatingCommand 는 design D4 마지막 줄 · tasks 2.7 이다 — `--expired` 는 같은
// 명령의 bool 플래그이고(별도 하위 명령 없음), 명령은 mutating=true 를 유지한다.
func TestVerifyReconcileExpiredIsAFlagOfTheSameMutatingCommand(t *testing.T) {
	root := newRootCmd()
	cmd, _, err := root.Find([]string{"verify", "reconcile"})
	if err != nil || cmd == nil || cmd.Name() != "reconcile" {
		t.Fatalf("verify reconcile: %v", err)
	}
	flag := cmd.Flags().Lookup("expired")
	if flag == nil || flag.Value.Type() != "bool" || flag.DefValue != "false" {
		t.Fatalf("--expired flag %+v, want a bool defaulting to false", flag)
	}
	if len(cmd.Commands()) != 0 {
		t.Fatalf("verify reconcile has subcommands %v — the expiry mode is a flag", cmd.Commands())
	}
	for _, sub := range cmd.Parent().Commands() {
		if strings.Contains(sub.Name(), "expir") {
			t.Fatalf("verify has a separate %q command", sub.Name())
		}
	}
	if cmd.Annotations["mutating"] != "true" {
		t.Fatalf("annotations %v", cmd.Annotations)
	}
}

// TestReconcileApprovalQuestionNamesTheMode 는 P3-6 의 질문 쪽이다 — 만료 모드는 「reconciled expired 로 기록할까?」,
// 부재 모드는 a121 그대로.
func TestReconcileApprovalQuestionNamesTheMode(t *testing.T) {
	for mode, want := range map[string]string{"": "reconciled absent 로 기록할까?", verifylive.ReconcileModeExpired: "reconciled expired 로 기록할까?"} {
		out := setReconcileTerminal(t, "y\n", true)
		a := reconcileApprovalFixture()
		a.Mode = mode
		if err := verifyReconcileApprove(context.Background(), a); err != nil {
			t.Fatalf("mode %q: %v", mode, err)
		}
		if !strings.Contains(out.String(), want) {
			t.Fatalf("mode %q question %q lacks %q", mode, out.String(), want)
		}
	}
}

// TestReconcileByIDReadRechecksIdentityBeforeItIsSent 는 P2-4 다 — 만료 모드의 첫 계좌 범위 호출인 by-id 도 읽기용
// 클라이언트의 Accounts 재조회 **뒤**에만 나가고, 신원이 바뀌었으면 나가지 않는다.
func TestReconcileByIDReadRechecksIdentityBeforeItIsSent(t *testing.T) {
	for name, second := range map[string]string{
		"same":          accountsBody(accountRow("123-45-678901", 5)),
		"other-account": accountsBody(accountRow("999-99-998901", 5)),
	} {
		t.Run(name, func(t *testing.T) {
			var mu sync.Mutex
			var seen []string
			accounts := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				seen = append(seen, r.Method+" "+r.URL.Path)
				mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.URL.Path == "/oauth2/token":
					fmt.Fprint(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
				case r.URL.Path == "/api/v1/accounts":
					mu.Lock()
					accounts++
					n := accounts
					mu.Unlock()
					if n == 1 {
						fmt.Fprint(w, accountsBody(accountRow("123-45-678901", 5)))
					} else {
						fmt.Fprint(w, second)
					}
				case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/api/v1/conditional-orders/"):
					fmt.Fprint(w, expiredDetailBody("CO-1", "005930"))
				default:
					w.WriteHeader(http.StatusNotFound)
				}
			}))
			t.Cleanup(srv.Close)
			reader, _, err := newVerifyReconcileReader(context.Background(), official.Credentials{APIKey: "k", SecretKey: "s"},
				filepath.Join(t.TempDir(), "token.json"), official.WithBaseURL(srv.URL), official.WithHTTPClient(srv.Client()))
			if err != nil {
				t.Fatal(err)
			}
			got, err := reader.ReconcileConditionalOrder(context.Background(), "CO-1")
			mu.Lock()
			defer mu.Unlock()
			var order []string
			for _, s := range seen {
				if s != "POST /oauth2/token" {
					order = append(order, s)
				}
			}
			if name == "same" {
				if err != nil || got.Status != "EXPIRED" || got.ID != "CO-1" {
					t.Fatalf("by-id %+v %v", got, err)
				}
				if want := []string{"GET /api/v1/accounts", "GET /api/v1/accounts", "GET /api/v1/conditional-orders/CO-1"}; strings.Join(order, ",") != strings.Join(want, ",") {
					t.Fatalf("request order %v, want %v", order, want)
				}
				return
			}
			if !errors.Is(err, errReconcileIdentityChanged) {
				t.Fatalf("a changed identity must stop the by-id read; got %v", err)
			}
			for _, s := range order {
				if strings.HasPrefix(s, "GET /api/v1/conditional-orders/") {
					t.Fatalf("the by-id read went out under a changed identity: %v", order)
				}
			}
		})
	}
}

// TestM0TriggerHoldFollowsAReconciledExpiredArtifact 는 tasks 2.5 의 cmd 쪽 HOLD(validateM0TriggerMode B9)다 — 미결 조건주문은
// 막고, 같은 기록에 reconciled-expired 줄이 붙으면 그 HOLD 가 풀린다.
func TestM0TriggerHoldFollowsAReconciledExpiredArtifact(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	opts := &verifyOptions{market: verifylive.MarketKR, includeTrigger: true, confirmEach: true, resume: true,
		redo: []string{string(verifylive.StepConditionalTrigger)}, triggerReceiptDir: dir}
	var prior []verifylive.Entry
	for _, step := range verifylive.Steps() {
		if step.ID != verifylive.StepConditionalTrigger {
			prior = append(prior, verifylive.Entry{Kind: verifylive.KindStep, StepID: step.ID, Verdict: verifylive.VerdictPass})
		}
	}
	art := verifylive.Artifact{Kind: verifylive.KindConditional, ID: "CO-EXPIRED", Symbol: "333430", Deliberate: true}
	prior = append(prior, verifylive.Entry{Kind: verifylive.KindStep, StepID: verifylive.StepConditionalRegister,
		Verdict: verifylive.VerdictPass, Artifacts: []verifylive.Artifact{art}})
	if err := validateM0TriggerMode(opts, prior); err == nil || !strings.Contains(err.Error(), "outstanding") {
		t.Fatalf("control: an outstanding conditional must hold --include-trigger; got %v", err)
	}
	ended := art
	ended.ReconciledExpired = true
	prior = append(prior, verifylive.Entry{Kind: verifylive.KindReconcile, StepID: verifylive.StepReconcileExpired,
		Artifacts: []verifylive.Artifact{ended}})
	if err := validateM0TriggerMode(opts, prior); err != nil {
		t.Fatalf("a reconciled-expired artifact still holds --include-trigger: %v", err)
	}
}
