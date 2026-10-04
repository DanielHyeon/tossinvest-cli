# M-B1 receipt addition (lot 0.7b.3, 2026-08-16): `candle-crawl.json`

Written by the collector after the candle loop in **both** outcomes, through the same
live-guarded `store.writeJSON` path as `preflight.json`, so it is a current-UID regular 0600
payload and a member of the sealed self-excluding `manifest.json`. It is never written on a
HOLD path (any loop error returns before it and the run is tainted).

```json
{"schema":"a112-mb-us-candle-crawl:v1","pages":4,"terminal":"cap_exhausted","last_cursor_sha256":"sha256:<64 hex>"}
{"schema":"a112-mb-us-candle-crawl:v1","pages":1,"terminal":"null"}
```

| Field | Value | Notes |
|---|---|---|
| `schema` | `a112-mb-us-candle-crawl:v1` | constant |
| `pages` | 1..4 | number of candle pages read and written as `candle-NN.*` |
| `terminal` | `null` or `cap_exhausted` | `null` only when the raw JSON cursor was `null` (unchanged `nextCursor` typing); `cap_exhausted` when four pages all returned valid non-null, non-looping cursors |
| `last_cursor_sha256` | `sha256:` + hex of the **decoded** page-4 cursor value bytes | present only on `cap_exhausted` (`omitempty`); reviewers cross-check it against `candle-04.meta.json.cursor_value_sha256`; the raw cursor JSON is separately preserved as `candle-04.cursor.raw.json` |

Reviewer checks (added to task 0.7b.4): entry present and sealed in `manifest.json`; `pages`
equals the count of `candle-NN.raw.json` files; `terminal` consistent with page-N cursor bytes;
digest cross-check above. Absence of the file with a sealed manifest is a P0 (the collector
must have written it before orderbook).

Observed on the first sealed receipt (run 4 B, 2026-08-16): `candle-crawl.json` 172 bytes,
`pages 4`, `terminal cap_exhausted`, `last_cursor_sha256` equal to `candle-04.meta.json.cursor_value_sha256`
(both reviewers recomputed it). Two schema facts to keep in mind when reading any receipt:
`orderbook.meta.json`/`calendar.meta.json` carry `cursor_json_sha256`/`cursor_value_sha256` equal to
the SHA-256 of empty bytes (`digestBytes(nil)` is non-empty, so `omitempty` never fires — code-consistent,
not a defect), and `*.headers.raw.json` serialises an absent `Retry-After` as JSON `null`.

Decision recorded with this lot: the page-4 `requireRemaining` gate still HOLDs when
`X-Ratelimit-Remaining < 1` even though no fifth candle request follows — kept as the
conservative reading of the spec sentence "rate gate 실패는 변함없이 HOLD"; observed values
on 2026-08-16 were 19/19/18/17 of 20, so this is not a practical constraint.
