# 7.3.1 SHADOW 설계 브리프 v2 (코드 0 — 2026-10-04)

v1(`design-brief-v1.md`)은 freeze 4판(codex REJECT · 보이스 1 REJECT · 보이스 2 APPROVE-WITH-CHANGES · 보이스 3 — `analysis/review-shadow-freeze-2026-10/`)에서
설계 수정 → re-freeze 판정을 받았다. v2 는 Manager 합본 판정 12 항을 반영한다. 계약: spec four-family-strategy-runtime :38(「dispatch handoff 0 — SHADOW 반사실 포함」) ·
:88-94(개정 [결정 63-v2] — :89 digest 핀 SHALL · MUST NOT 둘, :91-94 시나리오 「유효한 shadow manifest 없이 — 핀 없음, 또는 핀은 있으나 파일 없음 · 핀 불일치 · 만료 · 폐기 ·
결속 불일치」), design 결정 61 · 62 · 63(-v2) · :293(「dormant/shadow 로 배선」) · :295(「shadow 관측 수집」은 배포 뒤). amendment: c7219640 · 2817064c.

## 0. 코드 사실(v1 정정 포함, 좌표 = HEAD 2817064c)

- **v1 §0 의 「이미 있는 제안 위의 레인 관문 반사실」 은 틀렸다**(codex · 보이스 1 P1-1 · 보이스 2 P1-1 · 보이스 3 P1-1). 레인 런타임 입력은 `strategyLaneInputs(accountRef,
  authority)`(strategy_lane_runtime.go:266)가 `authority.entries` — 조정자가 **고른 승자** 만 — 에서 만든다(strategy_market_coordinator.go entries · strategy_proposal_authority.go
  성공 반환). 선언된 시장에서 OFF 레인의 제안은 관문(`gate.admit`)에서 DORMANT 로 멈춰 레인 런타임에 닿지 않고, 닫힌 시장은 entries 0. 그대로 쓰면 SHADOW 반사실은 구조적으로
  NO_INPUT 뿐(빈 표본).
- 관문 **앞** 의 전체 제안은 `coordinateMarketProposals`(strategy_market_coordinator.go:57) 루프의 `batch.LanesFor(symbol)` → 레인별 `strategyworker.Input{Scope, SnapshotDigest,
  Proposal}`(admit 직전, :87)에 있다.
- 레인 관문 `FamilyWorker.Run`(worker.go:193) · `Desired/Effective`(:107-113) · `Lane.Run`(lane.go:211) · 레인 런타임 `evaluate`(:199) · `runLane`(:286, 유계 step · 마감 ·
  latch 정산) · 시장 주기 `runProductionStrategyMarketCycle`(strategy_entry_supervisor.go:514).
- 투영: `strategyLaneProjection`(strategy_lane_projection.go) — `Runtime` 은 **worker 생성 시 고정 필드**(`lane.Runtime()` → UNOBSERVED)에서 읽는다(보이스 3 P1-6). 검증
  `strategyprojection.validateLane`(lanes.go:274) 이 runtime == UNOBSERVED 정확 일치. `Validate` 는 저장소 · RPC 클라이언트 · httpapi · console 의 관문이지만
  **`Context.Read`(strategy_runtime_projection.go:49-60)는 동적 레인 투영을 덧씌운 뒤 Validate 를 다시 부르지 않는다**(codex P1) — v1 §9 의 「모든 읽기 경로」 는 「외부 경계
  (RPC · httpapi · console)」 로 좁힌다.
- router `RuntimeState`(types.go:40-42)는 `MarketRecord` 와 공유 타입 — SHADOW 를 여기 열면 시장 기록 어휘가 넓어진다(보이스 3 P1-6). OpenAPI `StrategyRuntimeLaneRuntime`:
  `runtime.enum=["UNOBSERVED"]`, `additionalProperties:false`, properties 전부 required — 그 enum 을 고정하는 시험은 **없다**(보이스 3 P1-5).
- 정규 바이트 등식은 **누가 썼는지** 를 증명하지 않는다(codex P1) — 신뢰 앵커는 「배포 핀을 쓸 권한」 이고, 생성기만 만든다는 것은 운영 정책(결정 61 이 이미 그렇게 적음).

## 1. 신뢰 앵커 · 형식 (판정 ① · F1 — 변경 없음, 문구만 정정)

digest 핀 `TOSSOS_STRATEGY_FAMILY_SHADOW_<MARKET>_MANIFEST_SHA256` + 정규 바이트 등식 + 0400 · 소유자 · 크기 상한 · 24h 수명 · 폐기 · 결속(활성화 다섯 — 경로 매니페스트 · 보정 ·
달력 · 빌드 · 위험 정책; ProtectionReady 제외). 별도 파일 `strategy-family-shadow-<MARKET>.json`, 서술자 = 그 시장 네 레인 각각 `shadow: ON|OFF` — desired/effective 필드 없음.
생성기 `tools/a112-family-shadow` · 커밋 골든(바이트 + 핀) · 골든을 **만드는** 함수를 부르는 시험 · `docs/operations.md` 절 · 생산 작성자 0 가드. 앵커의 의미는 「배포 핀을 쓸 수 있는
주체」(정규 등식은 바이트 동일성만) — 문서 · 결정 63 문구와 일치.

## 2. 패키지 · 경계 (판정 ② — (a) 별도 패키지)

- 새 패키지 **`internal/strategyshadow`**: `FamilyShadow`(불투명 — 필드 비공개, 영값 = 아무 레인도 shadow 아님) · `LoadProductionFamilyShadow` · `EncodeProductionFamilyShadow` ·
  shadow 서술자 표 대조. `strategyrouter` 를 import 하되 **`FamilyActivation` 을 만들 수 없다**(그 타입의 필드는 strategyrouter 비공개 — 다른 패키지는 영값 말고 만들 길이 없음).
  보이스 1 P1-2 · codex P1(같은 router 패키지 안 쌍둥이 구조체 변환으로 Verified=true)의 경로가 구조적으로 닫힌다.
- census(go/types 해소, ShadowBand 선례 `internal/candidate/band_test.go:327` `TestNoFunctionThatProducesAVerdictCanSeeAShadowBand` 모양):
  ① `strategyrouter.FamilyActivation` 의 **비영 composite literal 은 저장소 전체에서 정확히 1 자리**(LoadProductionFamilyActivation 의 반환) — 시험 seam 파일 제외 목록은 이름으로;
  ② `strategyshadow.FamilyShadow` 를 보는(매개변수 · 결과 · 필드 · 지역 — 별칭 · any · 제네릭 인스턴스 · 함수 값 포함 「품는다」 걸음) 함수는 **허용 목록**(레인 런타임 shadow 단계 ·
  투영)뿐이고 `strategyFamilyGate.admit` · `coordinateMarketProposals` · `dispatchHandoffs` · `dispatch*` · `strategyworker.Lane.Run/Desired/Effective` · `FamilyWorker.Run/Desired/Effective`
  에서 0; ③ shadow 단계 호출 폐포(전이)에 조정자 Submit · handoff · 게이트웨이 · 원장 쓰기 · 활성화 인코더 0 — 하한 단언(폐포가 실제로 깊었다 — 양성 대조) 동반;
  ④ 경계 패키지 소스 digest 동결(strategyhandoff `source_freeze_test.go` 선례) — census 가 못 보는 모양(제네릭 주조 등)의 종결.

## 3. 적재기 — 활성화 적재기 무편집, 사본 + 양쪽 AST 핀 (판정 ③)

- `strategyrouter.LoadProductionFamilyActivation` 과 그 도우미는 **편집하지 않는다**. shadow 적재기는 형태를 옮겨 적은 사본이고, 옮겨 적은 코드 교훈대로 **양쪽을 AST 로 못
  박는다**(같은 검사 순서 — 미선언 맨 앞 · ctx · 설정 결속 · 파일 읽기 · 핀 · 정규 · 폐기 · 결속 · 수명 · 서술자; 한쪽만 바뀌면 실패).
- 공유 헬퍼는 **중립 오류를 돌려줄 때만** 공유(예: `readProductionRouteFile` 은 이미 공유 — 각 적재기가 자기 sentinel 로 감쌈). sentinel 은 패키지마다 따로
  (`ErrProductionFamilyShadowUndeclared/Unavailable/Revoked/Expired`) + **errors.Is 배타 핀**(활성화 오류가 shadow sentinel 을, shadow 오류가 활성화 sentinel 을 만족하지 않음).
- 교차 디코드 거절 **양방향** 핀: shadow 바이트를 활성화 적재기에 · 활성화 바이트를 shadow 적재기에 → 둘 다 거절(스키마 · 도메인 · 정규 등식 · DisallowUnknownFields).

## 4. 입력 — 관문 앞 배치에서 분기 (판정 1)

- `coordinateMarketProposals` 루프에서 각 레인 제안의 `strategyworker.Input`(admit **앞**)을 모아 `strategyMarketArbitration` 의 **읽기 전용 필드**(새 타입 `strategyShadowInputs` —
  비공개 슬라이스 + 복사 반환 접근자)로 싣고, 성공 반환한 `strategyProposalMarketAuthority` 가 그대로 운반한다. **시장 닫힘 갈래(13 닫힘 전부)의 shadow = 「관측 없음」**(운반 0 —
  fail 클로저는 그 필드를 싣지 않음).
- 그 타입은 `dispatchHandoffs` · `ResultAuthority` · `entries` · 조정자 · admit 이 **닿을 수 없다** — census(§2 ② 의 같은 걸음으로 `strategyShadowInputs` 를 보는 함수 = 허용 목록뿐) +
  `dispatchHandoffs` 가 `entries` 만 읽는다는 구조 핀(보이스 1 P1-1 제안).
- FLM 확장(판정 1): `coordinateMarketProposals` · `strategyLaneInputs` · `strategyLaneRuntime.runLane` · `Context.runProductionStrategyMarketCycle` · `strategyprojection.validateLane` ·
  `strategyLaneRuntime.evaluate` · (분기 수가 바뀌면) `collectMarket` · `strategyLaneProjection`.

## 5. 실행 자리 · 고장 격리 (판정 4 · 보이스 1 P1-4)

- shadow 단계는 **`collect` 밖**(제안 수집 · 권한 파도와 무관), **레인의 유계 step 밖**(`runLane` 의 마감 · latch 정산을 쓰지 않음), 시장 주기에서 `evaluate` 뒤 별도 단계로 돈다:
  파도당 shadow 매니페스트 적재 1 회(레인 밖) → 레인마다 순수 함수 `FamilyWorker.ShadowVerdict(input) ShadowOutcome`(관문 결과만 — 봉투 · 오류 반환 없음) → 레인 런타임의
  shadow 관측 메모리에 기록. **자체 마감 · 무오류 · 자가 recover**: 적재 실패 · panic · 마감 초과는 그 파도의 shadow 관측을 「없음」으로 둘 뿐, 레인 latch · 원장 · 시장 주기
  오류 · 레거시 경로에 아무것도 남기지 않는다.
- fault 핀: shadow 단계에 panic · 지연 · 적재 오류를 주입 → 레인 latch 0 · 원장 행(lease · 레인 잠금 · 잠금 복구) 0 · 시장 주기 무오류 · 레거시(미선언) 스냅숏 불변 · 시장 비잠금.

## 6. 「OFF 레인」 술어 하나 (판정 5 · 보이스 1 P1-6)

- 공개 함수 하나 `strategyworker.ShadowEligible(activation FamilyActivation, worker FamilyWorker) bool` = `Desired(activation)==OFF ∧ Effective(activation)==OFF`. worker 의 shadow
  적용(활성화 ON 레인은 shadow 아님 — codex P1 「ON 우선을 판단할 정보 없음」 의 해법: 호출 경계가 같은 파도의 활성화를 받아 이 술어로 거른다)과 `validateLane` 의 교차 규칙이
  **같은 판정** 을 쓴다(투영 쪽은 같은 조건의 값 형태: runtime=SHADOW ⇒ desired=OFF ∧ effective=OFF).
- 교차 규칙 추가 조건: runtime=SHADOW ⇒ health≠nil ∧ cycleGeneration>0(관측된 레인만 SHADOW).
- **거부되는 정상 입력 열거**(보이스 3 P1-4 · fail-closed 교훈): 관측 전 레인(cycleGeneration 0) · 활성화 ON 레인 · 잠긴 레인(LATCHED — shadow 는 관측만이므로 허용 여부 명시:
  허용, health 그대로) · 닫힌 시장(관측 없음 → UNOBSERVED). 각 경우의 기대 투영을 표로 시험.

## 7. SHADOW 값의 출처 (판정 6)

- SHADOW 는 **projection 어휘(`strategyprojection.LaneRuntime`)와 레인 런타임 관측에만** 있다. router `RuntimeState` 는 {UNOBSERVED} 유지(MarketRecord 공유 — 열지 않음).
  worker 의 고정 `runtime` 필드도 그대로(UNOBSERVED).
- 투영은 **관측** 에서 runtime 을 읽는다: 그 파도에 shadow 관측이 있고 §6 술어가 참이면 SHADOW, 아니면 UNOBSERVED. 가산 필드 `shadowOutcome`(WOULD_EMIT · NOT_THIS_LANE ·
  NO_INPUT) — runtime≠SHADOW 이면 null.
- R2 핀: `TestTheRuntimeVocabularyIsExactlyUnobservedUntilAShadowLotExtendsIt` → 둘로: router `RuntimeState` = 정확히 {UNOBSERVED}(유지), projection `LaneRuntime` = 정확히
  {UNOBSERVED, SHADOW} — 어휘 census 는 **타입을 가진 const + 변환식(`LaneRuntime("…")`) 전수**(go/types).

## 8. dispatch 핀 = 차등 척도 (판정 7)

- 같은 입력 · 같은 활성화에서 **shadow 핀 유 / 무** 두 실행의 Submit · handoff · 게이트웨이 스파이 궤적이 **같다**(0 을 단언하지 않음 — 활성화 ON 레인이 있으면 정상 dispatch 가
  있으므로 「0」 은 틀린 척도, 보이스 3 P1-2).
- 두 배치: 미선언 시장 · 부분 ON 선언 시장(일부 가족 ON, 나머지 OFF + shadow). 전제 단언: shadow 쪽 WOULD_EMIT ≥ 1(빈 표본 통과 금지). 실행 단위: `runProductionStrategyMarketCycle`
  전체 주기(레인 런타임 · 투영 · dispatch 포함).

## 9. OpenAPI (판정 8)

- enum 동기 시험 신설: `StrategyRuntimeLaneRuntime.runtime.enum` == projection 어휘 함수(`LaneRuntimes()`) · `shadowOutcome.enum` == `ShadowOutcomes()`.
- wire 모양: properties 가 전부 required 라 `shadowOutcome` 은 **항상 직렬화**된다 → v1 의 「핀 0 = wire 불변」 을 **「핀 0 = shadowOutcome 값 null · runtime UNOBSERVED」** 로 정정.
  핀: shadowOutcome≠null ⇔ runtime=SHADOW.
- 내부 계약 판단 기록에 `additionalProperties:false` 추가: 외부 소비자 등록부 없음 · 저장소 안 읽기는 Validate 경유 · RPC 디코더는 모르는 필드 허용 · 엔진/콘솔 같은 이미지.

## 10. 재시작 핀 셋 (판정 9 · 시나리오 :91-94)

① 핀 없음 → 여덟 OFF/OFF/UNOBSERVED · dispatch/activation write 0. ② 핀 있음 + 못 씀(파일 없음 · 불일치 · 만료 · 폐기 · 결속 불일치 — 각 모양) → 같음. ③ 유효 핀 잔존 → 재시작 뒤
**첫 물결 전** 여덟 UNOBSERVED · SHADOW 는 재독(첫 물결)에서만 · 이전 프로세스의 wave/outcome 0(관측 메모리 비복원). 세 개 모두 R2 재시작 시험과 같은 파일에 짝.

## 11. census 품질 · 활성화 쓰기 0 (판정 10)

- 어휘 census: 타입 보유 const + 변환식 전수(§7). 호출 폐포: 전이 + 하한 단언(§2 ③). 「활성화 쓰기 0」: 설정 디렉터리 **목록과 각 파일 바이트** 의 전후 동일(4.5-e 시험 모양) —
  shadow 단계 실행 전후.

## 12. 8.6 (판정 ⑥ — 변경 없음)

SHADOW 는 능력이지 배포가 아니다. 생산 shadow 핀 0 → shadow 단계는 파도마다 미선언 판정만(파일 I/O 0 — 적재기 첫 줄이 핀 비었음). 8.6 BLOCKED 유지, :295 의 shadow 수집은
배포 뒤(사람 · A100).

## 13. 로트 계획

1. re-freeze: 같은 4판이 **자기 P1 종결만 표적 재검**(Manager 판정 12) — 재검 전원 PASS 뒤 진행.
2. Pre-Edit FLM(§4 목록 + 활성화 적재기 무편집 확인용 AST 기준) → RED(§2 · §3 · §5 · §6 · §8 · §9 · §10 핀 전부, 편집 전 실패 기록).
3. GREEN: `internal/strategyshadow` · worker `ShadowEligible`/`ShadowVerdict` · 운반 타입 · shadow 단계 · 투영 · validateLane · OpenAPI · 생성기 · 골든 · 문서 · ROADMAP 행 정정.
4. 변이(§2 census 각 항 · 교차 디코드 · sentinel 배타 · 술어 공유 · fault 격리 · 차등 dispatch · 재시작 셋) · race 목록 · verify · 착지 → 재고정 영수증 최종 갱신.
