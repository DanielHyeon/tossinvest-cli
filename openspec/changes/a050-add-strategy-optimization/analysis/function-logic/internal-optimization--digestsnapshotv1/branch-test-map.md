# Branch Test Map: `digestSnapshotV1`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path accepts an authentic v1 snapshot for atomic v2 re-digest and rejects a legacy digest mismatch | legacy digest migration/tamper tests | migration could re-sign unverified legacy metadata | PASS |
