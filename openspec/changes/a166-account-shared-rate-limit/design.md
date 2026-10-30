# a166 4판 구현 계약 — 계좌 상한 우선

상태: **READY — 개발 착수, 2026-10-31 Manager 판정**. 사용자 개발 진행 지시, 연구 원문과 독립 검토, scoped strict·PM 증거를 반영해 proposal·design·spec delta·tasks를 정렬했다. 첫 로트는 기존 경로에 연결하지 않는 내부 모듈이다. 기존 함수 편집은 해당 로트의 최신 증거·Pre-Edit 후 진행한다. 제품 완료·운영·활성화 PASS는 아니다. 2판의 허용 초과·local fallback·보호 무대기 계약은 복원하지 않는다. 판정 근거는 `analysis/manager-review-v4.md`다.

## 결정과 범위

- 계좌 전체 3건/초를 우선한다. 보호 요청도 대기하며 우선순위는 대기열의 순서다. 이미 시작한 요청을 즉시 끝낼 수 있다고 가정하지 않는다.
- 이 제한은 동일한 공유 중재에 연결된 TossOS 참여자 전체에 적용한다. 앱의 수동 주문, 다른 소프트웨어, 다른 호스트 등 참여하지 않는 송신자의 총량은 통제하지 못한다. 배포 시 참여자 누락이 있으면 계좌 전체 보장을 선언하지 않는다.
- API 요청 시작의 제어·시험 경계는 직접 HTTPS 연결에서 HTTP 요청 plaintext를 TLS 연결의 첫 `Write`에 넘기는 지점이다. TLS handshake, TCP 재전송은 별도 API 요청이 아니다. 네트워크 지연 때문에 서버 도착이 몰릴 수 있으므로 브로커 도착률이나 429가 발생하지 않는다는 보장은 하지 않는다.
- 어떤 rolling 1초 구간에서도 이 시작 이벤트는 합계 3개 이하다. 보호·401 재전송·token 교환·redirect에도 예외가 없다. 이를 강제할 수 없는 전송 구성은 enforcing 모드에서 거부한다.
- 수치로 된 보호 wall-time 상한을 주장하지 않는다. 기존 요청 context/HTTP timeout은 대기와 I/O에 적용하고 상한을 위해 시한을 연장하지 않는다. OS 선점·진행 중 쓰기·네트워크의 실제 종료가 지연될 수 있다.
- 배포·실계좌·운영 토글은 이번 개발과 분리한다. a087 admission nil, a169 리허설·측정 보류를 유지한다.

## 구현 방식

신규 외부 의존성 없이 기존 Linux flock 선례, Go 표준 `net/http`, `crypto/tls`, context를 확장한다. 프로세스별 token bucket과 예약시각 기반 허가는 사용하지 않는다.

1. 지원하는 동일 호스트의 참여자는 명시적으로 구성된 같은 계좌 예산 디렉터리와 안정적인 잠금 inode를 사용한다. config-dir 또는 앱 키별로 예산을 새로 만들지 않는다. 호스트·컨테이너 동등성은 배포 전 로컬 두 프로세스 시험으로 확인한다.
2. 하나의 소유자가 실제 요청 한 개의 수명을 독점한다. 소유권 획득 후 매번 단조 시간으로 최소 350ms를 기다린다. 이는 3건/초 이하를 위한 보수적 내부 간격이며 성능 목표나 지연 상한이 아니다. 재기동한 참여자도 예외 없이 기다려 벽시계·전 소유자의 timestamp 복구에 rate 안전성을 의존하지 않는다.
3. enforcing 전송은 fresh HTTP/1 연결로 한 요청만 보낸다. proxy·HTTP/2·연결 재사용·임의 transport 옵션 등 증명 밖 구성은 OFF로 돌아가는 대신 시작을 거부한다. redirect hop은 각각 독립 허가를 받아야 한다. 자동 transport 재시도가 발생하지 않는 구성은 현재 Go 버전 소스와 반례 시험으로 고정한다.
4. 요청의 모든 Write와 revoke는 같은 동기화 경계를 사용한다. revoke가 끝난 뒤 새 Write나 지연 dial 결과의 등록은 거부한다. 진행 중 Write가 있으면 실제 종료를 기다린다. 연결 수명이 끝나기 전에 다음 소유자에게 예산을 넘기지 않는다.
5. **응답 본문을 보존한다.** `RoundTrip`의 header 반환 시점에 연결을 닫지 않는다. 정상 body EOF/Close, 오류, 취소를 각각 다루고 실제 요청 writer의 revoke/drain이 끝난 뒤 소유권을 반환한다. 호출자가 body를 읽지 않는 경우에도 기존 전체 요청 deadline/cancel로 정리가 진행되어야 한다. 무한 종료 보장을 만들기 위해 새 요청을 우회시키지 않는다.
6. 기본 OFF transport와 authority 판정은 그대로다. enforcing transport도 공식 constructor가 만든 구성만 금전 권한을 가질 수 있다. 사용자 `WithHTTPClient`·endpoint override는 wrapping 후에도 authority가 없다. TLS 서버명·인증서 검증·ALPN·응답 `TLS` 상태·timeout 보존을 시험한다.

위 방식은 동시 연결을 줄여 처리량을 낮출 수 있다. 이 비용을 숨기거나 3건/초를 항상 달성한다고 주장하지 않는다. 연구와 독립 검토를 근거로 구현 방향을 채택하며 제품 수용 시험을 대체하지 않는다.

## 보호 사슬과 대기 결과

보호 판단 가격 조회, 보호 매도·그 취소, flatten의 선행 조회와 주문, 보호 주문의 접수 확인 및 이 사슬에서 필요한 token 교환은 critical로 분류한다. 신규 진입은 normal, a087 캐시 갱신과 a169 측정은 auxiliary다. 경로 이름이나 HTTP method만으로 매도/진입을 추측하지 않고 호출자의 명시 context를 따라간다. token 교환은 부모 등급을 계승한다.

- critical 대기자가 등록되면 새 normal/auxiliary가 앞지르지 않는다. 이미 전송 중인 소유자는 강제 추월하지 않는다. 동일 등급은 등록 순서로 처리하고, 소유자가 아닌 대기자의 사망·취소 정리는 PID 숫자만으로 판단하지 않는다.
- normal의 기아 가능성은 허용하되 관측 가능해야 한다. auxiliary는 대기열을 쌓지 않는 try-admit 방식으로, 거절되면 해당 갱신·표본을 생략한다. a087 만료 토큰은 전송 전에 버린다.
- 허가 대기 중 취소/시한 만료는 새 요청 0이다. 첫 HTTP Write 전 거절임을 확인한 경우에만 `NOT_SENT` 근거로 사용한다. 전송 뒤 취소·오류를 미전송으로 바꾸지 않는다. 토큰 교환과 주문 본체를 구분해 전송 여부를 판정한다.
- 전송 여부는 **현재 주문 attempt의 본체**에 결속한다. token 요청을 보냈다는 사실은 주문 본체를 보냈다는 뜻이 아니다. 반대로 앞선 본체 전송·취소 409 등이 이미 있으면 후속 admission 실패로 최초 결과를 지우지 않는다. a165의 첫 AMBIGUOUS 보존 규칙을 유지하는 회귀 시험이 필수다.
- 보호 ACK 뒤 확인 GET이 실패하면 기존 IN_DOUBT 절차를 보존한다. 예산 때문에 늦어졌다는 이유로 주문을 중복 전송하거나 원장을 확정하지 않는다.
- a087의 활성 경로에서는 고루틴의 동시 생존과 실제 요청 전송의 중첩을 구분한다. 공통 a166 gate에 보호 관측과 캐시 갱신을 모두 연결하고, 기존 「관측 지연 0」을 enforcing 모드에서도 성립한다고 주장하지 않는다. 이 변경은 a087 5.5 계약에도 반영해야 한다. 휴면 OFF 동작은 동일하다.

## 우선 중재 — 두 잠금의 역할 분리

요청 수명 동안 잡는 `wire.lock`과 짧은 대기열 변경에만 쓰는 `queue.lock`을 분리한다. 요청 소유자는 네트워크 I/O 중 `queue.lock`을 잡지 않는다. 따라서 느린 요청 동안에도 critical 대기자는 등록할 수 있다. 모든 경로의 잠금 순서는 queue → wire(비차단 try-lock)이며 wire를 잡은 채 queue를 기다리는 역순은 금지한다.

대기열은 queue 잠금 아래 등록한 단조 ticket 순서와 등급을 기록한다. 각 대기자는 무작위 ID의 개별 liveness 파일에 독점 flock을 자신의 전체 대기 수명 동안 잡는다. 다른 프로세스는 queue 잠금 아래 해당 파일의 비차단 flock으로 종료 여부를 확인한다. PID 단독 비교·wall-clock TTL로 살아 있는 대기자를 지우지 않는다. liveness 파일이 누락·손상되어 의미를 판별할 수 없으면 큐를 자동 비우지 않고 실패한다.

등록 순서는 안전한 liveness 파일 생성·충돌/링크 거부 → 독점 flock 획득 → queue 잠금 아래 ticket 공개다. 잠금을 잡기 전 ticket을 공개하지 않는다. 등록 실패는 공개되지 않은 파일만 정리하고, 선택·취소 뒤에는 queue 레코드 정리가 성공한 뒤 liveness 파일을 정리한다. queue 잠금 파일은 안정적인 inode이며 원자 교체 가능한 queue 데이터 파일과 분리한다.

선택은 queue 잠금 아래 죽은 대기자 정리 → critical 최저 ticket(없으면 normal 최저 ticket) → 선택된 본인만 wire 비차단 획득 → 큐에서 본인 제거·기록 완료 → queue 해제 순서다. wire가 바쁘면 queue를 풀고 context-aware 대기 후 재확인한다. auxiliary는 대기열에 등록하지 않고 같은 원자 판정에서 대기자 0 및 wire 비차단 획득일 때만 진행한다. 획득 이후의 350ms 간격도 context-aware이며 취소되면 전송 없이 wire를 반환한다.

정상 반환은 요청 revoke/drain을 끝내고 wire를 풀며 queue를 다시 잠그지 않는다. 프로세스 사망은 커널이 wire 및 liveness flock을 해제한다. 다음 후보도 최소 간격을 새로 기다린다. 큐 state는 queue 잠금 아래 검증 가능한 완전한 레코드로 갱신하고, fsync·읽기·쓰기·원자 교체 실패는 허가 전에 실패한다. 파일 형식·내구성과 전체 프로토콜은 다중 프로세스 RED/변이로 검증해야 하며 현재 구현된 기능이 아니다.

우선순위가 확정되는 시점은 queue 잠금 아래 wire 획득과 선택 레코드 갱신을 완료한 때다. 그 뒤 등록한 critical은 이미 선택된 normal을 선점하지 않는다. 선택 레코드 저장·동기화에 실패하면 요청을 보내지 않고 wire를 반환하며, 불확실한 queue 상태는 전 등급 차단으로 남긴다. 파일 descriptor는 close-on-exec이고 자식·dup에 소유권을 넘기지 않는다. 연구의 단일 프로세스 mutex 또는 PID kill fixture만으로 이 프로토콜을 검증했다고 보고하지 않는다.

### 상태 갱신 실패의 전파

rename 뒤 directory fsync가 실패하면 새 state가 다른 프로세스에게 이미 보일 수 있다. 실패한 선택자만 반환시키는 것으로는 불확실한 상태의 후속 허가를 막지 못한다. queue 잠금 아래 상태 갱신 전에 `dirty` 파일을 안전 생성하고 파일·디렉터리 sync 성공을 확인한다. 그 뒤 새 state의 write·file sync·rename·directory sync를 모두 성공시킨 후에만 dirty를 제거한다. 각 참여자는 state를 사용하기 전 dirty가 없는지 확인한다.

- dirty 내구성 확인 실패: state 갱신 0, 전송 0. 남은 dirty를 자동 제거하지 않는다.
- state 갱신/rename/directory sync 실패: dirty를 남기고 wire를 반환하며 전송하지 않는다. 후속 참여자는 모든 등급에서 거부한다.
- state가 완전히 내구화된 뒤 dirty 삭제의 마지막 directory sync만 실패: state의 내구성과 dirty 삭제의 내구성을 구별한다. runtime에서 dirty가 없고 state가 유효하면 이를 불확실한 state로 취급하지 않는다. 재기동 시 dirty가 재등장하면 보수적으로 거부한다.
- crash로 남은 dirty는 자동 초기화하지 않는다. 운영자가 모든 참여자를 정지하고 상태·namespace를 확인하는 별도 복구 절차를 명시한다. 실패한 파일시스템에 어떤 표지도 영구 기록할 수 있다고 주장하지 않는다. dirty가 내구화되기 전에 state를 변경하지 않는 순서가 핵심이다.

이는 제품 구현과 검증의 파일 상태 계약이다. 실제 kernel fsync EIO나 전원 장애를 주입한 증거 없이 fake fault 시험을 실측으로 보고하지 않는다.

## 장애와 429

공유 파일·잠금·상태 무결성·instance 동등성을 확인하지 못하면 전 등급의 새 전송을 막는다. 부분 쓰기·손상·잘못된 mount·파일 교체를 새 예산으로 자동 초기화하지 않는다. 의도된 OFF와 enforcing 필수인데 설정이 누락된 경우를 구분한다.

공유 namespace는 운영 중 외부 교체/삭제를 허용하지 않는 지원 조건을 명시해야 한다. 비협조 프로세스나 관리자에 의한 임의 filesystem 변경까지 flock으로 막았다고 주장하지 않는다. 실제 접근 권한과 container UID를 확인하기 전 지원환경 확정은 보류한다.

모든 429는 응답 body 읽기 성공 여부·Retry-After 존재·후속 재시도 성공과 독립적으로 공유 이벤트에 기록한다. 비밀키·토큰·계좌/주문 원문은 이벤트에 넣지 않는다. 이벤트의 계수·로그 관측과 전송 허가 패널티는 별개의 기능이다.

**Manager 최소 변경안:** a166에서 새 계좌 전체 5분/30초 cooldown을 도입하지 않는다. 계좌 전체 admission은 항상 350ms 최소 간격을 적용하고, 429에 대한 호출자 동작은 이미 확정된 `verify-execution-capability/analysis/retry-matrix-slo-2026-10-22.md` §2를 따른다. 엔진 폴링은 429가 나면 그 조회를 끝내며 같은 조회 안에서 재전송하지 않는다. 유효한 Retry-After ≤30초는 해당 폴링 소비자가 다음 호출의 earliest-start로 존중하는 방안을 명세·시험한다. 누락·invalid·30초 초과는 그 조회 포기이며 30초로 잘라 재전송하지 않는다. 별도 비지연 walk의 기존 15s/30s backoff는 유지한다. POST 429의 원장 판정을 바꾸지 않는다.

earliest-start는 `Retrier.Query`의 계좌 + 엔진 폴링 역할 + `RequiredQuery` 종류에만 적용한다. price 종류를 공유하는 보호·일반 가격 조회가 함께 보류될 수 있지만, 주문 본체·그 token 교환·flatten·별도 walk를 새로운 계좌 전역 패널티로 막지는 않는다. 429를 반환한 호출은 즉시 종료하고 후속 cycle이 유효한 제한 안에 있으면 새 wire 없이 그 조회를 생략한다. 이 생략은 freshness 성공으로 기록하지 않는다. 이미 시작한 동시 요청을 되돌렸다고 주장하지 않는다.

hint는 해당 429 응답에 직접 결속해 파싱한다. 503이나 다른 응답의 전역 마지막 값을 소비하는 기존 도구를 재사용하지 않는다. 유효한 ≤30s만 기록하며 기존 미만료 유효 값과 최대값을 취한다. 누락·invalid·초과 값은 새 deadline을 만들거나 기존 유효 값을 지우지 않는다. HTTP-date는 응답 시각 기준이며 반올림으로 조기 만료시키지 않는다. overflow와 음수는 invalid다.

재시작은 새 30초 일괄 기동 정지가 아니라 해당 유효 hint의 영속 레코드 복원으로 처리한다. 같은 boot/time namespace에서는 주입된 공유 boottime 표본과 duration으로 남은 시간을 구한다. host reboot로 boot ID가 바뀌면 이전 ≤30s hint가 새 boot 이후 uptime만큼은 이미 경과했다는 보수 경계로 처리한다. clock/namespace/레코드 무결성을 확인할 수 없으면 해당 engine poll을 fail-closed로 유지하고 오류를 드러낸다. rate 간격은 이 레코드와 독립적인 매 소유권 획득 후 상대 대기로 지킨다. 시스템 clock 직접 호출을 주문 경로에 흩뿌리지 않고 clock provider로 주입한다. 이 복구 방식은 해당 제품 로트에서 source 적합성과 수용 시험을 확인한다.

공유 이벤트를 읽는 a087/a169의 정지 규칙은 각 change에서 정한다. 공유 패널티를 새로 도입하려면 보호까지 추가로 멈추는 정책을 별도로 명시해야 하며, 사용자에게 확정되지 않은 2판의 기본값을 복원하지 않는다.

## 수용 증거

1. 실제 TLS 요청·정상 분할 response body·조기 응답·느린 request body·쓰기 직전 정지·취소·늦은 dial에서 start/drain/release 순서를 기록한다. header 반환과 종료를 구별하지 않은 변이를 잡는다.
2. 독립 프로세스 최소 두 개가 같은 inode를 쓰고, 각 요청 시작을 계수하여 rolling 1초 상한을 검사한다. 진행 중 소유자 종료·새 참여자·시계 변경 조건을 별도로 검사한다. 프로세스 내부 mutex만으로 통과시키지 않는다.
3. 실제 `official.Client` token 교환·자동 401·redirect와 TLS 인증 실패를 검사한다. 수동 두 번째 `client.Get`만으로 자동 인증 재전송 시험을 대체하지 않는다.
4. 보호/normal/auxiliary 경합, 대기자 취소·사망, state 오류와 partial write, 모든 429 이벤트·해당 RequiredQuery의 유효 hint 제한·350ms 간격을 검증한다. OFF wire golden과 authority override 거절을 유지한다.
5. 전체 호출자 배선·Pre-Edit·FLM/BTM·targeted RED/GREEN/race/변이·gstack 리뷰·독립 검토를 완료한다. 전체 시험이나 gate를 자동 실행하지 않는다.

## 제품 통합 순서와 계약

첫 로트는 **사용되지 않는 내부 모듈**이다. `internal/accountbudget`에 공유 소유권·큐·context·응답 수명·오류 타입을 구현하고 fake/실제 TLS·독립 프로세스 시험을 수행한다. 기본 클라이언트, 실행 경로, 환경 기본값은 이 로트에서 바꾸지 않는다. 이후 각 로트는 앞 로트의 고정 SHA와 집중 검증을 입력으로 받는다. 연구 코드를 무검토 복사하거나 새 내부 모듈의 PASS를 전체 배선 완료로 표시하지 않는다.

공식 클라이언트 로트는 constructor-owned `*http.Transport`와 새 guard의 결속을 private 상태로 봉인한다. 설정되지 않은 OFF에서는 기존 코드·timeout·HTTP2·proxy 설정·authority 판정이 같아야 한다. enforcing에서만 검증한 HTTP/1 직접 TLS 구성을 사용한다. 외부 `WithHTTPClient`/base override는 여전히 금전 권한이 없으며, 외부에서 만든 임의 wrapper를 unwrapping하여 권한을 주지 않는다. 원래 transport와 허가 모듈이 일치하는 clone만 측정 전용 경로에 연결한다.

배선 로트는 `official.New`의 모든 호출자뿐 아니라 token 교환, `a112MBUSReadAt`의 별도 clone, `HTTPReplayTransport`, a087 probe의 직접 POST를 포함한다. 원장 재생 본문·멱등키·a165 첫 AMBIGUOUS 보존을 바꾸지 않는다. WTS monitor와 a128 judge HTTP는 공식 API로 분류하지 않는다. WTS와 동일 브로커 예산인지는 미확인이므로 지원환경에서 별도로 확인하며, 실제 활성화 전에 영향을 모르는 외부 소비자를 동시에 켜놓고 계좌 전체 보장을 주장하지 않는다.

enforcing 배포는 **기존 예산을 연다**. 각 프로세스가 시작 때 임의 폴더·instance를 만들지 않는다. 첫 지원 방식은 설치의 모든 참여 profile/계좌를 **하나의 보수적 예산 그룹**에 넣는 것이다. 그룹 총량이 3건/초 이하이면 각 계좌의 부분집합도 그 상한 이하다. 여러 계좌를 함께 쓰면 합산 처리량도 제한된다. 계좌마다 독립 3건/초를 제공하거나 서로 다른 앱 키의 accountSeq가 같은 물리 계좌라고 주장하지 않는다. 기존 계좌 선택·권한·원장 검증은 유지한다. 기존 「계좌 식별에 결속한 scope」는 이 명시적 group scope로 대체한다.

신뢰된 공통 launch manifest가 모든 참여자에 동일 opaque scope·instance 및 예상 root/queue.lock/wire.lock device·inode를 고정한다. factory는 trusted clock.System 계열로 기존 예산을 열어 이 예상 파일 ID와 비교한 뒤 token/accounts를 포함한 첫 요청부터 같은 guard를 사용한다. 입력 누락·다른 ID·미지원 상태는 bootstrap 전 wire 0이다. 열려는 root에서 marker를 읽고 그 값을 자기 승인 기준으로 삼지 않는다. 같은 marker 문자열을 복사한 다른 디렉터리도 거절한다. manifest를 임의 교체하는 관리자까지 막는다고 주장하지 않으며 모든 launch 경로의 실제 일치는 운영 전 사람 확인이다. 같은 프로세스의 재생성 client는 Budget 객체를 재사용하고 다른 프로세스는 동일 파일/instance를 검증한다. 계좌번호·키·토큰을 scope/경로/event에 넣지 않고 새 계좌 동등성 데이터베이스를 만들지 않는다. bootstrap을 무예산으로 보낸 뒤 나중에 합치지 않는다. 상세 근거·수용 반례는 `analysis/budget-group-binding.md`다.

생성·점검은 운영자용 로컬 provisioning 절차로 분리한다. directory/파일 소유권·mode·symlink/hardlink·regular-file·fs 종류·device/inode·boot/time namespace를 검증한다. active 참여자가 있는 상태에서 root·lock inode·instance를 바꾸거나 자동 복구하지 않는다. 지원하지 않는 host/namespace/fs와 검증 실패는 enforcing에서 새 wire 0이며 upstream으로 돌아가지 않는다. non-Linux는 enforcing 미지원으로 명시하고 OFF의 기존 지원을 유지한다. CLI 플래그/env를 추가하더라도 실제 운영 값을 자동으로 설정하지 않는다.

마지막 통합 로트에서 엔진 폴링의 응답별 429 hint, replay·token·별도 측정 전송, a087 후속 활성화 API, a169 공유 이벤트 독자를 연결한다. a087 admission의 production nil은 이 change에서 해제하지 않는다. a169 도구 구현은 검증된 계좌 admission/event API를 받는 후속 작업이다.

## 착수와 검증의 구분

READY는 구현 계약·기술 가능성·실패 처분·증거 계획이 확정되고 독립 설계 리뷰가 통과했다는 뜻이다. 아직 작성하지 않은 제품 구현의 모든 RED/GREEN을 설계 단계 PASS로 앞당기지 않는다. source/AST/CodeGraph·Pre-Edit는 각 기존 함수 편집 전에 최신화한다. 연구의 제한된 GREEN은 구현 선택의 근거이며, 제품에 대한 최종 Manager 승인·아카이브·운영 활성화는 각각 후속 판정이다.

## skill 적용과 검증 상태

grill-me/grilling: 확정 정책을 재질문하지 않고 미증명 요구와 수용 기준을 분리했다. search-first: 기존 flock·stdlib 후보, TLS 국소 연구와 공식 Go 문서를 조사했다. 버전·라이선스·미확인 채널은 연구 README와 `analysis/manager-review-v4.md`에 기록했다. impeccable: N/A — backend/documentation-only; UI 변경 없음. 현재 판정은 개발 착수이며 제품 구현·실제 계좌 합산·배포 검증 완료를 뜻하지 않는다.

환경 차이: 초기 연구는 Go 1.26.5, 저장소 `go.mod`와 `Dockerfile`의 build 버전은 Go 1.25.0이다. 독립 Linux fixture는 두 버전에서 통과했다. 실제 Alpine 이미지·운영 환경 호환 PASS로 확대하지 않으며 배포 toolchain을 조용히 올리지 않는다.
