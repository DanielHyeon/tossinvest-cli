package journal

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestRefuseClaimedStrategyDispatchPreTransportPairedKRUS(t *testing.T) {
	for _, market := range []StrategyDispatchMarket{StrategyDispatchMarketKR, StrategyDispatchMarketUS} {
		t.Run(string(market), func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := claimedPreTransportRefusalFixture(t, j, market, "success")
			request := preTransportRefusalRequest(owner, plan, claimed.Revision, StrategyDispatchPreTransportDecisionRefused)

			terminal, err := j.RefuseClaimedStrategyDispatchPreTransport(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if terminal.State != StrategyDispatchLeaseRefused || terminal.Disposition != StrategyDispatchReservationReleased ||
				terminal.Revision != claimed.Revision+1 || terminal.RefusalCode != string(StrategyDispatchPreTransportDecisionRefused) ||
				!terminal.TransportStartedAt.IsZero() {
				t.Fatalf("terminal=%+v", terminal)
			}
			assertStrategyDispatchRealHolds(t, j, terminal, "RELEASED", 0)
			assertNoStrategyDispatchTransport(t, j)
		})
	}
}

func TestRefuseClaimedStrategyDispatchPreTransportRejectsInvalidReasonPairedKRUS(t *testing.T) {
	for _, market := range []StrategyDispatchMarket{StrategyDispatchMarketKR, StrategyDispatchMarketUS} {
		t.Run(string(market), func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := claimedPreTransportRefusalFixture(t, j, market, "reason")
			request := preTransportRefusalRequest(owner, plan, claimed.Revision, StrategyDispatchPreTransportRefusalReason("gateway_decision_refused "))
			if _, err := j.RefuseClaimedStrategyDispatchPreTransport(context.Background(), request); !errors.Is(err, ErrInvalidRequest) {
				t.Fatalf("invalid reason error=%v", err)
			}
			assertClaimedPreTransportUnchanged(t, j, claimed)
		})
	}
}

func TestRefuseClaimedStrategyDispatchPreTransportRejectsCrossScopePairedKRUS(t *testing.T) {
	for _, market := range []StrategyDispatchMarket{StrategyDispatchMarketKR, StrategyDispatchMarketUS} {
		market := market
		for _, mutate := range []struct {
			name string
			do   func(*StrategyDispatchLeasePlan)
		}{
			{name: "attempt", do: func(plan *StrategyDispatchLeasePlan) { plan.OperationID += "-peer" }},
			{name: "account", do: func(plan *StrategyDispatchLeasePlan) { plan.AccountRef += "-peer" }},
			{name: "market", do: func(plan *StrategyDispatchLeasePlan) {
				plan.Market = map[StrategyDispatchMarket]StrategyDispatchMarket{StrategyDispatchMarketKR: StrategyDispatchMarketUS, StrategyDispatchMarketUS: StrategyDispatchMarketKR}[market]
			}},
			{name: "symbol", do: func(plan *StrategyDispatchLeasePlan) { plan.Symbol += "X" }},
			{name: "candidate", do: func(plan *StrategyDispatchLeasePlan) { plan.CandidateID += "-peer" }},
			{name: "evidence", do: func(plan *StrategyDispatchLeasePlan) { plan.EvidenceDigest += "-peer" }},
			{name: "router-id", do: func(plan *StrategyDispatchLeasePlan) { plan.RouterID += "-peer" }},
			{name: "router-version", do: func(plan *StrategyDispatchLeasePlan) { plan.RouterVersion += "-peer" }},
			{name: "lane-id", do: func(plan *StrategyDispatchLeasePlan) { plan.LaneID += "-peer" }},
			{name: "lane-version", do: func(plan *StrategyDispatchLeasePlan) { plan.LaneVersion += "-peer" }},
			{name: "campaign", do: func(plan *StrategyDispatchLeasePlan) { plan.CampaignID += "-peer" }},
			{name: "leg", do: func(plan *StrategyDispatchLeasePlan) { plan.LegID += "-peer" }},
			{name: "reservation", do: func(plan *StrategyDispatchLeasePlan) { plan.RiskReservationID += "-peer" }},
			{name: "decision", do: func(plan *StrategyDispatchLeasePlan) { plan.GuardianDecisionID += "-peer" }},
			{name: "authority-revision", do: func(plan *StrategyDispatchLeasePlan) { plan.AuthorityRevision++ }},
			{name: "authority-digest", do: func(plan *StrategyDispatchLeasePlan) { plan.AuthorityDigest += "-peer" }},
			{name: "issued-at", do: func(plan *StrategyDispatchLeasePlan) { plan.IssuedAt = plan.IssuedAt.Add(time.Second) }},
			{name: "expires-at", do: func(plan *StrategyDispatchLeasePlan) { plan.ExpiresAt = plan.ExpiresAt.Add(time.Second) }},
		} {
			t.Run(string(market)+"-"+mutate.name, func(t *testing.T) {
				j := openTestJournal(t)
				owner, plan, claimed := claimedPreTransportRefusalFixture(t, j, market, "cross-"+mutate.name)
				request := preTransportRefusalRequest(owner, plan, claimed.Revision, StrategyDispatchPreTransportDecisionRefused)
				mutate.do(&request.Binding)
				if _, err := j.RefuseClaimedStrategyDispatchPreTransport(context.Background(), request); !errors.Is(err, ErrStrategyDispatchLeaseUnavailable) {
					t.Fatalf("cross-scope error=%v", err)
				}
				assertClaimedPreTransportUnchanged(t, j, claimed)
			})
		}
	}
}

func TestRefuseClaimedStrategyDispatchPreTransportRejectsOwnerABAPairedKRUS(t *testing.T) {
	for _, market := range []StrategyDispatchMarket{StrategyDispatchMarketKR, StrategyDispatchMarketUS} {
		t.Run(string(market), func(t *testing.T) {
			j := openTestJournal(t)
			ownerA, plan, claimed := claimedPreTransportRefusalFixture(t, j, market, "aba")
			if _, err := j.AcquireStrategyDispatchOwner(context.Background(), "owner-b-"+strings.ToLower(string(market))); err != nil {
				t.Fatal(err)
			}
			ownerA2, err := j.AcquireStrategyDispatchOwner(context.Background(), ownerA.OwnerInstance)
			if err != nil {
				t.Fatal(err)
			}
			if ownerA2.Epoch <= ownerA.Epoch || ownerA2.FencingToken == ownerA.FencingToken {
				t.Fatalf("ABA did not advance owner: old=%+v new=%+v", ownerA, ownerA2)
			}
			for _, reason := range []StrategyDispatchPreTransportRefusalReason{
				StrategyDispatchPreTransportDecisionRefused, StrategyDispatchPreTransportProtectionRefused,
				StrategyDispatchPreTransportReservationRefused, StrategyDispatchPreTransportAccountBaseFXRefused,
				StrategyDispatchPreTransportPolicyRefused,
			} {
				request := preTransportRefusalRequest(ownerA, plan, claimed.Revision, reason)
				if _, err := j.RefuseClaimedStrategyDispatchPreTransport(context.Background(), request); !errors.Is(err, ErrStrategyDispatchFenced) {
					t.Fatalf("reason=%s strengthened stale ABA owner: %v", reason, err)
				}
			}
			assertClaimedPreTransportUnchanged(t, j, claimed)
		})
	}
}

func TestRefuseClaimedStrategyDispatchPreTransportNormalizesSafePartialReleasePairedKRUS(t *testing.T) {
	for _, market := range []StrategyDispatchMarket{StrategyDispatchMarketKR, StrategyDispatchMarketUS} {
		t.Run(string(market)+"-bucket-released", func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := claimedPreTransportRefusalFixture(t, j, market, "cardinality")
			const priorUpdated = "2026-03-30T00:30:01Z"
			if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET state='RELEASED',held_minor='0',updated_at=?
				WHERE decision_id=? AND existing_reservation_id=? AND bucket_dimension='symbol'`, priorUpdated, plan.GuardianDecisionID, plan.RiskReservationID); err != nil {
				t.Fatal(err)
			}
			request := preTransportRefusalRequest(owner, plan, claimed.Revision, StrategyDispatchPreTransportReservationRefused)
			terminal, err := j.RefuseClaimedStrategyDispatchPreTransport(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if terminal.State != StrategyDispatchLeaseRefused || terminal.Disposition != StrategyDispatchReservationReleased {
				t.Fatalf("terminal=%+v", terminal)
			}
			var updated string
			if err := j.db.QueryRow(`SELECT updated_at FROM risk_bucket_reservations
				WHERE decision_id=? AND bucket_dimension='symbol'`, plan.GuardianDecisionID).Scan(&updated); err != nil {
				t.Fatal(err)
			}
			if updated != priorUpdated {
				t.Fatalf("pre-released bucket metadata changed: updated_at=%q want=%q", updated, priorUpdated)
			}
			assertStrategyDispatchRealHolds(t, j, terminal, "RELEASED", 0)
		})

		t.Run(string(market)+"-aggregate-released", func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := claimedPreTransportRefusalFixture(t, j, market, "aggregate-released")
			const releasedAt = "2026-03-30T00:30:02Z"
			if _, err := j.db.Exec(`UPDATE risk_reservations SET state='RELEASED',released_at=?,release_reason=? WHERE id=?`,
				releasedAt, ReleaseReasonBrokerTerminal, plan.RiskReservationID); err != nil {
				t.Fatal(err)
			}
			terminal, err := j.RefuseClaimedStrategyDispatchPreTransport(context.Background(),
				preTransportRefusalRequest(owner, plan, claimed.Revision, StrategyDispatchPreTransportDecisionRefused))
			if err != nil {
				t.Fatal(err)
			}
			var gotAt, gotReason string
			if err := j.db.QueryRow(`SELECT released_at,release_reason FROM risk_reservations WHERE id=?`,
				plan.RiskReservationID).Scan(&gotAt, &gotReason); err != nil {
				t.Fatal(err)
			}
			if gotAt != releasedAt || gotReason != ReleaseReasonBrokerTerminal {
				t.Fatalf("pre-released aggregate metadata changed: at/reason=%q/%q", gotAt, gotReason)
			}
			assertStrategyDispatchRealHolds(t, j, terminal, "RELEASED", 0)
		})
	}
}

func TestAttemptRefuseClaimedStrategyPreTransportIsOneCommitPairedKRUS(t *testing.T) {
	for _, market := range []struct {
		name, symbol, currency string
	}{{"KR", "005930", "KRW"}, {"US", "AAPL", "USD"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j,
				"attempt-refuse-"+strings.ToLower(market.name), market.name, market.symbol)
			attempt, _ := prepareCoreStrategyAttempt(t, j, plan, market.name, market.symbol, market.currency)
			terminal, err := attempt.RefuseClaimedStrategyPreTransport(context.Background(),
				preTransportRefusalRequest(owner, plan, claimed.Revision,
					StrategyDispatchPreTransportProtectionRefused), "protection unavailable")
			if err != nil {
				t.Fatal(err)
			}
			if terminal.State != StrategyDispatchLeaseRefused || terminal.Disposition != StrategyDispatchReservationReleased {
				t.Fatalf("terminal lease=%+v", terminal)
			}
			record, err := j.LookupAttempt(context.Background(), attempt.ID())
			if err != nil || record.State != StateNotDispatched || record.ReasonCode != string(StrategyDispatchPreTransportProtectionRefused) {
				t.Fatalf("core=%+v err=%v", record, err)
			}
			assertStrategyDispatchRealHolds(t, j, terminal, "RELEASED", 0)
		})
	}
}

func TestAttemptRefuseClaimedStrategyPreTransportRollsBackCoreAndClaimPairedKRUS(t *testing.T) {
	for _, market := range []struct {
		name, symbol, currency string
	}{{"KR", "005930", "KRW"}, {"US", "AAPL", "USD"}} {
		t.Run(market.name, func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := seedClaimedStrategyDispatchLease(t, j,
				"attempt-refuse-rollback-"+strings.ToLower(market.name), market.name, market.symbol)
			attempt, _ := prepareCoreStrategyAttempt(t, j, plan, market.name, market.symbol, market.currency)
			if _, err := j.db.Exec(`CREATE TRIGGER fail_composite_pretransport_refusal
				BEFORE UPDATE ON strategy_dispatch_leases WHEN OLD.lease_id='` + plan.LeaseID + `'
				BEGIN SELECT RAISE(ABORT,'injected composite refusal failure'); END`); err != nil {
				t.Fatal(err)
			}
			_, err := attempt.RefuseClaimedStrategyPreTransport(context.Background(),
				preTransportRefusalRequest(owner, plan, claimed.Revision,
					StrategyDispatchPreTransportReservationRefused), "reservation unavailable")
			if err == nil || !strings.Contains(err.Error(), "injected composite refusal failure") {
				t.Fatalf("refusal error=%v", err)
			}
			record, lookupErr := j.LookupAttempt(context.Background(), attempt.ID())
			if lookupErr != nil || record.State != StateRecorded {
				t.Fatalf("core survived partial mutation=%+v err=%v", record, lookupErr)
			}
			lease, lookupErr := j.LookupStrategyDispatchLease(context.Background(), plan.LeaseID)
			if lookupErr != nil || lease.State != StrategyDispatchLeaseClaimed || lease.Revision != claimed.Revision {
				t.Fatalf("lease survived partial mutation=%+v err=%v", lease, lookupErr)
			}
			assertStrategyDispatchRealHolds(t, j, lease, "HELD", 5)
		})
	}
}

func TestRefuseClaimedStrategyDispatchPreTransportIntegrityRollbackPairedKRUS(t *testing.T) {
	for _, market := range []StrategyDispatchMarket{StrategyDispatchMarketKR, StrategyDispatchMarketUS} {
		for _, corrupt := range []struct {
			name string
			do   func(*testing.T, *Journal, StrategyDispatchLeasePlan)
		}{
			{name: "missing-bucket", do: func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
				if _, err := j.db.Exec(`DELETE FROM risk_bucket_reservations WHERE decision_id=? AND bucket_dimension='symbol'`, plan.GuardianDecisionID); err != nil {
					t.Fatal(err)
				}
			}},
			{name: "filled-bucket", do: func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
				if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET state='FILLED',held_minor='0',filled_minor=reserved_minor
					WHERE decision_id=? AND bucket_dimension='symbol'`, plan.GuardianDecisionID); err != nil {
					t.Fatal(err)
				}
			}},
			{name: "partial-held", do: func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
				if _, err := j.db.Exec(`UPDATE risk_bucket_reservations SET held_minor='1'
					WHERE decision_id=? AND bucket_dimension='symbol'`, plan.GuardianDecisionID); err != nil {
					t.Fatal(err)
				}
			}},
			{name: "order-mapped", do: func(t *testing.T, j *Journal, plan StrategyDispatchLeasePlan) {
				now := formatJournalTime(j.clk.Now())
				if _, err := j.db.Exec(`INSERT INTO risk_bucket_orders(order_key,order_id,decision_id,order_quantity,
					quote_currency,base_currency,reservation_policy_digest,request_digest,state,created_at,updated_at)
					VALUES(?,?,?,?,?,?,?,?,'ACTIVE',?,?)`, "premap:"+plan.LeaseID, "premap-order:"+plan.LeaseID,
					plan.GuardianDecisionID, 1, "KRW", "KRW", "premap-policy", "premap-request", now, now); err != nil {
					t.Fatal(err)
				}
			}},
		} {
			t.Run(string(market)+"-"+corrupt.name, func(t *testing.T) {
				j := openTestJournal(t)
				owner, plan, claimed := claimedPreTransportRefusalFixture(t, j, market, "integrity-"+corrupt.name)
				corrupt.do(t, j, plan)
				_, err := j.RefuseClaimedStrategyDispatchPreTransport(context.Background(),
					preTransportRefusalRequest(owner, plan, claimed.Revision, StrategyDispatchPreTransportReservationRefused))
				if !errors.Is(err, ErrStrategyDispatchLeaseUnavailable) {
					t.Fatalf("integrity error=%v", err)
				}
				after, lookupErr := j.LookupStrategyDispatchLease(context.Background(), plan.LeaseID)
				if lookupErr != nil || after.State != StrategyDispatchLeaseClaimed || after.Revision != claimed.Revision {
					t.Fatalf("lease changed after integrity refusal: %+v err=%v", after, lookupErr)
				}
			})
		}

		t.Run(string(market)+"-rollback", func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := claimedPreTransportRefusalFixture(t, j, market, "rollback")
			if _, err := j.db.Exec(`CREATE TRIGGER fail_pretransport_refusal BEFORE UPDATE OF state ON strategy_dispatch_leases
				WHEN NEW.state='REFUSED' BEGIN SELECT RAISE(ABORT,'injected pretransport refusal failure'); END`); err != nil {
				t.Fatal(err)
			}
			request := preTransportRefusalRequest(owner, plan, claimed.Revision, StrategyDispatchPreTransportProtectionRefused)
			if _, err := j.RefuseClaimedStrategyDispatchPreTransport(context.Background(), request); err == nil {
				t.Fatal("injected rollback unexpectedly succeeded")
			}
			assertClaimedPreTransportUnchanged(t, j, claimed)
		})
	}
}

func TestRefuseClaimedStrategyDispatchPreTransportReplayAndConcurrencyPairedKRUS(t *testing.T) {
	for _, market := range []StrategyDispatchMarket{StrategyDispatchMarketKR, StrategyDispatchMarketUS} {
		t.Run(string(market)+"-replay", func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := claimedPreTransportRefusalFixture(t, j, market, "replay")
			request := preTransportRefusalRequest(owner, plan, claimed.Revision, StrategyDispatchPreTransportPolicyRefused)
			first, err := j.RefuseClaimedStrategyDispatchPreTransport(context.Background(), request)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := j.RefuseClaimedStrategyDispatchPreTransport(context.Background(), request); !errors.Is(err, ErrStrategyDispatchLeaseConsumed) {
				t.Fatalf("replay error=%v", err)
			}
			after, err := j.LookupStrategyDispatchLease(context.Background(), plan.LeaseID)
			if err != nil || after.State != first.State || after.Revision != first.Revision || after.RefusalCode != first.RefusalCode {
				t.Fatalf("replay changed terminal lease first=%+v after=%+v err=%v", first, after, err)
			}
		})

		t.Run(string(market)+"-concurrent", func(t *testing.T) {
			j := openTestJournal(t)
			owner, plan, claimed := claimedPreTransportRefusalFixture(t, j, market, "concurrent")
			request := preTransportRefusalRequest(owner, plan, claimed.Revision, StrategyDispatchPreTransportAccountBaseFXRefused)
			start := make(chan struct{})
			errs := make(chan error, 2)
			var wg sync.WaitGroup
			for range 2 {
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-start
					_, err := j.RefuseClaimedStrategyDispatchPreTransport(context.Background(), request)
					errs <- err
				}()
			}
			close(start)
			wg.Wait()
			close(errs)
			successes := 0
			for err := range errs {
				if err == nil {
					successes++
				}
			}
			if successes != 1 {
				t.Fatalf("concurrent successes=%d, want one", successes)
			}
			terminal, err := j.LookupStrategyDispatchLease(context.Background(), plan.LeaseID)
			if err != nil || terminal.State != StrategyDispatchLeaseRefused || terminal.Disposition != StrategyDispatchReservationReleased || terminal.Revision != claimed.Revision+1 {
				t.Fatalf("terminal=%+v err=%v", terminal, err)
			}
			assertStrategyDispatchRealHolds(t, j, terminal, "RELEASED", 0)
		})
	}
}

func claimedPreTransportRefusalFixture(t *testing.T, j *Journal, market StrategyDispatchMarket, suffix string) (StrategyDispatchOwner, StrategyDispatchLeasePlan, StrategyDispatchLease) {
	t.Helper()
	marketName := string(market)
	owner, err := j.AcquireStrategyDispatchOwner(context.Background(), "pretransport-owner-"+strings.ToLower(marketName)+"-"+suffix)
	if err != nil {
		t.Fatal(err)
	}
	account := "acct-pretransport-" + strings.ToLower(marketName) + "-" + suffix
	symbol := map[StrategyDispatchMarket]string{StrategyDispatchMarketKR: "005930", StrategyDispatchMarketUS: "AAPL"}[market]
	plan := seedBoundStrategyDispatchClaimLease(t, j, owner, "pretransport-"+strings.ToLower(marketName)+"-"+suffix, account, marketName, symbol)
	j.clk.(*clock.Fake).Set(plan.IssuedAt.Add(time.Second))
	claimed, err := j.ClaimStrategyDispatchLease(context.Background(), StrategyDispatchLeaseCAS{
		LeaseID: plan.LeaseID, ExpectedRevision: 1, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken,
	})
	if err != nil {
		t.Fatal(err)
	}
	if claimed.State != StrategyDispatchLeaseClaimed || claimed.Disposition != StrategyDispatchReservationReserved {
		t.Fatalf("claimed=%+v", claimed)
	}
	return owner, plan, claimed
}

func preTransportRefusalRequest(owner StrategyDispatchOwner, plan StrategyDispatchLeasePlan, revision uint64, reason StrategyDispatchPreTransportRefusalReason) StrategyDispatchPreTransportRefusalRequest {
	return StrategyDispatchPreTransportRefusalRequest{
		Lease:   StrategyDispatchLeaseCAS{LeaseID: plan.LeaseID, ExpectedRevision: revision, OwnerEpoch: owner.Epoch, FencingToken: owner.FencingToken},
		Binding: plan, Reason: reason,
	}
}

func assertClaimedPreTransportUnchanged(t *testing.T, j *Journal, claimed StrategyDispatchLease) {
	t.Helper()
	after, err := j.LookupStrategyDispatchLease(context.Background(), claimed.LeaseID)
	if err != nil || after.State != StrategyDispatchLeaseClaimed || after.Disposition != StrategyDispatchReservationReserved ||
		after.Revision != claimed.Revision || !after.TransportStartedAt.IsZero() || after.RefusalCode != "" {
		t.Fatalf("claimed lease changed: before=%+v after=%+v err=%v", claimed, after, err)
	}
	assertStrategyDispatchRealHolds(t, j, after, "HELD", 5)
	assertNoStrategyDispatchTransport(t, j)
}

func TestStrategyDispatchPreTransportRefusalSourceHasNoExecutionCapability(t *testing.T) {
	source, err := os.ReadFile("strategy_dispatch_refusal.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, forbidden := range []string{"internal/execgw", "internal/broker", "Submit(", "BeginStrategyDispatchSubmitting", "AcquireStrategyDispatchOwner", "IssueStrategyDispatchLease", "Recover"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("pre-transport refusal contains forbidden capability %q", forbidden)
		}
	}
	if !strings.Contains(text, "refuseClaimedStrategyDispatchSubmittingTx") {
		t.Fatal("pre-transport refusal omitted the atomic release/terminal helper")
	}
}
