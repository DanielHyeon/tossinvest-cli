# Branch Test Map: `ReconcileDriver.adoptOne`

- Source: `internal/app/engine/adoption.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:320` `if err != nil {` | 아니오 | **a095 2.2** — 편입 전 거절(범주 ①)도 시도 실패 critical에 든다 | no | no |
| B2 | `:339` `if err != nil {` | 아니오 | **a095 2.2** — 영속 실패(범주 ②)는 시도 실패 critical | no | no |
| B3 | `:344` `if _, err := d.opts.Journal.OpenAdoptedExitState(ctx, c.position.ID); err != nil {` | 예 | **a095 2.13** — 커밋 뒤 보호 미개설(범주 ③)은 critical 요구 밖임을 명명된 경계로 고정 — 후속 후보 I7 | no | no |

**미진입 분기 2개**: B1, B2
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
