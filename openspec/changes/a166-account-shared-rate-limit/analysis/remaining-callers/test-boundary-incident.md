# 시험 범위 이탈 — 2026-10-31

verify RED에서 `http.DefaultTransport` 차단이 `official.New` 자체 transport에도 적용된다고 잘못 가정했다. 계좌 식별 결과 HTTP 401이 관찰되어 외부 브로커 접속 가능성이 있다. 네트워크 trace나 요청별 로그는 없으므로 실제 전송한 method·endpoint와 wire 횟수는 확인하지 못했다. 소스상 첫 인증은 `POST /oauth2/token`, 이어지는 계좌 조회는 `GET /api/v1/accounts`이며, 어느 단계의 401인지는 원문만으로 확정하지 않는다. 생산 기본 URL을 호출할 수 있었고 proxy 환경이 개입할 수 있다.

`isolate`는 `TOSSCTL_OPENAPI_KEY`, `TOSSCTL_OPENAPI_SECRET`을 빈 값으로 설정하고 `XDG_CONFIG_HOME`, `XDG_CACHE_HOME`을 임시 디렉터리로 바꾼다. 생성된 credential 파일은 고정 시험값이다. 다른 환경 변수값 및 자격 내용은 기록하지 않는다. 주문 호출과 실계좌 자격 사용은 없었다. 해당 fixture는 삭제했고 의도된 RED 증거로 세지 않는다. 원문 `/tmp/a166-remaining-verify-red.log`.

`official.New`는 독립 `http.Transport`를 생성하고 token manager에도 연결한다. 이후 네트워크 fixture 재개 전 실제 client의 endpoint와 transport를 local TLS로 고정하고, 변이가 외부 전송을 할 수 없음을 증명하는 계획에 대해 Manager 확인이 필요하다.
