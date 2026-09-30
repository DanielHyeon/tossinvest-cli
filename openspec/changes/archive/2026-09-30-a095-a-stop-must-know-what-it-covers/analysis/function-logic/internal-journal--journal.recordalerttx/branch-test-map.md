# Branch Test Map: `Journal.recordAlertTx`

- Source: `internal/journal/outbox.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다. 따라서 「Test」 열은 **a095 3판이 요구하는 시험**이며 현존 증명이 아니다. `[비움 — Qn]`은 사용자 결정이 덮지 않아 Manager 에게 올린 질문이다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:280` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:292` `switch {` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:293` `case err == nil:` | 예 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B4 | `:295` `if rearm {` | 예 | **a095 2.11** — [비움 — Q8] 편입 성공 뒤 남는 PENDING 행의 처분 | no | no |
| B5 | `:334` `if _, uerr := tx.ExecContext(ctx,` | — | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:347` `case !errors.Is(err, sql.ErrNoRows):` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:355` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |
| B8 | `:359` `if err != nil {` | 아니오 | 기존 — a095는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 4개**: B1, B6, B7, B8
**자체 블록 없는 분기 2개**: B2, B5 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
