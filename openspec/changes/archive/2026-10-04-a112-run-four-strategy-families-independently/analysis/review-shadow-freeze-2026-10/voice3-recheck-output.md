# SHADOW re-freeze 재검 voice3 — SubagentHandback 원문 (agent ac4fbb2d05a76746e, 저장 2026-10-04)

**재검 결과: FAIL.** 1차 P1 여섯과 FLM 지적 가운데 CLOSED가 넷(P1-3·4·5·6), PARTIAL이 셋(P1-1·P1-2·FLM)이다. v2가 새로 연 P1은 둘이다. P0은 없다.

**안전 규칙:** `~/.codex` 아래는 읽지도 검색하지도 않았다. 리뷰 디렉터리 안의 `_codex/` 도 열지 않았다. 저장소 파일은 편집하지 않았다. 실험은 `…/scratchpad/v3recheck`(`git archive 2817064c` 사본)에서만 했고, 임시 시험 파일은 지웠다.

**기준:**
- 좌표는 2817064c다. 설계 브리프 v2의 sha256은 `7f1e840c…37bd` 로 일치했다.
- `git diff --stat c7219640 2817064c -- '*.go' docs/api` 결과 코드 변경은 0이다. 그래서 1차 때 낸 코드 증거는 이 좌표에서도 그대로 유효하다.

| 1차 ID | 판정 | v2 좌표 | 근거 / 재현 |
|---|---|---|---|
| P1-1 관문·중재 뒤 입력이라 반사실이 공허함 | **PARTIAL** | v2 §0 정정, §4(관문 앞 `Input` 을 `strategyShadowInputs` 에 담아 **성공 반환에서만** 운반, 13 닫힘 = 관측 없음) | 미선언 시장은 닫힌다. 패자 제안도 관문 앞에서 잡힌다. **선언 시장은 열려 있다.** OFF 레인만 제안한 종목이 하나라도 있으면 범위가 지워지고 시장이 `FAMILY_GATE_CLOSED` 로 닫힌다. 그러면 v2 §4 규칙대로 shadow 운반이 0이 된다. 사본 재현(`-tags tossos_testseams`, 픽스처 `collectUnderGate`, CONT만 ON): CONT+REV가 함께 제안한 경우 READY, REV만 제안한 경우 `ready=false FAMILY_GATE_CLOSED`, REV+BREAKOUT도 같음, 전부 OFF로 선언한 경우도 같음. 결국 선언 시장에서 SHADOW는 ON 레인과 겹친 종목에서만 관측되고, OFF 레인만 제안한 종목(shadow가 정작 보려는 대상)과 전부 OFF 선언 시장에서는 영원히 관측 0이다. 근거 코드: `strategy_market_coordinator.go:110-112`(erasedScopes), `strategy_proposal_authority.go:402-408`. **닫는 법:** 적재 뒤 닫힘(FAMILY_GATE_CLOSED·PRODUCTION_FAULT·ARBITRATION_REFUSED·QUEUE_OVERFLOW·NO_ACCEPTED_SCOPE)에서도 관문 앞 입력을 싣는다. 아니면 이 편향을 「시장 닫힘 = 관측 없음」으로 투영에 드러내고, 수용 판정과 함께 적는다. |
| P1-2 dispatch 핀의 척도·표본 | **PARTIAL** | §8(유/무 차등, 두 배치, 전제 `WOULD_EMIT ≥ 1`, 전체 주기), §10 | dispatch 척도는 닫혔다. 단 「부분 ON 선언」 배치의 전제는 위 P1-1 때문에 겹친 종목 픽스처로만 성립한다. 재시작 핀 셋(§10 ①②)은 여전히 두 가지를 요구하지 않는다. 하나는 「앞 프로세스가 전체 주기에서 WOULD_EMIT ≥ 1」이고, 다른 하나는 「재시작한 프로세스가 전체 주기를 1회 이상 돈 뒤 dispatch/원장을 측정」이다. 그래서 1차에 지적한 R2 원본 모양(`evaluate(…, nil)`, dispatch 경로 없음)의 빈 표본을 그대로 물려받을 수 있다. §10에 그 두 조건을 한 줄 넣으면 닫힌다. |
| P1-3 유계 기계 통합 미정 | **CLOSED** | §5 | shadow 단계는 collect 밖, 레인 유계 step 밖, `ShadowVerdict` 는 봉투도 오류도 돌려주지 않고, 스스로 마감·recover한다. latch·원장·주기 오류는 0이다. fault 주입 핀도 있다. (위치 문제는 아래 새 P1-B) |
| P1-4 교차 규칙이 정당한 입력을 거절 | **CLOSED** | §6(`ShadowEligible` = Desired OFF ∧ Effective OFF), §7(투영은 술어가 참일 때만 SHADOW를 냄), 거부 입력 표, 결정 63-v2 문단 | 투영이 술어가 참일 때만 SHADOW를 만들기 때문에, 구성상 desired ON/effective OFF 레인은 SHADOW가 되지 않는다. 그래서 503이 나는 경로가 없다. 남는 점: `strategyprojection` 은 import가 0인 잎 패키지라 `strategyworker.ShadowEligible` 을 부를 수 없다. 투영 쪽은 「값 형태」의 사본이다. 두 쪽을 같은 행렬로 시험하는 계획(§6 표)이 있으므로 CLOSED로 본다. |
| P1-5 OpenAPI enum 핀 부재 | **CLOSED** | §9(`runtime.enum == LaneRuntimes()`, `shadowOutcome.enum == ShadowOutcomes()`) | 시험 신설이 계획됐다. trigger/start/outcome enum은 범위 밖으로 남았고, 1차에서도 부가 제안이었다. |
| P1-6 runtime 출처 / router 어휘 | **CLOSED** | §7(SHADOW는 projection 어휘와 관측에만 있음, router {UNOBSERVED} 유지, worker 필드 불변, census는 go/types로 const와 변환식 전수), §6(health≠nil ∧ cycleGeneration>0) | 1차의 변환식 const 우회까지 포함해 닫혔다. |
| §7 FLM 목록(1차 P2-6) | **PARTIAL** | §4 FLM 확장 | `coordinateMarketProposals`·`strategyLaneInputs`·`runLane`·`runProductionStrategyMarketCycle`·`validateLane`·`evaluate`·`strategyLaneProjection` 은 들어갔다. 그러나 `collectMarket` 은 「(분기 수가 바뀌면)」 조건부로 남았다. v2 §4의 운반을 하려면 성공 반환 리터럴(`strategy_proposal_authority.go:453`)에 필드를 더해야 한다. High-risk 기존 함수의 내부 편집이므로 CLAUDE.md 4항상 면제가 안 된다. 무조건으로 바꿔야 한다. |

### v2가 새로 연 P1
- **새 P1-A. 관문 앞 운반 타입을 지키는 census의 걸음이 세탁 접근자를 못 본다.**
  - `strategyShadowInputs` 는 접근자가 `[]strategyworker.Input` **사본**을 돌려준다. 그 값의 필드 `{Scope, SnapshotDigest, Proposal}` 은 `strategycoordinator.Envelope` 의 세 필드와 정확히 같다(`coordinator.go:104-108`). 즉 OFF 레인 제안으로 봉투를 만들 재료가 된다.
  - §2 ②/§4의 「보는」 걸음은 매개변수·결과·필드·지역 선언 타입을 센다. 예를 들어 `strategyProposalMarketAuthority` 에 새 메서드 `shadowProposals() []strategyworker.Input` 를 두고 `dispatchHandoffs` 가 그것을 부르는 변이를 생각해 보자. `dispatchHandoffs` 의 선언 타입 어디에도 `strategyShadowInputs` 가 나오지 않으므로 이 걸음으로는 0이 센다.
  - 그 뒤의 방어는 봉인이 아니다. `strategyhandoff.AdmitEachOwnerScope` 는 소스 동결로 지켜지지만 엔진 어디서든 부를 수 있다. `dispatchHandoffs` 의 digest 대조는 코드 주석(`strategy_dispatch_handoff.go:46-52`)이 스스로 「봉인이 아님, 위조 가능」이라고 적었다.
  - **닫는 법:** 접근자 메서드의 `*types.Func` 사용처(`types.Info.Uses`)를 허용 목록으로 센다. 함수 본문의 모든 식 타입(`types.Info.Types`)에 `strategyShadowInputs` 가 들어 있는지도 걸음에 넣는다. 양성 대조로 위 변이를 시험한다.
- **새 P1-B. shadow 단계가 주문 경로 앞에 놓이는데, 차등 핀은 그것을 못 본다.**
  - 기존 AST 핀이 `runProductionStrategyMarketCycle` 의 **마지막 문장을 `return dispatchStrategyMarketHandoffs(…)` 로 고정**한다(`a112_market_delivery_structure_test.go:169-175`).
  - 따라서 v2 §5의 「evaluate 뒤 별도 단계」를 이 함수 안에 두면 dispatch **앞**이 된다. 핀이 있을 때 매 주기 주문 경로가 shadow 적재(파일 I/O)와 판정, 그리고 「자체 마감」(값 미정)만큼 늦어진다. 제안의 신선도 창을 넘으면 핀 유무에 따라 dispatch 결과가 달라진다.
  - §8 차등 핀은 가짜 시계에서 궤적을 비교하므로 이 지연을 못 본다.
  - **닫는 법:** 배치를 명시한다. dispatch 뒤에 돌리거나(`defer` 또는 주기 밖), 앞에 둔다면 마감 상한을 상수로 두고 「dispatch 입력은 shadow 단계 전에 확정」을 AST로 고정한다.

**1차 P2 중 v2가 악화시킨 것:** 없다.

**시도했지만 막힌 공격:**
- strategyshadow를 별도 패키지로 둔 것 → `FamilyActivation` 위조 경로가 닫혔다. 필드가 비공개라서이고, §2 ①의 리터럴 1자리 census도 더해졌다.
- desired ON/effective OFF 레인에 SHADOW를 다는 공격 → 술어 하나로 막혔다.
- worker 필드에 SHADOW를 굽는 공격 → §7이 막는다.

**저장소 무변경:**
- 시작과 끝 모두 `rev-parse HEAD = 2817064cc1f78c4d0d6abc04211de906a7c3efc5`.
- `status --short` 는 시작과 끝이 같다: ` M docs/ROADMAP.md`, ` M …/tasks.md`, `?? .reticle-setup-crash.log`, `?? …/review-shadow-freeze-2026-10/`, `?? …/shadow-2026-10/`, `?? w4.log`.
- 사본에 넣은 임시 시험은 삭제했다.

