# Lane-performance hard map

- Captured: 2026-08-04
- Implementation base: `c60fee070389557413e610314248a232a648fc5c`
- Scope owner: `internal/performance` new leaf attribution files and a073 analysis/status/review/tasks only

## Hard evidence

- `internal/performance/model.go` owns the a049 deterministic candidate → lane → decision → order → fill →
  position → close lineage and explicit `complete`, `link_missing` and `not_measured` states.
- `internal/performance/store.go` owns the isolated `performance.db`, immutable replay checks, 90-day raw
  retention, 24-hour cadence and at most 500 raw rows per prune transaction.
- `internal/performance/query.go` groups existing complete samples by exact market, lane/version and
  policy/version; it never queries a broker or operating setting.
- The current lineage has no campaign/leg/close-leg identity. The existing single-row `Trade` also cannot
  prove signed fill-delta deduplication, partial-entry/staged-close quantity conservation or fee/tax/FX
  gross-to-net conservation.
- CodeGraph surfaced `internal/performance.Lineage` as the existing attribution entry point. No execution,
  journal writer, engine or gateway path is part of this implementation scope.

## New pure boundary

- Add a new immutable derived attribution store inside `internal/performance`; do not change existing
  function bodies or the persisted a049 schema.
- Input consists only of caller-supplied authoritative position/cost-policy evidence and immutable signed
  fill deltas. Construction validates and copies the evidence; query methods are read-only.
- Exact composite identity is market + candidate + lane/version + campaign/leg + decision/attempt + order +
  fill + position + close/close-leg + cost-policy/version. Ticker is display-only and never repairs identity.
- Deduplicate exact event/fill identity; refuse divergent replays. Correction/bust rows must cite the exact
  original fill, stay inside its composite lineage and use a signed negative delta.
- Require acquired = closed + authoritative residual quantity and authoritative entry basis = allocated
  close basis + residual basis. Missing links are `link_missing`, never symbol/time inferred.
- Source-currency gross is exit proceeds minus allocated basis. Net is gross minus evidenced entry fees,
  exit fees, taxes and persisted FX cost. Missing fee evidence keeps gross but makes net `not_measured`.
- Reporting-currency values additionally require persisted FX source/rate/as-of/quote currency and exact
  rounding policy/version. This identity evidence is required even when source and reporting currency names
  match; the projector does not invent an implicit rate of one. Missing FX is `not_measured`, never
  zero/current-rate fallback.

## Function Logic Map disposition

Not applicable for the pre-edit wave: implementation adds new leaf functions and tests only. No existing Go
function body is changed, so there is no existing-function branch map to refresh for this isolated scope.
