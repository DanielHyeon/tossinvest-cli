# Branch Test Map: `validateTypedPayload`

- Source SHA-256: `a67ab059e4cba377f0faaeb80f1c88821de2198707779b7175f901adc4b1819a`; AST branch locations are authoritative (regenerated after the L1a edit).
- Lot L1a (2026-08-16/17) inserted B3–B5. RED for the new dispatch was captured as a build failure on the undefined kinds/decoders against the unmodified file, then the named tests below failed/passed as recorded in review.md.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 402:2 | not-applicable (unchanged decode-error branch; canonical input from NewEnvelope) — legacy suite green | not-applicable | yes |
| B2 | if at 405:2 | `TestClosedBarRejectsSecretLikeField` (breakout payload with a secret-like key refused here, before the dispatch) | yes | yes |
| B3 | switch at 410:2 | `TestClosedBarDispatchRunsBeforeTheLegacyTypeMap` (refusal detail is the strict decoder's, never the legacy map's) | yes | yes |
| B4 | case at 411:2 | `TestClosedBarRejectsUnknownField`, `TestClosedBarRejectsUnknownEnumValues`, `TestClosedBarRejectsDecimalPriceNumber`, `TestClosedBarRejectsMinorThatDisagreesWithRawDecimal`, `TestClosedBarRequiresTheSessionCalendarDay`, `TestClosedBarRequiresObservationAfterTheBarClosed`, `TestClosedBarRequiresSuccessorOpenAt`, `TestClosedBarEnvelopeAcceptsCanonicalUSAndKRBars` | yes | yes |
| B5 | case at 414:2 | `TestQuoteL1RejectsEveryContractViolation`, `TestQuoteL1RejectsUnknownEnumValues`, `TestQuoteL1EnvelopeAcceptsCanonicalQuote` | yes | yes |
| B6 | range at 419:2 | not-applicable (unchanged legacy map) — `TestLegacyKindCanonicalBytesAndDigestAreUnchanged` + legacy suite | not-applicable | yes |
| B7 | if at 421:3 | not-applicable (unchanged legacy) | not-applicable | yes |
| B8 | switch at 425:3 | not-applicable (unchanged legacy) | not-applicable | yes |
| B9 | case at 426:3 | not-applicable (unchanged legacy) | not-applicable | yes |
| B10 | case at 428:3 | not-applicable (unchanged legacy) | not-applicable | yes |
| B11 | case at 430:3 | not-applicable (unchanged legacy) | not-applicable | yes |
| B12 | if at 433:3 | `TestClosedBarDispatchRunsBeforeTheLegacyTypeMap` (mutant moving the dispatch below the map makes a breakout payload hit this message — killed) | yes (mutant) | yes |

Verification: `go test ./internal/strategyevidence -count=1` / `-race`, consumers, `go build ./...` green; two independent reviewers reproduced (review.md 2026-08-17 L1a sections).
