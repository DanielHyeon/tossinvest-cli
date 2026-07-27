package verifylive

// hours.go is one clock reading and nothing more.
//
// # What was measured, and what was not
//
// On 2026-07-26 (a Sunday) every mutating step of the first live run came back
// HTTP 422 with `code: "order-hours-closed"`, `message: "주문가능일이 아닙니다."`.
// That is a measured fact and it is in measurements.md as M1.
//
// What is NOT measured is the complement: the exact window in which the broker
// *does* accept an order. Pre-market and after-hours sessions, half-days, the
// exchange holiday calendar, and whatever the broker does with a queued order
// outside the continuous session are all unknown to this build. So this file
// draws the crudest line that the one measurement supports — Mon–Fri
// 09:00–15:30 Asia/Seoul, the KR continuous session — and reports which side of
// it the clock is on.
//
// # It is advisory, and the distinction matters
//
// Nothing here blocks anything. A holiday calendar hard-wired into the tool would
// be an unmeasured rail: it would refuse runs on days the broker might well
// accept, and — worse — it would imply the opposite guarantee on days it did not
// refuse. The tool's actual rails are the plan authorisation and the typed
// approval; a wall clock is not one of them, and the surfaces that call this
// render a warning rather than a "no".
//
// A fixed zone rather than time.LoadLocation, for steps.go's reason: a host with
// no tzdata would otherwise fall back to UTC silently, and near midnight that is
// the wrong day.

import (
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

// KRRegularOpen and KRRegularClose bound the KR continuous session in Asia/Seoul,
// as minutes past midnight.
const (
	krRegularOpen  = 9*60 + 0
	krRegularClose = 15*60 + 30
)

// SessionAdvisory is what a surface needs in order to warn without blocking.
type SessionAdvisory struct {
	// At is the reading, in Asia/Seoul.
	At time.Time
	// Outside reports that the clock is not inside Mon–Fri 09:00–15:30 KST.
	Outside bool
	// Label is the short session name: "weekend", "KR regular hours", or
	// "outside KR regular hours".
	Label string
	// Detail is the sentence an operator reads, in the language the only surface
	// that shows it is written in. It names the measured error code verbatim, so
	// the warning can be checked against the evidence rather than believed.
	Detail string
}

// KRSessionAdvisory reads the clock in Asia/Seoul.
//
// now may be in any zone; it is converted. The returned advisory is a statement
// about the wall clock only — it knows nothing about holidays, and says so.
func KRSessionAdvisory(now time.Time) SessionAdvisory {
	kst := time.FixedZone("KST", 9*60*60)
	at := now.In(kst)
	hhmm := at.Hour()*60 + at.Minute()
	weekend := at.Weekday() == time.Saturday || at.Weekday() == time.Sunday

	const closedWarning = "2026-07-26(일) 실측: 휴장 중 POST /api/v1/orders는 HTTP 422 " +
		"order-hours-closed(\"주문가능일이 아닙니다.\")로 거절됐다. mutation 단계는 같은 이유로 실패해 " +
		"측정이 아니라 fail 판정만 남길 가능성이 높다. 읽기 전용 단계는 영향이 없다."

	stamp := at.Format("2006-01-02 15:04")
	switch {
	case weekend:
		return SessionAdvisory{
			At: at, Outside: true, Label: "weekend",
			Detail: fmt.Sprintf("%s KST는 주말이다. %s", stamp, closedWarning),
		}
	case hhmm >= krRegularOpen && hhmm < krRegularClose:
		return SessionAdvisory{
			At: at, Outside: false, Label: "KR regular hours",
			Detail: fmt.Sprintf("%s KST는 KR 연속매매 시간(09:00–15:30) 안이다. 거래소 휴장일은 "+
				"확인하지 않는다 — 그 달력은 미측정이다.", stamp),
		}
	default:
		return SessionAdvisory{
			At: at, Outside: true, Label: "outside KR regular hours",
			Detail: fmt.Sprintf("%s KST는 KR 연속매매 시간(09:00–15:30) 밖이다. %s", stamp, closedWarning),
		}
	}
}

// SessionAdvisoryFor reads the clock for a market.
//
// KR keeps its own reading because it is the one backed by a measurement: this
// account has actually been refused with order-hours-closed, and the advisory
// quotes that code. US goes through internal/clock, which is the calendar the
// rest of the product already judges sessions with (America/New_York, so the
// DST transition is the zone's rather than a fixed offset's). Writing a second
// US calendar here would put two answers to "is the market open" in one
// repository.
//
// The US text says its closed-market response is unmeasured, and it says so
// rather than borrowing KR's code: nothing has observed what this broker returns
// for a US order outside the session, and an advisory that implied otherwise
// would be asserting a fact nobody has.
func SessionAdvisoryFor(market string, now time.Time) SessionAdvisory {
	if NormalizeMarket(market) != MarketUS {
		return KRSessionAdvisory(now)
	}
	return usSessionAdvisory(now)
}

func usSessionAdvisory(now time.Time) SessionAdvisory {
	loc, err := clock.MarketUS.Location()
	if err != nil {
		// No tzdata. Report the reading as unknown rather than guessing an
		// offset: near a DST boundary a guess is simply wrong, and this is an
		// advisory, so "cannot say" is an honest answer that blocks nothing.
		return SessionAdvisory{
			At: now.UTC(), Outside: false, Label: "US session unknown",
			Detail: "이 호스트에 시간대 데이터가 없어 US 장시간을 판정할 수 없다. 안내를 생략한다 — " +
				"판정 불가는 차단 사유가 아니다.",
		}
	}
	at := now.In(loc)
	inside, err := clock.MarketUS.InRegularSession(at)
	if err != nil {
		inside = false
	}
	const unmeasured = "US 휴장·시간외에 이 브로커가 주문에 무엇을 돌려주는지는 이 계좌에서 " +
		"아직 관측되지 않았다([미측정]) — 다른 시장의 실측을 이 시장의 근거로 쓰지 않는다. " +
		"mutation 단계는 실패할 수도, 접수될 수도 있다."
	stamp := at.Format("2006-01-02 15:04")
	zone, _ := at.Zone()
	if inside {
		return SessionAdvisory{
			At: at, Outside: false, Label: "US regular hours",
			Detail: fmt.Sprintf("%s %s는 US 정규장(09:30–16:00 ET) 안이다. 거래소 휴장일은 확인하지 "+
				"않는다 — 그 달력은 미측정이다.", stamp, zone),
		}
	}
	return SessionAdvisory{
		At: at, Outside: true, Label: "outside US regular hours",
		Detail: fmt.Sprintf("%s %s는 US 정규장(09:30–16:00 ET) 밖이다. %s", stamp, zone, unmeasured),
	}
}

// sessionLabel records which session a mutation was accepted in.
//
// It is a KST clock reading, not a market-calendar lookup: the record should say
// what time it was and let the reader judge, rather than bake a holiday table
// into evidence.
func (r *Runner) sessionLabel() string {
	a := SessionAdvisoryFor(r.market, r.now())
	zone, _ := a.At.Zone()
	if a.Label == "weekend" {
		return "weekend " + a.At.Format("2006-01-02 15:04 ") + zone
	}
	return a.Label + " " + a.At.Format("15:04 ") + zone
}
