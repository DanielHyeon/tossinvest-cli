# Function Logic Map: `reclaimStaleControlDirectory`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (206-282)
- AST evidence: `ast.json` — revision `current`. AST 분기 16
- Risk scan: `risk-pattern-report.md`

이 change의 중심 함수다. 첫 라운드가 회수의 **커버리지**를 넓혔고(D1·D2), Fix 라운드가
**발행이 만드는 상태**까지 회수 대상에 넣었다(D1-2). 분기는 13 → 16으로 늘었다:
staging 접두 case, socket 모양 검사 분리, rmdir의 `ErrNotExist` 절.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| control 디렉터리 | 실디렉터리 · symlink 아님 · 정확히 0700 | `os.Lstat` + `controlDirectoryModeIsSafe` | B1 거부, 아무것도 지우지 않는다 |
| 소유 uid | 유효 uid | `unix.Lstat` + `ownedByEffectiveUser` | B2 거부 |
| 엔트리 이름 | `endpoint.json` · `runtime.sock` · `.staging-` 접두 | `os.ReadDir` | B8 거부(낯선 엔트리 하나면 전체 보존) |
| socket 모양 | socket 타입 · symlink 아님 · `perm&0o077 == 0` · 우리 uid · nlink 1 | `verifyStaleSocketShape` | B10 거부 |
| 주인의 생사 | 수락하면 살아 있다 | `projectionSocketAccepts` (connect probe) | B11 거부 — **남의 것은 치우지 않는다** |
| descriptor 모양 | 정확히 0600 정규 파일 · no-follow · 열기 전후 같은 inode | `openVerifiedDescriptor` | B13 거부 |
| descriptor **내용** | 검사하지 않는다 | — | 사망 입증 뒤에는 파싱 실패가 회수다(design D1-2) |

### 완화의 폭과 그 조건 (design D1-2)

socket perm 검사는 **정확-0600**에서 **`perm&0o077 == 0`**으로 좁게 완화됐다. 근거는
A1 F1: 구버전은 최종 이름에 bind한 뒤 chmod했으므로 그 사이의 죽음이 umask가 정한
0700 socket을 남겼고, 정확-0600 검사가 **자기 자신이 만든 잔재**를 영구 거부했다.
완화가 안전한 이유는 나머지 다섯 조건이 전부 성립할 때만 회수하기 때문이다 —
우리 uid · 0700 디렉터리 안 · symlink 아님 · hard link 없음 · **아무도 수락하지 않음**.
group·other 비트가 하나라도 있으면 지금도 거부다(뮤테이션 M11이 그 폭을 잰다).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 디렉터리 모양이 안전하지 않다 | 없음 | `existing control directory is unsafe` | `TestStartRefusesUnsafeLeftoverShapes/디렉터리...` |
| B2 | 소유 uid가 우리가 아니다 | 없음 | `... ownership is unsafe` | `TestOwnershipClauseRefusesAnotherUser`(절 단위) |
| B3 | `ReadDir` 실패 | 없음 | 그 오류 | 없음 |
| B4 | 엔트리 순회 | 없음 | — | `TestStartRefusesControlDirectoryWithUnknownEntry` |
| B5 | 이름 분류 switch | 없음 | — | 위와 같음 |
| B6 | 최종 이름 둘 중 하나 | `seen` 표시 | — | 회수 테스트 전부 |
| B7 | `.staging-` 접두 | 회수 목록에 추가 | — | `TestStartRecoversFromUnpublishedStagingLeftover` |
| B8 | 그 밖의 이름 | 없음 | `unexpected entries` | `TestStartRefusesControlDirectoryWithUnknownEntry` |
| B9 | socket이 있다 | 없음 | — | `TestStartRecoversFromDeadSocketOnlyLeftover` |
| B10 | socket 모양 검사 실패 | 없음 | 그 오류 | `TestStartRefusesUnsafeLeftoverShapes/socket...` |
| B11 | **누가 수락한다** | 없음 | `projection owner is still alive` | `TestStartRefusesLiveSocketWhoseDescriptorPIDIsDead` |
| B12 | descriptor가 있다 | 없음 | — | `TestStartRecoversFromDescriptorOnlyLeftover` |
| B13 | descriptor **형식** 검사 실패 | 없음 | `stale descriptor is unsafe` | `TestStartRefusesUnsafeLeftoverShapes/descriptor...` |
| B14 | 제거 목록 순회 | **파일 제거** | — | 회수 테스트 전부 |
| B15 | 제거 실패이고 `ErrNotExist`가 아니다 | 없음 | 그 오류 | `TestCloseToleratesLeftoverAlreadyRemoved` |
| B16 | rmdir 실패이고 `ErrNotExist`가 아니다 | 없음 | 그 오류 | 뮤테이션 M17 **생존**(아래) |

## 판정의 순서가 왜 이 순서인가

**socket → descriptor** 순이다(첫 라운드는 반대였다). 주인의 생사가 descriptor를 얼마나
엄격하게 볼지 정하기 때문이다: 사망이 증명된 뒤에야 내용 파싱 실패를 회수로 읽을 수
있고, 증명 전에 그렇게 하면 살아 있는 주인의 파일을 지우게 된다.

사망의 증명은 둘 중 하나다 — socket 파일이 아예 없거나(`Start`가 listen 성공 뒤에만
descriptor를 쓰므로 listener가 없었다는 뜻), 있는데 아무도 수락하지 않거나.

## Calls and live bindings

| Callee | Why called | Error contract | Evidence |
|---|---|---|---|
| `controlDirectoryModeIsSafe` | 디렉터리 모양 세 절 | 순수 판정 | AST calls, B1 |
| `ownedByEffectiveUser` | 소유 uid 절 | 순수 판정 | AST calls, B2·`verifyStaleSocketShape` |
| `verifyStaleSocketShape` | socket 모양 다섯 절 | 오류를 그대로 올린다 | AST calls, B10 |
| `projectionSocketAccepts` | **생존 판정** | 판정 불가는 "살아 있다"로 읽는다(보수) | AST calls, B11 |
| `openVerifiedDescriptor` | descriptor 형식만 | 내용은 보지 않는다 | AST calls, B13 |
| `os.Remove` × (staging + 2 + dir) | 제거 | `ErrNotExist` 용인 | AST calls, B15·B16 |

## State mutations and fallbacks

- 지우는 것: staging 잔재 전부 → descriptor → socket → 디렉터리. 그 밖의 것은
  **하나도** 지우지 않는다(B8이 그 경계다).
- 경합 용인: `ErrNotExist`는 성공과 같다. 주인의 `Close`와 겹치면 양쪽이 같은 경로를
  지우는데 결과가 같으므로 양성이다(design D2).
- **probe와 제거 사이에 새 주인이 들어오는 경합**의 방어는 이 함수가 아니라 부팅
  1단계의 journal flock이다(`cmd/tossctl/engine.go`의 "flock on the journal directory
  FIRST"). 이 함수는 원자적이지 않고, 원자적일 필요도 없다 — 엔진이 하나라는 것이
  상위 계층에서 강제되기 때문이다. 코드 주석이 그 인용을 담는다(A1 F6).
- **B16의 측정 부재(선언)**: rmdir의 `ErrNotExist` 용인은 파일 제거와 rmdir **사이**에
  디렉터리가 사라져야 걸린다. seam 없이 결정적으로 만들 수 없어 뮤테이션 M17이
  살아남았고, 원장 §B3에 생존으로 적었다. A1 F6이 요구한 것은 핀이 아니라 비대칭
  제거였다 — 그 대칭은 코드에 있다.

## Safety conclusion

- Safe edit boundary: **판정 순서(socket → descriptor)와 완화의 폭.** 어느 쪽을 건드려도
  design D1-2의 조건 집합을 다시 세워야 한다.
- High-risk impact: yes — 잘못 회수하면 살아 있는 주인의 socket을 지우고, 잘못 거부하면
  8/13 사고(영구 기동 실패 → 손절 감시 소멸)가 재발한다. 두 방향 모두 핀이 있다.
