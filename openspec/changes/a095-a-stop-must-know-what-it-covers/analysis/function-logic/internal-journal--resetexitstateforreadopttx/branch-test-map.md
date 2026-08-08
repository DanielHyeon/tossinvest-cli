# Branch Test Map: `resetExitStateForReadoptTx`

- Source: `internal/journal/apply_hook.go`

> **「진입 실측」은 측정값이다** — 네 패키지의 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a095가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a095 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:687` `if err != nil {` | 아니오 | 기존 | no | no |
| B2 | `:692` `if err := tx.QueryRowContext(ctx, `SELECT adoption_id,instance_seq FROM positions WHERE id=?`,` | — | 기존 | no | no |
| B3 | `:700` `if err != nil {` | 아니오 | 기존 | no | no |
| B4 | `:715` `if err != nil {` | 아니오 | 기존 | no | no |
| B5 | `:719` `if err != nil {` | 아니오 | 기존 | no | no |
| B6 | `:722` `if affected != 1 {` | 예 | **a095 3.5** — 정확히 1행이라는 불변식은 새 경로에도 그대로 필요하다 | no | no |

**미진입 분기 4개**: B1, B3, B4, B5
**자체 블록 없는 분기 1개**: B2 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `switch {` 등)이며 미커버와 다르다.
