# Branch Test Map: `CommonLadderForPosition`

편집 전 좌표(AST 5 분기, `internal/exitpolicy/common_policy.go` 106–127). 「미측정」 = 덮는 시험을 이름으로 특정하지 않았음.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 미등록 ID → 오류 | 미측정 | n/a | no |
| B2 | 편입 RUNNER → 부분익절 0 사본 | `TestExternalRunnerHasNoAutomaticPartials` | n/a | no |
| B3 | rung 순회 | `TestExternalRunnerHasNoAutomaticPartials` | n/a | no |
| B4 | Identity 실패 | 미측정(도달 불가) | n/a | no |
| B5 | 편입 RUNNER digest 표류 | 미측정(도달 불가) | n/a | no |

a208 F3 새 갈래: `TestA208FloorOnlyCopyKeepsTheHybridFloors`(표 · 트레일 동일, 부분익절만 0, 정체성 고정) · `TestA208FloorOnlyCopyIsAdoptionOnly`(비편입 거부) · `TestA208AdoptedHybridRowsKeepTheirMeaning`(편입 HYBRID_50 행 무변).
