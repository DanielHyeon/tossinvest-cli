package journal

// a199_order_path_evidence_test.go — HasConfirmedPlace 가 세는 것과 세지 않는 것(a199 G2), 그리고 질의 계획.

import (
	"context"
	"strings"
	"testing"
)

// a199Attempt 는 원장의 생산 쓰기 경로(Prepare → Dispatch)로 attempt 하나를 만들고 최종 상태를 돌려줌.
func a199Attempt(t *testing.T, j *Journal, id, account, source string, kind MutationKind, class DispatchClass) AttemptState {
	t.Helper()
	ctx := context.Background()
	req := PrepareRequest{
		Intent: Intent{ID: "intent-" + id, Market: "kr", TradingDay: "2026-11-20", AccountRef: account, Symbol: "005930",
			Side: "SELL", OrderType: "LIMIT", Quantity: "1", Price: "70000", Currency: "KRW", Source: source,
			Fingerprint: "fp-" + id},
		Kind:      kind,
		AttemptID: "attempt-" + id,
	}
	if kind != KindPlace {
		req.TargetOrderID = "BO-target-" + id
	}
	a, err := j.Prepare(ctx, req)
	if err != nil {
		t.Fatalf("Prepare %s: %v", id, err)
	}
	res, err := a.Dispatch(ctx, func(context.Context, *Attempt) DispatchOutcome {
		return DispatchOutcome{Class: class, BrokerOrderID: "BO-" + id, ReasonCode: "a199", Detail: "a199"}
	})
	if err != nil && class == DispatchAcked {
		t.Fatalf("Dispatch %s: %v", id, err)
	}
	return res.Final
}

func TestA199HasConfirmedPlaceCountsOnlyAConfirmedPlaceOfThatSourceOnThatAccount(t *testing.T) {
	ctx := context.Background()
	const account, engine = "123-45", "engine"
	for _, tc := range []struct {
		name  string
		seed  func(t *testing.T, j *Journal)
		found bool
	}{
		{"empty ledger", func(*testing.T, *Journal) {}, false},
		{"a confirmed engine place on the account", func(t *testing.T, j *Journal) {
			if s := a199Attempt(t, j, "ok", account, engine, KindPlace, DispatchAcked); s != StateConfirmed {
				t.Fatalf("seed state = %s, want CONFIRMED", s)
			}
		}, true},
		{"the account stored with surrounding spaces", func(t *testing.T, j *Journal) {
			a199Attempt(t, j, "spaced", " "+account+" ", engine, KindPlace, DispatchAcked)
		}, true},
		{"another account only", func(t *testing.T, j *Journal) {
			a199Attempt(t, j, "other", "999-99", engine, KindPlace, DispatchAcked)
		}, false},
		{"another producer (flatten) only", func(t *testing.T, j *Journal) {
			a199Attempt(t, j, "flatten", account, "flatten-all", KindPlace, DispatchAcked)
		}, false},
		{"an in-doubt engine place only", func(t *testing.T, j *Journal) {
			if s := a199Attempt(t, j, "doubt", account, engine, KindPlace, DispatchAmbiguous); s == StateConfirmed {
				t.Fatalf("seed state = %s, want a non-confirmed one", s)
			}
		}, false},
		{"a rejected engine place only", func(t *testing.T, j *Journal) {
			if s := a199Attempt(t, j, "rejected", account, engine, KindPlace, DispatchRejected); s == StateConfirmed {
				t.Fatalf("seed state = %s, want a non-confirmed one", s)
			}
		}, false},
		{"a confirmed engine cancel only", func(t *testing.T, j *Journal) {
			if s := a199Attempt(t, j, "cancel", account, engine, KindCancel, DispatchAcked); s != StateConfirmed {
				t.Fatalf("seed state = %s, want CONFIRMED", s)
			}
		}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			j := openTestJournal(t)
			tc.seed(t, j)
			got, err := j.HasConfirmedPlace(ctx, account, engine)
			if err != nil || got != tc.found {
				t.Fatalf("HasConfirmedPlace = %v, %v; want %v", got, err, tc.found)
			}
		})
	}
}

func TestA199HasConfirmedPlaceReportsAReadFailure(t *testing.T) {
	j := openTestJournal(t)
	if err := j.Close(); err != nil {
		t.Fatal(err)
	}
	if got, err := j.HasConfirmedPlace(context.Background(), "123-45", "engine"); err == nil || got {
		t.Fatalf("a closed journal answered %v, %v; want an error", got, err)
	}
}

// 기동 때 한 번 · 감시가 한 시간에 한 번 부름 — attempts 를 상태 인덱스로 좁히고 intents 는 기본 키로 찾아야 함(전체 스캔 금지).
func TestA199HasConfirmedPlaceUsesTheStateIndex(t *testing.T) {
	j := openTestJournal(t)
	// 생산 질의 그 상수를 잼(사본 아님 — 구현 리뷰 r2 처분 2).
	rows, err := j.db.Query("EXPLAIN QUERY PLAN "+confirmedPlaceQuery, "CONFIRMED", "PLACE", "engine", "123-45")
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var plan []string
	for rows.Next() {
		var id, parent, unused int
		var detail string
		if err := rows.Scan(&id, &parent, &unused, &detail); err != nil {
			t.Fatal(err)
		}
		plan = append(plan, detail)
	}
	text := strings.Join(plan, " | ")
	// intents.id 는 TEXT PRIMARY KEY 라 SQLite 가 자동 인덱스(sqlite_autoindex_intents_1)로 찾음.
	if !strings.Contains(text, "SEARCH a USING INDEX idx_attempts_state") ||
		!strings.Contains(text, "SEARCH i USING INDEX sqlite_autoindex_intents_1") || strings.Contains(text, "SCAN") {
		t.Fatalf("query plan = %s, want the attempts state index and the intents primary key, no scan", text)
	}
}
