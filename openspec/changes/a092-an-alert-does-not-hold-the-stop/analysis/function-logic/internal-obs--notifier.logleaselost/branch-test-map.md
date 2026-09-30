# Branch Test Map: `Notifier.logLeaseLost`

- Source: `internal/obs/notifier.go` (:679-720); **편집 뒤** 측정 — `analysis/harness/coverage-post-r26d-obs.json`(연결 워크트리 `15b64676`, `internal/obs` 시험 118개를 하나씩, 실패 0). 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.logleaselost.json(AST)`에 보존.
- 재번호: 새 B6(`case res.Outcome != journal.SettleLeaseLost` — 모르는 결과), 편집 전 B6(default — 남의 임차) → B7.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 680:2 | 로거 없음 → 반환 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B2 | switch at 683:2 | 결과 분기 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 683.2-683.9을 시험 8개가 실행, PASS |
| B3 | case at 684:2 | 행 없음 → 오류 줄 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 684.45-692.19을 시험 4개가 실행, PASS |
| B4 | case at 693:2 | 이미 정산 → 선점 기록 | `TestA092AnAcknowledgementPreemptsASendInFlight`, `TestA092TheLeaseLossLogClassifiesByOutcome` | 해당 없음(분기 불변) | 블록 693.51-697.77을 시험 3개가 실행, PASS |
| B5 | case at 698:2 | LeaseLost · 토큰 없음 → 자기 반납 재확인 | `TestA092TheLeaseLossLogClassifiesByOutcome` | 해당 없음(분기 불변) | 블록 698.46-704.19을 시험 1개가 실행, PASS |
| B6 | case at 705:2 | 모르는 결과 → 원장 이상 오류 줄 | `TestA092TheLeaseLossLogClassifiesByOutcome` | Z02 CAUGHT(`TestA092TheLeaseLossLogClassifiesByOutcome/unknown`) | 블록 705.27-712.77을 시험 1개가 실행, PASS |
| B7 | case at 713:2 | LeaseLost · 남의 토큰 → 선점 경고 | `TestA092TheLeaseLossLogClassifiesByOutcome`, `TestASenderThatLosesTheLeaseStopsAtOnce` | 해당 없음(분기 불변) | 블록 713.10-718.55을 시험 2개가 실행, PASS |
