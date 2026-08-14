# Function Logic Map: `canonicalDecimal`

- Source: `internal/reconcile/snapshot.go`
- Evidence: `ast.json`, `risk-pattern-report.md`

## Inputs and invariants

| Input | Valid range | Source | Failure behavior |
|---|---|---|---|
| decimal text | blank, finite decimal, or unreadable broker evidence | snapshot/comparer payload | blank maps to zero; unreadable text stays visible verbatim |

## Branches and early returns

| Branch | Condition | Result | Test |
|---|---|---|---|
| B1 | trimmed input empty | `"0"` | zero canonicalization tests |
| B2 | exact canonicalization fails | trimmed original | invalid/non-finite comparer tests |
| Return | finite decimal | exact canonical spelling | 2^53 comparer tests |

## Calls and live bindings

| Callee | Purpose | Contract | Evidence |
|---|---|---|---|
| `riskcalc.CanonicalDecimal` | preserve exact finite digits | no float64 | AST B2 |

## State mutations and fallbacks

- Pure vocabulary helper shared by snapshot digest and comparer; malformed evidence is never invented as zero.

## Safety conclusion

- Safe boundary: replace only finite canonicalization, retaining blank and unreadable visibility semantics.
- High-risk impact: yes; lossy output can pool permanent promotion evidence.
