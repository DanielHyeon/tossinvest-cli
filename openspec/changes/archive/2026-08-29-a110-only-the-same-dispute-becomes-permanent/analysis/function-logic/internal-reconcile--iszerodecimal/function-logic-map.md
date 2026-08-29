# Function Logic Map: `isZeroDecimal`

- Source: `internal/reconcile/compare.go`
- Evidence: `ast.json`, `risk-pattern-report.md`

## Inputs and invariants

| Input | Valid range | Authority | Failure behavior |
|---|---|---|---|
| quantity text | finite decimal | raw validation/canonical comparer | invalid or nonzero is never treated as zero |

## Branches and early returns

| Branch | Condition | Result | Test |
|---|---|---|---|
| B1 | exact canonicalization fails | false | invalid raw tests |
| Return | canonical equals zero | true only for exact numeric zero | near-zero tolerance test |

## Calls and live bindings

| Callee | Purpose | Contract | Evidence |
|---|---|---|---|
| `riskcalc.CanonicalDecimal` | exact zero vocabulary | no epsilon | AST B1 |

## State mutations and fallbacks

- Pure predicate; no small nonzero exposure is erased as zero.

## Safety conclusion

- Safe boundary: exact canonical zero only.
- High-risk impact: yes; false zero can hide exposure.
