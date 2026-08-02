package console

// settings_operating_test.go covers the screen that ended the hand-edits
// (change console-owns-the-operating-toggles).
//
// Three properties, and the third is the one worth the file:
//
//	the toggles exist        an operator can reach every setting the engine needs
//	                         without opening config.json
//	the seams stay disjoint  read off the types, the way settings_limits_test.go
//	                         reads its own
//	the screen is honest     it renders what it cannot judge, refuses to promise a
//	                         start, and says what turning the gate on begins —
//	                         without asking anybody to type anything

import (
	"net/url"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
)

type fakeTrading struct {
	mu      sync.Mutex
	block   config.Trading
	loadErr error
	saveErr error
	saved   []config.TradingPolicy
}

func (f *fakeTrading) Load() (config.Trading, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.block, f.loadErr
}

func (f *fakeTrading) Save(p config.TradingPolicy) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, p)
	return nil
}

type fakeGate struct {
	mu      sync.Mutex
	saveErr error
	saved   []bool
}

func (f *fakeGate) Save(on bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, on)
	return nil
}

func operatingHarness(t *testing.T, gate config.AutomationGate, trading config.Trading) (
	*dashboardHarness, *fakeTrading, *fakeGate,
) {
	t.Helper()
	tr := &fakeTrading{block: trading}
	g := &fakeGate{}
	h := newDashboardHarness(t, func(o *Options) {
		o.Settings = &fakeSettings{}
		o.Limits = &fakeLimits{gate: gate}
		o.TradingPolicy = tr
		o.Gate = g
	})
	return h, tr, g
}

func fullLimits() config.AutomationGate {
	return config.AutomationGate{
		LimitCurrency: "USD", MaxOrderQuantity: 100, MaxOrderNotional: 500,
		MaxTotalExposure: 1_500, MaxDailyLossAmount: 50, MaxDailyLossRatio: 0.01,
	}
}

func fullPolicy() config.Trading {
	return config.Trading{Place: true, Sell: true, Cancel: true, AllowLiveOrderActions: true}
}

// --- the seams stay disjoint -------------------------------------------------

// TestTheGateSeamCarriesOnlyTheSwitch and TestTheTradingSeamCarriesOnlyItsFour
// are the type-level halves of the disjointness internal/config proves at the
// byte level. Both are read with reflection for the reason settings_limits_test
// gives: a property nobody measures is a property that drifts.
func TestTheGateSeamCarriesOnlyTheSwitch(t *testing.T) {
	rt := reflect.TypeOf((*GateSwitch)(nil)).Elem()
	if rt.NumMethod() != 1 {
		t.Fatalf("GateSwitch has %d methods; it saves one key and reads nothing", rt.NumMethod())
	}
	save, _ := rt.MethodByName("Save")
	if save.Type.NumIn() != 1 || save.Type.In(0).Kind() != reflect.Bool {
		t.Errorf("Save takes %v; a switch that could carry a limit is not a switch", save.Type)
	}
}

func TestTheTradingSeamCarriesOnlyItsFour(t *testing.T) {
	rt := reflect.TypeOf((*TradingPolicySettings)(nil)).Elem()
	save, ok := rt.MethodByName("Save")
	if !ok {
		t.Fatal("TradingPolicySettings has no Save")
	}
	arg := save.Type.In(0)
	if arg != reflect.TypeOf(config.TradingPolicy{}) {
		t.Fatalf("Save takes %s; it must take config.TradingPolicy", arg)
	}
	for i := range arg.NumField() {
		switch strings.ToLower(arg.Field(i).Name) {
		case "amend", "conditional", "fractional":
			t.Errorf("config.TradingPolicy carries %s; the screen does not offer it and the type "+
				"must not be able to move it", arg.Field(i).Name)
		}
	}
	if arg.NumField() != 4 {
		t.Errorf("config.TradingPolicy has %d fields, want exactly the four the exit path uses",
			arg.NumField())
	}
}

// --- the toggles exist -------------------------------------------------------

func TestTheScreenRendersTheFourTogglesAndTheirState(t *testing.T) {
	h, _, _ := operatingHarness(t, fullLimits(), config.Trading{Place: true, Sell: true})
	h.authenticate(t)

	page := body(t, h.get(t, pathSettingsDaily))
	for _, key := range []string{"place", "sell", "cancel", "allow_live_order_actions"} {
		if !strings.Contains(page, `name="`+key+`"`) {
			t.Errorf("the screen has no %s toggle:\n%s", key, page)
		}
	}
	// The three the engine never reaches are absent, not disabled: a toggle on
	// screen is a question, and this one has no useful answer.
	for _, key := range []string{`name="amend"`, `name="conditional"`, `name="fractional"`} {
		if strings.Contains(page, key) {
			t.Errorf("the screen offers %s, which no loop in this build uses", key)
		}
	}
	if !strings.Contains(page, `action="/settings/trading"`) {
		t.Error("the policy form has no action")
	}
	// The gate switch is on 상시 now — irreversible, approved by a person — and the
	// four toggles are on 당일. That they are on different tabs is the change; that
	// both still exist and still post to the same two routes is what this asserts.
	if standing := body(t, h.get(t, pathSettingsStanding)); !strings.Contains(standing, `action="/settings/gate"`) {
		t.Error("the gate form has no action")
	}
}

func TestSavingThePolicyPassesExactlyWhatWasTicked(t *testing.T) {
	h, tr, _ := operatingHarness(t, fullLimits(), config.Trading{})
	h.authenticate(t)

	h.post(t, "/settings/trading", url.Values{
		"csrf": {h.csrf}, "place": {"on"}, "sell": {"on"}, "cancel": {"on"},
		"allow_live_order_actions": {"on"},
	})
	if len(tr.saved) != 1 || tr.saved[0] != (config.TradingPolicy{
		Place: true, Sell: true, Cancel: true, AllowLiveOrderActions: true,
	}) {
		t.Fatalf("saved = %+v", tr.saved)
	}

	// An unticked box is off, not absent: a form that only sends what is checked
	// still has to be able to turn something off.
	h.post(t, "/settings/trading", url.Values{"csrf": {h.csrf}, "sell": {"on"}})
	if got := tr.saved[1]; got.Place || got.Cancel || got.AllowLiveOrderActions || !got.Sell {
		t.Errorf("saved = %+v, want only Sell", got)
	}
}

func TestSavingAPartialPolicySaysTheEngineWillStillRefuse(t *testing.T) {
	h, _, _ := operatingHarness(t, fullLimits(), config.Trading{})
	h.authenticate(t)

	// The client follows the redirect, so the notice arrives rendered.
	page := body(t, h.post(t, "/settings/trading", url.Values{"csrf": {h.csrf}, "sell": {"on"}}))
	for _, want := range []string{"trading.place", "trading.cancel", "인터록 3절이 거부한다"} {
		if !strings.Contains(page, want) {
			t.Errorf("the notice does not name %q:\n%s", want, page)
		}
	}
}

func TestSavingTheGatePassesTheSwitchAndNothingElse(t *testing.T) {
	h, _, g := operatingHarness(t, fullLimits(), fullPolicy())
	h.authenticate(t)

	h.post(t, "/settings/gate", url.Values{"csrf": {h.csrf}, "enabled": {"on"}})
	h.post(t, "/settings/gate", url.Values{"csrf": {h.csrf}})
	if len(g.saved) != 2 || !g.saved[0] || g.saved[1] {
		t.Fatalf("saved = %v, want [true false]", g.saved)
	}
}

// --- the screen is honest ----------------------------------------------------

// TestThePreflightRendersWhatItCannotJudge. Three green ticks and silence about
// the rest would be read as "this will start", which is the one thing this
// screen must not say.
func TestThePreflightRendersWhatItCannotJudge(t *testing.T) {
	h, _, _ := operatingHarness(t, fullLimits(), fullPolicy())
	h.authenticate(t)

	page := body(t, h.get(t, pathSettingsStanding))
	if !strings.Contains(page, "판정 불가") {
		t.Errorf("the pre-flight does not mark the clauses it cannot judge:\n%s", page)
	}
	if !strings.Contains(page, "기동이 보장되지 않는다") {
		t.Errorf("the screen does not disclaim the guarantee:\n%s", page)
	}
}

func TestThePreflightNamesTheMissingToggles(t *testing.T) {
	h, _, _ := operatingHarness(t, fullLimits(), config.Trading{Sell: true})
	h.authenticate(t)

	page := body(t, h.get(t, pathSettingsStanding))
	for _, want := range []string{"trading.place", "trading.cancel", "지금 켜면 기동이 거부된다"} {
		if !strings.Contains(page, want) {
			t.Errorf("the pre-flight does not say %q:\n%s", want, page)
		}
	}
}

func TestThePreflightNamesUnsetLimits(t *testing.T) {
	h, _, _ := operatingHarness(t, config.AutomationGate{}, fullPolicy())
	h.authenticate(t)

	page := body(t, h.get(t, pathSettingsStanding))
	if !strings.Contains(page, "지금 켜면 기동이 거부된다") {
		t.Errorf("an empty limit block does not block the pre-flight:\n%s", page)
	}
}

// TestTheGateSectionSaysWhatStarts is the SHALL: the operator reads what turning
// the switch on begins before they press it — including the part that cannot be
// undone.
func TestTheGateSectionSaysWhatStarts(t *testing.T) {
	h := newDashboardHarness(t, func(o *Options) {
		o.Settings = &fakeSettings{block: config.Adoption{
			Enabled: true, ExcludeSymbols: []string{"TSLA"},
		}}
		o.Limits = &fakeLimits{gate: fullLimits()}
		o.TradingPolicy = &fakeTrading{block: fullPolicy()}
		o.Gate = &fakeGate{}
	})
	h.authenticate(t)

	page := body(t, h.get(t, pathSettingsStanding))
	for _, want := range []string{
		"대사·exit 관측",        // the loops
		"되돌릴 수 없다",          // adoption is irreversible
		"첫 대사 주기에 편입이 일어난다", // and it happens immediately
		"TSLA", // what is exempt from it
		"프로세스가 죽으면 보호도 사라진다", // the protection's lifetime
		"게이트웨이에서 계속 거부된다",    // entry is still refused
	} {
		if !strings.Contains(page, want) {
			t.Errorf("the gate section does not say %q:\n%s", want, page)
		}
	}
}

// TestTheGateSectionAsksForNoTypedConfirmation is the user instruction of
// 2026-07-27, asserted rather than remembered.
func TestTheGateSectionAsksForNoTypedConfirmation(t *testing.T) {
	h, _, _ := operatingHarness(t, fullLimits(), fullPolicy())
	h.authenticate(t)

	page := body(t, h.get(t, pathSettingsStanding))
	gate := page[strings.Index(page, `id="gate"`):]
	for _, forbidden := range []string{`type="text"`, `type="password"`, "<textarea"} {
		if strings.Contains(gate, forbidden) {
			t.Errorf("the operating section has a %s input; approval here is a button, not a "+
				"typing exercise:\n%s", forbidden, gate)
		}
	}
	if n := strings.Count(gate, `action="/settings/gate"`); n != 1 {
		t.Errorf("the gate has %d forms; one press is the whole approval", n)
	}
}

// TestAnUnwiredSeamExplainsRatherThanDisappearing, for the same reason the limit
// section does it: a missing section reads as a missing feature.
func TestAnUnwiredSeamExplainsRatherThanDisappearing(t *testing.T) {
	h := newDashboardHarness(t, func(o *Options) {
		o.Settings = &fakeSettings{}
		o.Limits = &fakeLimits{gate: fullLimits()}
	})
	h.authenticate(t)

	daily := body(t, h.get(t, pathSettingsDaily))
	if !strings.Contains(daily, "거래 정책 저장이 배선되지 않았다") {
		t.Errorf("an unwired policy seam says nothing:\n%s", daily)
	}
	standing := body(t, h.get(t, pathSettingsStanding))
	if !strings.Contains(standing, "게이트 저장이 배선되지 않았다") {
		t.Errorf("an unwired gate seam says nothing:\n%s", standing)
	}
	if strings.Contains(standing, `action="/settings/gate"`) {
		t.Error("the gate form is drawn with no seam behind it")
	}
}
