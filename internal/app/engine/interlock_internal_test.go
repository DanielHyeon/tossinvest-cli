package engine

// interlock_internal_test.go covers the two interlock clauses that cannot be
// reached through engine.New (task 7.5).
//
// The refusal matrix in interlock_test.go drives the whole engine, which is the
// right shape for every clause an operator can actually trigger: a missing limit,
// a policy that cannot sell, a Guardian configured from somewhere else. Two
// clauses are different in kind — "no gateway was constructed" and "the transport
// cannot carry an idempotency key" — because construction always produces a wired
// gateway and an official broker. There is no configuration that makes them fail,
// which is exactly why the check exists: it is a guard against a future wiring
// change, and the only way to test a guard against a future is to hand it the
// future.

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

// keylessBroker is a trading.Broker that does not declare the key capability —
// the shape of a transport somebody wires in later without noticing that the
// replay procedure assumes the key was sent.
type keylessBroker struct{}

func (keylessBroker) PlacePendingOrder(context.Context, orderintent.PlaceIntent) (trading.MutationResult, error) {
	return trading.MutationResult{}, errors.New("not implemented")
}

func (keylessBroker) CancelPendingOrder(context.Context, orderintent.CancelIntent) (trading.MutationResult, error) {
	return trading.MutationResult{}, errors.New("not implemented")
}

func (keylessBroker) AmendPendingOrder(context.Context, orderintent.AmendIntent) (trading.MutationResult, error) {
	return trading.MutationResult{}, errors.New("not implemented")
}

func (keylessBroker) GetOrderAvailableActions(context.Context, string) (map[string]any, error) {
	return nil, nil
}

// lyingBroker declares the capability and then denies it, which is the case a
// plain type assertion would let through.
type lyingBroker struct{ keylessBroker }

func (lyingBroker) CarriesIdempotencyKey() bool { return false }

func TestKeylessTransportIsRefused(t *testing.T) {
	if err := checkKeyCapableTransport(nil); err == nil {
		t.Error("a nil transport must be refused")
	}
	if err := checkKeyCapableTransport(keylessBroker{}); err == nil {
		t.Error("a transport that does not declare the key capability must be refused")
	}
	if err := checkKeyCapableTransport(lyingBroker{}); err == nil {
		t.Error("a transport that declares the capability as false must be refused")
	}
	if err := checkKeyCapableTransport(&officialBroker{}); err != nil {
		t.Errorf("the engine's own broker must satisfy the check: %v", err)
	}
}

// openTestJournal opens a journal in an isolated directory, with the filesystem
// probe stubbed for the reason internal/journal's own tests stub it.
func openTestJournal(t *testing.T) *journal.Journal {
	t.Helper()
	j, err := journal.Open(context.Background(), journal.Options{
		Path:     filepath.Join(t.TempDir(), journal.DBFileName),
		FSProber: journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
	})
	if err != nil {
		t.Fatalf("open journal: %v", err)
	}
	t.Cleanup(func() { _ = j.Close() })
	return j
}

func TestGatewayClauseChecksTheWiringAndNotThePointer(t *testing.T) {
	j := openTestJournal(t)
	svc := trading.NewService(config.Trading{}, &officialBroker{})

	if err := checkGateway(nil); err == nil {
		t.Error("no gateway at all must be refused")
	}

	// A gateway that exists but reads no orders back: the identifier round trip
	// never runs and a place is confirmed on the broker's ack alone
	// (issues.md 2026-07-26).
	bare, err := execgw.New(execgw.Options{Journal: j, Trading: svc, AccountRef: "123-45"})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	err = checkGateway(bare)
	if err == nil {
		t.Fatal("a gateway with no order reader must be refused: the round trip is the requirement, " +
			"not the pointer")
	}
	if !strings.Contains(err.Error(), "round trip") {
		t.Errorf("refusal %q does not name the missing round trip", err)
	}

	// Wired the way the engine profile wires it.
	full, err := execgw.New(execgw.Options{
		Journal:    j,
		Trading:    svc,
		AccountRef: "123-45",
		Orders:     execgw.OfficialOrders{},
		Entry:      execgw.NewEntryGate(nil, map[execgw.RequiredQuery]time.Duration{}),
	})
	if err != nil {
		t.Fatalf("execgw.New: %v", err)
	}
	if err := checkGateway(full); err != nil {
		t.Errorf("a fully wired gateway must pass: %v", err)
	}
}

func TestTradingPolicyClauseNamesWhatIsOff(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy config.Trading
		want   string
	}{
		{"sell off", config.Trading{Place: true, Cancel: true, AllowLiveOrderActions: true}, "trading.sell"},
		{"live actions off", config.Trading{Place: true, Sell: true, Cancel: true}, "allow_live_order_actions"},
		// The two the clause did not check until console-owns-the-operating-toggles.
		// A sell is a place, and the exit observer cancels its own proposal, so a
		// policy without them is a policy that cannot liquidate — which is the
		// combination this clause's own sentence forbids.
		{"placing off", config.Trading{Sell: true, Cancel: true, AllowLiveOrderActions: true}, "trading.place"},
		{"cancelling off", config.Trading{Place: true, Sell: true, AllowLiveOrderActions: true}, "trading.cancel"},
		{"all off", config.Trading{}, "and"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := checkTradingPolicy(tc.policy)
			if err == nil {
				t.Fatal("the policy must be refused")
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("refusal %q does not name %q", err, tc.want)
			}
		})
	}
	full := config.Trading{Place: true, Sell: true, Cancel: true, AllowLiveOrderActions: true}
	if err := checkTradingPolicy(full); err != nil {
		t.Errorf("a policy that can place, sell, cancel and act live must pass: %v", err)
	}
	// amend, conditional and fractional are not required: no loop in this build
	// uses them, and an interlock that demands unused capability is a checklist.
	if err := checkTradingPolicy(full); err != nil {
		t.Errorf("the four required toggles must be sufficient: %v", err)
	}
}

// --- the position projection's producer (task 6.1/6.3 succession) ------------

// TestTheProjectionGuardAsksTheBindingAndNotTheData.
//
// Like the two clauses above, this is a guard against a future wiring change:
// NewContext always binds the hook, so there is no configuration that reaches
// the refusal, and the only way to test a guard against a future is to hand it
// the future. The failure it prevents is the worst shape there is — silent, and
// in the direction that lets entries through: with nothing bound the position
// projection stays empty, reconciliation believes the account holds nothing, and
// every holding the broker reports classifies as an external position instead of
// a disagreement that blocks new exposure.
func TestTheProjectionGuardAsksTheBindingAndNotTheData(t *testing.T) {
	if err := checkProjectionWired(nil); err == nil {
		t.Error("no journal at all must be refused")
	}

	unbound := openTestJournal(t)
	err := checkProjectionWired(unbound)
	if !errors.Is(err, ErrProjectionUnbound) {
		t.Fatalf("err = %v, want ErrProjectionUnbound", err)
	}
	// The message has to say what goes wrong, because the operator's instinct on
	// "reconciliation is quiet" is that nothing is wrong.
	if !strings.Contains(err.Error(), "external position") {
		t.Errorf("refusal %q does not name the fail-open it prevents", err)
	}

	bound := openTestJournal(t)
	if err := bindApplyHooks(bound); err != nil {
		t.Fatalf("bindApplyHooks: %v", err)
	}
	if err := checkProjectionWired(bound); err != nil {
		t.Errorf("a bound journal must pass: %v", err)
	}
	// Bound once. A second binding is refused, which is what makes task 7.x
	// extend the literal in gateway.go rather than add a second call site.
	if err := bindApplyHooks(bound); err == nil {
		t.Error("the hooks must not be bindable twice")
	}
}
