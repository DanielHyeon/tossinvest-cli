package engine

// a092 25.6 (archive 게이트 24.4): a066 운영자 완화의 통지는 세울 자기 사유가 없는 기록자이므로 원장에 직접 쓰지 않고
// 알림기의 기록 전용 입구(배제 잠금 아래 기록)를 써야 함(a092 델타 「critical 기록 부류」).

import (
	"context"
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// 원장 직접 적재 경로가 타입에서 사라졌는가 — 해제 원장 면에 EnqueueAlert 가 없어야 함.
func TestA092RelaxationRepositoryCannotEnqueue(t *testing.T) {
	iface := reflect.TypeOf((*riskRelaxationRepository)(nil)).Elem()
	if _, ok := iface.MethodByName("EnqueueAlert"); ok {
		t.Error("riskRelaxationRepository still offers EnqueueAlert — the relaxation notice can bypass the notifier entry")
	}
}

// 구조 핀: notifyRelaxation 은 기록자 입구(RecordCritical)를 한 번 부르고 원장 적재 호출은 없음.
func TestA092NotifyRelaxationCallsOnlyTheEntry(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "risk_relaxation_command.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	calls := map[string]int{}
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Name.Name != "notifyRelaxation" {
			continue
		}
		ast.Inspect(fd.Body, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				if s, ok := c.Fun.(*ast.SelectorExpr); ok {
					calls[s.Sel.Name]++
					// 재알림 창은 0 이어야 함 — 해제 통지는 해제마다 새 행이고 정착한 옛 통지를 재무장하지 않음.
					if s.Sel.Name == "RecordCritical" {
						last, ok := c.Args[len(c.Args)-1].(*ast.BasicLit)
						if !ok || last.Value != "0" {
							t.Errorf("RecordCritical's reminder window is not the literal 0")
						}
					}
				}
			}
			return true
		})
	}
	if calls["RecordCritical"] != 1 {
		t.Errorf("RecordCritical calls = %d, want 1", calls["RecordCritical"])
	}
	for _, forbidden := range []string{"EnqueueAlert", "RecordAlert", "ClaimAlertForDelivery", "Notify"} {
		if calls[forbidden] != 0 {
			t.Errorf("notifyRelaxation calls %s — the notice must go through the notifier entry only", forbidden)
		}
	}
}

// 행 모양은 이행 전과 같음 — 키 · 유형 · 등급 · 제목 · 본문 · payload 바이트.
func TestA092RelaxationNoticeRowShapeIsUnchanged(t *testing.T) {
	fx := a066RelaxEngine(t, true, nil)
	at := a066RelaxNow
	result := notifyRelaxation(context.Background(), fx.service.notices, "entry_lock", 7,
		"entry_loss_lock:acct-7/KR/SHORT", "entry_loss_lock:KR/SHORT", "ops", " OPS-1 ", at)
	if !result.Notified {
		t.Fatalf("result = %+v", result)
	}
	rows := fx.alerts(t)
	if len(rows) != 1 {
		t.Fatalf("notices = %d", len(rows))
	}
	payload, _ := json.Marshal(map[string]any{
		"kind": "entry_lock", "target": "entry_loss_lock:acct-7/KR/SHORT", "release_seq": int64(7), "operator": "ops",
		"approval": "OPS-1", "released_at": journal.RFC3339(at),
	})
	want := journal.Alert{EventKey: EventRiskRelaxation + "|entry_lock|7", Type: EventRiskRelaxation, Severity: "critical",
		Title: "RISK RELAXATION: entry_loss_lock:KR/SHORT",
		Body:  "operator ops released entry_loss_lock:KR/SHORT (release 7, approval: OPS-1)", Payload: string(payload)}
	got := rows[0]
	if got.EventKey != want.EventKey || got.Type != want.Type || got.Severity != want.Severity || got.Title != want.Title ||
		got.Body != want.Body || got.Payload != want.Payload {
		t.Errorf("row changed shape:\n got  %q %q %q %q %q %q\n want %q %q %q %q %q %q",
			got.EventKey, got.Type, got.Severity, got.Title, got.Body, got.Payload,
			want.EventKey, want.Type, want.Severity, want.Title, want.Body, want.Payload)
	}
	if got.ClaimedBy != "" {
		t.Errorf("the notice row carries a lease by %q", got.ClaimedBy)
	}
}

// 기록 실패 → 해제는 유지, 결과는 「완화됨 · 통지 실패」, 그리고 입구가 전달 실패 사유로 진입을 잠금(critical 기록 부류).
func TestA092RelaxationNoticeFailureLatchesEntries(t *testing.T) {
	fx := a066RelaxEngine(t, true, nil)
	clk := clock.NewFake(a066RelaxNow)
	broken, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatal(err)
	}
	_ = broken.Close()
	gate := execgw.NewEntryGate(clk, map[execgw.RequiredQuery]time.Duration{})
	fx.service.notices = &obs.Notifier{Journal: broken, Gate: gate}

	result, err := fx.client.ReleaseEntryLossLock(context.Background(), fx.request(t))
	if err != nil {
		t.Fatalf("release: %v", err)
	}
	if result.Notified || result.NotifyError == "" || result.ReleaseSeq <= 0 {
		t.Fatalf("result = %+v, want released but not notified", result)
	}
	if locks := fx.openLocks(t); len(locks) != 0 {
		t.Fatalf("the release did not stand: %+v", locks)
	}
	if rej := gate.CheckEntry(); rej == nil || rej.Reason != execgw.ReasonAlertUndelivered {
		t.Errorf("entry check = %v, want a %s latch — a critical record that failed must close entries",
			rej, execgw.ReasonAlertUndelivered)
	}
}

// 알림기가 배선되지 않은 엔진은 「통지됨」이라고 말하지 않음.
func TestA092RelaxationWithoutANotifierIsNotNotified(t *testing.T) {
	fx := a066RelaxEngine(t, true, nil)
	fx.service.notices = nil
	result, err := fx.client.ReleaseEntryLossLock(context.Background(), fx.request(t))
	if err != nil {
		t.Fatalf("release: %v", err)
	}
	if result.Notified || result.NotifyError == "" {
		t.Fatalf("result = %+v, want not notified with a reason", result)
	}
	if got := len(fx.alerts(t)); got != 0 {
		t.Fatalf("notices = %d without a notifier", got)
	}
}

// 조립: 엔진 알림기가 있으면 그것이 기록자이고, 없으면 nil 인터페이스(타입 있는 nil 금지).
func TestA092CommandServiceTakesTheEngineNotifier(t *testing.T) {
	j := &journal.Journal{}
	n := &obs.Notifier{}
	s, err := NewPositionPolicyCommandService(&Context{Journal: j, Notifier: n}, clock.NewFake(a066RelaxNow))
	if err != nil {
		t.Fatal(err)
	}
	if got, ok := s.notices.(*obs.Notifier); !ok || got != n {
		t.Errorf("notices = %T, want the engine notifier", s.notices)
	}
	s, err = NewPositionPolicyCommandService(&Context{Journal: j}, clock.NewFake(a066RelaxNow))
	if err != nil {
		t.Fatal(err)
	}
	if s.notices != nil {
		t.Errorf("notices = %#v, want a nil interface when no notifier is wired", s.notices)
	}
}
