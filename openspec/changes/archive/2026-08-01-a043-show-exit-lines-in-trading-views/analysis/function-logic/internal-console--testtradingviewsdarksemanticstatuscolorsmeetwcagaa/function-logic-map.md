# Function Logic Map: `TestTradingViewsDarkSemanticStatusColorsMeetWCAGAA`

- Source: `internal/console/trading_views_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `pageTemplates` | inline stylesheet contains a primary dark block followed by dark/mobile block | console template constant | fail if either boundary is absent |
| semantic contracts | `.ok #22c55e`, `.bad #f43f5e` inside primary dark media only | StockOS lane-console dark palette | fail on selector drift or contrast below 4.5:1 |
| dark section background | exact `#1d1d22` | console dark section token | used as WCAG comparison background |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | primary dark media is absent | test diagnostic | fatal |
| B2 | following dark/mobile media boundary is absent | test diagnostic | fatal |
| B3 | iterate fresh and stale/error semantic contracts | none | continue through both contracts |
| B4 | exact selector/token is not inside primary dark media | test diagnostic | error |
| B5 | computed contrast is below 4.5:1 | test diagnostic | error |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.Index` / `strings.Contains` | isolate dark CSS and pin scoped selectors | missing boundary/selector fails closed | B1-B4 |
| `wcagContrastRatio` | calculate WCAG relative-luminance ratio | deterministic local helper | B5 |

## State mutations and fallbacks

- Test only; it cannot change rendered state or operating authority.

## Safety conclusion

- Safe edit boundary: approved dark semantic tokens and their measured contrast.
- High-risk impact: no production mutation; high user-visibility regression coverage.
