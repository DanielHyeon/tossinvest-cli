# Function Logic Map: `renewalReasonMessage`

- Source: `internal/console/data.go` (342-363); SHA-256 `5b4c6a2b962e4e9459822a1863ff6708f136baaeb5ffaefed5be22f0364be23a` (post-edit). Pre-edit AST: `../../scope-absorb/pre-edit-ast/internal-console--renewalreasonmessage.json` (342-361, 9 branches).
- AST evidence: `ast.json`. Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `code` | closed `soak.RenewalReasonCode` | validated by `soak.LoadRenewalStatus` before this is called (`attestView.readRenewalStatus`) | default fixed message |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | switch on `code` | none | — | `TestDashboardShowsBoundedRenewalDiagnosticWithoutChangingAttestationUsability` |
| B2 | record_empty | none | fixed message | unmeasured by focused test |
| B3 | record_stale | none | fixed message | unmeasured by focused test |
| B4 | insufficient_streak | none | fixed message | `TestDashboardShowsBoundedRenewalDiagnosticWithoutChangingAttestationUsability` |
| B5 | token_observation_missing | none | fixed message | unmeasured by focused test |
| B6 | token_refresh_missing | none | fixed message | unmeasured by focused test |
| B7 | required_endpoint_missing | none | fixed message | unmeasured by focused test |
| B8 | completeness_failed | none | fixed message | unmeasured by focused test |
| B9 | **D6 new case** `live_endpoint_uncovered` | none | fixed message naming restart refusal / exit-only start when the automation gate is on (review r1 F2) | `TestA063DashboardTellsAnIssuedRenewalThatLostTheSupervisedScope` |
| B10 | default (input_error, supervised_evidence_invalid, account codes) | none | generic fixed message | `TestA063DashboardTellsAnIssuedRenewalThatLostTheSupervisedScope` asserts the new code does not fall here |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | — | — | CodeGraph + AST |

## State mutations and fallbacks

- Pure; returns constant strings only (no raw error text, path or account data).

## Safety conclusion

- Safe edit boundary: one case inserted before the default; existing messages unchanged.
- High-risk impact: no — console display text.
