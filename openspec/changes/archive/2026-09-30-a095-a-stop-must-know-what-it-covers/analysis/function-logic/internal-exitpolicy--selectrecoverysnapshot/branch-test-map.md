# Branch Test Map: `SelectRecoverySnapshot`

- Source: `internal/exitpolicy/recovery.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:135` `if err := validateRecoverySnapshot(recomputed); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:138` `if saved == nil {` | 예 | **a095 5.3** — issues I1에 「저장 스냅샷이 없으면 비교 없음」으로 인용 | no | no |
| B3 | `:141` `if err := validateRecoverySnapshot(*saved); err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B4 | `:144` `if saved.PositionID != recomputed.PositionID \|\| saved.PositionGeneration != recomputed.PositionGeneration \|\|` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:151` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:155` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:159` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B8 | `:162` `if stage == 0 && (saved.NextTarget != recomputed.NextTarget \|\| saved.NextProtection != recomputed.NextProtection) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B9 | `:165` `if protection >= 0 && high >= 0 && stage >= 0 {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B10 | `:168` `if protection <= 0 && high <= 0 && stage <= 0 {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 3개**: B5, B6, B7
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
