# Function Logic Map: `Block.Covers`

- Source: `internal/reconcile/mismatch.go`
- Evidence: `ast.json`, `risk-pattern-report.md`

## Inputs and invariants

| Input | Valid range | Source | Failure behavior |
|---|---|---|---|
| block scope/market/symbol | account, market, symbol, defensive unknown | tracker projection | unknown or unrepresentable scope covers conservatively |

## Branches and early returns

| Branches | Condition | Result | Test |
|---|---|---|---|
| B1 | account | covers all | permanent gate tests |
| B2–B3 | market | case-insensitive market coverage | scope tests |
| B4–B5 | symbol with blank normalized symbol | covers all without changing key/scope | blank-symbol journal regression |
| B6 | ordinary symbol | case-insensitive symbol coverage | symbol gate tests |
| Return | unknown scope | covers all | defensive state-table tests |

## Calls and live bindings

| Callee | Purpose | Contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace`/`EqualFold` | canonical coverage | pure | AST |

## State mutations and fallbacks

- Pure projection. Blank symbol remains `ScopeSymbol` in memory to avoid colliding with the real permanent account key.

## Safety conclusion

- Safe boundary: make `Tracker.EntryAllowed` agree with the account-safe gateway projection.
- High-risk impact: yes; adoption calls this projection before price reads.
