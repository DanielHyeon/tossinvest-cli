# Branch Test Map: `blocksFor` (reconcile)

- Source: `internal/reconcile/mismatch.go` (1439-1472). 커버 측정: `../../coverage/after-branch-counts.txt`(cover mode set — 1 = 실행됨). 편집 전 측정 `../../coverage/base-branch-counts.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 기존 분기 | 기존 시험(패키지 GREEN) | 해당 없음(기존 분기) | yes — cover cond=1 body=1 |
| B2 | missing-order 블록 | `TestMissingOrderRaisesASymbolBlock`, `TestA192MissingOrderBlockArmsThenReleasesAfterTheTerminalProof` | yes — `../../red/red-1-reconcile.txt`(단언 실패, 컴파일 실패 아님) | yes — cover cond=1 body=1 |
