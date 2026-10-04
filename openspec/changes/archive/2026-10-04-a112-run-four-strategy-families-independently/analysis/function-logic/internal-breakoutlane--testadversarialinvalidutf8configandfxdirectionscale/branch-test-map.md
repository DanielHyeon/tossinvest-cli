# Branch Test Map: `TestAdversarialInvalidUTF8ConfigAndFXDirectionScale (시험)`

- Source SHA-256: `878c75713f0d714c0a6569f8f7b314a7f9796611f38d28fb50939f8138982b3e`; AST branch locations are authoritative.
- Revision: **modified (a112 B2, 2026-10-01).** 역방향 FX 봉인 단언을 반전 — 앞 판은 digest 없는 역방향 봉인의 성공을 박아 결함(호출자 digest 미검증)을 핀하고 있었다. 이제 digest 없이는 거절, 주어진 모양으로 봉인하면 정규화 통과(Manager 판정 B2).
- 편집 전 번들: `analysis/measurements/lot-bk/pre-edit/internal-breakoutlane--testadversarialinvalidutf8configandfxdirectionscale/`. 변이 원장 `analysis/measurements/lot-bk/mutation-b2.tsv`.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 311:2 — 이 시험 자신의 판정 갈래(UTF-8 거절 · 역방향 digest 없음 거절 · 봉인된 역방향 정규화) | 이 시험 자신 | no — 시험 코드 | yes — 반전 전 판은 수리 뒤 실패(`red-b2.log` 의 반대 방향) |
| B2 | if at 317:2 — 이 시험 자신의 판정 갈래(UTF-8 거절 · 역방향 digest 없음 거절 · 봉인된 역방향 정규화) | 이 시험 자신 | no — 시험 코드 | yes — 반전 전 판은 수리 뒤 실패(`red-b2.log` 의 반대 방향) |
| B3 | if at 322:2 — 이 시험 자신의 판정 갈래(UTF-8 거절 · 역방향 digest 없음 거절 · 봉인된 역방향 정규화) | 이 시험 자신 | no — 시험 코드 | yes — 반전 전 판은 수리 뒤 실패(`red-b2.log` 의 반대 방향) |
