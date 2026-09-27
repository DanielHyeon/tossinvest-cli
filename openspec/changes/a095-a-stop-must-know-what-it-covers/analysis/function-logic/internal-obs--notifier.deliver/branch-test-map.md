# Branch Test Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:422` `if attempts <= 0 {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:428` `for attempt := 1; attempt <= attempts; attempt++ {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:429` `if n.Publisher == nil {` | 아니오 | **a095 2.5** — 알림 off에서 a095의 사실이 이 창에 오지 않는다(호출자 쪽에서 고정) | no | no |
| B4 | `:434` `if err == nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:436` `if markErr == nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:437` `switch settled.Outcome {` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:438` `case journal.SettleApplied:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B8 | `:440` `case journal.SettleLeaseLost, journal.SettleAlreadySettled:` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B9 | `:452` `case journal.SettleNotFound:` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B10 | `:454` `default:` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B11 | `:478` `if n.Log != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B12 | `:483` `if n.Gate != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B13 | `:495` `if markErr != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B14 | `:499` `} else if failed.Outcome != journal.SettleApplied {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B15 | `:496` `if n.Log != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B16 | `:499` `} else if failed.Outcome != journal.SettleApplied {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B17 | `:509` `if n.Log != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B18 | `:519` `if failed.Outcome == journal.SettleNotFound && n.Gate != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B19 | `:525` `if attempt < attempts {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B20 | `:526` `if !n.wait(ctx) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B21 | `:543` `switch {` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B22 | `:544` `case relErr != nil:` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B23 | `:545` `if n.Log != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B24 | `:548` `case released.Outcome == journal.SettleApplied:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B25 | `:551` `default:` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B26 | `:565` `if n.Log != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B27 | `:570` `if n.Gate != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 8개**: B1, B3, B8, B9, B10, B22, B23, B25
**자체 블록 없는 분기 2개**: B6, B21 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
