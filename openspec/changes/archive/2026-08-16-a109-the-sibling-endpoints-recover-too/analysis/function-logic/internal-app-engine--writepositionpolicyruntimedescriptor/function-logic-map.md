# Function Logic Map: `writePositionPolicyRuntimeDescriptor`

- Source: `internal/app/engine/position_policy_runtime_transport_unix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

`publishPrivateDescriptor`의 이름-고정 사본이다(design "descriptor 발행 3벌"). **병은
없다** — 이미 stage(`os.CreateTemp(dir, ".position-policy-runtime-*")`)+rename이고,
`publishPrivateDescriptor`에 없는 절(B9: rename 직전 디렉터리 재검증, B13: 발행된 이름의
이름-고정 검증)까지 더 갖고 있다. 그 비대칭(P2-6)은 **선언된 생략**으로 남긴다.

a109가 여기서 한 일은 하나: staging 접두를 상수 `positionPolicyRuntimeStagingPrefix`로
만들어 회수의 아는-이름 집합과 같은 정의를 쓰게 했다(D1b 완전성).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | runtime control 디렉터리 안의 `endpoint.json` | 호출자 | 디렉터리·이름 검증 error |
| `descriptor` | socket 이름·토큰·PID | 호출자 | `json.Marshal` error(도달 불가에 가깝다) |
| staging 이름 | `.position-policy-runtime-` + `os.CreateTemp` 숫자 | `os.CreateTemp` | 정규 파일이라 길이는 계약이 아니다 |
| staged/published 동일성 | `os.SameFile` ×2 | 커널 | error + rollback |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `json.Marshal` error | 없음 | 그 error | 도달 불가에 가깝다 |
| B2 | `ValidateRuntimeControlDirectory(dir) != nil` | 없음 | "validating runtime directory before staging" | 0700 아닌 디렉터리 |
| B3 | `os.CreateTemp(dir, positionPolicyRuntimeStagingPrefix+"*") != nil` | 없음 | "staging runtime descriptor" | **접두가 여기서 정해진다** — 회수의 아는-이름과 같은 상수 |
| B4 | `ValidatePrivateOpenFile != nil` | close | "validating staged runtime descriptor" | 커버 없음 |
| B5 | `temporary.Stat() != nil` | close | "inspecting staged runtime descriptor" | 커버 없음 |
| B6 | `writePositionPolicyDescriptorBody != nil` | temp 제거(defer) | "writing runtime descriptor" | chmod/write/short-write |
| B7 | `ValidatePrivateFile(temporaryPath) != nil` | temp 제거 | "validating closed runtime descriptor" | 커버 없음 |
| B8 | `Lstat err \|\| !SameFile` | temp 제거 | "staged runtime descriptor changed before publication" | inode 교체 거부 |
| B9 | `ValidateRuntimeControlDirectory(dir) != nil` (rename 직전 재검증) | temp 제거 | "validating runtime directory before publication" | `publishPrivateDescriptor`에 없는 절 |
| B10 | `os.Rename != nil` | temp 제거 | "publishing runtime descriptor" | 커버 없음 |
| B11 | (defer) `result == nil` | 없음 | rollback 없음 | 정상 발행 |
| B12 | (defer) `err == nil && SameFile` | **발행된 최종 이름 제거** | — | rename 뒤 실패의 잔재 방지 |
| B13 | `ValidateRuntimePublishedDescriptor(path) != nil` | rollback | "validating published runtime descriptor" | 이름 고정 + 0600 |
| B14 | `Lstat err \|\| !SameFile` | rollback | "published runtime descriptor does not match staged file" | 커버 없음 |
| B15 | `os.Open(dir) != nil` | rollback | "opening runtime directory for sync" | 커버 없음 |
| B16 | `directory.Sync() != nil` | rollback + close | "syncing runtime directory" | 커버 없음 |
| B17 | `directory.Close() != nil` | rollback | "closing runtime directory" | 커버 없음 |
| 분기 밖 종단 | — | — | `nil` | 정상 발행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal` | descriptor 직렬화 | error 전파 | AST calls |
| `positionpolicyrpc.ValidateRuntimeControlDirectory` | 발행 전·직전 두 번 | error 전파 | AST calls |
| `os.CreateTemp` | staging 이름 | error 전파 | AST calls |
| `positionpolicyrpc.ValidatePrivateOpenFile`/`ValidatePrivateFile`/`ValidateRuntimePublishedDescriptor` | 0600 정규 파일·소유·nlink·이름 | error 전파 | AST calls |
| `writePositionPolicyDescriptorBody` | chmod·write·sync·close | `io.ErrShortWrite` 포함 | AST calls |
| `os.Rename` | 원자적 발행 | error 전파 | AST calls |
| `os.Remove` (defer ×2) | staging 정리 · 실패 시 published 회수 | 오류 무시 | AST defers |
| `directory.Sync`/`Close` | 디렉터리 엔트리 내구성 | error 전파 | AST calls |

## State mutations and fallbacks

- staging 생성 → 기록 → rename → sync. 실패마다 되돌린다. crash로 죽으면 되돌리지
  못하고 `.position-policy-runtime-*` 잔재가 남는다 — a109 회수가 아는 이름으로 다룬다.

## Safety conclusion

- Safe edit boundary: `os.CreateTemp` 접두 리터럴 → 상수 참조(완료). 그 밖의 절은 불변.
- High-risk impact: yes — 토큰 파일의 위생과 기동 성공 여부가 달린 경로다.
