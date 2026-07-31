# Branch Test Map: `runConsole`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | existing startup failure | existing `cmd/tossctl` tests | baseline | yes |
| B2 | existing configuration validation failure | existing `cmd/tossctl` tests | baseline | yes |
| B3 | existing path/configuration failure | existing `cmd/tossctl` tests | baseline | yes |
| B4 | default path cannot resolve | existing unwired diagnostic/static guards | baseline | yes |
| B5 | explicit config directory selects its journal | `TestConsoleJournalPathFollowsTheEngineProfile/explicit` | yes (missing resolver) | yes |
| B6 | default profile selects data journal | `TestConsoleJournalPathFollowsTheEngineProfile/default` | yes (missing resolver) | yes |
| B7 | existing OpenAPI seam wiring branch | existing `cmd/tossctl` tests | baseline | yes |
| B8 | existing read-only journal open branch | existing `cmd/tossctl` tests | baseline | yes |
| B9 | existing option normalization branch | existing `cmd/tossctl` tests | baseline | yes |
| B10 | existing option normalization else branch | existing `cmd/tossctl` tests | baseline | yes |
| B11 | existing credential status branch | existing `cmd/tossctl` tests | baseline | yes |
| B12 | existing credential status else branch | existing `cmd/tossctl` tests | baseline | yes |
| B13 | existing bind-address branch | existing `cmd/tossctl` tests | baseline | yes |
| B14 | existing bind-address else branch | existing `cmd/tossctl` tests | baseline | yes |
| B15 | existing TLS-address branch | existing `cmd/tossctl` tests | baseline | yes |
| B16 | existing TLS-address else branch | existing `cmd/tossctl` tests | baseline | yes |
| B17 | existing origin-policy branch | existing `cmd/tossctl` tests | baseline | yes |
| B18 | existing origin-policy else branch | existing `cmd/tossctl` tests | baseline | yes |
| B19 | existing listen startup failure | existing `cmd/tossctl` tests | baseline | yes |
| B20 | existing TLS material branch | existing `cmd/tossctl` tests | baseline | yes |
| B21 | existing TLS material else branch | existing `cmd/tossctl` tests | baseline | yes |
| B22 | existing server failure branch | existing `cmd/tossctl` tests | baseline | yes |
| B23 | existing shutdown branch | existing `cmd/tossctl` tests | baseline | yes |
| B24 | existing lifecycle cleanup branch | existing `cmd/tossctl` tests | baseline | yes |
| B25 | existing terminal error branch | existing `cmd/tossctl` tests | baseline | yes |
