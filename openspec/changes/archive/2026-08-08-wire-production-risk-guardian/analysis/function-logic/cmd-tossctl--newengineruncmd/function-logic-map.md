# Function Logic Map: `newEngineRunCmd`

- Source: `cmd/tossctl/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root` | CLI root options; may be nil until `RunE` executes | Cobra command assembly | forwarded unchanged to `runEngineRun` |
| help contract | gate/interlock and ProtectionReady behavior must match the shipped engine | engine safety specs + gateway protection enforcement | misleading operator action if stale |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | happy path (AST is branchless) | constructs one Cobra command | returns command; `RunE` forwards runtime errors | help and annotation tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | normalize the multiline operator help | cannot fail | CodeGraph + AST |
| `runEngineRun` | execute the fixed startup sequence | all startup/runtime errors propagate | CodeGraph + AST |

## State mutations and fallbacks

- Builds metadata only; it does not assemble the engine or mutate account state.
- The `mutating=true` annotation remains because the runtime can submit verified
  reduce-only orders, even while UNWIRED protection blocks entries.

## Safety conclusion

- Safe edit boundary: operator help text only; command annotations and `RunE`
  binding must remain unchanged.
- High-risk impact: yes — inaccurate startup/protection instructions can make an
  operator bypass or misdiagnose safety controls.
