# a108 High-risk Pre-Edit 선언

`docs/WORKFLOW.md:387-401`의 형식을 따른다. 이 파일은 T1(겹1 —
`internal/strategyprojectionrpc/**`) 몫이다. T2(겹2·겹3)는 자기 항목을 아래에 덧붙인다.

이 change가 닿는 경로는 **엔진 기동**이다. 주문·손절·사이징 코드는 건드리지 않지만,
기동이 거부되면 보호 루프가 서지 못하므로(사고 실측: reconcile·exit·filldetect 정지)
면제 대상이 아니다.

---

## T1-1. `internal/strategyprojectionrpc.reclaimStaleControlDirectory` — 잔재 회수 판정

```text
Pre-Edit Gate:
- change id / task id: a108 / tasks 1.1~1.4, 2.1~2.4
- 대상 심볼(패키지.함수):
    internal/strategyprojectionrpc.reclaimStaleControlDirectory (transport_unix.go:118-167)
- 기존 동작 파악 근거:
    FLM analysis/function-logic/internal-strategyprojectionrpc--reclaimstalecontroldirectory/
      (AST 분기 13 · return 12 — 손으로 읽은 열거가 아니다)
    기존 테스트: transport_unix_test.go
      TestStartReclaimsExactDeadProjectionEndpoint (둘 다 있고 PID 죽음 → 회수)
      TestStartRefusesLiveProjectionOwnerWithoutRemovingIt (살아 있는 주인 → 거부)
    호출부: Start 하나 (transport_unix.go:48, AST start 분기 15에서 ErrExist 뒤)
- upstream 상속 테스트 영향: no (strategyprojectionrpc는 TossOS 고유 endpoint)
- 실패 테스트 선행 작성: yes — 새 파일
    internal/strategyprojectionrpc/a108_leftover_recovery_test.go 에 S0·S1·S2-사망·
    PID재사용 4종을 먼저 RED로 고정하고, 거부 유지 핀(S4·수락 중 socket·perm/uid/symlink)은
    편집 전후 모두 GREEN이어야 하는 항목으로 같이 넣는다.
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** 통과 조건 넷:
    (1) **검증을 완화하지 않는다.** 디렉터리 0700·소유 uid·symlink 금지·socket 0600·
        nlink 1·descriptor 형식 검사는 회수 가능한 상태에서도 전부 수행한다. 이 change는
        회수의 **커버리지**를 넓히는 것이지 검사를 지우는 것이 아니다(design D1 마지막 문단).
    (2) **우리가 만들지 않은 것은 치우지 않는다.** 낯선 엔트리(S4)가 하나라도 있으면
        디렉터리 전체를 그대로 두고 거부한다. 회수 대상 이름은 endpoint.json·runtime.sock
        둘뿐이다.
    (3) **살아 있는 주인의 socket을 지우지 않는다.** 생존 판정은 connect probe이고,
        수락하면 거부다. ECONNREFUSED(와 socket 부재)만 사망으로 읽고, 그 밖의 오류는
        "죽었다고 단정할 수 없음"이므로 보수적으로 거부한다.
    (4) **회수 실패를 성공으로 삼키지 않는다.** os.Remove 오류는 ErrNotExist(경합 양성)만
        용인하고 나머지는 그대로 반환한다.
```

## T1-2. `internal/strategyprojectionrpc.processAlive` — 판정에서 제거(함수 삭제)

```text
Pre-Edit Gate:
- change id / task id: a108 / tasks 2.2
- 대상 심볼(패키지.함수):
    internal/strategyprojectionrpc.processAlive (transport_unix.go:169-175)
- 기존 동작 파악 근거:
    FLM analysis/function-logic/internal-strategyprojectionrpc--processalive/ (AST 분기 1)
    호출부: reclaimStaleControlDirectory 하나뿐 (transport_unix.go:146).
      `grep -rn processAlive --include=*.go` 결과 정의 1 + 호출 1.
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: yes — PID 재사용 핀(tasks 1.3).
    descriptor PID = 테스트 자신의 PID(확실히 생존) + 죽은 socket → 현재 코드는
    "projection owner is still alive"로 영구 거부한다. 이 RED이 판정 교체의 근거다.
- 안전 불변식 §0 위반 여부 검토: **통과.** kill-0 판정을 남기면(로그 참고값이라도)
    다음 독자가 다시 판정에 쓴다(design D2). 함수를 지워서 되돌아갈 자리를 없앤다.
    삭제로 FLM은 `revision: base`로 남는다 — 증거를 지우는 것이 아니다.
```

## T1-3. `internal/strategyprojectionrpc.Start` — 편집하지 않는다(선언된 무변경)

`Start`(transport_unix.go:34-116, AST 분기 15)는 `ErrExist` → reclaim → 재Mkdir 순서를
이미 갖고 있고, 회수의 전체성은 reclaim 안에서 끝난다. **Start 본문은 건드리지 않는다.**
건드리면 listen→chmod→token→descriptor 순서(생존 판정의 전제, design D2)가 리뷰 대상이 된다.

## T1-3 철회 (Fix 라운드, 2026-08-14) — `Start`는 편집한다

**위 T1-3의 무변경 선언을 철회한다.** A1 적대 리뷰가 정확히 그 "건드리지 않기로 한
순서"에서 결함을 찾았다: `net.Listen`과 `os.Chmod` 사이는 원자적이지 않고, 그 사이의
죽음이 umask가 정한 0700 socket을 최종 이름에 남긴다. 회수의 정확-0600 검사가 그것을
영구 거부했다(A1 F1, 실측 3/3). **무변경 결정이 지키려던 순서가 사고의 생산자였다.**

```text
Pre-Edit Gate (Fix 라운드):
- change id / task id: a108 / tasks 6.1~6.6 (design D1-2·D2-2·D4-2)
- 대상 심볼(패키지.함수):
    internal/strategyprojectionrpc.Start (socket 발행 두 줄 → listenPrivateSocket 한 호출)
    internal/strategyprojectionrpc.writeDescriptor (제자리 O_EXCL 쓰기 → stage+rename)
    internal/strategyprojectionrpc.reclaimStaleControlDirectory (staging 분류 · 판정 순서 ·
      socket perm 완화 · descriptor 형식/내용 분리 · rmdir ErrNotExist 대칭)
    internal/strategyprojectionrpc.Server.Close (listener 소유권)
    internal/strategyprojectionrpc.Dial (connect probe)
    internal/strategyprojectionrpc.readDescriptor (형식 검사를 openVerifiedDescriptor로 분리)
- 기존 동작 파악 근거:
    FLM 6건을 **편집 후 재생성**(analysis/function-logic/, revision current).
      start 14 · reclaim 16 · close 5 · dial 3 · writedescriptor 12 · readdescriptor 5.
    A1 review.md §1 F1·F2·F5·F6의 실측.
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: yes — a108_publication_is_total_test.go 를 GREEN보다 먼저
    커밋했다(RED 5건 전부 실패 확인: stale socket is unsafe / stale descriptor is unsafe ×2 /
    unexpected entries / Dial 성공).
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** 완화는 하나뿐이고 그 폭이 측정된다.
    (1) socket perm 검사를 정확-0600 → perm&0o077==0으로 좁게 완화한다. group·other
        비트가 하나라도 있으면 여전히 거부이고(뮤테이션 M11이 그 폭을 잰다), 나머지
        조건(우리 uid·0700 디렉터리·비symlink·nlink 1·**아무도 수락하지 않음**)은
        전부 그대로 요구한다.
    (2) descriptor는 **형식**을 그대로 요구한다(0600 정규 파일·no-follow·inode 동일성).
        완화되는 것은 **내용**뿐이고, 그것은 D2 이후 어떤 판정에도 쓰이지 않는다.
    (3) 회수 대상 이름에 `.staging-` 접두를 더한다. 그 이름은 **우리만** 만들고,
        최종 이름을 가진 적이 없으므로 아무도 읽거나 연결하지 않았다.
    (4) `Dial`의 probe는 검사를 **더하는** 방향이다. 실패를 첫 Read에서 Dial로 앞당길
        뿐 새 성공 경로를 만들지 않는다.
    (5) 절 단위 판정 함수 추출(controlDirectoryModeIsSafe·ownedByEffectiveUser·
        sameDescriptorFile)은 호출부의 조건을 **그대로** 옮긴 것이고, 각각을 뒤집는
        뮤테이션(M18·M19·M20·M21)이 전부 죽는 것으로 등가성을 확인했다.
```

## T1-4. 다른 세 endpoint (tasks 2.5) — 테스트만 추가

`internal/app/engine/`의 policy command·policy runtime·alert control transport는
**production 코드를 편집하지 않는다.** crash-shape 핀(새 _test.go 파일)만 추가하고,
핀이 결함을 드러내면 거기서 멈추고 Manager에 보고한다(tasks 2.5 문구 그대로).
