# Branch Test Map: `TestImportsStayInsideTheAllowlist` (tools/a128-jev-shadow-probe)

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 193 파일 순회 | 자신 | yes | yes |
| B2 | 194 import 순회 | 자신 | yes | yes |
| B3 | 196 목록 밖 모듈 import | 자신 — RED 에서 두 예산 경로, 변이 a128-e-import(`_ ".../internal/orderintent"`) | yes(RED 원인) | 변이 CAUGHT |
