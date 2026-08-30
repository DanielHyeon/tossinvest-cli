# Function Logic Map: `TestCleanupIsNotAMeasuredStep`

- Source: `internal/verifylive/cleanup_test.go`
- Function: `internal/verifylive/cleanup_test.go:TestCleanupIsNotAMeasuredStep`
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
| 변경 여부 | 없음 | `git diff de14674974ab -- internal/verifylive/cleanup_test.go` | 이 함수 본문에 diff 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` (base line 228) — `if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {` | 없음 | 어서션 계속 | `TestCleanupIsNotAMeasuredStep` |
| B2 | `range` (base line 234) — `for i := range entries {` | 없음 | 어서션 계속 | `TestCleanupIsNotAMeasuredStep` |
| B3 | `if` (base line 235) — `if entries[i].StepID == StepCleanup {` | 없음 | 어서션 계속 | `TestCleanupIsNotAMeasuredStep` |
| B4 | `if` (base line 239) — `if cleanup == nil {` | 없음 | 어서션 계속 | `TestCleanupIsNotAMeasuredStep` |
| B5 | `if` (base line 242) — `if cleanup.Kind != KindCleanup {` | 없음 | 어서션 계속 | `TestCleanupIsNotAMeasuredStep` |
| B6 | `range` (base line 248) — `for _, e := range entries {` | 없음 | 어서션 계속 | `TestCleanupIsNotAMeasuredStep` |
| B7 | `if` (base line 249) — `if e.Kind == KindStep && e.StepID != StepCleanup {` | 없음 | 어서션 계속 | `TestCleanupIsNotAMeasuredStep` |
| B8 | `if` (base line 253) — `if stepsAfter != measured {` | 없음 | 어서션 계속 | `TestCleanupIsNotAMeasuredStep` |
| B9 | `range` (base line 257) — `for _, id := range RedoSet(entries) {` | 없음 | 어서션 계속 | `TestCleanupIsNotAMeasuredStep` |
| B10 | `if` (base line 258) — `if id == StepCleanup {` | 없음 | 어서션 계속 | `TestCleanupIsNotAMeasuredStep` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `withHolding` | ast.json calls (base line 223) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `newFakeBroker` | ast.json calls (base line 223) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `newHarness` | ast.json calls (base line 224) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `alwaysConfirm` | ast.json calls (base line 224) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `leftover` | ast.json calls (base line 225) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `StepCount` | ast.json calls (base line 227) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `h.entries` | ast.json calls (base line 227) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `h.run` | ast.json calls (base line 228) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `t.Logf` | ast.json calls (base line 229) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `t.Fatal` | ast.json calls (base line 240) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `t.Errorf` | ast.json calls (base line 243) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `RedoSet` | ast.json calls (base line 257) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `t.Error` | ast.json calls (base line 259) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |

라이브 바인딩 없음 — 테스트는 fake 브로커와 임시 디렉터리만 쓴다.

## State mutations and fallbacks

- 없음. 이 change는 이 함수의 상태·분기·부작용을 바꾸지 않았다.

## Safety conclusion

- Safe edit boundary: 편집하지 않음. base와 현재의 함수 본문이 동일하다.
- High-risk impact: no — 테스트 코드이며 이 change가 수정하지 않았다.
