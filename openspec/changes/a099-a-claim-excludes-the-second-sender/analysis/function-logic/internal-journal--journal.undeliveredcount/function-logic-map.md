# Function Logic Map: `Journal.UndeliveredCount`

- Source: `internal/journal/outbox.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **a099는 이 함수를 편집하지 않는다.** 이 산출물이 있는 이유는 반대다 —
> **편집하지 않는다는 것이 design D1의 안전 논거 전체**이고, 그 논거는 이 함수의
> 술어가 무엇을 읽는지에 달려 있다. **논거가 함수 내부를 근거로 쓰므로 산출물이 먼저다.**
>
> `outbox.go:407`의 주석이 이 값의 용도를 적는다 —
> *"UndeliveredCount is the number the entry gate reacts to."*
> **알림 하나가 진입 차단을 좌우하는 자리다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 취소 가능 | 호출자 | 취소는 질의 오류로 나온다 |
| `alert_outbox.state` | 셋 중 하나 | `schemaV3` `outbox.go:57` | **PENDING만 센다** (`:411`) |
| 입력 인자 | **없다** | — | 이 함수는 매개변수를 안 받는다 |

**불변식 — a099가 지켜야 하는 것**: *"발송 중인 행도 이 수에 잡힌다."*
오늘 자동으로 성립한다. 발송 중이라는 상태가 없기 때문이다.
**design D1의 안 A(`SENDING` 상태 추가)를 골랐다면 깨졌을 불변식이다.**

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — 분기 1 · 이탈 2 · 호출 3.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:410` | `QueryRowContext(...).Scan` 실패 | 없음 | 이탈 `:412` `0, 오류` | 기존 |
| — 이탈 `:414` | 정상 | 없음 — **읽기 전용** | `n, nil` | 기존 + **a099 R21** |

**분기가 하나뿐이고 그것은 오류 검사다.** 판정 전체가 SQL 술어
`WHERE state = ?`(`:411`, 인자 `AlertPending`)에 있다.

**a099는 이 술어를 안 바꾼다.** 임차는 상태가 아니라 열이므로 발송 중인 행도
여전히 `state = 'PENDING'`이고 여전히 세어진다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.db.QueryRowContext` `:410` | `COUNT(*) WHERE state = ?` | ctx 취소 전파. 트랜잭션 없음 | `ast.json` calls |
| `Scan` `:411` | 정수 하나 | 실패는 B1 | 같음 |
| `fmt.Errorf` `:412` | 오류 포장 | — | 같음 |

**live bindings — 프로덕션 호출자 둘, 둘 다 `internal/obs`**:
`notifier.go:460` (`Flush`의 반환 `remaining`) · `notifier.go:506` (`Acknowledge`).
**`Acknowledge`가 이 값이 0일 때만 진입 게이트를 푼다** — 그것이 「진입 게이트가
반응하는 수」의 실제 배선이다.

## State mutations and fallbacks

- **mutation 없다.** 읽기 전용 질의 하나.
- **폴백 없다.** 오류면 `0, err`을 돌려준다. **0이 「밀린 것 없음」과 같은 값이라는
  것이 이 함수의 가장 날카로운 자리다** — 호출자가 오류를 안 보면 밀린 알림이
  있는데 게이트를 푼다. `notifier.go:506-507`이 오류를 검사한다.
  **a099는 이 자리를 안 건드리고, 안 건드린다는 것이 §5.3의 확인 대상이다.**

## Safety conclusion

- **Safe edit boundary**: **없다 — a099는 이 함수를 편집하지 않는다.**
  편집 후 `source_sha256`이 그대로여야 한다. 그것이 §5.3의 확인이다.
- **High-risk impact**: **yes** — 이 수가 신규 진입 차단을 좌우한다(불변식 5).
  편집하지 않는 것이 이 change에서 가장 중요한 결정이다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **이 함수는 임차를 모른다.** 만료된 임차를 가진 PENDING 행도, 살아 있는 임차를
    가진 PENDING 행도 똑같이 센다. **그것이 옳다** — 둘 다 미전달이다.
  - **오류와 0의 구분**은 이 함수 밖에 있다. a099가 만든 문제가 아니고 a099가
    고칠 것도 아니다. **`not-applicable`: 이 change는 그 구분을 근거로 쓰지 않는다.**
