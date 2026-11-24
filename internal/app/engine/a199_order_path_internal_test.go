package engine

// a199_order_path_internal_test.go — a199 G2: 낡은 attestation 을 청산 전용으로 내리는 조건에 「이 계좌로 엔진이 낸 주문이 확정된
// 기록」이 더해짐. 그 기록이 없거나 · 다른 계좌 · 다른 생산자 · 확정 아닌 PLACE · CANCEL 뿐이거나 · 읽기가 실패하면 낡음은 미증명과
// 구별되지 않으므로 지금처럼 기동 거절(낡음 sentinel 없음 · 청산 전용 감사 0 · 래치 0). 유효한 attestation · 게이트 OFF 의 경로는
// 원장을 읽지 않음. 감시는 같은 질의로 「재시작하면 청산 전용 / 기동 거부」를 가름.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/audit"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// SeedEngineOrderForTest 는 원장의 생산 쓰기 경로(Prepare → Dispatch)로 attempt 하나를 만듦 — 게이트웨이가 쓰는 것과 같은 행 모양
// (intent.source · attempt.kind · attempt.state). 실제 게이트웨이 제출로 만든 행과 같은 모양인지는 외부 시험
// TestA199AStopLossFromAVerifiedRunIsTheEvidenceForTheNextExitOnlyStart 가 잼. 외부 시험 패키지도 씀.
func SeedEngineOrderForTest(t testing.TB, j *journal.Journal, id, account, source string, kind journal.MutationKind,
	class journal.DispatchClass,
) journal.AttemptState {
	t.Helper()
	ctx := context.Background()
	req := journal.PrepareRequest{
		Intent: journal.Intent{ID: "a199-seed-intent-" + id, Market: "kr", TradingDay: "2026-01-05", AccountRef: account,
			Symbol: "000270", Side: "SELL", OrderType: "LIMIT", Quantity: "1", Price: "90000", Currency: "KRW",
			Source: source, Fingerprint: "a199-seed-" + id},
		Kind:      kind,
		AttemptID: "a199-seed-attempt-" + id,
	}
	if kind != journal.KindPlace {
		req.TargetOrderID = "BO-a199-seed-target-" + id
	}
	a, err := j.Prepare(ctx, req)
	if err != nil {
		t.Fatalf("seeding %s: Prepare: %v", id, err)
	}
	res, _ := a.Dispatch(ctx, func(context.Context, *journal.Attempt) journal.DispatchOutcome {
		return journal.DispatchOutcome{Class: class, BrokerOrderID: "BO-a199-seed-" + id, ReasonCode: "a199_seed", Detail: "a199 seed"}
	})
	return res.Final
}

// a199Proven 은 「기록 있음」 — 증거가 거절 사유가 아님을 보이려는 시험용.
func a199Proven(string) (bool, error) { return true, nil }

func TestA199AStaleAttestationWithoutAnEngineOrderIsRefused(t *testing.T) {
	for _, tc := range []struct {
		name string
		seed func(t *testing.T, j *journal.Journal)
		read func(j *journal.Journal) orderPathProof
	}{
		{"no order at all", func(*testing.T, *journal.Journal) {}, nil},
		{"only another account's confirmed engine place", func(t *testing.T, j *journal.Journal) {
			SeedEngineOrderForTest(t, j, "other", "999-99", engineSource, journal.KindPlace, journal.DispatchAcked)
		}, nil},
		{"only a flatten place on the account", func(t *testing.T, j *journal.Journal) {
			SeedEngineOrderForTest(t, j, "flatten", "123-45", "flatten-all", journal.KindPlace, journal.DispatchAcked)
		}, nil},
		{"only an in-doubt engine place", func(t *testing.T, j *journal.Journal) {
			SeedEngineOrderForTest(t, j, "doubt", "123-45", engineSource, journal.KindPlace, journal.DispatchAmbiguous)
		}, nil},
		{"only a rejected engine place", func(t *testing.T, j *journal.Journal) {
			SeedEngineOrderForTest(t, j, "rejected", "123-45", engineSource, journal.KindPlace, journal.DispatchRejected)
		}, nil},
		{"only a confirmed engine cancel", func(t *testing.T, j *journal.Journal) {
			SeedEngineOrderForTest(t, j, "cancel", "123-45", engineSource, journal.KindCancel, journal.DispatchAcked)
		}, nil},
		{"the ledger read fails", func(t *testing.T, j *journal.Journal) {
			SeedEngineOrderForTest(t, j, "ok", "123-45", engineSource, journal.KindPlace, journal.DispatchAcked)
		}, func(j *journal.Journal) orderPathProof {
			_ = j.Close() // 기록은 있지만 읽기가 실패 — 내리지 않음
			return engineOrderProof(context.Background(), j)
		}},
		{"no ledger wired", func(*testing.T, *journal.Journal) {}, func(*journal.Journal) orderPathProof { return nil }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, stale := range []string{"expired", "scope_lost"} {
				dir := t.TempDir()
				status := a199StaleStatusMissing(t, dir)
				if stale == "expired" {
					status = a199StaleStatus(t)
				}
				path := filepath.Join(dir, "audit.log")
				log, err := audit.Open(audit.Options{Path: path, Clock: clock.NewFake(a199Now), Subject: "t"})
				if err != nil {
					t.Fatal(err)
				}
				entry := execgw.NewEntryGate(nil, map[execgw.RequiredQuery]time.Duration{})
				facts := a199WiredFacts(t, status, entry)
				j := openTestJournal(t)
				tc.seed(t, j)
				facts.orderProof = engineOrderProof(context.Background(), j)
				if tc.read != nil {
					facts.orderProof = tc.read(j)
				}
				got, err := runInterlock(status, a199Gate(), log, nil, facts)
				if !errors.Is(err, ErrAutomationGateRefused) || !errors.Is(err, ErrOrderPathUnproven) {
					t.Fatalf("%s: err = %v, want a refusal for an unproven order path", stale, err)
				}
				if errors.Is(err, attest.ErrExpired) || errors.Is(err, attest.ErrEndpointNotAttested) {
					t.Fatalf("%s: err = %v carries a stale sentinel", stale, err)
				}
				if got.ExitOnly || got.Verified || got.RunsExitSide() {
					t.Fatalf("%s: status = %+v, want a refused start", stale, got)
				}
				if _, latched := entry.Blocks()[execgw.ReasonCapabilityAttestationStale]; latched {
					t.Fatalf("%s: the stale latch was set on a refused start", stale)
				}
				refused, exitOnly := 0, 0
				for _, e := range a199AuditLines(t, path) {
					switch e.Action {
					case audit.ActionGateRefused:
						refused++
					case audit.ActionGateExitOnly:
						exitOnly++
					}
				}
				if refused != 1 || exitOnly != 0 {
					t.Fatalf("%s: audit refused=%d exit_only=%d, want 1 and 0", stale, refused, exitOnly)
				}
				if lines := UnmetInterlockClauses(err); len(lines) < 2 || !strings.Contains(lines[0], "확정된 기록이 없다") {
					t.Fatalf("%s: clause lines = %q — the refusal must name its clause", stale, lines)
				}
			}
		})
	}
}

// 대조: 같은 조립에 이 계좌의 확정된 엔진 PLACE 가 있으면 청산 전용(위 거절이 증거 때문임).
func TestA199AStaleAttestationWithAnEngineOrderStartsExitOnly(t *testing.T) {
	dir := t.TempDir()
	status := a199StaleStatusMissing(t, dir)
	log, err := audit.Open(audit.Options{Path: filepath.Join(dir, "audit.log"), Clock: clock.NewFake(a199Now), Subject: "t"})
	if err != nil {
		t.Fatal(err)
	}
	entry := execgw.NewEntryGate(nil, map[execgw.RequiredQuery]time.Duration{})
	got, err := runInterlock(status, a199Gate(), log, nil, a199WiredFacts(t, status, entry))
	if err != nil || !got.ExitOnly {
		t.Fatalf("start = %+v, %v, want exit-only with the evidence present", got, err)
	}
	if _, latched := entry.Blocks()[execgw.ReasonCapabilityAttestationStale]; !latched {
		t.Fatal("no stale latch on an exit-only start")
	}
}

// 유효한 attestation 과 게이트 OFF 의 경로는 원장을 읽지 않음(차등 — 그 경로는 한 글자도 안 바뀜).
func TestA199AValidOrGateOffStartReadsNoLedger(t *testing.T) {
	reads := 0
	counting := func(string) (bool, error) { reads++; return true, nil }

	dir := t.TempDir()
	path := filepath.Join(dir, attest.FileName)
	if err := attest.Save(path, attest.Attestation{FormatVersion: attest.FormatVersion, AccountRef: "123-45",
		IssuedAt: a199Now.Add(-24 * time.Hour), ExpiresAt: a199Now.Add(29 * 24 * time.Hour), Endpoints: RequiredEndpoints()}); err != nil {
		t.Fatal(err)
	}
	gate := a199Gate()
	gate.AttestationFile = path
	status := newAutomationStatus(gate, config.Paths{ConfigDir: dir}, ProtectionUnwired)
	status.AccountRef = "123-45"
	log, err := audit.Open(audit.Options{Path: filepath.Join(dir, "audit.log"), Clock: clock.NewFake(a199Now), Subject: "t"})
	if err != nil {
		t.Fatal(err)
	}
	facts := a199WiredFacts(t, status, execgw.NewEntryGate(nil, map[execgw.RequiredQuery]time.Duration{}))
	facts.orderProof = counting
	got, err := runInterlock(status, gate, log, nil, facts)
	if err != nil || !got.Verified {
		t.Fatalf("valid start = %+v, %v", got, err)
	}
	off := a199Gate()
	off.Enabled = false
	if _, err := runInterlock(newAutomationStatus(off, config.Paths{ConfigDir: dir}, ProtectionUnwired), off, log, nil, facts); err != nil {
		t.Fatalf("gate-off start: %v", err)
	}
	if reads != 0 {
		t.Fatalf("ledger reads = %d on a valid / gate-off start, want 0", reads)
	}
}

// 감시: 낡음 단계 × 증거. 증거가 없거나 못 읽으면 「재시작하면 기동이 거부된다」(a199 이전 문장이 참), 있으면 청산 전용.
func TestA199TheWatchTellsARefusalForAStaleFileWithoutAnEngineOrder(t *testing.T) {
	scopeLost := a135Attestation(20 * 24 * time.Hour)
	var kept []string
	for _, e := range RequiredEndpoints() {
		if !strings.HasPrefix(e, "POST ") {
			kept = append(kept, e)
		}
	}
	scopeLost.Endpoints = kept
	for _, file := range []struct {
		stage string
		att   attest.Attestation
	}{{"expired", a135Attestation(-time.Hour)}, {"scope_lost", scopeLost}} {
		for _, proof := range []struct {
			name  string
			proof orderPathProof
		}{
			{"no record", func(string) (bool, error) { return false, nil }},
			{"read failure", func(string) (bool, error) { return false, errors.New("disk I/O error") }},
			{"no ledger", nil},
		} {
			t.Run(file.stage+"/"+proof.name, func(t *testing.T) {
				n := &a135Notifier{}
				att := file.att
				w := a135Watch(clock.NewFake(a135Now), n, true, &att)
				w.proof = proof.proof
				w.checkOnce(t.Context())
				got := n.got()
				if len(got) != 1 || got[0].Fields[obs.FieldReason] != file.stage || got[0].Type != obs.EventCredentialExpiring {
					t.Fatalf("events = %+v, want the %s stage, critical", got, file.stage)
				}
				e := got[0]
				if !strings.HasPrefix(e.Body, "지금 엔진을 재시작하면 기동이 거부된다") || !strings.Contains(e.Title, "기동이 거부된다") {
					t.Errorf("an unproven stale file must say the restart is refused:\n%q\n%q", e.Title, e.Body)
				}
				if strings.Contains(e.Title+e.Body, "청산 전용") {
					t.Errorf("an unproven stale file must not promise exit-only: %q", e.Body)
				}
			})
		}
	}
}

// 감시: 날수 · 지남 단계의 기한 뒤 예측도 같은 증거로 가름. healthy 는 원장을 읽지 않음.
func TestA199TheWatchForecastFollowsTheEvidenceAndAHealthyFileReadsNoLedger(t *testing.T) {
	for name, att := range map[string]attest.Attestation{
		"notice (expiry)": a135Attestation(5 * 24 * time.Hour),
		"notice (proof)":  a196ProofDue(a135Now, 5*a196Day),
		"proof_overdue":   a196ProofDue(a135Now, -time.Hour),
	} {
		t.Run(name, func(t *testing.T) {
			n := &a135Notifier{}
			att := att
			w := a135Watch(clock.NewFake(a135Now), n, true, &att)
			w.proof = func(string) (bool, error) { return false, nil }
			w.checkOnce(t.Context())
			got := n.got()
			if len(got) != 1 {
				t.Fatalf("events = %+v", got)
			}
			text := got[0].Title + got[0].Body
			if !strings.Contains(text, "재시작이 거부") || strings.Contains(text, "청산 전용") {
				t.Errorf("an unproven install's forecast must say the later restart is refused: %q", text)
			}
		})
	}
	reads := 0
	healthy := a135Attestation(20 * 24 * time.Hour)
	w := a135Watch(clock.NewFake(a135Now), &a135Notifier{}, true, &healthy)
	w.proof = func(string) (bool, error) { reads++; return true, nil }
	w.checkOnce(t.Context())
	if reads != 0 {
		t.Fatalf("ledger reads = %d for a healthy file, want 0", reads)
	}
}
