# Branch Test Map: `TestTheReplacedReaderIsClosed`

편집 전(base `7bc35e5d1`) AST 갈래. RED 원문: `../../red/tossctl-old-test-prod-sleep.txt`(사본, 생산 unlock 과 close 사이 50 ms) ·
대조 `../../red/tossctl-old-test-control.txt`(무변이 사본).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if 208 `closes != 1` → FAIL | `TestTheReplacedReaderIsClosed` | yes — 사본 50 ms 지연에서 5/5 「0번 닫았다」, 무변이 사본 5/5 PASS | 구현 뒤 기록 |
