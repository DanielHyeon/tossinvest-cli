# Function Logic Map: `statPrivateSocket`

- Source: `internal/positionpolicyrpc/private_endpoint.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

`ValidatePrivateSocket`(클라이언트·발행 확인)과 `PreparePrivateSocket`(회수)이 **공유하는**
검사다. freeze P1-3이 지적한 위험이 바로 이 공유다: 회수를 완화하려고 이 함수의
`perm != 0o600`을 손대면 **클라이언트 검증까지 함께 넓어진다.**

a109의 결정: 이 함수는 그대로 두고, 완화(`perm&0o077 == 0`)는 회수 전용 함수에만 새로 쓴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `path` | 공백 제거 후 비어 있지 않음 | 호출자 두 곳 | 빈 문자열이면 즉시 error |
| `filepath.Dir(path)` | 유효한 private control 디렉터리 | `ValidatePrivateControlDirectory` | error 전파 |
| 경로 존재 | `os.Lstat` 성공 | 커널 | ErrNotExist 포함 error 그대로 — `PreparePrivateSocket`이 ErrNotExist만 성공으로 읽는다 |
| 모양·권한 | socket 비트 있음 · symlink 아님 · **정확 0600** | `info.Mode()` | `errors.New("private endpoint: path is not an exact 0600 Unix socket")` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `clean == ""` | 없음 | "socket path is empty" | 커버 없음(호출자가 항상 실제 경로를 넘긴다) |
| B2 | `ValidatePrivateControlDirectory(dir) != nil` | 없음 | 그 error | 0700 아닌 디렉터리·symlink 디렉터리에서 거부 |
| B3 | `os.Lstat` error | 없음 | 그 error(ErrNotExist 포함) | 첫 기동에서 ErrNotExist 경로 |
| B4 | `!socket \|\| symlink \|\| perm != 0o600` | 없음 | "not an exact 0600 Unix socket" | **pre-chmod 0700 잔재의 영구 거부가 이 절이다**(A1 F3) |
| 분기 밖 종단 | 통과 | 없음 | `(info, nil)` | 정상 0600 socket |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `strings.TrimSpace` | 정규화 | — | AST calls |
| `ValidatePrivateControlDirectory` | 부모 디렉터리 위생 | error 전파 | AST calls |
| `filepath.Dir` | 부모 경로 | — | AST calls |
| `os.Lstat` | symlink를 따라가지 않는 stat | error 전파 | AST calls |
| `info.Mode()` | 모양·권한 판정 | — | AST calls |

## State mutations and fallbacks

- 상태 변경 없음. 판정만 한다. 그래서 이 함수의 엄격함이 **회수와 클라이언트 양쪽에 동시에**
  적용된다는 사실이 a109의 설계 제약이 된다.

## Safety conclusion

- Safe edit boundary: **편집하지 않는다.** 완화는 회수 전용 신규 함수에만 존재한다.
- High-risk impact: yes — 이 절의 엄격함이 pre-chmod 잔재의 영구 거부(사고 기전)를 만들었고,
  느슨하게 하면 클라이언트 경계가 함께 넓어진다. 두 방향 모두 위험하므로 "그대로 두고 우회"가 답이다.
