# Branch Test Map: `ExitObserver.observe`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | retrier/source read fails | existing observation failure tests | existing | yes |
| B2 | non-positive quote is omitted | existing invalid quote tests | existing | yes |
| B3 | no valid symbols answered | existing empty answer test | existing | yes |
| B4 | fetched timestamp survives price normalization | snapshot provenance test | yes | yes |
