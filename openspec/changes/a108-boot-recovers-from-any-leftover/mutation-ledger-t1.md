# a108 뮤테이션 원장 — T1 (겹1 `internal/strategyprojectionrpc`)

tasks 2.4와 6.5. design D1/D1-2/D2/D2-2/D4-2의 판정을 **하나씩** 뒤집어 어떤 테스트가
죽는지 관측한 기록이다. 통과는 증거가 아니다 — 죽여 본 뒤에 근거로 쓴다
([[passing-test-is-not-evidence]]).

이 원장은 **두 라운드**를 담는다.

- §A 첫 구현 라운드(2026-08-13, tasks 2.4) — 회수의 전체성과 D2 판정 교체.
- §B Fix 라운드(2026-08-14, tasks 6.5) — A1 적대 리뷰가 연 발행·소유권·probe 판정.
  Fix 라운드의 코드는 §A의 코드가 아니므로, §A의 행도 **새 코드에 다시 적용해**
  결론이 유지되는지 확인했다(§B2). 옛 코드에서 죽었다는 기록은 새 코드의 증거가 아니다.

## 실행 방법

각 뮤테이션은 `internal/strategyprojectionrpc/transport_unix.go` 또는 `transport.go`에
치환 1건을 적용하고, **먼저 `go build ./...`가 통과하는지 확인한 뒤**(컴파일 에러는
뮤테이션이 아니다) `go test ./internal/strategyprojectionrpc/ -count=1 -json`을 돌려
`Action == "fail"` 이벤트를 수집했다.

원복은 `git checkout -- <파일>` 뒤 **두 가지로 검증**했다. 적용 전에 작업 트리가
커밋돼 있는지 먼저 확인한다 — 커밋 전 GREEN에 checkout을 쓰면 GREEN까지 지워지고
`git diff`는 그것을 초록으로 통과시킨다([[mutation-revert-needs-the-right-baseline]]).

- `git diff --exit-code -- <파일>` == 0
- 심볼 대조(baseline과 정확히 일치): `listenPrivateSocket` 4 · `stagingPrefix` 5 ·
  `verifyStaleSocketShape` 3 · `openVerifiedDescriptor` 4 · `ownedByEffectiveUser` 4 ·
  `controlDirectoryModeIsSafe` 4 · `sameDescriptorFile` 3 · `SetUnlinkOnClose` 4 ·
  `errDescriptorChanged` 3 · `os.Rename` 2 · `MUTANT` 0 · `processAlive` 0

모든 행에서 두 검증이 통과했다.

---

## §A. 첫 구현 라운드 (tasks 2.4) — 10건 적용, 10건 사망

| # | 뒤집은 판정 (D1 행) | 뮤테이션 | 죽은 테스트 |
|---|---|---|---|
| M1 | S1 회수 → 거부 | descriptor만 있으면 `incomplete` 반환 | `TestStartRecoversFromDescriptorOnlyLeftover` |
| M2 | S0 회수 → 거부 | 빈 디렉터리면 반환 | `TestStartRecoversFromEmptyControlDirectoryLeftover` |
| M3 | S2-사망 회수 → 거부 | socket만 있으면 반환 | `TestStartRecoversFromDeadSocketOnlyLeftover` |
| M4 | S2/S3-생존 거부 → 회수 | `if false && projectionSocketAccepts(...)` | `TestStartRefusesLiveSocketWhoseDescriptorPIDIsDead`, `TestStartRefusesLiveSocketWithoutDescriptor`, `TestStartRefusesLiveProjectionOwnerWithoutRemovingIt` |
| M5 | D2 판정에 kill-0 **추가** | descriptor PID에 `unix.Kill(pid,0)` 거부를 되살림 | `TestStartRecoversFromDeadEndpointWhoseDescriptorPIDIsAlive` |
| M5b | D2 판정 **제거** | connect probe 거부 블록 삭제 | `TestStartRefusesLiveSocketWhoseDescriptorPIDIsDead`, `TestStartRefusesLiveSocketWithoutDescriptor`, `TestStartRefusesLiveProjectionOwnerWithoutRemovingIt` |
| M5c | **정확히 옛 판정** (M5+M5b) | probe를 kill-0으로 교체 | `TestStartRecoversFromDeadEndpointWhoseDescriptorPIDIsAlive`, `TestStartRefusesLiveSocketWhoseDescriptorPIDIsDead`, `TestStartRefusesLiveSocketWithoutDescriptor` |
| M6 | S4 거부 → 회수 | 낯선 엔트리 거부 블록 삭제 | `TestStartRefusesControlDirectoryWithUnknownEntry` |
| M7 | socket perm 검사 삭제 | socket `Perm() != 0o600` 조건 삭제 | `TestStartRefusesUnsafeLeftoverShapes/socket_권한이_0600이_아니다` |
| M8 | 경합 용인 → 엄격 | `os.Remove`의 `ErrNotExist` 용인 삭제 | `TestStartRecoversFromDescriptorOnlyLeftover`, `TestStartRecoversFromEmptyControlDirectoryLeftover`, `TestStartRecoversFromDeadSocketOnlyLeftover` |

### M7 해석 정정 (A1 F1 수용, tasks 6.5)

첫 라운드는 M7을 이렇게 읽었다: **"검증을 완화하지 않는다가 측정된다 — 회수 범위를
넓히면서 검사 하나를 흘리면 그 순간 테스트가 죽는다."**

그 해석은 **틀렸다.** M7이 지키고 있던 조건은 `Perm() != 0o600`, 즉 **정확히 0600**이다.
그런데 그 시점의 발행 코드는 최종 이름에 바로 bind한 뒤 chmod했으므로, 두 줄 사이에서
죽으면 umask가 정한 권한(컨테이너 실측 077 → **0700**)의 socket이 최종 이름에 남았다.
정확-0600 검사는 그 잔재 — **우리 자신이 만든, 아무도 수락하지 않는, 우리 uid 소유의,
0700 디렉터리 안에 있는 socket** — 을 "unsafe"로 영구 거부했다. A1이 3/3으로 실측했다.

즉 M7이 증명한 것은 "보안 검사가 살아 있다"가 아니라 **"버그가 보안 핀으로 봉인돼
있다"**였다. 통과하는 테스트가 결함을 지키고 있었던 것이다. 정정된 판정은 D1-2다:
생산자가 stage+rename으로 발행하고(§B M9·M9b·M10), 구버전이 남긴 잔재를 위해 검사를
`perm&0o077 == 0`으로 **좁게** 완화한다. 완화의 폭 자체가 §B M11로 측정된다.

원 M7의 자리는 §B의 `M7'`이 잇는다 — 같은 검사를 새 판정에서 통째로 지운 뮤테이션이다.

---

## §B. Fix 라운드 (tasks 6.5) — 15건 적용, 14건 사망, 생존 1

### B1. 새 판정 뒤집기

| # | 뒤집은 판정 (design) | 뮤테이션 | 죽은 테스트 |
|---|---|---|---|
| M9 | D1-2 descriptor stage+rename | 최종 이름에 제자리 `O_EXCL` 쓰기(옛 발행) | `TestDescriptorPublicationIsAtomic` |
| M9b | D1-2 descriptor stage+rename | 최종 이름에 제자리 `O_TRUNC` 쓰기 | `TestDescriptorPublicationIsAtomic` |
| M10 | D1-2 socket stage+rename | 최종 이름에 제자리 bind(옛 발행) | `TestStartPublishesBothArtifactsByRename`, `TestPublishedListenerNeverUnlinksTheNameItRemembers` |
| M11 | D1-2 완화의 **폭** | perm 마스크 `0o077` → `0o022` | `TestStartRefusesUnsafeLeftoverShapes/socket에_group/other_비트가_있다` |
| M7' | D1-2 완화 후에도 남는 검사 | socket perm 검사 완전 삭제 | `TestStartRefusesUnsafeLeftoverShapes/socket에_group/other_비트가_있다` |
| M12 | D4-2 Dial connect probe | probe 블록 삭제 → lazy client 복원 | `TestDialRefusesSocketWithNoListener` |
| M13 | D2-2 `SetUnlinkOnClose(false)` | 플래그 제거 | `TestPublishedListenerNeverUnlinksTheNameItRemembers` |
| M14 | D2-2 `Close`의 listener 소유 | `Close`의 명시적 `s.listener.Close()` 삭제 | `TestCloseClosesItsOwnListenerWithoutServe` |
| M15 | D1-2 staging 잔재 회수 | `.staging` 접두를 낯선 엔트리로 되돌림 | `TestStartRecoversFromUnpublishedStagingLeftover` |
| M16 | D1-2 형식/내용 분리 | 회수의 descriptor 검사를 내용 파싱으로 되돌림 | `TestStartRecoversFromEmptyDescriptorLeftover`, `TestStartRecoversFromTruncatedDescriptorLeftover` |
| M17 | A1 F6 rmdir 대칭 | rmdir의 `ErrNotExist` 용인 삭제 | **생존** (아래 B3) |
| M18 | 절 — 디렉터리 symlink | `controlDirectoryModeIsSafe`의 symlink 절 삭제 | `TestControlDirectoryModeClausesEachRefuseOnTheirOwn` |
| M19 | 절 — 소유 uid | `ownedByEffectiveUser`를 항상 참으로 | `TestOwnershipClauseRefusesAnotherUser` |
| M20 | 절 — 연 뒤에도 그 파일인가 | `sameDescriptorFile`의 둘째 `SameFile` 삭제 | `TestDescriptorIdentityClausesRejectASwappedFile` |
| M21 | 절 — 검사한 파일을 열었는가 | `sameDescriptorFile`의 첫째 `SameFile` 삭제 | `TestDescriptorIdentityClausesRejectASwappedFile` |

### B2. §A 행의 재적용 — 옛 결론이 새 코드에서도 유지되는가

회수 함수를 다시 썼으므로 §A의 기록은 **다른 코드에 대한 관측**이다. 같은 판정을 새
코드에 다시 적용했고, 7건 전부 다시 죽었다(생존 0).

| # | 뮤테이션 | 죽은 테스트 |
|---|---|---|
| M1' | S1 회수 → 거부 | `TestStartRecoversFromDescriptorOnlyLeftover`, `TestStartRecoversFromEmptyDescriptorLeftover` |
| M2' | S0 회수 → 거부 | `TestStartRecoversFromEmptyControlDirectoryLeftover` |
| M3' | S2-사망 회수 → 거부 | `TestStartRecoversFromDeadSocketOnlyLeftover` |
| M4' | 생존 거부 무력화 | `TestStartRefusesLiveProjectionOwnerWithoutRemovingIt`, `TestStartRefusesLiveSocketWhoseDescriptorPIDIsDead`, `TestStartRefusesLiveSocketWithoutDescriptor` |
| M5c' | **정확히 옛 판정**: probe → descriptor PID kill-0 | `TestStartRecoversFromDeadEndpointWhoseDescriptorPIDIsAlive`, `TestStartRefusesLiveSocketWhoseDescriptorPIDIsDead`, `TestStartRefusesLiveSocketWithoutDescriptor` |
| M6' | S4 거부 → 무시 | `TestStartRefusesControlDirectoryWithUnknownEntry` |
| M8' | 파일 제거의 `ErrNotExist` 용인 삭제 | `TestStartRecoversFromDescriptorOnlyLeftover`, `TestStartRecoversFromEmptyControlDirectoryLeftover`, `TestStartRecoversFromDeadSocketOnlyLeftover`, `TestStartRecoversFromEmptyDescriptorLeftover`, `TestStartRecoversFromUnpublishedStagingLeftover` |

`M5c'`가 여전히 이 change의 핵심 증거다. 옛 판정(kill-0)으로 정확히 되돌리면 D2 핀이
**두 방향으로** 동시에 죽는다 — 죽은 endpoint를 "주인 생존"으로 오판한 영구 거부와,
수락 중인 socket을 "주인 사망"으로 오판한 남의 socket 삭제.

### B3. 살아남은 뮤테이션 — M17 (숨기지 않고 적는다)

**M17: `os.Remove(controlDir)`의 `ErrNotExist` 용인을 지워도 아무 테스트도 죽지 않는다.**

이 절이 실제로 걸리려면 파일 제거와 rmdir **사이**에 디렉터리가 사라져야 한다. 그
인터리빙은 두 번째 회수자(또는 주인의 `Close`)가 정확히 그 창에 들어와야 만들어지고,
production 코드에 seam을 새로 뚫지 않고는 결정적으로 재현할 수 없다. 두 회수를 동시에
돌리는 방법도 결정적이지 않다 — 진 쪽이 `Lstat` 단계에서 먼저 실패하는 경우가 있어
단언이 흔들린다(불안정한 테스트는 핀이 아니다).

A1 F6이 요구한 것도 "핀"이 아니라 **비대칭의 제거**였다(P3). 파일에는 용인하고
디렉터리에는 용인하지 않으면, 같은 경합이 파일에서는 양성이고 디렉터리에서는 기동
실패가 되는데 그 차이에 근거가 없다. 그 대칭은 코드와 주석에 있고, 측정은 없다.

### B4. 첫 라운드에서 살아남았다가 닫은 것

첫 §B 실행에서 **셋**이 살아남았다(M9·M17·M21). 둘을 테스트를 더해 닫았다.

- **M9** — 발행의 원자성을 재는 테스트가 없었다. 회수 테스트는 반쯤 쓰인 descriptor를
  디스크에 **직접** 만들므로, `writeDescriptor`가 어떻게 쓰든 통과한다. 재발행 300회와
  동시에 읽으면서 "반쯤 쓰인 것은 절대 보이지 않는다"를 고정하는
  `TestDescriptorPublicationIsAtomic`을 추가했고, M9와 M9b가 함께 죽는다.
  이때 rename이 발밑에서 파일을 **다른 완성 파일**로 바꾸는 경우가 관측됐다(513번째
  읽기). 그것은 손상이 아니라 원자성의 증거이므로 `errDescriptorChanged`라는 이름을
  줘서 내용 손상과 구분했다 — 이름이 없으면 정상 경합이 손상으로 읽힌다.
- **M21** — `sameDescriptorFile`의 두 절이 `&&`로 묶여 있어, 원래 세 줄은 둘째 절만으로도
  전부 통과했다. "검사 전에 바꿔치기당했고 그 뒤로는 그대로"인 줄
  (`sameDescriptorFile(one, two, two)`)을 넣어 첫 절만 겨눴다.

### B5. 절 단위 핀에 대하여 (A1 F6, 정직하게)

세 절 중 둘은 **디스크 상태로는 그 절만 실패시킬 수 없다.**

- symlink 절: 실제 `Lstat`이 돌려주는 symlink의 mode에는 `ModeDir` 비트가 없으므로
  `IsDir()` 절이 항상 먼저 걸린다. 파일 기반 테스트로는 symlink 절만 지운 뮤테이션이
  살아남는다.
- uid 절: 이 테스트는 비root(uid 1000)로 돌고, 비root는 파일 소유자를 바꿀 수 없다.
  "남이 만든 디렉터리"라는 상태 자체를 만들 수 없다.

그래서 세 절을 이름 있는 순수 판정 함수(`controlDirectoryModeIsSafe` ·
`ownedByEffectiveUser` · `sameDescriptorFile`)로 뽑고 값을 직접 먹였다. 검사를 약하게
만든 것이 아니라 **죽여 볼 수 있게** 만든 것이다 — 호출부의 조건은 이 함수를 부르는
그대로다. 파일 기반 거부 핀(`TestStartRefusesUnsafeLeftoverShapes`)은 그대로 남아 있고,
symlink 행은 대상 디렉터리 안에 **회수 가능한 잔재**를 넣도록 강화해서 "따라갔다면
지웠을 파일"의 생존으로 거부를 측정한다.

---

## §C. gstack Fix 라운드 (2026-08-14, Fix-First B1~B5) — 9건 적용, 8건 사망, 생존 1

baseline: 커밋 `d3d0cdcb`. 드라이버는 파일 원문을 **메모리에 들고** 되돌리고, 매 회
① 원문 문자열 전체 동일성 ② 아래 심볼 개수 일치를 확인했다. 9회 전부 일치.

```text
listenPrivateSocket 5 · stagingPrefix 6 · stagingPath 4 · discardPublication 4 ·
verifyStaleSocketShape 3 · openVerifiedDescriptor 4 · ownedByEffectiveUser 4 ·
controlDirectoryModeIsSafe 4 · sameDescriptorFile 3 · SetUnlinkOnClose 4 ·
errDescriptorChanged 3 · os.Rename 2 · processAlive 0 · MUTANT 0
```

### C1. RED 을 먼저 봤다 (관측 기록)

상수(`stagingSocketKind`·`stagingDescriptorKind`)와 `discardPublication` 을 **순수
추가**한 상태 — 판정은 옛것 그대로 — 에서 새 테스트를 돌려 아래를 관측했다. 그
다음에 판정을 바꿨다.

```text
임시 이름 ".staging-sock-8dd91c3a"(22자)이 최종 이름 "runtime.sock"(12자)보다 길다
잔재에서 기동이 거부됐다: listen unix …/.staging-sock-1e7ff879: bind: invalid argument
    (같은 부팅의 최종 경로는 107자로 상한 안이다)
projectionSocketAccepts = true, want false            (owner 쓰기 비트가 없는 socket)
잔재에서 기동이 거부됐다: projection owner is still alive   (같은 socket, Start 경유)
치우면 안 되는 상태에서 기동이 받아들여졌다              (빈 .s-* 디렉터리)
```

RED 을 **테스트만 담은 커밋**으로 남기지 못한 이유를 적는다: 새 핀 셋이 아직 없는
심볼(상수 2 · `discardPublication`)을 부르므로 테스트만 커밋하면 컴파일이 깨지고,
컴파일 실패는 RED 가 아니다. 그래서 순서를 「순수 추가 → RED 관측 → 판정 변경」으로
쪼갰고 관측 결과를 위에 그대로 남긴다.

### C2. 원장

| # | 뒤집은 판정 | 뮤테이션 | 죽은 테스트 |
|---|---|---|---|
| M22 | B1 임시 이름 길이 | `stagingPrefix` 를 `.staging-` 로 되돌림 | `TestStagingNamesAreNeverLongerThanTheNamesTheyBecome`, `TestStartPublishesWhereTheFinalSocketPathIsAtTheLimit`, `TestStartRecoversFromUnpublishedStagingLeftover` |
| M23 | B1 임시 이름 길이 | 종류 이름을 `sock`·`endpoint` 로 되돌림 | `TestStagingNamesAreNeverLongerThanTheNamesTheyBecome`, `TestStartPublishesWhereTheFinalSocketPathIsAtTheLimit` |
| M24 | B1 descriptor 이름 생성기 | `stagingPath` → `os.CreateTemp` (옛 발행) | **생존** (아래 C3) |
| M25 | B2① owner 쓰기 절 | 절 삭제 — EACCES 를 다시 「생존」으로 | `TestProjectionLivenessClausesEachDecideOnTheirOwn/owner_쓰기_비트가_없다`, `TestStartRecoversFromUnwritableSocketLeftover` |
| M26 | B2② 연결 거부·부재 절 | 절 삭제 — 죽은 socket 을 「생존」으로 | `TestProjectionLivenessClausesEachDecideOnTheirOwn/{경로가_없다,아무도_수락하지_않는다}`, `TestDialRefusesSocketWithNoListener`, `TestStartReclaimsExactDeadProjectionEndpoint`, `TestStartRecoversFrom{DeadEndpointWhoseDescriptorPIDIsAlive,DeadSocketOnlyLeftover,PreChmodSocketLeftover,TruncatedDescriptorLeftover}` |
| M27 | B2② 수락 절 | 수락을 「사망」으로 반전 | 18건 — `TestProjectionLivenessClausesEachDecideOnTheirOwn/수락한다` 와 `TestStartRefusesLive*` 셋을 포함 |
| M28 | B3 staging 모양 검증 | 검증 삭제 — 이름만 보고 지운다 | `TestStartRefusesStagingLeftoverOfAnUnexpectedShape/빈_디렉터리` |
| M29 | B4 실패 정리의 범위 | `discardPublication` 에서 descriptor 제거를 뺌 | `TestDiscardingAFailedPublicationLeavesNothingBehind` |
| M30 | B2③ 디렉터리 모드 검사 | `Perm() == 0o700` → `Perm()&0o077 == 0` | `TestControlDirectoryModeClausesEachRefuseOnTheirOwn` |

### C3. 살아남은 뮤테이션 — M24 (숨기지 않고 적는다)

**M24: descriptor 의 임시 이름을 `os.CreateTemp` 로 되돌려도 아무 테스트도 죽지 않는다.**

새 접두(`.s-`) 위에서 `os.CreateTemp` 는 `.s-e` + 임의 자릿수 십진수를 붙이므로
basename 이 최대 14자가 되어 「최종 이름(13자) 이하」 계약을 깬다. 그런데 그 초과에
**디스크에서 관측되는 결과가 없다.** 길이 계약이 실제로 무언가를 정하는 곳은 socket
쪽뿐이고(bind 의 sun_path 107자 상한), descriptor 는 PATH_MAX 근처에도 가지 않는다.
socket 쪽 계약은 M22·M23 과 `TestStartPublishesWhereTheFinalSocketPathIsAtTheLimit`
이 두 방향으로 잡는다.

결정적으로 죽이는 방법도 없다: `CreateTemp` 의 접미 길이가 난수라 자릿수가 회마다
다르고(1~10), 임시 이름은 rename 뒤 사라져 관측 창이 마이크로초다. 길이를 production
코드에서 단언하는 가드를 넣으면 그 가드가 **확률적으로** 걸려 불안정한 테스트가 된다
— 불안정한 테스트는 핀이 아니다([[passing-test-is-not-evidence]] 의 반대편).

그래서 descriptor 쪽에서 M24 가 지키는 것은 **일관성 결정**이다: 두 산출물이 같은
생성기·같은 접두·같은 길이 규칙을 쓴다. 그 결정은 코드와 주석에 있고, 측정은 없다.
(M17 과 같은 부류의 정직한 공백이다.)
