# Function Logic Map: `strategyFirstLegAdmissionBridge.admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go` (80-128, 편집 뒤)
- Qualified: `strategyFirstLegAdmissionBridge.admit`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` fd50f34c7a6915cf…) · `ast-base.json`(편집 전, base `a000cf29`, `source_sha256` 254b4a6abb0d95fe…, 편집 전 분기수 8) — 현재 branches 9 · returns 8
- Risk scan: `risk-pattern-report.md`

**역할.** 전략 첫 다리를 precheck → 원자 발급으로 승인하고 실패를 코드 · 원인으로 분류.

**a137 편집.** precheck 직전 `authority.Entry.DailyLoss = b.guardian.MeasureDailyLoss(ctx)`. precheck 거절이 `*IssueRefusal` · StageChain · DAILY_LOSS_LIMIT_REACHED 면(B6 추가) `DAILY_LOSS_LIMIT` + cause = `EscalateDailyLossLimit(ctx, err)`(성공 = 계좌 차단 표지, 실패 = join 오류). 그 밖의 precheck 거절도 cause = err(F8). 분기 8 → 9.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| result | 완결 계보 | 전략 흐름 | B1 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if refusal.Code != "" {` (if 82) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B2 | `if b == nil \|\| b.loader == nil \|\| b.guardian == nil {` (if 85) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B3 | `if err != nil {` (if 89) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B4 | `if err := validateStrategyFirstLegAuthority(accepted, authority); err != nil {` (if 94) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B5 | `if err != nil {` (if 100) | — | 분기 본문 참조 | `TestA137APrecheckRefusalKeepsItsCause` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · 기존 `TestStrategyFirstLegAdmissionClassifiesPairedFailures` |
| B6 | `if issue := (*execgw.IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == execgw.StageChain &&` (if 103) | — | 분기 본문 참조 | `TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault` · `TestA137AFailedEscalationIsJoinedNotLost` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` |
| B7 | `if qFinal := (*execgw.QFinalRefusal)(nil); errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalBucketCapExhausted {` (if 116) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B8 | `if key, keyed := strategyOwnerKeyOf(authority.Result.Lineage); keyed {` (if 117) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |
| B9 | `if err != nil {` (if 124) | — | 분기 본문 참조 | `go test -tags tossos_testseams ./internal/app/engine -run StrategyFirstLegAdmission` |

Exact AST return positions: 83 · 86 · 92 · 95 · 107 · 121 · 125 · 127

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
