# Function Logic Map: `admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go` (83-140)
- Qualified: `strategyFirstLegAdmissionBridge.admit`
- AST evidence: `ast.json`(현재) — 현재 AST branches 11 · returns 9
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 70815fb4533715dd…) — 편집 전 갈래 수 11
- Risk scan: `risk-pattern-report.md`

**역할.** 전략 첫 다리 승인 — 결과 검증 → 권한 수집 → 결속 검증 → 일일 손실 측정 → q_final precheck(거절 분류) → 원자 발급.

**a146 편집.** precheck 거절 갈래에서 a137 두 갈래 뒤 · `AUTHORITY_MISMATCH` 앞에 분류 하나: `QFinalRefusal` 코드 `EXISTING_GUARDIAN_CAP` 이고 원인 사슬에 `risk.ErrStrategyQuantityZero` 면 `GUARDIAN_CAP_ZERO` + 원인 `*strategyEntrySkip` + ctx 기록자 호출. 새 분기 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if refusal.Code != "" {` (if 85) | — | 결과 무효 거절 그대로(권한 수집 전) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B2 | `if b == nil \|\| b.loader == nil \|\| b.guardian == nil {` (if 88) | — | AUTHORITY_UNAVAILABLE | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B3 | `if err != nil {` (if 92) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B4 | `if err := validateStrategyFirstLegAuthority(accepted, authority); err != nil {` (if 97) | — | AUTHORITY_MISMATCH(결속 불일치 — 진입 · 손절가를 조건에 묶음, design F2) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B5 | `if err != nil {` (if 103) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B6 | `if unmeasurable := (*execgw.DailyLossUnmeasurable)(nil); errors.As(err, &unmeasurable) {` (if 107) | — | a137 측정 불가 갈래(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B7 | `if issue := (*execgw.IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == execgw.StageChain &&` (if 108) | a137 강화(한도 갈래만) | a137 DAILY_LOSS_UNMEASURABLE / DAILY_LOSS_LIMIT(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B8 | `if issue := (*execgw.IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == execgw.StageChain &&` (if 115) | a137 강화(한도 갈래만) | a137 DAILY_LOSS_UNMEASURABLE / DAILY_LOSS_LIMIT(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B9 | `if qFinal := (*execgw.QFinalRefusal)(nil); errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalBucketCapExhausted {` (if 128) | — | 버킷 고갈 → 원인을 범위 거절로(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B10 | `if key, keyed := strategyOwnerKeyOf(authority.Result.Lineage); keyed {` (if 129) | — | 범위 키가 있을 때만 범위 거절(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |
| B11 | `if err != nil {` (if 136) | — | 편집 전과 같음(소스 줄 참조) | `go test -tags tossos_testseams ./internal/app/engine -run 'A146|A137|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA146OtherGuardianCapRefusalsStayFaults` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA146TheRealPathEvaluatesEveryScopeAndPlacesNothing` |

Exact AST return positions: 86 · 89 · 95 · 98 · 112 · 119 · 133 · 137 · 139

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 12종 — 대표: `b.guardian.EscalateDailyLossLimit`, `b.guardian.IssuePrecheckedQFinalCampaignFirstLeg`, `b.guardian.MeasureDailyLoss`, `b.guardian.PrecheckQFinalCampaignFirstLeg`, `b.loader.collectStrategyFirstLegAuthority`, `err.Error`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a146 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a146 편집」에 적은 한 자리.
- High-risk impact: yes(전략 첫 다리 진입 분류 — 주문을 늘리지 않음: 건너뛴 진입은 브로커 0, 매도 · 손절 · 청산 경로 무관)
