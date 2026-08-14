# Branch Test Map: `TestAReleasedGenerationsQuarantineDoesNotCloseTheCurrentOne`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | AST branch `internal/console/a077_live_protection_test.go:183`: legacy lifecycle coverage retains its isolation assertion under the shared freshness renderer | `TestA111ConsoleHidesStoppedFutureSeedAndCorruptEvidence` | intentional A111 RED before production change | asserted by focused A111 suite |
