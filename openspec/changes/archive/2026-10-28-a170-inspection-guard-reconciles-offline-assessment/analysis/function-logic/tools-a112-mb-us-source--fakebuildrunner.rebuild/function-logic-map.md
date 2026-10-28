# Function Logic Map: `fakeBuildRunner.rebuild`

- Source: `tools/a112-mb-us-source/main_test.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum tools/a112-mb-us-source/main_test.go` = `8008db70f470…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 출처: `archive/2026-10-27-a176-overview-summary-and-verify-steps` 번들 복사, source sha 대조 일치 — a172 종결 주석(2026-10-28): a172 base `fd6afe23b` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum tools/a112-mb-us-source/main_test.go` = `8008db70f470…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- a176 종결 주석(2026-10-27): a176 base `0e8ba1c7c` 가 a178 착지보다 앞이라 a176 창(base→HEAD)에 들어온 함수. a176 의 편집은 0 이고 소스 sha 가 a178 번들과 같아(`sha256sum` 대조: identity.go `5512e488…`, main_test.go `8008db70…`) a178 아카이브 번들(`openspec/changes/archive/2026-10-27-a178-a112-rebuild-failure-reported-separately/analysis/function-logic/`)을 그대로 복사함.

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
