# Branch Test Map: `signalsView.freshness`

- Source: `internal/console/chrome.go` (323-334) — 편집 뒤. AST branches 1.
- 측정: 커버리지 시험 바이너리(`go test -c -cover`)로 top-level 시험 607 개를 하나씩 `-test.run '^<Test>$' -test.coverprofile` 실행, 분기 본문 블록 count>0 인 시험만 인용(2026-10-25, a173 close). n 은 그 블록을 지나는 시험 수.

| Branch | Anchor | Scenario | Test | 측정 n | a173 |
|---|---|---|---|---|---|
| B1 | `324:2` | !Known → unavailable · 사유 코드 | `TestAStoreThatCouldNotBeReadIsItsOwnReasonAndNotTheJournalsOrTheBrokers` · `TestEachUnmeasuredCauseIsExplainedOnce` | 22 | a173: Code 추가. 거짓 갈래는 327행 cache 반환 |

- 거짓 갈래(327행 반환): `TestAKnownRefreshHoldGetsAReasonInsteadOfATone` 등 22 개 시험이 지남(측정).
