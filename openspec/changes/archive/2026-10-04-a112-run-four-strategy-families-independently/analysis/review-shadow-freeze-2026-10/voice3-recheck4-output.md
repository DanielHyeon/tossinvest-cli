# SHADOW re-freeze 재검 4라운드 voice3 — SubagentHandback 원문 (agent ac4fbb2d05a76746e, 저장 2026-10-05)

**4라운드 판정 (보이스 ③, 좌표 4d22d726): FAIL.** 표적 3개 중 CLOSED 2개(P1-1, FLM 목록), PARTIAL 1개(P1-C)이고, v3.1이 새로 연 P1이 1개(P1-D) 있다. P0은 없다. 두 P1 모두 주문 경로에 노출 상승을 만들지 않고, 문장 한두 줄과 핀 하나씩이면 닫힌다.

**안전 규칙:** `~/.codex` 아래 파일은 읽지도 검색하지도 않았다. 리뷰 디렉터리의 `_codex/`도 열지 않았다. 저장소 파일은 편집하지 않았다. 코드는 기존 사본 `…/scratchpad/v3r3`(archive 4d22d726, 디렉터리 존재 확인)에서 판독만 했고, 이번 라운드에 사본 쓰기와 시험 실행은 없다.

**입력:** 브리프 v3.1 sha256 `c2ea7e821d87…671c` 일치. Go 코드는 2817064c 이후 변경 0이다.

| 항목 | 판정 | v3.1 좌표 | 근거 |
|---|---|---|---|
| P1-1 (계보 충돌 시 부분 묶음) | **CLOSED** | §4 운반 범위, 부재 값 문단, 구조 핀 ⓐⓑⓒ | 충돌 갈래는 「관측 없음」이며, 빈 묶음이 아니라 부재 값으로 정의됐다. 충돌 감지 여부와 상관없이 collectMarket은 coordinate가 돌려준 값을 그대로 싣는다. 그래서 충돌과 범위 지움이 함께 일어나 FAMILY_GATE_CLOSED로 닫혀도 부재 값이 그대로 간다(:402 → :409 순서와 맞음). **부재 값과 빈 묶음은 시험으로 가를 수 있다:** 빈 관측 묶음은 기존 경로로 만들 수 있다. 모든 경로에 레인이 없으면 `refused++; continue`(strategy_market_coordinator.go:66-69 부근)를 거쳐 NO_ACCEPTED_SCOPE 닫힘이 되고, 그 묶음은 관측 상태인 빈 묶음이다. 그러면 OFF 레인은 NO_INPUT, 부재 값은 UNOBSERVED로 갈린다. 인용 좌표 확인: 충돌 return :113-115, 루프 :64, `Arbitrate()` :123. 모두 맞다. |
| FLM 목록 | **CLOSED** | §4 FLM 목록과 무편집 사유 | 좌표가 모두 맞다: `NewPairedStrategyEntryProductionAssembly` :293-374(조립 리터럴 :366, `return` :373), `record` :322(생산 호출은 :249 한 자리), `projection` :31, `refreshPaired…` :575, evaluate의 생산 호출은 strategy_entry_supervisor.go:543-546 한 자리. `refreshPaired`는 바뀌지 않는 것이 구성상 맞다(묶음이 evaluate 경로로 오므로). `strategyLaneInputs`도 바뀌지 않는다(authority.entries만 읽음). 둘 다 본문 digest 대조로 확인하게 되어 있다. |
| P1-C (묶음 출처 · 파도 정의 · refresh 금지) | **PARTIAL** | §2 ③ 금지 집합, §5 묶음 출처(evaluate → record와 같은 잠금 안의 `{wave, batch}`), 클로저가 nil을 받은 뒤 동기 복사, 핀 ①④⑤ | 묶음과 파도는 구성상 짝이 되고, refresh·조립·파도 합류 함수와 캐시 필드는 금지됐다. 거기까지는 닫혔다. **남은 구멍:** §5는 클로저가 「그 시장 칸 `{wave, batch}`와 **그 시장 활성화**를 한 번 복사」한다고 쓴다. 그런데 칸에는 활성화가 없다. evaluate는 활성화(`promotion`)를 인자로 받지만(strategy_lane_runtime.go:199-201) record에는 넘기지 않는다. 클로저 쪽에는 `fresh`가 없다. 그래서 `ShadowEligible(activation, …)`에 넣을 활성화의 출처가 정해져 있지 않다. 자연스러운 우회는 shadow 단계에서 `loadFamilyActivation`이나 `familyGateFor`로 활성화 매니페스트를 다시 읽는 것인데, 둘 다 §2 ③ 금지 집합에 없다. 그러면 묶음과 다른 순간의 활성화를 쓰게 되고, 같은 파도라는 정의가 활성화 축에서 깨진다. 투영 쪽이 관측값으로 술어를 다시 확인하므로 안전 문제는 아니고 정합성 문제다. **닫는 법:** 칸을 `{wave, batch, activation}`으로 하고 ⑤의 같은 임계 구역에서 함께 쓴다(activation은 evaluate의 `promotion`). 그리고 §2 ③ 금지 집합에 `loadFamilyActivation`, `familyGateFor`, `strategyrouter.LoadProductionFamilyActivation`을 추가한다. |

**새 P1-D (v3.1 핀 ① 재진술이 연 구멍):** 주문 경로 앞에서 실행되는 shadow 메서드 하나의 본문이 아무 핀에도 묶여 있지 않다.
- 핀 ① (A)가 허용하는 형태는 `fresh.<shadow 필드>.forMarket(market)`이다. 이것은 메서드 **호출**이고, 실행 시점은 `runProductionStrategyMarketCycle` 안, evaluate 인자를 평가할 때, 즉 dispatch **앞**이다.
- 같은 핀이 「본문의 shadow 함수 호출 0」을 요구하므로 문면이 스스로 모순된다.
- 실질적으로 더 중요한 점: 그 메서드 본문을 묶는 핀이 없다.
  - ④⑤는 evaluate·record 안만 본다.
  - §2 ③의 금지 집합과 하한 단언은 「shadow 단계 호출 폐포」를 걷는데, `forMarket`은 shadow 단계가 아니라 주기 함수에서 불린다.
  - §2 ②는 이 메서드를 운반 허용 목록에 이름으로 넣는다.
- 그래서 `strategyShadowPair.forMarket` 본문에 매니페스트 적재나 판정 같은 일을 넣는 변이가 핀 ①③④⑤와 :169-175를 모두 통과한다. P1-B와 같은 부류인 「dispatch 앞의 shadow 일」이 허용된 한 자리로 다시 들어오는 셈이다.
- **닫는 법:** `forMarket` 본문을 AST로 「시장 비교 후 필드 반환만, 호출 0」으로 고정한다. 기존 `strategyProposalAuthorityPair.forMarket`(strategy_proposal_authority.go:167-175)이 같은 모양이다. 아니면 그 메서드도 §2 ③ 금지 폐포 걸음에 포함한다. 핀 ①의 「호출 0」 문구에는 이 한 자리 예외를 명시한다.

**공통 질문 확인 (새 P1 아님):**
- 핀 ① 재진술과 기존 :169-175 핀은 충돌하지 않는다. 기존 핀은 마지막 문장과 인자 4개만 본다(a112_market_delivery_structure_test.go:165-175). 같은 함수를 보는 다른 핀도 확인했다: `dispatchMentionCensus`는 `.dispatch` 셀렉터만 센다(strategy_dispatch_handoff_guard_test.go:1055-1058). 잠금 복구 세대 핀은 evaluate의 `Args[2]` 문자열만 본다(a112_lane_latch_durability_test.go:335-348). 따라서 shadow 인자를 **세 번째 이후**(인덱스 3 이상)에 붙이면 충돌이 없다. 그 앞에 끼우면 잠금 복구 핀이 정당하게 RED가 된다. 브리프가 인자 위치를 명시하면 좋다.
- evaluate의 오류·panic 갈래:
  - `recoverMarketLanes` 오류나 레인 panic 재던짐은 record보다 **먼저** 일어난다(:208, :244-248). 그러면 칸은 이전 파도 그대로이고, 주기가 오류를 반환하므로 shadow 단계는 시작하지 않는다.
  - `persistMarketLatches` 오류는 record **뒤**다(:252). 이때 칸과 파도는 N으로 올라가지만 shadow는 시작하지 않는다. §5.1 신선도 규칙에 따라 이전 관측은 버려지고 UNOBSERVED가 된다.
  - 두 경우 모두 안전하게 맞물린다.
- 좌표 정확성: v3.1 §0이 admit 자리로 인용한 「:87」은 실제로 `gate.admit(` **:91**이다. 이 줄은 v3에서 그대로 내려온 것이다. §0 머리말도 여전히 「좌표 = HEAD 2817064c」이지만 Go 코드는 같다. 핀 ⓐ는 줄 번호가 아니라 문장 이름으로 고정하므로 영향 없음(비등급 메모).

**저장소 무변경 확인:**
- 시작과 끝 모두 `rev-parse HEAD` = `4d22d72663dcf3a42be928b3ecf3db817ffc6ac5`.
- `status --short`도 시작과 끝이 같다: ` M docs/ROADMAP.md`, ` M …/tasks.md`, `?? .reticle-setup-crash.log`, `?? …/review-shadow-freeze-2026-10/`, `?? …/shadow-2026-10/`, `?? w4.log`.

