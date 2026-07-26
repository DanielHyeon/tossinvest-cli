# Issues: extend-execution-contract

> 구현 중 발견한 스펙·코드 마찰 기록. advisory only — 권위는 spec + 코드 + 테스트.
> 분류: **blocking**(구현 중단 + Manager 호출) / **safe local**(스펙 의도 명백, 구현하며 기록) / **observation**(후속 task 입력)

## 2026-07-26 [safe local] D9의 `mutation_attempts` UNIQUE에 필요한 `account_ref`가 그 테이블에 없다 (task 0.1)

- 사실: D9는 `UNIQUE(account_ref, client_order_id) WHERE client_order_id IS NOT NULL`을
  **`mutation_attempts`에** 건다. 그런데 v1 스키마에서 `account_ref`는 `intents`에만 있고
  `mutation_attempts`는 `intent_id` FK로만 계좌에 닿는다(schema.go:63, 80-94). SQLite 인덱스는
  FK를 타고 다른 테이블 컬럼을 참조할 수 없으므로, 표의 자구 그대로는 인덱스가 성립하지 않는다.
- 선택지: (a) `mutation_attempts`에 `account_ref`를 additive-nullable로 추가하고 D9 표대로
  인덱스를 건다, (b) UNIQUE를 `decisions(account_ref, client_order_id)`로 옮긴다 — 그러면
  "한 attempt가 키를 점유한다"가 아니라 "한 결정이 키를 발급한다"가 되어 재발급·재생 시
  attempt 수준 충돌을 못 잡는다, (c) 트리거로 intents를 조회해 강제 — STRICT·단일 커넥션
  전제에서 비용·복잡도 대비 이득 없음.
- 이번 처리: **(a)**. `ALTER TABLE mutation_attempts ADD COLUMN account_ref TEXT`(nullable,
  v4 행은 NULL로 유효) + D9 표 그대로의 partial UNIQUE. D9 의도(키는 계좌당 attempt 하나가
  점유)를 자구 변경 없이 성립시키는 최소 추가다.
- 후속 task 입력: **1.3**(PrepareRequest 확장)이 이 컬럼을 채워야 한다 — 값은
  `intents.account_ref`와 동일해야 하며, 불일치는 결정↔attempt 결속 검증에서 거부 대상이다.
  (0.1은 스키마만 전사하므로 이 일관성 검증은 1.3의 몫이다.)

## 2026-07-26 [safe local] `reconcile_states` 부분 unique가 계좌 전역 스코프를 구속하지 못한다 (task 0.1)

- 사실: D9는 `활성 상태 부분 unique(account_ref, symbol) WHERE released_at IS NULL`이고
  `symbol NULL = 계좌 전역`이다. SQLite는 UNIQUE 인덱스에서 NULL을 서로 구별되는 값으로
  취급하므로, 이 인덱스 하나만으로는 **계좌 전역 활성 행이 몇 개든 들어간다**. "활성 상태는
  스코프당 하나"라는 D9 문장이 정확히 계좌 전역에서만 무효가 된다.
- 이번 처리: D9의 인덱스를 자구 그대로 만들고, 계좌 전역 스코프 전용 부분 unique를 하나 더
  추가했다 — `UNIQUE(account_ref) WHERE released_at IS NULL AND symbol IS NULL`. 두 인덱스의
  합이 D9가 쓴 규칙과 같다. 스코프를 좁히지도 넓히지도 않는 순수 보강.
- 후속 task 입력: **4.1**은 계좌 전역 RECONCILE 진입이 유일함을 인덱스가 보장한다고 가정해도
  된다. 단 "이미 활성이면 재진입 no-op"은 여전히 코드가 판단해야 한다(제약은 거부만 한다).

## 2026-07-26 [observation] D9가 `reconcile_states.cause` enum 값을 열거하지 않아 CHECK를 걸지 않았다 (task 0.1)

- 사실: D9는 `cause enum NOT NULL`이라고만 쓰고 값 목록을 주지 않는다. 다른 enum
  (safety_class·preimage_kind·kind·state·release_reason)은 전부 목록이 표에 있다.
- 판단: SQLite에서 CHECK는 테이블 재작성 없이는 완화할 수 없고, 재작성은 schema.go의 additive
  규칙(컬럼 drop/rename 금지, 이력 재작성 금지)이 금지한다. 값 목록을 추측해서 CHECK를 걸면
  4.1/5.1이 새 원인을 발견했을 때 v6 재작성이 강제된다. 반대 방향(제약 없음 → 나중에 추가)은
  가능하지만 그것도 재작성이 필요하다 — 그래도 **잘못된 목록을 지금 굳히는 쪽이 더 비싸다**.
- 이번 처리: 컬럼은 `TEXT NOT NULL`, CHECK 없음. D6·D7에서 도출한 값은 Go 상수로 고정
  (`ReconcileCause*` — SNAPSHOT_UNAVAILABLE / SNAPSHOT_STALE / QUANTITY_MISMATCH /
  IDENTIFIER_CONFLICT / ATTRIBUTION_FAILED). 검증은 쓰기 경로가 한다.
- 후속 task 입력: **4.1**이 원인 목록을 확정하면, 그때 CHECK를 걸지(=v6 재작성 비용) 쓰기
  경로 검증으로 둘지 Manager 판단이 필요하다. `release_cause`도 같은 상태다.

## 2026-07-26 [safe local] UNIQUE 키에 들어가는 컬럼을 NOT NULL로 굳혔다 (task 0.1)

- 사실: D9 `execution_corrections`의 `UNIQUE(order_id, cumulative_qty, new_avg_price,
  new_filled_amount)`에는 "crash 재관측의 이중 삽입 방지"라는 목적이 명시돼 있다. 그런데
  `new_avg_price`/`new_filled_amount`를 nullable로 두면 SQLite의 NULL-구별 규칙 때문에
  **브로커가 평균가를 주지 않은 경우에 정확히 그 방지가 꺼진다**.
- 이번 처리: 두 컬럼만 `TEXT NOT NULL DEFAULT ''`. 미관측은 빈 문자열로 표기하며, 이는 바로
  옆 `fill_snapshots.average_price`가 이미 쓰는 규약이다(fills.go:600). `prev_*`는 D9대로
  nullable 유지(첫 관측에는 prev가 없다).
- 같은 이유로 `fill_snapshots.filled_amount`도 `TEXT NOT NULL DEFAULT ''`로 추가했다. v4 행은
  마이그레이션 후 ""(미관측)으로 읽히며, 5.3의 prev 비교가 NULL 특례 없이 성립한다.
- 후속 task 입력: **5.3**은 정정 이벤트를 쓸 때 미관측 값을 NULL이 아니라 ""로 기록해야 한다.

## 2026-07-26 [observation] `spent_nonces.decision_id`에 FK를 걸지 않았다 (task 0.1)

- 사실: D9는 `risk_reservations`·`execution_corrections`에는 `REFERENCES`를 명시하고
  `spent_nonces`에는 명시하지 않는다. 동시에 "보존 기간 ≥ 최대 결정 TTL — 어떤 결정도 자기
  소비 기록보다 오래 살 수 없다"는 보존 불변식을 요구한다.
- 판단: FK를 걸면 만료 결정 정리(prune)가 소비 기록에 막히거나 cascade로 소비 기록을 지우게
  된다 — 후자는 불변식을 정면으로 깬다. D9의 FK 생략은 의도로 읽었다.
- 이번 처리: `decision_id TEXT`(FK 없음, D9 자구대로 NOT NULL도 없음). 값은 쓰기 경로가 항상
  채운다.
- 후속 task 입력: **1.7**(NonceStore)이 보존 스윕을 구현할 때 이 불변식을 테스트로 고정해야
  한다.

## 2026-07-26 [observation] v2 마이그레이션 회귀 테스트의 픽스처 방식이 v5에서 더 이상 성립하지 않는다 (task 0.2)

- 사실: `TestOutboxSurvivesTheV2Migration`(outbox_test.go)은 "현재 스키마로 연 뒤 나중 버전이
  추가한 **테이블을 DROP**하고 user_version을 2로 되돌린다"로 구버전 DB를 흉내 냈다. v5는
  테이블만이 아니라 **컬럼도** 추가하므로(ALTER는 DROP TABLE로 되돌지 않는다) 이 픽스처는
  "duplicate column name"으로 깨진다.
- 이번 처리: 픽스처를 `Options.migrationOverride`(비공개 테스트 전용 필드)로 교체해 **실제
  v2 마이그레이션까지만 적용된 파일**을 만든다. 흉내가 아니라 진짜 v2 DB이므로 이전보다
  강한 픽스처다. **단언은 한 줄도 바꾸지 않았다**(버전 전이 + 알림·flatten 쓰기 가능).
- 8.2 참고: 이 파일은 "기존 단언 변경 사전 열거" 목록에 없지만, 변경된 것은 픽스처 구성이며
  단언은 동일하다.

## 2026-07-26 [observation] opaque 원문 저장 위반이 D7이 열거한 3개소 밖에 3곳 더 있다 (task 5.2)

- 사실: D7은 "위반 3개소"로 `execgw/classify.go:149`·`journal/resolution.go:42·47·126`·
  `execgw/indoubt.go:512·516`을 지목했고 5.2는 그 셋을 고쳤다. 같은 규칙("저장은 수신 원문
  그대로, 비교는 바이트 동일")을 어기는 곳이 세 군데 더 있다:
  - `journal/lineage.go:118` — `ResolveConfirmedWithLineage`가 attempt의 broker_order_id를
    `TrimSpace`해서 쓴다. **방금 원문 저장으로 고친 `resolution.go`의 `ResolveConfirmed`와
    같은 해소 쓰기 경로인데 파일이 달라 D7의 열거에서 빠졌다.** 그 결과 취소 해소는 원문,
    정정 해소는 trim으로 저장하는 비대칭이 남아 있다. 같은 함수의 144-145행은 lineage edge의
    parent/child 식별자도 trim한다.
  - `journal/fills.go:173·385·393` — 체결 스냅샷·이벤트의 order_id를 trim해서 쓰고 trim해서
    조회한다(쓰기·조회가 같은 규칙이라 현재 자기모순은 없지만, 원문 저장 경로에서 들어온
    식별자와는 매칭되지 않는다).
  - `filldetect/payload.go:84` — 관측 payload의 orderId를 trim해서 Snapshot에 담는다.
- 이번 처리: **고치지 않았다.** 세 파일 모두 5.2의 파일 범위 밖이고, `fills.go`·`payload.go`는
  5.3(EXECUTION_CORRECTION)이 동시에 편집하는 파일이다. 범위를 넘어 고치면 5.3과 충돌한다.
- 후속 task 입력: **Manager 배정 필요**. `lineage.go:118`은 5.2가 만든 비대칭이므로 우선순위가
  높다(정정 해소만 식별자를 변형해 저장한다). `fills.go`·`payload.go`는 5.3에 붙이는 것이
  자연스럽다. 스펙 근거는 order-execution "브로커 식별자의 opaque 취급" — 저장 SHALL 원문,
  비교 SHALL 바이트 동일.

## 2026-07-26 [observation] round-trip 확인은 Gateway에 orders 리더가 배선돼야 동작한다 (task 5.2)

- 사실: 스펙은 "생성 응답의 식별자는 상세조회 round-trip으로 실재를 확인하며(SHALL)"라고
  쓴다. 확인 자체는 `execgw/roundtrip.go` + `journal.Attempt.DispatchVerified`(MarkAcked 후·
  Settle 전)로 구현했으나, 읽기 주체가 `Options.Orders`(nil 허용)이므로 **배선되지 않은
  Gateway는 P1과 동일하게 ack만으로 CONFIRMED**가 된다.
- 판단: `Orders`를 필수로 만들면 P1부터 있던 Gateway 생성자 계약이 깨지고 5.2 범위 밖의
  호출자(테스트 포함)가 전부 바뀐다. 기본값을 "확인 없음"으로 두는 대신 SHALL은 배선 쪽에
  걸었다.
- 후속 task 입력: **7.3**(Gateway 구성)이 `Orders: execgw.OfficialOrders{Client: …}`를 반드시
  채워야 한다. 채우지 않으면 스펙 요구가 런타임에서 미충족이며, 7.5 인터록의 "Gateway 구성
  확인" 항목에 이 필드를 포함하는 것이 자연스럽다.
- §0.4 예산: place당 `GET /api/v1/orders/{orderId}` 1회 추가(취소·정정은 미적용 — 청산 경로에
  왕복을 얹지 않는다, §0.3). retry matrix의 단건 조회 라인(`in-flight 주문당 3s마다 1회`)
  대비 place 빈도가 훨씬 낮아 새 라인을 만들지 않는다.

## 2026-07-26 [safe local] 배정된 원문 저장 위반 3곳 수정 완료 — lineage_edges의 나머지 2행은 남겼다 (task 5.3)

- 처리: Manager 배정대로 `journal/lineage.go:118`(`ResolveConfirmedWithLineage`의
  broker_order_id), `journal/fills.go:173`(스냅샷 쓰기 키)·`385`(LookupFill)·`393`(FillEvents),
  `filldetect/payload.go:84`(Snapshot.OrderID)에서 trim을 제거했다. trim은 공백 검사에만 남았고
  (RecordFill의 빈 id 거부, detect.go의 "payload에 id가 없으면 tracked id로 채움"), 저장·조회는
  수신 원문이다. 테스트: `journal/corrections_test.go`의 verbatim 3건 + `filldetect/payload_test.go`
  2건.
- **남긴 것**: 같은 함수의 `lineage.go:144-145`(lineage_edges의 parent/child 식별자 trim). 배정은
  "118의 비대칭 회복"이었고, 이 두 행을 고치면 읽기 쪽 `LineageChildren`·`ResolveCurrentOrderID`가
  **파라미터를** trim하는 것까지 함께 바꿔야 한다(저장만 원문으로 바꾸면 공백 있는 id가 조회에서
  누락된다). 현재는 쓰기·조회가 같은 규칙이라 자기모순이 없다 — 5.2가 fills.go에 대해 기록했던 것과
  같은 상태다.
- 후속 task 입력: lineage_edges 계열은 한 묶음(쓰기 2행 + 조회 2곳)으로 처리해야 한다. 실 식별자에
  공백이 없다는 전제 위에서만 현재 동작이 맞으므로, 우선순위는 낮지만 규칙 위반인 것은 맞다.

## 2026-07-26 [observation] D9의 정정 dedup 키가 왕복 정정의 3번째 관측을 흡수한다 (task 5.3)

- 사실: `UNIQUE(order_id, cumulative_qty, new_avg_price, new_filled_amount)`는 값 자체가 키다.
  브로커가 같은 수량에서 평균가를 A→B→A→B로 재서술하면 마지막 B는 첫 B와 키가 같아
  `ON CONFLICT DO NOTHING`이 삼킨다. 감사 행이 3개가 아니라 2개가 된다.
- 판단: D9가 이 키를 "crash 재관측의 이중 삽입 방지" 목적으로 명시했고, 스키마는 이미 v5로
  커밋됐다. 잃는 것은 감사 행 하나이며 수량·스냅샷에는 영향이 없다(스냅샷은 항상 최신 관측을
  따른다 — 테스트 `TestReObservedCorrectionIsDedupedByTheUniqueKey`가 고정). 키에 observed_at을
  넣으면 crash 재관측이 그대로 이중 삽입되므로 반대 방향이 더 나쁘다.
- 참고: 스펙의 "평균 체결가는 … 중복 판정 키에 포함해서는 안 된다(SHALL NOT)"는 **체결(fill)의**
  중복 판정 키에 대한 문장으로 읽었다 — 평균가 변경이 신규 체결로 계상되면 안 된다는 뜻. 정정
  테이블의 dedup 키는 D9가 별도로 정의한 것이고 둘은 다른 키다. 자구가 겹치므로 기록해 둔다.
- 후속 task 입력: 정정 이력의 완전성이 필요해지면(2b 측정 이후) 키에 관측 순번을 넣는 v6가
  필요하다. 지금은 불필요하다고 판단했다.

## 2026-07-26 [safe local] `execution_corrections.account_ref`를 관측이 아니라 journal에서 유도했다 (task 5.3)

- 사실: D9는 `account_ref NOT NULL`인데 `filldetect.Snapshot`에는 계좌 차원이 없다(감지기는
  주문 id로만 폴링한다). 관측이 계좌를 모른다.
- 이번 처리: `FillObservation.AccountRef`를 선택 필드로 추가하고, 비어 있으면 RecordFill이
  트랜잭션 안에서 `mutation_attempts.broker_order_id → intents.account_ref` 조인으로 유도한다.
  로컬 intent가 없는 주문(외부 주문)은 ""로 남는다 — NOT NULL은 만족하되 "귀속 불가"를 그대로
  기록한다. 계좌를 지어내는 것보다 낫다는 판단.
- 후속 task 입력: filldetect 상시 루프를 배선하는 change가 계좌를 알고 있다면 `AccountRef`를
  채우는 쪽이 정확하다(유도는 fallback이다).

## 2026-07-26 [observation] D6a가 design.md에 절(節)로 존재하지 않는다 — 수치는 tasks.md에만 있다 (task 6.1)

- 사실: design.md는 Risks 절에서 "아래 D6a 보수 기본값을 6.1이 전사"라고 쓰지만(153행),
  **D6a라는 절이 design.md에 없다**. 수치(계좌 스냅샷 10초·환율 60초)는 tasks.md 6.1 본문에만
  적혀 있다. review.md에도 없다.
- 이번 처리: tasks.md 6.1의 수치를 정본으로 삼아 `riskcalc.AccountSnapshotStaleness = 10s`,
  `FXRateStaleness = 60s`로 전사했다. 근거는 코드 주석에 적었다 — filldetect 폴링 3초
  (`DefaultConfig.PollInterval`, retry matrix의 단건 조회 주기)×3 + 여유가 10초이고 이는
  fill-detection SLO 목표 10초와 같은 수, 환율은 포지션보다 느리게 움직이므로 60초.
- 보수 방향 강제를 코드로 굳혔다: 호출자가 `Staleness`로 창을 **좁히는** 것만 유효하고
  넓히면 기본값이 이긴다(`Staleness.resolve`). §0.9의 단방향 안전이 주석이 아니라 동작이다.
  넓히려면 상수 자체를 바꿔야 하고 그건 사람 승인·audit 대상이다(§0.5·§0.7).
- 후속 task 입력: **2d**가 수치를 확정할 때 design.md에 D6a 절을 실제로 추가하는 편이 낫다.
  지금은 tasks.md 한 줄이 유일한 출처다.

## 2026-07-26 [safe local] 6.1을 `execgw.Limits`가 아니라 새 leaf 패키지로 냈다 (task 6.1)

- 사실: 6.1은 "문서+타입" 산출물이고, 한도 소비자는 `execgw/guardian.go`의 `Limits`
  (MaxQuantity·MaxNotional·Currency)다. 그런데 execgw는 이 change의 다른 태스크(1.5·1.7·7.x)가
  동시에 편집 중이고, 계산 계약은 journal·execgw 어느 쪽에도 의존할 이유가 없다.
- 이번 처리: `internal/riskcalc` — stdlib만 쓰는 leaf. 시계·네트워크·설정 조회가 없고 모든
  입력이 명시적이다(`Now`, 스냅샷 as-of, 환율 as-of). 숨은 `time.Now()`가 stale 스냅샷을
  신선하게 만들 수 있는 자리를 아예 없앴다. 소수는 journal 관례대로 decimal string 입출력 +
  float64 산술 + `FormatFloat(v,'f',-1,64)` 출력이며, 한도 비교는 journal의 1e-9 상대 허용오차를
  재사용해 동률을 초과로 판정한다(`WithinLimit`).
- 후속 task 입력: **1.7**(Limits fail-closed)과 **7.5**(인터록)가 이 패키지를 소비한다.
  `riskcalc`는 아직 어디서도 import되지 않는다 — 배선은 그 태스크들의 몫이며, 6.1은 계약만
  낸다는 태스크 문언대로다.

## 2026-07-26 [safe local] `Limits` 항목별 configured 비트가 `app/engine/interlock.go`를 컴파일 깨뜨렸다 (task 1.7)

- 사실: 1.7은 `execgw.Limits`를 `float64` 3필드 → `Limit{Set,Value}` 5필드로 바꾼다
  (주문 수량·주문 notional·총 개방 노출·일일 손실 절대액·일일 손실 비율). 유일한 비-테스트
  소비자가 `internal/app/engine/interlock.go`(section 7 소유)이므로 그 파일이 컴파일되지
  않는다 — 트리를 빨갛게 두지 않으려면 손댈 수밖에 없다.
- 이번 처리: **의미 무변경 기계적 적응만.** `gateLimits`가 `boundIfPositive`로 두 항목의
  비트를 세우고(config에 양수일 때만 — 부재가 "0이라는 한도"가 되지 않는다), 인터록의
  `IsZero()` 검사를 `!MaxQuantity.Set && !MaxNotional.Set`으로 바꿨다(옛 `<=0 && <=0`과
  동일). audit 문자열은 `.Value`를 쓴다. 인터록 단언은 `interlock_test.go:176`의
  `MaxQuantity != 10` → `MaxQuantity.Value != 10` 한 줄뿐이다.
- 결과: config에는 아직 총 노출·일손실 항목이 없으므로 `gateLimits` 스냅샷은 5개 중 2개만
  설정된다. 그 스냅샷을 실은 EXPOSURE_RAISING 결정은 **Gateway가 거부한다** — 스펙이 요구하는
  fail-closed 방향이고("부분적으로 무제한인 게이트는 허가된 게이트가 아니다"), 2a에서 진입
  결정을 발급하는 생산 코드는 없으므로 현재 동작 영향은 0이다.
- 후속 task 입력: **7.5**가 (1) config에 총 개방 노출·일일 손실 절대액·일일 손실 비율을
  추가하고, (2) 인터록 검사를 `Limits.Validate()`로 교체해야 한다(항목별 fail-closed는 이미
  그 함수 안에 있다). **7.3**도 입력이 있다 — 태스크 문언의 "NonceStore 구성"은 이제 존재하지
  않는 배선이다(아래 항목).

## 2026-07-26 [observation] nonce 소비가 `MarkDispatchStarted`에 병합되면서 execgw의 `NonceStore`가 사라졌다 (task 1.7)

- 사실: D5·스펙이 "소비 기록은 전송 시작 기록과 같은 트랜잭션에서 남긴다(SHALL)"를 요구하므로
  소비는 `journal.Attempt.MarkDispatchStarted`의 `BEGIN IMMEDIATE` 안으로 들어갔다. 그 결과
  Gateway가 호출할 대상이 없어져 `execgw.NonceStore`·`NewMemoryNonceStore`·`Options.Nonces`를
  삭제했다(외부 배선처 0곳이었다). 재시작에 잊는 저장소를 엔진 프로필이 주입할 수 있는 구멍
  자체를 없앤 것이기도 하다.
- 부수 효과 2건, 둘 다 보수 방향이라 그대로 뒀다:
  1. **소비가 마지막 재검증보다 먼저 일어난다.** MarkDispatchStarted가 커밋된 뒤 dispatch
     클로저가 행을 다시 읽어 검증한다. 그 재검증이 거부하면 아무것도 전송되지 않았는데 결정은
     소비된 상태다 — 마지막 검사에서 떨어진 결정을 재사용 가능하게 두는 것보다 낫다.
  2. **재제출 거부는 journal에 attempt로 남지 않는다.** 소비 여부·키 점유는 Prepare 전에
     durable 조회로 걸러 `guardian_nonce_reused`를 낸다. 근거는 기존 심볼 in-flight 거부와
     같다 — 그 결정을 소비한 attempt가 이미 기록이다.
- 후속 task 입력: **7.3**의 "Gateway 구성(…·NonceStore·…)" 문언은 갱신이 필요하다. durable
  저장소는 journal 그 자체이며 주입할 필드가 없다. **보존 스윕**(`PruneSpentNonces`)은 호출자가
  아직 없다 — 불변식(보존 ≥ 최대 결정 TTL)은 함수가 스스로 강제하고 테스트로 고정했지만,
  주기적 호출 배선은 엔진 루프를 소유하는 change의 몫이다.

## 2026-07-26 [observation] `riskcalc.WithinLimit`을 1.7에 배선하지 않았다 — 동률 판정이 다르다 (task 1.7)

- 사실: 6.1의 issues 항목은 "1.7과 7.5가 riskcalc를 소비한다"고 적었다. 배선하지 않았다.
  이유 둘: (1) `WithinLimit`은 `Money`(통화 있는 총계) 비교이고 1.7이 검사하는 두 항목 중
  **주문 수량에는 통화가 없다** — notional만 바꾸면 한 스냅샷 안에서 두 비교의 규칙이 갈린다.
  (2) 판정이 다르다: `riskcalc.WithinLimit`은 **동률을 초과로** 본다("Reaching the limit is not
  staying under it"), execgw의 주문당 검사는 P1부터 `value > limit`로 **동률을 허용**한다.
- 판단: 동률 의미를 조용히 뒤집는 것은 1.7의 위임 범위 밖이다(사이징·한도 수치는 2d). 지금
  바꾸면 "한도 정확히 채우는 주문"이 거부로 전환되는데, 이는 보수 방향이지만 **한도 수치의
  의미 변경**이라 사람 판단이 필요하다.
- 후속 task 입력: **Manager 판정 필요** — 주문당 한도의 동률을 (a) 현행 유지(허용), (b)
  riskcalc과 통일(초과) 중 어느 쪽으로 할지. (b)라면 `execgw.checkLimits`가 riskcalc을
  import하고 수량 비교에도 같은 허용오차 규칙을 쓰는 형태가 자연스럽다. **7.5**가 인터록에서
  같은 한도를 쓰므로 그 전에 정해지는 것이 좋다.

## 2026-07-26 [observation] RISK_REDUCING preimage는 가격·주문유형을 결속하지 않는다 (task 1.4)

- 사실: P1의 `GuardianDecision.IntentHash`는 canonical intent 전체(주문유형·통화모드·fractional
  플래그 포함)를 결속했다. 이번 change는 그것을 클래스별 preimage로 대체했고, 스펙이 정한
  `ReductionIntent` 필드는 계좌·시장·심볼·방향·상한 수량·사유뿐이다. 따라서 **같은 심볼·방향·
  수량 이하의 매도라면 가격이나 주문유형(limit↔market)을 바꿔도 preimage 검증을 통과한다.**
  (`RiskIntent` 쪽은 진입가가 필드라 이 구멍이 없다.)
- 판단: 2a에서 RISK_REDUCING 발급자는 flatten 하나이고 발급자=호출자이므로 악용 경로가 없다.
  스펙이 필드 목록을 명시했으므로 임의로 필드를 늘리는 것은 자구 이탈이다.
- 후속 task 입력: 보호주문(2c)·Guardian(2d)에서 발급자와 제출자가 분리되면 재검토 대상이다.
  선택지는 `ReductionIntent`에 상한 가격을 추가하거나(스펙 개정), 청산 가격 결정을 발급자
  소관으로 못 박는 것이다.

---

## Manager 판정 (1차 물결 검증, 2026-07-26)

- **0.1 D9 편차 5건 전부 승인.** (1) attempts.account_ref additive — 1.3이 채우고 intents 값과 불일치 시 거부할 의무 승계. (2) reconcile_states 이중 부분 unique — D9 문장의 의도를 SQLite NULL 의미론에서 실제로 강제하는 유일한 형태. (3) cause CHECK 제외 — 승인하되 **4.1의 저장 계층이 Go 상수 집합을 쓰기 시점에 검증(미지 cause는 거부)하는 조건부**. (4) dedup 컬럼 NOT NULL DEFAULT '' — 승인, **5.3은 NULL이 아니라 ''를 쓴다**. (5) spent_nonces FK 제외 — 보존 불변식 우선, 승인.
- **5.2 잔여 원문 저장 위반의 배정**: `journal/lineage.go:118`(ResolveConfirmedWithLineage의 trim — cancel 해소는 원문·amend 해소는 trim으로 갈라진 상태), `journal/fills.go:173/385/393`, `filldetect/payload.go:84` → **5.3 담당자에게 배정**(fills·payload는 원래 범위, lineage는 동일 해소 쓰기 경로의 정합 회복).
- **round-trip이 Options.Orders 배선 전 무효** → 7.5 인터록 "Gateway 구성 확인"에 Orders 배선 검증 항목 추가로 승계.
- 5.1의 보수 판정(REPLACED without successor → UNKNOWN, CANCELED+successor → UNKNOWN, PENDING_* 무증거) 승인 — 전부 fail-closed 방향.
