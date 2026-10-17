# Function Logic Map: `reclaimStaleControlDirectory`

- Source: `internal/strategyprojectionrpc/transport_unix.go`
- AST evidence: `ast.json` — **구현 후 재생성**(:244–337, 분기 17·반환 11). 편집 전 base `54004f44` 는 :243–330.
- 구현 후 AST 대조: 분기 ID·종류·순서 17개 전부 동일(B11 이 `socketInfo, err :=` 대입 뒤의 `err != nil` 이 됐을 뿐), 호출 `projectionSocketAccepts` 가 `staleProjectionSocketAccepts` 로 바뀌었다.
- Risk scan: `risk-pattern-report.md`

a113 편집 지점은 **B11·B12 두 줄**이다: B11 이 검증한 socket 의 `FileInfo` 를 받아 두고, B12 의 생존
질문을 추정 없는 chmod-then-probe(`staleProjectionSocketAccepts(socketPath, info)`)로 바꾼다.
나머지 15 분기(디렉터리 검증·열거·staging 모양·descriptor 형식·제거 루프)는 a108 확정 의례 그대로다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `engineDir` | `Start` 가 이미 Lstat 으로 검증한 엔진 디렉터리 | 유일 호출자 `Start` :93 (CodeGraph callers) | — |
| control 디렉터리 | 정확 0700 · 우리 uid · 비symlink | `os.Lstat`·`unix.Lstat` | B1/B2 거부 |
| 엔트리 집합 | 최종 두 이름 + `.s-` staging | `os.ReadDir` | 낯선 이름 B9 거부, staging 모양 이상 B8 거부 |
| 주인의 생사 | 수락 중 / 사망 | B12 의 probe | 생존이면 거부(아무것도 지우지 않음) |

불변식(engine-safety spec): 수락 중인 socket 은 절대 unlink 하지 않는다. 사망 검증은 PID 가 아니라
connect probe 이고, **a113 이후 권한 비트로 사망을 추정하지 않는다.**
1차 방어는 journal flock(:317–320 주석)이다 — probe 는 flock 밖(엔진 아닌 same-uid 점유)의 심층 방어다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | control 디렉터리 Lstat 실패 또는 모드 불안전 | 없음 | "existing control directory is unsafe" (:247) | `TestStartRefusesUnsafeLeftoverShapes/디렉터리_권한이_0700이_아니다`·`/디렉터리가_symlink다` |
| B2 | unix.Lstat 실패 또는 소유 uid 불일치 | 없음 | "…ownership is unsafe" (:251) | 비root 로 재현 불가 — `ownedByEffectiveUser` 분리 함수로 절 단위 뮤테이션(a108 A1 F6) |
| B3 | ReadDir 실패 | 없음 | 그 error (:255) | 없음(0700 우리 소유 디렉터리의 ReadDir 실패를 비root 로 결정적으로 만들 수 없음) |
| B4–B7 | 엔트리 순회: 최종 이름이면 seen, `.s-` 면 staging 후보 | 맵·슬라이스만 | — | `TestStartRecoversFromUnpublishedStagingLeftover` |
| B8 | staging 엔트리가 정규 파일·socket 아님 | 없음 | "stale staging entry has an unexpected shape" (:278) | `TestStartRefusesStagingLeftoverOfAnUnexpectedShape` |
| B9 | 낯선 이름 | 없음 | "…unexpected entries" (:282) | `TestStartRefusesControlDirectoryWithUnknownEntry` |
| B10 | socket 이 있다 | — | — | 아래 두 행 |
| B11 | `verifyStaleSocketShape` 실패 | 없음 | 그 error (:291) — **a113: FileInfo 를 함께 받는다** | `TestStartRefusesUnsafeLeftoverShapes/socket에_group…`·`/일반_파일`·`/hard_link` |
| B12 | 생존 probe 가 true | 편집 전: connect 1회. **편집 후: chmod 0600 + Lstat + connect** | "projection owner is still alive (or its death could not be proven)" — a113 post-review P2-4 문구 정정 | `TestStartRefusesLiveSocketWithoutDescriptor`·`…PIDIsDead`·`TestStartRefusesLiveProjectionOwnerWithoutRemovingIt` + a113 신규 `TestTheReclaimRefusesALiveSocketWhoseOwnerWriteBitWasStripped` |
| B13–B14 | descriptor 가 있고 형식 검증 실패 | 없음 | "stale descriptor is unsafe" (:307) | `TestStartRefusesUnsafeLeftoverShapes/descriptor_권한이_0600이_아니다` |
| B15–B16 | 제거 루프, ErrNotExist 외 제거 실패 | staging·descriptor·socket unlink | "remove stale endpoint" (:323) | 없음(비root 로 결정적 재현 불가) |
| B17 | rmdir 실패(ErrNotExist 외) | rmdir | "remove stale control directory" (:327) | 없음(같은 사유) |
| 종단 | 모두 통과 | 잔재 전부 제거 | `nil` (:329) | `TestStartRecoversFrom*`(a108 11종) + `TestStartRecoversFromUnwritableSocketLeftover`(0500 죽은 socket) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `verifyStaleSocketShape` | socket 모양·소유·nlink | error → B11. **a113: `(os.FileInfo, error)` 반환** | AST :290, CodeGraph callers(유일 호출자) |
| `projectionSocketAccepts` → a113 후 `staleProjectionSocketAccepts` | 주인 생존 질문 | true → B12 거부. chmod 실패 = 생존(부재 제외), SameFile 불일치 = 생존 | AST :293 |
| `openVerifiedDescriptor` | descriptor 형식 | error → B14 | AST :305 |
| `os.Remove` ×2 | 잔재 제거 | ErrNotExist 용인 | AST :322·:326 |

## State mutations and fallbacks

- 편집 전: B12 까지 디스크를 바꾸지 않는다. **편집 후: B12 의 probe 가 우리 uid 소유·0700 디렉터리
  안·비symlink·nlink 1 로 검증된 socket 의 권한을 0600 으로 되돌린다.** 결과 권한이 0600 뿐이라
  접근이 넓어지지 않고, 그 값은 발행 계약 자체다(`listenPrivateSocket` 은 chmod 0600 뒤에만 rename).
  산 socket 이면 그 권한 복원이 남는다 — 외부 chmod 로 깎인 우리 socket 을 우리 계약으로 되돌린 것이다.
- staging 엔트리(`.s-*`)는 probe 없이 지운다(a108 무변경 — design 선언된 생략).

## Safety conclusion

- Safe edit boundary: B11 의 대입 한 줄 + B12 의 호출 한 줄. 열거·모양·descriptor·제거 순서 무변경.
- High-risk impact: 엔진 boot 경로(회수) — 주문·손절 코드는 아니다. 방향은 보수적이다: 사망 판정이
  추정에서 질문으로 바뀌어 **산 socket 을 지우는 경우가 사라지고**, 죽은 0500/0400 잔재는 chmod 후
  ECONNREFUSED 로 계속 회수된다(영구 거부를 다시 만들지 않는다).
