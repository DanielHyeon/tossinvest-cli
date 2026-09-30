# Function Logic Map: `alertDeliverer.recordFailedAttempt`

- Source: `internal/app/engine/alertdelivery.go`
- AST evidence: `ast.json` — **편집 뒤**, :348–393, 분기 9 · 반환 2 · 호출 20, source_sha256 `df8a4171e8dc…`, 추출 커밋 `b910173a`. 편집 전 번들은 `analysis/pre-edit/r26b/internal-app-engine--alertdeliverer.recordfailedattempt/`에 보존.
- Risk scan: `risk-pattern-report.md`
- 편집: (d8769cfb, codex P0 · #4) 반납 결과를 받아 행 없음 · 모르는 결과면 원칙 E 조건부 차단(승격 없음), 시도 기록 선점은 EventAlertClaimLost 로 기록. (b910173a, codex 재확인 R1) 반납 결과가 선점이어도 기록. 시도 한도 판정 · 승격 규칙 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` · `token` | 이 실행자가 claim 한 행과 임차 | 배달 실행자 사이클 | 아래 분기 |
| 반납 결과 | Applied · AlreadySettled · LeaseLost · NotFound · 모르는 값 | `ReleaseAlertClaim` | B2 · B3 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:350) | 시도 기록 오류 → 로그 | — | `TestAClearAfterTheLimitthRecordFailureStillEscalates`, `TestAClearBetweenRecordFailuresRestartsTheCount` |
| B2 | if (:357) | 반납이 행 없음 · 모르는 결과 → 로그 + 조건부 차단(승격 없음) | — | (미실행) |
| B3 | if (:365) | 반납이 선점(승인 · 남의 임차) → 선점 기록 | — | `TestA092TheDelivererRecordsAPreemptionSeenOnlyAtRelease` |
| B4 | if (:369) | 시도 기록 오류 → 행별 연속 기록 실패 계수, 반환 | — | `TestAClearAfterTheLimitthRecordFailureStillEscalates`, `TestAClearBetweenRecordFailuresRestartsTheCount` |
| B5 | switch (:375) | 시도 기록 결과 분기 | — | `TestA092AReleaseAfterTheEpochReadStandsForTheReleaseSite`, `TestA092TheAttemptLimitVerdictSurvivesAMissingRelease` |
| B6 | case (:376) | Applied → 연속 기록 실패 끝 | — | `TestA092AReleaseAfterTheEpochReadStandsForTheReleaseSite`, `TestA092TheAttemptLimitVerdictSurvivesAMissingRelease` |
| B7 | if (:379) | 한도 미만 → 반환 | — | `TestA092AReleaseAfterTheEpochReadStandsForTheReleaseSite`, `TestA092TheDelivererLatchesWhenTheReleaseFindsNoRow` |
| B8 | case (:383) | 시도 기록 선점 → 선점 기록 | — | `TestA092TheDelivererRecordsAPreemptedAttempt` |
| B9 | case (:388) | 행 없음 · 모르는 결과 → 조건부 차단(승격 없음) | — | `TestALatchOnlyJudgementNeverEscalatesEvenAfterAClear`, `TestARecordThatFindsNoRowLatchesButNeverEscalates` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `MarkAlertAttemptFailed` | 시도 기록 | 오류면 결과를 읽지 않음(F10) | AST |
| `release` → `ReleaseAlertClaim` | 반납 | (결과, ok) — 오류면 ok=false | AST |
| `judge` · `readEpoch` | 원칙 E | 근거 확정 뒤 세대 | AST |
| `logf(EventAlertClaimLost)` | 선점 기록 | 판정 없음 | AST |

## State mutations and fallbacks

- 원장 정산 · 반납 · 게이트 래치(judge) · 로그.

## Safety conclusion

- Safe edit boundary: 반납 결과 분류와 선점 기록만 — 시도 한도 판정 · 승격 규칙 불변(X05 CAUGHT).
- High-risk impact: yes — 배달 실행자의 진입 차단 판정(a124 정본 영역, a092 델타 「모든 발송자」).
