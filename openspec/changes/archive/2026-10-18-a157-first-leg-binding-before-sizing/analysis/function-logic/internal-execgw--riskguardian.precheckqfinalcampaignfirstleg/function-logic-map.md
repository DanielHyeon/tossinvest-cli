# Function Logic Map: `PrecheckQFinalCampaignFirstLeg`

- Source: `internal/execgw/riskguardian_first_leg.go` (55-138)
- Qualified: `RiskGuardian.PrecheckQFinalCampaignFirstLeg`
- AST evidence: `ast.json`(편집 뒤) — 현재 AST branches 11 · returns 13
- 편집 전 기준: `ast-base.json`(base `7b1748481`, `source_sha256` a31ccdfd4d985cca…) — 편집 전 갈래 수 9
- Risk scan: `risk-pattern-report.md`

**역할.** 전략 첫 다리 precheck — 진입 판정(`PrecheckQFinalEntry`: 사이징 · 버킷 · 체인 · 노출)과 첫 다리 고유 결속 판정 일곱(B2 정책 버전 · B3 소유자 prospective 세대 · B4 가격 단위 · B5 계보 결속 · B6 포지션 세대 · B7 attempt/활성화 · B8 주간 예약)을 거쳐 계보 투영(B9) 뒤 봉인을 돌려줌. 원장 · 브로커 호출 없음.

**a157 편집.** design D1 — ① nil 가드(`PrecheckQFinalEntry` B1 과 같은 거절) ② B2~B8 을 편집 전 상대 순서로 **요청 값**에 대해 판정하고 첫 결함을 `bindErr` 로 잡음(입력 출처: `request.Entry.Admission.TransactionID` · `.Owner` · `request.Entry.QCandidate` — F3 · F4 로 같은 값) ③ `PrecheckQFinalEntry` 호출 한 번 ④ 진입 거절이 있고 결속 결함이 없거나 진입 거절이 일일 손실 체인 거절(`execgw.DailyLossRefusalKindOf`)이면 진입 거절 ⑤ 결속 결함 ⑥ B9 투영(무편집). 판정식 · 코드 · field · 문구 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD(설계 근거는 design F1~F10) | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if g == nil {` (if 57) | — | RISK_CALCULATION_INVALID/guardian(**a157 nil 가드** — 결속 판정이 `g.policyVersion` · `g.accountRef` 를 읽기 전) | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ANilGuardianRefusesBeforeAnyBindingJudgement` |
| B2 | `if err != nil {` (if 69) | — | B2 정책 버전 결속 오류(`journal.ErrInvalidRequest`) 그대로 | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` |
| B3 | `if strings.TrimSpace(owner.Key.ProspectiveGeneration) != "" {` (if 74) | — | B3 OWNER_CONFLICT/prospective_generation | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` |
| B4 | `if !entryOK \|\| !stopOK \|\| !targetOK {` (if 80) | — | B4 RISK_CALCULATION_INVALID/first_leg_price_unit | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` |
| B5 | `if request.Result.Quantity != request.Entry.QCandidate \|\| lineage.AccountRef != g.accountRef \|\|` (if 84) | — | B5 RISK_CALCULATION_INVALID/first_leg_lineage(q_candidate · 계좌 · 시장 · 종목 · 가격 · 소유자 · 캠페인) | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` · `TestA157OneInputThatTripsBothJudgements` |
| B6 | `if currentGeneration < 0 \|\| currentGeneration >= math.MaxInt64 \|\| request.Campaign.ExpectedPositionVersion < 0 \|\|` (if 93) | — | B6 OWNER_CONFLICT/position_generation | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` |
| B7 | `if strings.TrimSpace(request.AttemptID) == "" \|\| request.AttemptID != strings.TrimSpace(request.AttemptID) \|\| request.Revision < 1 \|\|` (if 98) | — | B7 RISK_CALCULATION_INVALID/first_leg_strategy | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` |
| B8 | `if err != nil {` (if 104) | — | B8 RISK_CALCULATION_INVALID/weekly_reservation(+주간 충돌 원인) | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` |
| B9 | `if err != nil && (bindErr == nil \|\| DailyLossRefusalKindOf(err) != DailyLossRefusalKindNone) {` (if 114) | — | **a157 ④** 진입 거절 반환 — 결속 결함이 없거나 진입 거절이 일일 손실 체인 거절(1-1) | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157DailyLossIsReportedBeforeABindingDefect` · `TestA157AnEntryRefusalAloneIsReportedAsItself` · `TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect` · `TestA157EngineAFailedEscalationIsTheSameJoinedFaultWithABindingDefect` · `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |
| B10 | `if bindErr != nil {` (if 118) | — | **a157 ⑤** 결속 결함 반환(상한 0 · 버킷 · 그 밖의 진입 거절보다 먼저) | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157CapZeroAndABindingDefect` · `TestA157OtherEntryRefusalsAndABindingDefect` · `TestA157OneInputThatTripsBothJudgements` · `TestA157EngineACapZeroNoLongerHidesABindingDefect` |
| B11 | `if err != nil {` (if 131) | — | B9 RISK_CALCULATION_INVALID/first_leg_projection(무편집 — q_final 이 필요해 사이징 뒤) | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes` |

Exact AST return positions: 58 · 70 · 75 · 81 · 89 · 95 · 100 · 105 · 109 · 115 · 119 · 132 · 137

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 17종 — 대표: `DailyLossRefusalKindOf`, `MajorDecimal`, `UTC`, `g.PrecheckQFinalEntry`, `g.clk.Now`, `guardianWeeklyFirstLegBinding`, `journal.ProjectAcceptedStrategyflowLineage`, `journal.QFinalPolicyVersion`) | 본문 참조 | 소스 참조 — 원장 · 브로커 · 파일 호출은 design F2 · F9 | AST |

## State mutations and fallbacks

- 편집은 위 「a157 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a157 편집」에 적은 자리.
- High-risk impact: yes(Guardian 첫 다리 판정 순서 · 일일 손실 분류 · 감독자 조립 — 받아들이는 요청 집합 불변(design D2), 주문을 늘리는 갈래 없음,
  일일 손실 강화 보존(1-1), 매도 · 손절 · 청산 경로 무관)
