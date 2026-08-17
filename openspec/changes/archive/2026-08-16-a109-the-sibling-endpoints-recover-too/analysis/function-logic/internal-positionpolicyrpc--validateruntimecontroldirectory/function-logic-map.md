# Function Logic Map: `ValidateRuntimeControlDirectory`

- Source: `internal/positionpolicyrpc/runtime.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

`ValidatePrivateControlDirectory`의 **이름-고정판**이다. 차이는 단 하나 — leaf 이름이
`.position-policy-runtime`이어야 한다는 요구이며, 나머지 검사는 같은
`validatePrivateDirectory(clean, true)`가 한다. a109 회수 기계는 이름-독립이므로 이
함수를 부르지 않지만, runtime transport는 회수 **후** 재생성한 디렉터리를 여전히 이
함수로 확인한다(이름 고정은 호출자의 몫이라는 계약 유지).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path`의 leaf 이름 | `RuntimeControlDirectoryName` (`.position-policy-runtime`) | `runtime.go` 상수 | "unexpected control directory name" |
| 부모(엔진 디렉터리) | 비symlink · group/other 쓰기 없음 · 우리 uid | `ValidateEngineDirectory` | error 전파 |
| leaf 디렉터리 | 실제 디렉터리 · 정확 0700 · 우리 uid | `validatePrivateDirectory(clean, true)` | error 전파 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `filepath.Base(clean) != RuntimeControlDirectoryName` | 없음 | "position policy runtime: unexpected control directory name" | descriptor 경로 주입 거부 |
| B2 | `ValidateEngineDirectory(filepath.Dir(clean)) != nil` | 없음 | 그 error | 안전하지 않은 엔진 디렉터리에서 기동 거부 |
| 분기 밖 종단 | 통과 | 없음 | `validatePrivateDirectory(clean, true)` | 0700 확인 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace`·`filepath.Base`·`filepath.Dir` | 이름·부모 산출 | — | AST calls |
| `ValidateEngineDirectory` | 부모 위생 | error 전파 | AST calls |
| `validatePrivateDirectory(_, true)` | leaf 정확-0700·소유 | error 전파 | AST calls |

## State mutations and fallbacks

- 상태 변경 없음. 순수 검증.

## Safety conclusion

- Safe edit boundary: **편집하지 않는다.** a109의 회수는 이름-독립 함수를 새로 쓰고,
  이 함수는 기존 호출부(발행 전·후 확인)에서 그대로 쓰인다.
- High-risk impact: yes — 기동 경로의 검증이다.
