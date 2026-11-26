# Branch Test Map: `TestTheOverviewSpendsNothingHoweverOftenItReloads`

- 편집 후(a205 r1 #1 처분). 실패 원문 `../../verify/green-attempt1.txt`(상수 2s · 횟수 10), 통과 원문 `../../verify/green.txt`.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 재로드 창이 TTL 이하 | `TestTheOverviewSpendsNothingHoweverOftenItReloads` | yes — green-attempt1.txt `line_cadence_test.go:75: the redraw window is 20s and the cache TTL is 30s` | yes — 거짓 갈래(32s > 30s) |
| B2 | 16회 재로드 | 같은 시험 | 해당 없음 | yes |
| B3 | 개요가 브로커 호출 | 같은 시험 | 해당 없음(거짓 갈래) | yes — 거짓 갈래 |
| B4 | 양성 대조 실패 | 같은 시험 | 해당 없음(거짓 갈래) | yes — 거짓 갈래 |
