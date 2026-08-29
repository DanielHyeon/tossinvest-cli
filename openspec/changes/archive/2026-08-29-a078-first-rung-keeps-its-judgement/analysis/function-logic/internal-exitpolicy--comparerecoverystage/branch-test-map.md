# Branch Test Map: `compareRecoveryStage`

번호는 수정 **후**의 AST 기준이다. a062가 제거한 "한쪽만 NoRung이면 오류" 분기는
번호를 갖지 않으며, 그것이 사라졌다는 사실을 B1/B2가 함께 고정한다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 한쪽이라도 rung이 있으면 오류 없이 ladder 단계 비교로 들어간다 | `TestFirstRungActivationIsAForwardMove`, `TestARecomputationThatLostItsRungKeepsTheSavedCandidate` | yes | yes |
| B2 | rung 숫자 비교로 단계를 정한다 | 같음 | yes | yes |
| B3 | 재계산이 저장보다 낮은 rung이면 저장 후보가 앞선다 | `TestARecomputationThatLostItsRungKeepsTheSavedCandidate` | yes | yes |
| B4 | 재계산이 첫 rung을 활성화하면 재계산이 앞선다 | `TestFirstRungActivationIsAForwardMove` | yes | yes |
| B5 | 같은 rung은 동률이고, 그때 파생선이 다르면 거부된다 | `TestEqualStageWithDifferentDerivedLinesIsStillRefused` | no | pass |
| B6 | ratchet level을 순위로 바꾼다 | `TestRatchetLevelRankingIsUnchanged` | no | pass |
| B7 | 알려진 level은 순위를 얻는다 | 같음 | no | pass |
| B8 | 알 수 없는 level은 정체성 불일치로 거부한다 | `TestAnUnrankedRatchetLevelIsStillRefused` | no | pass |
| B9 | ratchet 순위 비교로 진입한다 | `TestRatchetLevelRankingIsUnchanged` | no | pass |
| B10 | 재계산 level이 낮으면 저장이 앞선다 | 같음 | no | pass |
| B11 | 재계산 level이 높으면 재계산이 앞선다 | 같음 | no | pass |
| B12 | 같은 level은 동률이다 | `TestRecoverySelectsOneWholeCoherentSnapshot` (기존) | no | pass |

RED observed = 수정 전 코드에서 실패함을 2026-08-03에 확인. 변이 검증 2건도 같은 날
수행했다(`review.md`).
