# Branch Test Map: `storedFirstRank`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 자격을 갖춘 위치를 만든다 | `TestASymbolAlreadyHighInItsListWhenWeFirstSawItIsSeenLate` · `TestPruningRawObservationsLeavesTheFirstRankToo`(B9) · `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry`(B18) — 셋 다 `Measured`를 단언한다 | yes (두 필드를 지우면 셋이 미측정으로 실패한다) | yes |
