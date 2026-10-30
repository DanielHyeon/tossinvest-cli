# a166 설계 3판 초안 — 엄격한 계좌 공유 상한 (NOT READY)

> **2026-10-31 권한 갱신:** 사용자는 남은 개발 진행을 승인했다(`issues.md` 최신 결정). 아래 3판의 「구현 승인이 아니다」는 당시 상태이며 현재 진행 권한을 취소하지 않는다. 기술적 READY 차단은 유지한다. Manager가 계약을 확정하고 Terra가 구현·시험한다. 운영 활성화는 여전히 별도다.

> **2판 (2026-10-29, 작성 Claude Opus).** 1판(`eff9608e6`)에 대한 독립 적대 리뷰(`review.md`, `713585453`, FAIL — P0 4·P1 4·P2 2)를 Manager 결정(2026-10-29)대로 처분했다.
> 처분 대조표는 §11. 기준 코드 트리는 HEAD `5e2facaee`(a165 착지 뒤 — 이 브랜치는 그 위에 문서 커밋만 있음). 코드·시험 변경 0.
>
> 분기·early return 을 근거로 쓰는 주장은 `analysis/evidence/ast/*.json`(목록 `analysis/evidence/README.md`)의 분기 ID 로 적는다. 「손 읽기」는 AST 밖의 사실(호출 자리·설정)이다.
> CodeGraph 질의는 이 worktree 인덱스가 초기화되지 않아 **같은 코드 트리(`5e2facaee`)인 주 체크아웃 인덱스**로 했다(`analysis/code-context/codegraph-*-{1st,2nd}.txt`, 1.6.0).


## 3판 결정과 차단 조건 (2026-10-30)

사용자 결정: **계좌 전체 3건/초 우선, 보호 요청도 대기 허용**. 이는 a166 설계 방향에 한정하며 지연 수치·장애 결과·구현·배포·활성화 승인과 다르다. 사용자는 a166 재설계에 한해 기존 즉시성 원칙보다 이 보호 대기 방향을 우선하기로 승인했다. 다른 change나 전역 지침은 바꾸지 않는다. a087 production admission은 nil을 유지하고 a169는 보류한다.

- 모든 참여 프로세스의 실제 공식 API wire 시도는 어떤 1초 창에서도 합계 3건 이하여야 한다. 보호·재전송·공유 파일 실패·시계 이상에 대한 우회는 없다. 우선순위는 순서이며 상한 면제가 아니다. 브로커 도착 시각의 보장은 아니다.
- 보호 판단 `/prices`, 보호 주문 접수 상태 확인, 보호 경로 토큰 교환을 critical 의존 사슬에 포함하는 것을 **권장**한다. 분류와 취소·timeout·429·실패 시 처분은 별도 결정이며 승인된 사실이 아니다. normal이 보호 대기자보다 먼저 새 자리를 얻지 않도록 중재해야 하나 flock 공정성이나 try-lock만으로는 증명되지 않는다. 대기자 등록·원자적 관찰·사망 cleanup과 기아 정책이 미결이다.
- 의도된 OFF 경로는 upstream을 보존한다. 상한 적용이 필요한 배포에서 예산 누락·무효는 전 등급 새 전송 거부다. 설정 기본값을 조용히 바꾸지 않으며 배포·활성화는 사람의 명시적 승인이 필요하다.
- 공유 파일 열기·쓰기·읽기·동기화 실패, 손상, 잠금 실패, 파일 교체/삭제, 다른 mount alias 또는 동일 inode 입증 실패는 전 등급 fail-closed다. 독립 로컬 fallback은 없다. 프로세스 crash·재시작·host reboot 및 부분 기록 복구에서 기존 전송과 중복 예산이 없음을 증명해야 하며, 안전하게 증명하지 못하면 전송을 막는다.
- 후보만: 모든 참여자가 안정적인 동일 inode flock을 획득한 뒤 매번 단조 상대시간 350ms 이상 기다리고 `base.RoundTrip` 반환 및 공유 429 기록까지 잠금을 유지한다. 자동 unlock 후 다음 소유자도 매번 기다리는 조건에서 **외부 RoundTrip 호출 시작** 간격을 보수적으로 제한하는 방안이다. 단일 host, fd 상속 방지, inode 교체 방지 및 모든 참여자의 준수가 전제이며 350ms는 승인된 수치가 아니다. 이 후보는 보호 우선 중재나 실제 wire 상한의 해결책이 아니다.
- `RoundTrip`의 응답 반환과 body EOF/Close는 별개다. 잠금 및 drain의 생명주기, 지연된 body와 취소·비정상 종료 처분은 미결이다. HTTP transport 내부 재전송은 외부 `RoundTrip` 한 번에 여러 wire 시도를 만들 수 있다. Go `net/http/transport.go:688-724,815-861`, `h2_bundle.go:7904,7965-8001`의 내부 retry 및 현행 `internal/official/client.go:125` HTTP/2 설정은 별도 조사 근거다(손으로 읽음; AST 증명 아님). 모든 실제 시도를 중재하는 전송 경계와 시험 가능한 증거가 없으므로 **wire 3/s 미증명**. 401 재전송과 token exchange도 빠짐없이 포함해야 한다.
- 추가 조사 후보(미채택): Go 1.26.5 표준 `http.Transport`에서 keep-alive 비활성 및 HTTP/1만 명시, 대체 프로토콜·커스텀 재시도 dialer 금지와 직접 연결을 전제로 내부 retry를 배제할 가능성. `net/http/transport.go:188-193,815-845,1063-1065,1175-1177` 손 읽기 근거이며 전체 Go 버전 또는 proxy에 대한 증명이 아니다. 현행 HTTP/2 설정 변경의 영향, redirect 각 hop·401·token 재진입, 인증 transport 봉인, OFF 동작 및 실제 HTTP 요청 계수 fixture가 선결이다. 표준 라이브러리 후보이므로 신규 의존성·라이선스 추가는 없으나 호환성과 버전별 재검증이 필요하다. 채택/확장은 미결; 외부 패키지·레지스트리는 조사하지 않았다.
- 추가 wire 차단: Go 1.26.5 `transport.go:2874-2883,2897-2907,2935-2940` 손 읽기상 요청 writer와 응답 처리가 동시이며 반환이 writer 종료를 항상 기다린다고 증명되지 않았다. 앞 요청 반환 후 대기만으로 첫 실제 write 시점을 보장할 수 없다. 조기 응답·취소·느린 body fixture 및 write 경계 동기화의 별도 증명이 필요하다.
- 실제 보호 요청의 유한 wall-time 상한은 없다. 큐 경합, 선점, 진행 중 전송, 토큰 잠금, 취소 및 네트워크가 무한히 지연될 수 있다. 측정치·간격만으로 안전 상한을 주장하지 않는다.
- **선행 차단**: wire 경계와 내부 재시도 통제 증명, priority 중재·crash 정합성, 파일 식별/복구, body·drain, 보호 사슬/실패 정책 및 지연 수용 기준의 사람 결정, 독립 적대 재리뷰. 완료 전 구현·activation 금지.

**2판 처분:** `analysis/review-v2-20261030.md` FAIL 유지. 아래 2판은 추적용 역사 기록이며 D3 ε 증명, A/B 선택, 공유 실패 local fallback, clock 초과 예외, 수치 지연 상한, 관련 시험·구현 지시는 무효다. 3판과 충돌하면 아래를 사용하지 않는다.

## 4판 연구 후보 — 요청 쓰기 경계 (2026-10-31, 미검증·NOT READY)

목표 측정점은 **클라이언트가 TLS에 HTTP 요청 plaintext를 처음 넘기는 시각**이다. 커널/NIC에 이미 들어간 바이트나 브로커 도착 시각은 제어 대상이 아니며, 이 정의가 계좌 한도에 충분한지 별도 판단이 필요하다. 이 후보는 구현 명령이나 운영 승인도 아니다.

1. 강제 적용 모드에서 직접 HTTPS, 요청 전용 fresh HTTP/1 연결, keep-alive·HTTP/2·자동 재시도·proxy를 금지한다. TLS handshake 후 반환한 연결을 감싸고 첫 요청 데이터의 `Write` 진입을 시작으로 센다. 인증·서버명 검증·ALPN·timeout을 기존과 동등하게 유지해야 한다.
2. 동일 계좌의 안정적인 동일 inode flock을 얻은 각 소유자는 단조 상대시간 350ms 이상 대기한 다음 전송을 허용한다(350ms는 설계 후보). 잠금은 응답 반환만으로 풀지 않고 요청 writer revoke 완료까지 유지한다. revoke와 모든 `Write`는 같은 mutex로 직렬화하여 진행 중인 쓰기가 끝날 때까지 기다리고 이후 쓰기를 거부한다. 취소 뒤 늦게 도착하는 dial 결과도 revoke를 확인해 닫는다. 비정상 종료 후 다음 소유자도 무조건 대기한다.
3. 하나의 outer 호출에서 실제 HTTP 요청이 최대 하나라는 성질, write와 revoke 사이의 순서 및 crash 이후 중복이 없다는 성질은 코드와 테스트로 증명해야 한다. 보류 중인 write가 끝나지 않으면 잠금과 보호 요청도 무기한 기다릴 수 있다. 복구·상태 파일 무결성 실패는 전 등급 fail-closed다. 우선순위는 별도 공유 대기자 중재가 필요하며 flock의 공정성에 기대지 않는다.
4. Go 1.26.5 `net/http/request.go:584-589`, `transport.go:2654-2672` 손 읽기상 `WroteRequest`만으로 실제 네트워크 flush 완료를 증명할 수 없다. `transport.go:1795`의 `*tls.Conn` 판별 때문에 연결 래핑이 TLS 응답 상태와 인증 권한에 미칠 영향도 미해결이다. 기존 authority 봉인과 OFF 동작의 동등성 없이는 채택하지 않는다.

검증 설계: 로컬 fixture에서 느린 request body, 조기 응답, 취소 중 write, 늦은 dial 완료, 401과 token 경로, redirect, HTTP/2 협상 차단, 프로세스 종료와 경쟁 프로세스, 동일 inode 및 mount alias를 각각 계수·반례 주입한다. 실제 클라이언트의 TLS 검증·authority를 검사한다. Go 버전별 구현 차이와 테스트 한계를 기록한다. 테스트 성공은 서버 도착률이나 무제한 일시정지에 대한 보장이 아니다. 우선 중재, 공유 패널티 원자성, 실행 가능성 및 모든 전송 경로의 배선에 대한 독립 적대 검토가 선행된다.

## 이전 2판 기록 — 계약 아님

## 0. 범위

- **안**: 같은 계좌로 공식 Open API 를 치는 모든 TossOS 프로세스의 wire 요청이 하나의 공유 예산을 거친다(사용자 결정 2026-10-22 결정 3 「3 req/s 계좌 전체 공유」). 엔진 폴링 GET 429 처리를 확정 표대로 바꾼다.
  모든 wire 429 가 비밀 없는 정규 이벤트와 공유 패널티가 된다. 정본 retry matrix §0.3·§1·§2·§3·§4 를 코드와 같은 커밋에서 고친다.
- **밖**: WTS(`internal/client/*`, `internal/monitor/probes.go`) — 공식 API 가 아님(같은 한도인지 [미측정]). 계좌 둘 이상. a169 의 감시기·marker 검사 구현(계약만 준다). 주문 접수 재전송(a171).
- **수치의 출처**: 사용자 결정은 「3 req/s」뿐이다. 간격 **350ms 는 유도값**(1000/3 ≈ 333ms 에 여유)이고 확정 표에 없다 — proposal 의 「확정 표의 값」 인용은 오기였다(2판에서 정정, `grep -rn 350` 결과 a166 문서 밖 0건).

## 1. 소비자 계약

| # | 요구 | 출처 | 2판의 답 |
|---|---|---|---|
| C1 | wire 요청마다 admission(401 재전송 포함), `send` 단위 금지 | a087 `issues.md` D-2 ⑦, design D15.4 조건 1 | D2 — `doRequest` 의 `c.hc` 가 쓰는 transport 안. `send` 의 세 `doRequest`(:329·:354)가 각각 `c.hc.Do` → `RoundTrip` 을 탄다 |
| C2 | a087 의 `*official.Client` 와 같은 admission 인스턴스, dummy 불가 | D-2 ⑦ | D2.3 |
| C3 | 실제 `official.Client` 통합 시험: 401→401→200 admission 3회 · 둘째 거절이면 셋째 wire 0 · 첫 거절이면 wire 0 | D-2 ⑦, D15.4 조건 2 | §9 T-INT-1~3 |
| C4 | 헤더 없는 429 도 admission 거절로 — 수명·재시작 | D-2 ⑦, D15.1 | D5 |
| C5 | 호스트 측정 도구(별도 앱 키·같은 계좌)도 같은 예산 | a169 review §10·§11, issues §6 | D1 |
| C6 | 모든 wire 429 에 정규 이벤트 | a169 §10 P1-9 | D6 |
| C7 | admission 거절과 호출 시작이 같은 동기화 경계, 닫은 뒤 새 호출 0, 이미 admission 한 수 상한 | a169 §10 P1-8 | D7 |
| C8 | 보호 청산 경로가 보조 GET 에 굶지 않음(불변식 4) | Manager, WORKFLOW §0.3 | D4 · §6 |

## 2. 전송 자리 지도 (코드 트리 `5e2facaee`)

근거: 비시험 `grep -rn`(`.Do(`·`RoundTrip(`·`official.New(`·`openapi.tossinvest.com`) + CodeGraph `callers`(1st·2nd) + `affected --filter '*_test.go'`(1st).

### 2.1 공식 API 로 바이트를 내보내는 자리 — 전수와 2판 처분

| # | 자리 | 무엇 | 2판 처분 |
|---|---|---|---|
| S1 | `internal/official/client.go:193` `doRequest` `c.hc.Do` | 모든 `Client` 동사 + `send` 401 재전송 | `c.hc` 의 transport = 예산 transport(D2.1) |
| S2 | `internal/official/token.go:144` `exchange` `m.hc.Do` | `POST /oauth2/token`, `m.hc` 는 `New`(:115)가 넘긴 **같은 `c.hc`** | S1 과 같은 transport — 자동 |
| S3 | `internal/official/a112_mbus_read.go:153` `a112MBUSReadAt` | 측정 GET. `client.authorityTransport.Clone()`(:121, 손 읽기)으로 **새 `http.Client`** 를 만든다 | 그 clone 을 같은 Budget 으로 감싼다(D2.4). CodeGraph: 생산 호출자는 `tools/a112-mb-us-source/official_reader.go:20·25·30` 뿐이지만 그 도구는 a063 이 참조하고 WORKFLOW 의 a112 e2e 시험 대상이라 **은퇴하지 않는다** |
| S4 | `tools/a087-market-sell-probe/send.go:85` `sendOnce` `client.Do`, client 는 `newSendClient`(:54–65, 분기 0) | 사람 승인 1회 POST | `newSendClient` 의 `Transport` 를 예산 transport 로 감싼다. CodeGraph: `sendOnce` 호출자 `main.go:183 execute` 하나지만 a087·a100·verify-execution-capability 가 이 도구를 참조 — **은퇴하지 않는다** |
| S5 | `internal/execgw/replay_transport.go:95` `HTTPReplay.ReplayPlace` `h.HTTP.Do`, `h.HTTP` 는 `buildGateway`(`gateway.go:323-327`)의 `&http.Client{Timeout: replayTimeout}`(기본 transport) | IN_DOUBT 재생 — `:331 Attested: nil` 로 휴면 | 휴면이어도 **같은 transport 로 만든다**(D2.5) — 「래퍼 없는 자리 0」 구조 시험의 예외를 두지 않기 위해 |
| — | `internal/monitor/official_read_contract.go:15` | ContractOnly(보내지 않음) | 구조 시험 허용 목록(사유: 보내지 않음) |
| — | `internal/strategyevidence` `Transport.Do` | 생산 구현 0, `NewOfficialAdapter` 생산 호출자 0 | 대상 아님 |
| — | WTS 전송들 | 공식 API 아님 | 구조 시험 허용 목록(사유: 다른 호스트) |

a165 의 flatten 취소 409 재전송(`cancel_refusal.go:44` `withCancelRefusal` → `:90` `dispatch` 재호출)은 자기 HTTP 가 없고 S1 을 지난다.

### 2.2 프로세스 지도 — 1판 §2.2 그대로(콘솔·엔진·soak 은 컨테이너 `tossos`, httpapi 는 컨테이너 `httpapi`, CLI·도구는 호스트; 각자 `official.New`)

| 프로세스 | `official.New` 자리 |
|---|---|
| 콘솔 · httpapi | `cmd/tossctl/instrument_names.go:125`(`console.go:794`, `httpapi.go:647` `newConsoleBroker`), `candidatepanel.go:67` |
| 엔진 | `internal/app/engine/engine.go:390` `NewOrderPath`(Client 하나를 모든 루프가 공유) |
| soak · verify | `cmd/tossctl/soak.go:727`, `verify.go:889·906`, `verify_reconcile.go:183·204` |
| CLI | `internal/app/app.go:132`, `cmd/tossctl/flatten.go:200`(→ `NewOrderPath`), `engine_reconcile.go:254`, `mcp.go:68`, `openapi.go:44` |
| 도구 | `tools/a087-market-sell-probe/main.go:95`, `a112-l1c-quote-probe/main.go:60`, `a112-mb-us-source/official_reader.go:16`, `a128-d2-receipts/main.go:99`, `a128-jev-shadow-probe/main.go:127`, `a130-outcome-backfill/main.go:89`, `a140-breakout-signal-shadow/main.go:133` |

### 2.3 파일시스템 — 1판 §2.3 그대로

호스트 `/home/daniel/.config/tossctl` ext4(측정), `/mnt/D` fuseblk(측정), 컨테이너 `/tmp`·`/run/tossos` 는 컨테이너별 tmpfs. 공유 가능한 곳은 bind mount `TOSSOS_CONFIG_DIR` → `/var/lib/tossos/config` 뿐이다.
`TOSSOS_CONFIG_DIR` 의 실제 값과 컨테이너↔호스트 flock 교차 배제는 **[측정 필요 — H10 사람 gate]**.

## 3. 지금 코드의 429 처리 (AST 근거)

1. `Client.send` B4@344 `for attempt<2 && code==401` 이 `doRequest`(:354)를 최대 2번 더 부른다 → `send` 한 번에 wire ≤ 3. `send` 단위 admission 이면 두 재전송이 우회한다.
2. `Retrier.Query` B10@370 `ClassRateLimited` → `rateLimitWait` → B11@372·B12@375 를 지나면 루프를 계속 돌아 **다시 보낸다**. B5@351 `ClassOK` 는 `RecordSuccess` 만 — 재시도 성공 경로에 429 흔적이 없다.
3. 생산 배선 `gateway.go:340` `newRetrier(…, nil, …)` → `RetryAfter` nil → `rateLimitWait` B1@431 거짓 → :439 `BackoffFor` = **명목 400ms·800ms(지터 ±25% 별도)**. `NewRetryAfterTransport` 생산 호출자 0.
   그래서 생산 동작은 「429 뒤 같은 조회 안 재전송 최대 2번, 명목 backoff 합 1.2s(지터 별도)」이고, 정본 §1 「`min(RA,30s)` 1회 대기」·§4 「RetryAfterTransport」는 생산에서 참이 아니다.
4. `ClassifyQueryError` 는 모르는 오류를 :97 transient 로 둔다 → 새 거절 sentinel 은 갈래를 따로 줘야 한다(D8).
5. `classifyMutation` B8@84 → `statusOf` B4@134: `ErrRateLimited` 는 429 → 모호(IN_DOUBT). 새 sentinel 은 B6@140 default → `ClassifyHTTPMutation(SendNotStarted, 0, err)` → NOT_SENT(리뷰가 AST 로 재확인).
6. `ShouldFallback` :71-75: `ErrRateLimited` 는 WTS fallback, 새 sentinel 은 아님(H8 수락).

## 4. 선택지 — ① 유지 (1판 §4 의 비교 그대로)

공유 상태 파일 + flock. ② 프록시는 flatten·soak 를 엔진에 묶고, ③④는 합을 보장하지 못한다. 2판은 ①의 **경계 자리**를 `doRequest` 에서 **공식 API 전용 `http.RoundTripper`** 로 옮긴다(Manager 결정, P0-1).

## 5. 설계

### D1. 예산 디렉터리 — 명시 주입과 같은 파일 계약 (P0-2 · H1)

- **주입되지 않으면 OFF**: Budget nil → transport 를 감싸지 않음 → upstream 과 같은 바이트(불변식 2). 개발과 upstream CLI 기본이 이 경우다. 자동 기본 경로는 **없다**(1판의 `<UserConfigDir>/…` 폐기).
- **주입 방법**: 엔진·콘솔·soak·httpapi 는 `compose.yaml` env `TOSSOS_ACCOUNT_BUDGET_DIR=/var/lib/tossos/config/account-budget`(두 서비스). 콘솔이 띄우는 `engine run`·`soak run` 은 env 를 상속한다.
  호스트 CLI·도구는 루트 플래그 `--account-budget-dir` 또는 같은 env. 운영 배포는 반드시 주입한다 — 배포 체크(task 5.1)에 「두 서비스 env 존재 + 엔진 기동 로그의 instance」를 넣는다.
- **instance 파일**: 디렉터리에 `instance`(무작위 128bit hex, `O_CREAT|O_EXCL` 로 한 번만 만듦 — 경합 시 진 쪽은 읽음)와 상태 파일 `budget`. 각 프로세스는 Budget 을 연 직후
  `account_budget_open instance=<id> dev=<st_dev> ino=<st_ino>`(budget 파일) 한 줄을 남긴다. 엔진은 같은 id 를 marker `engine-run.json` 의 새 키 `account_budget_instance` 에 싣는다.
- **같은 inode 확인**: 컨테이너와 호스트가 같은 id·inode 를 보는지는 **H10 사람 gate**(배포 전, 실제 bind path, 브로커 호출 0)로 확인한다. 자동 검사는 instance 비교다.
- **측정 도구(a169 등)의 시작 거부**: 예산 디렉터리가 주입되지 않았거나, 열 수 없거나, 그 instance 가 fresh 엔진 marker 의 `account_budget_instance` 와 다르면 **GET 0 으로 시작을 거부**한다(검사 구현은 도구 쪽, 이 change 는 값과 계약을 준다).
- **한 디렉터리 = 한 계좌**: 계좌 식별(번호·seq)을 파일에 쓰지 않는다(개인정보, 안전 부트스트랩 8). 계좌가 하나인 것은 현재 실측 사실(M14·M35).

### D2. 단일 경계 — 공식 API 전용 RoundTripper (P0-1 · H6)

1. **`accountbudget.Transport{base http.RoundTripper, budget *Budget}`**(새 패키지, 구체 타입). `official.New` 는 봉인 구간(:114–117) 안에서 Budget 이 주입됐으면 `c.hc` 를
   `&http.Client{Timeout: hc.Timeout, Transport: budget.Wrap(hc.Transport), CheckRedirect: hc.CheckRedirect, Jar: hc.Jar}` 로 바꾼 **뒤** `newTokenManager(…, c.hc)` 를 부른다 → S1·S2 가 같은 transport.
   `WithHTTPClient`(시험) 로 들어온 client 도 같이 감싼다.
2. **authority 판정**: `authorityOriginLocked`(분기 0)·`hcTransport`(B1@164)는 `c.hc.Transport.(*http.Transport) == c.authorityTransport` 를 본다. 감싸면 이 단언이 깨져 `AuthorityOrigin`·`AuthoritativeExchangeRate`·`a112MBUSReadAt` 이 전부 거절된다.
   → `hcTransport` 가 **정확히 한 겹의 `*accountbudget.Transport` 이고 그 budget 이 `c.budget` 과 같을 때만** 벗겨 base 를 돌려준다. 그 밖의 래퍼는 지금처럼 authority 없음.
3. **같은 인스턴스(C2)**: Option `official.WithAccountBudget(*accountbudget.Budget)`(봉인 전), 접근자 `(*Client).AccountBudget()`. 구체 타입·비공개 필드라 dummy 를 생산 배선에 넣을 길이 없다. 엔진은 `NewOrderPath` 에서 Budget 하나를 열어 넣고,
   a087 5.5 는 갱신기의 Client 의 `AccountBudget()` 가 non-nil 이고 엔진이 연 그 포인터임을 확인한다. a087 은 자기 preflight 요청을 만들지 않는다.
4. **S3**: `a112MBUSReadAt` 의 clone(:121)을 `client.budget.Wrap(clone)` 으로(Budget nil 이면 그대로 — OFF). **S4**: `newSendClient` 가 Budget 을 받아 `Transport: budget.Wrap(&http.Transport{…})`. probe `main.go` 는 `--account-budget-dir` 없으면 **시작 거부**(사람 승인 POST 는 OFF 로 돌지 않는다).
5. **S5**: `buildGateway` 의 재생 client 를 `&http.Client{Timeout: replayTimeout, Transport: in.budget.Wrap(nil)}` 로(`Wrap(nil)` = 기본 transport 감쌈).
6. **구조 시험 T-STR-1(Manager 지시)**: `go/packages`+`go/types` 로 모듈 전체(시험 제외)의 `net/http.Client` 복합 리터럴·`new(http.Client)`·`http.DefaultClient` 사용·`(*http.Client).Do` 호출 자리를 전수 수집한다. 각 자리를
   (a) `Transport` 필드가 `(*accountbudget.Budget).Wrap(...)` 의 결과, (b) `official.New` 안에서 감싸진 `c.hc`, (c) 허용 목록(사유 필수: WTS 호스트 · localhost RPC · ntfy · release update · ContractOnly) 중 하나로 분류하고, **분류되지 않은 자리 1개 이상이면 실패**한다.
   양성 대조군: S4 의 `Wrap` 을 지운 사본에서 시험이 실패해야 한다(변이 원장). 허용 목록 항목 수와 수집 자리 수를 둘 다 단언(표본 0 통과 금지).

### D3. wire 시점 계약 — 측정점은 `RoundTrip` 안의 `base.RoundTrip` 호출 (P0-4)

- `Transport.RoundTrip(req)` 순서: ① 등급 판정(D4) ② flock 아래 판정·예약 — 거절이면 wire 0 으로 `ErrAdmissionRefused` ③ 잠금 풀기 ④ 슬롯 시각 `s` 까지 대기(ctx 상한) ⑤ 프로세스 안 mutex 아래 `closed` 재확인·
   `now ≤ s + ε` 확인 — 넘었으면 **다시 ②**(재admission) — 통과면 in-flight +1 ⑥ **즉시 `base.RoundTrip(req)`** ⑦ 응답 처리(D5·D6) ⑧ in-flight −1.
- ⑤와 ⑥ 사이에는 사용자 코드·잠금·I/O 가 없다. **계약의 측정점은 ⑥의 호출 시작**이다.
- **증명**: 모든 시작 시각 t 는 [s, s+ε] 안이고 슬롯은 350ms 간격이다. 길이 W < 1s 인 창에 시작 k 개가 들면 슬롯 k 개는 길이 W+ε 안에 있으므로 (k−1)·350ms ≤ W+ε < 1000ms+ε. **ε ≤ 50ms 이면 k ≤ 3**. 2판 값 **ε = 40ms**.
- 남는 것(정직하게): ⑤의 확인과 ⑥의 호출 사이 Go 스케줄러 선점은 막을 수 없다. 그 구간은 분기 없는 연속 두 문장이고, 시험은 ④와 ⑤ 사이 지연 주입으로 재admission 을 고정한다(T-WIRE-1).
  브로커 도착 시각은 네트워크 지연만큼 더 퍼진다 — 브로커 쪽 집계는 [미측정]이다.

### D4. 등급과 우선순위 (P0-3 · H2 — **사람 결정**)

**critical 의 범위(두 안 공통, Manager 결정)**: 보호 청산 매도 · 그 취소 · flatten · 그 경로의 토큰 교환. 진입 주문과 일반 GET 은 normal.

| 요청 | 등급 | 지연·실패 처분 | 근거 |
|---|---|---|---|
| 보호 매도 POST(`isProtective`, `exitloop.go:1463`) | critical | — | Manager 범위 |
| 그 앞의 보호 취소 POST | critical | — | 같음 |
| 취소 precheck `OrderRawByID`(`precheck.go:103`, 2s 상한) | normal **TryOnce**(자리 없으면 즉시 거절) | 거절 → B1@103 `inconclusiveActions` → 취소 진행(오늘의 fail-open) — **지연 0** | AST B1@103 |
| 보호 매도 접수 확인 GET(`roundtrip.go:71` `confirmCreatedOrder`) | critical(보호 매도 경로) — **H11** | normal 이면 패널티 중 거절 → 접수 확인 실패 → IN_DOUBT + 종목 차단 | 손 읽기 |
| exit 관측 `/prices`(`exitloop.go:864`) | **normal**(Manager 범위) — **H11** | 패널티 중 거절 → 그 주기 판단 없음 → price 15s staleness → 다음 주기 | AST 아님(호출 자리) |
| a165 엔진 취소 409(재전송 0) · flatten 취소 409 재전송 | 보호 경로면 critical, flatten 은 critical | — | a165 |
| flatten 명령의 모든 요청(`cmd/tossctl/flatten.go` ctx 뿌리) | critical | — | Manager 범위 |
| 토큰 교환 | ctx 를 따름. normal 이면 대기 상한 `W_tok = 700ms` 뒤 거절 | `exchange` 는 `m.mu` 를 쥔 채 돈다(`token.go:61·109` 손 읽기) — 상한이 없으면 보호 요청이 `m.mu` 뒤에서 굶는다 | 우선순위 역전 방지 |
| a087 하한가 GET | normal | 거절 → 오늘 LIMIT(a087 D15.1) | a087 |
| 진입 POST·그 확인, 그 밖의 모든 GET | normal | 기존 | Manager 범위 |

표지: ctx 값 `accountbudget.WithCritical(ctx)` 를 보호 제안 실행(`isProtective` 참인 dispatch)과 flatten 명령 뿌리에서 단다. 그 ctx 아래 요청은 critical 이고, precheck 는 `accountbudget.WithTryOnce` 로 내린다.
보호 경로 ctx 아래에서 나가는 경로 집합은 T-STR-2(엔진 통합 시험)가 **닫힌 집합으로 고정**한다 — 새 경로가 생기면 실패.

**안 A — Manager 권장**: critical 도 간격 350ms 를 지키되 **최우선**. 구현: critical 은 flock 안에서 바로 `next_slot = max(next_slot, now) + 350ms` 로 **예약**하고 그 슬롯까지 기다린다. normal 은 `next_slot ≤ now` 일 때만 자리를 얻으므로
critical 예약이 남아 있는 동안 normal 은 새 슬롯을 못 받는다. 패널티(D5)는 critical 을 거절하지 않는다(간격만 지킨다). 근거: 하드 3/s 를 넘기면 브로커 429 패널티(최대 30s 로 관측·가정)를 받아 손절이 더 늦어진다.
**안 B**: 1판의 무대기 우회 — critical 은 기다리지 않고 자리만 소비. 계좌 3/s 를 critical 버스트 동안 위반할 수 있다.

| | 안 A | 안 B |
|---|---|---|
| 계좌 3/s | 지킴(D3 증명이 critical 포함 전체에 성립) | critical 버스트 동안 위반 가능 |
| 보호 매도 1건, 경합 없음 | ≤ 350ms + ε (직전 normal 시작이 남긴 슬롯) | 0 |
| 동시 보호 매도 n 건 | 마지막 건 ≤ **n×350ms + ε** | 0 |
| 취소 → 매도 사슬(직렬) | 매도 시작 ≥ 취소 시작 + 350ms(오늘 대비 +max(0, 350ms − 취소 RTT); RTT 최대 169~182ms 관측 R8 → 약 +170~180ms) | 0 |
| flatten m 종목(취소+매도 2m 개 직렬) | 약 2m × 350ms | RTT 만 |
| 브로커 429 위험 | 낮음 | critical 버스트가 패널티를 부를 수 있음 → 그 뒤 매도가 429(POST 429 = IN_DOUBT) |

Manager 지시문의 「≤ (n−1)×350ms + ε」는 직전 normal 슬롯 몫(최대 350ms)을 빼고 센 값이다. 직전 normal 이 방금 시작했으면 첫 critical 도 350ms 를 기다리므로 2판은 **n×350ms + ε** 로 적는다.
그 몫을 없애려면 critical 이 normal 보다 앞에 끼어들 수 있어야 하는데, 그것은 간격 규칙을 깬다(안 B 쪽).

### D5. 429 — 공유 패널티 (C4 · P0-1)

- 자리: `Transport.RoundTrip` ⑦ — `base.RoundTrip` 응답의 status==429 면 헤더 유무와 무관하게 `Record429`. S1~S5 전부 같은 transport 라 우회 자리가 없다(P0-1 처분).
- 길이 P: `Retry-After`(delta-seconds·HTTP-date)가 0 < RA ≤ 30s 면 RA, > 30s 면 30s, 없거나 못 읽으면 **15s**(R3) — H4. `penalty_until = max(penalty_until, now + P)`.
- 효과: 패널티 안 normal 은 즉시 거절(wire 0). 안 A 의 critical 은 간격만. 파일에 있으므로 재시작을 견딘다(재시작은 복구가 아님).

### D6. 정규 이벤트 (C6)

- 프로세스 간: 상태 파일의 `wire_429` 단조 계수·`last_429`. 프로세스 안: `Budget.OnEvent` → `obs.Logger` 한 줄
  `rate_limit event=wire_429|admission_refused|clock_anomaly|budget_reset path=<템플릿> retry_after=<초|absent> penalty_until=<RFC3339> process=<…>`.
- 경로는 질의를 버리고 숫자·ID 세그먼트를 `{id}` 로. 토큰·계좌 헤더·본문 없음. 자리가 transport 하나라 호출자의 재시도 방식(재시도 성공·soak `readRetry`·fallback)과 무관하게 빠짐없다.

### D7. 닫기 — drain 완료 (P1-1)

- `Close() (started int)`: 프로세스 mutex 아래 `closed = true`. 그 뒤 D3 ⑤ 에 도착하는 요청은 wire 없이 거절된다. `Close` 는 **in-flight(⑤ 통과 ~ ⑧ 전) 가 0 이 될 때까지 기다리고**, 기다리기 시작한 순간의 in-flight 수를 돌려준다.
  완료 정의: `Close` 반환 = (a) 이후 새 `base.RoundTrip` 시작 0, (b) 그 전에 시작한 전송 모두 끝남, (c) 반환값 = 닫는 순간 이미 시작한 전송 수(상한 = 그 Budget 인스턴스의 동시 전송 수, 직렬 도구는 ≤ 1).
- seam: D3 ⑤ 와 ⑥ 사이(같은 mutex 구간의 끝). 시험은 ④에서 멈춘 요청·⑥에서 응답을 붙잡은 요청을 각각 두고 Close 를 주입한다(T-SYNC-1).
- `StopOnShared429`(도구용): ② 의 flock 안에서 `wire_429` 가 열 때보다 크면 거절하고 Budget 을 closed 로 바꾼다. 도구는 측정을 재개하기 전에 `Close()`(drain)를 끝내야 한다 — 계약.

### D8. 엔진 폴링 429 — `Retrier.Query` (H9)

- `ClassifyQueryError` 에 `ErrAdmissionRefused` → `ClassRateLimited`(B6@79 앞). B10@370 은 **바로 반환** — 같은 조회 안 재시도 0, staleness 로 넘김.
- 「RA ≤30s 존중」은 D5 패널티가 담당(다음 주기 normal 거절). 「>30s」는 30s 상한 뒤 다시 429 면 다시 30s.
- H9 수락(리뷰): `rateLimitWait`·`Retrier.RetryAfter`·`RetryAfterSource`·`RetryAfterTransport`(+`NewRetryAfterTransport`·`ParseRetryAfter` 는 accountbudget 으로 이동)·`newRetrier` 의 `retryAfter` 인자와 `gateway.go:340` 의 `nil` 인자 삭제.
  삭제 대상 public 이름·시험·정본 §4 행을 FLM/BTM(revision base)과 삭제 RED 에 전부 넣는다(P2-1).
- 비지연 walk(soak·verify `ReadRetryBackoff` 15s→30s, reconcile 15s)는 그대로. 그 재시도도 transport 를 지나 패널티 안이면 wire 없이 거절된다(T-WALK-1).

### D9. 실패와 OFF (P1-2 · H5 · H7 — Manager 결정)

| 상태 | normal | critical | 엔진 |
|---|---|---|---|
| **주입 안 됨** | 제한 없음(upstream) | 제한 없음 | 기동 로그 `account_budget=off` 경고 |
| **주입됐는데 열기 실패**(권한·없는 경로·허용 밖 fs — ext4/xfs/btrfs/tmpfs 만 허용, fuseblk·nfs·overlay 등 거부) | **모두 거절**(wire 0) — local 3/s 로 돌아가지 않음 | 안 A 규칙을 **프로세스 안 간격**으로(공유 없음) | 기동은 함 + 신규 진입 latch `ReasonAccountBudgetUnavailable` + critical 알림. 기동 거부는 손절 없음이라 안 함 |
| flock 지연 | ctx 상한까지 재시도 후 거절 | 50ms 안에 못 잡으면 프로세스 안 간격으로 보냄 + 이벤트 `budget_lock_unavailable` | — |
| 손상(crc·파싱·`v`) | 보수 재설정(`penalty_until = now+15s`, `next_slot = now+350ms`) + 이벤트 `budget_reset` | 같음 | instance 는 바꾸지 않음(instance 는 별도 파일) |

- 열기 실패 시 다른 프로세스들(콘솔·soak·verify·CLI)도 normal 이 전부 거절되므로 「진입 latch 만으로 normal GET 이 계속 돈다」(리뷰 P1-2)는 성립하지 않는다 — 각 프로세스가 **스스로** 거절한다. 공유 파일이 열리는 프로세스와 안 열리는 프로세스가 섞이면, 안 열리는 쪽은 normal 0 이라 합을 늘리지 않는다.
- upstream CLI 는 주입하지 않으면 오늘 그대로다(H7 대체).

### D10. 시계 (P1-4)

- 공유 판정은 **wall clock 하나**(`clock.System().Now()`, 같은 호스트의 컨테이너는 호스트 CLOCK_REALTIME 을 공유)로 한다. 단위 시험만 패키지 안 생성자로 fake Clock 을 넣는다. 실시간은 대기의 ctx 상한으로만.
- 허용 오차 δ = 50ms. 상태 파일의 `next_slot` 이 `now + H + δ`(H = 정당한 최대 예약 지평 = critical 동시 대기 상한 8 × 350ms) 를 넘거나 `last_start > now + δ` 이면 **역행 또는 다른 offset 의 기록자**로 본다:
  normal 은 거절하고 `penalty_until = now + 30s`(이벤트 `clock_anomaly`), critical 은 `next_slot = now + 350ms` 로 다시 세우고 보낸다(안 A 기준) + 프로세스마다 **monotonic** 마지막 시작 시각으로 자기 간격 350ms 는 계속 지킨다.
- 성질을 둘로 적는다:
  - **P-wait(대기 상한)**: 어떤 시계 점프에도 normal 대기·거절 창은 30s 를 넘지 않고, critical 대기는 안 A 표의 상한(+ 이상 감지 1회당 350ms)을 넘지 않는다.
  - **P-rate(3/s)**: 같은 wall clock 을 쓰는 동안 D3 증명이 성립한다. 역행 사건 1회마다 그 1초 창에 **최대 (critical 을 내는 프로세스 수 − 1)건** 초과가 생길 수 있다 — 엔진·flatten 둘이면 +1. 사람 수락 대상(H12).

### D11. 배포 · UID · rollback

- 배포: 이미지 재빌드 + 두 서비스 재시작(KST 05:00~09:00, 사람). `compose.yaml` env. 저널·원장 스키마 변경 없음. 배포 체크에 주입 확인.
- 권한: 디렉터리 0700·파일 0600, 소유 = `TOSSOS_UID`. 다른 uid 는 열기 실패 → D9 둘째 줄.
- rollback: 이전 이미지 → admission 없음(오늘 동작). 남은 `account-budget/` 는 읽는 자 없음.
- 쓰기: `flock` 아래 `pread` → 판정 → `pwrite(0)` + `ftruncate`. rename 금지(flock 은 inode). fsync 없음.

## 6. 불변식 4 표 (손절·청산 즉시성) — 안 A 기준 최악 추가 지연

| # | 상황 | 오늘 대비 추가 지연 상한(안 A) | 안 B |
|---|---|---|---|
| 1 | 보호 매도 1건, 경합 없음 | ≤ 350ms + ε(40ms) | 0 |
| 2 | 동시 보호 매도 n 건 | 마지막 ≤ n × 350ms + ε | 0 |
| 3 | 보호 취소 → 매도 | 매도 시작 +max(0, 350ms − 취소 RTT) | 0 |
| 4 | 취소 precheck | 0(TryOnce, 오늘의 fail-open) | 0 |
| 5 | 그 경로의 토큰 교환이 필요 | +≤ 350ms(critical 슬롯) / normal 교환이 `m.mu` 를 쥐고 있으면 +≤ `W_tok` 700ms | 0 / +≤ 700ms |
| 6 | 브로커 429 패널티 중 | 0 추가(critical 은 패널티 무시, 간격만) — 브로커가 실제로 429 를 주면 그것은 오늘과 같은 결과 | 같음 |
| 7 | 예산 주입됐는데 열기 실패 | critical 은 프로세스 안 간격만(1·2·3 과 같은 식). **그러나 normal 이 전부 거절되므로 exit `/prices` 가 normal(H11)이면 보호 판단 자체가 멈춘다 — 열릴 때까지(재시작 필요) 손절 없음** | 같음 |
| 8 | 시계 이상 감지 | 1회당 +≤ 350ms | 0 |
| 9 | flatten m 종목 | 약 2m × 350ms 직렬 | RTT 만 |
| 10 | exit `/prices`(normal, H11) | 패널티 중이면 그 주기 판단 없음 — 최대 30s(패널티 상한) | 같음 |

7·10 번 칸이 안 A·B 공통으로 가장 큰 칸이다 — 둘 다 exit `/prices` 가 normal 이라서 생기고, H11 이 그것을 정한다.

## 7. 예산 계상 (정본 §3 갱신안)

| 출처 | 등급 | 정상 상태 |
|---|---|---|
| 엔진 미체결·잔고·보유·단건 폴링 | normal | ≈ 0.5 req/s(정본 §3) |
| 엔진 exit `/prices` | normal(H11) | 주기당 1 — 수치 [미측정] |
| a165 엔진 취소 409 지속 종목 | critical(보호 경로)·normal | ≈ 0.4 req/s/종목(a165 D2) |
| flatten | critical | (GET+POST)×최대 2 /취소 |
| soak walk | normal | 26 페이지 ≥ 9.1s |
| 토큰 교환 | ctx 를 따름 | 하루 1회 + 401 갈래 |

안 A 이면 critical 포함 전체가 ≤ 3 req/s(D3·D10 P-rate 조건 아래). 안 B 이면 normal 만.

## 8. 사람 결정

| # | 질문 | 상태 |
|---|---|---|
| H1 | 키 = 명시 주입 예산 디렉터리, 계좌 1:1, instance 파일 | **Manager 결정**(D1) |
| **H2** | critical **안 A**(간격 준수·최우선, §6 표의 지연) vs **안 B**(무대기 우회, 3/s 위반 가능) | **사람 결정** — Manager 권장 A |
| H3 | normal 끼리 프로세스별 몫 없음 | 리뷰 조건부 수락 — D1·D9 로 조건 충족 |
| H4 | 헤더 없는 429 패널티 15s, 상한 30s | 리뷰 조건부 수락 — D5 로 조건(모든 자리 기록) 충족, 값은 사람 확인 |
| H5 · H7 | 실패·OFF 처분 | **Manager 결정**(D9) |
| H6 | S3·S4 같은 transport, 은퇴 안 함(활성 change 가 참조) | **Manager 결정 + CodeGraph 계수**(§2.1) |
| H8 · H9 | sentinel fallback 제외 · RetryAfter 일가 삭제 | 리뷰 수락 |
| H10 | 컨테이너↔호스트 같은 instance·inode·flock 교차 배제, 다른 uid 실패, fuseblk 거부 — 배포 전 사람 gate(브로커 호출 0) | 리뷰 수락 — 승인 요청 |
| **H11 (새)** | exit `/prices`(`exitloop.go:864`)와 보호 매도 접수 확인 GET(`roundtrip.go:71`)을 critical 로 올릴지. Manager 범위는 「일반 GET = normal」. normal 이면 패널티 중 exit 판단이 최대 30s 멈추고(§6 #10), 예산 열기 실패(D9) 동안에는 exit 판단이 계속 멈추며(§6 #7), 접수 확인이 실패하면 보호 매도가 IN_DOUBT + 종목 차단이 된다. critical 이면 주기마다 critical 슬롯 1개(≈0.2 req/s)를 쓴다 | **사람 결정** — 작성자 의견: 둘 다 critical(보호 매도 경로의 앞·뒤) |
| **H12 (새)** | 시계 역행·offset 사건 1회당 1초 창 초과 최대 (critical 프로세스 수 − 1)건 수락(D10 P-rate) | **사람 결정** — 대안은 역행 중 critical 도 기다림(손절 지연이 역행 폭만큼 커짐) |
| **H13 (새)** | 간격 350ms(유도값 — 확정 표에 없음)와 ε 40ms | **사람 확인** — 증명 조건은 3 × 간격 > 1000ms + ε. 333ms 는 3×333 = 999 < 1000 이라 ε = 0 에서도 깨진다. 350ms 는 ε < 50ms 를 허용한다 |

## 9. 시험 (a087·a169 가 인용할 최소 집합 — 이름은 RED 에서 확정)

- **T-INT-1~3**(C3): 실제 `official.Client` + `httptest` + 실제 `*accountbudget.Budget`(fake Clock, 임시 ext4 디렉터리). 401→401→200 에서 transport admission 3회·`base.RoundTrip` 3회.
  핸들러 안에서 첫 요청 처리 중 `Close()` → 둘째 admission 거절 → 둘째·셋째 wire 0(서버 계수 1). 둘째 처리 중 → 셋째 wire 0(계수 2). 시작 전 Close → wire 0. 각각 `errors.Is(err, ErrAdmissionRefused)`, `Close` 반환값.
- **T-429-S**(P0-1): S1·S2(토큰 교환 429)·S3(a112 clone)·S4(probe client)·S5(재생 client) 각각에서 헤더 없는 429 → 공유 패널티 15s, 새 Budget(재시작 흉내)에서 normal 거절·critical 간격 통과, 이벤트 정확히 1.
- **T-WIRE-1**(P0-4): ④~⑤ 사이 지연 주입 seam 으로 한 요청을 멈춘 채 다른 요청들을 진행 → 멈춘 요청은 `s+ε` 초과로 재admission, `base.RoundTrip` 시작 시각 기준 모든 1초 창 ≤ 3.
- **T-MP-1**(다중 프로세스, `-race`): `TestMain` 재실행 자식 4개 + 각자 임의 지연 주입, 서버가 받은 **요청 도착 시각**이 아니라 각 자식이 기록한 `base.RoundTrip` 시작 시각(같은 wall clock)으로 1초 창 ≤ 3. 임시 디렉터리가 허용 fs 가 아니면 실패.
- **T-MP-CLK**(P1-4): 자식 B 에 wall offset ±1s·역행 1h 주입 → normal 거절·`clock_anomaly`, P-wait 30s, P-rate 초과 ≤ 1.
- **T-SYNC-1**(P1-1): ④에서 대기 중 normal 여럿 + ⑥에서 응답을 붙잡은 요청 1 → Close 는 붙잡은 것이 끝날 때까지 반환하지 않고 반환값 1, 반환 뒤 시작 0. `StopOnShared429` 자식 프로세스 판.
- **T-PRIO-A / T-PRIO-B**(H2 결과에 따라 하나): 안 A — normal 대기열이 찬 상태에서 critical 예약 뒤 normal 이 그 슬롯 전에 못 감, n 건 상한 n×350ms+ε. 안 B — critical 대기 0.
- **T-MUT-1~3**(P1-3): ① 주문 POST admission 거절 → `DispatchNotSent` ② a165 엔진·flatten **첫 cancel** admission 거절 → `DispatchNotSent` ③ flatten 409 already-processing 뒤 **재전송** admission 거절 → 첫 `DispatchAmbiguous` 보존(`cancel_refusal.go:92-100`) — 셋 다 `attempt.DispatchVerified` 종단. `ErrAdmissionRefused` 가 `ErrRateLimited` 로 안 감싸짐.
- **T-RET-1**: 429·admission 거절에 `fn` 1회·대기 0. **T-OFF-1**: Budget nil → 요청 수·순서·transport 동일(`AuthorityOrigin` 포함). **T-AUTH-1**: 감싼 Client 의 `AuthorityOrigin`·`a112` 통과, 다른 래퍼는 거절.
- **T-STR-1**(D2.6 전수), **T-STR-2**(보호 ctx 경로 닫힌 집합 + `WithCritical` 자리 + precheck TryOnce), **T-WALK-1**.
- **T-FS-1**(P2-2): `pwrite`/`ftruncate` 사이 kill(자식) → 손상 → 보수 재설정·instance 불변, fuseblk·nfs·overlay 판별 거부, 다른 uid open 실패 → D9 둘째 줄.
- 변이 원장: 350ms·ε·δ 상수, `<`/`≤`, 등급 판정, 패널티 갈래, 재admission 삭제, `closed` 재확인 삭제, drain 대기 삭제, `hcTransport` 벗기기 조건(같은 budget 검사 삭제), `Wrap` 삭제(S1~S5 각각 — 양성 대조군), `Record429` 삭제, sentinel 을 `ErrRateLimited` 로.

## 10. FLM 대상 (gate ⑤ 번들)

편집: `official.New`(B1@108 뒤 감싸기) · `official.Client.hcTransport`(B1@164) · `official.Client.authorityOriginLocked` · `official.a112MBUSReadAt` · `a087probe.newSendClient`·`main`(시작 거부) · `engine.buildGateway`(재생 client·`newRetrier` 인자) ·
`engine.NewOrderPath`(Budget 열기) · `engine.newRetrier` · `execgw.ClassifyQueryError` · `execgw.Retrier.Query` · 보호 dispatch 에 `WithCritical` 을 다는 exitloop 함수 · `engine.officialBroker.availableActions`(TryOnce) · flatten 명령 뿌리 ·
각 Client 생성 함수(Option 추가) · 삭제: `Retrier.rateLimitWait`·`RetryAfterTransport.*`·`NewRetryAfterTransport`(revision base).
**바꾸지 않음**: `Client.doRequest`·`Client.send`·`tokenManager.exchange`(transport 경계로 옮겨 본문 무변경 — 2판의 이득).

## 11. 리뷰 처분 대조 (review.md `713585453`)

| ID | 처분 | 자리 |
|---|---|---|
| P0-1 | 단일 RoundTripper — S1~S5 전부, 429 기록도 그 안 | D2 · D5 · T-429-S · T-STR-1 |
| P0-2 | 명시 주입 + instance 파일 + inode 로그 + 도구 시작 거부 + H10 | D1 · D11 |
| P0-3 | critical 범위 축소(Manager), 안 A/B 사람 결정, 불변식 4 표 수치 | D4 · §6 · H2 |
| P0-4 | 측정점 = `base.RoundTrip` 호출, ε 재admission, 증명 | D3 · T-WIRE-1 · T-MP-1 |
| P1-1 | Close = 시작 상한 + drain | D7 · T-SYNC-1 |
| P1-2 | 주입 없음 OFF / 주입 실패 normal 전부 거절 / 보호 관련 GET 처분 표 | D9 · D4 표 |
| P1-3 | a165 두 경우 + `DispatchVerified` 종단 | T-MUT-2·3 |
| P1-4 | wall clock 하나, δ, P-wait / P-rate 분리 | D10 · T-MP-CLK |
| P2-1 | 「명목 backoff 합, 지터 별도」, 삭제 범위 FLM | §3 3 · D8 |
| P2-2 | partial write·거부 fs·다른 uid 시험 | T-FS-1 |
