# Function Logic Map: `TestTheRemeasureScreenCarriesTheAdvisoryToo`

- Source: `internal/console/remeasure_test.go`
- Function: `internal/console/remeasure_test.go:TestTheRemeasureScreenCarriesTheAdvisoryToo`
- AST evidence: `ast.json` — **base revision**(`de14674974ab`)에서 추출했다
- Risk scan: `risk-pattern-report.md`
- Change: `verify-reopens-conditional-chain`

**이 change는 이 함수를 수정하지 않았다.** 이 파일 끝에 새 테스트를 덧붙이면서 unified diff의
문맥 줄이 이 함수의 base 범위와 겹쳤을 뿐이다. 그래서 증거는 base revision으로 고정하고,
아래 표는 base 소스에서 읽었다. 이 map은 "무엇이 바뀌었나"가 아니라 "겹친 함수가 무엇이며
왜 손대지 않았나"를 기록한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 fixture | 이 테스트가 만든 값 | base revision의 이 파일 | `t.Errorf`/`t.Fatalf` |
| 변경 여부 | 없음 | `git diff de14674974ab -- internal/console/remeasure_test.go` | 이 함수 본문에 diff 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` (base line 357) — `if !strings.Contains(page, "order-hours-closed") {` | 없음 | 어서션 계속 | `TestTheRemeasureScreenCarriesTheAdvisoryToo` |
| B2 | `if` (base line 360) — `if !strings.Contains(page, "재측정 1단계") {` | 없음 | 어서션 계속 | `TestTheRemeasureScreenCarriesTheAdvisoryToo` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `atKST` | ast.json calls (base line 349) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `newHarness` | ast.json calls (base line 350) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `seedVerdicts` | ast.json calls (base line 351) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `h.authenticate` | ast.json calls (base line 354) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `body` | ast.json calls (base line 356) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `h.get` | ast.json calls (base line 356) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `strings.Contains` | ast.json calls (base line 357) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `t.Errorf` | ast.json calls (base line 358) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `truncateForLog` | ast.json calls (base line 358) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |

라이브 바인딩 없음 — 테스트는 fake 브로커와 임시 디렉터리만 쓴다.

## State mutations and fallbacks

- 없음. 이 change는 이 함수의 상태·분기·부작용을 바꾸지 않았다.

## Safety conclusion

- Safe edit boundary: 편집하지 않음. base와 현재의 함수 본문이 동일하다.
- High-risk impact: no — 테스트 코드이며 이 change가 수정하지 않았다.
