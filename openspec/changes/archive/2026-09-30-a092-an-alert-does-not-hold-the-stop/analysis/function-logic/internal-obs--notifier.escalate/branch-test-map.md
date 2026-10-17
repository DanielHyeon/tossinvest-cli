# Branch Test Map: `Notifier.escalate`

- Source: `internal/obs/notifier.go` (:425-447); **편집 뒤** 측정 — `analysis/harness/coverage-post-r26d-obs.json`(연결 워크트리 `15b64676`, `internal/obs` 시험 118개를 하나씩, 실패 0). 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.escalate/`에 보존.
- 재번호: 없음.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 426:2 | 승격 미포함 | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseBeforeTheEvidenceDoesNotChangeTheVerdict` | 해당 없음(분기 불변) | 블록 426.63-428.3을 시험 16개가 실행, PASS |
| B2 | switch at 431:2 | 결과 분기 | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` | 해당 없음(분기 불변) | 블록 429.2-431.9을 시험 15개가 실행, PASS |
| B3 | case at 432:2 | 승격 실패 로그 | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` | W02 CAUGHT(`TestA092TheEscalationFailureErrorMasksTheAccount` — error 칸 가림). 키 이름(`trigger_event`)은 행동 시험이 세지 않음(로그 키, 판정 불변 — 비례 원칙) | 블록 432.34-437.41을 시험 8개가 실행, PASS |
| B4 | case at 438:2 | 승격 됨 로그 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 438.31-444.92을 시험 7개가 실행, PASS |
