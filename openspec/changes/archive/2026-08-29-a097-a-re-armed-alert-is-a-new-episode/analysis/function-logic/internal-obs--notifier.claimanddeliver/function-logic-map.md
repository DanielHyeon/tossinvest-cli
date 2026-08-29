# Function Logic Map: `Notifier.claimAndDeliver`

- Source: `internal/obs/notifier.go` (L220-241)
- AST evidence: `ast.json` (2 branches, 3 returns)
- Risk scan: `risk-pattern-report.md`

a097이 **본문을 바꾸는** 두 번째 함수다. 바뀌는 것은 `B1@227`(claim 실패) 분기의 내용이며,
반환 계약과 잠금 범위는 바뀌지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `record` | `EventKey`·`Type` 비공백 | `notifyCritical@184-191` | 원장이 B1로 오류 반환 |
| `e` | 전송할 이벤트 | 호출자 | — |
| `n.Journal` | 비-nil | `notifyCritical` B1@171이 이미 걸렀다 | nil이면 이 함수에 도달하지 않음 |
| `n.mu` | 이 함수가 **잡는다** (`Lock@223`) | — | 재진입 불가 (Go 뮤텍스는 비재귀) |
| `n.Gate` | nil 가능 (선택 배선) | 조립 지점 | **a097: nil 검사 후 Block** |

**불변식 (a096 round 1의 결론)**: claim과 send는 **하나의 배타 구간**이다. claim만
트랜잭션으로 감싸고 send를 밖에 두면 두 관측이 같은 미전달 행을 읽고 둘 다 보낸다.
`Lock@223`부터 `deliver@240` 반환까지가 그 구간이다.

**불변식 (a096의 deadlock 회피)**: 이 구간 안에서 `escalate`를 부르지 않는다. 승격은
`ModeAnnouncer`를 통해 알리고, 이 Notifier에 배선된 announcer는 `Notify`로 재진입해
같은 뮤텍스에서 교착한다. 그래서 승격은 `notifyCritical` B4@199, **잠금 밖**에 있다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@227 | claim 오류 | **현재: 없음.** a097: 구조화 로그 + `Gate.Block(ReasonAlertUndelivered)` | `false, false, err` @228 | **a097 신규** 2.6 |
| B2@230 | `!owed` | 없음 (로그는 `Notify`가 이미 남겼다) | `false, false, nil` @238 | 기존 (a096) |

정상 반환: `n.deliver(ctx, id, e), true, nil` @240.

**B1이 왜 위험한가.** 반환값이 `owed=false`이므로 `notifyCritical` B4@199의
`owed && !sent` 승격이 **구조적으로 도달 불가**다. 그리고 이 함수에도 `notifyCritical`에도
gate 호출이 없다(AST calls 목록으로 확인). 마지막 호출자
`flatten.Saga.event`(`flatten.go:694`)는 오류를 `_ =`로 버린다. 결과: critical 알림이
기록도 전송도 되지 않았는데 로그도 gate도 승격도 없다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.mu.Lock@223` / `Unlock@224` | claim+send 배타 구간 | — | AST calls |
| `n.Journal.ClaimAlertForDelivery@226` | owed 판단 + 재무장 | 오류는 B1 | AST calls |
| `n.remindAfter@226` | 창 길이 (`RemindAfter` 또는 `DefaultRemindAfter`) | 오류 없음 | AST calls |
| `n.deliver@240` | 재시도 예산 안에서 전송, 실패 시 gate 래치 | 오류 대신 `bool`(정착 여부) | AST calls |
| **a097 추가** `n.Log.Error` | claim 실패 기록 | 오류 없음 | design D2 |
| **a097 추가** `n.Gate.Block` | 신규 진입 차단 | 오류 없음. 멱등(`retry.go:501`) | design D2 |

**lock ordering 확인**: `internal/execgw`는 비테스트 코드에서 `internal/obs`를 import하지
않는다 → `Gate.Block`이 obs로 되돌아오는 경로가 없다. 그리고 같은 패키지의
`deliver`가 이미 `n.mu` 아래에서 `n.Gate.Block`을 부른다(`notifier.go:343`, `:368`).
a097은 **새로운 잠금 순서를 만들지 않는다.**

## State mutations and fallbacks

- 이 함수 자체의 mutation은 없다. 원장 변경은 `ClaimAlertForDelivery`가, gate 래치는
  `deliver`(그리고 a097 이후 B1)가 한다.
- a097이 추가하는 것은 **fallback이 아니라 결과 통지**다. 오류는 계속 반환된다.

## Safety conclusion

- Safe edit boundary: **B1 분기의 본문만.** 반환 튜플, 잠금 범위, B2와 정상 경로는 불변.
- High-risk impact: **yes** — 진입 게이트를 잠근다. Pre-Edit 선언 대상.
- 안전 방향: `EntryGate.Block`은 신규 진입만 막고 청산은 통과시킨다
  (`internal/execgw/retry.go:498`). 안전 불변식 §0-3(손절 즉시성)에 영향 없음.
