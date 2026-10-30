# verify factory·console broker·HTTP API 사전 분석

기준 bbe5066ec. 함수별 수정 전 AST와 source SHA는 옆 파일. CodeGraph worktree 인덱스 없음; 분기 설명은 원문 손읽기. 동결된 internal/console lifecycle 후보와 별도이며 해당 파일은 수정하지 않는다.

| 함수/경로 | 현재 분기와 불변식 | 시험할 변경 |
|---|---|---|
| buildVerifyBroker | 자격 실패·없음 거절; Accounts에서 ref/seq; seq 0 기존 lazy, 양수 새 client | 실제 ctx·동일 root factory; required 오류 wire0; seq 고정·재생성 |
| runVerifyRun/Abort, consoleVerifyStarter | 기존 lock·record/승인 순서 뒤 client 생성, console은 run마다 새 client | 각 현재 ctx 사용; 기존 lock·order 유지 |
| newConsoleBroker/resolve/instrumentMetadata | lazy 성공만 cache, 실패 재시도; 현재 resolve는 Background 대기, metadata 별도 ctx builder | request와 signal lifetime 모두 취소; concurrent 단일 성공과 실패 재시도 |
| TypedMarketCalendar, lazyHoldings, lazyOrders | 각 읽기별 기존 결과·실패 분류 | 동일 요청 ctx로 broker 대기 |
| buildConsoleAccountBroker | credentials 후 Accounts | 첫 token부터 root factory, 기존 선택 유지 |
| openHTTPAPIResources/readPositionBrokerSnapshot/Performance | accountRef context 없는 closure, 요청 context는 나머지 읽기에 존재 | accountRef 실제 요청 ctx 전달, JSON/status 불변 |
| runHTTPAPI | server 선오류 즉시 반환, signal만 5초 Shutdown, publisher cancel/join | 모든 종료 경로 5초 Shutdown, 오류 결합·실패 Close, 완료 과장 금지 |
| runConsole | signal child context 생성 후 lazy broker 구성 | child lifetime 전달, 다른 UI/재실행 로직 별도 후보 |

BTM: required invalid·취소 wire0; OFF/seq 0과 양수 account 일치; token/Accounts local TLS; 실패 후 retry·동시 cache; request와 lifetime 양쪽 취소; server 선오류와 shutdown 실패 원인 보존. 취소·토큰·계좌 테스트는 격리 network namespace와 loopback 고정 fixture만 사용. HTTP 응답·auth·mutation 경로 변경 금지. 전체 tests/gate·운영·live 금지.
