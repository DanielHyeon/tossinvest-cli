package engine

// a199_exit_only_guards_internal_test.go — 구현 리뷰 r1 처분 2 · 3 · 7(a) · 7(h).
//
//   - 처분 2: 래치가 진입 차단의 유일한 집행이므로 래치를 세울 진입 게이트가 없으면 청산 전용으로 내려가지 않고 거절(nil 역참조
//     panic 대신 읽을 수 있는 오류).
//   - 처분 3: 청산 전용 감사 기록이 실패하면 기동 거절 — 수락 경로와 같은 정책(기록 없는 청산 전용은 「언제부터 진입이 막혔나」를
//     재구성할 수 없게 함).
//   - 처분 7(a): 감사 Detail 의 missing= 에 빠진 endpoint 이름이 실리고 계좌 · 경로 · 원문 오류가 없음.
//   - 처분 7(h): 인터록(verifyGate)과 감시(judgeAttestation)가 같은 판정 함수 classifyAttestation 을 부르고, 어느 쪽도
//     Attestation.Verify 를 직접 부르지 않음(판정 둘 금지 — 구조로 못 박음).

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
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
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

func a199Gate() config.AutomationGate {
	return config.AutomationGate{Enabled: true, MaxOrderQuantity: 10, MaxOrderNotional: 1_000_000, MaxTotalExposure: 5_000_000,
		MaxDailyLossAmount: 200_000, MaxDailyLossRatio: 0.02, LimitCurrency: "KRW"}
}

// a199StaleStatusMissing 는 만료 전이지만 POST 주문 · 취소가 빠진(운영에서 난 모양) attestation 의 상태임.
func a199StaleStatusMissing(t *testing.T, dir string) AutomationStatus {
	t.Helper()
	gate := a199Gate()
	var kept []string
	for _, e := range RequiredEndpoints() {
		if !strings.HasPrefix(e, "POST ") {
			kept = append(kept, e)
		}
	}
	path := filepath.Join(dir, attest.FileName)
	if err := attest.Save(path, attest.Attestation{FormatVersion: attest.FormatVersion, AccountRef: "123-45",
		IssuedAt: a199Now.Add(-24 * time.Hour), ExpiresAt: a199Now.Add(29 * 24 * time.Hour), Endpoints: kept}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	gate.AttestationFile = path
	status := newAutomationStatus(gate, config.Paths{ConfigDir: dir}, ProtectionUnwired)
	status.AccountRef = "123-45"
	return status
}

func a199WiredFacts(t *testing.T, status AutomationStatus, entry *execgw.EntryGate) gateFacts {
	t.Helper()
	j := openTestJournal(t)
	gw, err := execgw.New(execgw.Options{Journal: j, Trading: trading.NewService(config.Trading{}, &officialBroker{}),
		AccountRef: "123-45", Orders: execgw.OfficialOrders{}, Entry: execgw.NewEntryGate(nil, map[execgw.RequiredQuery]time.Duration{})})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	// a199 G2: 이 계좌로 엔진 주문이 확정된 기록 하나(생산 쓰기 경로로) — 증거 없는 사례는 시험이 orderProof 를 바꿈.
	SeedEngineOrderForTest(t, j, "wired-facts", "123-45", "engine", journal.KindPlace, journal.DispatchAcked)
	return gateFacts{
		guardian:   a199LimitGuardian{limits: status.Limits},
		trading:    config.Trading{Place: true, Sell: true, Cancel: true, AllowLiveOrderActions: true},
		gateway:    gw,
		transport:  &officialBroker{},
		entry:      entry,
		orderProof: engineOrderProof(context.Background(), j),
		now:        a199Now,
	}
}

func a199AuditLines(t *testing.T, path string) []audit.Entry {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open audit: %v", err)
	}
	defer f.Close()
	var out []audit.Entry
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var e audit.Entry
		if err := json.Unmarshal(sc.Bytes(), &e); err != nil {
			t.Fatalf("audit line: %v", err)
		}
		out = append(out, e)
	}
	return out
}

func TestA199NoEntryGateMeansNoExitOnlyStart(t *testing.T) {
	dir := t.TempDir()
	status := a199StaleStatusMissing(t, dir)
	log, err := audit.Open(audit.Options{Path: filepath.Join(dir, "audit.log"), Clock: clock.NewFake(a199Now), Subject: "t"})
	if err != nil {
		t.Fatal(err)
	}
	got, err := runInterlock(status, a199Gate(), log, nil, a199WiredFacts(t, status, nil))
	if !errors.Is(err, ErrAutomationGateRefused) || !errors.Is(err, ErrGatewayRequired) {
		t.Fatalf("err = %v, want a gateway-clause refusal — no gate to carry the latch", err)
	}
	// G9: 거절된 기동은 청산 전용 표지를 들지 않음(거절 줄에 보호 수명 문장이 실리지 않게).
	if got.Verified || got.EntryPermitted || got.ExitOnly || got.RunsExitSide() {
		t.Fatalf("status = %+v, want neither verified, entry-permitted nor exit-only", got)
	}
	for _, e := range a199AuditLines(t, filepath.Join(dir, "audit.log")) {
		if e.Action == audit.ActionGateExitOnly || e.Action == audit.ActionGateAccepted {
			t.Fatalf("audit has %q on a refused start", e.Action)
		}
	}
}

func TestA199AnUnrecordableExitOnlyStartIsRefused(t *testing.T) {
	dir := t.TempDir()
	status := a199StaleStatusMissing(t, dir)
	path := filepath.Join(dir, "audit.log")
	log, err := audit.Open(audit.Options{Path: path, Clock: clock.NewFake(a199Now), Subject: "t"})
	if err != nil {
		t.Fatal(err)
	}
	// 두 기록 사이의 일시 실패를 흉내 냄 — 경로를 디렉터리로 바꿔 append 가 실패하게.
	_ = os.Remove(path)
	if err := os.Mkdir(path, 0o700); err != nil {
		t.Fatal(err)
	}
	entry := execgw.NewEntryGate(nil, map[execgw.RequiredQuery]time.Duration{})
	_, err = runInterlock(status, a199Gate(), log, nil, a199WiredFacts(t, status, entry))
	if err == nil || !strings.Contains(err.Error(), "recording the exit-only start") {
		t.Fatalf("err = %v, want the exit-only start refused for want of its audit line (same policy as acceptance)", err)
	}
}

func TestA199TheExitOnlyAuditDetailNamesTheMissingEndpointsAndNothingPrivate(t *testing.T) {
	dir := t.TempDir()
	status := a199StaleStatusMissing(t, dir)
	path := filepath.Join(dir, "audit.log")
	log, err := audit.Open(audit.Options{Path: path, Clock: clock.NewFake(a199Now), Subject: "t"})
	if err != nil {
		t.Fatal(err)
	}
	entry := execgw.NewEntryGate(nil, map[execgw.RequiredQuery]time.Duration{})
	got, err := runInterlock(status, a199Gate(), log, nil, a199WiredFacts(t, status, entry))
	if err != nil || !got.ExitOnly {
		t.Fatalf("start = %+v, %v, want exit-only", got, err)
	}
	var detail string
	for _, e := range a199AuditLines(t, path) {
		if e.Action == audit.ActionGateExitOnly {
			detail = e.Detail
		}
	}
	for _, want := range []string{"reason=capability_attestation_stale", "stale=scope_lost",
		"POST /api/v1/orders", "POST /api/v1/orders/{id}/cancel"} {
		if !strings.Contains(detail, want) {
			t.Errorf("detail %q lacks %q", detail, want)
		}
	}
	for _, banned := range []string{"123-45", "12345", "***", dir, attest.FileName, "attest:", "never verified"} {
		if strings.Contains(detail, banned) {
			t.Errorf("detail %q carries %q (account · path · raw error must not be recorded)", detail, banned)
		}
	}
}

func TestA199TheInterlockAndTheWatchCallTheSameClassifier(t *testing.T) {
	calls := func(file, fn string) map[string]int {
		t.Helper()
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		f, err := parser.ParseFile(token.NewFileSet(), file, src, 0)
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]int{}
		for _, d := range f.Decls {
			fd, ok := d.(*ast.FuncDecl)
			if !ok || fd.Name.Name != fn {
				continue
			}
			ast.Inspect(fd.Body, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok {
					switch f := c.Fun.(type) {
					case *ast.Ident:
						got[f.Name]++
					case *ast.SelectorExpr:
						got["."+f.Sel.Name]++
					}
				}
				return true
			})
		}
		return got
	}
	for _, site := range []struct{ file, fn string }{{"interlock.go", "verifyGate"}, {"attestation_watch.go", "judgeAttestation"}} {
		got := calls(site.file, site.fn)
		if got["classifyAttestation"] != 1 {
			t.Errorf("%s calls classifyAttestation %d time(s), want 1 — the interlock and the watch share one judgement", site.fn, got["classifyAttestation"])
		}
		if got[".Verify"] != 0 {
			t.Errorf("%s calls Verify directly — a second judgement beside classifyAttestation", site.fn)
		}
	}
}
