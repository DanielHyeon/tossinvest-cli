# Function Logic Map: `scanClosedBarRecord`

- Source: `internal/strategyevidence/breakout_series.go`
- Source SHA-256: `ece27d1ede03408e1f819f5d65f42fca9a252dd3e693b4cbadf834bee4e9abc5` (current worktree; verified with `sha256sum` 2026-08-17)
- Signature: `scanClosedBarRecord(rows scanner, query BarSeriesQuery) (ClosedBarRecord, string, error)`
- Source range: `118:1`–`173:2`
- AST evidence: `ast.json` generated 2026-08-17 (new function, not in the frozen base 016da624); branches 16, returns 17, defers 0, go statements 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Called only from `Store.SealBarSeries` (breakout_series.go:84) with a `sql.Rows` positioned on one `evidence_records` row already scoped by the SQL predicates, and with the normalised query (upper-case market/symbol, validated session, interval 60000).
- Rebuilds the envelope (`scanEnvelope` → `NewEnvelope` → canonical JSON + `validateTypedPayload` + stored-digest check) and decodes the payload strictly (`DecodeClosedBar1mPayload`), so any row reaching the switch already satisfies `checkClosedBar1m`. The switch then binds header ↔ payload ↔ query (fix round P1-2/P1-3/P1-4 + recheck g2/g4): the record id must equal `closedBar1mRecordID(payload.Market, payload.Symbol, payload.SessionID, payload.IntervalMS, header.SourceEventAt)` — which, because SQL scoped the record id to the query prefix, also proves market/symbol/session/interval agree with the query and closes identity moves disguised as corrections.
- Refusal = `invalid(RefusalIdentityMismatch, field, "evidence <EvidenceID> <detail>")` naming the offending row; the caller aborts the whole read. Returns `(record, header.SourceRecordID, nil)` on success; the record id is the winners key in the caller.

## Branches and early returns

Exact AST return nodes: `121, 126, 131 (mismatch closure), 135, 143, 145, 147, 149, 151, 153, 155, 157, 159, 161, 163, 165, 167`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 120:2 | `scanEnvelope` error (scan/stamp/`NewEnvelope`/stored-digest failure on a corrupted row) → return | not-applicable: needs out-of-band row corruption — the append-only trigger blocks UPDATE (`TestStoreSchemaTooNewAndAppendOnlyTriggers`) and no `SealBarSeries` fixture tampers a stored row (the drift tests store well-formed rows and drift them logically, which is refused at B5–B16, not here) |
| B2 | if | 125:2 | `DecodeClosedBar1mPayload` error → return | not-applicable: unreachable for a row that passed `scanEnvelope` (`NewEnvelope` already ran the same strict decoder on the same canonical bytes; both call `decodeClosedBar1mObject`) — defensive double decode (review a2 notes the cost) |
| B3 | if | 134:2 | `sessionDateFor(payload.Market, payload.SessionID)` error → mismatch | not-applicable: unreachable — the payload passed `checkClosedBar1m` B6 (same call, same arguments) inside B2's decode (defensive) |
| B4 | switch | 141:2 | header↔payload↔query binding | every read test |
| B5 | case | 142:2 | `SourceRecordID` ≠ record id derived from payload identity + header `SourceEventAt` → mismatch `source_record_id` | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /payload symbol differs from the header symbol, /a foreign session forged under our record prefix; `TestSealBarSeriesRefusesACorrectionThatMovesTheBar` (asserts `Field == "source_record_id"`); `TestSealBarSeriesRefusesAPayloadThatDisagreesWithItsHeader` (US header, KR payload) |
| B6 | case | 144:2 | `Authority` ≠ Toss Open API → mismatch | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header authority is not the official api |
| B7 | case | 146:2 | `SchemaVersion` ≠ v1 → mismatch | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header schema version is foreign |
| B8 | case | 148:2 | `IssuerIdentity` ≠ `<query market>:<query symbol>` → mismatch | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header issuer identity is foreign |
| B9 | case | 150:2 | `IssuerMappingVersion` ≠ `a112-bar-issuer-v1` → mismatch | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header issuer mapping version is foreign |
| B10 | case | 152:2 | `MarketEffectiveDate` ≠ payload session date → mismatch | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header market effective date is not the session date |
| B11 | case | 154:2 | `Unit` ≠ `minor` → mismatch | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header unit is not minor |
| B12 | case | 156:2 | `SourceAvailableAt` ≠ `SourceEventAt + interval` → mismatch | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /header availability is not the bar close (recheck: exact +60,000 ms unfinished bar refused at read) |
| B13 | case | 158:2 | payload `open_at_ms` ≠ header `SourceEventAt` → mismatch | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /payload minute drifts earlier while every other clock agrees; `TestSealBarSeriesRefusesRatherThanHidingADriftedMinute` (13:30 header / 19:30 payload) |
| B14 | case | 160:2 | payload `source_observed_at_ms` ≠ header `ObservedAt` → mismatch | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /payload observation instant differs from the header |
| B15 | case | 162:2 | payload `currency` ≠ header `Currency` → mismatch | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /payload currency differs from the header |
| B16 | case | 164:2 | header `RevisionIdentity` ≠ `r<payload revision>` → mismatch | `TestSealBarSeriesRefusesEveryHeaderPayloadDrift` /revision identity disagrees with the payload revision |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `scanEnvelope(rows)` | 119 | store.go:460 — column scan, stamp parse, `NewEnvelope` re-validation, stored-digest equality |
| `envelope.Header()`, `envelope.CanonicalPayload()`, `envelope.PayloadDigest()` | 123, 124, 170 | envelope accessors |
| `DecodeClosedBar1mPayload(canonical)` | 124 | canonical decode + `rejectSecretFields` + strict decoder |
| `sessionDateFor(payload.Market, payload.SessionID)` | 133 | session date for the `MarketEffectiveDate` equality |
| `closedBar1mRecordID(payload.Market, payload.Symbol, payload.SessionID, payload.IntervalMS, header.SourceEventAt ms)` | 139 | expected record id (identity binding) |
| `revisionIdentityFor(payload.Revision)` | 164 | `r<n>` |
| `mismatch(field, detail)` closure → `invalid(RefusalIdentityMismatch, …)` | 130 | typed refusal naming `EvidenceID` (asserted by the drift tests) |

## State mutations and fallbacks

- None. Locals only (AST 6 assignments); no defers/goroutines; no store writes. No fallback: any disagreement refuses the row and, through the caller, the whole read (no silent skip — `TestSealBarSeriesRefusesRatherThanHidingADriftedMinute`).

## Safety conclusion

- This function is the read-side second defence that closes the four round-1 P1 findings (off-day rows, header/payload drift, early visibility, moved-minute corrections): all twelve equality cases (B5–B16) are pinned one-per-case by the stage-aware drift table plus dedicated tests, and both reviewers re-verified refusal at read (review.md 2026-08-17 rechecks). The three untested branches are defensive re-checks of work already done by `scanEnvelope`/`checkClosedBar1m` (B2, B3) and the out-of-band corruption path (B1). Read-only; no order/auth/runtime surface.
