# Function Logic Map: `PrecheckQFinalCampaignFirstLeg`

- Source: `internal/execgw/riskguardian_first_leg.go` (49-112)
- Qualified: `RiskGuardian.PrecheckQFinalCampaignFirstLeg`
- AST evidence: `ast.json`(편집 전(= base 소스)) — 현재 AST branches 9 · returns 10
- 편집 전 기준: `ast-base.json`(base `7b1748481`, `source_sha256` a31ccdfd4d985cca…) — 편집 전 갈래 수 9
- Risk scan: `risk-pattern-report.md`

**역할.** 전략 첫 다리 precheck — 진입 판정(`PrecheckQFinalEntry`: 사이징 · 버킷 · 체인 · 노출)과 첫 다리 고유 결속 판정 일곱(B2 정책 버전 · B3 소유자 prospective 세대 · B4 가격 단위 · B5 계보 결속 · B6 포지션 세대 · B7 attempt/활성화 · B8 주간 예약)을 거쳐 계보 투영(B9) 뒤 봉인을 돌려줌. 원장 · 브로커 호출 없음.

**a157 편집.** design D1 — ① nil 가드(`PrecheckQFinalEntry` B1 과 같은 거절) ② B2~B8 을 편집 전 상대 순서로 **요청 값**에 대해 판정하고 첫 결함을 `bindErr` 로 잡음(입력 출처: `request.Entry.Admission.TransactionID` · `.Owner` · `request.Entry.QCandidate` — F3 · F4 로 같은 값) ③ `PrecheckQFinalEntry` 호출 한 번 ④ 진입 거절이 있고 결속 결함이 없거나 진입 거절이 일일 손실 체인 거절(`execgw.DailyLossChainRefusalOf`)이면 진입 거절 ⑤ 결속 결함 ⑥ B9 투영(무편집). 판정식 · 코드 · field · 문구 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD(설계 근거는 design F1~F10) | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 51) | — | 편집 전 B1: 진입 판정 거절 그대로(결속 판정은 돌지 않음 — a157 이 바꾸는 자리) | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157AnEntryRefusalAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` |
| B2 | `if err != nil {` (if 57) | — | B2 정책 버전 결속 오류(`journal.ErrInvalidRequest`) 그대로 | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` |
| B3 | `if strings.TrimSpace(owner.Key.ProspectiveGeneration) != "" {` (if 61) | — | B3 OWNER_CONFLICT/prospective_generation | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` |
| B4 | `if !entryOK \|\| !stopOK \|\| !targetOK {` (if 67) | — | B4 RISK_CALCULATION_INVALID/first_leg_price_unit | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` |
| B5 | `if request.Result.Quantity != entry.decision.QCandidate \|\| lineage.AccountRef != g.accountRef \|\|` (if 71) | — | B5 RISK_CALCULATION_INVALID/first_leg_lineage(q_candidate · 계좌 · 시장 · 종목 · 가격 · 소유자 · 캠페인) | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` · `TestA157OneInputThatTripsBothJudgements` |
| B6 | `if currentGeneration < 0 \|\| currentGeneration >= math.MaxInt64 \|\| request.Campaign.ExpectedPositionVersion < 0 \|\|` (if 80) | — | B6 OWNER_CONFLICT/position_generation | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` |
| B7 | `if strings.TrimSpace(request.AttemptID) == "" \|\| request.AttemptID != strings.TrimSpace(request.AttemptID) \|\| request.Revision < 1 \|\|` (if 85) | — | B7 RISK_CALCULATION_INVALID/first_leg_strategy | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` |
| B8 | `if err != nil {` (if 91) | — | B8 RISK_CALCULATION_INVALID/weekly_reservation(+주간 충돌 원인) | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestA157ABindingDefectAloneIsReportedAsItself` · `TestA157CapZeroAndABindingDefect` · `TestA157DailyLossIsReportedBeforeABindingDefect` |
| B9 | `if err != nil {` (if 105) | — | B9 RISK_CALCULATION_INVALID/first_leg_projection(무편집 — q_final 이 필요해 사이징 뒤) | `go test -tags tossos_testseams ./internal/execgw -run 'A157|QFinalCampaignFirstLeg|A136QFinalCampaignFirstLeg|A138QFinalCampaignFirstLeg'` · `TestQFinalCampaignFirstLegProjectsGuardianOwnedLineageAllEightLanes` |

Exact AST return positions: 52 · 58 · 62 · 68 · 76 · 82 · 87 · 92 · 106 · 111

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 16종 — 대표: `MajorDecimal`, `UTC`, `g.PrecheckQFinalEntry`, `g.clk.Now`, `guardianWeeklyFirstLegBinding`, `journal.ProjectAcceptedStrategyflowLineage`, `journal.QFinalPolicyVersion`, `qFinalRefusal`) | 본문 참조 | 소스 참조 — 원장 · 브로커 · 파일 호출은 design F2 · F9 | AST |

## State mutations and fallbacks

- 편집은 위 「a157 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a157 편집」에 적은 자리.
- High-risk impact: yes(Guardian 첫 다리 판정 순서 · 일일 손실 분류 · 감독자 조립 — 받아들이는 요청 집합 불변(design D2), 주문을 늘리는 갈래 없음,
  일일 손실 강화 보존(1-1), 매도 · 손절 · 청산 경로 무관)
