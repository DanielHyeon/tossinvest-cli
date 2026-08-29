# Function Logic Map: `BlockRules`

- Source: `internal/reconcile/mismatch.go`
- Evidence: `ast.json`, `risk-pattern-report.md`

## Inputs and invariants

| Input | Valid range | Source | Failure behavior |
|---|---|---|---|
| none | static state table | reconciliation specification | missing row makes release ownership unauditable |

## Branches and early returns

| Branch | Condition | Result | Test |
|---|---|---|---|
| Return | static construction | recovery, ordinary, same-dispute permanent, broker-unknown rows | block-rule table tests |

## Calls and live bindings

| Callee | Purpose | Contract | Evidence |
|---|---|---|---|
| none | static data | no I/O | AST |

## State mutations and fallbacks

- No mutation. The permanent condition text now names one canonical dispute rather than pooled failures.

## Safety conclusion

- Safe boundary: wording aligned to unchanged reason/scope/release values.
- High-risk impact: no runtime branch; high-value audit contract.
