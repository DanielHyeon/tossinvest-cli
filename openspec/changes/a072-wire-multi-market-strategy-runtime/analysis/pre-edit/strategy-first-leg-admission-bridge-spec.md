# Spec: KR/US Strategyflow to Atomic First-Leg Admission Bridge

**Author:** Codex (a072 implementation teammate)
**Date:** 2026-08-04
**Status:** Approved
**Reviewers:** a072 approved design; user KR/US simultaneous-delivery decision

## Context

`strategyflow.Result` now preserves an accepted lane's sealed lineage and exact entry, effective-stop
and target price provenance for KR and US. The journal also has one atomic first-leg operation that
commits q_final, aggregate and five-bucket reservations, strategy lineage, campaign, owner and leg 1
without issuing a dispatch lease. There is still no production authority loader that can collect the
account, exposure, five bucket, activation, official FX and account-base Guardian evidence required
to build that atomic request.

The shared `execgw.RiskGuardian` now owns an opaque first-leg precheck/finalize pair. It accepts the
sealed strategy result plus activation/attempt metadata and authoritative campaign current CAS,
derives q_final and its independent Guardian policy, projects the canonical strategyflow lineage,
and calls only the atomic first-leg transaction. The engine bridge therefore holds exactly this
two-method Guardian capability and fails closed while its production authority loader is absent. It
does not accept a caller-built `StrategyPlanRequest`, expose a raw journal request, or add runtime
wiring, dispatch or broker capability.

## Functional Requirements

- FR-1: The bridge MUST accept only a `strategyflow.Result` whose refusal code is empty, quantity is
  positive, lineage is complete and valid, execution terms are valid, and all scope, quantity,
  router and execution-term bindings are exact.
- FR-2: KR and US MUST use the same bridge implementation and the same table-driven test wave. A
  result for one market MUST NOT be a prerequisite for processing the peer market.
- FR-3: Production authority collection MUST be represented by an engine-package-owned interface
  with an unexported method. No exported constructor or caller-populated q_final/first-leg DTO MUST
  be added.
- FR-4: If the authority loader, two-method Guardian issuer or complete sealed authority is absent,
  the bridge MUST return a typed fail-closed result and MUST make zero Guardian calls.
- FR-5: A collected authority MUST exactly repeat the accepted account, market, symbol, campaign,
  lane, candidate, execution prices, q_candidate, activation/attempt metadata, empty caller
  prospective token and campaign current generation/version before it reaches the Guardian.
- FR-6: The only mutation capability reachable from the bridge MUST be the Guardian's
  `PrecheckQFinalCampaignFirstLeg` followed by `IssuePrecheckedQFinalCampaignFirstLeg`. The bridge MUST
  NOT hold a raw Journal, standalone q_final, dispatch lease, Gateway, broker, activation, approval
  or operating-toggle API.
- FR-7: Guardian-owned journal recollection MUST use the exact sealed entry authority captured by
  the opaque precheck. The bridge MUST NOT collect a caller-built journal request or repair a rejected
  partial authority.
- FR-8: Success MUST return only the journal's opaque first-leg receipt. It MUST NOT return a token,
  lease, broker order, Gateway or activation capability.

## Non-Functional Requirements

- NFR-1: Validation and authority collection MUST perform zero network calls inside the bridge; any
  future network reads belong to the package-owned loader before the journal transaction.
- NFR-2: Missing or mismatched authority MUST let zero requests reach the journal mutation
  transaction in deterministic unit tests.
- NFR-3: The bridge MUST contain no loop. Bounded account snapshot recollection remains owned by the
  Guardian/journal atomic issue path.
- NFR-4: Existing engine workers, safety loops, live defaults and operating settings MUST remain
  unchanged because this slice adds new files only.

## Acceptance Criteria

### AC-1: Paired missing-authority refusal (FR-2, FR-3, FR-4; NFR-2)
Given valid accepted KR and US strategyflow results
When each enters a bridge with no production authority loader
Then both return the typed `AUTHORITY_UNAVAILABLE` result
And both Guardian methods are called zero times for both markets.

### AC-2: Paired atomic-only handoff (FR-2, FR-5, FR-6, FR-7, FR-8)
Given valid accepted results for all six KR/US lanes and package-owned test authorities
When admission runs
Then the loader receives the matching sealed result once
And the bridge calls Guardian precheck then Guardian atomic issue exactly once in that order
And success exposes only the matching opaque first-leg receipt.

### AC-3: Invalid strategy result fails before authority (FR-1, FR-4)
Given a refused, incomplete, mutated or non-first-leg strategyflow result
When admission is attempted
Then the bridge returns `STRATEGY_RESULT_INVALID`
And neither loader nor Guardian is called.

### AC-4: Cross-market or execution substitution fails before mutation (FR-5; NFR-2)
Given valid paired KR and US results
When collected authority substitutes the peer market, symbol, campaign, lane, router, candidate,
quantity, execution price or campaign current generation
Then the bridge returns `AUTHORITY_MISMATCH`
And neither Guardian method is called.

### AC-5: Static capability boundary (FR-3, FR-6, FR-8; NFR-4)
Given the bridge source
When its imports and call surface are inspected
Then no standalone q_final issue, dispatch lease, Gateway, broker, activation, approval or toggle
capability is present
And no existing engine file is modified by this slice.

## Edge Cases

- EC-1: Nil bridge, loader or Guardian issuer -> typed `AUTHORITY_UNAVAILABLE`; zero mutation.
- EC-2: Context cancellation from the loader -> typed `AUTHORITY_COLLECTION_FAILED`; zero mutation.
- EC-3: Zero-value package-owned authority -> typed `AUTHORITY_MISMATCH`; zero mutation.
- EC-4: KR provenance with non-KRW or nonzero minor scale -> `STRATEGY_RESULT_INVALID`.
- EC-5: US provenance with non-USD or scale other than two -> `STRATEGY_RESULT_INVALID`.
- EC-6: Execution terms not bound to the exact lineage identity -> `STRATEGY_RESULT_INVALID`.
- EC-7: Leg ordinal other than 1 -> `STRATEGY_RESULT_INVALID`; scale-in remains out of scope.
- EC-8: Guardian atomic issue refusal -> typed `ATOMIC_ADMISSION_FAILED`; no fallback issuance path.
- EC-9: Guardian opaque precheck refusal -> `AUTHORITY_MISMATCH`; issue is not called.

## API Contracts

This is an internal Go contract, not an HTTP endpoint. The logical shapes below document the
package boundary; the authority types and loader methods remain unexported.
For schema-validation notation only, the operation is written as
`POST /internal-go/engine/strategy-first-leg-admission`; it MUST NOT be registered by an HTTP router.

```typescript
interface StrategyFirstLegAdmissionResult {
  code: "ADMITTED" | "STRATEGY_RESULT_INVALID" | "AUTHORITY_UNAVAILABLE" |
        "AUTHORITY_COLLECTION_FAILED" | "AUTHORITY_MISMATCH" | "ATOMIC_ADMISSION_FAILED";
  market: "KR" | "US" | "";
  receipt?: QFinalCampaignFirstLegReceipt; // opaque, no prospective token or lease
  detail: string;
}

interface PackageOwnedAuthorityLoader {
  // Go method is unexported; only package engine can implement it.
  collectStrategyFirstLegAuthority(ctx, acceptedResult): QFinalCampaignFirstLegIssuance;
}

interface StrategyFirstLegGuardian {
  PrecheckQFinalCampaignFirstLeg(issuance): QFinalCampaignFirstLegPrecheck;
  IssuePrecheckedQFinalCampaignFirstLeg(ctx, precheck): QFinalCampaignFirstLegReceipt;
}
```

No public method accepts `journal.QFinalCampaignFirstLegRequest` from an engine caller.

## Data Models

No database model or migration is added.

| Value | Type | Constraints |
| --- | --- | --- |
| accepted strategy | `strategyflow.Result` | valid complete lineage, valid exact execution terms, leg 1 |
| sealed authority | `execgw.QFinalCampaignFirstLegIssuance` | sealed result + loader-owned entry, activation/attempt and current campaign CAS; no `StrategyPlanRequest` |
| admission result | typed value | receipt only on `ADMITTED`; no dispatch or broker authority |
| durable write | existing journal v26 transaction | atomic q_final/campaign first-leg only |

## Out of Scope

- OS-1: Implementing account, exposure, five-bucket, fee, activation, official FX or account-base
  Guardian production loaders.
- OS-2: Production implementation of the engine-private account/FX/campaign authority loader.
- OS-3: Dispatch lease issuance or claim, `CLAIMED->SUBMITTING`, Gateway or broker transport.
- OS-4: Scale-in or a campaign leg other than leg 1.
- OS-5: Runtime assembly, autostart, activation, LIVE approval or operating-setting changes.
- OS-6: Journal projector, risk, strategyflow or existing production engine assembly modifications.

## Function Logic Map Applicability

Function Logic Maps and Branch Test Maps are recorded for the two Guardian methods and the engine
bridge admission method under this change's `analysis/function-logic/` tree.
