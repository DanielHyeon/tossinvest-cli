# Function Logic Map: `verifyPrescribedBinary`

- Source: `tools/a112-mb-us-source/identity.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- 출처: `archive/2026-10-27-a172-console-anti-ai-redesign` 번들 복사, source sha 대조 일치 — a170 종결 주석(2026-10-28): a170 base `0d168e0a3` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum tools/a112-mb-us-source/identity.go` = `5512e4886462…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- 출처: `archive/2026-10-27-a176-overview-summary-and-verify-steps` 번들 복사, source sha 대조 일치 — a172 종결 주석(2026-10-28): a172 base `fd6afe23b` 가 그 change 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum tools/a112-mb-us-source/identity.go` = `5512e4886462…` 이 번들의 `source_sha256` 과 같아 번들을 그대로 복사함.
- a176 종결 주석(2026-10-27): a176 base `0e8ba1c7c` 가 a178 착지보다 앞이라 a176 창(base→HEAD)에 들어온 함수. a176 의 편집은 0 이고 소스 sha 가 a178 번들과 같아(`sha256sum` 대조: identity.go `5512e488…`, main_test.go `8008db70…`) a178 아카이브 번들(`openspec/changes/archive/2026-10-27-a178-a112-rebuild-failure-reported-separately/analysis/function-logic/`)을 그대로 복사함.
- 출처: `archive/2026-10-28-a170-inspection-guard-reconciles-offline-assessment/analysis/function-logic/tools-a112-mb-us-source--verifyprescribedbinary` 번들 복사, source sha 대조 일치 — a164 종결 주석(2026-10-28): a164 base `4cd7af19d` 가 그 함수를 고친 형제 착지보다 앞이라 창(base→워킹트리)에 들어온 함수. `sha256sum tools/a112-mb-us-source/identity.go` = `5512e4886462…` 이 번들의 `source_sha256`(revision current)과 같아 번들을 그대로 복사함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `snapshot.RepoRoot`, `snapshot.ExecutableSHA256` | 둘 다 비어 있지 않음 | pre-request identity snapshot | B1에서 rebuild 입력 unavailable 오류 반환 |
| `builder` | nil 아님 | `dependencies.builder`; 시험에서는 fake, 운영에서는 fixed system builder | B1에서 rebuild 입력 unavailable 오류 반환 |
| `snapshot.goExecution` | snapshot의 Go binary/toolchain SHA와 동일. 네 값 모두 비면 메모리 시험 builder만 허용 | receipt identity snapshot | B2에서 메모리 전용 zero identity로 대체, B3에서 결속 불일치 HOLD 오류 반환 |
| private Go distribution | Go binary path가 있으면 distribution digest와 재검증이 모두 일치 | `goBinaryIdentity` | B4/B5에서 distribution 또는 binary drift 오류 반환 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | repo root/executable digest가 없거나 builder가 nil | 없음 | `rebuild inputs are unavailable` | 기존 직접 시험 없음 |
| B2 | Go execution과 관련 snapshot 네 값이 모두 비었음 | 지역 `goBinary`를 zero identity로 대체; 메모리 fake만 허용 | 계속 진행 | `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch` |
| B3 | Go execution path/SHA가 snapshot의 Go binary/toolchain SHA와 다름 | 없음 | `selected Go execution snapshot does not bind receipt identity` | 기존 직접 시험 없음 |
| B4 | Go path가 있는데 private distribution binding이 없거나 digest 불일치 | 없음 | `private Go distribution snapshot does not bind receipt identity` | 기존 직접 시험 없음 |
| B5 | Go path가 있는데 SHA가 없거나 rebuild 전 재검증 실패 | 없음 | `selected Go binary changed before prescribed rebuild` | 기존 직접 시험 없음 |
| B6 | builder가 오류 반환 | `builder.rebuild` 한 번 호출 | `prescribed rebuild failed: %w`; 원래 오류 보존 | `TestVerifyPrescribedBinaryReportsRebuildFailureSeparately` |
| B7 | builder command가 prescribed command와 다름 | 없음 | `prescribed rebuild used an unexpected command` | `TestVerifyPrescribedBinaryReportsUnexpectedCommandSeparately` |
| B8 | rebuilt digest가 executable digest와 다름 | `digestBytes(rebuilt)` 계산 | `running executable does not match the prescribed rebuild` | `TestVerifyPrescribedBinaryRejectsExecutableHashMismatch` |
| B9 | Go path가 있는데 rebuild 후 재검증 실패 | 없음 | `selected Go binary changed after prescribed rebuild` | 기존 직접 시험 없음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `goBinary.revalidate` | selected Go binary의 pre/post drift 거부 | 오류를 새 오류 문구로 변환, retry 없음 | AST B5, B7; `identity.go:484,491` |
| `builder.rebuild` | fixed prescribed build의 bytes와 실제 명령 수집 | caller context(collector의 120초 deadline)를 그대로 전달, retry 없음 | AST B6; `identity.go:488` |
| `fmt.Errorf` | rebuild 오류 원문을 보존한 채 분류 | `%w`로 `errors.Is` 가능, retry 없음 | AST B6; `identity.go:490` |
| `digestBytes` | rebuilt bytes를 running executable SHA와 비교 | 오류 없음 | AST B8; `identity.go:495` |
| `errors.New` | fail-closed identity 오류 생성 | 네트워크·receipt 쓰기·fallback 없음 | AST B1, B3~B5, B7~B9 |

## State mutations and fallbacks

- 지역 `goBinary` 외 상태 변경 없음. 이 함수는 reader, receipt, 네트워크, deadline 생성 또는 HOLD 변환을 수행하지 않음.
- 메모리 builder fallback은 네 Go identity 값이 모두 빈 시험 경로로 한정되며, production Go binary의 결속 검사를 우회하지 않음.

## Safety conclusion

- Safe edit boundary: 기존 B6의 오류 분류만 B6~B8로 분리. `builder.rebuild` 호출 횟수, caller context, 120초 collector deadline, B1~B5/B9 및 상위 `run`의 HOLD 변환은 불변.
- High-risk impact: 아니오. production evidence authority와 분리된 측정 도구의 진단 문구·원인 보존 변경이며 주문·위험·원장·운영 토글에 닿지 않음.
