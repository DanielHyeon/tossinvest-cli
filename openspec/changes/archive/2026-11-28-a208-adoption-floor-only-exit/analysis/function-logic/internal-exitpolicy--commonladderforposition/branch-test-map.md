# Branch Test Map: `CommonLadderForPosition`

편집 뒤 좌표(a208 F3 재추출 · AST 6 분기, `internal/exitpolicy/common_policy.go` 106–131). B1 이 새 갈래, 옛 B1~B5 → B2~B6.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 편입 보호선 사본 ID → HYBRID_50 표 + 부분익절 0, 비편입 거부 | `TestA208FloorOnlyCopyKeepsTheHybridFloors` · `TestA208FloorOnlyCopyIsAdoptionOnly` | yes — red-f3.txt(KeepsTheHybridFloors FAIL) | yes — 선택 시험 통과(F3) |
| B2 | 미등록 ID → 오류 | 미측정 | n/a | 미측정 |
| B3 | 편입 RUNNER → 부분익절 0 사본 | `TestExternalRunnerHasNoAutomaticPartials` | n/a | yes — 선택 시험 통과(F3) |
| B4 | rung 순회 | `TestExternalRunnerHasNoAutomaticPartials` | n/a | yes — 선택 시험 통과(F3) |
| B5 | Identity 실패 | 미측정(도달 불가) | n/a | 미측정 |
| B6 | 편입 RUNNER digest 표류 | 미측정(도달 불가) | n/a | 미측정 |

기존 갈래 무변 고정: `TestA208AdoptedHybridRowsKeepTheirMeaning`(편입 HYBRID_50 행 = 등록 표 · 1.0.0 정체성).
