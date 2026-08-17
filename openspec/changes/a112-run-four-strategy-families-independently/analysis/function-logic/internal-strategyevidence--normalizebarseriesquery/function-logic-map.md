# Function Logic Map: `normalizeBarSeriesQuery`

- Source: `internal/strategyevidence/breakout_series.go`
- Source SHA-256: `ece27d1ede03408e1f819f5d65f42fca9a252dd3e693b4cbadf834bee4e9abc5` (current worktree; verified with `sha256sum` 2026-08-17)
- Signature: `normalizeBarSeriesQuery(query BarSeriesQuery) (BarSeriesQuery, marketclock.Market, error)`
- Source range: `175:1`–`207:2`
- AST evidence: `ast.json` generated 2026-08-17 (new function, not in the frozen base 016da624); branches 7, returns 7, defers 0, go statements 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Called only from `Store.SealBarSeries` (breakout_series.go:60). Validates and normalises the caller's `BarSeriesQuery` before any SQL: market parsed and upper-cased (`KR`/`US`), symbol trimmed + upper-cased and canonical (no `:`), session id trimmed and `<CAL>:<date>` for the market, interval exactly 60000, `MaxBars` in `[0, 512]` with 0 → 512, both cutoffs non-zero and converted to UTC.
- The normalised query is what `SealBarSeries` returns in `BarSeries.Query` and feeds into `barSeriesDigest`, so this function fixes the digest preimage's query part (`TestSealBarSeriesDigestGoldenVector`).
- Refusals are typed `ValidationError`s naming `bar_series_query.<field>`; on error the zero query and empty market are returned.

## Branches and early returns

Exact AST return nodes: `178, 184, 188, 191, 195, 202, 206`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 177:2 | `ParseMarket` error (unknown/empty market) → refuse `bar_series_query.market` | `TestSealBarSeriesRefusesInvalidQuery` /unknown market, /empty market |
| B2 | if | 183:2 | normalised symbol not canonical (blank, `:`) → refuse `.symbol` | `TestSealBarSeriesRefusesInvalidQuery` /empty symbol |
| B3 | if | 187:2 | `sessionDateFor` error (empty, foreign calendar) → refuse `.session_id` | `TestSealBarSeriesRefusesInvalidQuery` /empty session, /session calendar does not match market |
| B4 | if | 190:2 | `IntervalMS` ≠ 60000 → refuse `.interval_ms` | `TestSealBarSeriesRefusesInvalidQuery` /interval is not one minute |
| B5 | if | 194:2 | `MaxBars` < 0 or > 512 → refuse `.max_bars` | `TestSealBarSeriesRefusesInvalidQuery` /max bars above the hard cap, /negative max bars |
| B6 | if | 198:2 | `MaxBars == 0` → default 512 | `TestSealBarSeriesReturnsOrderedBarsAndDeterministicDigest` (asserts `Query.MaxBars == 512`), `TestSealBarSeriesRefusesMoreThanFiveHundredTwelveBars` (default bound refuses 513) |
| B7 | if | 201:2 | zero `EvaluationAt` or `IngestionCutoff` → refuse `bar_series_query` | `TestSealBarSeriesRefusesInvalidQuery` /zero evaluation clock, /zero ingestion cutoff |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `marketclock.ParseMarket(query.Market)` | 176 | case/space-insensitive market parse (`internal/clock`) |
| `strings.ToUpper`, `strings.TrimSpace` | 180–182 | normalisation of market/symbol/session |
| `canonicalSymbolText`, `sessionDateFor` | 183, 187 | the same identity predicates the header helper uses |
| `formatUint`, `strconv.Itoa` | 192, 196 | refusal text |
| `query.EvaluationAt.UTC()`, `query.IngestionCutoff.UTC()` | 204–205 | UTC cutoffs for `stamp` |

## State mutations and fallbacks

- Operates on its by-value `query` copy (AST: 8 assignments to `query.*` fields of the local copy); no package state, no defers/goroutines. The only defaulting is `MaxBars 0 → 512` (documented bound, not a fallback past a failure); everything else refuses.

## Safety conclusion

- All seven branches are pinned by named tests through `SealBarSeries` (the ten-case invalid-query table plus the default-bound assertion), and the bound `MaxBarSeriesBars = 512` was mutation-tested (512→513 killed). Refusal happens before any SQL is issued, so an invalid query never touches the store. No order/auth/runtime surface.
