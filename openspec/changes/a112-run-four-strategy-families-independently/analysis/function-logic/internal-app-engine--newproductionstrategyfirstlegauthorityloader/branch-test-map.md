# Branch Test Map: `newProductionStrategyFirstLegAuthorityLoader`

- Source SHA-256: `c29e90e2a1e9f04e531a1cc000caa4bcc6a97a0cd796a73845a97dd23e1de7ca`; AST branch locations are authoritative.
- Revision: **modified (a112 6.2 봉인 리뷰 수리, 2026-10-01)** — 제안 쌍을 떼어 내어 든다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path — 제안 쌍을 떼어 낸 loader | `TestAnInPlaceSwapInTheDispatchCopyDoesNotReachTheSeal`(KR · US) | yes — `red-6.2-seal-fix.log`(편집 전 686b94e4: 제자리 교체된 쌍둥이가 발급됨, err=nil) | yes — 변이 S10(떼어 내지 않음) · S11(한 시장만) CAUGHT |
