# Function Logic Map: `requireStrategyProtectionPositionInTx`

Source `internal/journal/strategy_protection_read.go`; extracted during a150 before the generation correction, `ast.json`, `ast-grep.txt` and `risk-pattern-report.md`. New a150 helper, not present at base 28a800f5.

## Inputs and invariants

| Input | Source of truth | Failure behavior |
|---|---|---|
| transaction and campaign | one BEGIN IMMEDIATE snapshot | DB failure prevents new transport authority |
| admitted policy and RiskIntent | verified immutable v4 binding | mismatch prevents additional exposure |
| current position and exit | positions, campaign generation and exit snapshot | missing or different values fail closed |
| prefill expected generation/version | admission campaign snapshot | current absence is valid only for 0/0; closed row must match both |

## Branches and early returns

| Branch | Condition | Effect | Verification |
|---|---|---|---|
| B1 | explicit legacy marker | retain legacy transport behavior | legacy compatibility tests pending |
| B2 | campaign read fails | deny | DB failure test pending |
| B3 | position absent | allow only unbound actual plus expected 0/0 | missing expected row RED/GREEN |
| B4 | position read fails | deny | existing DB tests pending |
| B5 | latest closed and actual generation absent | require same expected generation and version | closed successor RED/GREEN |
| B6 | active position | require actual generation and decision to match | v4 position tests |
| B7 | exit missing, corrupt, completed or different | deny without deleting exit | a150 final position tests |
| B8 | all bindings match | authorize consistent read instant | a150 final transport test |

## Calls and live bindings

Campaign, position, version and exit read from same journal transaction. No broker call or lock across network. Prior authorized SUBMITTING may remain in flight and reserved.

## Safety conclusion

High-risk helper: yes. Expected generation correction prevents stale prefill authority from accepting disappeared or newer closed position.

## State mutations and fallbacks

Current transaction and fallback branches are enumerated in the current AST and branch map; mutation failures propagate, with protective-exit behavior verified separately.


Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

## Current implementation clarification

Final authority validates admitted generation as well as actual and expected position generation, prior projection version and current exit snapshot; a verified prefill may have no row only for expected 0/0. Missing evidence, drift and mismatched exit refuse. Prior authorized SUBMITTING remains possibly in flight.




Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.

Current AST and branch inventory reflect the implementation; `pre-edit-ast.json` retains the initial extraction where available. Focused receipts and independent reviews are separate evidence.
