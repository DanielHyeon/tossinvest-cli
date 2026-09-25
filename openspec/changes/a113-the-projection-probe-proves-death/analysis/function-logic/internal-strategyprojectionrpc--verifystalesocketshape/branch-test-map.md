# Branch Test Map: `verifyStaleSocketShape`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | group/other 비트 · 일반 파일 | `TestStartRefusesUnsafeLeftoverShapes` 두 행 (a108) | no(회귀 핀) | yes |
| B2 | hard link | `TestStartRefusesUnsafeLeftoverShapes/socket에_hard_link가_걸려_있다` (a108) | no(회귀 핀) | yes |
