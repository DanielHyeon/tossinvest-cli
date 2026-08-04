# Function Logic Map: `checkDailyLoss`

- Source: `internal/risk/chain.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| account equity | positive account-base money | authoritative account snapshot | zero/negative blocks daily-loss gate |
| realised loss | non-negative account-base magnitude | authoritative aggregate | malformed/negative refuses |
| absolute/ratio limits | account-base money and `(0,1]` | Guardian policy | reaching either ceiling blocks |
| market | KR or US, but does not select the aggregate currency | intent scope | unsupported market already rejected in preflight |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | equity currency/amount invalid | none | input unavailable | paired wrong-base snapshot test |
| B2 | equity decimal invalid | none | input unavailable | existing malformed equity test |
| B3 | equity ≤ 0 | none | daily loss reached | existing zero/negative test |
| B4 | loss invalid/negative/wrong base currency | none | input unavailable | existing magnitude + paired currency tests |
| B5 | absolute comparison invalid | none | input unavailable | policy mismatch test |
| B6 | absolute loss reaches cap | none | daily loss reached | paired KR/US base cap table |
| B7-B8 | loss or ratio parse invalid | none | input unavailable | existing policy validation tests |
| B9 | loss/equity reaches ratio | none | daily loss reached | existing ratio boundary + paired table |
| success | both limits remain unused | none | allowed | paired KR/US table |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Policy.LimitCurrency` | select the single account base independently of market quote | pure | a072 account-base Guardian decision |
| `moneyIn`/`magnitudeIn` | validate base snapshot money | pure; mismatch refuses | CodeGraph + AST |
| `riskcalc.WithinLimit` | preserve fail-closed aggregate ≥ boundary | pure; no retry | riskcalc contract |
| exact rationals | compare loss/equity without division rounding | pure | current HEAD |

## State mutations and fallbacks

- No FX conversion occurs here: authoritative account equity and realised loss must already be one account-base snapshot for both markets.
- No mutation, I/O, fallback, market-specific Guardian or per-market loss budget.

## Safety conclusion

- Safe edit boundary: currency selection changes from market quote to `Policy.LimitCurrency`; all boundary logic remains byte-for-byte equivalent.
- High-risk impact: yes. Reading US account-wide loss as USD under a KRW account cap could split or multiply permission, so mismatch fails closed.
