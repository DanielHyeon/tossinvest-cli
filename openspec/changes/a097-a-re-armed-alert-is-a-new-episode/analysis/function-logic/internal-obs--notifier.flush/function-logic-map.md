# Function Logic Map: `Notifier.Flush`

- Source: `internal/obs/notifier.go` (L391-426)
- AST evidence: `ast.json` (6 branches, 4 returns)
- Risk scan: `risk-pattern-report.md`

**a097은 이 함수의 본문을 바꾸지 않는다.** 산출물이 필요한 이유는 둘이다 — proposal R1이
"보낼 내용을 **행에서** 만든다"를 근거로 삼고, R4가 이 함수의 뮤텍스를 테스트로 고정한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Journal` | nil 가능 | 조립 지점 | B1: `0, 0, nil` — 조용히 무동작 |
| `n.Publisher` | nil 가능 | 조립 지점 | B4: 루프 중단, 남은 행은 그대로 |
| `pending` | `PendingAlerts(ctx, 0)` 전체 | 원장 | B2에서 오류 반환 |
| `n.mu` | 이 함수가 **잡는다** (`Lock@398`) | — | notify 경로와 **같은** 뮤텍스 |

**불변식**: 이 함수와 notify 경로는 같은 뮤텍스를 공유한다. 공유하지 않으면 flush와 관측이
같은 행을 동시에 발행한다 — 이 패키지가 막으려는 이중 전송이다(`:395-397` 주석,
a096 round 1 blocker 1).

**a097이 고정하는 것**: 이 뮤텍스에는 오늘 테스트가 없다. 제거해도 스위트가 초록이다
(a096 3판 뮤테이션 실측). R4가 publisher 재진입 카운터로 그것을 잡는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1@392 | `n.Journal == nil` | 없음 | `0, 0, nil` @393 | 없음 — 조립 실수 (`not-applicable`) |
| B2@402 | `PendingAlerts` 오류 | 없음 | `0, 0, err` @403 | 없음 — 장애 주입 (`not-applicable`) |
| B3@405 | `range pending` | 루프 | — | 기존 |
| B4@406 | `n.Publisher == nil` | `break` | 아래 `UndeliveredCount`로 | 없음 — 조립 실수 (`not-applicable`) |
| B5@415 | `Publish` 실패 | `MarkAlertAttemptFailed@416` — **오류를 버린다** | `continue` | 없음 (범위 밖, a096 C4) |
| B6@419 | `MarkAlertDelivered` 실패 | 없음 | `delivered, 0, merr` @420 | 없음 — 장애 주입 (`not-applicable`) |

정상 반환: `delivered, remaining, err` @425.

**R1의 근거가 여기 있다.** `msg`는 `:409-414`에서 `alert.Title`·`alert.Body`로 만들어진다 —
**행에서** 만든다. 반면 `deliver`는 `notificationFor(e, …)`(`:310`)로 살아 있는 이벤트에서
만든다. 따라서 재무장이 본문을 갱신하지 않으면 **이 경로만** 옛 원인을 보낸다.

**잠복인 이유**: `Flush`에 비테스트 호출자가 없다(a096 review C3). a097은 그것을 배선하지
않고, 잠금과 계약만 고정한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.mu.Lock@398` / `Unlock@399` | notify 경로와의 배타 | — | AST calls |
| `n.Journal.PendingAlerts@401` | backlog 전체 | 오류는 B2 | AST calls |
| `n.Publisher.Publish@415` | 전송 | 오류는 B5 (행별 계속) | AST calls |
| `n.Journal.MarkAlertAttemptFailed@416` | 실패 기록 | **오류를 버린다** — 기존 문제 | AST calls |
| `n.Journal.MarkAlertDelivered@419` | 전달 표시 | 오류는 B6 (전체 중단) | AST calls |
| `n.Journal.UndeliveredCount@424` | 남은 수 | 오류는 반환 | AST calls |

## State mutations and fallbacks

- 원장: `MarkAlertAttemptFailed`(B5), `MarkAlertDelivered`(B6 직전). 이 함수 자체의
  메모리 상태 변경은 `delivered` 카운터뿐이다.
- B5와 B6의 비대칭이 눈에 띈다 — 전송 실패는 행별로 넘어가고, 기록 실패는 전체를 멈춘다.
  a097은 이 비대칭을 바꾸지 않는다.

## Safety conclusion

- Safe edit boundary: **없음 — a097은 이 함수를 편집하지 않는다.** 테스트만 추가한다.
- High-risk impact: yes (알림 전달 경로)
- 범위 밖으로 남기는 것: 비테스트 호출자 부재, `MarkAlertAttemptFailed` 오류 버림
  (tasks §8).
