# Branch Test Map: `TestTheReplacedReaderIsClosed`

편집 뒤(현재 AST, 191-214) 갈래. 편집 전 표는 커밋 `bbf0d9dcd` 의 이 파일.
RED 원문(편집 전 시험): `../../red/tossctl-old-test-prod-sleep.txt`(사본 생산 unlock↔close 50 ms → 5/5 FAIL) · `../../red/tossctl-old-test-control.txt`(무변이 사본 5/5 PASS).
GREEN 원문(편집 뒤 시험): `../../green/tossctl-new-test-none.txt`(무변이 사본 5/5 PASS) · `../../green/tossctl-new-test-prod-sleep.txt`(같은 50 ms 5/5 PASS) ·
`../../green/tossctl-new-test-prod-noclose.txt`(양성 대조 — 생산 닫기 호출 삭제 → 5/5 FAIL) · `../../verify/count200-tossctl.txt`(200/200 PASS).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| (반환 208) | 새 대기 술어 `closes >= 1` — 닫힘 자체를 기다림 | 사본 양성 대조(닫기 삭제) | — | 5/5 FAIL 「밀려난 reader 닫기: 2s 안에 일어나지 않았다」(시험이 닫힘을 잰다), 생산 그대로 · 50 ms 지연 모두 PASS |
| B1 | if 210 `closes != 1` → FAIL | `TestTheReplacedReaderIsClosed` | yes — 편집 전 시험이 사본 50 ms 지연에서 5/5 「0번 닫았다」 | pass — 같은 지연 5/5 PASS, `-count=200` 200/200 |
