package engine

// a094 4.T · 4.Ta — park 해동 명령의 엔진 절반(제어 endpoint 를 거친 실제 원장 · 실제 audit 로그).

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attemptthaw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicyrpc"
)

type a094ThawFixture struct {
	dir, auditLog, positionID string
	j                         *journal.Journal
	client                    *positionpolicyrpc.Client
}

// a094ThawEngine 은 열린 포지션 하나에 손절 발의를 무장하고 그 발주 attempt 를 attemptState 로 남긴 엔진을 세움.
func a094ThawEngine(t *testing.T, withAudit bool, attemptState journal.AttemptState) *a094ThawFixture {
	t.Helper()
	ctx := context.Background()
	dir, err := os.MkdirTemp("", "a094-thaw-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	j, err := journal.Open(ctx, journal.Options{
		Path:     filepath.Join(dir, journal.DBFileName),
		Clock:    clock.NewFake(a066RelaxNow),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	if err := j.SetApplyHooks(journal.ApplyHooks{Project: journal.ProjectPosition, Exit: journal.ApplyExitFill}); err != nil {
		t.Fatal(err)
	}
	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	// 체결된 진입 하나 — 진입 결정이 있어야 exit 정책 대상임.
	limits, err := execgw.EncodeLimits(execgw.Limits{
		MaxQuantity: execgw.Bound(1000), MaxNotional: execgw.Bound(1e9),
		MaxTotalExposure: execgw.Bound(1e9), MaxDailyLossAmount: execgw.Bound(1e6),
		MaxDailyLossRatio: execgw.Bound(0.02), Currency: "KRW",
	})
	must(err)
	_, err = j.RecordDecision(ctx, journal.DecisionRequest{
		ID: "d-entry", AccountRef: "acct-7", SafetyClass: journal.SafetyClassExposureRaising, Kind: journal.KindPlace,
		Preimage: journal.RiskIntent{AccountRef: "acct-7", Market: "kr", Symbol: "005930", Side: "BUY", Quantity: "10",
			EntryPrice: "70000", StopPrice: "68000", TargetPrice: "999999", PolicyVersion: "test/v1"},
		LimitsJSON: limits, Nonce: "nonce-d-entry", IssuedAt: a066RelaxNow, ExpiresAt: a066RelaxNow.Add(time.Hour),
	})
	must(err)
	entry, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{ID: "i-entry", Market: "kr", TradingDay: "2026-03-30", AccountRef: "acct-7", Symbol: "005930",
			Side: "BUY", OrderType: "LIMIT", TimeInForce: "DAY", Quantity: "10", Price: "70000", Currency: "KRW",
			Source: "engine/test", Fingerprint: "fp-entry"},
		Kind: journal.KindPlace, AttemptID: "a-entry", AccountRef: "acct-7", DecisionID: "d-entry",
		SafetyClass: journal.SafetyClassExposureRaising, ClientOrderID: journal.DeriveClientOrderID("d-entry", 0),
	})
	must(err)
	must(entry.MarkDispatchStarted(ctx))
	must(entry.MarkAcked(ctx, "O-entry"))
	must(entry.Settle(ctx, journal.StateConfirmed, "broker_accepted", ""))
	_, err = j.RecordFill(ctx, journal.FillObservation{OrderID: "O-entry", Symbol: "005930", Market: "kr", AccountRef: "acct-7",
		TradingDay: "2026-03-30", Side: "BUY", State: "CLOSED_FILLED", Terminal: true, Quantity: "10", FilledQuantity: "10",
		AveragePrice: "70000", ObservedAt: "2026-03-30T00:30:00Z"})
	must(err)
	p, err := j.CurrentPosition(ctx, "acct-7", "kr", "005930")
	must(err)
	_, err = j.OpenExitState(ctx, journal.ExitStateSeed{PositionID: p.ID, EntryPrice: "70000", InitialStop: "68000"})
	must(err)
	must(j.RecordExitJudgement(ctx, journal.ExitJudgement{PositionID: p.ID, ObservedPrice: "67900", HighWater: "70000",
		Baseline: "68000", RatchetLevel: journal.RatchetNone, ActiveRung: exitpolicy.NoRung,
		Proposal: &journal.ExitProposal{Action: string(exitpolicy.ActionBaselineBreach), Level: "NONE", IntentID: "exit-stop"}}))
	// 손절 발주 attempt — 409 로 모호 → park.
	stop, err := j.Prepare(ctx, journal.PrepareRequest{
		Intent: journal.Intent{ID: "exit-stop", Market: "kr", TradingDay: "2026-03-30", AccountRef: "acct-7", Symbol: "005930",
			Side: "SELL", OrderType: "LIMIT", TimeInForce: "DAY", Quantity: "10", Price: "67900", Currency: "KRW",
			Source: "engine/exit", Fingerprint: "fp-stop"},
		Kind: journal.KindPlace, AttemptID: "a-stop",
	})
	must(err)
	must(stop.MarkDispatchStarted(ctx))
	must(stop.MarkInDoubt(ctx, "dispatch_outcome_unknown", "409"))
	if attemptState == journal.StateUnresolvedInDoubt {
		must(stop.ResolveUnresolved(ctx, "in_doubt_unresolved", "parked"))
	}

	fx := &a094ThawFixture{dir: dir, auditLog: filepath.Join(dir, "audit.log"), positionID: p.ID, j: j}
	ectx := &Context{Journal: j}
	if withAudit {
		log, err := audit.Open(audit.Options{Path: fx.auditLog, Subject: "engine"})
		must(err)
		ectx.Audit = log
	}
	service, err := NewPositionPolicyCommandService(ectx, clock.NewFake(a066RelaxNow))
	must(err)
	server, err := StartPositionPolicyCommandServer(dir, service)
	must(err)
	t.Cleanup(func() { _ = server.Close() })
	client, err := positionpolicyrpc.Dial(ctx, positionpolicyrpc.DescriptorPath(dir))
	must(err)
	fx.client = client
	return fx
}

func (fx *a094ThawFixture) req(target, order string) attemptthaw.Request {
	return attemptthaw.Request{AttemptID: "a-stop", Target: target, BrokerOrderID: order,
		Operator: "운영자", Approval: "OPS-2026-0930-1", Note: "브로커 미체결·체결 목록에 그 주문 없음"}
}

func (fx *a094ThawFixture) state(t *testing.T) (journal.AttemptState, bool) {
	t.Helper()
	rec, err := fx.j.LookupAttempt(context.Background(), "a-stop")
	if err != nil {
		t.Fatal(err)
	}
	st, err := fx.j.ExitState(context.Background(), fx.positionID)
	if err != nil {
		t.Fatal(err)
	}
	return rec.State, st.Pending()
}

// 4.T — 비수용으로 닫으면: audit 가 먼저, attempt 종결, 같은 명령 안에서 발의 해제.
func TestA094AnOperatorThawClosesTheParkAndReleasesTheProposal(t *testing.T) {
	fx := a094ThawEngine(t, true, journal.StateUnresolvedInDoubt)
	result, err := fx.client.ResolveParkedAttempt(context.Background(), fx.req("FAILED_CONFIRMED", ""))
	if err != nil {
		t.Fatalf("ResolveParkedAttempt: %v", err)
	}
	state, armed := fx.state(t)
	if state != journal.StateFailedConfirmed || armed || !result.ProposalReleased || result.PositionID != fx.positionID {
		t.Fatalf("state %s armed %v result %+v, want FAILED_CONFIRMED + released", state, armed, result)
	}
	raw, err := os.ReadFile(fx.auditLog)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), AuditActionAttemptThaw) || !strings.Contains(string(raw), "OPS-2026-0930-1") {
		t.Errorf("the audit log does not carry the thaw and its approval: %s", raw)
	}
	rec, _ := fx.j.LookupAttempt(context.Background(), "a-stop")
	if !strings.Contains(rec.Detail, "운영자") || rec.ReasonCode != journal.ReasonOperatorResolved {
		t.Errorf("the ledger does not name the operator: %+v", rec)
	}
}

// 4.T — audit 로그가 없으면 아무것도 바뀌지 않음.
func TestA094AThawWithoutAnAuditLogChangesNothing(t *testing.T) {
	fx := a094ThawEngine(t, false, journal.StateUnresolvedInDoubt)
	_, err := fx.client.ResolveParkedAttempt(context.Background(), fx.req("FAILED_CONFIRMED", ""))
	if !errors.Is(err, attemptthaw.ErrAuditUnavailable) {
		t.Fatalf("err = %v, want ErrAuditUnavailable", err)
	}
	if state, armed := fx.state(t); state != journal.StateUnresolvedInDoubt || !armed {
		t.Fatalf("state %s armed %v, want untouched", state, armed)
	}
}

// 4.T — park 가 아닌 attempt 는 거절(stale), 아무것도 바뀌지 않음.
func TestA094AThawOfAnUnparkedAttemptIsStale(t *testing.T) {
	fx := a094ThawEngine(t, true, journal.StateInDoubt)
	_, err := fx.client.ResolveParkedAttempt(context.Background(), fx.req("FAILED_CONFIRMED", ""))
	if !errors.Is(err, attemptthaw.ErrStale) {
		t.Fatalf("err = %v, want ErrStale", err)
	}
	if state, armed := fx.state(t); state != journal.StateInDoubt || !armed {
		t.Fatalf("state %s armed %v, want untouched", state, armed)
	}
	// stale 은 audit 줄 **앞**에서 거절 — audit 기록은 실제로 시도한 해소만 담는다(변이 M27).
	if raw, _ := os.ReadFile(fx.auditLog); strings.Contains(string(raw), AuditActionAttemptThaw) {
		t.Errorf("a stale thaw left an audit line: %s", raw)
	}
}

// 4.T — 필수 입력 누락은 거절, 접수 확정은 주문 번호 필수 · 발의를 풀지 않음.
func TestA094AThawNeedsItsFieldsAndConfirmKeepsTheProposal(t *testing.T) {
	fx := a094ThawEngine(t, true, journal.StateUnresolvedInDoubt)
	for _, bad := range []attemptthaw.Request{
		{AttemptID: "a-stop", Target: "FAILED_CONFIRMED", Operator: "op", Approval: "", Note: "n"},
		{AttemptID: "a-stop", Target: "FAILED_CONFIRMED", Operator: "", Approval: "a", Note: "n"},
		{AttemptID: "a-stop", Target: "FAILED_CONFIRMED", Operator: "op", Approval: "a", Note: ""},
		{AttemptID: "a-stop", Target: "CONFIRMED", Operator: "op", Approval: "a", Note: "n"},
		{AttemptID: "a-stop", Target: "NOT_DISPATCHED", Operator: "op", Approval: "a", Note: "n"},
	} {
		if _, err := fx.client.ResolveParkedAttempt(context.Background(), bad); !errors.Is(err, attemptthaw.ErrInvalidRequest) {
			t.Errorf("%+v: err = %v, want ErrInvalidRequest", bad, err)
		}
	}
	if state, armed := fx.state(t); state != journal.StateUnresolvedInDoubt || !armed {
		t.Fatalf("an invalid request changed state %s armed %v", state, armed)
	}
	result, err := fx.client.ResolveParkedAttempt(context.Background(), fx.req("CONFIRMED", "O-found"))
	if err != nil {
		t.Fatalf("confirm: %v", err)
	}
	if state, armed := fx.state(t); state != journal.StateConfirmed || !armed || result.ProposalReleased {
		t.Fatalf("confirm: state %s armed %v released %v, want CONFIRMED with the proposal still armed", state, armed, result.ProposalReleased)
	}
}

// 4.Ta — 두 쓰기 사이 충돌(해소는 커밋, 해제 실패): 명령은 해제 실패를 말하고, 다음 기동 따라잡기가 발의를 푼다.
func TestA094AThawWhoseReleaseFailsIsCaughtUpAtBoot(t *testing.T) {
	fx := a094ThawEngine(t, true, journal.StateUnresolvedInDoubt)
	db, err := sql.Open("sqlite", "file:"+fx.j.Path())
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(fmt.Sprintf(`CREATE TRIGGER a094_refuse_release BEFORE UPDATE ON exit_states
		WHEN OLD.position_id = '%s' AND NEW.pending_action IS NULL BEGIN SELECT RAISE(ABORT, 'a094 fixture'); END`, fx.positionID)); err != nil {
		t.Fatal(err)
	}
	result, err := fx.client.ResolveParkedAttempt(context.Background(), fx.req("FAILED_CONFIRMED", ""))
	if err != nil {
		t.Fatalf("the resolution must commit even when the release fails: %v", err)
	}
	if result.ProposalReleased || result.ReleaseError == "" {
		t.Fatalf("result = %+v, want a named release failure", result)
	}
	if state, armed := fx.state(t); state != journal.StateFailedConfirmed || !armed {
		t.Fatalf("state %s armed %v, want closed attempt with the proposal still armed", state, armed)
	}
	if _, err := db.Exec(`DROP TRIGGER a094_refuse_release`); err != nil {
		t.Fatal(err)
	}
	if released := catchUpExitProposals(context.Background(), fx.j, "acct-7", nil, nil); released != 1 {
		t.Fatalf("boot catch-up released %d, want 1", released)
	}
	if _, armed := fx.state(t); armed {
		t.Fatal("the proposal is still armed after the boot catch-up")
	}
}

// 4.T — 경로 상수는 클라이언트와 같고, 콘솔은 이 route 를 부르지 않음.
func TestA094TheThawRouteIsSharedAndNotOnTheConsole(t *testing.T) {
	if AttemptThawPath != positionpolicyrpc.AttemptThawPath {
		t.Fatalf("engine route %q != client route %q", AttemptThawPath, positionpolicyrpc.AttemptThawPath)
	}
	entries, err := os.ReadDir(filepath.Join("..", "..", "console"))
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		raw, err := os.ReadFile(filepath.Join("..", "..", "console", e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(raw), "attempt-thaw") || strings.Contains(string(raw), "ResolveParkedAttempt") {
			t.Errorf("internal/console/%s reaches the thaw route — the thaw has no console button (D−4.5)", e.Name())
		}
	}
}
