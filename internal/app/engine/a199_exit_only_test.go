package engine_test

// a199_exit_only_test.go — 낡은 capability 증명은 진입만 막고 청산은 막지 않는다(a199).
//
// 이 파일이 고정하는 것(design §3):
//   1. 낡음(만료 · 필요 endpoint 누락 · 둘 다)으로만 실패하는 attestation 은 기동을 거부하지 않고 「청산 전용」으로 세움 —
//      Verified=false(뜻 무변), EntryPermitted=false, Guardian 유지, 진입 래치(새 사유) 하나, 감사에 청산 전용 action 한 줄.
//   2. 그 Context 로 exit 관측 · 대사 루프가 만들어짐(청산 쪽 소비자는 새 표지를 읽음).
//   4. 거절이 유지되는 결함 — 없음 · 불완전 · 다른 계좌 · 만료가 가린 다른 계좌 · 낡았는데 빈 endpoint 항목 · 낡았는데 형식 < 1.
//   5. 낡음 + 다른 조항 실패 → 그 조항의 오류만(낡음 sentinel 을 싣지 않음).
//   8. 재기동: 낡은 채면 래치가 다시 서고, 유효한 attestation 이면 attestation 사유의 래치가 없고 감사는 수락.
//
// 사유 · 감사 action 은 철자(계약값)로 단언함 — 원장 · 알림에 그 문자열이 남기 때문.

import (
	"bytes"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

const (
	a199StaleReason   = execgw.ReasonCode("capability_attestation_stale")
	a199ExitOnlyAudit = "automation_gate.exit_only"
)

// a199Expired 는 만료된(계좌는 일치) attestation 변형임.
func a199Expired(a *attest.Attestation) {
	a.IssuedAt = interlockNow.Add(-31 * 24 * time.Hour)
	a.ExpiresAt = interlockNow.Add(-time.Hour)
}

// a199OpenProduction 은 생산 Guardian 으로 게이트 ON 엔진을 엶(exit 관측이 Guardian 의 위험 감소 발급을 요구하므로).
func a199OpenProduction(t *testing.T, mutate func(*attest.Attestation)) (*engine.Context, string, error) {
	t.Helper()
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, mutate)
	srv, _ := interlockServer(t, "123-45")
	eng, err := openGateWithGuardianOptions(t, dir, srv, false, nil)
	return eng, dir, err
}

func a199AuditCount(t *testing.T, dir, action string) int {
	t.Helper()
	n := 0
	for _, e := range readAudit(t, filepath.Join(dir, "audit.log")) {
		if e.Setting == "engine.automation_gate.enabled" && e.Action == action {
			n++
		}
	}
	return n
}

func TestA199AStaleAttestationStartsExitOnly(t *testing.T) {
	cases := map[string]func(*attest.Attestation){
		"expired":                   a199Expired,
		"misses the order endpoint": func(a *attest.Attestation) { a.Endpoints = withoutEndpoint("POST /api/v1/orders") },
		"predates the price read":   func(a *attest.Attestation) { a.Endpoints = withoutEndpoint("GET /api/v1/prices") },
		"expired and misses an endpoint": func(a *attest.Attestation) {
			a199Expired(a)
			a.Endpoints = withoutEndpoint("POST /api/v1/orders/{id}/cancel")
		},
		"same account written with hyphens": func(a *attest.Attestation) {
			a199Expired(a)
			a.AccountRef = "1-2-3-4-5"
		},
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			eng, dir, err := a199OpenProduction(t, mutate)
			if err != nil {
				t.Fatalf("a stale attestation must start the runtime exit-only, not refuse it: %v", err)
			}
			if eng.Automation.Verified {
				t.Error("Verified = true: its meaning (every clause held) must not change for an exit-only start")
			}
			if eng.Automation.EntryPermitted {
				t.Error("EntryPermitted = true on an exit-only start")
			}
			if eng.Guardian == nil {
				t.Fatal("Guardian = nil: the exit observer issues its reductions through it")
			}
			if _, latched := eng.Entry.Blocks()[a199StaleReason]; !latched {
				t.Fatalf("entry gate latches = %v, want %q", eng.Entry.Blocks(), a199StaleReason)
			}
			if r := eng.Entry.CheckEntryFor("kr", "005930"); r == nil || r.Reason != a199StaleReason {
				t.Fatalf("CheckEntryFor = %v, want the stale-attestation refusal", r)
			}
			if _, err := eng.ExitObserver(engine.ExitObserverOptions{Costs: costs.DefaultModel()}); err != nil {
				t.Fatalf("ExitObserver on an exit-only start: %v", err)
			}
			clk := clock.NewFake(interlockNow)
			if _, err := eng.ReconcileDriver(engine.ReconcileDriverOptions{
				Collector: eng.SnapshotCollector(clk), Clock: clk, DefaultMarket: "kr",
			}); err != nil {
				t.Fatalf("ReconcileDriver on an exit-only start: %v", err)
			}
			if got := a199AuditCount(t, dir, a199ExitOnlyAudit); got != 1 {
				t.Errorf("exit-only audit lines = %d, want 1", got)
			}
			if got := a199AuditCount(t, dir, audit.ActionGateAccepted) + a199AuditCount(t, dir, audit.ActionGateRefused); got != 0 {
				t.Errorf("accepted+refused audit lines = %d, want 0 — exit-only is neither", got)
			}
		})
	}
}

func TestA199RefusalsStayRefusals(t *testing.T) {
	cases := []struct {
		name    string
		write   bool
		mutate  func(*attest.Attestation)
		account string
		want    error
	}{
		{"no attestation", false, nil, "123-45", attest.ErrMissing},
		{"names no account", true, func(a *attest.Attestation) { a199Expired(a); a.AccountRef = "" }, "123-45", attest.ErrIncomplete},
		{"another account", true, nil, "999-99", attest.ErrAccountMismatch},
		{"expired and another account", true, a199Expired, "999-99", attest.ErrAccountMismatch},
		{"expired with an empty endpoint item", true, func(a *attest.Attestation) {
			a199Expired(a)
			a.Endpoints = append(engine.RequiredEndpoints(), "")
		}, "123-45", attest.ErrIncomplete},
		{"scope lost with a blank endpoint item", true, func(a *attest.Attestation) {
			a.Endpoints = append(withoutEndpoint("POST /api/v1/orders"), "   ")
		}, "123-45", attest.ErrIncomplete},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := isolate(t)
			writeGateConfig(t, dir, smallLiveGate())
			writeCredentials(t, dir, "test-api-key-000000", "test-secret")
			if tc.write {
				writeAttestation(t, dir, tc.mutate)
			}
			srv, _ := interlockServer(t, tc.account)
			eng, err := openGateWithGuardianOptions(t, dir, srv, false, nil)
			if err == nil || eng != nil {
				t.Fatalf("start = (%v, %v), want a refusal", eng != nil, err)
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want %v", err, tc.want)
			}
			if errors.Is(err, attest.ErrExpired) || errors.Is(err, attest.ErrEndpointNotAttested) {
				t.Errorf("err = %v carries a stale sentinel: a refused start must not read as stale", err)
			}
			if got := a199AuditCount(t, dir, a199ExitOnlyAudit); got != 0 {
				t.Errorf("exit-only audit lines = %d on a refused start", got)
			}
		})
	}
}

func TestA199AnInvalidFormatVersionDoesNotGoExitOnly(t *testing.T) {
	eng, _, err := a199OpenProduction(t, func(a *attest.Attestation) {
		a199Expired(a)
		a.FormatVersion = -1
	})
	if err == nil || eng != nil {
		t.Fatal("a stale attestation with format_version -1 must be refused, not exit-only")
	}
	if errors.Is(err, attest.ErrExpired) {
		t.Errorf("err = %v carries the stale sentinel", err)
	}
}

// TestA199NoRefusalCarriesAStaleSentinel 는 구현 리뷰 r1 처분 6 의 확인이다: 기동 **거절**의 오류 사슬에 ErrExpired ·
// ErrEndpointNotAttested 가 실리는 입력이 있는가. 낡음은 거절이 아니라 청산 전용이 되고, 낡음과 겹친 결함은 그 결함의 sentinel
// 만 싣는다. 이 표의 모든 거절에서 두 sentinel 이 0 이면 interlockClauses 의 두 줄은 도달할 수 없는 열거다(지움).
// 처분 7(i): 불완전(ErrIncomplete)의 세 갈래(계좌 칸 · 만료 칸 · endpoint 목록)가 낡음으로 내려가지 않음도 여기서 잼.
func TestA199NoRefusalCarriesAStaleSentinel(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*attest.Attestation)
		account string
		guard   *stubGuardian
		want    error
	}{
		{"expired and another account", a199Expired, "999-99", nil, attest.ErrAccountMismatch},
		{"expired with an empty endpoint item", func(a *attest.Attestation) {
			a199Expired(a)
			a.Endpoints = append(engine.RequiredEndpoints(), "")
		}, "123-45", nil, attest.ErrIncomplete},
		{"scope lost with a blank endpoint item", func(a *attest.Attestation) {
			a.Endpoints = append(withoutEndpoint("POST /api/v1/orders"), " ")
		}, "123-45", nil, attest.ErrIncomplete},
		{"expired with format_version -1", func(a *attest.Attestation) { a199Expired(a); a.FormatVersion = -1 }, "123-45", nil, attest.ErrMalformed},
		{"incomplete: no account (expired)", func(a *attest.Attestation) { a199Expired(a); a.AccountRef = "" }, "123-45", nil, attest.ErrIncomplete},
		{"incomplete: no expiry", func(a *attest.Attestation) { a.ExpiresAt = time.Time{} }, "123-45", nil, attest.ErrIncomplete},
		{"incomplete: no endpoints (expired)", func(a *attest.Attestation) { a199Expired(a); a.Endpoints = nil }, "123-45", nil, attest.ErrIncomplete},
		{"expired and the limits source differs", a199Expired, "123-45",
			&stubGuardian{limits: gateLimitsWith(func(l *execgw.Limits) { l.MaxQuantity = execgw.Bound(9) })}, engine.ErrGuardianLimitsMismatch},
		{"scope lost and the limits source differs", func(a *attest.Attestation) { a.Endpoints = withoutEndpoint("POST /api/v1/orders") }, "123-45",
			&stubGuardian{limits: gateLimitsWith(func(l *execgw.Limits) { l.MaxQuantity = execgw.Bound(9) })}, engine.ErrGuardianLimitsMismatch},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dir := isolate(t)
			srv, _ := interlockServer(t, tc.account)
			writeCredentials(t, dir, "test-api-key-000000", "test-secret")
			var (
				eng *engine.Context
				err error
			)
			if tc.guard != nil {
				writeGateConfig(t, dir, fullGate())
				writeAttestation(t, dir, tc.mutate)
				eng, err = openGateEngine(t, dir, srv, *tc.guard)
			} else {
				writeGateConfig(t, dir, smallLiveGate())
				writeAttestation(t, dir, tc.mutate)
				eng, err = openGateWithGuardianOptions(t, dir, srv, false, nil)
			}
			if err == nil || eng != nil {
				t.Fatalf("start = (%v, %v), want a refusal", eng != nil, err)
			}
			if !errors.Is(err, tc.want) {
				t.Errorf("err = %v, want sentinel %v", err, tc.want)
			}
			if errors.Is(err, attest.ErrExpired) || errors.Is(err, attest.ErrEndpointNotAttested) {
				t.Errorf("a refusal carries a stale sentinel: %v", err)
			}
			for _, line := range engine.UnmetInterlockClauses(err) {
				if strings.Contains(line, "미상 조항") {
					t.Errorf("the refusal enumerates as an unknown clause: %v", engine.UnmetInterlockClauses(err))
				}
			}
		})
	}
}

func TestA199StaleWithAnotherClauseRefusesByThatClauseOnly(t *testing.T) {
	cases := map[string]engineGuardianCase{
		"limits source": {guardian: stubGuardian{limits: gateLimitsWith(func(l *execgw.Limits) { l.MaxQuantity = execgw.Bound(9) })}},
		"daily buy (7b)": {guardian: stubGuardian{limits: fullGateLimits(),
			dailyBuy: riskcalc.Money{Amount: "1", Currency: "KRW"}}},
		"open positions (7c)": {guardian: stubGuardian{limits: fullGateLimits(), maxOpenPositions: 3}},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			dir := isolate(t)
			writeGateConfig(t, dir, fullGate())
			writeCredentials(t, dir, "test-api-key-000000", "test-secret")
			writeAttestation(t, dir, a199Expired)
			srv, _ := interlockServer(t, "123-45")
			eng, err := openGateEngine(t, dir, srv, tc.guardian)
			if err == nil || eng != nil {
				t.Fatal("stale + another unmet clause must refuse the start")
			}
			if !errors.Is(err, engine.ErrGuardianLimitsMismatch) {
				t.Errorf("err = %v, want the single-source clause", err)
			}
			if errors.Is(err, attest.ErrExpired) {
				t.Errorf("err = %v also carries the stale sentinel — the refusal is that clause's alone", err)
			}
		})
	}
}

type engineGuardianCase struct{ guardian stubGuardian }

func TestA199ARestartRederivesTheLatchFromTheFile(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, a199Expired)
	srv, _ := interlockServer(t, "123-45")

	open := func() *engine.Context {
		t.Helper()
		eng, err := openGateWithGuardianOptions(t, dir, srv, false, nil)
		if err != nil {
			t.Fatalf("start: %v", err)
		}
		return eng
	}
	first := open()
	if _, latched := first.Entry.Blocks()[a199StaleReason]; !latched {
		t.Fatal("first start: no stale latch")
	}
	_ = first.Close()

	second := open() // 고치지 않고 재기동 — 래치가 다시 서야 함
	if _, latched := second.Entry.Blocks()[a199StaleReason]; !latched {
		t.Fatal("a restart with the same stale file came up without the latch — a restart must not clear it")
	}
	_ = second.Close()

	writeAttestation(t, dir, nil) // 갱신된 유효한 attestation
	third := open()
	if _, latched := third.Entry.Blocks()[a199StaleReason]; latched {
		t.Fatal("a valid attestation still carries the stale latch after a restart")
	}
	if !third.Automation.Verified {
		t.Error("a valid attestation must verify again")
	}
	if got := a199AuditCount(t, dir, a199ExitOnlyAudit); got != 2 {
		t.Errorf("exit-only audit lines = %d, want 2 (two stale starts)", got)
	}
	if got := a199AuditCount(t, dir, audit.ActionGateAccepted); got != 1 {
		t.Errorf("accepted audit lines = %d, want 1 (the repaired start)", got)
	}
}

func TestA199TheExitOnlyStartSaysWhatItIs(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, fullGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, a199Expired)
	srv, _ := interlockServer(t, "123-45")

	var logs bytes.Buffer
	if _, err := openGateEngineLogging(t, dir, srv, matchedGuardian(), &logs); err != nil {
		t.Fatalf("start: %v", err)
	}
	line := decodeLastLog(t, logs.String())
	if line["level"] != "WARN" {
		t.Errorf("level = %v, want WARN — an exit-only start is not routine", line["level"])
	}
	if reason, _ := line["reason"].(string); !strings.Contains(reason, string(a199StaleReason)) {
		t.Errorf("reason = %q, want the stale-attestation reason", reason)
	}
	if detail, _ := line["detail"].(string); !strings.Contains(detail, "프로세스가 죽으면 보호도 사라진다") {
		t.Errorf("detail = %q — an exit-only start is a protection-unwired runtime and must say so (보호 미배선 기동)", detail)
	}
	if line["gate_exit_only"] != true {
		t.Errorf("gate_exit_only = %v, want true on an exit-only start", line["gate_exit_only"])
	}
}

// 구현 리뷰 r1 처분 5: `gate_exit_only` 칸은 청산 전용 기동에만 실린다 — 유효 · 게이트 OFF 경로의 로그 줄은 base 와 같은 칸 집합.
func TestA199TheExitOnlyFieldIsAbsentFromOtherStarts(t *testing.T) {
	for name, tc := range map[string]struct {
		gate     func() config.AutomationGate
		guardian execgw.Guardian
	}{
		"valid attestation": {fullGate, matchedGuardian()},
		"gate off":          {func() config.AutomationGate { return config.AutomationGate{} }, nil},
	} {
		t.Run(name, func(t *testing.T) {
			dir := isolate(t)
			writeGateConfig(t, dir, tc.gate())
			writeCredentials(t, dir, "test-api-key-000000", "test-secret")
			writeAttestation(t, dir, nil)
			srv, _ := interlockServer(t, "123-45")
			var logs bytes.Buffer
			if _, err := openGateEngineLogging(t, dir, srv, tc.guardian, &logs); err != nil {
				t.Fatalf("start: %v", err)
			}
			line := decodeLastLog(t, logs.String())
			if _, present := line["gate_exit_only"]; present {
				t.Errorf("gate_exit_only present on a %s start: %v — the line must match base", name, line)
			}
			if line["event"] != string(obs.EventOperatingMode) {
				t.Fatalf("last line is not the gate decision: %v", line)
			}
		})
	}
}
