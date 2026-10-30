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
