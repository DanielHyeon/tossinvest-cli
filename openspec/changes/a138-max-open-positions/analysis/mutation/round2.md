# a138 변이 원장

- 사본: 연결 워크트리 `7e292659` (시작 sha 단언 · 깨끗함 확인 · 무변이 대조군 39 시험 GREEN)

| 변이 | 파일 | 시험 | 결과 |
|---|---|---|---|
| M1 초과(>) → 도달(>=) | `internal/journal/open_positions.go` | `TestA138TheSeventhSymbolIsRefused` | CAUGHT |
| M2 든 종목 scale-in 통과 삭제 | `internal/journal/open_positions.go` | `TestA138AnEntryOnAHeldSymbolIsAllowedAtTheMaximum` | CAUGHT |
| M3 ② HELD 예약 종목 버림 | `internal/journal/open_positions.go` | `TestA138TheSeventhSymbolIsRefused` | CAUGHT |
| M4 ① 외부 · 편입 보유 버림 | `internal/journal/open_positions.go` | `TestA138EveryKindOfLivePositionCounts` | CAUGHT |
| M5 노출 미증가 조기 통과 반전 | `internal/journal/open_positions.go` | `TestA138TheSeventhSymbolIsRefused` | CAUGHT |
| M6 상한 없는 예약 통과 | `internal/journal/open_positions.go` | `TestA138AnExposureRaisingReservationWithoutTheMaximumIsRefused` | CAUGHT |
| M7 preimage 결함을 0 으로 | `internal/journal/open_positions.go` | `TestA138AnUncountableHoldFailsClosed` | CAUGHT |
| M8 IssueEntry 문 상한 누락 | `internal/execgw/riskguardian.go` | `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` | CAUGHT |
| M9 인터록 7c 삭제 | `internal/app/engine/interlock.go` | `TestA138TheInterlockRefusesAGuardianThatDisagreesOnTheOpenPositionMaximum` | CAUGHT |
| M10 digest 입력 삭제 | `internal/execgw/riskguardian.go` | `TestA138TheLimitsDigestBindsTheOpenPositionMaximum` | CAUGHT |
| M11 ③ 이번 종목 버림 | `internal/journal/open_positions.go` | `TestA138TheSeventhSymbolIsRefused` | CAUGHT |
| M12 CLOSED 필터 삭제 | `internal/journal/open_positions.go` | `TestA138PositionsAndHoldsOfOneSymbolCountOnceAndClosedOnesNotAtAll` | CAUGHT |
| M13 시장 차원 무시 | `internal/journal/open_positions.go` | `TestA138PositionsAndHoldsOfOneSymbolCountOnceAndClosedOnesNotAtAll` | CAUGHT |
| M14 HELD 필터 삭제 | `internal/journal/open_positions.go` | `TestA138EveryHeldReservationKindCountsAndReleasedOnesDoNot` | CAUGHT |
| M15 종류를 OPEN_EXPOSURE 만 | `internal/journal/open_positions.go` | `TestA138EveryHeldReservationKindCountsAndReleasedOnesDoNot` | CAUGHT |
| M16 ② safety_class 필터 삭제 | `internal/journal/open_positions.go` | `TestA138AHeldReservationOfANonEntryDecisionTakesNoSlot` | CAUGHT |
| M17 ④ RECONCILE 칸 버림 | `internal/journal/open_positions.go` | `TestA138AReconciledSymbolWithoutAPositionRowCounts` | CAUGHT |
| M18 정수 판정 IsInf 삭제 | `internal/config/open_positions.go` | `TestA138WholeOpenPositionsIsTheOneIntegerJudgement` | SURVIVED |
| M19 정수 판정 소수 검사 삭제 | `internal/config/open_positions.go` | `TestA138WholeOpenPositionsIsTheOneIntegerJudgement` | CAUGHT |
| M20 factory 사전 거부 삭제 | `internal/app/engine/guardian_wiring.go` | `TestA138ProductionPolicyResolvesTheOpenPositionMaximum` | CAUGHT |
| M21 Validate 1 미만 → 0 미만 | `internal/risk/input.go` | `TestA138PolicyValidationRefusesAnUnusableOpenPositionMaximum` | CAUGHT |
| M22 콘솔 칸 없는 폼도 씀 | `internal/console/settings_limits.go` | `TestA138TheAdvancedFormSavesTheOpenPositionMaximum` | CAUGHT |
| M23 편집기 검증 삭제 | `internal/config/limits.go` | `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` | CAUGHT |
| M24 기동 audit 줄 삭제 | `internal/app/engine/interlock.go` | `TestA138TheInterlockAuditsTheOpenPositionMaximum` | CAUGHT |
| M25 기본값 0 | `internal/config/open_positions.go` | `TestA138OpenPositionMaximumResolution` | CAUGHT |
| M26 q_final 문 상한 누락 | `internal/execgw/riskguardian_qfinal.go` | `TestA138QFinalEntryRefusesTheEntryThatExceedsTheMaximum` | CAUGHT |
| M27 첫 다리 문 상한 누락 | `internal/execgw/riskguardian_first_leg.go` | `TestA138QFinalCampaignFirstLegRefusesTheEntryThatExceedsTheMaximum` | CAUGHT |
| M28 콘솔 행 삭제 | `internal/console/settings_limits.go` | `TestA138TheScreenSaysWhatTheOpenPositionMaximumIs` | CAUGHT |
| M29 콘솔 저장 audit 삭제 | `cmd/tossctl/limitsettings.go` | `TestA138TheOpenPositionMaximumSaveIsAuditedFromTheFile` | CAUGHT |
| M30 매도가 판정을 지남(IssueReduction 에 판정 삽입) | `internal/execgw/riskguardian.go` | `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` | CAUGHT |
| M31 ④ 해제 필터 삭제 | `internal/journal/open_positions.go` | `TestA138ReconcileScopesMatchExactlyWhatTheyCover` | CAUGHT |
| M32 ④ 이번 종목 동일 삭제(시장 범위) | `internal/journal/open_positions.go` | `TestA138ReconcileScopesMatchExactlyWhatTheyCover` | CAUGHT |
| M33 ④ 이미 센 종목 무시(시장 범위) | `internal/journal/open_positions.go` | `TestA138ReconcileScopesMatchExactlyWhatTheyCover` | CAUGHT |
| M34 ④ 이번 종목 동일 삭제(전역 범위) | `internal/journal/open_positions.go` | `TestA138ReconcileScopesMatchExactlyWhatTheyCover` | CAUGHT |
| M35 시장 정규화 삭제 | `internal/journal/open_positions.go` | `TestA138ReconcileScopesMatchExactlyWhatTheyCover` | CAUGHT |
| M36 시장 범위를 전역으로 | `internal/journal/open_positions.go` | `TestA138ReconcileScopesMatchExactlyWhatTheyCover` | CAUGHT |
| M37 진술 6 고정 | `internal/execgw/open_positions.go` | `TestA138TheLimitsDigestBindsTheOpenPositionMaximum` | CAUGHT |
| M38 새 파일 갈래 값 버림 | `internal/config/limits_io.go` | `TestA138TheLimitWriterTouchesTheOpenPositionKeyOnlyWhenAsked` | CAUGHT |
| M39 정수 판정 IsNaN 삭제 | `internal/config/open_positions.go` | `TestA138WholeOpenPositionsIsTheOneIntegerJudgement` | SURVIVED |

## 분류

- **37 / 39 CAUGHT**(CAUGHT(build) 0 — 전부 행동 실패). round1(86442ad7) 은 30 중 28 CAUGHT — 생존 M16 은 `TestA138AHeldReservationOfANonEntryDecisionTakesNoSlot` 로 못 박아 round2 에서 CAUGHT.
- **동등 변이 2**: M18(IsInf 삭제) · M39(IsNaN 삭제) — `WholeOpenPositions` 의 범위 검사(`v < 1` · `v > maxOpenPositionsCeiling`)가 ±Inf 를, 소수부 검사(`v != Trunc(v)`)가 NaN 을
  이미 거부(NaN 비교는 언제나 거짓이라 `v < 1` · `v > 상한` 은 통과시키지만 `v != Trunc(v)` 는 참). 이름 붙인 검사는 읽기용 — 주석 · design D4 에 같은 분류(구현 리뷰 P3-3).
