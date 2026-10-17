# Branch Test Map: `NewReconcileDriver`

a095는 이 함수의 분기를 바꾸지 않는다. 편집(래치 map 두 개의 타입)은 분기 없는 초기값이며, 그 효과는 호출하는 쪽 시험
(`alertUnmanaged` · `checkExternalIncrease` 의 래치 시험 — tasks 2.12 · 2.17 · 3.4)이 잰다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 검증 switch | 기존 — 무변화 | n/a | n/a |
| B2 | journal 없음 | 기존 — 무변화 | n/a | n/a |
| B3 | collector 없음 | 기존 — 무변화 | n/a | n/a |
| B4 | tracker 없음 | 기존 — 무변화 | n/a | n/a |
| B5 | 조정 기록자 없음 | 기존 — 무변화 | n/a | n/a |
| B6 | 계정 없음 | 기존 — 무변화 | n/a | n/a |
| B7 | 편입 켜짐 ∧ 시세 읽기 없음 | 기존 — 무변화 | n/a | n/a |
| B8 | 공통 정책 id 있음 | 기존 — 무변화 | n/a | n/a |
| B9 | 모르는 공통 정책 | 기존 — 무변화 | n/a | n/a |
| B10 | 시계 없음 | 기존 — 무변화 | n/a | n/a |
