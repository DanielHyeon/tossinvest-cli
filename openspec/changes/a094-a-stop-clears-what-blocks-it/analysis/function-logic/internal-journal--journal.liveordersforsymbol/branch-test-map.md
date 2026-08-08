# Branch Test Map: `Journal.LiveOrdersForSymbol`

- Source: `internal/journal/fills.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:1851` `if err := j.guardTrackedFillIdentity(ctx, accountRef); err != nil {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B2 | `:1894` `if err != nil {` | 아니오 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B3 | `:1900` `for rows.Next() {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B4 | `:1902` `if err := rows.Scan(&o.OrderID, &o.IntentID, &o.AccountRef, &o.Market, &o.TradingDay,` | — | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B5 | `:1909` `if err := rows.Err(); err != nil {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B6 | `:1913` `for i := range out {` | 예 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |
| B7 | `:1921` `if err != nil {` | 아니오 | 기존 — a094는 이 함수를 바꾸지 않는다 | no | no |

**미진입 분기 2개**: B2, B7
**자체 블록 없는 분기 1개**: B4 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `default:` 등)이며 미커버와 다르다.
