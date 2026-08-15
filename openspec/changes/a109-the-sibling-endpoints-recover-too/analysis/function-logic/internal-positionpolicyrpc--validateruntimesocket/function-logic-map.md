# Function Logic Map: `ValidateRuntimeSocket`

- Source: `internal/positionpolicyrpc/runtime_unix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

runtime endpoint의 **클라이언트 검증**이다(`DialRuntime`이 socket에 붙기 전에 부른다).
그리고 발행 직후 확인에도 쓰인다. a109의 회수 완화(`perm&0o077 == 0`)는 **여기 오지
않는다** — 발행이 stage+rename으로 바뀌면 최종 이름은 항상 0600이므로 정확-0600 요구가
오히려 계약의 증거가 된다(freeze P1-3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | leaf가 `runtime.sock` · 유효한 runtime control 디렉터리 안 | `validateRuntimeSocketPath` | error 전파 |
| 존재 | `os.Lstat` 성공 | 커널 | error 전파 |
| 모양·권한 | socket · 비symlink · **정확 0600** | `info.Mode()` | "endpoint is not an exact 0600 Unix socket" |
| 소유 | 우리 uid | `validateOwnerAndLinks(info, false)` | error |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `validateRuntimeSocketPath(path) != nil` | 없음 | 이름/디렉터리 error | descriptor의 socket 경로 주입 거부 |
| B2 | `os.Lstat` error | 없음 | 그 error | 죽은 endpoint에 붙지 않음 |
| B3 | `!socket \|\| symlink \|\| perm != 0o600` | 없음 | "endpoint is not an exact 0600 Unix socket" | **정확-0600 유지 핀**(a109 뮤테이션 M-C1) |
| 분기 밖 종단 | 통과 | 없음 | `validateOwnerAndLinks(info, false)` | 우리 소유 socket 통과 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validateRuntimeSocketPath` | 이름 고정 + 디렉터리 위생 | error 전파 | AST calls |
| `os.Lstat` | symlink 비추적 stat | error 전파 | AST calls |
| `validateOwnerAndLinks` | 소유 uid | error 전파 | AST calls |

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- Safe edit boundary: **편집하지 않는다.** a109 이후에도 runtime transport가 발행 직후
  이 함수로 0600을 확인한다 — 완화가 회수 밖으로 새지 않았다는 증거가 그 호출이다.
- High-risk impact: yes — 클라이언트가 붙는 경계다.
