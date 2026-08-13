# a108 High-risk Pre-Edit 선언 — T2 (겹2·겹3)

`docs/WORKFLOW.md:387-401`의 형식을 따른다. T1의 선언(`internal/strategyprojectionrpc`)과는
별개 문서다 — 파일 소유가 겹치지 않는다.

이 change가 T2 쪽에서 닿는 경로는 **엔진 기동 시퀀스**다. 주문·손절·사이징·Guardian 코드는
건드리지 않는다. 다만 기동이 실패하면 보호 루프(reconcile·exit·filldetect)가 전부 서지 못하고,
그것이 2026-08-13 23:35 사고의 실제 피해였으므로 High-risk로 다룬다.

---

## 1. `cmd/tossctl.runEngineRun` — 겹2 (조회 endpoint 실패는 엔진을 죽이지 않는다)

```text
Pre-Edit Gate:
- change id / task id: a108 / tasks 3.1, 3.2
- 대상 심볼(패키지.함수): cmd/tossctl.runEngineRun (engine.go:183, AST 분기 19개)
- 기존 동작 파악 근거:
    FLM openspec/changes/a108-.../analysis/function-logic/cmd-tossctl--runenginerun/
      (ast.json: branches 19 · returns 14 · defers 9 · source_sha256
       ee527a6a917ab5342bdc4e853d6a8d015a9b56534c1971096220f2288a520d3f)
    편집 대상 분기: **B17 (engine.go:283)** — `strategyprojectionrpc.Start` 오류의
      `return err`(returns[11], line 284). 사고가 죽은 바로 그 줄이다.
    영향받는 defer: defers[6] `strategyRuntime.Close` (line 286) — 강등 시 수신자가 nil이 된다.
    기존 테스트: cmd/tossctl/engine_test.go (기동 순서 8건),
      cmd/tossctl/a102_ready_wiring_test.go (step6→step7 ready seam),
      cmd/tossctl/engine_runtime_branch_test.go (engineRuntime 구성 분기)
      — **이 셋 전부 `engineRuntimeFactory`에서 멈춘다(errStubRuntimeReached).**
        즉 B17을 실행하는 기존 테스트는 **하나도 없다**. 새 harness가 필요하다.
    호출부: newEngineRunCmd의 RunE 하나 (`tossctl engine run`).
- upstream 상속 테스트 영향: no (TossOS 고유 부팅 시퀀스)
- 실패 테스트 선행 작성: **yes**, 단 두 단계다.
    (a) 무행위 리팩터: `var engineStrategyProjectionStart = strategyprojectionrpc.Start`
        seam 추가(engineAssemble·engineRuntimeFactory와 같은 var seam 관행).
        동작 변화 없음 — 기존 테스트 전부 GREEN 유지로 확인한다.
    (b) 그 seam에 실패를 주입하는 RED 테스트를 쓰고, **현재 코드가 exit 1 하는 것**을
        먼저 관측한 뒤 강등으로 GREEN을 만든다.
    실제 잔재 디스크 상태로 실패를 만들지 않는다 — 그쪽은 T1이 병행 편집 중인
    `internal/strategyprojectionrpc`의 동작에 의존하게 되어 두 세션이 서로를 깨뜨린다.
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** 통과 조건 다섯:
    (1) **싱글턴 권위를 옮기지 않는다.** 엔진 배타성은 부팅 1단계의 journal flock
        (`enginelock.Acquire`, engine.go:196)이지 projection 디렉터리가 아니다.
        강등 기동이 둘째 엔진을 허용하면 안 된다 — tasks 3.2 ①로 고정한다.
    (2) **a102 ready 신호의 발행 시점을 옮기지 않는다.** ready는 step 7의 Recover가
        끝난 뒤에만 발행된다(a102 D5). 강등이 그 시점을 앞당기면 콘솔이 "대사 끝났다"를
        거짓으로 읽는다 — tasks 3.2 ②로 고정한다.
    (3) **automation interlock 평가 순서를 옮기지 않는다.** gate-off·interlock 미충족은
        projection보다 **먼저** 거절되어야 한다. 강등 코드가 그 앞으로 새어 나가면
        게이트가 꺼진 계좌에서 socket이 생긴다 — tasks 3.2 ③으로 고정한다.
    (4) **강등을 조용히 하지 않는다.** 실패는 durable critical 알림(원장 outbox)으로
        남긴다. 기동 시 stderr 한 줄만 찍는 형태는 「1회 유실형」이고, 그러면 이 change가
        만드는 것은 회복이 아니라 **은폐**다.
    (5) **판정 로직을 `runEngineRun` 안에 두지 않는다.** B17의 새 처리는 인자를 받는
        별도 함수에 두고 여기에는 호출만 남긴다 — a101이 배운 것과 같은 이유
        (기동 함수 안의 판정은 harness 없이는 어떤 테스트도 부르지 못한다).
```

### 강등이 안전한 이유 (이중 writer 논증)

design D3의 논증을 이 선언의 언어로 다시 쓴다. projection은
`strategyprojection.Reader`를 내보내는 **조회 전용 export**다
(`internal/app/engine/strategy_runtime_projection.go:14` — `Context` 자신을 돌려준다).
루프의 입력이 아니다. 따라서 강등으로 잃는 것은 **화면**이고 **판정**이 아니다.
그리고 배타성은 (1)의 flock이 이미 강제하므로, projection 실패를 무시하고 진행해도
두 번째 엔진이 생길 수 없다.

반대 방향 — 관측 표면 하나 때문에 손절을 든 프로세스를 죽이는 것 — 이 a102가 이미
지운 비대칭이고 2026-08-13 사고가 그 재발이다.

---

## 2. `cmd/tossctl.runHTTPAPI` — 겹3 (부재와 실패를 같은 강등으로)

```text
Pre-Edit Gate:
- change id / task id: a108 / tasks 3.3
- 대상 심볼(패키지.함수): cmd/tossctl.runHTTPAPI (httpapi.go:120, AST 분기 21개)
- 기존 동작 파악 근거:
    FLM openspec/changes/a108-.../analysis/function-logic/cmd-tossctl--runhttpapi/
      (ast.json: branches 21 · source_sha256
       4b30b64c47e1d70052d0d1cc4f35126fe51d3543ec36dbe9bb2b624d0aaeb4c8)
    편집 대상 분기: **B9 (httpapi.go:151)** — `Dial` 실패의 `return`(line 152).
    유지 대상 분기: **B10 (httpapi.go:155)** — 비-NotExist `os.Stat` 오류는 fatal 유지.
    대조 분기: B7(line 149, stat 성공)·B8(else, descriptor 부재 → 강등)이 이미 설계된
      강등이며, 이 편집은 B9를 B8과 같은 결과로 맞추는 것이다.
    기존 테스트: cmd/tossctl/httpapi_test.go — **runHTTPAPI를 한 번도 실행하지 않는다**
      (명령 등록·boundary·flag만 잰다). B9를 실행하는 기존 테스트는 없다.
    호출부: newHTTPAPICmd의 RunE 하나 (`tossctl httpapi`).
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: **yes** — descriptor는 있고 socket이 없는 tmpdir fixture로
    현재 `runHTTPAPI`가 오류 반환하는 것을 먼저 고정한다. seam 불필요(디스크 상태만으로
    `Dial`이 자체 검사로 거절하며, 이는 T1이 편집하는 `reclaimStaleControlDirectory`와
    무관한 `Dial`의 socket 검사 경로다).
- 안전 불변식 §0 위반 여부 검토: **통과.** httpapi는 조회 전용 daemon이고
    (`Annotations["mutating"] != "true"`), 이 편집은 그 daemon이 **뜨는가**만 바꾼다.
    주문·손절 경로에 닿지 않는다. 다만 (4)와 같은 이유로 강등은 경고 로그를 남긴다.
    B10을 fatal로 유지하는 것이 조건이다 — 권한 오류까지 강등하면 「조사 불가능한 환경」과
    「정상적 부재」를 구별할 수 없게 된다(design D4).
```

---

## 3. `cmd/tossctl.httpAPIReader.Snapshot` — 겹3 Fix (집계는 전략 하나로 죽지 않는다)

Fix 라운드(tasks 6.8②)가 **새로 편집 대상에 넣은 기존 함수**다. 원 라운드의 선언에는
없었다 — A2 가 이 함수를 사고 표면으로 지목하기 전이었기 때문이다.

```text
Pre-Edit Gate:
- change id / task id: a108 / tasks 6.8 ②
- 대상 심볼(패키지.함수): cmd/tossctl.httpAPIReader.Snapshot
    (httpapi_reader.go:450, 편집 전 AST 분기 9개 · return 9개)
- 기존 동작 파악 근거:
    FLM openspec/changes/a108-.../analysis/function-logic/cmd-tossctl--httpapireader.snapshot/
      (ast.json: branches 10 · returns 8 · defers 0 · source_sha256
       7abe45c66b68edbcb716f97feacbd2330bd2dbbde9c3e053ba66f7c67200397b — 편집 후)
      편집 전 판(분기 9·return 9)은 커밋 d8b27021 의 소스로 재추출해 확인했다.
    편집 대상 분기: **B9** (구 B9, httpapi_reader.go:481-484) — 전략 `Read` 오류의
      `return nil, err`. 여섯 조회가 전부 성공했는데도 집계 전체를 비우던 줄이다.
    만든 분기: **B10** (else) — 성공 값 대입을 명시 분기로 분리했다.
    유지 대상: **B1~B7** — 자기 저장소(원장·브로커·성과·최적화) 실패는 fail-closed
      그대로다. 이 비대칭이 이 편집의 전부다.
    기존 테스트: cmd/tossctl/httpapi_reader_test.go 는 `Positions`/`Orders`/투영
      헬퍼만 잰다. `Snapshot` 을 실행하는 기존 테스트는 **하나도 없다** —
      httpapi_snapshot_test.go 는 캐시(`httpAPISnapshotCache`)를 재고 이 함수를
      부르지 않는다. 새 fixture 가 필요하다.
    호출부: `newHTTPAPISnapshotCache(resources.reader)` 하나 (httpapi.go).
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: **yes** — 커밋 d8b27021 이 테스트만 담고, 그 커밋에서
    `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` 가
    「집계 스냅샷이 실패했다: … socket is invalid」로 죽는 것을 관측했다.
    같은 커밋에서 대조군(`...StaysDormantRatherThanUnavailable`)은 통과했다.
- 안전 불변식 §0 위반 여부 검토: **통과.** 이 함수는 조회 전용 데몬의 직렬화
    경로이고 주문·손절·사이징·Guardian 어디에도 닿지 않는다. 변경 방향은
    「화면이 더 자주 보인다」이며 **거짓으로** 보이는 방향이 아니다 — 실패는
    지워지지 않고 `RUNTIME_UNAVAILABLE` 이라는 이름으로 스냅샷 안에 실린다.
    조건 하나: 읽기 실패(RUNTIME_UNAVAILABLE)와 reader 부재(NOT_CONFIGURED)를
    **접지 않는다.** 접으면 운영자는 「엔진이 죽었다」와 「기능을 안 켰다」를
    구별할 수 없고, 그것은 화면을 살린 것이 아니라 거짓말을 살린 것이다.
```

---

## 두 함수의 계약이 Fix 라운드에서 바뀐 것 (선언 정정)

위 1절의 통과 조건 **(4)** 은 「강등을 조용히 하지 않는다 — 실패는 durable critical
알림(원장 outbox)으로 남긴다. stderr 한 줄만 찍는 형태는 「1회 유실형」이고, 그러면
이 change 가 만드는 것은 회복이 아니라 **은폐**다」였다.

**A2 적대 리뷰가 그 조건을 뒤집었다(design D3-2).** durable outbox 행은
`UndeliveredCount`(Type 무필터) → `restoreAlertEntryLatch` 를 타고 **다음 부팅의 진입
게이트를 잠근다**. publisher 미설정 배포에서는 영구 PENDING 이므로, 화면 하나를
잃었다는 보고가 실계좌의 신규 진입을 영구 차단한다. obs 교리가 금지하는 모양이다.

새 조건 (4'): **강등을 조용히 하지 않되, 그 보고가 게이트에 배선된 rail 을 타지
않는다.** 보고는 ① stderr 기동 경고, ② obs **Normal** 이벤트 로그, ③ 콘솔·httpapi 의
dormant 전략 화면(상시 표면)이다. ③ 이 「1회 유실형」걱정의 실제 답이다 — 원 선언은
그 표면이 이미 있다는 것을 세지 않았다.

2절의 통과 조건 중 「B10 을 fatal 로 유지하는 것이 조건이다」도 같은 라운드에서
뒤집혔다(D4-2, A2 F4): ENOTDIR 볼륨 오배치가 조회 데몬을 다시 crash loop 에 넣는다.
새 조건: **전략 endpoint 때문에 이 데몬이 죽는 경로는 하나도 없다.** 「환경 이상을
삼키지 않는다」는 요구는 fatal 이 아니라 **구별되는 경고 문구**가 진다.

---

## 이 선언이 명시적으로 포기하는 것

- httpapi의 lazy 재-dial: design D4가 범위 밖으로 선언했다. descriptor-부재로 뜬 httpapi도
  지금 그렇게 남으며, 대칭을 유지하는 쪽이 새 비대칭을 만드는 것보다 낫다.
- 강등 상태의 **콘솔** 표시: proposal의 Non-goals가 콘솔 무변경을 선언했다.
- ~~알림 event type을 `internal/obs`의 등급표에 등록하는 것: T2 소유 파일이 아니다.
  아래 "설계 이견"에 이유와 대안을 적는다.~~
  **정정(Fix 라운드, D3-2):** 이것은 「소유 경계 때문에 못 한 것」이 아니라 **해서는
  안 되는 것**이다. 등급표 등재는 이 이벤트를 outbox rail 에 태우고, 그 rail 의 끝이
  다음 부팅의 진입 차단이다. 미등재가 **의도된 설계**이며 그 사실을
  `engineStrategyProjectionDegradedEvent` 의 주석이 진다. 뮤테이션 M3b 가 그 결정을
  지킨다 — 등재된 critical 이름으로 바꾸면 두 테스트가 죽는다.
- `strategyprojection.Validate` 를 집계 스냅샷에서 부르는 것(콘솔은 부른다):
  6.8② 의 편집 범위를 「Read **오류** 흡수」로 좁힌 결과다. 성공했지만 형식이 깨진
  스냅샷의 처리는 이 change 이전과 같다 — 선언된 생략.
