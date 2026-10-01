# Branch Test Map: `TestClientIgnoresAdditiveFieldsFromANewerEngine (시험)`

- Source SHA-256: `3d365481c525d3099026c29999bdaf048b5f7da7bedcdacf0cb8bc8873b5a0c6`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3, 2026-10-01).** 7.3 이 `coordinators` · `lanes` 를 아는 필드로 만들어 심던 이름을 아직 모르는 이름으로 바꿨다(envelope · 시장 레코드 · 레인 자식 세 층). 재는 것 불변.
- 편집 전 번들: `analysis/measurements/lot-7.3/pre-edit/internal-strategyprojectionrpc--testclientignoresadditivefieldsfromanewerengine/`. 변이 원장 `analysis/measurements/lot-7.3/mutation-7.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 28:2 — arrangement 오류 가드 | 이 시험 자신 | no — 시험 코드 | yes |
| B2 | if at 36:2 — arrangement 오류 가드 | 이 시험 자신 | no — 시험 코드 | yes |
| B3 | if at 41:2 — arrangement 오류 가드 | 이 시험 자신 | no — 시험 코드 | yes |
| B4 | if at 46:2 — arrangement 오류 가드 | 이 시험 자신 | no — 시험 코드 | yes |
| B5 | if at 51:2 — arrangement 오류 가드 | 이 시험 자신 | no — 시험 코드 | yes |
| B6 | if at 55:2 — arrangement 오류 가드 | 이 시험 자신 | no — 시험 코드 | yes |
| B7 | if at 59:2 — arrangement 오류 가드 | 이 시험 자신 | no — 시험 코드 | yes |
| B8 | if at 65:3 — 토큰 불일치 → 401 | 이 시험 자신(서버 stub) | no — 시험 코드 | yes |
| B9 | if at 76:2 — 구 reader 가 additive 필드에 죽음 → 실패 | 이 시험 자신 | yes — 7.3 GREEN 직후 옛 주입(`coordinators` 부분 객체)이 새 계약에 거절됨을 관측(`green-7.3-first.log`) | yes |
| B10 | if at 80:2 — 기존 필드 의미 · 형식 유지 | 이 시험 자신 | no — 시험 코드 | yes |
