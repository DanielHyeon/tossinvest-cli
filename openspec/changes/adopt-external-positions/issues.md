# Issues: adopt-external-positions

> 구현 중 발견한 스펙·코드 마찰과 §0 검토 기록. advisory only — 권위는 spec + 코드 + 테스트.
> 분류: **blocking**(구현 중단 + Manager 호출) / **safe local**(스펙 의도 명백, 구현하며 기록) / **observation**(후속 task 입력)

## 2026-07-27 §0 최상위 안전 불변식 검토 기록 (task 3.1)

구현 완료 시점에 §0 아홉 항목을 이 change의 실제 코드에 대조한 기록이다. 각 항목은
"어디서 그렇게 되는가"를 코드·테스트로 지목한다.

| §0 | 판정 | 근거 |
|---|---|---|
| 1 승인 없는 LIVE 주문 side-effect 금지 | 통과 | 편입은 주문을 내지 않는다. 편입 tx는 `position_adoptions` 행 + `positions.adoption_id` + `exit_states` 개설이 전부이고, 매도는 기존 exit 루프가 자체 주기에 Guardian 경유로만 낸다. 자동 테스트는 전부 fake/httptest이며 실 endpoint 접근 없음 |
| 2 OFF = upstream 동작 보존 | 통과 | `adoption.enabled` 기본 false, zero-value 안전. `TestAdoptionDefaultsOff`·`TestAdoptionStaysOffOnAGateOffEngine`. OFF에서 무관리 보유 알림은 그대로 발화한다(`TestTheUnmanagedAlertSurvivesAdoptionBeingOff`) — 이것이 landed 동작이므로 침묵이 오히려 회귀다 |
| 3 손절·비상 청산 즉시성 약화 금지 | 통과 | exit 판정 경로·주기·Guardian 경유 무변경. 편입은 exit 자격을 **넓히기만** 한다. flatten이 편입 포지션을 동일하게 덮는 것을 behavioural로 고정(`TestFlattenCoversAnAdoptedPosition`) — design A5의 "편입 포지션의 자동 매도를 멈추는 스위치는 의도적으로 없다"가 성립하려면 이 테스트가 필요하다 |
| 4 rate limit 예산 계상 | 통과 | 구동 루프 §0.4 계상을 `reconcileloop.go` 헤더에 고정: 정상 상태 주기(60초)당 수집 2회 × (미체결 1페이지 + holdings 1 + 통화 1) = 6콜, 편입 후보 발생 시 배치 시세 1콜 → 0.10~0.12 req/s. 상한은 MaxPages 50 → 2 × (50+1+1) + 1 = 103콜/분(1.7 req/s). 시세는 후보 전체 1콜(`TestAdoptionPricesEveryCandidateInOneCall`) |
| 5 운영 설정 변경 audit | 통과 | `audit.ActionAdoptionToggle`/`ActionAdoptionSetting`으로 enabled·default_stop_pct·exclude_symbols를 게이트 토글과 **분리**해 기록. 거부된 블록도 사유와 함께 기록(`TestARefusedAdoptionBlockIsAudited`) |
| 6 원장 스키마 additive-nullable + rollback | 통과 | v7 = `position_adoptions` 신규 테이블 + `positions.adoption_id` DEFAULT 없는 nullable ADD COLUMN. 기존 컬럼·CHECK 무변경, 과거 행 무재작성. 구버전 바이너리는 ErrSchemaTooNew 거부(`TestOlderBuildRefusesTheV7Journal`), 복구는 pre-migration 백업 |
| 7 운영 토글 flip 사람 승인 | 통과 | 코드 어디에서도 `adoption.enabled`를 쓰지 않는다 — config 파일이 유일한 입력이고 엔진은 읽기만 한다. 게이트 인터록도 무접촉이므로 실효는 사람이 게이트를 켜고 adoption을 켠 뒤에만 |
| 8 change scope 밖 주문·위험·원장 코드 무변경 | 통과 | 주문 경로(execgw·trading·official write) 무변경. 원장은 v7 additive만. `internal/console`은 task 2.7 범위(자격 표시)만 |
| 9 손절·익절·사이징 단방향 안전 | 통과 | 래칫·부분익절·사이징 규칙 자체는 한 줄도 바뀌지 않았다. 편입은 "보호가 없던 포지션에 보호를 붙이는" 방향뿐이고, 편입 해제(보호 제거)는 명시적 범위 밖(design A5) |

### A2 무발의 / A4 알림 존치 / A5 긴급 중지 서술의 테스트 지점

- **편입 tx 무발의(A2, SHALL NOT)**: `TestTheAdoptionTransactionProposesNothing` — 편입 직후
  exit_events에 OPENED 외의 action이 없고, pending proposal도 taken ratio도 움직이지 않는다.
- **첫 틱 발의는 정상(A2)**: `TestTheFirstObservationAfterAdoptionAppliesTheRatchetNormally` —
  합성 손절 하회 관측에서 발의 1건이 나오는 것이 정상임을 고정한다. 절대 무발의를 주장하지 않는다.
- **원가 무관(A2)**: `TestTheCostBasisDoesNotChangeTheFirstJudgement`가 원가 ±50%·동일·부재
  네 경우에서 같은 판정을 내는 것을 고정. `TestAnAdoptedWinnerRatchetsFromTheAdoptionPrice`가
  "보호 바닥은 편입일 가격"이라는 귀결을 수치로 고정한다.
- **알림 존치(A4, §0.2)**: `TestTheUnmanagedAlertSurvivesAdoptionBeingOff`,
  `TestAnExcludedSymbolIsReportedNotAdopted`, `TestACandidateWithNoQuoteIsDeferred`(편입 실패도
  알림). 무알림은 전이 상태뿐 — `TestAdoptionIsSilentUnderReconcile`,
  `TestAnUnstableAccountDefersSilently`.
- **긴급 중지(A5)**: `TestFlattenCoversAnAdoptedPosition`.

---

## 2026-07-27 [safe local] A7 교차 케이스의 기본값 — 부분 엔진 매도 + 외부 종결은 성과 행 없음

리뷰 라운드 5의 비차단 구현 노트를 그대로 구현하고 기록한다.

- 결정: 편입 포지션의 성과 동결은 **귀속된 엔진 매도 수량 == 편입 수량**일 때만 행을 만든다.
  엔진이 40%만 팔고 사용자가 잔량을 외부 처분한 경우 귀속 수량이 모자라므로 행은 없다.
  `TestAPartlyEngineSoldAdoptedPositionFreezesNothing`.
- 근거: 합성 전량 매수 leg에 부분 매도 leg을 맞추면 손익이 왜곡된다. 없는 행이 정직하다 —
  그 이야기는 ADJUSTMENT_CLOSED 이벤트가 기록한다.
- **부수 효과 하나가 설계상 유용하다**: 같은 규칙이 *과다* 귀속도 잡는다. flatten 체인은
  (계좌·시장·심볼) 범위이므로 원리상 이전 인스턴스의 flatten 매도가 섞일 수 있는데, 그러면
  귀속 수량이 편입 수량을 **초과**하므로 역시 행이 없다. 즉 귀속 모호성은 잘못된 행이 아니라
  행 부재로만 나타난다(라운드 5가 하중 조항으로 지목한 성질).
- 편입 후 외부 추가 매수(A8)로 투영 수량이 편입 수량을 넘긴 뒤 엔진이 전량 매도하면 같은
  이유로 행이 없다. fail-closed 방향이며 A8이 "재산정 없음"으로 동결을 유지한 것과 정합한다.

## 2026-07-27 [safe local] 매도 leg 귀속에 additive 스키마 보강은 불필요했다

- 라운드 4가 "귀속의 명시 참조가 부족하면 additive 보강 후 편차 기록"을 허용했으나, 기존
  선언 컬럼만으로 충분했다: exit 루프는 `exit_events.proposed_intent_id → mutation_attempts.intent_id
  → broker_order_id → fill_events`, flatten은 `flatten_sagas.account_ref → flatten_steps(LIQUIDATE,
  symbol, market) → intent_id → mutation_attempts → broker_order_id → fill_events`.
- flatten 체인이 인스턴스가 아니라 심볼 범위인 갭은 위 항목의 수량 일치 규칙이 fail-closed로
  덮는다. 스키마는 v7 A1 DDL 그대로 유지했다(추가 컬럼 없음).
- 시간창 매칭은 어느 arm에도 없다.

## 2026-07-27 [safe local] `SetFloat64(pct)`가 합성 손절을 66500 대신 66499.99999999999980으로 만든다

- 사실: `adoption.default_stop_pct`는 JSON에서 float64로 온다. `new(big.Rat).SetFloat64(0.05)`는
  0.05의 **이진 배정도 정확값**(0.05000000000000000277…)이라 `observed × (1 − pct)`가
  66499.9999999999998057…이 된다. 그대로 저장하면 합성 손절이 12자리 소수로 남고, 운영자가
  화면에서 보는 숫자와 원장이 다르다.
- 처리: `exitpolicy.SyntheticStop`이 pct를 `strconv.FormatFloat(pct,'f',-1,64)`로 최단 십진
  표현으로 되돌린 뒤 파싱한다 — 즉 "설정 파일에 적힌 숫자"로 읽는다. 가격 산술 자체는 여전히
  exact rational이며 float를 경유하지 않는다.
- 안전 영향: 없음(표현 정정). 방향성 있는 반올림은 formatPrice의 보호적 올림 그대로다.

## 2026-07-27 [safe local] 관측 시각은 응답이 아니라 **요청** 시점에서 잰다

- 스펙은 "편입 tx 직전의 시세 관측 · staleness ≤ 15s"다. 응답 수신 시각을 기준으로 재면
  Retrier 재시도로 쿼리 예산(8초)을 다 쓴 느린 읽기가 묵은 가격을 "0초 된 관측"으로 보고한다.
- 처리: `observeCandidates`가 `readAt`을 요청 **직전**에 찍는다. 보수 방향(더 늙게 셈)이고,
  느린 읽기는 편입 연기로 떨어진다. `TestAStalePriceDefersTheAdoption`이 실제 느린 읽기로
  이 경로를 탄다.

## 2026-07-27 [observation] exit 관측 루프의 `alertUnmanaged`는 전이 상태를 구분하지 않는다

- 사실: landed `ExitObserver.alertUnmanaged`(exitloop.go)는 포지션당 in-memory 래치로 첫
  관측에 무조건 발화한다. RECONCILE 중인 심볼도 예외가 아니다.
- design A4의 구현 노트가 "전이 상태 구분은 새 구동 루프가 함께 구현한다"고 지목했고, 이
  change는 그대로 구동 루프(`ReconcileDriver.judgeHoldings`)에 구분을 구현했다. exit 루프
  쪽은 손대지 않았다 — 그쪽은 tracker를 들고 있지 않고, 관측 주기(5초)마다 원장을 한 번 더
  읽는 비용이 붙는다.
- 잔여 영향: 두 루프가 동시에 도는 빌드에서 RECONCILE 중인 무관리 보유에 대해 exit 루프가
  한 번 더 알릴 수 있다. 과다 알림 방향이며(무알림 아님) 알림 키가 같아 outbox 중복은 아니다.
  또한 exit 루프는 인터록 조항 6 때문에 이 빌드에서 기동하지 않는다.
- 후속 입력: 2c에서 exit 루프가 실제로 도는 시점에 tracker를 주입해 같은 구분을 붙일지
  판단이 필요하다.

## 2026-07-27 [observation] `internal/config`가 `internal/exitpolicy`를 import한다

- 사실: `adoption.default_stop_pct` 범위(0.02 ≤ pct < 1)의 근거와 하한 provenance는
  `internal/exitpolicy/adoption.go`에 있고, config가 그 상수·검증 함수를 쓴다.
- 대안은 config에 숫자를 복제하는 것이었는데, 하한이 "관측 노이즈·왕복 비용 규모"라는
  **정책 근거**를 가진 값이라 두 벌로 두면 근거와 값이 갈라진다.
- 순환 위험 없음: `internal/exitpolicy`는 내부 패키지를 하나도 import하지 않는다(stdlib만).

## 2026-07-27 [observation] `migration_v6_test.go`를 v6에 고정했다

- 사실: v6 테스트 4개가 기본 마이그레이션 계획(`openTestJournalAt`)으로 열고 있어
  SchemaVersion이 7이 되자 "v5→v6 계약"이 조용히 "v5→head 계약"으로 바뀌며 백업 이름·버전
  단언이 깨졌다.
- 처리: `openJournalAtSchema(t, path, 6)`으로 고정했다 — `migration_v5_test.go`가 v4→v5를
  고정하는 방식과 동일하다. v6→v7은 새 파일 `migration_v7_test.go`가 맡는다.
- 골든 목록: `schema_test.go`의 wantTables·wantColumns에 `position_adoptions`와
  `positions.adoption_id` 추가, `core_domain_test.go`의 STRICT 목록에 `position_adoptions` 추가.
