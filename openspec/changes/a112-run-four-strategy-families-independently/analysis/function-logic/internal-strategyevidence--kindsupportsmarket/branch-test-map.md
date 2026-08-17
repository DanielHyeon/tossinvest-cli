# Branch Test Map: `kindSupportsMarket`

- Source SHA-256: `a67ab059e4cba377f0faaeb80f1c88821de2198707779b7175f901adc4b1819a`; AST branch locations are authoritative (regenerated after the L1a edit).
- Lot L1a (2026-08-16/17) edited B2 only. RED was captured as a build failure on the undefined new kind constants against the unmodified file (implementer report), then GREEN.

| Branch | Scenario anchor | Required test disposition | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | switch at 218:2 | `TestKindSupportsMarketCoversBreakoutKindsInKRAndUS` (dispatch) | yes (build failure on new kinds) | yes |
| B2 | case at 219:2 | `TestKindSupportsMarketCoversBreakoutKindsInKRAndUS` — both new kinds × KR/US return true | yes | yes |
| B3 | case at 221:2 | `TestNewEnvelopeRejectsCrossMarketKindAndIncompleteIdentity` (legacy KR-only kind refused for US) | not-applicable (unchanged branch; legacy test kept green) | yes |
| B4 | case at 223:2 | `TestNewEnvelopeRejectsCrossMarketKindAndIncompleteIdentity` (legacy US-only kind) | not-applicable (unchanged branch; legacy test kept green) | yes |
| B5 | default at 225:2 | `TestKindSupportsMarketCoversBreakoutKindsInKRAndUS` (unknown kind → false) | not-applicable (unchanged branch) | yes |

Verification: `go test ./internal/strategyevidence -count=1` and `-race` green after the edit; adversary and gstack reviewers re-ran the suite read-only (review.md 2026-08-17 L1a sections).
