# a166 설계 1판 — 계좌 공유 admission(3 req/s)과 엔진 GET 429

> **1판 (2026-10-29, 작성 Claude Opus — Manager 최종 결정 전).** 기준 HEAD `5e2facaee`(a165 착지 뒤). 아래 「권장」은 작성자의 권장이고
> Manager·사람 결정 전에는 고른 것이 아니다. 사람 결정 후보는 §8 에 모았다. 코드·시험은 한 줄도 바뀌지 않았다.
>
> 분기·early return 을 근거로 쓰는 주장은 `analysis/evidence/ast/*.json`(HEAD 에서 뽑은 AST, 목록 `analysis/evidence/README.md`)의 분기 ID 로 적는다.
> 「손 읽기」라고 적은 것은 AST 밖의 사실(호출 자리·설정)이다.

## 0. 범위

- **안**: 같은 계좌로 공식 Open API 를 치는 모든 TossOS 프로세스의 wire 요청이 하나의 예산을 거친다(사용자 결정 2026-10-22 결정 3). 엔진 폴링 GET 의 429 처리를
  확정 표(`verify-execution-capability/analysis/retry-matrix-slo-2026-10-22.md` §2)대로 바꾼다. 모든 wire 429 가 비밀 없는 정규 이벤트로 남는다(a169 P1-9).
  정본 retry matrix §0.3·§1·§2·§3·§4 를 코드와 같은 커밋에서 고친다.
- **밖**: WTS(`internal/client/*`, `internal/monitor/probes.go`)— 다른 서비스이고 결정 3 의 대상이 아니다(WTS 가 같은 한도를 쓰는지는 [미측정]).
  계좌 둘 이상(§D1). a169 의 감시기 · marker 검사 구현(a169 소유 — 이 change 는 계약만 준다). 주문 접수 재전송(a171).

## 1. 소비자가 이미 요구한 계약

| # | 요구 | 출처 | 이 설계의 답 |
|---|---|---|---|
| C1 | admission 은 `official.Client.doRequest` 의 매 `c.hc.Do` **직전**. `send` 단위 금지, 401 재전송도 각각 | a087 `issues.md` D-2 ⑦ 첫 줄, design D15.4 조건 1 | D8 |
| C2 | a087 이 쓰는 `*official.Client` 와 **같은** admission 인스턴스. 별도 preflight 가 토큰을 또 쓰거나 dummy admission 이 활성화를 가능하게 하면 안 됨 | D-2 ⑦ 둘째 줄 | D8.3 — 구체 타입 `*accountbudget.Budget`, 인터페이스 아님 |
| C3 | 실제 `official.Client` 통합 시험: 401→401→200 에서 admission 3회 · 둘째 거절 뒤 셋째 `c.hc.Do` 0 · 첫 거절이면 `c.hc.Do` 0. a087 5.5 가 그 이름·착지 sha 를 인용 | D-2 ⑦ 셋째 줄, D15.4 조건 2 | §9 T-INT-1~3 |
| C4 | 헤더 없는 HTTP 429 도 admission 거절로 표현하는 수명·재시작 처분(a087 이 `doRequest` 429 기록을 a166 에 넘김) | D-2 ⑦ 넷째 줄, design D15.1 | D4 |
| C5 | 호스트 측정 도구(별도 앱 키, 같은 계좌)도 같은 계좌 예산 — 키가 아니라 계좌가 키 | a169 `review.md` §10·§11, `issues.md` §6 재개 조건 2 | D1 |
| C6 | 엔진이 `ErrRateLimited` 를 받은 **모든** 경로(재시도 전·재시도 성공·Retry-After 유무)에서 정규 rate-limit 이벤트 | a169 §10 P1-9 | D5 — 자리를 `doRequest` 하나로 |
| C7 | admission 거절과 호출 시작이 같은 동기화 경계, 취소 뒤 새 호출 0, 이미 admission 한 호출 수의 상한 명시 | a169 §10 P1-8 | D6 |
| C8 | 보호 판단 `/prices`·청산 전송이 측정·보조 GET 에 굶지 않음(불변식 4) | Manager 지시, WORKFLOW §0.3 | D3 |

## 2. 전송 자리 지도 (HEAD `5e2facaee`)

근거: `grep -rn` 전수(`.Do(`·`RoundTrip(`·`official.New(`·`openapi.tossinvest.com`, `_test.go` 제외) + CodeGraph 1.6.0 `callers`
(`analysis/code-context/codegraph-callers-1st.txt`) + `affected --filter '*_test.go'`(`codegraph-affected-1st.txt` — 기본 판별식은 Go 시험을 못 찾는다).

### 2.1 공식 Open API 로 바이트를 내보내는 자리 — 전수

| # | 자리 | 무엇 | 지금 상한 | 처분(권장) |
|---|---|---|---|---|
| S1 | `internal/official/client.go:193` `Client.doRequest` 의 `c.hc.Do` | 모든 `official.Client` 동사(GET·POST·DELETE)와 `send` 의 401 재전송. CodeGraph 호출자 `send`(:320) 하나, `send` 호출자 `deleteAcct`(:369)·`getWithHeaders`(:392)·`postWithHeaders`(:427) | 없음 | **admission**(D8.1) |
| S2 | `internal/official/token.go:144` `tokenManager.exchange` 의 `m.hc.Do` | `POST /oauth2/token`. 호출자 `token`(:78)·`refresh`(:124) — 손 읽기(CodeGraph 는 `exchange` 호출자를 못 찾음) | 없음 | **admission, critical 등급**(D8.2) |
| S3 | `internal/official/a112_mbus_read.go:153` `a112MBUSReadAt` 의 `hc.Do` | 일회성 측정 GET. 호출자 `tools/a112-mb-us-source/official_reader.go:21·26·31` | 도구 자체 직렬 | **admission**(D8.4) — 사람 결정 H6 |
| S4 | `tools/a087-market-sell-probe/send.go:85` `sendOnce` 의 `client.Do` | 사람 승인 1회 POST(`AuthHeaders` 로 헤더만 빌림) | 1회 | **admission**(D8.4) — 사람 결정 H6 |
| S5 | `internal/execgw/replay_transport.go:95` `HTTPReplay.ReplayPlace` 의 `h.HTTP.Do` | IN_DOUBT 재생. 엔진 배선 `internal/app/engine/gateway.go:323` 이 별도 `http.Client` 로 만들고 `:331 Attested: nil` 로 **휴면** | 휴면 | 코드 변경 0 + 구조 시험(재생을 켜는 커밋은 admission 을 같이 넣어야 함 — D8.5) |
| — | `internal/monitor/official_read_contract.go:15` | `ContractOnly: true` — 보내지 않는 계약 probe | — | 대상 아님 |
| — | `internal/strategyevidence` `Transport.Do` | 생산 구현 0(시험 더블만), `NewOfficialAdapter` 생산 호출자 0 | 프로세스 안 `SharedRateBudget` | 대상 아님(배선되면 S1 을 거치게 할 것) |
| — | `internal/client/*`·`internal/monitor/probes.go:183` | WTS(쿠키 세션) — 공식 API 아님 | — | 범위 밖(§0) |

a165 가 더한 것(`internal/execgw/cancel_refusal.go` flatten 취소 409 재전송 최대 2회)은 `journal.DispatchFunc` → `trading` → `official.Client.postAcct` 로 가므로 **S1 을 지난다**
(손 읽기: `cancel_refusal.go:44` `withCancelRefusal` 은 원 dispatch 를 다시 부를 뿐 자기 HTTP 가 없다). 예산 계상은 §7.

### 2.2 프로세스 지도 — 누가 어느 Client 를 만드는가

| 프로세스 | 어디서 도나 | `official.New` 자리 | 비고 |
|---|---|---|---|
| 콘솔 | 컨테이너 `tossos` | `cmd/tossctl/instrument_names.go:125` `buildConsoleAccountBroker`(콘솔 계좌 화면·`console.go:794`), `candidatepanel.go:67` | 화면 조회 |
| 엔진 | 컨테이너 `tossos`(콘솔이 `engineproc.go:445` 로 `tossctl engine run` 을 띄움) | `internal/app/engine/engine.go:390` `NewOrderPath` ← `NewContext` ← `cmd/tossctl/engine_assembly.go:40` | Client **하나**를 모든 감독 루프 goroutine 이 공유. 재생용 `http.Client` 는 따로(S5, 휴면) |
| soak | 컨테이너 `tossos`(콘솔이 띄움, `soakautostart.go`) 또는 호스트 | `cmd/tossctl/soak.go:727` `buildSoakWiring` | 비지연 walk, `ReadRetryBackoff` 15s→30s |
| verify | 콘솔 또는 호스트 CLI | `cmd/tossctl/verify.go:889`·`:906`, `verify_reconcile.go:183`·`:204` | 한 명령 안에서 Client 둘(미결속 → 결속) |
| httpapi | 컨테이너 `httpapi` | 콘솔 seam 재사용(`httpapi.go:647` `newConsoleBroker(root)` → `console.go:794` `buildConsoleAccountBroker`) | 별도 컨테이너·별도 프로세스 |
| CLI | 호스트 | `internal/app/app.go:132`(hybrid, `root.go:524`), `cmd/tossctl/flatten.go:200`(→ `NewOrderPath`), `engine_reconcile.go:254`, `mcp.go:68`, `openapi.go:44` | upstream 상속 표면 |
| 도구 | 호스트 | `tools/a087-market-sell-probe/main.go:95`, `a112-l1c-quote-probe/main.go:60`, `a112-mb-us-source/official_reader.go:16`, `a128-d2-receipts/main.go:99`, `a128-jev-shadow-probe/main.go:127`, `a130-outcome-backfill/main.go:89`, `a140-breakout-signal-shadow/main.go:133`(a169 모드 포함) | a169 모드는 별도 앱 키·별도 `--config-dir` |

결론: 한 시각에 같은 계좌를 치는 프로세스가 최소 넷(콘솔·엔진·soak·httpapi)이고 호스트 CLI·도구가 더해진다. 프로세스 안 상한으로는 합을 못 지킨다.

### 2.3 공유 가능한 파일시스템

| 경로 | 종류 | 누가 보나 | 근거 |
|---|---|---|---|
| 호스트 `TOSSOS_CONFIG_DIR` → 컨테이너 `/var/lib/tossos/config` | bind mount(두 서비스 모두) | 콘솔·엔진·soak·httpapi·호스트 | `compose.yaml` `volumes` |
| 호스트 `/home/daniel/.config/tossctl` | **ext4**(`stat -f` = ext2/ext3 계열, `findmnt -T` → `/` ext4) | 호스트 CLI 기본 | 2026-10-29 측정(메타데이터만, 파일 내용 안 읽음) |
| `/mnt/D/…`(저장소) | **fuseblk**(ntfs-3g) | — | 같은 측정. flock 의미 신뢰 불가 — 예산 디렉터리로 쓰지 않는다 |
| 컨테이너 `/tmp`·`/run/tossos` | 컨테이너마다 tmpfs | 그 컨테이너만 | `compose.yaml` `tmpfs` — 공유 불가 |

- `TOSSOS_CONFIG_DIR` 가 실제로 `~/.config/tossctl` 인지는 배포 override(`~/.local/share/tossos-deploy/`)에 있고 이 작업은 읽지 않았다 — **[사람 확인]**(파일 기억은 같다고 말하지만 재검증 안 함).
- 컨테이너는 `user: ${TOSSOS_UID}:${TOSSOS_GID}` 로 호스트 사용자와 같은 uid 로 돈다(`compose.yaml`). 그래서 0600 파일을 양쪽이 연다. 루트 FS 는 `read_only: true` 라 예산 파일은 bind mount 안에만 둘 수 있다.
- flock: 같은 커널·같은 ext4 inode 이면 bind mount·mount namespace 를 넘어 배제가 성립한다(Linux `flock(2)` 의미). 이 저장소의 선례 `internal/enginelock/flock_unix.go` 는 같은 디렉터리에서 flock 을 쓰지만
  **컨테이너↔호스트 교차 배제를 잰 영수증은 없다 — [측정 필요]**, 시험 task 3.4 의 다중 프로세스 시험과 사람 승인 task 3.7 의 컨테이너↔호스트 1회 확인으로 닫는다.

## 3. 지금 코드의 429 처리 (AST 근거)

1. `Client.doRequest`(AST B1@194, B2@201): wire 응답이 오면 상태와 무관하게 `c.rates.record`(:199, 손 읽기) 뒤 `(status, body, nil)` 를 돌린다. 429 판정은 하지 않는다.
2. `Client.send`: 첫 `doRequest`(:329) 뒤 B4@344 `for attempt<2 && code==401` 이 `refresh` → `doRequest`(:354)를 최대 2번 더 부르고, B8@358 `!adopted` 이면 멈춘다. 그래서 한 `send` 의 wire 는 최대 3이다.
   B9@362 non-2xx → `classifyStatus` → 429 는 `ErrRateLimited`(`errors.go:54-55`, 손 읽기). **`send` 단위 admission 이면 :354 의 두 wire 가 관문을 우회한다**(C1 의 근거).
3. `Retrier.Query`: B3@347 `for attempt ≤ MaxAttempts(3)` · B4@350 `switch ClassifyQueryError` · B10@370 `ClassRateLimited` → `rateLimitWait` → B11@372 거짓이면 반환, B12@375 `sleepWithin` 거짓이면 반환, 참이면 **루프를 계속 돌아 다시 보낸다**.
   B5@351 `ClassOK` 는 `RecordSuccess` 만 하고 반환한다 — **재시도 성공 경로에 429 흔적이 없다**(a169 P1-9 의 반례).
4. `Retrier.rateLimitWait`: B1@431 `r.RetryAfter != nil` 이 거짓이면 B 바깥 :439 `BackoffFor(attempt, BaseBackoff)` = 400ms·2^(n−1)·(1±0.25).
5. **생산 배선에서 `RetryAfter` 는 nil 이다**(손 읽기): `internal/app/engine/gateway.go:340` `newRetrier(…, nil, in.clock)` → `exitwiring.go:51 RetryAfter: retryAfter`. `execgw.NewRetryAfterTransport` 의 생산 호출자는 0 이다(grep).
   그래서 정본 표 §1 「`min(Retry-After, 30s)` 1회 대기」와 §4 「`RetryAfterTransport`(헤더 포착)」는 **생산에서 참이 아니다**. 실제는: 429 → 약 400ms 뒤 재전송 → 또 429 면 약 800ms 뒤 재전송 —
   **한 조회가 약 1.2s 안에 wire 3번**이다. 이것은 R1(300ms 안 3회 전부 429)이 "회복하지 못한다"고 보인 바로 그 모양이고, 확정 표 「같은 주기 안 즉시 재시도 금지」와 정면으로 다르다.
6. `ClassifyQueryError`: B7@81 `ErrServer`/`ErrTransport` → transient, 그 밖의 모르는 오류는 :97 transient. 새 admission 거절 sentinel 을 넣지 않으면 **transient 로 분류돼 400ms 뒤 재시도된다** — D9 가 막는다.
7. mutation 분류 `classifyMutation` B8@84 → `statusOf` B4@134(`ErrRateLimited` → 429 → 모호). admission 거절을 `ErrRateLimited` 로 감싸면 **보내지도 않은 주문이 IN_DOUBT** 가 된다 — D8.1 이 별도 sentinel 을 쓰는 이유.
   별도 sentinel 은 `statusOf` 의 어느 갈래에도 안 걸리고(B6@140 default), `journal.ClassifyHTTPMutation(send, 0, err)` 가 SendTracker 상태(전송 전 = `SendNotStarted`)로 판정한다 — NOT_SENT. 이 사슬은 RED 시험으로 고정한다(T-MUT-1).
8. `ShouldFallback`(B1@68 뒤 :71-75): `ErrRateLimited` 면 hybrid 가 WTS 로 넘어간다. 새 sentinel 은 목록에 없으므로 넘어가지 않는다 — §8 H8.

확정 표와의 차이 요약: 현행 생산 = 「같은 조회 안 백오프 400ms·800ms 재시도(헤더 무시)」, 정본 표 = 「`min(RA,30s)` 1회 대기」, 확정 = 「같은 주기 재시도 금지, RA ≤30s 만 존중, 없으면 포기」.

## 4. 선택지 비교

| | ① 공유 상태 파일 + flock | ② 엔진 소유 브로커 프록시 | ③ 정적 분할 | ④ 헤더 피드백 공유 |
|---|---|---|---|---|
| 합 ≤ 3 req/s 보장 | 예 — 같은 파일의 같은 예약 규칙 | 예 — 한 곳 | **아니오** — 같은 종류 둘이면 깨짐, 지키는지 아무도 모름 | **아니오** — 사후 신호, 버스트 못 막음 |
| 엔진이 꺼져 있을 때(soak·CLI·flatten) | 동작 | **CLI·soak·flatten 이 엔진에 의존**(엔진 정지 = 손절 없음 창에 flatten 도 막힘) | 동작 | 동작 |
| 새 표면 | 파일 1개 + flock | 로컬 소켓 IPC·권한·인증 | 없음 | 파일 1개 |
| upstream 보존(토글 OFF) | Budget nil = 오늘 코드 | 프록시 경로 = upstream 과 다른 전송 경로 | 예 | 예 |
| 크래시 | flock 은 커널이 회수. 상태는 시각 몇 개라 회복 쉬움 | 프록시 죽으면 전 프로세스 정지 | — | — |
| 429 패널티 공유(C4) | 같은 파일에 | 프록시 메모리 | 불가 | 부분(헤더 의미 [미측정]) |
| 우선순위(C8) | 등급별 예약 규칙 | 큐 하나로 정확 | 불가 | 불가 |
| 비용 | 요청마다 flock+작은 pread/pwrite(µs, 상한 3 req/s 라 무시 가능). fuseblk 금지 | 큼 | 작음 | 작음 |

**권장: ①.** ②는 정확하지만 엔진 의존이 flatten·soak 의 독립성을 깨고(불변식 3 쪽으로 위험), 새 IPC 표면이 크다. ③④는 사용자 결정의 「합」을 보장하지 못한다.

## 5. 권장안 상세 (①)

### D1. 키 — 예산 디렉터리 하나 = 계좌 하나

- 예산은 `--config-dir` 가 아니라 **예산 디렉터리**에 산다. 해석 순서: 환경 변수 `TOSSOS_ACCOUNT_BUDGET_DIR` → 없으면 호스트 기본 `<os.UserConfigDir()>/tossctl/account-budget`.
  `compose.yaml` 은 두 서비스에 `TOSSOS_ACCOUNT_BUDGET_DIR=/var/lib/tossos/config/account-budget` 를 준다(bind mount 안). 그래서 다른 `--config-dir`·다른 앱 키(a169)를 쓰는 호스트 도구도 같은 파일을 연다(C5).
- **운영 불변식: 한 예산 디렉터리는 한 계좌다.** 계좌 식별(계좌 번호·seq)을 파일에 쓰지 않는다 — 계좌 번호는 개인정보이고(안전 부트스트랩 8), 해시해도 번호 공간이 작아 되돌릴 수 있다. 계좌가 하나인 것은 현재 실측 사실이다(M14·M35 「단일 계좌」).
  계좌가 둘이 되면 이 change 의 범위 밖이고, 그때 계좌별 디렉터리를 운영자가 가른다(§8 H1).
- 왜 config 디렉터리가 키가 아닌가: 2026-10-19 토큰 캐시가 config 디렉터리별로 갈려 서로 무효화한 것과 같은 모양을 피한다(파일 기억 — 재검증 안 함).
- 같은 예산을 보는지 확인하는 장치: 예산 파일은 처음 만들 때 비밀 아닌 무작위 `instance` 를 갖는다. 엔진은 기동 때 그 값을 비밀 없는 marker(`engine-run.json` 의 새 키 `account_budget_instance`)에 싣고,
  호스트 도구(a169)는 자기가 연 예산의 `instance` 와 그 marker 의 값이 같을 때만 GET 을 낸다(a169 P1-10 의 입력 — 검사 구현은 a169).

### D2. 알고리즘 — 최소 간격 예약(GCRA 변형)

- 상태(파일 하나, 고정 작은 JSON 레코드): `v`, `instance`, `next_slot`(다음 빈 자리, Unix ns), `penalty_until`(Unix ns), `wire_429`(누적), `last_429`, `refused`(누적), `crc32`.
- 간격 상수 `Spacing = 350ms`(확정 표의 페이지 간 최소 간격). **증명**: 자리 k 개가 길이 W 인 창에 들면 (k−1)·350ms ≤ W 이므로 W < 1s 이면 k ≤ 3. 즉 normal 등급 admission 은 **어떤 1초 창에도 3 이하**다.
- 판정 시각은 주입 `clock.Clock` 의 `Now()` 하나로 한다(a165 와 같은 원칙). 실시간은 대기의 ctx 상한으로만 쓴다(`clock.Sleep(ctx, d)`).
- 파일 갱신: `open(O_RDWR|O_CREATE, 0600)` 을 프로세스가 열어 두고, 매 admission 에서 `flock(LOCK_EX|LOCK_NB)` 을 짧게 재시도 → `pread` → 판정 → `pwrite(0)` + `ftruncate` → `LOCK_UN`.
  **임시 파일 + rename 을 쓰지 않는다** — flock 은 inode 에 걸리므로 rename 은 잠금을 둘로 쪼갠다. fsync 하지 않는다 — 프로세스 간 가시성은 페이지 캐시로 즉시이고, 전원 상실 뒤의 옛 시각은 이미 지난 값이다.

### D3. 등급과 우선순위 (C8 · 불변식 4)

| 등급 | 무엇 | 규칙 |
|---|---|---|
| **critical** | ① method ≠ GET 인 모든 요청(주문 접수·취소·정정·flatten, 조건주문) ② ctx 에 `accountbudget.WithCritical` 표지가 있는 GET — 엔진 exit 판단 `/prices`(`exitloop.go:864`), flatten 명령의 조회 ③ 토큰 교환(S2) | **대기 0**: 지금 바로 보낸다. 대신 `next_slot = max(next_slot, now) + Spacing` 으로 자리를 **소비**해 normal 을 뒤로 민다. 패널티 무시(D4). |
| **normal** | 그 밖의 GET(엔진 폴링·reconcile·soak·verify·콘솔·httpapi·CLI·도구·a087 하한가 GET) | `next_slot ≤ now` 이고 패널티 밖일 때만 자리를 얻는다(`next_slot = now + Spacing`). 아니면 **자리를 예약하지 않고** `next_slot`(+ jitter)까지 자고 다시 본다. ctx 가 끝나면 거절. |

- 왜 critical 이 기다리지 않는가: 안전 불변식 3(손절·비상 청산 즉시성 약화 금지)이 사용자 결정 3 보다 위다(권위 순서). normal 은 예약을 쌓지 못하므로 critical 은 다른 누구의 대기열 뒤에도 서지 않는다.
  대가: critical 버스트 동안 합이 3 req/s 를 넘을 수 있다(spec 이 이 예외를 명시한다). critical 은 드물다(§7). **사람 결정 H2** — 대안은 critical 도 간격을 지키되 normal 보다 앞에 서는 것(최대 지연 350ms × 앞선 critical 수).
- 표지 기본값의 방향: POST·DELETE 는 표지 없이도 critical(잊어도 안전 쪽 — 진입 매수가 critical 이 되는 정도). GET 은 표지가 없으면 normal 이므로 **보호 GET 의 표지 누락이 남는 위험**이다 → 구조 시험(T-STR-2)이 보호 GET 자리 목록에 표지가 있음을 고정한다.
- normal 끼리 공정성: 자리는 flock 을 먼저 잡은 쪽이 얻는다(순서 보장 없음). 한 프로세스가 독점하지 못하는 이유: 한 admission 은 자리 하나이고 normal 은 미래 자리를 예약하지 못한다. 한 프로세스의 동시 대기자가 많으면 확률적으로 더 자주 얻는다 — 상한은 두지 않는다(YAGNI, §8 H3).
- flock 을 못 잡을 때(다른 프로세스가 잠금을 쥔 채 SIGSTOP 등): critical 은 50ms 안에 못 잡으면 **기록 없이 보낸다**(fail-open, 이벤트 `budget_lock_unavailable`). normal 은 ctx 상한까지 재시도 후 거절(fail-closed).

### D4. 429 — 공유 패널티, 수명, 재시작 (C4)

- 자리: `doRequest` 가 wire 응답을 받은 직후 status==429 면 `Budget.Record429(now, retryAfterHeader, pathTemplate)` — 헤더 유무와 무관(a087 이 넘긴 몫).
- 패널티 길이 P: `Retry-After` 를 읽어(delta-seconds·HTTP-date) 0 < RA ≤ 30s 이면 P = RA, RA > 30s 이면 P = 30s(상한), 없거나 못 읽으면 **P = 15s**(R3: 15s 대기 뒤 회복 관측). `penalty_until = max(penalty_until, now + P)`.
  기본 15s 는 **사람 결정 H4**(a087 은 자기 갱신기에 5분 정지를 따로 둔다 — 그것은 a087 의 정책이고 이 패널티를 대체하지 않는다).
- 효과: 패널티 안에서 normal admission 은 **즉시 거절**(wire 0, `ErrAdmissionRefused{Reason: penalty}`). critical 은 영향 없음(D3).
- 수명·재시작: 패널티는 파일에 있으므로 **프로세스 재시작을 견딘다**(재시작은 복구 수단이 아님). 파일을 지우면 패널티도 사라진다(간격은 다시 성립) — 운영자 수단으로 남긴다.
- 시계 역행(같은 Clock 판정): `penalty_until > now + 30s` 이면 `now + 30s` 로, `next_slot > now + 2·Spacing` 이면 `now` 로 깎고 이벤트 `clock_regression` 을 남긴다. 그래서 어떤 시계 점프도 대기를 30s 넘게 만들지 못한다.
  시계 전진은 저장 값을 과거로 만들 뿐이라 자리 하나 정도의 버스트가 생긴다(허용).
- 손상(crc 불일치·파싱 실패·`v` 모름): 보수 재설정 — `penalty_until = now + 15s`, `next_slot = now + Spacing`, `instance` 새로 생성, 이벤트 `budget_reset`. instance 가 바뀌므로 a169 같은 도구는 멈춘다(의도).

### D5. 정규 rate-limit 이벤트 (C6 · a169 P1-9)

- 자리를 **`doRequest` 하나**로 둔다. 그래서 `Retrier.Query` 의 재시도 성공·soak `readRetry`·reconcile 15s 백오프·hybrid fallback 등 호출자가 429 를 어떻게 다루든 wire 429 는 빠짐없이 기록된다(호출자별 사본 금지 — [[two-judgements-cover-for-each-other]]).
- 두 겹: (a) **프로세스 간** — 파일의 `wire_429` 단조 계수와 `last_429`. 감시자는 시작 때 값을 읽고 늘면 멈춘다(a169 의 입력). (b) **프로세스 안** — `Budget.OnEvent func(Event)` 를 엔진 배선이 `obs.Logger` 에 연결해 한 줄
  `rate_limit event=wire_429 path=<템플릿> retry_after=<초|absent> penalty_until=<RFC3339> process=<engine|soak|…>` 를 남긴다. admission 거절은 `event=admission_refused reason=<penalty|ctx|closed|lock>`.
- 비밀 없음: 경로는 질의 문자열을 버리고 숫자·ID 세그먼트를 `{id}` 로 바꾼 템플릿만. 토큰·계좌 헤더·본문은 쓰지 않는다.

### D6. 닫기와 동기화 경계 (C7 · a169 P1-8)

- `Budget.Close() (admittedInFlight int)`: 프로세스 안 mutex 아래 `closed=true` 로 바꾸고, **그 순간 이미 admission 을 받고 `done()` 을 아직 부르지 않은 요청 수**를 돌려준다.
  `Admit` 은 대기에서 깨어난 뒤에도 같은 mutex 아래 `closed` 를 다시 본다. 그래서 **Close 반환 뒤 새로 admission 을 얻는 요청은 0** 이고, Close 전에 받은 요청은 최대 `admittedInFlight` 다
  (admission 과 `c.hc.Do` 사이는 같은 goroutine 의 몇 문장 — 그 창은 이 수에 포함된다). a169 도구는 직렬이라 그 수가 ≤ 1 이다.
- 교차 프로세스 정지: `Budget` 정책 `StopOnShared429` — 켜면 admission 판정(flock 안)에서 `wire_429` 가 열 때보다 커졌으면 거절한다. 판정과 거절이 같은 flock 구간이라 "엔진이 429 를 맞은 뒤 도구의 새 wire" 는 0 이다.
  이 정책은 도구용이고 엔진·콘솔은 끈다.

### D7. 실패 처분 요약

| 상황 | 엔진 | 그 밖(soak·verify·콘솔·httpapi·도구) | CLI(upstream 상속 명령) |
|---|---|---|---|
| 예산 디렉터리 열기 실패·권한 거부 | 프로세스 안 대체 Budget(같은 규칙, 공유 없음) + critical 알림 + **신규 진입 차단 latch**(`ReasonAccountBudgetUnshared`, 열리면 자동 해제 아님 — 재시작) | 대체 Budget + 경고 | 대체 Budget, 경고 없음(§8 H7) |
| 파일시스템이 ext4/xfs/btrfs/tmpfs 가 아님(`statfs` — fuseblk·nfs·overlay 등) | 위와 같음 | 위와 같음 | 위와 같음 |
| flock 지연 | D3 | D3 | D3 |
| 손상·시계 역행 | D4 | D4 | D4 |

- 엔진이 기동을 **거부하지 않는** 이유: 엔진 정지는 손절 없음이다(파일 기억 「엔진 정지 = 손절 없음」). 공유가 깨지면 진입만 막고 보호는 계속한다. **사람 결정 H5**.

### D8. `official.Client` 통합

1. **`doRequest`**(C1): `c.hc.Do`(:193) 바로 앞에서 `c.budget != nil` 이면 `done, err := c.budget.Admit(req.Context(), classOf(req))`, 거절이면 wire 없이 `(0, nil, err)` 반환.
   응답 뒤 status==429 면 `Record429`. `defer done()`. `send` 는 바꾸지 않는다 — 401 재전송 두 wire 도 `doRequest` 를 지나므로 각각 admission 을 탄다(§3 2).
   거절 오류는 새 sentinel `official.ErrAdmissionRefused`(`ErrRateLimited`·`ErrTransport` 가 **아님**) — §3 6·7.
2. **토큰 교환**(S2): `exchange` 의 `m.hc.Do`(:144) 앞에 같은 admission, 등급 critical 고정. 이유: `exchange` 는 `m.mu` 를 쥔 채 돌고(`token`·`refresh` 가 잠금을 잡고 부름, 손 읽기 :61·:109), 그 잠금은 캐시 적중
   조회도 막는다. normal 로 기다리면 보호 GET 이 `m.mu` 뒤에서 굶는다(우선순위 역전). 토큰 교환을 브로커가 같은 한도에 세는지는 [미측정] — 보수적으로 센다.
3. **같은 인스턴스**(C2): 새 Option `official.WithAccountBudget(*accountbudget.Budget)`(봉인 전 1회), 접근자 `(*Client).AccountBudget() *accountbudget.Budget`. **구체 타입**이고 필드가 비공개라 시험 더블·dummy 를 생산 배선에 넣을 길이 없다.
   엔진은 `NewOrderPath` 에서 Budget 하나를 만들어 그 Client 에 넣고, a087 활성화(5.5)는 갱신기가 쓰는 Client 의 `AccountBudget()` 가 **nil 아님 + 엔진이 만든 그 포인터**임을 확인한다. a087 갱신기는 자기 preflight 요청을 만들지 않는다(같은 Client 의 동사만).
4. **우회 자리**(S3·S4): `a112MBUSReadAt` 은 같은 Client 의 `AccountBudget()` 로 `hc.Do`(:153) 앞에서 admission(normal), 거절이면 기존 hold 경로(`a112MBUSHoldf`)로. a087 프로브 `sendOnce` 는 `main.go` 가 같은 Budget 을 받아
   `client.Do`(:85) 앞에서 admission(critical — POST), 거절이면 `Sent=false`. 두 도구를 은퇴시키는 쪽이 더 짧으면 그쪽(§8 H6).
5. **재생**(S5): 코드 변경 0. 구조 시험 T-STR-3 이 「`gateway.go` 의 `Attested` 가 nil 이거나, `HTTPReplay` 가 Budget 을 지난다」를 고정한다 — 재생을 켜는 커밋이 예산을 빠뜨리면 시험이 깨진다.
6. 새 코드는 새 패키지 `internal/accountbudget`(새 파일·새 함수 — FLM 대상 아님)에 둔다. `official` 이 그것을 import 하고 반대 방향은 없다.

### D9. 엔진 폴링 429 — `Retrier.Query`

- `ClassifyQueryError` 에 `ErrAdmissionRefused` → `ClassRateLimited` 갈래를 넣는다(B6@79 앞).
- `ClassRateLimited` 갈래(B10@370)는 **바로 반환**한다 — 같은 조회 안 재시도 0. staleness 는 다음 성공까지 낡아 가고 임계를 넘으면 신규 진입이 막힌다(기존 B5 경로 그대로).
- 「Retry-After ≤30s 존중」은 D4 의 공유 패널티가 담당한다(다음 주기의 normal admission 이 그 시각까지 거절). 「>30s」는 패널티 30s 상한 뒤 다음 시도가 다시 429 면 다시 30s — 결과는 같은 「포기 → staleness」다.
- 그러면 `rateLimitWait`·`Retrier.RetryAfter`·`RetryAfterSource`·`RetryAfterTransport` 는 생산·시험 밖에서 소비자가 없다 → **삭제 권장**(정본 §4 의 거짓 행도 같이 지움). `newRetrier` 의 `retryAfter` 인자도 지운다. **사람 결정 H9**(남기면 죽은 코드가 정본 표와 다시 갈린다).
- 비지연 walk(soak·verify `ReadRetryBackoff` 15s→30s, reconcile `DefaultRateLimitBackoff` 15s)는 그대로(확정 표 「유지」). 그 대기 뒤의 재시도도 admission 을 지나므로 패널티 안이면 wire 없이 거절된다 — 그 재시도를 「한 번 쓴 것」으로 셀지는
  각 호출자의 기존 규칙대로다(시험 T-WALK-1 로 soak 이 거절을 429 와 같게 다룸을 확인).

### D10. 배포 · UID · rollback

- 배포: 이미지 재빌드 + 두 서비스 재시작(엔진 정지 창 KST 05:00~09:00, 사람). `compose.yaml` 에 env 한 줄씩. 저널·원장 스키마 변경 없음.
- 권한: 디렉터리 0700·파일 0600, 소유자 = `TOSSOS_UID`(호스트 사용자). 다른 uid 로 도는 도구는 열기 실패 → D7.
- rollback: 이전 이미지로 되돌리면 admission 이 없다(오늘 동작). 남은 `account-budget/` 파일은 아무도 읽지 않는다(지워도 됨). 배포 순서 제약 없음 — 섞여 돌면(새 엔진·옛 CLI) 옛 쪽이 예산을 안 탈 뿐 안전 방향은 같다.
- 토글 OFF = upstream: Budget nil 이면 `doRequest`·`exchange` 는 오늘 바이트 그대로다(시험 T-OFF-1 이 nil 경로의 요청 수·순서를 오늘과 비교).

## 6. 권한·불변식 점검

- 불변식 3·4(손절·청산 즉시성): critical 대기 0(D3), 토큰 교환 critical(D8.2), 패널티가 critical 을 막지 않음(D4), 예산 실패에도 엔진 보호 지속(D7). 남는 위험은 보호 GET 표지 누락(T-STR-2).
- 불변식 4(호출 추가 계상): 이 change 는 wire 요청을 **추가하지 않는다**(`/oauth2/token` 포함 기존 요청에 관문만). §7.
- 불변식 2/OFF: D10.
- 시크릿: 파일·이벤트에 토큰·계좌 식별·본문 없음(D1·D5).

## 7. 예산 계상 (정본 §3 갱신안)

| 출처 | 등급 | 정상 상태 |
|---|---|---|
| 엔진 미체결·잔고·보유·주문 단건 폴링 | normal | ≈ 0.5 req/s(정본 §3) |
| 엔진 exit `/prices`(보호 종목당 주기마다) | critical | 보호 종목 수 × 주기 — 수치 [미측정](a087 D15.3 칸 4 가 이 몫을 읽는다) |
| a165 엔진 취소 409 지속 종목 | critical(POST)·normal(사전 GET) | ≈ 0.4 req/s/종목(a165 D2) |
| a165 flatten 취소 1건 | critical | (GET+POST) × 최대 2 |
| soak 주기 | normal | 26 페이지 walk 가 ≥ 9.1s(=26×350ms) 로 늘어남 — 그동안 엔진 normal 과 자리를 나눔 |
| 토큰 교환 | critical | 하루 1회 + 401 갈래 |

normal 합은 구조상 ≤ 3 req/s(D2 증명). critical 은 상한이 없으므로 spec 이 예외로 명시하고, 실측은 배포 뒤 이벤트(`wire_429`)로 본다.

## 8. 사람 결정 후보

| # | 질문 | 권장 |
|---|---|---|
| H1 | 키 = 예산 디렉터리(계좌 1:1 운영 불변식, 계좌 식별을 파일에 안 씀) | 수락 |
| H2 | critical 은 간격·패널티를 건너뛰고 자리만 소비(손절 지연 0, critical 버스트 동안 3 req/s 초과 가능) vs critical 도 간격 준수(최대 지연 350ms×순번) | 건너뜀(불변식 3 우선) |
| H3 | normal 끼리 프로세스별 몫·상한 없음 | 없음(YAGNI) |
| H4 | 헤더 없는 429 의 패널티 15s(R3), 상한 30s | 15s |
| H5 | 예산을 못 열면 엔진은 기동하되 신규 진입 latch + critical 알림(기동 거부 아님) | 수락 |
| H6 | 우회 자리 S3(a112 측정)·S4(a087 프로브): admission 을 넣는다 vs 도구 은퇴 | admission(작음), 도구가 더 안 쓰이면 은퇴 |
| H7 | upstream CLI 명령도 기본으로 예산을 탄다(열기 실패 시 조용히 프로세스 안 대체) — 새 설정 토글을 만들지 않음 | 수락 |
| H8 | `ErrAdmissionRefused` 는 hybrid WTS fallback 대상이 아님 | 수락 |
| H9 | `RetryAfterTransport`·`rateLimitWait`·`Retrier.RetryAfter` 삭제(생산 소비자 0) | 삭제 |
| H10 | 컨테이너↔호스트 flock 교차 배제 1회 확인(실계좌 호출 0 — 예산 파일만, task 3.7) | 승인 요청 |

## 9. 시험 전략 (이름은 RED 에서 확정 — 아래는 a087·a169 가 인용할 최소 집합)

- **T-INT-1~3**(C3, `internal/official`): 실제 `official.Client` + `httptest` + 실제 `*accountbudget.Budget`(fake Clock, 임시 ext4 디렉터리). 401→401→200 에서 admission 3회·wire 3.
  `Close()` 를 첫 wire 처리 중(핸들러 안)에 부르면 → 둘째 admission 거절, **wire 1**(둘째·셋째 `c.hc.Do` 0). 둘째 wire 처리 중이면 → wire 2(셋째 0). 시작 전 Close → wire 0. 각 경우 `errors.Is(err, ErrAdmissionRefused)`.
- **T-429-1**(C4): 헤더 없는 429 → `penalty_until = now+15s`, 새 `Budget`(재시작 흉내 — 같은 디렉터리를 다시 엶)에서 normal 거절·critical 통과, 15s 뒤 normal 통과. RA=5s → 5s, RA=90s → 30s, HTTP-date 형식.
- **T-EVT-1**(C6): 429 뒤 200 으로 끝나는 네 경로(엔진 `Retrier.Query` · soak `readRetry` · RA 있음 · RA 없음)에서 `wire_429` 가 각각 정확히 1 증가.
- **T-SYNC-1**(C7): 대기 중 normal 여럿 + `Close()` → Close 반환 뒤 admission 0, 반환값 = 그 순간 in-flight 수. `StopOnShared429`: 다른 프로세스(자식)가 429 를 기록하면 이후 admission 0.
- **T-SPACE-1**: fake Clock 무작위 도착 10⁴건 — 어떤 1s 창에도 normal ≤ 3(전칭 판정 표본 0 금지: 도착 수 단언).
- **T-PRIO-1**: normal 대기열이 찬 상태에서 critical 대기 0, critical 뒤 normal 은 `next_slot` 이후.
- **T-CLK-1**: 시계 역행 1h → 대기 ≤ 30s, 이벤트 `clock_regression`. 손상 파일 → 보수 재설정.
- **T-MP-1**(다중 프로세스, `-race`): `TestMain` 재실행 자식 4개가 같은 디렉터리에서 normal admission 을 실시간으로 다툼 → 전체 admission 시각 간격 ≥ 350ms − ε. 임시 디렉터리가 D7 의 허용 파일시스템이 아니면 Skip 하지 않고 실패한다(표본 0 통과 금지). fuseblk 거부는 `statfs` 판별 단위 시험으로 따로 잰다.
- **T-RET-1**(D9): `Retrier.Query` 에 429 → `fn` 호출 1회·대기 0, `ErrAdmissionRefused` → 1회. 기존 `TestDefaultsMatchThePublishedMatrix` 갱신.
- **T-MUT-1**(§3 7): 주문 POST 가 admission 거절 → `DispatchNotSent`(IN_DOUBT 아님).
- **T-STR-1~3**: `doRequest` 의 `Admit` 이 `c.hc.Do` 바로 앞(AST 순서), `exchange` 도 같음 / 보호 GET 자리 목록(`exitloop.go:864` 등, RED 에서 전수)에 `WithCritical` / S5 재생 결속.
- **T-OFF-1**: Budget nil → 요청 수·순서 무변화(upstream 상속 시험 포함 영향 범위).
- **T-WALK-1**: soak walk 가 admission 거절을 429 와 같게 다룸.
- 변이 원장: `Spacing` 상수·`<`/`≤`·등급 규칙(method/표지)·패널티 갈래·상한 깎기·`Admit` 호출 삭제(양성 대조군)·`Record429` 삭제·`Close` 의 closed 재검사·`ErrAdmissionRefused` 를 `ErrRateLimited` 로 바꾸기.

## 10. FLM 대상(편집할 기존 함수 — gate ⑤ 번들 필요)

`official.Client.doRequest` · `official.tokenManager.exchange` · `official.New`(Option 추가가 본문을 안 바꾸면 제외 — RED 에서 확인) · `execgw.ClassifyQueryError` · `execgw.Retrier.Query` · (삭제) `execgw.Retrier.rateLimitWait`·`RetryAfterTransport.*` ·
`engine.newRetrier` 와 그 호출자 `gateway.go` 의 배선 함수 · `engine.NewOrderPath`(Budget 배선) · `official.a112MBUSReadAt` · `a087 probe sendOnce`/`main` · 보호 GET 표지를 다는 엔진 함수(`exitloop.go:864` 의 감싼 함수 등) · soak·verify·console·httpapi·CLI 의 Client 생성 함수(Option 추가 — 본문이 바뀌면 전부).
`Client.send` 는 **바꾸지 않는다**(AST 는 C1 의 근거로만 씀). 기존 시험 함수를 고치면 그것도 번들(gate ⑤).
