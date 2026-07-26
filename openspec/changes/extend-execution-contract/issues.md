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
