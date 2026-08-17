# Function Logic Map: `strictMinuteWalk`

- Source: `internal/official/strict_minute_candles.go`
- Source SHA-256: `d32181a939f298db306f492b488468b5925ac0ba97dad3f82cb1cb3286254ced` (current worktree; `sha256sum` verified equal to `ast.json` `source_sha256`, 2026-08-18)
- Signature: `strictMinuteWalk(decoder *json.Decoder, depth int) error`
- Source range: `495:1`–`543:2` (ast.json `start`/`end`)
- AST evidence: `ast.json` regenerated 2026-08-18 against the decision-30 sources; branches 15, returns 11, calls 13, assignments 10, defers 0, go statements 0.
- Disposition: New function (lot L1b, not in the frozen base 016da624); AST regenerated 2026-08-18 against the decision-30 sources; branch enumeration is the evidence for the L1b acceptance record.
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- The single whole-body authority for JSON shape: it recurses through the response once and refuses a duplicate object key **at any depth**, an unreadable token stream, and nesting deeper than the bound. Its caller `strictMinuteCheckJSON` adds the trailing-value rule after the walk returns.
- Why it exists: `strictMinuteObject` only sees one level, so a duplicate key inside a candle — or inside an envelope sibling the reader otherwise ignores — would pass unseen. The test `duplicate key inside an ignored sibling` is exactly that case, and only this walk can catch it.
- Depth bound D5: `strictMinuteMaxDepth = 16`, counted over boxes only (objects and arrays), with the envelope itself at depth 0 — so 17 nested levels refuse and 16 are accepted. The 2 MiB body cap alone would not bound the recursion stack; the adversary challenged D5 as "right value, no test" and ruling 29 added both edges.
- The walk validates shape only. Key semantics, string-ness, currency, timestamps and ordering all belong to the layers above it.

## Branches and early returns

Exact AST return nodes: `498`, `502`, `506`, `514`, `518`, `521`, `525`, `531`, `537`, `540`, `542`.

| Branch | AST kind | Source location | Meaning (one short clause) | Test disposition |
|---|---|---|---|---|
| B1 | if | 497:2 | the next token could not be read → return the decoder error | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `empty body`, where the first token is `io.EOF`) |
| B2 | if | 501:2 | the token is a scalar, not a delimiter → this leaf is fine | every accepting body test, e.g. `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage` |
| B3 | if | 505:2 | D5 depth bound: more than 16 nested boxes → refuse | refused at 17: `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `seventeen levels of nesting`); accepted at 16: `TestStrictMinuteCandlesAcceptsNestingUpToTheBound` |
| B4 | switch | 508:2 | dispatch on the delimiter kind | every accepting body test |
| B5 | case | 509:2 | an object: track the keys seen at this level | `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage` |
| B6 | for | 511:3 | walk the object's members | `TestStrictMinuteCandlesIgnoresUnknownEnvelopeKeys` (envelope with a `traceId` sibling) |
| B7 | if | 513:4 | the key token could not be read | untested: reachable only with syntactically malformed JSON inside an object (e.g. a missing colon); the fixtures use well-formed bodies apart from the empty one — recorded gap |
| B8 | if | 517:4 | the key token is not a string | not-applicable: inside an object `encoding/json` yields only string keys, and the closing delimiter is excluded by `More()` |
| B9 | if | 520:4 | this key was already seen at this level → `duplicate object key` | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `duplicate key in the envelope`, `duplicate key in the result`, `duplicate key in a candle`, `duplicate key inside an ignored sibling`) |
| B10 | if | 524:4 | a nested value refused → propagate | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtests `duplicate key in the result`, `duplicate key in a candle`, `duplicate key inside an ignored sibling`, `seventeen levels of nesting`) |
| B11 | case | 528:2 | an array | every body with a `candles` array |
| B12 | for | 529:3 | walk the array's elements | `TestStrictMinuteCandlesSendsTheCanonicalQueryAndReturnsThePage` (two candle elements) |
| B13 | if | 530:4 | an element refused → propagate | `TestStrictMinuteCandlesRefusesMalformedBodies` (subtest `duplicate key in a candle`) |
| B14 | case | 534:2 | any other delimiter → `unexpected JSON delimiter` | not-applicable: unreachable defence declared in the source comment at 535–536 — `encoding/json` emits only `{ [ ] }` and the closing pair is consumed by the `More()`/`Token()` pairing (ruling 29 required the comment) |
| B15 | if | 539:2 | the closing token could not be read → return the decoder error | untested: reachable only with a truncated body; the fixtures are well-formed apart from the empty one — recorded gap |

## Calls and live bindings

| Callee expression | Source location | Evidence |
|---|---|---|
| `decoder.Token()` ×3 | 496, 512, 539 | opening token, object keys, closing delimiter |
| `errors.New`, `strconv.Itoa(strictMinuteMaxDepth)` | 506 | the depth refusal names the bound it enforces |
| `make(map[string]struct{})` | 510 | the seen-key set is per object level, so the same key name may legitimately appear in sibling objects |
| `decoder.More()` ×2 | 511, 529 | member and element iteration bounds |
| `fmt.Errorf("duplicate object key %q", key)` | 521 | the refusal names the offending key |
| `strictMinuteWalk(decoder, depth+1)` ×2 | 524, 530 | the recursion; the only unbounded resource is stack depth, bounded by B3 |

## State mutations and fallbacks

- Locals only (10 AST assignments): `token`, `delimiter`/`ok`, `seen`, `keyToken`, `key`, `err`. The `*json.Decoder` is advanced — that is the shared state of the walk, and it is the caller's decoder, consumed exactly once per body. No package state, no I/O, no goroutines, no defers, no clock read.
- No fallback: the first violation aborts the whole walk and, through `strictMinuteDecode` B3, the whole read. Nothing is skipped or repaired.

## Safety conclusion

- This is the function that makes "the exact bytes" trustworthy: without it a duplicate key would let the standard decoder silently choose the last value, and the digest stored beside the bar would attest bytes whose meaning the reader never actually agreed on.
- The depth bound is a resource guard on a High-risk path — a body that is only deep, not large, would otherwise pass the 2 MiB cap and recurse freely. Both edges (16 accepted, 17 refused) are pinned.
- Two branches are declared not-applicable (the impossible non-string key, and the `default` arm whose unreachability the source comments at 535–536) and two are reachable-but-untested error paths on malformed token streams (B7, B15), recorded above rather than claimed as covered. None of them can admit a body: every one of them returns an error.
