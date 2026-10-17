package obs_test

// a092 21.3 · 22.3 C1/C16 · 23.3 K1/K17: 알림기의 기록 전용 입구.
//
// exit 관측 goroutine 은 critical 알림과 모드 전이 통지를 **기록까지만** 하고 반환해야 함(정본 요구 「등급화된 알림」).
// 발송은 배달 실행자(a124)가 함. 여기서는 입구가 원격 전송을 부르지 않는지, 기록이 임차 없는 PENDING 행인지,
// 기록 실패가 그 자리에서 잠기는지, 모드 통지의 신원이 전이 하나인지를 잼.

import (
	"bytes"
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

// stuckPublisher 는 원격 전송이 멈춘 상태를 흉내 냄 — 부르면 ctx 가 끝날 때까지 돌아오지 않음.
type stuckPublisher struct {
	mu    sync.Mutex
	calls int
}

func (p *stuckPublisher) Publish(ctx context.Context, _ obs.Notification) error {
	p.mu.Lock()
	p.calls++
	p.mu.Unlock()
	<-ctx.Done()
	return ctx.Err()
}

func (p *stuckPublisher) count() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.calls
}

// a092Within 은 f 가 d 안에 끝나는지 봄 — 끝나지 않으면 입구가 전송을 기다린 것.
func a092Within(t *testing.T, d time.Duration, f func() error) error {
	t.Helper()
	done := make(chan error, 1)
	go func() { done <- f() }()
	select {
	case err := <-done:
		return err
	case <-time.After(d):
		t.Fatalf("the record-only entry did not return within %s — it is waiting on the transport", d)
		return nil
	}
}

func a092ModeRecord(id, mode string) journal.OperatingModeRecord {
	return journal.OperatingModeRecord{
		ID: id, AccountRef: "acct-a092", Mode: mode,
		Cause: journal.ModeTriggerExitObservationOutage, Actor: journal.ModeActorAuto, CreatedAt: obsNow,
	}
}

// (a)(b)(d): critical 기록은 publish 없이 반환하고, 행은 임차 없는 PENDING 이며 배달 실행자가 곧바로 집음.
// 래치·승격 없음 — 아직 아무것도 실패하지 않았음.
func TestA092RecordOnlyCriticalNeverPublishes(t *testing.T) {
	pub := &stuckPublisher{}
	n, j, gate, _ := a096Notifier(t, pub)
	n.AccountRef = "acct-a092"
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := a092Within(t, 5*time.Second, func() error { return obs.RecordOnly{N: n}.Notify(ctx, a096Event()) }); err != nil {
		t.Fatalf("Notify: %v", err)
	}
	if pub.count() != 0 {
		t.Errorf("publish calls = %d, want 0 — the exit goroutine must not send", pub.count())
	}
	rows, err := j.PendingAlerts(ctx, 0)
	if err != nil || len(rows) != 1 {
		t.Fatalf("pending rows = %d (%v), want 1", len(rows), err)
	}
	if rows[0].ClaimedBy != "" || rows[0].ClaimExpiresAt != nil {
		t.Errorf("the recorded row carries a lease by %q", rows[0].ClaimedBy)
	}
	if rows[0].EventKey != a096Event().Key || rows[0].Severity != string(obs.SeverityCritical) {
		t.Errorf("row key/severity = %q/%q", rows[0].EventKey, rows[0].Severity)
	}
	claim, err := j.ClaimAlertByID(ctx, rows[0].ID, "deliverer")
	if err != nil || claim.Disposition != journal.ClaimAcquired {
		t.Errorf("deliverer cannot take the row: %v %v", claim.Disposition, err)
	}
	if gate.CheckEntry() != nil {
		t.Errorf("entries blocked after a successful record: %v", gate.CheckEntry())
	}
	if cur, err := j.CurrentOperatingMode(ctx, "acct-a092"); err != nil || cur.Mode != journal.ModeNormal {
		t.Errorf("mode = %q (%v), want NORMAL — a record is not a delivery failure", cur.Mode, err)
	}
}

// (c): 재알림 창이 지난 정착 행은 기록 전용 입구도 다시 무장함(정본 「재무장된 outbox 행은 통째로 이번 에피소드」).
func TestA092RecordOnlyRearmsPastTheReminderWindow(t *testing.T) {
	pub := &failingPublisher{}
	n, j, _, clk := a096Notifier(t, pub)
	ctx := context.Background()

	if err := n.Notify(ctx, a096Event()); err != nil { // 동기 경로로 한 번 전달
		t.Fatalf("Notify: %v", err)
	}
	if rows, _ := j.PendingAlerts(ctx, 0); len(rows) != 0 {
		t.Fatalf("setup: %d pending after a delivered send", len(rows))
	}
	clk.Advance(a096Remind + time.Minute)
	if err := (obs.RecordOnly{N: n}).Notify(ctx, a096Event()); err != nil {
		t.Fatalf("RecordOnly.Notify: %v", err)
	}
	if rows, _ := j.PendingAlerts(ctx, 0); len(rows) != 1 {
		t.Errorf("pending rows = %d, want 1 — a settled row past the window must be re-armed", len(rows))
	}
	if pub.callCount() != 1 {
		t.Errorf("publish calls = %d, want 1 (the setup send only)", pub.callCount())
	}
}

// (e) · K17: 기록 자체가 실패하면 그 자리에서 전달 실패 사유로 잠그고 로그를 남기고, durable 차단(승격)을 시도하고, 오류를 돌려줌.
func TestA092RecordOnlyFailureLatchesAndEscalates(t *testing.T) {
	for name, call := range map[string]func(obs.RecordOnly) error{
		"notify": func(r obs.RecordOnly) error { return r.Notify(context.Background(), a096Event()) },
		"announce": func(r obs.RecordOnly) error {
			return r.AnnounceOperatingMode(context.Background(), journal.ModeNormal,
				a092ModeRecord("mode-1", journal.ModeEntryBlocked))
		},
	} {
		t.Run(name, func(t *testing.T) {
			n, gate, buf := a097BrokenClaim(t)
			pub := &stuckPublisher{}
			n.Publisher = pub
			err := a092Within(t, 5*time.Second, func() error { return call(obs.RecordOnly{N: n}) })
			if err == nil {
				t.Fatal("nil error with a closed journal")
			}
			if rej := gate.CheckEntry(); rej == nil || rej.Reason != execgw.ReasonAlertUndelivered {
				t.Errorf("entry check = %v, want a %s latch after a critical record failed",
					rej, execgw.ReasonAlertUndelivered)
			}
			log := buf.String()
			if !strings.Contains(log, string(obs.EventAlertUndelivered)) {
				t.Errorf("no %s line; got:\n%s", obs.EventAlertUndelivered, log)
			}
			// 사건 자체의 로그 줄도 engine.operating_mode 이므로(모드 통지) 이벤트 이름이 아니라 승격 실패 줄의 문구로 가름.
			if !strings.Contains(log, "did not reach the operating mode") {
				t.Errorf("the durable block was never attempted; got:\n%s", log)
			}
			if pub.count() != 0 {
				t.Errorf("publish calls = %d after a failed record, want 0", pub.count())
			}
		})
	}
}

// C16: Journal 이 없으면 동기 publish 갈래(notifyCritical B1)를 쓰지 않음 — 경고만.
func TestA092RecordOnlyWithoutAJournalDoesNotPublish(t *testing.T) {
	pub := &stuckPublisher{}
	buf := &bytes.Buffer{}
	clk := clock.NewFake(obsNow)
	n := &obs.Notifier{
		Log:       obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true, Clock: clk}),
		Publisher: pub,
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	for _, call := range []func() error{
		func() error { return obs.RecordOnly{N: n}.Notify(ctx, a096Event()) },
		func() error {
			return obs.RecordOnly{N: n}.AnnounceOperatingMode(ctx, journal.ModeNormal,
				a092ModeRecord("mode-1", journal.ModeEntryBlocked))
		},
	} {
		if err := a092Within(t, 5*time.Second, call); err != nil {
			t.Errorf("err = %v, want nil", err)
		}
	}
	if pub.count() != 0 {
		t.Errorf("publish calls = %d, want 0", pub.count())
	}
	if !strings.Contains(buf.String(), "no journal") {
		t.Errorf("no warning that the critical record is not durable; got:\n%s", buf.String())
	}
}

// nil 알림기는 무동작(Notifier.AnnounceOperatingMode 의 n == nil 과 같은 계약).
func TestA092RecordOnlyNilNotifierIsANoOp(t *testing.T) {
	r := obs.RecordOnly{}
	if err := r.Notify(context.Background(), a096Event()); err != nil {
		t.Errorf("Notify: %v", err)
	}
	if err := r.AnnounceOperatingMode(context.Background(), journal.ModeNormal,
		a092ModeRecord("m", journal.ModeEntryBlocked)); err != nil {
		t.Errorf("AnnounceOperatingMode: %v", err)
	}
}

// 기록 전용 모드 통지: publish 없이 PENDING 행 하나, 키에 전이 행 신원(rec.ID).
func TestA092RecordOnlyAnnouncementIsARowNotASend(t *testing.T) {
	pub := &stuckPublisher{}
	n, j, _, _ := a096Notifier(t, pub)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	rec := a092ModeRecord("mode-row-7", journal.ModeEntryBlocked)
	if err := a092Within(t, 5*time.Second, func() error {
		return obs.RecordOnly{N: n}.AnnounceOperatingMode(ctx, journal.ModeNormal, rec)
	}); err != nil {
		t.Fatalf("AnnounceOperatingMode: %v", err)
	}
	if pub.count() != 0 {
		t.Errorf("publish calls = %d, want 0", pub.count())
	}
	rows, _ := j.PendingAlerts(ctx, 0)
	if len(rows) != 1 {
		t.Fatalf("pending rows = %d, want 1", len(rows))
	}
	if rows[0].Type != string(obs.EventOperatingMode) || !strings.Contains(rows[0].EventKey, rec.ID) {
		t.Errorf("row type/key = %s/%q, want %s keyed by the transition id %q",
			rows[0].Type, rows[0].EventKey, obs.EventOperatingMode, rec.ID)
	}
}

// K1: 동기 통지와 기록 전용 통지는 사건 구성 하나를 같이 씀 — 같은 전이면 제목 · 본문 · 필드 · 키가 같다.
func TestA092BothAnnouncersBuildTheSameEvent(t *testing.T) {
	pubA := &failingPublisher{}
	nA, jA, _, _ := a096Notifier(t, pubA)
	nB, jB, _, _ := a096Notifier(t, &stuckPublisher{})
	ctx := context.Background()
	rec := a092ModeRecord("mode-row-9", journal.ModeEntryBlocked)

	if err := nA.AnnounceOperatingMode(ctx, journal.ModeNormal, rec); err != nil {
		t.Fatalf("sync: %v", err)
	}
	if err := (obs.RecordOnly{N: nB}).AnnounceOperatingMode(ctx, journal.ModeNormal, rec); err != nil {
		t.Fatalf("record-only: %v", err)
	}
	a := a092OnlyRow(t, jA, true)
	b := a092OnlyRow(t, jB, false)
	if a.EventKey != b.EventKey || a.Title != b.Title || a.Body != b.Body || a.Payload != b.Payload || a.Type != b.Type {
		t.Errorf("the two announcers disagree:\n sync   %q %q %q %q\n record %q %q %q %q",
			a.EventKey, a.Title, a.Body, a.Payload, b.EventKey, b.Title, b.Body, b.Payload)
	}
	if !strings.Contains(a.EventKey, rec.ID) {
		t.Errorf("sync key %q does not carry the transition id", a.EventKey)
	}
}

// K1: 강화 → 완화 → 재알림 창 안의 재강화 → 재완화 — 전이마다 새 통지 행. 계정+모드 키였다면 셋째·넷째가 옛 정착 행에 흡수됨.
// 변화 없는 재강화는 행을 만들지 않음(원장의 무변화 규칙).
func TestA092EachTransitionIsItsOwnAnnouncement(t *testing.T) {
	pub := &failingPublisher{}
	n, j, _, _ := a096Notifier(t, pub)
	ctx := context.Background()
	aud := &a092Auditor{}
	const acct = "acct-a092"

	tighten := func() bool {
		_, changed, err := j.EscalateOperatingMode(ctx, acct, journal.ModeTriggerExitObservationOutage, n)
		if err != nil {
			t.Fatalf("escalate: %v", err)
		}
		return changed
	}
	relax := func() {
		if _, _, err := j.TransitionOperatingMode(ctx, journal.TransitionModeRequest{
			AccountRef: acct, Mode: journal.ModeNormal, Cause: "checked", Actor: journal.ModeActorOperator,
			Approval: "ticket-1", Auditor: aud, Announcer: n,
		}); err != nil {
			t.Fatalf("relax: %v", err)
		}
	}
	if !tighten() {
		t.Fatal("first tighten did not change anything")
	}
	if tighten() { // 변화 없음 → 통지 없음
		t.Fatal("a repeated tighten changed the mode")
	}
	relax()
	if !tighten() {
		t.Fatal("re-tighten did not change anything")
	}
	relax()

	if got := pub.callCount(); got != 4 {
		t.Errorf("publish calls = %d, want 4 — one per transition, none for the no-op", got)
	}
	keys := map[string]bool{}
	for _, id := range a092AllAlertIDs(t, j) {
		row, _ := j.LookupAlert(ctx, id)
		if row.Type == string(obs.EventOperatingMode) {
			keys[row.EventKey] = true
		}
	}
	if len(keys) != 4 {
		t.Errorf("distinct announcement rows = %d, want 4: %v", len(keys), keys)
	}
}

// 구조 핀: 기록 전용 입구는 알림기의 배제 잠금 아래에서 RecordAlert 를 부름(정본 「셈~해제 배제」의 부류).
// 잠금 → RecordAlert → 해제 순서이고 publish 계열 호출이 없음.
func TestA092RecordEntryRecordsUnderTheNotifierLock(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "record_only.go", nil, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var body *ast.BlockStmt
	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok && fd.Name.Name == "recordCritical" {
			body = fd.Body
		}
	}
	if body == nil {
		t.Fatal("recordCritical not found")
	}
	var order []string
	ast.Inspect(body, func(n ast.Node) bool {
		c, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		s, ok := c.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		switch s.Sel.Name {
		case "Lock", "Unlock":
			if x, ok := s.X.(*ast.SelectorExpr); ok && x.Sel.Name == "mu" {
				order = append(order, s.Sel.Name)
			}
		case "RecordAlert", "Publish", "publishBestEffort", "deliver", "claimAndDeliver", "Notify", "ClaimAlertForDelivery":
			order = append(order, s.Sel.Name)
		}
		return true
	})
	want := []string{"Lock", "RecordAlert", "Unlock"}
	if strings.Join(order, ",") != strings.Join(want, ",") {
		t.Errorf("call order = %v, want %v", order, want)
	}
}

type a092Auditor struct{}

func (a092Auditor) RecordAction(string, string, string, string) error { return nil }

func a092OnlyRow(t *testing.T, j *journal.Journal, delivered bool) journal.Alert {
	t.Helper()
	ids := a092AllAlertIDs(t, j)
	if len(ids) != 1 {
		t.Fatalf("rows = %d, want 1", len(ids))
	}
	row, err := j.LookupAlert(context.Background(), ids[0])
	if err != nil {
		t.Fatal(err)
	}
	if delivered != (row.State == journal.AlertDelivered) {
		t.Fatalf("state = %s, delivered=%v expected", row.State, delivered)
	}
	return row
}

// a092AllAlertIDs 는 id 1 부터 없는 id 가 나올 때까지 읽음 — outbox 는 AUTOINCREMENT 이고 이 시험들은 행을 지우지 않음.
func a092AllAlertIDs(t *testing.T, j *journal.Journal) []int64 {
	t.Helper()
	var ids []int64
	for id := int64(1); ; id++ {
		if _, err := j.LookupAlert(context.Background(), id); err != nil {
			if errors.Is(err, journal.ErrAlertNotFound) {
				return ids
			}
			t.Fatalf("LookupAlert(%d): %v", id, err)
		}
		ids = append(ids, id)
	}
}

// 25.6: obs 밖 기록자(a066 완화 통지)가 쓰는 공개 입구. 등급과 무관하게 critical 로 기록하고, durable 하지 못하면 오류.
func TestA092RecordCriticalIsTheEntryForOutsideRecorders(t *testing.T) {
	pub := &stuckPublisher{}
	n, j, _, clk := a096Notifier(t, pub)
	ctx := context.Background()
	e := obs.Event{Type: obs.EventType("engine.risk_relaxation"), Key: "engine.risk_relaxation|entry_lock|1",
		Title: "RISK RELAXATION: x", Body: "b", Fields: map[string]any{"kind": "entry_lock"}}
	if obs.SeverityOf(e.Type) == obs.SeverityCritical {
		t.Fatal("fixture: the type must be normal-grade so the test shows the entry does not grade")
	}
	if err := a092Within(t, 5*time.Second, func() error { return n.RecordCritical(ctx, e, 0) }); err != nil {
		t.Fatalf("RecordCritical: %v", err)
	}
	rows, _ := j.PendingAlerts(ctx, 0)
	if len(rows) != 1 || rows[0].Severity != string(obs.SeverityCritical) || rows[0].EventKey != e.Key || rows[0].Payload != `{"kind":"entry_lock"}` {
		t.Fatalf("rows = %+v, want one critical row keyed %q", rows, e.Key)
	}
	if pub.count() != 0 {
		t.Errorf("publish calls = %d, want 0", pub.count())
	}
	// 재알림 창 0: 정착 행은 재무장하지 않음(M15).
	claim, err := j.ClaimAlertByID(ctx, rows[0].ID, "deliverer")
	if err != nil || claim.Disposition != journal.ClaimAcquired {
		t.Fatalf("claim: %v %v", claim.Disposition, err)
	}
	if _, err := j.MarkAlertDelivered(ctx, rows[0].ID, claim.Token); err != nil {
		t.Fatal(err)
	}
	clk.Advance(1000 * time.Hour)
	if err := n.RecordCritical(ctx, e, 0); err != nil {
		t.Fatal(err)
	}
	if rows, _ := j.PendingAlerts(ctx, 0); len(rows) != 0 {
		t.Errorf("remindAfter=0 re-armed the settled row")
	}
}

func TestA092RecordCriticalRefusesWhenNotDurable(t *testing.T) {
	var none *obs.Notifier
	if err := none.RecordCritical(context.Background(), a096Event(), 0); !errors.Is(err, obs.ErrAlertNotDurable) {
		t.Errorf("nil notifier: err = %v, want ErrAlertNotDurable", err)
	}
	pub := &stuckPublisher{}
	n := &obs.Notifier{Publisher: pub}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	err := a092Within(t, 5*time.Second, func() error { return n.RecordCritical(ctx, a096Event(), 0) })
	if !errors.Is(err, obs.ErrAlertNotDurable) {
		t.Errorf("no journal: err = %v, want ErrAlertNotDurable", err)
	}
	if pub.count() != 0 {
		t.Errorf("publish calls = %d, want 0", pub.count())
	}
}

// Manager 판정 2b(2026-09-30): RecordCritical 의 구조화 로그 줄에 필드(계좌를 담은 대상)가 새지 않음. 원장 payload 에는 남음.
func TestA092RecordCriticalKeepsFieldsOutOfTheLog(t *testing.T) {
	buf := &bytes.Buffer{}
	n, j, _, _ := a096Notifier(t, &stuckPublisher{})
	n.Log = obs.NewLogger(obs.LogOptions{Writer: buf, JSON: true, Clock: clock.NewFake(obsNow)})
	e := obs.Event{Type: obs.EventType("engine.risk_relaxation"), Key: "engine.risk_relaxation|entry_lock|1",
		Title: "RISK RELAXATION: entry_loss_lock:KR/SHORT", Body: "operator ops released entry_loss_lock:KR/SHORT",
		Fields: map[string]any{"target": "entry_loss_lock:acct-SECRET-7/KR/SHORT"}}
	if err := n.RecordCritical(context.Background(), e, 0); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(buf.String(), "acct-SECRET-7") {
		t.Errorf("the account reference reached the structured log:\n%s", buf.String())
	}
	if !strings.Contains(buf.String(), "engine.risk_relaxation") {
		t.Errorf("no log line for the record:\n%s", buf.String())
	}
	rows, _ := j.PendingAlerts(context.Background(), 0)
	if len(rows) != 1 || !strings.Contains(rows[0].Payload, "acct-SECRET-7") {
		t.Errorf("the payload lost the full target: %+v", rows)
	}
}
