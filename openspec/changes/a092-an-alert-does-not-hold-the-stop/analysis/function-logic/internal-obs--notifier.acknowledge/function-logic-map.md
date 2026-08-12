# Function Logic Map: `Notifier.Acknowledge`

- Source: `internal/obs/notifier.go` (476-514)
- AST evidence: `ast.json` — branches 10, returns 6, calls 12, assignments 2,
  **defers 1, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**게이트를 푸는 유일한 경로이고, 프로덕션 호출자가 없다.**
17판이 진입 래치를 더 확실하게 걸게 만들므로, **푸는 쪽이 없다는 사실이
17판에서 더 무거워진다.** 그래서 이 함수를 열거한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `operator` | **공백이면 거부** | 호출자 | B1 `:477` → 오류 `:478` |
| `n.Journal` | nil이면 게이트만 풀고 끝 | 배선 | B2 `:480` → `:484` `nil` |
| `n.Gate` | nil일 수 있다 | 배선 | B3 `:481`, B10 `:510` |
| `ids` | 비었으면 **미전달 전부** | 호출자 | B4 `:490` → `PendingAlerts(ctx, 0)` |
| `n.mu` | 배달 뮤텍스 | `:487`, defer `:488` | **read-then-decide를 원자화한다** |

**`:470-475`의 주석이 뮤텍스의 근거를 적는다**: 해제는 "남은 수를 세고, 0이면
푼다"이다. 그 사이에 재무장이 settled 행을 pending으로 되돌리면
**미전달 critical 알림이 있는데 게이트가 열린다** — 게이트가 존재하는 이유 그 자체가
깨진다.

**17판은 이 잠금을 유지한다.** `claimAndDeliver`에서 뺀 것은 발행이지 배제가 아니고,
여기의 배제는 발행이 아니라 **읽고-판단하기**를 지킨다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| B1 `:477` | `operator` 공백 | 없음 | 오류 `:478` |
| B2 `:480` | `n.Journal == nil` | B3 참이면 `Gate.Clear` `:482` | `:484` `nil` |
| B3 `:481` | `n.Gate != nil` (journal 없는 경로) | 래치 해제 | — |
| **B4 `:490`** | **`len(ids) == 0`** | `PendingAlerts(ctx, 0)` `:491` — **무제한** | — |
| B5 `:492` | 그 조회 실패 | 없음 | 오류 `:493` |
| B6 `:495` | `range pending` | `ids`에 누적 `:496` | — |
| B7 `:499` | `range ids` | 행마다 `AcknowledgeAlert` `:500` | — |
| **B8 `:500`** | **승인 실패이면서 `ErrAlertNotFound`가 아니다** | 없음 | 오류 `:502` |
| B9 `:507` | `UndeliveredCount` 실패 | 없음 | 오류 `:508` |
| **B10 `:510`** | **`remaining == 0 && n.Gate != nil`** | **`Gate.Clear` `:511`** | — |
| — `:513` | — | — | `nil` |

**B8의 `errors.Is(err, ErrAlertNotFound)` 예외가 멱등성이다**(`:501`).
이미 승인된 행을 다시 승인해도 실패하지 않는다 — 운영자가 같은 명령을 두 번 쳐도
안전하다.

**B10이 유일한 해제다.** `remaining == 0`이 아니면 게이트는 그대로 남는다.
`:466-468`의 주석이 근거를 적는다: 전송이 회복된 것만으로는 부족하고,
알림은 **사람이 보게 하려고** 존재했으므로 "네트워크가 돌아왔다"는 그 사람이 아니다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` `:477` | 신원 검증 | 순수 | AST calls |
| `n.Gate.Clear` `:482`·`:511` | **래치 해제 — 이 두 곳뿐** | void | AST calls |
| `n.mu.Lock` `:487` | read-then-decide 원자화 | **17판도 유지** | AST calls |
| `n.mu.Unlock` `:488` | **defer** | — | AST defers **1** |
| `n.Journal.PendingAlerts` `:491` | 대상 수집 | `limit = 0` — 여기서는 전부가 맞다 | AST calls |
| `n.Journal.AcknowledgeAlert` `:500` | 행 승인 | CAS, `ErrAlertNotFound` 허용 | AST calls |
| `n.Journal.UndeliveredCount` `:506` | 남은 수 | 전역 | AST calls |

**네트워크 없음.** 이 함수는 전송하지 않는다 — 정산만 한다.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `alert_outbox` 행 | `:500` | `PENDING` → `ACKNOWLEDGED`, 조작자 이름 기록 |
| 진입 게이트 래치 | `:482`·`:511` | 인메모리 해제 |

- fallback: B8의 `ErrAlertNotFound` 관용 하나.

## Safety conclusion

- **Safe edit boundary**: a092는 이 함수의 **본문을 편집하지 않는다.**
  17판이 더하는 것은 이 함수를 **부를 수 있는 경로**다.
- **High-risk impact**: yes — 게이트 해제의 유일한 자리다.
- **측정한 결함.** `Notifier.Acknowledge`의 프로덕션 호출자는 **0이다**
  (`internal/`·`cmd/` 전체 `rg` 확인). 그리고 `Gate.Clear(ReasonAlertUndelivered)`는
  `:482`와 `:511` 두 곳에만 있으며 둘 다 이 함수 안이다.
  **따라서 `ReasonAlertUndelivered` 래치가 걸리면 프로세스 재시작 외에는
  풀 방법이 프로덕션에 없다.**
- **17판이 이것을 더 중요하게 만든다**: 배달 루프가 주기마다 실패를 기록하고
  래치를 걸므로, 래치가 걸릴 확률이 오른다. 푸는 쪽이 없으면
  **전송기가 회복돼도 신규 진입이 영구히 막힌다.**
- a092가 만드는 결함이 아니라 **발견한** 결함이다. §6.0 **R17-11**이
  `Acknowledge`가 게이트를 푸는 것을 관측하고, **운영자가 그것을 호출할 수 있는
  경로가 필요한지**는 review.md 17.9의 미결로 남긴다 — 새 CLI 표면은
  a092의 범위를 넘고, 사용자 결정이 필요하다.
