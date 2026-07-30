# Function Logic Map: `Console.routes`

- Source: `internal/console/console.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

라우트 둘을 등록했다 — /settings/trading과 /settings/gate. 둘 다 session+CSRF 이중 게이트.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| mux | *http.ServeMux | 호출자 | 중복 등록은 런타임 panic |

## Branches and early returns

분기 없는 등록이다. 분기는 전부 기존 라우트의 조건부 등록이다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — happy path | 없음 | 정상 반환 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| c.session0(c.mutating(...)) | 상태변경 라우트의 필수 이중 게이트 | 오류 없음 | static_test.go가 단언 |

## State mutations and fallbacks

- mux에 핸들러 둘을 등록한다.

## Safety conclusion

- Safe edit boundary: 등록 두 줄. 기존 라우트는 무수정이고, 새 둘은 static_test의 상태변경 목록에 같은 커밋에서 들어갔다.
- High-risk impact: yes
