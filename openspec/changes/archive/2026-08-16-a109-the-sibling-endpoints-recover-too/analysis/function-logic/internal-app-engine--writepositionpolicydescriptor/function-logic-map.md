# Function Logic Map: `writePositionPolicyDescriptor`

- Source: `internal/app/engine/position_policy_transport.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

descriptor 발행 3벌 중 command endpoint 판이다. 절 구성은
`writePositionPolicyRuntimeDescriptor`와 동일하고 이름만 다르다(`.position-policy-control-*`
접두, `ValidatePublishedDescriptor`). **병은 없다** — 이미 stage+rename이다.

a109가 한 일은 하나: staging 접두를 상수 `positionPolicyControlStagingPrefix`로 만들어
command endpoint의 staging 위생(design D2a)이 **같은 정의**를 쓰게 했다. 접두가 두 곳에
따로 적히면 위생이 자기 잔재를 못 알아본다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | control 디렉터리 안의 `endpoint.json` | 호출자 | 디렉터리·이름 검증 error |
| `descriptor` | 주소·토큰·PID | 호출자 | `json.Marshal` error |
| staging 이름 | `.position-policy-control-` + `os.CreateTemp` 숫자 | `os.CreateTemp` | 정규 파일이라 길이는 계약이 아니다 |
| staged/published 동일성 | `os.SameFile` ×2 | 커널 | error + rollback |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `json.Marshal` error | 없음 | 그 error | 도달 불가에 가깝다 |
| B2 | `ValidateControlDirectory(dir) != nil` | 없음 | "validating control directory before staging" | 0750 디렉터리 거부 |
| B3 | `os.CreateTemp(dir, positionPolicyControlStagingPrefix+"*") != nil` | 없음 | "staging control descriptor" | **접두가 여기서 정해진다** — 위생과 같은 상수 |
| B4 | `ValidatePrivateOpenFile != nil` | close | "validating staged control descriptor" | 커버 없음 |
| B5 | `temporary.Stat() != nil` | close | "inspecting staged control descriptor" | 커버 없음 |
| B6 | `writePositionPolicyDescriptorBody != nil` | temp 제거(defer) | "writing control descriptor" | chmod/write/short-write |
| B7 | `ValidatePrivateFile(temporaryPath) != nil` | temp 제거 | "validating closed control descriptor" | 커버 없음 |
| B8 | `Lstat err \|\| !SameFile` | temp 제거 | "staged control descriptor changed before publication" | inode 교체 거부 |
| B9 | `ValidateControlDirectory(dir) != nil` (rename 직전 재검증) | temp 제거 | "validating control directory before publication" | 커버 없음 |
| B10 | `os.Rename != nil` | temp 제거 | "publishing control descriptor" | 커버 없음 |
| B11 | (defer) `result == nil` | 없음 | rollback 없음 | 정상 발행 |
| B12 | (defer) `err == nil && SameFile` | **발행된 최종 이름 제거** | — | rename 뒤 실패의 잔재 방지 |
| B13 | `ValidatePublishedDescriptor(path) != nil` | rollback | "validating published control descriptor" | 이름 고정 + 0600 |
| B14 | `Lstat err \|\| !SameFile` | rollback | "published control descriptor does not match staged file" | 커버 없음 |
| B15 | `os.Open(dir) != nil` | rollback | "opening control directory for sync" | 커버 없음 |
| B16 | `directory.Sync() != nil` | rollback + close | "syncing control directory" | 커버 없음 |
| B17 | `directory.Close() != nil` | rollback | "closing control directory" | 커버 없음 |
| 분기 밖 종단 | — | — | `nil` | 정상 발행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `json.Marshal` | 직렬화 | error 전파 | AST calls |
| `positionpolicyrpc.ValidateControlDirectory` | 발행 전·직전 두 번 | error 전파 | AST calls |
| `os.CreateTemp` | staging 이름 | error 전파 | AST calls |
| `positionpolicyrpc.ValidatePrivateOpenFile`/`ValidatePrivateFile`/`ValidatePublishedDescriptor` | 0600 정규 파일·소유·nlink·이름 | error 전파 | AST calls |
| `writePositionPolicyDescriptorBody` | chmod·write·sync·close | `io.ErrShortWrite` 포함 | AST calls |
| `os.Rename` | 원자적 발행 | error 전파 | AST calls |
| `os.Remove` (defer ×2) | staging 정리 · 실패 시 published 회수 | 오류 무시 | AST defers |
| `directory.Sync`/`Close` | 디렉터리 엔트리 내구성 | error 전파 | AST calls |

## State mutations and fallbacks

- staging 생성 → 기록 → rename → sync. 정상·실패 경로는 전부 되돌린다.
  **crash만 되돌리지 못하고** `.position-policy-control-*` 잔재를 남긴다 — a109 §1.4의
  staging 위생이 그것을 받는다.

## Safety conclusion

- Safe edit boundary: `os.CreateTemp` 접두 리터럴 → 상수 참조(완료). 그 밖의 절은 불변.
- High-risk impact: yes — 격리 해제 표면의 토큰 파일이다.
