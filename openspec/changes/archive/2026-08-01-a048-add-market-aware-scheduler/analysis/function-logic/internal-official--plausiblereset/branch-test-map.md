# Branch Test Map: `plausibleReset`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | zero endpoints are implausible | parser zero-time tests | existing reset zero guard | yes |
| B2 | exactly -1m and +24h are plausible | official parser boundary table | bounds not reused by scheduler | yes |
| B3 | one second outside and extreme epoch values are implausible without overflow | official parser/implausible tests | scheduler omitted bounds and used overflow-prone absolute delta | yes |
