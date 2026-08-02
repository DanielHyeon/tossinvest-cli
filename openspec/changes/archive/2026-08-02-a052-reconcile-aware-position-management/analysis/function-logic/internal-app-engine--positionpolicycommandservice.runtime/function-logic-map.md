# Function Logic Map: `PositionPolicyCommandService.Runtime`

- Source: `internal/app/engine/position_policy_command.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| startup adoption | enabled/default stop/include/exclude/rejected copied into service | engine-start config snapshot | always emitted as effective-known; no config file reread |
| reconcile tracker projection | nil or active `Blocks()` reader | same tracker used by adoption entry gate | nil yields known empty block list; no journal-wide completeness claim |
| block identity/detail | scope/market/symbol/reason/detail/since/permanent | reconcile tracker | constructor sanitizes detail; DTO contains no account/capability/token/command |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 109 | `if s.blocks == nil {` | none beyond initialized immutable DTO | returns known effective settings and empty block list | `TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks` nil-reader subcase |
| B2 | `range` at line 112 | `for _, block := range s.blocks.Blocks() {` | appends sanitized transport-neutral block values | returns full current adoption-blocking tracker projection | `TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positionpolicy.NewAdoptionSettings` | normalize immutable startup settings | deterministic; rejected input remains explicitly represented | runtime DTO test |
| `s.blocks.Blocks` | read the same in-memory projection that gates adoption | no retry/remote call; snapshot may change next cycle | B2 + runtime DTO test |
| `append` | execute the explicit dependency at line 113 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `positionpolicy.NewReconcileBlock` | mask account-like text and construct typed read-only block | deterministic sanitization; no release authority | security assertions in runtime DTO test |
| `positionPolicyReconcileScope` | execute the explicit dependency at line 114 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `string` | execute the explicit dependency at line 115 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- Only the local response slice is mutated; the tracker, journal, config, and engine gate are read-only.
- `BlockSource` explicitly says adoption-blocking tracker projection, avoiding a false claim that this is every journal cause.
- No account reference, capability, token, resolution command, or live operation is returned.

## Safety conclusion

- Safe edit boundary: Only the approved read/projection or fail-closed adoption boundary may change; no order placement, reconciliation resolution, or live toggle mutation is authorized.
- High-risk impact: yes; adoption provenance, reconciliation blocking, or persisted position lifecycle is money-sensitive.
