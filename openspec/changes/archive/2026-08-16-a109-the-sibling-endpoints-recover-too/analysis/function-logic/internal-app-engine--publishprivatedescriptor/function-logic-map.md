# Function Logic Map: `publishPrivateDescriptor`

- Source: `internal/app/engine/alert_control_transport_unix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**descriptor에는 병이 없다.** 이 함수는 이미 stage(`os.CreateTemp(dir, ".endpoint-*")`)
+rename이고, rename은 원자적이라 최종 이름은 완성된 파일에만 붙는다(design "병의 재확인").
a109가 여기서 한 일은 **하나뿐**이다: staging 접두 `.endpoint-`를 상수
`privateDescriptorStagingPrefix`로 만들어 회수의 아는-이름 집합과 **같은 정의**를 쓰게 했다
(design D1b의 완전성 요구). 접두가 두 곳에 따로 적히면 회수가 자기 잔재를 낯선 것으로 거부한다.

descriptor 3벌 fold는 **선언된 생략**이다(design 말미) — 병이 없는 표면의 High-risk 리팩터링.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 유효한 control 디렉터리 안의 최종 descriptor 경로 | 호출자 | 디렉터리 검증 error |
| `body` | 직렬화된 descriptor JSON | 호출자 | 쓰기 실패 error |
| staging 이름 | `.endpoint-` + `os.CreateTemp`의 임의 숫자 | `os.CreateTemp` | 길이가 계약이 아니다(정규 파일이라 sun_path 무관) |
| staged 파일 | 0600 정규 파일 · 우리 uid · nlink 1 | `ValidatePrivateOpenFile`/`ValidatePrivateFile` | error |
| 발행 동일성 | staged와 published가 같은 inode | `os.SameFile` ×2 | error + rollback |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `ValidatePrivateControlDirectory(dir) != nil` | 없음 | "validating the control directory before staging" | 0700 아닌 디렉터리 |
| B2 | `os.CreateTemp(dir, privateDescriptorStagingPrefix+"*") != nil` | 없음 | "staging a descriptor" | **접두가 여기서 정해진다** — 회수의 아는-이름과 같은 상수 |
| B3 | `ValidatePrivateOpenFile != nil` | 파일 close | "validating a staged descriptor" | 커버 없음 |
| B4 | `temporary.Stat() != nil` | 파일 close | "inspecting a staged descriptor" | 커버 없음 |
| B5 | `writePositionPolicyDescriptorBody != nil` | temp 제거(defer) | "writing a descriptor" | chmod/write/short-write 오류 보존 |
| B6 | `ValidatePrivateFile(temporaryPath) != nil` | temp 제거(defer) | "validating a closed descriptor" | 커버 없음 |
| B7 | `Lstat err \|\| !SameFile(staged, closed)` | temp 제거 | "a staged descriptor changed before publication" | inode 교체 거부 |
| B8 | `os.Rename(temporaryPath, path) != nil` | temp 제거 | "publishing a descriptor" | 커버 없음 |
| B9 | (defer) `result == nil` | 없음 | 정상 종료면 rollback 없음 | 정상 발행 |
| B10 | (defer) `err == nil && SameFile(staged, published)` | **발행된 최종 이름 제거** | — | rename 뒤 실패해도 잔재를 남기지 않는다 |
| B11 | `Lstat err \|\| !SameFile(staged, published)` | rollback | "a published descriptor does not match the staged file" | 커버 없음 |
| B12 | `ValidatePrivateFile(path) != nil` | rollback | "validating a published descriptor" | 커버 없음 |
| B13 | `os.Open(dir) != nil` | rollback | "opening the control directory for sync" | 커버 없음 |
| B14 | `directory.Sync() != nil` | rollback + close | "syncing the control directory" | 커버 없음 |
| 분기 밖 종단 | — | — | `directory.Close()` | 정상 발행 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `positionpolicyrpc.ValidatePrivateControlDirectory` | 발행 전 디렉터리 위생 | error 전파 | AST calls |
| `os.CreateTemp` | staging 이름 생성 | error 전파 | AST calls |
| `positionpolicyrpc.ValidatePrivateOpenFile`/`ValidatePrivateFile` | 0600 정규 파일·소유·nlink | error 전파 | AST calls |
| `writePositionPolicyDescriptorBody` | chmod·write·sync·close | 짧은 쓰기는 `io.ErrShortWrite` | AST calls |
| `os.Lstat`·`os.SameFile` | inode 동일성 3회 확인 | error 전파 | AST calls |
| `os.Rename` | 원자적 발행 | error 전파 | AST calls |
| `os.Remove` (defer ×2) | staging 정리 · 실패 시 published 회수 | 오류 무시 | AST defers |
| `directory.Sync` | 디렉터리 엔트리 내구성 | error 전파 | AST calls |

## State mutations and fallbacks

- staging 파일 생성 → 내용 기록 → rename → 디렉터리 sync. 실패 경로마다 되돌린다:
  staging은 `defer os.Remove`, rename 후 실패는 두 번째 defer가 최종 이름을 회수한다.
- **crash로 죽으면 되돌리지 못한다** — 그 결과가 `.endpoint-*` staging 잔재이고,
  a109 회수가 그것을 아는 이름으로 다룬다.

## Safety conclusion

- Safe edit boundary: `os.CreateTemp`의 접두 리터럴을 상수 참조로 바꾼 것만. 검증·
  rollback·sync 순서는 그대로다.
- High-risk impact: yes — 기동 경로이고, 잘못 건드리면 토큰 파일의 위생이 무너진다.
