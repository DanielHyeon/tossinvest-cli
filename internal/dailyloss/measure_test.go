// a137 측정기 순수 함수 시험. RED 는 태그 a137red 아래의 컴파일 실패로 관찰함(openspec/changes/a137-daily-loss-is-measured/analysis/red/red-dailyloss.txt).
// GREEN 에서 태그를 지움.
package dailyloss

import (
	"errors"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

func amount(s string) *string { return &s }

// 기준 시각 — 측정 순간(AsOf)으로도 쓰임.
var now = time.Date(2026, 10, 14, 6, 0, 0, 0, time.UTC)

func managed(market, closedAt string, pnl *string) Closed {
	return Closed{PositionID: market + "-" + closedAt, Market: market, ClosedAt: closedAt, Managed: true, Outcome: pnl}
}

// Q12: 손실만 셈 — 이익은 상쇄하지 않음.
func TestA137ProfitsDoNotOffsetLosses(t *testing.T) {
	got, err := Collect("2026-10-14", []Closed{
		managed("KR", "2026-10-14T01:00:00Z", amount("-60000")),
		managed("KR", "2026-10-14T02:00:00Z", amount("80000")),
		managed("KR", "2026-10-14T03:00:00Z", amount("0")),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ByCurrency["KRW"] != "60000" {
		t.Fatalf("KRW loss = %q, want 60000 (profits must not offset)", got.ByCurrency["KRW"])
	}
}

// Q4 · 리뷰 P1-1: 귀속은 계좌 일 키(clock.AccountTradingDay — KST 08:00 롤오버) 하나. 시장 현지일로 귀속하지 않음 —
// closed_at 은 원장 커밋 시각이라 늦게 기록된 손실이 끝난 날로 떨어지면 영원히 세어지지 않음(fail-open).
// D = 2026-10-14 의 창 = [2026-10-13T23:00Z, 2026-10-14T23:00Z).
func TestA137AttributionIsTheOneAccountDay(t *testing.T) {
	got, err := Collect("2026-10-14", []Closed{
		managed("KR", "2026-10-13T22:59:59Z", amount("-1")),    // KST 10-14 07:59:59 → 전 일 키, 빠짐
		managed("KR", "2026-10-13T23:00:00Z", amount("-2")),    // KST 10-14 08:00 → 들어감(경계 포함)
		managed("KR", "2026-10-14T14:00:00Z", amount("-1000")), // KST 10-14 23:00 → 들어감
		managed("KR", "2026-10-14T15:00:00Z", amount("-7")),    // KST 10-15 00:00 → 아직 일 키 D, 들어감
		managed("US", "2026-10-14T03:00:00Z", amount("-5")),    // NY 10-13 23:00 EDT 지만 KST 10-14 12:00 → 들어감
		managed("US", "2026-10-14T23:00:00Z", amount("-9")),    // KST 10-15 08:00 → 다음 일 키, 빠짐
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.ByCurrency["KRW"] != "1009" || got.ByCurrency["USD"] != "5" {
		t.Fatalf("by currency = %v, want KRW 1009 · USD 5", got.ByCurrency)
	}
}

// 늦은 기록: 브로커에서 전날 체결된 손절이 KST 08:00 뒤에 원장에 반영되면 오늘 일 키에 들어감.
func TestA137ALateRecordedLossLandsInTodaysKey(t *testing.T) {
	got, err := Collect("2026-10-14", []Closed{managed("US", "2026-10-14T00:30:00Z", amount("-40"))})
	if err != nil || got.ByCurrency["USD"] != "40" {
		t.Fatalf("got %v err %v, want USD 40", got.ByCurrency, err)
	}
}

// Q13 · U1: 관리 대상이 아닌 포지션과 조정으로 닫힌 포지션은 세지도 막지도 않음.
func TestA137UnmanagedAndAdjustmentClosedAreNeitherCountedNorBlocking(t *testing.T) {
	adjusted := managed("KR", "2026-10-14T01:00:00Z", nil)
	adjusted.AdjustmentClosed = true
	adjustedWithRow := managed("KR", "2026-10-14T01:30:00Z", amount("-999"))
	adjustedWithRow.AdjustmentClosed = true
	unmanaged := Closed{PositionID: "x", Market: "KR", ClosedAt: "2026-10-14T02:00:00Z", Outcome: nil}
	got, err := Collect("2026-10-14", []Closed{adjusted, adjustedWithRow, unmanaged})
	if err != nil {
		t.Fatalf("adjustment-closed / unmanaged must not fail closed: %v", err)
	}
	if v := got.ByCurrency["KRW"]; v != "" && v != "0" {
		t.Fatalf("KRW loss = %q, want none", v)
	}
}

// Q13 · 리뷰 P1-2: 조정 종결의 제외는 사람의 앱 매도에만 — 수명 안에 발송된 엔진 SELL 이 있으면 측정 불가.
func TestA137AnAdjustmentCloseWithADispatchedEngineSellIsUnmeasurable(t *testing.T) {
	row := managed("KR", "2026-10-14T01:00:00Z", nil)
	row.AdjustmentClosed, row.EngineSellDispatched = true, true
	if _, err := Collect("2026-10-14", []Closed{row}); !errors.Is(err, ErrUnmeasurable) {
		t.Fatalf("err = %v, want ErrUnmeasurable", err)
	}
}

// Q3: 그날 체결로 닫힌 관리 포지션의 성과 행 부재 → 측정 불가(진입 fail-closed).
func TestA137AManagedCloseWithoutAnOutcomeIsUnmeasurable(t *testing.T) {
	_, err := Collect("2026-10-14", []Closed{managed("KR", "2026-10-14T01:00:00Z", nil)})
	if !errors.Is(err, ErrUnmeasurable) {
		t.Fatalf("err = %v, want ErrUnmeasurable", err)
	}
}

// 다른 날의 성과 행 부재는 오늘을 막지 않음(일 경계에서 풀림).
func TestA137AMissingOutcomeOnAnotherDayDoesNotBlockToday(t *testing.T) {
	if _, err := Collect("2026-10-14", []Closed{managed("KR", "2026-10-13T22:00:00Z", nil)}); err != nil {
		t.Fatalf("err = %v, want nil", err)
	}
}

func TestA137UnreadableLedgerValuesAreUnmeasurable(t *testing.T) {
	for name, row := range map[string]Closed{
		"closed_at":       managed("KR", "yesterday-ish", amount("-1")),
		"empty closed_at": managed("KR", "", amount("-1")),
		"amount":          managed("KR", "2026-10-14T01:00:00Z", amount("lots")),
		"market":          managed("JP", "2026-10-14T01:00:00Z", amount("-1")),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := Collect("2026-10-14", []Closed{row}); !errors.Is(err, ErrUnmeasurable) {
				t.Fatalf("err = %v, want ErrUnmeasurable", err)
			}
		})
	}
}

func TestA137AnEmptyDayKeyIsUnmeasurable(t *testing.T) {
	if _, err := Collect("", nil); !errors.Is(err, ErrUnmeasurable) {
		t.Fatalf("err = %v, want ErrUnmeasurable", err)
	}
}

// Total: 기준 통화로 환산해 riskcalc.DailyLoss 로 합산. 환산은 호출자(Guardian)가 봉인된 권한으로 제공.
func TestA137TotalConvertsThroughTheSuppliedAuthority(t *testing.T) {
	losses := Losses{DayKey: "2026-10-14", ByCurrency: map[string]string{"KRW": "60000", "USD": "10"}}
	convert := func(currency, magnitude string) (string, error) {
		if currency != "USD" || magnitude != "10" {
			t.Fatalf("convert(%q, %q)", currency, magnitude)
		}
		return "14000", nil
	}
	got, err := Total(losses, "KRW", convert, now, now)
	if err != nil {
		t.Fatal(err)
	}
	if got != (riskcalc.Money{Amount: "74000", Currency: "KRW"}) {
		t.Fatalf("total = %+v, want 74000 KRW", got)
	}
}

func TestA137AMissingAuthorityIsUnmeasurable(t *testing.T) {
	losses := Losses{DayKey: "2026-10-14", ByCurrency: map[string]string{"USD": "10"}}
	_, err := Total(losses, "KRW", func(string, string) (string, error) { return "", errors.New("no US authority in this wave") }, now, now)
	if !errors.Is(err, ErrUnmeasurable) {
		t.Fatalf("err = %v, want ErrUnmeasurable", err)
	}
}

func TestA137ANoLossDayIsZero(t *testing.T) {
	got, err := Total(Losses{DayKey: "2026-10-14"}, "KRW", nil, now, now)
	if err != nil || got != (riskcalc.Money{Amount: "0", Currency: "KRW"}) {
		t.Fatalf("total = %+v err = %v, want 0 KRW", got, err)
	}
}
