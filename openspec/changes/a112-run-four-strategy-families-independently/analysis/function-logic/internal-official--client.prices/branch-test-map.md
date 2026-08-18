# Branch Test Map: `Client.Prices`

- Source: `internal/official/market_reads.go`, SHA-256 `b356524b48df5dc550c07d74c00f6f7c695e6364c78036e04df8df94ef2ce1ea`; branch IDs follow `ast.json` (generated 2026-08-18).
- AST counts: branches 1, returns 2, calls 4, defers 0, go statements 0. Source range `142:1`-`150:2`. Signature `(c *Client) Prices(ctx context.Context, symbols []string) ([]domain.Quote, error)`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1c brief. Any later body edit requires a fresh RED/BTM.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | transport/status/decode failure from `c.get` | not-applicable: `TestPricesIntegration` serves only a success body | n/a (not edited) | not-applicable |
| fall-through | 200 with rows for the requested symbols | `TestPricesIntegration` (`market_reads_test.go:118-152`) | n/a (not edited) | existing suite green |
| fall-through, empty | 200 with an empty `result` | `TestAdaptPricesEmpty` covers the adapter half only; no test drives an empty body through this method | n/a (not edited) | not-applicable |

Property the L1c brief cites: a missing symbol is not an error on this path. The quote producer must supply its own refusal instead of inheriting this silence.

Verification: `go test ./internal/official -count=1` green on 2026-08-18. No RED round applies - a112 does not edit this function.
