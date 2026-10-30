# TLS 전송 연구 (2026-10-31)

기준 `e32606d4bd44de3142f3c732ab86ecee913e3ab0`; Go 1.26.5; 표준 라이브러리만. 이 디렉터리는 실험이며 제품 코드가 아니다. 테스트 전용 자체 서명 TLS server만 사용했다.

`GO111MODULE=off go test -race -count=1 -timeout=30s ./openspec/changes/a166-account-shared-rate-limit/analysis/transport-prototype` → PASS, 3.326s. 최초 실행은 TLS `ServerName` 누락으로 실패했으며 보완 후 통과했다.

- HTTP/1, 연결 재사용 비활성, 직접 TLS 연결 및 요청별 Write/revoke 상호배제. 요청 종료 뒤 연결을 폐쇄하고 다음 호출 전 350ms 대기한다.
- 로컬 307 redirect → 401 → 별도 호출 세 요청의 TLS plaintext 첫 Write 시작 간격, 첫 Write 선점 중 다음 진입 차단, 취소 뒤 늦은 dial의 Write 0, 느린 요청 본문과 조기 응답을 확인했다. 첫 쓰기와 락 경합은 프로세스 내부 계측이다. 401은 수동 호출이며 공식 클라이언트/토큰 갱신 시험이 아니다.
- `httptrace.WroteRequest`는 outer buffered flush보다 먼저일 수 있어 배제했다. [Go net/http](https://pkg.go.dev/net/http), [Go httptrace](https://pkg.go.dev/net/http/httptrace); 설치 소스 `/usr/local/go/src/net/http/request.go:582-589`, `transport.go:2648-2675`.
- 미검증: 다중 프로세스 flock, 종료·재기동 복구, 소켓 내부 재시도 실패 주입, 401 토큰 교환, 보호 우선·429, TLS 인증 실패와 timeout, 공식 클라이언트 authority·OFF 호환, 변이시험 및 production 전 경로. 서버 도착·커널/NIC wire 시각 보장 불가. 실제 사용 가능한 엄격 wire 3/s 증명이나 활성화 승인이 아니다.
- gstack review 체크리스트 수동 적용: 전역 락은 `sync.Mutex`여서 프로세스 간 공유가 되지 않고 실제 운영에 부적합; `time.Sleep`은 context 취소를 무시; fixture TLS config 및 오류 처리도 제품용 아님. 제품 구현에 복사하지 않는다. 외부 모델 CLI 검토 미실행.
