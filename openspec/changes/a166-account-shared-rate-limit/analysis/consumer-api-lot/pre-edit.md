# 사전 분석

기준 6e0968222. 기존 ObservationStore.Snapshot은 pin→cursor 검사→event lock/read→view 사본→pending이면 wire lock 및 generation 재검사, 오류 그대로 반환한다. 기존 Budget/Factory는 observed store를 private 보유하며 Factory.Close가 소유한다. 신규 Budget.ObservationSnapshot은 nil/context/closed/unobserved 거절 후 기존 Snapshot에 위임하며 store pointer나 새 권한을 내놓지 않는다. 기존 함수 본문 미변경, 신규 함수만 추가하므로 기존함수 FLM 면제. 기존 Snapshot의 AST/source는 변경하지 않는다.

BTM: missing method compile RED→정상·nil ctx/nil budget/closed/unobserved·wrong/future/past cursor·변경된 view 사본·ring/pending 보존; guard 및 store 위임 제거 유효 변이. 별도 v2 두 factory 포인터 불일치+각 client UsesAccountBudget, token/Other429 공유 event, PriceLimits Auxiliary 첫 admission 거절 wire0·401 각 시도 및 소유권 보존 local TLS fixture. 모든 network는 user/net namespace lo-only/no-route 및 exact DialContext; fake만 사용. 생산 a087 nil board/token, a169 미변경. source search: 기존 Snapshot/Factory/transport 재사용, dependency 추가 없음; UI N/A. 전체 test/gate·운영 금지.
