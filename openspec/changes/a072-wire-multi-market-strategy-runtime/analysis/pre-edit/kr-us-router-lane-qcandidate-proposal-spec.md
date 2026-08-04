# KR/US production router/lane q_candidate proposal contract

## Scope

This pre-edit contract closes the ordering ambiguity between the six pure strategy lanes and a066.
It applies to KR and US in one implementation wave. A KR-only or US-only implementation, test result,
or production assembly is incomplete and cannot unblock the peer market.

No market activation, LIVE approval, order transport, journal mutation, signer, policy writer or
operating-toggle capability is added by this contract.

## Problem

The current lane evaluators require a sealed a066 `RiskCap` before they return `Quantity`, while a066
defines that input quantity as the lane's canonical `q_candidate` and must calculate `q_final` from it.
Using the existing final lane outcome as a066 input therefore creates a cycle:

```text
lane quantity requires sealed a066 q_final
                 ^                 |
                 |                 v
       a066 q_final requires lane q_candidate
```

The cycle MUST NOT be broken with an unlimited/fabricated cap, a caller-declared `q_final`, or a
test-only authority constructor. `q_final` is downstream risk authority and cannot feed backward into
the proposal that it constrains.

## Two-stage decision contract

### Stage A — pure lane proposal

For continuation, reversal and weekly-value, the KR and US evaluators SHALL expose one sealed,
mutation-free proposal path. It SHALL:

1. validate the approved candidate, exact router owner scope, market lifecycle record, lane binding,
   market-specific evidence/config, campaign plan, current leg, stop authority, invalidation state,
   freshness and execution terms;
2. calculate canonical `q_candidate` as the current immutable planned-leg remaining, before any a066
   cap is applied;
3. preserve exact entry, non-retreating effective stop, target, price provenance, policy identity and
   complete candidate→router→lane→campaign/leg lineage;
4. perform no Guardian call, bucket admission, journal mutation, Gateway call or broker request; and
5. return a typed refusal if any required source is missing, stale, unsealed, cross-market or
   inconsistent.

Campaign-risk latches and known filled/held campaign state remain Stage-A inputs because a latched or
already exhausted campaign must not become a proposal. The proposed quantity is nevertheless only a
ceiling; it is not an admission, reservation or executable order quantity.

### Stage B — authoritative q_final admission

The existing account-base Guardian and five-bucket authority consume the sealed Stage-A result as
`q_candidate`. They SHALL calculate:

```text
q_final = min(q_candidate, account-base Guardian cap,
              horizon cap, market cap, strategy cap, sector cap, symbol cap)
```

Only the journal-owned v26 first-leg transaction may atomically issue the Guardian decision and bind
`q_final`, prospective campaign owner, aggregate/five HELD reservations and leg 1. `q_final == 0` is a
typed refusal and produces no dispatch lease or transport authority.

After Stage B, execution terms SHALL be resealed or projected with quantity exactly equal to
`q_final`; all other Stage-A price, policy and lineage identities must remain byte-exact. The final
quantity MUST NOT exceed or alter `q_candidate`, and Stage B MUST NOT cause Stage A to be re-evaluated
with a forged cap.

## Production authority sources

The production adapter SHALL derive router and lane input only from package-owned, read-only
authority loaders:

- exact scheduler activation/calendar observation already loaded for the same market and frozen time;
- exact approved candidate and threshold/evidence digest already loaded for that market;
- journal-backed owner/campaign/leg state read through `mode=ro` / `query_only`, with no caller-minted
  owner snapshot or prospective token;
- digest-pinned, owner-only, Ed25519-signed market lane manifest containing the exact three lane
  activation/config/evidence bindings for that market; and
- for weekly-value, a durable exact reservation record. A fresh in-memory reservation synthesized on
  restart is forbidden.

The manifest path is closed to `strategy-lane-authority-KR.json` and
`strategy-lane-authority-US.json`. Trust pins are market-specific digests plus one configured public
key/key ID. TossOS provides no writer or signer. Symlinks, wrong owner/mode, duplicate keys, unknown
fields, non-canonical values, signature/digest mismatch and cross-market content fail that market
closed without invalidating a valid peer.

## Paired delivery and verification matrix

Every implementation slice runs in this order:

```text
KR RED + US RED
      ↓
KR GREEN + US GREEN
      ↓
one paired integration checkpoint
```

The same wave covers all six bindings:

| strategy family | KR | US |
| --- | --- | --- |
| continuation | `kr-flow-continuation` 8:4:2 | `us-participation-continuation` 8:4:2 |
| reversal | `kr-absorption-reversal` 2:4:8 | `us-dislocation-reversal` 2:4:8 |
| weekly value | `kr-opendart-weekly-value` | `us-edgar-weekly-value` |

Required paired tests cover exact proposal quantity, q_final reduction, no quantity increase,
market-local evidence/calendar/manifest failure, router tie/refusal, stop non-retreat, weekly durable
reservation restart, concurrent account-wide cap contention and zero mutation/transport before the
v26 atomic admission.

KR operational stability is not a prerequisite for US design, RED, GREEN, production wiring or
verification, and the inverse is also forbidden. Phase 5 E2E/race/restart/rollback/gate work runs on
one paired release candidate; a single-market pass is not a completion milestone.

## Existing-function edit boundary

Before changing `continuationlane.evaluate`, `reversallane.evaluate`,
`weeklyvaluelane.evaluate`, `strategyflow.evaluateWith` or `strategyrouter.Route`, their current
Function Logic Map and Branch Test Map must be present and refreshed. The implementation should share
validation logic between proposal and admitted evaluation so the two stages cannot drift, while
preserving every existing refusal and invalidation branch.
