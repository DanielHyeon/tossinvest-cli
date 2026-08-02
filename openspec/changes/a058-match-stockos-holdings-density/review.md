# Review — a058 match StockOS holdings density

## Proposal freeze — 2026-08-03

Voices: Manager self-review, rendered design comparison, safety review.

### Evidence

- StockOS desktop measurement: 10 px headers, 12 px cells, 15/18 px line-height, 8×12 px padding, widths 355/72/122/130/201/143/129 px, 96 px holding row.
- TossOS desktop measurement: 14.4 px headers/cells, 23.04 px line-height, 12×10.4 px padding, seven equal 159 px columns, first holding row 268 px.
- Root cause: fixed-layout widths are assigned only to body `.position-row` cells while the header is the first layout row.

### Findings and decisions

1. **High — header/value axes do not match the reference.** Accepted: use a seven-column `colgroup` so header and body share width definitions.
2. **High — verbose primary diagnostics destroy scan density.** Accepted: keep concise safety verdicts visible; fold verbose reason/explanation text into the existing native detail disclosure.
3. **Medium — type and padding are materially larger than StockOS.** Accepted: scope compact 10/12 px typography and 8×12 px padding to `.positions-table` only.
4. **Safety — missing/stale evidence must remain fail-closed.** Accepted unchanged: no line inference, concise visible verdict plus `—`, raw evidence in details.
5. **Security — no new inputs, script, route, broker call, mutation, or CSP change.** Accepted.

Function Logic Map: not-applicable — production edits are limited to inline CSS/template constants; no existing Go function body is modified.

Status: approved for implementation.
