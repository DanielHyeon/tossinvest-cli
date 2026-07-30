# Branch Test Map: `Updater.Inspect`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Branchless wrapper takes the updater mutex and returns one internally consistent current/candidate view | `TestUpdaterSerializesInspectStageAndInstall` plus existing inspection tests | inspect could observe half-published stage/install | pass |
