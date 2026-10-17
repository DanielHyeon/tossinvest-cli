package journal

// a124 tasks 2.7 · 2.4/2.5(원장 절반) — 판정 입력은 **정산 트랜잭션이 커밋한** attempts 다.
//
// 배달 실행자가 「이 행이 재시도 한도에 이르렀다」를 판정할 때 쓸 수 있는 값은 둘이었다.
// 나열 시점의 `Alert.Attempts` 와 정산이 방금 쓴 값. 앞의 것은 낡는다 — 배치 뒤쪽 행은
// 나열 뒤 한참 뒤에 임차되고, 그 사이 다른 발송자가 올리거나 재무장이 0 으로 되돌린다
// (design D1, freeze F1). 그래서 판정은 CAS(id · PENDING · 토큰) 아래 **같은 트랜잭션**에서
// 읽은 값에 서야 하고, HEAD 의 `SettleResult` 에는 그 값을 담을 칸이 없다(`alert_claim.go:140-145`).
//
// 가산 읽기의 원자성(Z4): 읽기나 커밋이 실패하면 정산은 없던 일이어야 하고, 돌아온 결과는
// 판정에 쓸 수 없어야 한다(영값 `SettleResult{}` 의 Outcome 은 `SettleApplied` 다 — F10,
// 그래서 호출자는 err 를 먼저 본다; 여기서는 원장이 err 를 돌려주는지와 저장 상태가 그대로인지를 잰다).

import (
	"context"
	"database/sql"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
	"testing"
	"time"
)

func a124Row(t *testing.T, j *Journal, key string) int64 {
	t.Helper()
	id, err := j.EnqueueAlert(context.Background(), Alert{EventKey: key, Type: "order.in_doubt", Severity: "critical"})
	if err != nil {
		t.Fatalf("EnqueueAlert: %v", err)
	}
	return id
}

func a124Claim(t *testing.T, j *Journal, id int64) string {
	t.Helper()
	claim, err := j.ClaimAlertByID(context.Background(), id, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertByID(%d): %v", id, err)
	}
	if claim.Disposition != ClaimAcquired {
		t.Fatalf("claim disposition = %v, want acquired", claim.Disposition)
	}
	return claim.Token
}

func a124Stored(t *testing.T, j *Journal, id int64) Alert {
	t.Helper()
	row, err := j.LookupAlert(context.Background(), id)
	if err != nil {
		t.Fatalf("LookupAlert(%d): %v", id, err)
	}
	return row
}

// TestAFailedAttemptReturnsTheAttemptsItCommitted — 한 임차 아래 세 번의 실패 기록이
// 1 · 2 · 3 을 돌려준다.
func TestAFailedAttemptReturnsTheAttemptsItCommitted(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()
	id := a124Row(t, j, "a124-committed")
	token := a124Claim(t, j, id)
	for want := 1; want <= 3; want++ {
		clk.Advance(time.Second)
		res, err := j.MarkAlertAttemptFailed(ctx, id, token, "transport is down")
		if err != nil {
			t.Fatalf("MarkAlertAttemptFailed #%d: %v", want, err)
		}
		if res.Outcome != SettleApplied || res.Attempts != want {
			t.Fatalf("attempt #%d: outcome=%v attempts=%d, want applied %d", want, res.Outcome, res.Attempts, want)
		}
	}
}

// TestTheJudgementIgnoresAListingThatWentStale — 나열은 0 이라고 말했지만 그 사이 다른
// 발송자가 두 번 실패를 기록했다. 정산이 돌려주는 값은 커밋된 3 이다(나열 값 + 1 = 1 이 아니다).
func TestTheJudgementIgnoresAListingThatWentStale(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	id := a124Row(t, j, "a124-stale-listing")

	listed, err := j.PendingAlerts(ctx, 10)
	if err != nil || len(listed) != 1 || listed[0].Attempts != 0 {
		t.Fatalf("arrangement: listed=%+v err=%v", listed, err)
	}

	other := a124Claim(t, j, id)
	for i := 0; i < 2; i++ {
		if _, err := j.MarkAlertAttemptFailed(ctx, id, other, "other sender"); err != nil {
			t.Fatalf("other sender's attempt: %v", err)
		}
	}
	if res, err := j.ReleaseAlertClaim(ctx, id, other); err != nil || res.Outcome != SettleApplied {
		t.Fatalf("other sender's release: %+v %v", res, err)
	}

	mine := a124Claim(t, j, id)
	res, err := j.MarkAlertAttemptFailed(ctx, id, mine, "mine")
	if err != nil {
		t.Fatalf("MarkAlertAttemptFailed: %v", err)
	}
	if res.Attempts != 3 {
		t.Fatalf("attempts = %d, want the committed 3 (the stale listing said %d)", res.Attempts, listed[0].Attempts)
	}
}

// TestARearmedRowCountsFromZero — 재무장은 새 에피소드다. attempts 는 0 부터 다시 센다.
func TestARearmedRowCountsFromZero(t *testing.T) {
	j, clk := outboxJournal(t)
	ctx := context.Background()
	alert := Alert{EventKey: "a124-rearm", Type: "order.in_doubt", Severity: "critical"}
	claim, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant)
	if err != nil || claim.Disposition != ClaimAcquired {
		t.Fatalf("first claim: %+v %v", claim, err)
	}
	for i := 0; i < 3; i++ {
		if _, err := j.MarkAlertAttemptFailed(ctx, claim.ID, claim.Token, "down"); err != nil {
			t.Fatalf("attempt: %v", err)
		}
	}
	if res, err := j.MarkAlertDelivered(ctx, claim.ID, claim.Token); err != nil || res.Outcome != SettleApplied {
		t.Fatalf("deliver: %+v %v", res, err)
	}
	clk.Advance(claimRemind + time.Minute)
	again, err := j.ClaimAlertForDelivery(ctx, alert, claimRemind, testClaimant)
	if err != nil || again.Disposition != ClaimAcquired || again.ID != claim.ID {
		t.Fatalf("re-arm claim: %+v %v", again, err)
	}
	res, err := j.MarkAlertAttemptFailed(ctx, again.ID, again.Token, "down again")
	if err != nil {
		t.Fatalf("MarkAlertAttemptFailed: %v", err)
	}
	if res.Attempts != 1 {
		t.Fatalf("attempts after re-arm = %d, want 1 (a new episode counts from zero)", res.Attempts)
	}
}

// TestAResultThatWroteNothingCarriesNoAttempts — 적용되지 않은 결과는 판정에 쓰지 않는다.
// 값이 0 인 것을 못 박아, 호출자가 실수로 읽어도 한도에 닿지 않게 한다.
func TestAResultThatWroteNothingCarriesNoAttempts(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	id := a124Row(t, j, "a124-not-applied")
	token := a124Claim(t, j, id)
	for i := 0; i < 3; i++ {
		if _, err := j.MarkAlertAttemptFailed(ctx, id, token, "down"); err != nil {
			t.Fatalf("attempt: %v", err)
		}
	}
	lost, err := j.MarkAlertAttemptFailed(ctx, id, "not-my-token", "down")
	if err != nil || lost.Outcome != SettleLeaseLost || lost.Attempts != 0 {
		t.Fatalf("lease lost: %+v %v", lost, err)
	}
	if err := j.AcknowledgeAlert(ctx, id, "operator"); err != nil {
		t.Fatalf("AcknowledgeAlert: %v", err)
	}
	settled, err := j.MarkAlertAttemptFailed(ctx, id, token, "down")
	if err != nil || settled.Outcome != SettleAlreadySettled || settled.Attempts != 0 {
		t.Fatalf("already settled: %+v %v", settled, err)
	}
	missing, err := j.MarkAlertAttemptFailed(ctx, 999, token, "down")
	if err != nil || missing.Outcome != SettleNotFound || missing.Attempts != 0 {
		t.Fatalf("not found: %+v %v", missing, err)
	}
}

// a124SettleCaller 는 settleUnderClaim 을 지나는 세 호출자다(evidence R9).
type a124SettleCaller struct {
	name string
	call func(j *Journal, id int64, token string) (SettleResult, error)
}

var a124SettleCallers = []a124SettleCaller{
	{"MarkAlertAttemptFailed", func(j *Journal, id int64, token string) (SettleResult, error) {
		return j.MarkAlertAttemptFailed(context.Background(), id, token, "down")
	}},
	{"MarkAlertDelivered", func(j *Journal, id int64, token string) (SettleResult, error) {
		return j.MarkAlertDelivered(context.Background(), id, token)
	}},
	{"ReleaseAlertClaim", func(j *Journal, id int64, token string) (SettleResult, error) {
		return j.ReleaseAlertClaim(context.Background(), id, token)
	}},
}

// a124FailTheAttemptsRead 는 가산 읽기를 실패시킨다(Z4 첫째 주입).
func a124FailTheAttemptsRead(t *testing.T) {
	t.Helper()
	orig := readSettledAttemptsTx
	readSettledAttemptsTx = func(context.Context, *sql.Tx, int64) (int, error) {
		return 0, errors.New("a124 synthetic attempts read failure")
	}
	t.Cleanup(func() { readSettledAttemptsTx = orig })
}

// a124FailTheCommit 은 읽기는 성공시키고 **커밋**을 실패시킨다(Z4 둘째 주입).
//
// 같은 트랜잭션 안에 지연 외래 키 위반 하나를 남긴다 — 커밋 순간에만 검사되므로 UPDATE ·
// SELECT 는 성공하고 COMMIT 만 실패한다. 원장은 foreign_keys(on) 으로 열린다(journal.go:223).
// a124CommitFaultArmed 는 커밋 주입이 끝까지 걸렸는지(읽기 성공 + 위반 행 삽입)를 기록한다 — 아니면 시험이
// 「아무 오류」로 통과해 커밋 단계에 닿았다는 증거가 없다(codex 구현 1회차 I5).
var a124CommitFaultArmed bool

func a124FailTheCommit(t *testing.T) {
	t.Helper()
	orig := readSettledAttemptsTx
	a124CommitFaultArmed = false
	readSettledAttemptsTx = func(ctx context.Context, tx *sql.Tx, id int64) (int, error) {
		n, err := orig(ctx, tx, id)
		if err != nil {
			return n, err
		}
		for _, stmt := range []string{
			`CREATE TEMP TABLE IF NOT EXISTS a124_parent (id INTEGER PRIMARY KEY)`,
			`CREATE TEMP TABLE IF NOT EXISTS a124_child (pid INTEGER REFERENCES a124_parent(id) DEFERRABLE INITIALLY DEFERRED)`,
			`INSERT INTO a124_child (pid) VALUES (1)`,
		} {
			if _, xerr := tx.ExecContext(ctx, stmt); xerr != nil {
				return 0, xerr
			}
		}
		a124CommitFaultArmed = true
		return n, nil
	}
	t.Cleanup(func() { readSettledAttemptsTx = orig })
}

// TestAFailedAttemptsReadLeavesTheLedgerAsItWas — 세 호출자 × 두 주입. 실패하면 오류가
// 돌아오고, 저장된 attempts · 상태 · 임차는 그대로이며, 같은 토큰으로 다시 정산할 수 있다.
func TestAFailedAttemptsReadLeavesTheLedgerAsItWas(t *testing.T) {
	for _, inject := range []struct {
		name string
		do   func(*testing.T)
	}{{"read", a124FailTheAttemptsRead}, {"commit", a124FailTheCommit}} {
		for _, caller := range a124SettleCallers {
			t.Run(inject.name+"/"+caller.name, func(t *testing.T) {
				j, _ := outboxJournal(t)
				id := a124Row(t, j, "a124-atomic")
				token := a124Claim(t, j, id)
				if _, err := j.MarkAlertAttemptFailed(context.Background(), id, token, "first"); err != nil {
					t.Fatalf("arrangement attempt: %v", err)
				}
				before := a124Stored(t, j, id)

				trueRead := readSettledAttemptsTx // 주입 전의 참 읽기 — SQL 을 베끼지 않는다(gstack /review 유지보수 전문가)
				inject.do(t)
				res, err := caller.call(j, id, token)
				if err == nil {
					t.Fatalf("%s returned %+v with no error while the injected fault fired", caller.name, res)
				}
				if res != (SettleResult{}) {
					t.Fatalf("%s returned a populated result %+v alongside its error", caller.name, res)
				}
				if inject.name == "commit" {
					if !a124CommitFaultArmed || !strings.Contains(err.Error(), "committing") {
						t.Fatalf("%s: the commit fault did not reach the commit (armed=%v, err=%v)", caller.name, a124CommitFaultArmed, err)
					}
				} else if !strings.Contains(err.Error(), "reading the attempts") {
					t.Fatalf("%s: the error is not the attempts read's (%v)", caller.name, err)
				}
				after := a124Stored(t, j, id)
				if after.State != before.State || after.Attempts != before.Attempts ||
					after.ClaimedBy != before.ClaimedBy || after.LastError != before.LastError {
					t.Fatalf("%s changed the row despite failing: before=%+v after=%+v", caller.name, before, after)
				}
				// 임차(토큰)가 살아 있다 — 같은 토큰의 정산이 다시 적용된다.
				readSettledAttemptsTx = trueRead
				again, err := j.MarkAlertAttemptFailed(context.Background(), id, token, "retry")
				if err != nil || again.Outcome != SettleApplied || again.Attempts != before.Attempts+1 {
					t.Fatalf("the lease did not survive the failed %s: %+v %v", caller.name, again, err)
				}
			})
		}
	}
}

// TestDeliverySelectionPutsRowsBelowTheLimitFirst — 한도 아래 행이 먼저, 그 안에서 오래된 것 먼저,
// 한도 행은 잔여 자리에만(2.4 원장 절반). 한도 행을 버리지 않는다(2.5).
func TestDeliverySelectionPutsRowsBelowTheLimitFirst(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()
	var exhausted []int64
	for i := 0; i < 3; i++ {
		id := a124Row(t, j, "a124-exhausted-"+string(rune('a'+i)))
		token := a124Claim(t, j, id)
		for k := 0; k < 3; k++ {
			if _, err := j.MarkAlertAttemptFailed(ctx, id, token, "down"); err != nil {
				t.Fatalf("attempt: %v", err)
			}
		}
		if _, err := j.ReleaseAlertClaim(ctx, id, token); err != nil {
			t.Fatalf("release: %v", err)
		}
		exhausted = append(exhausted, id)
	}
	fresh1 := a124Row(t, j, "a124-fresh-1")
	fresh2 := a124Row(t, j, "a124-fresh-2")

	got, err := j.PendingAlertsForDelivery(ctx, 3, 3)
	if err != nil {
		t.Fatalf("PendingAlertsForDelivery: %v", err)
	}
	want := []int64{fresh1, fresh2, exhausted[0]}
	if len(got) != len(want) {
		t.Fatalf("selected %d rows, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i].ID != want[i] {
			t.Fatalf("selection order = %v, want %v", a124IDs(got), want)
		}
	}
	all, err := j.PendingAlertsForDelivery(ctx, 0, 3)
	if err != nil || len(all) != 5 {
		t.Fatalf("an exhausted row went missing from the full selection: %v %v", a124IDs(all), err)
	}
	// PendingAlerts 는 그대로다 — 오래된 것 먼저(AA4, 호출자 무편집).
	legacy, err := j.PendingAlerts(ctx, 0)
	if err != nil || len(legacy) != 5 || legacy[0].ID != exhausted[0] {
		t.Fatalf("PendingAlerts order changed: %v %v", a124IDs(legacy), err)
	}
}

func a124IDs(alerts []Alert) []int64 {
	out := make([]int64, 0, len(alerts))
	for _, a := range alerts {
		out = append(out, a.ID)
	}
	return out
}

// readSettledAttemptsTx 는 시험이 결함을 주입하려고 바꾸는 패키지 변수다(Z4). 생산 코드가 그것을 바꾸면 원장
// 정산의 판정 입력이 조용히 다른 함수가 된다 — 비시험 파일에서의 대입은 선언 하나뿐이어야 한다(Eng 리뷰 F5).
func TestOnlyTestsReassignTheAttemptsRead(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}
	var assigned []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, name, nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", name, err)
		}
		ast.Inspect(f, func(n ast.Node) bool {
			as, ok := n.(*ast.AssignStmt)
			if !ok {
				return true
			}
			for _, lhs := range as.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && id.Name == "readSettledAttemptsTx" {
					assigned = append(assigned, fset.Position(as.Pos()).String())
				}
			}
			return true
		})
	}
	if len(assigned) != 0 {
		t.Fatalf("production code reassigns readSettledAttemptsTx at %v", assigned)
	}
}
