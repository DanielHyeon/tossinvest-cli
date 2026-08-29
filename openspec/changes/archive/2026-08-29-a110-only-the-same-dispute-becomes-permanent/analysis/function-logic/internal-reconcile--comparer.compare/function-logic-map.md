# Function Logic Map: `Comparer.Compare`

- Source: `internal/reconcile/compare.go`
- Evidence: `ast.json`, `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Authority | Failure behavior |
|---|---|---|---|
| snapshot/local state | one account comparison | broker snapshot + journal projection | unreadable present quantity becomes ordinary mismatch, never external/clean |

## Branches and early returns

| Branches | Decision family | Result | Tests |
|---|---|---|---|
| B1–B7 | timestamp and normalized broker/local symbol union | deterministic symbol walk | comparer baseline |
| B8 | validate present raw broker/local finite decimals | invalid evidence becomes QuantityMismatch | A110 unreadable/raw tables |
| B9–B14 | exact zero, positive broker-only external, negative broker-only mismatch, disagreement, match | correct long-only exposure classification | external/negative/zero/tolerance tests |
| B15–B25 | order candidate/identity matching and missing/external orders | unchanged order reconciliation | existing order comparer suite + A110 missing-order tests |

## Calls and live bindings

| Callee | Purpose | Contract | Evidence |
|---|---|---|---|
| `riskcalc.CanonicalDecimal` | raw finite validation | before zero/external classification | B8 |
| `canonicalDecimal` | exact spelling | unreadable remains visible | A110 raw tests |
| `isZeroDecimal`/`quantitiesAgree` | tolerance-zero classification | exact zero + finite one-ULP artifact only | B9–B14 |
| order identity helpers | canonical opaque-order matching | unchanged six-field vocabulary | B15–B25 |

## State mutations and fallbacks

- Pure comparison result construction; invalid present quantities are carried as visible mismatch spellings so promotion validation rejects their streak evidence.
- Only a truly absent local position plus a valid positive broker holding is external/nonblocking.

## Safety conclusion

- Safe boundary: raw quantity validation precedes exposure classification; order logic remains unchanged.
- High-risk impact: yes; a false clean or external classification bypasses the entry interlock.
