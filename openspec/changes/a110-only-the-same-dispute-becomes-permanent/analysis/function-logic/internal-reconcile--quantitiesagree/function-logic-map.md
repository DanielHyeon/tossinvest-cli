# Function Logic Map: `quantitiesAgree`

- Source: `internal/reconcile/compare.go`
- Evidence: `ast.json`, `risk-pattern-report.md`

## Inputs and invariants

| Input | Valid range | Source | Failure behavior |
|---|---|---|---|
| `a`, `b` | finite decimal strings | snapshot/local position projection | invalid or non-finite values disagree and remain visible |

## Branches and early returns

| Branch | Condition | Result | Test |
|---|---|---|---|
| B1 | either exact canonicalization fails | mismatch | `TestA110ComparerRejectsIdenticalInvalidQuantityStrings` |
| B2 | canonical decimals equal | agreement | `TestAgreementIsClean` |
| B3 | defensive float parse fails | mismatch | invalid quantity suite |
| B4 | distinct decimals collapse to one float64 | mismatch | `TestA110ComparerDoesNotTreatDistinctLargeIntegersAsEqual` |
| Return | distinct floats | delegate to proof-shaped short-decimal/binary-expanded artifact predicate | `TestFractionalQuantitiesSurviveTheFloatRoundTrip`, exact-integer A110 test |

## Calls and live bindings

| Callee | Purpose | Contract | Evidence |
|---|---|---|---|
| `riskcalc.CanonicalDecimal` | exact finite validation/equality | no float64 | AST B1–B2 |
| `strconv.ParseFloat` | legacy round-trip artifact measurement | exact-equal collision is rejected first | AST B3–B4 |
| `binaryRoundTripArtifact` | admit only a symmetric short fractional spelling and its shortest adjacent expanded spelling | generic one-ULP proximity and exact integers are rejected | F7 RED/GREEN |

## State mutations and fallbacks

- Pure function; it cannot open a gate itself. `Comparer.Compare` turns false into a visible quantity mismatch.

## Safety conclusion

- Safe boundary: exact validation and collision rejection before the explicit round-trip spelling proof.
- High-risk impact: yes; a false agreement can bypass reconciliation entry blocking.
