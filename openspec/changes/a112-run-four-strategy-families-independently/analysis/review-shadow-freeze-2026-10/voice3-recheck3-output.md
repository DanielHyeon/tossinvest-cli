# SHADOW re-freeze 재검 3라운드 voice3 — SubagentHandback 원문 (agent ac4fbb2d05a76746e, 저장 2026-10-05)

**3라운드 재검(보이스 ③) 결과: FAIL.** 남은 5항목 중 CLOSED 3개(P1-2, 새 P1-A, 새 P1-B), PARTIAL 2개(P1-1, FLM 목록)다. v3가 새로 연 P1이 1개 있다(아래 P1-C). P0은 없다.

**안전 규칙:** `~/.codex` 아래는 읽지도 검색하지도 않았다. 리뷰 디렉터리 안의 `_codex/`도 열지 않았다. 저장소 파일은 편집하지 않았다. 사본은 `…/scratchpad/v3r3`(`git archive 4d22d726`)에 만들었고, 이번 라운드에서는 사본에서 시험을 실행하지 않았다(코드 판독만).

**좌표:** `4d22d726`. 설계 브리프 v3 sha256 `f7883a47…30fc` 일치를 확인했다. `git diff --stat 2817064c 4d22d726 -- '*.go'` 결과 Go 변경은 0개이므로, 1·2라운드의 코드 증거와 재현은 이 좌표에서도 그대로 유효하다.

| 2라운드 항목 | 판정 | v3 좌표 | 근거 |
|---|---|---|---|
| P1-1 (선언 시장에서 OFF 레인만 제안한 범위 → `FAMILY_GATE_CLOSED` → shadow 관측 0) | **PARTIAL** | §4 운반 범위: 조정 뒤 닫힘 6개와 성공은 운반, 조정 앞 닫힘 7개는 「관측 없음」 | 2라운드에서 재현한 세 모양(REV만 제안, REV+BREAKOUT, 선언했으나 전부 OFF)은 모두 조정 뒤 닫힘인 `FAMILY_GATE_CLOSED`이므로 이제 운반된다. 주된 구멍은 닫혔다. **남은 것 하나:** 「계보 충돌」 닫힘은 `coordinateMarketProposals` 루프 **도중**에 `return`한다(`strategy_market_coordinator.go:113-115`, 루프는 :64). 그래서 v3 문구대로 「이미 모은 관문 앞 입력」을 싣게 되면 경로 순서상 뒤쪽 종목의 제안은 빠진 채 운반되고, 그 레인들은 `NO_INPUT`으로 거짓 보고된다. 이는 1차 P1-1과 같은 부류(없던 입력으로 보고)다. **닫는 법:** 충돌 갈래는 「관측 없음」으로 두거나, 관문 앞 수집을 루프 안 충돌 검사보다 먼저 끝까지 마친다. |
| P1-2 (재시작 핀의 빈 표본) | **CLOSED** | §10 공통 전제(앞 프로세스가 전체 주기에서 `WOULD_EMIT ≥ 1`, 재시작 뒤 전체 주기를 1회 이상 돈 뒤 측정), §8(cycle 클로저 전체, OFF 단독 종목 포함) | 2라운드에서 요구한 두 조건이 문면 그대로 들어갔다. |
| §7 FLM 목록 | **PARTIAL** | §4 FLM 확정 목록 | `collectMarket`이 무조건 포함됐고, `collect`와 cycle 클로저 두 곳도 들어갔다. 그러나 조립 자리를 잘못 가리킨다. 목록의 `refreshPairedStrategyEntryProductionAssembly(조립 필드)`는 캐시 합류만 하고 위임할 뿐이다(`strategy_entry_supervisor.go:575-596`). 실제로 `collect` 반환 모양이 바뀌어 편집될 곳은 `NewPairedStrategyEntryProductionAssembly`(:292-376)다. 이 함수는 `proposalAuthority`를 `ResultAuthority`·계좌·1차 레그·dispatch cycle·worker 생성에 나눠 주고, 조립 리터럴(:366)도 여기 있다. 즉 High-risk 조립 함수가 목록에서 빠졌다. 또 §5.1 신선도 규칙(shadow 파도 = 최신 evaluate 파도)을 적용하려면 `strategyLaneRuntime.projection`(`strategy_lane_projection.go:31`)이 관측을 조회하는 방식이 바뀐다. 이것도 목록에 넣거나 「무편집」 사유를 적어야 한다. |
| 새 P1-A (세탁 접근자가 census를 피함) | **CLOSED** | §2 ②(`types.Info.Uses`로 타입과 접근자 `*types.Func` 사용처를 세고, 본문 모든 식 타입을 보며, 세탁 접근자 변이를 양성 대조), §4(opaque `strategyworker.ShadowInput`은 `Input`이 아니므로 admit·Submit이 받을 수 없음; 운반은 authority 밖) | 2라운드에서 제안한 걸음과 양성 대조가 그대로 반영됐다. 운반 값이 `dispatchHandoffs`·`dispatchHandoff`·`authorityForOwnerScope`의 수신자 밖으로 나간 것도 확인했다. |
| 새 P1-B (shadow가 주문 경로 앞에서 지연을 만듦) | **CLOSED** | §5(cycle 함수 밖, `runProductionStrategyMarketCycle`이 돌아온 **뒤** 비동기·단일 비행·상수 마감, 분리된 ctx), AST 핀 ①②③, fault 핀(가짜 시계 마감+δ, 차등 궤적 동일) | 기존 마지막 문장 핀(`a112_market_delivery_structure_test.go:169-175`)을 유지하면서 dispatch 뒤로 옮겼다. |

### v3가 새로 연 P1
**P1-C. 비동기 shadow 단계가 같은 파도의 shadow 묶음을 어디서 얻는지 정해지지 않았다. 그리고 유일하게 존재하는 접근 경로는 새 원격 권한 파도를 일으킨다.**
- v3 §5에 따르면 「주기 함수는 그 필드를 읽지 않는다」. cycle 클로저가 받는 것은 `runProductionStrategyMarketCycle`의 오류뿐이다. 묶음이 든 조립(`fresh`)은 그 함수 안의 지역값이다(`strategy_entry_supervisor.go:515`).
- 클로저가 묶음을 얻는 기존 길은 `refreshPairedStrategyEntryProductionAssembly`뿐이다. 그런데 캐시 창은 1초다(`strategy_refresh_wave.go:73`: `now.Sub(c.strategyRefreshAt) < time.Second`).
- dispatch가 끝난 뒤 이 함수를 부르면 대개 창 밖이다. 그러면 지도자로서 `NewPairedStrategyEntryProductionAssembly` 전체가 돈다. 일정·후보·경로·FX·제안·위험·계좌 권한을 다시 원격으로 수집하고, 그 결과를 두 시장이 함께 쓰는 캐시에 게시한다(:83-99).
- 그 결과 shadow가 원격 I/O를 만들고, 다른 시장의 주문 주기가 합류할 권한 스냅숏을 바꾼다. 이는 §12의 「파일 I/O 0」, §5의 「레거시 경로에 아무것도 남기지 않는다」와 충돌한다. §8 차등 핀은 같은 입력을 쓰는 가짜 실행이므로 이 차이를 보지 못한다.
- 캐시 포인터(`c.strategyRefresh`)를 직접 읽는 방법도 있다. 다만 그 사이 다른 시장이 새 파도를 게시했을 수 있다. 그 경우 shadow 묶음과 레인 evaluate 파도가 어긋나고, §5.1의 「파도 번호」를 무엇에 묶는지가 문면상 정의되지 않는다. evaluate 파도는 레인 런타임이 시장마다 세는 수(`strategy_lane_runtime.go:322-338`)이고 조립의 신원과 무관하다.
- **닫는 법(셋 다 필요):**
  1. shadow 단계는 refresh·조립 함수를 부르지 않는다는 AST 핀(호출 폐포에 `refreshPairedStrategyEntryProductionAssembly`·`NewPairedStrategyEntryProductionAssembly`가 0개).
  2. 묶음을 얻는 경로를 명시한다. 예: 레인 런타임이 evaluate 시점에 같은 파도 번호와 함께 묶음을 값으로 보관하고, shadow 단계는 그것만 읽는다. 이렇게 하면 `runProductionStrategyMarketCycle` 본문에 shadow 식별자가 0이라는 핀과 충돌하지 않도록 evaluate 인자 경로를 설계해야 한다.
  3. 파도 묶기를 「묶음을 실어 온 evaluate 파도 번호」로 정의한다.

**1차·2차 P2 가운데 v3가 악화시킨 것:** 없다.

**저장소 무변경 확인:**
- 시작과 끝 모두 `rev-parse HEAD = 4d22d72663dcf3a42be928b3ecf3db817ffc6ac5`.
- `status --short`도 시작과 끝이 같다: ` M docs/ROADMAP.md`, ` M …/tasks.md`, `?? .reticle-setup-crash.log`, `?? …/review-shadow-freeze-2026-10/`, `?? …/shadow-2026-10/`, `?? w4.log`.

