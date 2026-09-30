package riskbucket

// a126 D1 · D2 — 떠남 규칙의 유일한 적용 자리(aggregateProductionRiskUsage)를 행 단위로 잼. SQL 이 싣는 사실(영수증 · owner
// released_at · scope latch · owner 키 사본 일치)은 journal 패키지의 실 원장 시험이 따로 잼.

import (
	"errors"
	"testing"
)

func a126Row(id, filled string) productionRiskUsageRow {
	return productionRiskUsageRow{ReservationID: id, PolicyVersion: "policy-v1", HeldMinor: "0", FilledMinor: filled, State: "FILLED",
		SnapshotID: "snapshot-" + id, PolicyRecordDigest: "digest-" + id, OwnerKeyMatches: 1}
}

// released 는 영수증이 있고 owner released_at 이 영수증과 같은 행 — 건강한 해제.
func a126Released(row productionRiskUsageRow) productionRiskUsageRow {
	row.Receipted, row.DecisionReceipted = 1, 1
	row.ReceiptReleasedAt, row.OwnerReleasedAt = "2026-03-30T00:45:00Z", "2026-03-30T00:45:00Z"
	return row
}

func TestA126AggregateLeavesOnlyReceiptedRows(t *testing.T) {
	usage, err := aggregateProductionRiskUsage([]productionRiskUsageRow{a126Row("active", "30"), a126Released(a126Row("gone", "50"))})
	if err != nil {
		t.Fatal(err)
	}
	if usage.FilledMinor != "30" || usage.HeldMinor != "0" {
		t.Fatalf("filled=%s held=%s, want only the active 30 (the receipted 50 leaves)", usage.FilledMinor, usage.HeldMinor)
	}
}

// M1 — 활성 owner 의 행은 떠나지 않음(영수증 없음).
func TestA126AggregateKeepsAnActiveOwnersRows(t *testing.T) {
	usage, err := aggregateProductionRiskUsage([]productionRiskUsageRow{a126Row("a", "30"), a126Row("b", "20")})
	if err != nil || usage.FilledMinor != "50" {
		t.Fatalf("usage=%+v err=%v, want every active row counted", usage, err)
	}
}

// M2 — owner released_at 만 있고 영수증이 없으면 떠나지 않음.
func TestA126AggregateDoesNotLeaveOnAReleaseMarkWithoutAReceipt(t *testing.T) {
	row := a126Row("marked", "50")
	row.OwnerReleasedAt = "2026-03-30T00:45:00Z"
	usage, err := aggregateProductionRiskUsage([]productionRiskUsageRow{row})
	if err != nil || usage.FilledMinor != "50" {
		t.Fatalf("usage=%+v err=%v, want the row counted without a receipt", usage, err)
	}
}

// D4 — 영수증 뒤 scope latch 가 있으면 되돌림(합에 다시 계상).
func TestA126AggregateScopeLatchRevertsTheDeparture(t *testing.T) {
	row := a126Released(a126Row("reverted", "50"))
	row.ScopeLatched = 1
	usage, err := aggregateProductionRiskUsage([]productionRiskUsageRow{row})
	if err != nil || usage.FilledMinor != "50" {
		t.Fatalf("usage=%+v err=%v, want the reverted row counted", usage, err)
	}
}

// M6 · M6b · codex #4 — 영수증 있는 행의 손상과 불일치는 떠남 여부와 무관하게, scope latch 판정 앞에서 거절.
func TestA126AggregateRefusesCorruptReceiptedRows(t *testing.T) {
	cases := map[string]func(productionRiskUsageRow) productionRiskUsageRow{
		"HELD state": func(r productionRiskUsageRow) productionRiskUsageRow { r.State = "HELD"; r.HeldMinor = "5"; return r },
		// 1.5 R4(X3) — HELD 상태 단독(held 0)도 손상. held≠0 절이 대신 막지 못하게 held 를 0 으로 둠.
		"HELD state with held 0":    func(r productionRiskUsageRow) productionRiskUsageRow { r.State = "HELD"; return r },
		"held remainder on FILLED":  func(r productionRiskUsageRow) productionRiskUsageRow { r.HeldMinor = "5"; return r },
		"owner released_at missing": func(r productionRiskUsageRow) productionRiskUsageRow { r.OwnerReleasedAt = ""; return r },
		"owner released_at differs": func(r productionRiskUsageRow) productionRiskUsageRow {
			r.OwnerReleasedAt = "2026-03-30T00:46:00Z"
			return r
		},
		"owner key differs":          func(r productionRiskUsageRow) productionRiskUsageRow { r.OwnerKeyMatches = 0; return r },
		"decision-side receipt only": func(r productionRiskUsageRow) productionRiskUsageRow { r.Receipted, r.OwnerKeyMatches = 0, 0; return r },
	}
	for name, corrupt := range cases {
		for _, latched := range []int{0, 1} {
			row := corrupt(a126Released(a126Row("bad", "50")))
			row.ScopeLatched = latched
			if _, err := aggregateProductionRiskUsage([]productionRiskUsageRow{row}); !errors.Is(err, ErrJournalUsageInvalid) {
				t.Errorf("%s (scope latched=%d): err=%v, want ErrJournalUsageInvalid", name, latched, err)
			}
		}
	}
}

// M7 — 떠난 행의 latch 플래그는 집계에 남음(합에서만 빠짐).
func TestA126AggregateKeepsADepartedRowsLatchFlags(t *testing.T) {
	row := a126Released(a126Row("gone", "50"))
	row.OverageLatched, row.UnknownLatched = 1, 1
	usage, err := aggregateProductionRiskUsage([]productionRiskUsageRow{row})
	if err != nil || usage.FilledMinor != "0" || !usage.Latched || !usage.OverageLatched || !usage.UnknownLatched {
		t.Fatalf("usage=%+v err=%v, want the sum to leave and every latch flag to stay", usage, err)
	}
}

// D2 — RowDigest 형식 불변: 같은 행 집합은 떠남 여부와 무관하게 같은 RowDigest(파생 합은 snapshot digest 의 filled 가 담음).
func TestA126AggregateRowDigestFormatIsUnchanged(t *testing.T) {
	active, err := aggregateProductionRiskUsage([]productionRiskUsageRow{a126Row("r", "50")})
	if err != nil {
		t.Fatal(err)
	}
	gone, err := aggregateProductionRiskUsage([]productionRiskUsageRow{a126Released(a126Row("r", "50"))})
	if err != nil {
		t.Fatal(err)
	}
	if active.RowDigest != gone.RowDigest {
		t.Fatalf("RowDigest moved %s → %s; the digest domain must not change (D2)", active.RowDigest, gone.RowDigest)
	}
	if active.FilledMinor == gone.FilledMinor {
		t.Fatalf("filled %s == %s; the departure must reach the sum", active.FilledMinor, gone.FilledMinor)
	}
}
