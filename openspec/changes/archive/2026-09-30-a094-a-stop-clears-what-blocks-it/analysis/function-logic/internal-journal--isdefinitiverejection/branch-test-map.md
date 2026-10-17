# Branch Test Map: `isDefinitiveRejection`

- Source: `internal/journal/dispatch.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:350` `switch statusCode {` | — | a094 2.6 — 목록 무변화를 표로 고정 | no | no |
| B2 | `:351` `case 400, 401, 403, 404, 405, 415, 422:` | 예 | a094 2.6 | no | no |
| B3 | `:353` `default:` | 예 | a094 2.4 — code 없는 409는 종전대로 Ambiguous | no | no |

**미진입 분기 0개**: 없음
**자체 블록 없는 분기 1개**: B1 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `default:` 등)이며 미커버와 다르다.
