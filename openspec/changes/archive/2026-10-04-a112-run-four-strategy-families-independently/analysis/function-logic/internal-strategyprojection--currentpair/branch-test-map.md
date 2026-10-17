# Branch Test Map: `currentPair (시험 도우미)`

- Source SHA-256: `44ffddfa4f011fdfb8484929cb28babca7a8367a43daeec90c1042667ed74dd0`; AST branch locations are authoritative.
- Revision: **modified (a112 7.3, 2026-10-01).** 시험 도우미. 옛 모양 리터럴(자식 없음) 대신 DormantSnapshot 에서 시작해 시장 레코드만 바꾼다 — envelope 계약이 자식을 요구한다.
- 편집 전 번들: `analysis/measurements/lot-7.3/pre-edit/internal-strategyprojection--currentpair/`. 변이 원장 `analysis/measurements/lot-7.3/mutation-7.3.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 116:2 — KR · US 시장 레코드 만들기 | `TestMarketFailureReplacesOnlyExactMarketWithoutFallback` · `TestEitherMarketFailurePreservesTheExactPeer` | no — 시험 코드 | yes |
| B2 | if at 138:2 — 만든 스냅숏이 Validate 를 통과 | 같은 시험들(도우미가 t.Fatal) | no — 시험 코드 | yes |
