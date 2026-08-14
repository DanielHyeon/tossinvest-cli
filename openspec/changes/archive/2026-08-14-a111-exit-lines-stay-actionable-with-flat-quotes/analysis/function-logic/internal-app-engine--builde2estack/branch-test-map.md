# Branch Test Map: `buildE2EStack`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch `internal/app/engine/exit_e2e_test.go:258`: test-stack setup retains the observer and journal seams required to assert a durable heartbeat without a live order | `TestA111FlatAndInBandQuotesRefreshProvenanceWithoutHistoryOrOrders` | intentional A111 RED before production change | asserted by focused A111 suite |
| B2 | AST branch `internal/app/engine/exit_e2e_test.go:262`: test-stack setup retains the observer and journal seams required to assert a durable heartbeat without a live order | `TestA111FlatAndInBandQuotesRefreshProvenanceWithoutHistoryOrOrders` | intentional A111 RED before production change | asserted by focused A111 suite |
| B3 | AST branch `internal/app/engine/exit_e2e_test.go:272`: test-stack setup retains the observer and journal seams required to assert a durable heartbeat without a live order | `TestA111FlatAndInBandQuotesRefreshProvenanceWithoutHistoryOrOrders` | intentional A111 RED before production change | asserted by focused A111 suite |
| B4 | AST branch `internal/app/engine/exit_e2e_test.go:280`: test-stack setup retains the observer and journal seams required to assert a durable heartbeat without a live order | `TestA111FlatAndInBandQuotesRefreshProvenanceWithoutHistoryOrOrders` | intentional A111 RED before production change | asserted by focused A111 suite |
