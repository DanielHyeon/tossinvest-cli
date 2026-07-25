# Issues: harden-execution-base

> 구현 중 발견한 스펙·코드 마찰 기록. advisory only — 권위는 spec + 코드 + 테스트.
> 분류: **blocking**(구현 중단 + Manager 호출) / **safe local**(스펙 의도 명백, 구현하며 기록) / **observation**(후속 task 입력)

## 2026-07-26 [observation] official 어댑터가 canceledAt·페이지 커서를 domain.Order에 싣지 않는다 (task 2.3)

- 사실: `internal/official.adaptOrder`(orders_reads.go:156)는 API의 `canceledAt`을 읽지만
  `domain.Order`에 해당 필드가 없어 버린다. `Orders`도 `nextCursor`/`hasNext`를 버린다
  (design.md D4가 이미 지적).
- 영향: 브로커 상태 파생은 `(status, canceledAt, filledQuantity, quantity, lineage)`가
  필요한데(order-execution "브로커 상태 파생") `domain.Order`만으로는 canceledAt을 얻을 수 없다.
- 이번 처리(2.3): `internal/brokerstate`가 자체 입력 타입 `OrderView`를 두고
  ① `FromDomainOrder(o, canceledAt, lineage)` — 호출자가 canceledAt을 별도 공급
  ② `ParseOfficialOrder(raw JSON, lineage)` — 공식 응답 본문(`{"result":…}` 봉투 포함)에서
  직접 canceledAt까지 파싱. upstream 파일 수정 없음(§불변 규칙 준수).
- 후속 task 입력: **2.9**(pagination 완주 어댑터)와 **3.1/4.1**(체결 감지·cancel/amend 사전
  확인)도 같은 벽을 만난다. 선택지는 (a) execgw 어댑터가 공식 엔드포인트를 raw JSON으로
  직접 읽어 `brokerstate.ParseOfficialOrder`에 넘긴다, (b) `domain.Order`에
  additive-nullable 필드(canceledAt)를 추가한다 — (b)는 upstream 파일 수정이므로 Pre-Edit
  선언 대상이고 CLI/MCP 직렬화 계약(json 태그)에 영향. 2.9 착수 전 Manager 판단 필요.
- 안전 영향: 없음. canceledAt 부재는 파생에서 "취소 증거 없음"으로 처리되어
  CLOSED+미체결/부분체결이 `UNKNOWN_BROKER_STATE`(fail-closed)로 떨어진다 — 오판이 아니라
  차단 방향.

### Manager 결정 (2026-07-26)

**(c) 안 채택**: `internal/official`에 **신규 파일**(orders_raw.go)로 additive 메서드를 추가한다 —
`OrdersPageRaw(ctx, filter, cursor)` 형태로 주문 목록의 raw JSON 항목들과 nextCursor/hasNext를
반환하고 기존 send/token 경로를 재사용한다. 기존 메서드·`domain.Order`·직렬화 계약은 무변경.
brokerstate.ParseOfficialOrder가 raw 항목을 소비한다. (a)는 인증 재구현 중복, (b)는 CLI/MCP
계약 영향으로 기각. 신규 파일 추가라 기존 함수 수정은 없으나 High-risk 패키지이므로 축약
Pre-Edit(대상·근거·테스트) 보고는 유지. task 2.9에서 구현.

## 2026-07-26 [safe local] engine.Context의 raw mutator 필드 봉인 (task 2.5)

- 사실: `internal/app/engine.Context`가 `Broker trading.Broker`·`Conditional
  ConditionalMutator`를 **exported 필드**로 노출했다. 이 두 필드를 잡은 호출자는
  GuardianDecision·journal 기록·IN_DOUBT 처리를 모두 건너뛰고 주문을 낼 수 있다 —
  engine-safety "Guardian 결정 없는 제출 경로는 컴파일·API 수준에서 존재하지
  않아야 한다(raw mutator 미노출)"와 정면 충돌.
- 처리: task 2.5가 execgw.Gateway를 도입하면서 두 필드를 unexported(`broker`,
  `conditional`)로 바꾸고, WTS-isolation 테스트가 요구하는 접근은
  `export_test.go`(빌드 산출물에 없음)의 `BrokerForTest()/ConditionalForTest()`로
  옮겼다. 구조 테스트 `seal_test.go`가 exported mutator 필드의 재등장을 막는다.
  `TradingService`는 config 정책·confirm token 게이트를 지닌 래핑 대상이고 두 mutator
  인터페이스를 만족하지 않으므로 노출 유지.
- 잔존 리스크(후속 task 입력): `Context.Official *official.Client`는 조회용으로
  노출되어 있으나 `PlaceOrder/CancelOrder/ModifyOrder` 메서드도 가진다. 즉 엔진
  wiring을 쥔 코드가 이론적으로 공식 클라이언트로 직접 주문할 수 있다. 봉인하려면
  읽기 전용 인터페이스로 좁혀야 하는데, 2.7/2.9/3.x가 이 클라이언트로 조회를 하고
  게이트웨이 배선이 확정되는 시점은 **task 4.2(기동 인터록)**이므로 그때 함께 좁히는
  것이 최소 변경이다. 현재 노출은 엔진 내부 wiring 한정(CLI/MCP 무관).

## 2026-07-26 [safe local] Manager 결정(c) 구현 시 단건 raw 조회도 필요 (task 2.9)

- 결정문은 `OrdersPageRaw`만 명시했으나, 2.8의 CANCEL/AMEND 해소가 **원주문 단건**을
  `(status, canceledAt, filledQuantity, quantity, lineage)`로 파생해야 한다.
  `OrderByID`는 `domain.Order`를 돌려주므로 canceledAt이 없다 — 같은 벽이다.
- 처리: 동일 신규 파일 `internal/official/orders_raw.go`에 `OrderRawByID(ctx, orderID)
  (json.RawMessage, error)`를 함께 추가했다. 결정(c)의 근거(기존 send/token 경로 재사용,
  기존 메서드·domain.Order·직렬화 계약 무변경)를 그대로 만족하며 기존 함수 수정 0건.
- 테스트: httptest로 path escaping·계좌 헤더·envelope 언랩·canceledAt 보존 검증.

## 2026-07-26 [safe local] journal 스키마 v2 — 체결 스냅샷 테이블 (task 3.2)

- 사실: 3.2의 "누적 스냅샷 멱등 반영"은 **직전 관측이 durable**해야 성립한다. 잃어버리면
  재시작 후 첫 폴이 누적 filledQuantity 전체를 신규 체결로 보고해 포지션을 두 배로 센다.
  journal v1에는 체결 스냅샷을 담을 테이블이 없다.
- 처리: 신규 파일 `internal/journal/fills.go`에 `schemaV2`(`fill_snapshots`,
  `fill_events` + 인덱스)와 읽기·쓰기 API를 추가했다. `schema.go`는 2줄만 바뀐다 —
  `SchemaVersion = 1 → 2`, `migrations`에 `{Version: 2, SQL: schemaV2}` append.
  schema.go가 스스로 규정한 additive migration 절차(“Never edit a released step.
  Append a new one.”)와 §0.6(additive-nullable 선호)을 그대로 따른 것이고, 기존 컬럼·
  기존 함수는 무변경이다.
- 테스트 영향: `schema_test.go`의 `wantTables`·`wantColumns`·인덱스 목록과 schema_meta
  미러 비교(하드코딩 `"1"` → `strconv.Itoa(SchemaVersion)`)를 새 스키마에 맞게 확장했다.
  기존 단언을 약화한 것이 아니라 신규 테이블을 계약에 추가한 것이다.
- rollback: 구버전 바이너리는 `user_version=2`를 보고 `ErrSchemaTooNew`로 **기동 거부**한다
  (오독이 아니라 정지 — 라이브 계좌에서 안전한 방향). 데이터 손실 없음.
- 안전 영향: 없음. 새 테이블은 기존 주문 경로가 읽지 않는다.

## 2026-07-26 [observation] task 3.4는 upstream 파일을 수정하지 않았다 — Pre-Edit 불요

- 사실: tasks.md 머리말이 Pre-Edit 전문 대상으로 **3.4**를 지목한다. 계획 시점에는
  reconcile 스냅샷이 `internal/official`·`domain.Order`를 건드려야 할 것으로 봤기 때문이다.
- 실제: 2.9가 이미 `OrdersPageRaw`/`OrderRawByID`를 추가해 두었고 execgw가 완주 루프를
  갖고 있어, 3.4는 **신규 패키지 `internal/reconcile`만으로** 구현됐다. upstream 파일
  수정 0건이므로 Pre-Edit 전문 작성 조건("upstream 파일을 실제 수정할 때")에 해당하지 않는다.
- 유일한 기존 패키지 변경은 `internal/journal/fills.go`(3.2에서 신규 생성한 파일)에
  `NetPositions` **함수 추가** — additive, 기존 함수 무수정.
  gross 합계(`FilledQuantities`)는 거래량 질문에 답하지 실물 포지션 질문에 답하지 못한다
  (SELL 체결이 노출을 줄이므로). 왕복매매마다 허위 불일치가 나므로 부호 있는 집계가 필요했다.

## 2026-07-26 [observation] EntryGate에 심볼 차원이 없다 (task 3.2/3.6)

- 사실: fill-detection 스펙은 UNKNOWN_BROKER_STATE 시 "해당 **심볼**이 차단"을,
  reconciliation 스펙은 "차단 범위(계좌/시장/심볼)"를 요구한다. 그러나
  `execgw.EntryGate`의 latch는 reason 단위 계좌 전역이고, Gateway의 심볼 단위 검사
  (`checkSymbolFree`)는 journal의 pending/unresolved attempt만 본다 — 외부에서 심볼
  차단을 주입할 훅이 없다.
- 이번 처리: (a) `filldetect`는 스냅샷이 fail-closed면 `ReasonBrokerStateUnknown`으로
  **계좌 전역** latch를 건다 — 스펙보다 넓은 차단이지만 §0.9(불명확하면 보수적)의
  방향이고, 청산 경로는 §0.3대로 열려 있다. (b) `reconcile.Blocks`(3.6)가 계좌/시장/
  심볼 범위와 해제 조건을 상태표로 보유하고, 계좌 범위 항목만 EntryGate에 동기화한다.
  심볼 범위 항목은 조회 API로 노출한다.
- 후속 task 입력: **4.2(기동 인터록)**에서 게이트웨이 배선이 확정될 때
  `EntryGate.CheckEntry()`에 심볼 인자를 더하는 대신 `reconcile.Blocks`를 게이트웨이가
  참조하도록 배선하면 심볼 단위로 좁힐 수 있다. Gateway 기존 함수 수정이 필요하므로
  Pre-Edit 대상이다.

### 해소 (2026-07-26, task 4.2)

`reconcile.Blocks` 참조 대신 **EntryGate에 심볼 차원을 추가**하는 쪽을 택했다. 이유:
`filldetect`도 같은 벽을 만나는데(3.2) reconcile을 참조하게 하면 filldetect →
reconcile 의존이 생기고, 두 생산자가 같은 게이트에 서로 다른 경로로 접근하게 된다.
차단 상태의 단일 소유자는 게이트다.

- 신규 파일 `internal/execgw/symbolgate.go`: `BlockSymbol`/`ClearSymbol`/
  `ClearSymbolReason`/`SymbolBlocks`/`CheckEntryFor(market, symbol)` — 전부 additive.
  `retry.go`는 필드 1줄(`symbolLatches`) 추가와 `CheckEntry`를 `CheckEntryFor("","")`
  위임으로 바꾼 것뿐이다. **`CheckEntry()`의 의미는 불변**(계좌 전역 질문)이라 기존
  호출자의 답이 바뀌지 않는다 — 좁아진 것은 게이트웨이가 던지는 질문이다.
- `execgw.Gateway.checkEntry`(unexported) 1줄: `CheckEntry()` → `CheckEntryFor(market, symbol)`.
- `reconcile.Tracker.syncGate`: symbol 범위 행은 `BlockSymbol`, account 범위 행은
  기존대로 `Block`. 이 함수는 3.6에서 신규 작성한 것이고 상태표(BlockRules)가 이미
  규정한 scope를 그대로 따르게 만든 것이다. 기존 테스트
  `TestQuantityMismatchBlocksEntries`가 "계좌 전역 latch"를 단언하고 있어 스펙 상태표
  기준(심볼 차단 + 다른 심볼은 거래 가능)으로 갱신했다 — 단언 약화가 아니라 3.6이
  기록한 "스펙보다 넓은 차단"의 해소다.
- §0.3: 심볼 차단은 진입 전용이다. `symbolgate.go`에 청산을 막을 수 있는 메서드는
  없고, `TestGatewayNeverGatesAnExitOnASymbolBlock`이 차단된 심볼의 cancel 성공을
  고정한다.

## 2026-07-26 [safe local] Context.Official 봉인 완료 (task 4.2)

- 위 2.5 항목의 잔존 리스크를 해소했다. `Context.Official`의 타입을
  `*official.Client` → 신규 read-only 인터페이스 `engine.OfficialReads`(신규 파일
  `internal/app/engine/reads.go`)로 좁혔다. 이 인터페이스는 `PlaceOrder`,
  `CancelOrder`, `ModifyOrder`, `Create/Cancel/ModifyConditionalOrder`를 선언하지
  않으므로 엔진 wiring을 쥔 코드가 그 호출을 **표기할 수 없다**.
- 구체 클라이언트는 `Context.official`(unexported)로 남는다. 테스트 접근은
  `export_test.go`의 `OfficialClientForTest()` — 빌드 산출물에 없다.
- `seal_test.go`에 구조 테스트 2건 추가: 필드가 인터페이스일 것 + mutator 메서드
  부재, 그리고 그 mutator 이름들이 실제로 `*official.Client`에 존재할 것(오타로
  테스트가 공허해지는 것 방지).
- 사용처 영향 0건: `Context.Official`을 읽던 production 코드는 없었다(engine_test 2곳뿐).

## 2026-07-26 [safe local] 엔진 테스트가 실 데이터 디렉터리에 쓸 수 있었다 (task 4.2)

- 사실: `internal/app/engine`의 `isolate(t)`는 `XDG_CONFIG_HOME`·`XDG_CACHE_HOME`만
  격리했다. 4.2가 audit 로그를 `journal.DataDir()`(= `$TOSSOS_DATA_DIR` >
  `$XDG_DATA_HOME/tossos` > `~/.local/share/tossos`) 아래에 두면서, 격리되지 않은
  데이터 디렉터리에 테스트가 파일을 쓰게 될 뻔했다.
- 처리: `isolate`가 `XDG_DATA_HOME`과 `TOSSOS_DATA_DIR`도 `t.Setenv`로 임시 경로에
  고정한다. task **4.6**의 격리 헬퍼가 이 패턴을 공용화한다.
- 안전 영향: 없음(발견 시점에 커밋된 적 없음). 다만 "문구가 아니라 테스트 인프라가
  막는다"는 불변 규칙의 실례이므로 기록한다.
