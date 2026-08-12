package journal

// a099 4라운드 리뷰가 낸 것들. 두 가지를 다룬다.
//
//	1. 이 change 의 제목이 주장하는 "두 발송자가 경합한다"를 **실제로 경합시킨다**.
//	   기존 R1 테스트는 핸들 하나를 공유해서 database/sql 이 둘을 줄 세웠다.
//	2. 정산 경로 넷 중 AcknowledgeAlert 만 임차를 안 지운다.
//
// 초등학생 설명: 1번은 "두 사람이 진짜로 동시에 손을 뻗게" 만드는 것이고,
// 2번은 "사람이 알림을 확인했으면 그 줄에 남은 이름표도 떼는" 것이다.

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// sharedPathJournals opens n 개의 서로 다른 Journal 핸들을 **한 파일**에 연다.
//
// 이것이 요점이다. Open 은 db.SetMaxOpenConns(1) 을 걸고, 그래서 한 핸들 안에서는
// database/sql 이 모든 문장을 커넥션 하나에 줄 세운다 — 같은 핸들을 공유한 두
// goroutine 은 경합하지 않고 **차례로** 실행된다. 핸들이 달라야 커넥션이 달라지고,
// 그래야 BEGIN IMMEDIATE 두 개가 진짜로 같은 파일을 두고 다툰다. 이 change 가
// 막겠다고 한 것(두 번째 Notifier, 다른 프로세스)이 바로 그 모양이다.
func sharedPathJournals(t *testing.T, n int, clk clock.Clock) []*Journal {
	t.Helper()
	path := filepath.Join(t.TempDir(), "journal.db")
	out := make([]*Journal, 0, n)
	for i := 0; i < n; i++ {
		j, err := Open(context.Background(), Options{
			Path:     path,
			Clock:    clk,
			FSProber: FixedFSProber(FSInfo{Name: "ext4", Magic: MagicExt}),
		})
		if err != nil {
			t.Fatalf("Open handle %d: %v", i, err)
		}
		t.Cleanup(func() { _ = j.Close() })
		out = append(out, j)
	}
	return out
}

// TestSeparateHandlesRaceForOnePendingRow is R1 을 실제로 경합시킨 것이다.
//
// 4라운드 지적: 기존 TestTwoSendersReachingOnePendingRowLeaveWithOneRightToSend 는
// 핸들 하나(j)를 두 goroutine 이 공유한다. SetMaxOpenConns(1) 때문에 두 번째
// goroutine 은 첫 번째가 **커밋하고 커넥션을 반납할 때까지** BeginTx 에서 막힌다.
// 그 테스트가 증명하는 것은 "커밋된 claim 뒤의 claim 은 거부된다"이고, 그것은
// 순차 속성이다. 동시성은 한 번도 시험되지 않았다.
//
// 여기서는 핸들을 발송자 수만큼 연다. 커넥션이 달라지므로 BEGIN IMMEDIATE 가
// 진짜로 겹치고, SQLite 의 쓰기 락이 배제의 실제 근거가 된다.
func TestSeparateHandlesRaceForOnePendingRow(t *testing.T) {
	const senders = 8
	clk := clock.NewFake(time.Date(2026, 7, 26, 8, 0, 0, 0, time.UTC))
	handles := sharedPathJournals(t, senders, clk)
	ctx := context.Background()

	id, err := handles[0].EnqueueAlert(ctx, a099Alert())
	if err != nil {
		t.Fatalf("arranging the pending row: %v", err)
	}

	var (
		mu       sync.Mutex
		acquired int
		held     int
		settled  int
		failed   []error
		wg       sync.WaitGroup
	)
	start := make(chan struct{})
	for i := 0; i < senders; i++ {
		wg.Add(1)
		go func(j *Journal, n int) {
			defer wg.Done()
			<-start
			claim, err := j.ClaimAlertByID(ctx, id, fmt.Sprintf("sender-%d", n))
			mu.Lock()
			defer mu.Unlock()
			switch {
			case err != nil:
				failed = append(failed, err)
			case claim.Disposition == ClaimAcquired:
				acquired++
			case claim.Disposition == ClaimHeldElsewhere:
				held++
			case claim.Disposition == ClaimSettled:
				settled++
			}
		}(handles[i], i)
	}
	close(start)
	wg.Wait()

	for _, err := range failed {
		t.Fatalf("ClaimAlertByID: %v", err)
	}
	if acquired != 1 {
		t.Errorf("acquired = %d, want 1 — %d senders each believing the send is theirs "+
			"is the double delivery this change exists to stop", acquired, acquired)
	}
	// 진 쪽이 **무엇을 들었는지**까지 본다. ClaimSettled 는 "보낼 것이 없다"이고,
	// 아직 PENDING 인 critical 알림에 그렇게 답하면 그것은 배제가 아니라 알림 누락이다.
	if settled != 0 {
		t.Errorf("%d senders were told the row was settled — it is PENDING and unsent; "+
			"that answer is a missed alert, not exclusion", settled)
	}
	if held != senders-1 {
		t.Errorf("held-elsewhere = %d, want %d", held, senders-1)
	}
}

// TestAcknowledgingClearsTheLease.
//
// 4라운드 지적: 정산 경로는 넷인데(배달, 재무장, 반납, 승인) alertClaimCleared 를
// 안 쓰는 것은 AcknowledgeAlert 하나다. 그래서 발송 중이던 행을 운영자가 승인하면
// 그 행은 ACKNOWLEDGED 가 되면서 **토큰을 그대로 안고 남는다**.
//
// 배제가 깨지지는 않는다 — 임차를 읽는 술어는 전부 state = PENDING 안에서만 돈다.
// 문제는 토큰이다. 코드 자신이 그것을 "the one column that must never leave the
// ledger"라고 부르는데, backup.go 의 VACUUM INTO 는 저널을 통째로 복사하므로
// 백업마다 그 토큰이 따라 나간다. 정산된 행이 그것을 무기한 들고 있을 이유가 없다.
func TestAcknowledgingClearsTheLease(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()

	claim, err := j.ClaimAlertForDelivery(ctx, a099Alert(), a099Remind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}
	if claim.Disposition != ClaimAcquired {
		t.Fatalf("disposition = %v, want acquired — the arrangement is wrong", claim.Disposition)
	}

	if err := j.AcknowledgeAlert(ctx, claim.ID, "daniel"); err != nil {
		t.Fatalf("AcknowledgeAlert: %v", err)
	}

	var (
		state     string
		token     string
		by        string
		at        sql.NullString
		expiresAt sql.NullString
	)
	if err := j.db.QueryRowContext(ctx,
		`SELECT state, claim_token, claimed_by, claimed_at, claim_expires_at
		   FROM alert_outbox WHERE id = ?`, claim.ID).
		Scan(&state, &token, &by, &at, &expiresAt); err != nil {
		t.Fatalf("reading the acknowledged row: %v", err)
	}
	if state != AlertAcknowledged {
		t.Fatalf("state = %q, want %q", state, AlertAcknowledged)
	}
	if token != "" {
		t.Error("the acknowledged row still carries the claim token — every backup " +
			"(VACUUM INTO) copies it, and whoever reads it can settle somebody else's send")
	}
	if by != "" || at.Valid || expiresAt.Valid {
		t.Errorf("the acknowledged row still names a holder: by=%q at=%v expires=%v — "+
			"a settled row holds no lease", by, at, expiresAt)
	}
}

// TestReleasingAClaimTwiceNamesNoHolder 는 원장이 무엇을 답하는지를 고정한다.
//
// 반납은 claim_token 을 지우므로 같은 토큰의 두 번째 반납은 아무 행도 안 맞고,
// explainSettleTx 는 행이 아직 PENDING 이라는 이유로 SettleLeaseLost 를 낸다 —
// **보유자 이름이 빈 채로**. 원장의 대답 자체는 정직하다: "네 토큰은 이 행에 없다".
//
// 틀린 것은 그것을 읽는 쪽이었다. 발송자는 이 결과를 "누가 네 임차를 가져갔다"는
// 경고로 찍었고, 아무도 안 가져갔다. 그 고침은 obs 쪽 logLeaseLost 에 있고
// (a099_round4_test.go 의 발송자 절반), 여기서는 그 고침이 딛고 선 사실 —
// 빈 이름 — 을 고정해서 원장이 조용히 이름을 지어내지 않게 한다.
func TestReleasingAClaimTwiceNamesNoHolder(t *testing.T) {
	j, _ := outboxJournal(t)
	ctx := context.Background()

	claim, err := j.ClaimAlertForDelivery(ctx, a099Alert(), a099Remind, testClaimant)
	if err != nil {
		t.Fatalf("ClaimAlertForDelivery: %v", err)
	}

	first, err := j.ReleaseAlertClaim(ctx, claim.ID, claim.Token)
	if err != nil {
		t.Fatalf("first release: %v", err)
	}
	if first.Outcome != SettleApplied {
		t.Fatalf("first release outcome = %v, want applied", first.Outcome)
	}

	second, err := j.ReleaseAlertClaim(ctx, claim.ID, claim.Token)
	if err != nil {
		t.Fatalf("second release: %v", err)
	}
	if second.Outcome != SettleLeaseLost {
		t.Fatalf("second release outcome = %v, want lease-lost — the row is still PENDING "+
			"and the token is gone", second.Outcome)
	}
	if second.ClaimedBy != "" {
		t.Errorf("the ledger named %q as the holder of a row that holds no lease. "+
			"Nobody took this claim; it was handed back", second.ClaimedBy)
	}
}
