# Function Logic Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json` — **편집 뒤**, :477–645, 분기 27 · 반환 7 · 호출 42, source_sha256 `d705f78d68c1…`, 추출 커밋 `15b64676`. 편집 전 번들은 `analysis/pre-edit/r26b/internal-obs--notifier.deliver/`에 보존.
- Risk scan: `risk-pattern-report.md`
- 편집: (d8769cfb, codex #6) 시도 기록의 행 없음 갈래 조건 `failed.Outcome == SettleNotFound` → `!isPreemption(failed.Outcome)`, 반납 갈래 조건 → `!isPreemption(released.Outcome)` — 한 판정(모르는 결과도 선점 아님). 분기 수 · 순서 불변(27).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` · `token` | claim 한 행과 임차 | `claimAndDeliver`(잠금 안 claim) | 토큰이 안 맞으면 정산 거절 |
| `n.mu` | **쥐지 않음** | — | 배제는 임차 |
| 해제 세대 | 근거 확정 직후 `readVerdict` 가 읽음 | `EntryGate.ClearEpoch` | 원칙 E |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if (:479) | 기본 시도 수 | — | (미실행) |
| B2 | for (:485) | 시도 루프 | — | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` |
| B3 | if (:486) | 발행기 없음 | — | (미실행) |
| B4 | if (:491) | 발행 성공 | — | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` |
| B5 | if (:493) | 정산 오류 없음 | — | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` |
| B6 | switch (:494) | 정산 결과 분기 | — | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` |
| B7 | case (:495) | 정산됨 | — | `TestA092ARecordDoesNotWaitForAnotherSendersTransport`, `TestA092ATakeoverIsAWarningThatNamesTheDeadSender` |
| B8 | case (:497) | 선점(승인 · 남의 임차) — 래치 없음 | — | `TestA092AnAcknowledgementPreemptsASendInFlight`, `TestAcknowledgeCannotClearTheGateMidSend` |
| B9 | case (:509) | 행 없음 → 미정산 | — | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` |
| B10 | case (:511) | 모르는 결과 → 미정산 | — | (미실행) |
| B11 | if (:537) | **unrecorded 판정** | — | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` |
| B12 | if (:551) | 시도 기록 오류 | — | `TestACancelledSenderStillHandsTheLeaseBack` |
| B13 | else (:555) | 오류 없음 분기 | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| B14 | if (:552) | 로그 | — | `TestACancelledSenderStillHandsTheLeaseBack` |
| B15 | if (:555) | 임차 상실 · 정산됨 · 행 없음 | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| B16 | if (:565) | 전송 실패 로그 | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| B17 | if (:575) | **시도 기록이 선점 아님(행 없음 · 모르는 결과 — `isPreemption`) → 원칙 E 조건부 차단** | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| B18 | if (:584) | 대기 | — | `TestACancelledSenderStillHandsTheLeaseBack`, `TestACriticalAlertStillEscalatesThroughTheSameNotifier` |
| B19 | if (:585) | 문맥 종료 | — | `TestACancelledSenderStillHandsTheLeaseBack` |
| B20 | switch (:603) | 반납 결과 분기 | — | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` |
| B21 | case (:604) | 반납 오류 | — | (미실행) |
| B22 | if (:605) | 로그 | — | (미실행) |
| B23 | case (:608) | 반납 적용 → 소진 판정 | — | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` |
| B24 | case (:611) | **반납이 선점 아님(행 없음 · 모르는 결과 — `isPreemption`) → 로그** | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| B25 | if (:616) | **그 자리의 원칙 E 조건부 차단(새 분기)** | — | `TestA092AReleaseAfterTheEpochReadIsHonoured`, `TestA092AReleaseBeforeTheEpochReadRelatches` |
| B26 | case (:623) | 반납 선점(AlreadySettled · LeaseLost) → lost | — | (미실행) |
| B27 | if (:639) | **exhausted 판정** 로그 | — | `TestA092AFailedEscalationLatchKeepsTheAccountOut`, `TestA092AFailedEscalationLatchesUnconditionally` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `isPreemption` | 선점 판정 한 곳 | AlreadySettled · LeaseLost 만 참 | AST · `TestA092DeliverClassifiesThroughOneJudgement` |
| `MarkAlertAttemptFailed` · `ReleaseAlertClaim` · `MarkAlertDelivered` | 정산 | 결과 분류 | AST |
| `readVerdict` · `BlockUnlessClearedSince` | 원칙 E | — | AST |

## State mutations and fallbacks

- 원장 정산 · 게이트 조건부 래치 · 로그.

## Safety conclusion

- Safe edit boundary: 두 조건식만 — 분기 구조 · 래치 순서 불변.
- High-risk impact: yes — 동기 critical 발송의 진입 차단 판정.
