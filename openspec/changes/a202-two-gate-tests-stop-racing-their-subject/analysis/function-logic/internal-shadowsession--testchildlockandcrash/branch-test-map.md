# Branch Test Map: `TestChildLockAndCrash`

편집 전(base `7bc35e5d1`) AST 갈래. 시험 함수 자신이 대상이므로 「Test」 열은 그 갈래를 실제로 지나는 실행을 적는다.
RED 원문: `../../red/shadowsession-child-before.txt`(자식 직접 실행) · `../../red/shadowsession-old-test-parent-delay.txt`(사본, 부모 300 ms 지연).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 211 자식 갈래 | `TestChildLockAndCrash`(자식 실행) | yes — 자식이 `READY` 뒤 0.10 s 에 exit 2 | 구현 뒤 기록 |
| B2 | if 215 자식 `Begin` 오류 → exit 3 | 없음(정상 픽스처에서 안 지남) | no | 구현 뒤 기록 |
| B3 | select 221 `select {}` — 편집 대상 | 자식 직접 실행 `../../red/shadowsession-child-before.txt` | yes — `fatal error: all goroutines are asleep - deadlock!` | 구현 뒤 기록 |
| B4 | if 227 `StdoutPipe` 오류 | 없음(환경 결함) | no | 구현 뒤 기록 |
| B5 | if 230 `Start` 오류 | 없음(환경 결함) | no | 구현 뒤 기록 |
| B6 | if 234 defer 정리 | 실패 경로에서만 | no | 구현 뒤 기록 |
| B7 | if 240 `READY` 아님 | 없음 | no | 구현 뒤 기록 |
| B8 | if 243 부모 `Begin` 이 Busy 아님 | 사본 부모 지연 3/3 FAIL `lock: shadowsession: invalid or unavailable state` | yes | 구현 뒤 기록 |
| B9 | if 246 `Kill` 오류 | 없음(환경 결함) | no | 구현 뒤 기록 |
| B10 | if 251 lock 열기 오류 | 없음 | no | 구현 뒤 기록 |
| B11 | if 254 잠금 미해제 | 없음 | no | 구현 뒤 기록 |
| B12 | if 257 unlock 오류 | 없음(환경 결함) | no | 구현 뒤 기록 |
| B13 | if 260 close 오류 | 없음(환경 결함) | no | 구현 뒤 기록 |
| B14 | if 264 state 열기 오류 | 없음 | no | 구현 뒤 기록 |
| B15 | if 269 state 가 active 아님 | 없음 | no | 구현 뒤 기록 |
| B16 | if 272 죽은 active 재개가 Invalid 아님 | 없음 | no | 구현 뒤 기록 |
