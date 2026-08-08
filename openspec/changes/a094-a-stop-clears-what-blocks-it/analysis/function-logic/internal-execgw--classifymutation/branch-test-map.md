# Branch Test Map: `classifyMutation`

- Source: `internal/execgw/classify.go`

> **「진입 실측」은 측정값이다** — 패키지 시험 전체를 `-covermode=set`으로 돌린 프로파일에서 그 분기가 만든 블록의 count다. 어떤 **개별** 시험이 그 분기를 밟는지는 이 실행이 답하지 않는다(시험별 프로파일이 필요하다). 따라서 「Test」 열은 **a094가 요구하는 시험**이며 현존 증명이 아니다.

| Branch | 조건 | 진입 실측 | Test (a094 요구) | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:22` `if err == nil {` | 예 | a094 2.7 — 성공 경로 무변화 | no | no |
| B2 | `:32` `if reason, refused := policyRefusal(err); refused {` | 예 | a094 2.7 | no | no |
| B3 | `:46` `if reason, refused := ClassifyBrokerRefusal(err); refused {` | 예 | **a094 2.1·2.2·2.5** — 새 code가 여기서 잡힌다 | no | no |
| B4 | `:49` `if errors.As(err, &branch) && branch.Source == trading.BranchSourcePostPrepareConfirmation {` | 아니오 | a094 2.7 — post-prepare 승격 무변화 | no | no |
| B5 | `:64` `if status, known := statusOf(err); known {` | 예 | **a094 2.3·2.4** — code가 없으면 여기로 와야 한다 | no | no |
| B6 | `:67` `if outcome.Detail == "" {` | 아니오 | a094 2.7 | no | no |
| B7 | `:69` `} else {` | 예 | a094 2.7 | no | no |

**미진입 분기 2개**: B4, B6
**자체 블록 없는 분기 0개**: 없음 — 컴파일러가 별도 블록을 만들지 않는 형태(빈 `default:` 등)이며 미커버와 다르다.
