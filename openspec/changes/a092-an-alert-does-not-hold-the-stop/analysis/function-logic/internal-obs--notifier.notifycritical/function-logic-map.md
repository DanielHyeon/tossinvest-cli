# Function Logic Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go` (153-190)
- AST evidence: `ast.json` — branches 4, returns 3, calls 11, assignments 2,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**이 함수가 이 change의 절단면이다.** 기록(`:177`)과 발송(`:182`)이 한 함수 안에서
연달아 일어나고, 앞의 것만 반환 전에 끝나야 한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | nil 허용 | 조립 (`exitwiring.go:76`은 항상 채운다) | nil이면 B1이 최선노력으로 강등 + 경고 |
| `n.Log` | nil 허용 | 조립 | B2가 nil을 흡수 |
| `e` | critical 등급이 확정된 이벤트 | `Notify:115` | — |
| `n.eventKey(e)` | 비어 있지 않은 문자열 | `eventKey:385-396` — `e.Key` 우선, 없으면 종류+attempt/order/symbol | 같은 조건이 같은 키를 만든다(중복 억제의 근거) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | durable? |
|---|---|---|---|---|
| B1 `:154` | `n.Journal == nil` | (B2) + `publishBestEffort(ctx, e, Critical)` `:163` | `return nil` `:164` | ❌ **강등** |
| B2 `:158` | B1 안 · `n.Log != nil` | `Log.Warn(EventAlertUndelivered, …)` `:159-162` | — | — |
| B3 `:178` | `EnqueueAlert`가 오류 | 없음 | `return fmt.Errorf(…)` `:179` | ❌ 실패 |
| B4 `:182` | `!n.deliver(ctx, id, e)` | `n.escalate(ctx, e)` `:187` | — | ✅ 행은 PENDING으로 남음 |
| — | 정상 | — | `return nil` `:189` | ✅ |

**이탈 3개 중 outbox 행이 남는 것은 B4 경로와 정상 경로 둘뿐이다.**

## Calls and live bindings

| Callee | Why called | **Error/timeout/retry contract** | Evidence |
|---|---|---|---|
| `n.Log.Warn` `:159` | journal 미배선 경고 | 네트워크 없음 | AST calls |
| `n.publishBestEffort` `:163` | journal 없을 때의 강등 발송 | publish 1회, **10s** | AST calls + publishBestEffort FLM |
| `n.eventKey` `:168` | 중복 키 | 순수 문자열 조립 | AST calls |
| `encodeFields` `:173` | payload JSON | 실패 시 빈 문자열(`:410-419`) | AST calls |
| **`n.Journal.EnqueueAlert` `:177`** | **durable 기록** | **로컬 SQLite 트랜잭션.** 네트워크 없음 — 이 호출은 밀리초 단위이고 비동기화 대상이 **아니다** | AST calls + `journal/outbox.go:111-151` |
| `n.deliver` `:182` | **발송** | **최대 34s, `n.mu` 보유.** 이 호출이 §0.3이 걸리는 자리다 | `internal-obs--notifier.deliver/ast.json` |
| `n.escalate` `:187` | ENTRY_BLOCKED 강화 | 로컬 journal 쓰기. 주석 `:181-186`이 "deliver 밖"인 이유를 쓴다 — deliver가 `n.mu`를 잡고 있어 재진입하면 교착 | AST calls |

### 기록과 발송의 비대칭이 절단선이다

| | 대상 | 비용 | 반환 전에 끝나야 하는가 |
|---|---|---|---|
| `EnqueueAlert` `:177` | 로컬 SQLite | 밀리초 | **예** — 주석 `:175-176`: "a record that only exists in memory is a record that does not survive the crash it is warning about" |
| `deliver` `:182` | 원격 HTTPS | 최대 34s | **아니오** — 이 함수의 반환값은 발송 성공을 말하지 않는다(주석 `:103-106`) |

## State mutations and fallbacks

- `EnqueueAlert`가 `alert_outbox`에 행을 만들거나 **기존 행의 id를 돌려준다**
  (`outbox.go:131` `case err == nil: return existing, tx.Commit()` — 상태 불문).
  이 재사용이 실측 로그의 `journal: no such alert: N (or it is no longer pending)`
  6줄을 만든다(2026-08-05, `engine.log`) — a089의 범위이고 이 change는 건드리지 않는다.
- fallback: B1이 journal 부재를 최선노력으로 강등하되 **로그로 알린다**(조용한 강등 금지).
- **goroutine 없음**(`go_statements: 0`).

## Safety conclusion

- **Safe edit boundary**: `:177`(기록)과 `:182`(발송) 사이. 그 사이를 자르면
  durability는 보존되고 체류만 사라진다.
- **High-risk impact**: **yes** — exit 관측 루프가 `:182`에서 최대 34초 머문다.
- **자르면 깨지는 것 셋**(이 change가 보존해야 할 계약):
  1. **B4의 escalate 순서** — `deliver`가 실패해야 escalate한다. 비동기화하면 escalate도
     비동기가 된다. 게이트 래치(`deliver:284`)도 마찬가지.
  2. **재진입 금지** — 주석 `:181-186`. 발송 goroutine이 `Notify`를 다시 부르면
     `n.mu`에서 교착한다.
  3. **B1의 강등 경로** — journal이 없으면 durable 큐 자체가 없으므로 비동기화의
     전제가 사라진다. 이 경로는 지금 형태를 유지해야 한다.
