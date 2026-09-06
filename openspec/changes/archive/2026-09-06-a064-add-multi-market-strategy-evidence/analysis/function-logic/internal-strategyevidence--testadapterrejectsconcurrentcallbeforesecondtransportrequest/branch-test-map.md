# Branch Test Map: `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | the first fetch reaches the transport and blocks there | `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` | existing coverage | PASS |
| B2 | the concurrent second fetch is rate limited | `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` | existing coverage | PASS |
| B3 | exactly one transport call happened | `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` | existing coverage | PASS |
| B4 | releasing the first fetch completes it cleanly | `TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest` | existing coverage | PASS |
