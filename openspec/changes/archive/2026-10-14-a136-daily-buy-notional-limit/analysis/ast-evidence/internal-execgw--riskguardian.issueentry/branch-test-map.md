# Branch Test Map: `RiskGuardian.IssueEntry`

- Source: `internal/execgw/riskguardian.go` (408-555) — base 7b073f55, 편집 예정(D1·D2 결정 뒤 재확정)

| Branch | Anchor | Scenario | Test | a136 |
|---|---|---|---|---|
| B1 | 409 | 수집기 없음 | `TestRiskGuardianIssueEntryRejectsMissingCollectorBeforeIssuingAuthority` | 영향 없음 |
| B2 | 413 | 다른 계좌 | `TestTheGuardianRefusesAnIntentForAnotherAccount` | 영향 없음 |
| B3 | 416 | 매수 아님 | `TestAReductionIsIssuedWithNoLimitsAndNoReservation`(대응 문) | 근거(매도 비차단) |
| B4 | 430 | 체인 거부 | `TestAChainRefusalIssuesNothing` | 영향 없음 |
| B5 | 439 | 예약 금액 불가 | `TestAChainRefusalIssuesNothing` 계열 | 영향 없음 |
| B6 | 474 | 수집 오류 | `TestReCollectionDoesNotReRunTheChain` | 편집 예정 지점(클로저) |
| B7 | 478 | 노출 사용량 오류 | 구현 단계에서 확정 | 편집 예정 지점(클로저) |
| B8 | 509 | 일반 진입 | `TestTheGuardianIssuesTheDecisionAndItsReservationTogether` | 새 시험 추가 예정 |
| B9 | 511 | 전략 진입 | 전략 진입 시험 | 새 시험 추가 예정 |
| B10 | 515 | 전략 수집 오류 | B6 | 영향 없음 |
| B11 | 520 | 전략 성공 | 전략 진입 시험 | 영향 없음 |
| B12 | 526 | 원자 발급 거부 | `TestARefusedReservationLeavesNoSubmittableDecision` · `TestTheIssuanceReasonsAreDistinguishable` | 새 거부 사유 시험 추가 예정 |
| B13 | 533 | 발급 단계 관측 | `TestTheIssuanceReasonsAreDistinguishable` | 영향 없음 |
