# Function Logic Map: `TestTheReadOnlyHandleHasNoWriteMethods`

- Source: `internal/journal/readonly_test.go`
- AST evidence: `ast.json` (revision `current`)
- Risk scan: `risk-pattern-report.md` (매칭 0건)

기존 테스트 — 이 change는 allowlist 맵에 `"BrokerOrderIDs": true` **한 항목과 그 근거 주석**을
추가했다(diff `@@ -122,6 +122,10 @@`, 4줄 삽입·0줄 삭제). 단언 구조와 reflect 순회는 무변경이다.

이 테스트가 `internal/journal`의 read-only 보증의 **첫 번째 자물쇠**다: 런타임 거절
(`mode=ro` + `query_only`)이 두 번째이고, 이쪽은 "쓰는 문장에 닿을 메서드가 아예 없다"를
컴파일 대상 타입에서 열거로 확인한다. allowlist가 손으로 적혀 있다는 것이 요점이다 —
메서드를 하나 더 다는 일이 여기, 눈에 보이는 곳에서 내려지는 결정이 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `allowed` | 8개 메서드 이름 | 이 파일(손으로 유지) | 누락 시 `t.Errorf` |
| `reflect.TypeOf(&ReadOnly{})` | `*ReadOnly`의 exported 메서드 집합 | 현재 HEAD의 타입 | — |

불변식: allowlist에 없는 exported 메서드가 하나라도 있으면 실패한다. 즉 이 change가
`BrokerOrderIDs`를 추가하면서 여기를 갱신하지 않았다면 **테스트가 먼저 깨졌다.**

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (for, L131) | `i < typ.NumMethod()` 순회 | 없음 | — | 자체 실행 |
| B2 (if, L133) | `!allowed[name]` | 없음 | `t.Errorf` — 열거되지 않은 메서드 | 자체 실행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reflect.TypeOf` / `typ.Method(i).Name` | 메서드 집합 열거 | — | ast.json calls |
| `t.Errorf` | 미열거 메서드 보고 | — | ast.json calls |

DB도 파일도 열지 않는다 — 순수 타입 검사다.

## State mutations and fallbacks

- 없음(테스트). 계좌·원장·브로커에 닿지 않는다.

## Safety conclusion

- Safe edit boundary: allowlist 1항목 + 주석 3줄. 순회·단언 무변경.
- High-risk impact: **yes** — 이것은 원장 read-only 보증을 강제하는 가드 자체다.
  가드를 **넓히는** 편집이므로 방향이 문제다: 여기서 위험한 변경은 allowlist에 쓰기 메서드를
  넣거나 검사를 느슨하게 만드는 것인데, 이번 편집은 `SELECT DISTINCT` 한 건을 하는 읽기
  메서드 하나를 명시적으로 등록했을 뿐이고 그 근거를 주석으로 남겼다. 검사 로직은 그대로다.
