# Branch Test Map: `Store.PruneDue`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | completed drain within 24h skips | cadence test | existing | existing |
| B2 | each transaction deletes <=500 under 100ms target | bounded prune test | yes | yes |
| B3 | finite backlog drains and records cadence | drain test | yes | yes |
| B4 | bounded run with remaining/incoming backlog keeps cadence due | continuous influx/reschedule test | yes | yes |
| B5 | invalid cadence/transaction error fails closed | store error tests | yes | yes |
| B6 | malformed cadence fails closed | invalid metadata branch contract | no — defensive | yes |
| B7 | completed drain inside 24h skips | cadence test | no — existing | yes |
| B8 | cadence-read commit error propagates | DB error contract | no — defensive | yes |
| B9 | DELETE error rolls back | transaction contract | no — defensive | yes |
| B10 | row-count error propagates | driver contract | no — defensive | yes |
| B11 | backlog-check error rolls back | transaction contract | no — defensive | yes |
| B12 | only a drained cutoff writes cadence | bounded backlog test | yes | yes |
| B13 | cadence-write error rolls back | transaction contract | no — defensive | yes |
| B14 | commit error exposes no partial metadata | transaction/crash contract | no — defensive | yes |
| B15 | maximum deleted per tx never exceeds 500 | prune batch test | yes | yes |
| B16 | maximum per-tx lock duration stays below target | prune batch test | yes | yes |
| B17 | drain returns; remaining backlog reschedules | backlog/influx test | yes | yes |
