# Function Logic Map: `canonicalPositiveMinor`

- Source: `internal/weeklyvaluelane/evaluate.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| raw decimal | canonical positive base-10 minor-unit integer | explicit execution request | false without fallback |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | empty or whitespace-altered input | none | false | missing-target test |
| B2 | parse/sign/canonical mismatch | none | false | constructor/terms tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `parseUnsigned` | checked integer parse | false | AST |

## State mutations and fallbacks

- Pure validation; no estimation or normalization of caller values.

## Safety conclusion

- Safe edit boundary: strict execution-term decimal validation.
- High-risk impact: yes, therefore fail closed.
