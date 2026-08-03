package deployguard_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/deployguard"
)

var deployObservedAt = time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
var deployStartAt = deployObservedAt.Add(time.Second)

func TestActionWindowAndCanonicalObservationAreSealed(t *testing.T) {
	preimage := validPreimage()
	plan, err := deployguard.Freeze(preimage)
	if err != nil {
		t.Fatal(err)
	}
	execution, action, err := deployguard.Start(plan, deployStartAt)
	if err != nil {
		t.Fatal(err)
	}
	if action.IssuedAt != deployStartAt || action.Deadline != deployStartAt.Add(action.Timeout) ||
		action.Deadline.Sub(action.IssuedAt) > deployguard.MaxServiceTimeout {
		t.Fatalf("unsealed action window: %+v", action)
	}

	valid := healthyResult(action, preimage, 0)
	valid.Health.ObservedAt = action.IssuedAt.Add(time.Second)
	sealResult(t, action, &valid)
	for _, test := range []struct {
		name       string
		receivedAt time.Time
		mutate     func(*deployguard.Result)
	}{
		{"at issue", action.IssuedAt, func(r *deployguard.Result) { r.Health.ObservedAt = action.IssuedAt }},
		{"before issue", action.IssuedAt.Add(time.Second), func(r *deployguard.Result) { r.Health.ObservedAt = action.IssuedAt.Add(-time.Nanosecond) }},
		{"after deadline", action.Deadline, func(r *deployguard.Result) { r.Health.ObservedAt = action.Deadline.Add(time.Nanosecond) }},
		{"received after deadline without timeout", action.Deadline.Add(time.Nanosecond), func(*deployguard.Result) {}},
		{"tampered schema after seal", action.IssuedAt.Add(time.Second), func(r *deployguard.Result) { r.SchemaVersion++ }},
		{"replayed digest from another action", action.IssuedAt.Add(time.Second), func(r *deployguard.Result) {
			other := action
			other.ID = string(digest("other-action"))
			r.Health.EvidenceDigest, _ = deployguard.ObservationDigest(other, *r)
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			result := valid
			test.mutate(&result)
			if test.name != "tampered schema after seal" && test.name != "replayed digest from another action" {
				sealResult(t, action, &result)
			}
			got, next, err := deployguard.Advance(execution, result, test.receivedAt)
			if err == nil || next != nil || got.Status() != deployguard.StatusRunning {
				t.Fatalf("invalid bounded observation advanced: status=%s next=%+v err=%v", got.Status(), next, err)
			}
		})
	}
}

func TestAppliedReplacementTimeoutStartsReverseRollback(t *testing.T) {
	preimage := validPreimage()
	plan, _ := deployguard.Freeze(preimage)
	execution, first, _ := deployguard.Start(plan, deployStartAt)
	firstResult := healthyResult(first, preimage, 0)
	execution, second, _ := deployguard.Advance(execution, firstResult, firstResult.Health.ObservedAt)
	timedOut := failedReplaceResult(*second, preimage, 1, deployguard.ReplaceApplied)
	timedOut.TimedOut = true
	timedOut.Health.ObservedAt = second.Deadline
	sealResult(t, *second, &timedOut)
	execution, inspect, err := deployguard.Advance(execution, timedOut, second.Deadline.Add(time.Nanosecond))
	if err != nil || inspect == nil {
		t.Fatalf("applied timeout did not enter rollback: action=%+v err=%v", inspect, err)
	}
	assertAction(t, *inspect, deployguard.ActionReadRollbackCompatibility, "tossos", preimage.Services[1].TargetImageDigest)

	compatTimeout := compatibilityResult(*inspect, preimage, 1, 22)
	compatTimeout.TimedOut = true
	compatTimeout.Health.Healthy = false
	compatTimeout.Health.ObservedAt = inspect.Deadline
	sealResult(t, *inspect, &compatTimeout)
	_, next, err := deployguard.Advance(execution, compatTimeout, inspect.Deadline.Add(time.Nanosecond))
	if err != nil || next == nil || next.Kind != deployguard.ActionReadRollbackCompatibility || next.Service != "httpapi" {
		t.Fatalf("compatibility timeout emitted destructive rollback or lost prior subset: next=%+v err=%v", next, err)
	}
}

func TestUnhealthyCompatibilityReadNeverEmitsRollback(t *testing.T) {
	preimage := validPreimage()
	plan, _ := deployguard.Freeze(preimage)
	execution, first, _ := deployguard.Start(plan, deployStartAt)
	failure := failedReplaceResult(first, preimage, 0, deployguard.ReplaceApplied)
	execution, inspect, _ := deployguard.Advance(execution, failure, failure.Health.ObservedAt)
	unhealthy := compatibilityResult(*inspect, preimage, 0, 22)
	unhealthy.Health.Healthy = false
	sealResult(t, *inspect, &unhealthy)
	execution, next, err := deployguard.Advance(execution, unhealthy, unhealthy.Health.ObservedAt)
	if err != nil || next != nil || execution.Status() != deployguard.StatusRecoveryRequired {
		t.Fatalf("unhealthy schema read authorized rollback: status=%s next=%+v err=%v", execution.Status(), next, err)
	}
	for _, recovery := range execution.Recoveries() {
		if recovery.Code == deployguard.RecoveryRollbackFailed {
			t.Fatalf("rollback was attempted after unhealthy compatibility read: %+v", execution.Recoveries())
		}
	}
}

func TestMissingCompatibilityEvidenceIsUnknownAndNeverEmitsRollback(t *testing.T) {
	preimage := validPreimage()
	plan, _ := deployguard.Freeze(preimage)
	execution, first, _ := deployguard.Start(plan, deployStartAt)
	failure := failedReplaceResult(first, preimage, 0, deployguard.ReplaceApplied)
	execution, inspect, _ := deployguard.Advance(execution, failure, failure.Health.ObservedAt)
	missing := compatibilityResult(*inspect, preimage, 0, 22)
	missing.State.Markets = nil
	sealResult(t, *inspect, &missing)
	execution, next, err := deployguard.Advance(execution, missing, missing.Health.ObservedAt)
	if err != nil || next != nil || execution.Status() != deployguard.StatusRecoveryRequired {
		t.Fatalf("missing schema evidence authorized rollback: status=%s next=%+v err=%v", execution.Status(), next, err)
	}
	recoveries := execution.Recoveries()
	last := recoveries[len(recoveries)-1]
	if last.Code != deployguard.RecoveryRollbackReadFailed || last.EntryEffective != deployguard.StateUnknown {
		t.Fatalf("missing compatibility recovery=%+v", recoveries)
	}
}

func TestRollbackTimeoutRequiresTypedRecovery(t *testing.T) {
	preimage := validPreimage()
	plan, _ := deployguard.Freeze(preimage)
	execution, replace, _ := deployguard.Start(plan, deployStartAt)
	failure := failedReplaceResult(replace, preimage, 0, deployguard.ReplaceApplied)
	execution, inspect, _ := deployguard.Advance(execution, failure, failure.Health.ObservedAt)
	compatible := compatibilityResult(*inspect, preimage, 0, 22)
	execution, rollback, _ := deployguard.Advance(execution, compatible, compatible.Health.ObservedAt)
	timedOut := healthyResult(*rollback, preimage, 0)
	timedOut.TimedOut = true
	timedOut.Health.Healthy = false
	timedOut.Health.ObservedAt = rollback.Deadline
	sealResult(t, *rollback, &timedOut)
	execution, next, err := deployguard.Advance(execution, timedOut, rollback.Deadline.Add(time.Nanosecond))
	if err != nil || next != nil || execution.Status() != deployguard.StatusRecoveryRequired {
		t.Fatalf("rollback timeout status=%s next=%+v err=%v", execution.Status(), next, err)
	}
	recoveries := execution.Recoveries()
	if len(recoveries) != 2 || recoveries[1].Code != deployguard.RecoveryRollbackFailed ||
		recoveries[1].RetainedImageDigest != preimage.Services[0].TargetImageDigest ||
		recoveries[1].EntryEffective != deployguard.StateOff {
		t.Fatalf("rollback timeout recovery=%+v", recoveries)
	}
}

func TestRecoveryNeverInventsOffForDriftedOrMissingEntryEvidence(t *testing.T) {
	for _, test := range []struct {
		name string
		want deployguard.State
		edit func(*deployguard.Result)
	}{
		{"mixed market entry is unknown", deployguard.StateUnknown, func(result *deployguard.Result) {
			kr := result.State.Markets[deployguard.MarketKR]
			kr.EntryEffective = deployguard.StateOn
			result.State.Markets[deployguard.MarketKR] = kr
		}},
		{"complete matching on is on", deployguard.StateOn, func(result *deployguard.Result) {
			for _, market := range []deployguard.Market{deployguard.MarketKR, deployguard.MarketUS} {
				state := result.State.Markets[market]
				state.EntryEffective = deployguard.StateOn
				result.State.Markets[market] = state
			}
		}},
		{"missing market evidence is unknown", deployguard.StateUnknown, func(result *deployguard.Result) {
			delete(result.State.Markets, deployguard.MarketUS)
		}},
		{"missing environment evidence is unknown", deployguard.StateUnknown, func(result *deployguard.Result) {
			result.EnvironmentKeys = nil
		}},
	} {
		t.Run(test.name, func(t *testing.T) {
			preimage := validPreimage()
			plan, _ := deployguard.Freeze(preimage)
			execution, action, _ := deployguard.Start(plan, deployStartAt)
			result := healthyResult(action, preimage, 0)
			test.edit(&result)
			sealResult(t, action, &result)
			execution, next, err := deployguard.Advance(execution, result, result.Health.ObservedAt)
			if err != nil || next == nil || next.Kind != deployguard.ActionReadRollbackCompatibility {
				t.Fatalf("drift did not enter recovery: next=%+v err=%v", next, err)
			}
			recoveries := execution.Recoveries()
			if len(recoveries) != 1 || recoveries[0].Code != deployguard.RecoveryStateDrift ||
				recoveries[0].EntryEffective != test.want {
				t.Fatalf("recovery=%+v want entry=%s", recoveries, test.want)
			}
		})
	}
}

func TestPreservationDriftTakesPriorityOverUnhealthyReplacement(t *testing.T) {
	preimage := validPreimage()
	plan, _ := deployguard.Freeze(preimage)
	execution, action, _ := deployguard.Start(plan, deployStartAt)
	result := failedReplaceResult(action, preimage, 0, deployguard.ReplaceApplied)
	kr := result.State.Markets[deployguard.MarketKR]
	kr.EntryEffective = deployguard.StateOn
	result.State.Markets[deployguard.MarketKR] = kr
	sealResult(t, action, &result)
	execution, next, err := deployguard.Advance(execution, result, result.Health.ObservedAt)
	if err != nil || next == nil {
		t.Fatalf("combined failure did not start recovery: next=%+v err=%v", next, err)
	}
	recoveries := execution.Recoveries()
	if len(recoveries) != 1 || recoveries[0].Code != deployguard.RecoveryStateDrift ||
		recoveries[0].EntryEffective != deployguard.StateUnknown {
		t.Fatalf("combined failure lost state drift truth: %+v", recoveries)
	}
}

func TestFreezeRejectsIncompleteOrMutablePreimageBeforeFirstReplace(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*deployguard.Preimage)
	}{
		{"mutable current image", func(p *deployguard.Preimage) { p.Services[0].CurrentImageDigest = "tossos:local" }},
		{"mutable target image", func(p *deployguard.Preimage) { p.Services[0].TargetImageDigest = "tossos:v2" }},
		{"missing rendered compose digest", func(p *deployguard.Preimage) { p.State.RenderedComposeDigest = "" }},
		{"missing config digest", func(p *deployguard.Preimage) { p.State.ConfigDigest = "" }},
		{"missing activation digest", func(p *deployguard.Preimage) { p.State.ActivationDigest = "" }},
		{"missing protection digest", func(p *deployguard.Preimage) { p.State.ProtectionDigest = "" }},
		{"environment key set not canonical", func(p *deployguard.Preimage) {
			p.Services[0].EnvironmentKeys = []string{"Z_KEY", "A_KEY"}
		}},
		{"mount identity missing", func(p *deployguard.Preimage) { p.Services[0].Mounts[0].IdentityDigest = "" }},
		{"target cannot read current schema", func(p *deployguard.Preimage) {
			p.Services[0].TargetSchema.Readable = deployguard.VersionRange{Min: 23, Max: 24}
		}},
		{"target cannot write post schema", func(p *deployguard.Preimage) {
			p.Services[0].TargetSchema.Writable = deployguard.VersionRange{Min: 21, Max: 21}
		}},
		{"rollback cannot read post schema", func(p *deployguard.Preimage) {
			p.Services[0].RollbackSchema.Readable = deployguard.VersionRange{Min: 1, Max: 21}
		}},
		{"rollback cannot write post schema", func(p *deployguard.Preimage) {
			p.Services[0].RollbackSchema.Writable = deployguard.VersionRange{Min: 1, Max: 21}
		}},
		{"baseline unhealthy", func(p *deployguard.Preimage) { p.Services[0].BaselineHealth.Healthy = false }},
		{"baseline stale", func(p *deployguard.Preimage) {
			p.Services[0].BaselineHealth.ObservedAt = p.CapturedAt.Add(-deployguard.MaxBaselineHealthAge - time.Nanosecond)
		}},
		{"zero timeout", func(p *deployguard.Preimage) { p.Services[0].Timeout = 0 }},
		{"timeout over five minutes", func(p *deployguard.Preimage) { p.Services[0].Timeout = 5*time.Minute + time.Nanosecond }},
		{"service omitted", func(p *deployguard.Preimage) { p.Services = p.Services[:1] }},
		{"service order changed", func(p *deployguard.Preimage) { p.Services[0], p.Services[1] = p.Services[1], p.Services[0] }},
		{"duplicate service", func(p *deployguard.Preimage) { p.Services[1].Name = p.Services[0].Name }},
		{"unknown service", func(p *deployguard.Preimage) { p.Services[1].Name = "unknown" }},
		{"autostart not dormant", func(p *deployguard.Preimage) { p.State.Autostart = deployguard.StateOn }},
		{"automation not dormant", func(p *deployguard.Preimage) { p.State.Automation = deployguard.StateOn }},
		{"live approval not dormant", func(p *deployguard.Preimage) { p.State.LiveApproval = deployguard.StateApproved }},
		{"US lane not dormant", func(p *deployguard.Preimage) {
			state := p.State.Markets[deployguard.MarketUS]
			state.LaneDesired = deployguard.StateOn
			p.State.Markets[deployguard.MarketUS] = state
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			preimage := validPreimage()
			test.mutate(&preimage)
			plan, err := deployguard.Freeze(preimage)
			if err == nil {
				t.Fatalf("invalid preimage produced plan digest %q", plan.Digest())
			}
			if _, _, startErr := start(plan); startErr == nil {
				t.Fatal("invalid preflight emitted a first replacement")
			}
		})
	}
}

func TestRenderedComposeTargetImagesMustBeExactImmutableReferences(t *testing.T) {
	preimage := validPreimage()
	plan, err := deployguard.Freeze(preimage)
	if err != nil {
		t.Fatal(err)
	}
	valid := map[string]string{
		"httpapi": "registry.example/tossos@" + string(preimage.Services[0].TargetImageDigest),
		"tossos":  "registry.example/tossos@" + string(preimage.Services[1].TargetImageDigest),
	}
	if err := deployguard.ValidateRenderedTargetImages(plan, valid); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name   string
		mutate func(map[string]string)
	}{
		{"mutable tag", func(images map[string]string) { images["httpapi"] = "tossos:local" }},
		{"wrong digest", func(images map[string]string) {
			images["httpapi"] = "registry.example/tossos@" + string(digest("wrong"))
		}},
		{"missing service", func(images map[string]string) { delete(images, "tossos") }},
		{"unknown service", func(images map[string]string) { images["other"] = images["tossos"]; delete(images, "tossos") }},
	} {
		t.Run(test.name, func(t *testing.T) {
			images := map[string]string{"httpapi": valid["httpapi"], "tossos": valid["tossos"]}
			test.mutate(images)
			if err := deployguard.ValidateRenderedTargetImages(plan, images); err == nil {
				t.Fatal("invalid rendered Compose image set accepted")
			}
		})
	}
}

func TestReplacementIsFrozenSequentialAndBounded(t *testing.T) {
	preimage := validPreimage()
	plan, err := deployguard.Freeze(preimage)
	if err != nil {
		t.Fatal(err)
	}
	execution, action, err := start(plan)
	if err != nil {
		t.Fatal(err)
	}
	assertAction(t, action, deployguard.ActionReplaceAndVerify, "httpapi", preimage.Services[0].TargetImageDigest)
	if action.Timeout <= 0 || action.Timeout > 5*time.Minute {
		t.Fatalf("first timeout=%s", action.Timeout)
	}

	before := execution
	wrong := healthyResult(action, preimage, 0)
	wrong.ActionID = "wrong-action"
	if _, _, err := advance(execution, wrong); err == nil {
		t.Fatal("out-of-order result advanced the plan")
	}
	if before.Status() != execution.Status() {
		t.Fatal("rejected result mutated execution")
	}

	execution, next, err := advance(execution, healthyResult(action, preimage, 0))
	if err != nil {
		t.Fatal(err)
	}
	if next == nil {
		t.Fatal("second service was not scheduled after first health success")
	}
	assertAction(t, *next, deployguard.ActionReplaceAndVerify, "tossos", preimage.Services[1].TargetImageDigest)
	if next.Timeout <= 0 || next.Timeout > 5*time.Minute {
		t.Fatalf("second timeout=%s", next.Timeout)
	}
	execution, next, err = advance(execution, healthyResult(*next, preimage, 1))
	if err != nil || next != nil || execution.Status() != deployguard.StatusSucceeded {
		t.Fatalf("final execution status=%s next=%+v err=%v", execution.Status(), next, err)
	}
}

func TestSecondServiceFailureRollsBackOnlySuccessfulSubsetInReverse(t *testing.T) {
	preimage := validPreimage()
	plan, err := deployguard.Freeze(preimage)
	if err != nil {
		t.Fatal(err)
	}
	execution, first, err := start(plan)
	if err != nil {
		t.Fatal(err)
	}
	execution, second, err := advance(execution, healthyResult(first, preimage, 0))
	if err != nil || second == nil {
		t.Fatal("second replacement not reached")
	}
	execution, inspect, err := advance(execution, failedReplaceResult(*second, preimage, 1, deployguard.ReplaceApplied))
	if err != nil || inspect == nil {
		t.Fatalf("failure did not start bounded rollback: action=%+v err=%v", inspect, err)
	}
	assertAction(t, *inspect, deployguard.ActionReadRollbackCompatibility, "tossos", preimage.Services[1].TargetImageDigest)
	if inspect.Timeout <= 0 || inspect.Timeout > 5*time.Minute {
		t.Fatalf("read-only compatibility action is not bounded: %s", inspect.Timeout)
	}
	execution, rollback, err := advance(execution, compatibilityResult(*inspect, preimage, 1, 22))
	if err != nil || rollback == nil {
		t.Fatalf("compatible rollback was not planned: action=%+v err=%v", rollback, err)
	}
	assertAction(t, *rollback, deployguard.ActionRollbackAndVerify, "tossos", preimage.Services[1].CurrentImageDigest)
	execution, inspect, err = advance(execution, healthyResult(*rollback, preimage, 1))
	if err != nil || inspect == nil {
		t.Fatalf("reverse rollback did not continue to first service: action=%+v err=%v", inspect, err)
	}
	assertAction(t, *inspect, deployguard.ActionReadRollbackCompatibility, "httpapi", preimage.Services[0].TargetImageDigest)
	execution, rollback, err = advance(execution, compatibilityResult(*inspect, preimage, 0, 22))
	if err != nil || rollback == nil {
		t.Fatalf("first service rollback was not planned: action=%+v err=%v", rollback, err)
	}
	assertAction(t, *rollback, deployguard.ActionRollbackAndVerify, "httpapi", preimage.Services[0].CurrentImageDigest)
	execution, next, err := advance(execution, healthyResult(*rollback, preimage, 0))
	if err != nil || next != nil || execution.Status() != deployguard.StatusRolledBack {
		t.Fatalf("rollback status=%s next=%+v err=%v", execution.Status(), next, err)
	}
}

func TestReplacementFailureBeforeApplyExcludesCurrentAttemptFromRollbackSubset(t *testing.T) {
	preimage := validPreimage()
	plan, err := deployguard.Freeze(preimage)
	if err != nil {
		t.Fatal(err)
	}
	execution, first, _ := start(plan)
	execution, second, _ := advance(execution, healthyResult(first, preimage, 0))
	execution, inspect, err := advance(execution, failedReplaceResult(*second, preimage, 1, deployguard.ReplaceNotApplied))
	if err != nil || inspect == nil {
		t.Fatalf("pre-apply failure did not begin prior-subset rollback: action=%+v err=%v", inspect, err)
	}
	assertAction(t, *inspect, deployguard.ActionReadRollbackCompatibility, "httpapi", preimage.Services[0].TargetImageDigest)
}

func TestRuntimeRollbackIncompatibilityKeepsNewServiceEntryOff(t *testing.T) {
	preimage := validPreimage()
	plan, err := deployguard.Freeze(preimage)
	if err != nil {
		t.Fatal(err)
	}
	execution, first, _ := start(plan)
	execution, second, _ := advance(execution, healthyResult(first, preimage, 0))
	execution, inspect, _ := advance(execution, failedReplaceResult(*second, preimage, 1, deployguard.ReplaceApplied))
	execution, next, err := advance(execution, compatibilityResult(*inspect, preimage, 1, 99))
	if err != nil || next == nil {
		t.Fatalf("incompatible current service did not continue prior-subset recovery: action=%+v err=%v", next, err)
	}
	assertAction(t, *next, deployguard.ActionReadRollbackCompatibility, "httpapi", preimage.Services[0].TargetImageDigest)
	execution, rollback, err := advance(execution, compatibilityResult(*next, preimage, 0, 22))
	if err != nil || rollback == nil {
		t.Fatalf("compatible prior service did not receive rollback action: action=%+v err=%v", rollback, err)
	}
	assertAction(t, *rollback, deployguard.ActionRollbackAndVerify, "httpapi", preimage.Services[0].CurrentImageDigest)
	execution, next, err = advance(execution, healthyResult(*rollback, preimage, 0))
	if err != nil || next != nil {
		t.Fatalf("prior rollback did not finish: action=%+v err=%v", next, err)
	}
	if execution.Status() != deployguard.StatusRecoveryRequired {
		t.Fatalf("status=%s", execution.Status())
	}
	recoveries := execution.Recoveries()
	if len(recoveries) != 2 {
		t.Fatalf("recoveries=%+v", recoveries)
	}
	var incompatible *deployguard.Recovery
	for index := range recoveries {
		if recoveries[index].Code == deployguard.RecoveryRollbackIncompatible {
			incompatible = &recoveries[index]
		}
	}
	if incompatible == nil || incompatible.Service != "tossos" || incompatible.EntryEffective != deployguard.StateOff ||
		incompatible.RetainedImageDigest != preimage.Services[1].TargetImageDigest {
		t.Fatalf("incompatible recovery=%+v", recoveries)
	}
	for _, recovery := range recoveries {
		if recovery.EntryEffective != deployguard.StateOff {
			t.Fatalf("recovery raised entry: %+v", recovery)
		}
	}
}

func TestHealthDriftCannotBecomeSuccessfulOrSchedulePeer(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*deployguard.Result)
	}{
		{"rendered compose", func(r *deployguard.Result) { r.State.RenderedComposeDigest = digest("compose-drift") }},
		{"config", func(r *deployguard.Result) { r.State.ConfigDigest = digest("config-drift") }},
		{"activation", func(r *deployguard.Result) { r.State.ActivationDigest = digest("activation-drift") }},
		{"lane", func(r *deployguard.Result) { r.State.LaneDigest = digest("lane-drift") }},
		{"autostart", func(r *deployguard.Result) { r.State.Autostart = deployguard.StateOn }},
		{"automation", func(r *deployguard.Result) { r.State.Automation = deployguard.StateOn }},
		{"live approval", func(r *deployguard.Result) { r.State.LiveApproval = deployguard.StateApproved }},
		{"protection", func(r *deployguard.Result) { r.State.ProtectionDigest = digest("protection-drift") }},
		{"journal", func(r *deployguard.Result) { r.State.JournalDigest = digest("journal-drift") }},
		{"market entry", func(r *deployguard.Result) {
			state := r.State.Markets[deployguard.MarketKR]
			state.EntryEffective = deployguard.StateOn
			r.State.Markets[deployguard.MarketKR] = state
		}},
		{"environment keys", func(r *deployguard.Result) { r.EnvironmentKeys[0] = "TAMPERED" }},
		{"mount identity", func(r *deployguard.Result) { r.Mounts[0].IdentityDigest = digest("mount-drift") }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			preimage := validPreimage()
			plan, err := deployguard.Freeze(preimage)
			if err != nil {
				t.Fatal(err)
			}
			execution, action, _ := start(plan)
			result := healthyResult(action, preimage, 0)
			test.mutate(&result)
			sealResult(t, action, &result)
			execution, inspect, err := advance(execution, result)
			if err != nil || inspect == nil {
				t.Fatalf("applied drift was abandoned instead of recovered: action=%+v err=%v", inspect, err)
			}
			assertAction(t, *inspect, deployguard.ActionReadRollbackCompatibility, "httpapi", preimage.Services[0].TargetImageDigest)
			execution, rollback, err := advance(execution, compatibilityResult(*inspect, preimage, 0, 22))
			if err != nil || rollback == nil {
				t.Fatalf("drift rollback compatibility was not honored: action=%+v err=%v", rollback, err)
			}
			execution, next, err := advance(execution, healthyResult(*rollback, preimage, 0))
			if err != nil || next != nil || execution.Status() != deployguard.StatusRolledBack {
				t.Fatalf("drift rollback status=%s action=%+v err=%v", execution.Status(), next, err)
			}
			recoveries := execution.Recoveries()
			wantEntry := deployguard.StateOff
			if test.name == "market entry" {
				wantEntry = deployguard.StateUnknown
			}
			if len(recoveries) != 1 || recoveries[0].Code != deployguard.RecoveryStateDrift || recoveries[0].EntryEffective != wantEntry {
				t.Fatalf("drift recoveries=%+v", recoveries)
			}
		})
	}
}

func TestObservedServiceImageAndHealthEvidenceMustMatchSealedAction(t *testing.T) {
	preimage := validPreimage()
	plan, err := deployguard.Freeze(preimage)
	if err != nil {
		t.Fatal(err)
	}
	execution, action, _ := start(plan)
	tests := []struct {
		name   string
		mutate func(*deployguard.Result)
	}{
		{"wrong service", func(r *deployguard.Result) { r.Service = "tossos" }},
		{"wrong image", func(r *deployguard.Result) { r.ImageDigest = preimage.Services[0].CurrentImageDigest }},
		{"missing health digest", func(r *deployguard.Result) { r.Health.EvidenceDigest = "" }},
		{"missing health as-of", func(r *deployguard.Result) { r.Health.ObservedAt = time.Time{} }},
		{"pre-preimage health", func(r *deployguard.Result) { r.Health.ObservedAt = preimage.CapturedAt.Add(-time.Nanosecond) }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := healthyResult(action, preimage, 0)
			test.mutate(&result)
			got, next, err := advance(execution, result)
			if err == nil || next != nil || got.Status() != deployguard.StatusRunning {
				t.Fatalf("mismatched observation advanced: status=%s next=%+v err=%v", got.Status(), next, err)
			}
		})
	}
	mutatedAction := action
	mutatedAction.Service = "tossos"
	mutatedAction.ImageDigest = preimage.Services[1].TargetImageDigest
	if _, _, err := advance(execution, healthyResult(mutatedAction, preimage, 0)); err == nil {
		t.Fatal("caller-mutated action was accepted")
	}
	if _, next, err := advance(execution, healthyResult(action, preimage, 0)); err != nil || next == nil {
		t.Fatalf("rejected observations consumed the sealed action: next=%+v err=%v", next, err)
	}
}

func TestFrozenPlanDeepCopiesPreimageAndActionsCarryNoOperatingMutation(t *testing.T) {
	preimage := validPreimage()
	wantImage := preimage.Services[0].TargetImageDigest
	plan, err := deployguard.Freeze(preimage)
	if err != nil {
		t.Fatal(err)
	}
	preimage.Services[0].TargetImageDigest = digest("tampered")
	preimage.Services[0].EnvironmentKeys[0] = "TAMPERED"
	state := preimage.State.Markets[deployguard.MarketKR]
	state.EntryEffective = deployguard.StateOn
	preimage.State.Markets[deployguard.MarketKR] = state
	_, action, err := start(plan)
	if err != nil {
		t.Fatal(err)
	}
	if action.ImageDigest != wantImage {
		t.Fatalf("frozen target changed to %q", action.ImageDigest)
	}
	wantFields := []string{"ID", "Kind", "Service", "ImageDigest", "Timeout", "IssuedAt", "Deadline"}
	actionType := reflect.TypeOf(action)
	gotFields := make([]string, 0, actionType.NumField())
	for i := 0; i < actionType.NumField(); i++ {
		gotFields = append(gotFields, actionType.Field(i).Name)
	}
	if !reflect.DeepEqual(gotFields, wantFields) {
		t.Fatalf("action gained operating authority fields=%v want=%v", gotFields, wantFields)
	}
	for _, method := range []string{"Apply", "Execute", "Activate", "Order", "Protect", "WriteJournal", "SetToggle"} {
		if _, ok := reflect.TypeOf(plan).MethodByName(method); ok {
			t.Fatalf("plan exposes mutation method %s", method)
		}
	}
}

func validPreimage() deployguard.Preimage {
	state := deployguard.StateEvidence{
		RenderedComposeDigest: digest("rendered-compose"), ConfigDigest: digest("config"),
		ActivationDigest: digest("activation"), LaneDigest: digest("lane"), AutostartDigest: digest("autostart"),
		AutomationDigest: digest("automation"), LiveApprovalDigest: digest("live-approval"),
		ProtectionDigest: digest("protection"), JournalDigest: digest("journal"),
		Autostart: deployguard.StateOff, Automation: deployguard.StateOff, LiveApproval: deployguard.StateUnapproved,
		Markets: map[deployguard.Market]deployguard.MarketState{
			deployguard.MarketKR: dormantMarket(),
			deployguard.MarketUS: dormantMarket(),
		},
	}
	return deployguard.Preimage{SchemaVersion: deployguard.PreimageSchemaVersion, CapturedAt: deployObservedAt, State: state,
		Services: []deployguard.ServicePreimage{
			service("httpapi", 2*time.Minute, []string{"TOSSOS_API_PUBLIC_URL", "TOSSOS_DATA_DIR"}),
			service("tossos", 5*time.Minute, []string{"TOSSOS_DATA_DIR"}),
		}}
}

func service(name string, timeout time.Duration, environment []string) deployguard.ServicePreimage {
	return deployguard.ServicePreimage{Name: name, CurrentImageDigest: digest(name + "-current"),
		TargetImageDigest: digest(name + "-target"), Timeout: timeout,
		EnvironmentKeys: environment,
		Mounts: []deployguard.MountIdentity{
			{Type: deployguard.MountBind, Source: "/srv/tossos/config", Target: "/var/lib/tossos/config", Mode: deployguard.MountReadWrite, IdentityDigest: digest(name + "-config-mount")},
			{Type: deployguard.MountBind, Source: "/srv/tossos/data", Target: "/var/lib/tossos/data", Mode: deployguard.MountReadWrite, IdentityDigest: digest(name + "-data-mount")},
		},
		CurrentSchemaVersion: 22, PostReplaceSchemaVersion: 22,
		TargetSchema: deployguard.SchemaCompatibility{
			Readable: deployguard.VersionRange{Min: 1, Max: 22}, Writable: deployguard.VersionRange{Min: 22, Max: 22}},
		RollbackSchema: deployguard.SchemaCompatibility{
			Readable: deployguard.VersionRange{Min: 1, Max: 22}, Writable: deployguard.VersionRange{Min: 22, Max: 22}},
		BaselineHealth: deployguard.HealthEvidence{Healthy: true, ObservedAt: deployObservedAt, EvidenceDigest: digest(name + "-health")},
	}
}

func dormantMarket() deployguard.MarketState {
	return deployguard.MarketState{LaneDesired: deployguard.StateOff, LaneEffective: deployguard.StateOff,
		ActivationDesired: deployguard.StateOff, ActivationEffective: deployguard.StateOff,
		EntryEffective: deployguard.StateOff, Refusal: deployguard.RefusalNotConfigured}
}

func healthyResult(action deployguard.Action, preimage deployguard.Preimage, serviceIndex int) deployguard.Result {
	service := preimage.Services[serviceIndex]
	outcome := deployguard.ReplaceOutcome("")
	if action.Kind == deployguard.ActionReplaceAndVerify {
		outcome = deployguard.ReplaceApplied
	}
	result := deployguard.Result{ActionID: action.ID, Service: action.Service, ImageDigest: action.ImageDigest,
		ReplaceOutcome: outcome, Health: health(action, true), SchemaVersion: service.PostReplaceSchemaVersion,
		State: cloneStateEvidence(preimage.State), EnvironmentKeys: append([]string(nil), service.EnvironmentKeys...),
		Mounts: append([]deployguard.MountIdentity(nil), service.Mounts...)}
	mustSealResult(action, &result)
	return result
}

func failedReplaceResult(action deployguard.Action, preimage deployguard.Preimage, serviceIndex int, outcome deployguard.ReplaceOutcome) deployguard.Result {
	service := preimage.Services[serviceIndex]
	image := service.CurrentImageDigest
	if outcome == deployguard.ReplaceApplied {
		image = service.TargetImageDigest
	}
	result := deployguard.Result{ActionID: action.ID, Service: service.Name, ImageDigest: image, ReplaceOutcome: outcome,
		Health: health(action, false), SchemaVersion: service.PostReplaceSchemaVersion,
		State: cloneStateEvidence(preimage.State), EnvironmentKeys: append([]string(nil), service.EnvironmentKeys...),
		Mounts: append([]deployguard.MountIdentity(nil), service.Mounts...)}
	mustSealResult(action, &result)
	return result
}

func compatibilityResult(action deployguard.Action, preimage deployguard.Preimage, serviceIndex int, schema uint64) deployguard.Result {
	service := preimage.Services[serviceIndex]
	result := deployguard.Result{ActionID: action.ID, Service: service.Name, ImageDigest: action.ImageDigest,
		Health: health(action, true), SchemaVersion: schema, State: cloneStateEvidence(preimage.State),
		EnvironmentKeys: append([]string(nil), service.EnvironmentKeys...),
		Mounts:          append([]deployguard.MountIdentity(nil), service.Mounts...)}
	mustSealResult(action, &result)
	return result
}

func health(action deployguard.Action, healthy bool) deployguard.HealthEvidence {
	return deployguard.HealthEvidence{Healthy: healthy, ObservedAt: action.IssuedAt.Add(time.Second)}
}

func start(plan deployguard.Plan) (deployguard.Execution, deployguard.Action, error) {
	return deployguard.Start(plan, deployStartAt)
}

func advance(execution deployguard.Execution, result deployguard.Result) (deployguard.Execution, *deployguard.Action, error) {
	return deployguard.Advance(execution, result, result.Health.ObservedAt)
}

func sealResult(t *testing.T, action deployguard.Action, result *deployguard.Result) {
	t.Helper()
	digest, err := deployguard.ObservationDigest(action, *result)
	if err != nil {
		t.Fatal(err)
	}
	result.Health.EvidenceDigest = digest
}

func mustSealResult(action deployguard.Action, result *deployguard.Result) {
	digest, err := deployguard.ObservationDigest(action, *result)
	if err != nil {
		panic(err)
	}
	result.Health.EvidenceDigest = digest
}

func cloneStateEvidence(input deployguard.StateEvidence) deployguard.StateEvidence {
	out := input
	out.Markets = make(map[deployguard.Market]deployguard.MarketState, len(input.Markets))
	for market, state := range input.Markets {
		out.Markets[market] = state
	}
	return out
}

func assertAction(t *testing.T, action deployguard.Action, kind deployguard.ActionKind, service string, image deployguard.Digest) {
	t.Helper()
	if action.Kind != kind || action.Service != service || action.ImageDigest != image || strings.TrimSpace(action.ID) == "" {
		t.Fatalf("action=%+v want kind=%s service=%s image=%s", action, kind, service, image)
	}
}

func digest(value string) deployguard.Digest { return deployguard.DigestBytes([]byte(value)) }
