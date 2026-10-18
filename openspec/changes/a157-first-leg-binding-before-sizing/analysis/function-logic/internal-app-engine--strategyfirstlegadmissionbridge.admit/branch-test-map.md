# Branch Test Map: `admit`

- Source: `internal/app/engine/strategy_first_leg_admission.go` (86-161)

| Branch | Anchor | Scenario | Test |
|---|---|---|---|
| B1 | if 88 | 결과 무효 거절 그대로(권한 수집 전) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestStrategyFirstLegAdmissionRejectsInvalidBeforeAuthority` |
| B2 | if 91 | AUTHORITY_UNAVAILABLE | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestStrategyFirstLegAdmissionDormantPairedKRUS` |
| B3 | if 95 | AUTHORITY_COLLECTION_FAILED + 원인(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestStrategyFirstLegAdmissionClassifiesPairedFailures` |
| B4 | if 100 | AUTHORITY_MISMATCH(결속 불일치 — B3~B7 꼴을 precheck 앞에서 막음, design F8) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestStrategyFirstLegAdmissionRejectsAuthorityTamperPairedKRUS` |
| B5 | if 106 | precheck 거절 → 아래 분류(일일 손실 · 상한 0 · 버킷 · 그 밖 AUTHORITY_MISMATCH) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA137ClassificationIsUnchangedForEveryPrecheckRefusal` · `TestA157EngineABindingDefectAloneIsAnAuthorityMismatch` |
| B6 | if 110 | 편집 전 a137 측정 불가 갈래(표지) — a157 이 판별 함수로 옮김 | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault` · `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |
| B7 | if 111 | 편집 전 a137 측정 불가 갈래의 체인 · 입력 불가 조건 → DAILY_LOSS_UNMEASURABLE — a157 이 판별 함수로 옮김 | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA137AnUnmeasurableDayIsAPolicyOutcomeNotALaneFault` · `TestA157EngineAnUnmeasurableDayStillWinsOverABindingDefect` |
| B8 | if 118 | 편집 전 a137 한도 갈래(체인 · 한도 도달) → DAILY_LOSS_LIMIT — a157 이 판별 함수로 옮김 | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA137ADailyLossRefusalEscalatesAndIsNotALaneFault` · `TestA157EngineADailyLossLimitStillEscalatesWithABindingDefect` |
| B9 | if 127 | a146 GUARDIAN_CAP_ZERO(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` · `TestA157EngineACapZeroAloneIsStillASkip` |
| B10 | if 133 | 스킵 알림의 FX 표시(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA146ACapZeroPrecheckIsATypedSkip` |
| B11 | if 149 | 버킷 고갈 → 원인을 범위 거절로(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA146CharacterizationABucketExhaustedScopeStillLatchesAnEffectiveMarket` |
| B12 | if 150 | 범위 키가 있을 때만 범위 거절(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestA146CharacterizationABucketExhaustedScopeStillLatchesAnEffectiveMarket` |
| B13 | if 157 | ATOMIC_ADMISSION_FAILED(무편집) | `go test -tags tossos_testseams ./internal/app/engine -run 'A157|A137|A146|StrategyFirstLegAdmission'` · `TestStrategyFirstLegAdmissionClassifiesPairedFailures` |
