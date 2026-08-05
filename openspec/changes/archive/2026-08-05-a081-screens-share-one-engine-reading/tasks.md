# a081 · Tasks

> **상태 2026-08-05 (독립 리뷰 반영 후).** 3인 독립 리뷰가 P1 7건을 냈고 그중
> 셋이 설계를 바꿨다 — 8장이 그 반영이다. 아래 1~7장의 체크는 당시 기록이며,
> 8장이 뒤집은 항목은 그 자리에 표시했다.
>
> **2026-08-05 land·배포 완료.** `df4407ed`가 `main`에 있고 그 코드로 만든
> `tossos@sha256:b083bc72…`가 두 service에 올라가 있다. 6.6 컨테이너 실측은 사람이
> 승인한 배포 뒤에 끝냈다 — 렌더 20회 묶음에서 엔진에 닿은 렌더는 1회였고 엔진
> `reconcile` 사이클은 62.6~63.7초로 흔들리지 않았다. 기록은 review.md 4차.

## 1. 근거 고정 (편집 전)

- [x] 1.1 `decoratePositionRows`가 렌더마다 `Runtime`·`List`를 호출하고 그 사이에
      캐시가 없다는 것을 현재 HEAD에서 확인해 기록한다.
- [x] 1.2 두 호출이 실제로 엔진 프로세스에 닿는다는 것을 확인한다 —
      `Runtime`은 descriptor 재해석 + 매 호출 dial, `List`는 엔진의 단일 쓰기
      커넥션 위의 SELECT.
- [x] 1.3 exit 루프가 그 커넥션을 쓰고 `작업 후 sleep` 구조라는 것을 확인한다
      (뺏은 시간이 판정 간격에 더해지는 근거).
- [x] 1.4 캐시가 실어 나를 세 값이 무엇에 따라 움직이는지 코드에서 확인한다.
      **결론이 틀렸다 → 8.2가 정정.** adoption effective는 생성자 고정이 맞지만
      blocks는 **살아 있는** reconcile tracker 상태이고, `Runtime`은 DB에 닿지도
      않는다. "셋 다 렌더 단위로 안 움직인다"는 이 change의 원래 전제였다.
- [x] 1.5 보호선의 값 자체가 `c.positions()`의 journal 읽기에서 온다는 것을
      확인한다 — 이 change가 신선도를 건드리지 않는 근거.
- [x] 1.6 `decoratePositionRows`의 Function Logic Map과 Branch Test Map을
      **편집 전에** 작성한다. `check_analysis.py` 통과.
- [x] 1.7 기존 요구사항과의 충돌 여부를 확인한다 — `rate budget 보호`(브로커),
      "편입 보조 상태…"(runtime unavailable SHALL), "포지션 관리는 실제 adoption
      desired/effective를 구분한다"(명령 화면).

## 2. RED — 지금은 세지 못하는 것을 세게 만든다

a080의 예산 테스트가 F1을 놓친 이유는 하네스가 `PositionPolicies`를 배선하지
않아 이 경로를 한 번도 실행하지 않았기 때문이다. 먼저 그 눈을 만든다.

- [x] 2.1 `PositionPolicies`를 배선하고 `Runtime`·`List` 호출 수를 세는 테스트
      하네스를 만든다. 캐시 없이 N회 렌더하면 각각 N이 나오는 것을 먼저 확인한다
      (RED 관측 기록).
- [x] 2.2 간격 하나 동안 여러 번 렌더해도 엔진 읽기가 각각 1회를 넘지 않는다.
- [x] 2.3 그 테스트가 무의미하게 통과할 수 없음을 자체 검사한다 — 렌더 횟수가
      허용 읽기 수보다 많지 않으면 `t.Fatal`.
- [x] 2.4 두 화면(`/positions`·`/dashboard`)이 같은 간격을 공유한다 — 화면마다가
      아니라 간격마다 1벌.

## 3. GREEN — 캐시

- [x] 3.1 `internal/console/position_policy_cache.go`에 `positionPolicyCache`와
      간격 상수를 둔다. 인터페이스는 `Runtime`·`List` 둘뿐이다 (design D7 —
      Preview·Apply를 명명할 수 없다는 것이 타입의 성질). **상수는 하나가 아니라
      둘이다 → 8.2.**
- [x] ~~3.2 저장 단위는 한 덩어리이고 갱신은 둘을 함께 읽는다.~~ **철회 → 8.3.**
      journal이 세 번째 시점이라 "한 시점의 짝"은 애초에 없었다. 두 절반은 각자의
      간격으로 만료되고, 보존되는 것은 불일치의 **방향**이다.
- [x] 3.3 마지막 **결과**를 서빙한다 — 실패는 실패로 캐시되고 직전 성공이
      되살아나지 않는다 (design D3).
- [x] 3.4 뮤텍스를 읽기 전체에 건다 — 동시 렌더 여러 건이 읽기 1벌 (design D1).
      **테스트가 없었다 → 8.6이 추가.**
- [x] 3.5 슬라이스는 복사해서 내보낸다 (`holdingsCache.snapshotLocked` 선례).
- [x] 3.6 `decoratePositionRows`가 커맨더 대신 캐시를 읽는다. 그 아래 분기 구조는
      바뀌지 않는다.

## 4. 무효화

- [x] 4.1 성공한 정책 Apply 뒤 캐시를 버린다.
- [x] ~~4.2 성공한 격리 해제 뒤 캐시를 버린다.~~ **철회 → 8.4.** 격리 상태는 이
      캐시를 지나지 않는다. 무효화는 엔진 읽기 1회를 더 사고 존재하지 않는 data
      flow를 주장하는 주석을 남길 뿐이었다.
- [x] 4.3 실패한 mutation은 버리지 않는다 (design D5).
- [x] 4.4 위 셋을 테스트로 고정한다.

## 5. 안 바뀌는 것

- [x] 5.1 읽기가 신선할 때 두 화면의 렌더 결과가 오늘과 같다.
- [x] 5.2 엔진 읽기 실패 시 렌더가 오늘과 같다 (effective unknown).
- [x] 5.3 `/position-management`와 설정 탭이 캐시를 거치지 않는다 (design D6).
- [x] 5.4 기존 `internal/console` 테스트가 손대지 않고 통과한다. 갱신이 필요한
      테스트가 나오면 그 이유를 `issues.md`에 남긴다.

## 6. VERIFY

- [x] 6.1 변이 검증: 캐시의 간격 gate를 제거하면 2.2가 RED가 되는지 확인하고
      되돌린다.
- [x] 6.2 변이 검증: 실패 시 직전 성공을 되살리게 바꾸면 3.3의 테스트가 RED가
      되는지 확인하고 되돌린다.
- [x] 6.3 변이 검증: 무효화를 지우면 4.1이 RED가 되는지 확인하고 되돌린다.
- [x] 6.4 `make test`, `make vet`, `make validate`, `make sdd-sync`, `make sdd-check`.
- [x] 6.5 `make gate CHANGE=a081-screens-share-one-engine-reading`.
- [x] 6.6 사람 승인 후 컨테이너 실측 — 탭 2개를 열어 둔 상태에서 엔진 프로세스로
      들어가는 읽기 횟수가 간격당 1벌인지, 엔진 사이클 시간이 나빠지지 않는지.
      **2026-08-05 배포 후 실측 완료.** 결과는 review.md 4차.

## 7. 리뷰와 기록

- [x] 7.1 proposal-freeze 리뷰.
- [x] 7.2 발견 사항을 `issues.md`에 남긴다 (같은 형태의 결합이 `StrategyRuntime`·
      `Settings.Load`에도 있는지 포함).
- [x] 7.3 PM story/tracker 동기화.
- [x] 7.4 별도 컨텍스트의 독립 리뷰 3인 (WORKFLOW.md §역할 분리) — 안전·테스트·
      스펙 렌즈. **P0 0건, P1 7건.** 결과와 처분은 review.md 3차.
- [x] 7.5 a080 재개 판단 — a081의 예산 테스트가 F1의 경로를 실제로 실행한 것을
      확인한 뒤 a080의 `[blocked]`를 해제하고 F2~F8을 정리한다. **해제한다.**
      근거는 셋이다 — 예산 테스트가 `PositionPolicies`를 배선해 F1의 경로를 실제로
      실행하고(I5), a081이 `df4407ed`로 land했으며, 배포 실측이 렌더 20회당 엔진
      도달 1회를 보였다(6.6). 문서 지적 F3·F4·F5·F8은 `15d25f80`에서 정리했고
      코드 지적 F2·F6·F7은 a080 8장이 코드 복원과 함께 처리한다.

## 8. 독립 리뷰 반영 (2026-08-05)

- [x] 8.1 갱신을 요청 컨텍스트에서 분리한다 (`context.WithoutCancel` + 자체
      타임아웃). 리뷰어 둘이 재현한 오염을 막는다. design D4b.
- [x] 8.2 두 읽기의 간격을 분리한다 — `List` 30초(쓰기 커넥션을 다툼),
      `Runtime` 5초(DB 미접촉, 살아 있는 대사 차단을 실어 나름). design D2.
- [x] 8.3 D4("한 시점의 짝")를 철회하고, 보존 대상을 불일치의 **방향**으로
      옮긴다. journal이 세 번째 시점이라는 사실을 명시한다.
- [x] 8.4 격리 해제 무효화를 제거한다 — 격리는 이 캐시를 지나지 않는다. 이유를
      코드 주석으로 남기고 그 테스트도 제거한다. design D5.
- [x] 8.5 lifecycle 절반을 실제로 검증한다 — 픽스처 `PositionID`를
      `pos-managed`로 정합하고 실패 시 `엔진 관리` → `관리 여부 불명` 전환을
      단정한다.
- [x] 8.6 동시성 SHALL에 테스트를 붙인다 (`TestConcurrentRendersCostOneReading`).
      읽기 창을 넓히는 hook으로 변이가 실제로 RED가 되게 한다.
- [x] 8.7 자체 검사의 죽은 코드(`const` 비교)를 간격에서 유도하도록 고친다.
- [x] 8.8 거부 테스트를 긍정 단정으로, 두 화면 공유 테스트에 내용 단정을 더한다.
- [x] 8.9 `Runtime.Blocks` 슬라이스를 복사해서 내보내고 테스트한다.
- [x] 8.10 스펙 면제 기준을 "명령을 발행하는 화면"에서 **"자동 재로드가 없는
      화면"**으로 바꾸고, 나이 상한·취소 금지·상한의 무효화 예외를 명문화한다.
- [x] 8.11 인용 줄번호 4건과 거짓 주석 2건을 정정한다.
- [x] 8.12 변이 검증을 7종으로 늘리고 전부 RED를 관측한다.
- [x] 8.13 Function Logic Map 4건의 AST를 재생성하고 본문을 개정 내용에 맞춘다.
      `check_analysis.py` 통과.
- [x] 8.14 `make test`(전 패키지 ok, 0 failed)·`vet`(clean)·`validate`(exit 0)·
      `sdd-sync`·`sdd-check`(exit 0) 재실행. 배포 직전에 한 번 더 돌렸다.
      `sdd-sync`에서 CodeGraph hard evidence는 갱신됐고 CodeGraphContext(timeout)와
      GBrain(다른 프로세스가 점유)은 갱신되지 않았다 — 둘 다 advisory라
      `sdd-check`는 통과한다.
