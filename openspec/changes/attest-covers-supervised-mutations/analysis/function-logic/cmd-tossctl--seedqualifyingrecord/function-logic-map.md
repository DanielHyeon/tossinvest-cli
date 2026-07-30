# Function Logic Map: `seedQualifyingRecord`

- Source: `cmd/tossctl/soak_test.go`
- Function: `cmd/tossctl/soak_test.go:seedQualifyingRecord`
- AST evidence: `ast.json` — **base revision**(`8fff3aa44569`)에서 추출했다
- Risk scan: `risk-pattern-report.md`
- Change: `attest-covers-supervised-mutations`

**이 change는 이 함수를 수정하지 않았다.** 이 파일 끝에 새 테스트를 덧붙이면서 unified diff의
문맥 줄이 이 함수의 base 범위와 겹쳤을 뿐이다. 증거는 base revision으로 고정하고 아래 표는
base 소스에서 읽었다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 fixture 경로 | 임시 디렉터리 | 호출 테스트 | `t.Fatalf` |
| 변경 여부 | 없음 | `git diff 8fff3aa44569 -- cmd/tossctl/soak_test.go` | 이 함수 본문에 diff 없음 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` (base line 552) — `if err != nil {` | 테스트 fixture 작성 | 계속 | `seedQualifyingRecord` |
| B2 | `for` (base line 558) — `for day := 0; day < 3; day++ {` | 테스트 fixture 작성 | 계속 | `seedQualifyingRecord` |
| B3 | `for` (base line 559) — `for half := 0; half < 2; half++ {` | 테스트 fixture 작성 | 계속 | `seedQualifyingRecord` |
| B4 | `range` (base line 576) — `for _, e := range soak.RequiredEndpoints() {` | 테스트 fixture 작성 | 계속 | `seedQualifyingRecord` |
| B5 | `if` (base line 581) — `if err := rec.Append(c); err != nil {` | 테스트 fixture 작성 | 계속 | `seedQualifyingRecord` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `t.Helper` | ast.json calls (base line 550) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `soak.OpenRecorder` | ast.json calls (base line 551) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `t.Fatalf` | ast.json calls (base line 553) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `rec.Close` | ast.json calls (base line 555) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `Truncate` | ast.json calls (base line 557) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `Add` | ast.json calls (base line 557) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `UTC` | ast.json calls (base line 557) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `time.Now` | ast.json calls (base line 557) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `start.AddDate` | ast.json calls (base line 560) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `time.Duration` | ast.json calls (base line 560) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `at.Add` | ast.json calls (base line 565) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `soak.RequiredEndpoints` | ast.json calls (base line 576) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `append` | ast.json calls (base line 577) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |
| `rec.Append` | ast.json calls (base line 581) | 테스트 헬퍼/표준 라이브러리 — error는 t.Fatalf로 보고 | ast.json |

라이브 바인딩 없음 — 임시 디렉터리에 soak 기록을 쓸 뿐이다.

## State mutations and fallbacks

- 없음. 이 change는 이 함수를 바꾸지 않았다.

## Safety conclusion

- Safe edit boundary: 편집하지 않음. base와 현재의 함수 본문이 동일하다.
- High-risk impact: no — 테스트 헬퍼이며 이 change가 수정하지 않았다.
