# Branch Test Map: `TestChildLockAndCrash`

편집 뒤(현재 AST, 211-278) 갈래. 시험 함수 자신이 대상이므로 「Test」 열은 그 갈래를 실제로 지나는 실행을 적는다. 편집 전 표(옛 B1~B16)는 커밋 `bbf0d9dcd` 의 이 파일.
RED 원문(편집 전): `../../red/shadowsession-child-before.txt`(자식 직접 실행 — 0.10 s exit 2) · `../../red/shadowsession-old-test-parent-delay.txt`(사본 부모 300 ms 지연 3/3 FAIL).
GREEN 원문(편집 뒤): `../../green/shadowsession-child-after.txt`(5 s 생존) · `../../green/shadowsession-new-test-parent-delay.txt`(같은 지연 3/3 PASS) ·
`../../verify/count200-shadowsession.txt`(200/200 PASS).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 212 자식 갈래 — 223 시한 있는 잠(옛 B3 `select {}` 자리) | 자식 직접 실행 | yes — 편집 전 `READY` 뒤 0.10 s exit 2, `fatal error: all goroutines are asleep - deadlock!` | pass — 5 s 뒤에도 살아 있음, stderr 빈 줄 |
| B2 | if 216 자식 `Begin` 오류 → exit 3 | 없음(정상 픽스처에서 안 지남) | no | no |
| B3 | if 230 `StdoutPipe` 오류 | 없음(환경 결함) | no | no |
| B4 | if 233 `Start` 오류 | 없음(환경 결함) | no | no |
| B5 | if 237 defer 정리 | 실패 경로에서만 | no | no |
| B6 | if 243 `READY` 아님 | 없음 | no | pass(거짓 갈래) |
| B7 | if 246 부모 `Begin` 이 Busy 아님 | 사본 부모 300 ms 지연 | yes — 편집 전 3/3 FAIL `lock: shadowsession: invalid or unavailable state` | pass — 같은 지연 3/3 PASS, `-count=200` 200/200 |
| B8 | if 249 `Kill` 오류 | 없음(환경 결함) | no | no |
| B9 | if 254 lock 열기 오류 | 없음 | no | no |
| B10 | if 257 잠금 미해제 | 이 시험(거짓 갈래) | no | pass(거짓 갈래 — kill 뒤 잠금 해제됨) |
| B11 | if 260 unlock 오류 | 없음(환경 결함) | no | no |
| B12 | if 263 close 오류 | 없음(환경 결함) | no | no |
| B13 | if 267 state 열기 오류 | 없음 | no | no |
| B14 | if 272 state 가 active 아님 | 이 시험(거짓 갈래) | no | pass(거짓 갈래) |
| B15 | if 275 죽은 active 재개가 Invalid 아님 | 이 시험(거짓 갈래) | no | pass(거짓 갈래) |
