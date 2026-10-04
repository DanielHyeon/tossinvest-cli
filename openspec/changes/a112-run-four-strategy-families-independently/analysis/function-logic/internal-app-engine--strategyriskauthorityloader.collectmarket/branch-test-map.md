# Branch Test Map: `strategyRiskAuthorityLoader.collectMarket`

- Source SHA-256: `90192bbebb241d780d2ab843a8c0536eda30596aeb976adadfbdd45623274746`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2 리뷰 수리, 2026-10-01).** 리뷰 수리: 적재 결과를 switch(B6~B9)로 나눠 실패 원인을 범위 칸에 **그대로** 운반한다(B7 — 편집 전에는 어떤 err 든 「준비 안 됨」 하나로 접었다, A #1 · codex #2). 준비 판정(B8)은 80ae96a5 와 같다.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2-fix/pre-edit/internal-app-engine--strategyriskauthorityloader.collectmarket/`. 변이 원장 `analysis/measurements/lot-5.2.2.2-fix/mutation-5.2.2.2-fix.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 194:2 — 결과 권한 준비 안 됨 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B2 | if at 197:2 — 환율 준비 안 됨 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B3 | if at 201:2 — 시장이 US 면 버킷 시장 US | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B4 | range at 207:2 — 범위마다 번들 하나 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B5 | if at 211:3 — 범위 키 정규화 성공 시에만 적재 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B6 | switch at 220:4 — **(새)** 적재 결과 분기 | `a112_owner_scope_trading_test.go` `TestACorruptLedgerRowStopsTheCycleWithItsCause` | no — 구조 | yes |
| B7 | case at 221:4 — **(새)** 적재 실패 → 적재기 원인 그대로 운반(범위 국소 신원 포함 여부는 1차 레그가 가름) | `a112_owner_scope_trading_test.go` `TestACorruptLedgerRowStopsTheCycleWithItsCause` · `TestARiskScopeOutsideTheSignedPolicyIsRefusedAloneInEitherOrder` | yes — `analysis/measurements/lot-5.2.2.2-fix/red-fix.log` · 변이 Y06(원인을 범위 국소로 덮음) CAUGHT | yes |
| B8 | case at 224:4 — 적재 성공 · 시장 · 계좌 · 시각 · 항목 5 일치 → 그 범위 준비 | 편집 전 번들 서술 그대로 | no — 이 로트가 바꾸지 않음 | 편집 전 번들의 측정 |
| B9 | case at 227:4 — **(새)** 적재 성공인데 시장 · 계좌 · 시각 · 항목 수 불일치 → 결함 원인 | 진입 0 — 생산 적재기가 같은 값으로 봉인(seam 으로 못 만듦) | no | 진입 0 |

> 재기준화: a112 게이트 준비(2026-10-04): a127 82080177 이 이 함수 본문을 편집(분기 · return 구조 동일, 호출 · 줄 이동) — 좌표 · 호출 표는 새 AST, 분기 의미 · a112 시험 인용은 그대로. a127 편집의 증거는 아카이브 번들 `openspec/changes/archive/2026-10-01-a127-strategy-authorities-read-the-current-ledger/analysis/function-logic/internal-app-engine--strategyriskauthorityloader.collectmarket/`(편집 뒤 FLM/BTM · a127 시험) — 분기 (id · 종류) · return 수 동일, 좌표는 id 끼리 사상, 호출 표는 새 AST.
