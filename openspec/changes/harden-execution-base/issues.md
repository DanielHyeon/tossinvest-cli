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
