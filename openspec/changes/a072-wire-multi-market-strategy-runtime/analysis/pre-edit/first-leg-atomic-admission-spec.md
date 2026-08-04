# Spec: KR/US q_final Campaign First-Leg Atomic Admission

**Author:** Codex (TossOS Manager context)
**Date:** 2026-08-04
**Status:** Approved
**Reviewers:** user KR/US simultaneous-delivery decision; a072 proposal-freeze review
**Related specs:** a065 position campaign, a066 multi-horizon risk buckets, a072 strategy runtime

## Context

The six KR/US lanes can now preserve sealed entry, stop and target execution terms, and q_final can
atomically persist a Guardian decision, one aggregate HELD reservation, one risk owner and exactly
five monetary bucket reservations. `CreatePositionCampaign`, however, requires that decision and its
strategy lineage to exist before it can mint a prospective campaign claim. The q_final transaction
requires the prospective token before it can acquire the risk owner. Calling the two existing APIs in
sequence therefore creates an ordering cycle and a crash window.

The approved a072 ordering contract resolves this with one journal-owned first-leg transaction. KR
and US may prepare concurrently; SQLite serializes only their commits. The transaction creates a
durable but unsubmitable plan. It has no dispatch-lease, Gateway, broker, activation or operating-
setting capability, so completing it cannot place an order or enable either market.

## Functional Requirements

- FR-1: The journal MUST mint the prospective campaign token itself and MUST NOT accept a caller-
  supplied prospective token for a first-leg admission.
- FR-2: One transaction MUST validate and persist the q_final Guardian decision, aggregate HELD
  reservation, immutable strategy decision/attempt lineage, position campaign, prospective scope
  claim, exact risk owner, five monetary HELD reservations and campaign leg 1.
- FR-3: The first campaign leg MUST be bound to the same q_final decision through the campaign
  decision identity and MUST store requested/residual quantity equal to q_final.
- FR-4: The transaction MUST validate the current position generation/version and absence of a
  competing active campaign/claim before inserting any durable authority.
- FR-5: Exact retry with the same transaction, decision, campaign, lineage and first-leg preimage
  MUST return the original opaque receipt without inserting or repairing rows.
- FR-6: Reuse of any transaction, decision, campaign, command, attempt or plan identity with a
  different preimage MUST fail closed and MUST preserve the original rows.
- FR-7: A KR conflict/refusal MUST NOT create or alter US campaign/risk authority, and a US
  conflict/refusal MUST NOT create or alter KR campaign/risk authority.
- FR-7A: KR/KRW and US/USD support MUST be designed, tested, implemented and integrated in the same
  delivery wave. A passing implementation for only one market MUST remain incomplete and MUST NOT
  be treated as a prerequisite for beginning the peer market.
- FR-8: A committed first-leg plan MUST remain unsubmitable until a later, separately authorized
  fenced dispatch lease exists; this API MUST NOT create a dispatch lease or broker attempt.
- FR-9: Missing, stale, tampered, duplicate or cross-market strategy/risk/FX evidence MUST cause
  zero first-leg rows to commit.
- FR-10: The implementation MUST append schema v26 and MUST NOT edit or rebuild released schema
  v20 through v25 tables, triggers or historical rows. Existing standalone q_final and
  PositionCampaign validation and replay contracts MUST remain at least as strict.
- FR-11: Schema v26 MUST add one immutable `strategy_first_leg_bindings` companion that durably
  binds the q_final decision and aggregate reservation to the exact strategy decision/attempt,
  fixed production router identity/release, campaign/prospective claim, risk owner, first campaign
  leg and journal-minted token. A separate
  additive lease-insert guard MUST reject any future v25 dispatch lease whose campaign, leg plan,
  router, lane, market, symbol, aggregate reservation or Guardian decision does not match that
  binding, or whose operation identity is not the bound attempt's exact client order identity.
- FR-12: Fresh KR and US admission MUST accept only the exact fixed `strategyrouter.RouterID` and
  `strategyrouter.RouterRelease`. Missing, forged or stale-release router values MUST fail before
  opening the journal transaction and MUST create zero rows.

## Non-Functional Requirements

- NFR-1: Atomicity MUST be enforced by one SQLite transaction; every injected failure before commit
  MUST leave zero rows for the new decision, reservations, lineage, campaign, claim, owner and leg.
- NFR-2: The transaction MUST perform zero network, Gateway, broker, scheduler, toggle or activation
  calls.
- NFR-3: The prospective token MUST contain at least 256 bits of cryptographic randomness and be
  bounded/canonical before persistence.
- NFR-4: The transaction work MUST be bounded to one decision, one aggregate reservation, one
  campaign/claim, one first leg, one owner, and exactly five risk buckets; no unbounded external loop
  is permitted.
- NFR-5: Existing risk-reducing, protection, fill, reconciliation and exit paths MUST remain
  unchanged.

## Acceptance Criteria

### AC-1: KR first leg commits atomically (FR-1, FR-2, FR-3, FR-4)
Given a valid KR/KRW q_final request, sealed strategy lineage, current flat position generation and five fresh KR bucket authorities
When the journal admits campaign leg 1
Then exactly one decision, aggregate HELD reservation, strategy decision/attempt lineage, campaign, claim, risk owner and leg are present
And exactly five monetary reservations are HELD
And the leg requested/residual quantity equals q_final.

### AC-2: KR and US progress in the same wave (FR-2, FR-7, FR-7A)
Given independent valid KR/KRW and US/USD first-leg requests for the same base-currency account and different market symbols
When both admissions start concurrently and the account-wide CAS loser performs bounded fresh recollection
Then both complete with their own market/currency/token/campaign/owner authority
And neither request consumes or substitutes the peer market's evidence
And the feature is not accepted when only one market's case passes.

### AC-3: Competing owner has one winner (FR-4, FR-7)
Given two valid first-leg requests for the same account, market and symbol generation
When they race to commit
Then exactly one campaign/claim/owner is committed
And the loser returns a typed generation or owner conflict with no decision or reservation rows.

### AC-4: Statement failure rolls back all families (FR-2, NFR-1)
Given a valid first-leg request and an injected failure at each durable write boundary
When the transaction executes
Then the transaction returns an error
And no partial decision, reservation, lineage, campaign, claim, owner or leg remains.

### AC-5: Exact retry returns the original receipt (FR-5)
Given a committed first-leg admission
When the identical request is retried after restart
Then the original decision, campaign, leg, owner and reservation identities are returned
And row counts and the journal-minted prospective token are unchanged.

### AC-6: Divergent retry fails closed (FR-6)
Given a committed first-leg admission
When the same transaction/campaign/command identity is retried with changed lineage, quantity, plan or position generation
Then replay mismatch or command conflict is returned
And all original rows remain byte-for-byte authoritative.

### AC-7: Committed plan cannot dispatch (FR-8, FR-11, NFR-2)
Given a committed first-leg admission with no dispatch lease
When restart discovery and Gateway-facing reads run
Then broker/Gateway request count is zero
And no synthetic lease, activation or LIVE approval exists.

### AC-8: Cross-market authority substitution is refused (FR-7, FR-9)
Given a valid KR request and a valid US request
When either request substitutes the peer market's token, campaign, bucket, currency, strategy or FX authority
Then the affected request commits zero first-leg rows
And the peer market's durable state is unchanged.

### AC-9: Entropy failure creates no authority (FR-1, NFR-1, NFR-3)
Given a token entropy source that returns an error
When a fresh first-leg admission attempts to mint a prospective token
Then the admission fails before the first insert
And all first-leg authority row counts remain zero.

### AC-10: Additive v26 migration preserves released authority (FR-10, FR-11, NFR-5)
Given a valid v25 journal containing representative KR and US authority rows
When schema v26 is applied
Then every v25 row remains byte-for-byte unchanged and the new binding table and lease guard exist
And an injected v26 migration failure rolls back the entire step to v25
And a v25 build refuses a v26 journal with `ErrSchemaTooNew`
And the existing q_final, campaign, fill, reconciliation, protection and exit suites retain behavior.

### AC-11: Fixed router provenance is fail-closed (FR-6, FR-11, FR-12)
Given fresh valid KR and US first-leg inputs
When either input omits or substitutes the fixed production router identity or release
Then `ErrInvalidRequest` is returned before a transaction opens
And no first-leg binding, reservation, campaign, owner, leg or lease row is created.

## Edge Cases

- EC-1: Empty/whitespace campaign, command, plan or attempt identity -> `ErrInvalidRequest`; zero writes.
- EC-2: Negative expected position generation/version -> `ErrInvalidRequest`; zero writes.
- EC-3: Open/legacy-unversioned position or active campaign claim -> `ErrGenerationConflict`; zero writes.
- EC-4: q_final is zero or decision quantity differs from calculated q_final -> risk refusal or snapshot mismatch; zero writes.
- EC-5: Entropy reader fails or returns fewer than 32 bytes -> token-mint error; zero writes.
- EC-6: Prospective-token uniqueness collision -> typed generation conflict; no partial rows.
- EC-7: Context cancellation or database failure before commit -> rollback all authority families.
- EC-8: Crash after commit but before lease issuance -> complete HELD plan remains discoverable but unsubmitable.
- EC-9: Existing q_final rows with missing/divergent campaign, claim, lineage or leg -> reconstruction/replay mismatch; do not repair.
- EC-10: More or fewer than five ordered bucket dimensions -> snapshot mismatch; zero writes.
- EC-11: Raw binding insert with a mismatched decision, aggregate reservation, attempt, campaign,
  claim, owner, first-leg plan or token -> SQLite abort; no binding is created.
- EC-12: Raw dispatch-lease insert without the exact v26 first-leg binding or with a different leg
  plan, router pair or client operation identity -> SQLite abort before any Gateway capability can observe it.
- EC-13: Fresh first-leg request with a well-formed but non-production router ID/release ->
  `ErrInvalidRequest`; zero writes for both KR and US.

## API Contracts

This is an internal Go journal API; no HTTP endpoint or operator mutation surface is added.
For schema-validation notation only, its logical operation is written as
`POST /internal-go/journal/q-final-campaign-first-leg`; this path MUST NOT be registered by an HTTP router.

```typescript
interface QFinalCampaignFirstLegRequest {
  issue: QFinalIssueRequest;          // exact Guardian decision + aggregate reservation
  strategy: StrategyPlanRequest;      // sealed decision/attempt lineage; no dispatch execution link
  routerId: "kr-us-horizon-router";
  routerVersion: "a070-kr-us-horizon-router-v1";
  campaign: {
    campaignId: string;
    expectedPositionGeneration: int64;
    expectedPositionVersion: int64;
    createCommandKey: string;
    firstLegCommandKey: string;
    firstLegPlanId: string;
  };
  // prospectiveToken is intentionally absent; journal mints it.
}

interface QFinalCampaignFirstLegReceipt {
  decisionId: string;
  campaignId: string;
  legSequence: 1;
  qFinal: uint64;
  aggregateReservationId: string;
  bucketReservationIds: string[5];
  idempotent: boolean;
  // no dispatch lease, broker order, activation or LIVE approval field
}
```

Stable errors include `ErrInvalidRequest`, `ErrGenerationConflict`,
`ErrRiskBucketOwnerConflict`, `ErrRiskBucketSnapshotMismatch`,
`ErrRiskBucketReplayMismatch`, `ErrCampaignCommandConflict` and the original typed risk refusal.

## Data Models

Schema v26 is additive. It preserves every v20-v25 table and row, then adds a companion binding and
an additional lease-insert guard. The first-leg transaction writes the existing families and this
new immutable binding in one commit.

| Entity/table | Fields used | Constraints in this feature |
| --- | --- | --- |
| `decisions` | id, account, RiskIntent, nonce, issued/expires | one exact exposure-raising q_final decision |
| `risk_reservations` | aggregate reservation identity/state | exactly one HELD reservation matching q_final |
| `strategy_decision_lineage` | market/symbol/lane/evidence/prices/quantity/digests | immutable and exact-bound to RiskIntent |
| `strategy_attempt_lineage` | attempt, decision, Guardian, manifest, revision | PLANNED only; no execution link |
| `position_campaigns` | campaign/scope/lane/decision/evidence/token/version | first-leg campaign starts PLANNED, version 2 |
| `position_campaign_claims` | account/market/symbol/generation/version/token/campaign | one active scope claim |
| `campaign_legs` | campaign, sequence, plan, requested/residual quantity | sequence 1, q_final quantity, PLANNED |
| `campaign_commands/events` | CREATE and PLAN_LEG digests/results | append-only exact replay evidence |
| `risk_bucket_owners` | scope/token/lane/campaign | exact same journal-minted token |
| `risk_bucket_final_decisions` | transaction/decision/q values/owner/snapshots | exact combined q_final authority |
| `risk_bucket_reservations` | five dimensions and HELD amounts | exactly horizon/market/strategy/sector/symbol |
| `risk_bucket_state_snapshots/events` | state digest and admission event | append-only deterministic replay state |
| `strategy_first_leg_bindings` (v26) | decision, aggregate reservation, strategy decision/attempt, fixed router ID/release, campaign, leg sequence/plan, owner generation, token, lane, scope, digest | immutable exact cross-family authority; one row per q_final decision/campaign/first leg |

## Out of Scope

- OS-1: Existing-campaign scale-in admission; it requires a separate leg-version contract after the
  first-leg transaction is GREEN.
- OS-2: Production snapshot, price, fee, FX, account-state or exposure loaders; this transaction
  consumes already sealed inputs and does not mint external facts.
- OS-3: Dispatch market authority, v25 lease issuance/claim, Gateway call or broker outcome.
- OS-4: KR/US lane, scheduler, autostart, automation or LIVE approval activation.
- OS-5: Editing or rebuilding released v20-v25 tables/triggers, or backfilling historical rows. The
  only schema change is the additive v26 companion and lease-insert guard.
- OS-6: Operator/API UI changes and production deployment wiring; those remain a072 integration tasks.
