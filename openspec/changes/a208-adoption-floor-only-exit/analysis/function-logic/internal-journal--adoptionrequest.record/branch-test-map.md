# Branch Test Map: `AdoptionRequest.record`

편집 뒤 좌표(a208 F3 재추출 · AST 13 분기, `internal/journal/adoption.go` 357–427). 분기 수 무변(B6 조건만 확장, 주석 한 줄로 B6 이후 +1 줄).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `switch` 필수 필드 | — | n/a | 미측정 |
| B2 | PositionID 없음 | 미측정 | n/a | 미측정 |
| B3 | Symbol/Market 없음 | 미측정 | n/a | 미측정 |
| B4 | ObservedAt 없음 | 미측정 | n/a | 미측정 |
| B5 | `if a.ExitPolicyID != ""` | `TestA208OffAdoptionIsUnchanged` | n/a | yes — 선택 시험 통과(F3) |
| B6 | `if _, ok := CommonPolicyByID(...); !ok && !IsAdoptedFloorOnlyPolicyID(...)` — **a208 F3 조건 확장** | `TestA208JournalRefusesOtherUnregisteredPolicies` | yes — red-f3.txt(사본 ID 수용 쪽은 RED 때 픽스처 결함으로 먼저 실패 — 변이 M5 로 의미 확인) | yes — 선택 시험 통과(F3) |
| B7 | 수량 정규화 실패 | 미측정 | n/a | 미측정 |
| B8 | 수량 0 | 미측정 | n/a | 미측정 |
| B9 | 관측가 실패 | 미측정 | n/a | 미측정 |
| B10 | 합성 손절 실패 | 미측정 | n/a | 미측정 |
| B11 | 비교 실패 | 미측정 | n/a | 미측정 |
| B12 | 손절 ≥ 관측가 | 미측정 | n/a | 미측정 |
| B13 | `if a.CostBasis != ""` | 미측정 | n/a | 미측정 |
