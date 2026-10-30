# TLS 전송 연구 (2026-10-31)

기준 `e32606d4bd44de3142f3c732ab86ecee913e3ab0`; Go 1.26.5; 표준 라이브러리만. 이 디렉터리는 실험이며 제품 코드가 아니다. 테스트 전용 자체 서명 TLS server만 사용했다.

`GO111MODULE=off go test -race -count=1 -timeout=30s ./openspec/changes/a166-account-shared-rate-limit/analysis/transport-prototype` → PASS, 3.326s. 최초 실행은 TLS `ServerName` 누락으로 실패했으며 보완 후 통과했다.

- HTTP/1, 연결 재사용 비활성, 직접 TLS 연결 및 요청별 Write/revoke 상호배제. 요청 종료 뒤 연결을 폐쇄하고 다음 호출 전 350ms 대기한다.
- 로컬 307 redirect → 401 → 별도 호출 세 요청의 TLS plaintext 첫 Write 시작 간격, 첫 Write 선점 중 다음 진입 차단, 취소 뒤 늦은 dial의 Write 0, 느린 요청 본문과 조기 응답을 확인했다. 첫 쓰기와 락 경합은 프로세스 내부 계측이다. 401은 수동 호출이며 공식 클라이언트/토큰 갱신 시험이 아니다.
- `httptrace.WroteRequest`는 outer buffered flush보다 먼저일 수 있어 배제했다. [Go net/http](https://pkg.go.dev/net/http), [Go httptrace](https://pkg.go.dev/net/http/httptrace); 설치 소스 `/usr/local/go/src/net/http/request.go:582-589`, `transport.go:2648-2675`.
- 미검증: 다중 프로세스 flock, 종료·재기동 복구, 소켓 내부 재시도 실패 주입, 401 토큰 교환, 보호 우선·429, TLS 인증 실패와 timeout, 공식 클라이언트 authority·OFF 호환, 변이시험 및 production 전 경로. 서버 도착·커널/NIC wire 시각 보장 불가. 실제 사용 가능한 엄격 wire 3/s 증명이나 활성화 승인이 아니다.
- gstack review 체크리스트 수동 적용: 전역 락은 `sync.Mutex`여서 프로세스 간 공유가 되지 않고 실제 운영에 부적합; `time.Sleep`은 context 취소를 무시; fixture TLS config 및 오류 처리도 제품용 아님. 제품 구현에 복사하지 않는다. 외부 모델 CLI 검토 미실행.

## 보강 실험 (2026-10-31)

`body-red.log`: 기존 `RoundTrip` 반환 시 연결을 닫는 구현에서 분할 500 KB 본문은 100 byte만 읽고 실패했다. revoke에서 연결 close를 제거해 쓰기 금지와 읽기 생명주기를 분리했다. `green.log`: Go 1.26.5, `go test -race -count=1 -timeout=40s ./openspec/changes/a166-account-shared-rate-limit/analysis/transport-prototype` 9.136s 통과. 정상 분할 500 KB io.ReadAll, 다른 인증서·잘못된 hostname 거절, 각 두 요청을 보내는 독립 프로세스 2개의 같은 파일 flock 및 첫 plaintext Write 네 시점 간격, lock 보유 프로세스 강제 종료 뒤 잠금 재획득을 확인했다. 강제 종료 시험은 lock만 잡은 owner를 죽였으며 전송 중 종료/재시작 검증은 아니다.

실제 `official.Client`를 로컬 TLS fixture에 `WithHTTPClient`로 연결하여 token→401→token→200 네 호출과 네 plaintext Write 및 `AuthorityOrigin` false를 확인했다. 래핑 연결에서는 `resp.TLS == nil`을 별도 확인해 호환성 결함으로 남겼다. 생산 constructor-owned authority와 TLS 정보 보존은 미해결이다. 자료는 fixture이며 운영 권한이 아니다.

아직 blocking `sync.Mutex`·`flock`·`time.Sleep`에서 context 취소를 존중하지 않는다. 이 점은 **구현 차단 결함**이다. 첫 Write evidence는 wall clock nanos이며 일회성 표본이고, clock 역행·프로세스 손상·inode 교체·정지·kill 시 진행 중 전송·실제 broker arrival·kernel buffer·중간 proxy·priority·429·전송 내부 재시도 오류 주입은 검증되지 않았다. `lease.revoke`가 기존 응답 read를 허용하도록 바뀌었으나 느린/취소 body의 완전한 연결 종료와 누수, write 대기 영구 정지에 대한 유한 상한은 미입증이다. 공식 Client 관문 시험은 fixture로 production authority를 증명하지 않는다.

## 경계 추가 시험

TLS handshake를 마친 동일 연결의 `ConnectionState`를 보관하고 `VerifiedChains`가 있는 경우에만 response의 빈 TLS 상태를 복구했다. `tls-state-red.log`의 복구 전 실패와 `final-green.log`의 이후 성공, 잘못된 CA/hostname 거절 fixture를 확인했다. fixture에서의 결과이며 production constructor-owned authority·TLS 전체 호환 증거는 아니다.

별도 `acquireContext` 연구 함수는 context-aware 채널, `LOCK_NB` 재시도와 350ms timer에서 각각 deadline 종료를 확인했고 취소된 대기의 mock transport 호출 수는 0이다. `cancel-mutant.log`는 cooldown 취소 오류를 nil로 바꾸면 시험이 실패함을 보인다. 이 함수는 기존 TLS limiter에 통합되지 않았다. 채널/파일 이름·inode 검증·우선순위·fairness·deadline 경계의 전체 계약은 미완이다.

TLS 첫 plaintext Write 이후 서버 handler 대기 중 owner child를 SIGKILL하고 새 child가 같은 flock을 획득해 후속 Write까지 350ms 이상 간격을 둔 일회성 fixture 통과. wall clock 계측이고 다양한 OS·중지·clock 이상·kernel 잔류 바이트 증거가 아니다. HTTP/1 새 연결에서 서버가 요청을 받은 뒤 연결을 강제 종료했을 때 hidden retry 없이 서버 요청과 첫 Write 각 1회인 fixture 통과. Go 1.26.5 `net/http/transport.go:815-845`의 reused 여부 분기 손 읽기와 부합하며 다른 모든 실패·버전의 증명이 아니다. AST 도구는 실행하지 않았다.

실제 source `internal/official/client.go:376,399,436`은 nil 본문 GET/DELETE와 `bytes.NewReader(encoded)` POST를 사용하고, `internal/official/token.go:138`은 `strings.NewReader(form.Encode())`를 쓴다. 자체 `Request.Write`/`ReadResponse` one-shot 전송은 가능성만 비교했으며 TLS, 취소, proxy, redirect, 오류, body 생명주기와 기존 권위·호환 계약을 새로 책임져야 한다. 표준 Transport 래핑 후보가 아직 실제 계약을 충족한다고 보지 않는다.

수동 gstack 점검 결과: 유효한 상태만 복구하고 인증 오류를 보존해야 하며, deadline 경합·무한 Write 및 body 누수·owner 죽음 후 inode 교체·재시도·동시 프로세스 우선권은 추가 검증이 필요하다. 전체 게이트·외부 CLI review 미실행. 최종 국소 테스트 `go test -race -count=1 -timeout=40s ./openspec/changes/a166-account-shared-rate-limit/analysis/transport-prototype` PASS 11.830s.

## 통합 TLS limiter 연구

동일 prototype RoundTrip에 context-aware process-local channel, `LOCK_NB` + context 재시도, acquire 후 350ms timer, write revoke, 검증된 TLS state 복원을 연결했다. `go test -race -count=1 -timeout=40s ./openspec/changes/a166-account-shared-rate-limit/analysis/transport-prototype` → `integrated-green.log` 13.031s. 이 test-only 구현은 nil lockPath를 허용하므로 production에 사용하지 않는다.

| 경계 | 통합 prototype fixture 결과 | 미검증·한계 |
|---|---|---|
| 획득 취소 | local·다른 fd flock·cooldown timeout 시 실제 limiter Write/서버 요청 0 | 취소와 허가 경합 전체, goroutine 정지 상한 |
| 응답 수명 | 분할 500 KB `ReadAll`, EOF·조기 Close·context cancel에서 서버 연결 종료 | 강제 block Write 중 cancel→Close→revoke의 실제 종료, 서버 악의적 행위 |
| TLS | 검증된 연결 상태 복원, 미신뢰 CA·잘못된 hostname 거절 | production authority, 모든 TLS state·버전·프록시 호환 |
| 재전송 | official.Client fixture token→401→token→200, 307, 새 HTTP1 연결 서버 강제종료 각 시도 계수 | 전 운용 진입점·모든 retry 원인 |
| 다중 프로세스 | 두 child 네 시작 간격, 첫 Write 후 owner kill과 새 child | wall clock 이상·부팅 재시작·inode 변경·공유장애·운영 부하 |
| 변이 | `revoke-mutant.log`: revoke 삭제 후 write 허용해 실패. `integrated-cancel-mutant.log`: cooldown 취소 오류 nil 변이에서 통합 호출 panic으로 실패 | 두 개 변이 외 전체 분기 변이 아님 |

기록된 시작은 TLS plaintext 첫 `Write` 진입이며 kernel/TLS record 전송 또는 브로커 도착률이 아니다. 시간 기록은 wall UnixNano이며 단일 실험이다. lock 단일 inode·단일 host·모든 참여자 준수 가정, priority·429·파일 무결성·authority 계약·rollout·경합 중 deadline에는 독립 설계/시험이 필요하다. gstack 수동 검토: wait 취소 함수가 실제 limiter에 연결됐지만 local/공유 획득 사이 취소 경합, 거부 시 request body Close, Close/error 처리, 느린 TLS Write와 취소, late dial, 장기간 누수는 추가 증거가 필요하다. 표준 `net.Conn.Close`는 block된 read/write를 깨우는 계약이고 Go1.26.5 TLS Close는 진행 중 Write에 underlying close를 사용한다; 강제-block 통합 재현은 미실행. 외부 CLI review·전체 gate 미실행. 제품 채택 결정이 아니라 기술 가능성의 제한된 근거다.

## transport 후보 검증 추가

전송에 넘기기 전 거절의 request Body.Close 누락을 `bodyclose-red.log`로 확인해 수정했다. body Close 삭제 변이 `bodyclose-mutant.log`도 세 대기 단계에서 실패했다. 로컬 잠금·flock·cooldown 각 단계의 실제 limiter 취소는 body Close 1회, TLS Write 0을 확인했다.

서버가 request body를 읽지 않는 실제 로컬 TLS fixture에서 128 MiB body 전송 중 guarded TLS Write가 진행 중임을 확인하고 context 취소했다. 호출이 종료되고 Body.Close 1회, 후속 GET 및 body read가 성공했다. response body read가 막힌 상황의 context 취소와 EOF·early Close 연결 정리도 확인했다. 첫 실험의 서버 handler는 미읽은 body 때문에 서버 context 종료만 기다리다가 fixture `Server.Close` timeout이 발생했다. 제어 채널을 추가해 fixture를 정리했고 초기 timeout의 원문 로그는 덮어써 보존하지 못했다. 현재 테스트는 소켓에서 임의 스케줄 중단이나 절대 지연 상한을 보장하지 않는다.

`transport-green.log`: `go test -race -count=1 -timeout=40s ./openspec/changes/a166-account-shared-rate-limit/analysis/transport-prototype` PASS 14.545s. 401·redirect·2 process·kill·TLS verify·fresh HTTP1 retry fault와 앞선 변이 기록을 포함한다. 취소 직전/동시 전송 경합은 하나의 검증 완료 증명이 아니라 기존 fixture 범위의 관찰이다.

판단: 직접 TLS·HTTP/1·keepalive OFF·요청당 연결 및 신뢰할 수 있는 단일 host 참여자라는 조건 아래, 이 transport 생명주기는 **후속 독립 리뷰를 거쳐 설계 후보로 채택 가능**하다. 제품 준비나 strict wire/server arrival 보장은 아니다. 취소와 첫 실제 write의 모든 동시 경계, 반복 시험·버전 고정·권위 및 OFF 호환, 공통 inode 신뢰·priority·429·운영 정책은 별도 증거/설계가 필요하다. 적용 전 독립 리뷰가 본 source의 context 경합과 io.Close 오류 처분을 확인해야 한다. gstack 수동 점검상 production-grade 보안·지원환경·실패 분류는 여전히 미완이며 전체 gate·외부 CLI review 미실행.

## 응답 본문 전체 소유권 (v4 검증)

앞선 `52d64223a`는 response header 반환 시 lock을 해제하므로 v4의 본문 전체 소유권을 충족하지 않았다. 이번 연구는 response wrapper가 EOF·read error·Close 또는 원래 request context 취소 시 원본 body close 후 write revoke/drain, 공유 flock과 local gate 해제를 `sync.Once`로 수행한다. 추가 timeout을 만들지 않는다. `http.Client.Timeout=2s`와 Background 요청을 사용해 호출자가 body를 닫지 않아도 기존 client timeout이 해제하는 fixture를 확인했다. Background 및 timeout 0이며 호출자도 닫지 않는 경우 무기한 보유는 명시적 미완 사용 계약이다.

느린 body를 보낸 첫 요청이 유지되는 동안 둘째 요청을 500ms 관찰했다(350ms cooldown보다 김). EOF·Close·cancel·기존 client timeout 전 둘째 요청은 서버에 도착하지 않았고 각 종료 뒤 진행했다. header에서 release하는 변이 `earlyrelease-mutant.log`는 `second request during first body`로 실패한다. 최초 fixture는 채널 변수 재할당으로 data race 및 이중 close 실패했고 해당 원문은 덮어쓰기로 보존하지 못했다. 이를 `sync.Once`로 고친 뒤 targeted race GREEN.

최종 전체 국소 회귀 `go test -race -count=1 -timeout=50s ./openspec/changes/a166-account-shared-rate-limit/analysis/transport-prototype` → `fullownership-green.log` PASS 20.546s. 128MiB TLS blocked Write 취소·본문 Close·후속 요청, pre-admission body close, 분할 본문, verified TLS state, 401/307, 두 프로세스·kill 재시작 시험을 포함한다. 표준 HTTP1 신선한 연결의 단일 실패 fixture이지 모든 장애/버전의 숨은 재시도 증명은 아니다. production authority·OFF·inode 신뢰·공정성/priority·429·lock 오류·전 경로 wiring은 별도 scope. 서버 도착률과 무한 scheduler 정지의 유한 상한은 보장하지 않는다. 수동 gstack 관점에서 `sync.Once`에 body.Close를 포함한 상태에서 요청 취소와 사용자 Read/Close가 경쟁할 때의 모든 구현체 동작, Close 오류 처리 및 writer drain·timeout의 production 통합은 독립 검토가 필요하다.
