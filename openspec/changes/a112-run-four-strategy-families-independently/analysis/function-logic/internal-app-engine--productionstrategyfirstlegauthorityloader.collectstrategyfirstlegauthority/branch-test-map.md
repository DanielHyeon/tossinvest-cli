# Branch Test Map: `collectStrategyFirstLegAuthority`

- Source SHA-256: `c29e90e2a1e9f04e531a1cc000caa4bcc6a97a0cd796a73845a97dd23e1de7ca`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2 리뷰 수리, 2026-10-01).** 리뷰 수리: B5(**활성화 없는 시장의 개수 관문** — codex #1, 6.2 위치), B6(키 정규화 실패 → 결함, A #4), B8(위험 범위 권한 부재가 범위 국소 원인일 때만 범위 거절, 아니면 타입 없는 결함 — A #1 · codex #2)를 더했고, B9(계좌 범위 권한 부재)는 **언제나 결함**(codex 재확인 P1 → 판정 (A) — 계좌 매니페스트는 시장 단위). 나머지는 80ae96a5 와 같은 분기(번호 이동). 2차 편집 전 번들: `lot-5.2.2.2-fix2/pre-edit/`.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2-fix/pre-edit/internal-app-engine--productionstrategyfirstlegauthorityloader.collectstrategyfirstlegauthority/`. 변이 원장 `analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 289:2 — loader · ctx · 시계 · 원장 · Guardian 부재 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 295:2 — 시장 권한 준비 미완 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 302:2 — 소유자 범위 선택 실패 — 위조 의심(타입 없음) | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | if at 306:2 — 봉인된 identity 대조 — 불일치는 타입 없는 오류 | `a112_first_leg_owner_scope_seal_test.go` `TestTheFirstLegSealRefusesEveryForgeryAxis` · `a112_owner_scope_trading_test.go` `TestAForgedScopeStopsTheCycleBeforeTheNextValidScope` | 80ae96a5 변이 X07 | yes |
| B5 | if at 311:2 — **(새)** 서명 활성화 **없는** 시장에서 항목이 정확히 하나가 아님 → `paired production authority is incomplete for market`(편집 전 문구) | `a112_first_leg_owner_scope_seal_test.go` `TestAnUnactivatedMultiEntryPairIsStillRefusedAtTheFirstLeg` | yes — `analysis/measurements/lot-5.2.2.2-fix/red-fix.log`(편집 전 err=nil) · 변이 Y12(제거) · Y13(활성화 시장에도) CAUGHT(`analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`) | yes |
| B6 | if at 318:2 — **(새)** 범위 키 정규화 실패 → 결함(타입 없음) | 진입 0 — 봉인 선택이 정규화를 보장(도달 불가) | M14 SURVIVED — 예상(도달 불가) | 진입 0 |
| B7 | if at 324:2 — 그 범위의 준비된 위험 권한 없음 | `a112_owner_scope_trading_test.go` `TestARiskScopeOutsideTheSignedPolicyIsRefusedAloneInEitherOrder` · `TestACorruptLedgerRowStopsTheCycleWithItsCause` | yes — 아래 B8 | yes |
| B8 | if at 326:3 — **(새)** 원인이 범위 국소가 아님(원장 결함 · 무결성 · 항목 부재) → 타입 없는 결함 `production risk authority fault …: %w`(주기 멈춤 · 원인 보존); 범위 국소(`riskbucket.ErrProductionRiskScopeRefused`)면 범위 거절 타입(원인 Unwrap) | `a112_owner_scope_trading_test.go` `TestACorruptLedgerRowStopsTheCycleWithItsCause` · `TestAScopeLatchIsRefusedAloneButALatchReadFaultStops` · `TestARiskScopeOutsideTheSignedPolicyIsRefusedAloneInEitherOrder` · `a112_owner_scope_handoff_test.go` `TestAScopeTheRiskAuthorityDoesNotHoldIsAFaultInEitherOrder` | yes — `analysis/measurements/lot-5.2.2.2-fix/red-fix.log`(편집 전 손상 행 → placed=[000660] · 범위 거절) · 변이 Y06 · Y07 · Y08 CAUGHT | yes |
| B9 | if at 332:2 — 그 범위의 준비된 계좌 권한 없음 → **언제나 결함**(타입 없는 오류 `production account authority fault …: %w` — 주기 멈춤 · 원인 보존). 계좌 매니페스트는 시장 단위 파일이라 범위 국소 원인이 없다(codex 재확인 P1 → Manager 판정 (A) — 앞 판의 「ctx 만 결함」 대체) | `a112_owner_scope_trading_test.go` `TestAnAccountLoadFailureOnOneScopeIsAFaultThatStopsTheCycle`(두 순서) · `TestAnAccountLoadCancelledByItsContextStopsTheCycle` | yes — `analysis/measurements/lot-5.2.2.2-fix2/red-fix2.log`(편집 전 범위 거절 · placed=000660) · 변이 Z01(범위 국소로 되돌림) · Z02(%v) CAUGHT | yes |
| B10 | if at 339:2 — 계보 시장 통화 미지 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B11 | if at 343:2 — 위험 권한 범위 불일치(범위 번들) | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B12 | if at 348:2 — 포지션 캠페인 CAS 변경 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B13 | range at 355:2 — 위험 버킷 항목 순회 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B14 | if at 366:2 — 가격 단위 무효 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B15 | if at 375:3 — 노출 스냅숏 만료(collect 클로저) | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B16 | if at 379:3 — 예약 버전 읽기 실패(collect 클로저) | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
