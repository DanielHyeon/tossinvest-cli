# Branch Test Map: `Notifier.escalate`

- Source: `internal/obs/notifier.go` (:419-441); **편집 뒤** 측정 — `analysis/harness/coverage-post-r25fix.json`(연결 워크트리 `55963f29`, `./internal/obs` 시험 99개를 하나씩). 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존.
- 재번호: 없음.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 423:2 | 승격 미포함 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseBeforeTheEvidenceDoesNotChangeTheVerdict` | 변이 L07 | 블록 423.63-425.3을 시험 16개가 실행, PASS |
| B2 | switch at 428:2 | 결과 분기 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 426.2-428.9을 시험 12개가 실행, PASS |
| B3 | case at 429:2 | 승격 실패 로그 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092RecordOnlyFailureLatchesAndEscalates` | 해당 없음(분기 불변) | 블록 429.34-434.41을 시험 5개가 실행, PASS |
| B4 | case at 435:2 | 승격 됨 로그 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 435.31-441.92을 시험 7개가 실행, PASS |
