# Branch Test Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:479` `if attempts <= 0 {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:485` `for attempt := 1; attempt <= attempts; attempt++ {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:486` `if n.Publisher == nil {` | 아니오 | **a095 2.5** — 알림 off에서 a095의 사실이 이 창에 오지 않는다(호출자 쪽에서 고정) | no | no |
| B4 | `:491` `if err == nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:493` `if markErr == nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:494` `switch settled.Outcome {` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:495` `case journal.SettleApplied:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B8 | `:497` `case journal.SettleLeaseLost, journal.SettleAlreadySettled:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B9 | `:509` `case journal.SettleNotFound:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B10 | `:511` `default:` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B11 | `:537` `if n.Log != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B12 | `:551` `if markErr != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B13 | `:555` `} else if failed.Outcome != journal.SettleApplied {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B14 | `:552` `if n.Log != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B15 | `:555` `} else if failed.Outcome != journal.SettleApplied {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B16 | `:565` `if n.Log != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B17 | `:575` `if !isPreemption(failed.Outcome) && n.Gate != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B18 | `:584` `if attempt < attempts {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B19 | `:585` `if !n.wait(ctx) {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B20 | `:603` `switch {` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B21 | `:604` `case relErr != nil:` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B22 | `:605` `if n.Log != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B23 | `:608` `case released.Outcome == journal.SettleApplied:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B24 | `:611` `case !isPreemption(released.Outcome):` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B25 | `:616` `if n.Gate != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B26 | `:623` `default:` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B27 | `:639` `if n.Log != nil {` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 6개**: B1, B3, B10, B21, B22, B26
**자체 블록 없는 분기 2개**: B6, B20 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
