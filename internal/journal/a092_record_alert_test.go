package journal

// a092 22.3 C3 · 24.3 M15: 임차 없는 기록 입구(RecordAlert)의 원장 쪽 계약.
//
// exit 관측 goroutine 의 critical 기록은 발송 임차를 잡지 않아야 하고(잡으면 배달 실행자가 임차 길이만큼 못 집음),
// 남의 임차를 풀거나 덮지 않아야 하며, 재알림 창이 지난 정착 행은 다시 무장해야 함. 입구가 0 을 넘기면 재무장하지 않음.

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
	"time"
)

const a092Remind = time.Hour

func a092Alert(key string) Alert {
	return Alert{EventKey: key, Type: "exit.proposal_refused", Severity: "critical", Title: "t", Body: "b"}
}

// 새 행을 넣고 발송 빚을 보고하되 임차는 없음 — 배달 실행자가 곧바로 집을 수 있어야 함.
func TestA092RecordAlertTakesNoLease(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()

	id, owed, err := j.RecordAlert(ctx, a092Alert("k-fresh"), a092Remind)
	if err != nil {
		t.Fatalf("RecordAlert: %v", err)
	}
	if id == 0 || !owed {
		t.Fatalf("id=%d owed=%v, want a new owed row", id, owed)
	}
	row, err := j.LookupAlert(ctx, id)
	if err != nil {
		t.Fatalf("LookupAlert: %v", err)
	}
	if row.State != AlertPending {
		t.Errorf("state = %s, want PENDING", row.State)
	}
	if row.ClaimedBy != "" || row.ClaimedAt != nil || row.ClaimExpiresAt != nil {
		t.Errorf("row carries a lease (by=%q at=%v exp=%v) — a recorder that claims leaves the row unsendable",
			row.ClaimedBy, row.ClaimedAt, row.ClaimExpiresAt)
	}
	claim, err := j.ClaimAlertByID(ctx, id, "deliverer")
	if err != nil {
		t.Fatalf("ClaimAlertByID: %v", err)
	}
	if claim.Disposition != ClaimAcquired {
		t.Errorf("deliverer disposition = %v, want acquired right after the record", claim.Disposition)
	}
}

// 전달된 행 + 재알림 창 경과 → 재무장(PENDING, 새 내용). 정본 「재무장된 outbox 행은 통째로 이번 에피소드」.
func TestA092RecordAlertRearmsASettledRowPastTheWindow(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()
	id := a092Delivered(t, j, "k-rearm")

	clk.Advance(a092Remind + time.Minute)
	next := a092Alert("k-rearm")
	next.Title = "second episode"
	got, owed, err := j.RecordAlert(ctx, next, a092Remind)
	if err != nil {
		t.Fatalf("RecordAlert: %v", err)
	}
	if got != id || !owed {
		t.Fatalf("id=%d owed=%v, want the same row (%d) owed again", got, owed, id)
	}
	row, _ := j.LookupAlert(ctx, id)
	if row.State != AlertPending || row.Title != "second episode" || row.DeliveredAt != nil {
		t.Errorf("row not re-armed as a new episode: state=%s title=%q delivered=%v", row.State, row.Title, row.DeliveredAt)
	}
	if row.ClaimedBy != "" {
		t.Errorf("re-armed row carries a lease by %q", row.ClaimedBy)
	}
}

// 창 안이면 정착 그대로, 빚 없음.
func TestA092RecordAlertInsideTheWindowOwesNothing(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()
	id := a092Delivered(t, j, "k-quiet")

	clk.Advance(a092Remind / 2)
	_, owed, err := j.RecordAlert(ctx, a092Alert("k-quiet"), a092Remind)
	if err != nil {
		t.Fatalf("RecordAlert: %v", err)
	}
	if owed {
		t.Error("owed inside the reminder window")
	}
	if row, _ := j.LookupAlert(ctx, id); row.State != AlertDelivered {
		t.Errorf("state = %s, want DELIVERED untouched", row.State)
	}
}

// 입구 remindAfter=0 은 정착 행을 재무장하지 않음(M15) — 창이 아무리 지나도.
func TestA092RecordAlertZeroWindowNeverRearms(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()
	id := a092Delivered(t, j, "k-zero")

	clk.Advance(1000 * time.Hour)
	_, owed, err := j.RecordAlert(ctx, a092Alert("k-zero"), 0)
	if err != nil {
		t.Fatalf("RecordAlert: %v", err)
	}
	if owed {
		t.Error("remindAfter=0 re-armed a settled row")
	}
	if row, _ := j.LookupAlert(ctx, id); row.State != AlertDelivered {
		t.Errorf("state = %s, want DELIVERED", row.State)
	}
}

// 다른 발송자가 쥔 임차는 풀지도 덮지도 않음.
func TestA092RecordAlertLeavesAForeignLeaseAlone(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()

	claim, err := j.ClaimAlertForDelivery(ctx, a092Alert("k-held"), a092Remind, "other-sender")
	if err != nil || claim.Disposition != ClaimAcquired {
		t.Fatalf("setup claim: %v %v", claim.Disposition, err)
	}
	before, _ := j.LookupAlert(ctx, claim.ID)

	if _, _, err := j.RecordAlert(ctx, a092Alert("k-held"), a092Remind); err != nil {
		t.Fatalf("RecordAlert: %v", err)
	}
	after, _ := j.LookupAlert(ctx, claim.ID)
	if after.ClaimedBy != before.ClaimedBy || !a092SameTime(after.ClaimExpiresAt, before.ClaimExpiresAt) {
		t.Errorf("lease changed: by %q→%q exp %v→%v", before.ClaimedBy, after.ClaimedBy,
			before.ClaimExpiresAt, after.ClaimExpiresAt)
	}
	// 토큰이 그대로면 원래 발송자가 정산할 수 있음.
	if res, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token); err != nil || res.Outcome != SettleApplied {
		t.Errorf("the original holder can no longer settle: res=%v err=%v", res, err)
	}
}

// 잘못된 알림은 쓰기 트랜잭션 전에 거절.
func TestA092RecordAlertRefusesAKeylessAlert(t *testing.T) {
	j, _ := outboxJournal(t)
	if _, _, err := j.RecordAlert(context.Background(), Alert{Type: "x"}, a092Remind); err == nil {
		t.Error("a keyless alert was recorded")
	}
}

// 구조 핀: 알림 하나에 트랜잭션 하나 · 임차 획득 호출 0 · 기록 몸체는 recordAlertTx 하나.
func TestA092RecordAlertIsOneTransactionWithoutAClaim(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "record_alert.go", nil, 0)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	var body *ast.BlockStmt
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if ok && fd.Name.Name == "RecordAlert" && fd.Recv != nil {
			body = fd.Body
		}
	}
	if body == nil {
		t.Fatal("Journal.RecordAlert not found in record_alert.go")
	}
	calls := map[string]int{}
	ast.Inspect(body, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if s, ok := c.Fun.(*ast.SelectorExpr); ok {
				calls[s.Sel.Name]++
			}
			if id, ok := c.Fun.(*ast.Ident); ok {
				calls[id.Name]++
			}
		}
		return true
	})
	if calls["BeginTx"] != 1 || calls["Commit"] != 1 || calls["recordAlertTx"] != 1 {
		t.Errorf("BeginTx=%d Commit=%d recordAlertTx=%d, want 1 each", calls["BeginTx"], calls["Commit"], calls["recordAlertTx"])
	}
	for _, forbidden := range []string{"acquireAlertClaimTx", "ClaimAlertForDelivery", "ClaimAlertByID", "ReleaseAlertClaim"} {
		if calls[forbidden] != 0 {
			t.Errorf("RecordAlert calls %s — the record-only entry must not take or touch a lease", forbidden)
		}
	}
}

func a092Delivered(t *testing.T, j *Journal, key string) int64 {
	t.Helper()
	ctx := context.Background()
	claim, err := j.ClaimAlertForDelivery(ctx, a092Alert(key), a092Remind, testClaimant)
	if err != nil || claim.Disposition != ClaimAcquired {
		t.Fatalf("setup claim: %v %v", claim.Disposition, err)
	}
	if _, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token); err != nil {
		t.Fatalf("MarkAlertDelivered: %v", err)
	}
	return claim.ID
}

func a092SameTime(a, b *time.Time) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Equal(*b)
}
