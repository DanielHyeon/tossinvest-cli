# a109 High-risk Pre-Edit 선언 — T1 (`internal/*`)

`docs/WORKFLOW.md`의 "Pre-Edit 선언 (High-risk 전용)" 형식을 따른다. 이 파일은 T1 몫이다
(표면: `internal/positionpolicyrpc/*`, `internal/app/engine`의 세 transport). T2는
`pre-edit-gate-t2.md`를 따로 쓴다.

이 change가 닿는 경로는 **엔진 기동**이다. 주문·손절·사이징 코드는 건드리지 않지만,
기동이 거부되면 autostart가 영구 기동 루프에 들어가 **장중 손절이 서지 못한다**
(2026-08-13 사고 실측). 면제 대상이 아니다.

작성 시점: §1a 완료 직후, 첫 편집 **전**. FLM 맵 17개는 이 선언보다 먼저 완성했다
(`python3 tools/logic-map/check_analysis.py --change a109-the-sibling-endpoints-recover-too`
가 T1 슬러그 17개에 대해 오류 0 — 남은 3건은 T2 표면).

---

## T1-1. 두 socket endpoint의 발행 — 최종 경로 bind → staged bind

```text
Pre-Edit Gate:
- change id / task id: a109 / tasks 1.1·1.4 (design D1·D1a·D1b)
- 대상 심볼(패키지.함수):
    internal/app/engine.StartAlertControlServer          (AST 분기 15 — B9·B10·B11 구간)
    internal/app/engine.StartPositionPolicyRuntimeServer (AST 분기 16 — B9·B10·B11 구간)
    internal/positionpolicyrpc.<신규 파일>               (staged listen — 새 코드, FLM 불요)
- 기존 동작 파악 근거:
    FLM analysis/function-logic/internal-app-engine--startalertcontrolserver/
        · internal-app-engine--startpositionpolicyruntimeserver/ (AST 열거, 손으로 읽은
      분기가 아니다). 두 함수 모두 B9(Prepare*Socket) → B10(net.Listen 최종 경로) →
      B11(os.Chmod 0600) 순서이고, B10과 B11 사이가 pre-chmod 잔재의 생산 구간이다.
    기존 테스트:
      TestTheAlertSocketIsPrivateToThisUser (0700 디렉터리·0600 socket·0600 descriptor 실측)
      TestRuntimeOnlyUnixControlCrossNamespaceTopologyAndRouteAbsence (같은 실측 + 라우트)
      TestAlertControlServerStartsOverALeftover · TestPositionPolicyRuntimeServerStartsOverALeftover
        (a108 2.5 관용 핀 — 유지해야 한다)
    호출부: cmd/tossctl/engine.go:277·:313 각각 하나씩(T2 표면).
    이식 원형: internal/strategyprojectionrpc/transport_unix.go listenPrivateSocket —
      원형은 수정하지 않는다(a108 소유).
- upstream 상속 테스트 영향: no — 세 endpoint 전부 TossOS 고유다(upstream tossinvest-cli에 없다).
- 실패 테스트 선행 작성: yes — §1.1(pre-chmod 0700 socket)·§1.2(산 주인)·§1.3(staging·낯선
    엔트리)을 **GREEN보다 먼저 각각 커밋**하고 실패를 관측한다. 잔재는 umask 조작이 아니라
    **명시적 chmod 0700**으로 만든다(umask는 프로세스 전역이라 병렬 테스트를 오염시킨다 —
    freeze QA①).
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** 통과 조건 넷:
    (1) 검증을 지우지 않는다. staged bind 뒤에도 발행 확인(ValidatePrivateSocket /
        ValidateRuntimeSocket, 둘 다 **정확-0600**)을 그대로 부른다. 그 호출이 남아 있는
        것이 "완화가 회수 밖으로 새지 않았다"의 증거다.
    (2) staging basename 은 11자 고정이다(`.s-`+종류1+hex7). alerts.sock 이 11자이므로
        a108의 12자를 그대로 쓰면 "staging ≤ 최종" 계약이 깨진다. 길이는 §1.5가 직접
        센다(절대 경로 103 요구는 넣지 않는다 — Linux 실측 상한 107).
    (3) SetUnlinkOnClose(false)를 반드시 건다. 걸지 않으면 listener 가 자기가 기억하는
        이름을 늦게 unlink 하고, 그 늦은 정리가 후계자의 socket 을 지운다(A1 F5).
    (4) 최종 이름은 chmod 0600 을 지난 socket 에만 붙는다.
        ⛔ **정정(2026-08-15, §1-fix F1).** 이 조건은 처음에 "회수의 owner-write 사망
        판정이 이식 적법한 근거"로 쓰였다. 그 판정 자체가 틀렸으므로 근거의 용도가
        바뀐다: 지금 이 사실이 뒷받침하는 것은 **probe 전 chmod 0600 이 안전하다**는
        것이다(발행 계약상 최종 socket 은 원래 0600 이므로 chmod 는 우리 계약의 복원이고,
        결과 권한은 어떤 경우에도 owner 전용이다).
```

## T1-2. 두 socket endpoint의 회수 — Prepare*Socket → 전체성 회수 + 사망 증명

```text
Pre-Edit Gate:
- change id / task id: a109 / tasks 1.2·1.3·1.4 (design D2·D2a·D2b)
- 대상 심볼(패키지.함수):
    internal/app/engine.StartAlertControlServer          (B9 호출 교체 + ErrExist 경로)
    internal/app/engine.StartPositionPolicyRuntimeServer (같음)
    internal/app/engine.StartPositionPolicyCommandServer (staging 위생 한 줄 추가, D2a)
    internal/positionpolicyrpc.<신규 파일>               (회수·probe — 새 코드)
  **편집하지 않는 것(선언):**
    internal/positionpolicyrpc.PreparePrivateSocket · PrepareRuntimeSocket — 본문을 고치지
      않고 기동 경로에서 뺀다. 둘 다 기존 테스트를 가진 공개 API이고
      (TestPrepareRuntimeSocketNeverDeletesNonSocket), 고치면 그 완화가
      statPrivateSocket 을 통해 클라이언트 검증까지 번진다(freeze P1-3).
    internal/positionpolicyrpc.statPrivateSocket · ValidatePrivateSocket ·
      ValidateRuntimeSocket · ValidatePrivateControlDirectory ·
      ValidateRuntimeControlDirectory · validatePrivateDirectory — 전부 무변경.
      회수 완화는 **회수 전용 신규 함수 안에만** 존재한다.
    internal/app/engine.PositionPolicyCommandServer.Close — 무변경(TCP라 지울 socket이 없다).
- 기존 동작 파악 근거:
    FLM 8개(prepare×2 · stat · validate socket×2 · validate dir×3) + start×3의 AST 열거.
      PreparePrivateSocket: 분기 3 · return 4 — **생존 판정 절이 없다**(마지막 줄이
      무조건 os.Remove). PrepareRuntimeSocket: 분기 5 · return 6 — 같은 모양.
    a108 review §1 F3·F4 실측(각 3회 반복).
    이식 원형: reclaimStaleControlDirectory · verifyStaleSocketShape ·
      projectionSocketAccepts (strategyprojectionrpc, 수정 금지).
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: yes — §1.2(산 주인 위 두 번째 Start)·§1.3(staging 잔재 + 낯선
    엔트리)을 GREEN 전에 커밋한다.
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** 통과 조건 여섯:
    (1) **완화는 하나뿐이다** — 잔재 socket 의 perm 을 정확-0600 → `perm&0o077 == 0`.
        group/other 비트가 하나라도 있으면 여전히 거부다. 소유 uid · 비symlink ·
        nlink 1 · 0700 디렉터리는 전부 그대로 요구한다.
    (2) **완화는 회수 전용 함수 안에만 있다.** 클라이언트 검증 두 개는 정확-0600 유지.
        뮤테이션 원장이 "클라이언트 검증을 완화하면 테스트가 죽는가"를 확인한다.
    (3) **수락 중인 socket 은 이름과 무관하게 절대 unlink 하지 않는다.** 제거 전
        connect probe 를 하고, 사망으로 읽는 것은 **둘뿐**이다 — ECONNREFUSED · ENOENT.
        그 밖의 오류(타임아웃 등)는 생존으로 간주하고 **거부**한다. PID 는 판정에 쓰지
        않는다(컨테이너 재생성이 PID 를 재배정한다, a102 D4b-2).
        ⛔ **정정(2026-08-15, §1-fix F1·F5).** 처음에는 셋이었고("owner 쓰기 비트 없음"이
        셋째), 대상도 **최종 이름의 socket 뿐**이었다. 둘 다 A1 이 반증했다:
          · owner-write 절은 묻는 대신 **추정**이었고, 쓰기 비트가 깎인 **수락 중인**
            socket 을 지웠다(P1-A). 지금은 probe 전에 0600 으로 chmod 하고 묻는다 —
            EACCES 가 사라지므로 산 socket 과 죽은 socket 이 결정적으로 갈린다.
          · staging 이름의 socket 은 probe 없이 지웠다(P2-D). 발행의 첫 걸음이 staging
            이름에 bind 하는 것이므로, 그 창의 후계자 socket 이 정확히 그 모양이다.
            지금은 같은 probe 를 건다.
    (4) **우리가 만들지 않은 것은 치우지 않는다.** 아는 이름은 호출자가 넘기고, 낯선
        엔트리가 하나라도 있으면 아무것도 건드리지 않고 거부한다(socket endpoint 한정).
        staging 엔트리는 모양(정규 파일|socket)에 더해 **소유 uid**도 검증한다(P1-7③).
    (5) **command endpoint 에는 새 실패 경로를 만들지 않는다.** 열거+거부를 넣으면 이물
        하나가 격리 해제 표면을 매 부팅 지운다(freeze P1-2). 위생은 자기 staging 접두만
        보고, 그 밖의 것은 오늘처럼 무시하며, 실패로 기동을 막지 않는다.
    (6) **회수 실패를 성공으로 삼키지 않는다.** os.Remove 오류는 ErrNotExist(경합 양성)만
        용인한다. 1차 방어는 여전히 journal flock 이고, 코드 주석이 그것을 명시 인용한다.
```

## T1-3. Close의 listener 소유권

```text
Pre-Edit Gate:
- change id / task id: a109 / tasks 1.4 (design D2b)
- 대상 심볼(패키지.함수):
    internal/app/engine.AlertControlServer.Close          (AST 분기 3 — listener 필드 자체가 없다)
    internal/app/engine.PositionPolicyRuntimeServer.Close  (AST 분기 3 — 필드는 있는데 안 쓴다)
    internal/app/engine.AlertControlServer (타입 — listener 필드 추가)
- 기존 동작 파악 근거:
    FLM 2개의 AST 열거: 두 Close 모두 Shutdown → os.Remove 루프(descriptor→socket→dir) →
      ErrNotExist 관용. listener.Close() 호출이 **없다**(AST calls 에 부재).
    기존 테스트: TestTheAlertSocketIsPrivateToThisUser · 
      TestRuntimeOnlyUnixControlCrossNamespaceTopologyAndRouteAbsence 가 Close 후 세 경로의
      부재를 확인한다 — 이 계약을 깨면 안 된다.
    a108 A1 F5 실측(300라운드 중 3회 재현).
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: no — 이 절은 **경합 방어**라서 결정적 RED 를 만들 수 없다
    (a108 도 300라운드 중 3회로 관측했다). 대신 ① 기존 Close 계약(세 경로 부재)이
    유지되는지 ② SetUnlinkOnClose(false) 로 경로 제거 권한이 Close 루프 하나뿐인지를
    뮤테이션으로 잰다. **RED 선행 없음을 여기 선언한다**(침묵한 생략 금지).
- 안전 불변식 §0 위반 여부 검토: **통과.** 더하는 것은 "우리가 직접 닫는다"뿐이고,
    제거 루프·제거 순서·ErrNotExist 관용은 그대로다. net.ErrClosed 는 성공과 같게 읽는다
    (이미 Shutdown 이 닫은 경우).
```

## T1-4. staging 접두의 상수화 (descriptor 발행 3벌)

```text
Pre-Edit Gate:
- change id / task id: a109 / tasks 1.4 (design D1b 완전성)
- 대상 심볼(패키지.함수):
    internal/app/engine.publishPrivateDescriptor            (B2 — ".endpoint-*")
    internal/app/engine.writePositionPolicyRuntimeDescriptor (B3 — ".position-policy-runtime-*")
    internal/app/engine.writePositionPolicyDescriptor        (B3 — ".position-policy-control-*")
- 기존 동작 파악 근거:
    FLM 3개의 AST 열거(분기 14·17·17). 세 함수 모두 이미 stage+rename 이고 **병이 없다** —
    바꾸는 것은 os.CreateTemp 의 접두 리터럴을 상수 참조로 만드는 것 하나뿐이다.
    이유: 접두가 발행과 회수에 따로 적히면 회수가 자기 잔재를 낯선 것으로 거부한다.
    descriptor 발행 3벌의 fold 는 **선언된 생략**이다(design 말미) — 하지 않는다.
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: yes — §1.3 의 staging 잔재 회수 RED 가 이 이름들을 쓴다.
    추가로 §1.5 의 이름-집합 완전성 테스트가 "각 발행 경로가 실제로 만드는 이름이 그
    transport 가 넘기는 집합에 속함"을 잰다(freeze P2-5).
    ⛔ **정정(2026-08-15, §1-fix F2 · A1 P1-B).** 위 줄은 처음에 그 측정을 "각 발행 경로의
    **생성기를 실제로 돌려서**"라고 적었다. **과장이었다.** 실제로는 이렇다:
      · socket staging — 발행이 쓰는 그 생성기(`StagedSocketName`)를 실제로 돌린다. 참.
      · descriptor staging — 테스트가 **상수를 가져다 자기가** `os.CreateTemp` 를 돌린다.
        발행 경로는 실행되지 않는다. 그래서 발행 쪽이 리터럴로 갈라져도 통과한다 —
        A1 의 뮤테이션 A1-N1(`.endpoint2-`)이 정확히 그렇게 살아남았다.
    "발행이 이 상수를 쓴다"는 **동적으로 잴 수 없는 성질**이므로 정적으로 잰다:
    `TestEveryPublishedStagingPrefixIsTheSharedConstant` 가 go/parser 로 세 transport 를
    파서, 파일을 만드는 호출(`os.CreateTemp`·`os.OpenFile`)의 이름 인자에 접두 리터럴이
    있으면 실패시키고, 그 상수가 선언·발행·회수 세 자리에서 읽히는지 센다. A1-N1 을
    재적용해 사망을 확인했다(원장 §D).
- 안전 불변식 §0 위반 여부 검토: **통과.** 검증·rollback·sync 순서를 건드리지 않는다.
    리터럴 위치만 바뀌고 만들어지는 이름은 같다.
```

---

## 선언된 무변경 (not-applicable)

- `internal/strategyprojectionrpc/**` — a108 확정 코드. 읽기만 하고 수정하지 않는다
  (staging 12자 통일도 하지 않는다 — `runtime.sock` 12자에 적법).
- `cmd/tossctl/**`, `internal/httpapi/**`, `internal/console/**` — T2 표면.
- descriptor 발행 3벌의 fold, `publishPrivateDescriptor`의 rename 직전 재검증 비대칭(P2-6) —
  병이 없는 표면의 High-risk 리팩터링(design 말미의 선언된 생략).
