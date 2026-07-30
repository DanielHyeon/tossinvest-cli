# Operations

운영 측면의 가이드 — API 회귀 감시 cron 설정, 알림 채널 등.

## VPN 원격 콘솔

기본 `tossctl console`은 계속 `127.0.0.1` HTTP와 터미널에 출력된 session URL만
사용한다. Compose 원격 모드는 이 기본값의 완화가 아니라 별도 배포 모드다. 단일
운영자의 명시적 선택으로 host loopback 또는 인증된 VPN membership을 application
접근 경계로 사용하며 별도 로그인은 없다. HTTPS, 실제 peer CIDR, 정확한 Host/Origin,
기존 CSRF는 계속 통과해야 한다. 원격 접근은 조회 전용이 아니며 현재 콘솔이 제공하는
설정·엔진 제어·검증 기능을 그대로 쓸 수 있지만,
새 주문 API나 자동 승인을 추가하지 않는다.

### 사전 준비

호스트의 실제 VPN 주소와 VPN client 대역을 먼저 확인한다. 아래 값은 예시이므로
그대로 쓰지 않는다.

```bash
ip -brief address
ip route
```

모바일에서 해석 가능한 VPN DNS 이름(예: `console.vpn.example`)을 준비하고 그
이름이 SAN에 포함된 인증서를 발급한다. 사설 CA를 쓰면 해당 CA를 모바일 OS의
신뢰 저장소에 설치한다. `--insecure` 우회는 없다.

native 실행 예:

```bash
tossctl console \
  --port 37085 \
  --bind 10.8.0.1 \
  --allowed-cidr 10.8.0.0/24 \
  --public-url https://console.vpn.example:37085 \
  --tls-cert /srv/tossos/secrets/tls.crt \
  --tls-key /srv/tossos/secrets/tls.key \
  --trusted-network
```

모든 remote flag는 한 세트다. 일부만 주거나, 전역 CIDR(`/0`), HTTP URL,
URL host와 맞지 않는 인증서 또는 접근 모드 누락이면 기동 전에
실패한다. `X-Forwarded-For`는 신뢰하지 않으므로 NAT가 client VPN 주소를 보존하지
않는 환경에서는 요청이 닫힌 상태로 거부된다.

Compose는 기본적으로 전용 `172.30.85.0/24` bridge를 만들고 그 대역과 실제 VPN
client 대역을 각각 허용한다. 이 bridge는 호스트의 loopback publish를 거쳐 접근할
때 보이는 source 주소를 한정하기 위한 것이다. 해당 대역이 호스트·VPN의 기존 route와
겹치면 `.env`의 `TOSSOS_CONTAINER_CIDR`을 사용하지 않는 사설 대역으로 바꾼다.
컨테이너 종료 유예는 75초다. 엔진이 실행 중이면 최대 60초 동안 현재 loop와 journal
정합 close를 기다린 후 컨테이너를 종료하기 때문이다.

### Docker Compose

`.env.example`을 `.env`로 복사하고 실제 값을 입력한다. 세 secret 원본은 모두
저장소 밖에 두고 `chmod 600`으로 제한한다. `TOSSOS_UID=$(id -u)`와
`TOSSOS_GID=$(id -g)`를 설정하고 두 값이 0이 아닌지 확인한다. Compose는 이
non-root identity로 실행하므로 0600 secret을 읽을 수 있고 config/data directory도
같은 사용자가 읽고 쓸 수 있다. image 자체의 기본 identity는 `10001:10001`이다.

```bash
cp .env.example .env
install -d -m 700 /srv/tossos/config /srv/tossos/data
docker compose config
docker compose build
docker compose up -d
docker compose ps
```

Compose의 host publish는
`${TOSSOS_VPN_BIND_IP}:${TOSSOS_CONSOLE_PORT}:37085`이며 bind IP가 없으면
`docker compose config`부터 실패한다. `.env`에 `0.0.0.0`이나 LAN/public interface
주소를 넣지 않는다. 컨테이너 안의 `--bind 0.0.0.0`은 port NAT를 받기 위한
것이며 host publish 제한과 애플리케이션 CIDR 검사를 없애지 않는다. 호스트
방화벽에서도 해당 VPN interface와 client CIDR만 허용한다.

`TOSSOS_DATA_DIR`은 `journal.db`와 `audit.log`가 직접 들어 있는 기존 TossOS
애플리케이션 데이터 디렉터리를 가리킨다. Compose는 이를
`/var/lib/tossos/data`에 직접 마운트하고 컨테이너의 `TOSSOS_DATA_DIR`도 같은
경로로 고정한다. XDG base directory를 가리키는 값이 아니므로
`.../data/tossos`처럼 한 단계 더 중첩하지 않는다.

`/healthz`는 GET/HEAD에 `ok`만 반환하며 계좌·설정 상태를 공개하지 않는다.
모바일에서는 먼저 VPN 연결 후 `TOSSOS_PUBLIC_URL/healthz`, 이어서 `/`를 확인한다.
별도 application login이나 token 입력은 없다.

### LIVE 주문과 엔진 자동 시작

설정 화면의 **거래 정책** 네 토글과 **자동화 게이트**는 엔진이 프로그램 주문을
낼 수 있는지 승인한다. 별도의 **엔진 자동 시작**은 프로세스 수명주기만 승인하며
기본 OFF다.

- 자동 시작 ON 저장은 기존 [엔진 시작]과 같은 경로로 즉시 기동을 한 번 시도한다.
- Docker/콘솔이 재기동될 때 ON이면 같은 시도를 한 번 반복한다.
- automation gate, Guardian 한도, capability attestation, 거래 정책, journal
  flock 중 하나라도 기존 startup interlock을 통과하지 못하면 엔진은 거부된다.
- 자동 시작 OFF는 다음 기동만 막는다. 현재 엔진은 대시보드의 [엔진 정지]로
  graceful shutdown해야 한다.
- Compose의 `restart: unless-stopped`와 enabled 상태의 Docker service가 부팅 후
  콘솔을 다시 띄우고, 콘솔이 이 설정을 판정한다. 별도 privileged daemon이나
  Docker socket mount는 없다.

배포·업데이트 절차는 `engine.autostart`를 임의로 켜지 않는다. ON은 trusted-network
브라우저의 CSRF 설정 폼에서 운영자가 직접 저장하고 audit에 남긴다.

### 교체·회수·복구

- TLS 인증서는 같은 public host SAN을 유지해 교체하고 container를 재생성한다.
- broker session secret은 host에서 `tossctl auth login`으로 사람이 갱신한 뒤
  container를 재생성한다. container 안에서 QR 인증을 자동화하지 않는다.
- container에서는 콘솔의 self-update를 운영 절차로 사용하지 않는다.
  `docker compose build` 또는 서명·검증된 image pull 후
  `docker compose up -d`로 교체한다.
- rollback은 이전 image tag로 `docker compose up -d`하거나 remote flag를 제거한
  native loopback 실행으로 돌아간다. rollback도 automation gate를 켜거나 engine을
  시작하지 않는다.

인터넷 router port-forward, `0.0.0.0:37085` host publish, public cloud security
group 개방은 이 명세 밖이며 금지한다. 다중 사용자가 필요하면 trusted-network
범위를 넓히지 말고 IdP/RBAC를 별도 설계한다.

## API 회귀 감시 (`tossctl monitor api`)

토스 웹 API는 예고 없이 변경됩니다. 과거 두 차례 user-facing 회귀가 있었습니다:

- [#15 / #17](https://github.com/JungHoonGhae/tossinvest-cli/issues/15) — User-Agent 핑거프린팅 차단 (v0.3.6 fix)
- [#29](https://github.com/JungHoonGhae/tossinvest-cli/issues/29) — `/sections/all` body 계약 변경 (v0.4.8 fix)

`monitor api` 명령은 6개 read-only endpoint 를 schema-invariant probe 로 호출해 이런 변경을 사용자보다 먼저 감지합니다.

### 동작 흐름

```
[당신 머신: tossctl monitor api]
       ↓ (당신 세션 쿠키)
[토스 서버] ← 본인 계좌 조회
       ↓ (응답)
[당신 머신: 응답 schema 체크]
       ↓ (실패 시만)
[당신이 설정한 Discord webhook] → 당신 채널
```

`monitor api` 는 본인 머신에서만 실행되며, 본인 세션으로 본인 계좌만 조회합니다. webhook URL 은 코드에 기본값이 없어 사용자가 직접 설정합니다.

### Probe 목록

| 이름 | Endpoint | 보호하는 명령 |
| --- | --- | --- |
| `account-list` | `GET /api/v1/account/list` | `account list` |
| `account-summary-overview` | `GET /api/v3/my-assets/summaries/markets/all/overview` | `account summary` |
| `portfolio-positions` | `POST /api/v2/dashboard/asset/sections/all` (`SORTED_OVERVIEW`) | `portfolio positions` |
| `watchlist` | `POST /api/v2/dashboard/asset/sections/all` (`WATCHLIST`) | `watchlist list` |
| `quote-stock-infos` | `GET /api/v2/stock-infos/A005930` | `quote get` |
| `pending-orders` | `GET /api/v1/trading/orders/histories/all/pending` | `orders list` |

각 probe 는 status 200 + 핵심 JSON 경로 존재 + 타입을 검사합니다. Toss 가 새 필드를 추가하는 변경은 통과시키고, 핵심 필드가 사라지거나 빈 응답을 받으면 실패합니다.

### Cron + 알림 합성

`monitor api` 는 exit 0/1 만 반환합니다. 알림 채널은 cron 라인의 `||` 우항에서 사용자가 자유롭게 합성합니다. `crontab -e`:

```cron
# 매시간 정각, 실패 시 Discord 알림
0 * * * * /usr/local/bin/tossctl monitor api --quiet || \
  curl -sS -X POST -H 'Content-Type: application/json' \
    -d '{"content":"⚠️ tossctl regression"}' \
    'https://discord.com/api/webhooks/...'
```

Discord 외 Slack · ntfy · macOS notification · 이메일 등 다른 채널 합성 예시는 [`AGENTS.md`](../AGENTS.md). launchd · systemd timer 등 다른 스케줄러도 동일하게 동작합니다 (exit code 기반).

### 출력 예시

정상 (모든 probe 통과):

```
  ✓ account-list — status=200 (43ms)
  ✓ account-summary-overview — status=200 (53ms)
  ✓ portfolio-positions — status=200 (52ms)
  ✓ watchlist — status=200 (16ms)
  ✓ quote-stock-infos — status=200 (44ms)
  ✓ pending-orders — status=200 (19ms)

6 passed, 0 failed
```

실패 (예: #29 같은 body-contract 회귀):

```
  ✓ account-list — status=200 (43ms)
  ✓ account-summary-overview — status=200 (53ms)
  ✓ watchlist — status=200 (15ms)
  ✓ quote-stock-infos — status=200 (44ms)
  ✓ pending-orders — status=200 (19ms)
  ✗ portfolio-positions — status=200: result.sections is empty — likely body-contract regression (#29-class)

5 passed, 1 failed
```

webhook 페이로드:

```
🚨 tossctl API regression detected (0.4.9)
2026-05-13 10:00 UTC — 1/6 probes failed

❌ portfolio-positions — POST wts-cert-api.tossinvest.com/api/v2/dashboard/asset/sections/all
    status=200, result.sections is empty — likely body-contract regression (#29-class)
```

### 새 probe 추가

새 read-only endpoint 의존이 생기면 `internal/monitor/probes.go` 의 `Probes()` 반환 슬라이스에 항목 추가:

```go
{
    Name:   "new-endpoint",
    Method: "POST",
    URL:    cert + "/api/v2/...",
    Body:   `{"types":["..."]}`,
    Check: func(status int, body []byte) error {
        if err := expectStatus(status, body, 200); err != nil {
            return err
        }
        return expectPath(body, "result.someKey", "array")
    },
},
```

새 Check 함수는 schema 진단 메시지만 반환하면 됩니다 — `expectStatus` / `expectPath` 가 기본 패턴.
## 서명된 공식 릴리스 시스템 업데이트

`tossctl console` → 설정 → 시스템 업데이트에서 **서명된 최신 릴리스
확인·다운로드**를 누른다. 요청 폼의 URL·repository·tag·대상 경로 값은 선택자로
사용하지 않는다. 바이너리에 고정된 공식 repository, 현재 GOOS/GOARCH asset,
GitHub Actions workflow와 Sigstore public-good 신뢰 루트만 사용한다.

검증 조건:

- GitHub의 canonical latest stable tag이며 현재 release semver보다 새 버전이다.
  `dev` 빌드는 최초 release bootstrap으로만 예외 처리하고 화면에 표시한다.
- archive SHA-256과 SLSA v1 subject digest가 일치한다.
- Fulcio chain/SCT, Rekor inclusion/checkpoint와 integrated timestamp가 검증된다.
- 인증서 issuer와 SAN이 공식 `release.yml@refs/tags/<tag>`와 정확히 일치한다.
- provenance의 repository, workflow path, tag ref, source commit과 다운로드한
  Go binary의 VCS revision이 일치하고 modified build가 아니다.

네트워크는 `api.github.com`, GitHub의 고정 release asset host,
`tmaproduction.blob.core.windows.net/attestations/` 및 Sigstore public-good
TUF 서비스가 필요하다. 최초 실행에서는 TUF metadata 갱신 때문에 시간이 더 걸릴
수 있다. 응답·redirect·pagination·bundle 수·archive 확장 크기는 모두 제한되며
검증 실패 시 현재 실행 파일과 기존 candidate를 유지한다.

성공해도 설치는 일어나지 않는다. 설정 화면에 tag, asset, archive SHA-256,
signer workflow, candidate SHA-256이 표시되고, 사람이 별도 **설치 및 재기동**을
눌러야 한다. 콘솔을 재시작하면 candidate 자체는 검사할 수 있지만 process-local
서명 영수증은 사라지므로 출처를 “확인 안 됨”으로 표시한다.

기존 candidate 보존 후 새 candidate의 directory sync가 실패하면 자동 복구를
시도한다. 복구까지 실패한 경우 오류에 표시되는 recovery 경로의 파일을 보존하고
운영자가 수동으로 비교·복구한다. 다운로드는 엔진 정지·재기동, gate 변경,
계좌 접근이나 주문을 수행하지 않는다.

시스템 업데이트 메뉴가 없는 구버전은 이 기능을 호출할 수 없다. 메뉴가 포함된
바이너리를 최초 한 번 설치한 뒤에는 공식 릴리스마다 수동 `install`을 반복할
필요가 없다. `tossctl update`는 checksum-only 레거시 경로이므로 자동화하지
않고, 확인만 필요하면 `tossctl update --check`를 사용한다.

## 로컬 개발 빌드 시스템 업데이트

임시 Claude/Codex 작업 디렉터리의 바이너리를 `install` 명령으로 직접 덮어쓰지
않는다. 저장소 gate가 끝난 뒤 다음 target으로 현재 설치 경로의 고정 sibling
candidate를 만든다.

```bash
make stage-local-update
```

`tossctl console` → 설정 → 시스템 업데이트에서 candidate SHA-256과 build
metadata를 확인하고 설치한다. 설치는 다음 조건을 모두 만족해야 한다.

- 실계좌 검증이 진행 중이지 않다.
- 실제 engine journal flock을 획득할 수 있다.
- 실행 중인 콘솔이 시작할 때 본 현재 바이너리와 설치 직전 바이너리가 같다.
- 화면에서 검토한 SHA-256과 준비된 candidate 바이트가 같다.

성공하면 현재 경로는 원자적으로 교체되고 직전 바이너리는
`<tossctl>.rollback`에 남는다. 후보 경로는 HTTP 요청으로 선택할 수 없고 항상
`<tossctl>.candidate`이다. target이나 콘솔은 gate를 켜거나 주문을 실행하지
않는다.
