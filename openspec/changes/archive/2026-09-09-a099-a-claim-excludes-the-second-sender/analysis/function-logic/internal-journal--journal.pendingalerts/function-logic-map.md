# Function Logic Map: `Journal.PendingAlerts`

- Source: `internal/journal/outbox.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **a099는 이 함수를 편집하지 않는다.** `UndeliveredCount`와 같은 이유로 산출물이 있다 —
> **편집하지 않는다는 것이 design D1의 논거**이고 그 논거가 이 함수의 술어를 근거로 쓴다.
>
> 이 함수는 두 가지를 동시에 진다: **운영자가 읽는 밀린 목록**과
> **`Flush`가 도는 재시도 대상**. a099의 D7이 `Flush`를 바꾸므로 이 함수의 반환이
> 어떻게 소비되는지가 달라진다 — **그러나 이 함수 자체는 안 바뀐다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `limit` | `<= 0`이면 전부 | 호출자 | B1 `:395`가 그것을 판정한다 |
| `alert_outbox.state` | 셋 중 하나 | `schemaV3` `outbox.go:57` | **PENDING만 고른다** (`:393`) |
| 정렬 | `ORDER BY id` — 오래된 것 먼저 | `:393` | id가 AUTOINCREMENT이므로 삽입 순 |
| **임차 열** | **오늘 없다** | — | a099 이후에도 **이 질의는 안 읽는다** |

**불변식 — a099가 지켜야 하는 것**: *"발송 중인 행도 이 목록에 보인다."*
운영자가 밀린 것을 볼 때 **누군가 보내는 중이라는 이유로 사라지면 안 된다.**

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — 분기 2 · 이탈 2 · 호출 5 · defer 1.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:395` | `limit > 0` | 질의에 `LIMIT ?`를 붙이고 인자를 append (`:396-397`) | — | 기존 |
| B2 `:400` | `QueryContext` 실패 | 없음 | 이탈 `:401` `nil, 오류` | 기존 |
| — 이탈 `:404` | 정상 | 없음 — **읽기 전용** | `scanAlerts(rows)` | 기존 + **a099 R21** |

**a099는 두 분기의 조건도 이탈도 안 바꾼다.**

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `append` `:397` | `LIMIT` 인자 | B1 안에서만 | `ast.json` calls |
| `j.db.QueryContext` `:399` | `alertSelect + WHERE state = ? ORDER BY id` | ctx 취소 전파 | 같음 |
| `fmt.Errorf` `:401` | 오류 포장 | — | 같음 |
| `rows.Close` `:403` | **defer** | — | 같음 |
| `scanAlerts` `:404` | 행 → `Alert` | 스캔 오류를 그대로 올린다 (`outbox.go:434-463`) | 같음 |

**live bindings — 프로덕션 호출자 둘, 둘 다 `internal/obs`**:
`notifier.go:437` (`Flush`가 도는 목록) · `notifier.go:491` (`Acknowledge`가
승인할 대상).

**a099의 D7이 바꾸는 것은 `Flush` 쪽의 소비 방식이다** — 목록을 받은 뒤
행마다 `ClaimAlertByID`를 부르고 못 잡은 것을 건너뛴다.
**이 함수는 여전히 전부를 돌려준다.**

## State mutations and fallbacks

- **mutation 없다.** 읽기 전용.
- **폴백 없다.** 오류면 `nil, err`.
- **`scanAlerts`가 `Alert` 구조체를 채운다.** a099가 임차 열을 더해도
  `alertSelect`(`outbox.go:386-388`)의 열 목록을 안 바꾸면 `Alert`에 안 나타난다.
  **a099는 안 바꾼다** — 운영자 표면에 임차를 보이는 것은 a098의 읽기 표면 결정이다.

## Safety conclusion

- **Safe edit boundary**: **없다 — a099는 이 함수를 편집하지 않는다.**
  편집 후 `source_sha256`이 그대로여야 하고, 그것이 §5.3의 확인이다.
- **High-risk impact**: **yes** — 운영자가 밀린 알림을 보는 유일한 읽기다.
  여기서 행이 사라지면 운영자는 그것이 처리됐다고 읽는다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **`limit=0`으로만 프로덕션에서 불린다** (`notifier.go:437`·`:491` 둘 다 `0`).
    B1의 참 경로는 테스트에만 있다. a099가 만든 것이 아니다.
  - **이 함수는 임차를 안 읽는다.** `Flush`가 도는 목록에 **다른 발송자가 이미
    잡은 행이 들어 있을 수 있고**, 그 행은 claim 실패로 건너뛰어진다.
    질의로 걸러내는 편이 효율적이지만 **a099는 안 한다** — 술어를 바꾸면
    이 함수의 두 소비자 중 운영자 쪽 의미가 같이 바뀌기 때문이다.
    **효율은 contract 단계의 문제다.**
