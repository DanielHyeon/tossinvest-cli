# Function Logic Map: `admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go` (86-161)
- Qualified: `strategyFirstLegAdmissionBridge.admit`
- AST evidence: `ast.json`(편집 전(= base 소스)) — 현재 AST branches 13 · returns 10
- 편집 전 기준: `ast-base.json`(base `7b1748481`, `source_sha256` ce30183e9d936008…) — 편집 전 갈래 수 13
- Risk scan: `risk-pattern-report.md`

**역할.** 전략 첫 다리 승인 — 결과 검증 → 권한 수집 → 결속 검증 → 일일 손실 측정 → q_final precheck(거절 분류) → 원자 발급.

**a157 편집.** precheck 거절 갈래의 a137 두 조건(측정 불가: 표지 + 체인 · 입력 불가 / 한도: 체인 · 한도 도달)을 execgw 판별 함수 `DailyLossChainRefusalOf` 의 결과로 가름 — 조건을 글자 그대로 옮겼으므로 분기 결과(코드 · 원인 · 강화 호출)는 같음(design D1 「판별은 한 자리」).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD(설계 근거는 design F1~F10) | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if refusal.Code != "" {` (if 88) | — | 결과 무효 거절 그대로(권한 수집 전) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestStrategyFirstLegAdmissionRejectsInvalidBeforeAuthority` |
| B2 | `if b == nil \|\| b.loader == nil \|\| b.guardian == nil {` (if 91) | — | AUTHORITY_UNAVAILABLE | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestStrategyFirstLegAdmissionDormantPairedKRUS` |
| B3 | `if err != nil {` (if 95) | — | AUTHORITY_COLLECTION_FAILED + 원인(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestStrategyFirstLegAdmissionClassifiesPairedFailures` |
| B4 | `if err := validateStrategyFirstLegAuthority(accepted, authority); err != nil {` (if 100) | — | AUTHORITY_MISMATCH(결속 불일치 — B3~B7 꼴을 precheck 앞에서 막음, design F8) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestStrategyFirstLegAdmissionRejectsAuthorityTamperPairedKRUS` |
| B5 | `if err != nil {` (if 106) | — | precheck 거절 → 아래 분류(일일 손실 · 상한 0 · 버킷 · 그 밖 AUTHORITY_MISMATCH) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA157EngineABindingDefectAloneIsAnAuthorityMismatch` |
| B6 | `if unmeasurable := (*execgw.DailyLossUnmeasurable)(nil); errors.As(err, &unmeasurable) {` (if 110) | — | 편집 전 a137 측정 불가 갈래(표지) — a157 이 판별 함수로 옮김 | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault` · `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |
| B7 | `if issue := (*execgw.IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == execgw.StageChain &&` (if 111) | — | 편집 전 a137 측정 불가 갈래의 체인 · 입력 불가 조건 → DAILY_LOSS_UNMEASURABLE — a157 이 판별 함수로 옮김 | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault` · `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |
| B8 | `if issue := (*execgw.IssueRefusal)(nil); errors.As(err, &issue) && issue.Stage == execgw.StageChain &&` (if 118) | a137 강화(`EscalateDailyLossLimit`) | 편집 전 a137 한도 갈래(체인 · 한도 도달) → DAILY_LOSS_LIMIT — a157 이 판별 함수로 옮김 | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault` · `TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect` |
| B9 | `if qFinal := (*execgw.QFinalRefusal)(nil); errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalExistingGuardianCap &&` (if 127) | ctx 기록자 호출(`recordStrategyEntrySkip`) | a146 GUARDIAN_CAP_ZERO(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA157EngineACapZeroAloneIsStillASkip` |
| B10 | `if digest := authority.Entry.FXAuthority.Digest(); digest != "" {` (if 133) | — | 스킵 알림의 FX 표시(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` |
| B11 | `if qFinal := (*execgw.QFinalRefusal)(nil); errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalBucketCapExhausted {` (if 149) | — | 버킷 고갈 → 원인을 범위 거절로(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA146CharacterizationABucketExhaustedScopeStillLatchesAnEffectiveMarket` |
| B12 | `if key, keyed := strategyOwnerKeyOf(authority.Result.Lineage); keyed {` (if 150) | — | 범위 키가 있을 때만 범위 거절(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA146CharacterizationABucketExhaustedScopeStillLatchesAnEffectiveMarket` |
| B13 | `if err != nil {` (if 157) | 원장 원자 발급 | ATOMIC_ADMISSION_FAILED(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestStrategyFirstLegAdmissionClassifiesPairedFailures` |

Exact AST return positions: 89 · 92 · 98 · 101 · 115 · 122 · 140 · 154 · 158 · 160

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 20종 — 대표: `MajorDecimal`, `accepted.result.ExecutionTerms.EffectiveStop`, `accepted.result.ExecutionTerms.Entry`, `authority.Entry.FXAuthority.AccountCurrency`, `authority.Entry.FXAuthority.Digest`, `authority.Entry.FXAuthority.QuoteCurrency`, `b.guardian.EscalateDailyLossLimit`, `b.guardian.IssuePrecheckedQFinalCampaignFirstLeg`) | 본문 참조 | 소스 참조 — 원장 · 브로커 · 파일 호출은 design F2 · F9 | AST |

## State mutations and fallbacks

- 편집은 위 「a157 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a157 편집」에 적은 자리.
- High-risk impact: yes(Guardian 첫 다리 판정 순서 · 일일 손실 분류 · 감독자 조립 — 받아들이는 요청 집합 불변(design D2), 주문을 늘리는 갈래 없음,
  일일 손실 강화 보존(1-1), 매도 · 손절 · 청산 경로 무관)
