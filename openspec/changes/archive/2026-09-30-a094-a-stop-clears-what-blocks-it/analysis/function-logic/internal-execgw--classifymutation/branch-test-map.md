# Branch Test Map: `classifyMutation`

- Source: `internal/execgw/classify.go`

| Branch | 조건 | 진입 실측 | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `:22` `if err == nil {` | 예 | `TestTransportOutcomeTable` (accepted) — 무변화 | n/a (무변화) | yes |
| B2 | `:32` `if reason, refused := policyRefusal(err); refused {` | 예 | `TestGatewayRefusesFailClosedBranchesBeforeDispatch` — 무변화 | n/a (무변화) | yes |
| B3 | `:44` `switch reason, verdict := classifyRefusalCode(err); verdict {` | 예 | `TestA094RefusalCodeClassifiesTheAttempt` | yes | yes |
| B4 | `:45` `case refusalCodeDefinitive:` | 예 | `TestA094RefusalCodeClassifiesTheAttempt` (2.1 · 2.2 · 2.5b · 2.5c) | yes — base 에서 409 → IN_DOUBT, 422 → reason `dispatch_rejected_by_broker` | yes |
| B5 | `:52` `case refusalCodeContradictory:` | 예 | `TestA094RefusalCodeClassifiesTheAttempt` (2.5e · 2.5f) | yes — base 에서 422 모순 → FAILED_CONFIRMED | yes |
| B6 | `:66` `if reason, refused := ClassifyBrokerRefusal(err); refused {` | 예 | `TestA094ExistingRefusalCodesAreUnchanged` (2.7) | n/a (무변화) | yes |
| B7 | `:69` `if errors.As(err, &branch) && branch.Source == trading.BranchSourcePostPrepareConfirmation {` | 아니오 | 미진입 — a094 가 바꾸지 않는 post-prepare 승격(기존 공백) | n/a | n/a |
| B8 | `:84` `if status, known := statusOf(err); known {` | 예 | `TestA094RefusalCodeClassifiesTheAttempt` (2.3 · 2.4 · 2.5 · 2.5d · 2.5g — 판정 없음이 status 로 감) · `TestA094TheStatusTableIsUnchanged` (2.6) | n/a (무변화) | yes |
| B9 | `:87` `if outcome.Detail == "" {` | 아니오 | 미진입 — `ClassifyHTTPMutation` 은 status 분기에서 늘 detail 을 채운다(기존 공백) | n/a | n/a |
| B10 | `:89` `} else {` | 예 | `TestTransportOutcomeTable` — 무변화 | n/a (무변화) | yes |

RED 로그: `analysis/implementation/r1-red.log`(편집 전 트리에서 7 실패, 나머지 무변화 통과).
