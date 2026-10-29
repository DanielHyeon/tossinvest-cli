## design 1판 독립 적대 리뷰 (GPT-5.6 Terra, 2026-10-29)

### 결론: **FAIL (blocker 4)**

`flock` 상태 파일 자체는 적은 표면으로 합산 조정을 만들 수 있는 방향이다. 그러나 현재 1판은 (1) 우회 wire의 429를 공유 상태에 기록하지 않고, (2) 컨테이너·호스트가 같은 계좌 파일을 연다는 계약이 없고, (3) critical 예외와 `Admit` 시각만의 증명이 사용자의 계좌 전체 3 req/s 결정을 만족하지 못하며, (4) `Close`/Stop이 admission 뒤·실제 전송 전 창을 닫지 않는다. 이 넷이 닫히기 전 구현 착수 불가다.

### 발견

| ID | 심각도 | file:line | 근거 | 처분 제안 |
|---|---|---|---|---|
| P0-1 | P0 | `design.md:147-160,184-193`; `internal/official/token.go:144`; `internal/official/a112_mbus_read.go:153`; `tools/a087-market-sell-probe/send.go:85` | **AST.** D4/D5는 `doRequest` 응답만 `Record429` 한다. 하지만 S2 토큰 교환, S3 a112, S4 a087 probe는 각각 독자 `Do`다. D8은 이들 앞 admission만 적고 응답 429 기록을 적지 않았다. 따라서 헤더 없는 429의 재시작 패널티(C4), 모든 wire 429 이벤트(C6/a169 P1-9)가 셋에서 빠진다. | 모든 S1~S5가 공유 `Record429`와 정규 이벤트를 거치는 하나의 transport 경계로 모으거나, 각 우회 응답의 429도 같은 함수로 기록하라. S2/S3/S4별 `429 → 재시작 normal 거절`과 이벤트 1회 RED를 추가하라. |
| P0-2 | P0 | `design.md:116-122,205-210`; `compose.yaml:42-56,123-140` | **손으로 읽음.** 컨테이너에는 새 env를 넣는다고 하나, 호스트 CLI/도구는 env가 없으면 `<os.UserConfigDir()>/tossctl/account-budget`로 간다. compose bind source와 그 경로가 같은 실제 디렉터리라는 계약·배선·실패 처분이 없다. 다른 `--config-dir`/별도 앱 키 호스트 도구도 자동으로 같은 파일을 열지 않는다. C5와 “모든 TossOS 프로세스”가 성립하지 않는다. | 계좌 하나당 canonical budget directory를 배포·호스트 CLI·도구 모두에 명시 주입하고, 시작 전 동일 inode/instance를 검증하라. 다르면 normal wire 0으로 거부할 것. 실제 bind mount의 경로·UID·mode 검증은 H10 사람 gate에 넣어라. |
| P0-3 | P0 | `design.md:127,136-143,230,237`; `specs/order-execution/spec.md:31-32` | **AST 불필요(설계 자체).** 증명은 normal admission에만 해당한다. 그러나 모든 non-GET(진입 주문 포함), 토큰 교환, lock 50ms 실패 critical이 무대기·무패널티로 전송되어 임의 burst를 만든다. spec도 계좌 전체 3 req/s 결정을 “critical 제외 normal”로 축소한다. `next_slot` 소비는 이미 보낸 critical의 합을 제한하지 않는다. | H2를 현 문구로 수락하지 말 것. 실제 보호 청산·flatten·필수 보호 토큰만 명시적 critical로 최소화하고, 진입 mutation은 normal로 둘 것. 보호 critical의 불가피한 예외는 수치 상한이 아니라 별도 안전 예외·이벤트·사람 수락으로 명시하라. 사용자 결정 3의 의미를 바꿀 권한은 이 change에 없다. |
| P0-4 | P0 | `design.md:127-130,184-185,249-257`; `internal/official/client.go:191-206` | **AST.** 350ms는 flock 아래 예약/admission 시각만 간격을 둔다. `Admit` 반환 뒤 goroutine이 멈추면 먼저 예약한 요청이 나중에 `c.hc.Do`를 시작해 뒤 예약들과 한 1초 창에 몰릴 수 있다. T-MP-1도 admission 시각만 재며, 실제 wire 시작/지연 주입을 재지 않는다. 그러므로 “wire 요청 합 ≤3” 증명이 아니다. | 요구의 측정점을 `c.hc.Do` 시작으로 명시하고 그 시점까지 유효한 commit 경계/전송 직렬화 설계를 정하라. 그것이 불가능하면 계약을 “admission 시각”으로 축소하고 브로커 한도 보장이라고 말하지 말 것. 전송 직전 scheduler 지연을 주입해 wire 시작 창을 세는 다중 프로세스 RED를 추가하라. |
| P1-1 | P1 | `design.md:165-168,250,253`; `specs/order-execution/spec.md:50-52` | **AST/손으로 읽음.** `Close`는 새 admission만 막고 이미 admission한 요청의 `done()` 전 수를 반환한다. 그 요청은 Close 반환 뒤 처음 `c.hc.Do`를 시작할 수 있다. “같은 goroutine의 몇 문장”은 a169 P1-8의 같은 동기화 경계가 아니다. 현재 T-INT는 handler 안 Close만 보아 이 반례를 못 낸다. | Stop/Close의 완료 정의를 “이미 admission한 전송 시작 수 상한과 drain 완료”까지 정하고, `Admit` 뒤·`Do` 전을 막는 seam에서 Close를 주입하라. a169 StopOnShared429은 이 drain 전에는 측정을 재개하지 않도록 계약화하라. |
| P1-2 | P1 | `design.md:141,175-180,242`; `internal/app/engine/precheck.go:23-43,98-105` | **손으로 읽음.** 공유 파일 실패 뒤 모든 프로세스가 독립 3/s 대체 Budget을 쓰므로 N개 프로세스의 합은 다시 N×3/s다. 신규 진입 latch만으로 콘솔·soak·verify·CLI normal GET은 멈추지 않아 `/prices`를 다시 굶길 수 있다. 또한 보호 GET 목록은 “RED에서 전수”라서 지금 검증 가능한 계약이 아니다. 취소 전 `OrderRawByID` GET은 실패해도 취소를 진행시키므로 normal이어도 되지만, 이처럼 각 보호 관련 GET의 지연/실패 처분을 지금 열거해야 한다. | H5/H7의 local fallback을 shared-cap 충족으로 보고하지 말 것. 공유 불능 시 noncritical 실행을 전 계좌 범위에서 중지/명시 거절하고, 엔진은 보호 mutation과 fail-open 취소만 지속하는 처분을 정하라. `/prices`, flatten 조회, cancel precheck, cancel 확인/재조정, a087 하한가 갱신 각각을 critical 또는 normal+의존 action fail-open으로 design/spec/RED에 고정하라. |
| P1-3 | P1 | `design.md:191-193,259-260`; `internal/execgw/cancel_refusal.go:43-104`; `internal/execgw/classify.go:84-98` | **AST.** a165 `withCancelRefusal`은 raw HTTP가 아니라 `dispatch`를 다시 부르므로 S1 배선이면 retry POST는 관문을 지난다. 그러나 T-MUT-1은 일반 주문 POST만 다룬다. 첫 cancel admission 거절은 `DispatchNotSent`여야 하고, 첫 409 뒤 retry admission 거절은 첫 `DispatchAmbiguous`를 보존해야 한다(`cancel_refusal.go:95-100`). 이 상호작용을 고정하지 않으면 NOT_SENT/IN_DOUBT 계약이 깨질 수 있다. | T-MUT를 a165 두 경우로 확장하고 `DispatchVerified` 경로까지 종단 시험하라. `ErrAdmissionRefused`를 `ErrRateLimited`로 감싸지 않는 RED도 유지하라. |
| P1-4 | P1 | `design.md:128,152-153,256-257` | **설계 반례.** 상태 파일은 Unix ns인데 판정 Clock은 각 프로세스에 주입된다. 서로 다른 Clock/offset에서 뒤 프로세스가 `next_slot > now+700ms`를 clock regression으로 잘라 즉시 자리·wire를 얻을 수 있다. T-CLK-1은 단일 Clock의 1h 역행만, T-MP-1은 같은 실시간 Clock만 잰다. | 공유 파일 판정은 하나의 실제 wall-clock 의미와 허용 오차를 정하거나, process A/B offset·역행을 넣은 다중 프로세스 RED에서 보수 거절을 증명하라. “대기 최대 30s”와 3/s 보장은 별도 성질로 기록할 것. |
| P2-1 | P2 | `internal/app/engine/gateway.go:340`; `internal/execgw/retry.go:69-98,335-440`; `design.md:201,267-268` | **AST.** `newRetrier(..., nil, ...)`은 확인했고 `NewRetryAfterTransport`의 비시험 생산 호출은 0이다. H9 삭제 방향은 현재 사실과 맞다. 다만 현재는 최대 3회이며 400ms·800ms 기본 backoff(+jitter)라 “약 1.2s 안 wire 3회”는 명목값이지 AST가 보장한 실측값은 아니다. | H9는 수락 가능. 삭제 대상 public type·시험·문서·`newRetrier` 호출부를 FLM/BTM 및 삭제 RED에 전부 넣고, 시간 주장은 “명목 backoff 합(지터 별도)”으로 고쳐라. |
| P2-2 | P2 | `design.md:129-130,154,257`; `tasks.md:30-36` | **손으로 읽음.** SIGKILL이면 flock은 풀리고 반쪽 JSON은 D4의 CRC/parse 실패→15s reset으로 normal을 보수 처리한다. 이 방향은 맞다. 다만 pwrite/ftruncate 사이 크래시, 새 instance로 a169 중지, 허용 fs(ext4/xfs/btrfs/tmpfs)·거부 fs(NFS/fuseblk/overlay)를 묶은 시험은 아직 구체 RED가 아니다. bind mount·동일 UID는 compose 선언상 가능하지만 실제 배포 override는 금지 범위라 확인하지 못했다. | H10을 실제 canonical bind path에서의 사람 승인 gate로 유지하고, partial write/손상·NFS/fuseblk 거부·다른 UID open 실패를 단위/자식 프로세스 RED에 넣어라. |

### 사람 결정 H1~H10 권고

| 결정 | 권고 | 이유 |
|---|---|---|
| H1 | **보류/수정** | 계좌=예산 디렉터리는 계좌 식별을 파일에 쓰지 않는 방향은 맞지만, host/container/도구의 canonical 같은 inode 계약이 없다. P0-2를 먼저 닫아야 한다. |
| H2 | **현 문구 거부** | 보호 즉시성은 우선이나 모든 mutation·token의 무상한 bypass는 계좌 전체 3 req/s를 무효화한다(P0-3). |
| H3 | **조건부 수락** | P0-2/P1-2가 닫힌 단일 공유 파일에서만 normal 프로세스별 quota는 불필요하다. |
| H4 | **조건부 수락** | 15s/30s는 헤더 없는 429의 보수 기본값으로 가능하나 S1~S5 모든 429 기록·재시작 RED가 먼저다. |
| H5 | **수정 수락** | 엔진 기동 거부는 손절을 없애므로 피한다. 다만 local fallback은 noncritical 계좌 전체를 중지/거절해야 하며 entry latch만으로는 부족하다. |
| H6 | **보류** | admission만 추가하면 S2~S4의 429 패널티/이벤트가 빠진다. 공통 response 기록까지 넣거나 도구를 은퇴해야 한다. |
| H7 | **거부** | “기본으로 예산”과 “열기 실패 시 조용한 local fallback”은 OFF=upstream 및 계좌 공유 상한을 둘 다 증명하지 못한다. 명시된 ON 주입과 nil OFF, 실패 시 처리 계약이 필요하다. |
| H8 | **수락** | 별도 sentinel은 현재 `statusOf` default를 지나 SendNotStarted면 `DispatchNotSent`가 된다. hybrid fallback 제외와 mutation RED를 함께 고정할 것. |
| H9 | **수락** | 현재 생산 `RetryAfter` 주입은 nil이고 `NewRetryAfterTransport` 생산 호출도 0이다. P2-1의 삭제 범위 검증을 전제로 한다. |
| H10 | **수락(배포 전 사람 gate)** | broker 호출 없이 실제 canonical bind path·host/컨테이너 두 PID·동일 UID에서 flock 교차 배제를 확인해야 한다. NFS/fuseblk/UID 실패도 함께 거부 확인할 것. |

### 생산 사실 대조

- **AST 확인:** `Client.doRequest`의 유일한 `c.hc.Do`는 `client.go:193`; `Client.send`은 `doRequest`를 첫 1회와 401 loop에서 최대 2회 더 호출한다(`:329,:344-355`). C1의 wire 단위 삽입 자리는 맞다.
- **AST 확인:** `tokenManager.exchange`는 독립 `m.hc.Do`(`token.go:144`), a112는 독립 `hc.Do`(`a112_mbus_read.go:153`), a087 probe는 독립 `client.Do`(`send.go:85`)다. P0-1의 근거다.
- **AST 확인:** `ClassifyQueryError`는 현재 `ErrRateLimited`만 `ClassRateLimited`로 보내고(`retry.go:69-98`), `Retrier.Query`는 그 갈래에서 대기 뒤 loop를 계속 돈다(`:347-388`). D9의 새 sentinel 갈래·즉시 반환이 필요하다.
- **AST 확인:** `statusOf`는 `ErrRateLimited`만 429로 돌리고 새 sentinel은 default `(0,false)`가 된다(`classify.go:125-143`). `journal.ClassifyHTTPMutation(SendNotStarted,0,err)`는 `DispatchNotSent`다. D8의 오류 분리 자체는 맞다.
- **AST 확인:** a165는 `withCancelRefusal`의 `dispatch` 재호출(`cancel_refusal.go:51,90`)이지 자체 HTTP 전송이 아니다. S1 배선이 빠지지 않으면 retry POST도 계상된다.

### 적용성·검증·미실행

- `impeccable`: N/A — backend/OpenSpec 설계 리뷰만, UI/CLI 렌더 변경·검증 0.
- `search-first`: 구현·의존성 채택 0. 저장소 정적 검색만 했고 외부 레지스트리/MCP/GitHub 채널은 보지 않았다. tasks 0.4는 여전히 미완료다.
- `grill-me`/`grilling`: N/A — 이 문서는 독립 적대 리뷰이며 H1~H10의 사용자 결정을 대체하지 않고 위 권고만 남긴다.
- ai-memory: 현재 project `a166 OR rate limit OR account budget` 검색 결과 0. 보조 파일 memory의 a087/a169 문맥은 현재 source·OpenSpec으로 재대조했다.
- 실행: `openspec validate a166-account-shared-rate-limit --strict` 성공. `tools/logic-map` AST를 `Client.doRequest`, `Client.send`, `tokenManager.exchange`, `Retrier.Query`, `rateLimitWait`, `ClassifyQueryError`, `classifyMutation`, `statusOf`, `a112MBUSReadAt`, `sendOnce`에 실행했다. 현재 worktree의 CodeGraph는 초기화되지 않아 질의 불가(`CodeGraph not initialized`); 기존 산출물은 보조로만 썼다.
- 미실행: Go 시험·build·vet·lint·race·`make sdd-*`·gstack, Docker/컨테이너, 운영 config/marker, 브로커/API, live order 및 모든 `mutating: true` 명령. 소스·시험·운영 파일 편집 0.

### 구현 착수 전 필수 순서

1. P0-1~P0-4를 design/spec/tasks에 반영하고 H1·H2·H5·H6·H7 결정을 다시 받는다.
2. 계좌별 canonical 경로·모든 bypass response 기록·정확한 critical 범위를 먼저 RED로 고정한다.
3. 전송 시작/Close 경계와 다중 Clock 반례를 재현하는 다중 프로세스 RED를 통과한 뒤에만 FLM/BTM·production 편집에 들어간다.
