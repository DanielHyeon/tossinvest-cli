# Branch Test Map: `TestSystemClockNowIsUTC`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `internal/clock/clock_test.go:19`: preserve frozen-base UTC and wall-nearness assertions while monotonic behavior is tested separately | `TestSystemClockNowIsUTC` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
| B2 | `internal/clock/clock_test.go:22`: preserve frozen-base UTC and wall-nearness assertions while monotonic behavior is tested separately | `TestSystemClockNowIsUTC` | intentional A111 RED before the corresponding production correction | focused A111 suite GREEN |
