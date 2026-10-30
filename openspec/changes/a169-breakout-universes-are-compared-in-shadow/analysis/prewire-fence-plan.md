# 안정 기준점과 요청별 관측 fence

2026-10-31 Manager 제한 설계 READY. 기준 `b0f7f1d93`. 사용자 결정: **현재 안정된 시점부터 시작하고 과거 gap은 제외**. 과거 기록을 삭제하거나 복구된 것으로 바꾸지 않는다. 측정 시작 이전 429를 새 사건으로 세지 않는 기존 계약도 유지한다. 이 로트는 내부 budget/transport 기능과 fixture만이며 실제 a169 실행 연결·활성 엔진 신원·당일 abort 영속성·운영 전환은 후속이다. 전체 a169 freeze/측정/archive 승인이 아니다.

## 1. 기준점 생성

새 `(*Budget).NewObservationFence(ctx)`는 required observed Budget의 기존 Snapshot을 읽어 불변 fence를 만든다. nil ctx/budget, closed/unobserved, dirty/손상·identity·읽기·잠금 오류, generation 경합, pending Busy/Orphan 모두 생성 거부다. 내부 대기 재시도나 자동 root/instance 전환은 하지 않는다. caller는 별도 시작 deadline 아래 다시 시도할 수 있으나 이 로트에 시작 loop는 만들지 않는다.

성공 조건은 검증된 stable view에서 Pending=nil/None인 한 번의 snapshot이다. 그 snapshot 읽기를 기준점의 선형화 시점으로 삼고 scope/instance와 현재 Event/Gap cursor를 private immutable 필드로 저장한다. 직후 다른 writer가 시작할 수 있으며 그것을 전송 전 검사가 다룬다. 과거 Event/Gap 목록 및 zero cursor로부터의 ring overrun은 **기준점 생성 때만** 과거 이력으로 제외한다. 현재 state의 유효성 오류나 pending을 과거 gap으로 간주하지 않는다. poll poison·hint·dirty를 지우지 않는다. 생성 뒤 cursor를 갱신/재기준화하는 API는 없다.

새 `(*ObservationFence).Context(ctx) (context.Context,error)`는 유효한 fence와 nonnil ctx를 private context 값으로 결속한다. 외부에서 zero-value fence를 만들거나 다른 Budget에 붙여도 실제 전송 경계에서 거절한다. absence는 기존 동작이며 명시적 잘못된 fence를 absence로 낮추지 않는다. scope/instance가 같아도 다른 Budget 포인터이면 거절한다. 생성 함수 context는 읽기 수명만, 요청 context는 개별 전송 수명이다. fence는 Budget을 영구 pin하지 않고 실제 사용 시 기존 pin/lease 소유권을 검사한다.

## 2. 각 실제 wire 전 검사

기존 `Transport.RoundTrip`에서 fence가 있을 때만 sticky failure를 확인하고, 정상 Acquire 뒤 **자기 observation Begin 전**에 fence 검사를 한다. 동일 Budget의 유효 lease이며 observationStarted가 아직 거짓이어야 한다. 기존 withLease·event lock·state read를 재사용하는 좁은 내부 메서드로 읽고, wire→event 잠금 순서를 유지한다. Snapshot을 lease 아래 재호출하거나 범용 callback/외부 mutable store를 만들지 않는다.

검사 실패: identity/cursor 역행/읽기 오류, Pending 존재, 기준점 뒤 Event 또는 Gap 증가. Event/Gap cursor는 누적 값이므로 ring에서 새 기록이 덮여도 증분을 놓치지 않는다. 새 overrun도 통과시키지 않는다. 현재 lease를 가진 동안 다른 정상 writer가 wire를 가질 수 없으므로 기존 pending은 미확정/orphan으로 거절하고, 그것을 지우거나 Begin으로 대신 정착시키지 않는다.

검사 성공 후 event lock을 놓되 wire lease는 계속 유지한다. 기존 Begin→context 확인→TLS 첫 plaintext Write→response/body/Close 수명을 그대로 쓴다. 자기 Begin의 pending을 다시 검사해 스스로 거절하지 않는다. 검사와 첫 Write 사이 다른 정상 writer의 새 Finish가 끼어들 수 없다는 근거는 lease다. 비협조 프로그램·broker 도착 시각까지 보장하지 않는다.

## 3. 실패 수명과 token/401

fence는 mutex로 보호한 첫 실패 하나를 sticky 보존한다. 새 shared 사건/상태 오류 및 자기 observation Begin/Finish/cleanup 오류·429가 발견되면 그 fence의 후속 요청을 거절한다. 새 논리 호출마다 collector를 바꿔도 같은 fence가 유지되므로 오류를 잊지 않는다. 일반 Auxiliary admission busy/refusal 자체를 shared 사건으로 만들지 않는다. nil fence의 기존 동작은 그대로다.

거절한 미전송 요청 body는 닫고 lease를 정확히 해제한다. fence 검사에서 Begin 전 실패했으면 가짜 sequence나 NoResponse Finish를 쓰지 않는다. Begin에 성공한 뒤 실패한 경로만 기존 bounded settlement를 재사용한다. 이미 얻은 HTTP 응답과 기존 AMBIGUOUS 의미는 Finish 오류 때문에 NOT_SENT로 바꾸지 않는다. wire0인 실제 요청만 ErrNotSent다. 기존 collector/cleanup error에도 결과를 남긴다. 첫 실패 이후 추가 오류 문자열이 무한히 누적되지 않게 한다.

`WithTokenObservation`은 parent context 값을 유지하므로 fence도 token/401/adopt/mint 경로에 전달한다. source 손읽기만으로 완료하지 않고 실제 official.Client fixture로 증명한다. shared fence를 적용하지 않은 엔진/flatten/OFF 동작을 변경하지 않는다. 생산 a169 caller 연결은 이번 로트에 없다. 후속 a169 runtime은 세션당 fence 하나를 유지하고 abort를 새 fence로 우회할 수 없게 영속화해야 한다.

취소는 기존 context/gatedConn 첫 Write 선형화를 따른다. 이미 첫 Write 허가 경계를 통과한 요청을 소급0으로 주장하지 않는다. 감시기·engine identity·같은 날 재시작 금지는 별도 계약이며 이 API만으로 해결됐다고 표시하지 않는다. 외부 monitor/Stop callback·timer는 추가하지 않는다.

## 4. 검증과 영향

변경 예상: accountbudget 새 fence 파일/Linux 및 unsupported 경계, 기존 Transport.RoundTrip의 선택 분기와 기존 cleanup closure, wire census의 해당 reviewed shape. token 구현은 수정 없이 문맥 전달 시험으로 확인하는 것이 우선이다. 기존 함수 편집 전 FLM/BTM·AST·Pre-Edit를 고정하고 변경 범위가 늘면 Manager에게 보고한다. schema·운영 manifest·env default는 변경하지 않는다.

RED/fixture: 과거 gap·429·historical ring overrun이 있어도 stable baseline 생성 후 정상 wire; pending Busy/Orphan·dirty·closed·wrong owner 거절; baseline 뒤 상대 프로세스429/새gap→다음 token/GET0; same fence+새collector에도 sticky 거절; own Begin 정상, Begin/Finish fault, response 의미 보존, 취소/lease/body Close; 실제 official.Client cached/adopt/mint401 전달; absent fence 기존 선택 회귀. gap/event counter 역행·overrun 경계도 검사한다. 실제 전송은 fake credential·temp manifest·exact local TLS dial과 lo-only/no-route만 허용한다. 동등변이는 검출로 세지 않는다.

독립 targeted normal/tagged race·필요 다중 process·유효 변이·수동 scoped gstack와 Manager 검증. 전체 tests/gate·실브로커·실계좌·운영·활성화 금지. CodeGraph 최신 index가 없으면 원문 AST/FLM과 도구 제한을 명시하고 PASS라고 하지 않는다.

search-first: 기존 ObservationStore.Snapshot/withLease/read, Budget Acquire, Transport response ownership, context 및 ObservationCollector 재사용. `analysis/prewire-fence-research.md` 원문과 현재 source를 대조했다. repository/Go stdlib에서 충분해 외부 registry/web 패키지 조사는 생략, 새 의존성·라이선스 변경 없음. ai-memory CLI 해당 scope 검색은 결과 없음. grill-me/grilling에서 과거 gap 정책은 사람 답으로 닫고 안정성·실패·수명 반례를 위에 고정했다. impeccable N/A — backend 전송 경계와 시험만.

독립 설계 리뷰 뒤 Manager가 제한 READY를 기록하기 전 구현하지 않는다. source/test 로트 완료와 사람 운영 단계는 구별하며 active change를 허위 archive하지 않는다.

## 5. 독립 리뷰 P1 정정 — cleanup과 다음 admission 순서

독립 리뷰 `/tmp/a169-prewire-plan-review.md`의 A cleanup 실패와 대기 중 B 간 경합을 수용한다. Acquire 전 sticky 확인은 빠른 거절일 뿐이다. B는 **Acquire 뒤 fence mutex를 잡아 sticky를 다시 확인한 뒤에만** lease/state 검사를 통과할 수 있다.

같은 fence의 lease 해제는 좁은 helper로 fence mutex 아래 수행한다. gate.stop/Begin·Finish 등 이미 알려진 오류를 먼저 sticky에 보존하고, mutex를 유지한 채 Lease.Release를 호출한 뒤 그 반환 오류도 sticky에 보존하고 mutex를 놓는다. 따라서 Release가 wire lock을 실제로 놓은 뒤 오류를 반환해도 B는 fence mutex에서 기다렸다가 그 오류를 본다. lock 순서는 wire 소유→fence mutex→lease mutex→필요 event lock이며, fence mutex를 잡고 Acquire를 호출하지 않는다. store Begin/Finish가 lease mutex를 보유한 채 fence를 잠그는 역순도 만들지 않는다. fence mutex 아래 네트워크 전송이나 전체 body drain은 수행하지 않는다. 잠금/OS 정리의 유한 wall-time 상한을 주장하지 않는다.

응답 body Close 오류는 좁은 private ReadCloser wrapper가 원래 Close 결과를 fence에 기록한 뒤 기존 responseOwner의 finish(lease 해제)가 실행되도록 한다. 기존 responseOwner.Close 본문은 수정하지 않는 안을 우선한다. gate.stop과 body.Close는 fence mutex 밖에서 끝내고 알려진 오류를 해제 helper에 전달한다. 상태/cleanup 오류와 429는 sticky지만, 단순 Auxiliary admission busy 및 Begin 성공 뒤 context 취소+정상 NoResponse Finish는 스스로 관측 무결성 실패가 아니므로 context 오류만으로 latch하지 않는다. Begin 실패는 durable pending 여부가 불명일 수 있어 보수적으로 latch한다. 이미 얻은 응답/기존 ErrNotSent 판정은 보존한다.

fence 검사 통과가 이 요청의 permission 선형화다. 그 뒤 생긴 별도 취소/오류는 이미 허가 경계를 통과한 요청을 소급 없애지 않는다. 특히 미래 monitor 취소 계약을 이 로트가 대신 구현했다고 쓰지 않는다.

추가 RED: B가 sticky=nil을 읽고 Acquire 대기 중 A의 postcommit Finish 오류; body.Close 오류; 실제 wire unlock을 먼저 수행한 뒤 반환하는 Release 오류. B는 lease 후 sticky 재검사에서 첫 Write0이어야 한다. 반대조로 정상 cleanup→B 정상 전송과 Auxiliary busy 재시도 허용, 정상 cancellation settlement 이후 다른 유효 context의 fence 사용을 확인한다. lock 순서·double release·goroutine 누수·닫힌 body 및 원 HTTP 결과도 검증한다.

## 6. 제한 READY

독립 재검토에서 §5의 경합 P1 해소, 새 P0/P1 없음. Manager가 이 내부 선택적 fence 로트만 구현 승인한다. 검사 실패 시 fence mutex를 놓고 cleanup helper가 재획득하는 순서를 명시하여 self-deadlock을 막는다. fixture 재사용 조사는 `prewire-fixture-research.md`이며 sleep을 동시성 증거로 사용하지 않는다. 기존 hook으로 결정적 경계가 불가능하면 시험 전용 seam과 영향 목록을 먼저 기록·보고한다. 전체 a169 freeze나 운영 활성화 승인은 아니다.
