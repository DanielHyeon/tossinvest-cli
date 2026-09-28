# Function Logic Map: `claimOwed`

- Source: `internal/journal/outbox.go` (269-315)
- AST evidence: `ast.json` — branches 8, returns 7, calls 2, assignments 2,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**증거용.** a092는 이 함수를 편집하지 않는다. 17판 D0.3의 *"기록 경로가 배달
잠금을 잡지 않아도 PENDING 행의 내용이 덮어써지지 않는다"*는 주장이 **오직 B2의
반환값 하나**에 걸려 있다. 그래서 이 함수를 따로 열거한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `state` | `PENDING`/`DELIVERED`/`ACKNOWLEDGED`/그 밖의 무엇이든 | `alert_outbox.state` — **CHECK 제약 없음**(`:161`) | 미상은 B8이 fail-open |
| `deliveredAt`·`acknowledgedAt` | `sql.NullString`, RFC3339 또는 파싱 불가 | 같은 행 | 파싱 불가면 B5가 fail-open |
| `now` | 호출자의 `j.clk.Now()` | 주입 시계 | 과거로 튀면 B6이 fail-open |
| `remindAfter` | `<= 0`이면 재알림 안 함 | `Notifier.remindAfter()` | B4가 `(false,false)` |

**순수 함수다.** 오류 반환이 없고, I/O가 없고, `latestStamp`와 `now.Sub` 둘만
부른다(AST calls 2). 그래서 이 판정에는 **기한도 실패 모드도 없다** — 17판이
관측 사이클에 남기는 몫에 이 함수가 기여하는 시간은 상수다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return `(owed, rearm)` |
|---|---|---|---|
| B1 `:275` | `switch state` | 없음 | — |
| **B2 `:276`** | **`state == PENDING`** | 없음 | **`:278` `(true, false)`** |
| B3 `:279` | `state ∈ {DELIVERED, ACKNOWLEDGED}` | 없음 | 아래 B4~B7 |
| B4 `:280` | `remindAfter <= 0` | 없음 | `:281` `(false, false)` |
| B5 `:284` | 타임스탬프를 하나도 못 읽음 | 없음 | `:287` **`(true, true)`** — fail-open |
| B6 `:290` | `elapsed < 0` (시계 역행) | 없음 | `:302` **`(true, true)`** — fail-open |
| B7 `:304` | `elapsed < remindAfter` | 없음 | `:305` `(false, false)` |
| — `:307` | 창이 지났다 | 없음 | `(true, true)` |
| **B8 `:308`** | **`default` — 미상 상태** | 없음 | `:313` **`(true, true)`** — fail-open |

**B2가 D0.3이 서 있는 한 줄이다.** PENDING 행에 대해 `rearm=false`이므로
`ClaimAlertForDelivery`의 B6(`outbox.go:197`) UPDATE가 실행되지 않는다. 즉
**아직 배달되지 않은 행의 내용을 다른 관측이 덮어쓰는 경로가 없다.**

세 개의 fail-open(B5·B6·B8)은 전부 **"모르면 보낸다"**로 같은 방향이다. 이것들은
`(true, true)`이므로 재무장이 일어나지만, 재무장의 대상은 PENDING이 **아닌** 행뿐이고
`PendingAlerts`는 PENDING만 돌려주므로(`outbox.go:393`) 17판의 배달 루프가 손에
쥔 행과 겹치지 않는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `latestStamp` `:283` | 두 타임스탬프 중 최신 | 파싱 실패는 `ok=false`로만 표현, 오류 없음 (`:318-334`) | AST calls |
| `now.Sub` `:289` | 경과 시간 | 순수 | AST calls |

**호출 2개뿐이고 둘 다 순수하다.** DB도 네트워크도 잠금도 없다.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| — | — | **없음.** 이 함수는 아무것도 바꾸지 않는다 |

- fallback은 세 곳(B5·B6·B8)이고 전부 **보내는 쪽**이다. 침묵보다 중복 발송을
  고른 판단이며 a096 라운드 2가 근거를 적었다(`:294-301`).

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: yes — 알림이 나가느냐 마느냐의 전부가 여기다.
  17판이 이 함수에 **의존하는 방식은 B2 하나**이고, B2가 바뀌면 D0.3의
  "기록 경로가 잠금을 안 잡아도 된다"는 근거가 사라진다. 그러므로
  **17판 구현이 이 함수를 건드리면 D0.3을 다시 써야 한다.**
- fail-open 세 갈래는 17판에서 **더 안전해진다**: 재무장된 행은 배달 루프가
  다음 주기에 집어가고, 그 사이 관측 사이클은 기다리지 않는다.
- B4(`remindAfter <= 0`)는 프로덕션 경로에서 도달하지 않는다 —
  `Notifier.remindAfter()`가 0 이하를 기본값 1시간으로 바꾼다(`notifier.go:280-285`).
  기록 전용 호출자만 0을 준다.
