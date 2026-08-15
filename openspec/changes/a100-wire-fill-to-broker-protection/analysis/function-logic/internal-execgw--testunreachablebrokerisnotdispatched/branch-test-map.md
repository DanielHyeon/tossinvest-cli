# Branch Test Map: `TestUnreachableBrokerIsNotDispatched`

Source: `internal/execgw/gateway_test.go` (727-757).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Fixture construction unexpectedly rejects its valid local journal/gateway dependencies | `TestUnreachableBrokerIsNotDispatched` | no — construction is the prerequisite | yes — focused package test |
| B2 | A transport that rejects token acquisition before write must produce `NOT_DISPATCHED` | `TestUnreachableBrokerIsNotDispatched` | yes — the former released-port fixture could receive a reused-port HTTP response in CI | yes — focused package test |
| B3 | The synthetic transport observes more or fewer than one request | `TestUnreachableBrokerIsNotDispatched` | yes — any retry or extra request trips the assertion | yes — focused package test |
| B4 | The sole observed request is not `POST /oauth2/token` | `TestUnreachableBrokerIsNotDispatched` | yes — a mutation request reaching transport trips the assertion | yes — focused package test |

The B2-B4 assertions are deliberately tied to the same test: the recorder is local to that invocation, so the test proves both the transport boundary and the absence of a subsequent order POST without external ports, sockets, or scheduler timing.
