# Branch Test Map: `Client.Orderbook`

- Source: `internal/official/market_reads.go`, SHA-256 `b356524b48df5dc550c07d74c00f6f7c695e6364c78036e04df8df94ef2ce1ea`; branch IDs follow `ast.json` (generated 2026-08-18).
- AST counts: branches 1, returns 2, calls 3, defers 0, go statements 0. Source range `209:1`-`217:2`. Signature `(c *Client) Orderbook(ctx context.Context, symbol string) (domain.OrderBook, error)`.
- Citation-only bundle: this function is NOT edited by a112; its branch enumeration is evidence for the L1c brief. Any later body edit requires a fresh RED/BTM.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | transport/status/decode failure from `c.get` | not-applicable: `TestOrderbookIntegration` only serves a success body; the error path of this specific method has no test | n/a (not edited) | not-applicable |
| fall-through | 200 with a well-formed ladder; query carries `symbol=005930` | `TestOrderbookIntegration` (`market_reads_test.go:197-232`) asserts the query parameter and the adapted levels | n/a (not edited) | existing suite green |

Property the L1c brief cites: the request has exactly one parameter, so no test can pin a level count - depth is response data, not a contract constant.

Verification: `go test ./internal/official -count=1` green on 2026-08-18. No RED round applies - a112 does not edit this function.
