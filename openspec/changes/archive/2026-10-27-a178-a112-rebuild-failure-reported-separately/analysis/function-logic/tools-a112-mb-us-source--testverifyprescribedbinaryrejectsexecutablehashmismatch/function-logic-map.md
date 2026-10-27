# Function Logic Map: `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch`

- Source: `tools/a112-mb-us-source/main_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | Go testing harness | `go test` | `Fatal`/`Fatalf`로 시험 실패 |
| identity snapshot | running bytes의 digest와 `rebuilt` bytes가 다름 | 시험 fixture | verifier가 digest mismatch 오류를 반환해야 함 |
| `fakeBuildRunner.binary` | `rebuilt` | 시험 fixture | prescribed command 기본값을 유지 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | verifier 오류가 nil | temp directory만 생성됨 | `t.Fatal` | 이 함수 자체 |
| B2 | 오류 문구에 digest mismatch 리터럴이 없음 | 없음 | `t.Fatalf` | 이 함수 자체 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyPrescribedBinary` | digest 불일치 갈래의 공개 오류 계약 고정 | 오류가 nil이면 B1, 문구가 다르면 B2; retry 없음 | AST `main_test.go:1100` |
| `t.TempDir` | isolated repo root fixture 제공 | Go test가 정리 | AST `main_test.go:1100` |
| `digestBytes` | expected running digest 생성 | 오류 없음 | AST `main_test.go:1100` |

## State mutations and fallbacks

- production 상태·네트워크·receipt 변경 없음. risk scan의 panic 지적은 대상 함수 밖 `errorReader` 행 2004·2007에 있으므로 이 함수의 recovery boundary가 아님.

## Safety conclusion

- Safe edit boundary: 기존 nil 오류 단언 뒤 digest 문구 단언 한 줄 추가.
- High-risk impact: 아니오. 시험 전용이며 production 판정 또는 시한을 변경하지 않음.
