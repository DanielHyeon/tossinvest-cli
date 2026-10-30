# a166 엔진 poll 설계 독립 재검토

대상: 수정된 engine-poll-observation-plan.md, source 기준 origin/main e5c9d0256. 원문 손읽기, 코드 수정·시험·새 AST 실행 없음. 판정: 기존 두 P1 해소, 현재 문서 범위에서 새 P0/P1 발견 없음. 선행 runtime v2/token의 독립 검증·게시 뒤 문서에 적힌 사전 AST/FLM·RED/GREEN 및 독립 구현 검증을 조건으로 구현 착수 가능하다. 운영·활성화 승인이 아니다.

- 이전 P1 collector: 문서 Query 흐름 다음 단락은 같은 collector의 RateLimited/LocalError를 Token 포함 다음 wire의 Acquire/Begin 전에 검사하고 body exactly-once·ErrNotSent·원 관측 오류를 보존한다. 첫 HTTP 응답과 중간 401의 기존 의미는 바꾸지 않는다. token200/GET401 Finish 실패 후 실제 후속 wire를 막는 post-commit fault RED와 token429 fixture도 명시해 반례를 닫았다. 이미 진행 중인 동시 wire의 소급 중단은 주장하지 않는 범위다.
- 이전 P1 foreign reader: 신원과 키 절에서 required의 Context.ReconcileDriver 외 NewTracer, NewExitObserver, PositionPolicyCommandService reader를 동일 client 기준으로 검사하고 foreign 구현을 생성 전 거부한다. OFF와 별도 저수준 fixture 주입 보존, uncomparable interface panic 방지까지 명시했다. 값 흐름·mutation은 구현 수용 조건이다.
- 생산 Query 9곳을 손읽기로 대조했다: tracer.go:360/366/372/381 네 곳, exitloop.go:868 한 곳, exitwiring.go:207/231 두 곳, adoption.go:308 한 곳, position_policy_command.go:331 한 곳. 재편입은 position_policy_command.go:325–331의 기존 가격 read에 한정되며 원장 명령·권한은 변경하지 않는다는 설계가 반영됐다. 동일 account/kind의 보호·일반·adoption·재편입은 동일 slot을 쓴다.
- 현재 official send는 401에서만 refresh/retry하고 429는 classify 후 반환한다(client.go:335–378, token.go:110–127). 일반 GET429/token429가 정상 반환된 뒤 추가 GET이 시작되는 분기는 확인하지 못했다. response Finish fault 뒤 후속 wire 반례는 새 latch 수용 조건으로 별도 처리됐다.
- identity는 interlock.go:695–724의 같은 official.Client 첫 Accounts 행·양의 seq·SelectedAccountSeq 일치를 사용한다. accountRef·seq와 client/Budget 재검사, 길이 구분된 scope/instance/account/role/kind key, 동일 kind hint 공유, token/주문·flatten 제외, OFF 기존 retry/freshness 보존이 문서화됐다. 키 hash는 cross-key 물리 계좌 동등성 증명이 아니다.

P2 제안: 선택 수용 항목에 PositionPolicyCommandService의 재편입 QueryPrice가 동일 slot을 사용하고 원장 명령·권한/거절 결과가 그대로임을 직접 명시한 fixture를 한 줄 추가하면 9번째 경로의 확인 범위가 더 분명하다. 현재 본문과 생산 closure census 조건은 이 경로를 포함하므로 설계 P1로 분류하지 않는다.

정상 응답·장애시 wire 수와 freshness, AuthFatal latch/escalation, 503 단독, 동시 collector 격리 및 OFF는 구현 후 증거가 필요하다. 전체 tests/gate, 실제 broker·운영·a087/a169 활성화는 미실행.
