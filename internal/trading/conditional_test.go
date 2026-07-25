package trading

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
)

// Conditional-order gate + intent assembly tests (harden-execution-base 1.4).
//
// The gate used to live in cmd/tossctl, which meant the only caller that could
// enforce it was cobra: an engine placing a conditional order would have had to
// re-implement the same three checks, and issue #111 is the record of what
// happens when a write-policy rule is duplicated per surface instead of living
// with the module that owns it.

// fakeConditionalBroker records calls without performing any I/O.
type fakeConditionalBroker struct {
	created  []orderintent.ConditionalPlaceIntent
	canceled []orderintent.ConditionalCancelIntent
	modified []orderintent.ConditionalModifyIntent
	err      error
}

func (f *fakeConditionalBroker) CreateConditionalOrder(_ context.Context, intent orderintent.ConditionalPlaceIntent) (domain.ConditionalOrderRef, error) {
	f.created = append(f.created, intent)
	if f.err != nil {
		return domain.ConditionalOrderRef{}, f.err
	}
	return domain.ConditionalOrderRef{ID: "CO-1", ClientOrderID: intent.ClientOrderID}, nil
}

func (f *fakeConditionalBroker) CancelConditionalOrder(_ context.Context, intent orderintent.ConditionalCancelIntent) error {
	f.canceled = append(f.canceled, intent)
	return f.err
}

func (f *fakeConditionalBroker) ModifyConditionalOrder(_ context.Context, intent orderintent.ConditionalModifyIntent) error {
	f.modified = append(f.modified, intent)
	return f.err
}

func (f *fakeConditionalBroker) mutations() int {
	return len(f.created) + len(f.canceled) + len(f.modified)
}

func openConditionalPolicy() config.Trading {
	return config.Trading{Conditional: true, AllowLiveOrderActions: true}
}

func cancelIntentFixture() orderintent.ConditionalCancelIntent {
	return orderintent.ConditionalCancelIntent{ID: "co-1"}
}

func placeIntentFixture() orderintent.ConditionalPlaceIntent {
	return orderintent.ConditionalPlaceIntent{
		Symbol: "005930", Type: "SINGLE", OrderType: "LIMIT", ExpireDate: "2026-12-31", Quantity: 10,
		First: orderintent.ConditionLeg{OrderSide: "SELL", TriggerPrice: 70000, OrderPrice: 69900},
	}
}

func modifyIntentFixture() orderintent.ConditionalModifyIntent {
	return orderintent.ConditionalModifyIntent{
		ID: "co-1", Type: "SINGLE", OrderType: "LIMIT", ExpireDate: "2026-12-31", Quantity: 10,
		First: orderintent.ConditionLeg{OrderSide: "SELL", TriggerPrice: 70000, OrderPrice: 69900},
	}
}

// TestConditionalGateTruthTable is the moved cmd/tossctl test, at its new home.
// The four outcomes must stay distinguishable: a config refusal must never be
// mistaken for "show a preview".
func TestConditionalGateTruthTable(t *testing.T) {
	t.Parallel()

	canonical := orderintent.CanonicalConditionalCancel(cancelIntentFixture())
	token := orderintent.ConfirmToken(canonical)
	open := openConditionalPolicy()

	if err := ConditionalGate(config.Trading{}, canonical, true, token); err == nil || errors.Is(err, ErrConditionalPreviewOnly) {
		t.Errorf("config disabled: want refusal error, got %v", err)
	}
	if err := ConditionalGate(config.Trading{Conditional: true}, canonical, true, token); err == nil || errors.Is(err, ErrConditionalPreviewOnly) {
		t.Errorf("conditional on but live actions off: want refusal, got %v", err)
	}
	if err := ConditionalGate(config.Trading{AllowLiveOrderActions: true}, canonical, true, token); err == nil || errors.Is(err, ErrConditionalPreviewOnly) {
		t.Errorf("live actions on but conditional off: want refusal, got %v", err)
	}
	if err := ConditionalGate(open, canonical, false, ""); !errors.Is(err, ErrConditionalPreviewOnly) {
		t.Errorf("no execute: want ErrConditionalPreviewOnly, got %v", err)
	}
	if err := ConditionalGate(open, canonical, true, "wrong-token"); err == nil || errors.Is(err, ErrConditionalPreviewOnly) {
		t.Errorf("wrong confirm: want mismatch error, got %v", err)
	}
	if err := ConditionalGate(open, canonical, true, token); err != nil {
		t.Errorf("valid: want nil, got %v", err)
	}
}

// TestConditionalPreviewsCarryCanonicalAndToken pins that the token a user is
// shown is the token the gate will accept — computed from the same canonical
// string, for all three verbs.
func TestConditionalPreviewsCarryCanonicalAndToken(t *testing.T) {
	t.Parallel()

	svc := NewService(openConditionalPolicy(), nil)

	for _, tc := range []struct {
		name      string
		preview   Preview
		canonical string
		wantKind  string
	}{
		{"place", svc.PreviewConditionalPlace(placeIntentFixture()), orderintent.CanonicalConditionalPlace(placeIntentFixture()), "conditional_place"},
		{"cancel", svc.PreviewConditionalCancel(cancelIntentFixture()), orderintent.CanonicalConditionalCancel(cancelIntentFixture()), "conditional_cancel"},
		{"modify", svc.PreviewConditionalModify(modifyIntentFixture()), orderintent.CanonicalConditionalModify(modifyIntentFixture()), "conditional_modify"},
	} {
		if tc.preview.Kind != tc.wantKind {
			t.Errorf("%s: Kind = %q, want %q", tc.name, tc.preview.Kind, tc.wantKind)
		}
		if tc.preview.Canonical != tc.canonical {
			t.Errorf("%s: Canonical = %q, want %q", tc.name, tc.preview.Canonical, tc.canonical)
		}
		if tc.preview.ConfirmToken != orderintent.ConfirmToken(tc.canonical) {
			t.Errorf("%s: ConfirmToken does not match the canonical form", tc.name)
		}
		if !tc.preview.MutationReady {
			t.Errorf("%s: open policy must report MutationReady", tc.name)
		}
	}
}

func TestConditionalPreviewReportsClosedPolicy(t *testing.T) {
	t.Parallel()

	svc := NewService(config.Trading{}, nil)
	preview := svc.PreviewConditionalPlace(placeIntentFixture())
	if preview.MutationReady {
		t.Error("closed policy must not report MutationReady")
	}
	if !strings.Contains(strings.Join(preview.Warnings, "\n"), "disables") {
		t.Errorf("expected a config-disabled warning, got %v", preview.Warnings)
	}
}

// TestConditionalExecuteRequiresGate is the safety core: no mutation may reach
// the broker unless config, --execute and the confirm token all agree.
func TestConditionalExecuteRequiresGate(t *testing.T) {
	t.Parallel()

	type attempt struct {
		name string
		run  func(*Service, ExecuteOptions) error
	}
	attempts := []attempt{
		{"place", func(s *Service, o ExecuteOptions) error {
			_, err := s.ConditionalPlace(context.Background(), placeIntentFixture(), o)
			return err
		}},
		{"cancel", func(s *Service, o ExecuteOptions) error {
			return s.ConditionalCancel(context.Background(), cancelIntentFixture(), o)
		}},
		{"modify", func(s *Service, o ExecuteOptions) error {
			return s.ConditionalModify(context.Background(), modifyIntentFixture(), o)
		}},
	}

	for _, a := range attempts {
		t.Run(a.name+"/config disabled", func(t *testing.T) {
			broker := &fakeConditionalBroker{}
			svc := NewService(config.Trading{}, nil).WithConditional(broker)
			err := a.run(svc, ExecuteOptions{Execute: true, Confirm: "whatever"})
			if err == nil || errors.Is(err, ErrConditionalPreviewOnly) {
				t.Fatalf("want refusal, got %v", err)
			}
			if broker.mutations() != 0 {
				t.Error("broker must not be reached when config disables conditional orders")
			}
		})

		t.Run(a.name+"/no execute", func(t *testing.T) {
			broker := &fakeConditionalBroker{}
			svc := NewService(openConditionalPolicy(), nil).WithConditional(broker)
			if err := a.run(svc, ExecuteOptions{}); !errors.Is(err, ErrConditionalPreviewOnly) {
				t.Fatalf("want ErrConditionalPreviewOnly, got %v", err)
			}
			if broker.mutations() != 0 {
				t.Error("preview must not mutate")
			}
		})

		t.Run(a.name+"/wrong confirm", func(t *testing.T) {
			broker := &fakeConditionalBroker{}
			svc := NewService(openConditionalPolicy(), nil).WithConditional(broker)
			err := a.run(svc, ExecuteOptions{Execute: true, Confirm: "nope"})
			if err == nil || errors.Is(err, ErrConditionalPreviewOnly) {
				t.Fatalf("want mismatch error, got %v", err)
			}
			if broker.mutations() != 0 {
				t.Error("broker must not be reached on a confirm mismatch")
			}
		})

		t.Run(a.name+"/no broker wired", func(t *testing.T) {
			svc := NewService(openConditionalPolicy(), nil)
			// Correct token, open policy — the only thing missing is the broker.
			err := a.run(svc, ExecuteOptions{Execute: true, Confirm: conditionalTokenFor(t, svc, a.name)})
			if !errors.Is(err, ErrLiveMutationPending) {
				t.Fatalf("want ErrLiveMutationPending, got %v", err)
			}
		})
	}
}

// TestConditionalExecuteReachesBrokerWithCorrectToken is the positive half: the
// token from the preview opens the gate exactly once per verb, and the intent
// arrives at the broker unchanged.
func TestConditionalExecuteReachesBrokerWithCorrectToken(t *testing.T) {
	t.Parallel()

	broker := &fakeConditionalBroker{}
	svc := NewService(openConditionalPolicy(), nil).WithConditional(broker)

	place := placeIntentFixture()
	ref, err := svc.ConditionalPlace(context.Background(), place, ExecuteOptions{
		Execute: true, Confirm: svc.PreviewConditionalPlace(place).ConfirmToken,
	})
	if err != nil {
		t.Fatalf("ConditionalPlace: %v", err)
	}
	if ref.ID != "CO-1" {
		t.Errorf("ref: got %+v", ref)
	}
	if len(broker.created) != 1 || broker.created[0] != place {
		t.Errorf("broker received %+v, want %+v", broker.created, place)
	}

	cancel := cancelIntentFixture()
	if err := svc.ConditionalCancel(context.Background(), cancel, ExecuteOptions{
		Execute: true, Confirm: svc.PreviewConditionalCancel(cancel).ConfirmToken,
	}); err != nil {
		t.Fatalf("ConditionalCancel: %v", err)
	}
	if len(broker.canceled) != 1 || broker.canceled[0] != cancel {
		t.Errorf("broker received %+v", broker.canceled)
	}

	modify := modifyIntentFixture()
	if err := svc.ConditionalModify(context.Background(), modify, ExecuteOptions{
		Execute: true, Confirm: svc.PreviewConditionalModify(modify).ConfirmToken,
	}); err != nil {
		t.Fatalf("ConditionalModify: %v", err)
	}
	if len(broker.modified) != 1 || broker.modified[0].ID != modify.ID {
		t.Errorf("broker received %+v", broker.modified)
	}
}

// TestConditionalExecuteSurfacesBrokerError — a broker failure must reach the
// caller as-is, never be swallowed into a "looks like it worked" result.
func TestConditionalExecuteSurfacesBrokerError(t *testing.T) {
	t.Parallel()

	boom := errors.New("official path rejected the order")
	broker := &fakeConditionalBroker{err: boom}
	svc := NewService(openConditionalPolicy(), nil).WithConditional(broker)

	place := placeIntentFixture()
	if _, err := svc.ConditionalPlace(context.Background(), place, ExecuteOptions{
		Execute: true, Confirm: svc.PreviewConditionalPlace(place).ConfirmToken,
	}); !errors.Is(err, boom) {
		t.Fatalf("want the broker error, got %v", err)
	}
}

// --- intent assembly -------------------------------------------------------

// TestNewConditionalPlaceIntentSecondLeg pins the OCO/OTO rule that used to sit
// inline in the cobra command: a two-leg type without a second leg is refused
// before any token is computed, and SINGLE never carries one.
func TestNewConditionalPlaceIntentSecondLeg(t *testing.T) {
	t.Parallel()

	base := ConditionalPlaceInput{
		Symbol: "005930", Type: "SINGLE", OrderType: "LIMIT", ExpireDate: "2026-12-31", Quantity: 10,
		FirstSide: "SELL", FirstTrigger: 70000, FirstOrderPrice: 69900,
	}

	single, err := NewConditionalPlaceIntent(base)
	if err != nil {
		t.Fatalf("SINGLE: %v", err)
	}
	if single.Second != nil {
		t.Errorf("SINGLE must not carry a second leg: %+v", single.Second)
	}
	if single.First.OrderSide != "SELL" || single.First.TriggerPrice != 70000 {
		t.Errorf("first leg not mapped: %+v", single.First)
	}

	for _, twoLeg := range []string{"OCO", "OTO"} {
		in := base
		in.Type = twoLeg
		if _, err := NewConditionalPlaceIntent(in); err == nil {
			t.Errorf("%s without a second leg must be refused", twoLeg)
		} else if !strings.Contains(err.Error(), "--second-side") {
			t.Errorf("%s: error should name the missing flags, got %v", twoLeg, err)
		}

		in.SecondSide = "BUY"
		in.SecondTrigger = 60000
		in.SecondOrderPrice = 60100
		intent, err := NewConditionalPlaceIntent(in)
		if err != nil {
			t.Fatalf("%s with second leg: %v", twoLeg, err)
		}
		if intent.Second == nil || intent.Second.OrderSide != "BUY" || intent.Second.TriggerPrice != 60000 || intent.Second.OrderPrice != 60100 {
			t.Errorf("%s second leg not mapped: %+v", twoLeg, intent.Second)
		}
	}
}

func TestNewConditionalModifyIntentSecondLeg(t *testing.T) {
	t.Parallel()

	base := ConditionalModifyInput{
		ID: "co-1", Type: "OCO", OrderType: "LIMIT", ExpireDate: "2026-12-31", Quantity: 10,
		FirstSide: "SELL", FirstTrigger: 70000, FirstOrderPrice: 69900,
	}
	if _, err := NewConditionalModifyIntent(base); err == nil {
		t.Error("OCO modify without a second leg must be refused")
	}

	base.SecondSide = "BUY"
	base.SecondTrigger = 60000
	intent, err := NewConditionalModifyIntent(base)
	if err != nil {
		t.Fatalf("OCO modify: %v", err)
	}
	if intent.ID != "co-1" || intent.Second == nil || intent.Second.OrderSide != "BUY" {
		t.Errorf("modify intent not mapped: %+v", intent)
	}
}

// conditionalTokenFor returns the correct confirm token for the verb under test.
func conditionalTokenFor(t *testing.T, svc *Service, verb string) string {
	t.Helper()
	switch verb {
	case "place":
		return svc.PreviewConditionalPlace(placeIntentFixture()).ConfirmToken
	case "cancel":
		return svc.PreviewConditionalCancel(cancelIntentFixture()).ConfirmToken
	case "modify":
		return svc.PreviewConditionalModify(modifyIntentFixture()).ConfirmToken
	}
	t.Fatalf("unknown verb %q", verb)
	return ""
}
