package engine_test

// a199_order_path_test.go — a199 G2 의 생산 조립 시험 · gstack review G3.
//
//   - 원장에 이 계좌로 엔진 주문이 확정된 기록이 없는 설치: 낡은 attestation 이면 기동 거절(낡음 sentinel 없음 · 청산 전용 감사 0).
//     그 뒤 기록이 생기면 같은 파일로의 재기동은 청산 전용 — 판정은 매 기동 원장에서 다시 유도됨.
//   - 실제 게이트웨이가 낸 주문이 증거가 됨: 검증된 기동이 낸 손절 매도(CONFIRMED PLACE, source=engine)만 있는 원장으로, 파일이
//     낡은 뒤의 재기동이 청산 전용. 시드 헬퍼(SeedEngineOrderForTest)가 쓰는 행이 생산 행과 같은 모양임을 이 시험이 잼 — 청산 전용
//     중의 손절도 같은 게이트웨이 · 같은 source 이므로 다음 기동의 증거로 쌓임(순환이 아니라 누적).
//   - G3: 기동이 세운 래치는 조립된 게이트웨이가 묻는 바로 그 게이트임(전략 진입 관문 · 송신 경로가 같은 EntryGate 를 읽음).

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
)

// a199IsolateWithEngineOrder 는 격리된 설정 · 데이터 폴더에, 이 계좌(123-45)로 엔진이 낸 PLACE 하나가 확정된 원장을 둠.
func a199IsolateWithEngineOrder(t *testing.T) string {
	t.Helper()
	dir := isolate(t)
	j := a199OpenLedger(t, dir)
	if s := engine.SeedEngineOrderForTest(t, j, "boot", "123-45", "engine", journal.KindPlace, journal.DispatchAcked); s != journal.StateConfirmed {
		t.Fatalf("seed state = %s, want CONFIRMED", s)
	}
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	return dir
}

// a199OpenLedger 는 엔진이 여는 바로 그 원장 파일(<설정 폴더>/journal.db)을 엶.
func a199OpenLedger(t *testing.T, dir string) *journal.Journal {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(dir, journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	return j
}

func TestA199AnInstallWithoutAnEngineOrderRefusesAStaleStartUntilAnOrderIsRecorded(t *testing.T) {
	dir := isolate(t) // 기록 없는 설치
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, a199Expired)
	srv, _ := interlockServer(t, "123-45")

	eng, err := openGateWithGuardianOptions(t, dir, srv, false, nil)
	if err == nil {
		_ = eng.Close()
		t.Fatalf("a stale attestation on an install with no confirmed engine order started: %+v", eng.Automation)
	}
	if !errors.Is(err, engine.ErrOrderPathUnproven) || errors.Is(err, attest.ErrExpired) ||
		errors.Is(err, attest.ErrEndpointNotAttested) {
		t.Fatalf("err = %v, want the unproven-order-path refusal without a stale sentinel", err)
	}
	if n := a199AuditCount(t, dir, a199ExitOnlyAudit); n != 0 {
		t.Fatalf("exit_only audit lines = %d on a refused start", n)
	}
	if lines := engine.UnmetInterlockClauses(err); len(lines) == 0 || strings.Contains(lines[0], "미상") {
		t.Fatalf("clause lines = %q — the refusal must name its clause", lines)
	}

	// 같은 계좌로 엔진 주문이 확정된 기록이 생긴 뒤의 재기동 — 같은 낡은 파일로 청산 전용.
	j := a199OpenLedger(t, dir)
	engine.SeedEngineOrderForTest(t, j, "later", "123-45", "engine", journal.KindPlace, journal.DispatchAcked)
	_ = j.Close()
	eng, err = openGateWithGuardianOptions(t, dir, srv, false, nil)
	if err != nil {
		t.Fatalf("restart with the order recorded: %v", err)
	}
	defer eng.Close()
	if !eng.Automation.ExitOnly {
		t.Fatalf("automation = %+v, want exit-only once the account has a confirmed engine order", eng.Automation)
	}
}

func TestA199AStopLossFromAVerifiedRunIsTheEvidenceForTheNextExitOnlyStart(t *testing.T) {
	broker, srv := newA199Broker(t)
	dir := isolate(t) // 시드 없음 — 증거는 이 시험의 실제 게이트웨이 제출뿐
	valid := func(a *attest.Attestation, now time.Time) {
		a.IssuedAt = now.Add(-24 * time.Hour)
		a.ExpiresAt = now.Add(29 * 24 * time.Hour)
	}
	s, err := a199Boot(t, dir, broker, srv, valid, nil, false)
	if err != nil || !s.eng.Automation.Verified {
		t.Fatalf("a valid attestation must start verified: %+v, %v", s.eng, err)
	}
	stack := &e2eStack{t: t, engine: s.eng, clk: s.clk}
	stack.entry("005930", "10", "70000", "68000")
	broker.setPrice("67000")
	s.run(t, []engine.SupervisedLoop{s.exitLoop(t)}, func() bool { return len(broker.side("SELL")) > 0 })
	if len(broker.side("SELL")) == 0 {
		t.Fatalf("the verified run sent no stop-loss; engine log:\n%s", s.logs.String())
	}
	if ok, err := s.eng.Journal.HasConfirmedPlace(context.Background(), s.eng.AccountRef, "engine"); err != nil || !ok {
		t.Fatalf("the gateway's stop-loss is not a confirmed engine place in the ledger: %v, %v", ok, err)
	}
	_ = s.eng.Close()

	// 파일이 낡은 뒤의 재기동 — 원장에 있는 것은 위 실제 손절뿐.
	again, err := a199Boot(t, dir, broker, srv, a199ExpiredFile, nil, false)
	if err != nil {
		t.Fatalf("restart on a stale file after a real stop-loss: %v", err)
	}
	if !again.eng.Automation.ExitOnly || again.eng.Automation.Verified {
		t.Fatalf("automation = %+v, want exit-only on the evidence of the real stop-loss", again.eng.Automation)
	}
}

// G3: 기동이 세운 래치를 조립된 게이트웨이가 읽음 — 전략 진입 관문(송신 직전 관문과 같은 EntryGate)이 이 사유로 거절.
// 보호 WIRED 조립은 엔진에 만드는 길이 없음(SetProtectionReadyForTest 는 보고 칸만 바꿈) — WIRED 게이트웨이의 매수 거절 사유는
// execgw TestA199AWiredGatewayRefusesABuyForTheStaleAttestationAndAdmitsASell 이 잼.
func TestA199TheBootLatchIsTheGateTheAssembledGatewayChecks(t *testing.T) {
	for _, ready := range []bool{false, true} {
		s := a199OpenExitOnly(t, a199ScopeLostFile, nil, ready)
		_, err := s.eng.Gateway.ObserveStrategyEntryGate(context.Background(), "kr", "000660")
		if err == nil || !strings.Contains(err.Error(), "capability_attestation_stale") {
			t.Fatalf("protection-ready=%v: the assembled gateway's entry gate = %v, want the boot's stale latch", ready, err)
		}
	}
}
