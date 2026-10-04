# Branch Test Map: `adaptOrderbook`

- Source: `internal/official/market_reads.go`, SHA-256 `b356524b48df5dc550c07d74c00f6f7c695e6364c78036e04df8df94ef2ce1ea`; branch IDs follow `ast.json` (generated 2026-08-18).
- AST counts: branches 2, returns 1, calls 12, defers 0, go statements 0. Source range `229:1`-`261:2`. Signature `adaptOrderbook(symbol string, raw apiOrderbook) domain.OrderBook`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1c brief and decision 32. Any later body edit requires a fresh RED/BTM.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | two ask levels; offers and `TotalOffer` built from them | `TestAdaptOrderbookUnit` (`market_reads_test.go:158-193`) asserts `TotalOffer == 11900` under the literal comment `// TotalOffer = 8500+3400` | n/a (not edited) | existing suite green |
| B2 | two bid levels; bids and `TotalBid` built from them | `TestAdaptOrderbookUnit` asserts `TotalBid == 9300` (`5200+4100`) | n/a (not edited) | existing suite green |
| B1/B2 empty | absent `asks`/`bids` | not-applicable: no existing test drives an empty ladder through this adapter | n/a (not edited) | not-applicable |
| whole body | end-to-end through `c.get` | `TestOrderbookIntegration` (`market_reads_test.go:197-232`) serves a body that **does** carry `"timestamp":"2026-03-25T09:30:00.123+09:00"` and asserts nothing about it - the field has no destination | n/a (not edited) | existing suite green |

Properties the L1c brief cites: (1) `TotalOffer`/`TotalBid` are adapter arithmetic, pinned as such by `TestAdaptOrderbookUnit`'s own comment - so their agreement with the visible levels is not evidence about book depth; (2) the response `timestamp` is dropped, proven by `TestOrderbookIntegration` supplying one that no assertion can reach; (3) `ProductCode == ""` is asserted by `TestAdaptOrderbookUnit`, which is what makes it usable as the official-path fingerprint in a console probe.

Verification: `go test ./internal/official -count=1` green on 2026-08-18. No RED round applies - a112 does not edit this function.
