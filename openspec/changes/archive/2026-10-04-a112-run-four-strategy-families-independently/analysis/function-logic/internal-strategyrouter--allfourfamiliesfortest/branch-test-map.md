# Branch Test Map: `AllFourFamiliesForTest`

- Source SHA-256: `58436a4ec9e3f6893eeeea65dbd187910d3fe4950a958463ebd04ff094fb6d34` (base); AST branch locations are authoritative.
- Revision: **base-pinned (a112 7.3.1).** 본문 불변 — 바로 뒤 새 seam 추가로 diff 조각이 맞닿음.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range at 65:2 — 서술자 표 순회 | 이 seam 을 쓰는 시험 전부(engine · strategyworker 태그 스위트) | no — 본문 불변 | yes |
