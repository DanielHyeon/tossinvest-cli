# Branch Test Map: `TestProductionFirstLegAuthorityLoaderPairedKRUS`

- Source SHA-256: `47f555f84382d2768f30dbce87c834c0ad832c28664c595b0779b7c295b1ce1e`; AST branch locations are authoritative.
- Revision: **modified (a112 5.2.2.2, 2026-10-01).** a112 5.2.2.2: 손으로 만든 계좌 권한에 범위 목록을 싣는 한 줄(`a112ScopedAccount`)만 더했다 — 1차 레그 권한이 이제 범위별 계좌 권한을 `forScope` 로 고르므로(봉투 폴백 없음) 범위 없는 fixture 는 거절된다. 분기 불변.
- 편집 전 번들: `analysis/measurements/lot-5.2.2.2/pre-edit/internal-app-engine--testproductionfirstlegauthorityloaderpairedkrus/`. 변이 원장 `analysis/measurements/lot-5.2.2.2/mutation-5.2.2.2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 26:2 — 시험 본문 분기(편집 불변) | 이 시험 자신 | no — 시험 코드 | yes |
| B2 | if at 30:3 — 시험 본문 분기(편집 불변) | 이 시험 자신 | no — 시험 코드 | yes |
| B3 | if at 37:3 — 시험 본문 분기(편집 불변) | 이 시험 자신 | no — 시험 코드 | yes |
| B4 | if at 48:3 — 시험 본문 분기(편집 불변) | 이 시험 자신 | no — 시험 코드 | yes |
| B5 | else at 50:10 — 시험 본문 분기(편집 불변) | 이 시험 자신 | no — 시험 코드 | yes |
| B6 | if at 58:2 — 시험 본문 분기(편집 불변) | 이 시험 자신 | no — 시험 코드 | yes |
| B7 | range at 63:2 — 시험 본문 분기(편집 불변) | 이 시험 자신 | no — 시험 코드 | yes |
| B8 | if at 66:3 — 시험 본문 분기(편집 불변) | 이 시험 자신 | no — 시험 코드 | yes |
| B9 | if at 70:3 — 시험 본문 분기(편집 불변) | 이 시험 자신 | no — 시험 코드 | yes |
| B10 | if at 74:3 — 시험 본문 분기(편집 불변) | 이 시험 자신 | no — 시험 코드 | yes |
| B11 | if at 82:2 — 시험 본문 분기(편집 불변) | 이 시험 자신 | no — 시험 코드 | yes |
