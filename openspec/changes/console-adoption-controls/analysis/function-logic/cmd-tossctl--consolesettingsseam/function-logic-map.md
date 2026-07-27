# Function Logic Map: `consoleSettingsSeam`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (구현 후 추출)
- Risk scan: `risk-pattern-report.md`

신규 leaf — 구체 seam의 typed-nil이 인터페이스에 실리지 않게 하는 어댑터.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (함수 서술 참조 — 이 change의 delta·design D1~D6이 계약) | — | 현재 HEAD + 위 테스트 | 테스트 실패로 관측 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if | 없음 | — | cmd/tossctl 빌드·기존 테스트(import 가드 포함) green |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ast.json calls 참조 | 목적 서술 참조 | 기존 계약 무변경 | ast.json + HEAD |

## State mutations and fallbacks

- 없음

## Safety conclusion

- Safe edit boundary: nil 가드 + 인터페이스 변환만
- High-risk impact: no (콘솔·config·배선 — 주문·원장 무접촉)
