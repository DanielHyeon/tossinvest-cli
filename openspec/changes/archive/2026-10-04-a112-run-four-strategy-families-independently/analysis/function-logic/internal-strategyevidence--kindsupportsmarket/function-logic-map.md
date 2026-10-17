# Function Logic Map: `kindSupportsMarket`

- Source: `internal/strategyevidence/model.go`
- Post-L1a source SHA-256: `a67ab059e4cba377f0faaeb80f1c88821de2198707779b7175f901adc4b1819a` (frozen-base SHA was `c49652af…`; the only edit is the B2 case list)
- Signature: `kindSupportsMarket(kind EvidenceKind, market marketclock.Market) bool`
- Source range: `217:1`–`228:2`
- AST evidence: `ast.json`, regenerated 2026-08-17 after the L1a GREEN edit (5 branches, unchanged count).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- Exact switch on `kind`; `default` returns false so an unknown kind is refused by `normalizeHeader` (model.go:168) before any payload work.
- L1a change: the market-agnostic case (B2) now also lists `KindOfficialClosedBar1m` and `KindOfficialQuoteL1` (both KR and US). No other case changed; legacy kinds keep their exact market binding.

## Branches and early returns

- Exact AST return nodes: `220:3, 222:3, 224:3, 226:3`.

| Branch | AST kind | Source location | L1a disposition |
|---|---|---|---|
| B1 | switch | 218:2 | dispatch on kind — exercised by every envelope test |
| B2 | case (market-agnostic kinds incl. the two new breakout kinds) | 219:2 | **edited** — `TestKindSupportsMarketCoversBreakoutKindsInKRAndUS` (both new kinds × KR/US true; unknown kind false) |
| B3 | case KindKRNetFlow → KR only | 221:2 | unchanged — legacy `TestNewEnvelopeRejectsCrossMarketKindAndIncompleteIdentity` (model_test.go:40) |
| B4 | case KindUSParticipation → US only | 223:2 | unchanged — same legacy test |
| B5 | default false | 225:2 | unchanged — `TestKindSupportsMarketCoversBreakoutKindsInKRAndUS` (unknown kind → false) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| (no call expressions in this function) | — |

## State mutations and fallbacks

- None. Pure function; no assignments, defers or goroutines (AST).

## Safety conclusion

- The edit widens only the market-agnostic case; a mutant removing either new kind from B2 fails `TestKindSupportsMarketCoversBreakoutKindsInKRAndUS` (implementer report). Legacy canonical bytes/digest vector `7e30f2af…` unchanged (`TestLegacyKindCanonicalBytesAndDigestAreUnchanged`).
