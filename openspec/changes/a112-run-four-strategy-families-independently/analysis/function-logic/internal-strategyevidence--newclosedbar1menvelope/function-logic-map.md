# Function Logic Map: `NewClosedBar1mEnvelope`

- Source: `internal/strategyevidence/breakout_bar.go`
- Source SHA-256: `ea18740bf672ced97c4bad9d5ed54ab0d4d447f10c6c03f12a9307487fccac0b` (current worktree; verified with `sha256sum` 2026-08-17)
- Signature: `NewClosedBar1mEnvelope(input ClosedBar1mInput) (Envelope, error)`
- Source range: `299:1`–`360:2`
- AST evidence: `ast.json` generated 2026-08-17 (new function, not in the frozen base 016da624); branches 6, returns 7, defers 0, go statements 0.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- The only supported write API for closed-bar evidence (review.md a1: header helpers are unexported; L1b's producer must use this constructor exclusively). Builds header (via `newClosedBar1mHeader`) and payload from one typed `ClosedBar1mInput`, so header↔payload mismatch is impossible through this path; `SealBarSeries` keeps refusing mismatches on read as second defence.
- Derives currency and scale from the market (`marketMoney`), recomputes every minor from the raw strings (`recomputeMinors`), copies the raw strings byte-for-byte, and writes payload clocks/identity from the header it just built (`open_at_ms = header.SourceEventAt`, `source_observed_at_ms = header.ObservedAt`, `symbol = header.Symbol`).
- Pins the constants `schema`, `interval_ms=60000`, `bar_label=open_at`, `finality=successor_observed`, `closed=true`; `successor_open_at_ms` is the structural successor-observed claim (decision 6; L1b must fill it from an observed successor).
- Ends in `NewEnvelope(header, payload)`, which re-runs the full strict decoder — every payload-level rule (`checkClosedBar1m`) applies to constructor output too.

## Branches and early returns

Exact AST return nodes: `303, 307, 310, 315, 326, 357, 359`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 302:2 | header helper refused (market/symbol/session/calendar/currency/revision/clocks/day) → refuse | `TestNewClosedBar1mEnvelopeRefusesInconsistentInput` (currency, session calendar, unknown market, revision zero, open_at off-minute, observed at/before open, empty symbol, empty calendar version), `TestClosedBarRequiresTheSessionCalendarDay` (constructor path) |
| B2 | if | 306:2 | `marketAndCode` error → refuse | not-applicable: unreachable — the same market already passed `marketAndCode` inside `newClosedBar1mHeader` at B1 (defensive) |
| B3 | if | 309:2 | zero `SuccessorOpenAt` → refuse | `TestNewClosedBar1mEnvelopeCarriesTheObservedSuccessor` /zero |
| B4 | if | 314:2 | `successorMS <= 0` (pre-1970 instant) → refuse | not-applicable: no fixture uses a non-zero pre-epoch instant; the other successor violations (`the bar itself`, `not on the minute`, `after the observation instant`) pass this guard and are refused by `checkClosedBar1m` B25–B27 through `NewEnvelope` — pinned by `TestNewClosedBar1mEnvelopeCarriesTheObservedSuccessor` (defensive) |
| B5 | if | 325:2 | `recomputeMinors` error (raw malformed / over-precise / signed / leading zero) → refuse | `TestNewClosedBar1mEnvelopeRefusesInconsistentInput` /over precise raw close, /fractional raw volume, /signed raw open |
| B6 | if | 356:2 | `json.Marshal` error → refuse | not-applicable: unreachable — the map holds only strings, `uint64`, `bool` and a nested string map (defensive) |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `newClosedBar1mHeader(...)` | 300 | header identity/clocks; see `internal-strategyevidence--newclosedbar1mheader` |
| `marketAndCode(input.Market)` | 305 | upper-case market code for the payload |
| `marketMoney(code)` | 317 | single scale table (KR→KRW/0, US→USD/4) |
| `recomputeMinors(KindOfficialClosedBar1m, ...)` → `minorFromRawDecimal` | 318 | integer recomputation of open/high/low/close at the market scale and volume at scale 0 |
| `json.Marshal(map[string]any{...})` | 328 | payload bytes; canonicalised again inside `NewEnvelope` |
| `NewEnvelope(header, payload)` | 359 | canonical JSON + `validateTypedPayload` → strict decoder; refusals for low>high, digest, calendar day, successor bounds surface here (`TestNewClosedBar1mEnvelopeRefusesInconsistentInput` /low above high, /malformed digest) |

## State mutations and fallbacks

- None. Locals only (AST 6 assignments); no defers/goroutines; no store or clock access. No fallback: any refusal returns `Envelope{}` (asserted `PayloadDigest() == ""` in the refusal test).

## Safety conclusion

- The constructor is the single supported producer seam and it derives scale, minors, identity and clocks from one input, so the header/payload mismatch class found by both reviewers in round 1 cannot be produced through it (`TestNewClosedBar1mEnvelopeDerivesScaleMinorsAndIdentityTogether` asserts constructor header == helper header and payload clocks == header clocks; `TestSealBarSeriesRoundTripsEnvelopesBuiltByTheConstructor` proves store round-trip). The three defensive branches (B2, B4, B6) are unreachable by construction; every other branch is pinned by a named test. No production caller exists yet (L1b/L3 wiring is a later lot).
