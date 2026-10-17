# Branch Test Map: `alertDeliverer.deliverOne`

- Source: `internal/app/engine/alertdelivery.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:272` `if err != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:278` `switch claim.Disposition {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:279` `case journal.ClaimAcquired:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B4 | `:283` `case journal.ClaimHeldElsewhere:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:292` `if d.reportHeld(alert.ID, claim.ExpiresAt) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:298` `default:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:305` `if claim.Stole {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B8 | `:316` `if d.Publisher == nil {` | 예 | **a095 2.5** — 알림 off에서 a095의 사실이 critical 행을 만들지 않으므로 이 창에 오지 않는다 | no | no |
| B9 | `:326` `} else {` | 예 | **a095 2.11** — 10판 Q8 답 (b): 실행자는 저장된 문구를 보냄 — 문구가 시점 사건이라 늦게 보내도 참 | no | no |
| B10 | `:333` `if perr != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B11 | `:337` `if perr != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 0개**: 없음
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
