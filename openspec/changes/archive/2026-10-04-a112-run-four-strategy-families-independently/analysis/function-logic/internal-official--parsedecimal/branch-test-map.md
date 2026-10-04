# Branch Test Map: `parseDecimal`

- Source: `internal/official/market_reads.go`, SHA-256 `b356524b48df5dc550c07d74c00f6f7c695e6364c78036e04df8df94ef2ce1ea`; branch IDs follow `ast.json` (generated 2026-08-18).
- AST counts: branches 2, returns 3, calls 1, defers 0, go statements 0. Source range `15:1`-`24:2`. Signature `parseDecimal(s string) float64`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1c brief. Any later body edit requires a fresh RED/BTM.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | empty/absent decimal string | `TestAdaptPricesEmpty` reaches the adapter with no rows, so B1 is exercised only indirectly; no test in this package asserts `parseDecimal("") == 0` directly | n/a (not edited) | existing suite green |
| B2 | non-numeric decimal string | not-applicable: no existing test in `internal/official` feeds a malformed decimal through this function - the fail-open `0` is untested | n/a (not edited) | not-applicable |
| fall-through | ordinary decimal string | `TestAdaptPricesUnit` (`"185.70"` -> `185.70`), `TestAdaptOrderbookUnit` (`"72100"` -> `72100`) | n/a (not edited) | existing suite green |

Property the L1c brief cites and no test covers: B2's swallowed error. That absence is the reason the quote producer may not read prices through this function.

Verification: `go test ./internal/official -count=1` green on 2026-08-18. No RED round applies - a112 does not edit this function.
