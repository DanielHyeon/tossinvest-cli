package engine

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/exitpolicy"
	"github.com/JungHoonGhae/tossinvest-cli/internal/positionpolicy"
)

func TestPositionPolicyCommandServiceRequiresEngineOwnedJournal(t *testing.T) {
	if _, err := NewPositionPolicyCommandService(nil, clock.System()); err == nil {
		t.Fatal("nil engine context unexpectedly built a command service")
	}
	if _, err := NewPositionPolicyCommandService(&Context{}, clock.System()); err == nil {
		t.Fatal("context without owned journal unexpectedly built a command service")
	}
}

type fakePolicyRepository struct {
	state          positionpolicy.State
	applied        positionpolicy.Request
	previewed      positionpolicy.Request
	positionErr    error
	applyErr       error
	resultMismatch bool
	applies        int
}

func (f *fakePolicyRepository) PositionPolicies(context.Context) ([]positionpolicy.State, error) {
	return []positionpolicy.State{f.state}, nil
}
func (f *fakePolicyRepository) PositionPolicy(context.Context, string) (positionpolicy.State, error) {
	if f.positionErr != nil {
		return positionpolicy.State{}, f.positionErr
	}
	return f.state, nil
}
func (f *fakePolicyRepository) PreviewPositionPolicy(_ context.Context,
	req positionpolicy.Request) (positionpolicy.Preview, error) {
	f.previewed = req
	after := f.state
	after.Version++
	after.UpdatedAt = req.At.UTC().Format(time.RFC3339Nano)
	switch req.Action {
	case positionpolicy.ActionOverride:
		after.DesiredPolicyID = req.PolicyID
	case positionpolicy.ActionRelease:
		after.Status = positionpolicy.StatusReleased
		after.EffectivePolicyID = ""
	case positionpolicy.ActionReadopt:
		after.Status = positionpolicy.StatusManaged
		after.AdoptionGeneration++
		after.Version = 1
		after.EffectivePolicyID = req.ReAdoption.PolicyID
	}
	return positionpolicy.Preview{Before: f.state, After: after, Action: req.Action, Reason: req.Reason}, nil
}
func (f *fakePolicyRepository) ApplyPositionPolicy(_ context.Context,
	req positionpolicy.Request) (positionpolicy.State, error) {
	f.applies++
	f.applied = req
	if f.applyErr != nil {
		return positionpolicy.State{}, f.applyErr
	}
	preview, err := f.PreviewPositionPolicy(context.Background(), req)
	if err != nil {
		return positionpolicy.State{}, err
	}
	f.state = preview.After
	if f.resultMismatch {
		f.state.Version++
	}
	return f.state, nil
}

type fakePolicyPrices struct {
	quotes []domain.Quote
	calls  int
}

func (f *fakePolicyPrices) Prices(context.Context, []string) ([]domain.Quote, error) {
	f.calls++
	return f.quotes, nil
}

func TestPositionPolicyServiceDerivesReadoptObservationInsideEngine(t *testing.T) {
	now := time.Date(2026, 8, 1, 2, 3, 4, 0, time.UTC)
	repository := &fakePolicyRepository{state: positionpolicy.State{
		PositionID: "p-1", Symbol: "005930", ExitEligible: true,
		Provenance:  positionpolicy.ProvenanceExternalAdoption,
		Eligibility: positionpolicy.EligibilityExternalLifecycle,
		Status:      positionpolicy.StatusReleased, AdoptionGeneration: 1, Version: 1,
		DesiredPolicyID: exitpolicy.CommonLadderRunner,
	}}
	prices := &fakePolicyPrices{quotes: []domain.Quote{{Symbol: "005930", Last: 72000}}}
	service := &PositionPolicyCommandService{
		j: repository, prices: prices, adoption: config.Adoption{DefaultStopPct: 0.05},
		commonPolicy: exitpolicy.CommonLadderBalanced, clk: clock.NewFake(now),
	}
	request := positionpolicy.Request{
		PositionID: "p-1", ExpectedGeneration: 1, ExpectedVersion: 1,
		Action: positionpolicy.ActionReadopt, Actor: positionpolicy.ActorLocalOperator,
		Reason: positionpolicy.ReasonReadopt, At: now.Add(-time.Hour),
		ReAdoption: &positionpolicy.ReAdoptionObservation{
			ObservedPrice: "1", SyntheticStop: "0.5", ObservedAt: "attacker",
		},
	}
	preview, err := service.Preview(context.Background(), request)
	if err != nil {
		t.Fatal(err)
	}
	if preview.Capability == "" || preview.After.EffectivePolicyID != exitpolicy.CommonLadderRunner {
		t.Fatalf("readopt preview=%+v", preview)
	}
	got := repository.previewed
	if prices.calls != 1 || got.At != now || got.ReAdoption == nil ||
		got.ReAdoption.ObservedPrice != "72000" || got.ReAdoption.SyntheticStop != "68400" ||
		got.ReAdoption.PolicyID != exitpolicy.CommonLadderRunner || got.ReAdoption.ObservedAt != now.Format(time.RFC3339Nano) {
		t.Fatalf("engine-derived request=%+v price calls=%d", got, prices.calls)
	}
}

func TestPositionPolicyServiceRefusesIneligibleBeforePriceRead(t *testing.T) {
	repository := &fakePolicyRepository{state: positionpolicy.State{PositionID: "p-1", Symbol: "005930"}}
	prices := &fakePolicyPrices{quotes: []domain.Quote{{Symbol: "005930", Last: 72000}}}
	service := &PositionPolicyCommandService{j: repository, prices: prices, clk: clock.System()}
	_, err := service.Preview(context.Background(), positionpolicy.Request{Action: positionpolicy.ActionReadopt})
	if !errors.Is(err, positionpolicy.ErrIneligible) || prices.calls != 0 {
		t.Fatalf("error=%v price calls=%d", err, prices.calls)
	}
}

func TestPositionPolicyCapabilityRequiresPreviewIsTimedAndOneTime(t *testing.T) {
	now := time.Date(2026, 8, 1, 3, 0, 0, 0, time.UTC)
	clk := clock.NewFake(now)
	repository := &fakePolicyRepository{state: positionpolicy.State{
		PositionID: "p-1", AccountRef: "acct-1", Market: "kr", Symbol: "005930",
		AdoptionGeneration: 1, Version: 4, Status: positionpolicy.StatusManaged,
		Provenance:   positionpolicy.ProvenanceExternalAdoption,
		Eligibility:  positionpolicy.EligibilityExternalLifecycle,
		ExitEligible: true, PositionState: "OPEN",
	}}
	service := &PositionPolicyCommandService{j: repository, clk: clk}
	req := positionpolicy.Request{
		PositionID: "p-1", ExpectedGeneration: 1, ExpectedVersion: 4,
		Action: positionpolicy.ActionOverride, PolicyID: exitpolicy.CommonLadderHybrid50,
		Actor: positionpolicy.ActorLocalOperator, Reason: positionpolicy.ReasonPolicyOverride,
	}

	if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
		Capability: "no-preview",
	}); !errors.Is(err, positionpolicy.ErrCapabilityInvalid) {
		t.Fatalf("apply without preview error=%v", err)
	}
	preview, err := service.Preview(context.Background(), req)
	if err != nil || preview.Capability == "" {
		t.Fatalf("preview=%+v err=%v", preview, err)
	}
	got, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
		Capability: preview.Capability,
	})
	if err != nil || got != preview.After || repository.applies != 1 {
		t.Fatalf("apply=%+v err=%v applies=%d", got, err, repository.applies)
	}
	if repository.applied.PositionID != req.PositionID || repository.applied.Action != req.Action ||
		repository.applied.PolicyID != req.PolicyID || repository.applied.Reason != req.Reason {
		t.Fatalf("capability did not preserve exact request: %+v", repository.applied)
	}
	if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
		Capability: preview.Capability,
	}); !errors.Is(err, positionpolicy.ErrCapabilityInvalid) || repository.applies != 1 {
		t.Fatalf("replay error=%v applies=%d", err, repository.applies)
	}
}

func TestPositionPolicyDangerousCapabilityRequiresServerSideConfirmation(t *testing.T) {
	now := time.Date(2026, 8, 1, 3, 0, 0, 0, time.UTC)
	clk := clock.NewFake(now)
	repository := &fakePolicyRepository{state: positionpolicy.State{
		PositionID: "p-1", AdoptionGeneration: 1, Version: 4,
		Status: positionpolicy.StatusManaged, PositionState: "OPEN",
		Provenance:  positionpolicy.ProvenanceExternalAdoption,
		Eligibility: positionpolicy.EligibilityExternalLifecycle, ExitEligible: true,
	}}
	service := &PositionPolicyCommandService{j: repository, clk: clk}
	preview, err := service.Preview(context.Background(), positionpolicy.Request{
		PositionID: "p-1", ExpectedGeneration: 1, ExpectedVersion: 4,
		Action: positionpolicy.ActionRelease, Actor: positionpolicy.ActorLocalOperator,
		Reason: positionpolicy.ReasonRelease,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
		Capability: preview.Capability, Confirmed: true,
	}); !errors.Is(err, positionpolicy.ErrCapabilityTooEarly) || repository.applies != 0 {
		t.Fatalf("early confirmed release error=%v applies=%d", err, repository.applies)
	}
	clk.Advance(positionPolicyCapabilityDelay)
	if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
		Capability: preview.Capability,
	}); !errors.Is(err, positionpolicy.ErrConfirmationRequired) || repository.applies != 0 {
		t.Fatalf("unconfirmed release error=%v applies=%d", err, repository.applies)
	}
	if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
		Capability: preview.Capability, Confirmed: true,
	}); err != nil || repository.applies != 1 {
		t.Fatalf("confirmed release error=%v applies=%d", err, repository.applies)
	}
}

func TestReadoptCapabilityHasShortFreshnessAndRejectsClockRollback(t *testing.T) {
	now := time.Date(2026, 8, 1, 3, 0, 0, 0, time.UTC)
	newService := func() (*PositionPolicyCommandService, *fakePolicyRepository, *clock.Fake) {
		clk := clock.NewFake(now)
		repository := &fakePolicyRepository{state: positionpolicy.State{
			PositionID: "p-1", Symbol: "005930", AdoptionGeneration: 1, Version: 2,
			Status: positionpolicy.StatusReleased, PositionState: "OPEN",
			Provenance:  positionpolicy.ProvenanceExternalAdoption,
			Eligibility: positionpolicy.EligibilityExternalLifecycle, ExitEligible: true,
		}}
		return &PositionPolicyCommandService{
			j: repository, prices: &fakePolicyPrices{quotes: []domain.Quote{{Symbol: "005930", Last: 72000}}},
			adoption:     config.Adoption{DefaultStopPct: 0.05},
			commonPolicy: exitpolicy.CommonLadderBalanced, clk: clk,
		}, repository, clk
	}
	request := positionpolicy.Request{
		PositionID: "p-1", ExpectedGeneration: 1, ExpectedVersion: 2,
		Action: positionpolicy.ActionReadopt, Actor: positionpolicy.ActorLocalOperator,
		Reason: positionpolicy.ReasonReadopt,
	}

	t.Run("exact freshness boundary is accepted", func(t *testing.T) {
		service, repository, clk := newService()
		preview, err := service.Preview(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		clk.Advance(positionPolicyReadoptFreshness)
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
			Capability: preview.Capability, Confirmed: true,
		}); err != nil || repository.applies != 1 {
			t.Fatalf("exact boundary error=%v applies=%d", err, repository.applies)
		}
	})

	t.Run("stale observation is consumed", func(t *testing.T) {
		service, repository, clk := newService()
		preview, err := service.Preview(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		clk.Advance(positionPolicyReadoptFreshness + time.Nanosecond)
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
			Capability: preview.Capability, Confirmed: true,
		}); !errors.Is(err, positionpolicy.ErrCapabilityStale) || repository.applies != 0 {
			t.Fatalf("stale readopt error=%v applies=%d", err, repository.applies)
		}
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
			Capability: preview.Capability, Confirmed: true,
		}); !errors.Is(err, positionpolicy.ErrCapabilityInvalid) {
			t.Fatalf("stale grant was reusable: %v", err)
		}
	})

	t.Run("clock rollback is consumed", func(t *testing.T) {
		service, repository, clk := newService()
		preview, err := service.Preview(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		clk.Set(now.Add(-time.Nanosecond))
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
			Capability: preview.Capability, Confirmed: true,
		}); !errors.Is(err, positionpolicy.ErrCapabilityStale) || repository.applies != 0 {
			t.Fatalf("rollback error=%v applies=%d", err, repository.applies)
		}
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
			Capability: preview.Capability, Confirmed: true,
		}); !errors.Is(err, positionpolicy.ErrCapabilityInvalid) {
			t.Fatalf("rollback grant was reusable: %v", err)
		}
	})
}

func TestPositionPolicyCapabilityRejectsExpiryOtherEngineStaleAndConcurrentReplay(t *testing.T) {
	newService := func(now time.Time) (*PositionPolicyCommandService, *fakePolicyRepository, *clock.Fake) {
		clk := clock.NewFake(now)
		repository := &fakePolicyRepository{state: positionpolicy.State{
			PositionID: "p-1", AdoptionGeneration: 1, Version: 2,
			Status: positionpolicy.StatusManaged, PositionState: "OPEN",
			Provenance:  positionpolicy.ProvenanceExternalAdoption,
			Eligibility: positionpolicy.EligibilityExternalLifecycle, ExitEligible: true,
		}}
		return &PositionPolicyCommandService{j: repository, clk: clk}, repository, clk
	}
	request := positionpolicy.Request{
		PositionID: "p-1", ExpectedGeneration: 1, ExpectedVersion: 2,
		Action: positionpolicy.ActionInherit, Actor: positionpolicy.ActorLocalOperator,
		Reason: positionpolicy.ReasonPolicyInherit,
	}
	now := time.Date(2026, 8, 1, 3, 0, 0, 0, time.UTC)

	t.Run("expired and other engine", func(t *testing.T) {
		issuer, _, clk := newService(now)
		preview, err := issuer.Preview(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		other, otherRepo, otherClock := newService(now)
		otherClock.Advance(3 * time.Second)
		if _, err := other.Apply(context.Background(), positionpolicy.ApplyRequest{Capability: preview.Capability}); !errors.Is(err, positionpolicy.ErrCapabilityInvalid) || otherRepo.applies != 0 {
			t.Fatalf("other engine error=%v applies=%d", err, otherRepo.applies)
		}
		clk.Advance(positionPolicyCapabilityTTL + time.Nanosecond)
		if _, err := issuer.Apply(context.Background(), positionpolicy.ApplyRequest{Capability: preview.Capability}); !errors.Is(err, positionpolicy.ErrCapabilityExpired) {
			t.Fatalf("expired error=%v", err)
		}
	})

	t.Run("stale failure consumes fail closed", func(t *testing.T) {
		service, repository, clk := newService(now)
		preview, err := service.Preview(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		repository.state.Version++
		clk.Advance(3 * time.Second)
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{Capability: preview.Capability}); !errors.Is(err, positionpolicy.ErrVersionMismatch) || repository.applies != 0 {
			t.Fatalf("stale error=%v applies=%d", err, repository.applies)
		}
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{Capability: preview.Capability}); !errors.Is(err, positionpolicy.ErrCapabilityInvalid) {
			t.Fatalf("stale capability was reusable: %v", err)
		}
	})

	t.Run("repository failure consumes fail closed", func(t *testing.T) {
		service, repository, clk := newService(now)
		preview, err := service.Preview(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		repository.applyErr = errors.New("storage failed")
		clk.Advance(positionPolicyCapabilityDelay)
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
			Capability: preview.Capability,
		}); err == nil || repository.applies != 1 {
			t.Fatalf("repository failure error=%v applies=%d", err, repository.applies)
		}
		repository.applyErr = nil
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
			Capability: preview.Capability,
		}); !errors.Is(err, positionpolicy.ErrCapabilityInvalid) || repository.applies != 1 {
			t.Fatalf("failed capability was reusable: error=%v applies=%d", err, repository.applies)
		}
	})

	t.Run("current read failure consumes fail closed", func(t *testing.T) {
		service, repository, clk := newService(now)
		preview, err := service.Preview(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		repository.positionErr = errors.New("read failed")
		clk.Advance(positionPolicyCapabilityDelay)
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
			Capability: preview.Capability,
		}); err == nil || repository.applies != 0 {
			t.Fatalf("current read failure error=%v applies=%d", err, repository.applies)
		}
		repository.positionErr = nil
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
			Capability: preview.Capability,
		}); !errors.Is(err, positionpolicy.ErrCapabilityInvalid) {
			t.Fatalf("read-failed capability was reusable: %v", err)
		}
	})

	t.Run("result mismatch is fail closed", func(t *testing.T) {
		service, repository, clk := newService(now)
		preview, err := service.Preview(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		repository.resultMismatch = true
		clk.Advance(positionPolicyCapabilityDelay)
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
			Capability: preview.Capability,
		}); !errors.Is(err, positionpolicy.ErrCapabilityInvalid) || repository.applies != 1 {
			t.Fatalf("result mismatch error=%v applies=%d", err, repository.applies)
		}
		if _, err := service.Apply(context.Background(), positionpolicy.ApplyRequest{
			Capability: preview.Capability,
		}); !errors.Is(err, positionpolicy.ErrCapabilityInvalid) || repository.applies != 1 {
			t.Fatalf("mismatched capability was reusable: error=%v applies=%d", err, repository.applies)
		}
	})

	t.Run("concurrent duplicate has one winner", func(t *testing.T) {
		service, repository, clk := newService(now)
		preview, err := service.Preview(context.Background(), request)
		if err != nil {
			t.Fatal(err)
		}
		clk.Advance(3 * time.Second)
		errs := make([]error, 2)
		var wg sync.WaitGroup
		for i := range errs {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				_, errs[i] = service.Apply(context.Background(), positionpolicy.ApplyRequest{Capability: preview.Capability})
			}(i)
		}
		wg.Wait()
		wins, rejected := 0, 0
		for _, err := range errs {
			if err == nil {
				wins++
			} else if errors.Is(err, positionpolicy.ErrCapabilityInvalid) {
				rejected++
			}
		}
		if wins != 1 || rejected != 1 || repository.applies != 1 {
			t.Fatalf("wins=%d rejected=%d applies=%d errors=%v", wins, rejected, repository.applies, errs)
		}
	})
}

func TestPositionPolicyServiceRequiresExternalAdoptedProvenanceForLifecycleActions(t *testing.T) {
	for _, action := range []positionpolicy.Action{positionpolicy.ActionRelease, positionpolicy.ActionReadopt} {
		repository := &fakePolicyRepository{state: positionpolicy.State{
			PositionID: "p-1", Symbol: "005930", Status: positionpolicy.StatusManaged,
			AdoptionGeneration: 1, Provenance: positionpolicy.ProvenanceEngineEntry,
			Eligibility: positionpolicy.EligibilityExitOnly, ExitEligible: true,
		}}
		service := &PositionPolicyCommandService{j: repository, clk: clock.System()}
		_, err := service.Preview(context.Background(), positionpolicy.Request{
			PositionID: "p-1", ExpectedGeneration: 1, Action: action,
			Actor: positionpolicy.ActorLocalOperator,
			Reason: map[positionpolicy.Action]positionpolicy.Reason{
				positionpolicy.ActionRelease: positionpolicy.ReasonRelease,
				positionpolicy.ActionReadopt: positionpolicy.ReasonReadopt,
			}[action],
		})
		if !errors.Is(err, positionpolicy.ErrIneligible) {
			t.Errorf("%s error=%v", action, err)
		}
	}
}
