# consumer API 인계 후보

기준 6e0968222. 새 Budget.ObservationSnapshot은 nil budget/context 및 closed/unobserved를 거절하고 private store Snapshot을 위임한다. 기존 함수 body 변경 없음; 사전 분석/BTM은 pre-edit.md. 소비자는 Factory 소유권을 얻지 않고 snapshot이 다음 wire의 원자 허가·readiness·운영 승인이라는 주장도 하지 않는다. a087 production nil, a169 미구현 보존.

- 동일 group/pin의 서로 다른 두 v2 factory는 별도 Budget 포인터이고 각 official client가 자기 Budget에 결속된다. 한 factory token/Other429를 다른 factory의 Snapshot이 관측하고 과거 cursor 재조회·future/wrong cursor 거절, view 사본 변조 무영향, closed/unobserved/nil 및 Snapshot 중 Close pin을 확인했다. 기존 cursor/ring 및 generation-race 시험도 선택 실행했다.
- 실제 official.ProtectivePriceLimits Auxiliary: critical 보유 최초 거절 token/GET0; token이 이미 전송된 후 경쟁은 token1/GET0로 별도 확인. 실제 token200→GET401→token200→GET200 네 wire의 first plaintext Write 간격 ≥350ms 및 pending sequence/종류를 확인. 본문 처리 중 보호 Critical 대기 및 EOF/cancel 후 획득, 별도 observed response Body.Close 후 획득 확인. 429 Finish postcommit fsync 오류는 cleanup 보고와 event 상태를 남기며 원래 ErrRateLimited/HTTP 의미를 NOT_SENT로 낮추지 않는다. timestamp 의미·실거래·board/new token/5분정지 및 a169 event abort는 미검증.
- 위임 제거·nil Budget guard 제거 변이 RED. nil context guard만 제거한 변이는 기존 Snapshot이 별도 nil 검사를 수행하므로 동등변이로 살아남았으며 유효 변이로 세지 않는다. 선택 Go1.25 normal/tagged race의 원문과 컴파일 RED 및 변이 원문은 gzip 보존. fake credentials, local TLS exact DialContext, 별도 user/net namespace lo-only/no-route. Darwin accountbudget compile-only는 구현 컴파일 확인이지 운영 지원 PASS가 아니다.
- `TestA087ProductionBoardAdmissionFlowIsOnlyLiteralNil`, `TestA087NilAdmissionMakesNoLowerLimitWire` 선택 실행. a087 활성화/인간 실측 및 a169 소비자 구현은 별도 미완. 전체 tests/gate·실제 broker/운영·activation/archive 미실행.
- gstack 관련 API 수명·오류·동시성·사본·wire/취소 항목 수동 리뷰만 수행, 자동 mixed-model 전체 pipeline·독립 리뷰는 미실행. 신규 dependency 없음, impeccable N/A backend-only. 테스트 전용 unsafe는 firstWrite 관측 및 postcommit fsync fault만 주입하며 제품 authority를 부여하지 않는다.

후속 delta는 `delta-review.md`와 `handoff-table.md` 참조. 실제 PriceLimits 단일 논리 요청의 401→401→200은 cached 최초 GET + 채택 token GET + 새 token POST + 성공 GET, 총 4 wire다. 기존 generic `TestA166RuntimeToken401AndCached`의 5 wire는 401 한 번을 포함한 두 논리 호출의 합이다. 두 증거를 합쳐 한 시험이 모두 수행했다고 주장하지 않는다.
