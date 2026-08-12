package engine

// a098 — 한 주기가 **무엇을 집는가**. 세 R 이 여기 모인다:
//
//	R22  `PendingAlerts(ctx, N)` 이 id 오름차순 앞의 N 행을 준다      (회귀 핀)
//	R23  한 주기가 batch 를 안 넘고, 나머지는 다음 주기에 나온다
//	R20  **만료된 임차를 루프가 집는다 — 조건을 다시 안 일으켜도**
//
// 셋을 한 파일에 두는 이유는 셋이 같은 질문의 세 부분이기 때문이다: *"이 주기가
// 집는 집합은 무엇인가."* 하나만 재면 나머지가 공허해진다 — 예를 들어 R23 만 재면
// 배치 상한은 맞는데 **집는 순서가 뒤죽박죽인 구현**이 통과하고, 그러면 오래된 행이
// 영원히 뒤로 밀린다.

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

var a098BatchNow = time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)

// a098LeasedJournal opens a ledger whose **clock and lease are the test's**.
//
// ⛔ 이것이 R20 을 잴 수 있게 하는 유일한 조각이다. 임차 만료는 원장의 시계로
// 판정되므로(`alertClaimable`), 원장이 실제 시계를 쓰면 만료를 보려면 81초를
// 기다려야 한다 — 즉 **안 재게 된다.**
func a098LeasedJournal(t *testing.T, lease time.Duration) (*journal.Journal, *clock.Fake) {
	t.Helper()
	fake := clock.NewFake(a098BatchNow)
	j, err := journal.Open(context.Background(), journal.Options{
		Path:       t.TempDir() + "/" + journal.DBFileName,
		Clock:      fake,
		AlertLease: lease,
		FSProber:   journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	return j, fake
}

// TestThePendingListingTakesTheOldestNFirst is R22.
//
// **빨간 시점이 없다 — 오늘도 참이다.** 그런데도 핀을 두는 이유는 실측이다:
// `PendingAlerts` 의 호출 지점 열일곱이 **전부 `limit=0`** 이고, 상한 갈래의
// 커버리지 블록이 **count=0** 이었다. 즉 4.0 이 `batch` 를 넘기기 시작한 이 change
// 가 **그 갈래를 처음 켠다.** 한 번도 실행된 적 없는 코드는 참인 것이 아니라
// 안 재어진 것이다.
func TestThePendingListingTakesTheOldestNFirst(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	ids := make([]int64, 0, 5)
	for i := range 5 {
		ids = append(ids, a098ParkedAlert(t, j, "a098-r22-"+string(rune('a'+i))))
	}

	limited, err := j.PendingAlerts(ctx, 2)
	if err != nil {
		t.Fatalf("PendingAlerts(2): %v", err)
	}
	if len(limited) != 2 {
		t.Fatalf("PendingAlerts(2) returned %d rows, want 2", len(limited))
	}
	if limited[0].ID != ids[0] || limited[1].ID != ids[1] {
		t.Errorf("PendingAlerts(2) = [%d %d], want the two oldest [%d %d] — "+
			"앞의 N 이 아니면 오래된 행이 뒤로 밀린다", limited[0].ID, limited[1].ID, ids[0], ids[1])
	}

	// 대조군: 상한 0 은 전부다. 이것이 없으면 **항상 둘만 주는 구현**도 통과한다.
	all, err := j.PendingAlerts(ctx, 0)
	if err != nil {
		t.Fatalf("PendingAlerts(0): %v", err)
	}
	if len(all) != 5 {
		t.Fatalf("PendingAlerts(0) returned %d rows, want all 5", len(all))
	}
	for i, row := range all {
		if row.ID != ids[i] {
			t.Fatalf("PendingAlerts(0)[%d] = %d, want %d — 오름차순이 아니다", i, row.ID, ids[i])
		}
	}
}

// TestOneCycleTakesTheBatchAndTheNextTakesTheRest is R23.
//
// **4.0b 가 켠 갈래다.** 오늘까지 `PendingAlerts` 의 상한은 프로덕션에서 한 번도
// 안 쓰였다(위 R22의 실측), 그래서 「한 주기가 batch 를 넘지 않는다」는 규약이 아니라
// 희망이었다.
//
// 두 방향을 다 본다: 한 주기가 **넘지 않는 것**과 나머지가 **다음 주기에 나오는 것**.
// 앞엣것만 보면 **아무것도 안 보내는 구현**이 통과하고, 뒤엣것만 보면 한 주기에
// 전부 보내는 구현이 통과한다.
func TestOneCycleTakesTheBatchAndTheNextTakesTheRest(t *testing.T) {
	ctx := context.Background()
	j := openTestJournal(t)
	pub := &a098RecordingPublisher{}
	ids := make([]int64, 0, 5)
	for i := range 5 {
		ids = append(ids, a098ParkedAlert(t, j, "a098-r23-"+string(rune('a'+i))))
	}

	d := &alertDeliverer{
		Journal: j, Publisher: pub, Clock: clock.NewFake(a098BatchNow),
		Interval: alertDeliveryInterval, Batch: 3, Claimant: "a098-test",
	}
	if err := d.cycle(ctx); err != nil {
		t.Fatalf("cycle 1: %v", err)
	}
	if n := pub.count(); n != 3 {
		t.Fatalf("cycle 1 published %d rows, want exactly the batch of 3", n)
	}
	// 그리고 **어느 셋인지** 본다. 수만 세면 매번 같은 셋을 집는 구현도 통과한다.
	for i := range 3 {
		if got := a098AlertState(t, j, ids[i]); got != journal.AlertDelivered {
			t.Errorf("alert %d (the %dth oldest) is %s after cycle 1, want %s",
				ids[i], i+1, got, journal.AlertDelivered)
		}
	}
	for i := 3; i < 5; i++ {
		if got := a098AlertState(t, j, ids[i]); got != journal.AlertPending {
			t.Errorf("alert %d went out in cycle 1; 한 주기가 batch 를 넘었다", ids[i])
		}
	}

	if err := d.cycle(ctx); err != nil {
		t.Fatalf("cycle 2: %v", err)
	}
	if n := pub.count(); n != 5 {
		t.Fatalf("after cycle 2 the publisher saw %d rows, want all 5 — 나머지가 안 나온다", n)
	}
	for i, id := range ids {
		if got := a098AlertState(t, j, id); got != journal.AlertDelivered {
			t.Errorf("alert %d (index %d) is %s after cycle 2", id, i, got)
		}
	}
}

// TestAnExpiredLeaseIsTakenOverWithoutTheConditionHappeningAgain is R20.
//
// **a099 가 만든 만료 술어는 다음 claim 이 일어날 때만 평가된다.** 그 claim 을 하는
// 주체가 이 루프다 — 없으면 죽은 발송자가 쥔 행은 임차가 만료된 뒤에도 **아무도 안
// 집는다.** 그것을 재려면 조건을 다시 일으키면 안 된다: 다시 일으키면 새 행이 생겨
// **옛 행이 안 나가도 초록으로 보인다.**
//
// ⛔ 두 시점을 다 본다. 만료 **전** 주기는 아무것도 안 보내야 하고, 만료 **후**
// 주기가 보내야 한다. 앞엣것이 없으면 「임차를 무시하고 늘 집는 구현」이 통과하고,
// 그 구현은 살아 있는 발송자의 행을 중복 발송한다 — 2026-08-08 의 그 형태다.
func TestAnExpiredLeaseIsTakenOverWithoutTheConditionHappeningAgain(t *testing.T) {
	ctx := context.Background()
	const lease = 10 * time.Second
	j, ledgerClock := a098LeasedJournal(t, lease)
	pub := &a098RecordingPublisher{}
	var logs bytes.Buffer

	id := a098ParkedAlert(t, j, "a098-r20-abandoned")
	// 죽은 발송자: 임차만 쥐고 정산하지 않는다.
	claim, err := j.ClaimAlertByID(ctx, id, "engine.alert_delivery.dead")
	if err != nil {
		t.Fatalf("ClaimAlertByID: %v", err)
	}
	if claim.Disposition != journal.ClaimAcquired {
		t.Fatalf("the dead sender did not get the lease: %v", claim.Disposition)
	}

	d := &alertDeliverer{
		Journal:   j,
		Publisher: pub,
		Log:       obs.NewLogger(obs.LogOptions{Writer: &logs, JSON: true, Clock: ledgerClock}),
		Clock:     clock.NewFake(a098BatchNow),
		Interval:  alertDeliveryInterval, Batch: alertDeliveryBatch, Claimant: "a098-test",
	}

	// --- 만료 전 --------------------------------------------------------------
	if err := d.cycle(ctx); err != nil {
		t.Fatalf("cycle before expiry: %v", err)
	}
	if n := pub.count(); n != 0 {
		t.Fatalf("the executor published %d rows while another sender's lease was live; "+
			"살아 있는 임차를 무시하면 중복 발송이다", n)
	}
	if got := a098AlertState(t, j, id); got != journal.AlertPending {
		t.Fatalf("state before expiry = %s, want %s", got, journal.AlertPending)
	}
	if !strings.Contains(logs.String(), "held by another sender") {
		t.Errorf("the held row was skipped silently: %s", logs.String())
	}

	// --- 만료 후 — 조건은 **다시 안 일으킨다** ---------------------------------
	ledgerClock.Advance(lease + time.Second)
	if err := d.cycle(ctx); err != nil {
		t.Fatalf("cycle after expiry: %v", err)
	}
	if n := pub.count(); n != 1 {
		t.Fatalf("the executor published %d rows after the lease expired, want 1 — "+
			"만료된 임차를 아무도 안 집으면 그 행은 영원히 앉아 있다", n)
	}
	if got := a098AlertState(t, j, id); got != journal.AlertDelivered {
		t.Fatalf("state after expiry = %s, want %s", got, journal.AlertDelivered)
	}
	// 인수는 조용하면 안 된다 — 누군가 죽었거나 멈췄다는 뜻이다.
	if !strings.Contains(logs.String(), "expired alert lease was taken over") {
		t.Errorf("the steal was silent; 발송자가 죽은 것을 아무도 모른다: %s", logs.String())
	}
	if !strings.Contains(logs.String(), "engine.alert_delivery.dead") {
		t.Errorf("the log does not name who was stolen from: %s", logs.String())
	}
	// 그리고 토큰은 절대 안 나간다 (a099 ClaimResult.Token · 안전 불변식 8).
	if strings.Contains(logs.String(), claim.Token) {
		t.Error("a claim token reached the log; 그것을 읽는 누구나 남의 발송을 정산할 수 있다")
	}
}
