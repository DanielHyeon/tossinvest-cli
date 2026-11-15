# Branch Test Map: `TestImportsStayInsideTheAllowlist` (tools/a128-jev-shadow-probe)

변이 원문: 앞 로트 `../../mutation/mutation-after-priority-pin.txt`(a128-e-import), 결정 1 `../../mutation-2/mutation2-run10.txt`(d1*).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 220 파일 순회 | 자신 | yes | yes |
| B2 | 221 import 순회 | 자신 | yes | yes |
| B3 | 223 목록 밖 모듈 import | 자신 — RED 에서 두 예산 경로, 변이 a128-e-import(`_ ".../internal/orderintent"`) | yes(RED 원인) | 변이 CAUGHT |
| B4 | 226 세 패키지의 이름 붙인 import | 자신 — d1a-a128(별칭)·d1b-a128(dot)·d1c-a128(blank) FAIL; PRE1 판본은 d1a·d1b PASS(못 잡음) | yes | yes |
