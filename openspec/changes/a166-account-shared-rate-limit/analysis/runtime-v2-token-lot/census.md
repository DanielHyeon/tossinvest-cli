# source census — 고정 기준 e5c9d0256

`rg`의 비test Go 정적 검색에서 Budget open 호출은 internal/accountbudgetruntime/runtime_linux.go의 OpenObserved 한 곳이다. production NewTransport 호출은 official/account_budget.go, official/a112_mbus_read.go, execgw/replay_transport.go, tools/a087-market-sell-probe/send.go다. 런타임의 official.NewWithAccountBudget는 같은 factory Budget을 넘긴다. 해당 wrapper는 인자로 전달된 Budget을 사용하며 raw low-level API의 직접 사용 가능성은 자체적으로 없애지 않는다. 이 검색은 함수명 정적 census이며 runtime 전체 값 흐름과 모든 외부 참여자 완료 증명이 아니다.

v2 test-only fixture는 root/queue/wire/event.lock을 임시 provision 후 외부 manifest에 pin한다. existing child-process 첫 Write fixture는 RootCA 유지·exact DialContext pin으로 local TLS에 제한했다. a112 테스트는 기존 official constructor 소유 private authorityTransport의 RootCAs와 DialContext만 reflect/unsafe로 사전 설정한다. field 이름·타입·non-nil 및 UsesAccountBudget를 검사하며 authority pointer/base/flag를 바꾸지 않는다. fake broker hostname 인증서와 local TLS만 사용한다; 운영 authority 증명 아님. replay/probe 테스트도 fake 주문 body와 local TLS fixture만 사용하고 도구를 실행하지 않는다.

Retrier 소비·token 429 후 Query 중단 정책·production 전 caller value-flow 전체 검사·운영 manifest 전환은 후속 범위다. event 관측은 runtime required에서 켜졌지만 a166 전체 완료·활성화 증거는 아니다.
