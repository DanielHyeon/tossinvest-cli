# Function Logic Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go` (341-407)
- AST evidence: `ast.json` — **branches 12, returns 3, calls 16**
- Risk scan: `risk-pattern-report.md`

> **12판 재기준화.** base `ec29dc72` → `285c7619`. a096 2라운드와 a097이 같은 파일을
> 편집해 이 함수가 **103줄 아래로 밀렸고 분기가 10 → 12로 늘었다.** 늘어난 둘은
> a096이 만든 **publish 성공 + 기록 실패** 경로의 로그·게이트 분기다.
>
> **12라운드 보이스 A가 이 산문이 구판이라고 지적했다(T).** BTM은 12판이 고쳤는데
> 이 파일은 안 고쳤다 — `check_analysis`는 산문 drift를 보지 않는다. 아래는 재열거다.
- Risk scan: `risk-pattern-report.md`

**34초는 여기서 만들어진다.** 그리고 그동안 `n.mu`가 잡혀 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | outbox 행 id | `notifyCritical:177`의 `EnqueueAlert` 반환값 | **기존 행 재사용 가능** (`outbox.go:131`) → `MarkAlertDelivered`가 `state=PENDING` 술어에 걸려 실패 |
| `n.Attempts` | 0이면 기본값 | 조립 | B1 `:343`이 `DefaultCriticalAttempts` = **3**으로 대체. 프로덕션 `newNotifier`는 안 채운다 |
| `n.RetryDelay` | 0이면 기본값 | 조립 | `wait`가 `DefaultRetryDelay` = **2s**로 대체 |
| `n.Publisher` | nil 허용 | 조립 | B3 `:350`이 즉시 `break` — **대기 0초** |
| `n.Gate` | nil 허용 | 조립 | B7 `:378`·B12 `:403`이 nil을 흡수 — **게이트를 잠그는 자리가 둘이다** |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:343` | `attempts <= 0` | `attempts = DefaultCriticalAttempts`(3) | — |
| B2 `:349` | `for attempt := 1; attempt <= attempts; attempt++` | 최대 **3회전** | — |
| B3 `:350` | `n.Publisher == nil` | `lastErr = "no notification publisher is configured"` → `break` | — |
| B4 `:355` | `Publish` 성공 | (B5~B7) | — |
| B5 `:357` | `MarkAlertDelivered` 성공 | outbox 행 DELIVERED | `return true` `:358` |
| B6 `:373` | 기록 실패 **and** `Log != nil` | `Log.Error(EventAlertUndelivered, markErr, event, alert_id)` | — |
| B7 `:378` | 기록 실패 **and** `Gate != nil` | **`Gate.Block(ReasonAlertUndelivered, detail)`** | `return false` `:381` |
| B8 `:384` | `MarkAlertAttemptFailed` 오류 **and** `Log != nil` | `Log.Error(…)` | — |
| B9 `:387` | `attempt < attempts` | (B10) | — |
| B10 `:388` | `!n.wait(ctx)` (ctx 종료) | `break` | — |
| B11 `:398` | `Log != nil` | `Log.Error(EventAlertUndelivered, lastErr, event, alert_id)` | — |
| B12 `:403` | `n.Gate != nil` | **`Gate.Block(ReasonAlertUndelivered, detail)`** | — |
| — | 예산 소진 | — | `return false` `:406` |

**이탈 3개**: 성공(`:358` true) / **기록 실패**(`:381` false, a096 신설) / 소진(`:406` false).
소진은 `notifyCritical` B4를 켠다.

**a096이 만든 것이 a092에 주는 것 하나**: 게이트를 잠그는 자리가 **둘**이 됐다(B7·B12).
B7은 **예산 밖이다** — publish가 성공한 뒤에 일어나므로 재시도 예산이 그것을 유계로
만들지 않는다. a092의 예산은 B2 회전만 덮는다.

## Calls and live bindings

| Callee | Why called | **Error/timeout/retry contract** | Evidence |
|---|---|---|---|
| — (**이 함수는 잠그지 않는다**) | 직렬화는 호출자가 한다 | **`claimAndDeliver`(`:238`)가 `:241`/`:242`에서 잡고 `:276`에서 이 함수를 부른다.** 그러므로 보유는 여전히 이 함수 전체에 걸린다 — **최대 34s**. 주석 `:338`: "goroutines publishing the same backlog would double-send" | 이 산출물의 `ast.json`: `defers: null`, `mu` 호출 0 |
| `notificationFor` `:346` | 메시지 조립 | 순수 | AST calls |
| **`n.Publisher.Publish` `:354`** | **원격 발송** | 프로덕션 `*obs.Ntfy` → **1회 10s**(`ntfy.go:96`·`:122`). 최대 **3회** | `internal-obs--ntfy.publish/ast.json` |
| `n.Journal.MarkAlertDelivered` `:356` | 성공 표시 | 로컬. **`WHERE id = ? AND state = 'PENDING'`**(`outbox.go:159`) — 이미 DELIVERED면 실패 | AST calls |
| `n.Journal.MarkAlertAttemptFailed` `:384` | 시도 실패 기록 | 로컬. 같은 술어(`outbox.go:174`) | AST calls |
| `n.wait` `:388` | 시도 간 대기 | **2s × 2회**. ctx 종료 시 false → `break` | AST calls + `notifier.go:290-300` |
| `n.Gate.Block` `:404` (소진) · `:379` (발송 후 기록 실패) | 신규 진입 차단 | 인메모리 래치. `detail`은 **여기에만** 간다 — 로그에는 안 간다(B9는 `lastErr`만 쓴다) | AST calls |

### 최악 예산의 산술

```
3 × Publish(10s)  +  2 × wait(2s)  =  34s
   B2 회전 3회        B7이 참인 2회
```

**단, Publisher가 nil이면 B3이 첫 회전에서 break하므로 0초다.** 2026-08-04 이전
outbox 행 1~9가 전부 `attempts=0`인 것이 그 경로의 실측 증거다(`journal.db`).
비용을 만드는 것은 **응답하지 않는 transport**이지 없는 transport가 아니다.

## State mutations and fallbacks

- `n.mu`는 **이 함수가 아니라 호출자 `claimAndDeliver`**가 잡는다(획득 `:241`, 해제 defer `:242`). 이 함수는 그 잠금 **안에서** 불린다(`:276`) — 보유 시간은 함수 전체다.
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
