# Branch Test Map: `adaptPrices`

- Source: `internal/official/market_reads.go`, SHA-256 `b356524b48df5dc550c07d74c00f6f7c695e6364c78036e04df8df94ef2ce1ea`; branch IDs follow `ast.json` (generated 2026-08-18).
- AST counts: branches 1, returns 1, calls 6, defers 0, go statements 0. Source range `168:1`-`180:2`. Signature `adaptPrices(raw []apiPrice) []domain.Quote`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1c brief. Any later body edit requires a fresh RED/BTM.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | two rows, KRW integer and USD two-decimal | `TestAdaptPricesUnit` (`market_reads_test.go:85-104`) asserts `72000` and `185.70` and that `Name` stays empty | n/a (not edited) | existing suite green |
| B1 zero-iteration | empty `result` array | `TestAdaptPricesEmpty` (`market_reads_test.go:107`) | n/a (not edited) | existing suite green |
| whole body | end-to-end through `c.get` | `TestPricesIntegration` (`market_reads_test.go:118`) | n/a (not edited) | existing suite green |

Properties the L1c brief cites and no test covers: the response `timestamp` has no destination, and no test asserts that returned rows match the requested symbol set. Both are absences, recorded here so the quote brief does not treat the console path as an observation instrument.

Verification: `go test ./internal/official -count=1` green on 2026-08-18. No RED round applies - a112 does not edit this function.
