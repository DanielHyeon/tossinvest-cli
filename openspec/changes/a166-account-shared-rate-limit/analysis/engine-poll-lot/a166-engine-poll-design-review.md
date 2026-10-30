# a166 엔진 poll 설계 독립 검토 — 보완 필요

대상 engine-poll-observation-plan.md 현재 초안, source 기준 origin/main e5c9d0256c4cbc8dafacc483c3fdf13bc928f3f4. e5와 확인용 checkout 사이 engine/execgw/official 파일 diff 없음. 코드·시험 실행 없음; 제어 흐름은 원문 손읽기이며 새 AST 증거를 생성하지 않았다. 3/s·보호 대기는 이미 확정된 정책이다.

## P1 — collector의 관측 실패 후 같은 공식 호출 내부 후속 wire 차단이 명시되지 않음

설계 23–27행은 fn 반환 후 collector를 검사한다. e5의 internal/accountbudget/transport_linux.go:215–227은 Acquire 뒤 scope를 얻고 기존 collector 상태를 검사하지 않는다. 같은 파일 329–345행에서 header 이후 Finish 실패는 collector/callback에만 기록하고 응답은 유지한다. internal/official/client.go:335–375에서 실제 token 200의 Finish 실패 후에도 send는 GET을 만들 수 있고, GET 401의 Finish 실패 후에도 refresh 및 재전송을 시도한다. 이것은 429 자체의 재시도가 아니라 관측 불완전 상태에서 fn 반환 전에 생기는 후속 wire다. runtime-v2-token-plan의 token collector 상속이 구현되면 발생 가능한 조합이며, 현재 미연결 단계에서 실제 재현됐다고 주장하지 않는다.

최소 처분: observed Query collector에 이미 RateLimited 또는 LocalError가 있으면 Token을 포함한 같은 논리 Query의 후속 실제 전송을 Acquire/Begin 전 거절하고 body close 및 ErrNotSent·원인 보존을 명시한다. 최초 HTTP 응답·body는 바꾸지 않는다. 각 Query의 collector만 사용하고 이미 진행 중인 동시 요청에 대한 소급 중단은 주장하지 않는다. token200→Finish 실패→GET 0, GET401→Finish 실패→refresh/재GET 0, token429→GET 0의 실제 official.Client local TLS fixture와 collector 간 격리·body close 테스트가 필요하다. 무관한 collector 없는 OFF/다른 요청의 의미는 유지한다. 원래 최종 fn 오류가 AuthFatal이면 기존 latch/escalation을 먼저 수행한다. 선행 401이 기존 send의 refresh에 의해 흡수된 경우까지 새 AuthFatal 정책으로 확대하는지는 별도 결정 없이 가정하지 않는다.

## P1 — adoption의 동일 client 값 흐름 범위를 명시적으로 닫아야 함

설계 7·13·35행은 tracer/보호 조회와 동일 client proof를 요구하지만, internal/app/engine/adoption.go:300–308의 adoption price도 같은 Retrier.Query(QueryPrice)를 사용한다. internal/app/engine/reconcileloop.go:365–372는 Retrier를 c.Retrier로 고정하되 caller가 opts.Prices를 지정하면 이를 보존한다. 실제 cmd/tossctl/engine.go:662–666은 Prices를 전달하지 않아 기본 c.Official을 사용하지만, 공개 Context.ReconcileDriver 경로 전체가 반드시 같은 client라고는 증명되지 않는다. 이 차이를 실제 현재 CLI 공격 경로라고 주장하지 않는다.

최소 처분: required production 조립에서 adoption 포함 모든 Query closure의 read 구현이 검증된 동일 official client 또는 검증 가능한 동일 adapter임을 구체적으로 고정한다. caller 제공 다른 Prices는 required에서 생성/호출 전에 거절하거나 그 경로를 명시적으로 비지원 처리한다. OFF와 저수준 fake fixture의 기존 주입을 보존한다. 실제 cmd의 기본 경로·foreign Prices 거절·fn 0, exitSideRetrier 복사 및 동일 kind 보호/일반/adoption 공유를 시험한다. 새 범용 권한 계층은 필요 없다.

## 확인한 경계

- internal/app/engine/interlock.go:695–724의 resolveAccountRef는 동일 official.Client Accounts 첫 행의 trim된 accountRef·양의 seq와 SelectedAccountSeq 일치를 확인한다. engine.go:500 및 gateway 결속을 유지해야 하며, 임의 설정 문자열이나 hash 자체를 계좌 동등성 증명으로 사용할 수 없다. 설계의 길이 구분된 scope/instance/account/role/kind key와 seq 별도 검사 방향은 적절하다. alias·cross-key 물리 계좌 동일성은 미증명이다.
- tracer.go:360–384, exitloop.go:868 및 exitwiring.go:207–240의 각 closure는 직접 한 read를 호출한다. 확인한 official Orders/Holdings/BuyingPower/Prices/SellableQuantity는 호출별 단일 API 함수를 사용하며, getAcct의 선행 계좌 확인은 실패 시 반환한다. client.go:359–378의 자동 재시도 조건은 401뿐이다. 일반 GET 429 또는 token exchange 429가 정상적으로 반환된 뒤 추가 GET으로 넘어가는 분기는 손읽기에서 발견하지 못했다. 단 header 429와 body 오류는 collector가 구분해야 하고, 위 Finish 실패+401/200 반례는 별도다.
- 설계 23–25행은 AuthFatal 우선과 collector429 또는 fn nil에서 RecordSuccess 0·같은 Query 재시도 0을 명시한다. Query의 기존 classifier는 context 취소를 auth보다 먼저 분류한다(internal/execgw/retry.go:69–80). 실제 구현에서는 원 fn 오류의 기존 분류와 관측 오류를 합치는 순서를 시험해 기존 latch 동작을 보존해야 한다. 503만으로 새 429를 만들지 않는다.
- canonical design:73의 동일 accountRef+RequiredQuery hint를 보호·일반 조회가 공유하는 방향과 critical priority의 hint 비우회는 일치한다. token과 주문·flatten에 poll hint를 붙이지 않는 구분도 명시됐다. OFF는 원래 retry/freshness를 유지해야 한다. 실제 v2 factory·token collector는 후속 로트이므로 지금 구현·운영 완료를 뜻하지 않는다.

P0 없음. 위 P1 두 항목이 설계에 구체화되고 source/AST 값 흐름과 선택 fixture로 검증되기 전 구현 승인 판정을 권고하지 않는다. 전체 테스트/gate·실 브로커·운영·코드 수정은 하지 않았다.
