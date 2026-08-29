# Function Logic Map: `Recovery.stableSnapshot`

- Source: `internal/reconcile/recovery.go` (lines 371-405)
- AST evidence: `ast.json` (7 branches)
- Risk scan: `risk-pattern-report.md`

이 change 는 이 함수를 **편집하지 않는다.** 묶음을 만든 이유는 proposal 과 design 이 이 함수의
분기를 근거로 "21은 정책이 아니라 산술이다"라고 쓰기 때문이다. 근거로 쓰는 문서보다 열거가
먼저 나와야 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.stab.MaxAttempts` | ≥1 | `Stabilisation`(`recovery.go:73`) | 초과 시 B6 로 루프 이탈 |
| `r.stab.Required` | ≥1 | `Stabilisation`(`recovery.go:70`) | 미달 시 마지막 return 으로 fail-closed |
| `r.stab.Interval` | ≥0 | `Stabilisation`(`recovery.go:67`) | B7 의 `clk.Sleep` 간격 |
| `attempt` (루프 변수) | 1..MaxAttempts | 이 함수 | **rate limit 팔은 소모하지 않는다** |
| `progress` | 누적만 | `snapshotProgress` | Report 로 흘러 나간다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 :376 | `attempt <= r.stab.MaxAttempts` 루프 | — | — | `TestRateLimitDoesNotConsumeAStabilisationAttempt` |
| B2 :378 | `Collect` 오류 | — | — | `TestRateLimitedCollectDoesNotEndRecovery` |
| B3 :379 | rate limit 이 **아닌** 오류 | — | `ErrRecoveryIncomplete` 즉시 | `TestNonRateLimitedCollectStillFailsImmediately` |
| B4 :382 | `waitOutRateLimit` 이 오류 | — | 그 오류를 그대로 | `TestRateLimitWaitBudgetExhaustionFailsClosed` |
| B5 :390 | `stabiliser.Offer` 가 Stable | — | 스냅샷 반환(유일한 성공 출구) | `TestRecoveryReleasesTheLatchOnlyWhenItCompletes` |
| B6 :393 | `attempt > MaxAttempts` | — | 루프 이탈 → fail-closed | `TestRecoveryFailsClosedWhenTheAccountWillNotSettle` |
| B7 :396 | `clk.Sleep` 이 오류(취소) | — | `ErrRecoveryIncomplete` | `TestRateLimitBackoffStopsOnContextCancel` |

**이 change 가 근거로 쓰는 사실.** B4 뒤의 `continue` 는 `attempt++` 를 **하지 않는다**
(`recovery.go:385` 의 "Deliberately no attempt++: the account was never read."). 따라서 429 가
반복되는 경로에서 B1 의 루프 상한은 아무것도 멈추지 않고, 멈추는 것은 `waitOutRateLimit` 의
예산 검사뿐이다. 읽기 횟수가 `MaxRateLimitWait / RateLimitBackoff + 1` 인 이유가 이것이다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.opts.Collector.Collect` | 계좌 스냅샷 읽기(조회 3종) | 429 는 B3 이 아니라 B4 로 | AST calls :377 |
| `r.waitOutRateLimit` | 유한한 예산 안에서 백오프 | 예산 초과 시 오류 | AST calls :382 |
| `stabiliser.Offer` | 연속 일치 판정 | — | AST calls :390 |
| `clk.Sleep` | 안정화 간격 | 취소를 오류로 통과 | AST calls :396 |

## State mutations and fallbacks

- 브로커 뮤테이션 호출 없음. 이 루프가 부르는 것은 조회뿐이다.
- 부분 스냅샷을 반환하지 않는다 — 성공 출구는 B5 하나이고 나머지는 빈 스냅샷과 오류다.

## Safety conclusion

- Safe edit boundary: 이 change 는 편집하지 않는다. 뮤테이션 M2 로 잠시 뒤집고 원복했다.
- High-risk impact: yes — 복구는 진입 게이트를 래치한 채 돈다. 다만 이 change 의 편집은 0줄이다.
