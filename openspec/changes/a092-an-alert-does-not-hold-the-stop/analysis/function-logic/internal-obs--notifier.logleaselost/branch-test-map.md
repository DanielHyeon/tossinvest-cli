# Branch Test Map: `Notifier.logLeaseLost`

- Source: `internal/obs/notifier.go` (:677-718); **편집 뒤** 측정 — `analysis/harness/coverage-post-r26b-obs.json`(연결 워크트리 `b910173a`, `internal/obs` 시험 116개를 하나씩, 실패 0). 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.logleaselost.json(AST)`에 보존.
- 재번호: 새 B6(`case res.Outcome != journal.SettleLeaseLost` — 모르는 결과), 편집 전 B6(default — 남의 임차) → B7.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 678:2 | 로거 없음 → 반환 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B2 | switch at 681:2 | 결과 분기 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 681.2-681.9을 시험 8개가 실행, PASS |
| B3 | case at 682:2 | 행 없음 → 오류 줄 | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` | 해당 없음(분기 불변) | 블록 682.45-690.19을 시험 4개가 실행, PASS |
| B4 | case at 691:2 | 이미 정산 → 선점 기록 | `TestA092AnAcknowledgementPreemptsASendInFlight`, `TestA092TheLeaseLossLogClassifiesByOutcome` | 해당 없음(분기 불변) | 블록 691.51-695.77을 시험 3개가 실행, PASS |
| B5 | case at 696:2 | LeaseLost · 토큰 없음 → 자기 반납 재확인 | `TestA092TheLeaseLossLogClassifiesByOutcome` | 해당 없음(분기 불변) | 블록 696.46-702.19을 시험 1개가 실행, PASS |
| B6 | case at 703:2 | 모르는 결과 → 원장 이상 오류 줄 | `TestA092TheLeaseLossLogClassifiesByOutcome` | Z02 CAUGHT(`TestA092TheLeaseLossLogClassifiesByOutcome/unknown`) | 블록 703.27-710.77을 시험 1개가 실행, PASS |
| B7 | case at 711:2 | LeaseLost · 남의 토큰 → 선점 경고 | `TestA092TheLeaseLossLogClassifiesByOutcome`, `TestASenderThatLosesTheLeaseStopsAtOnce` | 해당 없음(분기 불변) | 블록 711.10-716.55을 시험 2개가 실행, PASS |
