# SHADOW freeze voice3 — SubagentHandback 원문 (agent ac4fbb2d05a76746e, 저장 2026-10-04)

## 보이스 ③ (증거 · 사실 · 실현 가능성) — a112 7.3.1 SHADOW freeze 리뷰

**안전 규칙:** `~/.codex` 아래는 읽지도 검색하지도 않았다. 사본에 들어 있는 저장소 자체의 `.codex/` 디렉터리도 grep 대상에서 뺐다. 저장소 파일은 편집하지 않았다. 실험은 모두 스크래치 사본(`…/scratchpad/v3copy`, c7219640 archive)에서 했고, 실험이 끝난 뒤 사본도 되돌렸다.

**판정: REJECT — 설계 §0·§3을 다시 쓴 뒤 재리뷰 필요.**
- P0은 0건이다. 실주문이나 노출 상승 경로는 찾지 못했다.
- 그러나 설계의 핵심 전제(§0 「이미 있는 제안 위의 레인 관문 반사실」)가 c7219640 코드와 맞지 않는다. 그대로 지으면 SHADOW는 반사실을 하나도 만들지 못한다. 그리고 dispatch MUST NOT을 지키는 행동 핀은 빈 표본이거나 잘못된 척도라서 통과만 하게 된다.

### P1

**P1-1. 레인 런타임의 입력은 「이미 있는 제안」이 아니다. 관문과 중재를 통과하고 READY인 경우에만 남은 항목이다. 그래서 SHADOW 반사실이 구조적으로 공허하다.**
- 근거 (코드):
  - 레인 입력은 `strategy_entry_supervisor.go:543-546`의 `strategyLaneInputs(fresh.proposals…)`가 만든다.
  - `strategyLaneInputs`(`strategy_lane_runtime.go:266`)는 `authority.entries`만 읽는다.
  - `entries`는 성공 갈래(`strategy_proposal_authority.go:453`)에서만 채워진다. 그 값은 중재 선택(`arbitration.entries()`), 즉 소유자 범위당 하나다.
  - 관문이 막은 제안은 Submit 전에 `continue`로 빠진다(`strategy_market_coordinator.go:96-100`).
  - 범위가 통째로 지워지면 `FAMILY_GATE_CLOSED`가 되고 entries는 nil이다(`strategy_proposal_authority.go:402-408`). 다른 닫힘 12갈래도 모두 entries가 nil이다.
- 재현: 사본에 임시 시험을 넣고 `go test -tags tossos_testseams` 로 돌렸다. 기존 픽스처 `collectUnderLoad/collectUnderGate` 를 썼고 시험 파일은 삭제했다.
  - 미선언, CONT+REV가 005930에 제안: 입력 = `{CONTINUATION:1 (000660), REVERSAL:1}`. 005930의 CONT 패자 제안은 없다.
  - 선언, CONT만 ON: 입력 = `{CONTINUATION:2}`. SHADOW 대상인 REV(OFF) 레인의 입력은 0이고 `gated=[DORMANT]` 이다.
  - 선언, CONT만 ON, 005930은 REV만 제안: `ready=false reason=FAMILY_GATE_CLOSED`, 입력 `{}`. 000660의 CONT 레인 입력까지 0이 된다.
- 결과:
  - 선언 시장에서는 shadow가 적용되는 OFF 레인이 언제나 `NO_INPUT` 이다.
  - 미선언 시장에서는 이미 레거시 경로로 거래되는 승자 레인만 `WOULD_EMIT` 이다. 반사실이 아니라 사실이고, coordinators `selected[]` 의 사본일 뿐이다.
  - 골든 `arbitration.selection = "… after pure evaluation"` 에 따르면 계약 어휘의 「pure evaluation」은 중재 **앞** 단계다. 설계의 해석(중재 뒤 관문)과 다르다.
  - 브리프 질문 2의 「반사실 결과 = 레인 관문 반사실」은 공허하게만 충족된다(전칭 판정·빈 표본 교훈).
- 제안:
  - shadow 입력은 관문 **앞**의 배치(`batch.LanesFor` 단위)에서 따야 한다. 예: `coordinateMarketProposals` 가 관문 전에 입력을 권한에 실어 내보낸다(읽기 전용 필드).
  - 그러면 High-risk 함수 `coordinateMarketProposals` 가 편집 대상이 되므로 §7 FLM 목록에 넣어야 한다.
  - 시장 닫힘 갈래(FX·authority·fault·gate·arbitration)에서 shadow를 「관측 없음」으로 둘지도 정해야 한다.

**P1-2. §3 ③ dispatch 행동 핀의 척도가 틀렸고, 표본은 비어 있다.**
- 「shadow 매니페스트로 WOULD_EMIT이 여럿이어도 Submit 0」이라는 척도 자체가 맞지 않는다.
  - 미선언 시장에서는 레거시 경로가 모든 제안을 Submit한다(`strategy_market_coordinator.go:94-108`, `admit` 은 `!installed` 일 때 통과). 그래서 Submit은 shadow와 무관하게 0이 아니다.
  - 선언 시장에서는 OFF 레인의 WOULD_EMIT이 P1-1 때문에 나올 수 없다. 핀은 빈 표본 위에서 통과한다.
- 재시작 짝(§4)은 R2 원본 `TestARestartAfter…` 의 모양을 그대로 물려받는다. 원본은 `lanes.evaluate(..., nil)` 로 입력 0이고 dispatch를 부르지 않는다. 따라서 「lease 0」은 dispatch 경로가 없는 측정이다.
- 제안:
  - 척도를 **차등**으로 바꾼다. shadow 매니페스트 유무에 따른 Submit 집합·`dispatchHandoffs`·게이트웨이 스파이 호출이 동일해야 한다.
  - 미선언 시장과 부분 ON 선언 시장 두 배치에서 모두 잰다.
  - 전제조건으로 WOULD_EMIT ≥ 1을 단언한다.
  - `runProductionStrategyMarketCycle` 전체(refresh→evaluate→dispatch)를 생산 적재기 seam nil 상태로 돌린다.

**P1-3. Shadow 문을 레인의 유계 기계에 어떻게 붙일지 정하지 않았다. 자연스러운 구현은 §3 ①의 자기 census와 모순된다.**
- `Step` 은 `func(ctx, Input) (Cycle, error)` 이다(`bounded.go:67`). `Cycle` 은 `Envelope` 필드를 갖는다(`worker.go:173-178`).
- `RunBounded` 안에서 Shadow를 돌리면 결과가 `Cycle` 로 운반된다. 그러면 「ShadowCycle에는 봉투 필드 없음」이 무의미해진다.
  - 이 경우 deadline·panic이 abnormal로 처리되어 레인이 잠기고, `persistMarketLatches` 가 원장에 durable latch를 남긴다(`bounded.go:200-211`, `strategy_lane_runtime.go:252`). shadow가 유발한 원장 쓰기가 재시작을 넘는다.
- 밖에서 돌리면 마감 시한이 없다.
- 어느 쪽인지, 그리고 「shadow 고장은 잠그지 않는다 / 잠근다」를 설계에 적어야 한다.

**P1-4. 투영 교차 규칙(§9)이 정당한 입력을 거절해 투영 전체가 사라질 수 있다.**
- 활성화는 desired ON / effective OFF를 정당한 상태로 받는다(`production_family_activation.go:620-623`).
- 레인 관문은 `Effective` 만 본다(`worker.go:194`).
- 설계의 「SHADOW는 OFF 레인만」을 `Effective != ON` 으로 구현하면, desired ON/effective OFF 레인이 runtime=SHADOW를 받는다.
- 그런데 새 규칙 「SHADOW ⇒ desired=OFF ∧ effective=OFF」는 그 스냅숏을 거절한다. RPC 서버 `transport_unix.go:141-144` 가 503 `runtime_unavailable` 을 내므로 모든 화면의 투영이 사라진다.
- 제안: shadow 적용 술어와 교차 규칙을 **한 함수**로 만든다. 거절할 정상 입력도 열거한다.

**P1-5. §5 「OpenAPI 대조 시험 갱신」은 존재하지 않는 핀이다.**
- `a112_lane_children_openapi_test.go` 는 필드 이름·required·strict만 대조하고 enum 값은 보지 않는다. 같은 디렉터리의 다른 OpenAPI 시험도 레인 enum을 보지 않는다.
- 재현(사본): `LaneRuntimeShadow = "SHADOW"` 를 추가하고 `validateLane` 을 넓혔다. OpenAPI enum은 `["UNOBSERVED"]` 그대로 두었다. 결과는 `go test ./internal/httpapi -run OpenAPI` ok, `./internal/strategyprojection` ok.
- 제안: 새 시험을 만든다. OpenAPI의 `runtime.enum`(그리고 trigger/start/outcome enum) == `strategyprojection` 어휘 함수. `shadowOutcome` enum도 포함한다.

**P1-6. runtime=SHADOW의 출처가 정해지지 않았다. 지금 투영은 프로세스 수명 동안 고정된 worker 필드를 읽는다.**
- `strategy_lane_projection.go:55` 는 `Runtime: lane.Runtime()` 을 관측 여부 판정보다 먼저 쓴다. `lane.Runtime()` 은 `worker.runtime` 이고, 생성 시 고정된다(`worker.go:162`).
- Manager ④는 「router `RuntimeState` census = {UNOBSERVED, SHADOW}」라고 했다. 그런데 `strategyrouter.RuntimeState` 는 시장 단위 봉인 레코드 `MarketRecord` 도 쓰는 타입이다. 이 레코드는 ON이 될 수 있고(`scheduler.go:10-26, :94, :105`), `newMarketRecord` 는 어휘 검사 없이 받는다(:52-55).
  - 이 타입에 SHADOW 값을 열면 worker 필드에 SHADOW를 구우는 구현이 생기기 쉽다. 그러면 매니페스트를 철회해도 프로세스가 사는 동안 SHADOW가 남는다.
- 제안:
  - SHADOW는 projection `LaneRuntime` 과 관측(observation)에만 둔다.
  - router `RuntimeState` census는 {UNOBSERVED}로 유지한다.
  - 투영은 runtime을 관측에서 읽는다.
  - 교차 규칙에 「SHADOW ⇒ health≠nil ∧ cycleGeneration>0」을 더한다.

### P2

| # | 지적 | 근거 / 재현 | 제안 |
|---|---|---|---|
| 1 | R2 어휘 census를 피해 갈 수 있다 | 사본에서 `const LaneRuntimeShadow = LaneRuntime("SHADOW")`(타입이 변환식에 있음)를 넣고 `validateLane` 을 넓혔다 → census 시험 ok. census는 `ValueSpec.Type` 이 Ident인 것만 센다(`a112_shadow_absent_restart_test.go:94`). 대조군: 타입을 명시한 const는 FAIL(정상) | go/types로 「타입이 LaneRuntime인 상수 전부 + `LaneRuntime(...)` 변환 지점」을 센다. `strategyflow.RuntimeState`·`strategyruntime.RuntimeState`(같은 UNOBSERVED 어휘, 범위 밖)도 포함하거나, 제외 사유를 적는다 |
| 2 | §3 ① census의 대상이 틀렸다 | `FamilyActivation` 은 비공개 필드만으로 지켜진다. 같은 패키지(strategyrouter)에 들어올 shadow 적재기는 `FamilyActivation{generation:…}` 를 직접 만들 수 있다(`production_family_activation.go:493`). 「FamilyShadow에 DesiredState 반환 메서드 0」은 `Promote() FamilyActivation`, 패키지 함수, bool 반환으로 뚫린다. 「ShadowCycle에 Envelope 없음」도 마찬가지다. Envelope는 공개 필드 구조체라 Input만 있으면 누구나 만든다(`coordinator.go:104-108`) | census를 바꾼다. (a) 비시험 strategyrouter 안의 비영 `FamilyActivation{` 리터럴 = 정확히 1자리. (b) FamilyShadow는 엔진에서 허용 함수(레인 런타임·투영)에서만 참조되고, `admit`·`coordinateMarketProposals`·`dispatchHandoffs` 에서는 참조 0(go/types uses). `Lane.Desired/Effective`(`lane_view.go:47-51`)도 census에 넣는다 |
| 3 | 호출 허용 목록 모양의 한계 | `TestTheRollbackPathOnlyReads` 는 이름 붙인 함수 본문만 센다(`a112_family_rollback_test.go:275-310`). 헬퍼를 하나 빼내면 범위 밖이 되고, `lane.Run` 같은 텍스트가 수신자와 무관하게 허용된다 | shadow 진입점에서 전이 호출 폐포를 세고, 셈의 하한을 단언한다 |
| 4 | 「validateLane이 모든 읽기 경로의 관문」은 부분적으로만 맞다 | lanes는 `Context.Read` 에서 store 검증 **뒤에** 덧씌워진다(`strategy_runtime_projection.go:53-58`). 그래서 store.go:18/:43은 lanes를 검증하지 않는다. 실제 관문은 RPC 서버 `transport_unix.go:141-144` 이고, 그 뒤 클라이언트(`transport.go:97`)·httpapi·console이 이중으로 본다. 엔진 시험 헬퍼 `a112Read` 는 Validate를 부른다(`a112_lane_coordinator_projection_test.go:56`) | §9 서술을 정정한다. 교차 규칙 핀은 `Validate`/RPC 서버 경로로 잰다 |
| 5 | 로트가 놓친 파일·시험 | `LaneJSONFields`(OpenAPI 시험이 properties = required = 목록을 요구하므로 `shadowOutcome` 은 핀이 0이어도 늘 직렬화됨 → 「핀 0 = 동작 변화 0」이 wire 모양에서는 거짓). `cloneLanes` 의 포인터 복제. `validateLane` 미관측·무트리거 갈래. `a112_lane_coordinator_children_test.go:138` 거절표. `strategyworker/golden_contract_test.go:108`(worker runtime = 골든). evaluate 서명을 바꾸면 호출자 약 30곳 + AST 핀(`a112_lane_latch_durability_test.go:335-348`) | §5·§7에 목록화한다 |
| 6 | §7 FLM 목록이 불완전하다 | P1-1 수리에는 `coordinateMarketProposals` 와 `strategyLaneInputs`, P1-3에는 `runLane`, 매개변수 전달에는 `runProductionStrategyMarketCycle`(High-risk, 구조 핀 있음), `validateLane` 이 필요하다 | FLM 범위를 넓힌다 |
| 7 | 공유 헬퍼의 오류 신원 | 공유하자는 「수명」 판정 `familyActivationRemaining` 은 `ErrProductionFamilyActivationExpired` 를 돌려준다(:384-390). 미선언 판정도 활성화 sentinel이다. shadow 오류가 활성화 신원을 갖게 된다(8.5 P2-1과 같은 모양). 공유 사실을 못 박는 핀도 계획에 없다 | 중립 오류를 반환하고 각 적재기가 자기 sentinel로 감싼다. 「shadow 적재기는 공유 헬퍼를 부르고 자기 파일 읽기·디코드 사본이 0」을 AST로 고정한다. 도메인 교차 디코드 거절 시험(shadow 바이트 → 활성화 디코더, 그 반대)을 둔다 |
| 8 | shadow 적재 자리 | collectMarket 안에서 적재하면 panic이 `collect` 의 recover(`strategy_proposal_authority.go:271-275`)로 가서 INTERNAL_FAILURE가 된다. shadow 결함이 레거시 진입을 닫는다(토글 OFF ≠ upstream) | 적재를 collect 밖(evaluate 쪽)에 두고, 그 사실을 시험한다 |
| 9 | 「활성화 쓰기 0」 핀 | 「파일 바이트 전후 동일」은 원래 없던 파일이 새로 생기는 경우를 못 본다 | 설정 디렉터리 목록까지 전후로 같아야 한다고 단언한다 |
| 10 | 개정 뒤에도 낡은 문장이 남았다 | design:242 「spec.md에서 "signed"는 SHADOW manifest에만 붙는다 … 그대로 둔다」(이제 spec에 signed가 없음 — 브리프 §1이 「그 줄도 함께」라고 했지만 패치는 :246만 고침). tasks.md:566(c7219640 기준) 「서명 shadow 매니페스트」, ROADMAP.md:250, R2 시험 실패 문구 「signed shadow manifest, golden amendment」(:120, Manager ④ 골든 무변과 모순) | 문구를 정리한다 |
| 11 | 시나리오 :91이 좁아졌다 | 「signed manifest 없이」→「핀 없이」로 바뀌어, 「핀은 있는데 파일 없음·만료·폐기」 재시작이 spec 시나리오에서 빠졌다. 설계 §4는 다룬다 | 「유효한 shadow manifest 없이(핀 없음·파일 없음·불일치·만료·폐기)」로 고친다 |

**범위 밖 (등급 없음):** spec :38 「SHADOW counterfactual 외 dispatch handoff 0」과 design :84 「SHADOW projection은 raw score를 counterfactual로」는 조정자 수준 반사실을 전제한다. Manager ③의 좁힘 뒤에는 가리키는 대상이 없는 문장이 된다.

### §0·§9 사실 다섯 실측 (c7219640)
1. 「평가는 활성화와 무관하게 돈다」
   - 제안 배치 적재에 대해서는 참이다(`collectMarket` :355, 관문과 무관).
   - evaluate는 배선이 준비되면 dormant 갱신 worker에서도 돈다(:427-433, :543).
   - 그러나 레인이 보는 입력은 관문과 중재 뒤의 것이라, 「레인 관문 반사실」의 전제로는 **거짓**이다 → P1-1.
2. 「골든은 runtime 어휘를 열거하지 않는다」 — **참**. 골든 :17-24는 서술자 기본값만 담고 enum은 없다.
3. 「OpenAPI가 열거한다」 — **참**. `StrategyRuntimeLaneRuntime.runtime.enum=["UNOBSERVED"]`, `additionalProperties:false`. 다만 그 enum을 고정하는 시험은 없다 → P1-5.
4. 「validateLane이 모든 읽기 경로의 관문」 — **대체로 참이고 좌표 하나가 틀렸다** → P2-4.
   - grep 범위: `internal`·`cmd` 의 비시험 `.go` 중 strategyprojection을 import하는 14개 파일 전수. 저장소 자체의 `.codex/` 는 제외.
   - Validate를 거치지 않는 외부 경로는 없다. 집계 스냅샷(`httpapi_reader.go:576-583`)도 RPC 클라이언트의 Validate를 거친다.
   - lanes의 runtime 값을 직접 읽는 소비자: 0. 콘솔의 `UNOBSERVED` 는 `templates_optimization.go:436` 의 고정 문구 하나다.
5. 「RPC 디코더는 모르는 필드를 허용」 — **참**(`transport.go:77-90`, DisallowUnknownFields 없음).

### 시도했지만 막힌 공격
- **shadow 세대를 레인 잠금 복구 세대로 넘기기:** 막혔다. `a112_lane_latch_durability_test.go:335-348` 이 evaluate의 3번째 인자 식을 정확한 문자열로 고정한다. 앞쪽에 매개변수를 끼워 넣어도 실패한다.
- **엔진이 SHADOW+ON 레인을 만들어 시험을 통과하기:** 막혔다. `a112Read` 가 Validate를 부른다.
- **구판 콘솔이 가산 필드 때문에 죽기:** 막혔다. 디코더가 모르는 필드를 허용한다. 다만 enum 값 SHADOW는 구판 검증기가 거절한다. 설계 §9가 이미 인정하고 수용했다.
- **활성화 핀 없는 시장에서 shadow가 관문을 세우기:** 설계대로라면(shadow가 `installed()` 에 관여하지 않음) 막힌다. 다만 고정하는 핀은 P2-2의 (b)가 필요하다.

### 저장소 무변경 측정
- 시작: `rev-parse HEAD = c7219640d08408a650941c62cf0df6efff77c036`. status: ` M …/tasks.md`, `?? .reticle-setup-crash.log`, `?? …/review-shadow-freeze-2026-10/`, `?? …/shadow-2026-10/`, `?? w4.log`
- 끝: status는 **동일**하다. **HEAD는 `1dfda9e1` 로 바뀌었다.** 바꾼 것은 내가 아니다.
  - `1dfda9e1 2026-10-04 23:19:33 daniel docs(a100): R0 — stale 번들 7→0 …` 은 병행 세션의 a100 커밋이다.
  - 내가 실행한 git 명령은 status/rev-parse/archive/show/log/diff(읽기 전용)뿐이다.
  - 리뷰 근거는 모두 c7219640 archive 사본에서 읽었으므로 HEAD 이동의 영향을 받지 않는다.
- 사본 정리: 임시 probe 시험은 삭제했고, `lanes.go` 는 c7219640 원본과 같음을 diff로 확인했다.

