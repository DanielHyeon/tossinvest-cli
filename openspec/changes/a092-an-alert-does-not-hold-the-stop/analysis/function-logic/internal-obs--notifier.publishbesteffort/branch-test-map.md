# Branch Test Map: `Notifier.publishBestEffort`

- Source: `internal/obs/notifier.go` (:179-192); **편집 뒤** 측정 — `analysis/harness/coverage-post-r26b-obs.json`(연결 워크트리 `b910173a`, `internal/obs` 시험 116개를 하나씩, 실패 0). 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.publishbesteffort.json(AST)`에 보존.
- 재번호: 없음.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 180:2 | 발행기 없음 → 반환 | `TestA092DropAndNoJournalLinesKeepTheirOwnEventKey`, `TestTheTransitionLogLineIsCountable` | 해당 없음(분기 불변) | 블록 180.24-182.3을 시험 2개가 실행, PASS |
| B2 | if at 183:2 | 발행 실패 → 경고 줄 | `TestA092DropAndNoJournalLinesKeepTheirOwnEventKey`, `TestObservationFailureAlertNeverReachesTheGateOrTheMode` | Z13 CAUGHT | 블록 183.95-191.3을 시험 4개가 실행, PASS |
