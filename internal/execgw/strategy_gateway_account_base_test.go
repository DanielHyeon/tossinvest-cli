//go:build tossos_testseams

package execgw_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/orderintent"
	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
	"github.com/JungHoonGhae/tossinvest-cli/internal/trading"
)

func TestStrategyGatewayBeginsSubmittingImmediatelyBeforeBrokerPairedKRUS(t *testing.T) {
	for _, test := range []struct {
		name, market, symbol, currency, entry, stop, target, rate string
		costMarket                                                costs.Market
	}{
		{name: "KR", market: "KR", symbol: "005930", currency: "KRW", entry: "70000", stop: "69000", target: "72000", rate: "1", costMarket: costs.MarketKR},
		{name: "US", market: "US", symbol: "AAPL", currency: "USD", entry: "50", stop: "45", target: "60", rate: "1400", costMarket: costs.MarketUS},
	} {
		t.Run(test.name, func(t *testing.T) {
			decisionID := "strategy-gateway-" + strings.ToLower(test.market)
			rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) {
				options.NewID = fixedIDs(decisionID, decisionID+"-nonce")
			})
			entry := qFinalAccountBaseRequest(t, rig, "gateway-"+strings.ToLower(test.market), test.market)
			entry.EntryPrice, entry.StopPrice, entry.TargetPrice = test.entry, test.stop, test.target
			fx, err := risk.TestOnlySealAccountBaseFX(fixedNow, test.costMarket, guardianPolicy(), test.rate, "1")
			if err != nil {
				t.Fatal(err)
			}
			bindPairedAccountBaseFXForTest(&entry, fx)
			precheck, err := rig.guardian.PrecheckQFinalEntry(entry)
			if err != nil {
				t.Fatal(err)
			}
			issued, err := rig.guardian.IssuePrecheckedQFinalEntry(context.Background(), precheck)
			if err != nil {
				t.Fatal(err)
			}

			broker := &fakeBroker{result: domain.MutationResult{Kind: "place", Status: "accepted", OrderID: "order-" + strings.ToLower(test.market)}}
			gateway, err := execgw.New(execgw.Options{Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker),
				Clock: rig.clock, AccountRef: "acct-7", Source: "paired-strategy-test"})
			if err != nil {
				t.Fatal(err)
			}
			leaseCAS := journal.StrategyDispatchLeaseCAS{LeaseID: "lease-" + strings.ToLower(test.market), ExpectedRevision: 2,
				OwnerEpoch: 7, FencingToken: "fence-" + strings.ToLower(test.market)}
			decision, err := rig.journal.LookupDecision(context.Background(), issued.Decision.ID)
			if err != nil {
				t.Fatal(err)
			}
			claimed := claimedGatewayLease(test.market, test.symbol, leaseCAS, decision, issued.Reservations[0].ID)
			beginCalls := 0
			gateway.SetStrategyDispatchLeaseForTest(func(_ context.Context, leaseID string) (journal.StrategyDispatchLease, error) {
				if leaseID != claimed.LeaseID {
					t.Fatalf("%s loaded lease %q, want %q", test.market, leaseID, claimed.LeaseID)
				}
				return claimed, nil
			}, func(_ context.Context, got journal.StrategyDispatchLeaseCAS) (journal.StrategyDispatchLease, error) {
				if places, _, _ := broker.totals(); places != 0 {
					t.Fatalf("%s broker called before SUBMITTING CAS: %d", test.market, places)
				}
				if got != leaseCAS {
					t.Fatalf("%s lease CAS = %+v, want %+v", test.market, got, leaseCAS)
				}
				beginCalls++
				started := claimed
				started.State, started.Revision, started.TransportStartedAt = journal.StrategyDispatchLeaseSubmitting, got.ExpectedRevision+1, fixedNow
				return started, nil
			})
			request := execgw.StrategyPlaceRequest{Intent: orderintent.PlaceIntent{Symbol: test.symbol, Market: strings.ToLower(test.market),
				Side: "buy", OrderType: "limit", Quantity: float64(precheck.QFinal()), Price: mustTestFloat(t, test.entry), CurrencyMode: test.currency},
				Decision: issued.Decision, Lease: leaseCAS, FinalAuthorityCheck: func(context.Context) error { return nil }}
			request.SetAccountBaseFXForTest(fx)
			out, err := gateway.PlaceClaimedStrategy(context.Background(), request)
			if err != nil || out.State != journal.StateConfirmed || beginCalls != 1 {
				t.Fatalf("%s outcome=%+v beginCalls=%d err=%v", test.market, out, beginCalls, err)
			}
			if places, _, _ := broker.totals(); places != 1 {
				t.Fatalf("%s broker places=%d, want 1", test.market, places)
			}
		})
	}
}

func TestStrategyGatewayRechecksEntryGateAfterSubmittingFencePairedKRUS(t *testing.T) {
	for _, market := range []string{"KR", "US"} {
		t.Run(market, func(t *testing.T) {
			rig, fx, issued, intent := setupQFinalGatewayDecision(t, market, "entry-gate-race")
			broker := &fakeBroker{}
			gate := execgw.NewEntryGate(rig.clock, map[execgw.RequiredQuery]time.Duration{})
			gateway, err := execgw.New(execgw.Options{Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker),
				Clock: rig.clock, AccountRef: "acct-7", Source: "paired-strategy-test", Entry: gate})
			if err != nil {
				t.Fatal(err)
			}
			decision, err := rig.journal.LookupDecision(context.Background(), issued.Decision.ID)
			if err != nil {
				t.Fatal(err)
			}
			cas := journal.StrategyDispatchLeaseCAS{LeaseID: "lease-entry-gate-race-" + strings.ToLower(market),
				ExpectedRevision: 2, OwnerEpoch: 7, FencingToken: "fence-" + strings.ToLower(market)}
			claimed := claimedGatewayLease(market, intent.Symbol, cas, decision, issued.Reservations[0].ID)
			gateway.SetStrategyDispatchLeaseForTest(func(context.Context, string) (journal.StrategyDispatchLease, error) {
				return claimed, nil
			}, func(context.Context, journal.StrategyDispatchLeaseCAS) (journal.StrategyDispatchLease, error) {
				gate.Block(execgw.ReasonReconcileMismatch, "durable reconcile started after initial entry check")
				started := claimed
				started.State, started.Revision, started.TransportStartedAt = journal.StrategyDispatchLeaseSubmitting, cas.ExpectedRevision+1, fixedNow
				return started, nil
			})
			request := execgw.StrategyPlaceRequest{Intent: intent, Decision: issued.Decision, Lease: cas}
			request.SetAccountBaseFXForTest(fx)
			out, err := gateway.PlaceClaimedStrategy(context.Background(), request)
			var rejected *execgw.RejectedError
			if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonReconcileMismatch || out.State != journal.StateNotDispatched {
				t.Fatalf("%s outcome=%+v rejection=%v err=%v", market, out, rejected, err)
			}
			if places, _, _ := broker.totals(); places != 0 {
				t.Fatalf("%s broker places=%d, want 0 after final entry-gate drift", market, places)
			}
		})
	}
}

func TestStrategyGatewayFinalGateWaitCannotUseExpiredReplacedOwnerKRUS(t *testing.T) {
	for _, market := range []string{"KR", "US"} {
		t.Run(market, func(t *testing.T) {
			rig, fx, issued, intent := setupQFinalGatewayDecision(t, market, "final-gate-expiry")
			broker := &fakeBroker{}
			gate := execgw.NewEntryGate(rig.clock, map[execgw.RequiredQuery]time.Duration{})
			refreshEntered, releaseRefresh := make(chan struct{}), make(chan struct{})
			refreshCalls := 0
			gate.SetAuthorityRefresh(func() error {
				refreshCalls++
				if refreshCalls == 4 {
					close(refreshEntered)
					<-releaseRefresh
				}
				return nil
			})
			gateway, err := execgw.New(execgw.Options{Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker),
				Clock: rig.clock, AccountRef: "acct-7", Source: "paired-strategy-test", Entry: gate})
			if err != nil {
				t.Fatal(err)
			}
			authority, err := gateway.ObserveStrategyEntryGate(context.Background(), strings.ToLower(market), intent.Symbol)
			if err != nil {
				t.Fatal(err)
			}
			decision, err := rig.journal.LookupDecision(context.Background(), issued.Decision.ID)
			if err != nil {
				t.Fatal(err)
			}
			cas := journal.StrategyDispatchLeaseCAS{LeaseID: "lease-final-gate-expiry-" + strings.ToLower(market),
				ExpectedRevision: 2, OwnerEpoch: 7, FencingToken: "fence-" + strings.ToLower(market)}
			claimed := claimedGatewayLease(market, intent.Symbol, cas, decision, issued.Reservations[0].ID)
			gateway.SetStrategyDispatchLeaseForTest(func(context.Context, string) (journal.StrategyDispatchLease, error) {
				return claimed, nil
			}, func(context.Context, journal.StrategyDispatchLeaseCAS) (journal.StrategyDispatchLease, error) {
				started := claimed
				started.State, started.Revision, started.TransportStartedAt = journal.StrategyDispatchLeaseSubmitting, cas.ExpectedRevision+1, fixedNow
				return started, nil
			})
			var replaced atomic.Bool
			gateway.SetStrategyFinalTransportProofForTest(func(context.Context, journal.StrategyDispatchLeaseCAS) error {
				if !replaced.Load() {
					t.Fatal("final transport proof ran before blocked refresh advanced owner")
				}
				return journal.ErrStrategyDispatchFenced
			})
			request := execgw.StrategyPlaceRequest{Intent: intent, Decision: issued.Decision, Lease: cas,
				EntryGateAuthority: authority, FinalAuthorityCheck: func(context.Context) error { return nil }}
			request.SetAccountBaseFXForTest(fx)
			type result struct {
				out execgw.Outcome
				err error
			}
			done := make(chan result, 1)
			go func() {
				out, callErr := gateway.PlaceClaimedStrategy(context.Background(), request)
				done <- result{out: out, err: callErr}
			}()
			select {
			case <-refreshEntered:
				replaced.Store(true)
				close(releaseRefresh)
			case <-time.After(time.Second):
				t.Fatal("final EntryGate refresh did not block")
			}
			got := <-done
			var rejected *execgw.RejectedError
			if !errors.As(got.err, &rejected) || rejected.Reason != execgw.ReasonStrategyDispatchFenced || got.out.State != journal.StateNotDispatched {
				t.Fatalf("%s outcome=%+v rejection=%v err=%v", market, got.out, rejected, got.err)
			}
			if places, _, _ := broker.totals(); places != 0 {
				t.Fatalf("%s expired/replaced owner broker places=%d", market, places)
			}
		})
	}
}

func TestStrategyGatewayRechecksSourceBackedActivationAfterSubmittingFencePairedKRUS(t *testing.T) {
	for _, market := range []string{"KR", "US"} {
		t.Run(market, func(t *testing.T) {
			rig, fx, issued, intent := setupQFinalGatewayDecision(t, market, "activation-race")
			broker := &fakeBroker{}
			gateway, err := execgw.New(execgw.Options{Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker),
				Clock: rig.clock, AccountRef: "acct-7", Source: "paired-strategy-test"})
			if err != nil {
				t.Fatal(err)
			}
			decision, err := rig.journal.LookupDecision(context.Background(), issued.Decision.ID)
			if err != nil {
				t.Fatal(err)
			}
			cas := journal.StrategyDispatchLeaseCAS{LeaseID: "lease-activation-race-" + strings.ToLower(market),
				ExpectedRevision: 2, OwnerEpoch: 7, FencingToken: "fence-" + strings.ToLower(market)}
			claimed := claimedGatewayLease(market, intent.Symbol, cas, decision, issued.Reservations[0].ID)
			gateway.SetStrategyDispatchLeaseForTest(func(context.Context, string) (journal.StrategyDispatchLease, error) {
				return claimed, nil
			}, func(context.Context, journal.StrategyDispatchLeaseCAS) (journal.StrategyDispatchLease, error) {
				started := claimed
				started.State, started.Revision, started.TransportStartedAt = journal.StrategyDispatchLeaseSubmitting, cas.ExpectedRevision+1, fixedNow
				return started, nil
			})
			request := execgw.StrategyPlaceRequest{Intent: intent, Decision: issued.Decision, Lease: cas,
				FinalAuthorityCheck: func(context.Context) error { return errors.New("human changed desired state to OFF") }}
			request.SetAccountBaseFXForTest(fx)
			out, err := gateway.PlaceClaimedStrategy(context.Background(), request)
			var rejected *execgw.RejectedError
			if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonStrategyDispatchFenced || out.State != journal.StateNotDispatched {
				t.Fatalf("%s outcome=%+v rejection=%v err=%v", market, out, rejected, err)
			}
			if places, _, _ := broker.totals(); places != 0 {
				t.Fatalf("%s broker places=%d, want 0 after activation drift", market, places)
			}
		})
	}
}

func TestStrategyGatewayRejectsMissingOrFencedAuthorityPairedKRUS(t *testing.T) {
	for _, market := range []string{"KR", "US"} {
		t.Run(market, func(t *testing.T) {
			t.Run("ordinary place lacks lease and FX", func(t *testing.T) {
				rig, _, issued, intent := setupQFinalGatewayDecision(t, market, "missing")
				broker := &fakeBroker{}
				gateway, err := execgw.New(execgw.Options{Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker), Clock: rig.clock, AccountRef: "acct-7"})
				if err != nil {
					t.Fatal(err)
				}
				_, err = gateway.Place(context.Background(), execgw.PlaceRequest{Intent: intent, Decision: issued.Decision})
				var rejected *execgw.RejectedError
				if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonStrategyDispatchAuthorityMissing {
					t.Fatalf("%s missing strategy authority error=%v", market, err)
				}
				if places, _, _ := broker.totals(); places != 0 {
					t.Fatalf("%s missing authority broker places=%d", market, places)
				}
			})

			t.Run("claimed lease fenced at final boundary", func(t *testing.T) {
				rig, fx, issued, intent := setupQFinalGatewayDecision(t, market, "fenced")
				broker := &fakeBroker{}
				gateway, err := execgw.New(execgw.Options{Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker), Clock: rig.clock, AccountRef: "acct-7"})
				if err != nil {
					t.Fatal(err)
				}
				decision, err := rig.journal.LookupDecision(context.Background(), issued.Decision.ID)
				if err != nil {
					t.Fatal(err)
				}
				leaseCAS := journal.StrategyDispatchLeaseCAS{LeaseID: "lease-" + strings.ToLower(market), ExpectedRevision: 2, OwnerEpoch: 7, FencingToken: "fence"}
				claimed := claimedGatewayLease(market, intent.Symbol, leaseCAS, decision, issued.Reservations[0].ID)
				gateway.SetStrategyDispatchLeaseForTest(func(context.Context, string) (journal.StrategyDispatchLease, error) {
					return claimed, nil
				}, func(context.Context, journal.StrategyDispatchLeaseCAS) (journal.StrategyDispatchLease, error) {
					return journal.StrategyDispatchLease{}, journal.ErrStrategyDispatchFenced
				})
				request := execgw.StrategyPlaceRequest{Intent: intent, Decision: issued.Decision,
					Lease: leaseCAS}
				request.SetAccountBaseFXForTest(fx)
				_, err = gateway.PlaceClaimedStrategy(context.Background(), request)
				var rejected *execgw.RejectedError
				if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonStrategyDispatchFenced {
					t.Fatalf("%s fenced strategy error=%v", market, err)
				}
				if places, _, _ := broker.totals(); places != 0 {
					t.Fatalf("%s fenced broker places=%d", market, places)
				}
			})
		})
	}
}

func TestStrategyGatewaySettlesDefinitiveRefusalAndAmbiguityPairedKRUS(t *testing.T) {
	for _, market := range []string{"KR", "US"} {
		for _, outcome := range []struct {
			name      string
			brokerErr error
			wantCore  journal.AttemptState
		}{
			{name: "definitive-refusal", brokerErr: official.ErrAuth, wantCore: journal.StateFailedConfirmed},
			{name: "ambiguous", brokerErr: official.ErrServer, wantCore: journal.StateInDoubt},
		} {
			t.Run(market+"/"+outcome.name, func(t *testing.T) {
				rig, fx, issued, intent := setupQFinalGatewayDecision(t, market, "settle-"+outcome.name)
				broker := &fakeBroker{err: outcome.brokerErr}
				gateway, err := execgw.New(execgw.Options{Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker),
					Clock: rig.clock, AccountRef: "acct-7", Source: "paired-strategy-test"})
				if err != nil {
					t.Fatal(err)
				}
				decision, err := rig.journal.LookupDecision(context.Background(), issued.Decision.ID)
				if err != nil {
					t.Fatal(err)
				}
				cas := journal.StrategyDispatchLeaseCAS{LeaseID: "lease-settle-" + strings.ToLower(market) + "-" + outcome.name,
					ExpectedRevision: 2, OwnerEpoch: 7, FencingToken: "fence-" + strings.ToLower(market)}
				claimed := claimedGatewayLease(market, intent.Symbol, cas, decision, issued.Reservations[0].ID)
				gateway.SetStrategyDispatchLeaseForTest(func(context.Context, string) (journal.StrategyDispatchLease, error) {
					return claimed, nil
				}, func(context.Context, journal.StrategyDispatchLeaseCAS) (journal.StrategyDispatchLease, error) {
					started := claimed
					started.State, started.Revision, started.TransportStartedAt = journal.StrategyDispatchLeaseSubmitting, cas.ExpectedRevision+1, fixedNow
					return started, nil
				})
				request := execgw.StrategyPlaceRequest{Intent: intent, Decision: issued.Decision, Lease: cas,
					FinalAuthorityCheck: func(context.Context) error { return nil }}
				request.SetAccountBaseFXForTest(fx)
				got, callErr := gateway.PlaceClaimedStrategy(context.Background(), request)
				if callErr == nil || got.State != outcome.wantCore {
					t.Fatalf("%s/%s core=%+v err=%v", market, outcome.name, got, callErr)
				}
				if places, _, _ := broker.totals(); places != 1 {
					t.Fatalf("%s/%s broker places=%d, want 1", market, outcome.name, places)
				}
			})
		}
	}
}

func TestStrategyGatewayRejectsStaleAndCrossMarketFXBeforeSubmittingPairedKRUS(t *testing.T) {
	for _, market := range []string{"KR", "US"} {
		for _, failure := range []string{"stale", "cross-market"} {
			t.Run(market+"/"+failure, func(t *testing.T) {
				rig, _, issued, intent := setupQFinalGatewayDecision(t, market, "fx-"+failure)
				broker := &fakeBroker{}
				gateway, err := execgw.New(execgw.Options{Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker), Clock: rig.clock, AccountRef: "acct-7"})
				if err != nil {
					t.Fatal(err)
				}
				loadCalls, beginCalls, refusalCalls := 0, 0, 0
				cas := journal.StrategyDispatchLeaseCAS{LeaseID: "lease-fx-" + strings.ToLower(market), ExpectedRevision: 2, OwnerEpoch: 7, FencingToken: "fence"}
				decision, err := rig.journal.LookupDecision(context.Background(), issued.Decision.ID)
				if err != nil {
					t.Fatal(err)
				}
				claimed := claimedGatewayLease(market, intent.Symbol, cas, decision, issued.Reservations[0].ID)
				gateway.SetStrategyDispatchLeaseForTest(func(context.Context, string) (journal.StrategyDispatchLease, error) {
					loadCalls++
					return claimed, nil
				}, func(context.Context, journal.StrategyDispatchLeaseCAS) (journal.StrategyDispatchLease, error) {
					beginCalls++
					return journal.StrategyDispatchLease{}, nil
				})
				gateway.SetStrategyPreTransportRefusalForTest(func(_ context.Context, got journal.StrategyDispatchPreTransportRefusalRequest) (journal.StrategyDispatchLease, error) {
					refusalCalls++
					if got.Lease != cas || got.Binding != claimed.StrategyDispatchLeasePlan || got.Reason != journal.StrategyDispatchPreTransportAccountBaseFXRefused {
						t.Fatalf("%s/%s refusal request=%+v", market, failure, got)
					}
					terminal := claimed
					terminal.State = journal.StrategyDispatchLeaseRefused
					terminal.Disposition = journal.StrategyDispatchReservationReleased
					return terminal, nil
				})
				fxMarket, rate, at := costs.MarketKR, "1", fixedNow
				if market == "US" {
					fxMarket, rate = costs.MarketUS, "1400"
				}
				if failure == "stale" {
					at = fixedNow.Add(-time.Second)
				} else if fxMarket == costs.MarketKR {
					fxMarket, rate = costs.MarketUS, "1400"
				} else {
					fxMarket, rate = costs.MarketKR, "1"
				}
				fx, err := risk.TestOnlySealAccountBaseFX(at, fxMarket, guardianPolicy(), rate, "1")
				if err != nil {
					t.Fatal(err)
				}
				request := execgw.StrategyPlaceRequest{Intent: intent, Decision: issued.Decision,
					Lease: cas}
				request.SetAccountBaseFXForTest(fx)
				_, err = gateway.PlaceClaimedStrategy(context.Background(), request)
				var rejected *execgw.RejectedError
				if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonAccountBaseFXMismatch {
					t.Fatalf("%s/%s FX refusal=%v", market, failure, err)
				}
				if places, _, _ := broker.totals(); places != 0 || loadCalls != 1 || beginCalls != 0 || refusalCalls != 1 {
					t.Fatalf("%s/%s places=%d load/begin/refusal=%d/%d/%d, want 0/1/0/1", market, failure, places, loadCalls, beginCalls, refusalCalls)
				}
			})
		}
	}
}

func TestStrategyGatewayRejectsValidLeaseFromDifferentDecisionPairedKRUS(t *testing.T) {
	for _, market := range []string{"KR", "US"} {
		t.Run(market, func(t *testing.T) {
			rig, fx, issued, intent := setupQFinalGatewayDecision(t, market, "lease-substitution")
			broker := &fakeBroker{}
			gateway, err := execgw.New(execgw.Options{Journal: rig.journal, Trading: trading.NewService(openPolicy(), broker), Clock: rig.clock, AccountRef: "acct-7"})
			if err != nil {
				t.Fatal(err)
			}
			decision, err := rig.journal.LookupDecision(context.Background(), issued.Decision.ID)
			if err != nil {
				t.Fatal(err)
			}
			cas := journal.StrategyDispatchLeaseCAS{LeaseID: "lease-other-" + strings.ToLower(market), ExpectedRevision: 2, OwnerEpoch: 7, FencingToken: "fence"}
			other := claimedGatewayLease(market, intent.Symbol, cas, decision, issued.Reservations[0].ID)
			other.GuardianDecisionID = "different-valid-decision"
			beginCalls := 0
			gateway.SetStrategyDispatchLeaseForTest(func(context.Context, string) (journal.StrategyDispatchLease, error) {
				return other, nil
			}, func(context.Context, journal.StrategyDispatchLeaseCAS) (journal.StrategyDispatchLease, error) {
				beginCalls++
				return other, nil
			})
			request := execgw.StrategyPlaceRequest{Intent: intent, Decision: issued.Decision, Lease: cas}
			request.SetAccountBaseFXForTest(fx)
			_, err = gateway.PlaceClaimedStrategy(context.Background(), request)
			var rejected *execgw.RejectedError
			if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonStrategyDispatchFenced {
				t.Fatalf("%s substituted lease error=%v", market, err)
			}
			if places, _, _ := broker.totals(); places != 0 || beginCalls != 0 {
				t.Fatalf("%s substituted lease places=%d beginCalls=%d, want 0/0", market, places, beginCalls)
			}
		})
	}
}

func TestStrategyLeaseCannotBeConsumedByReductionPairedKRUS(t *testing.T) {
	for _, test := range []struct{ market, symbol, currency string }{{"kr", "005930", "KRW"}, {"us", "AAPL", "USD"}} {
		t.Run(strings.ToUpper(test.market), func(t *testing.T) {
			broker := &fakeBroker{}
			gateway, journalDB, clk := newGateway(t, broker)
			intent := orderintent.PlaceIntent{Symbol: test.symbol, Market: test.market, Side: "sell", OrderType: "limit", Quantity: 1, Price: 10, CurrencyMode: test.currency}
			decision := exitDecision(t, journalDB, clk, journal.KindPlace, test.market, test.symbol, "SELL", 1)
			leaseCAS := journal.StrategyDispatchLeaseCAS{LeaseID: "lease-reduction-" + test.market, ExpectedRevision: 2, OwnerEpoch: 7, FencingToken: "fence"}
			loads, begins := 0, 0
			gateway.SetStrategyDispatchLeaseForTest(func(context.Context, string) (journal.StrategyDispatchLease, error) {
				loads++
				return journal.StrategyDispatchLease{StrategyDispatchLeasePlan: journal.StrategyDispatchLeasePlan{LeaseID: leaseCAS.LeaseID},
					State: journal.StrategyDispatchLeaseClaimed, Disposition: journal.StrategyDispatchReservationReserved, Revision: leaseCAS.ExpectedRevision}, nil
			}, func(context.Context, journal.StrategyDispatchLeaseCAS) (journal.StrategyDispatchLease, error) {
				begins++
				return journal.StrategyDispatchLease{}, nil
			})
			_, err := gateway.PlaceClaimedStrategy(context.Background(), execgw.StrategyPlaceRequest{Intent: intent, Decision: decision, Lease: leaseCAS})
			var rejected *execgw.RejectedError
			if !errors.As(err, &rejected) || rejected.Reason != execgw.ReasonStrategyDispatchFenced {
				t.Fatalf("reduction lease error=%v", err)
			}
			if places, _, _ := broker.totals(); places != 0 || loads != 1 || begins != 0 {
				t.Fatalf("reduction path places=%d loads=%d begins=%d, want 0/1/0", places, loads, begins)
			}
		})
	}
}

func claimedGatewayLease(market, symbol string, cas journal.StrategyDispatchLeaseCAS, decision journal.Decision, reservationID string) journal.StrategyDispatchLease {
	return journal.StrategyDispatchLease{StrategyDispatchLeasePlan: journal.StrategyDispatchLeasePlan{
		LeaseID: cas.LeaseID, OperationID: decision.ClientOrderID, AccountRef: decision.AccountRef,
		Market: journal.StrategyDispatchMarket(strings.ToUpper(market)), Symbol: symbol,
		RiskReservationID: reservationID, GuardianDecisionID: decision.ID, OwnerEpoch: cas.OwnerEpoch, FencingToken: cas.FencingToken,
	}, State: journal.StrategyDispatchLeaseClaimed, Disposition: journal.StrategyDispatchReservationReserved, Revision: cas.ExpectedRevision}
}

func setupQFinalGatewayDecision(t *testing.T, market, suffix string) (*guardianRig, risk.AccountBaseFX, execgw.Issued, orderintent.PlaceIntent) {
	t.Helper()
	decisionID := "strategy-gateway-" + suffix + "-" + strings.ToLower(market)
	rig := newGuardian(t, func(options *execgw.RiskGuardianOptions) { options.NewID = fixedIDs(decisionID, decisionID+"-nonce") })
	entry := qFinalAccountBaseRequest(t, rig, "gateway-"+suffix+"-"+strings.ToLower(market), market)
	costMarket, rate := costs.MarketKR, "1"
	if market == "US" {
		costMarket, rate = costs.MarketUS, "1400"
	}
	fx, err := risk.TestOnlySealAccountBaseFX(fixedNow, costMarket, guardianPolicy(), rate, "1")
	if err != nil {
		t.Fatal(err)
	}
	bindPairedAccountBaseFXForTest(&entry, fx)
	precheck, err := rig.guardian.PrecheckQFinalEntry(entry)
	if err != nil {
		t.Fatal(err)
	}
	issued, err := rig.guardian.IssuePrecheckedQFinalEntry(context.Background(), precheck)
	if err != nil {
		t.Fatal(err)
	}
	intent := orderintent.PlaceIntent{Symbol: entry.Symbol, Market: strings.ToLower(market), Side: "buy", OrderType: "limit",
		Quantity: float64(precheck.QFinal()), Price: mustTestFloat(t, entry.EntryPrice), CurrencyMode: entry.Currency}
	return rig, fx, issued, intent
}

func mustTestFloat(t *testing.T, raw string) float64 {
	t.Helper()
	var value float64
	if _, err := fmt.Sscan(raw, &value); err != nil {
		t.Fatal(err)
	}
	return value
}
