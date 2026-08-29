# Function Logic Map: `mergeEngine`

- Source: `internal/config/engine.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cfg` | non-nil destination built from safe defaults | `Service.load` | caller owns pointer validity |
| `raw` | nil or decoded optional engine block | config JSON | nil leaves every engine feature OFF |
| optional children | nil or parsed values | `rawEngine` | absent child leaves its zero/default value |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `raw == nil` | none | return with all defaults | missing-engine test |
| B2 | `raw.ExitPolicy != nil` | normalize/validate exit policy | rejected policy is represented, not executed | existing exit-policy tests |
| B3 | `raw.AutomationGate == nil` | adoption/exit policy already merged | return with gate OFF | missing-gate tests |
| B4 | `gate.Enabled != nil` | copy explicit bool | absent remains false | gate default tests |

The change adds one branch immediately after B1: explicit
`raw.Autostart != nil` copies the boolean; nil leaves false.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `mergeAdoption` | normalize and validate adoption | refusal zeros adoption and carries a verdict | AST |
| `strings.TrimSpace` | normalize common policy ID | no I/O | AST |
| `ExitPolicy.validate` | reject unknown policy IDs | verdict string, no panic | AST |

## State mutations and fallbacks

- Mutates only the supplied `cfg` snapshot during config load.
- There is no file write or broker call.
- Missing data always falls back to disabled behavior.

## Safety conclusion

- Safe edit boundary: copy one optional autostart boolean before existing child merges.
- High-risk impact: yes — this value later controls whether an engine start is attempted.
