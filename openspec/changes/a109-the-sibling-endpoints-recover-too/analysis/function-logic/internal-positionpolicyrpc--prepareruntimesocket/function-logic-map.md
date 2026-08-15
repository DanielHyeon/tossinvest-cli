# Function Logic Map: `PrepareRuntimeSocket`

- Source: `internal/positionpolicyrpc/runtime_unix.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

`PreparePrivateSocket`과 **같은 병의 다른 사본**이다(design 병 표): 정확-0600 요구가
pre-chmod 0700 잔재를 영구 거부하고, 생존 probe가 없어 산 주인의 socket을 unlink한다.
두 사본이 같은 병을 공유한다는 사실이 "이름-독립 기계 하나"라는 a109 D1b의 근거다.

a109는 이 함수를 편집하지 않고 runtime transport의 기동 경로에서 **뺀다**. 함수는 기존
공개 API·기존 테스트(`TestPrepareRuntimeSocketNeverDeletesNonSocket`)를 위해 남는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | leaf가 `runtime.sock` · 유효한 runtime control 디렉터리 안 | `validateRuntimeSocketPath` | error 전파 |
| 경로 부재 | 정상 | `errors.Is(err, os.ErrNotExist)` | `nil` |
| 잔재 모양 | socket · 비symlink · **정확 0600** | `info.Mode()` | "stale endpoint is not an exact 0600 Unix socket" |
| 소유 | 우리 uid | `validateOwnerAndLinks(info, false)` | error |
| 주인의 생사 | **검사하지 않는다** | — | 산 주인의 socket도 제거된다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `validateRuntimeSocketPath(path) != nil` | 없음 | 이름·디렉터리 error | 잘못된 socket 이름 거부 |
| B2 | `errors.Is(err, os.ErrNotExist)` | 없음 | `nil` | 첫 기동 |
| B3 | `err != nil` (그 밖의 Lstat 실패) | 없음 | error | — |
| B4 | `!socket \|\| symlink \|\| perm != 0o600` | 없음 | "stale endpoint is not an exact 0600 Unix socket" | **pre-chmod 0700 잔재의 영구 거부**(A1 F3) · 정규 파일 비삭제 |
| B5 | `validateOwnerAndLinks(info, false) != nil` | 없음 | error | 남의 소유 socket 비삭제(비root 재현 불가) |
| 분기 밖 종단 | 위 다섯 통과 | **`os.Remove(path)`** — 생존 확인 없는 제거 | `os.Remove` 결과 | a109 §1.2 RED가 탈취를 고정한다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `validateRuntimeSocketPath` | 이름 고정 + 디렉터리 위생 | error 전파 | AST calls |
| `os.Lstat` | symlink 비추적 stat | error 전파 | AST calls |
| `errors.Is` | ErrNotExist를 성공으로 | — | AST calls |
| `validateOwnerAndLinks` | 소유 uid | error 전파 | AST calls |
| `os.Remove` | 잔재 unlink | error 전파 | AST calls |

## State mutations and fallbacks

- 유일한 상태 변경은 마지막 `os.Remove`. fallback 없음.

## Safety conclusion

- Safe edit boundary: 본문 편집 없음 — 호출부에서 제거하고 회수 기계로 대체.
- High-risk impact: yes — 산 socket 제거는 조회 표면의 이중 기동을 만든다.
