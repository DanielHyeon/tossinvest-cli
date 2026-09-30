# Branch Test Map: `Notifier.escalate`

- Source: `internal/obs/notifier.go` (:423-445); **편집 뒤** 측정 — `analysis/harness/coverage-post-r26b-obs.json`(연결 워크트리 `b910173a`, `internal/obs` 시험 116개를 하나씩, 실패 0). 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.escalate/`에 보존.
- 재번호: 없음.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 424:2 | 승격 미포함 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AFailedRecordKeepsTheKeyOutOfTheGate` | 해당 없음(분기 불변) | 블록 424.63-426.3을 시험 17개가 실행, PASS |
| B2 | switch at 429:2 | 결과 분기 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 427.2-429.9을 시험 12개가 실행, PASS |
| B3 | case at 430:2 | 승격 실패 로그 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092RecordOnlyFailureLatchesAndEscalates` | 전용 시험 없음 — 이 줄의 키 이름은 행동 시험이 세지 않음(로그 키 변경, 판정 불변 — 비례 원칙) | 블록 430.34-435.41을 시험 5개가 실행, PASS |
| B4 | case at 436:2 | 승격 됨 로그 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 436.31-442.92을 시험 7개가 실행, PASS |
