# Branch Test Map: `Notifier.escalate`

- Source: `internal/obs/notifier.go` (:419-441); **편집 뒤** 측정 — `analysis/harness/coverage-post-unit3.json`(연결 워크트리 `fbc6df5f`, `./internal/obs` 시험 99개를 하나씩). 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존.
- 재번호: 없음.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 420:2 | 승격 미포함 → `(false, nil)` | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseBeforeTheEvidenceDoesNotChangeTheVerdict` | 변이 L07 | 블록 420.63-422.3을 시험 16개가 실행, PASS |
| B2 | switch at 425:2 | 결과 분기 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 해당 없음(분기 불변) | 블록 423.2-425.9을 시험 12개가 실행, PASS |
| B3 | case at 426:2 | 승격 실패 로그 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092RecordOnlyFailureLatchesAndEscalates` | 해당 없음(분기 불변) | 블록 426.34-431.41을 시험 5개가 실행, PASS |
| B4 | case at 432:2 | 승격 됨 로그 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 432.31-438.92을 시험 7개가 실행, PASS |
