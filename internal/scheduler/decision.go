package scheduler

import (
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

type DecisionState string

const (
	DecisionEntryAllowed   DecisionState = "ENTRY_ALLOWED"
	DecisionWaitMarket     DecisionState = "WAIT_MARKET"
	DecisionDisabled       DecisionState = "DISABLED"
	DecisionBudgetDeferred DecisionState = "BUDGET_DEFERRED"
)

type DecisionReason string

const (
	ReasonSchedulerOff        DecisionReason = "SCHEDULER_OFF"
	ReasonNotActivated        DecisionReason = "NOT_ACTIVATED"
	ReasonMarketNotSelected   DecisionReason = "MARKET_NOT_SELECTED"
	ReasonCalendarMissing     DecisionReason = "CALENDAR_MISSING"
	ReasonCalendarInvalid     DecisionReason = "CALENDAR_INVALID"
	ReasonCalendarDayMismatch DecisionReason = "CALENDAR_DAY_MISMATCH"
	ReasonMarketClosed        DecisionReason = "MARKET_CLOSED"
	ReasonHoliday             DecisionReason = "HOLIDAY"
	ReasonBudgetUnavailable   DecisionReason = "BUDGET_UNAVAILABLE"
	ReasonEntryPermitted      DecisionReason = "ENTRY_PERMITTED"
)

type DecisionInput struct {
	Now        time.Time
	Desired    DesiredState
	Activation *Activation
	Market     marketclock.Market
	Calendar   *CalendarSnapshot
	Budget     *BudgetCoordinator
	BudgetKey  string
}

type Decision struct {
	State          DecisionState
	Reason         DecisionReason
	NextTransition time.Time
	Budget         BudgetGrant
}

func Decide(in DecisionInput) Decision {
	if !in.Desired.Enabled {
		return Decision{State: DecisionDisabled, Reason: ReasonSchedulerOff}
	}
	if in.Activation == nil || !in.Activation.matches(in.Desired, in.Market) {
		return Decision{State: DecisionDisabled, Reason: ReasonNotActivated}
	}
	if !in.Desired.Market.allows(in.Market) {
		return Decision{State: DecisionDisabled, Reason: ReasonMarketNotSelected}
	}
	if in.Calendar == nil {
		return Decision{State: DecisionWaitMarket, Reason: ReasonCalendarMissing}
	}
	if in.Calendar.Version == "" || in.Calendar.Version != in.Desired.CalendarVersion ||
		in.Activation.binding.CalendarVersion != in.Calendar.Version {
		return Decision{State: DecisionDisabled, Reason: ReasonNotActivated}
	}
	if in.Calendar.Market != in.Market || in.Calendar.ValidityAt(in.Now) != CalendarValid {
		return Decision{State: DecisionWaitMarket, Reason: ReasonCalendarInvalid}
	}
	loc, err := in.Market.Location()
	if err != nil || in.Calendar.Today.Date != in.Now.In(loc).Format("2006-01-02") {
		return Decision{State: DecisionWaitMarket, Reason: ReasonCalendarDayMismatch, NextTransition: nextSession(in.Calendar)}
	}
	regular := in.Calendar.Today.Regular
	if regular == nil {
		return Decision{State: DecisionWaitMarket, Reason: ReasonHoliday, NextTransition: nextSession(in.Calendar)}
	}
	if in.Now.Before(regular.Open) {
		return Decision{State: DecisionWaitMarket, Reason: ReasonMarketClosed, NextTransition: regular.Open}
	}
	if !in.Now.Before(regular.Close) {
		return Decision{State: DecisionWaitMarket, Reason: ReasonMarketClosed, NextTransition: nextSession(in.Calendar)}
	}
	if in.Budget == nil {
		return Decision{State: DecisionBudgetDeferred, Reason: ReasonBudgetUnavailable}
	}
	grant := in.Budget.TryAcquire(in.BudgetKey, PollEntry, in.Now)
	if !grant.Allowed {
		return Decision{State: DecisionBudgetDeferred, Reason: ReasonBudgetUnavailable, Budget: grant, NextTransition: grant.Reset}
	}
	return Decision{State: DecisionEntryAllowed, Reason: ReasonEntryPermitted, NextTransition: regular.Close, Budget: grant}
}

func (a *Activation) matches(desired DesiredState, market marketclock.Market) bool {
	if a == nil {
		return false
	}
	binding := a.binding
	return binding.SchedulerVersion == desired.Version && binding.DesiredRevision == desired.Revision &&
		binding.CalendarVersion == desired.CalendarVersion &&
		binding.Market == desired.Market && binding.Session == desired.Session && binding.ConfigVersion == desired.ConfigVersion &&
		binding.Actor == desired.Actor && binding.ApprovedAt.Equal(desired.ApprovedAt) && desired.Market.allows(market)
}

func nextSession(calendar *CalendarSnapshot) time.Time {
	if calendar != nil && calendar.NextBusinessDay.Regular != nil {
		return calendar.NextBusinessDay.Regular.Open
	}
	return time.Time{}
}

func (m MarketScope) allows(market marketclock.Market) bool {
	switch m {
	case MarketScopeKR:
		return market == marketclock.MarketKR
	case MarketScopeUS:
		return market == marketclock.MarketUS
	default:
		return false
	}
}

type MarketChoice struct {
	Value MarketScope
	Label string
}

type SessionChoice struct {
	Value SessionScope
	Label string
}

type ScheduleDescriptor struct {
	Category       string
	Section        string
	Default        DesiredState
	MarketChoices  []MarketChoice
	SessionChoices []SessionChoice
}

func MarketScheduleDescriptor() ScheduleDescriptor {
	return ScheduleDescriptor{
		Category: "strategy-runtime", Section: "시장·일정", Default: DefaultDesiredState(),
		MarketChoices:  []MarketChoice{{MarketScopeNone, "선택 안 함"}, {MarketScopeKR, "한국"}, {MarketScopeUS, "미국"}},
		SessionChoices: []SessionChoice{{SessionRegular, "정규장"}},
	}
}
