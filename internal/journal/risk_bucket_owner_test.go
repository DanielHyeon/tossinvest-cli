package journal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/riskbucket"
)

func TestRiskBucketOwnerBindsAuthoritativePositionGenerationForKRAndUS(t *testing.T) {
	for _, market := range []riskbucket.Market{riskbucket.MarketKR, riskbucket.MarketUS} {
		t.Run(string(market), func(t *testing.T) {
			j, key := seedRiskBucketOwnerLifecycle(t, market, "bind-"+string(market))
			result, err := j.bindRiskBucketOwnerActual(context.Background(), key)
			if err != nil || !result.Bound || result.ActualGeneration != "1" {
				t.Fatalf("bind result=%+v err=%v", result, err)
			}
			retry, err := j.bindRiskBucketOwnerActual(context.Background(), key)
			if err != nil || !retry.AlreadyBound || retry.ActualGeneration != "1" {
				t.Fatalf("retry result=%+v err=%v", retry, err)
			}
			var actual string
			var events int
			if err := j.db.QueryRow(`SELECT actual_generation FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&actual); err != nil {
				t.Fatal(err)
			}
			if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_events WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND event_type='OWNER_BOUND'`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&events); err != nil {
				t.Fatal(err)
			}
			if actual != "1" || events != 1 {
				t.Fatalf("actual=%q events=%d", actual, events)
			}
		})
	}
}

func TestRiskBucketOwnerReleaseDerivesEveryCleanPredicateAndIsIdempotent(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "release")
	if _, err := j.bindRiskBucketOwnerActual(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	closeRiskBucketOwnerLifecycle(t, j, key, false)

	before := riskBucketOwnerLifecycleWrites(t, j, key)
	if _, err := j.releaseRiskBucketOwner(context.Background(), key); !errors.Is(err, ErrRiskBucketOwnerReleaseBlocked) {
		t.Fatalf("missing broker-zero reconciliation error=%v", err)
	}
	if after := riskBucketOwnerLifecycleWrites(t, j, key); after != before {
		t.Fatalf("refusal wrote state before=%s after=%s", before, after)
	}

	if _, err := j.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause) VALUES(?,?,?,'QUANTITY_MISMATCH','broker quantity zero','2026-03-30T00:37:00Z','2026-03-30T00:40:00Z','RECHECK_MATCHED')`, "owner-reconcile-"+key.ProspectiveGeneration, key.AccountID, key.Symbol); err != nil {
		t.Fatal(err)
	}
	recordOfficialRiskBucketBrokerZeroAt(t, j, key, "official-zero-stale-"+key.ProspectiveGeneration, time.Date(2026, 3, 30, 0, 39, 0, 0, time.UTC))
	if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET updated_at='2026-03-30T00:39:30Z' WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	refreshRiskBucketOwnerSnapshot(t, j, key, "stale-official-zero")
	before = riskBucketOwnerLifecycleWrites(t, j, key)
	if _, err := j.releaseRiskBucketOwner(context.Background(), key); !errors.Is(err, ErrRiskBucketOwnerReleaseBlocked) {
		t.Fatalf("stale broker-zero reconciliation error=%v", err)
	} else {
		var lifecycle *RiskBucketOwnerLifecycleError
		if !errors.As(err, &lifecycle) || lifecycle.BlockingField != "broker_reconciliation_stale" {
			t.Fatalf("stale reconciliation lifecycle=%+v", lifecycle)
		}
	}
	if after := riskBucketOwnerLifecycleWrites(t, j, key); after != before {
		t.Fatalf("stale refusal wrote state before=%s after=%s", before, after)
	}
	if _, err := j.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause) VALUES(?,?,?,'QUANTITY_MISMATCH','fresh broker quantity zero','2026-03-30T00:40:00Z','2026-03-30T00:42:00Z','RECHECK_MATCHED')`, "owner-reconcile-fresh-"+key.ProspectiveGeneration, key.AccountID, key.Symbol); err != nil {
		t.Fatal(err)
	}
	recordOfficialRiskBucketBrokerZeroAt(t, j, key, "official-zero-fresh-"+key.ProspectiveGeneration, time.Date(2026, 3, 30, 0, 41, 0, 0, time.UTC))
	result, err := j.releaseRiskBucketOwner(context.Background(), key)
	if err != nil || !result.Released {
		t.Fatalf("release result=%+v err=%v", result, err)
	}
	retry, err := j.releaseRiskBucketOwner(context.Background(), key)
	if err != nil || !retry.AlreadyReleased {
		t.Fatalf("retry result=%+v err=%v", retry, err)
	}
	var released string
	var events int
	if err := j.db.QueryRow(`SELECT released_at FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&released); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_events WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND event_type='OWNER_RELEASED'`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if released == "" || events != 1 {
		t.Fatalf("released_at=%q events=%d", released, events)
	}
}

func TestRiskBucketOwnerReleaseRejectsFreeformReconcileWithoutOfficialZeroObservation(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "freeform-zero")
	if _, err := j.bindRiskBucketOwnerActual(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	closeRiskBucketOwnerLifecycle(t, j, key, false)
	if _, err := j.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause) VALUES('freeform-zero',?,?,'QUANTITY_MISMATCH','trust me: broker zero','2026-03-30T00:37:00Z','2026-03-30T00:40:00Z','RECHECK_MATCHED')`, key.AccountID, key.Symbol); err != nil {
		t.Fatal(err)
	}
	before := riskBucketOwnerLifecycleWrites(t, j, key)
	_, err := j.releaseRiskBucketOwner(context.Background(), key)
	var lifecycle *RiskBucketOwnerLifecycleError
	if !errors.Is(err, ErrRiskBucketOwnerReleaseBlocked) || !errors.As(err, &lifecycle) || lifecycle.BlockingField != "broker_zero_observation" {
		t.Fatalf("freeform release error=%v lifecycle=%+v", err, lifecycle)
	}
	if after := riskBucketOwnerLifecycleWrites(t, j, key); after != before {
		t.Fatalf("freeform refusal wrote state before=%s after=%s", before, after)
	}
}

func TestRiskBucketAdjustmentReleaseRequiresExactZeroAdjustmentAndLaterOfficialRecheck(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketKR, "adjustment-zero")
	if _, err := j.bindRiskBucketOwnerActual(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	closeRiskBucketOwnerLifecycle(t, j, key, false)
	if _, err := j.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause)
		VALUES('adjustment-zero-reconcile',?,?,'QUANTITY_MISMATCH','adjustment applied',
		'2026-03-30T00:37:00Z','2026-03-30T00:40:00Z','ADJUSTMENT_APPLIED')`, key.AccountID, key.Symbol); err != nil {
		t.Fatal(err)
	}
	capability := officialZeroCapabilityForTest(t, key, "adjustment-zero", time.Date(2026, 3, 30, 0, 39, 0, 0, time.UTC))
	if _, err := j.recordRiskBucketBrokerZeroObservation(context.Background(), capability); !errors.Is(err, ErrRiskBucketOwnerReleaseBlocked) {
		t.Fatalf("adjustment release without adjustment recorded official zero err=%v", err)
	} else {
		var lifecycle *RiskBucketOwnerLifecycleError
		if !errors.As(err, &lifecycle) || lifecycle.BlockingField != "zero_adjustment" {
			t.Fatalf("missing zero adjustment lifecycle=%+v", lifecycle)
		}
	}
	var positionID string
	if err := j.db.QueryRow(`SELECT id FROM positions WHERE account_ref=? AND market=? AND symbol=? AND instance_seq=1`, key.AccountID, normaliseMarket(string(key.Market)), key.Symbol).Scan(&positionID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO position_adjustments(id,position_id,kind,expected_prev_quantity,prev_quantity,new_quantity,
		prev_avg_price,new_avg_price,broker_as_of,evidence,created_at) VALUES('adjustment-zero-row',?,'UNKNOWN','0','0','0',
		'0','0','2026-03-30T00:39:00Z','official adjustment to zero','2026-03-30T00:39:00Z')`, positionID); err != nil {
		t.Fatal(err)
	}
	if _, err := j.recordRiskBucketBrokerZeroObservation(context.Background(), capability); !errors.Is(err, ErrRiskBucketOwnerReleaseBlocked) {
		t.Fatalf("same-time adjustment observation accepted err=%v", err)
	} else {
		var lifecycle *RiskBucketOwnerLifecycleError
		if !errors.As(err, &lifecycle) || lifecycle.BlockingField != "post_adjustment_recheck" {
			t.Fatalf("same-time adjustment lifecycle=%+v", lifecycle)
		}
	}
	capability = officialZeroCapabilityForTest(t, key, "adjustment-zero-later", time.Date(2026, 3, 30, 0, 39, 30, 0, time.UTC))
	if result, err := j.recordRiskBucketBrokerZeroObservation(context.Background(), capability); err != nil || !result.Recorded {
		t.Fatalf("later official zero result=%+v err=%v", result, err)
	}
}

func TestRiskBucketBrokerZeroRejectsUnsealedAndCallerMutatedCapabilities(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "unsealed-zero")
	if _, err := j.bindRiskBucketOwnerActual(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	closeRiskBucketOwnerLifecycle(t, j, key, false)
	if _, err := j.db.Exec(`INSERT INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause)
		VALUES('unsealed-zero-reconcile',?,?,'QUANTITY_MISMATCH','caller claims zero',
		'2026-03-30T00:37:00Z','2026-03-30T00:40:00Z','RECHECK_MATCHED')`, key.AccountID, key.Symbol); err != nil {
		t.Fatal(err)
	}
	for name, capability := range map[string]riskBucketOfficialZeroCapability{
		"zero value": {},
		"unsealed caller fields": {
			owner: key, brokerAsOf: time.Date(2026, 3, 30, 0, 39, 0, 0, time.UTC),
			capabilityVersion: "caller", buildVersion: "caller", sourceVersion: "caller", payloadDigest: "caller-zero",
		},
		"mutated after seal": func() riskBucketOfficialZeroCapability {
			capability := officialZeroCapabilityForTest(t, key, "sealed-before-mutation", time.Date(2026, 3, 30, 0, 39, 0, 0, time.UTC))
			capability.payloadDigest = "caller-mutated-zero"
			return capability
		}(),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := j.recordRiskBucketBrokerZeroObservation(context.Background(), capability); !errors.Is(err, ErrRiskBucketOwnerReleaseBlocked) {
				t.Fatalf("arbitrary capability created authority err=%v", err)
			} else {
				var lifecycle *RiskBucketOwnerLifecycleError
				if !errors.As(err, &lifecycle) || lifecycle.BlockingField != "official_capability" {
					t.Fatalf("arbitrary capability lifecycle=%+v", lifecycle)
				}
			}
		})
	}
	var observations int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_broker_zero_observations`).Scan(&observations); err != nil || observations != 0 {
		t.Fatalf("arbitrary caller created observations=%d err=%v", observations, err)
	}
}

func TestRiskBucketOwnerReleasedReplayRequiresExactSealedReceiptAndEvent(t *testing.T) {
	for _, corrupt := range []struct {
		name string
		run  func(*testing.T, *Journal)
	}{
		{
			name: "missing event",
			run: func(t *testing.T, j *Journal) {
				j.db.SetMaxOpenConns(1)
				conn, err := j.db.Conn(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				defer conn.Close()
				if _, err := conn.ExecContext(context.Background(), `PRAGMA foreign_keys=OFF`); err != nil {
					t.Fatal(err)
				}
				if _, err := conn.ExecContext(context.Background(), `DROP TRIGGER risk_bucket_events_no_delete`); err != nil {
					t.Fatal(err)
				}
				if _, err := conn.ExecContext(context.Background(), `DELETE FROM risk_bucket_events WHERE event_type='OWNER_RELEASED'`); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "divergent receipt",
			run: func(t *testing.T, j *Journal) {
				if _, err := j.db.Exec(`DROP TRIGGER risk_bucket_owner_release_receipts_no_update`); err != nil {
					t.Fatal(err)
				}
				if _, err := j.db.Exec(`UPDATE risk_bucket_owner_release_receipts SET observation_digest='forged'`); err != nil {
					t.Fatal(err)
				}
			},
		},
	} {
		t.Run(corrupt.name, func(t *testing.T) {
			j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketKR, "receipt-"+corrupt.name)
			if _, err := j.bindRiskBucketOwnerActual(context.Background(), key); err != nil {
				t.Fatal(err)
			}
			closeRiskBucketOwnerLifecycle(t, j, key, true)
			if result, err := j.releaseRiskBucketOwner(context.Background(), key); err != nil || !result.Released {
				t.Fatalf("initial release result=%+v err=%v", result, err)
			}
			corrupt.run(t, j)
			_, err := j.releaseRiskBucketOwner(context.Background(), key)
			var lifecycle *RiskBucketOwnerLifecycleError
			if !errors.Is(err, ErrRiskBucketOwnerReleaseBlocked) || !errors.As(err, &lifecycle) || lifecycle.BlockingField != "release_receipt" {
				t.Fatalf("corrupt replay error=%v lifecycle=%+v", err, lifecycle)
			}
		})
	}
}

func TestRiskBucketOwnerBindRefusalWritesNothingWithoutAuthoritativeGeneration(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketKR, "bind-refusal")
	if _, err := j.db.Exec(`UPDATE position_campaigns SET actual_position_generation=NULL WHERE prospective_token=?`, key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	before := riskBucketOwnerLifecycleWrites(t, j, key)
	_, err := j.bindRiskBucketOwnerActual(context.Background(), key)
	var lifecycle *RiskBucketOwnerLifecycleError
	if !errors.Is(err, ErrRiskBucketOwnerBindBlocked) || !errors.As(err, &lifecycle) || lifecycle.BlockingField != "actual_generation" {
		t.Fatalf("bind refusal error=%v lifecycle=%+v", err, lifecycle)
	}
	if after := riskBucketOwnerLifecycleWrites(t, j, key); after != before {
		t.Fatalf("bind refusal wrote state before=%s after=%s", before, after)
	}
}

func TestRiskBucketFillHookLatchesBindGapWithoutReturningError(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "hook-refusal")
	var decisionID string
	if err := j.db.QueryRow(`SELECT decision_id FROM risk_bucket_final_decisions WHERE owner_prospective_generation=?`, key.ProspectiveGeneration).Scan(&decisionID); err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrderScope(t, j, "hook-refusal-intent", "hook-refusal-attempt", "hook-refusal-order", FillSnapshotScope{
		AccountRef: key.AccountID, Market: "us", TradingDay: "2026-03-30", Symbol: key.Symbol, Side: "BUY",
	})
	bindRiskOrderAttemptDecision(t, j, "hook-refusal-order", decisionID)
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{
		OrderID: "hook-refusal-order", DecisionID: decisionID, OrderQuantity: 10,
		ReservedMinor: riskReservedMap("50"), CreatedAt: riskFillNow,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE position_campaigns SET actual_position_generation=NULL WHERE prospective_token=?`, key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	exitRan := false
	if err := j.SetApplyHooks(ApplyHooks{
		Campaign: func(context.Context, *ApplyTx, AppliedFill) error { return nil },
		Exit: func(context.Context, *ApplyTx, AppliedFill) error {
			exitRan = true
			return nil
		},
	}); err != nil {
		t.Fatal(err)
	}
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	fill := AppliedFill{OrderID: "hook-refusal-order", AccountRef: key.AccountID, Market: "us", Symbol: key.Symbol,
		Side: "BUY", Delta: "1", CumulativeQuantity: "1", TradingDay: "2026-03-30", CommittedAt: "2026-03-30T00:32:00Z"}
	if err := j.runApplyHooks(context.Background(), tx, fill); err != nil {
		_ = tx.Rollback()
		t.Fatalf("authoritative fill hook returned semantic bind refusal: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	if !exitRan {
		t.Fatal("semantic owner-bind refusal blocked the existing exit hook")
	}
	var latched, events int
	if err := j.db.QueryRow(`SELECT unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&latched); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_events WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=? AND event_type='OWNER_BIND_REFUSED'`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&events); err != nil {
		t.Fatal(err)
	}
	if latched != 1 || events != 1 {
		t.Fatalf("latched=%d bind-refused-events=%d", latched, events)
	}
}

func TestRunApplyHooksBindsRiskBucketOwnerAfterCampaignInSameTransaction(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "hook-bind")
	var decisionID string
	if err := j.db.QueryRow(`SELECT decision_id FROM risk_bucket_final_decisions WHERE owner_prospective_generation=?`, key.ProspectiveGeneration).Scan(&decisionID); err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrderScope(t, j, "hook-bind-intent", "hook-bind-attempt", "hook-bind-order", FillSnapshotScope{
		AccountRef: key.AccountID, Market: "us", TradingDay: "2026-03-30", Symbol: key.Symbol, Side: "BUY",
	})
	bindRiskOrderAttemptDecision(t, j, "hook-bind-order", decisionID)
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{
		OrderID: "hook-bind-order", DecisionID: decisionID, OrderQuantity: 10,
		ReservedMinor: riskReservedMap("50"), CreatedAt: riskFillNow,
	}); err != nil {
		t.Fatal(err)
	}
	if err := j.SetApplyHooks(ApplyHooks{Campaign: func(context.Context, *ApplyTx, AppliedFill) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	fill := AppliedFill{OrderID: "hook-bind-order", AccountRef: key.AccountID, Market: "us", Symbol: key.Symbol,
		Side: "BUY", Delta: "1", CumulativeQuantity: "1", TradingDay: "2026-03-30", CommittedAt: "2026-03-30T00:32:00Z"}
	if err := j.runApplyHooks(context.Background(), tx, fill); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	var actual string
	if err := tx.QueryRow(`SELECT actual_generation FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&actual); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if actual != "1" {
		_ = tx.Rollback()
		t.Fatalf("same-transaction actual=%q", actual)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func TestRiskBucketOwnerReleaseBlocksTerminalButDirtyEvidence(t *testing.T) {
	tests := []struct {
		name  string
		dirty func(*testing.T, *Journal, riskbucket.OwnerKey)
		want  string
	}{
		{
			name: "protection saga",
			dirty: func(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
				_, err := j.db.Exec(`INSERT INTO protection_sagas(saga_id,account_ref,profile,market,symbol,generation,revision,state,trigger,quantity,client_order_id,updated_at) VALUES('dirty-saga',?,'default',?,?,1,1,'ACTIVE',1,1,'dirty-client','2026-03-30T00:39:00Z')`, key.AccountID, string(key.Market), key.Symbol)
				if err != nil {
					t.Fatal(err)
				}
			},
			want: "protection_saga",
		},
		{
			name: "protection mutation",
			dirty: func(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
				if _, err := j.db.Exec(`INSERT INTO protection_sagas(saga_id,account_ref,profile,market,symbol,generation,revision,state,trigger,quantity,client_order_id,updated_at) VALUES('dirty-closed-saga',?,'default',?,?,1,1,'CLOSED',1,1,'dirty-closed-client','2026-03-30T00:39:00Z')`, key.AccountID, string(key.Market), key.Symbol); err != nil {
					t.Fatal(err)
				}
				if _, err := j.db.Exec(`INSERT INTO protection_mutation_attempts(attempt_id,saga_id,generation,kind,state,serializer_version,canonical_body,created_at,updated_at) VALUES('dirty-protection-attempt','dirty-closed-saga',1,'CANCEL','DISPATCHED',1,'{}','2026-03-30T00:39:00Z','2026-03-30T00:39:00Z')`); err != nil {
					t.Fatal(err)
				}
			},
			want: "protection_order",
		},
		{
			name: "sell mutation",
			dirty: func(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
				_, err := j.db.Exec(`INSERT INTO intents(id,created_at,market,trading_day,account_ref,symbol,side,order_type,quantity,currency,source,fingerprint) VALUES('dirty-sell','2026-03-30T00:39:00Z',?,'2026-03-30',?,?,'SELL','MARKET','1','USD','exit','dirty-sell')`, normaliseMarket(string(key.Market)), key.AccountID, key.Symbol)
				if err != nil {
					t.Fatal(err)
				}
			},
			want: "sell_mutation",
		},
		{
			name: "unknown sell mutation state",
			dirty: func(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
				if _, err := j.db.Exec(`INSERT INTO intents(id,created_at,market,trading_day,account_ref,symbol,side,order_type,quantity,currency,source,fingerprint) VALUES('unknown-state-sell','2026-03-30T00:39:00Z',?,'2026-03-30',?,?,'SELL','MARKET','1','USD','exit','unknown-state-sell')`, normaliseMarket(string(key.Market)), key.AccountID, key.Symbol); err != nil {
					t.Fatal(err)
				}
				if _, err := j.db.Exec(`INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,fingerprint,recorded_at) VALUES('unknown-state-attempt','unknown-state-sell','PLACE','FUTURE_STATE',1,'unknown-state','2026-03-30T00:39:00Z')`); err != nil {
					t.Fatal(err)
				}
			},
			want: "sell_mutation",
		},
		{
			name: "unknown sell mutation kind",
			dirty: func(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
				if _, err := j.db.Exec(`INSERT INTO intents(id,created_at,market,trading_day,account_ref,symbol,side,order_type,quantity,currency,source,fingerprint) VALUES('unknown-kind-sell','2026-03-30T00:39:00Z',?,'2026-03-30',?,?,'SELL','MARKET','1','USD','exit','unknown-kind-sell')`, normaliseMarket(string(key.Market)), key.AccountID, key.Symbol); err != nil {
					t.Fatal(err)
				}
				if _, err := j.db.Exec(`INSERT INTO mutation_attempts(id,intent_id,kind,state,attempt_no,fingerprint,recorded_at) VALUES('unknown-kind-attempt','unknown-kind-sell','FUTURE_KIND','NOT_DISPATCHED',1,'unknown-kind','2026-03-30T00:39:00Z')`); err != nil {
					t.Fatal(err)
				}
			},
			want: "sell_mutation",
		},
		{
			name: "campaign claim",
			dirty: func(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
				var campaign string
				if err := j.db.QueryRow(`SELECT campaign_id FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&campaign); err != nil {
					t.Fatal(err)
				}
				if _, err := j.db.Exec(`INSERT INTO position_campaign_claims(account_ref,market,symbol,position_generation,position_version,version,prospective_token,campaign_id,actual_position_generation,created_at,updated_at) VALUES(?,?,?,0,0,1,?,?,1,'2026-03-30T00:30:00Z','2026-03-30T00:39:00Z')`, key.AccountID, normaliseMarket(string(key.Market)), key.Symbol, key.ProspectiveGeneration, campaign); err != nil {
					t.Fatal(err)
				}
			},
			want: "sell_claim",
		},
		{
			name: "unresolved fill",
			dirty: func(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
				if _, err := j.db.Exec(`INSERT INTO risk_bucket_events(event_id,account_ref,market,symbol,prospective_generation,event_type,event_digest,payload,created_at) VALUES('dirty-unresolved-fill',?,?,?,?, 'FILL_UNACCOUNTED','dirty','{}','2026-03-30T00:39:00Z')`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
					t.Fatal(err)
				}
			},
			want: "unresolved_fill",
		},
		{
			name: "unknown actual fill",
			dirty: func(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
				_, err := j.db.Exec(`UPDATE risk_bucket_owners SET unknown_actual_latched=1 WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration)
				if err != nil {
					t.Fatal(err)
				}
				refreshRiskBucketOwnerSnapshot(t, j, key, "dirty-actual")
			},
			want: "owner_latch",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "dirty-"+tc.name)
			if _, err := j.bindRiskBucketOwnerActual(context.Background(), key); err != nil {
				t.Fatal(err)
			}
			closeRiskBucketOwnerLifecycle(t, j, key, true)
			tc.dirty(t, j, key)
			before := riskBucketOwnerLifecycleWrites(t, j, key)
			_, err := j.releaseRiskBucketOwner(context.Background(), key)
			var lifecycle *RiskBucketOwnerLifecycleError
			if !errors.Is(err, ErrRiskBucketOwnerReleaseBlocked) || !errors.As(err, &lifecycle) || lifecycle.BlockingField != tc.want {
				t.Fatalf("error=%v lifecycle=%+v", err, lifecycle)
			}
			if after := riskBucketOwnerLifecycleWrites(t, j, key); after != before {
				t.Fatalf("refusal wrote state before=%s after=%s", before, after)
			}
		})
	}
}

func TestRiskBucketOwnerReleaseDoesNotUseOtherMarketReconciliation(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "market-isolation")
	if _, err := j.bindRiskBucketOwnerActual(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	closeRiskBucketOwnerLifecycle(t, j, key, false)
	if _, err := j.db.Exec(`INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at,closed_at) VALUES('kr-same-symbol',?,'kr',?,1,NULL,'CLOSED','0','0','2026-03-30T00:31:00Z','2026-03-30T00:35:00Z')`, key.AccountID, key.Symbol); err != nil {
		t.Fatal(err)
	}
	seedReleasedBrokerZeroReconciliation(t, j, key)
	result, err := j.releaseRiskBucketOwner(context.Background(), key)
	if err != nil || !result.Released {
		t.Fatalf("exact-market official observation must survive other-market same-symbol position result=%+v err=%v", result, err)
	}
}

func TestRiskBucketOwnerReleaseRaceAndRestartRemainIdempotent(t *testing.T) {
	j, key := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketKR, "race-restart")
	if _, err := j.bindRiskBucketOwnerActual(context.Background(), key); err != nil {
		t.Fatal(err)
	}
	closeRiskBucketOwnerLifecycle(t, j, key, true)
	restarted := openTestJournalAt(t, j.path)

	results := make(chan RiskBucketOwnerReleaseResult, 2)
	errorsCh := make(chan error, 2)
	var wg sync.WaitGroup
	for _, journal := range []*Journal{j, restarted} {
		wg.Add(1)
		go func(journal *Journal) {
			defer wg.Done()
			result, err := journal.releaseRiskBucketOwner(context.Background(), key)
			results <- result
			errorsCh <- err
		}(journal)
	}
	wg.Wait()
	close(results)
	close(errorsCh)
	for err := range errorsCh {
		if err != nil {
			t.Fatalf("race release error=%v", err)
		}
	}
	released, already := 0, 0
	for result := range results {
		if result.Released {
			released++
		}
		if result.AlreadyReleased {
			already++
		}
	}
	if released != 1 || already != 1 {
		t.Fatalf("released=%d already=%d", released, already)
	}
}

func TestRiskBucketLateFillCannotBindReopenedOwner(t *testing.T) {
	j, oldKey := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "late-old")
	var oldDecision string
	if err := j.db.QueryRow(`SELECT decision_id FROM risk_bucket_final_decisions WHERE owner_prospective_generation=?`, oldKey.ProspectiveGeneration).Scan(&oldDecision); err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrderScope(t, j, "late-old-intent", "late-old-attempt", "late-old-order", FillSnapshotScope{
		AccountRef: oldKey.AccountID, Market: "us", TradingDay: "2026-03-30", Symbol: oldKey.Symbol, Side: "BUY",
	})
	bindRiskOrderAttemptDecision(t, j, "late-old-order", oldDecision)
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{
		OrderID: "late-old-order", DecisionID: oldDecision, OrderQuantity: 10,
		ReservedMinor: riskReservedMap("50"), CreatedAt: riskFillNow,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.releaseRiskBucketOrder(context.Background(), RiskBucketOrderRelease{
		Owner: oldKey, DecisionID: oldDecision, OrderID: "late-old-order",
		Reason: RiskBucketReleaseCancel, ReleasedAt: riskFillNow.Add(time.Second),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE mutation_attempts SET state='FAILED_CONFIRMED' WHERE id='late-old-attempt'`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.bindRiskBucketOwnerActual(context.Background(), oldKey); err != nil {
		t.Fatal(err)
	}
	closeRiskBucketOwnerLifecycle(t, j, oldKey, true)
	if result, err := j.releaseRiskBucketOwner(context.Background(), oldKey); err != nil || !result.Released {
		t.Fatalf("old release result=%+v err=%v", result, err)
	}

	seedExistingRiskReservation(t, j, "owner-existing-late-new", oldKey.AccountID)
	plan := riskBucketAdmissionFixture(t, "owner-late-new", oldKey.AccountID, "lane-medium", "campaign-late-new", "prospective-late-new", "100", "0")
	plan.ExistingReservationID = "owner-existing-late-new"
	plan.Owner.Key.Market = riskbucket.MarketUS
	plan.Owner.Key.Symbol = "AAPL"
	plan.Admission.Policy.QuoteCurrency = "USD"
	rebindRiskBucket(t, &plan, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: "US", PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &plan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "AAPL", PolicyVersion: "policy-v1"})
	if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE mutation_attempts SET state='CONFIRMED' WHERE id='late-old-attempt'`); err != nil {
		t.Fatal(err)
	}
	if err := j.SetApplyHooks(ApplyHooks{Campaign: func(context.Context, *ApplyTx, AppliedFill) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	late := observation("late-old-order", "1")
	late.AccountRef = oldKey.AccountID
	late.Market = "us"
	late.Symbol = "AAPL"
	late.ObservedAt = "2026-03-30T00:45:00Z"
	if result, err := j.RecordFill(context.Background(), late); err != nil || !result.Changed {
		t.Fatalf("late RecordFill result=%+v err=%v", result, err)
	}
	var newActual *string
	if err := j.db.QueryRow(`SELECT actual_generation FROM risk_bucket_owners WHERE account_ref=? AND market='US' AND symbol='AAPL' AND prospective_generation=?`, oldKey.AccountID, plan.Owner.Key.ProspectiveGeneration).Scan(&newActual); err != nil {
		t.Fatal(err)
	}
	if newActual != nil {
		t.Fatalf("late predecessor fill bound reopened owner actual=%q", *newActual)
	}
	var orphanLatch, activeReconcile, unaccounted int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market='US'
		AND symbol='AAPL' AND prospective_generation=? AND latch='ORPHAN_FILL'`, oldKey.AccountID, plan.Owner.Key.ProspectiveGeneration).Scan(&orphanLatch); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM reconcile_states WHERE account_ref=? AND symbol='AAPL' AND released_at IS NULL`, oldKey.AccountID).Scan(&activeReconcile); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_events WHERE account_ref=? AND market='US' AND symbol='AAPL'
		AND prospective_generation=? AND event_type='FILL_UNACCOUNTED'`, oldKey.AccountID, plan.Owner.Key.ProspectiveGeneration).Scan(&unaccounted); err != nil {
		t.Fatal(err)
	}
	if orphanLatch != 1 || activeReconcile != 1 || unaccounted != 1 {
		t.Fatalf("late fill durable blockers orphan=%d reconcile=%d unaccounted=%d", orphanLatch, activeReconcile, unaccounted)
	}
	if _, err := j.releaseRiskBucketOwner(context.Background(), oldKey); !errors.Is(err, ErrRiskBucketOwnerReleaseBlocked) {
		t.Fatalf("released replay accepted state changed by late fill err=%v", err)
	} else {
		var lifecycle *RiskBucketOwnerLifecycleError
		if !errors.As(err, &lifecycle) || lifecycle.BlockingField != "release_receipt" {
			t.Fatalf("late fill replay lifecycle=%+v", lifecycle)
		}
	}
	seedExistingRiskReservation(t, j, "owner-existing-late-blocked", oldKey.AccountID)
	blocked := plan
	blocked.TransactionID = "owner-late-blocked-tx"
	blocked.DecisionID = "owner-late-blocked-decision"
	blocked.ExistingReservationID = "owner-existing-late-blocked"
	blocked.CreatedAt = blocked.CreatedAt.Add(time.Minute)
	if _, err := j.CommitRiskBucketAdmission(context.Background(), blocked); !errors.Is(err, ErrRiskBucketEntryBlocked) {
		t.Fatalf("late fill allowed new exposure err=%v", err)
	}
}

func TestReleasedOwnerLateFillBlocksFirstFreshAdmissionOnlyInExactMarket(t *testing.T) {
	j, oldKey := seedRiskBucketOwnerLifecycle(t, riskbucket.MarketUS, "late-before-reopen")
	var oldDecision string
	if err := j.db.QueryRow(`SELECT decision_id FROM risk_bucket_final_decisions WHERE owner_prospective_generation=?`, oldKey.ProspectiveGeneration).Scan(&oldDecision); err != nil {
		t.Fatal(err)
	}
	recordConfirmedFillOrderScope(t, j, "late-before-intent", "late-before-attempt", "late-before-order", FillSnapshotScope{
		AccountRef: oldKey.AccountID, Market: "us", TradingDay: "2026-03-30", Symbol: oldKey.Symbol, Side: "BUY",
	})
	bindRiskOrderAttemptDecision(t, j, "late-before-order", oldDecision)
	if err := j.RegisterRiskBucketOrder(context.Background(), RiskBucketOrderPlan{
		OrderID: "late-before-order", DecisionID: oldDecision, OrderQuantity: 10,
		ReservedMinor: riskReservedMap("50"), CreatedAt: riskFillNow,
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.releaseRiskBucketOrder(context.Background(), RiskBucketOrderRelease{
		Owner: oldKey, DecisionID: oldDecision, OrderID: "late-before-order",
		Reason: RiskBucketReleaseCancel, ReleasedAt: riskFillNow.Add(time.Second),
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE mutation_attempts SET state='FAILED_CONFIRMED' WHERE id='late-before-attempt'`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.bindRiskBucketOwnerActual(context.Background(), oldKey); err != nil {
		t.Fatal(err)
	}
	closeRiskBucketOwnerLifecycle(t, j, oldKey, true)
	if result, err := j.releaseRiskBucketOwner(context.Background(), oldKey); err != nil || !result.Released {
		t.Fatalf("old release result=%+v err=%v", result, err)
	}
	if _, err := j.db.Exec(`UPDATE mutation_attempts SET state='CONFIRMED' WHERE id='late-before-attempt'`); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO reconcile_states
		(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause,scope_market)
		VALUES('late-before-kr-reconcile',?,?,'QUANTITY_MISMATCH','KR already blocked',?,NULL,NULL,'KR')`,
		oldKey.AccountID, oldKey.Symbol, formatJournalTime(riskFillNow.Add(14*time.Minute))); err != nil {
		t.Fatal(err)
	}
	if err := j.SetApplyHooks(ApplyHooks{Campaign: func(context.Context, *ApplyTx, AppliedFill) error { return nil }}); err != nil {
		t.Fatal(err)
	}
	late := observation("late-before-order", "1")
	late.AccountRef, late.Market, late.Symbol = oldKey.AccountID, "us", oldKey.Symbol
	late.ObservedAt = "2026-03-30T00:45:00Z"
	if result, err := j.RecordFill(context.Background(), late); err != nil || !result.Changed {
		t.Fatalf("late RecordFill result=%+v err=%v", result, err)
	}
	var activeOwners, oldLatch, usReconcile, activeReconciles int
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_owners WHERE account_ref=? AND market='US' AND symbol=? AND released_at IS NULL`, oldKey.AccountID, oldKey.Symbol).Scan(&activeOwners); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market='US' AND symbol=? AND latch='ORPHAN_FILL'`, oldKey.AccountID, oldKey.Symbol).Scan(&oldLatch); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM reconcile_states WHERE account_ref=? AND symbol=? AND scope_market='US' AND released_at IS NULL`, oldKey.AccountID, oldKey.Symbol).Scan(&usReconcile); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM reconcile_states WHERE account_ref=? AND symbol=? AND released_at IS NULL`, oldKey.AccountID, oldKey.Symbol).Scan(&activeReconciles); err != nil {
		t.Fatal(err)
	}
	if activeOwners != 0 || oldLatch != 1 || usReconcile != 1 || activeReconciles != 2 {
		t.Fatalf("late-before-reopen facts active=%d latch=%d US reconcile=%d all reconcile=%d", activeOwners, oldLatch, usReconcile, activeReconciles)
	}
	if replay, err := j.RecordFill(context.Background(), late); err != nil || replay.Changed {
		t.Fatalf("same-market late fill replay=%+v err=%v", replay, err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM reconcile_states WHERE account_ref=? AND symbol=? AND released_at IS NULL`, oldKey.AccountID, oldKey.Symbol).Scan(&activeReconciles); err != nil || activeReconciles != 2 {
		t.Fatalf("late replay reconcile count=%d err=%v", activeReconciles, err)
	}
	if _, err := j.db.Exec(`INSERT INTO reconcile_states
		(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause,scope_market)
		VALUES('late-before-us-duplicate',?,?,'QUANTITY_MISMATCH','duplicate US',?,NULL,NULL,'US')`,
		oldKey.AccountID, oldKey.Symbol, formatJournalTime(riskFillNow.Add(16*time.Minute))); err == nil {
		t.Fatal("duplicate active US reconcile was accepted")
	}

	seedExistingRiskReservation(t, j, "late-before-us-existing", oldKey.AccountID)
	usPlan := riskBucketAdmissionFixture(t, "late-before-us", oldKey.AccountID, "lane-us", "campaign-us", "prospective-us", "100", "0")
	usPlan.ExistingReservationID = "late-before-us-existing"
	usPlan.Owner.Key.Market, usPlan.Owner.Key.Symbol = riskbucket.MarketUS, oldKey.Symbol
	usPlan.Admission.Policy.QuoteCurrency = "USD"
	rebindRiskBucket(t, &usPlan, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: "US", PolicyVersion: "policy-v1"})
	rebindRiskBucket(t, &usPlan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: oldKey.Symbol, PolicyVersion: "policy-v1"})
	if _, err := j.CommitRiskBucketAdmission(context.Background(), usPlan); !errors.Is(err, ErrRiskBucketEntryBlocked) {
		t.Fatalf("US first fresh admission escaped late-fill gate err=%v", err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_owners WHERE account_ref=? AND market='US' AND symbol=? AND released_at IS NULL`, oldKey.AccountID, oldKey.Symbol).Scan(&activeOwners); err != nil || activeOwners != 0 {
		t.Fatalf("blocked US admission inserted owner=%d err=%v", activeOwners, err)
	}

	seedExistingRiskReservation(t, j, "late-before-kr-existing", oldKey.AccountID)
	krPlan := riskBucketAdmissionFixture(t, "late-before-kr", oldKey.AccountID, "lane-kr", "campaign-kr", "prospective-kr", "100", "0")
	krPlan.ExistingReservationID = "late-before-kr-existing"
	krPlan.Owner.Key.Market, krPlan.Owner.Key.Symbol = riskbucket.MarketKR, oldKey.Symbol
	rebindRiskBucket(t, &krPlan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: oldKey.Symbol, PolicyVersion: "policy-v1"})
	for i, bucket := range krPlan.Admission.Buckets {
		key := bucket.Key
		key.PolicyVersion = "policy-kr-isolation-v1"
		rebindRiskBucket(t, &krPlan, i, key)
	}
	if _, err := j.CommitRiskBucketAdmission(context.Background(), krPlan); !errors.Is(err, ErrRiskBucketEntryBlocked) {
		t.Fatalf("KR active reconcile did not block KR admission err=%v", err)
	}
	if released, ok, err := j.ReleaseReconcile(context.Background(), ReleaseReconcileRequest{
		AccountRef: oldKey.AccountID, Symbol: oldKey.Symbol, ScopeMarket: "KR",
		Cause: ReconcileReleaseOperator, Evidence: "KR independently reconciled",
	}); err != nil || !ok || released.ScopeMarket != "KR" {
		t.Fatalf("KR scoped release state=%+v released=%v err=%v", released, ok, err)
	}
	if result, err := j.CommitRiskBucketAdmission(context.Background(), krPlan); err != nil || result.DecisionID != krPlan.DecisionID {
		t.Fatalf("US late fill contaminated KR admission result=%+v err=%v", result, err)
	}
}

func seedRiskBucketOwnerLifecycle(t *testing.T, market riskbucket.Market, suffix string) (*Journal, riskbucket.OwnerKey) {
	t.Helper()
	j := openTestJournal(t)
	existing := "owner-existing-" + suffix
	seedExistingRiskReservation(t, j, existing, "acct-owner")
	plan := riskBucketAdmissionFixture(t, "owner-"+suffix, "acct-owner", "lane-short", "campaign-owner-"+suffix, "prospective-owner-"+suffix, "100", "0")
	plan.ExistingReservationID = existing
	plan.Owner.Key.Market = market
	if market == riskbucket.MarketUS {
		plan.Owner.Key.Symbol = "AAPL"
		plan.Admission.Policy.QuoteCurrency = "USD"
		rebindRiskBucket(t, &plan, 1, riskbucket.BucketKey{Dimension: riskbucket.DimensionMarket, Value: "US", PolicyVersion: "policy-v1"})
		rebindRiskBucket(t, &plan, 4, riskbucket.BucketKey{Dimension: riskbucket.DimensionSymbol, Value: "AAPL", PolicyVersion: "policy-v1"})
	}
	if _, err := j.CommitRiskBucketAdmission(context.Background(), plan); err != nil {
		t.Fatal(err)
	}
	var legacyDecision string
	if err := j.db.QueryRow(`SELECT legacy.decision_id FROM risk_bucket_final_decisions d JOIN risk_reservations legacy ON legacy.id=d.existing_reservation_id WHERE d.decision_id=?`, plan.DecisionID).Scan(&legacyDecision); err != nil {
		t.Fatal(err)
	}
	marketText := normaliseMarket(string(market))
	positionID := "position-owner-" + suffix
	if _, err := j.db.Exec(`INSERT INTO positions(id,account_ref,market,symbol,instance_seq,entry_decision_id,state,quantity,avg_price,opened_at) VALUES(?,?,?,?,1,?,'OPEN','1','100','2026-03-30T00:31:00Z')`, positionID, plan.Owner.Key.AccountID, marketText, plan.Owner.Key.Symbol, legacyDecision); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO position_campaigns(id,account_ref,market,symbol,lane_id,lane_version,decision_id,evidence_digest,expected_position_generation,expected_position_version,prospective_token,actual_position_generation,state,version,entry_blocked,created_at,updated_at) VALUES(?,?,?,?,?,'v1',?,'evidence',0,0,?,1,'ACTIVE',1,0,'2026-03-30T00:30:00Z','2026-03-30T00:31:00Z')`, plan.Owner.CampaignID, plan.Owner.Key.AccountID, marketText, plan.Owner.Key.Symbol, plan.Owner.LaneID, legacyDecision, plan.Owner.Key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`INSERT INTO position_campaign_claims(account_ref,market,symbol,position_generation,position_version,version,prospective_token,campaign_id,actual_position_generation,created_at,updated_at) VALUES(?,?,?,0,0,1,?,?,1,'2026-03-30T00:30:00Z','2026-03-30T00:31:00Z')`, plan.Owner.Key.AccountID, marketText, plan.Owner.Key.Symbol, plan.Owner.Key.ProspectiveGeneration, plan.Owner.CampaignID); err != nil {
		t.Fatal(err)
	}
	return j, plan.Owner.Key
}

func closeRiskBucketOwnerLifecycle(t *testing.T, j *Journal, key riskbucket.OwnerKey, reconcile bool) {
	t.Helper()
	if _, err := j.db.Exec(`UPDATE positions SET state='CLOSED',quantity='0',closed_at='2026-03-30T00:35:00Z' WHERE account_ref=? AND market=? AND symbol=? AND instance_seq=1`, key.AccountID, normaliseMarket(string(key.Market)), key.Symbol); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE position_campaigns SET state='CLOSED',entry_blocked=1,version=version+1,updated_at='2026-03-30T00:35:00Z' WHERE id=(SELECT campaign_id FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?)`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`DELETE FROM position_campaign_claims WHERE campaign_id=(SELECT campaign_id FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?)`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE risk_reservations SET state='RELEASED',released_at='2026-03-30T00:36:00Z',release_reason='BROKER_TERMINAL' WHERE id IN (SELECT existing_reservation_id FROM risk_bucket_final_decisions WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?)`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET held_minor='0',state='FILLED',updated_at='2026-03-30T00:36:00Z' WHERE account_ref=? AND market=? AND symbol=? AND owner_prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration); err != nil {
		t.Fatal(err)
	}
	refreshRiskBucketOwnerSnapshot(t, j, key, "closed")
	if reconcile {
		seedReleasedBrokerZeroReconciliation(t, j, key)
	}
}

func seedReleasedBrokerZeroReconciliation(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
	t.Helper()
	id := "owner-reconcile-" + key.ProspectiveGeneration
	if _, err := j.db.Exec(`INSERT OR IGNORE INTO reconcile_states(id,account_ref,symbol,cause,evidence,entered_at,released_at,release_cause) VALUES(?,?,?,'QUANTITY_MISMATCH','broker quantity zero','2026-03-30T00:37:00Z','2026-03-30T00:40:00Z','RECHECK_MATCHED')`, id, key.AccountID, key.Symbol); err != nil {
		t.Fatal(err)
	}
	recordOfficialRiskBucketBrokerZero(t, j, key)
}

func recordOfficialRiskBucketBrokerZero(t *testing.T, j *Journal, key riskbucket.OwnerKey) {
	t.Helper()
	recordOfficialRiskBucketBrokerZeroAt(t, j, key, key.ProspectiveGeneration, time.Date(2026, 3, 30, 0, 39, 0, 0, time.UTC))
}

func recordOfficialRiskBucketBrokerZeroAt(t *testing.T, j *Journal, key riskbucket.OwnerKey, authorityID string, brokerAsOf time.Time) {
	t.Helper()
	result, err := j.recordRiskBucketBrokerZeroObservation(context.Background(), officialZeroCapabilityForTest(t, key, authorityID, brokerAsOf))
	if err != nil || (!result.Recorded && !result.Idempotent) {
		t.Fatalf("record official broker zero result=%+v err=%v", result, err)
	}
}

func officialZeroCapabilityForTest(t *testing.T, key riskbucket.OwnerKey, authorityID string, brokerAsOf time.Time) riskBucketOfficialZeroCapability {
	t.Helper()
	capability := riskBucketOfficialZeroCapability{
		owner: key, brokerAsOf: brokerAsOf,
		capabilityVersion: "official-account-v1", buildVersion: "test-build-v1",
		sourceVersion: "toss-open-api-v1", payloadDigest: "payload-zero-" + authorityID,
	}
	seal, err := capability.expectedSeal()
	if err != nil {
		t.Fatal(err)
	}
	capability.seal = seal
	return capability
}

func refreshRiskBucketOwnerSnapshot(t *testing.T, j *Journal, key riskbucket.OwnerKey, suffix string) {
	t.Helper()
	tx, err := j.db.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := j.recordRiskBucketStateTx(context.Background(), tx, key, "TEST_REFRESH", "test-refresh-"+suffix+"-"+key.ProspectiveGeneration, suffix, "2026-03-30T00:36:00Z"); err != nil {
		_ = tx.Rollback()
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func riskBucketOwnerLifecycleWrites(t *testing.T, j *Journal, key riskbucket.OwnerKey) string {
	t.Helper()
	var released *string
	var snapshots, events int
	if err := j.db.QueryRow(`SELECT released_at FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&released); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_state_snapshots WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&snapshots); err != nil {
		t.Fatal(err)
	}
	if err := j.db.QueryRow(`SELECT count(*) FROM risk_bucket_events WHERE account_ref=? AND market=? AND symbol=? AND prospective_generation=?`, key.AccountID, string(key.Market), key.Symbol, key.ProspectiveGeneration).Scan(&events); err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%v/%d/%d", released, snapshots, events)
}
