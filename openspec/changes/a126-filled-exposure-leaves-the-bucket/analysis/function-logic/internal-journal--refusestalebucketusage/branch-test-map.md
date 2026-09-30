# Branch Test Map: `refuseStaleBucketUsage`

주석 한 줄 편집 — 분기 무변. 떠남이 이 함수에 닿는 효과(B3 거절 전파 · B4 latch · B12 기록 한도 cap)는 a126 시험이 잰다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 길이 불일치 | 기존 | n/a | n/a |
| B2 | 순회 | 기존 | n/a | n/a |
| B3 | 판독 오류 | a126 1.1 손상 · 불일치 거절 | no | no |
| B4 | latch 차단 | a126 M7 | no | no |
| B5 | ledger 금액 아님 | 기존 | n/a | n/a |
| B6 | 주장 금액 아님 | 기존 | n/a | n/a |
| B7 | stale | 기존 | n/a | n/a |
| B8 | cap 키 불일치 | 기존 | n/a | n/a |
| B9 | 한도 판독 오류 | 기존 | n/a | n/a |
| B10 | 기록 한도 없음 | 기존 | n/a | n/a |
| B11 | 합 금액 아님 | 기존 | n/a | n/a |
| B12 | cap 소진 | a126 M4 | no | no |
