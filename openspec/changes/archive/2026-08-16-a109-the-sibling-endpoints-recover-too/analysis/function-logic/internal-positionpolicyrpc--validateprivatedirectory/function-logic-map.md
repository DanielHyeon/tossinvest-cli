# Function Logic Map: `validatePrivateDirectory`

- Source: `internal/positionpolicyrpc/client.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

> scaffold의 `Source:`는 `private_fs_unix.go`로 적혀 있었으나 `ast.json`이 가리키는
> 실제 정의 파일은 `client.go`다. AST를 정본으로 정정했다.

control 디렉터리 위생의 **바닥**이다. 이름-고정판 셋(`ValidateControlDirectory`,
`ValidateRuntimeControlDirectory`, `ValidatePrivateControlDirectory`)이 전부 여기로
모이므로, 여기를 완화하면 세 endpoint의 경계가 한 번에 넓어진다. a109가 디렉터리 perm
완화를 **두지 않기로 한** 근거가 이 수렴이다(design D2, P1-7①).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 공백 제거 후 비어 있지 않음 | 호출자 | "private directory path is empty" |
| symlink traversal | 절대경로 == EvalSymlinks 결과 | `rejectSymlinkTraversal` | "symlink traversal is forbidden" |
| 존재·모양 | 실제 디렉터리, symlink 아님 | `os.Lstat` + `info.Mode()` | "private directory is not a real directory" |
| 소유 | 우리 uid | `validateOwnerAndLinks(info, false)` | "wrong owner" |
| `exact == true` (control leaf) | perm **정확 0700** | 호출자 | "control directory mode is %04o, want 0700" |
| `exact == false` (엔진 디렉터리) | group/other **쓰기** 비트 없음 | 호출자 | "engine directory is writable by group or other" |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `clean == ""` | 없음 | "private directory path is empty" | 커버 없음 |
| B2 | `rejectSymlinkTraversal(clean) != nil` | 없음 | 그 error | symlink control 디렉터리 거부 |
| B3 | `os.Lstat` error | 없음 | 그 error | 없는 디렉터리 |
| B4 | `symlink \|\| !IsDir` | 없음 | "not a real directory" | 디렉터리 자리의 이물 거부 |
| B5 | `validateOwnerAndLinks(info, false) != nil` | 없음 | 그 error | 남의 uid 디렉터리 거부(비root 재현 불가) |
| B6 | `exact && perm != 0o700` | 없음 | mode 불일치 error | 0750 control 디렉터리 거부 |
| B7 | `!exact && perm&0o022 != 0` | 없음 | "writable by group or other" | 0770 엔진 디렉터리 거부 |
| 분기 밖 종단 | 전부 통과 | 없음 | `nil` | 정상 기동 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | 정규화 | — | AST calls |
| `rejectSymlinkTraversal` | 경로 전체의 symlink 거부 | error 전파 | AST calls |
| `os.Lstat` | symlink 비추적 stat | error 전파 | AST calls |
| `info.Mode()`·`Perm()` | 모양·권한 | — | AST calls |
| `validateOwnerAndLinks` | 소유 uid | error 전파 | AST calls |
| `fmt.Errorf`·`errors.New` | 사유 문자열 | — | AST calls |

## State mutations and fallbacks

- 상태 변경 없음. 완화·재시도·fallback 모두 없다.

## Safety conclusion

- Safe edit boundary: **편집하지 않는다.** a109는 회수 대상 **파일**의 판정만 완화하고
  디렉터리 판정은 그대로 둔다. owner 비트를 깎는 umask 같은 디렉터리 perm 변형은 환경
  이상으로 분류되어 D3 강등·보고가 받는다.
- High-risk impact: yes — 세 endpoint의 접근 경계가 이 함수 하나로 수렴한다.
