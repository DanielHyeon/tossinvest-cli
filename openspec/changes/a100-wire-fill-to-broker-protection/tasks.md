# Tasks — a100-wire-fill-to-broker-protection

순서는 `design.md`의 Migration Plan을 따른다.

**0절이 끝나기 전에 1절을 시작하지 않고, 1절이 끝나기 전에 2절 이후를 시작하지 않는다.**
0절은 이 change의 전제가 참인지 묻고(선행 실측), 1절은 배선 대상 함수의 거부 분기 9개가 한 번도
실행된 적이 없다는 측정에 대응한다. 그대로 켜면 미검증 결함이 배선과 동시에 프로덕션으로 나간다.

## 0. 착수 전 조건

- [x] 0.1 `base-commit.txt`가 현재 작업 base와 일치하는지 확인한다. 병행 세션이 커밋을 쌓았으면
  `python3 tools/sdd/capture_change_base.py`로 재고정한 뒤 `make sdd-sync`를 다시 돌린다.
  — 2026-08-11 확인: base `eb41e19a`, HEAD `ce78b0db`. 사이의 커밋 3개는 **전부 a100 자신의
  문서 커밋**이고 Go 변경이 없다. 따라서 함수 귀속의 비교 기준은 여전히 유효하며 재고정하지
  않는다(재고정은 fingerprint 재sync만 더 들고 얻는 것이 없다). **병행 세션 커밋은 없었다.**
- [ ] 0.2 **선행 실측 M-A — 조건주문이 실제로 발동해 체결되는가.** `measurement-prereq.md`의
  절차를 따른다. **실계좌 주문이므로 실행 직전에 사람이 승인한다**(§0-1, §0-7).
  **실패하면 여기서 멈춘다** — 발동하지 않는 상주 주문은 보호가 아니므로 설계가 아니라
  전제가 무너진다(design D10).
- [x] 0.3 **선행 실측 M-B — 엔진 사망부터 exit observer 재무장까지의 무보호 창.** 이 change가
  사려는 안전의 크기다. 측정하지 않으면 개선 폭을 주장할 수 없다.
  — 2026-08-11 14:00Z, 0.10b의 배포 재시작에서 측정. **≤ 6.7초**(명령 발행 → 첫 `exit.*` 관측).
  기록은 `measurement-prereq.md`. **다만 이 값은 하한이다** — a056의 8분은 *기동 실패* 경우이고
  이번 측정은 정상 기동 경우다. a100이 사는 안전은 6.7초가 아니라 그 두 값 사이다.
- [x] 0.4 **편집 전 Function Logic Map 대상 확정.** 6건을 특정하고 AST·FLM을 만들었다.
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
- [x] 0.5 산출물의 source SHA-256 일치를 `check_analysis.py`로 확인한다.
- [x] 0.6 `strategyDispatchCycle.dispatch` D7 면제 유효. 편집하지 않고, 분기 수치는 2차 개정에서
  삭제했으므로 내부 분기를 근거로 인용하지도 않는다.
- [x] 0.7 High-risk Pre-Edit 선언을 `pre-edit-gate.md`에 남겼다. 6개 대상 중 **5개가
  「조건부 통과」**이며 각 조건이 그 편집의 통과 요건이다. 「실패 테스트 선행 작성」은 각 task
  착수 시점에 갱신한다.
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
    새로 3일 이상 돌려야 한다.**
  ⇒ 배포 순서 고정: **(a) soak probe 추가 → (b) soak 재실행(≥3일, 기록 신선)
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
- [ ] 0.11 **raw conditional status를 도메인 타입에 실을지 결정한다(D2).** 어댑터가
  `WATCHING/PAUSED/ORDERING/ORDERED`를 같은 값으로 접으므로(`protectionofficial/gateway.go:308-310`)
  일시정지된 주문과 무장된 주문이 구별되지 않는다. 싣지 않기로 하면 M-A가 `PAUSED`의 실재와
  발동 여부를 관측하고, 실재하며 발동하지 않으면 **싣는 것이 필수 task가 된다.**

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
  최소한 브로커 order id, lifecycle 상태, desired가 생긴 시각(미설치 시간 측정용)을 담는다.
- [ ] 2.4 새 컬럼을 모르는 이전 바이너리로 여는 회귀 테스트 — 체결 감지·대사·reduce-only 청산이
  그대로 동작하고 값이 없는 행은 「보호 미설치」로 읽힌다.
- [ ] 2.5 `SchemaVersion` 증가와 main 대조. 낮은 SchemaVersion 바이너리가 뜨면 콘솔만 뜨고 엔진이
  조용히 죽는다(2026-08-04 실발생). 배포 전 대조를 7절에 건다.
- [ ] 2.6 **전송 전 pending 커밋.** `recoverSubmit`은 영속된 pending command가 없으면
  "no exact submit pending"으로 시장을 latch한다(`protectionlifecycle/lifecycle.go:63`, `:71`).
  저장 순서를 「plan → **pending 커밋** → 전송 → 응답 커밋」으로 고정하고, 전송 직후 프로세스가
  죽는 시나리오를 테스트한다. 2.3의 컬럼 목록이 이 레코드를 포함해야 한다.

## 3. 수렴 워커 (Migration 3)

이 절이 끝나도 **프로덕션 관찰 동작은 바뀌지 않는다.** 워커가 조립되지 않기 때문이다(조립은 4절).

- [ ] 3.1 수렴 워커 패키지를 만든다. `internal/filldetect`도 `ReconcileDriver`도 편집하지 않는다.
  자기 주기로 돌고 입력은 journal의 커밋된 상태뿐이다(D3).
- [ ] 3.2 desired/observed 비교를 D2의 표대로 구현한다. 각 행이 하나의 테스트를 갖는다.
- [ ] 3.3 **수렴 완료의 정의를 코드로 고정한다** — 보유 수량과 ACTIVE 보호 수량이 정확히 같고
  trigger가 journal 값에서 유도된 값과 같을 때만 수렴이다. "대략 맞음"을 허용하지 않는다.
- [ ] 3.4 **보호 미설치 시간**을 포지션별로 측정하고 상한 초과 시 알림을 낸다(D2).
  상한이 없으면 「수렴 중」과 「영원히 실패 중」이 구별되지 않는다.
- [ ] 3.5 실패 격리 — 수렴 실패가 typed reconcile reason과 알림으로 끝나고 다른 루프의
  outage·SLO·게이트를 건드리지 않음을 테스트한다.
- [ ] 3.6 재시도와 백오프. **수렴한 포지션도 주기를 늘려 계속 재확인한다** — 상주 주문은
  나중에 취소·만료·일시정지될 수 있고 그때 보호 미설치 시간이 다시 시작되어야 한다.
  rate limit은 주기로 다루고 판정의 신선도를 버리지 않는다(리뷰 정정).
- [ ] 3.7 워커가 조립되지 않았을 때 관측 가능한 동작이 배선 이전과 완전히 동일함을 증명하는 테스트.
- [ ] 3.8 **대사와의 경합을 막는다.** 전송 **직전**에 포지션 수량·generation을 다시 읽어 계획
  시점과 다르면 전송하지 않고 그 주기를 버린다. 응답 후에도 다시 읽어 그 사이 수량이 바뀌었으면
  수렴 완료로 기록하지 않는다. 경합 원문: 워커가 10주를 읽는다 → 대사가 0주를 커밋한다
  (`reconcileloop.go:413`, `reconcile/converge.go:209`, `:244`) → 워커가 SELL 10을 등록한다.
- [ ] 3.9 **실행 주체와 기동 순서.** 조립은 `gateway.go`, **기동·취소·감독은 `cmd/tossctl`의
  런타임**(`engine.go:377`의 감독 루프 옆). `buildGateway` 안에서 goroutine을 띄우면 automation
  interlock 평가보다 먼저 돈다(`app/engine/engine.go:489`, `:533`).
  **automation gate가 verified가 아니면 워커는 기동하지 않는다** — 대사·exit 루프와 같은 규율이다.

## 4. 봉인 해제와 조립 (Migration 4)

- [ ] 4.0 **다섯 번째 봉인 — `protectionlifecycle` API 가드(D1). 이것 없이는 컴파일되지 않는다.**
  lifecycle 함수가 전부 unexported이고 `external_api_test.go:11`이 exported 패키지 레벨 함수를
  금지하므로 **외부 호출 표면이 0**이다. `dependency_test.go:12`가 journal·protection·execgw·app
  의존을 금지하므로 워커를 그 안에 둘 수도 없다.
  - [ ] 4.0.1 상태 전이당 하나의 명명된 exported 함수만 노출한다. 인자는 `State`와 값 타입뿐이며
    transport·approval·toggle 타입을 받지 않는다.
  - [ ] 4.0.2 가드를 **대체 단언으로 바꾼다** — exported 표면이 허용 목록과 정확히 일치하고
    어느 것도 transport/approval/toggle 타입을 시그니처에 갖지 않음을 정적으로 검사한다.
  - [ ] 4.0.3 `dependency_test.go`의 의존 금지는 **그대로 둔다.**
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
  - [ ] 4.5.5 D9의 trigger 유도를 구현한다 — 0.8의 확인 결과에 따라 `InitialStop`만 쓰거나
    더 높은 스칼라로 교체한다. 후퇴는 거부한다.
  - [ ] 4.5.6 **취소 대상을 `isProtective`가 아니라 `isFullExit`로 넓힌다.** 익절도 완전
    청산이지만 보호 경로가 아니다(`exitloop.go:1207-1213`) — 좁게 두면 익절이 상주 주문을
    남긴 채 포지션을 닫는다.
  - [ ] 4.5.7 **발동한 child의 order id를 살린다.** 어댑터가 `TriggeredOrderID`를 bool로 접고
    (`protectionofficial/gateway.go:271`) 체결 감지는 `mutation_attempts`가 소유한 id만 추적하므로
    (`journal/fills.go:1538-1548`) **브로커가 우리 손절을 실행한 사실을 원장이 알 방법이 없다.**
    `BrokerProtection`에 child id를 싣고, D2의 terminal 행을 「보유 재확인 후에만 재등록」으로 한다.
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
- [ ] 5.11 **발동 후 원장 귀속 재현.** 상주 주문이 발동해 child가 체결되면 그 사실이 원장에
  반영되고, 재등록이 일어나지 않음을 증명한다(4.5.7).

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
- [ ] 6.2 보호 미설치 시간과 수렴 실패가 콘솔·알림에 보이는지 확인한다.
- [ ] 6.3 lane은 6개 모두 `Desired=OFF, Effective=OFF`로 남는다. 이 change는 어떤 토글도 flip하지 않는다.
- [ ] 6.4 운영 절차 문서에 세 가지를 적는다. 운영자가 보는 화면에 없던 것이 생기는 유일한
  관찰 가능 변화이므로 오해할 여지를 남기지 않는다.
  - [ ] 6.4.1 **브로커에 상주 주문이 생긴다**는 사실과 수동 확인 방법.
  - [ ] 6.4.2 **상주 trigger는 엔진의 현재 청산선이 아니라 재난 하한이다**(D9). 둘이 다를 수 있다.
  - [ ] 6.4.3 포지션이 비-보호 경로로 닫히면 **다음 수렴 주기까지 상주 주문이 남고**, 그 창에
    수동 매수가 들어오면 그 주식에 대해 발동할 수 있다(M13 — 수량 예약 없음).
- [ ] 6.5 **`protection.Controller`(824줄) + `Repository`(540줄) 정리 change를 지금 등록한다.**
  a100 완료 후로 미루지 않는다 — 리뷰의 두 보이스가 모두 선행을 권고했고, 등록만이라도
  앞당기는 것이 절충이다(D1).

## 7. 게이트

- [ ] 7.1 편집한 모든 기존 함수의 Function Logic Map을 **편집 후 재생성**해 SHA-256을 맞춘다.
- [ ] 7.2 `python3 tools/logic-map/check_analysis.py --change a100-wire-fill-to-broker-protection`
- [ ] 7.3 영향 패키지 `go test` + `go test -race` + `go vet`
  (`protectionlifecycle`, `protectionofficial`, `protection`, `execgw`, `filldetect`, `app/engine`,
  `journal`, `exitpolicy`, `cmd/tossctl`, 새 워커 패키지).
- [ ] 7.4 `openspec validate --all --strict`
- [ ] 7.5 `make sdd-sync` → `make sdd-check` → `make gate CHANGE=a100-wire-fill-to-broker-protection`
  (병행 세션이 커밋을 쌓았으면 base 재고정 후 연속 실행)
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
