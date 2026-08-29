# a065 campaign / a066 q_final atomic ordering
## Decision scope

This document freezes the missing production ordering contract for both KR and US. It adds no Go
surface, schema, runtime activation, broker capability, Guardian permission or operating setting.
The existing functions are unchanged, so a Function Logic Map is not applicable to this mapping
checkpoint.

## Hard constraints from current HEAD

1. `Journal.CreatePositionCampaign` accepts a prospective token only inside the authoritative
   position-generation/version CAS. Before it inserts a campaign, it requires an existing
   exposure-raising decision and matching immutable `strategy_decision_lineage`.
2. `RecordQFinalDecisionAndReserveWithRecollection` creates the exposure-raising q_final Guardian
   decision, aggregate reservation, exactly five monetary reservations and risk-bucket owner in
   one transaction, but its admission input already contains the campaign ID and prospective token.
3. A router `PositionGeneration` is an expected position projection generation, not an a065
   prospective token. A digest or string conversion cannot replace the journal CAS.
4. A decision without a v25 dispatch lease cannot reach the production Gateway, but leaving an
   independently reusable decision between campaign and risk transactions would still create an
   unnecessary recovery and capacity-leak state.

## Frozen target transaction

The production integration SHALL use one journal transaction for the first leg of a new campaign:

```text
BEGIN IMMEDIATE
  validate current position generation/version and absence of an active scope claim
  validate complete sealed strategyflow lineage + exact execution terms
  calculate/revalidate q_final against current five-bucket snapshot references
  insert q_final Guardian decision + aggregate HELD reservation
  insert immutable strategy decision lineage for that exact decision
  insert position campaign + prospective token + scope claim
  insert risk-bucket owner using the exact same prospective token/lane/campaign
  insert exactly five monetary HELD reservations
  insert campaign leg 1 bound to q_final quantity and decision
COMMIT
```

The transaction result is opaque and carries the exact decision, campaign, leg, owner and
reservation identities. It is not a dispatch lease and grants no broker capability. A later central
owner transaction may issue a v25 lease only after activation/calendar/protection/reconciliation and
all current authorities are loaded.

For an existing active campaign/scale-in, the transaction SHALL reload the current campaign and
owner, require the same prospective token/lane/campaign, validate the next leg/version and append the
new q_final decision/reservations/leg atomically. It SHALL NOT create or replace a scope claim.

## Refusal and recovery matrix

| Condition | Durable result | Broker calls |
| --- | --- | --- |
| position generation/version stale | no inserts; typed generation refusal | 0 |
| active competing campaign/owner | no inserts; owner conflict | 0 |
| campaign/owner token mismatch | no inserts; reconstruction mismatch | 0 |
| missing/tampered execution terms or lineage | no inserts; lineage refusal | 0 |
| stale/missing/duplicate/cross-market bucket authority | no inserts; bucket refusal | 0 |
| q_final zero | no decision/campaign/leg/reservation | 0 |
| crash before commit | rollback all rows | 0 |
| crash after commit but before dispatch lease | complete durable plan remains HELD and unsubmitable; restart discovers it by exact IDs | 0 |
| exact idempotent retry | return the original opaque receipt; create no second row | 0 |
| same transaction key with different preimage | replay mismatch; preserve original rows | 0 |

## KR/US concurrency rule

KR and US may evaluate and precheck concurrently. Their final journal commits are serialized by the
single SQLite writer and account-wide reservation version, not by a KR-before-US dependency. A KR
generation conflict refuses only the KR transaction; a US FX or bucket refusal refuses only the US
transaction. Neither market can supply the other market's campaign claim, owner, bucket references,
Guardian currency authority or dispatch lease.

## Required RED tests before implementation

- simultaneous KR and US first-leg requests both complete when their independent authorities are
  valid, regardless of commit order;
- competing same `(account, market, symbol)` first-leg requests have exactly one winner;
- all nine inserted authority families roll back together at every injected statement boundary;
- first-leg retry returns the exact original receipt, while divergent retry is rejected;
- scale-in reuses only the exact current campaign/owner token and advances one leg/version;
- a committed plan without a dispatch lease produces zero Gateway/broker calls across restart;
- KR/US token, campaign, bucket, currency and authority cross-substitution always fails closed.
