# consumer 증거 보완 delta

부모 89c9a5b7eb9b1ea303e68fe60272ef84502eb557에서 생산 source 변경 없이 `a166_consumer_handoff_test.go`에 두 테스트를 추가했다. Go1.25 선택 normal/tagged race 각각 두 이름만 실행, 원문 gzip와 SHA 보존. 별도 user/network namespace lo-only/no-route 및 local TLS exact DialContext, 가짜 token/cache만 사용했다.

- `TestA166ConsumerToken429CrossFactorySnapshot`: 서로 다른 factory Budget에 자기 client 결속, token429 뒤 GET0 및 다른 factory의 Budget.ObservationSnapshot에서 Token event1·cursor 전진.
- `TestA166ConsumerPriceLimitsTwo401AdoptAndMint`: 단일 actual PriceLimits logical 요청에서 기존 cached token GET401→다른 cached token 채택 GET401→새 token 교환→GET200. 실제 count GET3/token1/first plaintext Write4이며 pending 종류 Other/Other/Token/Other와 각 시작 간격 ≥350ms. 같은 logical context/collector를 전달한다.

과거 `TestA166RuntimeToken401AndCached`의 5 wire는 첫 GET 401 이후 성공한 첫 논리 호출과 별도의 cache-hit GET을 합친 횟수다. 두 GET이 401이었다는 이전 해석은 오류이며 기존 영수증은 삭제하지 않는다. 기존 `TestA166ConsumerPriceLimits401EachWire`는 token200→GET401→token200→GET200의 별도 4 wire fixture다. source SHA 및 최종 착지 SHA는 Manager 인계 기록에서 고정한다.

관련 gstack 수동 점검: fixture가 다른 주소의 연결을 거부, factory 별 pointer/ownership, 동시 pending, cache/401 조건·attempt 및 기존 nil 유지 확인. 전체 gstack 자동 pipeline·외부 모델 독립 리뷰는 미실행. live/운영/전체 tests·gate/a087 활성화/a169 구현 미실행.
