# Function Logic Map: `Store.SealBarSeries`

- Source: `internal/strategyevidence/breakout_series.go`
- Source SHA-256: `ece27d1ede03408e1f819f5d65f42fca9a252dd3e693b4cbadf834bee4e9abc5` (current worktree; verified with `sha256sum` 2026-08-17)
- Signature: `(s *Store) SealBarSeries(ctx context.Context, query BarSeriesQuery) (BarSeries, error)`
- Source range: `59:1`–`114:2`
- AST evidence: `ast.json` generated 2026-08-17 (new function, not in the frozen base 016da624); branches 11, returns 9, defers 0, go statements 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- New kind-scoped bounded ordered read (Manager decision 3); `SnapshotQuery`/`SealSnapshot`/`snapshotDigest` untouched. SELECT-only by construction and by structural test (`TestSealBarSeriesIsStructurallySelectOnly`); writes no snapshot row (`TestSealBarSeriesReturnsOrderedBarsAndDeterministicDigest` counts `snapshots` = 0).
- Scope: `evidence_kind = official_closed_bar_1m`, `market`, `symbol`, plus the `source_record_id` byte range `[<MARKET>:<SYMBOL>:<session>:60000:, upperBound)` (D2 — session scope carried by the record-id prefix); dual cutoff `source_event_at <= EvaluationAt AND source_available_at <= EvaluationAt` and `ingested_at <= IngestionCutoff` (decision 5). All values are bound SQL parameters; the only string concatenation is the static column list (`envelopeColumns("e")`).
- Each row is re-decoded through the strict decoder and header↔payload-bound by `scanClosedBarRecord`; one refusal aborts the whole read (no truncation, no partial series). Latest visible revision per `source_record_id` wins in memory (the NOT EXISTS supersedes subquery was removed in the fix round — in-memory max is the single authority).
- Output: bars sorted by `open_at_ms`, bound `len(bars) <= MaxBars` (default and cap 512, refusal above — no truncation), digest `barSeriesDigest(normalized, bars)` (domain-separated, golden vector `51d80380…`).
- Measured cost (review a2): 390-bar session 42.4 ms at 1 rev/bar, 168.4 ms at 4 rev/bar — recorded as an L3/L5 input; not for a per-minute hot path as-is.

## Branches and early returns

Exact AST return nodes: `62, 67, 80, 87, 96, 99, 108 (sort.Slice closure), 110, 113`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 61:2 | `normalizeBarSeriesQuery` refused → return the typed refusal | `TestSealBarSeriesRefusesInvalidQuery` (10 cases) |
| B2 | if | 66:2 | `textPrefixUpperBound` cannot form an upper bound → refuse | not-applicable: unreachable — the prefix always ends in `:` (0x3A < 0xFF) so the last byte increments (defensive) |
| B3 | if | 79:2 | `QueryContext` error → return driver error | not-applicable: SQLite driver failure; not injectable through the package's public API in these tests |
| B4 | for | 83:2 | iterate result rows | every series test with rows, e.g. `TestSealBarSeriesReturnsOrderedBarsAndDeterministicDigest` |
| B5 | if | 85:3 | `scanClosedBarRecord` refused a row → close rows, return refusal (whole read fails) | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift`, `TestSealBarSeriesRefusesAPayloadThatDisagreesWithItsHeader`, `TestSealBarSeriesRefusesRatherThanHidingADriftedMinute`, `TestSealBarSeriesRefusesACorrectionThatMovesTheBar` (all assert zero bars returned) |
| B6 | if | 90:3 | same record already seen with an equal/higher payload revision → skip this row (`continue`) | taken side pinned by `TestSealBarSeriesPicksTheHighestRevisionNumericallyNotTextually` (revisions r1..r11; textual scan order `r1,r10,r11,r2,…` asserted in-test; numeric max r11 / r10 under the two cutoffs; mutant neutralising the guard degrades the winner to r9 — killed); untaken side by `TestSealBarSeriesPrefersTheLatestVisibleRevision` and `TestSealBarSeriesCorrectionIsAppendOnlyAndReplayStable` |
| B7 | if | 95:2 | `rows.Close` error → return | not-applicable: driver failure path (see B3) |
| B8 | if | 98:2 | `rows.Err` error → return | not-applicable: driver failure path (see B3) |
| B9 | range | 102:2 | iterate winners into the slice | every series test with rows |
| B10 | if | 103:3 | `RegularSessionOnly` and the bar is extended-hours → skip | `TestSealBarSeriesRegularSessionFilter` (3 → 2 bars; digest changes) |
| B11 | if | 109:2 | more visible bars than `MaxBars` → refuse (no truncation) | `TestSealBarSeriesRefusesMoreThanFiveHundredTwelveBars` (513 refused with zero bars/empty digest, 512 accepted; mutant 512→513 killed) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `normalizeBarSeriesQuery` | 60:29 |
| `strings.Join` | 64:12 |
| `formatUint` | 64:94 |
| `textPrefixUpperBound` | 65:20 |
| `invalid` | 67:23 |
| `stamp` | 69:16 |
| `stamp` | 70:15 |
| `s.db.QueryContext` | 71:15 |
| `envelopeColumns` | 71:48 |
| `make` | 82:13 |
| `rows.Next` | 83:6 |
| `scanClosedBarRecord` | 84:28 |
| `rows.Close` | 86:8 |
| `rows.Close` | 95:12 |
| `rows.Err` | 98:12 |
| `make` | 101:10 |
| `len` | 101:37 |
| `append` | 106:10 |
| `sort.Slice` | 108:2 |
| `len` | 109:5 |
| `invalid` | 110:23 |
| `strconv.Itoa` | 111:23 |
| `len` | 111:36 |
| `strconv.Itoa` | 111:84 |
| `barSeriesDigest` | 113:58 |

### 손으로 쓴 주석 — 완전성 주장이 아니다

위 표가 `ast.json` 의 호출 전부이고 `tools/logic-map/role_check.py` 가 1:1 로 대조한다.
아래는 그 자리에 있던 손으로 쓴 분석이다. 줄 번호만 적거나 한 줄이 호출 여럿을 묶어서
기계가 읽지 못했고, 그래서 잘려 있어도 게이트가 조용했다(a112 4차 리뷰가 센 39 개 중 하나).
근거로서의 값은 남으므로 지우지 않는다. **좌표는 위 표가 정본이다** — 아래 산문의
줄 번호는 그때 손으로 읽은 값이고, 어긋나면 위 표가 맞다.

| Callee (hand-written note) | Source location | Evidence |
|---|---|---|
| `normalizeBarSeriesQuery(query)` | 60 | see `internal-strategyevidence--normalizebarseriesquery` |
| `strings.Join(...)`, `textPrefixUpperBound(prefix)` | 64–65 | record-id byte range for the session scope |
| `stamp(...)` ×2 | 69–70 | cutoff timestamps in the store's stamp format |
| `s.db.QueryContext(ctx, SELECT … WHERE kind/market/symbol/record-id range/dual cutoff ORDER BY source_record_id, revision_identity, evidence_id, params…)` | 71 | parameterised SELECT; scope predicates pinned by `TestSealBarSeriesIgnoresOtherKindsSymbolsAndSessions`, `TestSealBarSeriesScopeExcludesLaterSessionsOtherKindsAndOtherSymbols` (later session, legacy kind under the prefix, forged-record-id MSFT bar); cutoffs by `TestSealBarSeriesDualCutoffIndependence` |
| `scanClosedBarRecord(rows, normalized)` | 84 | strict re-decode + header↔payload binding; see `internal-strategyevidence--scanclosedbarrecord` |
| `rows.Close`, `rows.Err` | 86, 95, 98 | resource release / iteration error |
| `sort.Slice(bars, by OpenAtMS)` | 108 | ordering; out-of-order inserts read back ascending (`TestSealBarSeriesReturnsOrderedBarsAndDeterministicDigest`) |
| `barSeriesDigest(normalized, bars)` | 113 | golden vector `TestSealBarSeriesDigestGoldenVector`, sensitivity `TestSealBarSeriesDigestChangesWithEveryInput`, domain separation `TestSealBarSeriesDigestIsDomainSeparatedFromSnapshotDigest` |

## State mutations and fallbacks

- No store mutation (SELECT only; structural AST guard forbids INSERT/UPDATE/DELETE/ExecContext/BeginTx in this function). Locals only: `winners` map, `bars` slice (AST 15 assignments, all local). No fallback: refusal returns `BarSeries{}` (asserted zero bars / empty digest in every refusal test); no truncation to `MaxBars`, no silent skip of a drifted row.

## Safety conclusion

- Read-only evidence access with fail-closed semantics: invalid query, any drifted/mismatched row, or a series above the bound refuses the whole read rather than returning a partial or reordered series. Append-only correction visibility under the dual cutoff is pinned (`TestSealBarSeriesCorrectionIsAppendOnlyAndReplayStable`: earlier-cutoff digest unchanged after `r2`). Untested branches are three driver-failure paths, one unreachable prefix guard, and the taken side of the revision `continue` (needs revision ≥ 10) — recorded above, none affects order/auth/runtime surfaces. Nothing in production calls this yet (L3/L5 wiring later; a2 cost note applies).
