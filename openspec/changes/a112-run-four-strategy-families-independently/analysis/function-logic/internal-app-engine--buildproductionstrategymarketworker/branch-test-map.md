# Branch Test Map: `buildProductionStrategyMarketWorker`

- Source SHA-256: `9cb510c1c9c7f44ac109de5fd8ddac8e6c559adb4d654b1acd65dce19cc9e577`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2 리뷰 수리, 2026-10-01).** 리뷰 수리(A #3): 승격 근거 범위가 자기 위험 · 계좌 권한을 갖도록 B5(키) · B6(위험 forScope) · B7(계좌 forScope)를 더했고, 만료를 준비된 계좌 범위의 최소값(`earliestFreshUntil`)으로 바꿨다(B11 조건의 만료 항). 나머지는 80ae96a5 와 같은 분기(번호 이동).
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2-fix/pre-edit/internal-app-engine--buildproductionstrategymarketworker/`. 변이 원장 `analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 434:2 — 배선 미완/nil → dormant | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 447:2 — 시장 권한 준비 미완 → dormant | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | range at 456:2 — 주문 경로와 같은 handoff 목록 순회 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | if at 458:3 — handoff 거절 · 봉인 깨짐 → 그 범위 건너뜀 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B5 | if at 464:3 — **(새)** 범위 키 정규화 실패 → 건너뜀 | 진입 0 — 조립 결과는 늘 유효 키(seam 으로 못 만듦) | no | 진입 0 |
| B6 | if at 467:3 — **(새)** 그 범위의 위험 권한 준비 안 됨 → 건너뜀(승격 근거 아님) | `a112_owner_scope_trading_test.go` `TestAWorkerPromotesOnlyOnAScopeWithBothAuthorities/000660_without_risk_authority` · `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope` | yes — `analysis/measurements/lot-5.2.2.2-fix/red-fix.log`(편집 전 권한 없는 범위로 승격) · 변이 Y15 CAUGHT(`analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`) | yes |
| B7 | if at 470:3 — **(새)** 그 범위의 계좌 권한 준비 안 됨 → 건너뜀 | `a112_owner_scope_trading_test.go` `TestAWorkerPromotesOnlyOnAScopeWithBothAuthorities/000660_without_account_authority` | yes — 변이 Y16 CAUGHT | yes |
| B8 | if at 476:3 — 보호 관측 실패 → 그 범위 건너뜀 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B9 | if at 479:3 — 진입 관문 관측 실패 → 그 범위 건너뜀(J3) | `a112_owner_scope_trading_test.go` `TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope` | 80ae96a5 변이 X15 · X17 | yes |
| B10 | if at 485:2 — 승격 근거 범위 없음 → dormant | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B11 | if at 508:2 — digest / revision / **만료(준비된 계좌 범위의 최소 FreshUntil)** → dormant | `a112_owner_scope_trading_test.go` `TestAWorkerOverTwoReadyScopesExpiresWithItsEarliestAccountScope` | yes — `analysis/measurements/lot-5.2.2.2-fix/red-fix.log` · 변이 Y17 CAUGHT | yes |
