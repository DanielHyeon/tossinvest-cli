# Branch Test Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go` (:191-246); **편집 뒤** 측정 — `analysis/harness/coverage-post-unit3.json`(연결 워크트리 `fbc6df5f`, `./internal/obs` 시험 99개를 하나씩). 편집 전 번들은 `analysis/pre-edit/unit3/`에 보존.
- 재번호: 없음.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 192:2 | 원장 없음 | `TestCriticalWithoutAJournalIsLoudRatherThanSilent`, `TestTheTransitionLogLineIsCountable` | 해당 없음(분기 불변) | 블록 192.22-196.19을 시험 2개가 실행, PASS |
| B2 | if at 196:3 | 로그 | `TestCriticalWithoutAJournalIsLoudRatherThanSilent`, `TestTheTransitionLogLineIsCountable` | 해당 없음(분기 불변) | 블록 196.19-200.4을 시험 2개가 실행, PASS |
| B3 | if at 216:2 | claim 실패 → 승격(반환값 무시 — 래치는 잠금 안에서 이미 무조건) | `TestAClaimThatFailsAttemptsTheDurableBlock`, `TestAClaimThatFailsBlocksNewEntries` | 해당 없음(분기 불변) | 블록 216.16-236.3을 시험 4개가 실행, PASS |
| B4 | if at 238:2 | 미전달 → `judge`(조건부 차단 → 승격 → 실패면 무조건 차단) | `TestA092AFailedEscalationLatchesUnconditionally`, `TestA092AReleaseAfterTheEpochReadIsHonoured` | 변이 L02 · L06 · L07 · L08 | 블록 238.19-244.3을 시험 22개가 실행, PASS |
