# Branch Test Map: `FlattenAuthorization.Consume`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | absent/mismatched scope, broker, or quantity | permit mismatch table | enum could be replayed anywhere | rejected |
| B2 | clock rollback or delayed replay | pre-issue/+1h tests | caller supplied time | rejected |
| B3 | same or copied permit consumed twice | copied permit test | replayable enum | exactly one succeeds |
