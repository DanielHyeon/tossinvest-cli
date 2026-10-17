# a142 independent adversarial implementation review

2026-10-16. Reviewer: `/root/a142_adversarial_reviewer`, independent of Terra implementation. Baseline `b36798f0ec47ee9b0d0d26d5492d0eafe8dd5172`; reviewed production commit `c8ab315a`, final test repair `3aeabc4c`. Initial verdict: CHANGES_REQUIRED (test evidence). Final verdict: **PASS CODE** after closure of all three P2 findings below. This is not the additional gstack review or final completion gate.

## Actionable findings

1. **[P2] Nonempty partial-market failure is not exercised.** `internal/client/completed_orders_test.go:112`: first US response is `body:[]`, then KR fails. Thus the named test never accumulates an order before failure. A defect that only returns accumulated nonempty orders on a later failure could pass. Change fixture to a valid nonempty first-market order, assert both market requests occurred, and assert nil orders plus error for malformed and HTTP-failed second responses. Keep market sequence from the current adapter rather than changing it. Also include `{ "result": {} }` as the explicit missing-body case required by the frozen contract.

2. **[P2] Watchlist compatibility assertions leave core requirements unprotected.** `internal/client/watchlist_test.go:14`: all item ordering values are zero, groups have distinct values, and assertions omit Group, Name, positive Base/Last and USD output. Removing item sorting, losing group/name, or zeroing all positive prices could pass. Add shuffled non-equal item ordering and equal group ordering, then compare complete expected domain items (including both currencies and positive/zero/null/omitted price values). Preserve differing ticker/code fixture to protect `Symbol=code`. `:43` also omits explicit `items:null` and malformed later group after a valid populated group; add these with nil result/error assertions.

3. **[P2] High-risk consumer failure propagation lacks a concrete regression receipt.** Frozen tasks require affected consumer/error propagation checks. Current new tests stop at list methods; the existing FindOrder tests exercise successful fallback, not malformed pending data. Add an offline `HasPendingOrder` case proving unknown data returns error instead of authoritative false, and `FindOrder` case proving malformed pending data propagates error without consulting completed history. Do not modify those consumers' policy. Record exact test names in branch/evidence records rather than generic “existing client coverage”.

## Scope and source inspection

- Baseline diff confines production changes to account.go, completed_orders.go, watchlist.go and watchlist_manage.go. No engine, journal, Guardian, token/auth sharing, operation-ID, operational configuration or mutation method body change appears.
- Reviewed current AST records for the order readers and read-adapter changes alongside the diff. Collection guards reject nil decoded slices and leave explicit empty slice behavior intact. Wrong-type handling remains JSON decoder errors. Completed error paths still return nil aggregate/error; no new fallback policy introduced.
- Watchlist uses the existing bulk GET via ListWatchlistGroups, exact query parameters retained. Internal wire types contain ordering/currency; stable group/item ordering and product-code identity match revised design. Optional price conversion stays the existing helper. Read decoder extraction does not require domain expansion.
- Mutation methods and existing mutation tests are unchanged. Existing mutation contract test covers create/rename/delete; its pass must not be described as new exhaustive add/remove validation. Unchanged mutation bodies provide the narrow preservation evidence here.
- The strictness intentionally covers collection boundaries, not every malformed item or missing identity field; broader upstream item-count/positive-ID reconciliation rules remain excluded by the frozen design.

## Source hashes

| File | SHA-256 |
|---|---|
| internal/client/account.go | cda0d5d3cb12e91e55641aa66cc54f02f7e77ed48b57a4052b83c09d39b0a90f |
| internal/client/completed_orders.go | 2c389d043a94552ec5dd4e0179f3f5c3cdbd1eb26eaefec8cbbfcae5959cc486 |
| internal/client/watchlist.go | fc5440b66ddc096a9daeb2a8a125cf132d69bb69b7c191aaab864a44eb74f33b |
| internal/client/watchlist_manage.go | de925d43fbf32a246efaf6c6df56fbf6a2f7158d00188818dcbc6dbb39dca4f1 |

## Independent offline checks

- `go test ./internal/client -run 'Test(ListPendingOrders|ListCompletedOrders|ListWatchlist|WatchlistMutation|Watchlist.*XSRF|FindOrder)' -count=1`: PASS, 0.010s.
- `go test ./internal/ops ./cmd/tossctl -run 'Watchlist|Order.*(List|Pending|Completed)|FindOrder' -count=1`: cmd/tossctl PASS, 0.031s; internal/ops **no tests to run**, explicitly not consumer coverage evidence.
- No tests or production files written by reviewer. No live/account calls, settings, restart or deployment. Terra owns test repairs. Re-review repaired tests and final source hashes before changing this verdict.

## Re-review of test repairs (5246594a)

P2-1 closed: nonempty US result, explicit US/KR request sequence, schema and HTTP failure of the second market, nil output/error assertions, and missing body case are present. P2-3 closed: HasPendingOrder rejects unknown data; FindOrder rejects malformed pending data with zero non-pending requests. Independent focused suite including these tests passed (0.011s); four production hashes remain unchanged.

P2-2 remains open: revised groups have wire ordering `1,1,2`, so removing group sorting now passes; revised item order `3,1,2` has no item ties. The fix must retain all four dimensions: unsorted numeric groups, group ties, unsorted numeric items, item ties. Complete domain comparison now correctly checks group/name/prices/currency, and explicit null items/later-group failure are covered. Add omitted and explicit-zero prices alongside the nullable case to preserve the original requested optional-value distinctions. This is test-only repair; production logic need not change.

## Final re-review (3aeabc4c)

P2-2 now closed: wire group order `1,2,1` requires numeric sorting while preserving the tied groups' order; wire item order `3,1,1` requires numeric sorting and preserving tied items. Full expected-item equality checks Group, Symbol=code despite differing ticker, Name, KRW/USD currencies, positive prices, explicit zero, null and omitted prices. Explicit null-items and malformed later-group tests remain. Together with P2-1/P2-3 closure above, no actionable finding remains in reviewed scope.

Independently reran `go test ./internal/client -run 'Test(ListPendingOrders|ListCompletedOrders|ListWatchlist|WatchlistMutation|Watchlist.*XSRF|FindOrder|HasPendingOrder)' -count=1`: PASS, 0.015s. Rechecked all four production SHA-256 values against the table above: unchanged. Inspected the four local mutation raw logs referenced by `analysis/verification-receipt.md`; each contains the corresponding negative fixture failure. These mutation experiments were run by Terra, not independently rerun by this reviewer. Initial RED was not retained as a raw file; no independent historical-RED reproduction claim is made.

Final test SHA-256:

| File | SHA-256 |
|---|---|
| internal/client/orders_parse_test.go | 68c536332b40ea4d407232486ec7adf43f3dc902b3ec84e4a36d002b120fc182 |
| internal/client/completed_orders_test.go | edbc76c6f47787af02a153444a21ef7cc8b84ceff1c7835419fb1e2d8f67bf1a |
| internal/client/watchlist_test.go | de3ec778a43b365513e8d9ae4070f3bd19a039c6ac5c5bf397d46af9288e7bf7 |

Remaining limits: offline local-fixture review only; no claim of live WTS availability, deployment readiness or end-to-end account reconciliation. Existing mutation methods are unchanged; create/rename/delete/XSRF receipts are narrow coverage, not a new exhaustive mutation audit. Additional gstack review, full required gate and Manager completion verification remain required. No product/test code, operating configuration or live state was changed by this reviewer.

## Probe amendment implementation re-review (10be99cf)

**PASS CODE for the complete amended a142 scope.** Inspected the commit diff, registered probe test and current `wtsOperations` AST (source hash matches below). The only operation-registry source edit is the existing watchlist ProbeSpec: GET with exact bulk URL/query and empty body; status-200 check retained; JSON decode, required watchlists and each group's items arrays validated. Explicit empty arrays succeed. Operation ID/handler and all other probes, including the pre-existing watchlist-groups definition, remain unchanged. No new monitor schedule/configuration or runner-policy change.

Independently ran `go test ./internal/ops -run TestWatchlistProbe -count=1`: PASS, 0.003s. Test directly exercises the registered callback with missing/null/object collections, malformed later group, invalid JSON, explicit empty arrays and status 500. No HTTP/account request is made. A populated valid probe fixture is not explicitly present; the unchanged collection element content is intentionally opaque to this schema-only probe, and the populated reader contract has separate full-item coverage. This is a coverage limitation, not a blocking implementation finding.

All four client production hashes and all three client test hashes rechecked unchanged from the tables above. Additional final hashes:

| File | SHA-256 |
|---|---|
| internal/ops/wts_operations.go | e622a677b56a55caaf4eec189086219112c9b48ba3d4a49e22f690d00cdf1fd2 |
| internal/ops/wts_operations_watchlist_probe_test.go | a5e6a64d65490c0f395e80a8a5ac16c14be04e0c950554ba412372232109a0dd |

No outstanding actionable finding. Additional gstack amendment review, mandatory full gate and Manager verification remain separate completion requirements. Live correctness is not asserted.
