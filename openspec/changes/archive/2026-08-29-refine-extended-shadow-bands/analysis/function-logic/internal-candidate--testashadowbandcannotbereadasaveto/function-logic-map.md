# Function Logic Map: `TestAShadowBandCannotBeReadAsAVeto`

- Source: `internal/candidate/band_test.go`
- Function: `internal/candidate/band_test.go:TestAShadowBandCannotBeReadAsAVeto`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

Changed in this change: it now reads both method sets. A pointer-receiver Dangerous() was invisible to it.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| a measured ShadowBand | any | MeasureSeenLateBand | the value is not what is examined; its type is |
| the forbidden vocabulary | Dangerous, Clear, Raised, Vetoed, Passed, State | this test | a name not on the list is not refused |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | NEW - range over the value type and the pointer type | none | no return | mutation 0.5 - `func (b *ShadowBand) Dangerous() bool` is RED |
| B2 | range over the forbidden names | none | no return | same |
| B3 | the type has such a method | t.Errorf naming the type | no return | same |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| reflect.TypeOf | the value method set | total | ast.json calls |
| reflect.PointerTo | NEW - the pointer method set, which contains both | total | ast.json calls |
| typ.MethodByName | the lookup | total | ast.json calls |

## State mutations and fallbacks

- None.

## Safety conclusion

- Safe edit boundary: Reading only the value method set. Go's method sets differ and a pointer-receiver method is callable on every addressable band in the package.
- High-risk impact: no.
