# Function Logic Map: `strategyFirstLegAdmissionBridge.admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go` (83-140, 편집 뒤)
- Qualified: `strategyFirstLegAdmissionBridge.admit`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 70815fb4533715dd…) · `ast-base.json`(편집 전, base `a000cf29`, `source_sha256` 254b4a6abb0d95fe…, 편집 전 분기수 8) — 현재 branches 11 · returns 9
- Risk scan: `risk-pattern-report.md`

**역할.** 전략 첫 다리를 precheck → 원자 발급으로 승인하고 실패를 코드 · 원인으로 분류.

**a137 편집.** precheck 직전 `authority.Entry.DailyLoss = b.guardian.MeasureDailyLoss(ctx)`. precheck 거절이 `*IssueRefusal` · StageChain · DAILY_LOSS_LIMIT_REACHED 면(B6 추가) `DAILY_LOSS_LIMIT` + cause = `EscalateDailyLossLimit(ctx, err)`(성공 = 계좌 차단 표지, 실패 = join 오류). 그 밖의 precheck 거절도 cause = err(F8). 분기 8 → 9. 구현 리뷰 P1-3(Manager 결정 ①): `*execgw.DailyLossUnmeasurable` 표지 + StageChain · INPUT_UNAVAILABLE 이면(B6 · B7) 강화 없이 `DAILY_LOSS_UNMEASURABLE` + cause = err — 레인 결함 아님. 분기 9 → 11.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| result | 완결 계보 | 전략 흐름 | B1 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if refusal.Code != "" {` (if 85) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B2 | `if b == nil \|\| b.loader == nil \|\| b.guardian == nil {` (if 88) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B3 | `if err != nil {` (if 92) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B4 | `if err := validateStrategyFirstLegAuthority(accepted, authority); err != nil {` (if 97) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B5 | `if err != nil {` (if 103) | — | 분기 본문 참조 | `TestA137APrecheckRefusalKeepsItsCause` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · 기존 `TestStrategyFirstLegAdmissionClassifiesPairedFailures` |
| B6 | `if unmeasurable := (*execgw.DailyLossUnmeasurable)(nil); errors.As(err, &unmeasurable) {` (if 107) | — | 분기 본문 참조 | `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault` |
| B7 | `if issue := (*execgw.IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == execgw.StageChain &&` (if 108) | — | 분기 본문 참조 | `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault` |
| B8 | `if issue := (*execgw.IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == execgw.StageChain &&` (if 115) | — | 분기 본문 참조 | `TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault` · `TestA137AFailedEscalationIsJoinedNotLost` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` |
| B9 | `if qFinal := (*execgw.QFinalRefusal)(nil); errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalBucketCapExhausted {` (if 128) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B10 | `if key, keyed := strategyOwnerKeyOf(authority.Result.Lineage); keyed {` (if 129) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B11 | `if err != nil {` (if 136) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |

Exact AST return positions: 86 · 89 · 95 · 98 · 112 · 119 · 133 · 137 · 139

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `b.guardian.MeasureDailyLoss` | 측정 토큰(a137) | 실패도 토큰에 봉인 | AST calls |
| `b.guardian.PrecheckQFinalCampaignFirstLeg` | precheck | 거절 → B5 | AST calls |
| `b.guardian.EscalateDailyLossLimit` | 강화(a137) | 실패 → join | AST calls |

## State mutations and fallbacks

- 이 함수는 원장 능력이 없음(금지 문자열 시험 유지) — 원장 읽기 · 강화는 Guardian 안.

## Safety conclusion

- Safe edit boundary: precheck 앞 토큰 대입 한 줄 · B5 안의 분류.
- High-risk impact: yes — 진입을 더 막는 방향만.
