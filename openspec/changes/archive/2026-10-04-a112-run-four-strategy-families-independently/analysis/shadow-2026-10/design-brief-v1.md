# 7.3.1 SHADOW 설계 브리프 (코드 0 — 2026-10-04, 사용자 결정 「선택지 1: a112 안에서 구축」)

정본 계약: `specs/four-family-strategy-runtime/spec.md` :88-93 — SHALL 하나(새 설치 · migration · 재시작 = OFF/OFF/UNOBSERVED), 허용 절(server-owned **signed** shadow
manifest 가 있을 때만 pure evaluation · counterfactual projection), MUST NOT 둘(desired/effective/activation ON · dispatch capability 소유 / process-local shadow 상태의
재시작 자동 복구), 시나리오 :91-93(앞 프로세스가 SHADOW 관측 → 매니페스트 없이 재시작 → OFF/OFF/UNOBSERVED · dispatch/activation write 0). design :246(같은 문장),
:289(「8 FamilyWorker/2 MarketCoordinator 를 dormant/shadow 로 배선」), :291(배포 뒤 「shadow 관측과 counterfactual 결과만 수집」).

## 0. 지금 코드에서 측정한 것(설계가 읽을 값 · 경로)

- **평가는 이미 활성화와 무관하게 돈다.** 시장 주기의 제안 수집(`strategyProposalAuthorityLoader.collectMarket`)은 경로에 오른 범위의 순수 제안을 활성화 여부와
  상관없이 적재한다(미선언 시장 = 기존 경로도 제안을 적재). 레인(`strategyworker.FamilyWorker.Run`)은 평가자가 아니라 **관문**이다: `Effective(activation) != ON` 이면
  DORMANT, 아니면 봉인된 제안이 제 레인인지(`owns`) 보고 coordinator 봉투(`strategycoordinator.Envelope`)를 낸다. → SHADOW 의 「pure evaluation」 은 새 평가가 아니라
  **이미 있는 제안 위에서 레인 관문을 반사실로 도는 것**이다(새 원격 I/O 0).
- runtime 어휘: router `RuntimeState` · projection `LaneRuntime` 이 정확히 {UNOBSERVED}(R2 핀 `TestTheRuntimeVocabularyIsExactlyUnobservedUntilAShadowLotExtendsIt`).
  worker 의 `runtime` 은 생성 시 고정 필드(`productionWorker` → UNOBSERVED) — 주기마다 변하는 값이 아니다.
- **골든 four-family-runtime-v1.json 은 runtime 어휘를 열거하지 않는다** — 서술자별 기본값(`runtime: "UNOBSERVED"`)만 얼린다. SHADOW 는 기본값이 아니므로
  **골든 내용 개정이 필요 없다**(기본값 불변이 spec SHALL 그 자체). 대신 **OpenAPI 계약이 열거한다**: `docs/api/openapi-v1.json`
  `components/schemas/StrategyRuntimeLaneRuntime/properties/runtime.enum = ["UNOBSERVED"]` → SHADOW 를 투영하면 enum 확장(계약 변경 — 아래 §4).
- 형제 매니페스트(활성화)의 신뢰 앵커는 결정 61 로 ed25519 를 뺀 **외부 digest 핀**(`TOSSOS_STRATEGY_FAMILY_ACTIVATION_<M>_MANIFEST_SHA256`) + 0400 · 소유자 · 정규 바이트
  등식 · 24h 수명 · 폐기 · 다섯 결속, 생성기 `tools/a112-family-activation`, 커밋 골든 + 핀 시험, 문서 `docs/operations.md` :277-320. 생산 핀 0.

## 1. 신뢰 앵커 — 서명 vs digest 핀 (핵심 축)

| | (S1) spec 문면 그대로 — ed25519 서명 | (S2) 결정 61 과 같은 digest 핀 |
|---|---|---|
| 막는 것 | 파일 치환(서명 검증) + 승인자/배포자 분리(개인키가 호스트 **밖**일 때만) | 파일 치환(핀 ≠ 바이트면 거절) — 서명과 독립적으로 이미 막는다(결정 61 근거 (b)) |
| 이 배포에서 서명이 더 사는 것 | 없음 — 사람이 키를 호스트 밖에 둘 생각이 없다고 답했다(결정 61). 키가 config 옆이면 env 를 쓸 수 있는 것은 무엇이든 파일 · 핀을 함께 바꾼다 | — |
| 위험 비례 | **역전**: 읽기 전용(노출 0)인 SHADOW 가 노출을 여는 활성화보다 강한 앵커를 갖게 된다 | 형제와 같다. SHADOW 위조의 최악 = 투영 소음(주문 · lease · 활성화 0 — §3 핀) |
| 비용 | 키 생성 · 보관 · 교체 · 분실 절차, 서명 생성기, 서명 골든(테스트 키), 검증 코드 + 변이 | 활성화 적재기 · 생성기 · 골든 패턴 재사용(코드 대부분이 형태 복제) |
| spec 정합 | 문면 그대로 | **spec 문구 수정 필요** — :89 「server-owned signed shadow manifest」 → 「server-owned, deployment-pinned(digest) shadow manifest」 류. design :238 이 「spec 의 signed 는 SHADOW 에만 붙는다 — 이 개정 범위 밖」 이라 적었으므로 그 줄도 함께 |

**권고: (S2) + spec · design amendment(Manager 작성).** 근거: 서명이 이 배포에서 사는 성질이 0 이고(결정 61 과 같은 사실), 서명을 고르면 위험이 낮은 쪽이 더 강한
앵커를 갖는 역전이 생긴다. **정지-보고 항목 ①:** (S2) 는 spec 문구 수정이 전제다 — amendment 없이 구현하면 spec/코드 불일치. 사람이 「SHADOW 만은 서명」 을 원하면 (S1)
비용(키 절차)을 사용자에게 다시 물어야 한다.

## 2. 매니페스트 형식 · 스키마 · 생성기 · 골든 · 문서 (8.5 선례: 넷 중 하나라도 없으면 「배포 불가능」)

선택지: (F1) **별도 파일** `strategy-family-shadow-<MARKET>.json` + 별도 핀 `TOSSOS_STRATEGY_FAMILY_SHADOW_<MARKET>_MANIFEST_SHA256` vs (F2) 활성화 매니페스트에 서술자별
`runtime: SHADOW` 필드 추가.
- (F2) 는 활성화 정규 바이트 · 골든 · 핀을 바꾼다(이미 커밋된 활성화 골든 `TestTheCommittedGoldenManifestMatchesItsPinAndPromotesTheFourLanes` 재생성) 그리고 SHADOW 만 원하는
  배포가 「모두 OFF 인 활성화 매니페스트」 를 둬야 한다 — 미선언(토글 OFF = upstream, 결정 62)과 「선언됐는데 전부 OFF」 가 섞인다.
- **권고 (F1).** 형태는 활성화와 같은 정규 직렬화 등식 · 0400 · 소유자 · 크기 상한 · 수명 상한(24h) · 폐기 · 결속(경로 매니페스트 · 보정 · 달력 · 빌드 · 위험 정책 —
  활성화와 같은 다섯; ProtectionReady 는 노출이 없으므로 결속 안 함 — 정지-보고 항목 ② 로 Manager 확인) + 서술자 = 그 시장의 정확히 네 레인 각각 `shadow: ON|OFF`.
  desired/effective 필드 **자체가 없다**(값으로 ON 을 표현할 자리 0 — §3 ①).
- 스키마 상수 · 도메인(`tossos.strategy-family-shadow.v1` 류) · 생성기 `tools/a112-family-shadow`(활성화 생성기와 같은 플래그 모양, 핀 출력) · 커밋 골든(바이트 + 핀) ·
  골든을 **만드는** 함수를 부르는 시험(8.8.3 교훈) · `docs/operations.md` 절(파일 위치 · 핀 env · 생성 · 폐기 · 「SHADOW 는 노출을 열지 않는다」).
- 생산 작성자 0 핀: 활성화 인코더 가드(`TestOnlyTheAuthoringToolCanBuildActivationBytes`)와 같은 모양으로 shadow 인코더도 도구만 부를 수 있게.

## 3. 허용 = pure evaluation + counterfactual projection **만** — MUST NOT 둘의 집행 핀

설계: 레인에 **둘째 문** `FamilyWorker.Shadow(shadow strategyrouter.FamilyShadow, input Input) ShadowCycle` — `Run` 과 분리. `FamilyShadow` 는 적재기만 만드는 불투명
타입(필드 비공개, 영값 = 아무 레인도 shadow 아님), `ShadowCycle` 은 반사실 결과(`WOULD_EMIT` · `NOT_THIS_LANE` · `NO_INPUT` · `NOT_SHADOWED`)와 진단만 싣고
**coordinator 봉투 필드가 없다.** 활성화가 그 레인을 ON 으로 둔 경우 shadow 는 그 레인에 적용되지 않는다(ON 이 우선, SHADOW 는 OFF 레인만).

| MUST NOT / 성질 | 집행 핀(구조 + 행동, 변이로 잼) |
|---|---|
| desired/effective/activation ON 금지 | ① `FamilyShadow` 에 `DesiredState` 를 돌려주는 메서드 0 · `FamilyWorker.Desired/Effective` 는 `FamilyActivation` 만 받음(타입 census — strategyhandoff mint census 모양) ② 행동: 유효한 shadow 매니페스트 + 활성화 없음 → 여덟 레인 desired/effective OFF, `Run` 결과 DORMANT ③ 변이: Shadow 가 Effective 를 ON 으로 · shadow 적재기가 FamilyActivation 을 만듦 → RED |
| dispatch capability 소유 금지 | ① `ShadowCycle` 이 `strategycoordinator.Envelope` · handoff 타입을 **품지 않음**(타입 census) ② shadow 경로 함수의 호출 허용 목록(되돌림 경로 `TestTheRollbackPathOnlyReads` 모양) ③ 행동: shadow 매니페스트로 WOULD_EMIT 이 여럿이어도 coordinator Submit 0 · dispatchHandoffs 전달 0 · 게이트웨이 스파이 0 · 원장 lease/잠금/복구 행 0 ④ 변이: ShadowCycle 이 봉투를 실어 조정에 넣음 → RED |
| 활성화 쓰기 0 | shadow 적재기 · 투영은 읽기만(파일 바이트 전후 동일 — 4.5-e 시험 모양), 활성화 인코더 가드에 shadow 경로 추가 |
| 노출 0(spec 의 상위 SHALL) | 위 ③ 의 같은 시험 안에서 매수 브로커 0 |

counterfactual projection: `lanes[i].runtime = SHADOW` + 가산 필드 `shadowOutcome`(위 넷) · 선택: 코디네이터 반사실 선택(`coordinators[i].shadowSelected` — 순수 함수
`coordinateMarketProposals` 를 shadow 관문으로 한 번 더 돈 결과). **정지-보고 항목 ③:** 코디네이터 반사실까지 할지(spec 「counterfactual projection」 의 범위 — 레인 단위만으로
문면은 충족) Manager 판정.

## 4. 재시작 — process-local 자동 복구 금지

활성화 적재기 패턴 재사용: **매 파도마다 매니페스트를 다시 읽고**, 핀 없음 = 미선언(SHADOW 없음), 핀 있음 + 쓸 수 없음(없음 · 불일치 · 만료 · 폐기 · 결속) = SHADOW 없음
(되돌림 — 노출이 없으므로 「닫힘」 의 의미는 투영이 UNOBSERVED 로 돌아오는 것뿐). SHADOW 상태는 원장 · 파일 어디에도 쓰지 않는다(관측 저장은 레인 런타임 메모리 —
프로세스와 함께 사라짐). 핀 시나리오 :91-93: 앞 프로세스가 shadow 매니페스트 아래 SHADOW 관측 → 같은 원장 · **매니페스트 없이** 새 Context → 여덟 OFF/OFF/UNOBSERVED,
dispatch lease · 레인 잠금 · 복구 기록 0(R2 시험 `TestARestartAfterAnObservedPromotionComesBackOffOffUnobservedWithNothingWritten` 의 shadow 판 — 같은 파일에 짝).

## 5. R2 핀 · 계약 개정(로트 안에서 짝으로)

- `TestTheRuntimeVocabularyIsExactlyUnobservedUntilAShadowLotExtendsIt` → 정확히 {UNOBSERVED, SHADOW}(이름도 바꿈 — 「…ExactlyUnobservedAndShadow」). 이 로트가 그 flip 의 주인.
- 골든 four-family-runtime-v1.json: **내용 개정 불필요**(§0 — 기본값만 얼림, 어휘 열거 없음). 어휘를 골든에 선언하고 싶다면 그것은 Manager amendment(정지-보고 항목 ④ —
  권고: 하지 않음, 기본값 불변이 spec SHALL 이고 어휘는 코드 핀 + OpenAPI 가 진다).
- OpenAPI `StrategyRuntimeLaneRuntime.runtime.enum` 에 `SHADOW` 추가 + `shadowOutcome` 가산 필드 — 계약 변경(7.3 투영과 같은 규율: `internal/httpapi/a112_lane_children_openapi_test.go`
  류 대조 시험 갱신, older readers 는 가산 필드를 무시 — 단 **enum 값 확장은 엄격한 클라이언트에 비호환일 수 있음**: 정지-보고 항목 ⑤ — 콘솔 템플릿 · 외부 소비자 영향 확인).

## 6. design :289/:291 과 8.6 BLOCKED 의 정합

SHADOW 구축은 **능력** 이지 배포가 아니다. 생산 shadow 핀 0 → 오늘 동작 변화 0(활성화와 같은 「핀 없음 = 미선언」). :291 의 「dormant 배포 뒤 shadow 관측 수집」 은
여전히 8.6 의 if 갈래(A100 ProtectionReady + 모든 의존 게이트 + 사람 승인 배포) 뒤 — **8.6 은 다시 열리지 않는다**(BLOCKED 기록 유지). :289 「dormant/shadow 로 배선」
은 이 로트가 문면을 처음으로 참으로 만든다(지금까지는 dormant 만).

## 7. 로트 계획(승인 뒤)

1. Manager amendment(spec :89 문구 · design :238/:246 — S2 채택 시) → freeze 급 리뷰(보이스 + codex — 새 매니페스트 표면).
2. Pre-Edit FLM(편집될 기존 함수: collectMarket · strategyLaneRuntime.evaluate · strategyLaneProjection · familyGateFor 주변 · 투영 조립) → RED(§3 · §4 핀 전부, 편집 전 실패 기록).
3. GREEN: strategyrouter shadow 적재기(활성화 적재기 형태) · worker `Shadow` 문 · 레인 런타임 반사실 관측 · 투영 가산 필드 · OpenAPI · 생성기 · 골든 · 문서.
4. 변이(§3 표 전부 + 적재기 결속 항) · 엔진 race 목록 · verify · 착지 → 재고정 영수증 최종 갱신.

## 정지-보고 항목(Manager 판정 필요 — 코드 전)

① 신뢰 앵커 (S1)/(S2) — (S2) 면 spec :89 · :91 시나리오 제목 · design :238 · :246 amendment. ② shadow 매니페스트 결속 집합(활성화 다섯 그대로 · ProtectionReady 제외) 확인.
③ 코디네이터 반사실 선택까지 투영할지(레인 단위만이면 문면 충족). ④ 골든 어휘 선언 여부(권고: 안 함). ⑤ OpenAPI enum 확장의 소비자 영향(콘솔 · 외부) 확인 방식.

## 8. Manager 판정(2026-10-04) — 브리프 확정

① **(S2) digest 핀 채택** — spec :89/:91 · design :238/:246 amendment 는 Manager 작성(사용자 보고에 거부권 명시). **(F1) 별도 파일 + 전용 핀 env 승인** — 스키마는
`shadow: ON|OFF` 만, desired/effective 필드 없음(표현 불가 = 구조적 MUST NOT). 생성기 · 골든 · 골든을 만드는 함수를 부르는 시험 · operations 문서 · 생산 작성자 0 가드 전부.
② **결속 = 활성화 다섯, ProtectionReady 제외 확인.** 단서: shadow 적재기가 활성화 적재기의 사본이 되면 양쪽을 다 못 박아야 한다(옮겨 적은 코드 교훈) — **공유할 수 있는
것은 공유**(파일 읽기 · 정규 바이트 등식 · 0400/소유자 · 수명 · 폐기 · 결속 비교 — 활성화와 같은 함수), **갈리는 지점(결속 집합 · 서술자 스키마)은 타입으로 가른다.**
③ **레인 한정.** 조정자 반사실 선택은 **범위 밖**(두 번째 중재 경로 = 분기 위험; spec 문면은 레인 평가로 충족) — 이 좁힘을 tasks 7.3.1 노트에도 기록.
④ **골든 무변.** R2 핀 rename + 어휘 census = {UNOBSERVED, SHADOW}.
⑤ **OpenAPI: 저장소 안 소비자 감사로 닫는다**(외부 클라이언트 등록부는 저장소에 없음 → 내부 계약으로 취급, 그 판단을 기록) — 감사 결과 §9.
⑥ **8.6 불재개** 기록.

## 9. ⑤ 저장소 안 소비자 감사(실측, 2026-10-04 — 코드 0)

- **lanes[] 의 runtime 값을 직접 읽는 소비자: 0.** console(html/template) · cmd/tossctl · httpapi 어디에도 `LaneRuntime` 값 비교 · switch 없음(grep 전수 — console 의
  `UNOBSERVED` 는 고정 안내 문구 `templates_optimization.go:436` 하나).
- **그러나 의미 검증이 한 곳에서 정확 일치를 요구한다:** `strategyprojection.validateLane`(lanes.go:274-276) `lane.Runtime != LaneRuntimeUnobserved` → 거절. 그리고
  `strategyprojection.Validate` 는 **모든 읽기 경로의 관문**이다 — 엔진 저장소(store.go:18 · :43), Unix RPC 클라이언트(strategyprojectionrpc transport.go:97 ·
  transport_unix.go:142), httpapi(router.go:164 · strategy_runtime.go:40), console(settings_tabs.go:297 · strategy_runtime_multimarket.go:55). 즉 SHADOW 값을 실은 스냅숏은
  **검증기를 고치지 않으면 모든 화면에서 스냅숏 전체가 거절된다**(fail-closed — 노출은 없지만 투영 전체 소실). → 로트 안에서 `validateLane` 을 {UNOBSERVED, SHADOW} 로 넓히고
  **교차 규칙을 더한다: runtime=SHADOW ⇒ desired=OFF ∧ effective=OFF**(투영 층에서도 「SHADOW 는 ON 을 말할 수 없다」 를 집행 — 핀 + 변이).
- **판본 엇갈림(skew):** RPC 클라이언트 디코더는 모르는 필드를 **허용**한다(transport.go:77-90, 7.3 의 legacy reader 시나리오) → 가산 필드 `shadowOutcome` 은 구판 콘솔에서
  무시된다. 단 **구판 검증기는 SHADOW 값을 거절**한다 — 엔진만 새 판이고 콘솔이 구 판이면 그 기간 투영 소실. 이 저장소의 배포는 엔진 · 콘솔이 같은 이미지(`make image` 하나)라
  엇갈림 창은 교체 순간뿐이고, 생산 shadow 핀 0 이면 SHADOW 값 자체가 나오지 않는다. 기록하고 수용(외부 소비자 없음 — 내부 계약).
- OpenAPI `StrategyRuntimeLaneRuntime.runtime.enum` 확장 + `shadowOutcome` 가산 — `internal/httpapi/a112_lane_children_openapi_test.go` 류 대조 시험을 같은 로트에서 갱신.
