# Function Logic Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go` (238-287)
- AST evidence: `ast.json` — branches 10, returns 2, calls 15, assignments 11,
  **defers 1**(`n.mu.Unlock` `:242`), **go_statements 0**
- Risk scan: `risk-pattern-report.md`

**34초는 여기서 만들어진다.** 그리고 그동안 `n.mu`가 잡혀 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | outbox 행 id | `notifyCritical:177`의 `EnqueueAlert` 반환값 | **기존 행 재사용 가능** (`outbox.go:131`) → `MarkAlertDelivered`가 `state=PENDING` 술어에 걸려 실패 |
| `n.Attempts` | 0이면 기본값 | 조립 | B1 `:245`가 `DefaultCriticalAttempts` = **3**(`:45`)로 대체. 프로덕션 `newNotifier`는 안 채운다 |
| `n.RetryDelay` | 0이면 기본값 | 조립 | `wait:290-300`이 `DefaultRetryDelay` = **2s**(`:48`)로 대체 |
| `n.Publisher` | nil 허용 | 조립 | B3 `:252`가 즉시 `break` — **대기 0초** |
| `n.Gate` | nil 허용 | 조립 | B10 `:283`이 nil을 흡수 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:245` | `attempts <= 0` | `attempts = DefaultCriticalAttempts`(3) | — |
| B2 `:251` | `for attempt := 1; attempt <= attempts; attempt++` | 최대 **3회전** | — |
| B3 `:252` | `n.Publisher == nil` | `lastErr = errors.New("no notification publisher is configured")` → `break` | — |
| B4 `:257` | `Publish` 성공 | (B5) | `return true` `:261` |
| B5 `:258` | `MarkAlertDelivered` 오류 **and** `Log != nil` | `Log.Error(EventAlertUndelivered, markErr)` `:259` | — |
| B6 `:264` | `MarkAlertAttemptFailed` 오류 **and** `Log != nil` | `Log.Error(…)` `:265` | — |
| B7 `:267` | `attempt < attempts` | (B8) | — |
| B8 `:268` | `!n.wait(ctx)` (ctx 종료) | `break` | — |
| B9 `:278` | `Log != nil` | `Log.Error(EventAlertUndelivered, lastErr, event, alert_id)` `:279-281` | — |
| B10 `:283` | `n.Gate != nil` | **`Gate.Block(ReasonAlertUndelivered, detail)`** `:284` | — |
| — | 예산 소진 | — | `return false` `:286` |

**이탈 2개**: 성공(`:261` true) / 소진(`:286` false). 소진은 `notifyCritical` B4를 켠다.

## Calls and live bindings

| Callee | Why called | **Error/timeout/retry contract** | Evidence |
|---|---|---|---|
| `n.mu.Lock` `:241` / `n.mu.Unlock` `:242` (**defer**) | 배달 루프 직렬화 | **보유 시간 = 이 함수 전체 = 최대 34s.** 주석 `:239-240`: "two goroutines publishing the same backlog would double-send" | AST calls + defers 1 |
| `notificationFor` `:248` | 메시지 조립 | 순수 | AST calls |
| **`n.Publisher.Publish` `:256`** | **원격 발송** | 프로덕션 `*obs.Ntfy` → **1회 10s**(`ntfy.go:96`·`:122`). 최대 **3회** | `internal-obs--ntfy.publish/ast.json` |
| `n.Journal.MarkAlertDelivered` `:258` | 성공 표시 | 로컬. **`WHERE id = ? AND state = 'PENDING'`**(`outbox.go:159`) — 이미 DELIVERED면 실패 | AST calls |
| `n.Journal.MarkAlertAttemptFailed` `:264` | 시도 실패 기록 | 로컬. 같은 술어(`outbox.go:174`) | AST calls |
| `n.wait` `:268` | 시도 간 대기 | **2s × 2회**. ctx 종료 시 false → `break` | AST calls + `notifier.go:290-300` |
| `n.Gate.Block` `:284` | 신규 진입 차단 | 인메모리 래치. `detail`은 **여기에만** 간다 — 로그에는 안 간다(B9는 `lastErr`만 쓴다) | AST calls |

### 최악 예산의 산술

```
3 × Publish(10s)  +  2 × wait(2s)  =  34s
   B2 회전 3회        B7이 참인 2회
```

**단, Publisher가 nil이면 B3이 첫 회전에서 break하므로 0초다.** 2026-08-04 이전
outbox 행 1~9가 전부 `attempts=0`인 것이 그 경로의 실측 증거다(`journal.db`).
비용을 만드는 것은 **응답하지 않는 transport**이지 없는 transport가 아니다.

## State mutations and fallbacks

- `n.mu` 보유(획득 `:241`, 해제 defer `:242`) — **함수 전체**.
- outbox 행 상태: PENDING → DELIVERED(`:258`) 또는 attempts 증가(`:264`).
- `Gate` 래치(`:284`) — 프로세스 내. 재기동하면 사라지므로 `escalate`가
  `notifyCritical:187`에서 durable 짝을 만든다.
- fallback: `lastErr`가 비어 있을 수 없다 — B3이 nil publisher에도 메시지를 넣는다.
- **goroutine 없음**(`go_statements: 0`).

## Safety conclusion

- **Safe edit boundary**: 이 함수의 **본문은 바꾸지 않아도 된다.** 바꿔야 하는 것은
  **누가 이 함수를 부르는가**다. 지금은 `notifyCritical:182`(호출자 goroutine)뿐이다.
- **High-risk impact**: **yes** — exit 관측 루프가 여기서 최대 34초, `n.mu`를 잡은 채
  머문다. 같은 뮤텍스를 대사 루프도 다툰다.
- **비동기화가 보존해야 할 성질 셋**:
  1. `n.mu` 직렬화 — `TestNotifierIsConcurrencySafe`(`obs_test.go:620`)가 고정한다.
  2. B10의 게이트 래치가 **여전히 일어난다**(시점만 늦어진다).
  3. B4의 `return true`가 outbox 행을 DELIVERED로 만든다.
- **관측되지 않는 것 하나**: B9의 `detail`(`:276-277`)은 `Gate.Block`에만 가고
  로그에는 없다. 그래서 "N회 시도 후 실패"라는 문자열은 `engine.log`에서 검색해도
  나오지 않는다 — a089 1라운드 리뷰가 이것으로 잘못된 결론을 냈다. 기록해 둔다.
