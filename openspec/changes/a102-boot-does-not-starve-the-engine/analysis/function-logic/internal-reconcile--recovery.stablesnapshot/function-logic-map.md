# Function Logic Map: `Recovery.stableSnapshot`

- Source: `internal/reconcile/recovery.go` (371-405)
- AST evidence: `ast.json` — AST 기준 branches **7** / returns 5 / calls 9 / defers 0 / go_statements 0
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `f32ab95497925c87fbd750dcd60772f75f30a39190dc8be03a1a7c8704622dc5`

**a102 §1(겹1)이 편집한 함수다.**

## 세 판 — 편집 전, GREEN 뒤, A1 리뷰 뒤

1판은 **편집 전**(HEAD `03139000`)에 만들어졌고 그것이 아래 「Branches」의 근거였다.
2판은 GREEN(`1c76a580`) 뒤에 다시 뽑았다. 3판은 A1 적대 리뷰 반영(§1.9) 뒤이며,
**위 헤더의 좌표·수·해시는 전부 3판이다.**

| | 1판 (편집 전) | 2판 (`1c76a580`) | **3판 (§1.9, 이 문서)** |
|---|---|---|---|
| 위치 | `recovery.go:333-359` | `:366-400` | **`:371-405`** |
| 분기 | 5 | 7 | **7** |
| 이탈 | 4 | 5 | **5** |
| 호출 | 7 | 9 | **9** |
| source SHA-256 | `80ee029c…` | `e0d5690f…` | `f32ab954…` |

늘어난 분기 둘은 429 분류(`:379`)와 백오프 실패(`:382`)다. 늘어난 호출 둘은
`errors.Is`와 `r.waitOutRateLimit`이다.

**3판에서 이 함수 자체는 안 바뀌었다** — 분기·이탈·호출이 그대로이고 좌표만 다섯 줄
밀렸다(위쪽 `withDefaults`의 주석 때문). §1.9가 바꾼 것은 `withDefaults`의 하한 규칙과
`waitOutRateLimit`의 두 오류 문장이며, 그 둘은 B4가 부르는 쪽에 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.stab.MaxAttempts` | ≥1 (0이면 5) | `Options.Stabilise` → `New`(`:176`) | B1의 상한. 소진하면 `:402`의 미완료 실패 |
| `r.stab.Required` | ≥1 (0이면 2) | 같음 | `Stabiliser.Required`. `:403` 메시지에 인용된다 |
| `r.stab.Interval` | >0 (0이면 2s) | 같음 | B7의 `clk.Sleep` 인자이자 `Stabiliser.MinInterval` |
| `r.stab.RateLimitBackoff` | **≥15s** — 더 짧은 값은 `withDefaults`가 올린다(§1.9 F6) | 같음 | **신규.** B4가 소비 → `waitOutRateLimit` |
| `r.stab.MaxRateLimitWait` | >0 (0이면 5m) | 같음 | **신규.** 예산 소진은 B5의 실패로 나온다 |
| `r.opts.Collector` | non-nil (`New`가 강제) | `New`(`:172`) | B2 → B3에서 **429와 그 밖**이 갈린다 |
| `clk` | non-nil | `r.clock()`(`:420`) | 없음 |
| `ctx` | 취소 가능 | `Run`의 인자 | B5·B7이 `ctx.Err()`를 올린다 |

> **불변식 1 — 세는 것은 읽기다.** `attempt++`(`:388`)와 `progress.Taken++`(`:389`)는
> 둘 다 **성공한 `Collect` 뒤에만** 실행된다. 429는 `continue`(`:386`)로 되돌아가며
> 어느 쪽도 늘리지 않는다. 읽기가 없었으므로 안정화 판정의 입력이 아니다.
>
> **불변식 2 — 기다림은 유한하다.** 429의 총 대기는 `MaxRateLimitWait`이 덮는다
> (`ratelimit.go:88`). 예산이 남지 않으면 **기다리지 않고** 실패한다. 경계는 `>`이므로
> 예산에 **정확히 닿는** 대기는 허용되고 넘기는 대기는 아니다 — `TestRateLimitBudgetStopsExactlyAtTheBoundary`가
> 그 한 글자를 지킨다(§1.9 F4). 예산이 백오프 한 번보다 작으면 한 번도 기다리지 않고,
> 오류 문장이 그렇게 말한다(`ratelimit.go:93`, §1.9 F8).
>
> **불변식 3 — 실패는 언제나 fail-closed다.** 이 함수의 모든 오류 경로는 `Run`이 그대로
> 올리고, 게이트는 `New`가 잠근 채 남는다. 429 대기 중에도 게이트는 닫혀 있다.

## Branches and early returns

`ast.json`의 열거를 그대로 옮긴다.

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:376` | `for attempt := 1; attempt <= r.stab.MaxAttempts;` (**증분 없음**) | — | 소진 시 `:402` |
| B2 | `:378` | `err != nil` (`Collect` 실패) | 없음 | — (B3로) |
| B3 | `:379` | `!errors.Is(err, official.ErrRateLimited)` | 없음 | `:380` `%w: %v` — **오늘과 동일한 즉시 실패** |
| B4 | `:382` | `r.waitOutRateLimit(...) != nil` | `progress.RateLimitWaits/Waited` | `:383` 예산 소진 또는 종료 신호 |
| B5 | `:390` | `stabiliser.Offer(snap).Stable` | stabiliser의 streak/digest | `:391` `snap, progress, nil` |
| B6 | `:393` | `attempt > r.stab.MaxAttempts` | 없음 | `break` → `:402` |
| B7 | `:396` | `clk.Sleep(ctx, r.stab.Interval) != nil` | 없음 | `:397` `%w: %v` |

Returns: `:380` · `:383` · `:391` · `:397` · `:402` (AST 5개).

> **B3이 이 change의 전부다.** 편집 전에는 B2가 곧 실패였고, 브로커가 "지금은 안 된다"고
> 한 것과 "계좌를 못 읽는다"고 한 것이 한 팔로 합쳐져 있었다. 2026-08-13 02:03:30.545Z에
> 부팅 429가 그 팔로 들어와 복구를 끝냈다. 이제 그 갈림이 B3에 있다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.clock` | 시계 seam | 실패 없음 (nil → `clock.System()`) | `:372` |
| `r.opts.Collector.Collect` | 계좌 1회 읽기 | 오류는 B2 → B3에서 분류된다 | `:377` (+ `internal-reconcile--collector.collect`) |
| `errors.Is` | **429 판별** — 문자열 매칭 금지 | 판정만, 부작용 없음 | `:379` (sentinel `internal/official/errors.go:15`) |
| `r.waitOutRateLimit` | 백오프 1회 + 예산 검사 | 예산 소진·`ctx` 취소가 오류 | `:382` (`ratelimit.go`) |
| `stabiliser.Offer` | 안정화 판정 | 오류 없음 | `:390` |
| `clk.Sleep` | 안정화 간격 대기 | `ctx` 취소가 즉시 통과 | `:396` |
| `fmt.Errorf` ×3 | `ErrRecoveryIncomplete` wrap | 원인은 `%v` (기존 형태 유지) | `:380` · `:397` · `:402` |

**호출자는 하나다**: `Recovery.Run`의 3단계(`recovery.go:298`, 번들
`internal-reconcile--recovery.run`). 반환한 `snapshotProgress`가 `Report`의
`SnapshotsTaken`·`RateLimitWaits`·`RateLimitWaited`로 옮겨진다(`:299-301`).

## State mutations and fallbacks

- `Recovery`의 필드는 여전히 쓰지 않는다. 상태는 지역 `stabiliser`와 `progress` 둘이고
  호출마다 새로 시작한다.
- 게이트를 직접 만지지 않는다. 잠금은 `New`(`:177`), 해제는 `Run`의 5단계(`:322`).
- **새 로거 배관은 없다.** 429의 관측은 `Report`의 두 필드로만 나간다 — 기존 report 소비
  지점에 그대로 실린다(design D3).
- fallback은 하나뿐이다: 429일 때 다시 읽는다. 그 밖의 모든 오류는 fallback이 없다.

## Safety conclusion

- Safe edit boundary: B1의 증분 위치와 B3·B4의 신설. 안정화 판정(B5)·상한(B6)·간격
  대기(B7)의 의미는 편집 전과 **같다**.
- High-risk impact: **yes** — 재시작 복구. 이 함수가 성공하지 않으면 신규 진입 게이트가
  열리지 않는다.
- 보수 방향: 편집은 **실패를 더 늦게** 만들 뿐 게이트를 더 일찍 열지 않는다. 429 대기
  중에도, 예산 소진 뒤에도 게이트는 닫힌 채다.
- 불변식 4(종료 지연 금지): 새 대기도 기존 `clk.Sleep(ctx, …)` seam을 쓴다. 뮤테이션 (d)
  (`ctx` 대신 `context.Background()`)와 (N2)(`Sleep`의 오류를 삼킨다)로 반증했고
  `TestRateLimitBackoffStopsOnContextCancel`이 둘 다에서 죽는다.
- **§1.9에서 고친 것**: 취소의 정체가 `%w`로 살아 나오고(F3), 종료 뒤 브로커를 한 번도
  더 부르지 않는다는 것을 **호출 수로** 고정했다(F2). A1이 "오류를 삼켜도 전 스위트가
  통과한다"를 실증했고, 그 상태의 운영 의미는 SIGTERM 순간 스로틀 중인 브로커에 대한
  20회 즉시 재조회였다 — 지금은 그 뮤테이션이 `reads = 21, want 1`로 죽는다.
