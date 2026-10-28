# Function Logic Map: `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch`

- Source: `tools/a112-mb-us-source/main_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum tools/a112-mb-us-source/main_test.go` = `8008db70f470…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 출처: `archive/2026-10-27-a176-overview-summary-and-verify-steps` 번들 복사, source sha 대조 일치 — a172 종결 주석(2026-10-28): a172 base `fd6afe23b` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum tools/a112-mb-us-source/main_test.go` = `8008db70f470…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- a176 종결 주석(2026-10-27): a176 base `0e8ba1c7c` 가 a178 착지보다 앞이라 a176 창(base→HEAD)에 들어온 함수. a176 의 편집은 0 이고 소스 sha 가 a178 번들과 같아(`sha256sum` 대조: identity.go `5512e488…`, main_test.go `8008db70…`) a178 아카이브 번들(`openspec/changes/archive/2026-10-27-a178-a112-rebuild-failure-reported-separately/analysis/function-logic/`)을 그대로 복사함.
- 출처: `archive/2026-10-28-a170-inspection-guard-reconciles-offline-assessment/analysis/function-logic/tools-a112-mb-us-source--testverifyprescribedbinaryrejectsexecutablehashmismatch` 번들 복사, source sha 대조 일치 — a164 종결 주석(2026-10-28): a164 base `4cd7af19d` 가 그 함수를 고친 형제 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum tools/a112-mb-us-source/main_test.go` = `8008db70f470…` 이 번들의 `source_sha256`(revision current)과 같아 번들을 그대로 복사함.

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
