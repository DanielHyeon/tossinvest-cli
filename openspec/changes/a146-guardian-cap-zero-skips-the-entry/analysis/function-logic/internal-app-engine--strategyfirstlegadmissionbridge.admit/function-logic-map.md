# Function Logic Map: `admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go` (86-157)
- Qualified: `strategyFirstLegAdmissionBridge.admit`
- AST evidence: `ast.json`(현재) — 현재 AST branches 12 · returns 10
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 70815fb4533715dd…) — 편집 전 갈래 수 11
- Risk scan: `risk-pattern-report.md`

**역할.** 전략 첫 다리 승인 — 결과 검증 → 권한 수집 → 결속 검증 → 일일 손실 측정 → q_final precheck(거절 분류) → 원자 발급.

**a146 편집.** precheck 거절 갈래에서 a137 두 갈래 뒤 · `AUTHORITY_MISMATCH` 앞에 분류 하나: `QFinalRefusal` 코드 `EXISTING_GUARDIAN_CAP` 이고 그 거절 **자신의** 원인 사슬에 `risk.ErrStrategyQuantityZero` 면 `GUARDIAN_CAP_ZERO` + 원인 `*strategyEntrySkip` + ctx 기록자 호출. 새 분기 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if refusal.Code != "" {` (if 88) | — | 결과 무효 거절 그대로(권한 수집 전) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B2 | `if b == nil \|\| b.loader == nil \|\| b.guardian == nil {` (if 91) | — | AUTHORITY_UNAVAILABLE | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B3 | `if err != nil {` (if 95) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B4 | `if err := validateStrategyFirstLegAuthority(accepted, authority); err != nil {` (if 100) | — | AUTHORITY_MISMATCH(결속 불일치 — 진입 · 손절가를 조건에 묶음, design F2) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B5 | `if err != nil {` (if 106) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B6 | `if unmeasurable := (*execgw.DailyLossUnmeasurable)(nil); errors.As(err, &unmeasurable) {` (if 110) | — | a137 측정 불가 갈래(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B7 | `if issue := (*execgw.IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == execgw.StageChain &&` (if 111) | a137 강화(한도 갈래만) | a137 DAILY_LOSS_UNMEASURABLE / DAILY_LOSS_LIMIT(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B8 | `if issue := (*execgw.IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == execgw.StageChain &&` (if 118) | a137 강화(한도 갈래만) | a137 DAILY_LOSS_UNMEASURABLE / DAILY_LOSS_LIMIT(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B9 | `if qFinal := (*execgw.QFinalRefusal)(nil); errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalExistingGuardianCap &&` (if 127) | ctx 기록자 호출(`recordStrategyEntrySkip`) | **a146** GUARDIAN_CAP_ZERO + 원인 `*strategyEntrySkip` — 코드 · sentinel 둘 다 맞을 때만, 나머지는 아래 AUTHORITY_MISMATCH | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B10 | `if qFinal := (*execgw.QFinalRefusal)(nil); errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalBucketCapExhausted {` (if 145) | — | 버킷 고갈 → 원인을 범위 거절로(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B11 | `if key, keyed := strategyOwnerKeyOf(authority.Result.Lineage); keyed {` (if 146) | — | 범위 키가 있을 때만 범위 거절(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B12 | `if err != nil {` (if 153) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |

Exact AST return positions: 89 · 92 · 98 · 101 · 115 · 122 · 136 · 150 · 154 · 156

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 17종 — 대표: `MajorDecimal`, `accepted.result.ExecutionTerms.EffectiveStop`, `accepted.result.ExecutionTerms.Entry`, `b.guardian.EscalateDailyLossLimit`, `b.guardian.IssuePrecheckedQFinalCampaignFirstLeg`, `b.guardian.MeasureDailyLoss`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a146 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a146 편집」에 적은 한 자리.
- High-risk impact: yes(전략 첫 다리 진입 분류 — 주문을 늘리지 않음: 건너뛴 진입은 브로커 0, 매도 · 손절 · 청산 경로 무관)
