# Function Logic Map: `OpenReadOnly`

- Source: `internal/journal/readonly.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **a099가 이 함수에서 바꾼 것은 식별자 하나다.** `defaultBusyTimeout` →
> `DefaultBusyTimeout` (`:183`). 상수가 export된 이유는 `obs`가 배달 상한을
> 유도할 때 그 값을 **읽어야** 하기 때문이다 — 함수 안에 숨은 숫자는 유도가
> **복사**해야 하는 숫자이고, 복사한 숫자는 언젠가 안 맞는다.
>
> **그래서 이 번들은 「무엇이 안 바뀌었나」의 증거다.** 값도, 분기도, 이탈도
> 그대로다. `check_analysis.py`가 이 함수를 요구하는 이유는 diff가 그 줄을
> 건드렸기 때문이지 로직이 바뀌었기 때문이 아니다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `opts.Path` | 빈 문자열이면 `DefaultPath()` | B1 `:156` | B2 `:158` → 이탈 `:159` |
| 그 경로의 파일 | **존재해야 하고 디렉터리가 아니어야 한다** | B3 `:168` · B6 `:173` | `ErrJournalMissing` |
| `opts.BusyTimeout` | **`<= 0`이면 `DefaultBusyTimeout`** | B7 `:183` | — |
| 파일시스템 allowlist | **일부러 적용 안 한다** | 주석 `:176-181` | 읽기 전용 연결은 fsync 계약이 필요 없다 |

**불변식**: *"이 연결은 절대 쓰지 않는다."* `readOnlyDSN`이 그것을 진다.
a099는 그 DSN도, 이 함수의 어떤 판정도 안 건드렸다.

## Branches and early returns

AST 열거 — 분기 10 · 이탈 8 · defer 0.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:156` | `opts.Path == ""` | `DefaultPath()` | — | 기존 |
| B2 `:158` | `DefaultPath` 실패 | 없음 | `:159` err | 없음 |
| B3 `:168` | `os.Stat` 실패 | 없음 | (B5로) | 기존 |
| B5 `:169` | `errors.Is(err, os.ErrNotExist)` | 없음 | `:170` `ErrJournalMissing` | 기존 |
| — 이탈 `:172` | 그 밖의 stat 오류 | 없음 | 포장 오류 | 없음 |
| B4 `:173` | B3의 `else` | — | — | — |
| B6 `:173` | `info.IsDir()` | 없음 | `:174` `ErrJournalMissing` | 기존 |
| B7 `:183` | **`busy <= 0`이면 `DefaultBusyTimeout`** | 값 채우기 | — | **없음 — 기본값 쪽 미검증** |
| B8 `:188` | `sql.Open` 실패 | 없음 | 포장 오류 | 없음 |
| B9 `:197` | `PingContext` 실패 | `db.Close()` | 포장 오류 | 없음 |
| B10 `:203` | `ro.checkSchema` 실패 | `db.Close()` | err | 기존 (스키마 상한 테스트) |
| — 이탈 `:207` | 정상 | 연결 하나를 연다 | `ro, nil` | 기존 (콘솔 테스트 다수) |

**B4는 B3의 `else` 절 그 자체다.** AST가 `else`와 그 안의 `if`(B6)를 따로 센다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `DefaultPath` `:157` | 경로 기본값 | 실패면 B2 | `ast.json` calls |
| `os.Stat` `:168` | **드라이버에 묻기 전에 존재를 확인** | 없으면 타입 있는 답 | 같음 |
| `sql.Open` `:188` + `readOnlyDSN` | 읽기 전용 연결 | DSN이 쓰기를 막는다 | 같음 |
| `db.PingContext` `:197` | 실제 연결 | 실패면 닫고 오류 | 같음 |
| `ro.checkSchema` `:203` | **스키마가 이 빌드보다 새로우면 거절** | 실패면 닫고 오류 | 같음 |

**live bindings**: 콘솔·`tossctl`의 조회 경로. a099가 호출자를 안 바꿨다.

## State mutations and fallbacks

- **이 함수는 원장을 안 바꾼다.** 연결을 열거나 안 열거나다.
- **실패는 전부 `db.Close()`와 함께 나간다** (B9·B10) — 연결이 새지 않는다.
- **폴백 없음.**

## Safety conclusion

- **Safe edit boundary**: **식별자 이름 하나.** 값(`5초`)도 분기도 그대로다.
  §5.6 실측: 분기 10 · 이탈 8로 base와 같다.
- **High-risk impact**: **no** — 읽기 전용 경로다. 이 함수가 실패하면 콘솔이
  조회를 못 할 뿐 엔진과 주문 경로는 영향이 없다.
  **다만 `DefaultBusyTimeout`은 High-risk 유도의 항이 되었다** —
  `obs.AlertDeliveryBound`가 그 값을 읽어 배달 상한을 만들고,
  그 상한이 임차 길이를 정한다. **상수의 값을 바꾸면 임차 길이가 따라 바뀐다.**
- **덮이지 않은 것을 이름으로 적는다**:
  - **B7의 「기본값이 쓰인다」쪽에 테스트가 없다.** `Open` 쪽 `BusyTimeout`과
    같은 공백이고 a099 이전부터 그렇다.
    **`not-applicable`: 이 change는 그 분기를 근거로 아무것도 주장하지 않는다.**
  - **B2·B8·B9에 테스트가 없다.** 경로 해결·드라이버 개방·핑 실패.
  - **`DefaultBusyTimeout`을 export한 것이 API 표면을 넓혔다.** 패키지 밖에서
    이 값을 읽는 곳은 오늘 `obs/alert_lease.go` 하나다.
