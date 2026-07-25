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
