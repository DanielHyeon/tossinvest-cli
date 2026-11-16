# Branch Test Map: `TestImportsStayInsideTheAllowlist` (tools/a140-breakout-signal-shadow)

변이 원문: 앞 로트 `../../mutation/mutation-after-priority-pin.txt`(a140-e-import), 결정 1 `../../mutation-2/mutation2-run10.txt`(d1*).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 154 파일 순회 | 자신 | yes | yes |
| B2 | 155 import 순회 | 자신 | yes | yes |
| B3 | 157 표준 라이브러리 건너뜀 | 자신 | yes | yes |
| B4 | 160 목록 밖 모듈 import | 자신 — RED 두 예산 경로, 변이 a140-e-import(`_ ".../internal/orderintent"`) | yes(RED 원인) | 변이 CAUGHT |
| B5 | 163 세 패키지의 이름 붙인 import | 자신 — d1a-a140(별칭)·d1b-a140(dot)·d1c-a140(official blank) FAIL; PRE1 판본은 d1a·d1b PASS(못 잡음) | yes | yes |
