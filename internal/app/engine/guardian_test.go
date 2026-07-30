package engine_test

// guardian_test.go is task 4.2: the real Guardian, injected into the engine
// profile, measured against the startup interlock.
//
// The stub Guardian in interlock_test.go can pass clause 4 by construction — it
// is handed the limits it declares. That proves the *check* works and nothing
// about the thing being checked. These tests inject execgw.RiskGuardian, whose
// limit snapshot is derived from the risk.Policy it sizes against, and ask
// whether the two independent paths from "the operator's numbers" to "a limit
// snapshot" — the config through the interlock, and the policy through the
// Guardian — actually arrive at the same value.

import (
	"context"
	"errors"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/attest"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/riskcalc"
)

// smallLiveGate is the conservative default set risk-management specifies for an
// operator who has not chosen numbers (정책 수치의 provenance: 주문당 notional
// 1,000,000 KRW·주문당 수량 100주·총 노출 10,000,000 KRW·일일 손실 100,000 KRW·
// 일일 손실 자본비 1%·통화 KRW).
//
// It is written out here rather than derived from risk.DefaultPolicy() on
// purpose: this is the *config* half of the single-source claim, and deriving it
// from the policy would make the comparison a tautology.
func smallLiveGate() config.AutomationGate {
	return config.AutomationGate{
		Enabled:            true,
		MaxOrderQuantity:   100,
		MaxOrderNotional:   1_000_000,
		MaxTotalExposure:   10_000_000,
		MaxDailyLossAmount: 100_000,
		MaxDailyLossRatio:  0.01,
		LimitCurrency:      "KRW",
	}
}

// realGuardian builds the Guardian the engine would be injected with, from a
// policy the caller supplies. Its journal is its own: the engine opens one in
// the config directory, and two handles on one SQLite file is not a wiring
// anybody ships.
func realGuardian(t *testing.T, policy risk.Policy) *execgw.RiskGuardian {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), journal.DBFileName),
		Clock:    clock.NewFake(interlockNow),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("journal.Open: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })

	g, err := execgw.NewRiskGuardian(execgw.RiskGuardianOptions{
		Journal:       j,
		Clock:         clock.NewFake(interlockNow),
		AccountRef:    "123-45",
		Policy:        policy,
		Costs:         costs.DefaultModel(),
		PolicyVersion: "add-core-domain/4.2",
	})
	if err != nil {
		t.Fatalf("NewRiskGuardian: %v", err)
	}
	return g
}

// TestTheRealGuardianClearsEveryClause is the combination test.
//
// The gate carries the small_live set, the attestation is valid and covers every
// endpoint, the policy can exit, the gateway is wired, and the injected Guardian
// is the real one — the numbers an operator actually gets from a preset. Every
// clause has to pass on that exact configuration, and clause 4 is the one this
// task is about: the two paths from the operator's numbers, config → interlock
// and policy → Guardian, have to arrive at the same snapshot.
//
// It used to assert a refusal, because clause 6 stood at the end and nothing an
// operator configured could get past it. Reaching that refusal was the proof that
// clauses 1-5 passed. Since interlock-gates-entry-not-exit the same proof is the
// start succeeding, which is a stronger statement: there is no last clause left
// to hide behind.
func TestTheRealGuardianClearsEveryClause(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	eng, err := openGateEngine(t, dir, srv, realGuardian(t, risk.DefaultPolicy()))
	// Clause 4 by name. A mismatch here would mean the two paths from the
	// operator's numbers disagree, and then the audit trail describes a gate that
	// does not exist.
	if errors.Is(err, engine.ErrGuardianLimitsMismatch) {
		t.Errorf("the real Guardian's limits do not match the audited configuration: %v", err)
	}
	for _, other := range []error{
		engine.ErrGuardianRequired,
		engine.ErrLimitsRequired,
		engine.ErrTradingPolicyRefused,
		engine.ErrGatewayRequired,
		engine.ErrKeylessTransport,
		attest.ErrExpired,
		attest.ErrAccountMismatch,
		attest.ErrEndpointNotAttested,
	} {
		if errors.Is(err, other) {
			t.Errorf("the small_live default set was refused by %v", other)
		}
	}
	if err != nil {
		t.Fatalf("the small_live default set must produce a verified gate: %v", err)
	}
	if eng == nil {
		t.Fatal("a verified start must return an engine")
	}
	if eng.Automation.EntryPermitted {
		t.Error("EntryPermitted = true: this build still leaves no protective order at the broker")
	}
}

// TestTheRealGuardianIsPublishedWhenTheGateVerifies goes one step further than
// the refusal above: with clause 6 satisfied through the test-only seam, the
// engine starts, publishes the Guardian, and the snapshot it declares is
// field-for-field the one the interlock audited — Set bits included.
//
// This is the assertion clause 4 is written for, stated positively. Reaching a
// clause-6 refusal proves clause 4 did not fail; this proves what it compared.
func TestTheRealGuardianIsPublishedWhenTheGateVerifies(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	guardian := realGuardian(t, risk.DefaultPolicy())
	eng, err := openProtectedGateEngine(t, dir, srv, guardian)
	if err != nil {
		t.Fatalf("every clause is satisfied and the gate must come up: %v", err)
	}
	if !eng.Automation.Verified {
		t.Fatal("Verified = false after a successful interlock")
	}
	if eng.Guardian == nil {
		t.Fatal("the injected Guardian must be published once verified")
	}
	limiter, ok := eng.Guardian.(engine.ExposureLimiter)
	if !ok {
		t.Fatalf("the published Guardian (%T) does not report its exposure limits", eng.Guardian)
	}
	if !reflect.DeepEqual(limiter.ExposureLimits(), eng.Automation.Limits) {
		t.Errorf("Guardian limits %+v, audited limits %+v — one source, or the audit is fiction",
			limiter.ExposureLimits(), eng.Automation.Limits)
	}
	// And the audited limits are the small_live numbers, so the comparison above
	// is not two copies of the same mistake.
	want := execgw.Limits{
		MaxQuantity:        execgw.Bound(100),
		MaxNotional:        execgw.Bound(1_000_000),
		MaxTotalExposure:   execgw.Bound(10_000_000),
		MaxDailyLossAmount: execgw.Bound(100_000),
		MaxDailyLossRatio:  execgw.Bound(0.01),
		Currency:           "KRW",
	}
	if !reflect.DeepEqual(eng.Automation.Limits, want) {
		t.Errorf("audited limits = %+v, want the small_live set %+v", eng.Automation.Limits, want)
	}
}

// TestAGuardianSizedAgainstOtherNumbersIsRefused: the equality check is not a
// formality that any RiskGuardian satisfies. A Guardian whose policy relaxes one
// ceiling — here the daily-loss ratio, from 1% to 2% — is refused before the
// protection clause is even reached.
func TestAGuardianSizedAgainstOtherNumbersIsRefused(t *testing.T) {
	dir := isolate(t)
	writeGateConfig(t, dir, smallLiveGate())
	writeCredentials(t, dir, "test-api-key-000000", "test-secret")
	writeAttestation(t, dir, nil)
	srv, _ := interlockServer(t, "123-45")

	relaxed := risk.DefaultPolicy()
	relaxed.MaxDailyLossRatio = "0.02"

	_, err := openProtectedGateEngine(t, dir, srv, realGuardian(t, relaxed))
	if err == nil {
		t.Fatal("a Guardian authorising against unaudited numbers must not start the gate")
	}
	if !errors.Is(err, engine.ErrGuardianLimitsMismatch) {
		t.Fatalf("err = %v, want the single-source clause", err)
	}
	if !strings.Contains(err.Error(), "0.02") {
		t.Errorf("message %q does not show the value that disagreed", err)
	}
}

// TestTheMoneyCeilingsAlsoHaveToAgree walks the remaining bounds one at a time,
// so a mapping that happened to be right for the ratio and wrong for a money
// amount cannot hide behind the case above.
func TestTheMoneyCeilingsAlsoHaveToAgree(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*risk.Policy)
	}{
		{"order quantity", func(p *risk.Policy) { p.MaxOrderQuantity = "50" }},
		{"order notional", func(p *risk.Policy) {
			p.MaxOrderNotional = riskcalc.Money{Amount: "900000", Currency: "KRW"}
		}},
		{"total exposure", func(p *risk.Policy) {
			p.MaxOpenExposure = riskcalc.Money{Amount: "9000000", Currency: "KRW"}
		}},
		{"daily loss", func(p *risk.Policy) {
			p.MaxDailyLoss = riskcalc.Money{Amount: "90000", Currency: "KRW"}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := isolate(t)
			writeGateConfig(t, dir, smallLiveGate())
			writeCredentials(t, dir, "test-api-key-000000", "test-secret")
			writeAttestation(t, dir, nil)
			srv, _ := interlockServer(t, "123-45")

			policy := risk.DefaultPolicy()
			tc.mutate(&policy)

			_, err := openProtectedGateEngine(t, dir, srv, realGuardian(t, policy))
			if !errors.Is(err, engine.ErrGuardianLimitsMismatch) {
				t.Fatalf("err = %v, want the single-source clause for a changed %s", err, tc.name)
			}
		})
	}
}
