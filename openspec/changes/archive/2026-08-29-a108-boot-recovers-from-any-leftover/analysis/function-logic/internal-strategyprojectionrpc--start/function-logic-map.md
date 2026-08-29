# Function Logic Map: `Start`

- Source: `internal/strategyprojectionrpc/transport_unix.go` (79-162)
- AST evidence: `ast.json` — revision `current`. AST 분기 14
- Risk scan: `risk-pattern-report.md`

## base 선언의 철회 (숨기지 않고 적는다)

첫 라운드의 이 문서는 **`Start` 본문을 한 줄도 바꾸지 않았다**고 적었고, Pre-Edit
선언 T1-3이 그것을 무변경으로 못박았다. 근거는 "listen → chmod → token → descriptor"
순서가 생존 판정의 전제라 건드리면 그 순서 자체가 리뷰 대상이 된다는 것이었다.

**A1 적대 리뷰가 정확히 그 순서에서 결함을 찾았다.** `net.Listen`과 `os.Chmod` 사이는
원자적이지 않고, 그 사이에서 죽으면 umask가 정한 권한(컨테이너 실측 077 → 0700)의
socket이 최종 이름에 남는다. 회수의 정확-0600 검사가 그것을 영구 거부했다 — 무변경
결정이 지키려던 "순서"가 사고의 생산자였다. Fix 라운드는 그 두 줄을
`listenPrivateSocket` 한 호출로 바꾼다(design D1-2). 분기 수는 15 → 14로 줄었다:
chmod 실패 분기가 이 함수 밖으로 갔다.

Pre-Edit 선언 T1-3의 철회는 `pre-edit-gate.md`의 T1-fix 절에 기록했다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reader` | non-nil | 호출부(`internal/app/engine`의 projection 주입) | B1 거부 |
| `engineDir` | 실디렉터리 · symlink 아님 · group/other 쓰기 비트 없음 | `os.Lstat` | B2 `engine directory is unsafe` |
| `.strategy-runtime-read/` | 없어야 한다. 있으면 회수 대상 | `os.Mkdir`의 `ErrExist` | B3-B6: 회수 실패는 그대로 기동 실패 |
| socket 발행 | 임시 이름 bind → 0600 → rename. **listen 성공이 descriptor 발행의 선행 조건** | `listenPrivateSocket` | B7 거부 + 디렉터리 되돌림 |
| token | 32바이트 난수 | `crypto/rand` | B8 거부 + listener 되돌림 |
| descriptor 발행 | stage → write → sync → rename | `writeDescriptor` | B14 거부 + listener 되돌림 |

**이 함수가 회수 판정에 주는 전제**: descriptor는 listen·token이 전부 성공한 뒤에만
쓰인다(B7 → B8 → B14 순). 그래서 "수락하지 않는 socket 파일"은 죽은 주인의 것이라고
결론할 수 있다(design D2). Fix 라운드는 여기에 하나를 더한다 — **최종 이름은 완성된
산출물에만 붙는다**(D1-2). 그래서 최종 이름의 socket이 0600이 아닌 상태는 이 판
바이너리에서는 만들어지지 않고, 구버전이 남긴 것만 남는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `reader == nil` | 없음 | 거부 | 없음(무변경) |
| B2 | engine 디렉터리가 안전하지 않다 | 없음 | 거부 | 없음(무변경) |
| B3 | `Mkdir` 실패 | 없음 | B4/B5/B6로 | `TestStartRecoversFromDescriptorOnlyLeftover` |
| B4 | 실패가 `ErrExist`가 아니다 | 없음 | 거부 | 없음(무변경) |
| B5 | 회수가 거부했다 | 없음(잔재 보존) | 거부 | `TestStartRefusesControlDirectoryWithUnknownEntry` |
| B6 | 회수 후 재`Mkdir` 실패 | 없음 | 거부 | 없음 |
| B7 | socket 발행 실패 | 디렉터리 되돌림 | 거부 | 없음 |
| B8 | 토큰 생성 실패 | listener·socket·디렉터리 되돌림 | 거부 | 없음 |
| B9 | 토큰 불일치 | 없음 | 401 | `TestUnixEndpointStrictMethodBodyQueryAuthAndRouteGuards` |
| B10 | GET·HEAD가 아니다 | 없음 | 405 | `TestUnixEndpointStrictMethodBodyQueryAuthAndRouteGuards` |
| B11 | 본문·쿼리가 있다 | 없음 | 400 | `TestUnixEndpointStrictMethodBodyQueryAuthAndRouteGuards` |
| B12 | 스냅샷 검증 실패 | 없음 | 503 | 없음 |
| B13 | HEAD | 헤더만 | 200 | 없음 |
| B14 | descriptor 발행 실패 | listener·socket·디렉터리 되돌림 | 거부 | 없음 |

## Calls and live bindings

| Callee | Why called | Error contract | Evidence |
|---|---|---|---|
| `reclaimStaleControlDirectory` | 잔재 회수 | 거부는 그대로 기동 실패 | AST calls, B5 |
| `listenPrivateSocket` | socket 발행(stage+rename+unlink 끔) | 실패 시 자기 임시 파일까지 되돌린다 | AST calls, B7 |
| `writeDescriptor` | descriptor 발행(stage+rename) | 실패 시 listener 되돌림 | AST calls, B14 |
| `server.server.Serve(listener)` | goroutine | `Close`가 listener를 직접 닫는다(D2-2) | AST calls |

## State mutations and fallbacks

- 디스크에 만드는 것: control 디렉터리 · 임시 이름 둘 · 최종 이름 둘. **임시 이름은
  rename으로 사라지거나 실패 경로에서 지워진다.** 그 이름으로 죽은 잔재는 회수가
  자기 잔재로 치운다(design D1-2).
- 되돌림은 층층이다: `cleanupDir` → `cleanupListener`(listener + socket + dir).
  listener는 자기 경로를 unlink하지 않으므로 socket 파일 제거는 여기서만 한다.

## Safety conclusion

- Safe edit boundary: socket·descriptor **발행 순서**. listen이 descriptor보다 먼저라는
  것은 회수의 생존 판정 전제이므로 바꾸면 design D2를 다시 봐야 한다.
- High-risk impact: yes — 엔진 기동 7단계. 실패는 강등(엔진은 계속 돈다, design D3)
  이지만 이 함수가 잘못 성공하면 회수가 남의 socket을 지울 수 있다.

## gstack Fix 라운드가 이 함수에서 바꾼 것 (2026-08-14)

분기 수는 14 로 불변이다. 바뀐 것은 **실패 정리의 범위**와 그것을 부르는 이름이다.

- `cleanupListener` 가 `discardPublication(controlDir, descriptorPath, socketPath)` 를
  부른다. 예전에는 socket 과 디렉터리만 치웠는데, `writeDescriptor` 의 rename 이
  성공한 **뒤에** 실패하는 줄이 있다(발행 확인의 SameFile, 디렉터리 sync). 그 경로의
  descriptor 는 이미 최종 이름을 갖고 있으므로, 치우지 않으면 당회 부팅 내내 S1 잔재가
  남고 디렉터리 제거도 ENOTEMPTY 로 실패한다.
- socket 임시 이름의 종류 상수가 `stagingSocketKind` 가 됐다(길이 계약 — 아래 참조).

그 인터리빙(rename 성공 후 실패)은 seam 을 새로 뚫지 않고 결정적으로 만들 수 없어서,
핀은 정리의 **커버리지**를 잰다: `TestDiscardingAFailedPublicationLeavesNothingBehind`
가 세 산출물을 놓고 부른 뒤 디렉터리가 사라졌는지 본다(하나라도 빠지면 남는다).
뮤테이션 M29 가 그 행을 지키고, 사유는 `mutation-ledger-t1.md` §C 에 적었다.
