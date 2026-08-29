# Branch Test Map: `Recovery.Run`

Source: `internal/reconcile/recovery.go` (238-329). AST 기준 branches **12** / returns 8.

## 커버리지는 주장이 아니라 측정값이다

`go test ./internal/reconcile/ -count=1 -coverprofile`(exit 0 · **147건 통과** · 86.6%)의
블록 카운트를 `recovery.go:238-329`로 잘라 읽은 것이다.

| Branch | 위치 | 조건 평가 | 본문 실행 | 근거 블록 | 지는 테스트 |
|---|---|---|---|---|---|
| B1 | `:245` `RecoverPending` 오류 | yes | **no** | `245.16,248.3` count=0 | — |
| B2 | `:258` `PendingAttempts` 오류 | yes | **no** | `258.16,260.3` count=0 | — |
| B3 | `:261` `range pending` | yes | **yes** | `261.30,262.40` count=1 | `TestCrashMidDispatchBecomesInDoubtAndIsResolved` |
| B4 | `:262` IN_DOUBT가 아님 | yes | **no** | `262.40,268.19` count=0 | — |
| B5 | `:268` `blockedSymbol` 오류 | **no — 도달 없음** | **no** | `268.19,270.5` count=0 | — |
| B6 | `:276` `replay` 오류 | yes | **yes** | `276.18,279.4` count=1 | `replay_recovery_test.go`의 replay 실패 경로 |
| B7 | `:280` replay가 해결함 | yes | **yes** | `280.14,281.12` count=1 | 같음 |
| B8 | `:285` `Resolve` 오류 | yes | **no** | `285.18,288.4` count=0 | — |
| B9 | `:290` 미해결 IN_DOUBT | yes | **no** | `290.50,292.4` count=0 | — |
| B10 | `:302` `stableSnapshot` 오류 | yes | **yes** | `302.16,304.3` count=1 | `TestRecoveryFailsClosedWhenTheAccountWillNotSettle` · `TestNonRateLimitedCollectStillFailsImmediately` · `TestRateLimitWaitBudgetExhaustionFailsClosed` · `TestPermanentlyRefusingBrokerExhaustsTheBudget` · `TestBudgetTooSmallForOneBackoffSaysSo` |
| B11 | `:309` `LocalStateFromJournal` 오류 | yes | **no** | `309.16,311.3` count=0 | — |
| B12 | `:324` `Diff.BlocksEntry()` | yes | **yes** | `324.31,327.3` count=1 | `TestRecoveryThatEndsInDisagreementBlocksEntriesForADifferentReason` |

**a102가 편집한 자리는 B10의 바로 앞 네 줄(`:298-301`)이고, B10의 본문은 실행된다.**
그 밖의 미덮임(B1·B2·B4·B5·B8·B9·B11)은 전부 **a102 이전부터 있던 것**이고 이 change는
그 팔들을 건드리지 않았다 — 편집 전후 AST 수치가 동일하다는 것이 그 근거다.

> ⚠ **물려받은 공백을 조용히 넘기지 않는다.** 저널·해소기의 오류 팔 일곱은 이 함수에서
> 한 번도 실행되지 않는다(각 패키지 자체 테스트는 별개다). a102 §1은 그것을 고치지 않는다 —
> 범위는 3단계의 429 처리다.

## a102 §1이 이 표에 지는 것

| 성질 | 지는 테스트 |
|---|---|
| 3단계가 429를 넘기고 완주하면 게이트가 열린다 | `TestRateLimitedCollectDoesNotEndRecovery` |
| 3단계의 실측이 `Report`에 실린다 (`SnapshotsTaken`·`RateLimitWaits`·`RateLimitWaited`) | `TestRateLimitDoesNotConsumeAStabilisationAttempt` |
| **실패한 복구에도** 대기 실측이 남는다 (대입이 오류 검사 앞) | `TestRateLimitWaitBudgetExhaustionFailsClosed` · `TestPermanentlyRefusingBrokerExhaustsTheBudget` |
| 429가 아닌 오류는 오늘처럼 즉시 미완료 + 게이트 유지 | `TestNonRateLimitedCollectStillFailsImmediately` |
| 기존 절차 무회귀 | `TestRecoveryReleasesTheLatchOnlyWhenItCompletes` · `TestCrashMidDispatchBecomesInDoubtAndIsResolved` · `TestRecordedButNeverSentIsClosedSafely` · `TestRecoveryThatEndsInDisagreementBlocksEntriesForADifferentReason` |

## 뮤테이션 정산

이 함수의 편집을 직접 반증한 것은 (b)다: `ratelimit.go`의 예산 검사를 지우면
`TestRateLimitWaitBudgetExhaustionFailsClosed`가 `err = <nil>`로 죽는다 — 즉 B10을 지나
성공 경로로 빠지는 것이 관측된다. 전체 표는
`internal-reconcile--recovery.stablesnapshot/branch-test-map.md`에 있다.

## 산출물 근거

- 분기·이탈 열거: `ast.json` (branches 12, returns 8) — `go run ./tools/logic-map`
- 커버리지: `go test ./internal/reconcile/ -count=1 -coverprofile` exit 0 · **147건 통과** · 86.6% (§1.9)
- 편집 전 기준선: 같은 명령, 130건 통과 (HEAD `03139000`), 분기·이탈 수 동일
