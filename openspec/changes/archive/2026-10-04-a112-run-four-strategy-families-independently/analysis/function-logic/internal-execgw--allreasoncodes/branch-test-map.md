# Branch Test Map: `AllReasonCodes`

- Source SHA-256: `fbe97b79db74c5e7801fd3e1c6301d13f261d0d6c5e880308b36323ca328ebd4`; AST branch locations are authoritative.
- Revision: **modified (태스크 5.6.2.1, 2026-09-30).** (5.6.2.1, 커밋 `3260f4eb`) 열거에 `ReasonStrategyCentralIntegrity` 한 줄 — 분기 없음. 골든 재생성 · a098 census 한 줄. 편집 전 번들은 `analysis/measurements/lot-5.6.2-5.2.2/pre-edit/`.
- 측정: 분기 없음 — 행동 증거는 RED 칸(`TestReasonCodeEnumIsStable` · 순서 시험).

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path — 분기 없는 열거 함수(정렬된 전체 목록) | `TestReasonCodeEnumIsStable` | 변이 E10(등록 누락) CAUGHT — `TestReasonCodeEnumIsStable` | 통과 |
