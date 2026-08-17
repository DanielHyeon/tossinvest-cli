# Function Logic Map: `strictMinuteObject`

- Source: `internal/official/strict_minute_candles.go`
- Source SHA-256: `441bed46f81bc928cab03d512b3ff1305c0c663cb1b58027986e2e91b739977d` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-17)
- Signature: `strictMinuteObject(source []byte) (map[string]json.RawMessage, error)`
- Source range: `432:1`–`465:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` generated 2026-08-17; branches 8, returns 8, calls 12, assignments 10, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST generated 2026-08-17 after GREEN; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- A pure extractor: one object's keys mapped to the exact raw bytes of their values (`bytes.Clone`, so no aliasing into the response buffer). It is called three times per body — envelope (`decode:256`), `result` (`decode:265`) and each candle (`strictMinuteCandle:325`).
- Duplicate keys and trailing values are **not** re-checked here. The fix round made `strictMinuteCheckJSON` the single authority for them (source comment at 428–431): with the rule in two places, deleting either one leaves the other masking it, and neither can then be proved by a test. That is the review's "masked by sibling" finding applied by construction.
- Consequence for reachability: every byte slice this function sees has already been walked and proven well-formed JSON by `strictMinuteCheckJSON`, and candle values are `bytes.Clone`d sub-values of that same body. Only the "is this value an object?" question can still fail.

## Branches and early returns

Exact AST return nodes: `436, 439, 445, 449, 453, 459, 462, 464`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 435:2 | the opening token could not be read → return the decoder error | not-applicable: the caller only passes bytes `strictMinuteCheckJSON` already proved well-formed (an empty or truncated body is refused at `strictMinuteDecode` B3 before this call) |
| B2 | if | 438:2 | the value is not an object → `value is not an object` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `body is not an object`, `result is not an object`, `candle is not an object`) |
| B3 | for | 442:2 | walk the object's members into the map | every accepting body test, e.g. `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage`, `TestStrictMinuteCandlesIgnoresUnknownEnvelopeKeys` |
| B4 | if | 444:3 | the key token could not be read | not-applicable: unreachable behind the whole-body walk (well-formed input; `More()` guarantees a following token) |
| B5 | if | 448:3 | the key token is not a string | not-applicable: inside an object `encoding/json` yields only string keys, and the closing delimiter is excluded by `More()` |
| B6 | if | 452:3 | the value could not be decoded | not-applicable: unreachable behind the whole-body walk |
| B7 | if | 458:2 | the closing token could not be read | not-applicable: unreachable behind the whole-body walk |
| B8 | if | 461:2 | the closing token is not `}` | not-applicable: the member loop ends only at the matching `}` |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `json.NewDecoder(bytes.NewReader(source))` | 433 | token-level read of the exact bytes, not a struct unmarshal |
| `decoder.Token()` ×3 | 434, 443, 457 | opening delimiter, keys, closing delimiter |
| `decoder.More()` | 442 | member iteration bound |
| `decoder.Decode(&value)` | 452 | value captured as `json.RawMessage` |
| `bytes.Clone(value)` | 455 | the stored value owns its bytes; no aliasing into the decoder's buffer |

## State mutations and fallbacks

- Locals only (10 AST assignments): `decoder`, `opening`, `delimiter`/`ok`, `fields`, `keyToken`, `key`, `value`, `closing`. No package state, no I/O, no goroutines, no defers, no clock read.
- Last-writer-wins on the map is *not* a fallback here: a body with a duplicate key never reaches this function, because `strictMinuteCheckJSON` refuses it first. That ordering is the whole reason the rule was removed from this function.

## Safety conclusion

- Pure, allocation-bounded extractor whose only judgement is "object or not". It cannot admit a value the contract layers above it would reject, and it cannot silently repair one.
- Six of its eight branches are declared not-applicable because the caller's ordering (size cap → UTF-8 → whole-body walk → extract) removes them. That declaration is a design property of the pipeline, not an untested claim about JSON: the walk itself is enumerated in `internal-official--strictminutewalk` and its duplicate-key and trailing-value rules are driven by seven body subtests.
- Recorded residual (review.md 2026-08-17): making the deep walk the single authority was the fix-round resolution of the masked duplicate/trailing mutant pair, so no test can distinguish the two implementations of that rule any more — by design, since only one now exists.
