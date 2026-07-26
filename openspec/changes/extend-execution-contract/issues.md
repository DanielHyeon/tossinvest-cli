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

## 2026-07-26 [safe local] 예약 산술을 `riskcalc`의 새 exact decimal 층으로 냈다 — journal이 riskcalc을 import한다 (task 3.1)

- 사실: 스펙은 "예약 산술은 decimal 문자열 연산이며 float 누적을 사용하지 않는다(SHALL NOT)"를
  요구한다. 그런데 6.1이 낸 `riskcalc`의 총계 계산은 **의도적으로** float64 누적 + 1e-9 허용오차다
  (그 패키지 doc에 근거가 적혀 있다). 두 규칙이 한 패키지 안에서 공존해야 한다.
- 이번 처리: `internal/riskcalc/decimal.go` — 자릿수 문자열 위에서 동작하는 정확 산술
  (`AddDecimal`/`SubDecimal`/`CompareDecimal`/`Min`/`Max`/`Canonical`). 기존 `aggregate.go`의
  float64 평가는 **한 줄도 바꾸지 않았다**: 그쪽은 가격×수량·환율 평가이고 자체 허용오차 규칙이
  스펙과 정합한다. 예약 원장은 평가가 아니라 합산이므로 규칙이 다르다는 것을 양쪽 doc에 적었다.
- 결과로 **`internal/journal` → `internal/riskcalc` 의존이 새로 생겼다.** 근거: (1) 태스크 문언이
  "use riskcalc staleness constants"를 지시하므로 의존 방향은 이미 정해져 있었고, (2) riskcalc은
  stdlib만 쓰는 leaf라 순환이 불가능하며, (3) "decimal 두 개를 어떻게 더하는가"는 저장 세부가
  아니라 계산 계약이다. 회귀 테스트: `TestReservationArithmeticIsExact`(0.1 × 10 = 정확히 1,
  float64 누적이면 0.9999999999999999로 한도를 통과해 11번째를 허용한다).
- 후속 task 입력: **4.2**의 확정 하한도 같은 층을 쓴다. 7.x가 execgw에서 총계를 계산할 때
  평가(riskcalc float64)와 예약 합산(exact)의 경계를 섞지 않아야 한다.

## 2026-07-26 [safe local] as-of 재검증을 "예약 원장 버전"으로 구현했다 — 스키마 무변경 (task 3.1)

- 사실: 스펙은 "안에서 as-of·staleness를 검증한 뒤 예약을 삽입하며, 불충족이면 롤백·재수집한다"를
  요구하고, 3.3은 "as-of 조건이 실제 재검증을 유발"함을 증명하라고 한다. staleness만으로는 부족하다 —
  단일 커넥션(`SetMaxOpenConns(1)`)에서 두 트랜잭션은 겹칠 수 없으므로, 스냅샷이 **이미 반영된
  다른 결정의 예약을 모른다**는 사실은 시간만으로 드러나지 않는다.
- 이번 처리: `ReservationVersion(account)` = `1 + 행 수 + RELEASED 행 수`. 삽입도 해제도 값을
  전진시키고 감소는 불가능하다(행 삭제 없음 — 고아 회수도 삭제가 아니라 해제다). 호출자는 스냅샷
  수집 **전에** 읽고, `Reserve`가 트랜잭션 안에서 다시 읽어 다르면 `ErrSnapshotSuperseded`로
  롤백한다. **스키마 변경 0** — v5 표에 컬럼을 더하지 않았다.
- 0을 "미조회"로 남기려고 버전을 1부터 시작시켰다. 버전을 안 읽은 호출자는 통과가 아니라 거부된다.
- 후속 task 입력: **7.3**이 Gateway에 예약 저장소를 배선할 때 "버전 읽기 → 스냅샷 수집 →
  Reserve" 순서가 계약이다. `ReserveWithRecollection`이 그 순서를 강제하는 형태(수집 콜백이
  요청 전체를 만든다)로 제공된다.

## 2026-07-26 [safe local] 재수집 루프를 execgw가 아니라 journal에 두었다 (task 3.1)

- 사실: 태스크는 "재수집 루프는 CALLER-side helper (execgw 또는 journal — 선택하고 이유를 문서화)".
- 이번 처리: `journal.ReserveWithRecollection`. 이유 셋: (1) 상한·데드라인·fail-closed는
  트랜잭션이 **거부하기 때문에** 존재하는 것이라 같은 계약의 반쪽이다, (2) 재시도 대상 오류
  (`ErrSnapshotStale`·`ErrSnapshotSuperseded`)와 재시도 금지 대상(`ErrReservationLimitExceeded`)의
  구분이 여기 정의돼 있어 두 패키지가 그 목록에 합의할 필요가 없다, (3) journal은 여전히 I/O를
  하지 않는다 — 수집은 호출자의 콜백이고 journal이 보는 것은 데이터뿐이다.
- 기본값: 상한 3회(스펙 명시), 예산 10초 = `riskcalc.AccountSnapshotStaleness`. 데이터가 신선하게
  유지되는 시간보다 오래 도는 루프는 성공할 수 없고 진입 결정만 붙잡는다.

## 2026-07-26 [safe local] 종결 해제 훅을 `transition()` 한 곳에 걸었다 — CONFIRMED는 해제하지 않는다 (task 3.2)

- 사실: D5의 해제 트리거 중 "파생된 브로커 종결 상태 도달"은 attempt 종결(NOT_DISPATCHED·
  FAILED_CONFIRMED)과 주문 종결(FILLED/CANCELLED/REJECTED 계열) 둘 다를 포함한다. 전자는
  `Settle`·`ResolveFailed`·재시작 복구가 각각 부르고, 후자는 `RecordFill`이 부른다.
- 이번 처리: attempt 쪽은 메서드마다가 아니라 **`transition()` 안에서 목적 상태로 판정**한다
  (`releasesReservations`). 그래서 gateway의 refuse, 재시작 복구의 NOT_DISPATCHED, 해소의
  FAILED_CONFIRMED가 전부 자동으로 같은 트랜잭션 안에서 해제된다 — 생산자가 기억할 것이 없다.
  주문 쪽은 `RecordFill`의 같은 `BEGIN IMMEDIATE` 안에서 `broker_order_id` 조인으로 해제한다.
- **CONFIRMED는 해제하지 않는다.** 주문이 실재한다는 뜻이므로 노출은 살아 있고, 해제하면 살아 있는
  노출에 대해 한도 여유를 열어준다. UNRESOLVED_IN_DOUBT도 해제하지 않는다(운영자 유일 출구).
  둘 다 테스트로 고정(`TestFailedConfirmedReleasesAndConfirmedDoesNot`).
- 8.2 참고: `transition()`은 1.3·1.7이 이미 손댄 함수다. 이번 추가는 **예약이 없는 attempt에는
  no-op**이므로 기존 단언 변경 0건이다(전체 스위트 재실행 green).

## 2026-07-26 [safe local] 운영자 해제의 audit을 인터페이스로 뺐다 — `audit.Log.RecordAction` 추가 (task 3.2)

- 사실: 스펙은 운영자 해제에 "근거 기록·audit 필수"를 건다. 그런데 `internal/journal`은
  `internal/audit`을 import하지 않으며, `audit.Log`에는 `RecordChange`(값이 바뀔 때만 기록)뿐이라
  **같은 해제를 두 번 하면 두 번째가 삼켜진다** — 사건 로그에는 틀린 의미론이다.
- 이번 처리: (1) `internal/audit`에 `RecordAction(action, setting, value, detail)` 추가 —
  무조건 append. `RecordChange`는 한 줄도 바꾸지 않았다. (2) journal은 그 시그니처만 요구하는
  1-메서드 인터페이스 `ReservationAuditor`를 정의한다. 의존 방향이 생기지 않고, 배선은
  `Auditor: auditLog` 한 줄이다. (3) action 문자열은 journal이 소유
  (`AuditActionReservationRelease = "risk_reservation.release"`) — 사건의 주인이 journal이므로.
- **audit 기록이 실패하면 해제도 실패한다**(commit 전에 기록). 근거 없이 풀린 홀드가 남는 것보다
  풀리지 않는 편이 낫다. 테스트: `TestAFailedAuditWriteAbortsTheRelease`.
- 후속 task 입력: **7.3/7.5**가 운영자 해제 CLI/배선을 만들 때 `audit.Open(...)`의 `*Log`를
  그대로 넘기면 된다.

## 2026-07-26 [observation] 거래일 소멸의 시장을 결정 preimage에서 유도했다 (task 3.2)

- 사실: `risk_reservations`에는 market 컬럼이 없다(D9 표). `trading_day`는 시장-로컬 날짜인데
  어느 시장의 날짜인지가 행에 없다. attempt→intent 조인은 예약 시점에 attempt가 없어서 못 쓴다.
- 이번 처리: `decisions.risk_preimage`(RiskIntent/ReductionIntent 둘 다 market 필드를 가진다)를
  파싱해 `clock.Market.TradingDay(now)`와 비교한다. 파싱·시장 해석 실패는 **소멸시키지 않고
  보존**하며 `ReservationsAwaitingOperator`에 `UNKNOWN_MARKET`으로 뜬다 — 한도를 필요보다 좁게
  두는 방향이 보수적이다.
- 소멸은 **lazy**다: 원장을 다음에 쓸 때(`Reserve` 트랜잭션 안)와 기동 시(3.3)만 계산하며 타이머는
  없다. 위험 홀드를 푸는 백그라운드 goroutine은 트랜잭션 규율 없는 두 번째 writer이고, 엔진이
  거래를 멈춘 뒤에도 계속 돈다.

## 2026-07-26 [safe local] 재생 횟수를 전송 *전에* 세고 409만 환불한다 (task 2.1/2.2)

- 사실: D3은 "회수 상한(기본 2회)"과 "`409 request-in-progress` → 상한 미소비"를 함께 요구한다.
  둘을 동시에 만족시키는 순서는 하나뿐이다. 응답을 보고 세면 "전송 후 · 계수 전" 크래시가
  재생 1회를 공짜로 돌려주고, 크래시 루프에서는 무한히 돌려준다.
- 이번 처리: `journal.MarkReplayStarted`가 전송 **전에** BEGIN IMMEDIATE 안에서 상한 검사와
  증가를 함께 하고, 409에서만 `RefundReplay`가 카운트를 되돌린다. 환불은 `last_replay_at`을
  건드리지 않는다 — "처리 중"에 대한 답은 대기이므로 최소 간격은 유지돼야 한다.
- 부작용: 크래시로 쓰지 못한 재생 1회를 잃을 수 있다. 조회 폴백이 그대로 돌므로 비용은 0이고,
  반대 방향은 키 유효창을 무한 루프에 쓰는 것이라 비대칭이 크다.

## 2026-07-26 [observation] 409가 상한을 소비하지 않으므로 대기 횟수 상한을 새로 만들었다 (task 2.2)

- 사실: 409는 재생의 최빈 응답이고(D3) 상한을 소비하지 않는다. 따라서 상한만으로는 "영원히
  409를 답하는 브로커"에 대해 루프가 유계가 아니다. TTL−margin이 상한이긴 하지만 기본값으로
  (10분−60초)/5초 ≈ 108회 왕복이다.
- 이번 처리: `ReplayConfig.MaxWaits`(기본 3)를 추가했다. 스펙·design에 없는 값이며,
  "원 요청이 얼마나 오래 처리 중일 수 있는가"는 `[미측정 — 2b]`다. 소진 시 조회 폴백으로
  전환하므로 방향은 보수적이다.
- 후속 task 입력: **2b**가 왕복 p99와 함께 이 값을 측정해야 한다(margin 60초와 같은 묶음).

## 2026-07-26 [safe local] 재생의 critical 알림을 journal outbox에 직접 넣었다 — obs를 import할 수 없다 (task 2.2)

- 사실: `internal/obs`가 `internal/execgw`를 import한다(Notifier가 EntryGate를 래치한다).
  따라서 execgw는 obs를 import할 수 없고, `obs.EventOrderUnresolved`·`obs.SeverityCritical`을
  참조할 수 없다. 현재 코드에서 UNRESOLVED 알림을 실제로 내는 곳은 flatten뿐이며(양쪽을 다
  import한다), Resolver의 `park()`는 게이트 래치만 한다.
- 이번 처리: 422 key-conflict 주차 시 `journal.EnqueueAlert`로 outbox에 직접 기록한다 —
  obs의 critical 경로의 durable 절반이며 Notifier의 `Flush`가 그대로 배달한다. 이벤트 타입
  문자열은 execgw에 비공개 상수로 복제하고, `execgw_test`(외부 테스트 패키지 → obs import
  가능)에서 실제 outbox 행의 `Type`을 `obs.EventOrderUnresolved`와 대조해 고정했다.
- 대안 기각: execgw에 별도 Alerter 인터페이스를 새로 만드는 것 — obs와 평행한 두 번째 알림
  표면이 생기고 배선처가 늘어난다. outbox는 이미 "durable 우선, 배달 나중"이 계약이다.

## 2026-07-26 [observation] 생산 재생 transport가 어디에도 배선되지 않았다 (task 2.1/2.2)

- 사실: `execgw.HTTPReplay`(저장 바이트를 `POST /api/v1/orders`로 그대로 보내는 transport)를
  execgw 안에 두었다. `internal/official`에 만들지 않은 이유는 (a) official의 모든 메서드가
  구조화 필드에서 본문을 **구성**하는데 재생은 정확히 그것을 하면 안 되고, (b) official은
  이 태스크의 파일 범위 밖(Pre-Edit 대상)이기 때문이다.
- 결과: `Options.Replay`·`Options.Attested`를 채우는 생산 코드가 **0곳**이다. attestation
  기본값이 OFF이므로 지금은 의도된 상태(2b 전 비활성)이고, 진입점은 `replay_not_attested`로
  즉시 폴백한다.
- 후속 task 입력: **7.3**(Gateway 구성)이 배선할 때 `HTTPReplay.Headers`가 official의
  토큰 매니저·계좌 헤더를 재사용할 수 있어야 한다. official의 토큰 접근이 비공개이므로
  Pre-Edit이 필요할 수 있다 — 또는 배선을 2b가 attestation과 함께 가져가는 편이 자연스럽다.
  현재 `HTTPReplay`는 401 재시도를 하지 않는다(계수된 재생을 transport가 몰래 두 번 보내면
  상한이 무의미해진다).

## 2026-07-26 [safe local] dedup을 `findSuccessors`(amend 해소)에도 적용했다 (task 2.4)

- 사실: 2.4 문언은 "조회 폴백"의 유일성 판정을 지목하고 place 경로의 `scanBoth`를 겨냥한다.
  그런데 amend 해소의 `findSuccessors`(amend_indoubt.go)도 OPEN·CLOSED를 걷고 호출부가
  `switch len(successors) { case 1: ... }`로 **유일성을 판정**한다. `PARTIAL_FILLED`가 양쪽
  그룹에 속하므로(openapi) 부분 체결 승계 주문은 두 번 잡히고 주차된다 — place와 동일한 결함이다.
- 이번 처리: 같은 규칙(orderId 바이트 동일 dedup)을 적용했다. 파일 범위 안이고, 스펙 문장
  ("유일성 판정 전에 orderId 기준 dedup을 수행해야 한다(SHALL)")은 경로를 한정하지 않는다.

## 2026-07-26 [observation] 관측 창의 시작점을 `dispatch_started_at`으로 잡았다 (task 2.4)

- 사실: 스펙은 "관측 창 동안 같은 심볼에 다른 mutation이 전송되었다면"이라고만 쓰고 창의
  시작을 정의하지 않는다. 후보는 (a) 해소 절차 시작 시각, (b) 이 attempt의 전송 시작 시각.
- 판단: delta 교차확인이 비교하는 baseline은 **전송 직전** 스냅샷이므로, 오염을 판정해야 하는
  구간은 baseline이 유효하지 않게 된 시점부터다 → (b). (a)를 쓰면 크래시 후 재시작까지의
  구간(가장 오염되기 쉬운 구간)이 창 밖으로 빠진다.
- 부수 결정: `MutationsDispatchedSince`는 `NOT_DISPATCHED`를 제외한다. `MarkDispatchStarted`가
  브로커 호출 **전에** 커밋되므로 그 사이 거부는 전송 시각을 가진 채 아무것도 보내지 않았다.

## 2026-07-26 [observation] `indoubt.go` 헤더의 "멱등키 없음" 서술을 고치지 않았다 (task 2.1~2.4)

- 사실: `indoubt.go:9-12`는 "The official API has no idempotency key, so 'try again' and 'place
  a second live order' are the same action"이라고 쓴다. 이 change가 뒤집은 전제이며, 같은 파일에
  내가 2.4 수정을 넣었으므로 헤더와 본문이 어긋난 상태다.
- 이번 처리: **손대지 않았다.** 이 문구의 정정은 tasks.md **2.5 [M]** — Manager 태스크로 명시
  배정돼 있다(`retry.go:8-10,77`·아카이브 스펙과 한 묶음). 팀메이트가 먼저 고치면 2.5의
  일관성 검토가 부분적으로 이미 적용된 상태에서 시작된다.
- 참고: 같은 헤더의 "the Resolver has no trading service, no broker writer, no submit path"는
  **여전히 참이다**(재생 문은 Gateway에 있다). 무효가 된 것은 멱등키 문장 하나다.

## 2026-07-26 [observation] 기동 스윕의 "고아"는 앞으로 생길 수 없는 상태다 — 그래도 남겼다 (task 3.3)

- 사실: 3.2가 해제를 종결 기록과 같은 트랜잭션에 넣었으므로, "attempt는 종결인데 홀드는 HELD"는
  **이 빌드가 만들 수 없는 상태**다. 그런데도 `SweepReservations`가 그 케이스를 회수한다.
- 근거: (1) v5 이전 빌드가 남긴 행, (2) 마이그레이션 실패 후 백업 복원본, (3) 사람이 SQL로 손댄
  DB. 셋 다 실재 가능하고, 회수되지 않는 홀드는 계좌 한도를 **영구히** 좁힌다(만료도 거래일도
  걸리지 않는다 — OPEN_EXPOSURE는 날짜 속성이 없다).
- 테스트는 그 상태를 SQL로 직접 만든다(`TestStartupSweepRecoversAnOrphanedTerminalHold`) —
  `Settle`을 통과시키면 live 경로가 해제해 버려 스윕을 검증할 수 없다.
- fail-closed 관측은 스윕에서도 제외된다(`f.fail_closed = 0` 조건): 만료 추정 해제 금지가 기동
  경로에도 적용된다는 뜻이다(`TestOrphanSweepDoesNotReleaseAFailClosedObservation`).

## 2026-07-26 [observation] as-of 동시성 테스트는 goroutine이 아니라 버전 주입으로 썼다 (task 3.3)

- 사실: 태스크 문언대로 `SetMaxOpenConns(1)` 아래에서 두 예약 트랜잭션은 겹칠 수 없다. goroutine
  두 개를 `-race`로 돌리면 뮤텍스가 동작한다는 것만 증명되고, **as-of 조건이 재수집을 유발한다**는
  것은 증명되지 않는다.
- 이번 처리: `TestASnapshotThatPredatesACommittedReservationIsRecollected` — 수집 콜백이
  (1) 버전을 먼저 읽고, (2) "브로커 왕복 중"에 다른 결정이 예약을 커밋하고, (3) 낡은 버전으로
  `Reserve`를 부른다. 1회차 `ErrSnapshotSuperseded` → 재수집 → 2회차 성공, 그리고 **두 홀드가
  모두 합산에 잡힌다**(세 번째 400이 1000 한도에 도달)까지 단언한다. 음성 대조군
  (`TestAnUnchangedLedgerNeedsNoRecollection`)이 "항상 실패하는 버전 검사"로는 통과 못 하게 막는다.
- 동시 goroutine 테스트는 3.1의 `TestConcurrentDecisionsCannotBothTakeTheLastSlot`에 별도로 있다
  (스펙 시나리오 "동시 다심볼 결정"). 두 테스트는 다른 것을 증명한다.

## 2026-07-26 [safe local] `Tracker.Observe`·`Resolve` 시그니처를 바꿨다 — 비-테스트 호출자 0곳 (task 4.1)

- 사실: 4.1은 "Tracker 상태를 journal 투영으로 이전"을 요구한다. 그러려면 `Observe`가 durable
  write를 해야 하고, durable write는 실패할 수 있다. 기존 시그니처 `Observe(Diff) Outcome`에는
  실패를 알릴 자리가 없다.
- 이번 처리: `Observe(ctx, Diff) (Outcome, error)`, `Resolve(ctx, operator, note) error`.
  **비-테스트 호출자는 0곳**이다(grep 확인 — Tracker는 아직 어디에도 배선되지 않았다). 테스트
  호출부 16+3곳은 헬퍼 `observe(t, tracker, diff)` 도입으로 기계적 적응했고 **단언은 한 줄도
  바꾸지 않았다**. `ObserveContext` 같은 병행 메서드를 두지 않은 이유: 영속하지 않는 쪽을
  실수로 계속 쓸 수 있는 문을 남기게 된다.
- 8.2 참고: 이 파일은 "기존 단언 변경 사전 열거" 목록에 없다. 변경된 것은 호출 형태이며 단언은
  동일하다(0.2의 outbox 픽스처 항목과 같은 종류).

## 2026-07-26 [safe local] RECONCILE cause → EntryGate reason 매핑을 "자동 해제 가능한가"로 갈랐다 (task 4.1)

- 사실: `reconcile_states.cause`는 5값인데 EntryGate의 reconcile 계열 reason은 2개
  (`reconciliation_mismatch` = clean reconcile로 자동 해제, `reconciliation_mismatch_permanent`
  = 운영자만). 새 reason 코드를 추가하면 `latchOrder`(운영자 runbook과의 계약)를 건드린다.
- 이번 처리: `execgw.ReconcileReasonFor` — SNAPSHOT_UNAVAILABLE·SNAPSHOT_STALE·QUANTITY_MISMATCH
  → mismatch(재조회가 반증 가능), IDENTIFIER_CONFLICT·ATTRIBUTION_FAILED → permanent(반증
  불가 — CANCEL_REJECTED 별도 레코드의 형태가 `[미측정 — 2b 2.1]`이므로 "또 봐도 안 맞더라"는
  해소의 증거가 아니다). **미지 cause도 permanent** — 구버전 행이 넓게 막는 방향.
- 부수 변경 1건: `reconcile.isReconcileReason`을 `ReasonReconcileMismatch` 하나로 좁혔다.
  좁히지 않으면 clean reconcile이 투영된 permanent 심볼 블록(식별자 충돌)을 지운다. Tracker가
  심볼 스코프에서 내는 reason은 원래 mismatch 하나뿐이라 의미 변화는 없다.
- `reconcile_states`에 market 컬럼이 없다(D9). 심볼 스코프 상태는 `BlockSymbol("", symbol, …)`로
  **모든 시장**을 막는다 — 스키마가 기록하지 않는 스코프의 보수적 해석.

## 2026-07-26 [observation] Tracker는 자기가 넣은 행만 해제한다 — permanent 승격은 계좌 전역 행이다 (task 4.1)

- 사실: Tracker의 3연속 실패 승격은 "원인"이 아니라 정책이다. 5개 cause 중 여기 해당하는 값이
  없다. cause 목록을 늘리면 0.1이 굳힌 집합이 흔들리고 4.1의 재량을 넘는다.
- 이번 처리: 승격은 **계좌 전역(symbol NULL) + QUANTITY_MISMATCH** 행으로 영속한다. `Restore`는
  계좌 전역 행을 permanent 블록으로 복원하고 `failures`를 임계치로 되돌린다 — 0으로 두면 재시작
  직후 clean 한 번이 사람이 안 푼 블록을 "해제"한다(`TestARestartKeepsAPermanentMismatchPermanent`).
- 해제는 전부 `ExpectCause: QUANTITY_MISMATCH`를 건다. 다른 생산자(식별자 충돌·귀속 실패)의 행은
  Tracker의 clean pass가 건드리지 못한다(`TestRestoreProjectsStatesThisTrackerDidNotEnter`).
- 후속 task 입력: **7.3**이 Gateway 구성 시 `Tracker.Journal`을 채우고 기동 시 `Restore(ctx)`를
  1회 호출해야 한다. 호출하지 않으면 재시작이 차단을 잃는다(스펙 SHALL 미충족). 인터록의
  "Gateway 구성 확인"(7.5) 항목 후보다.

## 2026-07-26 [safe local] 확정 하한을 `riskcalc`에 두었다 — 소비자는 아직 0곳 (task 4.2)

- 배치 근거: 계산 계약이지 게이트웨이 배선이 아니다. 순수·stdlib·모든 입력 명시(시계 없음)이므로
  숨은 `time.Now()`가 stale 스냅샷을 신선하게 만드는 자리가 없고, 3.1이 도입한 exact decimal 층을
  그대로 쓴다(분수 주식 하한이 부동소수로 새지 않는다). execgw에 두면 네트워크 타입이 따라 들어온다.
- 스냅샷 부재·stale은 **에러가 아니라 하한 0**이다. 에러로 만들면 호출자가 "계산 불가" 별도 경로를
  갖게 되고, 그 경로가 결국 다른 일을 한다. 0은 이 공식이 정의한 답이다. 대신 `Bound` 필드로
  "미조회"와 "낡음"을 구분한다 — 운영자의 다음 행동이 다르다.
- `LocalOpenSellQuantity`의 빈 문자열은 **거부**한다(0이 아니다). 미체결 SELL을 모르는 채 0으로
  간주하면 이미 걸린 주문이 잡아둔 수량까지 매도를 허가한다. 음수도 거부 — 뺄셈의 피감수가
  음수면 하한이 올라간다(SHALL NOT의 유일한 구멍).
- 매도가능은 `min()`에서만 쓰이고 주석에 `[미측정 — 2b 2.8]` 태그를 달았다.
- **소비자는 아직 없다.** 4.2 문언이 "공식 구현"이고, RECONCILE 중 청산 경로에 이 하한을 적용하는
  배선은 엔진 루프/Gateway를 소유하는 change의 몫이다. 후속: **7.3/7.5**가 자동 청산 경로에
  이 함수를 걸 때 flatten은 제외해야 한다(아래).

## 2026-07-26 [observation] flatten 면제를 정적 import 검사로 고정했다 — 전이 의존은 검사하지 않는다 (task 4.2)

- 사실: §0.3은 수동 flatten이 확정 하한의 대상이 아니라고 못 박는다(자체 신선 조회). 그런데
  `internal/flatten`은 `internal/journal`을 거쳐 `internal/riskcalc`에 **전이적으로** 닿는다
  (3.1이 예약 산술을 거기 두었기 때문). 전이 그래프를 금지하면 journal 자체를 금지하게 된다.
- 이번 처리: `internal/flatten`의 **비-테스트 파일 직접 import**만 검사한다
  (`TestFlattenCannotReachTheConfirmedFloorRule`). 호출이 생기려면 누군가 리뷰에서 정당화해야 하는
  import 한 줄이 추가돼야 한다. 행동 회귀는 `TestFlattenSizesFromItsOwnFreshReadNotAConfirmedFloor`
  — 스냅샷을 하나도 주지 않은 상태에서 자체 조회한 sellable 전량으로 주문이 나간다(확정 하한이라면
  0이라 주문이 없다). 기존 `TestLiquidationSizesFromTheSellableQuantity`가 더 넓은 characterization이다.

## 2026-07-26 [safe local] 게이트 OFF의 "브로커 호출 0회" 단언이 반전됐다 (task 7.1)

- 사실: `interlock_test.go`의 `TestGateOffStartsAndTouchesNothing`은 "게이트 OFF 경로는
  브로커를 한 번도 부르지 않는다"를 §0.2 근거로 고정하고 있었다. 7.1은 계좌 해석을 게이트
  검증 밖으로 빼서 기동 경로로 옮기므로(D8 1단계), 이 단언은 성립할 수 없다.
- 이번 처리: 테스트를 `TestGateOffStartsAndDoesNoGateWork`로 바꾸고 단언을 좁혔다 —
  계좌 읽기는 **정확히 1회**, attestation 미읽기(ExpiresAt 0), Guardian 미공개,
  `AccountRef` 채워짐. §0.2가 실제로 보호하는 것(게이트가 아무 일도 하지 않는다)은
  그대로 남고, "네트워크 0회"라는 더 강한 부수 성질만 내려놓았다.
- 8.2 참고: 이 단언 변경은 8.2의 사전 열거 목록(1.1·1.3·5.1·5.2·7.4)에 **없다**. 7.1이
  필연적으로 만드는 변경이므로 Manager 확인 대상으로 여기 기록한다. 같은 파일의
  `fullGate()` 헬퍼도 7.5에서 3개 한도가 추가되면 갱신이 필요하다.
- 기동 비용 변화: 게이트 OFF 엔진이 기동당 `GET /api/v1/accounts` 1회를 새로 낸다.
  §0.4 예산에서 기동 1회는 라인이 아니다(재시작 빈도 ≪ 폴링 빈도).


## 2026-07-26 [safe local] 보존 스윕을 "거부"가 아니라 "보존 기간 확장"으로 냈다 (task 7.2)

- 사실: `PruneSpentNonces(ctx, now, retention)`은 `retention < 최대 결정 TTL`이면
  `ErrRetentionTooShort`로 **아무것도 지우지 않고** 거부한다. 기동 스윕이 이 오류를 그대로
  기동 거부로 올리면, 과거에 긴 TTL로 발급된 결정 행 하나가 영구적으로 엔진 기동을 막는다 —
  아무것도 삭제되지 않는(=안전한) 상황을 장애로 바꾸는 셈이다.
- 이번 처리: 기동 시 `retention = max(30일, MaxDecisionTTL())`로 넓혀서 호출한다. 불변식
  (보존 ≥ 최대 결정 TTL)은 **정의상** 위반될 수 없고, 스윕은 항상 수행된다. DB 오류는 그대로
  기동 거부(journal이 답하지 못하는 상태에서 거래를 시작하지 않는다).
- 30일은 튜닝 값이 아니라 여유 큰 하한이다(현 빌드의 결정 TTL은 분 단위).

## 2026-07-26 [safe local] journal 편입으로 엔진 기동 조건이 넓어졌다 — 의도된 상속 (task 7.2)

- 결과 명시(D8 2단계): 파일시스템 allowlist(ext4/xfs/btrfs)와 무결성 검사가 이제 **엔진**
  기동 조건이다. tmpfs·NFS·fuse에서 기동하면 파일시스템 이름을 말하는 거부가 나온다.
  `openEngineJournal`의 doc comment에 근거를 적었다.
- 경로는 flatten 관례 보존: `--config-dir`가 있으면 `<config-dir>/journal.db`, 없으면
  journal 자체 기본값(`$TOSSOS_DATA_DIR` > `$XDG_DATA_HOME/tossos` > `~/.local/share`).
- 테스트 seam: `engine.Options`에 **비공개** 필드 `journalFSProber` + `export_test.go`의
  setter. `journal.Options.migrationOverride`와 같은 형태이며 빌드 산출물에는 없다.
  `testenv`의 `TestFixedFSProberIsTestOnly`는 생산 파일의 `FixedFSProber` 사용만 금지하므로
  타입 이름(`journal.FSProber`)을 받는 필드는 대상이 아니다.
- 기동 비용 변화: 기동당 journal open(마이그레이션 확인 포함) 1회 + `PruneSpentNonces` 1회.
- 과도기 1건: 7.4 전까지 `tossctl flatten-all`은 engine Context(journal 1개)와 flatten 자체
  journal(같은 경로)을 **둘 다** 연다. 단일 프로세스·순차 오픈이라 동작하지만 의도된 상태는
  아니며 7.4의 자체 배선 전환이 없앤다.

## 2026-07-26 [blocking→해소] Pre-Edit 범위 이탈 1건: `internal/official`에 `AuthHeaders` 추가 (task 7.3)

- 사실: 승계 목록은 "HTTPReplay transport 배선 — 헤더는 official 토큰 매니저에 연결"을
  7.3에 배정했다. `HTTPReplay.Headers`는 `func(ctx) (map[string]string, error)`인데
  official의 토큰 접근(`c.tm.token`)과 계좌 seq 해석(`c.ensureAccountSeq`)이 **둘 다
  비공개**다. 토큰 캐시 파일을 엔진이 직접 읽는 우회는 갱신 정책이 둘로 갈라진다.
- 이번 처리: `internal/official/auth_headers.go` **신규 파일**에 additive 메서드
  `(*Client).AuthHeaders(ctx) (map[string]string, error)` 하나. 기존 파일 0줄 변경,
  기존 호출자 0곳 영향, 새 능력 부여 없음(이 클라이언트를 쥔 코드는 이미 모든 엔드포인트를
  호출할 수 있다). 헤더 이름 상수 2개(`HeaderAuthorization`·`HeaderAccount`)를 함께 내서
  호출자가 철자를 새로 발명하지 않게 했다.
- 위임 범위 판단: 위임된 Pre-Edit는 `app/engine/{engine,interlock}.go`·`config/engine.go`·
  `cmd/tossctl/flatten.go` 4개였고 `internal/official`은 그 밖이다. 태스크 지시문이
  "small official accessor가 필요할 수 있다 — additive 메서드를 넘으면 issues.md에 기록하고
  transport를 nil로 두라"고 했고, 실제로 **additive 메서드 하나로 끝났으므로** 배선했다.
  넘었다면 nil + `[미측정 — 2b]`였을 것이다. **Manager 확인 요망**.
- 안전성: attestation이 nil이므로 재생 진입점은 `replay_not_attested`로 즉시 폴백한다.
  transport는 컴파일·리뷰되지만 실행되지 않는다(dark).

## 2026-07-26 [safe local] `execgw.Gateway.Wiring()` 추가 — 인터록이 "구성됨"을 물어볼 수 있게 (task 7.3)

- 사실: `Options`의 Orders·Entry·Preflight·Replay·Attested는 전부 nil 허용이고, nil이면
  보장이 조용히 사라진다(Orders nil → ack만으로 CONFIRMED). 7.5 인터록의 "Gateway 구성
  확인" 항목은 gateway 포인터가 non-nil인지가 아니라 **무엇이 배선됐는지**를 물어야 한다.
- 이번 처리: `execgw.Gateway.Wiring() Wiring`(bool 5개) 추가. 순수 읽기, 내부 값을 넘겨주지
  않는다(포인터가 아니라 bool). 7.3이 배선하고 7.5가 검증한다.

## 2026-07-26 [safe local] Preflight도 배선했다 — 태스크 문언에 없지만 nil은 "검사 생략"이 아니다 (task 7.3)

- 7.3 문언은 journal·EntryGate·해소기·예약 저장소·Orders·Replay만 열거한다. `Preflight`를
  nil로 두면 주문 형태 검사와 자금 확인이 **사라진다**(failclosed.go: "A nil reader does not
  mean skip the check" — 그건 Preflight 내부 규칙이고, Preflight 자체가 nil이면 호출이 없다).
- 생산 gateway가 fail-closed 검사 없이 서는 것은 스펙 방향과 반대이므로 배선했다.
  `execgw.OfficialAccount{Client: off}`가 입력이며, 진입 결정 발급자가 없는 2a에서 동작
  영향은 0이다.

## 2026-07-26 [safe local] 예약 저장소는 배선할 필드가 없다 — journal 자체다 (task 7.3)

- 7.3 문언의 "예약 저장소"는 별도 주입 지점이 아니다. 3.1이 예약을 `journal.Reserve`/
  `ReserveWithRecollection`으로 냈고 nonce 소비는 1.7이 `MarkDispatchStarted` 트랜잭션에
  병합했으므로, Gateway에 넘길 것은 journal 하나뿐이다(1.7 issues 항목이 예고한 대로
  `execgw.NonceStore`·`Options.Nonces`는 존재하지 않는다).
- 예약 배선의 계약("버전 읽기 → 스냅샷 수집 → Reserve")은 진입 결정을 발급하는 쪽의 몫이고
  2a에는 그 발급자가 없다 — Gateway 구성 단계에서 할 일은 없다.

## 2026-07-26 [safe local] 재생 transport가 자체 http.Client를 든다 (task 7.3)

- 사실: `official.Client`의 `hc`는 비공개이고, 접근자를 하나 더 내는 것은 7.3의 범위를 더
  넓힌다. `HTTPReplay`는 `*http.Client`를 요구한다(타임아웃 없는 재생은 자기가 쓰는 키보다
  오래 사는 재생이다).
- 이번 처리: 엔진이 `&http.Client{Timeout: 15s}`를 만들어 넘긴다. official의 401 재시도를
  **상속하지 않는 것이 옳다** — 재생은 전송 전에 journal에 계수되므로 transport 레벨 재시도는
  계수되지 않은 요청을 보낸다(replay_transport.go의 "No retry" 근거와 같다).
- 결과: 테스트에서 재생 transport의 BaseURL은 httptest 서버지만 HTTP 클라이언트는 기본
  transport다. attestation nil이라 호출되지 않으므로 관측되지 않는다.
- `ReplayConfig`는 기본값 그대로다 — `MaxWaits`·margin 수치는 `[미측정 — 2b]`(왕복 p99 측정
  전 임의 조정 금지).

## 2026-07-26 [safe local] 새 파일 `internal/app/engine/gateway.go` (task 7.3)

- 위임된 Pre-Edit 파일은 `engine.go`였다. 배선 전체를 거기 넣으면 500줄이 넘어 리뷰가
  어려워지므로 같은 패키지의 새 파일로 냈다. 공개 계약 변경은 `Context`의 필드 4개
  (`Gateway`·`Entry`·`Resolver`·`Reconcile`)뿐이고 그것들은 `engine.go`에 있다.
- 네 필드 모두 mutator가 아니다: Gateway는 GuardianDecision 없이는 아무것도 제출하지 못하고,
  Entry·Resolver·Tracker는 브로커 mutation 경로를 갖지 않는다. 봉인 테스트(7.4)가 이를
  `trading.Broker`/`ConditionalBroker` 구현 여부로 확인한다.

---

## Manager 판정 (1차 물결 검증, 2026-07-26)

- **0.1 D9 편차 5건 전부 승인.** (1) attempts.account_ref additive — 1.3이 채우고 intents 값과 불일치 시 거부할 의무 승계. (2) reconcile_states 이중 부분 unique — D9 문장의 의도를 SQLite NULL 의미론에서 실제로 강제하는 유일한 형태. (3) cause CHECK 제외 — 승인하되 **4.1의 저장 계층이 Go 상수 집합을 쓰기 시점에 검증(미지 cause는 거부)하는 조건부**. (4) dedup 컬럼 NOT NULL DEFAULT '' — 승인, **5.3은 NULL이 아니라 ''를 쓴다**. (5) spent_nonces FK 제외 — 보존 불변식 우선, 승인.
- **5.2 잔여 원문 저장 위반의 배정**: `journal/lineage.go:118`(ResolveConfirmedWithLineage의 trim — cancel 해소는 원문·amend 해소는 trim으로 갈라진 상태), `journal/fills.go:173/385/393`, `filldetect/payload.go:84` → **5.3 담당자에게 배정**(fills·payload는 원래 범위, lineage는 동일 해소 쓰기 경로의 정합 회복).
- **round-trip이 Options.Orders 배선 전 무효** → 7.5 인터록 "Gateway 구성 확인"에 Orders 배선 검증 항목 추가로 승계.
- 5.1의 보수 판정(REPLACED without successor → UNKNOWN, CANCELED+successor → UNKNOWN, PENDING_* 무증거) 승인 — 전부 fail-closed 방향.

## Manager 판정 (2차 물결 검증, 2026-07-26)

- **경계값(tie) 의미론**: 변경 불필요 — 두 의미론은 각자의 스펙과 정합한다. 주문 단위 Max는 포함 상한("largest quantity this decision authorises" — 10 허용), 총계 한도는 도달 시 차단(risk-management "한도 중 하나라도 **도달**한 상태 → 거부"). riskcalc.WithinLimit(tie=초과)와 execgw 주문 검사(tie=허용)를 그대로 유지하고 이 구분을 여기 기록한다.
- **interlock.go 기계적 적응(선언된 Pre-Edit 이탈)**: 승인 — 의미 보존·컴파일 수복 한정, 3개 누락 한도와 `Limits.Validate()` 전환은 7.5 승계 확인.
- **d30eaa8의 광역 git add**: 섹션 1 담당자가 5.3 체크박스를 자기 커밋에 포함시킴 — "체크박스=산출물 동일 커밋" 규칙 위반(내용은 정확, 되돌리지 않음). 8.1 diff 리뷰 기록 대상. 이후 물결 지시문에 경로 지정 스테이징 재강조.
- **ReductionIntent의 가격·유형 미결속**: 2a에서 비악용(발급자=호출자). 2c/2d가 발급자를 분리할 때 결속 확장 필수 — 2c 선행 조건 목록에 승계.
- **design.md "D6a" 미존재 참조**: Manager 편집 실수 — 스펙 요구명 참조로 정정.

## Manager 판정 (3차 물결 검증, 2026-07-26)

- **journal → riskcalc 의존 신설**: 승인(태스크가 staleness 상수 사용을 지시). riskcalc는 stdlib-only leaf라 순환 위험 없음.
- **정밀 decimal 계층**: 승인 — 예약 원장의 "합이 한도를 넘지 않는다" 속성에 float 누적 불가(0.1×10 판별 테스트 확인). aggregate 평가의 문서화된 float64는 별개 유지.
- **transition() 해제 훅 + CONFIRMED·UNRESOLVED는 해제하지 않음**: 스펙 그대로 — 승인.
- **audit 실패 시 해제 중단**: 보수 방향 — 승인.
- **영구 불일치 승격을 계좌 전역 QUANTITY_MISMATCH 행으로 영속(신규 cause 미추가)**: 승인 — 0.1 cause 집합 불변 유지가 우선, Restore가 임계값 복원으로 클린 패스 해제를 차단함을 확인.
- **durability.go 비계약 편집 2건(예약 backfill·해제 훅) Pre-Edit 공개**: 승인 — 공개 계약 무변경, 무예약 시 no-op.
- **2.5 [M] 수행**: indoubt.go·retry.go의 "no idempotency key" 서술을 정정(재생=정체 회수, replay.go 참조). P1 아카이브 스펙 정정은 델타(MODIFIED IN_DOUBT 해소·Retry Matrix)가 담당 — archive 시 반영 확인 예정.
- **7.x 승계 확정 목록**: interlock의 누락 한도 3종+Limits.Validate() 전환(7.5), Orders 배선 확인(7.5), Tracker.Restore 기동 호출(7.3), reconcile_states 투영 재구성(7.3), PruneSpentNonces 기동 1회(7.2), HTTPReplay 헤더 배선(7.3 — attestation OFF라 dark), MaxWaits 수치는 [미측정 — 2b].
- 확정 하한의 소비자 부재: 의도됨 — 자동 청산 경로는 엔진 루프 소유 change(2d)의 몫.
