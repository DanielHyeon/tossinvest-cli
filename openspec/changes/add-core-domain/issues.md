# Issues: add-core-domain

> 구현 중 발견한 스펙·코드 마찰 기록. advisory only — 권위는 spec + 코드 + 테스트.
> 분류: **blocking**(구현 중단 + Manager 호출) / **safe local**(스펙 의도 명백, 구현하며 기록) / **observation**(후속 task 입력)

## 2026-07-26 [safe local] D7 표가 nullability를 적지 않은 컬럼의 해석 (task 0.1)

- 사실: D7 표는 일부 컬럼에만 `NOT NULL`을 명시한다(`account_ref·market·symbol NOT NULL`,
  `entry_price·initial_stop·initial_risk TEXT NOT NULL`, `baseline_price NOT NULL`,
  `high_water NOT NULL`, `expected_prev_quantity TEXT NOT NULL`, `broker_as_of NOT NULL`).
  enum 컬럼은 `state CHECK(...)`, `kind CHECK(...)`, `mode CHECK(...)`,
  `actor CHECK(AUTO|OPERATOR)`, `policy_kind CHECK(...)`, `ratchet_level CHECK(...)`처럼
  CHECK만 적혀 있고, 기본값 컬럼은 `taken_ratio_total TEXT NOT NULL DEFAULT '0'`(NOT NULL 명시)와
  `completed INTEGER DEFAULT 0`(미명시)로 갈린다.
- 문제: SQLite에서 `CHECK (x IN (...))`는 x가 NULL이면 결과가 NULL이라 **제약을 통과한다**.
  자구 그대로 옮기면 `state`, `mode`, `actor`, `policy_kind`, `ratchet_level`이 전부 NULL을
  허용하고, 열거값 밖의 상태가 "NULL"이라는 이름으로 존재하게 된다. DEFAULT만 있는 컬럼도
  명시적 `NULL` 쓰기로 기본값을 무력화할 수 있다.
- 이번 처리: **v5 전사가 `risk_reservations.state`에 내린 판정을 그대로 승계**한다 —
  D9는 `state CHECK(HELD|RELEASED) DEFAULT HELD`였고 전사는
  `state TEXT NOT NULL DEFAULT 'HELD' CHECK (...)`였다(execution_contract.go:174).
  규칙: **v6 신규 테이블에서 enum CHECK를 가진 컬럼과 DEFAULT를 가진 컬럼은 NOT NULL**.
  적용 대상: `positions.state`, `position_adjustments.kind`, `operating_modes.mode`,
  `operating_modes.actor`, `exit_states.policy_kind`, `exit_states.ratchet_level`,
  `exit_states.completed`. 표의 열거·기본값 자구는 하나도 바꾸지 않았다.
- 검증: `TestPositionsStateCheck`(NULL state 거부), `TestExitStatesEnumsAndDefaults`
  (taken_ratio_total·completed NOT NULL), `TestOperatingModesConstraints`.

## 2026-07-26 [safe local] `exit_events`의 "id PK"를 INTEGER AUTOINCREMENT로 전사 (task 0.1)

- 사실: D7은 `position_adjustments`·`operating_modes`·`exit_events` 셋 다 "id PK"로만 적는다.
  journal의 기존 관례는 두 갈래다 — 호출자 발행 안정 PK(`intents.id`, `reconcile_states.id`)와
  append-only 이벤트 로그의 자동증가 정수(`attempt_transitions`, `fill_events`,
  `lineage_edges`).
- 이번 처리: `exit_events`만 `INTEGER PRIMARY KEY AUTOINCREMENT`. 근거는 그 테이블의 성격이다 —
  5초 주기 관측 판정의 순서 자체가 기록 대상이고, 호출자가 재시도로 되찾을 안정 id가 없다
  (`fill_events`와 같은 모양). `position_adjustments`(compare-and-append 재시도 인식)와
  `operating_modes`(전환 멱등)는 호출자 발행 TEXT PK로 남겼다.
- 후속 task 입력: **7.4**가 exit_events를 쓸 때 "같은 판정 재관측"의 중복은 PK가 막지 않는다 —
  중복 회피가 필요하면 판정 루프 쪽의 조건(레벨/워터마크 무변화 시 미기록)으로 처리해야 한다.

## 2026-07-26 [safe local] `position_adjustments`의 prev/new avg_price 표현 (task 0.1)

- 사실: D7은 "prev/new quantity·avg_price"라고만 적는다(4개 컬럼, nullability·기본값 미기재).
  외부 편입 포지션은 브로커가 취득단가를 주지 않을 수 있다.
- 이번 처리: 수량 쌍은 `TEXT NOT NULL`(수량 없는 수량 조정은 조정이 아니다), 단가 쌍은
  `TEXT NOT NULL DEFAULT ''` — `''`가 "미관측"이라는 규약은 바로 옆
  `fill_snapshots.average_price`(schemaV2)와 `execution_corrections.new_avg_price`(schemaV5)가
  이미 쓰는 것이다. NULL을 쓰지 않은 이유도 v5와 같다: NULL은 인덱스·비교에서 서로 구별되는
  값이라 "미관측"의 동치 비교를 조용히 깨뜨린다.

## 2026-07-26 [safe local] `exit_states.pending_intent_id`에 FK를 걸지 않았다 (task 0.1)

- 사실: D7은 `pending_intent_id TEXT`로만 적는다(FK 미기재). 반면 exit-policy 스펙의
  "제출 전 크래시" 시나리오는 **발의 기록 → 제출**의 순서를 요구한다.
- 이번 처리: FK 없음(자구대로). `intents(id)` FK를 걸면 intent가 아직 없는 시점에 pending을
  무장하는 것이 구조적으로 불가능해지고, 그 순서가 정확히 크래시 복원을 가능하게 하는
  순서다. 스키마 주석에 이 이유를 남겼다.

## 2026-07-26 [safe local] `operating_modes`의 "현재=최신 행"과 초 해상도 타임스탬프 (task 0.1)

- 사실: D7은 "append-only, 현재=최신 행"이라고 적고, journal의 타임스탬프 규약은 RFC3339
  **초 해상도**다(decision.go `formatJournalTime`). 한 초 안에 두 번 전환하면 `created_at`만으로는
  순서가 결정되지 않는다.
- 이번 처리: 스키마는 그대로 두고, 읽기 규약을 `(created_at, rowid)` 정렬로 못박아 컬럼 주석에
  남겼다(fills.go의 `ORDER BY a.recorded_at DESC, a.rowid DESC`와 같은 처리).
- 후속 task 입력: **3.1**의 현재 모드 조회는 반드시 `ORDER BY created_at DESC, rowid DESC LIMIT 1`.
  `created_at`만으로 정렬하면 보수 강화 직후의 완화가 먼저 읽힐 수 있다.

## 2026-07-26 [observation] `positions.opened_at`/`closed_at`은 nullable로 남겼다 (task 0.1)

- D7은 두 컬럼에 제약을 적지 않는다. `closed_at`은 정의상 CLOSED 전까지 NULL이고, `opened_at`은
  체결 전 상태(FLAT/OPENING)의 행이 존재할 수 있어 nullable로 두었다. 시각 컬럼의 허용 방향이라
  안전 제약을 약화하지 않는다. **6.1**이 상태기계를 구현할 때 "OPEN 이상은 opened_at NOT NULL"을
  코드 불변식으로 강제하면 된다(스키마 재작업 불필요 — 나중에 CHECK를 추가하려면 테이블 재작성이
  필요하므로 코드 쪽이 맞다).

## 2026-07-26 [safe local] 비용 override의 미지 키는 무시가 아니라 거부 (task 1.1)

- 사실: StockOS `costs.py:170-176` `_resolve_rate`는 env에 없는 키를 기본값으로 통과시키고,
  `tests/test_costs_env_override.py:57-63`은 "무관한 키가 모델을 흔들지 않는다"를 단언한다.
  env에는 수백 개의 무관한 키가 정상적으로 존재하므로 원본에서는 그것이 옳다.
- 문제: TossOS의 override는 전용 설정 블록이다. 인식되지 않는 키는 무관한 키가 아니라 **오타**이고,
  오타를 조용히 기본값으로 통과시키면 운영자가 설정했다고 믿는 값이 설정되지 않은 채 동작한다.
  §0.5(설정 변경 추적 가능)와 어긋난다.
- 이번 처리: 미지 키는 `ErrInvalidRate`로 거부(원본보다 엄격한 방향). 원본 케이스 2를
  `TestUnknownOverrideKeyIsRefused`로 **반전 이식**했고, 원본이 실제로 지키던 성질(모델이 모르는
  override가 모델을 흔들지 않는다)은 `TestNoOverridesReturnsDefaults` + 거부로 보존된다.
- 후속 task 입력: **4.x**가 설정 스키마를 배선할 때 미지 키 거부는 기동 실패로 이어진다
  (costs.py 모듈 docstring의 "Failure mode"와 같은 계약).

## 2026-07-26 [safe local] 실질 본전의 최소이익 하한은 native 통화 (task 1.1)

- 사실: `costs.py:269-284` `_native_profit_floor`는 KRW 최소이익 하한을 `usd_krw_rate`로 나눠
  USD로 환산하고, `tests/test_costs_env_override.py:263-283`이 그 환산을 단언한다.
- 문제: 통화 정규화는 `internal/riskcalc`의 소관이고 거기에는 환율 신선도 상한
  (`FXRateStaleness=60s`)과 "환율 없으면 암묵적 1:1 금지"가 이미 있다(riskcalc.convert).
  비용 모델 안에 상한 없는 두 번째 환산을 두면 stale 환율로 계산된 본전가가 조용히 통과한다.
- 이번 처리: `BreakEvenSellPriceWithFloor`의 하한은 **거래 자체 통화**로 받는다. 케이스 16은
  `TestUSBreakEvenAppliesNativeProfitFloorAndFXFee`로 이식 — 검증 대상 산술(하한 가산 + 매도측
  비율 그로스업)은 동일하고 환산 단계만 호출자 책임으로 옮겼다.

## 2026-07-26 [safe local] KRX 보드 차원(KOSPI/KOSDAQ/KONEX) 미이식 (task 1.1)

- 사실: `costs.py:76-78`은 보드별 거래세율을 따로 갖는다.
- 문제: TossOS의 시장 어휘는 `internal/clock.Market`("kr"/"us")이고 journal·intent·브로커
  클라이언트 어디에도 보드 차원이 없다. 보드를 구별할 수 없는 프로그램에 보드별 요율을 두면
  어느 요율이 쓰였는지 아무도 모른다.
- 이번 처리: KR 단일 요율을 보드 최대치 방향(과대 추정)으로 둔다. 요율은 전부 `[미검증]`이므로
  **2b 실측**이 이 결정을 함께 재검토한다. 보드를 구별할 입력이 생기면 그때 차원을 추가한다.

## 2026-07-26 [safe local] `migration_v5_test.go`가 head 버전을 따라다니고 있었다 (task 0.1)

- 사실: v5 전이 테스트 4건이 `openTestJournalAt`(= head까지 마이그레이션)을 썼다. SchemaVersion이
  6이 되는 순간 `TestMigrationV4ToV5PreservesEveryRow`는 v4→v6 테스트가 되고,
  `TestOlderBuildRefusesTheV5Journal`의 "두 버전 이름이 메시지에 있다" 단언과
  `TestMigrationBacksUpBeforeApplying`의 `v4-pre-v5` 파일명 단언은 그냥 깨진다.
- 이번 처리: 그 4건을 `openJournalAtSchema(t, path, 5)`로 고정해 **v4→v5 전이 테스트로 유지**하고,
  head 전이(v5→v6)는 `migration_v6_test.go`로 따로 세웠다. 백업 검사 헬퍼는
  `assertBackupAtVersion(backup, version, want, absentTable)`로 일반화해 두 파일이 공유한다.
  단언은 하나도 약해지지 않았다(같은 검사를 두 전이에 각각 건다).

## 2026-07-26 [safe local] 만료 sentinel은 `ErrInvalidRequest`를 감싼다 (task 0.2)

- 사실: 태스크는 DECISION_EXPIRED에 **신규 sentinel**을 요구한다(현재 만료는
  `checkDecisionReservable`이 `ErrInvalidRequest`로 낸다 — reservations.go). 그런데 landed
  `TestReservationRequiresARecordedUnexpiredDecision`은 만료가 `ErrInvalidRequest`임을 단언하고,
  단발 `Reserve`의 공개 계약도 그렇다.
- 이번 처리: `ErrDecisionExpired = fmt.Errorf("%w: decision expired", ErrInvalidRequest)`.
  `errors.Is(err, ErrInvalidRequest)`는 그대로 참이므로 기존 API·호출자·테스트는 하나도 약해지지
  않고, 새로 얻는 것은 "이 거부를 다른 invalid request와 구별할 수 있다"뿐이다 — 그게 정확히
  reason 매핑에 필요한 것이다.
- 검증: `TestIssuanceRefusesAnExpiredDecision`(두 sentinel 동시 만족),
  `TestSingleShotReserveStillRefusesTheExpiredDecision`(단발 API 무변경).

## 2026-07-26 [safe local] 재수집 변형을 함께 구현했다 — 그래야 reason 매핑이 도달 가능하다 (task 0.2)

- 사실: 태스크의 reason 매핑은 `SNAPSHOT_RECOLLECTION_EXHAUSTED←ErrRecollectionExhausted`를
  요구하는데, 그 오류는 재수집 루프에서만 나온다. `RecordDecisionAndReserve` 단발 API만으로는
  네 reason 중 하나가 구조적으로 도달 불가능하고, 도달 불가능한 매핑은 테스트할 수 없다.
- 이번 처리: `RecordDecisionAndReserveWithRecollection`(+`CollectIssue`)을 같이 추가했다.
  루프 본체는 landed `ReserveWithRecollection`에서 `recollectLoop[T]`로 추출해 **두 API가
  같은 루프를 공유**한다 — 재시도 가능 오류 집합·데드라인·소진=거부라는 세 성질이 사본 사이에서
  갈라지지 않게 하려는 것이다. `ReserveWithRecollection`의 동작은 무변경(기존 테스트 green).
- 부수 효과: 재수집 예산(10초)이 결정 TTL(60초)보다 짧다는 D1의 논증이 테스트로 고정된다 —
  `TestIssuanceWithRecollectionExpiresWithTheDecision`은 TTL을 예산보다 짧게 만들어 루프가
  결정보다 오래 살면 DECISION_EXPIRED로 끝남을 보인다.

## 2026-07-26 [observation] 원자 발급의 거부 순서는 precheck → 결정 삽입 → 예약 (task 0.2)

- 트랜잭션 안 순서: 스냅샷 신선도·원장 버전(결정 행이 필요 없는 거부) → 결정 INSERT →
  `checkDecisionReservable`(만료·계좌·1회) → 총계 → 예약 INSERT. 그래서 VERSION_CONFLICT는
  결정을 쓰기 전에, DECISION_EXPIRED·LIMIT_REACHED는 쓴 뒤에 롤백으로 거부된다.
- 후속 task 입력: **4.1** 발급자가 reason을 기록할 때 한 요청이 두 reason에 해당할 수
  있다(예: 만료된 결정 + 소진된 한도). 위 순서가 곧 우선순위이며, `IssueRefusalReason`의
  분기 순서(한도 → 만료 → 소진 → 버전)는 **오류 래핑** 우선순위라 서로 다르다 —
  `ErrRecollectionExhausted`가 마지막 stale/superseded를 감싸므로 버전보다 먼저 검사해야 한다.

## 2026-07-26 [safe local] apply hook의 범위는 "해소"이고 "무장"은 7.3의 몫이다 (task 0.3)

- 사실: D7의 괄호는 "taken_ratio·pending **해소**는 여기서만"이고, exit-policy 스펙의 발의
  무장(pending_action/level/intent_id 세팅)은 관측 판정 경로(별도 트랜잭션)에서 일어난다.
  태스크 0.3 문장("hook 밖에서 taken_ratio·pending을 쓸 수 없음")은 그보다 넓게 읽힌다.
- 이번 처리: `ApplyTx`는 `MoveTakenRatioTotal`·`ResolvePending`만 갖는다(무장 API 없음).
  결과적으로 **지금은 네 컬럼의 writer가 apply_hook.go 하나뿐**이고,
  `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook`이 다른 production 파일에서 컬럼 이름이
  등장하기만 해도 실패한다.
- 후속 task 입력: **7.3**이 무장 writer를 추가할 때는 그 함수도 `apply_hook.go`에 두어야 한다
  (다른 파일에 두면 layout 테스트가 실패한다 — 그게 의도된 강제 검토 지점이다). 무장은 fill
  트랜잭션이 아니라 발의 트랜잭션에서 일어나므로 `*Journal` 메서드 형태가 맞다.

## 2026-07-26 [safe local] `ApplyTx.Exec`은 투영 테이블용이고 guarded 컬럼은 거부한다 (task 0.3)

- 사실: 투영(6.1)은 `positions`·`position_adjustments`의 모양을 소유하고, journal이 그 상태기계를
  소유해서는 안 된다(position-ledger: journal이 도메인 상태기계를 직접 소유하는 것도 아니다).
  그래서 hook에는 일반 `Exec`가 필요하다. 그런데 일반 Exec가 있으면 hook이 그것으로
  `taken_ratio_total`을 쓸 수 있고, 그러면 layout 테스트가 볼 수 없는 writer가 생긴다.
- 이번 처리: `ApplyTx.Exec`는 쿼리 문자열에 guarded 컬럼 이름이 있으면 `ErrInvalidRequest`로
  거부한다(런타임). 같은 목록이 layout 테스트에도 있고, 둘은 한쪽이 바뀌면 사람이 대조하도록
  일부러 이중화했다. 문자열 검사라는 한계는 명시한다 — 우회하려면 SQL을 동적으로 조립해야 하고,
  그건 리뷰에서 보인다.

## 2026-07-26 [observation] hook은 거부·no-op 스냅샷에서 호출되지 않는다 (task 0.3)

- 계약: fail-closed 스냅샷(호출자 판정, 수량 감소, 역순 관측)과 바이트 동일 재관측에서는 hook이
  돌지 않는다. 반영된 것이 없으므로 적용할 것도 없고, 거부된 체결을 투영하는 hook은 원장이
  방금 믿지 않기로 한 값을 포지션에 쓰는 것이 된다.
- 반대로 delta 0이어도 **정정(EXECUTION_CORRECTION)과 terminal 전이**에서는 호출된다 — 수량은
  안 움직였지만 취득단가와 주문 종결성은 움직였고, 둘 다 포지션의 상태다.
- 후속 task 입력: **6.1**의 투영 함수는 `AppliedFill.Delta == "0"`인 호출을 정상 입력으로
  다뤄야 한다(정정·terminal).

## 2026-07-26 [safe local] 현금 부족 사유 코드를 하나로 합쳤다 (task 2.1)

- 사실: guardian.py:462-466은 `INSUFFICIENT_CASH`(원금 기준)와 `INSUFFICIENT_CASH_AFTER_COSTS`
  (비용 포함)를 분리한다. risk-management 이식 분류표는 "현금(INSUFFICIENT_CASH·비용 포함)"
  하나만 열거한다.
- 이번 처리: 스펙대로 단일 코드. 비용 포함 검사가 항상 더 엄격하므로 두 코드의 거부 집합은
  포함 관계이고, 코드를 둘로 두면 같은 사건이 원장에 두 이름으로 남는다. 어느 항목이
  모자랐는지는 Detail이 숫자로 싣는다("the entry needs 90090 KRW including estimated cost
  but 90000 KRW is available").
- 이식된 테스트: `test_guardian_requires_cash_for_order_plus_costs` →
  `TestCashIsMeasuredWithCostsIncluded`(같은 입력, 코드만 단일화).

## 2026-07-26 [safe local] 최소 RR provenance는 원본 lock의 **초기값**이다 (task 2.2)

- 사실: 스펙은 "provenance: StockOS parker_vwap §22 lock 2.0"이라고 적는데, 실제
  `strategies/parker_vwap/default_lock.py:35-38`의 **현재 값은 1.3**이다 — Plan 044가
  2.0→1.3으로 완화했고 그 이력이 같은 파일 :83에 남아 있다. 2.0은 그 lock의 초기값이고,
  현행으로 살아 있는 2.0은 `live_entry_contract.py:53`
  (`_DEFAULT_US_MIN_RR = Decimal("2.0")`)다.
- 이번 처리: 스펙의 2.0을 유지하되 코드 주석의 provenance를 정확히 적었다(초기 lock 값 +
  live 게이트 현행 기본값, 그리고 Plan 044의 완화를 따르지 않는 이유가 §0.9라는 것).
  원본의 현재 값을 따라가는 것은 손절·사이징 인접 파라미터의 완화 방향이라 금지다.

## 2026-07-26 [safe local] 크기 검사가 최소 RR보다 앞이라 a090 케이스 하나를 재작성했다 (task 2.2)

- 사실: TossOS 체인 순서는 "손절 계약 → 주문 크기 → 최소 RR → 현금"(risk-management 표)인데
  a090의 순서는 "stop → rr → grade → size"다. 그래서 a090
  `test_us_market_stop_rr_rungs_still_enforced`가 의존하는 성질("사이징이 막혀도 RR이 먼저
  보인다")은 TossOS에서 그대로 성립하지 않는다 — 통화 불일치로 사이징이 먼저 닫힌다.
- 이번 처리: 순서는 스펙 표가 권위이므로 바꾸지 않고, 케이스를 **의도 통화와 같은 통화의
  정책**으로 재작성해 이식했다(`TestForeignCurrencyIntentIsNotSizedAgainstADomesticBudget`
  후반부). 원본이 지키던 성질(stop·RR rung은 시장을 건너 동일)은 그대로 검증된다.
- 부수 기록: 크기 단계 **안**의 순서는 수량 → 위험예산 상한 → 설정 주문당 상한이다. 위험예산
  상한이 앞인 이유는 그것이 방금 검증한 손절에서 파생된 사이징 규칙이고, 설정 상한은 그
  바깥의 봉투이기 때문이다(docs/guardian-chain.md §1).

## 2026-07-26 [safe local] 스펙에 없는 두 기본값 — 위험예산과 비용 모델 부재 (task 2.1/2.2)

- 사실: 체인을 평가하려면 risk-management의 보수 기본값 6개(주문당 notional·수량·총 노출·
  일손실·자본비·통화) 외에 **per-trade 위험예산**이 필요하다. 스펙에 값이 없다.
- 이번 처리: `RiskBudget = MaxDailyLoss`로 **유도**했다. 측정값이 아니라, 일일 손실 한도가
  이미 함의하는 최대 per-trade 위험이다 — 즉 이 상한은 없던 허가를 만들지 않는다. 더 작게
  사이징하는 발급자는 자유롭고(§0.9 보수 방향), 더 크게 요구하는 의도는 거부된다.
  코드 주석에 "측정값 아님·유도값"으로 명시했다.
- 함께 처리: `costs.Model`의 zero value는 모든 요율이 0이라 **거래가 공짜로 보인다**(현금
  과소 계상 + 실질 본전이 진입가로 내려앉음). 요율 0은 합법적인 설정값이라 값으로는 구별할
  수 없으므로 `configured` 비트를 추가하고(`Model.Configured()`), 진입 경로 preflight가
  모델 부재를 `INPUT_UNAVAILABLE`로 거부한다. **위험 감소 경로는 이 질문을 하지 않는다** —
  "청산 비용을 계산할 수 없다"가 "그러니 들고 있어라"가 되면 §0.3 위반이다.

## 2026-07-26 [observation] `stop = "0"`은 STOP_MISSING이 아니라 INVALID_TARGET_STOP (task 2.2)

- 사실: a090 `test_stop_missing_rejected`는 `stop=0.0`을 `stop_missing`으로 판정한다.
  test_target_stop_contract의 `test_candidate_rejects_zero_stop`은 같은 값을 생성자
  ValueError로 막는다. 두 원본이 같은 입력에 다른 이름을 붙인다.
- 이번 처리: TossOS 입력은 decimal **문자열**이라 "부재"(`""`)와 "값은 있으나 가격이 아님"
  (`"0"`)이 구별된다. 부재 = `STOP_MISSING`, 비가격 = `INVALID_TARGET_STOP`. 구별의 실익은
  운영자에게 "신호가 아무것도 내지 않았다"와 "말이 안 되는 값을 냈다"를 알려주는 것이다.
  두 경우 다 거부이므로 안전 방향의 차이는 없다.

## 2026-07-26 [safe local] `AppliedFill`에 직전 스냅샷 쌍을 추가했다 (task 6.1)

- 사실: 투영의 원가는 주문 기여분 `누적체결 × 평균가`의 **변화**다. hook은 스냅샷 upsert
  **뒤에** 돌아서(fills.go 순서) 직전 값이 그 시점에 이미 사라져 있다.
- 문제: 직전 쌍 없이 쓸 수 있는 근사는 "델타 × 주문의 현재 평균가"뿐인데, 한 주문이 서로 다른
  가격으로 두 번 이상 체결되면 틀린다(50@100 후 50@110 → 현재평균 105, 한계원가는 5500인데
  근사는 5250). 취득단가는 실질 본전·실현 R의 분자라 §0.3 인접이다.
- 이번 처리: `AppliedFill`에 `PrevCumulativeQuantity`·`PrevAveragePrice`·`OrderedQuantity`
  3필드 **가산**(apply_hook.go 선언 + fills.go에서 `prev`·`obs.Quantity` 대입). 기존 필드·
  hook 계약·guarded 컬럼 규칙 무변경. 한 식이 체결·정정·terminal 세 경로를 모두 덮는다.
- 검증: `TestTheProjectionCarriesThePreviousSnapshotForItsCostBasis`(첫 관측은 빈 쌍,
  두 번째는 (5, 70000), 5@70000+5@80000 → 75000).

## 2026-07-26 [safe local] OPENING 종료는 **주문별** 원주문 수량이고 lineage가 합성한다 (task 6.1)

- 사실: 스펙은 "OPENING 종료 판단(원주문 수량)"이라고만 적는다. D7의 `positions`에는 추적 중인
  주문을 담을 컬럼이 없다(정정 교체 시 브로커가 새 주문번호를 준다 — lineage.go).
- 이번 처리: 완료 판정은 **체결 중인 주문 자신의** 주문수량 대비다. 정정 교체가 없으면 그것이
  곧 원주문이고, 있으면 부모는 체결분에서 `SUCCEEDED`(전이표의 lineage 차원)로 끝나고 자식의
  주문수량이 잔여라서 체인 전체가 원요청에서 정확히 완료된다. "기억된 원주문"으로 판정하려면
  스키마에 없는 주문 식별자를 투영이 들고 있어야 하고, 수량을 바꾼 정정에서는 그쪽이 틀린다.
- 부수 규칙: 주문수량이 `""`(미상)이면 완료를 판정할 수 없으므로 브로커가 terminal이라고 할
  때까지 `WORKING`이다 — 완료를 가정하면 아직 체결 중인 진입을 OPEN으로 닫는다.
- 검증: `TestOpeningCompletesAtTheOrderedQuantity`, `TestLineageSuccessionKeepsThePositionOpening`,
  `TestAnAmendmentKeepsOneInstance`(journal), `TestNoOrderedQuantityLeavesTheOrderWorking`.

## 2026-07-26 [safe local] 금지 전이는 투영자가 같은 트랜잭션에서 RECONCILE을 쓰고 체결은 커밋된다 (task 6.1)

- 사실: 스펙은 "허용되지 않은 전이는 오류이며 RECONCILE로 전이한다(산식 보정 금지)"인데,
  `positions.state` CHECK에 RECONCILE은 없다 — RECONCILE은 `reconcile_states`의 시스템 상태다.
- 문제: hook이 오류를 반환하면 체결 트랜잭션 전체가 롤백되어 **스냅샷이 전진하지 않고**
  다음 폴에서 같은 실패가 반복된다(체결 검출 영구 정지). 브로커가 알려준 사실을 잃는 쪽이
  그것과 불일치하는 쪽보다 나쁘다.
- 이번 처리: 투영은 수량·단가를 **동결**(산식 보정 금지)하고, 같은 apply tx 안에서
  `reconcile_states`에 심볼 스코프 행을 넣는다(활성 행이 있으면 no-op — `entered_at`이 폴마다
  전진하면 영원히 새 상태로 보인다). 체결 스냅샷·이벤트는 정상 커밋된다. 사유 매핑은 둘뿐이다:
  초과매도 = `QUANTITY_MISMATCH`(지역 수량과 브로커 체결의 문자 그대로의 불일치), 나머지 셋 =
  `ATTRIBUTION_FAILED`. hook에서 exported `*Journal` 메서드를 못 쓰므로(단일 커넥션, apply_hook.go
  규칙 4) `EnterReconcile`의 규칙을 handle 위에 다시 적었다.
- 후속 task 입력: **6.3**의 자동 해제(ADJUSTMENT_APPLIED)가 이 경로로 들어온 상태도 대상이다.
  조정만으로는 해제하지 않는다 — 해제 규칙은 6.3의 몫(`TestAnAdjustmentConvergesAFrozenProjection`이
  "블록은 그대로"를 단언하므로 6.3이 이 단언을 바꿔야 한다 = 사전 열거 대상 1건).

## 2026-07-26 [safe local] CLOSING 중 매수 체결(E31–E33)을 금지 전이로 판정했다 (task 6.1)

- 사실: 스펙이 표가 다뤄야 한다고 열거한 것은 즉시 전량체결·OPENING 종료·SCALING·lineage·
  매도 귀속·CLOSED 종결성이고, "청산 주문 진행 중 진입 체결"은 열거에 없다. 초과매도·귀속불가와
  달리 이것은 산술적 불가능이 아니라 **판정**이다.
- 이번 처리: 금지(RECONCILE). 근거는 §0.9 보수 방향 — 한 포지션에 상반된 두 지시가 동시에
  살아 있고 투영은 엔진이 아직 어느 쪽을 믿는지 알 수 없다. SCALING으로 처리하면 보호 경로가
  줄어드는 중이라고 믿는 동안 포지션이 커진다. 잃는 것은 없다(체결은 기록되고 수량 권위는
  계좌이며 6.2가 수렴시킨다). **수량이 움직이지 않는 delta 0 관측(정정·terminal)은 금지가
  아니다** — 금지는 수량이 움직였을 때만 발화한다(E13–E15는 허용).
- 후속 task 입력: **7.4**가 t0 하회 전량 청산을 발의할 때 작업 중인 진입 주문을 먼저 취소하지
  않으면 그 창에서 들어온 체결이 이 RECONCILE을 만든다. 취소가 7.x 설계에 들어가야 한다.

## 2026-07-26 [safe local] 미관측 평균가는 0이 아니고, 인스턴스 원가를 영구 미상으로 만든다 (task 6.1)

- 사실: `averageFilledPrice`는 nullable이고(openapi) journal의 규약은 `""`=미관측이다
  (fill_snapshots.average_price). D7의 `positions.avg_price`는 `TEXT NOT NULL`이라 `""`가 쓸 수 있다.
- 문제: 미관측을 0으로 접으면 취득단가가 과소 계상되고 → 실질 본전이 내려가고 → **손실 구간을
  본전으로 오인해 청산**한다. 방향이 fail-open이다.
- 이번 처리: 미관측 가격은 `position.Unknown`(`""`)이고, 한 번 미상이 된 인스턴스의 원가는
  수명 내내 미상으로 남는다(빠진 조각은 뒤 조각들의 평균으로 복구되지 않는다). 수량 투영은
  정상 진행한다 — 수량은 알고 있다. 읽는 쪽이 `""`에서 fail-closed 하는 것이 계약이다.
- 함께: `avg_price`는 투영에서 **유일하게 반올림되는 값**이다(원가 ÷ 수량은 종료하지 않을 수
  있다). 소수 12자리·half-away-from-zero(`big.Rat.FloatString`), 수량 산술은 riskcalc 정확 연산
  그대로. 근거와 누적 오차 상한은 `internal/position/decimal.go`의 `avgPriceScale` 주석.
- 검증: `TestAnUnpricedFillPoisonsTheCostBasis`, `TestAnUnpricedFillLeavesTheBasisUnknown`.

## 2026-07-26 [observation] 인스턴스 귀속은 (계좌·시장·심볼)의 최신 인스턴스다 (task 6.1)

- 투영은 체결을 그 심볼의 `instance_seq` 최대 행에 귀속시킨다. 스키마에 주문→인스턴스 링크가
  없고(D7 표에 없다) 추가하려면 v6를 다시 열어야 하기 때문이다.
- 한계: **닫힌 인스턴스에 속한 주문의 지연 정정**(CLOSED 후 새 인스턴스가 열린 뒤 도착하는
  EXECUTION_CORRECTION)은 현 인스턴스의 원가로 간다. 수량은 움직이지 않으므로 노출·손절 계약에는
  영향이 없고, 잘못될 수 있는 것은 취득단가다. 지금은 CLOSED 인스턴스에 대한 delta 0 관측이
  종결성 규칙으로 무시되므로(E16–E18, X16–X18) 오귀속은 "새 인스턴스가 이미 열려 있을 때"로
  한정된다.
- 후속 task 입력: **6.4**의 provenance 질의가 주문→인스턴스 조인을 필요로 하면 그때
  `fill_events`에 인스턴스 참조를 더하는 것이 자연스럽다(추가 컬럼 = 새 스키마 버전).

## 2026-07-26 [safe local] 체결 watermark는 심볼 단위이고 정정은 제외한다 (task 6.2)

- 사실: 스펙은 "체결 watermark의 불변을 재검증"이라고만 적고 범위를 정하지 않는다.
- 이번 처리: `fill_events`의 심볼별 `MAX(id)`. 계좌 전역으로 잡으면 아무 심볼에서 체결이 날
  때마다 무관한 심볼의 조정이 폐기되고, 조정 루프가 바쁜 계좌에서 영원히 수렴하지 못한다.
  `execution_corrections`는 **제외** — 정정은 취득단가만 움직이고 수량을 움직이지 않으므로
  수량에 관한 조정을 무효화할 수 없다.
- 기대 이전 값만으로 부족한 이유(= watermark가 존재하는 이유)를 테스트로 고정했다:
  수집과 커밋 사이에 같은 수량의 매수·매도가 들어오면 수량은 같고 세계는 다르다
  (`TestAMovedFillWatermarkIsDiscardedEvenWhenTheQuantityMatches`).

## 2026-07-26 [safe local] 조정 id는 파생값이라 재적용이 stale이 아니라 멱등이다 (task 6.2)

- 문제: 커밋 직후 크래시로 호출자가 결과를 못 받고 재시도하면, 수량은 **정당하게** 이동한 뒤라
  기대 이전 값 검사가 그 재시도를 stale로 판정한다 — 복구를 폐기로 오인한다.
- 이번 처리: id를 조정의 내용(스코프·kind·기대 이전 값·새 수량·새 단가·broker_as_of)에서
  파생하고, 트랜잭션 안에서 **비교보다 먼저** id 조회를 한다. 이미 있으면 저장된 행과 수렴된
  포지션을 `Applied=false`로 돌려준다. 조정 행과 포지션 수렴은 한 트랜잭션이므로 크래시는
  둘 다이거나 둘 다 아니다.
- 검증: `TestReapplyingTheSameAdjustmentIsANoOp`(행 1개 유지), `TestAnAdjustmentSurvivesARestart`.

## 2026-07-26 [safe local] MANUAL은 운영자 선언에서만 나오고, 조정은 CLOSED를 되살리지 않는다 (task 6.2)

- 분류(`position.Classify`): 운영자 선언 → MANUAL, 지역 인스턴스 없음 → EXTERNAL, 그 외 →
  UNKNOWN. "사람이 했을 것 같다"는 추측이고 `kind` 컬럼은 증거이므로 MANUAL은 선언에서만 나온다.
  UNKNOWN은 "안 봤다"가 아니라 "귀속에 실패했다"는 판정이다(core_domain.go 주석과 동일).
- 수렴 상태 규칙: 새 수량 0 → CLOSED(+`closed_at`), 양수인데 현재 상태가 살아 있으면
  **상태 유지**(계좌 스냅샷은 작업 중인 주문의 존재를 모른다), 없거나 CLOSED면 OPEN.
- CLOSED 인스턴스에 계좌가 수량을 보고하면 되살리지 않고 **다음 인스턴스**를 연다
  (`entry_decision_id` NULL — exit 대상 아님). CLOSED 종결성과 외부 편입이 같은 규칙으로 만난다.
- 검증: `TestAnExternalHoldingOnAClosedSymbolOpensTheNextInstance`,
  `TestClassifyNamesTheProvenance`, `TestAnAdjustmentToZeroClosesTheInstance`.

## 2026-07-26 [safe local] 집계 입력의 음수는 비교하지 않고 거부 (task 2.3)

- 사실: 2.1/2.2 커밋이 2.3의 검사(크기·현금·allowlist·재진입·총계·중복)까지 함께 실장했고,
  스펙 표와 대조한 결과 순서·경계·사유 코드는 전부 일치했다. 2.3에서 새로 필요한 코드는
  하나뿐이었다.
- 문제: `AccountState.OpenExposure`·`DailyRealizedLoss`는 생산자 계약상 **크기**다
  (`riskcalc.DailyLoss`는 `loss := 0.0; if net < 0 { loss = -net }`, aggregate.go:252-256).
  그런데 체인은 그것을 검사하지 않고 `WithinLimit`에 넘긴다. 호출자가 부호 있는 손익을
  넘기면 10만원 잃은 날이 `-100000`으로 도착하고, 절대·비율 두 비교가 모두 "한도 여유"로
  읽는다 — 이 체인에서 유일하게 게이트를 여는 방향의 입력 오류다.
- 이번 처리: `magnitudeIn`(moneyIn + 음수 거부)을 두 집계 입력에 적용, `INPUT_UNAVAILABLE`.
  0은 통과한다(개방 없음·손실 없음이 정상 케이스). 경계 규칙표(docs/guardian-chain.md §2)에
  행 추가.
- 검증: `TestNegativeAggregatesAreRefusedRatherThanCompared`.

## 2026-07-26 [safe local] 진입 latch는 execgw가 생산하고 체인은 소비만 한다 (task 2.4)

- 사실: 태스크는 "기존 EntryGate 사유 매핑, 중복 판정 없음"을 요구한다. 스펙이 부르는 네 조건
  (401/403·SLO·RECONCILE·recovery)을 `internal/risk`가 다시 유도하면 게이트 규칙의 사본이
  생기고, 사본은 해제 규칙이 사유마다 다르기 때문에(자동 해제/운영자 전용/심볼 범위) 정확히
  중요한 날에 갈라진다.
- 이번 처리: `execgw.(*EntryGate).EntryLatchFor(market, symbol)`가 `CheckEntryFor` — 봉인된
  제출 시퀀스가 부르는 그 호출 — 의 답을 체인의 두 평문 값으로 옮긴다. `internal/risk`는
  execgw를 import하지 않는다(그러면 journal이 체인 뒤로 딸려 들어온다).
- 파일 배치 편차: 태스크는 `internal/execgw/retry.go`를 지목했으나 신규
  `internal/execgw/entrylatch.go`에 두었다 — retry.go는 재시도 매트릭스 파일이고, 2.4의
  비중복 논거를 그 파일 헤더에 섞으면 둘 다 읽기 어려워진다. 3.1의 투영도 같은 이유로
  `modegate.go`.
- **열거 밖 사유도 전달한다**: 체인이 게이트 사유의 부분집합만 보면 Guardian이 허용한 의도를
  Gateway가 곧바로 거부하고, 그 사이에 결정·예약 왕복이 낀다. 신선도·flatten·IN_DOUBT·
  알림 미전달도 같은 값으로 도착한다.
- 검증: `TestEveryEntryBlockingConditionReachesTheChain`(네 조건을 production 경로로 발생),
  `TestTheLatchSurfaceNeverDisagreesWithTheGate`(2^5 조합 — 사유·detail이 게이트와 동일),
  `TestTheChainNeverRederivesTheLatch`.

## 2026-07-26 [safe local] 모드 전환은 journal이 소유하고 게이트가 구현한다 (task 3.1)

- 사실: 스펙은 "모드 전환 = journal 영속 + EntryGate 계좌 latch 투영"을 SHALL로 적는다.
  두 패키지에 걸친 요구라 소유자를 정해야 했고, 방향은 이미 정해져 있다(execgw → journal).
- 이번 처리: journal이 `ModeProjector` 인터페이스(1메서드)를 들고 `TransitionOperatingMode`가
  커밋 후 호출한다. `*execgw.EntryGate`가 구현체다. 공개 API에 "투영 없이 append"도
  "append 없이 투영"도 없다 — 전자는 게이트가 모르는 강화를, 후자는 재시작이 복원할 수 없는
  차단을 남긴다. 2a의 `ReservationAuditor` 인터페이스와 같은 모양이다.
- **순서는 커밋 → 투영**이고 근거는 완화 방향이다: 투영이 먼저면 근거 행이 durable해지기 전에
  latch가 풀리고, 그 사이 크래시는 "게이트는 열렸는데 journal은 ENTRY_BLOCKED"로 복귀한다.
  반대 방향(강화 후 미투영)은 한 프로세스 안 마이크로초 창이고 `RestoreOperatingModeProjection`이
  기동 때 닫는다.
- `latchOrder` 말미에 추가(관례: append). 앞의 사유들은 운영자가 고칠 구체적 결함이고 모드는
  대개 그 결과이므로, 원인을 먼저 보여주는 것이 조치 가능한 순서이기도 하다. 체인은 어차피
  모드 단계(2)가 latch 단계(3)보다 앞이라 `OPERATING_MODE_BLOCKED`로 보고한다.
- 검증: `TestAnAutomaticTighteningRefusesTheNextPlace`(게이트웨이 경유),
  `TestTheRestartRebuildsTheModeLatch`, `TestModeClassTableIsComplete`(9셀 + fail-closed 4).

## 2026-07-26 [safe local] 트리거 목표가 현재 모드보다 느슨하면 오류가 아니라 no-op (task 3.1/3.2)

- 사실: "동시 적용 시 보수 우선(SHALL)"과 "자동 강화는 즉시(SHALL)"가 한 요청에서 만나는
  경우가 있다 — 운영자가 HALT_ALL을 건 계좌에서 일손실 트리거(목표 ENTRY_BLOCKED)가 발화.
- 이번 처리: 오류가 아니라 **no-op**(행 없음·투영 없음·알림 없음·`changed=false`). 오류로 하면
  5초 주기 트리거가 매 주기 오류를 내고 호출자마다 특례가 생긴다. 반대로 AUTO의 `NORMAL` 지정과
  `HALT_ALL` 지정은 요청 자체가 틀린 것이라 오류다(`ErrModeRelaxationRequiresOperator`,
  `ErrHaltAllIsNeverAutomatic`).
- **부수 효과가 본질적이다**: 이 no-op이 알림 되먹임 고리를 닫는다. 강화 → critical 알림 →
  전달 실패 → (트리거) critical outbox 실패 → 강화? 이미 ENTRY_BLOCKED라 no-op이고, 전환이
  없으므로 알림도 없다. `TestTheAlertEscalationLoopTerminates`가 이것을 고정한다.

## 2026-07-26 [safe local] 완화의 승인은 `cause`에 실어 durable하게 남긴다 (task 3.2)

- 사실: 스펙은 완화에 "사람 승인(§0.7) + audit"을 요구한다. D7의 `operating_modes` 표에는
  승인 컬럼이 없고, 0.1 스키마는 이 change에서 재작업 대상이 아니다.
- 이번 처리: 요청에 `Approval` 필드를 두고 완화 시 필수로 만든 뒤, 저장은
  `cause + " | approved-by: " + approval`로 한다 — `reconcile_states`의 해제 evidence가 쓰는
  것과 같은 관용구(`evidence + " | released: " + evidence`). audit 줄은 커밋 **전에** 쓰고,
  실패하면 전환 전체가 롤백된다(2a `OperatorReleaseReservation`과 같은 순서·같은 이유).
- 결과: journal 단독으로 "누가 승인했나"에 답할 수 있고, audit 파일은 전/후 값·주체·시각을
  따로 갖는다.
- 검증: `TestTheApprovedRelaxationIsAuditedWithBothValues`,
  `TestAFailedAuditWriteAbortsTheRelaxation`, `TestRelaxationRequiresAnApproval`.

## 2026-07-26 [observation] 자동 강화 트리거의 **생산자** 배선은 이 태스크 밖 (task 3.2 → 4.x·7.4)

- 3.2가 실장한 것은 트리거→목적 상태의 닫힌 열거와 전환 API(`EscalateOperatingMode`)다.
  실제로 그것을 부르는 쪽은 아직 없다:

  | 트리거 | 생산자 | 어느 task |
  |---|---|---|
  | `DAILY_LOSS_LIMIT_REACHED` | 발급자(체인이 DAILY_LOSS_LIMIT_REACHED로 거부한 뒤) | 4.x |
  | `BROKER_AUTH_REJECTED` | `execgw.Retrier`(현재는 게이트 latch만) | 4.x |
  | `CRITICAL_ALERT_UNDELIVERED` | `obs.Notifier.deliver`(현재는 게이트 latch만) | 4.x |
  | `EXIT_OBSERVATION_OUTAGE` | 판정 루프 60초 두절 | 7.4 |

- 지금 넷 중 둘만 배선하면 절반만 동작하는 상태로 남고, 두 생산자가 journal 핸들을 새로
  가져야 한다(엔진 배선 문제). **네 개를 한 번에 배선하는 것이 맞고 그 자리는 엔진 배선
  태스크다.** 현재도 안전한 이유: 401/403과 알림 미전달은 이미 EntryGate latch로 진입을
  막는다 — 모드 강화는 그 위에 "재시작을 건너 살아남는 영속" 층을 더한다.

## 2026-07-26 [safe local] `EventOperatingMode`를 critical로 승급 (task 3.3)

- 사실: `obs/event.go`는 이 이벤트를 "Phase 2 예약"으로 선언만 해뒀고 등급표에 없었다
  (미지 이벤트 = normal).
- 이번 처리: `criticalEvents`에 추가. 양방향 모두 사람을 깨울 사건이다 — 강화는 엔진이 스스로
  진입을 멈췄다는 뜻이고(네 트리거 전부 사람이 조치해야 풀린다), 완화는 라이브 계좌에서
  누군가 진입을 다시 켰다는 뜻이다(§0.7이 두 번째 눈을 원하는 바로 그 변경).
- critical이므로 outbox에 먼저 durable하게 쓰인다 — 전환을 알리고 죽은 프로세스의 알림도
  남는다. 되먹임 고리는 위 항목의 no-op 규칙이 닫는다.

## 2026-07-26 [safe local] 발급자를 `internal/execgw`에 두었다 (task 4.1)

- 사실: 태스크는 "`internal/risk` issuer type or new file"이라 적는다. 그런데 발급자는 셋을 동시에
  필요로 한다 — 체인(`internal/risk`), 원자 발급 API(`internal/journal`), 결정 참조·TTL·한도
  스냅샷(`internal/execgw`).
- 문제: `internal/risk`에 두면 그 패키지가 journal을 import한다. 2.4 판정이 세운 성질
  ("`internal/risk`는 execgw를 import하지 않는다 — 그러면 journal이 체인 뒤로 딸려 들어온다")이
  깨지고, 체인의 순수성(값만 받는 함수)이 패키지 의존성 수준에서 사라진다.
- 이번 처리: `internal/execgw/riskguardian.go`. 의존 방향은 이미 execgw → journal이고
  execgw → risk를 더해도 순환이 없다(risk는 costs·riskcalc·clock만 import). 결정 계약을 소유한
  패키지가 그 발급자도 소유하는 배치이고, 2.4의 파일 배치 편차와 같은 성격이다.
- 검증: `internal/risk`의 import 목록 무변경(chain.go에 `EntryExposureValue` 1개 추가만).

## 2026-07-26 [safe local] 진입 예약은 OPEN_EXPOSURE 하나다 (task 4.1)

- 사실: 예약 kind는 셋이다(`OPEN_EXPOSURE`·`DAILY_LOSS`·`CASH`, execution_contract.go:67-69).
  스펙은 "총계 한도의 최종 권위는 예약 트랜잭션"이라고만 적고 어느 kind를 거는지는 적지 않는다.
- 이번 처리: 진입은 `OPEN_EXPOSURE` 하나만 건다. 근거는 kind별로 다르다:
  - `DAILY_LOSS`는 **실현** 손실이다. 진입은 그것을 움직이지 않으므로 예약할 것이 없고,
    "이 거래가 질 것"을 전제로 손실 한도를 미리 깎는 것은 스펙에 없는 신규 정책이다.
  - `CASH`는 설정 한도가 아니라 브로커 사실이다. 예약의 한도 비교는 도달=차단(≥)인데 체인의 현금
    검사는 포함 상한(정확히 커버하면 통과)이라, 가용현금을 "한도"로 넘기면 두 규칙이 경계에서
    어긋난다. 현금은 체인이 비용 포함으로 이미 검사한다.
- 예약 금액은 `risk.EntryExposureValue`(지정가 × 수량 + 과대 추정 비용) — 체인의
  `checkOpenExposure`가 방금 여유를 확인한 바로 그 수이고, 그래서 새 export가 필요했다.
  같은 수를 호출자가 다시 계산하면 경계에서 갈라진다.
- 후속 task 입력: **8.1**이 실현 손실을 기록하면 `DAILY_LOSS` 예약의 생산자는 그쪽이다.

## 2026-07-26 [observation] VERSION_CONFLICT는 발급자 경로에서 도달 불가 (task 4.1)

- 사실: 태스크는 네 reason이 전부 표면화되는 테스트를 요구한다. 그런데 발급자는 항상 재수집 루프
  (`RecordDecisionAndReserveWithRecollection`)를 쓰고, `recollectLoop`는 stale·superseded를
  마지막에 `ErrRecollectionExhausted`로 감싼다(0.2 판정: "내부 재시도로 stale·superseded는
  종단에서 여기 수렴"). 그래서 발급자에서 나오는 버전 경합은 전부
  `SNAPSHOT_RECOLLECTION_EXHAUSTED`다.
- 이번 처리: 도달 가능한 셋(LIMIT_REACHED·SNAPSHOT_RECOLLECTION_EXHAUSTED·DECISION_EXPIRED)을
  end-to-end로 고정하고, VERSION_CONFLICT는 단발 `Reserve`의 공개 계약으로 남긴다(journal의
  `IssueRefusalReason` 테스트가 매핑을 고정). 테스트 주석에 도달 불가의 이유를 적었다 —
  "테스트가 없다"와 "구조적으로 일어날 수 없다"는 다른 사실이다.

## Manager 판정 (1차 물결 검증, 2026-07-26)

- **독립 재실행**: `go test ./... -race -count=1` 0 FAIL (1947 tests, 43 pkgs). tasks.md worktree의 미커밋 unchecking은 에이전트 경합 잔재로 확인·폐기(HEAD 정확).
- **0.x 편차 10건 승인**: nonce 재사용 가능성으로 고아 없음 증명(더 강함), 재수집 발급 변형(EXHAUSTED 도달 가능성 확보), apply hook 3중 강제(런타임·AST 도달성·파일 배치 — 7.3 arming을 apply_hook.go로 강제하는 배치 규칙 포함), ErrDecisionExpired가 ErrInvalidRequest를 wrap(기존 계약 비약화), operating_modes 최신 행 정렬 (created_at, rowid) → 3.1 입력, hooks가 delta-0 정정·terminal 전이에도 발화 → 6.1 입력.
- **1.x·2.x 판정 8건 승인**: unknown override key **거부**(원본은 무시 — 오타가 조용히 기본값이 되는 것보다 낫다, 의도적 역전 승인), zero-value 비용 모델의 무료 취급 차단(configured 비트 — 진입 거부·감소 경로 비대상, 스스로 찾은 fail-open), math/big 유리수(수량 floor·RR 경계 — 이진 전개가 아니라 규칙의 경계), INSUFFICIENT_CASH 병합, size-before-RR 순서(스펙 순서 준수 — a090 재작성 명시), stop="0"→INVALID_TARGET_STOP.
- **min-RR provenance 정정**: 스펙 인용을 live_entry_contract.py:53으로 교체(default_lock은 Plan 044에서 1.3 완화 — §0.9상 추종 안 함, 2.0 유지). 값 무변경.
- 이식 대조: costs 20/20, guardian 8+4대체/8제외, target_stop 13/16제외(P3 신호층), a090 15/21제외 — 제외 사유 전수 파일 헤더 기록 확인.

## Manager 판정 (2차 물결 검증, 2026-07-26)

- 독립 재실행 0 FAIL(2084). C·E 판정 16건 전부 승인 — 특히: 집계 입력 부호 fail-open 봉합(magnitudeIn), commit-then-project 순서(완화 방향 기준 선택·강화 창은 Restore로 폐쇄), 보수 우선 no-op(알림 실패 피드백 루프 차단), latch 전체 통과(부분 뷰였다면 Guardian 허용→Gateway 즉시 거부의 왕복 낭비), E31 ENTRY_WHILE_CLOSING 거부(§0.9), 거부 시 detector 비정지.
- 승계 확정: ProjectPosition SetApplyHooks 바인딩 → 3차 F(엔진 배선 영역), Exit applier 바인딩 → 4차 7.x. 자동 강화 트리거 생산자 배선(일손실·401/403·critical outbox) → 3차 F, exit 관측 두절 → 4차 7.4. 6.3의 사전 열거 단언 1건(조정 후 차단 유지 → ADJUSTMENT_APPLIED 해제 반전) → 3차 G. 주문→인스턴스 링크 → G의 6.4 검토. 7.x 입력: 전량 청산 발의 전 working 진입 취소 선행(E31 회피).
