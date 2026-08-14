# Branch Test Map: `TestTheProtectionLineSurvivesAJudgementThatChangedNothing`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch `internal/console/a077_live_protection_test.go:88`: legacy expectation now depends on a true A111 heartbeat rather than a running-engine age bypass | `TestA111ConsoleConsumesTheSharedFreshnessVerdict` | intentional A111 RED before production change | asserted by focused A111 suite |
| B2 | AST branch `internal/console/a077_live_protection_test.go:90`: legacy expectation now depends on a true A111 heartbeat rather than a running-engine age bypass | `TestA111ConsoleConsumesTheSharedFreshnessVerdict` | intentional A111 RED before production change | asserted by focused A111 suite |
