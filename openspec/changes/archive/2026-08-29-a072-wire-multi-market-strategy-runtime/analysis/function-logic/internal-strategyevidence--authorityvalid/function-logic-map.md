# Function Logic Map: `authorityValid`

- Source: `internal/strategyevidence/model.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `authority` | exact enumerated official evidence source | a064 evidence source contract plus frozen Toss official read contract | unknown string returns false and the envelope is rejected before storage/replay |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | OpenDART, KRX, SEC EDGAR or Toss official Open API | none | true | authority matrix test |
| B2 | any other value, including empty/WTS/KIS aliases | none | false | authority matrix test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| none | pure enum classification | no error, timeout or retry path | AST |

## State mutations and fallbacks

- No state mutation, network call or fallback. Toss official is an explicit source identity, never an alias for WTS.

## Safety conclusion

- Safe edit boundary: add only the exact official Toss source constant and keep the default refusal.
- High-risk impact: yes; an over-broad source acceptance could admit unofficial evidence.
