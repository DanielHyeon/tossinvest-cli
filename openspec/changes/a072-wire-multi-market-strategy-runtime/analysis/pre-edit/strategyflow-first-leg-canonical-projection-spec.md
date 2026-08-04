# Strategyflow accepted-result canonical first-leg projection

**Author:** a072 implementation teammate
**Date:** 2026-08-04
**Status:** Approved refinement of approved a072 strategy-runtime requirements
**Reviewer/approver:** a072 Manager (`/root` task assignment)
**Scope:** read-only `strategyflow.Result` projection and journal RiskIntent binding verification

## Context

The six production strategy lanes already return an accepted `strategyflow.Result` containing a sealed
`Lineage` and opaque sealed `ExecutionTerms`. Journal first-leg admission currently accepts only the
legacy `strategyengine.DecisionRecord` JSON shape mirrored as `strategyDecisionPayload`. Consequently,
the production strategyflow result cannot be transformed into the exact `StrategyDecisionLineage`
shape that `RecordQFinalCampaignFirstLeg` verifies without a caller inventing an unrelated legacy
record.

This refinement adds a versioned canonical evidence projection. It is deliberately not an execution
capability: it performs deterministic normalization and verification only. The existing q_final,
reservation, campaign, router, dispatch lease, Gateway and broker gates remain the sole downstream
authorities. KR continuation/reversal/weekly and US continuation/reversal/weekly are delivered in one
paired wave.

## Functional Requirements

- FR-1: Strategyflow MUST emit canonical schema `strategyflow-accepted:v1` only when the
  result is accepted, pure (`GuardianCalls=BrokerCalls=Mutations=0`), has a complete valid sealed
  `Lineage`, and has valid sealed `ExecutionTerms` exactly bound to that lineage and quantity.
- FR-2: The canonical strategyflow projection MUST include every sealed lineage field,
  every execution price provenance field, the complete execution-policy preimage and both existing
  lineage/execution-term identities. It MUST NOT reseal an invalid or mutated result.
- FR-3: The projection verifier MUST require the fixed production router identity/release
  and exactly one of the six registered KR/US market+horizon+lane+version+release descriptors.
- FR-4: Journal MUST wrap the verified strategyflow projection in canonical schema
  `journal-strategyflow-risk-binding:v2`, bind it to one canonical RiskIntent, and derive
  `strategy-decision:v2:sha256:<digest>` from the canonical identity-free outer payload.
- FR-5: Journal v2 verification MUST compare RiskIntent and decision account, market,
  symbol, entry, stop and target exactly with the outer payload and verified strategyflow terms.
  RiskIntent q_final MUST be a positive canonical integer no greater than the sealed strategyflow
  candidate quantity; the persisted lineage quantity MUST be q_final, not the candidate quantity.
- FR-11: The full Guardian q_final policy version MUST be bound exactly in the outer RiskIntent and
  persisted lineage policy field. It MUST remain independent from the lane execution-policy identity
  sealed in the inner strategyflow result; neither identity may be substituted for the other.
- FR-6: Journal MUST compare the v2 projection back to every persisted
  `StrategyDecisionLineage` column: decision identity, candidate, market/symbol, thresholds,
  evidence, lane/version, lane evidence/config digests, entry/stop/target, q_final quantity, policy,
  settings digest, payload and payload digest.
- FR-7: The legacy `strategyDecisionPayload` / `strategy-decision:v1` verifier MUST remain
  accepted without field, identity, digest or canonicalization changes. Unknown non-empty schema
  versions MUST fail closed.
- FR-8: Projection and verification MUST be record-only. They MUST NOT write journal rows,
  mint prospective tokens or leases, call Gateway/broker, activate a lane, change a toggle, or
  create LIVE approval.
- FR-9: KR and US continuation, reversal and weekly projections MUST be implemented and
  tested together; a market or lane family MAY NOT be reported complete independently.
- FR-10: Strategyflow prospective `PositionGeneration` MUST remain sealed in the inner evidence but
  MUST NOT be equated with a flat campaign's expected generation/version `0/0` during first entry.

## Non-Functional Requirements

- NFR-1: JSON decoding MUST reject unknown fields, trailing values,
  non-canonical encodings, invalid seals, unsupported descriptors and cross-market substitution.
- NFR-2: Identical accepted input and RiskIntent MUST produce byte-identical
  canonical payload, payload digest and decision identity across replay.
- NFR-3: Any mismatch MUST return before the first journal transaction in
  first-leg admission, leaving decision/reservation/campaign/attempt/lease rows at zero.
- NFR-4: Existing untagged journal strategy-lineage tests MUST remain green,
  and the v14-v26 schema bytes MUST remain unchanged.
- NFR-5: The projection contains no mutable global state and MUST pass affected
  package race tests.

## Acceptance Criteria

### AC-1: Paired six-lane projection (FR-1, FR-2, FR-3, FR-9)

Given accepted results for all six
  registered KR/US lanes, when each result is projected and verified, then every projection is
  byte-canonical, retains both valid identities and maps to its exact descriptor.
### AC-2: Exact journal binding (FR-4, FR-5, FR-6)

Given paired KR/US accepted results and matching
  q_final RiskIntents, when journal creates v2 lineage and first-leg validation runs, then exact
  account/market/symbol/entry/stop/target/Guardian-policy bindings pass and q_final is persisted.
### AC-3: Drift refusal before writes (FR-5)

Given quantity, account, market, symbol, entry, stop, target or
  policy drift, when first-leg validation runs, then it refuses before any journal authority row.

### AC-8: Guardian downsizing and first-entry generations (FR-5, FR-10, FR-11)

Given paired KR/US accepted results with candidate quantity 20, prospective position generation 1,
Guardian q_final 10, independent q_final policy `engine.automation_gate/risk-policy-v1`, and flat campaign generation/version 0/0
When journal projects and records the first leg
Then both markets persist quantity 10, retain candidate quantity 20 only inside the sealed payload,
and admit the first entry without equating the two policy identities or generation domains.
### AC-4: Tamper refusal (FR-2, FR-3)

Given tampered lineage identity, execution-terms
  identity, router, lane descriptor, provenance, unknown JSON field, trailing JSON or non-canonical
  whitespace, when projection verification runs, then it fails closed.
### AC-5: Legacy compatibility (FR-7)

Given an unchanged legacy v1 fixture, when existing strategy
  issuance runs, then its prior identity/digest and acceptance behavior remain unchanged.
### AC-6: Read-only projection (FR-8)

Given any successful or refused projection, when static and runtime
  assertions inspect side effects, then Guardian, Gateway, broker, lease, activation and toggle
  calls/writes are zero.
### AC-7: Deterministic replay (FR-4)

Given the same accepted result and RiskIntent twice, when projected, then
  payload bytes, digest, v2 decision identity and populated lineage are exactly equal.

## Edge Cases

- EC-1: Refused, incomplete, non-pure, zero-quantity or zero-value `Result` is rejected.
- EC-2: A valid KR result paired with US RiskIntent, or the reverse, is rejected.
- EC-3: q_final zero, fractional, non-canonical or greater than candidate quantity is rejected;
  a positive q_final smaller than candidate quantity is accepted and persisted exactly.
- EC-4: A syntactically valid payload with an unsupported schema version is rejected.
- EC-5: A valid seal whose lane descriptor is not one of the six frozen bindings is rejected.
- EC-6: Decimal aliases, surrounding whitespace, lowercase venue substitution and trailing
  JSON are rejected by the existing canonical RiskIntent/parser or v2 canonical verifier.

## API Contracts

```typescript
// Opaque read-only value; callers cannot set its fields.
interface AcceptedProjection {
  payload(): string;
  payloadDigest(): `sha256:${string}`;
  lineage(): SealedLineageCopy;
  executionTerms(): OpaqueSealedExecutionTermsCopy;
}

function projectAccepted(result: strategyflow.Result): AcceptedProjection | Error;
function verifyAcceptedProjection(payload: string): AcceptedProjection | Error;

interface StrategyflowLineageProjectionRequest {
  result: strategyflow.Result;
  riskIntent: journal.RiskIntent;
  activationManifestDigest: string;
  createdAt: timestamp;
}

// Returns evidence for StrategyPlanRequest; performs no write and grants no authority.
function projectAcceptedStrategyflowLineage(
  request: StrategyflowLineageProjectionRequest
): journal.StrategyDecisionLineage | Error;
```

Errors are typed Go errors wrapping `strategyflow: canonical accepted projection` or
`journal strategy issuance: strategyflow projection`; they contain no broker capability and no
retry instruction.

N/A -- internal Go-only API; no HTTP method or route is added, and specifically no
`POST /strategyflow/projection` endpoint exists.

## Data Models

| Entity | Field | Type | Constraint |
|---|---|---|---|
| Strategyflow accepted payload | `schema_version` | string | exact `strategyflow-accepted:v1` |
| Strategyflow accepted payload | `lineage` | object | all `Lineage` fields including valid identity |
| Strategyflow accepted payload | `execution_terms` | object | all term/provenance/policy fields including valid identity |
| Strategyflow accepted payload | `quantity` | uint64 | candidate quantity before Guardian downsizing |
| Strategyflow accepted payload | `common_safety_independent` | bool | exact `true` |
| Strategyflow accepted payload | call/mutation counters | uint64 | exact zero |
| Journal outer payload | `schema_version` | string | exact `journal-strategyflow-risk-binding:v2` |
| Journal outer payload | `strategyflow` | object | byte-canonical verified inner payload |
| Journal outer payload | `risk_intent` | object | exact account/market/symbol/q_final/prices/policy |
| Journal outer payload | `identity` | string | exact `strategy-decision:v2:sha256:<64 lowerhex>` |
| Existing SQL tables | N/A | N/A | no schema/table/column change |

## Out of Scope

- OS-1: Gateway construction, broker transport, lease issuance/claim/submitting, recovery and reconciliation.
- OS-2: Guardian sizing or q_final calculation; this change only checks the already produced exact q_final.
- OS-3: Account/FX authority loaders, activation manifests, operating toggles, LIVE approval and deployment.
- OS-4: Replacing or migrating legacy v1 rows; v1 remains additive-compatible.
- OS-5: Adding a seventh lane, changing lane constants, prices, sizing or execution-policy algorithms.
