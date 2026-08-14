# Branch Test Map: `stableObservationID`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `internal/app/engine/exitloop.go:1094`: bind official identity to fetched-at and fallback identity to durable cycle:N | `TestA111FallbackObservationSourceIsDurableAndRestartMonotone` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B2 | `internal/app/engine/exitloop.go:1098`: bind official identity to fetched-at and fallback identity to durable cycle:N | `TestA111FallbackObservationSourceIsDurableAndRestartMonotone` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B3 | `internal/app/engine/exitloop.go:1101`: bind official identity to fetched-at and fallback identity to durable cycle:N | `TestA111FallbackObservationSourceIsDurableAndRestartMonotone` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B4 | `internal/app/engine/exitloop.go:1106`: bind official identity to fetched-at and fallback identity to durable cycle:N | `TestA111FallbackObservationSourceIsDurableAndRestartMonotone` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
