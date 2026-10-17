# Function Logic Map: `Notifier.Acknowledge`

- Source: `internal/obs/notifier.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> **a099는 이 함수를 편집하지 않는다.** 이 산출물이 있는 이유는 반대다 —
> **design C6이 이 함수 안의 분기 둘을 근거로 쓴다.**
> *"오늘 게이트를 푸는 자리는 `:481-482`와 `:510-511`이고, a099는 그 둘을 안 건드린다."*
> `.claude/CLAUDE.md`의 규칙은 편집 여부가 아니라 **「근거로 쓰는가」**에 걸린다.
>
> **이 산출물은 3라운드에 만들어졌다.** 2판은 이 함수를 *"게이트를 유도로 평가한다"*며
> **편집 대상**으로 D5에 넣었다. 사용자 결정 5-1이 그 편집을 없앴고, 그래도
> **인용은 남아서** 산출물이 필요하다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `operator` | **공백이 아니어야 한다** | 호출자 | B1 `:477`이 오류로 거절한다 — 승인은 **이름 있는 사람**의 행위다 |
| `ids` | 비었으면 **밀린 것 전부** | 호출자 | 비면 B4 `:490`이 `PendingAlerts`로 목록을 만든다 |
| `n.Journal` | nil일 수 있다 | 배선 | B2 `:480` — 원장이 없으면 **게이트만 풀고 끝낸다** |
| `n.Gate` | nil일 수 있다 | 배선 | B3 `:481`·B10 `:510`이 각각 검사한다 |
| `n.mu` | 이 함수가 **잡는다** (`:487` Lock, `:488` defer Unlock) | — | 발송과 승인이 **같은 잠금**을 공유한다 |

**불변식 — a099가 지켜야 하는 것**: *"게이트를 푸는 조건은 **운영자의 승인 + 남은 수 0**
둘 다이고, 어느 하나만으로는 안 풀린다."*

이것은 승인된 정본 `openspec/specs/engine-safety/spec.md:147-152`가 요구하는 것이고,
**`TestRecoveredDeliveryDoesNotReleaseTheGateByItself`(`obs_test.go:427`)가 이미
단언하고 있다** — 전달이 복구돼도 게이트는 안 풀린다.

> **⛔⛔ a099의 3판이 이 불변식을 깼다 — 3라운드 A-P1.**
> 3판 C6은 *"미전달 수가 0이 되면 사람 개입 없이 풀린다"*를 규범으로 적었다.
> 그러면 위 테스트가 **거짓이 되어야 한다.** 승인된 규범을 `MODIFIED` 없이 뒤집는
> 변경이었고, **이 함수의 산출물이 있었다면 더 일찍 잡혔을 자리다.**

## Branches and early returns

`ast.json`의 열거를 그대로 쓴다 — 분기 10 · 이탈 6 · 호출 12 · defer 1.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:477` | `strings.TrimSpace(operator) == ""` | 없음 | 이탈 `:478` 오류 | 기존 — `TestAcknowledgeRequiresAnIdentity` |
| B2 `:480` | `n.Journal == nil` | — | (B3로) | **없음 — nil 원장 경로에 테스트가 없다** |
| B3 `:481` | `n.Gate != nil` | **`Gate.Clear`** `:482` | 이탈 `:484` `nil` | **없음 — 같음** |
| B4 `:490` | `len(ids) == 0` | — | (B5로) | 기존 — `TestRecoveredDeliveryDoesNotReleaseTheGateByItself` |
| B5 `:492` | `PendingAlerts` 오류 | 없음 | 이탈 `:493` 오류 | **없음 — DB 오류 주입 없음** |
| B6 `:495` | `range pending` | `ids`에 append `:496` | — | 기존 |
| B7 `:499` | `range ids` | — | — | 기존 |
| B8 `:500` | `AcknowledgeAlert` 오류이고 `ErrAlertNotFound`가 **아니다** | 없음 | 이탈 `:502` 오류 | **없음 — 없는 id는 조용히 넘어간다** |
| B9 `:507` | `UndeliveredCount` 오류 | 없음 | 이탈 `:508` 오류 | **없음** |
| **B10 `:510`** | **`remaining == 0 && n.Gate != nil`** | **`Gate.Clear`** `:511` | (이탈 `:513` `nil`) | 기존 — `TestAcknowledgeWhileStillPendingKeepsTheBlock`(거짓 쪽) · `TestRecoveredDeliveryDoesNotReleaseTheGateByItself`(참 쪽) |
| — 이탈 `:513` | 정상 | — | `nil` | 기존 |

**게이트를 푸는 자리는 둘이고 둘 다 조건이 붙어 있다** — B3(원장이 없다)과
B10(**승인했고 남은 수가 0이다**). a099는 **둘 다 안 건드린다**(design C6).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `n.Gate.Clear` `:482`·`:511` | 진입 차단 해제 | 오류 없음 — 메모리 맵 (`execgw/retry.go:495-510`) | AST |
| `n.mu.Lock`/`Unlock` `:487`·`:488` | 발송과의 배제 | — | AST (defer 1) |
| `n.Journal.PendingAlerts` `:491` | 밀린 목록 | 오류면 이탈 `:493` | AST |
| `n.Journal.AcknowledgeAlert` `:500` | 행을 ACKNOWLEDGED로 | `ErrAlertNotFound`는 **삼킨다** | AST |
| `n.Journal.UndeliveredCount` `:506` | 해제 판정의 근거 | 오류면 이탈 `:508` | AST |

**`:487`의 잠금이 이 함수의 안전 논거다.** `:470-475`의 doc이 그 이유를 적는다 —
count와 clear 사이에 재무장이 끼면 *"미전달 critical이 있는데 게이트가 열린다."*
**a099는 이 구간을 안 줄인다.**

## State mutations and fallbacks

- 승인은 **행 단위**로 일어나고, 없는 행은 조용히 넘어간다 (B8의 `errors.Is`).
- 해제는 **전체 판정**이다 — 한 행을 승인해도 다른 미전달이 있으면 안 풀린다(B10).
- `n.Journal`이 nil이면 **원장을 안 보고 푼다**(B3). 배선이 없는 빌드의 경로다.
- **a099가 더하는 것**: `AcknowledgeAlert`의 술어가 임차를 **안 보므로**(design C5)
  발송 중인 행도 승인된다. 그때 발송자의 정산은 `SettleAlreadySettled`가 되고,
  **그것은 오류가 아니다**(design C3). 이 함수 자체는 안 바뀐다.

## Safety conclusion

- **Safe edit boundary**: **없다 — a099는 이 함수를 편집하지 않는다.**
  §5.3b가 `:481-482`와 `:510-511`의 diff가 **비어 있음**을 확인한다.
  하나라도 바뀌면 승인된 규범을 건드린 것이고 `MODIFIED` 없이는 못 나간다.
- **High-risk impact**: **yes** — 진입 게이트(불변식 5). 이 함수가 **유일하게**
  `ReasonAlertUndelivered`를 정상 경로에서 푸는 자리다.
- **덮이지 않은 것을 이름으로 적는다**:
  - **B2·B3·B5·B8·B9에 테스트가 없다.** nil 원장, DB 오류 주입, 없는 id 경로.
    **a099는 이 다섯을 안 건드리므로 `not-applicable`**이지만, 안 덮였다는 사실은
    적어 둔다 — 침묵한 생략은 금지다.
  - **기동 직후 이 함수가 안 불리면 게이트는 비어 있다.** 그것이 C6의 두 번째
    변경(기동 복원)이 필요한 이유이고, **a098이 진다** — `not-applicable`: a099 밖이다.
