# 7.3.1 SHADOW 설계 브리프 v3.2 (코드 0 — 2026-10-05)

**v3.2 변경(재검 4라운드 3판 — codex FAIL(P1, 자가 무효: pipefail 누락) · 보이스 1 FAIL(N2 PARTIAL · NP1) · 보이스 3 FAIL(P1-C PARTIAL · P1-D), P0 0 — Manager 판정 2026-10-05 「1 — v3.2
승인 그대로」):** §5 칸 = `{wave, batch, activation}` 같은 임계 구역(보이스 3 P1-C) · §2 ③ 금지 집합에 활성화 적재 셋 · §5 핀 ① 을 타입 규칙으로 재진술 + `forMarket` shape 핀 + 인자 index 5
(보이스 1 NP1-1 · 보이스 3 P1-D) · §4 수집 helper · 생성자 shape 핀(보이스 1 NP1-2 — 주입 seam 미채택 사유 기록) · §5 비-nil 반환 시 관측 즉시 폐기 + epoch CAS 게시(codex P1 · 보이스 1 N2) ·
§5.1 관측 나이 backstop(보이스 1 N2) · ④ 문구 · §4 「shadow 식별자 0」 범위 · §0 좌표 · §2 unsafe/reflect 핀 범위(생산 그래프만). v3.1 은 `design-brief-v3.1.md`.

**v3.1 변경(재검 3라운드 4판 — codex FAIL(N2) · 보이스 1 FAIL(N2 PARTIAL · N3) · 보이스 2 PASS(노트 3) · 보이스 3 FAIL(P1-1 · FLM PARTIAL · P1-C), P0 0 — Manager 판정 2026-10-05 「1 — v3.1
승인」):** §2 unsafe/reflect 는 모듈 패키지의 직접 import 로 정밀화(codex N2) · §2 ③ shadow 단계 금지 집합에 refresh · 조립 · 파도 합류 함수와 캐시 필드 · §3 wrapper 읽기 전용 명시(보이스 2
노트 3) · §4 계보 충돌 닫힘은 「관측 없음」 + 수집 helper 문장 핀(보이스 3 P1-1 · 보이스 2 노트 1) · §4 FLM 목록 정정(보이스 3) · §5 **묶음 출처 = evaluate 가 파도 번호와 같은 잠금에서
보관한 값(evaluate 인자 한 자리 — 핀 ① 재진술, Manager 판정 (A)), 스탬프는 시작 시점**(보이스 1 N3 · 보이스 3 P1-C · 보이스 2 노트 2) + 교차 시장 fault 핀 · race 목록 · §5.1 만료를 투영 시점에 검사 + 핀 (e)(보이스 1 N2). v3 는 `design-brief-v3.md`.

**v3 변경(재검 2라운드 4판 — 보이스 1 · 2 · 3 FAIL · codex FAIL, P0 0 — Manager 확정 8항 + amendment v3 4d22d726):** §1 일치 주장 삭제 · §2 unsafe/reflect 금지 + census 를 types.Info.Uses · 본문 식 타입으로 · §3 활성화 헬퍼는 새 파일의 중립 wrapper export(기존 함수 무편집) · §4 운반은 authority 밖 별도 값 + opaque `ShadowInput` + 조정 **뒤** 닫힘 여섯도 운반 · §5 shadow step 은 dispatch **뒤**(cycle 함수 밖) · 새 §5.1 프로세스 내 소거 · §10 재시작 전제 · FLM 목록 확정(collectMarket 무조건). v2 는 `design-brief-v2.md`.

v1(`design-brief-v1.md`)은 freeze 4판(codex REJECT · 보이스 1 REJECT · 보이스 2 APPROVE-WITH-CHANGES · 보이스 3 — `analysis/review-shadow-freeze-2026-10/`)에서
설계 수정 → re-freeze 판정을 받았다. v2 는 Manager 합본 판정 12 항을 반영한다. 계약: spec four-family-strategy-runtime :38(「dispatch handoff 0 — SHADOW 반사실 포함」) ·
:88-94(개정 [결정 63-v2] — :89 digest 핀 SHALL · MUST NOT 둘, :91-94 시나리오 「유효한 shadow manifest 없이 — 핀 없음, 또는 핀은 있으나 파일 없음 · 핀 불일치 · 만료 · 폐기 ·
결속 불일치」), design 결정 61 · 62 · 63(-v2) · :293(「dormant/shadow 로 배선」) · :295(「shadow 관측 수집」은 배포 뒤). amendment: c7219640 · 2817064c · 4d22d726(결정 63 v3 — 61 의 (a)·(b) 만 적용, 61(c) 비재채택).

## 0. 코드 사실(v1 정정 포함, 좌표 = HEAD 4d22d726 — 2817064c 이후 Go 변경 0)

- **v1 §0 의 「이미 있는 제안 위의 레인 관문 반사실」 은 틀렸다**(codex · 보이스 1 P1-1 · 보이스 2 P1-1 · 보이스 3 P1-1). 레인 런타임 입력은 `strategyLaneInputs(accountRef,
  authority)`(strategy_lane_runtime.go:266)가 `authority.entries` — 조정자가 **고른 승자** 만 — 에서 만든다(strategy_market_coordinator.go entries · strategy_proposal_authority.go
  성공 반환). 선언된 시장에서 OFF 레인의 제안은 관문(`gate.admit`)에서 DORMANT 로 멈춰 레인 런타임에 닿지 않고, 닫힌 시장은 entries 0. 그대로 쓰면 SHADOW 반사실은 구조적으로
  NO_INPUT 뿐(빈 표본).
- 관문 **앞** 의 전체 제안은 `coordinateMarketProposals`(strategy_market_coordinator.go:57) 루프의 `batch.LanesFor(symbol)` → 레인별 `strategyworker.Input{Scope, SnapshotDigest,
  Proposal}`(admit 직전 — `gate.admit(` 는 :91, v3.2 좌표 정정 · 보이스 3 재검 4)에 있다.
- 레인 관문 `FamilyWorker.Run`(worker.go:193) · `Desired/Effective`(:107-113) · `Lane.Run`(lane.go:211) · 레인 런타임 `evaluate`(:199) · `runLane`(:286, 유계 step · 마감 ·
  latch 정산) · 시장 주기 `runProductionStrategyMarketCycle`(strategy_entry_supervisor.go:514).
- 투영: `strategyLaneProjection`(strategy_lane_projection.go) — `Runtime` 은 **worker 생성 시 고정 필드**(`lane.Runtime()` → UNOBSERVED)에서 읽는다(보이스 3 P1-6). 검증
  `strategyprojection.validateLane`(lanes.go:274) 이 runtime == UNOBSERVED 정확 일치. `Validate` 는 저장소 · RPC 클라이언트 · httpapi · console 의 관문이지만
  **`Context.Read`(strategy_runtime_projection.go:49-60)는 동적 레인 투영을 덧씌운 뒤 Validate 를 다시 부르지 않는다**(codex P1) — v1 §9 의 「모든 읽기 경로」 는 「외부 경계
  (RPC · httpapi · console)」 로 좁힌다.
- router `RuntimeState`(types.go:40-42)는 `MarketRecord` 와 공유 타입 — SHADOW 를 여기 열면 시장 기록 어휘가 넓어진다(보이스 3 P1-6). OpenAPI `StrategyRuntimeLaneRuntime`:
  `runtime.enum=["UNOBSERVED"]`, `additionalProperties:false`, properties 전부 required — 그 enum 을 고정하는 시험은 **없다**(보이스 3 P1-5).
- 정규 바이트 등식은 **누가 썼는지** 를 증명하지 않는다(codex P1) — 신뢰 앵커는 「배포 핀을 쓸 권한」 이고, 생성기만 만든다는 것은 운영 정책(결정 61 이 이미 그렇게 적음).

## 1. 신뢰 앵커 · 형식 (판정 ① · F1)

digest 핀 `TOSSOS_STRATEGY_FAMILY_SHADOW_<MARKET>_MANIFEST_SHA256` + 정규 바이트 등식 + 0400 · 소유자 · 크기 상한 · 24h 수명 · 폐기 · 결속(활성화 다섯 — 경로 매니페스트 · 보정 ·
달력 · 빌드 · 위험 정책; ProtectionReady 제외). 별도 파일 `strategy-family-shadow-<MARKET>.json`, 서술자 = 그 시장 네 레인 각각 `shadow: ON|OFF` — desired/effective 필드 없음.
생성기 `tools/a112-family-shadow` · 커밋 골든(바이트 + 핀) · 골든을 **만드는** 함수를 부르는 시험 · `docs/operations.md` 절 · 생산 작성자 0 가드.
**신뢰의 근거(결정 63 v3 와 같은 진술):** 정규 바이트 등식은 바이트 동일성만 보장한다 — 신뢰는 배포 핀을 쓸 권한에서 오고, 생성기만 만든다는 것은 운영 정책이다(기술 증명 아님).

## 2. 패키지 · 경계 (판정 ② — (a) 별도 패키지)

- 새 패키지 **`internal/strategyshadow`**: `FamilyShadow`(불투명 — 필드 비공개, 영값 = 아무 레인도 shadow 아님) · `LoadProductionFamilyShadow` · `EncodeProductionFamilyShadow` ·
  shadow 서술자 대조. `strategyrouter` 를 import 하되 **`FamilyActivation` 을 만들 수 없다**(필드가 strategyrouter 비공개 — 재검에서 보이스 1 · codex 가 별도 패키지 주조 시도 = 컴파일
  거절로 실측). **폐포 금지: `unsafe` · `reflect` 의 직접 import**(비공개 필드 우회 차단 — 보이스 1 재검 부기) — 대상은 strategyshadow 와 그 의존 폐포의 **모듈 안 패키지**(`github.com/JungHoonGhae/tossinvest-cli/…`) 의 import 목록; 표준 라이브러리 내부 사용(`encoding/json` 등이 reflect · unsafe 에 닿음 — codex 재검 3 N2 실측)은 예외. 2026-10-05 실측: strategyrouter 폐포의 모듈 패키지 중 둘을 직접 import 하는 것 0. 핀은 **`strategyshadow` 를 루트로 한 생산 그래프 둘**(`ListDeps` 의 deps · deps-tagged)에서 모듈 노드의 `Imports` 를 본다 — 시험 그래프(deps-test · deps-test-tagged)는 범위 밖(v3.2 — codex 재검 4 실측: router 의 태그 시험 그래프는 `a127_real_journal_test.go` → journal · exitpolicy 의 생산 reflect 에 닿는다; 이 핀이 막는 것은 생산 바이너리의 비공개 필드 우회이고 시험 코드는 그 경로가 아님 — Manager 판정 5). router · worker 자체 시험 그래프가 0 이라고 주장하지 않는다.
- census(go/types 해소, ShadowBand 선례 `internal/candidate/band_test.go:327` 모양) — **선언 타입이 아니라 사용으로 센다**(보이스 3 P1-A · 보이스 1 N1):
  ① `strategyrouter.FamilyActivation` 의 비영 composite literal 은 저장소 전체에서 정확히 1 자리(LoadProductionFamilyActivation 반환) — 시험 seam 파일은 이름으로 제외;
  ② `strategyshadow.FamilyShadow` · `strategyworker.ShadowInput`(§4) · 운반 값 타입의 **사용**: `types.Info.Uses` 로 그 타입 · 그 접근자 `*types.Func` 를 가리키는 식별자, 그리고 함수 본문
  모든 식의 타입(`types.Info.Types` — 별칭 · any 변환 · 제네릭 인스턴스 · 함수 값 포함 「품는다」 걸음)에 그 타입이 나타나는 함수 = **허용 목록**(shadow 운반 · shadow 단계 · 투영 — 운반에는 `runProductionStrategyMarketCycle` 의 `evaluate` 인자 한 자리를 이름으로 포함, §5 핀 ①)뿐,
  `strategyFamilyGate.admit` · `coordinateMarketProposals` 의 조정 경로 · `dispatchHandoffs` · `dispatchHandoff` · `authorityForOwnerScope` · `dispatch*` · `strategyworker.Lane.Run/Desired/Effective` ·
  `FamilyWorker.Run/Desired/Effective` 에서 0. **양성 대조:** 사본 변이(새 접근자 메서드 경유 세탁 — 보이스 3 P1-A 의 모양)를 census 가 잡는 시험을 같이 둔다;
  ③ shadow 단계 호출 폐포(전이)에 조정자 Submit · handoff · 게이트웨이 · 원장 쓰기 · 활성화 인코더 0, **그리고 `refreshPairedStrategyEntryProductionAssembly` ·
  `NewPairedStrategyEntryProductionAssembly` · `joinStrategyRefreshWave` · `collectStrategyRefreshWave` · `awaitStrategyRefreshWave` 0 + 공유 캐시 필드(`Context.strategyRefresh` ·
  `strategyRefreshAt` · `strategyRefreshWave`)의 `types.Info.Uses` 0 + 활성화 적재 `strategyProposalAuthorityLoader.loadFamilyActivation`(strategy_family_activation.go:154) ·
  `familyGateFor`(:115) · `strategyrouter.LoadProductionFamilyActivation` 0(v3.2 — 보이스 3 재검 4 P1-C: shadow 단계의 활성화는 칸의 복사본뿐)**(v3.1 — 보이스 1 N3 · 보이스 3 P1-C: 이 함수들은 원격 권한 파도를 일으키거나 두 시장이 함께 읽는 스냅숏을 바꾼다) — 하한 단언(양성 대조:
  사본에서 shadow 단계에 refresh 호출을 심으면 잡힘) 동반;
  ④ 경계 패키지 소스 digest 동결(strategyhandoff `source_freeze_test.go` 선례).

## 3. 적재기 — 활성화 적재기 무편집, 사본 + 양쪽 AST 핀 (판정 ③ · 재검 확정 2 = (a))

- `strategyrouter.LoadProductionFamilyActivation` 과 그 도우미 **기존 함수는 편집하지 않는다**.
- 공유가 필요한 strategyrouter 비공개 도우미(`readProductionRouteFile`(production_owner_unix.go:12 · `!unix` 스텁 production_owner_other.go:8) · `productionRouteOwnerUID` · `productionRouteDigest`
  (production.go:768) · `productionRouteTime`(:751) · `productionRouteIdentity`(:756) · `productionRouteDigestValid`(:760) · `productionRouteDescriptors`(:586))는 **새 파일**
  (`internal/strategyrouter/production_shared_export.go`)에서 **중립 wrapper 로 export** — 기존 함수를 부르기만 하는 한 줄 함수, **읽기 전용**(파일 · 상태 쓰기 wrapper 0 — 새 파일의
  export 목록을 이름으로 고정하는 AST 핀, 보이스 2 재검 3 노트 3), 서술자 표는 **하나**(복사 금지 — 적재기 주석
  production_family_activation.go:537-539). 읽기 wrapper 는 strategyrouter sentinel 을 그대로 돌려주므로 shadow 적재기는 그 오류를 **`%v` 로 접어** 자기 sentinel 로 감싼다
  (`errors.Is(shadowErr, ErrProductionRouteUnavailable)==false` 핀 — 8.5 P2-d 와 같은 규칙).
- shadow 적재기 본문은 형태를 옮긴 사본이고 **양쪽을 AST 로 못 박는다**(같은 검사 순서 — 미선언 맨 앞 · ctx · 설정 결속 · 파일 읽기 · 핀 · 정규 · 폐기 · 결속 · 수명 · 서술자; 한쪽만
  바뀌면 실패). wrapper 층도 그 핀에 넣는다(wrapper 가 기존 함수 하나만 부름 — AST).
- sentinel 은 패키지마다 따로 + **errors.Is 배타 핀**; 교차 디코드 거절 **양방향** 핀.

## 4. 입력 — 관문 앞 배치에서 분기, authority 밖 별도 값 (판정 1 · 재검 확정 1 · 3)

- `coordinateMarketProposals` 루프에서 각 레인 제안을 admit **앞** 에서 `strategyworker.ShadowInput`(**opaque** — 필드 비공개, 생성자는 engine 운반 한 자리만, 받는 것은
  `FamilyWorker.ShadowVerdict` 하나; `strategyworker.Input` 이 아님 — 그래서 `admit` · `Submit` 이 받을 수 없다)으로 모은다.
- 운반은 **authority 구조체 밖 별도 값**: `coordinateMarketProposals` 가 조정 결과와 **나란히** `strategyShadowBatch`(ShadowInput 목록 — 값 타입)를 돌려주고, `collectMarket` 이
  `(strategyProposalMarketAuthority, strategyShadowBatch)` 로, `collect` 가 짝 구조 옆의 `strategyShadowPair` 로, 조립(`StrategyEntryProductionAssembly`)이 `proposals` 와
  **별개 필드** 로 든다. `strategyProposalMarketAuthority` · `strategyMarketArbitration`(dispatchHandoffs · dispatchHandoff · authorityForOwnerScope · entries 의 수신자)에는 **넣지
  않는다**(codex N1 · 보이스 1 N1).
- **운반 범위(확정 1 · v3.1 정정):** 조정 **뒤** 닫힘 다섯 — FAMILY_GATE_CLOSED · QUEUE_OVERFLOW · ARBITRATION_REFUSED · 미해결 선택 · NO_ACCEPTED_SCOPE — 과 성공은 루프를 끝까지
  돈 뒤의 관문 앞 입력을 **싣는다**(OFF-단독 제안 종목 = SHADOW 의 대상 모집단 — 보이스 1 · 2 · 3 재검 P1-1 재현: 선언 시장에서 OFF 가족만 제안한 종목이 범위를 지워 FAMILY_GATE_CLOSED).
  **계보 충돌 닫힘은 「관측 없음」**(v3.1 — Manager 판정 3): 그 갈래는 루프 **도중** `return` 하므로(strategy_market_coordinator.go:113-115, 루프 :64) 그때까지 모은 입력은 뒤쪽 종목이
  빠진 부분 묶음이고, 실으면 그 레인들이 NO_INPUT 으로 거짓 보고된다(보이스 3 재검 3 P1-1 · 보이스 2 노트 1). 조정 루프 순서는 바꾸지 않는다(High-risk 경로 불필요 교란, 계보 충돌은 병리
  케이스라 진단 손실 수용). 조정 **앞** 닫힘 일곱(ROUTE_NOT_READY · FX · 설정 · 열쇠 · 중복 · 적재 · 고장)도 「관측 없음」(수집 전 — 정직).
  - 「관측 없음」 은 **빈 묶음이 아니라 부재 값**(`strategyShadowBatch` 의 관측 여부 필드 거짓)이다 — 빈 목록이면 OFF 레인 전부가 NO_INPUT 이 되어 같은 거짓이 된다. 시험: 부재 값 → 그
    시장 여덟 투영 UNOBSERVED · null; 빈 관측 묶음(제안 0 의 정상 파도) → OFF 레인 NO_INPUT.
  - 구조 핀(AST): ⓐ `coordinateMarketProposals` 안 shadow 수집은 **안쪽 루프의 문장 수준 helper 호출 하나**(`gate.admit` 문장 **앞**; 조정 경로 — admit · Submit · 계보 색인 · Arbitrate — 의
    식에 shadow 사용 0, 수집 helper 호출과 반환 운반은 허용 — v3.2 범위 정정); ⓑ 충돌 `return` 문의 shadow
    반환값은 부재 값 리터럴; ⓒ 정상 `return`(루프 뒤 `Arbitrate` 다음) 만 수집된 묶음을 돌려줌. 변이: 충돌 갈래가 부분 묶음을 돌려주면 ⓑ 와 행동 시험(뒤쪽 종목 OFF 레인이 NO_INPUT 이 아님)
    둘 다 실패.
  - **수집 helper · 생성자 shape 핀(v3.2 — 보이스 1 재검 4 NP1-2):** helper 본문 = `strategyworker.ShadowInput` 생성자 결과를 묶음에 `append` 하는 문장 하나(다른 호출 · 분기 · 순회 0),
    생성자 본문 = 값 필드의 composite literal 반환 하나(호출 0). 이 모양에는 panic 거리가 없으므로(nil 역참조 · 색인 · 변환 0) `collect` 의 recover(strategy_proposal_authority.go:270-276)가
    shadow 때문에 INTERNAL_FAILURE 로 닫는 길이 구조로 막힌다. **panic 주입 fault 핀 미채택 사유:** 구조적으로 panic 거리가 없는 helper 에 주입하려면 High-risk 조정자에 시험 seam 을
    새로 뚫어야 한다 — seam 을 위한 seam(Manager 판정 4).
- 구조 핀: `dispatchStrategyMarketHandoffs` 의 handoff 원천은 `proposals.forMarket(market).dispatchHandoffs()` 그대로(기존 :169-175 핀), shadow 값은 그 식에 없음(§2 ② census).
- FLM 확정 목록(재검 확정 7 · v3.1 정정 — 보이스 3 재검 3): `coordinateMarketProposals` · `strategyProposalAuthorityLoader.collectMarket`(**무조건** — 반환 모양이 바뀜) ·
  `strategyProposalAuthorityLoader.collect` · **`Context.NewPairedStrategyEntryProductionAssembly`**(strategy_entry_supervisor.go:293-374 — `collect` 반환을 받아 나눠 주고 조립 리터럴
  :366 이 있는 실제 편집 자리, High-risk 조립 함수) · `strategyLaneRuntime.evaluate` · `strategyLaneRuntime.record`(§5 — 묶음 보관) · `strategyLaneRuntime.runLane` ·
  `Context.runProductionStrategyMarketCycle`(§5 — evaluate 인자 한 자리) · `Context.NewRefreshingPairedStrategyEntrySupervisor`(§5 — cycle 클로저) · `Context.productionStrategyWorker` ·
  **`strategyLaneRuntime.projection`**(strategy_lane_projection.go:31 — §5.1 신선도 · 만료 판정이 관측 조회를 바꿈) · `strategyLaneProjection` · `strategyprojection.validateLane`.
  **무편집 사유:** `Context.refreshPairedStrategyEntryProductionAssembly` — 캐시 합류 · 위임만 하고(:575-596) 서명 · 본문이 바뀌지 않는다(§5 의 묶음 출처는 evaluate 경로라 refresh 서명
  무변경) — FLM 에서 「무편집」 AST 대조(편집 전후 본문 digest 동일)로 확인. `strategyLaneInputs` — authority.entries 만 읽고 shadow 와 무관(§4 운반이 authority 밖) — 같은 방법.
  전부 Pre-Edit FLM 먼저.

## 5. 실행 자리 · 고장 격리 (판정 4 · 재검 확정 4 — dispatch 뒤, cycle 함수 밖)

- `runProductionStrategyMarketCycle` 은 **바꾸지 않는 쪽이 기본**: 마지막 문장이 dispatch(기존 핀 `a112_market_delivery_structure_test.go:169-175`)이고, 그 함수가 shadow 를 하면 dispatch
  앞(주문 경로 지연)이 된다(보이스 1 P1-4 · 보이스 3 P1-B). 대신:
  - 조립(`NewPairedStrategyEntryProductionAssembly`)이 shadow 묶음 짝을 `StrategyEntryProductionAssembly` 의 별개 필드로 둔다(§4).
  - **묶음 출처(v3.1 — 보이스 1 N3 · 보이스 3 P1-C · Manager 판정 1):** 그 시장 묶음은 `evaluate` 가 받아 `record` 가 **파도 번호를 올리는 같은 잠금 안에서** 레인 런타임에 값으로
    보관한다(시장별 칸 하나 = **`{wave, batch, activation}`** — activation 은 evaluate 가 이미 받는 `promotion`(strategy_lane_runtime.go:199-201), 셋을 한 임계 구역에서 함께 씀 · 덮어씀,
    누적 없음 — v3.2, 보이스 3 재검 4 P1-C · 보이스 1 노트). 그래서 「파도」 는 **묶음을 실어 온 evaluate 의 파도 번호** 로 정의되고 묶음과 파도는 구성으로 짝이다(공유 캐시
    `c.strategyRefresh` 는 그사이 다른 시장이 새 파도를 게시할 수 있어 쓰지 않는다 — §2 ③ 금지).
  - shadow 단계는 **cycle 클로저**(NewRefreshingPairedStrategyEntrySupervisor · productionStrategyWorker 의 `Cycle: func…`)에서 `runProductionStrategyMarketCycle` 이 **nil 로 돌아온
    뒤** 시작한다(오류면 시작하지 않음). **스탬프는 시작 시점**(보이스 2 노트 2): 클로저가 **동기로** 레인 런타임 읽기 잠금 아래에서 그 시장 칸 `{wave, batch, activation}` 과 그 시장
    `shadowEpoch` 를 한 번 복사하고, 비동기 작업은 그 복사본만 쓴다 — 작업 도중 새 파도가 와도 결과는 복사한 wave 로 찍히고 §5.1 이 낡은 것을 버린다.
  - **실패 시 즉시 폐기 + epoch CAS 게시(v3.2 — codex 재검 4 P1 · 보이스 1 N2, Manager 판정 1(a)):** record **전** 실패(`recoverMarketLanes` 오류 strategy_lane_runtime.go:208-209 ·
    레인 panic 재던짐 :244-247)는 파도를 올리지 않으므로 파도 등식만으로는 이전 SHADOW 가 남고, panic 은 시장을 잠가(`latchMarket`) 주기 자체가 멈춘다(`evaluationState`
    strategy_entry_supervisor.go:1035). 그래서:
    - cycle 클로저는 **recover 하지 않는 defer** 하나를 둔다 — 주기 함수가 nil 로 돌아오지 않았으면(오류 · panic 전파 중) 레인 런타임 `mu` 아래에서 그 시장 `shadowEpoch++` 와 그 시장
      shadow 관측 삭제를 한 번에 한다. panic 은 그대로 위로 간다(기존 `invokeStrategyCycle` 회복 · latch 경로 불변). defer 본문 = 레인 런타임 메서드
      `invalidateShadow(market)` 호출 하나, 그 메서드 본문 = 잠금 · epoch 증가 · 관측 삭제뿐(다른 호출 · I/O 0) — 둘 다 AST shape 핀.
    - shadow 단계의 게시는 **CAS**: `mu` 아래에서 현재 `shadowEpoch` == 복사한 epoch ∧ 현재 칸 wave == 복사한 wave 일 때만 관측을 쓴다 — 실패 뒤 늦게 끝난 in-flight 작업이 관측을
      되살리지 못한다.
    - 핀: SHADOW 관측(wave W) → 만료 전 (i) 복구 오류 (ii) record 전 panic → **즉시**(다음 투영) UNOBSERVED · null; (iii) W 의 in-flight 작업을 실패 **뒤** 완료시켜도 UNOBSERVED · null 유지;
      (iv) 주기 오류 반환 · panic 전파 · dispatch 궤적이 shadow 없음과 동일. record **뒤** 실패(`persistMarketLatches` · `staleLatchError`)는 파도가 올라 §5.1 신선도로 이미 버려지지만
      defer 가 같이 폐기한다(둘 중 하나만 지워도 통과하는 쌍이 되지 않도록 (i)~(iii) 은 파도 불변 경로로만 잰다).
    시장당 단일 비행 비동기 작업(이전 것이 아직 돌면 이번 파도는 건너뜀 · 건너뛴 수를 셈), 자체 상수 마감 `strategyShadowStepDeadline`(예: 2s — 상수, 레인 마감 · 주기 한도 30s 와
    독립), 주기 ctx 와 분리된 ctx(주기 취소가 shadow 를 닫되 shadow 가 주기를 붙잡지 않음). 그 클로저의 반환값은 주기 함수의 오류 그대로(shadow 는 무오류).
  - 파도당 shadow 매니페스트 적재 1 회 → 레인마다 순수 함수 `FamilyWorker.ShadowVerdict(ShadowInput) ShadowOutcome` → 레인 런타임 shadow 관측 메모리(복사한 wave · 매니페스트
    `expiresAt` 과 함께). 자가 recover(panic → 그 파도 「관측 없음」), 마감 초과 → 「관측 없음」. shadow 단계가 **읽는 것은 복사본 · 매니페스트 파일 · 시계뿐**(원격 I/O 0 — §2 ③).
- AST 핀(v3.1 — Manager 판정 (A), 2026-10-05 · v3.2 타입 규칙 재진술): ① `runProductionStrategyMarketCycle` — **타입 규칙**: 본문에서 shadow 타입을 갖는 식은 `evaluate` 호출의 **인자
  index 5**(마지막 — 기존 잠금 복구 세대 핀이 `Args[2]` 를 보므로 그 앞에 끼우지 않음, a112_lane_latch_durability_test.go:335-348) 의 부분 트리 하나뿐이고, 그 부분 트리의 유일한 호출은
  접근자 `strategyShadowPair.forMarket`(`fresh.<shadow 필드>.forMarket(market)` 꼴) · 마지막 문장 dispatch(기존 :169-175 핀) 유지. 그 자리는 §2 ② 허용 목록에 **이름으로** 박는다(운반 —
  은닉이 아니라 명시). **`forMarket` shape 핀(v3.2 — 보이스 1 NP1-1 · 보이스 3 P1-D):** 본문 = 시장 비교 후 필드 반환뿐 · 호출 0 — 형제 `strategyProposalAuthorityPair.forMarket`
  (strategy_proposal_authority.go:167-175) 과 같은 모양을 AST 로 고정(그 메서드에 적재 · 판정을 넣는 변이가 잡힘) ② cycle 클로저에서 shadow
  시작 문장이 `runProductionStrategyMarketCycle` 호출 **뒤**, 그 오류가 nil 인 갈래 안 ③ dispatch 입력(`dispatchHandoffs()` 원천)이 shadow 값과 무관(§2 census) ④ `evaluate` · `record` 안에서
  묶음에 대한 메서드 호출 · 순회 0 — 허용은 `evaluate` → `record` 인자 전달과 `record` 안 보관 대입 한 자리뿐(v3.2 문구) ⑤ 그 대입은 `record` 의 잠금 안에서 파도 번호 증가와 **같은 임계
  구역**(시장 칸 `{wave, batch, activation}` 을 한 번에 씀).
  - 근거: 핀 ① 의 실체는 「dispatch 앞에 shadow **일** 0」(보이스 1 P1-4 · 보이스 3 P1-B)이지 식별자 철자가 아니다. `evaluate` 의 생산 호출은 그 함수 안 한 자리뿐이고
    (strategy_entry_supervisor.go:543-546 — 2026-10-05 grep) 묶음은 그 함수의 지역값 `fresh` 에만 있으므로, evaluate 시점 보관은 그 인자 한 자리를 요구한다. 인자 묶음 타입으로 식별자만
    감추면 §2 ② 가 같은 사용을 세므로 세탁이 된다 — 그래서 감추지 않고 허용 목록에 이름을 박는다. 주기 경로가 하는 shadow 일은 잠금 안 값 대입 하나이고 I/O · 판정 · 적재 0(④ · ⑤).
  - 기각한 대안 (B): refresh 시점 시장별 칸 + 클로저 동기 읽기 — 철자를 지키려고 High-risk 경로에 「같은 시장 주기 비중첩」 이라는 새 동시성 불변식을 세우고 `refreshPaired` 를 편집하는
    거꾸로 된 거래(Manager 판정 (A) 사유).
- fault 핀: shadow 단계에 panic · **마감 초과 지연**(가짜 시계로 마감 + δ) · 적재 오류 주입 → 같은 주기의 dispatch 가 정상 수행(차등 궤적 = shadow 없음과 동일) · 레인 latch 0 · 원장 행
  (lease · 레인 잠금 · 잠금 복구) 0 · 주기 무오류 · 시장 비잠금 · 레거시 스냅숏 불변.
  **교차 시장(v3.1 — 보이스 1 N3):** 시장 A 의 shadow 단계가 마감 초과로 멈춘 동안 시장 B 가 주기 도중(refresh 합류 · evaluate · dispatch) → B 는 잠기지 않고(B 레인 latch 0 · B 원장
  잠금 행 0) B 의 dispatch 궤적 = shadow 없음과 동일 · B 의 refresh 파도 수(지도자 호출 수)가 shadow 없음과 동일(shadow 가 파도를 일으키지 않음).
- race: shadow 비동기 단계 · 교차 시장 fault 시험 이름을 `make test-race` 의 엔진 동시성 이름 목록에 등재(완전성 가드가 새 이름을 요구 — race-detector 교훈).

### 5.1 프로세스 내 소거 (재검 확정 5 · 보이스 1 N2)

- shadow 관측은 **그 파도 번호가 최신 evaluate 파도와 같을 때만** 투영에 쓰인다(신선도 규칙). 그 파도에 shadow 관측이 없으면(매니페스트 철회 · 만료 · 결속 불일치 · 핀 제거 · 시장 닫힘 ·
  shadow 실패 · 건너뜀) 투영은 UNOBSERVED · `shadowOutcome` null.
- **만료는 투영 시점에도 본다(v3.1 — 보이스 1 N2 PARTIAL):** 파도가 더 오지 않으면(주기 정지 · 시장 장 마감 등) 신선도 규칙만으로는 마지막 SHADOW 관측이 매니페스트 만료 뒤에도 남는다.
  그래서 shadow 관측은 그 파도에 적재한 매니페스트의 `expiresAt` 을 함께 들고, `strategyLaneRuntime.projection` 은 런타임 시계(`runtime.clk.Now()` — 호출자 서명 무변경)가
  `expiresAt` 이상이면 그 관측을 쓰지 않는다(UNOBSERVED · null). 판정 함수는 하나(`shadowObservationUsable(now, latestWave, observation)` — 파도 등식 ∧ 미만료 ∧ 나이 상한)이고 투영은
  그 함수만 부른다.
- **나이 상한 backstop(v3.2 — 보이스 1 재검 4 N2, Manager 판정 1(b)):** 실패 없이 파도가 멈추는 경우(supervisor 정지 · accepting 거짓 — §5 의 defer 가 돌지 않음)에도 폐기 · 철회가 무한히
  가려지지 않도록 관측은 게시 시각 `observedAt` 을 들고, `now − observedAt ≥ strategyShadowObservationMaxAge` 면 쓰지 않는다. **상수 값은 유도한다(⚠ Manager 확인 — 판정 문면 「2×PollInterval」
  에서 바꿈):** 건강한 두 관측 사이 최대 간격 = 폴 간격 `DefaultStrategyCycleLimit`(5s) + 주기 한도 `MaximumStrategyCycleLimit`(30s) + `strategyShadowStepDeadline`(2s) = 37s 이므로
  2×PollInterval(10s)은 **정상 입력을 거부**한다 — 10s 넘게 걸리는 건강한 주기마다 SHADOW 가 UNOBSERVED 로 깜박인다(fail-closed 는 무엇을 거부하는지 말해야 한다). 그래서
  `strategyShadowObservationMaxAge = 2 × (DefaultStrategyCycleLimit + MaximumStrategyCycleLimit + strategyShadowStepDeadline)` = 74s, 세 상수에서 식으로 유도(숫자 사본 금지 — 등식을 시험).
  핀: 정확히 MaxAge → UNOBSERVED · null, MaxAge − 1ns → SHADOW 유지; **거부 안 함 대조:** 주기 한도 30s 를 다 쓴 건강한 주기 연속에서 SHADOW 가 끊기지 않음.
- 핀: 같은 프로세스에서 SHADOW 관측 → 다음 파도에 (a) 매니페스트 폐기 (b) 만료 (c) 시장 닫힘(조정 앞 닫힘 · 계보 충돌) (d) shadow 단계 실패 — 각 경우 투영이 UNOBSERVED · null(이전
  파도의 SHADOW/WOULD_EMIT 이 남지 않음); **(e) 새 파도 없음 + 가짜 시계를 매니페스트 `expiresAt` 으로 → 다음 투영 UNOBSERVED · null**(경계: `expiresAt − 1ns` 는 SHADOW 유지 —
  등호 쪽 변이를 잡음); (f) 새 파도 없음 + 실패 없음(supervisor 정지) + 가짜 시계를 `observedAt + MaxAge` 로 → UNOBSERVED · null.

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
- 두 배치: 미선언 시장 · 부분 ON 선언 시장(일부 가족 ON, 나머지 OFF + shadow). 전제 단언: shadow 쪽 WOULD_EMIT ≥ 1(빈 표본 통과 금지). 실행 단위: **cycle 클로저 전체**(`runProductionStrategyMarketCycle` + 그 뒤의
  shadow 단계 — 레인 런타임 · 투영 · dispatch 포함). 전제의 두 배치는 OFF-단독 제안 종목을 포함한다(§4 운반 범위 — 조정 뒤 닫힘에서도 shadow 관측이 서는지).

## 9. OpenAPI (판정 8)

- enum 동기 시험 신설: `StrategyRuntimeLaneRuntime.runtime.enum` == projection 어휘 함수(`LaneRuntimes()`) · `shadowOutcome.enum` == `ShadowOutcomes()`.
- wire 모양: properties 가 전부 required 라 `shadowOutcome` 은 **항상 직렬화**된다 → v1 의 「핀 0 = wire 불변」 을 **「핀 0 = shadowOutcome 값 null · runtime UNOBSERVED」** 로 정정.
  핀: shadowOutcome≠null ⇔ runtime=SHADOW.
- 내부 계약 판단 기록에 `additionalProperties:false` 추가: 외부 소비자 등록부 없음 · 저장소 안 읽기는 Validate 경유 · RPC 디코더는 모르는 필드 허용 · 엔진/콘솔 같은 이미지.

## 10. 재시작 핀 셋 (판정 9 · 시나리오 :91-94 · 재검 확정 6)

**공통 전제(빈 표본 금지 — 보이스 3 재검 P1-2):** 앞 프로세스가 **전체 주기**(`runProductionStrategyMarketCycle` + cycle 클로저의 shadow 단계)에서 WOULD_EMIT ≥ 1 을 관측했음을 단언하고,
재시작한 프로세스가 **전체 주기를 1 회 이상** 돈 뒤 dispatch 궤적 · 원장을 잰다.
① 핀 없음 → 여덟 OFF/OFF/UNOBSERVED · dispatch/activation write 0(shadow 기여분 0 — 차등). ② 핀 있음 + 못 씀(파일 없음 · 불일치 · 만료 · 폐기 · 결속 불일치 — 각 모양) → 같음.
③ 유효 핀 잔존 → 재시작 뒤 **첫 물결 전** 여덟 UNOBSERVED · SHADOW 는 재독(첫 물결)에서만 · 이전 프로세스의 wave/outcome 0(관측 메모리 비복원). 세 개 모두 R2 재시작 시험과 같은 파일에 짝.

## 11. census 품질 · 활성화 쓰기 0 (판정 10)

- 어휘 census: 타입 보유 const + 변환식 전수(§7). 호출 폐포: 전이 + 하한 단언(§2 ③). 「활성화 쓰기 0」: 설정 디렉터리 **목록과 각 파일 바이트** 의 전후 동일(4.5-e 시험 모양) —
  shadow 단계 실행 전후.

## 12. 8.6 (판정 ⑥ — 변경 없음)

SHADOW 는 능력이지 배포가 아니다. 생산 shadow 핀 0 → shadow 단계는 파도마다 미선언 판정만(파일 I/O 0 — 적재기 첫 줄이 핀 비었음). 8.6 BLOCKED 유지, :295 의 shadow 수집은
배포 뒤(사람 · A100).

## 13. 로트 계획

1. re-freeze: 같은 4판이 **자기 P1 종결만 표적 재검**(Manager 판정 12) — 재검 전원 PASS 뒤 진행. 4라운드(v3.1): codex · 보이스 1 · 보이스 3 표적, 보이스 2 PASS 유지(노트 2 는 §5
   스탬프로 반영). §5 핀 ① 은 Manager 판정 (A)(2026-10-05). 5라운드(v3.2): codex **clean run**(4라운드 자가 무효) · 보이스 1 · 보이스 3 표적.
2. Pre-Edit FLM(§4 목록 + 활성화 적재기 무편집 확인용 AST 기준) → RED(§2 · §3 · §5 · §6 · §8 · §9 · §10 핀 전부, 편집 전 실패 기록).
3. GREEN: `internal/strategyshadow` · worker `ShadowEligible`/`ShadowVerdict` · 운반 타입 · shadow 단계 · 투영 · validateLane · OpenAPI · 생성기 · 골든 · 문서 · ROADMAP 행 정정.
4. 변이(§2 census 각 항 · 교차 디코드 · sentinel 배타 · 술어 공유 · fault 격리 · 차등 dispatch · 재시작 셋) · race 목록 · verify · 착지 → 재고정 영수증 최종 갱신.
