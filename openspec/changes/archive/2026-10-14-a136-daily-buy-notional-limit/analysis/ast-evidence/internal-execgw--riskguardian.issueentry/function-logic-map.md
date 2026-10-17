# Function Logic Map: `RiskGuardian.IssueEntry`

- Source: `internal/execgw/riskguardian.go` (408-555, base 7b073f55)
- Qualified: `RiskGuardian.IssueEntry`
- 위치: `analysis/ast-evidence/` — **proposal 단계 근거(게이트 번들 아님)**. 편집하게 되는 함수는 구현 착수 때 `analysis/function-logic/` 에 정식 번들을 만든다.
- AST evidence: `ast.json` (`source_sha256` 88cf7c183018aae8…) — AST branches 13 · returns 12
- Risk scan: `risk-pattern-report.md`
- 용도: a136 proposal 의 주장 「진입 발급은 체인 1회 뒤 **OPEN_EXPOSURE 한 종류만** 원자 예약한다 — 일일 매수 합계를 세는 칸은 없다」의 근거. a136 구현은 이 함수의 `collectIssue` 클로저(474-505)가 만드는 `IssueRequest` 에 일일 매수 판정 입력을 더하게 될 것이므로, **구현 착수 전에 이 번들을 Pre-Edit 기준으로 다시 확정**한다(D1·D2 결정 뒤).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `req.Collect` | nil 아님 | 호출자 | 오류(409) |
| `req.Intent.Side` | `BUY` | 호출자 | 오류(416) — 매수만 이 문으로 들어옴 |
| `req.Account` | 체인 입력 | 호출자 수집 | 체인 거부 |
| `g.policy` | 생성 시 검증된 정책 | `riskPolicyFromAutomationGate` | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `req.Collect == nil` (409) | 없음 | 오류 | `TestRiskGuardianIssueEntryRejectsMissingCollectorBeforeIssuingAuthority` |
| B2 | `scopedIntent` 실패 (413) | 없음 | 오류 | `TestTheGuardianRefusesAnIntentForAnotherAccount` |
| B3 | `intent.Side != SideBuy` (416) | 없음 | 오류 — **매도는 IssueEntry 로 못 들어옴** | `TestAReductionIsIssuedWithNoLimitsAndNoReservation`(대응 문) |
| B4 | 체인 거부 (430) | 관측 1행 · 일손실이면 모드 강화 | `chainRefusal` (+ escalate) | `TestAChainRefusalIssuesNothing` |
| B5 | `EntryExposureValue` 거부 (439) | 관측 1행 | `chainRefusal` | `TestAChainRefusalIssuesNothing` 계열 |
| B6 | collectIssue: `req.Collect` 오류 (474) | 없음 | 오류(재수집 루프로) | `TestReCollectionDoesNotReRunTheChain` |
| B7 | collectIssue: `exposureUsage` 오류 (478) | 없음 | 오류 | `TestNegative…`(execgw) 계열 확인 필요 |
| B8 | `req.strategyPlan == nil` (509) | 원자 발급 `RecordDecisionAndReserveWithRecollection` | — | `TestTheGuardianIssuesTheDecisionAndItsReservationTogether` |
| B9 | else (511) 전략 계획 | `RecordStrategyDecisionAndReserveWithRecollection` | — | 전략 진입 시험(strategy_gateway_*) |
| B10 | collectIssue 오류 (515) | 없음 | 오류 | B6 와 같음 |
| B11 | `strategyErr == nil` (520) | out·receipt 채움 | — | 전략 진입 시험 |
| B12 | 원자 발급 오류 (526) | 관측(발급 거부) | `issuanceRefusal` | `TestARefusedReservationLeavesNoSubmittableDecision` · `TestTheIssuanceReasonsAreDistinguishable` |
| B13 | 발급 단계 거부일 때만 관측 (533) | 관측 1행 | — | `TestTheIssuanceReasonsAreDistinguishable` |

Exact AST return positions: 410 · 414 · 417 · 436 · 441 · 475 · 479 · 481 · 516 · 518 · 537 · 548

핵심(collectIssue 481-505): `SnapshotUsage` · `Limits` · `Reservations` 가 **각 한 항목, 종류 OPEN_EXPOSURE** 뿐이다. 일일 손실(DAILY_LOSS)조차 이 문에서는 예약되지 않는다 — 일일 손실은 체인 rung(`checkDailyLoss`)의 입력으로만 판정된다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `evaluateChain` (= `risk.Evaluate`) | 체인 1회 | 반환값 | AST calls · `TestReCollectionDoesNotReRunTheChain` |
| `risk.EntryExposureValue` | 예약 금액(비용 포함, account-base) | 반환값 | AST calls |
| `g.journal.RecordDecisionAndReserveWithRecollection` | 결정 + OPEN_EXPOSURE 예약 한 tx | 재수집 3회·10초, 한도 거부는 즉시 | journal/issuance.go |
| `g.journal.RecordStrategyDecisionAndReserveWithRecollection` | 전략 계획 포함 | 같음 | AST calls |
| `g.observeEntry` | 관측(비차단) | 실패가 발급을 되돌리지 않음 | spec |

## State mutations and fallbacks

- 모든 쓰기는 journal 원자 tx 안(결정 + 예약). 거부 시 아무것도 남지 않음. fallback 없음.

## Safety conclusion

- Safe edit boundary(예정): `collectIssue` 가 만드는 요청에 일일 매수 판정 입력을 더하고, journal 원자 tx 안에서 판정. 체인 · 관측 · 재수집 의미는 무변경. 매도(IssueReduction)는 이 함수를 지나지 않으므로(B3) 새 판정에 구조적으로 닿지 않음.
- High-risk impact: yes — 편집 시 Pre-Edit 선언 · RED 선행 · 변이 원장 · 적대 리뷰 의무.
- 같은 판정이 필요한 다른 발급 문: `IssuePrecheckedQFinalEntry`(riskguardian_qfinal.go:177) · `IssuePrecheckedQFinalCampaignFirstLeg`(riskguardian_first_leg.go:117, 생산 호출자 `strategy_first_leg_admission.go:105`). 구현 단계에서 각각 번들을 만든다.
