# Branch Test Map: `collect`

- Source SHA-256: `2e2e7dd7ade2428ad72eddde648e1798c1e6c19b082052a61804c1c226a6dcf6`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 분기 불변(6). 두 번째 반환값 strategyShadowPair 를 더했다 — 시장 goroutine 이 자기 묶음을 싣고, recover 갈래는 묶음을 부재 값으로 되돌린다.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-app-engine--strategyproposalauthorityloader.collect/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 259:2 — 입력 결함 → 실패 짝 + 부재 shadow 짝 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B2 | range at 268:2 — 시장 순회(KR · US goroutine) | `TestCollectReturnsTheShadowPairBesideTheAuthorityPair` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B3 | if at 275:6 — recover — **편집: shadow 를 부재 값으로** | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
| B4 | range at 288:2 — 결과 두 개 수신 | `TestCollectReturnsTheShadowPairBesideTheAuthorityPair` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B5 | if at 290:3 — KR 결과 | `TestCollectReturnsTheShadowPairBesideTheAuthorityPair` | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | yes — shadow 시험 1개가 진입(측정); 패키지 합집합 진입=True |
| B6 | else at 292:10 — US 결과 | 진입 없음 | no — 이 로트가 분기를 바꾸지 않음(편집 전 번들 `pre-edit/` 와 같은 조건) | no — 측정상 진입 0(커버리지 공백, 통과가 아님) |
