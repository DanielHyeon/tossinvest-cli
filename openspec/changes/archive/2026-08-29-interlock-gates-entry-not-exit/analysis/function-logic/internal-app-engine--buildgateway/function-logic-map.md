# Function Logic Map: `buildGateway`

- Source: `internal/app/engine/gateway.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## 이 change가 한 일

게이트웨이 생성에 `ProtectionOverrideForTest: in.protectionOverride` 한 줄이 추가됐다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| in.protectionOverride | nil 또는 포인터 | engine.Options (프로덕션은 nil) | nil이면 게이트웨이가 빌드 상수로 판정 |

## Branches and early returns

분기는 전부 기존 구성 오류 처리다. 추가된 것은 구조체 리터럴의 필드 하나.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if` @ internal/app/engine/gateway.go:201 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B2 | `if` @ internal/app/engine/gateway.go:223 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |
| B3 | `if` @ internal/app/engine/gateway.go:257 | 없음 | 해당 분기의 단언/거부 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| execgw.New | 게이트웨이 구성 | 오류를 감싸 반환 | 기존 |

## State mutations and fallbacks

- 없음 — 구성만 한다.

## Safety conclusion

- Safe edit boundary: 전달 한 줄. 이 값을 만들어내는 곳은 없고, execgw의 AST 단언이 그것을 증명한다.
- High-risk impact: yes
