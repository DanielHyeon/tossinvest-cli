# a109 구현 중 발견한 계약 결함

분류 규칙: ① blocking(안전·동작 모순) → 구현 중단 + 보고 종료 · ② safe local → 구현하며
사후 기록 · ③ editorial → 즉시 수정.

## T2-1 (safe local) — D4 가 열거한 nil 검사 2곳 밖에 production REST 경로가 하나 더 있다

- **발견**: design D4 와 freeze P1-4 는 소비자측 nil 검사를 두 곳으로 열거했다 —
  `cmd/tossctl/httpapi_reader.go:566` 과 `internal/httpapi/strategy_runtime.go:18`.
  실제로는 **세 곳**이고, 셋 중 SSE helper 쪽(`strategy_runtime.go:18`)은
  `StrategyRuntimeSnapshotFunc` 의 것인데 **production 호출자가 없다**
  (`rg StrategyRuntimeSnapshotFunc` = 정의 1 + contract test 1).
  열거에서 빠진 `internal/httpapi/router.go:154` 가 REST `/api/v1/strategy-runtime`
  경로의 진짜 부재 판정이다:

  ```go
  case "strategy-runtime":
      if r.strategyRuntime == nil {
          return strategyprojection.DormantSnapshot(r.now().UTC()), nil
      }
  ```

- **왜 결함인가**: 재부착 wrapper 는 정의상 non-nil 이므로 이 검사가 영원히 거짓이 된다.
  전략 화면을 안 쓰는 배포에서 REST 응답이 **dormant 스냅샷 → 오류**로 바뀐다. 그것이
  http-api-service delta 의 「재부착 전의 응답 값은 기존 부재·unavailable 구분을 그대로
  유지해야 한다(SHALL)」를 REST 경로에서 거짓으로 만든다. a108 D4-2 가 금지한 접힘의
  같은 모양이다.
- **분류**: safe local. 설계 의도(부재/unavailable 구분 보존)를 **바꾸지 않고 완성**하는
  기계적 동일 수정이다 — 세 자리 모두 같은 공유 판정 하나로 교체한다.
- **처리**: `internal/httpapi/router.go` 를 T2 표면에 더하고(1줄), 판정은
  `internal/httpapi/strategy_runtime.go` 에 **한 벌만** 둔다
  (`StrategyRuntimeAbsent` + `StrategyRuntimePresence`). 복사한 검사는 어긋나기 시작한
  검사다(a098 D7.1) — 세 곳에 세 벌을 두지 않는다.
- **Manager 확인 요청**: T2 표면 목록에 `internal/httpapi/router.go` 추가.

## T2-2 (editorial) — 격리 해제 오귀속 문구는 한 곳이 아니라 세 곳이다

- **발견**: design D3a-2 는 `internal/console/exit_quarantine.go:227–229` 한 자리를
  인용했다. 같은 문자열이 **세 벌** 있다 — `:161`(release preview),
  `:196`(release apply), `:229`(`writeQuarantineError` 의 `ErrUnwired` 가지).
- **왜 결함인가**: 인용된 file:line 만 고치면 같은 값의 사본 둘이 살아남고, 운영자가
  실제로 먼저 만나는 것은 preview 경로(:161)다.
- **분류**: editorial — 정정의 단위는 줄이 아니라 **값**이다. 즉시 세 곳 모두 고친다.
- **처리**: 세 문자열을 같은 문구로 바꾸고, 핀 테스트가 **세 경로 전부**를 확인한다.

## T2-3 (editorial) — proposal 의 fatal 줄번호는 freeze 판결대로 :274/:279/:315 다

- proposal 본문의 `:274/:279/:315` 는 현재 HEAD 와 일치함을 AST 로 재확인했다
  (`analysis/function-logic/cmd-tossctl--runenginerun/ast.json` returns[8]=274,
  returns[9]=279, returns[11]=315). freeze P1-9 가 정정한 :294 표기는 남아 있지 않다.
  **수정 불요** — 확인만 기록한다.

## T2-5 (safe local) — 재부착 wrapper 의 로그는 두 goroutine 이 함께 쓴다

- **발견**: 첫 `-race` 전체 실행이 `strategyRuntimeAttachment` 의 로그 writer 에서 데이터
  경합을 잡았다 — 부착 전이는 **시도 goroutine**이, 탈착 전이는 **요청 goroutine**이
  말하는데 둘 다 잠금 없이 `fmt.Fprintf(a.log, …)` 였다.
- **왜 결함인가**: 설계는 「보고는 상태 전이 시 1회」만 정했고 그 보고가 **두 goroutine
  에서 나온다**는 것은 D4 의 백그라운드 시도 결정이 만든 결과다. 설계에 없던 상태다.
- **분류**: safe local — 구현이 만든 경합이고 수정도 구현 안에서 끝난다.
- **처리**: 로그 전용 잠금(`logMu`)과 `report` 한 곳. 상태 잠금(`mu`)을 쓰지 않은 이유는
  느린 writer 하나가 모든 요청의 상태 조회를 멈추게 하면 안 되기 때문이다.

## T2-6 (safe local) — a098 테스트가 alerts 문구의 부분 문자열을 핀하고 있었다

- **발견**: `cmd/tossctl/a098_the_operator_command_names_a_person_test.go:341`
  (`TestTheCommandsRefuseWhenNoEngineIsRunning`)가
  `strings.Contains(err.Error(), "엔진이 없다")` 를 요구한다. D3a-2 의 첫 문구 시안
  (「…엔진이 없거나, …강등 부팅했다」)은 그 부분 문자열을 깨뜨려 이 테스트가 **실제로
  실패**했다.
- **왜 중요한가**: a098 의 의도는 「운영자가 자기 경로를 의심하게 만들지 마라」이고,
  a109 의 요구는 「엔진 부재를 **단정**하지 마라」다. 둘은 충돌하지 않는다 — 충돌한 것은
  단정문을 통째로 지운 첫 시안이었다.
- **분류**: safe local (문구 재작성으로 해소, 기존 테스트는 손대지 않았다).
- **처리**: 문구를 **조건문 두 갈래**로 바꿨다 — 「엔진이 없다면 `engine run` 을 살려라 —
  엔진이 돌고 있다면 강등 부팅한 것이고 원인은 엔진 로그에 있다」. 단정이 사라지면서
  부재일 때의 행동 안내(a098 의 요구)는 그대로 남는다.
- **남는 결합**: ~~a098 은 여전히 부분 문자열로 이 문구에 묶여 있다.~~ **해소됨
  (§2-fix F6)** — a098 의 단정을 `errors.Is(err, errEngineAlertsUnavailable)` 로
  바꿨다. 이제 a098 은 거부의 **정체**를, a109 는 그 거부가 무슨 말을 해야 하는지를
  각자 잰다. 의도가 계속 측정되는지는 M27b 로 확인했다(sentinel 을 날것의 오류로
  되돌리면 a098 이 FAIL).

## T2-7 (safe local) — 표면 밖 신규 테스트 파일 둘

- `internal/httpapi/a109_absence_is_a_state_not_a_nil_test.go` (T2-1 의 REST 자리를 잰다)
- `internal/console/a109_a_degraded_surface_is_not_an_old_build_test.go` (§2.5 문구 핀)
- 둘 다 **신규 파일**이고 기존 파일을 고치지 않는다. 그 패키지의 코드를 바꾼 이상 핀도
  그 패키지에 있어야 한다 — cmd/tossctl 에서 문자열로 흉내 내면 그것은 동작이 아니라
  소스를 재는 테스트가 된다. T1 표면(`internal/positionpolicyrpc`,
  `internal/app/engine`)과는 겹치지 않는다.

## T2-8 (safe local) — 재부착 wrapper 는 `cmd/tossctl` 의 **새 파일**에 둔다

- **발견**: wrapper 를 `httpapi.go` 안에 두자 `check_analysis.py` 가 삽입 지점 바로 위의
  `unavailableStrategyRuntime.Read` 를 「수정됨」으로 잡았다 — 아무도 편집하지 않은
  함수의 증거 묶음을 요구한다(`--unified=0` 의 인접 hunk 가 그 함수의 줄 범위에 닿는다).
- **분류**: safe local. a102·a108 이 같은 이유로 세운 관례(「새 코드는 새 파일에」)를
  따른다.
- **처리**: `cmd/tossctl/httpapi_strategy_attach.go` 신설. `httpapi.go` 에는 부팅 1회
  해석(`resolveStrategyRuntimeReader`)과 그것을 감싸는 `strategyRuntimeReaderFor` 만
  남는다. T2 표면 목록에 없는 **신규 production 파일**이므로 여기 기록한다.

## T2-9 (기록) — 뮤테이션이 테스트 결함 둘을 잡았다

- M10(single-flight)과 M22(SSE 부재 판정)가 **생존**했고, 둘 다 「테스트가 다른 이유로
  초록」이었다. 테스트를 고친 뒤 재측정해 CAUGHT 로 바꿨다 — 상세는
  `mutation-ledger-t2.md` §5.
- 이것이 원장의 실제 수확이다. 두 테스트는 이 change 안에서 **한 번도 실패한 적이
  없었고**, 그래서 그 초록은 증거가 아니었다.

## T2-4 (기록) — D5b 실측은 편집 전에 끝냈고 결과는 "사전 wipe 불요"다

- tasks 3.3 은 §3 소속이라 T2 가 체크하지 않는다. 측정 결과는 완료 보고에 담고
  Manager 가 배포 절차 prose 에 반영한다.
- 요약: 구버전(HEAD, pre-a109) 세 Start × 잔재 6모양 = 18/18 START OK, 잔재는 그대로
  생존. 코드 근거: `internal/positionpolicyrpc`·`internal/app/engine` production 코드에
  `os.ReadDir` 호출 0건(테스트 3건뿐) — 구버전 회수는 디렉터리를 열거하지 않으므로
  `.s-*` 잔재를 **볼 수 없다.**

## T2-10 (정산) — A2 적대 리뷰 P2 여덟 건의 §2-fix 처리

Manager 판결(2026-08-15) 후 브랜치 `feat/a109-t2-degrade-and-redial` 위에서 F1~F9 로
구현했다. RED 가 의미 있는 항목은 RED 커밋을 선행했다.

| 항목 | A2 발견 | 처리 | 측정 |
|---|---|---|---|
| F1 | P2-1 요청자 ctx 취소를 endpoint 실패로 읽는다 | `observe` 가 취소 계열 오류 + **호출자 ctx 종료**를 상태 무변경으로 가린다 | RED 실측(시도 1회·탈착 보고·live client 교체) → GREEN, 뮤테이션 M29 CAUGHT |
| F2 | P2-2 늦은 read 실패가 새 부착을 뒤엎는다 | 읽기가 쓴 **자리의 세대**(seat)가 현재일 때만 전이 | RED 실측(탈착 보고·재-dial 2회) → GREEN, M30 CAUGHT |
| F3 | P2-3 재부착 wake 가 집계 성공에 얹혀 있다 | publisher 루프가 `Refresh` 밖에서 부재 판정을 먼저 부른다(주기 유지·새 ticker 없음) | RED 실측(2초 안에 시도 0회) → GREEN, M31 CAUGHT |
| F4 | P2-4 주석 핀의 공허한 단정 2개 | 고유 구절로 교체 | M24·M25 생존(재현) → M24b·M25b·M26 CAUGHT |
| F5 | P2-5 격리 거부의 **제목**이 빌드를 탓한다 | `quarantineUnwiredTitle` 상수 1벌 + a079 핀을 값 단위로 | RED(세 경로 FAIL) → GREEN, 핀 둘 |
| F6 | P2-6 a098 의 우연한 문자열 결합 | `errors.Is(err, errEngineAlertsUnavailable)` | M27b CAUGHT(의도 유지 확인) |
| F7 | P2-10 T1 병합 대비 nil-safe 계약 | 소비자 쪽 계약 테스트 1개 | M28 CAUGHT |
| F8 | P2-7 부작용 있는 술어 | `StrategyRuntimePresence`·`StrategyRuntimeAbsent` 주석에 명문화 | 문서 |

## T2-11 (잔존 선언) — `exitquarantine.ErrUnwired` 문자열은 고치지 않았다

`internal/exitquarantine/model.go:49` 의

```go
ErrUnwired = errors.New("exit quarantine: the engine control plane does not offer quarantine release")
```

는 F5 가 고친 **운영자 화면 문구**와 같은 오귀속("control plane 이 제공하지 않는다" =
사실상 「빌드가 낡았다」)을 담고 있다. 그럼에도 손대지 않았다:

- 이 값은 화면에 **도달하지 않는다**. 콘솔은 `errors.Is` 로 갈래를 고르고 자기 제목·
  본문(`quarantineUnwiredTitle`·`quarantineUnwiredDetail`)을 쓴다. 운영자가 이 문자열을
  보는 곳은 로그·디버그 출력뿐이다.
- `internal/exitquarantine` 은 T2 표면이 아니고, sentinel 문자열 변경은 그 값을 문자열로
  잡는 다른 패키지의 테스트를 건드릴 수 있다.

**잔존 위험**: 로그를 읽는 사람은 여전히 「control plane 이 제공하지 않는다」를 본다.
후속 change 후보로 남긴다.

## T2-12 (기록) — publisher 서명이 하나 늘었다

`publishHTTPAPISnapshots` 에 `strategyRuntime httpapi.StrategyRuntimeReader` 인자를
더했다(F3). 호출자는 둘뿐이다 — `runHTTPAPI` 와 그 테스트. 재부착 wake 를 집계 밖으로
꺼내려면 루프가 그 값을 **알아야** 하고, 대안(캐시가 source 를 되돌려주게 하기)은
`httpAPISnapshotCache` 에 이 change 와 무관한 결합을 새로 만든다.

## T2-13 (판결 이행 방식 보고) — F2 의 "같은 reader 인가"는 **세대 번호**로 물었다

Manager 판결의 문면은 「`Read` 가 사용한 reader 를 `observe` 에 넘겨 현재 `a.reader` 와
**동일할 때만** 전이」다. 구현은 reader 값을 넘기되 비교를 `==` 가 아니라 **자리의 세대
번호**(`seat uint64`)로 한다. 이유는 안전이다:

- 인터페이스 값의 `==` 는 두 동적 타입이 같고 그 타입이 **비교 불가**면 런타임 패닉이다.
- `httpapi.StrategyRuntimeReader` 의 구현 중 값 타입이 스냅샷을 들고 있으면 정확히 그
  경우가 된다 — `strategyprojection.Snapshot` 은 `map[Market]MarketProjection` 을 들고
  있고, 이 저장소 안에 이미 그 모양의 reader 가 있다(테스트의 `a109Projection`).
- 즉 문면대로 쓰면 「전략 읽기가 실패했을 때 패닉하는 조회 데몬」이 될 수 있다.

세대 번호는 같은 질문에 더 강한 답을 준다: 같은 값이 **다시 앉은** 경우도 새 자리로
세므로, 옛 실패는 언제나 버려진다(보수 방향). 판정 자체는 바뀌지 않는다.
