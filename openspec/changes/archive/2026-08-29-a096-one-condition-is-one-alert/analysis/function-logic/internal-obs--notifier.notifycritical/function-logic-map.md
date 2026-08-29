# Function Logic Map: `Notifier.notifyCritical`

- Source: `internal/obs/notifier.go` (170–207)
- AST evidence: `ast.json` (sha256 `c9dee3479706…`, 4분기, 반환 3곳)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `e` | critical로 등급된 이벤트만 | `Notify` B1@128이 걸러 넘긴다 | 등급 재판정 없음 |
| `n.Journal` | nil 가능 | 조립 시점 | B1: best-effort로 강등하고 반환 |
| `n.Log` | nil 가능 | 조립 시점 | B2: 강등 사실을 알릴 수 없으면 조용히 진행 |
| `n.eventKey(e)` | 조건에서 조립, 시계 아님 | `e.Key` 또는 타입 기반 | 같은 조건 = 같은 key |
| `sent`, `owed` | bool 두 개 | `claimAndDeliver` | B4: owed였는데 못 보냈을 때만 escalate |

계약(주석): "It returns an error only when a *critical* event could not be made durable —
that is, when the outbox write itself failed. A failed send is not an error to the caller."
a096은 이 계약을 바꾸지 않는다. 전송을 **하지 않은** 경우도 오류가 아니다.

## Branches and early returns

RED(base `ec29dc72`)에서도 이 함수는 4분기였다. 분기 **수**는 같고 B4의 조건이 바뀌었다:
`!n.deliver(...)`(전송 실패) → `owed && !sent`(보내야 했는데 못 보냄). claim과 전송은
`claimAndDeliver`로 옮겼다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@171 | `n.Journal == nil` | `publishBestEffort` (전송 시도) | `nil` @181 | 기존 `TestCriticalWithoutAJournalIsLoudRatherThanSilent` |
| B2@175 | `n.Log != nil` | 경고 한 줄 | — | 위와 같음 |
| B3@195 | claim/전송이 오류 | 없음 | `fmt.Errorf` @196 — **유일하게 올라가는 오류** | 없음(측정: 미진입) |
| B4@199 | `owed && !sent` | `escalate` | `nil` @206 | `TestPersistentDeliveryFailureBlocksEntries`, `TestADeadTransportIsStillFoundAfterASuccessfulDelivery` |

**`owed &&`가 B4의 안전 조건이다.** 보낼 필요가 없어서 안 보낸 것은 전달 실패가 아니므로
진입을 차단해서는 안 된다. `!sent` 하나만으로 판정하면 억제될 때마다 gate가 잠긴다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.publishBestEffort` | journal 없을 때의 강등 경로 | 오류를 삼키고 로그만 | AST :180 |
| `n.eventKey` | 중복 제거 key | 없음 | AST :185 |
| `encodeFields` | payload 직렬화 | 실패 시 빈 문자열 | AST :190 |
| `n.claimAndDeliver` | 기록 + 전달 필요 여부 + 전송, 한 잠금 안에서 | B3에서 유일하게 올라간다 | AST :194 |
| `n.escalate` | 전달 실패의 지속 결과 | `claimAndDeliver` 밖에서 호출 — 재진입 교착 회피 | AST :204 |

`escalate`가 잠금 밖에 있어야 하는 이유는 주석에 있다: 그것이 `ModeAnnouncer`를 통해
알리고, 이 Notifier에 배선된 announcer는 `Notify`로 재진입해 Go가 재진입 허용하지 않는
뮤텍스에서 교착한다.

## State mutations and fallbacks

- 이 함수 자체는 원장에 쓰지 않는다. 쓰기는 `ClaimAlertForDelivery`(행 생성·재무장)와
  `deliver` 안의 `MarkAlertDelivered`/`MarkAlertAttemptFailed`가 한다.
- 구조화 로그는 이 함수가 아니라 `Notify`의 `logEvent`(:126)가 남기며, 등급 분기보다
  **먼저** 실행된다. 따라서 전송이 억제돼도 관측 기록은 매번 남는다.

## Safety conclusion

- Safe edit boundary: B4의 조건에 `owed &&`를 더한 것과, claim+전송을 `claimAndDeliver`로
  옮긴 것. 반환 계약과 escalate의 위치는 그대로다.
- High-risk impact: **no** — 주문·손절·익절·사이징·체결 경로에 닿지 않는다.
- 실질 위험 두 가지 모두 `owed`의 정의에 달려 있다: 과잉 억제(보낼 것을 안 보냄)와
  과잉 escalate(안 보내도 될 것에 gate를 잠금). 전자는 `claimOwed`가, 후자는 B4의
  `owed &&`가 막는다.
