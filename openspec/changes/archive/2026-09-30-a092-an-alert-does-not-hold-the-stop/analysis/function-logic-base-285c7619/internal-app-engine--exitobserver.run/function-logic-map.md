# Function Logic Map: `ExitObserver.Run`

- Source: `internal/app/engine/exitloop.go` (353-363)
- AST evidence: `ast.json` — branches 3, returns 2, calls 5, assignments 2,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

이 함수가 a092의 예산이 걸리는 자리다. **한 사이클의 체류가 주기에 더해진다**는 사실이
여기 AST에 그대로 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 감독 루프의 컨텍스트 | `runtime.go:277-283`이 goroutine으로 띄운다 | B2 `:355`가 취소를 보면 즉시 반환 |
| `o.Interval()` | > 0 | `:326-331` — `o.opts.Interval`이 0 이하면 `DefaultExitObservationInterval` **5s**(`:97`) | 없음 |
| `o.clk` | `clock.Clock` | 프로덕션은 `clock.System()` | 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:354` | `for {}` — 무한 루프 | — | — |
| B2 `:355` | `ctx.Err() != nil` | 없음 | `return err` `:356` |
| B3 `:359` | `o.clk.Sleep(ctx, o.Interval())`가 오류 | 없음 | `return err` `:360` |

**이탈은 둘 다 컨텍스트 종료다.** 사이클 실패로는 루프가 멈추지 않는다 —
`ObserveOnce`의 결과는 `reportCycle`로만 흘러간다(`:358`).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ctx.Err` `:355` | 종료 확인 | 즉시 | AST calls |
| `o.reportCycle` `:358` | 실패 사이클 1줄 | 네트워크 없음 | AST calls |
| `o.ObserveOnce` `:358` | **한 사이클 전체** | **무기한** — 이 함수는 기한을 씌우지 않는다(AST calls 5에 `context.*` 없음, defers 0) | AST calls |
| `o.clk.Sleep` `:359` | 주기 대기 | ctx 취소로만 중단 | AST calls |
| `o.Interval` `:359` | 주기 조회 | — | AST calls |

### 사이클 체류는 주기에 **더해진다**

AST의 호출 순서가 `:358` `ObserveOnce` → `:359` `Sleep(Interval)`이다. ticker가 아니라
**작업 후 고정 수면**이므로 실제 관측 주기는

```text
period = ObserveOnce가 걸린 시간 + Interval(5s)
```

| 사이클 상황 | ObserveOnce 체류 | 실제 주기 | 의도 대비 |
|---|---|---|---|
| 알림 없음 | ~0 | 5s | 1× |
| critical 알림 1건, transport 죽음 | 34s | 39s | **7.8×** |
| critical 알림 N건 (아래) | 34N s | 34N+5 s | — |
| 두절 사이클, **NORMAL 계정** (P4+P1a) | 68s | 73s | 14.6× |
| 두절 사이클, **오늘의 계정** | 34s | 39s | 7.8× |

**두절 사이클이 오늘 68초가 아닌 이유**: P1a(`checkOutage:796` → Announcer)는
계정이 이미 `ENTRY_BLOCKED`이므로 `direction == 0`에서 announce 전에 반환한다
(`operating_mode.go:409-415`). 계정은 2026-07-31부터 그 상태이고, 로그 전체에서
`AnnounceOperatingMode`가 쓴 줄은 **1개**다(`analysis/delivery-latency.md` §0).

**N이 1이 아닌 이유**: `alertProposalRefused`(`exitloop.go:1548`)에는 래치가 없고
`judge`가 포지션마다 부를 수 있다. 그러므로 **이 함수의 체류는 알림당 예산으로는
유계가 되지 않는다** — a092가 정하는 것은 알림 하나의 예산이다.

## State mutations and fallbacks

- **이 함수는 아무 상태도 변경하지 않는다.** AST assignments 2는 둘 다 `err :=`다.
- fallback 없음. 사이클이 실패해도 다음 사이클이 온다.
- goroutine 없음(AST `go_statements: 0`) — `ObserveOnce`는 이 goroutine에서 동기로 돈다.

## Safety conclusion

- **Safe edit boundary**: 이 함수 자체는 a092에서 **편집하지 않는다.** 여기 있는 것은
  예산의 *기준*(5s)이지 예산이 걸리는 코드가 아니다.
- **High-risk impact**: **yes** — 이 루프가 멈추거나 느려지면 손절이 늦는다.
  `.claude/CLAUDE.md` §0.4가 걸리는 자리다.
- a092의 목표를 이 함수의 용어로 쓰면: **알림 하나의 동기 체류가 `Interval()`을
  넘지 않게 한다** — 그래야 실제 주기가 의도의 2배를 넘지 않는다.
