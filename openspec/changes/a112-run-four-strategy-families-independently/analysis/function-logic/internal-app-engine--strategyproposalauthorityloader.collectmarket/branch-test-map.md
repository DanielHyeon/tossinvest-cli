# Branch Test Map: `collectMarket`

- Source SHA-256: `2e2e7dd7ade2428ad72eddde648e1798c1e6c19b082052a61804c1c226a6dcf6`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 분기 불변(15). out 인자 `shadow` 를 더했다: 첫 문장에서 부재 값으로 비우고(조정 앞 닫힘 일곱 = 관측 없음), 조정 바로 뒤 대입 하나로 수집 묶음 + 결속 설정(`loader.shadowConfig`)을 싣는다. 반환 갈래 열다섯 불변.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--strategyproposalauthorityloader.collectmarket/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 325:2 — 경로 · 스케줄 미준비 → ROUTE_NOT_READY | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B2 | if at 338:2 — FX 미준비 | `TestCollectMarketCarriesTheShadowBatchOnlyFromCoordination` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B3 | if at 341:2 — 적재기 설정 결함 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B4 | if at 346:2 — 제안 공개 열쇠 무효 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B5 | if at 350:2 — US env 이름 | `TestCollectReturnsTheShadowPairBesideTheAuthorityPair` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B6 | range at 356:2 — 경로 항목 순회 | `TestCollectMarketCarriesTheShadowBatchOnlyFromCoordination`, `TestCollectReturnsTheShadowPairBesideTheAuthorityPair` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 2개가 진입(측정); 패키지 합집합 진입=True |
| B7 | if at 358:3 — 종목 중복 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B8 | if at 370:2 — 제안 적재 실패 · digest 불일치 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B9 | if at 375:2 — 받아들인 범위가 제안을 잃음 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B10 | if at 412:2 — 관문이 범위를 지움 → FAMILY_GATE_CLOSED(묶음 실음) | `TestCollectMarketCarriesTheShadowBatchOnlyFromCoordination` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B11 | if at 419:2 — 계보 충돌(조정자의 부재 값 그대로) | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B12 | if at 428:2 — 큐 넘침 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B13 | if at 436:2 — 중재 거절 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
| B14 | if at 446:2 — 미해결 선택 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B15 | if at 454:2 — 받아들인 범위 0 | shadow 시험 밖의 패키지 시험(합집합 측정) | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — 패키지 합집합 진입(측정) |
