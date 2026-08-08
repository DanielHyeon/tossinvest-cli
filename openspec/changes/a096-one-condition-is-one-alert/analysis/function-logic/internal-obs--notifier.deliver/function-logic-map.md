# Function Logic Map: `Notifier.deliver`

- Source: `internal/obs/notifier.go` (303–347)
- AST evidence: `ast.json` (sha256 `c9dee3479706…`, 10분기, 반환 2곳)
- Risk scan: `risk-pattern-report.md`
# Function Logic Map: `Notifier.deliver`

a096은 이 함수의 **본문을 한 줄도 바꾸지 않았다.** 바꾼 것은 잠금의 소유자다:
1판까지 이 함수가 `n.mu`를 잡았고, 2판은 호출자 `claimAndDeliver`가 claim부터 함께 잡는다.
이 map이 있는 이유는 ① 여기가 `Publish`를 실행하는 곳이라 폭주의 실제 지점이고,
② B5가 이 change의 측정점이기 때문이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `id` | outbox 행 id | `ClaimAlertForDelivery` | PENDING이 아니면 표시가 실패한다 |
| `n.mu` | **호출자가 보유** | `claimAndDeliver` | 이 함수는 잡지 않는다 |
| `n.Attempts` | ≤0이면 기본값 | 조립 시점 | B1: `DefaultCriticalAttempts`(3) |
| `n.Publisher` | nil 가능 | 조립 시점 | B3: 예산을 돌지 않고 즉시 실패로 |
| `n.Gate` | nil 가능 | 조립 시점 | B10: 없으면 차단하지 않는다 |

불변식: **한 번에 하나의 claim-and-send.** 1판의 "한 번에 하나의 전달 루프"보다 넓다 —
루프만 배타적이면 두 호출자가 각자 owed를 얻은 뒤 차례로 보낼 수 있다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@305 | `attempts <= 0` | 기본값 대입 | — | 없음(미진입) |
| B2@311 | 재시도 루프 | — | — | 기존 |
| B3@312 | `n.Publisher == nil` | `lastErr` 설정 후 break | — | 없음(미진입) |
| B4@317 | `Publish`가 nil | 전달 표시 시도 | `true` @321 | 기존 |
| **B5@318** | `MarkAlertDelivered`가 오류 | 로그 한 줄뿐 — `true`는 그대로 | — | 아래 |
| B6@324 | `MarkAlertAttemptFailed`가 오류 | 로그 한 줄 | — | 없음(미진입) |
| B7@327 | `attempt < attempts` | 대기 | — | 기존 |
| B8@328 | `wait`가 false(ctx 종료) | break | — | 없음(미진입) |
| B9@338 | `n.Log != nil` | 소진 로그 | — | 기존 |
| B10@343 | `n.Gate != nil` | `Gate.Block` | `false` @346 | 기존 + `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` |

## B5가 무엇을 말하는가

이 블록은 `MarkAlertDelivered`가 오류를 돌려줬을 때만 들어간다. 그 오류의 실질 경로는
`WHERE id = ? AND state = 'PENDING'`이 0행을 갱신하는 것 — **이미 전달된 행에 다시 전달
표시를 시도했다**는 뜻이고, 그 시점에 `Publish`는 이미 성공한 뒤다(B4가 참이어야 도달).
운영 로그의 `journal: no such alert: 14 (or it is no longer pending)`가 정확히 여기다.

RED 측정값은 `notifier.go:258.88,260.5 1 1`(진입)이었고, 격리 측정으로
`TestNotifierIsConcurrencySafe`가 만든 것임을 특정했다. 결과는 discard writer로 가는
로그 한 줄이었고, 그래서 아무도 보지 않았다.

`-covermode=set`은 실행 여부를 0/1로 기록하며 **횟수를 세지 않는다.** 이 칸에서 읽을 수
있는 것은 "그 경로가 발생했다/않았다"뿐이다. 1판 문서가 이 칸을 횟수로 읽었고 독립 리뷰
1라운드가 그것을 지적했다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `notificationFor` | 이벤트 → 전송 메시지 | 없음 | AST :308 |
| `n.Publisher.Publish` | 실제 전송 | 오류는 예산 안에서 소비 | AST :316 |
| `n.Journal.MarkAlertDelivered` | 성공을 원장에 | 오류는 로그만, `true`를 막지 않는다 | AST :318 |
| `n.Journal.MarkAlertAttemptFailed` | 실패를 원장에 | 오류는 로그만 | AST :324 |
| `n.wait` | 재시도 간 대기 | ctx 종료 시 false | AST :328 |
| `n.Gate.Block` | 예산 소진의 결과 | 없음 | AST :344 |

호출자(CodeGraph): `Notifier.claimAndDeliver` 한 곳(`notifier.go:240`).

## State mutations and fallbacks

- 성공: 행 → DELIVERED. 실패: attempts 증가, 행은 PENDING 유지.
- `Publish` 성공 + 표시 실패면 행은 PENDING으로 남고, 다음 관측에서 **다시 보낸다** —
  원장이 "아직 전달 못 했다"고 말하는 것을 그대로 따르는, 안전한 방향이다.

## Safety conclusion

- Safe edit boundary: 본문은 편집하지 않았다. 바뀐 것은 잠금 획득 위치이며,
  그 전제를 함수 주석에 `PRECONDITION`으로 명시했다.
- High-risk impact: no (본문 편집 없음).
