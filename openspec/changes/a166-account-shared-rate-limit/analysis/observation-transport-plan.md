# 공유 429 관측의 전송·poll 통합 — 검토안

기준 main `bbe5066ec`와 store leaf API; store 최종 delta `04fdd134c`는 독립 검증 후 `5f681f2a5`로 게시됐다. 독립 설계 검토를 거쳐 마지막 절의 미배선 observed-open/Transport 첫 로트만 구현 승인했다. runtime v2·token·Retrier·전체 caller 통합은 후속이다. `429-store-draft.md`의 상태·복구 계약을 바꾸지 않는다. 기존 Transport.RoundTrip/connectionGate/responseOwner, runtime Factory, official token.exchange/doRequest, execgw.Retrier.Query 원문을 Manager가 손으로 읽었다. source 검색 결과 기존 ParseRetryAfter는 초 overflow와 HTTP-date 반올림 위험이 있으므로 신규 required 관측의 parser로 그대로 재사용하지 않는다. 기존 OFF parser·동작은 보존한다.

## 소유와 constructor

기존 Budget와 leaf ObservationStore를 하나의 required runtime factory가 소유한다. factory는 version 2 manifest의 영속 event.lock ID를 별도로 검증하고, budget Open와 store Open이 모두 성공한 뒤에만 외부에 객체를 내준다. version 1·누락·복사 lock·dirty 실패는 첫 전송 전에 거절한다. 기존 운영 manifest를 수정하거나 leaf provisioning을 자동 호출하지 않는다.

선택할 최소 구현은 새로운 observed-open helper가 Budget와 store를 함께 열고, Budget의 private observation 설정을 외부 노출 전에 한 번 고정하는 방식이다. 이후 Attach/Set 같은 변경 API는 제공하지 않는다. 생성된 Transport는 그 불변 설정을 읽어 같은 store를 사용한다. 기존 저수준 미관측 Open은 core/legacy fixture용이며 production required factory는 observed-open만 사용한다. official·token·a112 clone·budgeted replay·probe가 모두 factory의 동일 Budget를 쓰는 구조/통합 증거를 갱신한다. 별도 client 생성자 인자를 매 호출마다 전달하다 누락되는 대안보다 기존 Budget 전달 경로를 재사용한다. 이 결속은 거래 권한이나 계좌 동등성을 증명하지 않는다.

미관측 저수준 API의 존재를 production 포괄성 증거로 대체하지 않는다. 구조 검사는 모든 production Budget Open/Transport/NewWithAccountBudget/별도 HTTP constructor와 tools/CLI/console/engine 사용 자리를 열거해 required 가지가 observed-open의 같은 Budget에 도달하는 값 흐름을 확인한다. 테스트 패키지·명시적 OFF와 미배선 fixture 예외를 이름으로 숨기지 않고 분류한다. actual official token/401, a112 clone, replay/probe 및 서로 다른 process의 factory에서 event 기록을 읽는 통합 시험을 추가한다. 하나라도 unobserved required 경로가 남으면 전체 연결 완료·a087/a169 인계는 거절한다.

factory는 신뢰 UTC clock과 BootSource를 구성한다. BOOTTIME은 store가 검증하고 UTC는 HTTP-date 변환에만 쓴다. 고정 event.lock ID는 manifest version 2에 추가하며 중복/unknown/배열/trailing JSON 및 모든 기존 파일 경계를 유지한다. state/dirty inode를 manifest에 넣지 않는다. 실패한 초기화는 Budget을 닫고 원 오류와 정리 오류를 모두 보존한다. store가 별도 영속 fd를 소유하지 않는 현재 leaf 계약을 확인한 뒤 Close 순서를 확정한다.

## 전송의 한 attempt

1. 기존 URL/proxy/priority 검사를 거쳐 기존 Acquire로 실제 wire lease를 얻는다. observation 추가로 우선순위나 rolling cap을 우회하지 않는다.
2. 같은 활성 lease로 Begin을 내구 저장한다. 이전 orphan의 gap·poison 처리는 Begin의 상태 기계를 사용한다. Begin 실패면 HTTP Write 0이며 ErrNotSent와 오류를 반환한다. 성공 뒤 요청이 취소돼도 Finish 또는 남은 pending 공백이 사라지지 않는다.
3. engine-poll이면 **Begin 뒤** PollRemaining을 재검사한다. Begin이 이전 orphan을 poison으로 바꿀 수 있으므로 그 전에 한 번 읽은 결과로 전송하지 않는다. hint/poison/상태 오류면 Write 0, 명시적 local skip/fault를 수집하고 이번 pending은 정상 무응답으로 종료한다. poison은 지우지 않는다. 나머지 종류에는 poll hint를 적용하지 않는다.
4. 기존 격리 HTTP/1 TLS transport를 실행한다. 첫 Write 취소·revoke·drain·sent 판정과 body 소유 기간을 유지한다. 새 보안 우회 transport나 keepalive를 도입하지 않는다.
5. 응답 header를 받으면 body 소비 전에 status 429 여부를 request-local collector에 먼저 남기고 Finish한다. 429이면 header 부재·유효하지 않은 hint·body 오류·후속 성공과 무관하게 event를 기록한다. event+hint+pending 완료는 store의 원자 transaction이다. 다른 status는 기존 hint를 지우지 않는다.
6. 응답이 없으면 먼저 기존 gate stop/drain으로 전송 상태를 확정하고 정상 무응답 Finish를 시도한 뒤 lease를 해제한다. HTTP 응답이 있으면 Finish 뒤에도 lease는 기존 body EOF/Close/cancel까지 유지한다. Finish 실패가 원래 HTTP response/status/body를 교체하지 않는다. 응답 없는 오류는 원 오류를 보존한 errors.Join을 사용할 수 있으나 sent/AMBIGUOUS 의미를 바꾸지 않는다.

RoundTrip 이전 거절 경로의 자원 책임도 명시한다. Begin 실패·성공 뒤 hint/poison/상태 오류·요청 취소는 먼저 collector에 원래 localSkip/storeFault/cancellation을 독립적으로 보존한다. 이후 정리 성공 여부가 이 사실을 지우지 않는다. 아직 내부 tr.RoundTrip에 넘기지 않은 req.Body는 exactly-once Close하고 오류를 join한다. Begin 성공 sequence가 있을 때만 정상 무응답 Finish를 시도하고, Begin이 실패해 sequence를 주지 않았다면 추정값으로 Finish하지 않는다. 그 경우 남은 pending은 공백으로 보존한다. 마지막에 lease를 반드시 Release하며 정리 실패를 숨기지 않는다. 이 경로는 gate Write가 시작되지 않았으므로 ErrNotSent다. 이미 내부 transport에 Body 소유를 넘긴 뒤에는 별도의 중복 Close를 추가하지 않는다. Body Close·Finish·Release 각각 실패, Begin 후 취소, poison skip의 wire0·collector·잔여 상태·exactly-once를 시험한다.

Finish는 취소된 요청 context만 재사용하지 않는다. 요청 값은 유지하되 취소를 분리한 정리 context에 2초의 실제 timeout을 둔다. 이는 event lock 대기의 상한이며 fsync·커널 IO·callback의 실행 시간을 유한하게 보장한다는 주장은 아니다. 이 시간은 retry/poll 사업 판단용 clock이 아니라 정리 취소 상한이다. 실패하면 pending/dirty를 유지하고 callback 및 local fault를 남긴다. 자동 재시도·성공 주장·무한 Background loop는 없다. 기존 cleanup callback은 동시 안전하고 신속히 반환해야 한다. HTTP 응답을 받은 뒤 callback/저장 실패만으로 원래 주문 수용 결과를 NOT_SENT로 만들지 않는다.

## request-local scope와 Retry-After

공유 mutable last429 대신 논리 Query 하나에 collector 하나를 만든다. collector는 mutex 또는 atomic으로 header429/localSkip/storeFault를 누적하고 뒤의 성공으로 지우지 않는다. 다른 Query와 섞지 않는다. context priority/deadline/cancel은 그대로 계승한다.

engine-poll scope는 canonical account reference와 고정 RequiredQuery enum에서 만든 opaque key를 가진다. slot 생성은 group instance와 domain/길이 구분을 포함하며 key·원문 account를 로그에 남기지 않는다. 실제 엔진 client 계좌 선택과 AccountRef의 byte 결속을 검증해야 하며, account alias가 다른 참여자를 같은 slot으로 합친다고 주장하지 않는다. 조회 함수가 다른 계좌에 대해 실행될 수 있다면 scope 생성에서 거절하거나 분리해야 한다. 이 항목은 엔진 호출자 원문 대조 후 확정한다.

현재 소스 근거: `engine.resolveAccountRef`는 같은 official.Client의 Accounts 첫 행 DisplayName을 trim한 값과 양의 seq를 읽고 SelectedAccountSeq가 그 seq인지 비교한다. 이 값이 gatewayInputs.accountRef와 기존 Retrier.AccountRef로 전달된다. `official.accountsLocked`는 implicit 선택 seq의 drift를 거절한다. required poll binding은 이 검증이 끝난 실제 client/선택 seq/정확 accountRef를 불변 묶음으로 만들고 Query의 AccountRef와 일치하지 않으면 전송 전에 거절한다. 옵션의 임의 accountRef나 다른 client를 받아 ready라고 만들지 않는다. 생산 Query closure가 이 client 또는 그것을 보존한 adapter만 쓰는지 전체 생산 호출 자리(AST 및 손읽기)로 고정한다. 테스트용 교체 reads와 cross-key 실제 계좌 동일성은 별도이며 fixture를 운영 입증으로 쓰지 않는다. 처음 account resolution 자체는 아직 poll slot이 없으므로 non-poll 관측이며 부트스트랩을 추가 호출하지 않는다.

official token.exchange는 명시적으로 token wire-kind를 붙이고 poll key는 제거하지만 collector·priority·deadline은 유지한다. URL 문자열로 token 권한을 추정하지 않는다. 실제 engine-poll GET 및 그 Accounts 선행 GET만 poll-kind이다. POST·replay·flatten·probe는 non-poll이며 poll hint를 쓰지 않는다. poll scope가 비GET에 남는 실수는 전송 전 거절하는 구조와 fixture로 고정한다. default는 non-poll이다. 모든 enforcing kind의 429 event는 동일 store에 기록한다.

required hint parser는 delta-seconds의 부호·overflow·30초 초과를 검사하고, HTTP-date는 주입 UTC에서 정확 duration을 계산한다. 과거/invalid/missing/>30은 hint 없음이며 429 사건은 그대로 기록한다. 0초는 허용하고 duration을 아래로 반올림하지 않는다. 기존 미만료 값과의 max·재부팅 anchor·poison은 leaf가 담당한다. 신뢰 UTC 오차에 대한 보장은 하지 않는다. 기존 OFF RetryAfterTransport/ParseRetryAfter를 소급 변경하지 않는다.

## Retrier와 엔진

required factory의 관측 결속을 받은 Retrier에만 새 경로를 적용한다. OFF는 기존 fn 호출·재시도·freshness와 byte 계약을 유지한다. Query 진입 시 slot 잔여를 검사하고, fn에 collector/poll scope가 붙은 context를 넘긴다. budget 대기 후 재검사는 위 transport가 책임진다.

fn 반환 뒤 ClassAuthFatal의 기존 latch/escalation을 보존하면서, collector의 429·skip·fault가 있으면 같은 Query 안에서 더 보내거나 Gate.RecordSuccess를 호출하지 않는다. fn이 nil을 반환해도 이러한 관측이 있으면 명시적 관측 오류로 반환한다. token의 429도 그 논리 Query를 중단하지만 poll hint로 저장하지 않는다. header 없는 429와 503은 동일하지 않으며 503만으로 429 사건/공유 hint를 만들지 않는다. 429가 없는 기존 transient 재시도는 각 실제 wire의 admission을 유지한다. collector 없이 전역 마지막 사건을 소비하는 방식은 사용하지 않는다.

이 단계는 a087/a169를 켜지 않는다. 이후 두 change가 store snapshot의 pending/gap/overrun/instance 변경과 정확한 factory 결속을 소비해야 한다. engine.log 문자열을 통해 누락 없는 관측이라고 주장하지 않는다.

## 수용·미해결

독립 검토가 확인할 것: observed-open 불변 결속·모든 clone/POST의 공유 store, Begin 이후 poison 재검사, Finish context/응답 보존과 sent 분류, token kind와 collector 상속, engine 계좌 byte 결속, 기존 OFF, v2 cutover. 실제 filesystem IO에 유한 상한을 주장하지 않으며 record 실패는 보수적 가용성 저하를 만들 수 있다.

수용 fixture: missing/invalid/overflow/date hint, header429+body 실패·후속 성공·token429, Begin 실패 wire0, pending 직후 취소·normal no-response, header 직후 Finish 실패 원 POST/status/body 유지, 다음 owner orphan·poison skip, 대기 중 다른 request hint 생성, collector 동시 격리, auth fatal 보존, v1거부와 v2복사lock거부, OFF golden. 기존 wire/body drain과 401 개별 admission fixture도 정확 영향 범위에서 재실행한다. 모든 네트워크/변이는 local TLS와 loopback-only namespace, fake credentials·임시 manifest만 사용한다.

신규 의존성 없음. stdlib context/errors/http/sha256과 기존 store/clock/runtime를 확장한다. source/AST/FLM/BTM→RED/GREEN→변이·수동 gstack→독립 검증 순서. impeccable N/A — backend-only. 전체 tests/gate·실제 브로커·운영 cutover·활성화는 범위 밖이다.

## 독립 검토와 단계별 승인

독립 Terra 검토의 P1(pre-transport 정리와 collector 순서)은 자원 책임 문단으로 해소했다. P2(미관측 저수준 경로의 production 유입)는 전 호출자 값 흐름·실제 client/clone/두 process 통합을 수용 조건으로 추가했다. store 구현자의 API 대조는 Begin 이후 poison, 성공 seq만 Finish, 별도 settle ctx, store가 별도 영속 fd를 소유하지 않음을 확인했으며 독립 PASS를 대체하지 않는다.

store leaf의 독립 발견과 추가 delta가 닫힌 뒤, 첫 통합 로트는 **아직 production factory에 연결하지 않는 observed-open + 기존 Transport의 관측 분기**로 제한한다. required 기존 factory와 v1 manifest는 이 첫 로트에서 변경하지 않는다. 새 helper로 만들어진 Budget만 불변 store 결속을 가지며 기존 unobserved Budget은 같은 동작이다. 이 단계는 token/engine poll의 생산 지정이나 전체 429 기록 완료가 아니다. 별도 local TLS fixture가 typed context를 명시해 header·Begin/Finish·skip·body·sent와 오류 처리를 검증한다. 기존 모듈 내부의 private 구조·기존 함수 편집은 사전 FLM/BTM이 필수다.

첫 통합 로트 독립 검증 뒤에 runtime v2 factory 강제·공식 token 종류·엔진 Retrier 배선과 모든 fixture 갱신을 별도 로트로 수행한다. 중간에 low-level helper가 존재한다는 이유로 v2/전체 coverage가 완료됐다고 체크하지 않는다. 전체 API 이름은 기존 store leaf API를 재사용하며 추가 추상 서비스나 retry framework를 만들지 않는다.
