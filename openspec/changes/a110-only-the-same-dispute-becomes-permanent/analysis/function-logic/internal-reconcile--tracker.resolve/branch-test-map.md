# Branch Test Map: `Tracker.Resolve`

| Branch | Scenario | Test | RED | GREEN |
|---|---|---|---|---|
| B1 | missing operator rejected | existing Resolve validation | preserve | yes |
| B2 | missing note rejected | existing Resolve validation | preserve | yes |
| B3 | journal-backed release path selected | operator release suite | preserve | yes |
| B4 | every active block receives exact-cause request | operator release suite | preserve | yes |
| B5 | journal failure keeps gate/permanent | operator failure suite | preserve | yes |
| B6 | known-nondurable blank pending is excluded from journal requests | A110 operator blank guard test | yes | yes |
| B7 | empty durable request batch is skipped safely | same test + Journal no-op mutation note | preserve | yes |
| Return | confirmed operator release clears runtime | durable permanent A110 test | preserve | yes |
