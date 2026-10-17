# Branch Test Map: `alertDeliverer.recordFailedAttempt`

- Source: `internal/app/engine/alertdelivery.go` (:348-393); **편집 뒤** 측정 — `analysis/harness/coverage-post-r26b-engine.json`(연결 워크트리 `b910173a`, `internal/app/engine` 시험 37개를 하나씩, 실패 0). 편집 전 번들은 `analysis/pre-edit/r26b/internal-app-engine--alertdeliverer.recordfailedattempt/`에 보존.
- 재번호: 편집 전(e55102f0) B1 → B1, B2 → B4, B3~B7 → B5~B9. 새 B2(반납 행 없음 · 모르는 결과 — d8769cfb) · B3(반납 선점 기록 — b910173a).

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 350:2 | 시도 기록 오류 → 로그 | `TestAClearAfterTheLimitthRecordFailureStillEscalates`, `TestAClearBetweenRecordFailuresRestartsTheCount` | 해당 없음(분기 불변) | 블록 350.16-352.3을 시험 4개가 실행, PASS |
| B2 | if at 357:2 | 반납이 행 없음 · 모르는 결과 → 로그 + 조건부 차단(승격 없음) | (미실행) | X01 · X02 · X03 · X05 CAUGHT(`mutation-r26/ledger.tsv`) | 블록 좌표 없음(조건이 여러 줄이거나 본문이 비어 하네스가 같은 줄 블록을 못 잡음) — 행동 증거는 RED 칸 |
| B3 | if at 365:2 | 반납이 선점(승인 · 남의 임차) → 선점 기록 | `TestA092TheDelivererRecordsAPreemptionSeenOnlyAtRelease` | Z01 CAUGHT(`ledger-r26b.tsv`) | 블록 365.116-368.3을 시험 1개가 실행, PASS |
| B4 | if at 369:2 | 시도 기록 오류 → 행별 연속 기록 실패 계수, 반환 | `TestAClearAfterTheLimitthRecordFailureStillEscalates`, `TestAClearBetweenRecordFailuresRestartsTheCount` | 해당 없음(분기 불변) | 블록 369.16-374.3을 시험 4개가 실행, PASS |
| B5 | switch at 375:2 | 시도 기록 결과 분기 | `TestA092AReleaseAfterTheEpochReadStandsForTheReleaseSite`, `TestA092TheAttemptLimitVerdictSurvivesAMissingRelease` | 해당 없음(분기 불변) | 블록 375.2-375.21을 시험 16개가 실행, PASS |
| B6 | case at 376:2 | Applied → 연속 기록 실패 끝 | `TestA092AReleaseAfterTheEpochReadStandsForTheReleaseSite`, `TestA092TheAttemptLimitVerdictSurvivesAMissingRelease` | 해당 없음(분기 불변) | 블록 376.29-379.39을 시험 13개가 실행, PASS |
| B7 | if at 379:3 | 한도 미만 → 반환 | `TestA092AReleaseAfterTheEpochReadStandsForTheReleaseSite`, `TestA092TheDelivererLatchesWhenTheReleaseFindsNoRow` | X05 CAUGHT(한도 판정 보존) | 블록 379.39-381.4을 시험 11개가 실행, PASS |
| B8 | case at 383:2 | 시도 기록 선점 → 선점 기록 | `TestA092TheDelivererRecordsAPreemptedAttempt` | X04 CAUGHT | 블록 383.61-387.52을 시험 1개가 실행, PASS |
| B9 | case at 388:2 | 행 없음 · 모르는 결과 → 조건부 차단(승격 없음) | `TestALatchOnlyJudgementNeverEscalatesEvenAfterAClear`, `TestARecordThatFindsNoRowLatchesButNeverEscalates` | 해당 없음(분기 불변) | 블록 388.10-391.66을 시험 2개가 실행, PASS |
