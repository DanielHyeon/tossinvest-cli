# Branch Test Map: `newProductionStrategyFirstLegAuthorityLoader`

- Source SHA-256: `d0d6281292dafcc979edce741a3a2bf98ed348f023267d8198d2436c71ec7291`; AST branch locations are authoritative.
- Revision: **modified (a112 6.2 봉인 리뷰 수리, 2026-10-01)** — 제안 쌍을 떼어 내어 든다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | happy path — 제안 쌍을 떼어 낸 loader | `TestAnInPlaceSwapInTheDispatchCopyDoesNotReachTheSeal`(KR · US) | yes — `red-6.2-seal-fix.log`(편집 전 686b94e4: 제자리 교체된 쌍둥이가 발급됨, err=nil) | yes — 변이 S10(떼어 내지 않음) · S11(한 시장만) CAUGHT |
