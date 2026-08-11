# Tasks — a100-wire-fill-to-broker-protection

순서는 `design.md`의 Migration Plan을 따른다. **1절이 끝나기 전에 2절 이후를 시작하지 않는다** —
배선 대상 함수의 거부 분기 9개가 한 번도 실행된 적이 없고, 그대로 켜면 미검증 결함이 배선과
동시에 프로덕션으로 나가기 때문이다(`proposal.md` 「Evidence produced」).

## 0. 착수 전 조건

- [ ] 0.1 `base-commit.txt`가 현재 작업 base와 일치하는지 확인한다. 병행 세션이 커밋을 쌓았으면
  `python3 tools/sdd/capture_change_base.py`로 재고정한 뒤 `make sdd-sync`를 다시 돌린다.
- [ ] 0.2 **편집 전 Function Logic Map 대상 확인.** 아래 세 함수는 a100이 내부를 편집하므로
  편집 **전에** `tools/logic-map` AST 산출물을 만든다. 이미 만든 것은 재생성해 SHA-256을 맞춘다.
  - `internal/protectionreadiness/production.go` — `ProductionProvider.initialize` (`:252`).
    D2의 진단 가능한 refusal이 이 함수의 early return을 바꾼다. **아직 산출물이 없다.**
  - `internal/app/engine/protection_wiring.go` — `productionProtectionAssemblies`. 산출물 있음(분기 0).
    시그니처가 바뀌므로 편집 후 재생성한다.
  - `cmd/tossctl/engine.go` — `engineFillDetector`. `Ledger` 데코레이터를 꽂는 지점이다.
    **아직 산출물이 없다.**
- [ ] 0.3 `strategyDispatchCycle.dispatch`를 편집하지 않는다는 D7 면제가 여전히 유효한지 확인한다.
  편집하거나 그 내부 분기를 근거로 인용하게 되면 **그 시점에** AST를 만들고 1절 착수 전에 완성한다.
- [ ] 0.4 High-risk Pre-Edit 선언을 남긴다. 이 change는 체결·보호주문·진입 인터록에 모두 닿는다.

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
- [ ] 1.1.5 **B7** 잔량 0 → `Terminal` 전이 — **완전 체결로 보호가 닫히는 경로.** 종료 후
  `EntryOpen`이 열리지 않고 추가 전이가 멱등하게 거부된다.

### 1.2 `protectionlifecycle.prepareRegister` — 미실행 4개

- [ ] 1.2.1 **B2** entry latched / phase 부적합 — 진입이 닫힌 상태의 등록 시도가 거부된다.
- [ ] 1.2.2 **B3** 이미 pending인 operation — **중복 제출 방지.** 재시도가 두 번째 제출을 만들지 않는다.
- [ ] 1.2.3 **B4** 보호가 이미 active — 같은 포지션에 두 번째 보호주문이 나가지 않는다.
- [ ] 1.2.4 **B5** 브로커가 정확한 operation 조회를 못 함 — capability 부재가 실제로 판정을 막는다.

### 1.3 `engine.runInterlock` — 미실행 1개

- [ ] 1.3.1 **B3** 수락 audit 기록 실패 시 기동 실패 — 진입이 허용된 사실이 기록되지 못했는데
  기동이 계속되면 추적 불가능한 자동매매가 된다(안전 불변식 §0-5의 실패 방향).

### 1.4 측정

- [ ] 1.4.1 `go test -covermode=set -coverprofile`로 **true 결과 본문 행**의 실행 여부를 다시
  측정한다. 조건 statement의 covered는 근거가 아니다(`branch-test-map.md`의 측정 방법과 동일).
- [ ] 1.4.2 세 `branch-test-map.md`를 측정값으로 갱신한다. 10개 모두 `yes`가 아니면 2절로 가지 않는다.

## 2. 도메인 매핑과 journal 스키마 (Migration 2)

- [ ] 2.1 `protectionlifecycle` ↔ `internal/protection` 도메인 타입 매핑을 순수 함수로 만든다
  (`Scope`, `ConditionalBody`, `BrokerTarget`, `BrokerProtection`).
- [ ] 2.2 왕복(round-trip) property 테스트로 매핑이 정보를 잃지 않음을 증명한다. 손실이 있으면
  lifecycle이 증명한 불변식이 전송 단계에서 깨진다(design Risks).
- [ ] 2.3 journal에 보호 컬럼을 **additive-nullable로만** 추가한다. 기존 컬럼의 의미를 바꾸지 않는다.
- [ ] 2.4 새 컬럼을 모르는 이전 바이너리로 여는 회귀 테스트 — 체결 감지·대사·reduce-only 청산이
  그대로 동작하고 값이 없는 행은 「보호 미설치」로 읽힌다.
- [ ] 2.5 `SchemaVersion` 증가와 main 대조. 낮은 SchemaVersion 바이너리가 뜨면 콘솔만 뜨고 엔진이
  조용히 죽는다(2026-08-04 실발생). 배포 전 대조를 `tasks 7`에 건다.

## 3. supervisor와 assembly 파라미터화 (Migration 3)

이 절이 끝나도 **프로덕션 관찰 동작은 바뀌지 않는다.** supervisor 입력이 없으면 두 시장 모두
`false`이고 그것이 현재 상태와 같기 때문이다.

- [ ] 3.1 `internal/protectionsupervisor`(신규 패키지)를 만든다. `internal/protection` 안에 두지
  않는다 — `protection.NewSupervisor` 금지 심볼을 유지하기 위해서다(D6).
- [ ] 3.2 시장별 `Wired` 판정을 D2의 세 조건으로 구현한다. 하나라도 아니면 **그 시장만** false.
- [ ] 3.3 `productionProtectionAssemblies`에 `wired map[Market]bool` 파라미터를 추가한다.
  함수는 순수 생성자로 남는다(분기 0을 유지한다).
- [ ] 3.4 `"fill-lifecycle-unwired"` identity 문자열을 배선된 상태를 뜻하는 값으로 바꾼다(D3).
- [ ] 3.5 **`ProductionProvider.initialize`의 진단 가능한 refusal.** 현재는 assembly가 하나라도
  `Wired: false`면 `initialize()`를 포기해 `globalRefusal`이 `RefusalInvalid` 초기값으로 남는다
  (`production.go:293`, `:117`). 미배선·digest 불일치·중복 시장·잘못된 시장을 각각 구별되는
  refusal로 분리한다. **0.2의 FLM이 선행 조건이다.**
- [ ] 3.6 `a071_security_review_test.go`의 `TestUnprovenFillLifecycleKeepsBothProductionAssembliesUnwired`를
  무조건 단언에서 supervisor 입력에 따른 조건부 단언으로 바꾼다.
- [ ] 3.7 supervisor 입력이 없을 때 두 시장 모두 `Wired`가 아니고 관측 동작이 배선 이전과 동일함을
  증명하는 테스트.

## 4. 봉인 해제와 조립 (Migration 4)

- [ ] 4.1 `dormant_test.go`와 `a071_security_review_test.go`에서 **`protectionofficial.New` 하나만**
  금지 목록에서 뺀다. `protection.NewSupervisor`, `protection.db`, `GatewayFactory`는 유지한다.
- [ ] 4.2 해제한 하나에 대체 단언을 붙인다 — 조립된 보호 경로가 journal-backed이고 별도 DB에
  의존하지 않으며 조립 지점이 정확히 하나임을 정적으로 검사한다.
- [ ] 4.3 「app 코드 중 `internal/protection`을 import할 수 있는 파일은 `gateway.go` 하나」 규칙을
  **바꾸지 않는다**(`dormant_test.go:61`). a100의 조립도 `gateway.go` 안에서 한다.
- [ ] 4.4 `filldetect.Ledger` 데코레이터를 만든다(D8). `internal/filldetect`는 **편집하지 않는다** —
  같은 인터페이스를 만족하는 구현을 `cmd/tossctl/engine.go:428-430`의 조립 지점에서 감싼다.
- [ ] 4.5 데코레이터의 촉발 조건: `applied.Delta > 0 && !applied.FailClosed`.
  `applied.Corrected`(누적 수량 불변)는 교체를 촉발하지 않는다.
- [ ] 4.6 **실패 격리 테스트 — 이 절에서 가장 중요하다.** 보호 계획이 실패해도 데코레이터가
  `Apply` 에러를 만들지 않는다. `PollOnce`는 `Ledger.Apply` 에러를 사이클 실패 + outage로 다루므로
  (`detect.go:317-322`) 보호 실패를 에러로 올리면 **체결 감지 루프가 멈춘다.** 그것은 안전 불변식
  §0-3(손절·비상 청산의 즉시성을 약화·지연하지 않는다) 위반이다.
- [ ] 4.7 보호 계획 실패는 coverage latch를 닫고 typed reconcile reason을 기록하는 것으로 끝난다.
- [ ] 4.8 coverage latch를 `internal/app/engine/strategy_entry_supervisor.go`에서 소비한다(D5).
  `ReadinessRequest`에 포지션 식별자를 추가하지 않는다 — a071의 봉인된 계약이다.
- [ ] 4.9 `guardian_test.go:131-132`와 `interlock_entry_test.go:70-71`의 `EntryPermitted == true`
  실패 단언을 갱신한다. `WIRED`가 생산 가능해지면 전제가 바뀐다.

## 5. official fixture 계약 검증 (Migration 5)

전부 `httptest`다. **실계좌 주문은 0건이다.** 격리된 config 디렉터리를 강제한다.

- [ ] 5.1 KR·US 각각 등록 → 확인 → journal에 broker order id 커밋.
- [ ] 5.2 부분체결 수렴 — 한 주기의 누적 delta 하나에 수렴 1회. 부분체결 건수만큼 왕복이
  발생하지 않음을 왕복 횟수로 증명한다.
- [ ] 5.3 **겹침 재현.** 앞선 교체가 브로커 응답을 기다리는 중에 다음 delta가 도착하는 경우를
  fixture로 실제로 만든다. 두 번째 수렴이 두 번째 보호주문을 만들지 않아야 한다.
  attested idempotency 증명이 없으면 재제출하지 않고 `RECONCILE_REQUIRED`로 고정된다(a071 계약).
- [ ] 5.4 더 안전한 방향으로만 교체. trigger 후퇴 거부.
- [ ] 5.5 취소와 재기동 복구 — stable operation identity와 exact broker 조회로 귀속.
- [ ] 5.6 manifest digest 불일치(재서명 누락)가 조용한 미배선이 아니라 원인을 지목하는 refusal로
  드러난다.
- [ ] 5.7 시장 격리 — KR 실패가 US 판정과 두 시장의 청산·대사를 바꾸지 않는다.

## 6. 롤백과 운영 절차 (Migration 6)

- [ ] 6.1 롤백은 신규 진입을 OFF로 유지하고 **이미 브로커에 상주하는 보호주문을 취소하지 않는다.**
  회귀 테스트로 고정한다.
- [ ] 6.2 배포 절차 문서에 **manifest 재서명** 단계를 추가한다. supervisor digest가 build-bound이므로
  (`vcs.revision` 포함) 재빌드마다 이전 manifest는 무효다(D3).
- [ ] 6.3 재서명 누락이 조용한 UNWIRED가 되지 않도록 3.5의 refusal이 콘솔·상태에 보이는지 확인한다.
- [ ] 6.4 lane은 6개 모두 `Desired=OFF, Effective=OFF`로 남는다. 이 change는 어떤 토글도 flip하지 않는다.

## 7. 게이트

- [ ] 7.1 편집한 모든 기존 함수의 Function Logic Map을 **편집 후 재생성**해 SHA-256을 맞춘다.
- [ ] 7.2 `python3 tools/logic-map/check_analysis.py --change a100-wire-fill-to-broker-protection`
- [ ] 7.3 영향 패키지 `go test` + `go test -race` + `go vet`
  (`protectionlifecycle`, `protectionofficial`, `protectionreadiness`, `protectionsupervisor`,
  `protection`, `execgw`, `filldetect`, `app/engine`, `journal`, `cmd/tossctl`).
- [ ] 7.4 `openspec validate --all --strict`
- [ ] 7.5 `make sdd-sync` → `make sdd-check` → `make gate CHANGE=a100-wire-fill-to-broker-protection`
  (병행 세션이 커밋을 쌓았으면 base 재고정 후 연속 실행)
- [ ] 7.6 gstack 독립 리뷰. High-risk이므로 adversarial Eng voice가 필수다.
- [ ] 7.7 배포 전 main과 `SchemaVersion` 대조(2.5).
- [ ] 7.8 `review.md` 작성 — 생략한 단계가 있으면 `not-applicable` 사유를 명시한다.
  **침묵한 생략은 금지다.**
- [ ] 7.9 PM 동기화 — `STORY-TOS-a100.yaml`의 acceptance 8개를 증거와 대조한다.
- [ ] 7.10 후속 정리 change 등록 — `protection.Controller`(824줄) + `Repository`(540줄)는 D1에 따라
  프로덕션에서 죽은 코드로 남는다. a100 완료 시 삭제 change를 등록한다(design D1의 약속).

## 범위 밖 (확인용)

- 레인 활성화(a105), threshold 승인(a101), 라이브 평가(a103), 사이징 역산(a104).
- 실계좌 주문 1회는 a106이다.
- attestation 스키마·서명·키 수명·trusted-time floor는 a071이 만들었고 소비만 한다.
- a071이 openspec `engine-safety`에서 MODIFY 중인 "ProtectionReady는 attestation 범위에서만 WIRED다"
  요구사항은 **건드리지 않는다.** a100의 delta는 전부 ADDED이므로 두 change의 archive 순서가
  서로의 텍스트를 덮어쓰지 않는다.
