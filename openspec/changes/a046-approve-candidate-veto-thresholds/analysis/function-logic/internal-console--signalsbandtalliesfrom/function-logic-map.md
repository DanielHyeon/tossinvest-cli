# Function Logic Map: `signalsBandTalliesFrom`

- Source: `internal/console/signals.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| band tally map | zero or more D3 keys | candidate shadow measurement | missing code is omitted, present code rendered in fixed order |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate copied D3 order | local result | continue | signals band tests |
| B2 | code absent | none | skip | sparse tally test |
| B3 | iterate configured bands | local append | continue | band crossing test |
| B4 | iterate missing reasons | local map fill | sorted result | missing-band test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `OrderedVetoCodes`, `BandsFor`, render helpers | stable projection | pure | CodeGraph + AST |

## State mutations and fallbacks

- Fresh DTO only; no numeric fallback for absent tally entries.

## Safety conclusion

- Safe edit boundary: immutable order accessor.
- High-risk impact: no; shadow display only.
