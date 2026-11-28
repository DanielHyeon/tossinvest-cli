# Branch Test Map: `AdoptionRequest.record`

편집 전 좌표(AST 13 분기, `internal/journal/adoption.go` 357–426).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | switch | — | n/a | no |
| B2 | PositionID 없음 | 미측정 | n/a | no |
| B3 | Symbol/Market 없음 | 미측정 | n/a | no |
| B4 | ObservedAt 없음 | 미측정 | n/a | no |
| B5 | 정책 ID 있음 | `TestA208OffAdoptionIsUnchanged` | n/a | no |
| B6 | 미등록 정책 ID → 거부 | `TestA208JournalRefusesOtherUnregisteredPolicies` | n/a | no |
| B7 | 수량 정규화 실패 | 미측정 | n/a | no |
| B8 | 수량 0 | 미측정 | n/a | no |
| B9 | 관측가 실패 | 미측정 | n/a | no |
| B10 | 합성 손절 실패 | 미측정 | n/a | no |
| B11 | 비교 실패 | 미측정 | n/a | no |
| B12 | 손절 ≥ 관측가 | 미측정 | n/a | no |
| B13 | 원가 있음 | 미측정 | n/a | no |

a208 F3: 사본 ID 수용 `TestA208FloorOnlyAdoptionRecordsTheHybridCopy`(엔진 경유), 다른 미등록 ID 거부 `TestA208JournalRefusesOtherUnregisteredPolicies`.
