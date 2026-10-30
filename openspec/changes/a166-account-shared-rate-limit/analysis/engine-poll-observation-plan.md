# 엔진 Retrier의 응답별 429와 영속 hint 결속

독립 재검토에서 P1 두 건 해소·새 P0/P1 없음. Manager는 선행 runtime v2/token 로트의 독립 검증·게시 뒤 이 범위 구현을 승인한다. Manager가 main `2938622a8`의 retry.go, engine NewContext/buildGateway/resolveAccountRef, tracer.freshenGateInputs, ExitObserver.observe, reconcileFloor, exitSideRetrier와 official Accounts/SelectedAccountSeq/UsesAccountBudget를 손으로 읽었고 이후 adoption/position policy까지 생산 Query 9곳을 대조했다. `runtime-v2-token-plan.md`와 기존 store/관측 transport 계약을 재사용한다.

## 정책과 역할

canonical design:71–77을 구현한다. 검증된 engine Retrier의 역할은 `engine-poll` 하나다. tracer의 네 필수 조회, 보호 가격 조회와 보호 floor의 holdings/sellable 조회, adoption의 QueryPrice, 엔진 소유 PositionPolicyCommandService의 재편입 사전 가격 QueryPrice가 동일 accountRef·RequiredQuery이면 같은 hint를 공유한다. 여기서 역할은 실행 주기의 이름이 아니라 검증된 엔진 Retrier의 필수 조회 묶음이다. 재편입의 원장 명령·권한 절차 자체는 변경하지 않는다. critical 우선순위는 보류 hint를 우회하지 않는다. 주문 본체, token, flatten과 별도 walk에는 이 hint를 붙이지 않는다. 보호 read를 제외하는 새 정책이나 계좌 전역 cooldown을 만들지 않는다.

required 결속이 없는 OFF Retrier는 기존 fn·retry·freshness·AuthFatal 동작을 유지한다. required production engine에서는 결속 생성 실패를 startup 오류로 돌려주며 nil로 내려가 기존 경로를 쓰지 않는다. readiness나 거래 권한을 이 결속으로 승인하지 않는다.

## 신원과 키

NewContext의 기존 resolveAccountRef가 같은 official.Client의 Accounts 첫 행을 읽고 trim한 DisplayName, 양의 seq, SelectedAccountSeq 일치를 확인한 직후 그 결과를 사용한다. 추가 broker identity GET이나 선택 계좌 정책 변경은 없다. private engine assembly에서 이 증거를 보존해 기존 accountRef/official/borrowed Budget와 함께 poll binding을 만든다. 임의 옵션 문자열을 검증된 계좌로 간주하지 않는다. 생성 및 Query마다 동일 client의 UsesAccountBudget와 저장된 양의 selected seq, Retrier.AccountRef의 정확한 byte 일치를 검사한다. 변동·nil·교체 client는 fn 호출 전 거절한다. source/AST 값 흐름으로 실제 Query closure가 그 client/동일 adapter를 쓰는지 확인한다. 저수준 시험 생성자와 fake reads는 명시적으로 분리한다.

키는 SHA-256(domain `tossos/a166/engine-poll/v1`, Budget의 immutable scope와 instance, 정확한 accountRef, role `engine-poll`, RequiredQuery 값)을 각 필드 길이 구분으로 인코딩한다. separator 연결이나 문자열 normalize를 추가하지 않는다. 계좌 seq는 key가 아니라 실제 client 결속 검사에 사용한다. 같은 계좌 번호가 key별 seq 차이로 나뉘어 영속 hint가 사라지는 설계를 피한다. RequiredQuery는 네 실제 값 전체만 허용하며 unknown은 required에서 fn 0이다. accountRef·키를 로그/이벤트에 출력하지 않는다. 같은 byte 계좌 번호의 재시작/동일 group 공유를 보장하며 cross-key alias의 물리 계좌 동등성은 입증하지 않는다.

최소 API는 기존 Budget의 불변 observed store에 제한된 key 생성과 PollRemaining을 위임하는 방법을 우선한다. mutable store setter, 별도 global map, 신규 authority framework를 만들지 않는다. store 포인터를 반환해야 한다면 borrowed 수명과 Factory.Close 소유권을 명시하고 nil/closed/unobserved를 오류로 취급한다. 실제 API 이름·표현은 구현 전 FLM에 고정하며 위 계약을 바꾸는 편의 API는 독립 재검토한다. Retrier의 결속은 생성 후 불변이며 exitSideRetrier 값 복사도 같은 결속을 읽는다. Query collector는 매 호출 새로 만들고 Retrier 필드에 저장하지 않는다.

API 대조 후 Manager 확정: Budget.EnginePollKey(accountRef, kind)와 Budget.PollRemaining(ctx, key)의 좁은 메서드를 추가하고 store pointer를 노출하지 않는다. 두 메서드는 nil/closed/unobserved를 fail-closed로 처리하며 전자는 위 여섯 필드의 길이 구분 hash만 반환한다. 네 RequiredQuery 값 검증은 execgw binding이 맡는다. Retrier에는 unexported 불변 binding을 두고 기존 설정을 복사해 binding을 붙이는 execgw 생성자를 engine의 private assembly에서만 호출한다. 새로운 일반 setter나 전역 상태를 추가하지 않는다. 동일 reader 판정은 구체 *official.Client type assertion과 pointer 비교로 구현하여 interface의 비교 불가 값에 panic을 내지 않는다. 키 생성이나 생성자 호출 자체가 broker 계좌 증거를 만드는 것은 아니며 앞 문단의 resolveAccountRef source 결속을 유지한다.

required의 공개 assembly 경계도 닫는다. Context.ReconcileDriver의 caller-supplied Prices는 nil이면 기존 공식 client를 쓰고, non-nil이면 검증된 바로 그 official.Client인 경우만 허용한다. foreign reader/adapter는 생성 전에 거절하며 OFF 주입은 보존한다. NewTracer의 Reads, NewExitObserver의 Prices와 PositionPolicyCommandService의 reader도 required binding에 대해 같은 검사를 적용한다. Context.ExitObserver는 현재처럼 Context의 공식 client를 강제로 쓰고 private reconcileFloor는 같은 client로 만들어짐을 고정한다. 단순 interface 동등 비교로 uncomparable 구현에 panic을 내지 않는다. 실제 지원 adapter가 있으면 구체 값 흐름으로 동일 client를 입증하고 허용 목록을 문서화한다. 임의 closure나 실행 중 악의적 Go 메모리 편집을 일반 권한 체계로 감시하는 설계는 아니며 생산 Query 호출 자리 전체 census를 수용 증거로 남긴다.

## Query 흐름

1. required binding·account·kind 검증, shared PollRemaining 검사. 남은 hint/poison/오류면 fn 호출과 freshness 기록 없이 반환한다. 기존 caller context를 연장하거나 새 sleep하지 않는다.
2. 논리 Query당 collector 하나를 만들고 해당 poll key와 함께 context를 파생한다. 부모 priority/deadline/cancel은 보존한다. token helper는 같은 collector를 유지하고 Token/zero-key로 전환한다. 각 실제 GET의 Begin 이후 재검사는 선행 transport가 수행한다.
3. fn 반환 직후 collector Snapshot을 읽는다. 원래 오류가 ClassAuthFatal이면 기존 gate latch와 durable escalation을 먼저 수행하고 원 오류·escalation 오류·관측 오류를 합쳐 반환한다. collector의 429나 local fault 때문에 이 부작용을 생략하지 않는다.
4. collector RateLimited, LocalError 또는 fn의 ClassRateLimited가 있으면 같은 Query 안에서 재시도하지 않는다. nil 반환으로 429가 숨겨져도 성공 처리하지 않는다. 적절한 기존 rate-limit sentinel 및 원 오류를 보존하고 Gate.RecordSuccess 0이다. 503만으로 shared429를 생성하지 않는다. token429도 Query를 종료하지만 poll hint는 생성하지 않는다.
5. 관측 문제가 없으면 기존 permanent/canceled/transient 분기와 deadline/backoff를 유지한다. transient 재시도의 모든 wire는 admission 및 관측을 다시 거친다. 진정한 ClassOK만 기존 RecordSuccess를 호출한다. 후속 호출의 hint 대기는 sleep이 아니라 새 wire 없는 skip이며, 만료는 store boottime/anchor에 따른다.

fn 반환 검사만으로는 부족하다. 실제 official.send의 token200 또는 GET401 뒤 Finish 저장 실패는 원 HTTP 결과를 유지하므로 다음 GET/refresh로 진행할 수 있다. 같은 logical Query collector에 이미 RateLimited/LocalError가 있으면 observed Transport가 다음 요청의 Acquire/Begin 전에 거절한다. Token helper를 거친 후속 wire에도 적용한다. body는 아직 내부 transport에 넘기기 전이므로 정확히 한 번 Close하고 오류·원 관측 원인·ErrNotSent를 보존한다. 새 pending이나 새 event를 만들지 않으며 최초 HTTP 응답/body/수용 의미는 변경하지 않는다. 다른 collector 또는 collector 없는 요청은 이 Query latch와 무관하다. 이미 진행 중인 동시 wire의 소급 중단은 주장하지 않는다. 원래 최종 fn 오류가 AuthFatal이면 기존 부작용을 보존하지만, 기존 send가 흡수한 중간401까지 새로운 AuthFatal로 승격하지 않는다.

현재 engine closure와 official.send 원문상 일반 GET429/token429가 정상 반환된 뒤 추가 GET를 수행하는 분기는 발견하지 못했다. 이 사실을 Finish 오류 반례와 혼동하지 않는다. 새로운 후속 wire가 시작되기 직전의 collector 검사는 위 local latch를 보존하며, 기존 Begin 이후 poll hint 재검사와 둘 다 필요하다.

## 선택 수용 시험

- 실제 official.Client + observed v2 factory + local TLS로 header 없는429, valid delta/date hint, invalid/overflow/>30, token429, 429와 body오류, collector 429 뒤 fn nil, record failure, 503-only 및 기존 transient retry를 확인한다. wire 수와 response 종류를 따로 센다.
- 429 직후 같은 Query retry0, 다음 cycle hint중 fn0/RecordSuccess0, 만료 뒤 read/성공, 다른 account/kind 및 non-poll 주문의 hint 비적용. 보호·일반 price 동일 slot, priority 유지. 기존 원장 POST분류 시험과 분리한다.
- 재시작/두 factory에서 같은 key 및 hint, 다른 instance/계좌/kind 분리, 모호한 field 접합 충돌 방지. reboot/clock 위조의 상세 leaf 시험을 중복 구현하지 않고 필요한 binding 경계를 선택 재실행한다.
- concurrent Query collector 격리, exitSideRetrier 복사 뒤 같은 binding, unknown kind·accountRef mismatch·selected-seq drift·foreign client·unobserved Budget는 fn0. AuthFatal+collector429 조합도 gate latch/escalation을 보존한다. 실제 official.Client의 token200→Finish 실패→GET0, GET401→Finish 실패→token/reGET0, token429→GET0, local latch 뒤 body exactly-once/원인 보존을 검증한다. store fault는 단순 dirty Begin 실패와 별개로 이미 commit된 상태에서 발생시켜 collector guard가 없을 때 후속 wire가 실제로 나가는 RED를 만든다.
- adoption·보호·일반 price의 같은 slot 및 Context.ReconcileDriver/NewTracer foreign reader 거부, 실제 CLI 기본 assembly를 검증한다. PositionPolicyCommandService의 재편입 사전 price도 같은 slot을 사용하고 기존 원장 권한·거절 결과가 유지되는 직접 fixture를 추가한다(독립 재검토 P2 반영). 동일 client guard 제거·collector prewire guard 제거를 각각 변이한다.
- 생산 constructor와 Query closure의 실제 값 흐름, authority guard clone·token helper·OFF golden. fake closure가 임의 client를 호출하는 capability를 production 보장으로 쓰지 않는다.

사전 AST/FLM/BTM 완료 후 RED/GREEN/변이, 정확한 snapshot의 normal/tagged race, scoped gstack 및 독립 검증. 네트워크는 fake 입력·명시 local TLS/exact DialContext·loopback-only/no-route namespace. 전체 tests/gate·브로커·운영·a087/a169 활성화는 미실행 범위다.

search-first: 기존 Retrier·EntryGate·clock·store/collector 및 stdlib sha256/encoding 재사용. 공식 계좌 identity authority helper는 추가 Accounts 재조회와 다른 authority 의미를 가지므로 무작정 호출하지 않는다. 새 의존성 없음. grill-me/grilling: 사용자 strict 정책과 기존 보호 hint 범위를 유지하며 미해결은 위 구현 경계/독립 반례다. impeccable: N/A — backend poll/control-only.
