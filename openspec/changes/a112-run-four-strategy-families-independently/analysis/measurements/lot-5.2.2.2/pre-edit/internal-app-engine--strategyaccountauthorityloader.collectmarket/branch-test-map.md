# Branch Test Map: `strategyAccountAuthorityLoader.collectMarket`

- Source SHA-256: `e6c12de7902b15de91de03a8da004f1ad167ce37d37580cf52766f4707ed0a4d`; AST branch locations are authoritative.
- Revision: 편집 전 현행(5.2.2.2 Pre-Edit). 이 표의 시험 칸은 편집 로트가 채운다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 155:2 — 제안 항목이 정확히 하나가 아니거나 그 제안이 유효하지 않음 → `StrategyAccountProposalNotReady`(시장당 하나 가정 — 5.2.2.2 가 범위별로 바꿈) | (편집 전 — 측정 안 함) | no | no |
| B2 | if at 158:2 — loader 구성 불완전(load · 키 · 설정 경로 · 계좌) → `StrategyAccountInternalFailure` | (편집 전 — 측정 안 함) | no | no |
| B3 | if at 163:2 — 시장이 US 면 계좌 시장 US | (편집 전 — 측정 안 함) | no | no |
| B4 | if at 169:2 — 계좌 권한 적재 실패 · 시장 · 매니페스트 불일치 → `StrategyAccountAuthorityUnavailable` | (편집 전 — 측정 안 함) | no | no |
