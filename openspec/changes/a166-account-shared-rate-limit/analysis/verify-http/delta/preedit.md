# Read lifetime P1 수정 전 분석

기준 3dc0c776b. 기존 resolve/instrumentMetadata가 만든 child context는 반환 때 취소되지만, 네 개의 read 진입점은 원래 request ctx로 후속 network read를 수행한다. signal lifetime이 bootstrap 뒤 read에 전파되지 않는 P1. 함수별 AST는 옆 파일이다.

TypedMarketCalendar·lazyHoldings.Holdings·lazyOrders.Orders·lazyInstrumentNames.Names는 각 진입 시 request와 signal lifetime을 결합하고 실제 read가 끝날 때까지 유지한다. 원래 request 취소도 유지하며, 기존 resolve 자체의 결합은 직접 accountRef 호출자를 위해 유지한다. 기존 반환·오류/부분결과 분기와 cache·retry는 바꾸지 않는다. cached client bootstrap 뒤 실제 read 대기 중 signal 취소, read 진입 전 취소 wire0, 원래 request 취소, cache 실패 후 재시도와 우회 변이를 확인한다. local fixture만 사용한다.
