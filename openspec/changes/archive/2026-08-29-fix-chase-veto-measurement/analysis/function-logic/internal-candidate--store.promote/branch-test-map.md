# Branch Test Map: `Store.Promote`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | market/symbol 없는 승격 | **커버 없음** | no | no |
| B2 | instant 없는 승격 | **커버 없음** | no | no |
| B3 | BeginTx 실패 | **커버 없음** (I/O) | no | no |
| B4 | readCandidate 실패 | **커버 없음** (I/O) | no | no |
| B5 | 시계 역행 승격 거부 | `TestABackwardClockStepDoesNotEndTheDiscoveryLoop` · `TestOneRejectedSymbolDoesNotAbortTheMarket` | — (기존 동작) | yes |
| B6 | 냉각 중 재승격은 first_seen_at을 유지한다 | `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry` | — (기존 동작) | yes |
| B7 | Exec 실패 | **커버 없음** (I/O) | no | no |
| B8 | Commit 실패 | **커버 없음** (I/O) | no | no |
| B9 | 만료 후 재승격은 provenance와 first_rank를 버린다 | `TestTheFirstRankFollowsFirstSeenAtThroughCoolingAndExpiry`(reset 확인) · `TestAMigratedCandidateBecomesMeasurableWhenItsLifeEnds`(두 자격 칼럼이 새 읽기로 채워짐) | yes (두 새 CASE 절이 없으면 후자가 낡은 자격으로 측정된다) | yes |

**정직한 커버리지 기록**: 두 새 CASE 절이 만료 시 NULL을 쓰는지를 **칼럼 값으로 직접**
읽는 테스트는 없다. `TestAMigratedCandidateBecomesMeasurableWhenItsLifeEnds`가 결과
(만료 후 새 읽기의 자격으로 측정 가능)로 확인하고, `TestNoteFirstRankKeepsTheStoredPositionWhateverIsOfferedNext`가
반대 방향(같은 생명 안에서는 덮이지 않음)을 고정한다. 두 테스트가 함께 절의 존재를
요구하지만 SELECT로 NULL을 확인하지는 않는다.
