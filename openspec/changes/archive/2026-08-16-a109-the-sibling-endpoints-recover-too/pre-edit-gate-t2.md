# a109 High-risk Pre-Edit 선언 — T2 (기동 강등 · 소비자 재부착)

`docs/WORKFLOW.md:387-401`의 형식을 따른다. T1의 선언(`internal/positionpolicyrpc`,
`internal/app/engine`)과는 별개 문서다 — 파일 소유가 겹치지 않는다.

T2가 닿는 경로는 **엔진 기동 시퀀스**와 **조회 전용 데몬의 화면 경로**다. 주문·손절·
사이징·Guardian·원장 코드는 건드리지 않는다. 그럼에도 High-risk 로 다루는 이유는
하나다: 기동이 실패하면 보호 루프(reconcile·exit·filldetect)가 전부 서지 못하고,
autostart 앞에서 그것이 **영구 기동 루프**가 된다 — 2026-08-13 23:35 사고의 실제 피해다.

착수 전 실측(D5b, tasks 3.3 몫)은 **편집 전에** 끝냈다: 구버전(=현재 HEAD) 세 Start 가
신버전 잔재 모양(`.s-` 정규 파일·socket·pre-chmod 0700 socket·현행 CreateTemp staging)
에서 18/18 START OK 였고 잔재는 그대로 남았다. 코드 근거도 같은 방향이다 —
`internal/positionpolicyrpc`·`internal/app/engine` 의 **production 코드에 `os.ReadDir`
호출이 하나도 없다**(테스트 3건뿐). 구버전은 신버전 잔재를 **볼 수 없다.**

---

## 1. `cmd/tossctl.runEngineRun` — §2.1·§2.2 (세 형제 fatal 강등)

```text
Pre-Edit Gate:
- change id / task id: a109 / tasks 2.1, 2.2
- 대상 심볼(패키지.함수): cmd/tossctl.runEngineRun (engine.go:183-328, AST 분기 20개)
- 기존 동작 파악 근거:
    FLM openspec/changes/a109-.../analysis/function-logic/cmd-tossctl--runenginerun/
      (ast.json: branches 20 · returns 13 · defers 9 · source_sha256
       8111c1c9e20f501b6221e231836fb02d7d03d127b3592892175c1beb38788381)
    편집 대상 분기: **B15(:273) · B16(:278) · B20(:314)** — 각각
      `engine.StartPositionPolicyCommandServer`(:272) · `StartPositionPolicyRuntimeServer`(:277)
      · `StartAlertControlServer`(:313) 오류의 `return err`(:274 · :279 · :315).
      design D3 표가 지목한 세 자리이고 proposal 의 :294 표기는 freeze P1-9 가 정정했다.
    **편집하지 않는 이웃**: B14(:269 `NewPositionPolicyCommandService`)와
      B19(:310 `ectx.AlertOperations()`)는 endpoint 기동이 아니라 in-process 조립이므로
      design D3 표에 없다 — fatal 유지가 계약이다. 이 둘을 같이 강등하면 「표면 없이
      부팅」이 아니라 「서비스 없이 부팅」이 되어 강등 논증의 전제가 깨진다.
    영향받는 defer: :276 `policyControl.Close` · :281 `policyRuntime.Close` ·
      :317 `alertControl.Close` — 강등 시 수신자가 nil 이 될 수 있으므로 a108 이 B17 에서
      쓴 nil 가드 형태를 그대로 쓴다.
    기존 테스트: cmd/tossctl/engine_test.go(기동 순서), a102_ready_wiring_test.go,
      engine_runtime_branch_test.go — **셋 다 `engineRuntimeFactory` 에서 멈춘다**
      (errStubRuntimeReached). 즉 B15·B16·B20 을 실행하는 기존 테스트는 없다.
      a108_the_engine_outlives_its_read_endpoint_test.go 는 B15·B16·B20 을 **성공**
      경로로 지난다(진짜 Start 를 쓴다) — 그것이 이 편집의 회귀 감시자다.
    호출부: newEngineRunCmd 의 RunE 하나 (`tossctl engine run`).
- upstream 상속 테스트 영향: no (TossOS 고유 부팅 시퀀스)
- 실패 테스트 선행 작성: **yes**, 두 단계다.
    (a) 무행위 리팩터: 세 Start 를 package var seam 으로 뺀다
        (`engineStrategyProjectionStart`(engine.go:416) 관례 그대로 — freeze P1-5).
        `cli_testseams.go` 관례는 인용하지 않는다: 그 파일은 internal/app/engine,
        즉 T1 표면이다.
    (b) 그 seam 에 실패를 주입하는 RED 를 쓰고 **현재 코드가 엔진을 죽이는 것**을 먼저
        관측한 뒤 강등으로 GREEN 을 만든다. 잔재 디스크 상태로 실패를 만들지 않는다 —
        그러면 T1 이 지금 편집 중인 회수 규칙에 이 테스트가 묶인다.
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** 통과 조건 다섯:
    1. 강등은 **표면 부재**이지 접근 확대가 아니다 — 어떤 토큰·경로·권한도 넓어지지 않는다.
    2. 엔진 싱글턴 권위(1단계 journal flock, B3)를 건드리지 않는다.
    3. 강등 보고에 **critical 등급 · obs 등급표 등재 · 원장 outbox** 중 무엇도 쓰지
       않는다(engine.go:330–359 주석이 정본). 미전달 PENDING 행은 다음 부팅의 진입
       게이트를 잠근다.
    4. 보고는 부팅을 붙잡지 않는다 — stderr 만 동기, 발행은 `WithoutCancel` + goroutine.
    5. defer **등록 순서**를 바꾸지 않는다: `lock.Release`(:200)가 가장 먼저 등록돼
       가장 나중에 돌고, 그래서 모든 endpoint Close 가 flock 을 쥔 채로 돈다. §2.4 의
       정적 핀이 이것을 고정한다.
    잃는 것에 대한 정직한 기록(freeze P0-4): policy command 강등은 **격리 해제 표면**을
    잃는다. 격리 해제는 격리된 포지션의 손절 포함 미판정 상태를 푸는 유일한 장중
    경로이므로, 강등은 「아무것도 느슨해지지 않는다」가 아니라 **「격리된 포지션의
    무보호가 유지된다」**이다. fatal(전 포지션 무보호)보다 엄격히 낫다는 비교로만
    이 판정이 성립한다.
```

## 2. `cmd/tossctl.reportStrategyProjectionDegraded` — §2.2 (보고 일반화)

```text
Pre-Edit Gate:
- change id / task id: a109 / tasks 2.2
- 대상 심볼(패키지.함수): cmd/tossctl.reportStrategyProjectionDegraded
  (engine.go:369-405, AST 분기 1개)
- 기존 동작 파악 근거:
    FLM .../cmd-tossctl--reportstrategyprojectiondegraded/ (branches 1 · returns 1 ·
      go 1 · source_sha256 8111c1c9…)
    부작용은 둘뿐이다: 동기 stderr 한 줄, `WithoutCancel` goroutine 안의 obs Normal
      이벤트. 원장·디스크에 닿지 않는다 — 그래서 `ectx.Close()` 와 겹쳐도 안전하다.
    기존 테스트: TestAFailedStrategyProjectionDoesNotStopTheEngine ·
      TestTheDegradedBootWritesNoUndeliveredOutboxRow ·
      TestASecondDegradedBootLeavesTheNextBootsEntryGateUnlatched ·
      TestTheDegradedBootDoesNotWaitForTheNotifier (전부 a108, cmd/tossctl).
    호출부: runEngineRun B18 하나(engine.go:304). 일반화 후 B15·B16·B20 이 추가된다.
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: yes — 세 형제 강등이 **어느 표면인지 말한다**는 것을 먼저
  RED 로 요구한다(문구·obs scope 구분). 「강등했다」만 재면 「전부 projection 이라고
  찍는다」가 통과한다.
- 안전 불변식 §0 위반 여부 검토: **조건부 통과.** 금지 3종(§1 조건 3)과 비차단
  구조(§1 조건 4)를 유지하는 한 통과. 이름은 일반화하되(`reportEngineEndpointDegraded`)
  기존 obs 이벤트 타입 값 `engine.strategy_projection_unavailable` 은 **바꾸지 않는다** —
  형제용으로는 등급표에 없는 새 타입을 하나 더 둔다. 등급표 미등재가 Normal 판정의
  전부이므로(`obs.SeverityOf`), 새 타입은 자동으로 Normal 이고 테스트가 그것을 핀한다.
```

## 3. `cmd/tossctl.strategyRuntimeReaderFor` — §2.3 (소비자 재부착)

```text
Pre-Edit Gate:
- change id / task id: a109 / tasks 2.3
- 대상 심볼(패키지.함수): cmd/tossctl.strategyRuntimeReaderFor
  (httpapi.go:254-286, AST 분기 4개)
- 기존 동작 파악 근거:
    FLM .../cmd-tossctl--strategyruntimereaderfor/ (branches 4 · returns 4 ·
      source_sha256 37c2ed8990337096cb98973cba78f1d0e9cc61abade8df115106d754761ef2df)
    반환 3종: nil(부재→dormant) · unavailableStrategyRuntime(sentinel→unavailable) ·
      live client. **세 값 모두 부팅 1회로 굳는다** — 재시도 경로가 없다.
    소비자 nil 검사 **3곳**(설계는 2곳만 열거했다 — issues.md T2-1):
      ① cmd/tossctl/httpapi_reader.go:566 (집계 스냅샷)
      ② internal/httpapi/router.go:154 (REST /api/v1/strategy-runtime — **설계 누락**)
      ③ internal/httpapi/strategy_runtime.go:18 (SSE helper, 현재 production 호출자 없음)
    기존 테스트: TestADialFailureRendersUnavailableRatherThanNotConfigured(세 모양의
      화면 값) · TestADeadDescriptorDoesNotStopTheDaemon ·
      TestAnUninspectableDescriptorDegradesLikeTheConsole ·
      TestASocketFileWithNoOwnerDegradesTheDaemon ·
      TestAnAbsentDescriptorAndADeadOneBootTheSame · 그리고
      internal/httpapi/strategy_runtime_contract_test.go 의 REST/SSE 대비.
    호출부: runHTTPAPI(httpapi.go:147) 하나이고 그 값을 **두 소비자**에 같이 꽂는다
      (`resources.reader.strategyRuntime`, `httpapi.NewRouter(Options{StrategyRuntime})`).
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: yes — 냉부팅 순서(엔진 나중)와 가동 중 재시작(새 socket·새
  토큰) 두 시나리오를 RED 로 먼저 세운다. 하나만 재면 P0-1 이 지적한 구멍(live 부착
  후 재시작)이 다시 열린다.
- 안전 불변식 §0 위반 여부 검토: **통과.** 이 데몬은 조회 전용이고 주문·손절·원장 어디에도
  닿지 않는다. 지켜야 할 것 셋:
    1. 요청 goroutine 에서 `strategyprojectionrpc.Dial`·connect probe 를 부르지 않는다
       (Dial 본문에 200ms probe 가 있다 — transport_unix.go:402-424).
    2. 부재(dormant)/실패(unavailable) 화면 값을 접지 않는다 — wrapper 는 정의상
       non-nil 이므로 세 nil 검사를 **하나의 공유 판정**으로 교체한다(복사한 검사는
       어긋나기 시작한 검사다, a098 D7.1).
    3. 실패 반복을 stderr 에 찍지 않는다 — 보고는 상태 전이 시 1회.
```

## 4. 문구 정직화 — §2.5 (`engine_alerts_client_unix.go`, `internal/console/exit_quarantine.go`)

```text
Pre-Edit Gate:
- change id / task id: a109 / tasks 2.5 (design D3a-2, freeze P0-3)
- 대상 심볼: cmd/tossctl.errEngineAlertsUnavailable (engine_alerts_client_unix.go:33-35)
  · internal/console.Console.handleQuarantineRelease{,Apply}·writeQuarantineError 의
  refuse 본문 문자열 (exit_quarantine.go:161 · :196 · :229)
- 기존 동작 파악 근거:
    두 메시지 모두 강등을 **다른 원인으로 단정한다**: 「엔진이 없다」(엔진은 돌고 있다),
    「control plane 이 제공하지 않는다」(빌드가 낡았다는 뜻). 운영자는 그 말을 따라
    엔진을 재시작하고, 원인이 결정적이므로 같은 강등이 재현된다.
    같은 문자열이 exit_quarantine.go 에 **세 벌** 있다(161·196·229). 설계는 229 만
    인용했다 — 값 단위로 고친다(issues.md T2-2).
    기존 테스트: internal/console/a079_quarantine_release_test.go:298 이 **제목**
      "판정 격리 command seam 미배선" 만 핀한다. 본문은 미핀 — 이번에 핀을 깐다.
      errEngineAlertsUnavailable 을 이름으로 핀하는 테스트는 없다.
- upstream 상속 테스트 영향: no
- 실패 테스트 선행 작성: yes — 메시지 텍스트 핀 테스트를 먼저 쓴다.
- 안전 불변식 §0 위반 여부 검토: **통과.** 문구만이다. 새 화면·새 디스크 상태·새 마커
  없음. 사용자 지시(확인 문구 금지)와도 무관하다 — 타이핑 확인을 더하지 않는다.
```

---

## 표면 밖으로 나가지 않는다는 선언

T2 가 쓰는 파일은 `cmd/tossctl/{engine.go, httpapi.go, httpapi_reader.go,
engine_alerts_client_unix.go}` + cmd/tossctl 의 a109 신규 테스트,
`internal/httpapi/{strategy_runtime.go, router.go}`,
`internal/console/exit_quarantine.go`(문구만)뿐이다. `internal/positionpolicyrpc/*`
와 `internal/app/engine/*` 는 **읽기·import 만** 한다.

`internal/httpapi/router.go` 는 설계의 T2 표면 목록에 없다 — 왜 넣었는지는
`issues.md` T2-1 에 사후 기록한다(설계가 열거한 nil 검사 2곳 밖에 production REST
경로의 세 번째가 있었고, 그것을 고치지 않으면 http-api-service delta 의
「부재·unavailable 구분 유지 SHALL」이 REST 경로에서 거짓이 된다).

## 선언된 생략 (not-applicable)

- **콘솔 lifecycle client 재부착**(console.go:397): design 의 선언된 생략 그대로.
  같은 병이지만 이 change 는 httpapi 만 고친다.
- **`templates_position_policy.go` 의 "배선되지 않아 조회만 가능"**: 원인 단정이
  아니므로 유지(D3a-2).
- **B14·B19 의 fatal**: endpoint 기동이 아니므로 D3 표 밖. 강등하지 않는다.
