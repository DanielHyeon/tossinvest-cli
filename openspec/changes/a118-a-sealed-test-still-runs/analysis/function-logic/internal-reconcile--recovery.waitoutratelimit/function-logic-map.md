# Function Logic Map: `Recovery.waitOutRateLimit`

- Source: `internal/reconcile/ratelimit.go` (lines 82-113)
- AST evidence: `ast.json` (3 branches)
- Risk scan: `risk-pattern-report.md`

이 change 는 이 함수를 **편집하지 않는다.** 묶음을 만든 이유는 proposal 과 design 이 이 함수의
예산 검사를 "실제로 멈추는 것"으로 지목하기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.stab.RateLimitBackoff` | ≥ `DefaultRateLimitBackoff`(15s, floor) | `ratelimit.go:51` | 조회 전용 서베이보다 성급해지지 않게 하는 바닥 |
| `r.stab.MaxRateLimitWait` | >0, 기본 5분 | `ratelimit.go:59` | 초과 시 B1 |
| `progress.RateLimitWaited` | 누적 대기 | `snapshotProgress` | 예산 판정의 좌변 |
| `progress.RateLimitWaits` | 대기 횟수 | `snapshotProgress` | B2 가 사유 문구를 가르는 데 쓴다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 :88 | `RateLimitWaited + backoff > MaxRateLimitWait` | — | `ErrRecoveryIncomplete` | `TestRateLimitBudgetStopsExactlyAtTheBoundary` |
| B2 :91 | 그중 `RateLimitWaits == 0` | — | "예산이 backoff 하나도 못 덮는다"는 **다른** 문구 | `TestBudgetTooSmallForOneBackoffSaysSo` |
| B3 :107 | `clk.Sleep` 이 오류(취소) | — | `%w` 두 겹으로 원인 보존 | `TestRateLimitBackoffStopsOnContextCancel` |

**이 change 가 근거로 쓰는 사실.** B1 이 유일한 상한이다. 이 함수가 정상 반환하면
`RateLimitWaits++` 와 `RateLimitWaited += backoff` 가 일어나고(:110-111), 그래서 예산은
`MaxRateLimitWait / RateLimitBackoff` 번의 대기 뒤 소진된다. B2 가 존재하기 때문에
"예산을 backoff 보다 작게 준다"는 대안 시험 방식은 **다른 계약**을 시험하게 된다 —
그것이 design.md §3 이 그 대안을 거부한 이유다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `clk.Sleep` | 백오프 대기 | 종료 신호를 즉시 통과(안전 불변식 4) | AST calls :107 |
| `fmt.Errorf` | 사유 문구 구성 | `%w` 로 원인 보존 | AST calls |

## State mutations and fallbacks

- `progress` 두 필드만 누적한다. 브로커 호출도 뮤테이션도 없다.
- 대기 중에도 진입 게이트는 닫힌 채다(`recovery.go:177` 에서 이미 래치).

## Safety conclusion

- Safe edit boundary: 이 change 는 편집하지 않는다. 뮤테이션 M1·M3 으로 잠시 뒤집고 원복했다.
- High-risk impact: yes — 부팅 복구를 지연시킬 수 있는 유일한 대기다. 이 change 의 편집은 0줄이다.
