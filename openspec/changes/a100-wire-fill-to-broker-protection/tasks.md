# Tasks — a100-wire-fill-to-broker-protection

순서는 `design.md`의 Migration Plan을 따른다.

**0절이 끝나기 전에 1절을 시작하지 않고, 1절이 끝나기 전에 2절 이후를 시작하지 않는다.**
0절은 이 change의 전제가 참인지 묻고(선행 실측), 1절은 배선 대상 함수의 거부 분기 9개가 한 번도
실행된 적이 없다는 측정에 대응한다. 그대로 켜면 미검증 결함이 배선과 동시에 프로덕션으로 나간다.

## 2026-08-15 재동결 — 실행 로트와 독립 검토 계약

**다른 change로 옮기지 않는다.** 아래 미구현 범위는 모두
`a100-wire-fill-to-broker-protection` 안에서 수행한다. Manager(SOL)는 OpenSpec·분석 문서 작성과
최종 증거 대조만 소유하고, 프로덕션 코드와 테스트의 작성·수정·실행은 Terra 구현 에이전트가
소유한다. 각 구현 로트는 **서로 다른 Terra 적대 리뷰 에이전트**가 read-only로 검토하고,
P0/P1을 구현자가 고친 뒤 같은 리뷰어가 재검토해 `P0=0, P1=0`을 선언해야 다음 로트로 간다.

| 로트 | 포함 task | 구현 소유권 | 별도 적대 리뷰 | 종료 조건 |
| --- | --- | --- | --- | --- |
| R0 | 0.1, 0.4~0.7, 문서 모순 정정 | Manager 문서 + Terra 증거 수집 | A0 Terra | 현재 main base·FLM/BTM·strict validate 동결 |
| M0 | 0.2a | Terra-M0: `verify run --include-trigger` causal receipt 도구 | A-M0 Terra | lossless parent/child receipt·fsync barrier·INCONCLUSIVE gate GREEN |
| M-A | 0.2, 0.11 | **사람이 실행·승인하는 실계좌 측정** | A-M Terra read-only | 발동·child 체결·조회 귀속·raw status 표가 모두 PASS |
| T1 | 1.1~1.3 | Terra-1: `internal/protectionlifecycle` RED와 필요한 최소 core 수정 | A1 Terra | 9개 거부 분기 본문 실행 + 회귀 GREEN |
| T2-A | 2.1~2.6, 3.10 | Terra-2A: 도메인 매핑·journal schema·child ownership/apply | A2-A Terra | durable pending/causal child 귀속/선행-fill 거부 원자성 |
| T2-B | 4.0 | Terra-2B: `protectionlifecycle` pure API seal/allowlist | A2-B Terra | exact public surface + authority-minting 0; T3 compile seam 준비 |
| T3 | 3.1~3.8 | Terra-3: 독립 수렴 워커 core | A3 Terra | worker-only 실패 격리·수량/generation 재확인 |
| T4-A | 4.5~4.6 | Terra-4A: exit 상주주문 취소·flat/child 권위 경계 | A4-A Terra | 기존 손절 즉시성·child 귀속 보존 + double-sell 창 최소화 |
| T4-B | 3.9, 4.1~4.4 | Terra-4B: import 가드·gateway/runtime auxiliary 조립 | A4-B Terra | gate verified/recovery 이후 auxiliary 기동, `ProtectionWired` 불변 |
| T5 | 5.1~6.4 | Terra-5: official fixture·운영 표면·rollback 회귀/문서 증거 | A5 Terra | KR/US fixture·console/alert·rollback 계약 GREEN |
| G | 7.1~7.10 | Terra 검증 실행 + Manager 결과 대조 | gstack 독립 리뷰 | gate PASS, gstack P0/P1=0, Manager acceptance |

M0가 GREEN·A-M0 accepted되기 전에는 M-A의 place를 실행하지 않는다. M-A PASS와 0.11의
raw-status 판정표 동결 전에는 T1을 포함한 어떤 제품 구현 로트도 시작하지 않는다.
T2-B의 exact lifecycle API seal이 GREEN·A2-B accepted되기 전에는 T3를 시작하지 않는다.
T4-A의 child 귀속·완전청산 취소·flat 권위 경계가 GREEN·A4-A accepted되기 전에는 T4-B의
worker construction/start와 broker mutation 도달 경로를 열지 않는다. 번호 순서와 무관하게
실행 순서는 **4.5~4.6 → 3.9·4.1~4.4**다.
한 구현자가 여러 로트를 맡을 수는 있지만, 같은 로트의 적대 리뷰어를 겸할 수 없고 소유 파일을
겹쳐 편집하지 않는다. 각 로트의 리뷰·재검토·Manager acceptance는 8절에 기록한다.

## 0. 착수 전 조건

- [x] 0.1 `base-commit.txt`가 현재 작업 base와 일치하는지 확인한다. 병행 세션이 커밋을 쌓았으면
  `python3 tools/sdd/capture_change_base.py`로 재고정한 뒤 `make sdd-sync`를 다시 돌린다.
  — 2026-08-11 확인: base `eb41e19a`, HEAD `ce78b0db`. 사이의 커밋 3개는 **전부 a100 자신의
  문서 커밋**이고 Go 변경이 없다. 따라서 함수 귀속의 비교 기준은 여전히 유효하며 재고정하지
  않는다(재고정은 fingerprint 재sync만 더 들고 얻는 것이 없다). **병행 세션 커밋은 없었다.**
  — **2026-08-15 재개방:** 위 2026-08-11 판정은 더 이상 유효하지 않다. 현재 main은
  `882a0b49`이고 그 사이 `gateway.go`, `exitloop.go`, `journal/apply_hook.go`,
  `cmd/tossctl/engine.go`가 바뀌었다. `base-commit.txt`를 현재 main으로 재고정하고 R0 증거를
  다시 만든 뒤에만 이 항목을 완료한다.
  — **2026-08-15 완료:** base와 HEAD가 모두
  `882a0b490b0b6d2eb7abe5c5040c514776f49f3e`이고 current-main evidence를 재생성했다.
  `make sdd-sync`는 hard CodeGraph/CodeGraphContext를 최신으로 만들고 `all indexes current`로
  종료했다(GBrain busy는 advisory previous freshness 유지 경고).
- [x] 0.2a **M-A causal receipt measurement-only 도구를 먼저 만든다.** 2026-08-15 preflight에서
  current CLI/verify record는 strict local order를 증명하지 못했다. 외부 shell timestamp나 WTS
  `order show`로 대신하지 않는다. M0는 exact
  `verify run --include-trigger --confirm-each --resume --redo conditional-trigger`만 열고
  runtime·engine·protection·attestation·trading journal에는 연결하지 않으며 새 mutation method도 만들지
  않는다. existing owner-only verify record/status/abort는 cleanup identity authority다.
  - [x] 0.2a.1 편집 전 FLM/BTM을 `verifylive.Runner.stepConditionalTrigger`, `pollTrigger`,
    `readConditional`, `readOrder`, `finishTrigger`, `createConditional`, `readRetry`, `verifylive.New`,
    `cmd/tossctl.newVerifyRunCmd`, `runVerifyRun`, `official.Client.doRequest`·`Client.send`에 만든다.
    cleanup/abort/status의 기존 함수를 실제로 편집하면 그것도 먼저 추가한다. 새 private helper는 기존
    account/token/401-refresh semantics를 공유하고 기존 raw reader의 public 계약을 바꾸지 않는다.
  - [x] 0.2a.2 M0는 새 receipt 경로, `--confirm-each --resume --redo conditional-trigger`가 아니거나
    `--include-ttl-edge`, 다른 redo, trigger 외 미완료 mutating step이면 broker factory·confirmer·모든
    mutation 전에 거부한다. prior verify record의 `Outstanding(...) != 0`이면 cleanup prologue를 실행하지
    않고 같은 지점에서 HOLD한다. CLI와 `verifylive.New` direct caller가 같은 조합을 강제한다.
  - [x] 0.2a.3 receipt parent/final path는 current uid 소유·non-symlink exact 0700/0600이다.
    no-follow/O_EXCL로 만들고 versioned header+run ID를 file fsync한 뒤 parent directory도 fsync한다.
    이 준비 전 broker factory·confirmer·mutation은 모두 0건이다. fresh resume은 허용하지만 과거 receipt
    sequence는 재사용하지 않는다.
  - [x] 0.2a.4 official transport는 모든 HTTP attempt의 request-start와 body-read-complete를 같은
    process monotonic anchor에서 포착한다. numeric status/no-response class, 401 refresh attempt, 성공
    `SHA-256(raw-result-bytes-v1)`, non-2xx/invalid-envelope `SHA-256(raw-response-body-bytes-v1)`를 decode 전에
    writer에 전달한다. helper-return 뒤 `now()`는 response receipt가 아니다. account selection,
    token refresh, rate-budget, error classification이 기존 send와 동일하다는 RED를 둔다.
  - [x] 0.2a.5 extracted schema v1은 parent request/response conditional ID tag, client ID tag,
    symbol/market/type/order type/quantity/first side·trigger/expiry/root·leg status/triggered-child tag와 child
    request/response ID tag, requested market scope, raw symbol/side/status/quantity/filled/execution fields를
    보존한다. causal receipt에는 account/token/opaque ID 원문을 쓰지 않는다. PASS는
    `pending client tag == parent raw client tag`, pending approved parent fields == parent raw fields,
    `parent child tag == child checkpoint tag == child request tag == child response tag`, child
    SELL/market scope/symbol/quantity == approved parent leg를 각각 요구한다.
  - [x] 0.2a.6 human gate 승인 뒤 broker create **전에** unique client ID, run ID와 approved
    market/symbol/SINGLE/MARKET/SELL/qty=1/trigger/expiry를 owner-only verify record의 pending intent로
    append+fsync한다. 일반 verify step은 이 checkpoint를 만들지 않는다. create-response→exact-parent
    checkpoint 사이 crash의 다음 M0 resume은 safe receipt를 연 뒤 official all-page raw read로 pending
    client+fields를 대조한다. unique match면 exact parent checkpoint를 fsync하고 그 run은 HOLD로 끝낸다.
    zero/multiple/mismatch도 HOLD이며 자동 cancel/recreate하지 않는다.
  - [x] 0.2a.6b parent raw에서 child ID를 얻으면 owner-only exact child reconciliation checkpoint를
    먼저 append+fsync하고, 다음 sanitized parent causal receipt를 append+fsync한 뒤에만 child GET한다.
    어느 kill point도 과거 sequence를 재개하지 않는다. status/abort-list는 exact/pending 객체를 보이되
    triggered-but-child-unobserved 객체를 자동 취소하지 않는다.
  - [x] 0.2a.7 durable parent child-ID receipt부터 durable child first-observed-fill receipt까지가 strict
    critical window다. 기존 401 refresh/429 retry의 모든 attempt를 receipt화하고 첫 401/429/read/decode/
    identity/write/sync gap은 irreversible HOLD다. 이후 성공 retry가 지우지 못한다. window 전 오류도
    기록하지만 이 strict verdict의 자동 HOLD는 아니다.
  - [x] 0.2a.8 PASS는 `parent_child_id.seq < child_first_observed_fill.seq`와
    `parent_fsync_done < child_request_start`를 모두 요구한다. durable child fill로 window가 끝난 뒤 parent
    terminal GET은 요구하지 않는다. 그 전 parent 404는 HOLD다. server/wall time은 대체 증거가 아니다.
  - [x] 0.2a.9 RED/GREEN은 wall rollback/same-time, every-attempt 401/429, helper-return delay, raw
    identity/decimal mismatch, symlink/owner/mode/O_EXCL/file+dir fsync, pending-before-create,
    create-response-before-parent-ID, parent-raw-before-child-ID, child-ID-before-parent-causal,
    parent-causal-before-child-GET kill points, outstanding cleanup-prologue refusal, zero/one/multiple recovery,
    restart/no sequence merge, terminal boundary, forbidden flags, pre-mutation refusal, static M0 import isolation,
    `MutationMethods()` exact 불변을 포함한다. Terra-M0 뒤 별도 A-M0 `P0=0/P1=0`이어야 0.2를 연다.
  — **2026-08-15 Manager acceptance:** same-instance `*official.Client`가 mutation과 parent/child raw
  result+attempt를 함께 소유하고 arbitrary/split Broker는 거부된다. receipt는 모든 path component를
  dirfd/no-follow로 열며 active exclusive lease의 typed writer만 허용한다. write/fsync 실패는 영구 poison,
  pending-create만 read-only recovery 진입, parent/child owner는 manual HOLD다. core A-M0와 receipt
  A-M0-R은 각각 최종 **P0=0/P1=0**으로 ACCEPT했고 normal/race/vet/strict/FLM 검증도 PASS했다.
- [ ] 0.2 **선행 실측 M-A — 조건주문이 실제로 발동해 체결되는가.** `measurement-prereq.md`의
  절차를 따른다. **실계좌 주문이므로 실행 직전에 사람이 승인한다**(§0-1, §0-7).
  **완전 통과가 아니면 여기서 멈춘다** — 발동하지 않는 상주 주문은 보호가 아니며,
  child id가 fill보다 늦거나 조회 불가능한 부분 통과도 3.10의 causal ownership을 증명하지 못한다.
  trigger·수량 비교 실패도 수렴 정의를 무너뜨린다. 이 세 경우 모두 설계·spec을 다시 리뷰하기
  전에는 1절로 가지 않는다(design D10).
  — **2026-08-15 10:27 KST fresh preflight:** installed M0 identity와 official OPEN 전 페이지
  (`PAUSED` 2, `WATCHING` 3)는 확인했지만 KR 폐장, 기존 5건의 사람별 retain/cancel 결정 부재,
  official sellable 불완전·미관리 whole-share 후보 부재, soak/attestation unready·binary-unbound 때문에
  **HOLD/NO-GO**다. corrected receipt `/tmp/a100-ma-preflight-corrected.kSn14D`는 같은 Terra
  adversary가 evidence `P0=0/P1=0`으로 수용했지만 live 실행을 수용한 것은 아니다. preview·주문·
  config·verify run/resume는 0건이며 이 체크박스와 0.11, T1은 계속 미완료다.
- [x] 0.3 **선행 실측 M-B — 엔진 사망부터 exit observer 재무장까지의 무보호 창.** 이 change가
  사려는 안전의 크기다. 측정하지 않으면 개선 폭을 주장할 수 없다.
  — 2026-08-11 14:00Z, 0.10b의 배포 재시작에서 측정. **≤ 6.7초**(명령 발행 → 첫 `exit.*` 관측).
  기록은 `measurement-prereq.md`. **다만 이 값은 하한이다** — a056의 8분은 *기동 실패* 경우이고
  이번 측정은 정상 기동 경우다. a100이 사는 안전은 6.7초가 아니라 그 두 값 사이다.
- [x] 0.4 **편집 전 Function Logic Map 대상 확정.** 기존 6건에 현재 main의 runtime/child 귀속
  경계 7건을 추가해 AST·FLM을 다시 만든다. 2026-08-11의 측정 수치와 줄 번호는 historical evidence다.
  - `internal/app/engine.buildGateway` — 수렴 워커 조립 지점. 분기 4개가 **전부 에러 검사**이고
    조건부 조립이 하나도 없다 ⇒ 워커는 무조건 생성하고 기동은 호출자가 한다.
  - `internal/journal.scanExitStateResult` — 보호 컬럼이 지나는 **단일 스캔 지점**. 분기 22개 중
    20개가 부패 판정이다. **보호 컬럼을 `v10Evidence`·`full`·평탄화 비교에 넣으면 멀쩡한 행이
    부패로 판정되어 exit 정책이 멈춘다.**
  - `internal/journal.Journal.OpenExitStates` — 워커의 대상 집합. 주석이 "deliberately not two
    functions"라고 못박았으므로 **별도 조회 함수를 만들지 않는다.**
  - `internal/app/engine.ExitObserver.record` — a087 청산 경로의 취소 판정 지점.
  - `internal/app/engine.ExitObserver.submit` — a087 매도 발행 지점.
  - `internal/protection.TestProtectionRemainsUnwired…` — 봉인 가드 본체.
  - `cmd/tossctl.engineRuntime` — interlock 이후 워커의 기동·취소·감독 지점.
  - `internal/protectionofficial.Gateway.adapt` — raw status·triggered child id 손실 지점.
  - `internal/journal.Journal.TrackedFillOrders` — unchanged fill detector가 child를 읽는 집합.
  - `internal/journal.confirmedFillOwners` — fill보다 엄격히 선행한 owner의 시간 권위.
  - `internal/journal.resolveFillOrigin` — ordinary/protection owner의 exact conflict 판정 지점.
  - `internal/app/engine.Runtime.runAuxiliary`(+ panic 경계 `runAuxiliaryBody`) — worker stop이 기존
    안전 loop를 내리지 않으면서 전용 event를 내는 authority.
  - `protectionlifecycle_test.TestProductionAPIExportsNoAuthorityMintingFunction` — T2-B exact public
    allowlist로 바꿀 기존 봉인 가드.
  — **2026-08-15 완료:** 13개 current 경계의 AST·FLM·BTM·risk bundle과 historical
  `Runner.RunCycle` 증거를 `analysis/function-logic/`에 동결했다.
- [x] 0.5 현재 main에서 재생성한 산출물의 source SHA-256 일치를 `check_analysis.py`로 확인한다.
  — **2026-08-15 완료:** `check_analysis.py` PASS, 발견된 current `ast.json` 17개의 source SHA
  직접 대조 17/17 일치, strict OpenSpec validate와 `git diff --check` PASS.
- [x] 0.6 `strategyDispatchCycle.dispatch` D7 면제 유효. 편집하지 않고, 분기 수치는 2차 개정에서
  삭제했으므로 내부 분기를 근거로 인용하지도 않는다.
- [x] 0.7 High-risk Pre-Edit 선언을 `pre-edit-gate.md`에 현재 main 기준으로 다시 남긴다.
  current 구현 경계 13개(기존 6개 + runtime/child/API 7개)의 조건과 citation-only 면제를
  구분하고, 완료된 0.10 probe의 historical `Runner.RunCycle` 선언은 별도 증거로 유지한다.
  각 조건이 그 편집의 통과 요건이다.
  「실패 테스트 선행 작성」은 각 task
  착수 시점에 갱신한다.
  — **2026-08-15 완료:** sections 1~6과 8~14에 current 경계 조건을 기록했고, section 7의
  완료된 probe 선언을 historical evidence로 유지했다. 독립 A0 재검토 `P0=0/P1=0/P2=0`.
- [x] 0.8 **ratchet 수준이 스칼라로 영속되는지 확인한다.** — **영속된다. D9의 전제가 틀렸었다.**
  `exit_states`는 `initial_stop`과 별도로 `baseline_price`를 갖고 exit 판정마다 갱신한다
  (`journal/exit_state.go:563`). t0 값은 `InitialStop`이고(`exitpolicy/ratchet.go:321-332`),
  `RunnerTrailPct`는 별도 경로가 아니라 `CandidateHighWaterRunner` 후보로 합성되어
  `baseline_price`에 들어간다(`ladder.go:391-411`). 단조 비감소는 이미 불변식이다
  (`candidate.go:86-88`, `apply_hook.go:559`).
  ⇒ **trigger는 `baseline_price`에서 유도한다.** D9·spec 재작성 완료.
  ⇒ **파생 비용: trigger가 움직이므로 교체가 자동 경로에 상시 존재한다**(0.10에 반영).

- [x] 0.9 **a100의 도달 범위를 확인하고 기록한다.** — 2026-08-11 실측:
  `adoption.enabled = true`, `default_stop_pct = 0.03`,
  `include_symbols = [272210, 333430, IONQ, TSLA]` — **제외 목록이 아니라 포함 목록이므로
  a100의 대상은 정확히 이 4종목이다.** `automation_gate.enabled = true`.
  capability attestation은 2026-07-30 발급 / **2026-08-29 만료**.
  ⇒ 포함 목록 밖 보유는 이 change의 대상이 아니다.
  ⇒ **333430에는 2026-07-31 verify 실행이 등록한 조건주문
  `DJwYn8P_dD9lQMEeYc-5l5_yDTzu2fo55FEJkhU8WVg`가 있고, 그것을 취소할 후속 단계는 실행되지
  않았다**(`conditional-persist` = `awaiting-restart`). 수렴 워커가 자기가 만들지 않은 상주
  주문을 만나는 첫 사례이므로 `discoverOrphan` 경로의 실제 입력이다. **착수 전에 현재 상태를
  조회해 기록한다.**
- [x] 0.10 **`RequiredEndpoints()` 갱신과 attestation 재발급의 선후를 확인한다(D11).** — 확인 결과
  **a071의 Ed25519 서명은 무관하고**(그것은 protection capability attestation이다) C4는 a100을
  막지 않는다. 대신 재발급 자체가 비싸다. 2026-08-11 실측:
  - soak 기록의 최신 cycle이 **2026-08-05**이고 `MaxRecordAge`는 48시간이므로
    (`soak/attest.go:57`, `:95-100`) **오늘 `soak attest`는 거부한다.**
  - verify 기록(2026-07-31)은 `POST /conditional-orders`, `GET /conditional-orders`,
    `GET /conditional-orders/{id}`, `GET /sellable-quantity`를 이미 증명한다.
    **`DELETE /conditional-orders/{id}`와 `POST /conditional-orders/{id}/modify`는 없다.**
    기존 증거는 **2026-08-30에 만료된다**(30일 validity).
  - read는 supervised에서 조달할 수 없다(`soak/attest.go:255-277`). 수렴 워커가 매 주기 부르는
    `GET /conditional-orders`는 자동 경로의 상시 read이므로 **read-only soak에 probe를 추가하고
    신선한 성공 cycle을 적어도 한 번 기록해야 한다.**
  ⇒ 배포 순서 고정: **(a) soak probe 추가 → (b) soak 재실행(성공 cycle 1회, 기록 신선)
  → (c) 남은 verify 단계 실행(DELETE·modify, 사람 승인) → (d) `tossctl soak attest` 재발급
  → (e) 새 바이너리 배포.** (a)만 배포하고 (d)를 건너뛰면 두 시장의 loop가 전부 거부된다.

  **정정(2026-08-11, 0.10 (a) 착수 시 실측).** 위 (a)의 원문은 세 곳이 틀렸다.
  - **read는 하나가 아니라 셋이다.** `protectionofficial.Gateway`가 자동 경로에서 부르는
    read는 `ProtectionConditionalOrdersRaw`(`GET /api/v1/conditional-orders`),
    `ConditionalOrderRaw`(`GET /api/v1/conditional-orders/{id}`),
    `SellableQuantityRaw`(`GET /api/v1/sellable-quantity`)다
    (`gateway.go:103,122,173,193,221,246`). 하나만 probe하면 나머지 둘 때문에 배포를 다시 한다.
  - **(a)에서 `RequiredEndpoints()`를 확장하면 안 된다.** `soak.RequiredEndpoints()`는
    `Evaluate`의 거부 기준이고(`attest.go:130`), `engine.RequiredEndpoints()`는 기동 기준이다
    (`interlock.go:518`). 확장은 **거부를 먼저 오게 하고 증거를 나중에 오게 한다.**
    probe만으로 attestation에 실린다 — `BuildAttestation`이 싣는 것은
    `Window.SuccessfulEndpoints()`이고 required 목록이 아니다(`attest.go:183`).
    세 목록의 확장은 a100 본체가 (e)에서 한 번에 한다
    (`TestSoakAndLiveEndpointsCoverTheEngineInterlock`이 셋의 정합을 강제하므로 같이 움직인다).
  - **(d)는 a100 바이너리로 돌려야 한다.** `supervisedProofs`가 `LiveOnlyEndpoints()`로
    미리 거르므로(`cmd/tossctl/soak.go:456-465`), 그 목록이 조건주문 mutation을 모르는
    바이너리로 attest하면 **M-A가 만든 증거가 조용히 버려진다.**
  - **probe는 자체 3일 시계를 갖지 않는다.** `Window`는 streak 창이므로
    (`summary.go:107,162`) 창 안에서 1회 성공하면 증명된다. 3일은 credential streak의 몫이고
    그것은 0.12로 이미 돌고 있다.
- [x] 0.10a **조건주문 read probe 구현.** `internal/soak`에 endpoint 3개와 probe 3개를 추가하고
  `cmd/tossctl/soak.go`의 어댑터에 배선했다. 편집 전 FLM은
  `analysis/function-logic/internal-soak--runner.runcycle/`이고 Pre-Edit 선언은
  `pre-edit-gate.md` 7번이다. 통과 조건 4개(credential 격리·completeness 격리·순서·거부 목록
  불변)를 각각 RED 테스트로 고정했다(`internal/soak/protection_probe_test.go`).
  어댑터를 `cmd/tossctl/soak.go` **안에** 둔 이유: `static_test.go:136`이 그 파일 하나만
  검사하므로 새 파일로 빼면 조건주문 어댑터가 가드 밖으로 나간다.
  단위 테스트는 stub을 통과하므로 어댑터 경로를 증명하지 않는다 ⇒ 실제 HTTP로 4건을 더 고정했다
  (`cmd/tossctl/soak_protection_test.go`): 세 endpoint OK, OPEN·CLOSED 두 그룹 순회,
  by-id가 목록이 준 식별자를 사용, **조건주문 경로가 열려도 GET 외 요청 없음**(read 경로가
  create·modify·cancel과 같은 경로를 공유하므로 read-only를 다시 세워야 한다).
  그리고 이 단계 전체가 서 있는 성질 —「probe만으로 attestation에 실린다」— 를
  `TestAProbedEndpointReachesTheAttestationWithoutBeingRequired`로 고정했다.
- [x] 0.10b **배포 완료 (2026-08-11 14:00Z).** 이미지 `sha256:86c6e4d2…`, 두 컨테이너 healthy.
  **KST 05:00~09:00 창이 아니라 미국 정규장 중에 수행했다** — 위험(US 3종목 무보호)을 통보한 뒤
  사용자가 즉시 진행을 선택했다. 결과 창은 6.7초였다(0.3).
  **그리고 배포가 결함을 하나 드러냈다: 컨테이너 재생성이 capability soak를 죽인다.**
  `engine.autostart = true`는 있지만 **soak에는 자동 기동이 없고**(`deploy/container-entrypoint.sh`에
  `soak` 문자열이 없다) 재생성 후 살아난 것은 engine·console·httpapi뿐이다.
  ⇒ **배포할 때마다 attestation 시계가 조용히 멈춘다.** attestation은 2026-08-29에 만료되므로
  이것은 a100과 무관하게 운영 결함이다. 6.5와 같은 자리에 정리 change로 등록한다.
- [x] 0.10b-1 **soak 재시작 완료 (2026-08-12 08:44:51 KST).** 사용자가 콘솔의 soak restart를
  눌렀다. 에이전트의 `docker exec` 시도는 차단됐고 우회하지 않았다 — 지원 경로가 실제로
  지원 경로였다. audit에 `soak.autostart` 두 줄이 남았다(`old:false→true`, 이어서
  `old:true→true`): 버튼이 3.6초 간격으로 두 번 눌렸고, 두 번째가 첫 서베이를 정상 종료시킨 뒤
  새로 세웠다. 그것이 버튼의 설계된 동작이다(`restartSoak`).
- [x] 0.10c **세 endpoint 실측 완료 (cycle 2026-08-11T23:44:54Z → 23:46:00Z, `endpoints 9/9`).**

  | endpoint | 결과 | 요청 | latency |
  |---|---|---|---|
  | `GET /api/v1/conditional-orders` | **ok** | **2** | 230ms |
  | `GET /api/v1/conditional-orders/{id}` | **ok** | 1 | 315ms |
  | `GET /api/v1/sellable-quantity` | **ok** | 1 | 134ms |

  세 가지가 확인됐다.
  ① **미지수가 풀렸다** — 미보유 종목 `005930`에 대해 `GET /sellable-quantity`가 **답한다.**
  보유 종목으로 재시작할 필요가 없다.
  ② `conditional-orders`의 **요청 2건**은 OPEN·CLOSED 두 그룹을 각각 세는 정정이 기록에
  그대로 나타난 것이다(정정 전에는 1로 기록되어 rate budget을 절반으로 과소보고했다).
  ③ `{id}`가 ok인 것은 **333430 고아 조건주문이 아직 살아 있다**는 뜻이다 — 그것이 이
  endpoint를 증명 가능하게 만드는 유일한 대상이므로, M-A가 그것을 소진하기 전에 probe를
  배포해야 한다던 판단(review.md S6)이 맞았고 그 순서를 실제로 지켰다.

  **직전 사이클은 실패했다(같은 기록에 남아 있다).** 두 번째 버튼 누름이 3.6초 만에 첫 서베이를
  끊었고, 그 서베이는 `credential rate_limited`로 사이클을 닫았다. 그 실패 기록도 세 endpoint를
  포함하며 `{id}`의 skip 사유가 **"the conditional-order list could not be read, so no id was
  available"**로 나온다 — 「계정에 조건주문이 없다」와 「목록을 못 읽었다」를 구분하도록 고친
  사유 문구가 실운영에서 그대로 확인됐다.

  성공 사이클 안에서도 `GET /api/v1/orders`는 429를 네 번 맞고 재시도로 통과했다(30요청·64초).
  M8의 penalty window가 여전히 존재하며 **probe 셋을 사이클 맨 끝에 둔 판단의 근거가 그대로
  유효하다** — 429를 맞은 것은 앞의 order walk이고 뒤의 세 probe는 전부 ok다.
- [x] 0.12 **soak 재가동 완료 (2026-08-12 08:44 KST).** attestation은 2026-08-29에 만료되고
  soak는 2026-08-05 이후 돌지 않았다. 그때까지 새 기록이 없으면 automation gate를 켠 엔진은
  **a100이 없어도** 뜨지 않는다. 0.10b-1이 이것을 함께 해소했다 — 성공 사이클 하나가
  기록됐고 15분 간격으로 계속 돈다. **다음 배포에서 자동으로 살아나는지가 a101의 남은 검증이다**
  (a101 tasks 5.4).
- [ ] 0.11 **raw conditional status 보존과 판정표를 M-A 결과로 동결한다(D2).** raw status를
  도메인과 journal에 싣는 것은 선택이 아니라 필수다. `PAUSED`의 부재 관측은 무장 증명이 아니며,
  `WATCHING/PAUSED/ORDERING/ORDERED`를 같은 값으로 접는 현 어댑터는 수렴 증거가 될 수 없다.
  M-A가 실제로 관측한 pre-trigger/triggering/terminal 문자열을 표로 남기고 다음을 지킨다.
  - `PAUSED`와 unknown은 **미수렴·mutation 금지·operator alert**다.
  - child 귀속이 끝나지 않은 triggering 상태는 ACTIVE도 terminal도 아니며 재등록하지 않는다.
  - M-A가 armed 상태와 child ID/체결의 조회 가능성을 증명하지 못하면 1절로 가지 않는다.

## 1. 거부 경로 RED 테스트 — 배선보다 먼저 (Migration 1)

이 절은 **프로덕션 조립을 바꾸지 않는다.** 순수 core에 대한 테스트만 추가한다.
각 항목은 RED(실패하는 테스트) → GREEN 순서로 진행하고, GREEN이 core 로직 수정을 요구하면
그 수정 자체가 이 change가 찾아낸 결함이므로 별도로 기록한다.

### 1.1 `protectionlifecycle.applyFill` — 미실행 5개

- [ ] 1.1.1 **B1** position 취득 실패(비-`InvalidObservation`) — 알 수 없는 포지션 키에 대한 체결이
  상태를 만들지 않고 typed refusal로 끝난다.
- [ ] 1.1.2 **B2** state seal 무효 — 봉인이 깨진 상태를 입력하면 전이가 거부되고 어떤 필드도
  복구·재봉인되지 않는다.
- [ ] 1.1.3 **B3** fill 식별자 무효 / broker order id 불일치 — **잘못된 체결이 남의 포지션에
  귀속되지 않는 유일한 방어선.** 다른 포지션의 broker order id를 실은 체결이 거부된다.
- [ ] 1.1.4 **B6** 수량 0 또는 claim 초과 — 보호 수량이 보유를 넘는 전이가 거부된다.
- [ ] 1.1.5 **B7** 잔량 0 → `Terminal` 전이 — **보호주문이 다 채워졌을 때 상태가 닫히는 경로.**
  종료 후 추가 전이가 멱등하게 거부된다.

### 1.2 `protectionlifecycle.prepareRegister` — 미실행 4개

- [ ] 1.2.1 **B2** entry latched / phase 부적합 — 진입이 닫힌 상태의 등록 시도가 거부된다.
- [ ] 1.2.2 **B3** 이미 pending인 operation — **중복 제출 방지.** 재시도가 두 번째 제출을 만들지 않는다.
- [ ] 1.2.3 **B4** 보호가 이미 active — 같은 포지션에 두 번째 보호주문이 나가지 않는다.
- [ ] 1.2.4 **B5** 브로커가 정확한 operation 조회를 못 함 — capability 부재가 실제로 판정을 막는다.

### 1.3 측정

- [ ] 1.3.1 `go test -covermode=set -coverprofile`로 **true 결과 본문 행**의 실행 여부를 다시
  측정한다. 조건 statement의 covered는 근거가 아니다.
- [ ] 1.3.2 두 `branch-test-map.md`를 측정값으로 갱신한다. 9개 모두 `yes`가 아니면 2절로 가지 않는다.

## 2. 도메인 매핑과 journal 스키마 (Migration 2)

- [ ] 2.1 `protectionlifecycle` ↔ `internal/protection` 도메인 타입 매핑을 순수 함수로 만든다
  (`Scope`, `ConditionalBody`, `BrokerTarget`, `BrokerProtection`).
- [ ] 2.2 왕복(round-trip) property 테스트로 매핑이 정보를 잃지 않음을 증명한다. 손실이 있으면
  lifecycle이 증명한 불변식이 전송 단계에서 깨진다.
- [ ] 2.3 journal에 보호 컬럼을 **additive-nullable로만** 추가한다. 기존 컬럼의 의미를 바꾸지 않는다.
  최소한 parent broker order id, client order id, raw status, triggered child order id, lifecycle 상태,
  position generation, desired가 생긴 시각(미설치 시간 측정용), child apply watermark를 담는다.
- [ ] 2.4 schema 호환 회귀를 두 방향으로 고정한다.
  - 새 바이너리가 **이전 schema의 기존 행**을 열면 nullable 보호 값은 「보호 미설치」로 읽고
    기존 체결 감지·대사·reduce-only 청산이 그대로 동작한다.
  - 보호 컬럼을 포함한 **더 새로운 schema를 이전 바이너리로 열면** `ErrSchemaTooNew`로 엔진을
    거부한다. 구 바이너리가 같은 journal로 계속 동작한다는 기대를 테스트해서는 안 된다(D4).
- [ ] 2.5 `SchemaVersion` 증가와 main 대조. 낮은 SchemaVersion 바이너리가 뜨면 콘솔만 뜨고 엔진이
  조용히 죽는다(2026-08-04 실발생). 배포 전 대조를 7절에 건다.
- [ ] 2.6 **전송 전 pending 커밋.** `recoverSubmit`은 영속된 pending command가 없으면
  "no exact submit pending"으로 시장을 latch한다(`protectionlifecycle/lifecycle.go:63`, `:71`).
  저장 순서를 「plan → **pending 커밋** → 전송 → 응답 커밋」으로 고정하고, 전송 직후 프로세스가
  죽는 시나리오를 테스트한다. 2.3의 컬럼 목록이 이 레코드를 포함해야 한다. conditional parent의
  canonical owner는 전송 전에 고정하며 응답의 opaque id로 scope를 바꾸지 않는다.

## 3. 수렴 워커 (Migration 3)

이 절이 끝나도 **프로덕션 관찰 동작은 바뀌지 않는다.** 워커가 조립되지 않기 때문이다(조립은 4절).

- [ ] 3.1 수렴 워커 패키지를 만든다. `internal/filldetect`도 `ReconcileDriver`도 편집하지 않는다.
  자기 주기로 돌고 입력은 journal의 커밋된 상태뿐이다(D3).
- [ ] 3.2 desired/observed 비교를 D2의 표대로 구현한다. 각 행이 하나의 테스트를 갖는다.
- [ ] 3.3 **수렴 완료의 정의를 코드로 고정한다** — 보유 수량과 ACTIVE 보호 수량이 정확히 같고
  trigger가 journal 값에서 유도된 값과 같을 때만 수렴이다. "대략 맞음"을 허용하지 않는다.
  ACTIVE는 0.11의 evidence-backed armed raw status만 허용한다. `PAUSED`, unknown, triggering,
  child 미귀속은 수렴 완료가 아니다.
- [ ] 3.4 **보호 미설치 시간**을 포지션별로 측정하고 상한 초과 시 알림을 낸다(D2).
  상한이 없으면 「수렴 중」과 「영원히 실패 중」이 구별되지 않는다. 기본 상한은 90초,
  dirty/pending 재시도 기본 5초(지수 백오프, 최대 60초), 수렴한 ACTIVE 재확인 기본 60초로 두고
  config key는 각각 `engine.protection_convergence.unprotected_alert_seconds`,
  `dirty_retry_initial_seconds`, `retry_max_seconds`, `active_recheck_seconds`로 고정한다. block 부재는
  90/5/60/60을 적용하고, 명시된 0·음수·`dirty_retry_initial_seconds > retry_max_seconds`는
  block 전체를 거부해 worker를 기동하지 않는다. 정확한 경계 테스트를 둔다. 알림 identity는
  `(account, market, position, generation, cause)`이며 같은 episode에는 한 번만 발행하고 해결 후
  같은 cause가 재발하면 다시 발행한다.
- [ ] 3.5 실패 격리 — 수렴 실패가 typed reconcile reason과 알림으로 끝나고 다른 루프의
  outage·SLO·게이트를 건드리지 않음을 테스트한다. ordinary cycle failure는 worker `Run`의 반환이
  아니라 데이터이며 내부 backoff로 계속 돈다. panic/예상 밖 return은 A100 전용 stable
  `protection.convergence.worker_stopped` event와 durable reconcile/alert를 남기지만 reconcile·exit·
  filldetect를 취소하거나 runtime nonzero 종료를 만들지 않는다.
- [ ] 3.6 재시도와 백오프. **수렴한 포지션도 주기를 늘려 계속 재확인한다** — 상주 주문은
  나중에 취소·만료·일시정지될 수 있고 그때 보호 미설치 시간이 다시 시작되어야 한다.
  rate limit은 주기로 다루고 판정의 신선도를 버리지 않는다(리뷰 정정).
- [ ] 3.7 워커가 조립되지 않았을 때 관측 가능한 동작이 배선 이전과 완전히 동일함을 증명하는 테스트.
- [ ] 3.8 **대사와의 경합을 막는다.** 전송 **직전**에 포지션 수량·generation을 다시 읽어 계획
  시점과 다르면 전송하지 않고 그 주기를 버린다. 응답 후에도 다시 읽어 그 사이 수량이 바뀌었으면
  수렴 완료로 기록하지 않는다. 경합 원문: 워커가 10주를 읽는다 → 대사가 0주를 커밋한다
  (`reconcileloop.go:413`, `reconcile/converge.go:209`, `:244`) → 워커가 SELL 10을 등록한다.
- [ ] 3.9 **실행 주체와 기동 순서.** 조립은 `gateway.go`, **기동·취소·drain은 `cmd/tossctl`의
  `engineRuntime` 경계다.** 과거 `engine.go:377` 인용은 현재 main에서 폐기한다. `buildGateway`
  안에서 goroutine을 띄우면 automation interlock 평가보다 먼저 돈다.
  **automation gate가 verified가 아니면 워커는 기동하지 않는다.** 단 worker는 current
  `engine.Runtime.Loops`에 넣지 않는다. 그 목록은 non-cancel return 하나로 모든 안전 loop를
  내리는 all-or-nothing contract이므로 D3과 모순된다. recovery 완료 뒤 `AuxiliaryExecutor`로 시작해
  같은 context에서 cancel/drain하고, 전용 stop event/reconcile/alert를 쓰되 entry gate와 다른 loop를
  건드리지 않는다. existing alert-delivery의 stop event를 빌려 쓰지 않는다.
- [ ] 3.10 **triggered child의 durable ownership과 fill apply 경로를 하나로 만든다.**
  `internal/filldetect`는 편집하지 않는다. worker가 M-A로 증명된 parent/client/scope/generation과
  `TriggeredOrderID`를 exact 비교한 뒤 journal의 좁은 registrar로 `protection_child_orders`에
  소유권을 원자 기록한다. 그 뒤 기존 `Journal.TrackedFillOrders`가 child를 읽고 기존 Detector /
  JournalLedger가 체결 권위를 유지한다. `TrackedFillOrders`, `confirmedFillOwners`,
  `resolveFillOrigin`은 confirmed ordinary attempt와 protection child를 합쳐 **정확히 한 canonical
  owner**만 허용하며 중복/충돌은 durable reconcile로 fail closed한다.

  **소유권은 체결 관측보다 먼저 커밋되어야 한다.** registrar 시점에 같은 canonical child의
  fill snapshot/event가 이미 존재하면 소급 귀속·delta 재생·hook 재실행을 하지 않는다. registrar를
  거부하고 `ATTRIBUTION_FAILED` reconcile + stable alert를 durable 기록하며, 공식 계좌 대사가
  position을 복구할 때까지 그 child와 position에 새 보호 mutation/재등록을 금지한다. 정상 순서의
  owned child fill은 기존 RecordFill transaction으로 position/exit에 반영되고, worker는 그 durable
  cumulative snapshot을 읽어 protection lifecycle의 applied watermark를 멱등하게 전진시킨 뒤에만
  다음 mutation을 계획한다. restart·중복 registrar·RecordFill 경합에서 소유권과 lifecycle watermark가
  한 번만 전진해야 한다.

## 4. 봉인 해제와 조립 (Migration 4)

이 절은 번호순으로 실행하지 않는다. 상주 SELL을 생성할 수 있는 runtime 배선보다 먼저
**4.5~4.6(T4-A)**의 child 귀속·완전청산 취소·flat 권위 경계를 GREEN/A4-A accepted로 닫는다.
그 뒤에만 **3.9와 4.1~4.4(T4-B)**가 worker를 construction/start하여 broker mutation에
도달시킬 수 있다. 중간 빌드가 runtime에서 상주 보호주문을 생성해서는 안 된다.

- [ ] 4.0 **다섯 번째 봉인 — `protectionlifecycle` API 가드(D1). 이것 없이는 컴파일되지 않는다.**
  lifecycle 함수가 전부 unexported이고 `external_api_test.go:11`이 exported 패키지 레벨 함수를
  금지하므로 **외부 호출 표면이 0**이다. `dependency_test.go:12`가 journal·protection·execgw·app
  의존을 금지하므로 워커를 그 안에 둘 수도 없다.
  - [ ] 4.0.1 T3 worker가 실제로 호출하는 상태 전이당 하나의 명명된 exported 함수만 노출한다.
    T2-B RED가 planned worker call graph에서 필요한 함수명을 먼저 추출해 exact allowlist 상수로
    고정한다. 최소 후보는 `NewState`, `View`, `PrepareRegister`, `ApplySubmitResult`, `RecoverSubmit`,
    `PrepareReplace`, `ApplyReplaceResult`, `RecoverReplace`, `PrepareCancel`, `ApplyCancelResult`,
    `RecoverCancel`, `ApplyFill`, `DiscoverOrphan`, `ErrorCode`이며, worker가 호출하지 않는 후보는
    allowlist에서 제거한다. 인자는 `State`와 pure value/evidence 타입뿐이고 transport·approval·toggle
    타입을 받지 않는다. capability constructor가 필요하면 verified evidence value를 받는 단 하나만
    허용하고 public bool/scalar로 권위를 만들 수 없음을 RED로 고정한다.
  - [ ] 4.0.2 가드를 **대체 단언으로 바꾼다** — exported 표면이 허용 목록과 정확히 일치하고
    어느 것도 transport/approval/toggle 타입을 시그니처에 갖지 않음을 정적으로 검사한다.
  - [ ] 4.0.3 `dependency_test.go`의 의존 금지는 **그대로 둔다.**
  - [ ] 4.0.4 이 task는 lot 순서상 T2-B이며 T3보다 먼저 끝낸다. T3가 test-only shim이나 package
    내부 우회로 API 부재를 숨겨서는 안 된다.
- [ ] 4.1 `dormant_test.go`와 `a071_security_review_test.go`에서 **`protectionofficial.New` 하나만**
  금지 목록에서 뺀다. `protection.NewSupervisor`, `protection.db`, `GatewayFactory`는 유지한다.
- [ ] 4.2 **import walk를 넓힌다.** 현재 walk는 `…/internal/protection`만 매칭하므로
  `protectionofficial`·`protectionlifecycle`은 아무 app 파일에서나 import할 수 있다(리뷰 발견).
  세 패키지 + 새 워커 패키지를 walk 대상에 넣고 허용 파일을 명시한다(D6).
- [ ] 4.3 해제한 하나에 대체 단언을 붙인다 — 조립된 보호 경로가 journal-backed이고 별도 DB에
  의존하지 않으며 조립 지점이 정확히 하나임을 정적으로 검사한다.
- [ ] 4.4 워커를 `gateway.go`에서 조립한다. 「app 코드 중 보호 패키지를 import할 수 있는 파일」
  규칙을 유지한다.
- [ ] 4.5 **D8의 이중 매도 권한 계약을 구현한다.**
  - [ ] 4.5.1 인프로세스 보호 매도 직전에 그 포지션의 상주 보호주문을 취소한다.
  - [ ] 4.5.2 취소 확인 실패가 **매도를 막지 않는다.** §0-3이 청산 지연을 금지한다. typed reason만 남긴다.
  - [ ] 4.5.3 상주 주문이 먼저 채워져 포지션이 이미 flat이면 **보호 청산을 시도하지 않고**
    `ProtectionCompletedByResting`으로 기록한다. a091이 `engine-safety`에 넣은 "제출 수량 0인
    보호 청산은 critical (SHALL)"에 **도달하지 않게** 하는 것이 목적이다 — 그 요구사항에
    예외를 파면 a100 delta가 MODIFIED가 되어 a071·a091과 archive 순서로 얽힌다(design D8-3).
  - [ ] 4.5.4 포지션이 flat이면 워커가 상주 주문을 취소한다. **다음 주기까지의 창**이 남는다는
    사실과 그 창에 수동 매수가 들어오면 발동할 수 있다는 것을 6.4에 적는다.
  - [ ] 4.5.5 D9의 trigger는 **오직 현재 generation의 `exit_states.baseline_price`**에서 유도한다.
    `InitialStop` 선택지나 시세 재계산은 없다. 후퇴는 typed refusal로 거부한다.
  - [ ] 4.5.6 `record`의 기존 `isFullExit` 진입 조건은 바꾸지 않는다. 그 조건이 참일 때 기존
    working order 정리와 **별도의 비차단 시도**로 상주 conditional protection도 취소 대상으로
    포함한다. 상주 취소 실패를 기존 `clearTheSymbol`의 err/cleared나 arm suppression에 섞지 않는다.
  - [ ] 4.5.7 **발동한 child의 order id를 살린다.** 어댑터가 `TriggeredOrderID`를 bool로 접고
    (`protectionofficial/gateway.go:271`) 체결 감지는 `mutation_attempts`가 소유한 id만 추적하므로
    (`journal/fills.go:1538-1548`) **브로커가 우리 손절을 실행한 사실을 원장이 알 방법이 없다.**
    `BrokerProtection`에 raw status와 child id를 싣고, 3.10의 durable owner를 기록한다. D2의
    terminal 행은 「child apply watermark와 공식 보유를 재확인한 후에만 재등록」으로 한다.
- [ ] 4.6 **`ProtectionWired`가 여전히 생산되지 않음을 증명하는 테스트.** a071의 "구조적으로
  UNWIRED" 보장은 이 change 이후에도 유효하다. `EntryPermitted == true`를 실패로 단언하는
  기존 테스트 2건(`guardian_test.go:131-132`, `interlock_entry_test.go:70-71`)은 **그대로 둔다.**

## 5. official fixture 계약 검증 (Migration 5)

전부 `httptest`다. **실계좌 주문은 0건이다.** 격리된 config 디렉터리를 강제한다.
(0절의 M-A는 전제 확인이며 별도 승인 대상이다 — 이 절의 검증 수단이 아니다.)

- [ ] 5.1 KR·US 각각 등록 → 확인 → journal에 broker order id 커밋.
- [ ] 5.2 **기존 보유 수렴** — 보호 컬럼이 NULL인 기존 포지션이 워커의 첫 주기에 등록된다.
  이것이 이 change가 오늘 사는 안전이다.
- [ ] 5.3 **겹침 재현.** 앞선 교체가 브로커 응답을 기다리는 중에 다음 주기가 오는 경우를
  fixture로 실제로 만든다. 두 번째 수렴이 두 번째 보호주문을 만들지 않아야 한다.
  attested idempotency 증명이 없으면 재제출하지 않고 `RECONCILE_REQUIRED`로 고정된다(a071 계약).
- [ ] 5.4 더 안전한 방향으로만 교체. trigger 후퇴 거부.
- [ ] 5.5 취소와 재기동 복구 — stable operation identity와 exact broker 조회로 귀속.
- [ ] 5.6 **이중 매도 권한 경합** — 4.5의 세 경로를 fixture로 재현한다.
- [ ] 5.7 시장 격리 — KR 실패가 US 수렴과 두 시장의 청산·대사를 바꾸지 않는다.
- [ ] 5.8 수렴 워커 실패가 체결 감지 사이클과 대사 사이클의 outage·SLO에 **아무 영향이 없음**을
  증명한다(D3의 핵심 주장).
- [ ] 5.9 **대사와의 경합 재현.** 워커가 수량을 읽은 뒤 브로커 왕복 중에 대사가 포지션을 닫는
  시퀀스를 fixture로 만든다. 3.8의 재확인이 등록을 막아야 한다. 5.3(워커 주기 두 개의 겹침)과
  다른 시나리오다.
- [ ] 5.10 **익절 경로 재현.** 익절 완전 청산 전에 상주 주문 취소가 시도됨을 증명한다(4.5.6).
- [ ] 5.11 **발동 후 원장 귀속 재현.** 상주 주문이 발동해 child가 체결되면 그 사실이 기존
  fill detector의 snapshot→journal apply transaction을 통해 position/exit에 반영되고, 재등록이
  일어나지 않음을 증명한다(3.10, 4.5.7). child fill이 registrar보다 먼저 도착하면 소급 적용 없이
  durable `ATTRIBUTION_FAILED` reconcile/alert가 남고 새 보호 mutation이 금지되어야 한다.
  registrar/RecordFill/restart가 겹치는 정상 순서에서는 lifecycle applied watermark가 정확히 한 번만
  이동하고, ordinary order ownership과 충돌하면 durable identifier reconcile이 남는지 단언한다.

## 6. 롤백과 운영 절차 (Migration 6)

- [ ] 6.1 **롤백 절차를 다시 쓴다(D4).** 이 journal은 더 새로운 스키마를 이전 바이너리가 여는
  것을 거부한다(`journal/journal.go:23-27`, `:240`) — 「구버전이 같은 journal을 연다」는 성립하지
  않는다. 스키마를 올린 뒤의 롤백은 백업 복원이고, **복원된 journal은 상주 보호주문의 기록을
  잃는다**(`journal/backup.go`). 그러면 구버전 엔진이 모르는 채 자기 청산을 내어 매도 권한이
  둘이 된다.
  - [ ] 6.1.1 절차 문서: 복원 **전에** 브로커 상주 보호주문 목록을 기록 → 복원 → 기록을 근거로
    사람이 각 주문의 처분을 정한다. **a100은 자동 취소도 자동 유지도 하지 않는다.**
  - [ ] 6.1.2 자동 롤백이 상주 주문을 취소하지 않음을 회귀 테스트로 고정하되, **그것이 안전을
    보장하지 않는다**는 문장을 문서에 함께 둔다.
- [ ] 6.2 보호 미설치 시간과 수렴 실패를 `/position-management`의 관리 행에 표시한다.
  raw status, desired/observed 수량·trigger, 현재 baseline과의 차, 마지막 성공/오류 시각,
  unprotected elapsed, stable alert cause를 한 shared projection에서 만들고 alert와 화면이 같은
  identity/cause를 쓰는지 확인한다. trigger·수량·ownership evidence가 불완전하면 actionable
  `보호됨`을 표시하지 않는다.
- [ ] 6.3 lane은 6개 모두 `Desired=OFF, Effective=OFF`로 남는다. 이 change는 어떤 토글도 flip하지 않는다.
- [ ] 6.4 운영 절차 문서에 세 가지를 적는다. 운영자가 보는 화면에 없던 것이 생기는 유일한
  관찰 가능 변화이므로 오해할 여지를 남기지 않는다.
  - [ ] 6.4.1 **브로커에 상주 주문이 생긴다**는 사실과 수동 확인 방법.
  - [ ] 6.4.2 **상주 trigger는 현재 generation의 영속 `baseline_price`로 수렴한다**(D9).
    교체 전에는 broker trigger와 baseline이 다를 수 있으므로 그 delta와 경과 시간을 표시한다.
  - [ ] 6.4.3 포지션이 비-보호 경로로 닫히면 **다음 수렴 주기까지 상주 주문이 남고**, 그 창에
    수동 매수가 들어오면 그 주식에 대해 발동할 수 있다(M13 — 수량 예약 없음).
- [x] 6.5 **`protection.Controller`(824줄) + `Repository`(540줄) 정리 change를 지금 등록한다.**
  a100 완료 후로 미루지 않는다 — 리뷰의 두 보이스가 모두 선행을 권고했고, 등록만이라도
  앞당기는 것이 절충이다(D1).
  — 2026-08-13 등록 완료: `openspec/changes/a107-retire-the-second-protection-core/` +
  STORY-TOS-a107. 이 항목만 6절에서 먼저 실행한 근거는 항목 자신의 "지금 등록한다"이며
  구현은 등록에 포함되지 않는다(a107 착수 조건: a100 land — 같은 봉인 파일 편집).
  a104~a106은 이 문서가 예약한 번호라 비워 두고 a107을 썼다.

## 7. 게이트

- [ ] 7.1 편집한 모든 기존 함수의 Function Logic Map을 **편집 후 재생성**해 SHA-256을 맞춘다.
- [ ] 7.2 `python3 tools/logic-map/check_analysis.py --change a100-wire-fill-to-broker-protection`
- [ ] 7.3 영향 패키지 `go test` + `go test -race` + `go vet`
  (`protectionlifecycle`, `protectionofficial`, `protection`, `execgw`, `filldetect`, `app/engine`,
  `journal`, `exitpolicy`, `cmd/tossctl`, 새 워커 패키지).
- [ ] 7.4 `openspec validate --all --strict`
- [ ] 7.5 `make sdd-sync` → `make sdd-check` → `make gate CHANGE=a100-wire-fill-to-broker-protection`
  (병행 세션이 커밋을 쌓았으면 base 재고정 후 연속 실행)
  2026-08-12의 a099 RED blocker 주장은 current main의 사실로 재사용하지 않는다. R0에서 Terra가
  baseline을 다시 실행해 실제 실패만 기록하고, 과거 외부 blocker를 선험적으로 면제하지 않는다.
- [ ] 7.6 gstack 독립 리뷰(구현 후). High-risk이므로 adversarial Eng voice가 필수다.
  proposal-freeze 리뷰는 2026-08-11에 완료했고 기록은 `review.md`다.
- [ ] 7.7 배포 전 main과 `SchemaVersion` 대조(2.5).
- [ ] 7.7b **`RequiredEndpoints()` 갱신과 attestation 재발급을 같은 배포 단위로 묶고 순서를
  고정한다(D11, 0.10).** 재발급 없이 카탈로그만 갱신한 배포는 **두 시장의 모든 루프를 거부시킨다.**
  이것이 a100이 배포되지 못하게 만들 수 있는 유일한 경로다.
- [ ] 7.8 `review.md`에 구현 리뷰 결과를 추가한다 — 생략한 단계가 있으면 `not-applicable` 사유를
  명시한다. **침묵한 생략은 금지다.**
- [ ] 7.9 PM 동기화 — `STORY-TOS-a100.yaml`의 acceptance를 새 범위에 맞게 고치고 증거와 대조한다.
  **원안 acceptance는 `Wired` 생산을 포함하므로 그대로 두면 통과할 수 없다.**
- [ ] 7.10 a105에 이관 항목을 기록한다(아래 「a105로 이관」 목록 전체).

## 8. 구현 로트별 독립 검토 원장

아래 항목은 해당 구현 task를 완료했다는 주장보다 **먼저** 닫혀야 한다. 리뷰어는 구현 파일을
편집하지 않고 반례·mutation·회귀 실행 근거로 판정한다. P0/P1 발견 시 구현자가 수정하고 같은
리뷰어가 재검토한다. Manager는 diff·task·증거를 대조할 뿐 코드·테스트를 작성하거나 고치지 않는다.

- [x] 8.0 M0 measurement-only 구현 Terra 보고 → A-M0 적대 리뷰 → 수정 → A-M0 `P0=0/P1=0`
  → Manager acceptance. M0는 제품 worker 구현 착수가 아니라 M-A 증거를 가능하게 하는 선행 도구다.
  — core reviewer ACCEPT(P0=0/P1=0, P2=RED chronology audit note), receipt same-reviewer recheck
  ACCEPT(P0=0/P1=0/P2=0). P2는 첫 trace seam의 독립 재생 가능한 pre-GREEN commit 영수증이 없다는
  과정 기록이며 결과 동작을 약화하지 않는다.
- [ ] 8.1 T1 구현 Terra 보고 → A1 적대 리뷰 → 수정 → A1 `P0=0/P1=0` → Manager acceptance.
- [ ] 8.2 T2-A 구현 Terra 보고 → A2-A 적대 리뷰 → 수정 → A2-A `P0=0/P1=0` → Manager acceptance.
- [ ] 8.2b T2-B 구현 Terra 보고 → A2-B 적대 리뷰 → 수정 → A2-B `P0=0/P1=0` → Manager acceptance.
- [ ] 8.3 T3 구현 Terra 보고 → A3 적대 리뷰 → 수정 → A3 `P0=0/P1=0` → Manager acceptance.
- [ ] 8.4 T4-A 구현 Terra 보고 → A4-A 적대 리뷰 → 수정 → A4-A `P0=0/P1=0` → Manager acceptance.
- [ ] 8.5 T4-B 구현 Terra 보고 → A4-B 적대 리뷰 → 수정 → A4-B `P0=0/P1=0` → Manager acceptance.
- [ ] 8.6 T5 구현 Terra 보고 → A5 적대 리뷰 → 수정 → A5 `P0=0/P1=0` → Manager acceptance.
- [ ] 8.7 모든 로트 후 gstack review를 실행하고 P0/P1을 0으로 닫는다. 그 뒤 Manager가
  OpenSpec task/FLM/BTM/mutation/gate/PM 증거를 독립 대조해 구현 완료 여부를 판정한다.

## a105로 이관 (proposal-freeze 리뷰, 2026-08-11)

이 목록은 원안 a100의 범위였고 리뷰가 잘라냈다. **사라진 것이 아니라 옮겨진 것이다.**

- `internal/protectionsupervisor` 신규 패키지와 시장별 `Wired` 판정(원안 D2)
- `productionProtectionAssemblies`의 `wired` 파라미터화와 identity 문자열 교체(원안 3.3·3.4)
- `ProductionProvider.initialize`의 refusal 분화(원안 3.5) — **FLM 선행 대상**
- manifest digest 불일치의 진단 가능한 refusal, 배포 절차의 재서명(원안 D3·6.2·6.3)
- **서명 도구.** 저장소에 서명자가 없다 — `internal/attest/protection_signature.go`는 검증
  전용이고 서명 함수는 `_test.go`에만 있으며 `supervisor_digest`를 발행하는 `cmd/` 경로가 없다.
  a105는 이것 없이 `Wired`를 켤 수 없다.
- 포지션 단위 coverage latch와 entry supervisor 소비(원안 D5·4.8)
- `engine.runInterlock` B3 — `WIRED`가 생산 가능해야 도달한다(원안 1.3.1)
- `a071_security_review_test.go`의 무조건 단언 반전(원안 3.6)
- `guardian_test.go:131-132`·`interlock_entry_test.go:70-71`의 `EntryPermitted` 전제 변경(원안 4.9)
- **flat 포지션에 상주 주문이 남는 창을 닫는 것.** a100은 다음 수렴 주기까지의 창을 운영 문서로
  처리한다(6.4.3). 자동 매수가 생기면 그 창은 「방금 산 주식에 남의 손절이 걸리는」 경로가
  되므로, **a105가 진입을 열기 전에 닫아야 한다.**

## 범위 밖 (확인용)

- 레인 활성화(a105), threshold 승인(a101), 라이브 평가(a103), 사이징 역산(a104).
- 전략적 실계좌 주문 1회는 a106이다. 0절의 M-A는 전제 확인이며 별개다.
- attestation 스키마·서명·키 수명·trusted-time floor는 a071이 만들었다. a100은 소비도 하지 않는다.
- a071이 openspec `engine-safety`에서 MODIFY 중인 요구사항은 **건드리지 않는다.**
  a100의 delta는 전부 ADDED이므로 archive 순서가 서로의 텍스트를 덮어쓰지 않는다.
