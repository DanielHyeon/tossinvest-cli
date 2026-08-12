# Function Logic Map: `Journal.PendingAlerts`

- Source: `internal/journal/outbox.go` (392-405)
- AST evidence: `ast.json` — AST 기준 branches 2 / returns 2 / defers 1 / go_statements 0
- Risk scan: `risk-pattern-report.md` (매치 없음)

**a098은 이 함수를 편집하지 않는다.** 이 map이 있는 이유는 a098의 배달 실행자가
이 함수를 **직접** 매 주기 부르게 되기 때문이고, 그 비용과 한계가 주기 선택을 지배하기 때문이다.

> **⛔ 5라운드 A-T4**: 이 문장은 *"`Flush`를 통해"*라고 적고 있었다.
> **2026-08-10 결정으로 `Flush`는 안 부른다**(design D1.1·D1.2) — 실행자가
> 자기 배달 경로에서 이 함수를 직접 부른다. 호출 경로가 달라도 **이 함수의 비용은 같으므로
> 이 map의 나머지는 그대로 유효하다.** 틀렸던 것은 경로 한 마디다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `limit` | 임의 정수 | 호출자 | B1 `:395` — `limit > 0`일 때만 `LIMIT` 절이 붙는다. **0 이하는 전부** (선언 주석 `:390`) |
| `state = PENDING` | 고정 | `:393` `WHERE state = ?` | 필터가 상수다. DELIVERED·ACKNOWLEDGED는 안 나온다 |
| 정렬 | 고정 | `:393` `ORDER BY id` | **오래된 것 먼저.** 공정성 손잡이가 없다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 `:395` | `limit > 0` | 질의에 `LIMIT ?` 추가 | 이탈 없음 | **없음 — 아래가 이 map의 핵심이다** |
| B2 `:400` | `QueryContext` 오류 | 없음 | `:401` `return nil, err` | 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `db.QueryContext` | 행 읽기 | 오류가 B2로 | `:399` |
| `rows.Close` | 정리 | `defer` | `:403` |
| `scanAlerts` | 행 → `[]Alert` | 오류가 그대로 반환 | `:404` |

## State mutations and fallbacks

- 읽기 전용. 원장을 안 바꾼다.
- `alertSelect`(`:386-388`)가 `attempts`를 고르고 `Alert.Attempts int`(`:95`)가 받는다.
  **그래서 a092가 계획한 공정성 정렬은 이 함수를 안 고치고 Go에서 할 수 있다.**

## Safety conclusion

- Safe edit boundary: **a098은 이 함수를 안 건드린다.**
- High-risk impact: **yes** — 원장 읽기이지만 mutation이 없다. a098의 관심은
  **비용**이다: `Flush`가 `limit=0`으로 부르므로 매 주기 PENDING 전체를 읽는다.
