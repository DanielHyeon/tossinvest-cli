# Branch Test Map: `strategyAccountAuthorityLoader.collectMarket`

- Source SHA-256: `674fb6c148cca915074491fed6e39397a2e655d67687a913362b7676e64abdc3`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2 리뷰 수리, 2026-10-01).** 리뷰 수리: 적재 결과를 switch(B6~B10)로 나눠 실패 원인을 범위 칸에 운반하고(B7), 실패 뒤 ctx 가 끝났으면 원인을 ctx 로 바꾼다(B8 — 생산 적재기는 ctx 종료를 자기 오류로 접으므로, 조건 ④). 준비 판정 자체(B9)는 80ae96a5 와 같다.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2-fix/pre-edit/internal-app-engine--strategyaccountauthorityloader.collectmarket/`. 변이 원장 `analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 162:2 — 항목 없음 · 활성화 없는 시장의 항목 하나 아님/무효 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 165:2 — loader 구성 불완전 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 169:2 — 시장이 US 면 계좌 시장 US | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | range at 173:2 — 항목(범위)마다 적재 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B5 | if at 178:3 — 범위 키 · 제안 유효할 때만 적재 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B6 | switch at 183:4 — **(새)** 적재 결과 분기 | `a112_owner_scope_trading_test.go` `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded` | no — 구조 | yes |
| B7 | case at 184:4 — **(새)** 적재 실패 → 원인 운반 | `a112_owner_scope_trading_test.go` `TestAScopeWithoutItsOwnAccountAuthorityIsRefusedAloneAndRecorded` · `TestAnAccountLoadCancelledByItsContextStopsTheCycle` | yes — 변이 Y10 CAUGHT(전부 결함이면 J3 시험 FAIL) | yes |
| B8 | if at 187:5 — **(새)** 실패 뒤 ctx 종료 → 원인 = ctx 오류(결함) | `a112_owner_scope_trading_test.go` `TestAnAccountLoadThatFailsUnderACancelledContextIsAFault` | yes — 변이 Y11 CAUGHT | yes |
| B9 | case at 190:4 — 적재 성공 · 시장 · 매니페스트 일치 → 그 범위 준비 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B10 | case at 192:4 — **(새)** 적재 성공인데 시장 · 매니페스트 불일치 → 결함 원인 | 진입 0 — 시험 적재기는 늘 일치(seam 으로 못 만듦) | no | 진입 0 |
