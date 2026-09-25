# Function Logic Map: `verifyStaleSocketShape`

- Source: `internal/strategyprojectionrpc/transport_unix.go`
- AST evidence: `ast.json` — **구현 후 재생성**(:361–371, 분기 2·반환 3). 편집 전 base `54004f44` 는 :353–363.
- 구현 후 AST 대조: 분기 2 동일. B2 의 입력이 두 번째 `unix.Lstat` 에서 같은 Lstat 의 `info.Sys().(*syscall.Stat_t)` 로 바뀌었다(freeze P2-1).
- Risk scan: `risk-pattern-report.md`

a113 편집: 반환을 `error` 에서 `(os.FileInfo, error)` 로 넓혀 **검증한 그 inode 의 FileInfo** 를
호출자에게 준다. chmod-then-probe 는 이름에 걸리므로, probe 가 "검증한 파일이 아직 그 이름인가"를
`os.SameFile` 로 확인하려면 검증 시점의 FileInfo 가 필요하다(a109 §2b.3 G7 원형 —
`verifyStalePrivateSocket` 이 같은 모양이다). 분기 조건은 한 글자도 바꾸지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `socketPath` | control 디렉터리 안 `runtime.sock` | 유일 호출자 `reclaimStaleControlDirectory` :290 | — |
| 모양 | socket · 비symlink · `perm&0o077 == 0` | `os.Lstat` | B1 |
| 소유·링크 | 우리 uid · nlink 1 | `unix.Lstat` | B2 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | Lstat 실패 · socket 아님 · symlink · group/other 비트 | 없음 | "stale socket is unsafe" (:356) | `TestStartRefusesUnsafeLeftoverShapes/socket에_group/other…`·`/socket_자리에_일반_파일` |
| B2 | unix.Lstat 실패 · 소유 불일치 · nlink ≠ 1 | 없음 | "stale socket ownership is unsafe" (:360) | `…/socket에_hard_link가_걸려_있다` |
| 종단 | 통과 | 없음 | `nil` (:362) → a113 후 `(info, nil)` | `TestStartRecoversFromPreChmodSocketLeftover` 등 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.Lstat` | 모양·권한 (symlink 안 따라감) | error → B1 | AST :354 |
| `info.Sys().(*syscall.Stat_t)` + `ownedByEffectiveUser` (편집 전 `unix.Lstat`) | uid·nlink — 돌려주는 inode 와 같은 stat | 타입 단언 실패 → B2 | AST :366 |

## State mutations and fallbacks

- 디스크를 바꾸지 않는다(편집 전·후 동일). 두 Lstat 사이의 이름 교체는 기존 창이고, 같은 uid
  위협 경계 안이다(a109 design D2) — a113 의 SameFile 재확인이 그 뒤 창을 좁힌다.

## Safety conclusion

- Safe edit boundary: 반환 값 추가만. 조건·오류 문구 무변경. 실패 경로는 `nil` FileInfo 를 준다.
- High-risk impact: no — 검증을 넓히지도 좁히지도 않는다.
