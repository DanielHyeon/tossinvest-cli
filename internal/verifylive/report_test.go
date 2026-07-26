package verifylive

// report_test.go covers the thing tasks 2.6 and 1.4 read.
//
// The property under test throughout is the same one: an attribute nobody
// measured must come out of the report as *unverified*, loudly, rather than as
// absent. An attestation is a permission slip for unattended trading, and the
// most dangerous thing a report can do is stay quiet about a gap.

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestBuildReportOnAnEmptyRecord(t *testing.T) {
	rep := BuildReport("/nowhere", nil, time.Now())
	if rep.ReplayEnabled {
		t.Fatal("idempotent replay is enabled on a record that proves nothing")
	}
	if len(rep.Unverified) == 0 {
		t.Fatal("an empty record produced no unverified properties")
	}
	var out bytes.Buffer
	rep.WriteText(&out)
	if !strings.Contains(out.String(), "verify run") {
		t.Errorf("the empty report does not point at the command that starts one:\n%s", out.String())
	}
}

// TestEveryChecklistPropertyIsReportedEvenWhenUnmeasured.
func TestEveryChecklistPropertyIsReportedEvenWhenUnmeasured(t *testing.T) {
	rep := BuildReport("/nowhere", nil, time.Now())
	want := 0
	for _, g := range requiredProperties() {
		want += len(g.Attributes)
	}
	got := 0
	for _, g := range rep.Groups {
		got += len(g.Attributes)
		for _, a := range g.Attributes {
			if a.Verified {
				t.Errorf("%s reported itself verified on an empty record", a.Key)
			}
		}
	}
	if got != want {
		t.Errorf("the report carries %d attributes, the checklist has %d", got, want)
	}
	if len(rep.Unverified) != want {
		t.Errorf("Unverified has %d entries, want all %d", len(rep.Unverified), want)
	}
}

// TestReplayStaysDisabledUnlessBothHalvesArePositive is the spec's rule:
// "검증되지 않은 상태에서 멱등 재생을 해소 절차로 사용해서는 안 된다".
func TestReplayStaysDisabledUnlessBothHalvesArePositive(t *testing.T) {
	cases := []struct {
		name    string
		sameID  string
		noExtra string
		want    bool
	}{
		{"both positive", "true", "true", true},
		{"replay returned a different order", "false", "true", false},
		{"a second order appeared", "true", "false", false},
		{"neither observed", "", "", false},
		{"unverified", "unverified", "unverified", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			entries := []Entry{{StepID: StepIdempotency, Observations: []Observation{
				{Key: "idempotency.replay_returns_same_order_id", Value: c.sameID},
				{Key: "idempotency.no_second_order_created", Value: c.noExtra},
			}}}
			if got := BuildReport("/r", entries, time.Now()).ReplayEnabled; got != c.want {
				t.Errorf("ReplayEnabled = %v, want %v", got, c.want)
			}
		})
	}
}

// TestUnverifiedValuesAreNotCountedAsAnswers. A step that says "unverified" has
// still not measured anything, and the report must not treat the string as data.
func TestUnverifiedValuesAreNotCountedAsAnswers(t *testing.T) {
	entries := []Entry{{StepID: StepConditionalTrigger, Observations: []Observation{
		{Key: "conditional.triggered_order_id_exposed", Value: "unverified"},
		{Key: "conditional.trigger_observed", Value: "false"},
	}}}
	rep := BuildReport("/r", entries, time.Now())

	exposed := findAttribute(t, rep, "conditional.triggered_order_id_exposed")
	if exposed.Verified {
		t.Error("an 'unverified' value was counted as a measurement")
	}
	observed := findAttribute(t, rep, "conditional.trigger_observed")
	if !observed.Verified || observed.Value != "false" {
		t.Errorf("a measured 'false' must count as an answer, got %+v", observed)
	}
}

// TestAFullRunProducesAReportWithNoIdempotencyGaps.
func TestAFullRunProducesAReportWithNoIdempotencyGaps(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 1)
	h := newHarness(t, broker, alwaysConfirm())
	runToCompletion(t, h, Options{HoldingSymbol: "005930"})

	rep := BuildReport(h.record, h.entries(), time.Now())
	if !rep.ReplayEnabled {
		t.Errorf("a broker that honoured the key still left replay disabled; unverified: %v", rep.Unverified)
	}
	for _, key := range []string{
		"conditional.survives_process_exit",
		"conditional.modify_issues_new_id",
		"conditional.cancel.gone_after",
		"order.amend.issues_new_id",
		"sell.boundary.over_holding_rejected",
	} {
		if a := findAttribute(t, rep, key); !a.Verified {
			t.Errorf("%s came back unverified after a complete run", key)
		}
	}
	// The trigger observation is the one that must still be outstanding.
	if a := findAttribute(t, rep, "conditional.trigger_observed"); a.Value != "false" {
		t.Errorf("conditional.trigger_observed = %q", a.Value)
	}
	if a := findAttribute(t, rep, "conditional.triggered_order_id_exposed"); a.Verified {
		t.Error("triggeredOrderId exposure cannot be verified without a trigger")
	}
	if len(rep.Outstanding) != 0 {
		t.Errorf("the report says the account still holds %+v", rep.Outstanding)
	}
}

func TestReportTextNamesTheUnverifiedProperties(t *testing.T) {
	broker := newFakeBroker()
	h := newHarness(t, broker, alwaysConfirm())
	if _, err := h.run(Options{}); err != nil {
		t.Fatalf("run: %v", err)
	}
	var out bytes.Buffer
	BuildReport(h.record, h.entries(), time.Now()).WriteText(&out)
	text := out.String()

	for _, want := range []string{
		"ProtectiveCapability", "Idempotency key", "unverified",
		"no-automatic-entry list", "conditional.survives_process_exit",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("the report does not mention %q:\n%s", want, text)
		}
	}
	if strings.Contains(text, "123-45-678901") {
		t.Error("the report printed the unmasked account number")
	}
}

// TestProgressPointsAtTheRestartWhenOneIsPending.
func TestProgressPointsAtTheRestartWhenOneIsPending(t *testing.T) {
	broker := newFakeBroker().withHolding("005930", 1)
	h := newHarness(t, broker, alwaysConfirm())
	if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {
		t.Fatalf("run: %v", err)
	}

	p := BuildProgress(h.record, h.entries())
	if p.AwaitingRestart != StepConditionalPersist {
		t.Fatalf("AwaitingRestart = %q, want the persistence step", p.AwaitingRestart)
	}
	if len(p.Outstanding) != 1 {
		t.Fatalf("Outstanding = %+v, want the live conditional", p.Outstanding)
	}
	var out bytes.Buffer
	p.WriteText(&out)
	text := out.String()
	if !strings.Contains(text, "--resume") {
		t.Errorf("the status does not tell the operator to resume:\n%s", text)
	}
	if !strings.Contains(text, p.Outstanding[0].ID) {
		t.Errorf("the status does not print the live conditional's id:\n%s", text)
	}
}

func TestProgressOnAnUnstartedVerification(t *testing.T) {
	var out bytes.Buffer
	BuildProgress("/nowhere", nil).WriteText(&out)
	if !strings.Contains(out.String(), "verify run --list") {
		t.Errorf("the status does not offer the read-only preview:\n%s", out.String())
	}
}

// TestWriteStepsIsAReadableProcedure — an operator must be able to see exactly
// what will happen before anything touches the account.
func TestWriteStepsIsAReadableProcedure(t *testing.T) {
	var out bytes.Buffer
	WriteSteps(&out, false)
	text := out.String()
	for _, s := range Steps() {
		if !strings.Contains(text, string(s.ID)) {
			t.Errorf("the listing omits %s", s.ID)
		}
	}
	for _, want := range []string{"mutating", "needs a holding", FlagIncludeTTLEdge, "deferred", "one share"} {
		if !strings.Contains(text, want) {
			t.Errorf("the listing does not mention %q", want)
		}
	}
}

func TestChecklistGroupsHaveNoDuplicateKeys(t *testing.T) {
	for _, g := range requiredProperties() {
		keys := sortedAttributeKeys(g)
		for i := 1; i < len(keys); i++ {
			if keys[i] == keys[i-1] {
				t.Errorf("%s lists %s twice", g.Name, keys[i])
			}
		}
	}
}

func findAttribute(t *testing.T, rep Report, key string) Attribute {
	t.Helper()
	for _, g := range rep.Groups {
		for _, a := range g.Attributes {
			if a.Key == key {
				return a
			}
		}
	}
	t.Fatalf("%s is not in the report's checklist", key)
	return Attribute{}
}
