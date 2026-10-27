# Function Logic Map: `fakeBuildRunner.rebuild`

- Source: `tools/a112-mb-us-source/main_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `f.binary` | verifier가 digest할 임의 bytes | 시험 fixture | 호출자 verifier가 결과를 판정 |
| `f.command` | 빈 값은 prescribed command, 비어 있지 않으면 주입한 command | 시험 fixture | 예상 밖 command는 verifier의 command 갈래로 전달 |
| `f.err` | nil 또는 주입 오류 | 시험 fixture | verifier의 rebuild 오류 갈래로 원문 전달 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `f.command == ""` | 지역 command를 prescribed command로 설정 | binary, prescribed command, injected error 반환 | `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 없음 | fake는 subprocess·네트워크·시간을 호출하지 않음 | error를 감싸거나 재시도하지 않고 그대로 반환 | AST calls 없음 |

## State mutations and fallbacks

- 지역 `command`만 변경하고 fixture 외 상태를 변경하지 않음. risk scan의 panic 지적은 대상 함수 밖 `errorReader` 행 2004·2007에 있으므로 무관함.

## Safety conclusion

- Safe edit boundary: zero-value fake의 기존 prescribed command 기본값을 보존하면서 command/error 주입만 추가.
- High-risk impact: 아니오. 시험 helper이며 production build command 또는 collector context를 바꾸지 않음.
