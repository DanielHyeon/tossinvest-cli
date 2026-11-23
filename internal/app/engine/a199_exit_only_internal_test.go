package engine

// a199_exit_only_internal_test.go — 인터록 · 감시의 내부 판정(a199).
//
//   - 낡음 + 게이트웨이 배선 결함 · 키 못 싣는 transport → 그 조항의 오류만(NewContext 로는 만들 수 없는 조항이라 verifyGate 를 직접 부름).
//   - 감시 판정은 인터록과 같은 닫힌 집합: 만료가 가린 다른 계좌는 expired 가 아니라 account_mismatch(D7 · D1a).
//   - **투영용 조립**(NewPairedStrategyEntryProductionAssembly → productionStrategyWorker)의 worker 조건에 `c.Automation.Verified`
//     항이 남아 있음을 구조로 못 박음. 이 조립은 상태 투영용이고 실행되는 런타임 worker 가 아님(구현 리뷰 r1 처분 1 정정) —
//     실제 런타임 worker(NewRefreshingPairedStrategyEntrySupervisor)는 Verified 를 읽지 않고, 청산 전용에서 진입을 막는 것은
//     진입 래치 하나다(a199_strategy_dispatch_latch_test.go 가 실제 dispatch 경로로 잼).

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

var a199Now = time.Date(2026, 11, 24, 3, 0, 0, 0, time.UTC)

// a199LimitGuardian 은 감사 한도와 같은 출처를 진술하는 최소 Guardian(조항 1 · 7 · 7b · 7c 통과용).
type a199LimitGuardian struct{ limits execgw.Limits }

func (a199LimitGuardian) Authorize(context.Context, execgw.AuthorizationRequest) (execgw.GuardianDecision, error) {
	return execgw.GuardianDecision{}, errors.New("not reached")
}
func (g a199LimitGuardian) ExposureLimits() execgw.Limits { return g.limits }
func (a199LimitGuardian) DailyBuyNotionalLimit() riskcalc.Money {
	return riskcalc.Money{Amount: "10000000", Currency: "KRW"}
}
func (a199LimitGuardian) MaxOpenPositionsLimit() int { return 6 }

func a199StaleStatus(t *testing.T) AutomationStatus {
	t.Helper()
	gate := config.AutomationGate{
		Enabled: true, MaxOrderQuantity: 10, MaxOrderNotional: 1_000_000, MaxTotalExposure: 5_000_000,
		MaxDailyLossAmount: 200_000, MaxDailyLossRatio: 0.02, LimitCurrency: "KRW",
	}
	dir := t.TempDir()
	path := filepath.Join(dir, attest.FileName)
	if err := attest.Save(path, attest.Attestation{
		FormatVersion: attest.FormatVersion, AccountRef: "123-45",
		IssuedAt: a199Now.Add(-31 * 24 * time.Hour), ExpiresAt: a199Now.Add(-time.Hour),
		Endpoints: RequiredEndpoints(),
	}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	gate.AttestationFile = path
	status := newAutomationStatus(gate, config.Paths{ConfigDir: dir}, ProtectionUnwired)
	status.AccountRef = "123-45"
	return status
}

func TestA199StaleWithAnUnwiredGatewayRefusesByTheGatewayClauseOnly(t *testing.T) {
	status := a199StaleStatus(t)
	err := verifyGate(&status, gateFacts{
		guardian: a199LimitGuardian{limits: status.Limits},
		trading:  config.Trading{Place: true, Sell: true, Cancel: true, AllowLiveOrderActions: true},
		gateway:  nil,
		now:      a199Now,
	})
	if !errors.Is(err, ErrGatewayRequired) {
		t.Fatalf("err = %v, want the gateway clause", err)
	}
	if errors.Is(err, attest.ErrExpired) {
		t.Fatalf("err = %v also carries the stale sentinel", err)
	}
}

func TestA199StaleWithAKeylessTransportRefusesByTheTransportClauseOnly(t *testing.T) {
	status := a199StaleStatus(t)
	j := openTestJournal(t)
	gw, err := execgw.New(execgw.Options{
		Journal: j, Trading: trading.NewService(config.Trading{}, &officialBroker{}), AccountRef: "123-45",
		Orders: execgw.OfficialOrders{}, Entry: execgw.NewEntryGate(nil, map[execgw.RequiredQuery]time.Duration{}),
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	err = verifyGate(&status, gateFacts{
		guardian:  a199LimitGuardian{limits: status.Limits},
		trading:   config.Trading{Place: true, Sell: true, Cancel: true, AllowLiveOrderActions: true},
		gateway:   gw,
		transport: keylessBroker{},
		now:       a199Now,
	})
	if !errors.Is(err, ErrKeylessTransport) {
		t.Fatalf("err = %v, want the transport clause", err)
	}
	if errors.Is(err, attest.ErrExpired) {
		t.Fatalf("err = %v also carries the stale sentinel", err)
	}
}

func TestA199TheWatchJudgesAnExpiredForeignAttestationAsTheInterlockDoes(t *testing.T) {
	att := a135Attestation(-time.Hour) // 만료
	v := judgeAttestation(att, nil, a135Now, "111-22-333444")
	if v.Stage != stageAccountMismatch {
		t.Fatalf("stage = %s, want account_mismatch — the interlock refuses this file, it does not go exit-only", v.Stage)
	}
	blank := a135Attestation(-time.Hour)
	blank.Endpoints = append(RequiredEndpoints(), " ")
	if v := judgeAttestation(blank, nil, a135Now, a135Account); v.Stage == stageExpired || v.Stage == stageScopeLost {
		t.Fatalf("stage = %s for a stale file with a blank endpoint item — the interlock refuses it", v.Stage)
	}
}

// TestA199TheProjectionAssemblyWorkerConditionStillReadsVerified 는 투영용 조립의 조건만 잰다 — 실행되는 런타임 worker 의 차단이
// 아님(그것은 진입 래치, a199_strategy_dispatch_latch_test.go).
func TestA199TheProjectionAssemblyWorkerConditionStillReadsVerified(t *testing.T) {
	src, err := os.ReadFile("strategy_entry_supervisor.go")
	if err != nil {
		t.Fatal(err)
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "strategy_entry_supervisor.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Name.Name != "productionStrategyWorker" {
			continue
		}
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			if id, ok := call.Fun.(*ast.Ident); !ok || id.Name != "buildProductionStrategyMarketWorker" || len(call.Args) < 4 {
				return true
			}
			// 넷째 인자(wiringReady)의 && 사슬에 `c.Automation.Verified` 가 그대로 있어야 함 — 청산 쪽 표지가 아님.
			ast.Inspect(call.Args[3], func(m ast.Node) bool {
				sel, ok := m.(*ast.SelectorExpr)
				if ok && sel.Sel.Name == "Verified" {
					if inner, ok := sel.X.(*ast.SelectorExpr); ok && inner.Sel.Name == "Automation" {
						found = true
					}
				}
				return true
			})
			return false
		})
	}
	if !found {
		t.Fatal("the projection assembly's productionStrategyWorker no longer requires c.Automation.Verified — on an exit-only " +
			"start the Guardian is non-nil, so the projected worker would be reported as wired (the runtime worker is stopped by the latch)")
	}
}
