# Branch Test Map: `TestImportsStayInsideTheAllowlist` (tools/a140-breakout-signal-shadow)

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 132 파일 순회 | 자신 | yes | yes |
| B2 | 133 import 순회 | 자신 | yes | yes |
| B3 | 135 표준 라이브러리 건너뜀 | 자신 | yes | yes |
| B4 | 138 목록 밖 모듈 import | 자신 — RED 두 예산 경로, 변이 a140-e-import(`_ ".../internal/orderintent"`) | yes(RED 원인) | 변이 CAUGHT |
