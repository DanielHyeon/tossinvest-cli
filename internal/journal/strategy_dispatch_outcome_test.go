package journal

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestDispatchStrategyVerifiedClosesACKFillRaceInOneCompositeTxPairedKRUS(t *testing.T) {
	for _, market := range []struct {
		name, symbol, currency string
	}{{"KR", "005930", "KRW"}, {"US", "AAPL", "USD"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Campaign: ApplyPositionCampaignFill}); err != nil {
				t.Fatal(err)
			}
			owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j,
				"composite-"+strings.ToLower(market.name), market.name, market.symbol)
			attempt, core := prepareCoreStrategyAttempt(t, j, plan, market.name, market.symbol, market.currency)
			orderID := "\tstrategy-composite-" + strings.ToLower(market.name) + " opaque\n"
			callerCtx, cancelCaller := context.WithCancel(context.Background())
			brokerCalls, verifyCalls := 0, 0
			result, err := attempt.DispatchStrategyVerified(callerCtx, StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
			}, func(context.Context, *Attempt) DispatchOutcome {
				brokerCalls++
				cancelCaller() // after-send durability must not inherit this cancellation
				return DispatchOutcome{Class: DispatchAcked, BrokerOrderID: orderID}
			}, func(settlementCtx context.Context, gotOrderID string) error {
				verifyCalls++
				if settlementCtx.Err() != nil {
					t.Fatalf("settlement context inherited caller cancellation: %v", settlementCtx.Err())
				}
				if gotOrderID != orderID {
					t.Fatalf("verify order ID=%q want byte-exact %q", gotOrderID, orderID)
				}
				fill, fillErr := j.RecordFill(settlementCtx, FillObservation{
					OrderID: orderID, AccountRef: plan.AccountRef, Market: strings.ToLower(market.name),
					TradingDay: "2026-03-30", Symbol: market.symbol, Side: "BUY", State: "OPEN_PARTIALLY_FILLED",
					Quantity: "10", FilledQuantity: "3", AveragePrice: "5", FilledAmount: "15",
					BrokerVisibleAt: "2026-03-30T00:30:20Z", ObservedAt: "2026-03-30T00:30:20Z",
				})
				if fillErr != nil || !fill.Changed {
					t.Fatalf("ACK-window fill=%+v err=%v", fill, fillErr)
				}
				return nil
			})
			if err != nil {
				t.Fatal(err)
			}
			if brokerCalls != 1 || verifyCalls != 1 || result.Final != StateConfirmed || result.BrokerOrderID != orderID {
				t.Fatalf("result=%+v broker/verify=%d/%d", result, brokerCalls, verifyCalls)
			}
			lease, err := loadStrategyDispatchLease(context.Background(), j.db, plan.LeaseID)
			if err != nil || lease.State != StrategyDispatchLeaseSubmitted || lease.Disposition != StrategyDispatchReservationTransferred || lease.BrokerOrderID != orderID {
				t.Fatalf("lease=%+v err=%v", lease, err)
			}
			attemptRecord, err := j.LookupAttempt(context.Background(), core.attemptID)
			if err != nil || attemptRecord.State != StateConfirmed || attemptRecord.BrokerOrderID != orderID {
				t.Fatalf("core attempt=%+v err=%v", attemptRecord, err)
			}
			var riskMappings int
			if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_order_reservations r
				JOIN risk_bucket_orders o ON o.order_key=r.order_key WHERE o.order_id=?`, orderID).Scan(&riskMappings); err != nil || riskMappings != 5 {
				t.Fatalf("risk mappings=%d err=%v", riskMappings, err)
			}
			var cumulative, remaining string
			if err := j.db.QueryRow(`SELECT cumulative_filled,remaining_quantity FROM campaign_order_watermarks
				WHERE campaign_id=? AND leg_sequence=1 AND order_id=?`, plan.CampaignID, orderID).
				Scan(&cumulative, &remaining); err != nil || cumulative != "3" || remaining != "7" {
				t.Fatalf("campaign cumulative/remaining=%s/%s err=%v", cumulative, remaining, err)
			}
			position, err := j.CurrentPosition(context.Background(), plan.AccountRef, strings.ToLower(market.name), market.symbol)
			if err != nil || position.Quantity != "3" {
				var snapshotCommitted, settledAt string
				_ = j.db.QueryRow(`SELECT committed_at FROM scoped_fill_snapshots WHERE order_id=?`, orderID).Scan(&snapshotCommitted)
				_ = j.db.QueryRow(`SELECT settled_at FROM mutation_attempts WHERE id=?`, core.attemptID).Scan(&settledAt)
				t.Fatalf("position=%+v err=%v snapshot=%s settled=%s", position, err, snapshotCommitted, settledAt)
			}
			var rawRisk, rawCampaign, rawStrategy int
			_ = j.db.QueryRow(`SELECT count(*) FROM risk_bucket_orders WHERE order_id=?`, orderID).Scan(&rawRisk)
			_ = j.db.QueryRow(`SELECT count(*) FROM campaign_order_watermarks WHERE order_id=?`, orderID).Scan(&rawCampaign)
			_ = j.db.QueryRow(`SELECT count(*) FROM strategy_execution_lineage WHERE kind='BROKER_ORDER' AND external_ref=?`, orderID).Scan(&rawStrategy)
			if rawRisk != 1 || rawCampaign != 1 || rawStrategy != 1 {
				t.Fatalf("byte-exact raw order links risk/campaign/strategy=%d/%d/%d", rawRisk, rawCampaign, rawStrategy)
			}
			j.clk.(*clock.Fake).Set(time.Date(2026, 3, 30, 0, 30, 20, 500000000, time.UTC))
			identical, err := j.RecordFill(context.Background(), FillObservation{
				OrderID: orderID, AccountRef: plan.AccountRef, Market: strings.ToLower(market.name),
				TradingDay: "2026-03-30", Symbol: market.symbol, Side: "BUY", State: "OPEN_PARTIALLY_FILLED",
				Quantity: "10", FilledQuantity: "3", AveragePrice: "5", FilledAmount: "15",
				BrokerVisibleAt: "2026-03-30T00:30:20Z", ObservedAt: "2026-03-30T00:30:20.5Z",
			})
			if err != nil || identical.Changed {
				t.Fatalf("rebased identical fill=%+v err=%v", identical, err)
			}

			j.clk.(*clock.Fake).Set(time.Date(2026, 3, 30, 0, 30, 21, 0, time.UTC))
			after, err := j.RecordFill(context.Background(), FillObservation{
				OrderID: orderID, AccountRef: plan.AccountRef, Market: strings.ToLower(market.name),
				TradingDay: "2026-03-30", Symbol: market.symbol, Side: "BUY", State: "OPEN_PARTIALLY_FILLED",
				Quantity: "10", FilledQuantity: "5", AveragePrice: "5", FilledAmount: "25",
				BrokerVisibleAt: "2026-03-30T00:30:21Z", ObservedAt: "2026-03-30T00:30:21Z",
			})
			if err != nil || !after.Changed || after.Delta != "2" {
				t.Fatalf("post-composite fill=%+v err=%v", after, err)
			}
			position, err = j.CurrentPosition(context.Background(), plan.AccountRef, strings.ToLower(market.name), market.symbol)
			if err != nil || position.Quantity != "5" {
				t.Fatalf("post-composite position=%+v err=%v", position, err)
			}
			var riskCumulative uint64
			if err := j.db.QueryRow(`SELECT cumulative_fill FROM risk_bucket_orders WHERE order_id=?`, orderID).Scan(&riskCumulative); err != nil || riskCumulative != 5 {
				t.Fatalf("risk cumulative=%d err=%v", riskCumulative, err)
			}
			if err := j.db.QueryRow(`SELECT cumulative_filled,remaining_quantity FROM campaign_order_watermarks
				WHERE campaign_id=? AND leg_sequence=1 AND order_id=?`, plan.CampaignID, orderID).
				Scan(&cumulative, &remaining); err != nil || cumulative != "5" || remaining != "5" {
				t.Fatalf("post-composite campaign=%s/%s err=%v", cumulative, remaining, err)
			}

			secondBrokerCalls := 0
			_, err = attempt.DispatchStrategyVerified(context.Background(), StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
			}, func(context.Context, *Attempt) DispatchOutcome {
				secondBrokerCalls++
				return DispatchOutcome{Class: DispatchAcked, BrokerOrderID: "must-not-resend"}
			}, func(context.Context, string) error { return nil })
			if !errors.Is(err, ErrStrategyDispatchLeaseConsumed) || secondBrokerCalls != 0 {
				t.Fatalf("duplicate dispatch err=%v second broker calls=%d", err, secondBrokerCalls)
			}
		})
	}
}

func TestDispatchStrategyVerifiedPreventsOwnerTakeoverBetweenSubmittingAndBrokerKRUS(t *testing.T) {
	for _, market := range []struct{ name, symbol, currency string }{{"KR", "005930", "KRW"}, {"US", "AAPL", "USD"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j,
				"owner-transport-race-"+strings.ToLower(market.name), market.name, market.symbol)
			attempt, _ := prepareCoreStrategyAttempt(t, j, plan, market.name, market.symbol, market.currency)
			cas := StrategyDispatchLeaseCAS{LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken}
			brokerCalls := 0
			result, err := attempt.DispatchStrategyVerified(context.Background(), cas, func(ctx context.Context, _ *Attempt) DispatchOutcome {
				if _, takeoverErr := j.AcquireStrategyDispatchOwner(ctx, "racing-replacement-"+strings.ToLower(market.name)); !errors.Is(takeoverErr, ErrStrategyDispatchOwnerBusy) {
					t.Fatalf("takeover after SUBMITTING error=%v", takeoverErr)
				}
				if fenceErr := j.RequireCurrentStrategyDispatchTransportAuthority(ctx, cas); fenceErr != nil {
					t.Fatalf("final transport fence after refused takeover=%v", fenceErr)
				}
				brokerCalls++
				return DispatchOutcome{Class: DispatchAcked, BrokerOrderID: "race-order-" + strings.ToLower(market.name)}
			}, func(context.Context, string) error { return nil })
			if err != nil || result.Final != StateConfirmed || brokerCalls != 1 {
				t.Fatalf("result=%+v brokerCalls=%d err=%v", result, brokerCalls, err)
			}
		})
	}
}

func TestDispatchStrategyVerifiedRequiresOfficialVerifierBeforeMutationPairedKRUS(t *testing.T) {
	for _, market := range []struct {
		name, symbol, currency string
	}{{"KR", "005930", "KRW"}, {"US", "AAPL", "USD"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j,
				"nil-verify-"+strings.ToLower(market.name), market.name, market.symbol)
			attempt, _ := prepareCoreStrategyAttempt(t, j, plan, market.name, market.symbol, market.currency)
			brokerCalls := 0
			_, err := attempt.DispatchStrategyVerified(context.Background(), StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
			}, func(context.Context, *Attempt) DispatchOutcome {
				brokerCalls++
				return DispatchOutcome{Class: DispatchAcked, BrokerOrderID: "must-not-send"}
			}, nil)
			if !errors.Is(err, ErrInvalidRequest) || brokerCalls != 0 {
				t.Fatalf("nil verifier err=%v broker calls=%d", err, brokerCalls)
			}
			record, lookupErr := j.LookupAttempt(context.Background(), attempt.ID())
			if lookupErr != nil || record.State != StateRecorded {
				t.Fatalf("attempt mutated=%+v err=%v", record, lookupErr)
			}
			lease, lookupErr := j.LookupStrategyDispatchLease(context.Background(), plan.LeaseID)
			if lookupErr != nil || lease.State != StrategyDispatchLeaseClaimed ||
				lease.Disposition != StrategyDispatchReservationReserved || lease.Revision != claimed.Revision {
				t.Fatalf("lease mutated=%+v err=%v", lease, lookupErr)
			}
		})
	}
}

func TestDispatchStrategyVerifiedReleasesTerminalACKWindowRemaindersPairedKRUS(t *testing.T) {
	for _, market := range []struct {
		name, symbol, currency string
	}{{"KR", "005930", "KRW"}, {"US", "AAPL", "USD"}} {
		for _, filled := range []string{"0", "3"} {
			t.Run(market.name+"/filled-"+filled, func(t *testing.T) {
				j := openTestJournal(t)
				if err := j.SetApplyHooks(ApplyHooks{Project: ProjectPosition, Campaign: ApplyPositionCampaignFill}); err != nil {
					t.Fatal(err)
				}
				owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j,
					"terminal-"+strings.ToLower(market.name)+"-"+filled, market.name, market.symbol)
				attempt, _ := prepareCoreStrategyAttempt(t, j, plan, market.name, market.symbol, market.currency)
				orderID := "terminal-ack-window-" + strings.ToLower(market.name) + "-" + filled
				_, err := attempt.DispatchStrategyVerified(context.Background(), StrategyDispatchLeaseCAS{
					LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
					OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
				}, func(context.Context, *Attempt) DispatchOutcome {
					return DispatchOutcome{Class: DispatchAcked, BrokerOrderID: orderID}
				}, func(ctx context.Context, got string) error {
					if got != orderID {
						t.Fatalf("order id=%q want=%q", got, orderID)
					}
					_, fillErr := j.RecordFill(ctx, FillObservation{
						OrderID: orderID, AccountRef: plan.AccountRef, Market: strings.ToLower(market.name),
						TradingDay: "2026-03-30", Symbol: market.symbol, Side: "BUY", State: "CLOSED_CANCELLED",
						Terminal: true, Quantity: "10", FilledQuantity: filled, AveragePrice: "5",
						FilledAmount: filled, BrokerVisibleAt: "2026-03-30T00:30:20Z", ObservedAt: "2026-03-30T00:30:20Z",
					})
					return fillErr
				})
				if err != nil {
					t.Fatal(err)
				}
				var orderState, releaseReason string
				var cumulative uint64
				if err := j.db.QueryRow(`SELECT state,release_reason,cumulative_fill FROM risk_bucket_orders WHERE order_id=?`, orderID).
					Scan(&orderState, &releaseReason, &cumulative); err != nil {
					t.Fatal(err)
				}
				if orderState != "RELEASED" || releaseReason != string(RiskBucketReleaseBrokerTerminal) {
					t.Fatalf("risk order state/reason=%s/%s", orderState, releaseReason)
				}
				var heldBuckets, releasedBuckets int
				if err := j.db.QueryRow(`SELECT count(*),COALESCE(sum(CASE WHEN state='RELEASED' AND held_minor='0' THEN 1 ELSE 0 END),0)
					FROM risk_bucket_reservations WHERE decision_id=?`, plan.GuardianDecisionID).
					Scan(&heldBuckets, &releasedBuckets); err != nil || heldBuckets != 5 || releasedBuckets != 5 {
					t.Fatalf("buckets total/released=%d/%d err=%v", heldBuckets, releasedBuckets, err)
				}
				var aggregateState string
				if err := j.db.QueryRow(`SELECT state FROM risk_reservations WHERE id=?`, plan.RiskReservationID).
					Scan(&aggregateState); err != nil || aggregateState != "RELEASED" {
					t.Fatalf("aggregate=%s err=%v", aggregateState, err)
				}
				wantCumulative := uint64(0)
				if filled == "3" {
					wantCumulative = 3
				}
				if cumulative != wantCumulative {
					t.Fatalf("cumulative=%d want=%d", cumulative, wantCumulative)
				}
			})
		}
	}
}

func TestLoadStrategyDispatchLeasePreservesEqualTimestampsPairedKRUS(t *testing.T) {
	for _, market := range []struct{ name, symbol string }{{"KR", "005930"}, {"US", "AAPL"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			owner, err := j.AcquireStrategyDispatchOwner(context.Background(), "equal-time-"+strings.ToLower(market.name))
			if err != nil {
				t.Fatal(err)
			}
			plan := seedBoundStrategyDispatchClaimLease(t, j, owner, "equal-time-"+strings.ToLower(market.name),
				"acct-"+strings.ToLower(market.name), market.name, market.symbol)
			var issued, expires, created, updated string
			if err := j.db.QueryRow(`SELECT issued_at,expires_at,created_at,updated_at FROM strategy_dispatch_leases WHERE lease_id=?`,
				plan.LeaseID).Scan(&issued, &expires, &created, &updated); err != nil {
				t.Fatal(err)
			}
			if issued != created || issued != updated {
				t.Fatalf("fixture does not exercise equal timestamps: %q/%q/%q", issued, created, updated)
			}
			loaded, err := loadStrategyDispatchLease(context.Background(), j.db, plan.LeaseID)
			if err != nil {
				t.Fatal(err)
			}
			wantIssued, _ := parseJournalTime(issued)
			wantExpires, _ := parseJournalTime(expires)
			wantCreated, _ := parseJournalTime(created)
			wantUpdated, _ := parseJournalTime(updated)
			if !loaded.IssuedAt.Equal(wantIssued) || !loaded.ExpiresAt.Equal(wantExpires) ||
				!loaded.CreatedAt.Equal(wantCreated) || !loaded.UpdatedAt.Equal(wantUpdated) ||
				loaded.IssuedAt.IsZero() || loaded.CreatedAt.IsZero() || loaded.UpdatedAt.IsZero() {
				t.Fatalf("loaded timestamps issued=%s expires=%s created=%s updated=%s", loaded.IssuedAt,
					loaded.ExpiresAt, loaded.CreatedAt, loaded.UpdatedAt)
			}
		})
	}
}

func TestDispatchStrategyVerifiedPairedKRUSRefusalAndAmbiguityMatrix(t *testing.T) {
	cases := []struct {
		name           string
		outcome        DispatchOutcome
		verifyError    error
		core           AttemptState
		lease          StrategyDispatchLeaseState
		disposition    StrategyDispatchReservationDisposition
		aggregate      string
		buckets        int
		durableOrderID string
	}{
		{name: "not-sent", outcome: DispatchOutcome{Class: DispatchNotSent}, core: StateNotDispatched,
			lease: StrategyDispatchLeaseRefused, disposition: StrategyDispatchReservationReleased, aggregate: "RELEASED"},
		{name: "rejected", outcome: DispatchOutcome{Class: DispatchRejected}, core: StateFailedConfirmed,
			lease: StrategyDispatchLeaseRefused, disposition: StrategyDispatchReservationReleased, aggregate: "RELEASED"},
		{name: "ambiguous", outcome: DispatchOutcome{Class: DispatchAmbiguous}, core: StateInDoubt,
			lease: StrategyDispatchLeaseAmbiguous, disposition: StrategyDispatchReservationHeld, aggregate: "HELD", buckets: 5},
		{name: "ack-blank", outcome: DispatchOutcome{Class: DispatchAcked, BrokerOrderID: " \t\n"}, core: StateInDoubt,
			lease: StrategyDispatchLeaseAmbiguous, disposition: StrategyDispatchReservationHeld, aggregate: "HELD", buckets: 5},
		{name: "ack-verify-failed", outcome: DispatchOutcome{Class: DispatchAcked, BrokerOrderID: "\topaque uncertain\n"},
			verifyError: errors.New("official lookup unavailable"), core: StateInDoubt,
			lease: StrategyDispatchLeaseAmbiguous, disposition: StrategyDispatchReservationHeld, aggregate: "HELD", buckets: 5,
			durableOrderID: "\topaque uncertain\n"},
		{name: "unknown-class", outcome: DispatchOutcome{Class: DispatchClass("FORGED")}, core: StateInDoubt,
			lease: StrategyDispatchLeaseAmbiguous, disposition: StrategyDispatchReservationHeld, aggregate: "HELD", buckets: 5},
	}
	for _, market := range []struct {
		name, symbol, currency string
	}{{"KR", "005930", "KRW"}, {"US", "AAPL", "USD"}} {
		for _, tc := range cases {
			t.Run(market.name+"/"+tc.name, func(t *testing.T) {
				j := openTestJournal(t)
				owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j,
					"matrix-"+strings.ToLower(market.name)+"-"+tc.name, market.name, market.symbol)
				attempt, core := prepareCoreStrategyAttempt(t, j, plan, market.name, market.symbol, market.currency)
				brokerCalls := 0
				result, err := attempt.DispatchStrategyVerified(context.Background(), StrategyDispatchLeaseCAS{
					LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
					OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
				}, func(context.Context, *Attempt) DispatchOutcome {
					brokerCalls++
					return tc.outcome
				}, func(context.Context, string) error { return tc.verifyError })
				if err != nil && tc.verifyError == nil {
					t.Fatal(err)
				}
				if tc.verifyError != nil && !errors.Is(result.Err, tc.verifyError) {
					t.Fatalf("verify error result=%+v call-error=%v", result, err)
				}
				if brokerCalls != 1 || result.Final != tc.core {
					t.Fatalf("result=%+v broker calls=%d", result, brokerCalls)
				}
				record, err := j.LookupAttempt(context.Background(), core.attemptID)
				if err != nil || record.State != tc.core || record.BrokerOrderID != tc.durableOrderID {
					t.Fatalf("core=%+v err=%v", record, err)
				}
				lease, err := loadStrategyDispatchLease(context.Background(), j.db, plan.LeaseID)
				if err != nil || lease.State != tc.lease || lease.Disposition != tc.disposition ||
					lease.BrokerOrderID != tc.durableOrderID {
					t.Fatalf("lease=%+v err=%v", lease, err)
				}
				assertStrategyDispatchRealHolds(t, j, lease, tc.aggregate, tc.buckets)
			})
		}
	}
}

func TestDispatchStrategyVerifiedRejectsCrossAttemptBeforeBrokerPairedKRUS(t *testing.T) {
	j := openTestJournal(t)
	owner, err := j.AcquireStrategyDispatchOwner(context.Background(), "cross-attempt-owner")
	if err != nil {
		t.Fatal(err)
	}
	type prepared struct {
		plan    StrategyDispatchLeasePlan
		claimed StrategyDispatchLease
		attempt *Attempt
	}
	var pair []prepared
	for _, market := range []struct{ suffix, name, symbol, currency string }{
		{"cross-kr", "KR", "005930", "KRW"}, {"cross-us", "US", "AAPL", "USD"},
	} {
		plan := seedBoundStrategyDispatchClaimLease(t, j, owner, market.suffix,
			"acct-"+strings.ToLower(market.name), market.name, market.symbol)
		claimed, err := j.ClaimStrategyDispatchLease(context.Background(), StrategyDispatchLeaseCAS{
			LeaseID: plan.LeaseID, ExpectedRevision: 1, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken})
		if err != nil {
			t.Fatal(err)
		}
		attempt, _ := prepareCoreStrategyAttempt(t, j, plan, market.name, market.symbol, market.currency)
		pair = append(pair, prepared{plan: plan, claimed: claimed, attempt: attempt})
	}
	for index := range pair {
		peer := pair[1-index]
		brokerCalls := 0
		_, err := pair[index].attempt.DispatchStrategyVerified(context.Background(), StrategyDispatchLeaseCAS{
			LeaseID: peer.plan.LeaseID, ExpectedRevision: peer.claimed.Revision,
			OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
		}, func(context.Context, *Attempt) DispatchOutcome {
			brokerCalls++
			return DispatchOutcome{Class: DispatchAcked, BrokerOrderID: "must-not-send"}
		}, func(context.Context, string) error { return nil })
		if (!errors.Is(err, ErrStrategyDispatchLeaseUnavailable) && !errors.Is(err, ErrStrategyDispatchLeaseConsumed)) || brokerCalls != 0 {
			t.Fatalf("cross index=%d err=%v broker=%d", index, err, brokerCalls)
		}
		record, lookupErr := j.LookupAttempt(context.Background(), pair[index].attempt.ID())
		if lookupErr != nil || record.State != StateRecorded {
			t.Fatalf("cross attempt mutated=%+v err=%v", record, lookupErr)
		}
	}
}

func TestDispatchStrategyVerifiedLateFailureRollsBackCompositePairedKRUS(t *testing.T) {
	for _, market := range []struct {
		name, symbol, currency string
	}{{"KR", "005930", "KRW"}, {"US", "AAPL", "USD"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j,
				"rollback-"+strings.ToLower(market.name), market.name, market.symbol)
			attempt, core := prepareCoreStrategyAttempt(t, j, plan, market.name, market.symbol, market.currency)
			orderID := "opaque-rollback-" + strings.ToLower(market.name)
			_, err := attempt.DispatchStrategyVerified(context.Background(), StrategyDispatchLeaseCAS{
				LeaseID: plan.LeaseID, ExpectedRevision: claimed.Revision,
				OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
			}, func(context.Context, *Attempt) DispatchOutcome {
				if _, triggerErr := j.db.Exec(`CREATE TRIGGER fail_confirmed_strategy_outcome
					BEFORE INSERT ON strategy_dispatch_outcomes WHEN NEW.to_state='SUBMITTED'
					BEGIN SELECT RAISE(ABORT,'injected confirmed composite failure'); END`); triggerErr != nil {
					t.Fatal(triggerErr)
				}
				return DispatchOutcome{Class: DispatchAcked, BrokerOrderID: orderID}
			}, func(context.Context, string) error { return nil })
			if err == nil || !strings.Contains(err.Error(), "injected confirmed composite failure") {
				t.Fatalf("late failure err=%v", err)
			}
			record, lookupErr := j.LookupAttempt(context.Background(), core.attemptID)
			if lookupErr != nil || record.State != StateAcked || record.BrokerOrderID != orderID {
				t.Fatalf("durable ACK=%+v err=%v", record, lookupErr)
			}
			lease, loadErr := loadStrategyDispatchLease(context.Background(), j.db, plan.LeaseID)
			if loadErr != nil || lease.State != StrategyDispatchLeaseSubmitting || lease.Disposition != StrategyDispatchReservationReserved {
				t.Fatalf("lease=%+v err=%v", lease, loadErr)
			}
			for table, query := range map[string]string{
				"risk":     `SELECT count(*) FROM risk_bucket_orders WHERE order_id=?`,
				"campaign": `SELECT count(*) FROM campaign_order_watermarks WHERE order_id=?`,
				"strategy": `SELECT count(*) FROM strategy_execution_lineage WHERE kind='BROKER_ORDER' AND external_ref=?`,
			} {
				var count int
				if err := j.db.QueryRow(query, orderID).Scan(&count); err != nil || count != 0 {
					t.Fatalf("%s partial rows=%d err=%v", table, count, err)
				}
			}
			assertStrategyDispatchRealHolds(t, j, lease, "HELD", 5)
		})
	}
}

type confirmedCoreStrategyAttempt struct {
	attemptID string
	intentID  string
	orderID   string
}

func prepareCoreStrategyAttempt(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan,
	market, symbol, currency string,
) (*Attempt, confirmedCoreStrategyAttempt) {
	t.Helper()
	suffix := strings.TrimPrefix(plan.LeaseID, "claim-lease-")
	intentID := "core-intent-" + suffix
	attemptID := "core-attempt-" + suffix
	attempt, err := j.Prepare(context.Background(), PrepareRequest{
		Intent: Intent{ID: intentID, Market: strings.ToLower(market), TradingDay: "2026-03-30",
			AccountRef: plan.AccountRef, Symbol: symbol, Side: "BUY", OrderType: "LIMIT", TimeInForce: "DAY",
			Quantity: "10", Price: "5", Currency: currency, Source: "strategy/test", Fingerprint: "fp-" + suffix},
		Kind: KindPlace, AttemptID: attemptID, AccountRef: plan.AccountRef, DecisionID: plan.GuardianDecisionID,
		SafetyClass: SafetyClassExposureRaising, ClientOrderID: plan.OperationID,
	})
	if err != nil {
		t.Fatal(err)
	}
	return attempt, confirmedCoreStrategyAttempt{attemptID: attemptID, intentID: intentID}
}
