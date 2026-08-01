# Private operator HTTP API

`tossctl httpapi`는 console/engine과 분리된 private REST/SSE 프로세스다. 엔진을
자동 시작하거나 중지하지 않고, 주문·LIVE·automation gate·kill switch·activation
manifest 경로를 갖지 않는다. 기본 모드는 application token과 login 화면이 없는
loopback/VPN 읽기 전용 모드다.

## 소비자 UX 계약

모바일·웹 소비자는 [OpenAPI 계약](api/openapi-v1.json)을 그대로 사용한다. 화면은
StockOS처럼 개요 카드, 상태 pill, 고정 category tab과 서버가 제공한 preset chip/tile로
구성한다. 종목·가격·수량·사유·숫자 범위·확인 문구를 입력받지 않는다. 조회 경로는
query와 body를 모두 거부한다. optimization도 owner descriptor가 제공한 option ID만
표시하며, 미승인 threshold를 `0`으로 바꾸지 않는다.

고정 resource는 다음과 같다.

- `GET /api/v1/engine`
- `GET /api/v1/positions`
- `GET /api/v1/orders`
- `GET /api/v1/candidates`
- `GET /api/v1/performance`
- `GET /api/v1/settings`
- `GET /api/v1/optimization`
- `GET /api/v1/stream`

SSE ID는 `<process-epoch>:<sequence>`다. 같은 epoch의 연속 ID만 replay하고, restart,
unknown ID 또는 64-event history gap은 full snapshot으로 수렴한다. 최대 32 client,
client당 queue 64, heartbeat 15초이며 느린 client만 끊는다.

브로커를 읽는 positions/orders는 process 전체에서 30초 cache와 singleflight를 공유한다.
따라서 REST 반복 조회나 32개 SSE 재연결이 official API 호출을 곱하지 않는다. 정상 상태의
고정 예산은 30초마다 holdings 1회와 order 화면의 bounded read 3회, 합계 최대 8회/분이다.
실패는 5초 negative cache로 묶고 자동 빠른 재시도는 하지 않는다. 다른 resource는 local
read-only DB/marker만 사용한다.

## Native 실행

인증서 SAN에 `api.vpn.example`이 있어야 하고 사설 CA는 소비자 OS가 신뢰해야 한다.

```bash
tossctl --config-dir /srv/tossos/config \
  --session-file /srv/tossos/secrets/session.json \
  httpapi \
  --port 37086 \
  --bind 10.8.0.1 \
  --allowed-cidr 127.0.0.0/8 \
  --allowed-cidr 10.8.0.0/24 \
  --public-url https://api.vpn.example:37086 \
  --tls-cert /srv/tossos/secrets/tls.crt \
  --tls-key /srv/tossos/secrets/tls.key
```

public IP, `/0`, HTTP origin, TLS 없는 listener, 일부만 설정한 proxy forwarding은
startup에서 거부한다. forwarded header는 `--trusted-proxy`에 정확한 한 IP를 지정하고
`--tls-forwarded`를 함께 쓴 경우에만 한 hop을 인정한다.

## 제한 mutation

기본 Compose는 mutation을 노출하지 않는다. 그러면 모든 POST가 404/405다. signed
capability를 쓰는 별도 운영 환경에서만 아래 두 값을 함께 지정한다.
감사 DB 부모는 서비스 계정(또는 root)이 소유한 기존 전용 디렉터리이며
mode `0700`이어야 한다. daemon은 symlink, 다른 계정 소유 component,
group/other writable component를 거부하고 이 디렉터리를 대신 만들지 않는다.

```bash
install -d -m 0700 -o tossos -g tossos /srv/tossos/httpapi-security
--security-db /srv/tossos/httpapi-security/security.db \
--capability-public-key /srv/tossos/secrets/httpapi-capability-public.key
```

private signing key는 daemon과 repository에 두지 않고 local human approval channel이
보관한다. capability는 actor/client, canonical HTTPS audience, exact POST resource,
canonical JSON digest, idempotency key, `If-Match`, one-time nonce와 최대 60초 expiry에 묶인다. server는
append-only audit와 actor-scoped a050 commander를 추가로 검사한다.

현재 writable owner field인 `exit.common-policy`는 safety direction이 `contextual`이다.
따라서 remote preview로 before/after를 볼 수는 있지만 application은 local human
channel 외부에서 항상 거부된다. 향후 owner가 `neutral`을 명시한 finite preset만 같은
generic application 경로를 사용할 수 있다. rollback-preview 경로는 제공하지 않는다.

## 배포·rollback

Compose의 `httpapi` service는 `${TOSSOS_HTTPAPI_CONTAINER_IP}` private 주소에 직접
bind하고 `${TOSSOS_VPN_BIND_IP}:${TOSSOS_API_PORT}:37086`에만 publish된다. wildcard
bind는 시작 단계에서 거부된다. console service와 engine lifecycle은 그대로 유지된다.
API만 회수할 때:

healthcheck는 container private IP로 연결하되 `TOSSOS_API_PUBLIC_URL`의 canonical
Host를 전송하므로 TLS/origin boundary를 우회하지 않는다. 실제 client는 private CA를
신뢰하고 public URL의 SAN과 유효기간을 검증해야 하며 `--no-check-certificate`를 사용하지 않는다.

1. mutation을 구성했다면 host의 `sqlite3`로 pending row가 0인지 확인한다.
2. `docker compose stop httpapi`로 API service만 정지한다.
3. console/engine service는 중지하거나 gate·autostart 상태를 바꾸지 않는다.

```bash
sqlite3 /srv/tossos/httpapi-security/security.db \
  "select count(*) from mutation_idempotency where state='pending';"
docker compose stop httpapi
```

0이 아니면 회수하지 않는다. pending request의 결과를 확인하고 동일 image로 복구한
뒤 durable outcome을 확정한다. database를 삭제하거나 pending row를 수동 수정하지
않는다.
