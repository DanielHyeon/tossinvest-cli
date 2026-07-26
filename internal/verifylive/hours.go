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

// sessionLabel records which session a mutation was accepted in.
//
// It is a KST clock reading, not a market-calendar lookup: the record should say
// what time it was and let the reader judge, rather than bake a holiday table
// into evidence.
func (r *Runner) sessionLabel() string {
	a := KRSessionAdvisory(r.now())
	if a.Label == "weekend" {
		return "weekend " + a.At.Format("2006-01-02 15:04 KST")
	}
	return a.Label + " " + a.At.Format("15:04 KST")
}
