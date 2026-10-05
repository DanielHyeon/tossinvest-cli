# Function Logic Map: `newVerifyCmd`

- Source: `cmd/tossctl/verify.go` (96-129)
- Qualified function: `newVerifyCmd`
- Revision: `current` (구현 base `de147cc2`, `source_sha256` 77d33e4d…)
- AST evidence: `ast.json` — AST branches 0, 반환 1(128:2), 호출 6
- Risk scan: `risk-pattern-report.md`
- 편집 예정: 대사 하위 명령 등록(`cmd.AddCommand` 목록에 한 줄) + `Long` 도움말의 명령 목록 (task 3.2)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root *rootOptions` | 루트 명령의 전역 옵션(`--config-dir` 등), 하위 명령에 그대로 전달 | `cmd/tossctl/root.go:31` | 이 함수는 검사하지 않음 — 하위 명령 RunE 가 해석 |
| `opts := &verifyOptions{}` | 네 하위 명령이 **공유**하는 플래그 구조체 하나 | `cmd/tossctl/verify.go:74` | 공유이므로 새 하위 명령이 같은 구조체에 플래그를 붙이면 형제 명령과 필드 이름이 겹칠 수 있음 |

불변식: 그룹 명령 `verify` 자체에는 `RunE` 가 없다(잎이 아님). 잎 명령 집합과 각자의 `source`/`mutating`
주석은 시험 두 개가 고정한다 — `TestVerifyCommandsAreRegisteredAndAnnotated`(`verify_test.go:234`, run·status·report
의 source/mutating 대조)와 `TestMutatingAnnotationOnTradeCommands`(`help_convention_test.go:96`, 전 잎 명령의
`mutating=true` 집합이 **정확히** 목록과 같아야 함: `tossctl verify run`·`tossctl verify abort` 포함).

## Branches and early returns

AST 분기 0. 조기 반환 없음 — 단일 경로로 cobra 명령을 만들어 돌려준다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — `verify` 그룹을 만들고 하위 4개(run·status·report·abort)를 `AddCommand` | 명령 트리에 4 잎 추가 | `*cobra.Command` | `TestVerifyCommandsAreRegisteredAndAnnotated` · `TestMutatingAnnotationOnTradeCommands` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | `Long` 도움말 정리 | 실패 없음 | AST (102행) |
| `cmd.AddCommand` | 하위 명령 등록 | 실패 없음(cobra) | AST (122행) |
| `newVerifyRunCmd` | `verify run` — 주석 `source=official`, `mutating=true`(`verify.go:177`) | 생성만, 부작용 없음 | AST + CodeGraph callees |
| `newVerifyStatusCmd` | `verify status` — `source=local`(`verify.go:229`) | 생성만 | AST + CodeGraph callees |
| `newVerifyReportCmd` | `verify report` — `source=local`(`verify.go:258`) | 생성만 | AST + CodeGraph callees |
| `newVerifyAbortCmd` | `verify abort` — `source=official`, `mutating=true`(`verify.go:576`) | 생성만 | AST + CodeGraph callees |

호출자: CodeGraph 1.6.0 `callers newVerifyCmd` = `newRootCmd`(`root.go:52`) 1건; HEAD grep 호출 자리 `root.go:170` 1곳.

## State mutations and fallbacks

- 상태 변경 없음. 생성된 명령은 실행되기 전까지 아무것도 하지 않는다.
- 새 하위 명령을 더하면 `TestMutatingAnnotationOnTradeCommands` 가 그 잎의 `mutating` 주석을 판정한다 — 주석이
  `"true"` 이면 그 시험의 목록에 `tossctl verify reconcile`(가칭)을 더해야 하고, 아니면 목록에 없어야 한다.
  어느 쪽인지는 design·spec 이 아직 정하지 않았다(대사는 브로커 변이 없음 · 로컬 기록 추가 있음 — 선례:
  원장을 쓰는 운영자 명령 `engine mode-release`·`engine alerts ack` 는 `mutating=true`). Manager 결정 항목.
- `TestLeafCommandsHaveSourceAnnotation`(`help_convention_test.go:80`)이 새 잎에도 `source` 주석을 요구한다.

## Safety conclusion

- Safe edit boundary: `AddCommand` 목록에 한 줄 추가와 `Long` 의 명령 목록 한 줄 추가만. 기존 네 하위 명령의
  생성·순서·주석은 그대로.
- High-risk impact: no (등록만) — 단 등록되는 명령이 계좌 읽기와 기록 추가를 하므로, 그 명령 자체의 경계
  (flock·rate-budget lease·변이 메서드 도달 불가)는 새 함수의 RED 시험(task 2.4.1)이 진다.
