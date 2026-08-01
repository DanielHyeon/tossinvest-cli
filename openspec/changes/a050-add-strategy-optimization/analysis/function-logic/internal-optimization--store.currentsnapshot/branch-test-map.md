# Branch Test Map: `Store.currentSnapshot`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | missing/corrupt control row fails | store corruption coverage | pointer digest absent | PASS |
| B2 | missing/corrupt pointed snapshot fails | snapshot digest corruption tests | baseline snapshot validation | PASS |
| B3 | valid-looking rollback of version without matching digest fails Read, Preview and Apply | `TestControlPointerRollbackTamperFailsReadPreviewAndApply` | pointer version could be rolled back independently | PASS |
