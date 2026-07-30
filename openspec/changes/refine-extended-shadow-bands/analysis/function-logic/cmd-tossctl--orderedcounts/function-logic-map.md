# Function Logic Map: `orderedCounts`

- Source: `cmd/tossctl/candidate.go`
- Function: `cmd/tossctl/candidate.go:orderedCounts`
- AST evidence: `ast.json` — the branch ids, returns and callees below are read from it
- Risk scan: `risk-pattern-report.md`
- Change: `refine-extended-shadow-bands`

Changed in this change: the body moved into orderedCountParts so the same list can also be folded. Behaviour is identical and the branch is base's.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| keys []string | the declared order | the caller | an empty key list renders "none" |
| counts map[string]int | any | the caller | a key with no count renders 0, so a column nobody crossed keeps its place |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | there are no keys | none | "none" rather than an empty string | existing scan report tests; a code with no scale renders "none" |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| orderedCountParts | NEW - the shared body | total | ast.json calls |
| strings.Join | unchanged | total | ast.json calls |

## State mutations and fallbacks

- None.

## Safety conclusion

- Safe edit boundary: Dropping a key whose count is zero.
- High-risk impact: no.
