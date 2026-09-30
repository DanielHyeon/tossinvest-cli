# Branch Test Map: `ClassifyHTTPMutation`

- Source: `internal/journal/dispatch.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:302` `if err != nil {` | 예 | a094 2.6 — 이 함수의 판정 무변화 | no | no |
| B2 | `:303` `if send == SendNotStarted {` | 예 | a094 2.6 — 이 함수의 판정 무변화 | no | no |
| B3 | `:320` `switch {` | 예 | a094 2.6 — 이 함수의 판정 무변화 | no | no |
| B4 | `:321` `case statusCode >= 200 && statusCode < 300:` | 예 | a094 2.6 — 이 함수의 판정 무변화 | no | no |
| B5 | `:323` `case isDefinitiveRejection(statusCode):` | 예 | a094 2.6 — 이 함수의 판정 무변화 | no | no |
| B6 | `:330` `case statusCode == 0:` | 예 | a094 2.6 — 이 함수의 판정 무변화 | no | no |
| B7 | `:337` `default:` | 예 | a094 2.6 — 이 함수의 판정 무변화 | no | no |

**미진입 분기 0개**: 없음
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `default:` 등)이며 미커버와 다르다.
