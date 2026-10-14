package journal

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// a136_daily_buy_test.go 는 하루 매수 총액 상한(risk-management 「하루 매수 총액 한도」)의 원장 판정을 고정함.
//
// 시계: openReservationJournal 은 2026-03-30 00:30 UTC(= 09:30 KST, 계좌 하루 라벨 2026-03-30)에서 시작함.
// 노출 한도는 넉넉하게(1e9) 두어 이 시험에서는 하루 매수 상한만 묶이게 함.

const a136Ceiling = "1000"

// a136Reserve 는 진입 결정을 기록하고 하루 매수 상한 1000 KRW 를 실어 예약함.
func a136Reserve(t *testing.T, j *Journal, id, amount string) (ReserveResult, error) {
	t.Helper()
	recordEntryDecision(t, j, id, "acct-1")
	req := exposureReserve(j, id, "acct-1", amount, "0", "1000000000", mustVersion(t, j, "acct-1"))
	req.DailyBuyLimit = &DailyBuyLimit{Amount: a136Ceiling, Currency: "KRW"}
	return j.Reserve(context.Background(), req)
}

func a136MustReserve(t *testing.T, j *Journal, id, amount string) Decision {
	t.Helper()
	if _, err := a136Reserve(t, j, id, amount); err != nil {
		t.Fatalf("Reserve(%s, %s): %v", id, amount, err)
	}
	dec, err := j.LookupDecision(context.Background(), id)
	if err != nil {
		t.Fatalf("LookupDecision(%s): %v", id, err)
	}
	return dec
}

func a136RequireRefused(t *testing.T, j *Journal, id, amount string) {
	t.Helper()
	_, err := a136Reserve(t, j, id, amount)
	if !errors.Is(err, ErrReservationLimitExceeded) || !strings.Contains(err.Error(), "DAILY_BUY_NOTIONAL") {
		t.Fatalf("Reserve(%s, %s) must be refused by the daily buy ceiling, got %v", id, amount, err)
	}
	rows, err := j.ReservationsForDecision(context.Background(), id)
	if err != nil {
		t.Fatalf("ReservationsForDecision: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("a refused reservation must write nothing, found %+v", rows)
	}
}

func a136Fill(t *testing.T, j *Journal, orderID, state, filled string, terminal bool) {
	t.Helper()
	if _, err := j.RecordFill(context.Background(), FillObservation{
		OrderID: orderID, AccountRef: "acct-1", Symbol: "005930", Market: "kr",
		TradingDay: "2026-03-30", Side: "BUY",
		State: state, Terminal: terminal,
		Quantity: "10", FilledQuantity: filled, AveragePrice: "70000",
		ObservedAt: j.nowString(),
	}); err != nil {
		t.Fatalf("RecordFill(%s): %v", orderID, err)
	}
}

func TestA136ReachingTheDailyBuyCeilingIsRefused(t *testing.T) {
	j, _ := openReservationJournal(t)
	a136MustReserve(t, j, "d-a", "600")
	// 600 + 400 = 1000 → 도달 = 거부(≥).
	a136RequireRefused(t, j, "d-b", "400")
	// 600 + 399 = 999 → 통과.
	a136MustReserve(t, j, "d-c", "399")
}

func TestA136AFilledBuyStillCountsAfterItsHoldIsReleased(t *testing.T) {
	j, _ := openReservationJournal(t)
	dec := a136MustReserve(t, j, "d-fill", "600")
	confirmAttempt(t, j, dec, "attempt-fill", "order-fill")
	a136Fill(t, j, "order-fill", "FILLED", "10", true)
	held, err := j.HeldReservations(context.Background(), "acct-1")
	if err != nil || len(held) != 0 {
		t.Fatalf("the fill must have released the hold (held=%+v, err=%v)", held, err)
	}
	a136RequireRefused(t, j, "d-next", "400")
	a136MustReserve(t, j, "d-small", "399")
}

func TestA136ABuyProvenToHaveFilledNothingDropsOut(t *testing.T) {
	ctx := context.Background()
	cases := map[string]func(t *testing.T, j *Journal, dec Decision){
		"Z2 not dispatched": func(t *testing.T, j *Journal, dec Decision) {
			attempt, err := j.Prepare(ctx, reservationPrepare(dec, "attempt-z"))
			if err != nil {
				t.Fatal(err)
			}
			if err := attempt.Settle(ctx, StateNotDispatched, "test", "never sent"); err != nil {
				t.Fatal(err)
			}
		},
		"Z2 failed confirmed": func(t *testing.T, j *Journal, dec Decision) {
			attempt, err := j.Prepare(ctx, reservationPrepare(dec, "attempt-z"))
			if err != nil {
				t.Fatal(err)
			}
			if err := attempt.MarkDispatchStarted(ctx); err != nil {
				t.Fatal(err)
			}
			if err := attempt.MarkInDoubt(ctx, "test", "no answer"); err != nil {
				t.Fatal(err)
			}
			if err := attempt.ResolveFailed(ctx, "", "proven absent"); err != nil {
				t.Fatal(err)
			}
		},
		"Z3 cancelled with nothing filled": func(t *testing.T, j *Journal, dec Decision) {
			confirmAttempt(t, j, dec, "attempt-z", "order-z")
			a136Fill(t, j, "order-z", "CANCELED", "0", true)
		},
	}
	for name, settle := range cases {
		t.Run(name, func(t *testing.T) {
			j, _ := openReservationJournal(t)
			dec := a136MustReserve(t, j, "d-zero", "600")
			settle(t, j, dec)
			a136MustReserve(t, j, "d-next", "999")
		})
	}
	t.Run("Z1 expired with no attempt", func(t *testing.T) {
		j, fake := openReservationJournal(t)
		a136MustReserve(t, j, "d-zero", "600")
		// 결정 TTL(10분) 뒤 다음 예약 tx 의 precheck sweep 이 EXPIRED_UNCONSUMED 로 해제함.
		fake.Advance(11 * time.Minute)
		a136MustReserve(t, j, "d-next", "999")
		res, err := j.ReservationsForDecision(ctx, "d-zero")
		if err != nil || len(res) != 1 || res[0].ReleaseReason != ReleaseReasonExpiredUnconsumed {
			t.Fatalf("the unconsumed expiry is the precondition of this case, got %+v (%v)", res, err)
		}
	})
}

func TestA136ABuyNotProvenToHaveFilledNothingStillCounts(t *testing.T) {
	cases := map[string]func(t *testing.T, j *Journal, dec Decision){
		"confirmed and still held": func(t *testing.T, j *Journal, dec Decision) {
			confirmAttempt(t, j, dec, "attempt-n", "order-n")
		},
		"partially filled then cancelled": func(t *testing.T, j *Journal, dec Decision) {
			confirmAttempt(t, j, dec, "attempt-n", "order-n")
			a136Fill(t, j, "order-n", "CANCELED", "3", true)
		},
		"zero fill but not terminal": func(t *testing.T, j *Journal, dec Decision) {
			confirmAttempt(t, j, dec, "attempt-n", "order-n")
			a136Fill(t, j, "order-n", "OPEN", "0", false)
		},
		"operator-released hold with only another day's zero-fill snapshot of the reused id": func(t *testing.T, j *Journal, dec Decision) {
			// freeze 리뷰 P1-1 반례: 오늘 주문 스냅숏은 없고 다른 날 같은 id 가 「종결 · 체결 0」 이며 보유는 운영자가 풀었음.
			confirmAttempt(t, j, dec, "attempt-n", "order-n")
			ctx := context.Background()
			if _, err := j.db.ExecContext(ctx,
				`INSERT INTO scoped_fill_snapshots (order_id, account_ref, market, trading_day, symbol, side,
				   state, terminal, quantity, filled_quantity, committed_at)
				 VALUES ('order-n','acct-1','kr','2026-03-27','005930','BUY','CANCELED',1,'10','0','2026-03-30T00:31:00Z')`); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.ExecContext(ctx,
				`UPDATE risk_reservations SET state='RELEASED', released_at=?, release_reason='OPERATOR' WHERE decision_id=?`,
				j.nowString(), dec.ID); err != nil {
				t.Fatal(err)
			}
		},
		"broker-terminal release with only another day's zero-fill snapshot": func(t *testing.T, j *Journal, dec Decision) {
			confirmAttempt(t, j, dec, "attempt-n", "order-n")
			ctx := context.Background()
			if _, err := j.db.ExecContext(ctx,
				`INSERT INTO scoped_fill_snapshots (order_id, account_ref, market, trading_day, symbol, side,
				   state, terminal, quantity, filled_quantity, committed_at)
				 VALUES ('order-n','acct-1','kr','2026-03-27','005930','BUY','CANCELED',1,'10','0','2026-03-30T00:31:00Z')`); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.ExecContext(ctx,
				`UPDATE risk_reservations SET state='RELEASED', released_at=?, release_reason='BROKER_TERMINAL' WHERE decision_id=?`,
				j.nowString(), dec.ID); err != nil {
				t.Fatal(err)
			}
		},
		"zero fill but the order was replaced": func(t *testing.T, j *Journal, dec Decision) {
			confirmAttempt(t, j, dec, "attempt-n", "order-n")
			a136Fill(t, j, "order-n", "CANCELED", "0", true)
			if _, err := j.db.ExecContext(context.Background(),
				`INSERT INTO scoped_lineage_edges (parent_order_id, child_order_id, relation, parent_filled_quantity,
				   requested_quantity, account_ref, market, trading_day, symbol, side, intent_id, attempt_id, created_at)
				 VALUES ('order-n','order-child','REPLACE','0','10','acct-1','kr','2026-03-30','005930','BUY',
				   'intent-attempt-n','attempt-n',?)`, j.nowString()); err != nil {
				t.Fatal(err)
			}
		},
		"operator release even with today's exact-scope zero-fill snapshot": func(t *testing.T, j *Journal, dec Decision) {
			// 해제 사유 가드(Z3 는 BROKER_TERMINAL 만)를 범위 검사와 따로 고정함 — 범위가 맞는 종결·체결 0 행이 있어도 운영자 해제는 셈.
			confirmAttempt(t, j, dec, "attempt-n", "order-n")
			ctx := context.Background()
			if _, err := j.db.ExecContext(ctx,
				`INSERT INTO scoped_fill_snapshots (order_id, account_ref, market, trading_day, symbol, side,
				   state, terminal, quantity, filled_quantity, committed_at)
				 VALUES ('order-n','acct-1','kr','2026-03-30','005930','BUY','CANCELED',1,'10','0','2026-03-30T00:31:00Z')`); err != nil {
				t.Fatal(err)
			}
			if _, err := j.db.ExecContext(ctx,
				`UPDATE risk_reservations SET state='RELEASED', released_at=?, release_reason='OPERATOR' WHERE decision_id=?`,
				j.nowString(), dec.ID); err != nil {
				t.Fatal(err)
			}
		},
		"operator-released before any attempt": func(t *testing.T, j *Journal, dec Decision) {
			// Z1 은 EXPIRED_UNCONSUMED 만 — 운영자가 attempt 없는 보유를 풀었으면 증명이 아니라 셈.
			if _, err := j.db.ExecContext(context.Background(),
				`UPDATE risk_reservations SET state='RELEASED', released_at=?, release_reason='OPERATOR' WHERE decision_id=?`,
				j.nowString(), dec.ID); err != nil {
				t.Fatal(err)
			}
		},
		"unresolved in doubt": func(t *testing.T, j *Journal, dec Decision) {
			ctx := context.Background()
			attempt, err := j.Prepare(ctx, reservationPrepare(dec, "attempt-n"))
			if err != nil {
				t.Fatal(err)
			}
			if err := attempt.MarkDispatchStarted(ctx); err != nil {
				t.Fatal(err)
			}
			if err := attempt.MarkInDoubt(ctx, "test", "no answer"); err != nil {
				t.Fatal(err)
			}
			if err := attempt.ResolveUnresolved(ctx, "", "neither proven"); err != nil {
				t.Fatal(err)
			}
		},
	}
	for name, settle := range cases {
		t.Run(name, func(t *testing.T) {
			j, _ := openReservationJournal(t)
			dec := a136MustReserve(t, j, "d-count", "600")
			settle(t, j, dec)
			a136RequireRefused(t, j, "d-next", "400")
		})
	}
}

func TestA136TheAccountDayRollsOverAtEightKST(t *testing.T) {
	t.Run("a terminal buy leaves the total at 08:00 KST", func(t *testing.T) {
		j, fake := openReservationJournal(t)
		dec := a136MustReserve(t, j, "d-day1", "600")
		confirmAttempt(t, j, dec, "attempt-day1", "order-day1")
		a136Fill(t, j, "order-day1", "FILLED", "10", true)

		fake.Set(time.Date(2026, 3, 30, 22, 59, 59, 0, time.UTC)) // 07:59:59 KST 3-31 — 아직 라벨 3-30
		a136RequireRefused(t, j, "d-late", "400")
		fake.Set(time.Date(2026, 3, 30, 23, 0, 0, 0, time.UTC)) // 08:00:00 KST 3-31 — 라벨 3-31
		a136MustReserve(t, j, "d-day2", "999")
	})
	t.Run("a decision issued before 08:00 that reserves after it is refused", func(t *testing.T) {
		j, fake := openReservationJournal(t)
		ctx := context.Background()
		fake.Set(time.Date(2026, 3, 30, 22, 59, 58, 0, time.UTC)) // 07:59:58 KST 3-31 — 라벨 3-30
		recordEntryDecision(t, j, "d-edge", "acct-1")
		fake.Set(time.Date(2026, 3, 30, 23, 0, 5, 0, time.UTC)) // 08:00:05 KST — 라벨 3-31
		req := exposureReserve(j, "d-edge", "acct-1", "1", "0", "1000000000", mustVersion(t, j, "acct-1"))
		req.DailyBuyLimit = &DailyBuyLimit{Amount: a136Ceiling, Currency: "KRW"}
		if _, err := j.Reserve(ctx, req); !errors.Is(err, ErrReservationLimitExceeded) {
			t.Fatalf("a buy decided on one account day cannot reserve on the next, got %v", err)
		}
	})
	t.Run("a buy still held from yesterday keeps counting", func(t *testing.T) {
		j, fake := openReservationJournal(t)
		dec := a136MustReserve(t, j, "d-day1", "600")
		confirmAttempt(t, j, dec, "attempt-day1", "order-day1")
		fake.Set(time.Date(2026, 3, 30, 23, 30, 0, 0, time.UTC))
		a136RequireRefused(t, j, "d-day2", "400")
	})
}

func TestA136AnExposureRaisingReservationWithoutTheCeilingIsRefused(t *testing.T) {
	j, _ := openReservationJournal(t)
	recordEntryDecision(t, j, "d-bare", "acct-1")
	req := exposureReserve(j, "d-bare", "acct-1", "100", "0", "1000000000", mustVersion(t, j, "acct-1"))
	req.DailyBuyLimit = nil
	if _, err := j.Reserve(context.Background(), req); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("an exposure-raising reservation without the daily buy ceiling must be refused, got %v", err)
	}
	for _, bad := range []DailyBuyLimit{{Amount: "0", Currency: "KRW"}, {Amount: "-1", Currency: "KRW"},
		{Amount: "1000", Currency: ""}, {Amount: "1000", Currency: "USD"}} {
		req.DailyBuyLimit = &bad
		if _, err := j.Reserve(context.Background(), req); !errors.Is(err, ErrInvalidRequest) {
			t.Fatalf("ceiling %+v must be refused as an invalid request, got %v", bad, err)
		}
	}
}

func TestA136AnUnreadableTotalFailsClosed(t *testing.T) {
	j, _ := openReservationJournal(t)
	dec := a136MustReserve(t, j, "d-bad", "600")
	confirmAttempt(t, j, dec, "attempt-bad", "order-bad")
	a136Fill(t, j, "order-bad", "FILLED", "10", true)
	// 해제된 행만 하루 합계가 읽으므로(heldTotals 는 HELD 만) 이 손상은 하루 매수 판정에만 닿음.
	if _, err := j.db.ExecContext(context.Background(),
		"UPDATE risk_reservations SET amount = 'not-a-number' WHERE decision_id = 'd-bad'"); err != nil {
		t.Fatal(err)
	}
	_, err := a136Reserve(t, j, "d-after", "1")
	if err == nil {
		t.Fatal("a total that cannot be computed must refuse the buy, not read the row as zero")
	}
	if rows, _ := j.ReservationsForDecision(context.Background(), "d-after"); len(rows) != 0 {
		t.Fatalf("a refused reservation must write nothing, found %+v", rows)
	}
}

func TestA136ConcurrentBuysCannotBothFitUnderTheCeiling(t *testing.T) {
	j, _ := openReservationJournal(t)
	ctx := context.Background()
	a136MustReserve(t, j, "d-base", "300")
	recordEntryDecision(t, j, "d-x", "acct-1")
	recordEntryDecision(t, j, "d-y", "acct-1")

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		okCount int
		errs    []error
	)
	for _, id := range []string{"d-x", "d-y"} {
		wg.Add(1)
		go func(decision string) {
			defer wg.Done()
			_, err := j.ReserveWithRecollection(ctx, func(ctx context.Context, attempt int) (ReserveRequest, error) {
				version, err := j.ReservationVersion(ctx, "acct-1")
				if err != nil {
					return ReserveRequest{}, err
				}
				req := exposureReserve(j, decision, "acct-1", "400", "0", "1000000000", version)
				req.DailyBuyLimit = &DailyBuyLimit{Amount: a136Ceiling, Currency: "KRW"}
				return req, nil
			}, RecollectPolicy{})
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				okCount++
				return
			}
			errs = append(errs, err)
		}(id)
	}
	wg.Wait()
	// 300 + 400 + 400 = 1100 ≥ 1000 — 둘 중 하나만.
	if okCount != 1 || len(errs) != 1 || !errors.Is(errs[0], ErrReservationLimitExceeded) {
		t.Fatalf("exactly one of two concurrent buys may fit, ok=%d errs=%v", okCount, errs)
	}
}
