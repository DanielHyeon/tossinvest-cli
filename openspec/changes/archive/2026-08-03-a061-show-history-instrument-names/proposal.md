# a061 · Show instrument names in trade history

## Problem

`/history` renders only the journal's symbol. US tickers such as `IONQ` look like
a readable name, while KR symbols such as `272210` are opaque numeric codes. The
journal deliberately stores market and symbol, not mutable instrument metadata,
so both completed trips and exit events lack a display name.

## Change

- Add one narrow read-only instrument-name seam to the console.
- Resolve the unique `market + symbol` references already present on the history
  page through the official batch stock-metadata endpoint.
- Render `symbol · name` in both completed trips and exit events. If metadata is
  unavailable, preserve the symbol and state that name enrichment failed.
- Keep the frozen performance values, journal schema, inputs, operating toggles,
  and order capabilities unchanged.

## Non-goals

- Persisting names in the journal or backfilling historical rows.
- Recomputing any trade outcome or exit evidence.
- Adding search, filters, forms, scripts, or a write route.
