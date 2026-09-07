# Function Logic Map: `Journal.CreatePositionCampaign`

- Source: `internal/journal/position_campaign.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| request identity | non-empty canonical account/market/symbol/lane/decision/token/key | caller + domain validation | `ErrInvalidRequest` / identity error |
| decision | persisted exposure-raising decision for same account | `decisions` | fail closed, no campaign |
| strategy identity | exact market/symbol/lane/version/evidence for the risk decision | `strategy_attempt_lineage` → immutable `strategy_decision_lineage` | invalid identity; no campaign or claim |
| Position generation/state/version | no Position (`0/0`) or latest authoritative `CLOSED` row with v20 companion version | `positions` + `position_projection_versions` | `ErrGenerationConflict` |
| prospective claim | no existing scope claim | `position_campaign_claims` | `ErrGenerationConflict` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | invalid request/domain identity | none | typed refusal | validation tests |
| B2 | exact CREATE retry | read existing command | existing record | CAS retry test |
| B3 | missing/wrong decision or missing/cross-scope strategy lineage | none | invalid identity | hardening + exact-lineage tests |
| B4 | legacy unversioned or non-CLOSED Position | none | generation conflict | OPEN/legacy CAS tests |
| B5 | generation/version mismatch or active claim | none | generation conflict | concurrent/stale tests |
| B6 | authoritative CAS succeeds | campaign+claim+command+event in one tx | record | create/race tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `campaignCommandResult` | deterministic retry | digest mismatch refuses | AST |
| immutable strategy lineage join | bind campaign identity to strategy authority | absent or any identity mismatch fails closed | AST + independent-review test |
| `campaignProjectionDigest` via event insert | complete replay checkpoint | tx error rolls all back | AST |

## State mutations and fallbacks

- No Position mutation, broker call, order dispatch, or runtime-toggle write.
- Campaign, claim, command and event commit atomically; event projection digest binds immutable identity, expected generation/version and the full durable claim row.

## Safety conclusion

- Safe edit boundary: additive journal CAS and evidence only.
- High-risk impact: yes; fail-closed generation ownership.
