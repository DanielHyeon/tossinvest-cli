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
