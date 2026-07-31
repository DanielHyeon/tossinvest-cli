# Function Logic Map: `TestTradingViewsCarryWholePageResponsiveAndFocusContracts`

- Source: `internal/console/trading_views_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| authenticated positions fixture | deterministic journal-backed HTML | console test harness | fail the test on setup/render error |
| required contracts | fixed responsive/focus token list | a043 accessibility contract | report every missing token |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | iterate required whole-page responsive/focus contracts | none | continue through all tokens | `TestTradingViewsCarryWholePageResponsiveAndFocusContracts` |
| B2 | rendered page lacks one contract | test diagnostic | `t.Errorf` | same test |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDashboardHarness` / `seedJournal` / `authenticate` | produce deterministic operator page | test-owned; no live account | AST calls |
| `strings.Contains` | exact static contract probe | absence is a test failure | B2 |

## State mutations and fallbacks

- Test-only harness state; production configuration, journal, and broker state are unchanged.

## Safety conclusion

- Safe edit boundary: rendered CSS/semantic markup assertions only.
- High-risk impact: no; this is a regression test for read-only UI contracts.
