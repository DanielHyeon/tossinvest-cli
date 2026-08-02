package console

import (
	"context"
	"encoding/hex"
	"net/http"
	"strings"
	"time"
)

// MarketScheduleReading contains only values rendered on the read-only screen.
type MarketScheduleReading struct {
	SchedulerDesired   bool
	AutoStartDesired   bool
	SchedulerEffective string
	AutoStartEffective bool
	Market             string
	Session            string
	ApplyTiming        string
	CalendarSource     string
	CalendarVersion    string
	CalendarFetchedAt  time.Time
	DecisionReason     string
	NextTransition     time.Time
}

type MarketScheduleReader interface {
	Read(context.Context) (MarketScheduleReading, error)
}

type marketSchedulePage struct {
	chrome
	Unwired            bool
	LoadErr            bool
	SchedulerDesired   string
	AutoStartDesired   string
	SchedulerEffective string
	AutoStartEffective string
	Market             string
	Session            string
	ApplyTiming        string
	CalendarSource     string
	CalendarVersion    string
	CalendarUpdatedAt  string
	DecisionReason     string
	NextTransition     string
	DecisionHelp       string
}

func (marketSchedulePage) Refresh() bool { return false }

func (c *Console) handleMarketSchedule(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", http.MethodGet+", "+http.MethodHead)
		c.refuse(w, http.StatusMethodNotAllowed, "읽기 전용 화면이다",
			"시장·일정 조회는 GET/HEAD만 허용한다. 아무것도 전송되지 않았다.")
		return
	}
	reading := MarketScheduleReading{
		SchedulerEffective: "DISABLED", Market: "선택 시장 없음", Session: "정규장",
		ApplyTiming: "다음 엔진 기동", CalendarSource: "없음", CalendarVersion: "없음",
		DecisionReason: "NOT_ACTIVATED",
	}
	// "optimization-sub" names no navigation item: see strategy_runtime.go.
	page := marketSchedulePage{chrome: c.chromeOnRequest("optimization-sub"), Unwired: c.opts.MarketSchedule == nil}
	if c.opts.MarketSchedule != nil {
		value, err := c.opts.MarketSchedule.Read(r.Context())
		if err != nil {
			page.LoadErr = true
		} else {
			reading = value
		}
	}
	page.SchedulerDesired = onOff(reading.SchedulerDesired)
	page.AutoStartDesired = onOff(reading.AutoStartDesired)
	page.SchedulerEffective = effectiveState(reading.SchedulerEffective)
	page.AutoStartEffective = onOff(reading.AutoStartEffective)
	page.Market = scheduleMarketLabel(reading.Market)
	page.Session = sessionLabel(reading.Session)
	page.ApplyTiming = applyTiming(reading.ApplyTiming)
	page.CalendarSource, page.CalendarVersion, page.CalendarUpdatedAt = calendarProvenance(reading)
	page.DecisionReason = typedDecisionReason(reading.DecisionReason)
	page.NextTransition = formatScheduleTime(reading.NextTransition)
	page.DecisionHelp = decisionHelp(page.SchedulerEffective, page.DecisionReason)
	c.render(w, "market-schedule", page)
}

func onOff(value bool) string {
	if value {
		return "ON"
	}
	return "OFF"
}

func effectiveState(value string) string {
	switch value {
	case "ENTRY_ALLOWED", "WAIT_MARKET", "DISABLED", "BUDGET_DEFERRED":
		return value
	default:
		return "DISABLED"
	}
}

func typedDecisionReason(value string) string {
	switch value {
	case "SCHEDULER_OFF", "NOT_ACTIVATED", "MARKET_NOT_SELECTED", "CALENDAR_MISSING",
		"CALENDAR_INVALID", "CALENDAR_DAY_MISMATCH", "MARKET_CLOSED", "HOLIDAY",
		"BUDGET_UNAVAILABLE", "ENTRY_PERMITTED":
		return value
	default:
		return "NOT_ACTIVATED"
	}
}

func applyTiming(value string) string {
	if value == "다음 엔진 기동" {
		return value
	}
	return "다음 엔진 기동"
}

func calendarProvenance(reading MarketScheduleReading) (string, string, string) {
	digest, err := hex.DecodeString(strings.TrimPrefix(reading.CalendarVersion, "sha256:"))
	if reading.CalendarSource != "official-openapi" || !strings.HasPrefix(reading.CalendarVersion, "sha256:") || err != nil || len(digest) != 32 || reading.CalendarFetchedAt.IsZero() {
		return "검증되지 않음", "검증되지 않음", "검증되지 않음"
	}
	return reading.CalendarSource, reading.CalendarVersion, formatScheduleTime(reading.CalendarFetchedAt)
}

func scheduleMarketLabel(value string) string {
	switch value {
	case "KR":
		return "한국"
	case "US":
		return "미국"
	default:
		return "선택 시장 없음"
	}
}

func sessionLabel(value string) string {
	if value == "regular" || value == "정규장" {
		return "정규장"
	}
	return "정규장"
}

func formatScheduleTime(value time.Time) string {
	if value.IsZero() {
		return "없음"
	}
	return value.UTC().Format(time.RFC3339)
}

func decisionHelp(state, reason string) string {
	switch state {
	case "BUDGET_DEFERRED":
		return "사용자 OFF가 아니라 API 예산 대기다. safety budget은 침범하지 않는다."
	case "WAIT_MARKET":
		return "시장·calendar 조건을 기다린다. stale calendar라면 official calendar를 장 시작 전에 갱신해야 한다."
	case "ENTRY_ALLOWED":
		return "시장·calendar·budget 조건이 충족됐다. 이 표시는 주문 승인을 만들지 않는다."
	default:
		return "scheduler가 꺼져 있거나 활성화 manifest가 없다: " + reason
	}
}
