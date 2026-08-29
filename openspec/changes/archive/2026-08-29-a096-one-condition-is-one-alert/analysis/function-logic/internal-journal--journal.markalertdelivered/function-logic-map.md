# Function Logic Map: `Journal.MarkAlertDelivered`

- Source: `internal/journal/outbox.go` (290–301)
- AST evidence: `ast.json` (sha256 `c6612c641a3a…`, 1분기, 반환 2곳)
- Risk scan: `risk-pattern-report.md`

a096은 **이 함수를 편집하지 않는다.** 여기 있는 이유는 이 함수가 운영에서 관측된
증상(`journal: no such alert: 14 (or it is no longer pending)`)을 만든 곳이고, 그 증상이
a096의 실측 완료 조건(design D7)이기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | 존재하는 outbox 행 | `EnqueueAlert` | 없거나 PENDING이 아니면 `ErrAlertNotFound` |
| 대상 행의 `state` | **PENDING이어야 한다** | `WHERE … AND state = ?` (:290) | 0행 갱신 → 오류 |

불변식: **PENDING만 DELIVERED가 된다.** 이미 DELIVERED·ACKNOWLEDGED인 행을 다시 전달
완료로 표시할 수 없다. 이것이 멱등이 아니라 **한 방향 전이**라는 뜻이고, 그래서 두 번째
호출이 조용히 성공하지 않고 오류가 된다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@297 | `ExecContext` 오류 | 없음 | `fmt.Errorf` @298 | 없음(측정: 미진입) |
| — | 정상 | `state=DELIVERED, delivered_at, last_attempt_at, attempts+1, last_error=''` | `requireOneRow` @300 | 기존 |

`requireOneRow`([outbox.go:413-421](../../../../../internal/journal/outbox.go#L413-L421))가
0행을 `ErrAlertNotFound`로 바꾼다. 메시지는 `"%d (or it is no longer pending)"`이며 —
운영 로그에 찍힌 문장이 정확히 이것이다.

## 이 오류가 무엇을 증명하는가

이 오류가 한 번 로그에 남을 때마다, 그 직전에 `Publish`가 **성공했다**.
`deliver`가 이 함수를 부르는 유일한 자리가 `if err == nil`(전송 성공) 안이기 때문이다
([notifier.go:317-318](../../../../../internal/obs/notifier.go#L317-L318)).

따라서 이 오류가 남는다는 것은 **불필요한 재전송이 일어났다**는 뜻이다. 운영 로그에
53건 있었고 a096 이후의 목표값은 0이다.

횟수까지 세려면 `-covermode=count`가 필요하다. 이 저장소의 측정은 `set`이므로 커버리지에서
읽을 수 있는 것은 발생 여부뿐이고, 운영 로그의 줄 수는 셀 수 있다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `RFC3339(j.clk.Now())` | `delivered_at`·`last_attempt_at` | 주입 시계 | AST :291 |
| `j.db.ExecContext` | 상태 전이 | B1 | AST :292 |
| `requireOneRow` | 0행을 오류로 | `ErrAlertNotFound` | AST :300 |

호출자(CodeGraph): `obs.Notifier.deliver` 한 곳뿐([notifier.go:258](../../../../../internal/obs/notifier.go#L258)).

## State mutations and fallbacks

- 단일 UPDATE. 트랜잭션 없음 — 한 문장이 원자적이다.
- 되돌림 없음. 실패는 호출자에게 오류로 가고, `deliver`는 그것을 로그만 하고 `true`를 유지한다.

## Safety conclusion

- Safe edit boundary: **없음 — 편집하지 않는다.** a096은 이 함수가 **덜 호출되게** 만들 뿐이다.
- High-risk impact: no (편집 없음). 원장 스키마·전이 규칙 그대로.
