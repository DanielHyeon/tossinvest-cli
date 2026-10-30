# Function Logic Map: `Transport.RoundTrip`

Source `internal/accountbudget/transport_linux.go`; baseline `b0f7f1d93`, original AST `ast.json` committed in pre-edit `9e6cd08b8`. **절차 예외:** 첫 source 수정 전 완전 FLM/BTM은 작성되지 않았다. 이 상세 지도는 source/test 초안 작성 후 원본 AST·`git show` 및 현재 변경안을 대조해 작성했다. 사전 준수로 소급하지 않는다. risk-pattern-report는 원본 복사본 대상. CodeGraph 이 worktree index 없음; 원본 AST·손읽기 한정이며 graph PASS 주장 없음.

## 입력과 불변식

| 입력/상태 | 경계 및 결속 | 실패 처리 |
|---|---|---|
| req/context/body | nonnil HTTP request는 net/http 계약; HTTPS host만 직접 연결, body 소유권 | 거절시 body.Close 오류 join; 실제 첫 write 전 오류만 ErrNotSent |
| priority/collector | Normal/Critical/Aux; collector는 context에 있고 nil-safe | 기록된 rate/local error는 첫 전송 전 거절 |
| t.base.Proxy / TLS | proxy가 해당 요청에서 nil 및 오류 nil, direct HTTP/1 TLS, 인증 보존 | proxy URL/error, dial/handshake 오류는 전송 전 종료 |
| budget/lease | 같은 budget Acquire, 응답 body 종료까지 wire 보유 | 실패/취소시 release 및 오류 전파, 첫 write 확정은 gate.stop drain 뒤 |
| observation | required budget의 store; Poll은 GET과 valid key, Begin 뒤 Finish | pending/오류는 collector와 cleanupError; fence 없는 기존 경로 보존 |
| fence 추가안 | context의 명시 fence는 같은 budget pointer·불변 cursor·generation, one sticky | 최초 확인→Acquire→동일 fence mutex 재확인 및 lease-held event 검사→Begin. 무 fence는 변경 전 동작 |

## 원본 흐름 그룹·조기 반환 (행은 b0 source)

| ID | 조건/호출 및 변경 | 종료·오류/의미 | 시험 근거 또는 남은 선택 |
|---|---|---|---|
| G01 190–195 | sent=false, deferred err && !sent | ErrNotSent join; sent=true면 미부착 | `TestAttemptClassificationAfterDrain`; 추가 fence 거절 write0 |
| G02 197–207 | 이전 collector rate/local; rate 여부, body 유무 | local/rate join, body close, nil response | `TestObservedTransportRateLimitAndPoison`; 오류 body 별도 선택 필요 |
| G03 208–216 | 잘못된 priority/URL/scheme/host; body 유무 | ErrUnavailable + close, wire0 | `TestTransportPreAdmissionIsNotSent`; URL 개별 case 별도 선택 미주장 |
| G04 217–226 | Proxy nil 또는 선택; URL/error; body 유무 | nil proxy만 통과, 나머지 ErrUnavailable | `TestProxySelectionRefusedBeforeAdmission` |
| G05 227–236 | Acquire error 또는 nil lease; body 유무, e nil | body close, nil 오류는 ErrUnavailable, wire0 | `TestTransportPreAdmissionIsNotSent`; fence Aux busy 선택 |
| G06 237–246 | store nil/nonnull; poll 비GET/그 외 Begin | nonGET unavailable, Begin 결과 sequence/errors | `TestObservedTransportPollRejectsPost`, `TestObservedTransportBeginFaultClosesBody` |
| G07 247–254 | Begin 성공 Poll만 PollRemaining; error/양수/0; context 취소 | error 또는 양수는 deferred, context error join | `TestObservedTransportRateLimitAndPoison`, `TestObservedTransportCancelAfterBeginSettles` |
| G08 255–273 | obsErr; collector 기록, body 유무, sequence!=0 여부 | sequence 성공만 WithoutCancel 2초 NoResponse Finish; finish error callback+join; release 한번; 반환 | `TestObservedTransportBeginFaultClosesBody`, `TestObservedTransportCancelAfterBeginSettles`, `TestObservedTransportPreSendReleaseError` |
| G09 275–280 | gate 및 transport clone | Proxy nil·HTTP1·keepalive off, fresh connection; 네트워크 전 fence mutex 보유 금지 | `TestTransportOwnsFullBody`; 실제 TLS 선택 |
| G10 281–285 | host:port parse 실패/성공 | 실패 반환, 성공 host 인증에 사용 | TLS fixture; parse 실패 개별 시험 미주장 |
| G11 286–294 | TLS config nil/clone | clone 후 ServerName 및 ALPN HTTP1 지정 | `TestTransportOwnsFullBody` |
| G12 295–302 | DialContext nil fallback/제공; dial 실패/성공 | 제공 dial 보존, nil은 기본 dialer, error 반환 | `TestProxySelectionRefusedBeforeAdmission`; fallback 개별 시험 미주장 |
| G13 303–317 | handshake timeout 유무, handshake 실패, gate 등록 실패/성공 | handshake 실패 raw.Close; 취소 등록 거절 ErrClosedPipe; 성공 gatedConn | `TestCancelBlockedTLSWriteDrains`, `TestAttemptClassificationAfterDrain` |
| G14 319–340 | underlying nil response; beforeStop 유무, gate.stop → sent 읽기; store 유무/Finish 오류 | 2초 NoResponse settle; cleanupError, stop/finish/release 오류 join; first write 전만 ErrNotSent | `TestObservedTransportNoResponseAfterSent`, `TestCancelBlockedTLSWriteDrains` |
| G15 341–359 | response 존재, store 유무; status429 여부, poll hint 여부, Finish 오류 | 429 collector/rate/hint; Finish 실패 callback이나 기존 response 반환·sent 의미 유지 | `TestObservedTransportRateLimitAndPoison`, `TestObservedTransportFinishFaultPreservesResponse` |
| G16 360–365 | response.TLS nil 및 gate TLS 존재 | 검증된 TLS state만 보완 | `TestTransportOwnsFullBody` (정확 TLS state 추가 확인 필요) |
| G17 366–377 | responseOwner Close/EOF 또는 context 취소 goroutine | body close→gate stop→release 한번, async 오류 callback; response 원형 반환 | `TestTransportOwnsFullBody`, `TestCancelUnreadResponseReleasesOwnership` |

## 새 fence 분기·잠금 순서 (초안 diff; 검증 전)

| ID | 조건/효과 | 실패·상태 | 명명 시험 |
|---|---|---|---|
| F01 | context fence 부재/존재; invalid/wrong owner/sticky | 부재 기존 G01–G17; 명시 오류 body close/collector, ErrNotSent | `TestA169FenceInvalidOwnerAndPending`, 추가 wrong-owner HTTP fixture 필요 |
| F02 | fenced initial hook→Acquire→acquired hook→check | lease-held owner/started/identity/counters/generation/pending 확인; 실패 mutex 반환 후 release helper | `TestA169FenceTransportAfterEventAndOwn429`, `TestA169FenceReleaseFailureConcurrentWaiter` |
| F03 | Begin 오류/ PollRemaining 오류 대 정상 context cancellation·deferred | Begin·저장 오류 sticky; 단순 취소/aux busy는 sticky 아님; sequence=0이면 Finish 금지 | `TestA169FenceAuxBusyAndCancellationNotSticky`; Begin fault fenced fixture 필요 |
| F04 | nil response stop/Finish 오류/Release | stop+finish를 release mutex 아래 최초 latch, 해제 오류도 unlock 전 publish; sent 판정 원래와 동일 | `TestA169FenceReleaseFailureConcurrentWaiter`; nil response fault 추가 필요 |
| F05 | 429/Finish 오류, response body close/gate stop/release | 최초 sticky; response 보존, body wrapper는 원래 close 오류를 release 전에 latch/collector | `TestA169FencePostcommitFinishConcurrentWaiter`, `TestA169FenceBodyCloseBeforeRelease`, `TestA169FenceTransportAfterEventAndOwn429` |
| F06 | 서로 다른 Budget·프로세스의 event/gap/후퇴 | event lock+lease 결속, 증분/rollback/pending fail closed; baseline 당시 이전 gap/overrun 허용 | `TestA169FenceBaselineAndEvents`, `TestA169FenceCrossProcessToken429`; gap·overrun fixture 추가 필요 |
| F07 | official 401/token context 보존 및 후속 429 | wire 별 검사, 새 collector도 fence sticky 유지 | `TestA169FenceOfficial401TokenAndSticky`; adopt 별도 fixture 필요 |

## 호출·변이·수명

`Acquire` 전 fence mutex 미보유. 정상 잠금 순서 wire → fence.mu → lease.mu → event lock; `withLease` callback 안에서 fence.mu 재획득하지 않음. `Snapshot`은 fence 생성에서만 호출, 현재 lease 아래 재호출 금지. `Begin`은 old pending을 정착시킬 수 있으므로 fence check 이후여야 함. `Finish`은 성공적인 Begin에만 적용; 정상 cancel은 2초 bounded 별도 settle context. `gate.stop`과 body.Close는 fence mutex 밖, 실제 wire Release 및 오류 latch는 같은 mutex 안. response close는 sync.Once; B가 사전 검사 후 대기했을 때 unlock 후 오류가 재검사보다 먼저 게시되어야 함. `WithTokenObservation`은 부모 context 값 유지; official 호출 fixture에서 별도 검증 필요. 유한 OS 정지/외부 비협조 writer 보장 없음. 소스에서 내부 재시도 없음, official의 401 재시도는 매 RoundTrip으로 재진입한다.

현재 초안 시험은 해당 이름의 관찰만 주장하며, 전체 branch coverage·실계좌·운영·전체 gate는 미검증. Manager가 이 사후 지도와 절차 예외를 검토할 때까지 source 편집·추가 시험은 중지한다.
