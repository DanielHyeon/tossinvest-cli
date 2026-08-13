# Function Logic Map: `reclaimStaleControlDirectory`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (144-197)
- AST evidence: `ast.json` (편집 **후** 재생성 — AST 분기 14, revision `current`)
- Risk scan: `risk-pattern-report.md`

편집 전 이 함수는 118-167이고 분기 13이었다. 아래 표의 분기 주장은 전부 재생성한
`ast.json`의 열거에서 나온다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `engineDir` | 이미 `Start`가 검사한 엔진 디렉터리 | 호출부 `Start`(base AST B2가 lstat·perm 검사 후 통과시킨 값) | 여기서는 재검사하지 않는다 — 이 함수의 단독 호출부는 `Start` 하나다 |
| `<engineDir>/.strategy-runtime-read/` | 실디렉터리 · symlink 아님 · perm 정확히 0700 · 소유 uid == euid | 디스크(`os.Lstat` + `unix.Lstat`) | B1·B2 거부. 회수 가능 상태에서도 면제 없음 |
| 디렉터리 엔트리 집합 | `endpoint.json`·`runtime.sock`의 **부분집합**(공집합 포함) | `os.ReadDir` | B5 거부 — 우리가 만들 수 있는 이름이 아니면 우리 잔재가 아니다 |
| `endpoint.json`(있을 때) | `readDescriptor` 계약: 0600 정규파일 · symlink 아님 · no-follow open 후 동일 inode · 스키마 v1 | 디스크 | B7 거부 |
| `runtime.sock`(있을 때) | socket 타입 · symlink 아님 · perm 정확히 0600 · uid == euid · nlink == 1 | 디스크 | B9·B10 거부 |
| socket의 주인 생존 | **연결이 수락되지 않을 것** | `projectionSocketAccepts`의 connect probe (PID 아님 — design D2) | B11 거부 |

관통 불변식 둘.

1. **회수는 우리 수명주기가 만든 상태에서만 일어난다.** 이름 집합(B5)·권한·소유권이
   그 판정이고, 하나라도 어긋나면 아무것도 지우지 않는다.
2. **빠진 파일은 낯선 것이 아니라 덜 만들어진 것이다.** 부분집합을 허용하는 것이
   이 change의 전부이며, 그것이 없으면 넷 중 셋이 영구 거부가 된다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `os.Lstat` 실패 · 디렉터리 아님 · symlink · perm != 0700 | 없음 | `existing control directory is unsafe` | `TestStartRefusesUnsafeLeftoverShapes` (0700 아님 · symlink 행) |
| B2 | `unix.Lstat` 실패 · uid != euid | 없음 | `existing control directory ownership is unsafe` | 소유권 행 — 같은 uid로 도는 테스트에서는 재현 불가(아래 "측정 한계") |
| B3 | `os.ReadDir` 실패 | 없음 | 원본 오류 그대로 | 없음 (0700 통과 후 읽기 실패는 환경 이상) |
| B4 | 엔트리 순회 | `seen` 채움 | — | 전 회수 테스트 |
| B5 | 엔트리 이름이 descriptor/socket 둘 다 아님 | 없음 | `stale control directory has unexpected entries` | `TestStartRefusesControlDirectoryWithUnknownEntry` |
| B6 | descriptor가 있다 | 없음 | — | `TestStartRecoversFromDescriptorOnlyLeftover` |
| B7 | `readDescriptor` 실패 | 없음 | `stale descriptor is unsafe: %w` | `TestStartRefusesUnsafeLeftoverShapes` (0600 아님 · 반쯤 쓰인 descriptor 행) |
| B8 | socket이 있다 | 없음 | — | `TestStartRecoversFromDeadSocketOnlyLeftover` |
| B9 | socket lstat 실패 · socket 아님 · symlink · perm != 0600 | 없음 | `stale socket is unsafe` | `TestStartRefusesUnsafeLeftoverShapes` (일반 파일 · 0600 아님 행) |
| B10 | uid != euid · nlink != 1 | 없음 | `stale socket ownership is unsafe` | `TestStartRefusesUnsafeLeftoverShapes` (hard link 행) |
| B11 | connect probe가 수락됨 | 없음 | `projection owner is still alive` | `TestStartRefusesLiveSocketWhoseDescriptorPIDIsDead`, `TestStartRefusesLiveSocketWithoutDescriptor`, `TestStartRefusesLiveProjectionOwnerWithoutRemovingIt` |
| B12 | descriptor·socket 경로 순회 | **파일 제거** | — | 전 회수 테스트 |
| B13 | `os.Remove` 실패이고 `ErrNotExist`가 아님 | 일부 제거된 상태로 남음 | `remove stale endpoint: %w` | `TestStartRecoversFrom*`(용인 쪽) — 뮤테이션 M8이 반대쪽을 죽인다 |
| B14 | 디렉터리 `os.Remove` 실패 | 파일 둘은 이미 제거됨 | `remove stale control directory: %w` | 없음 (B5가 비어 있지 않은 디렉터리를 먼저 막는다) |

**early return이 아닌 성공 경로**는 197줄의 `return nil` 하나이며, 그 앞에서 디렉터리까지
제거된다 — 회수 성공은 곧 `Start`가 `os.Mkdir`을 다시 성공시킬 수 있는 상태다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ControlDirectory` / `DescriptorPath` / `SocketPath` | 경로 계산 (transport.go) | 순수 함수 | AST calls |
| `os.Lstat` / `unix.Lstat` | 타입·권한·소유·nlink 검사 | 오류는 즉시 거부 | AST calls, B1/B2/B9/B10 |
| `os.ReadDir` | 엔트리 이름 집합 | 오류 그대로 전파 | AST calls, B3 |
| `readDescriptor` (transport.go:92) | descriptor의 형식·권한·inode 동일성 | 오류는 wrap 후 거부. **PID·Token은 더 이상 읽지 않는다** | AST calls, B7 |
| `projectionSocketAccepts` (신규, 같은 파일) | 이 경로에서 수락하는 자가 있는가 | `net.DialTimeout` 200ms. 수락=생존, ECONNREFUSED·부재=사망, 그 밖=생존으로 보수 판정 | AST calls, B11 |
| `os.Remove` | 잔재 제거 | `ErrNotExist`만 용인 | AST calls, B13/B14 |
| `errors.New` / `fmt.Errorf` | 거부 사유 | — | AST calls |

**제거된 live binding**: `processAlive`(kill-0). 판정에서 빼는 것이 아니라 함수째 지웠다 —
남기면 다음 독자가 다시 판정에 쓴다(design D2).

## State mutations and fallbacks

- 디스크 변경은 B12/B13의 `os.Remove` 2회와 B14의 디렉터리 `os.Remove` 1회뿐이다.
  그 앞의 모든 분기는 **읽기만** 한다 — 거부 경로는 디스크를 바꾸지 않는다.
- fallback은 없다. 회수하거나 거부하거나 둘 중 하나이고, 부분 회수 후 성공은 없다.
- 경합: 주인이 자기 `Close` 도중이면 같은 경로를 양쪽이 지운다. `Close`도(server.close B3)
  `ErrNotExist`를 용인하므로 결과가 같다 — 양성 경합으로 문서화하고 수용한다(design D2).
- **잔여**: `writeDescriptor`는 O_EXCL 생성 → write → sync 순이라 그 사이에서 죽으면
  0바이트 descriptor가 남고, B7이 그것을 거부한다. 이 상태는 a108이 닫지 못했다.

## Safety conclusion

- Safe edit boundary: 이 함수의 반환은 `Start`의 B5(base AST)에서 그대로 기동 실패가 된다.
  따라서 "거부를 늘리는 변경"은 부팅을 못 하게 만들 수 있고, "회수를 늘리는 변경"은 남의
  파일을 지울 수 있다. 두 방향 다 High-risk이며, 이 편집은 후자를 **connect probe로
  좁히면서** 전자를 넓혔다.
- High-risk impact: yes — 엔진 기동 경로. Pre-Edit 선언은 `../../pre-edit-gate.md` T1-1.
- 측정 한계: B2(uid 불일치)·B3(ReadDir 실패)·B14는 같은 uid·정상 파일시스템에서 도는
  테스트로 재현할 수 없다. 세 분기 모두 이 change가 **바꾸지 않은** 코드이고, B2는
  편집 전부터 무테스트였다. 숨기지 않고 적는다.
