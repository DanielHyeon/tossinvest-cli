# Function Logic Map: `verifyStrategyflowRiskBinding`

- Source: `internal/journal/strategyflow_projection.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| decision payload | canonical v2 historical or v3 current record | journal lineage | unknown/trailing/non-canonical bytes fail closed |
| embedded projection | verified accepted strategyflow payload | strategyflow verifier | tamper/seal/router/lane failure rejects |
| persisted lineage | exact schema-aware deterministic projection | journal row | any drift rejects |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | outer decode/schema/canonical check fails | none | decode error | tamper table |
| B2 | inner projection verification fails | none | wrapped projection error | inner tamper test |
| B3 | embedded digest or RiskIntent mismatch | none | mismatch error | outer drift test |
| B4 | schema-aware price binding fails | none | exact-binding error | paired drift test |
| B5 | identity/time/columns mismatch | none | exact-binding error | column drift table |
| B6 | exact v2 or v3 record | none | nil | compatibility + paired success |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `decodeStrategyflowRiskBinding` | strict outer decode and schema allowlist | deterministic; no retry | AST |
| `strategyflow.VerifyAcceptedProjection` | authenticate inner evidence | deterministic; no I/O | AST |
| `verifyProjectionRiskIntent` | apply v2/v3 unit semantics | deterministic; fail closed | pre-edit contract |
| `strategyflowDecisionIdentity` | recompute schema-bound identity | deterministic SHA-256 | AST |

## State mutations and fallbacks

- Read-only verifier with no history rewrite, write, Gateway, lease, or fallback.
- Unknown versions cannot downgrade to legacy.

## Safety conclusion

- Safe edit boundary: version-aware verification before authority use.
- High-risk impact: yes; paired unit, replay, tamper, and compatibility tests are mandatory.
