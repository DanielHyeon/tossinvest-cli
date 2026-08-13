# a108 뮤테이션 원장 — T1 (겹1 `internal/strategyprojectionrpc`)

tasks 2.4. design D1 상태 표의 판정을 **하나씩** 뒤집어(회수→거부, 거부→회수) 어떤
테스트가 죽는지 관측한 기록이다. 통과는 증거가 아니다 — 죽여 본 뒤에 근거로 쓴다
([[passing-test-is-not-evidence]]).

## 실행 방법

각 뮤테이션은 `internal/strategyprojectionrpc/transport_unix.go`에 문자열 치환 1건
(M5c만 2건)을 적용하고, **먼저 `go build`가 통과하는지 확인한 뒤**(컴파일 에러는
뮤테이션이 아니다) `go test ./internal/strategyprojectionrpc/ -count=1 -v`를 돌렸다.

원복은 `git checkout -- <file>` 뒤 **두 가지로 검증**했다.

- `git diff --exit-code -- <file>` == 0
- 심볼 대조: `projectionSocketAccepts` 3 · `processAlive` 0 · `MUTANT` 0 ·
  `reclaimStaleControlDirectory` 3

원장의 모든 행에서 `restored=True`였다(= 두 검증 모두 통과).

## 결과 — 10건 적용, 10건 사망, 생존 0

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
| M7 | 보안 검사 완화 | socket `Perm() != 0o600` 조건 삭제 | `TestStartRefusesUnsafeLeftoverShapes/socket_권한이_0600이_아니다` |
| M8 | 경합 용인 → 엄격 | `os.Remove`의 `ErrNotExist` 용인 삭제 | `TestStartRecoversFromDescriptorOnlyLeftover`, `TestStartRecoversFromEmptyControlDirectoryLeftover`, `TestStartRecoversFromDeadSocketOnlyLeftover` |

## 읽는 법 — 이 원장이 실제로 증명하는 것

**M5c가 이 change의 핵심 증거다.** 옛 판정(kill-0)으로 정확히 되돌리면 D2 핀 세 개가
동시에 죽는다. 죽는 방식이 **두 방향**이라는 것이 요점이다.

- `...DeadEndpointWhoseDescriptorPIDIsAlive` — 죽은 endpoint를 "주인 생존"으로 오판해
  **영구 거부**한다. 이것이 8/13 사고에 숨어 있던 두 번째 사고 형태다.
- `...LiveSocketWhoseDescriptorPIDIsDead` / `...LiveSocketWithoutDescriptor` — 수락 중인
  socket을 "주인 사망"으로 오판해 **남의 socket을 지운다.**

M5 단독이 한 방향만 죽이는 이유는 probe가 남아 있어 생존 거부를 여전히 잡기 때문이다.
kill-0을 판정에서 **빼는 것**만으로는 부족하고 probe가 **대신 답해야** 한다는 것을
M5·M5b·M5c 세 행이 함께 보여 준다.

M8은 D2 경합 창 문단(주인의 `Close`와 겹치는 제거)이 문장이 아니라 코드 조건임을
고정한다. 그 용인을 지우면 회수 가능한 세 상태가 전부 다시 거부로 돌아간다.

M7은 "검증을 완화하지 않는다"가 측정된다는 것이다 — 회수 범위를 넓히면서 검사 하나를
흘리면 그 순간 테스트가 죽는다.

## 남긴 것 (생략 아님, 관측 결과)

`TestStartRefusesUnsafeLeftoverShapes`의 다른 행(디렉터리 0700·symlink·socket 타입·
hard link·descriptor 0600)은 **편집 전에도 GREEN이었고 편집 후에도 GREEN이다.** 이
change가 건드리지 않은 검사이므로 뮤테이션은 대표로 M7 하나만 돌렸다.
