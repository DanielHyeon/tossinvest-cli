# Branch Test Map: `LaneJSONFields`

- Source SHA-256: `9b6b15b0ebe8e74fcd582a8a6d6abed626706542ba9706e5ba551a3c302315a4`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3.1 SHADOW, 2026-10-01).** 분기 없음. 계약 JSON 이름 목록에 `shadowOutcome` 을 더했다.
- 편집 전 번들: `analysis/measurements/lot-7.3.1-shadow/pre-edit/internal-strategyprojection--lanejsonfields/`. 변이 원장 `analysis/measurements/lot-7.3.1-shadow/mutation-7.3.1-S.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path (분기 없음) — 분기 없음 — 분기 없음. 계약 JSON 이름 목록에 `shadowOutcome` 을 더했다. | 이름 목록은 실제 직렬화(`TestLaneAndCoordinatorJSONNamesAreTheContract`)와 OpenAPI(`TestOpenAPIDocumentsTheLaneAndCoordinatorChildrenByTheirExactNames`)가 잰다. | no — 시험 코드 | yes |
