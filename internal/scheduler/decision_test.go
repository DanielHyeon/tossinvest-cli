package scheduler

import (
	"context"
	"reflect"
	"strings"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func activeKRCalendar(t *testing.T, fetched time.Time) CalendarSnapshot {
	t.Helper()
	got, err := AdaptOfficialCalendar(marketclock.MarketKR, krCalendar(t, "2026-03-25",
		"2026-03-25T09:00:00+09:00", "2026-03-25T15:30:00+09:00"), fetched)
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func TestActivationIsOpaqueAndBoundToExactDesiredState(t *testing.T) {
	typeOf := reflect.TypeOf(Activation{})
	for i := 0; i < typeOf.NumField(); i++ {
		if typeOf.Field(i).IsExported() {
			t.Fatalf("activation exposes forgeable field %s", typeOf.Field(i).Name)
		}
	}
	open := at(t, "2026-03-25T09:00:00+09:00")
	calendar := activeKRCalendar(t, open.Add(-time.Hour))
	desired := approvedDesired()
	desired.ApprovedAt = open.Add(-2 * time.Hour)
	desired.CalendarVersion = calendar.Version
	activation := activationFor(t, desired, calendar, open)
	desired.ConfigVersion = "changed-after-approval"
	got := Decide(DecisionInput{Now: open, Desired: desired, Activation: activation,
		Market: marketclock.MarketKR, Calendar: &calendar})
	if got.State != DecisionDisabled || got.Reason != ReasonNotActivated {
		t.Fatalf("stale activation accepted: %+v", got)
	}
	desired.ConfigVersion = "config-v7"
	desired.Revision++
	got = Decide(DecisionInput{Now: open, Desired: desired, Activation: activation,
		Market: marketclock.MarketKR, Calendar: &calendar})
	if got.State != DecisionDisabled || got.Reason != ReasonNotActivated {
		t.Fatalf("activation from an older desired revision accepted: %+v", got)
	}
}

func TestDecisionRejectsCalendarVersionChangedAfterActivation(t *testing.T) {
	open := at(t, "2026-03-25T09:00:00+09:00")
	calendar := activeKRCalendar(t, open.Add(-time.Hour))
	desired := approvedDesired()
	desired.ApprovedAt = open.Add(-2 * time.Hour)
	desired.CalendarVersion = calendar.Version
	activation := activationFor(t, desired, calendar, open)

	// The newly fetched calendar remains structurally fresh, but it is not the
	// exact evidence the operator/manifest approved.
	calendar.Version = "sha256:" + strings.Repeat("b", 64)
	got := Decide(DecisionInput{Now: open, Desired: desired, Activation: activation,
		Market: marketclock.MarketKR, Calendar: &calendar})
	if got.State != DecisionDisabled || got.Reason != ReasonNotActivated {
		t.Fatalf("changed calendar version accepted: %+v", got)
	}
}

func activationFor(t *testing.T, desired DesiredState, calendar CalendarSnapshot, now time.Time) *Activation {
	t.Helper()
	desired.CalendarVersion = calendar.Version
	current := CurrentBinding{SchedulerVersion: desired.Version, CalendarVersion: desired.CalendarVersion,
		Market: desired.Market, Session: desired.Session, ConfigVersion: desired.ConfigVersion, BuildDigest: "build-v1"}
	claim := testManifestClaim{binding: desired.ActivationBinding(current.BuildDigest), expiresAt: now.Add(time.Hour)}
	result := Restore(context.Background(), desired, current, claim, now)
	if !result.Restored {
		t.Fatalf("activation: %+v", result)
	}
	return result.Activation
}

func TestDecisionStatesAreTypedAndOrderedFailClosed(t *testing.T) {
	open := at(t, "2026-03-25T09:00:00+09:00")
	calendar := activeKRCalendar(t, open.Add(-time.Hour))
	desired := approvedDesired()
	desired.ApprovedAt = open.Add(-2 * time.Hour)
	desired.CalendarVersion = calendar.Version
	activation := activationFor(t, desired, calendar, open)
	b := NewBudgetCoordinator()
	b.Observe(budget(open, 100))

	cases := []struct {
		name string
		in   DecisionInput
		want DecisionState
	}{
		{"disabled", DecisionInput{Now: open, Desired: DefaultDesiredState(), Market: marketclock.MarketKR}, DecisionDisabled},
		{"not activated", DecisionInput{Now: open, Desired: desired, Market: marketclock.MarketKR, Calendar: &calendar, Budget: b, BudgetKey: "/api/v1/rankings"}, DecisionDisabled},
		{"calendar missing", DecisionInput{Now: open, Desired: desired, Activation: activation, Market: marketclock.MarketKR, Budget: b, BudgetKey: "/api/v1/rankings"}, DecisionWaitMarket},
		{"outside market", DecisionInput{Now: open.Add(-time.Minute), Desired: desired, Activation: activation, Market: marketclock.MarketKR, Calendar: &calendar, Budget: b, BudgetKey: "/api/v1/rankings"}, DecisionWaitMarket},
		{"budget missing", DecisionInput{Now: open, Desired: desired, Activation: activation, Market: marketclock.MarketKR, Calendar: &calendar, BudgetKey: "/api/v1/rankings"}, DecisionBudgetDeferred},
		{"allowed", DecisionInput{Now: open, Desired: desired, Activation: activation, Market: marketclock.MarketKR, Calendar: &calendar, Budget: b, BudgetKey: "/api/v1/rankings"}, DecisionEntryAllowed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := Decide(tc.in)
			if got.State != tc.want || got.Reason == "" {
				t.Fatalf("Decide = %+v, want %q with reason", got, tc.want)
			}
		})
	}
}

func TestHolidayWaitsWithNextSession(t *testing.T) {
	fetched := at(t, "2026-03-25T08:00:00+09:00")
	payload := krCalendar(t, "2026-03-25", "2026-03-25T09:00:00+09:00", "2026-03-25T15:30:00+09:00")
	payload.Today.Integrated = nil
	calendar, err := AdaptOfficialCalendar(marketclock.MarketKR, payload, fetched)
	if err != nil {
		t.Fatal(err)
	}
	desired := approvedDesired()
	desired.ApprovedAt = fetched.Add(-time.Hour)
	desired.CalendarVersion = calendar.Version
	activation := activationFor(t, desired, calendar, fetched)
	b := NewBudgetCoordinator()
	b.Observe(budget(fetched, 100))
	got := Decide(DecisionInput{Now: at(t, "2026-03-25T12:00:00+09:00"), Desired: desired,
		Activation: activation, Market: marketclock.MarketKR, Calendar: &calendar,
		Budget: b, BudgetKey: "/api/v1/rankings"})
	if got.State != DecisionWaitMarket || got.Reason != ReasonHoliday || got.NextTransition.IsZero() {
		t.Fatalf("holiday decision = %+v", got)
	}
}

func TestDescriptorFreezesServerOwnedDefaultsAndChoices(t *testing.T) {
	d := MarketScheduleDescriptor()
	if d.Category != "strategy-runtime" || d.Section != "시장·일정" {
		t.Fatalf("descriptor owner = %+v", d)
	}
	if d.Default != DefaultDesiredState() {
		t.Fatalf("descriptor default = %+v", d.Default)
	}
	if len(d.MarketChoices) != 3 || len(d.SessionChoices) != 1 {
		t.Fatalf("server choices = markets:%v sessions:%v", d.MarketChoices, d.SessionChoices)
	}
	for _, choice := range d.MarketChoices {
		if choice.Value == MarketScope("KR+US") {
			t.Fatal("combined scope was advertised without per-market calendar bindings")
		}
	}
}
