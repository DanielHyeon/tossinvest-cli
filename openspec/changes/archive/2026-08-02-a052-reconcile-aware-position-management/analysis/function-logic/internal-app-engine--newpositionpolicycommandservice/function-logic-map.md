# Function Logic Map: `NewPositionPolicyCommandService`

- Source: `internal/app/engine/position_policy_command.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| engine context | non-nil with the engine-owned open Journal | engine startup wiring | nil context/journal returns an error; it never opens another journal |
| clock | injected clock or nil | engine wiring | nil substitutes system clock |
| effective adoption | immutable startup config, including rejected entries | `ectx.Config.Engine.Adoption` | copied into service; desired file state is not substituted |
| block reader | the active `ectx.Reconcile` tracker | same engine instance that gates adoption | nil is allowed and produces an empty runtime block projection |

## Branches and early returns

| Branch | AST kind/location | Condition/control path | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|---|
| B1 | `if` at line 82 | `if ectx == nil \|\| ectx.Journal == nil {` | none | returns owned-journal requirement error | `TestEngineOwnsAuthenticatedPositionPolicyControlEndpoint` setup contract |
| B2 | `if` at line 85 | `if clk == nil {` | assigns system clock locally | continues with a complete service | constructor regressions + `TestPositionPolicyRuntimeReturnsStartupEffectiveAndSanitizedBlocks` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `errors.New` | execute the explicit dependency at line 83 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `clock.System` | execute the explicit dependency at line 86 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |
| `strings.TrimSpace` | execute the explicit dependency at line 92 | callee result/error is handled by the AST-recorded branch; no hidden retry is assumed | current AST + focused package tests |

## State mutations and fallbacks

- The returned service aliases the engine-owned journal, official price reader, retrier, startup adoption config, and active reconcile tracker; it does not reopen or migrate storage.
- Runtime truth is bound at engine construction, so later desired-file edits cannot masquerade as effective settings.
- Constructor wiring adds no live order, reconcile resolution, or toggle mutation.

## Safety conclusion

- Safe edit boundary: Only the approved read/projection or fail-closed adoption boundary may change; no order placement, reconciliation resolution, or live toggle mutation is authorized.
- High-risk impact: yes; adoption provenance, reconciliation blocking, or persisted position lifecycle is money-sensitive.
