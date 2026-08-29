# Function Logic Map: `Journal.PendingAlerts`

- Source: `internal/journal/outbox.go` (**517-530**)
- AST evidence: `ast.json` — AST 기준 branches 2 / returns 2 / defers 1 / go_statements 0
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `fb28da30fa8cc005d7d8dc9c8db0eb1a8c7bd18cfbebc3624465d19312145817`

**a098은 이 함수를 편집하지 않는다.** 이 map이 있는 이유는 a098의 배달 실행자가
이 함수를 **직접** 매 주기 부르게 되기 때문이고, 그 비용과 한계가 주기 선택을 지배하기 때문이다.

> **⛔ 5라운드 A-T4**: 이 문장은 *"`Flush`를 통해"*라고 적고 있었다.
> **2026-08-10 결정으로 `Flush`는 안 부른다**(design D1.1·D1.2) — 실행자가
> 자기 배달 경로에서 이 함수를 직접 부른다. 호출 경로가 달라도 **이 함수의 비용은 같으므로
> 이 map의 나머지는 그대로 유효하다.** 틀렸던 것은 경로 한 마디다.

## ✅ a099가 착지해서 다시 잰 것 (2026-08-12)

`check_analysis.py`가 `AST source hash is stale: internal/journal/outbox.go`로 막았다.
다시 뽑아 대조한 결과는 아래와 같다.

| | 1판 (base `285c7619`) | **재측정 (base `e6c4636a`)** | |
|---|---|---|---|
| 위치 | `:392-405` | **`:517-530`** | 줄만 밀렸다 |
| 분기 · 이탈 · 호출 · 대입 · defer | 2 · 2 · 5 · 5 · 1 | **2 · 2 · 5 · 5 · 1** | **전부 같다** |

**함수 본문은 안 바뀌었다.** 바뀐 것은 이 함수가 쓰는 **상수 하나**다 —
`alertSelect`(`:510-513`)에 a099가 **임차 세 열**(`claimed_by`·`claimed_at`·
`claim_expires_at`)을 더했다. 그것이 a099 task 4.10의 투영이고, **a098의 R9와 4.4
목록 명령이 기다리던 것**이다.

> **⚠ 같은 상수가 R9의 위험도 확정한다.** `alertSelect`는 **`body`와 `payload`를
> 여전히 싣고 온다**(`:510-511`). 즉 구조체를 그대로 출력하는 CLI는
> **알림 본문과 payload를 노출한다**(불변식 8). R9가 *"넷 전부"*를 보게 조인 것이
> 이 자리이고, **이제 그 위험이 추측이 아니라 착지한 코드에 있다.**

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `limit` | 임의 정수 | 호출자 | B1 `:520` — `limit > 0`일 때만 `LIMIT` 절이 붙는다. **0 이하는 전부** (선언 주석 `:515-516`) |
| `state = PENDING` | 고정 | `:518` `WHERE state = ?` | 필터가 상수다. DELIVERED·ACKNOWLEDGED는 안 나온다 |
| 정렬 | 고정 | `:518` `ORDER BY id` | **오래된 것 먼저.** 공정성 손잡이가 없다 |
| **임차 상태** | — | `alertSelect` `:512` | **필터가 아니다** — 임차된 행도 목록에 그대로 나온다 |

> **마지막 줄이 a098 4.0의 형태를 정한다.** 이 함수는 임차를 **투영만 하고 거르지 않는다.**
> 그래서 배달 실행자는 목록을 받은 뒤 **행마다 `ClaimAlertByID`로 갈라야** 하고,
> 그것이 4.0의 한 주기 형태가 `PendingAlerts → ClaimAlertByID → …`인 이유다.
> 목록 단계에서 거르는 설계는 **불가능하다** — 이 함수에 그 손잡이가 없다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error |
|---|---|---|---|
| B1 `:520` | `limit > 0` | 질의에 `LIMIT ?` 추가 | 이탈 없음 |
| B2 `:525` | `QueryContext` 오류 | 없음 | `:526` `return nil, err` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `db.QueryContext` | 행 읽기 | 오류가 B2로 | `:524` |
| `rows.Close` | 정리 | `defer` | `:528` |
| `scanAlerts` | 행 → `[]Alert` | 오류가 그대로 반환 | `:529` |

## State mutations and fallbacks

- 읽기 전용. 원장을 안 바꾼다.
- `alertSelect`(`:510-513`)가 `attempts`를 고르고 `Alert.Attempts int`가 받는다.
  **그래서 a092가 계획한 공정성 정렬은 이 함수를 안 고치고 Go에서 할 수 있다.**
- **B1이 a098의 배치 손잡이다.** 4.0b가 정하는 한 주기 배치 크기는 이 인자로 들어간다 —
  `Flush`가 쓰는 `0`(전부)과 **다른 값**이어야 하고, 그것이 a098이 잠금 없이
  도는 것과 함께 D1.1의 대가를 갚는 방식이다.

## Safety conclusion

- Safe edit boundary: **a098은 이 함수를 안 건드린다.** tasks 5.2가 `internal/journal`의
  diff가 비어 있음을 요구한다.
- High-risk impact: **yes** — 원장 읽기이지만 mutation이 없다. a098의 관심은 둘이다:
  **비용**(`limit=0`이면 매 주기 PENDING 전체)과 **노출**(`body`·`payload`가 실려 온다).
