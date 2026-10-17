<!-- 22라운드 보이스 A(생산·안전 렌즈) — Claude code-reviewer 서브에이전트(독립 컨텍스트, 같은 하네스). 프롬프트 prompt-common.md + 호출 렌즈.
     대상 7cf80832(리뷰 중 HEAD 90e5170d — a066 파일 5개만 들어옴). 원문 그대로. -->

**판정: BLOCK** (P0 0 · P1 3 · P2 10)

손절·청산의 즉시성과 exit goroutine의 원격 전송 대기 0은 22판 설계대로 선다. 다만 구현 전에 고칠 결함 셋이 남아 있다.
- a092가 새로 만드는 완화 경로에서 critical 통지가 조용히 사라진다.
- Q4 핀의 정의가 오늘의 `parkAlert`와 서로 모순된다.
- 병행 change(a090·a094·a066)의 직접 기록자와 a092의 SHALL·핀이 충돌한다.

리뷰 중 HEAD가 `90e5170d`(a066 5.5.4)로 움직였다. a092가 인용한 운영 파일의 sha는 `head-ast-21/MANIFEST.txt`와 같다. 새로 들어온 것은 a066 파일 5개뿐이다.

## 발견 표

| # | 등급 | 주장 | 증거(파일:줄) | 권고 |
|---|---|---|---|---|
| 1 | **P1** | 운영 모드 통지 key에 에피소드 신원이 없다. a092가 완화 경로를 열면 「강화 → 완화 → 재강화」나 두 번째 완화가 재알림 창(1h) 안에서 일어날 수 있다. 그때 통지는 옛 정착 행에 흡수되어 나가지 않는다. 401 경로는 모드 통지가 유일한 알림이라, 운영자는 진입이 다시 막힌 사실을 모른다. a094 R6-2와 같은 결함 부류이고, a092 이전에는 완화 경로가 없어 잠복해 있었다 | `internal/obs/mode.go:59-62`(key = 계정+모드, "transition id … would defeat that") · `internal/obs/notifier.go:59`(1h) · `internal/journal/outbox.go:289-294,346` · `claimOwed :407-409`(창 안이면 owed=false, 재무장 없음) · `notifier.go:285-293`(ClaimSettled → 전송 없음) · `internal/execgw/retry.go:361-365,413-414` · 델타 ADDED `:306`「완화도 통지되어야 한다(SHALL)」 · 정본 risk-management `:137`·`:157` | key에 전이 신원(22판이 이미 레코드에 싣는 삽입 rowid)을 넣는다. a094 D−5.2 방향과 같다. RED: 1h 안의 재강화와 두 번째 완화가 각각 새 PENDING 행을 만든다 |
| 2 | **P1** | Q4 핀의 금지 형태 ③(「`entry == nil` 갈래 · 삽입을 감싼 조건이 잠금을 감싼 조건보다 넓음」)을 오늘의 `parkAlert`가 그대로 갖고 있다. `Block`은 `if g.entry != nil` 안에 있고 `EnqueueAlert`는 무조건 실행된다. `execgw.New`는 nil Entry를 거절하지 않는다. 그런데 tasks는 「양성 대조군: 오늘의 parkAlert는 통과」라고 적어 모순이다. D0.3g 4의 「B2 `:548` 안에서 `EnqueueAlert`(`:551`)」는 오독이다 — B2는 `if err != nil { payload = nil }`(`:548-550`)이고 `:551`은 그 밖이다. AST 추출물에 끝 좌표가 없어 「안에서」를 뒷받침하지 못한다 | `internal/execgw/replay.go:535-538,548-551` · `head-ast-21/internal-execgw-replay--gateway.parkalert.json`(B2 at 548, 끝 좌표 없음) · `internal/execgw/gateway.go`의 `New`(Journal·Trading·AccountRef만 검사) · `tasks.md:31-32` · `design.md:658-666` | `parkAlert` 편집 task를 더한다(Entry 필수화, 또는 nil이면 삽입도 하지 않음). 아니면 형태 ③을 다시 정의한다. FLM이 먼저다 |
| 3 | **P1** | 직접 기록자 census와 SHALL이 병행 change와 충돌한다. ① 「오늘은 parkAlert 하나」는 현재 HEAD에서 거짓이다. a066의 완화 통지가 인터페이스 경유 `repo.EnqueueAlert`로 critical 행을 넣는데, 먼저 잠그지 않고 `n.mu` 밖에서 넣는다. ② a094(D−4.6·D−5.3)와 a090은 exit 루프에 enqueue-only critical 기록자를 새로 둔다. 이들은 삽입 전에 자기 사유를 세우지 않는다. 적재가 실패하면 적재 **전**에 읽은 세대로 `BlockUnlessClearedSince`(조건부)만 하고 승격은 하지 않는다. 이것은 a092 델타 `:62`(직접 기록자는 삽입 전에 자기 사유)와 `:72`(durable 기록 실패는 그 자리에서 래치+승격)에 어긋난다. ③ 결과: `Acknowledge`의 셈(`:870`)과 해제(`:875`) 사이에 이 행들이 들어오면, PENDING critical 행이 남은 채로 게이트가 열린다(a124 판정이 다시 잠글 때까지). ④ 완화 통지는 「먼저 잠근다」가 성립할 수 없는 기록자라서 SHALL이 너무 넓다 | `internal/app/engine/risk_relaxation_command.go:40,158`(HEAD `90e5170d`) · `openspec/changes/a094-…/design.md:57-63,214-220` · `openspec/changes/a090-…/design.md:69-76` · `internal/obs/notifier.go:851-876` · a092 `design.md:666`「오늘은 parkAlert 하나」 | exit 루프 critical 기록의 **단일 입구**를 a092가 정의한다: `n.mu` 아래 `RecordAlert`, 기록자별 remindAfter 0 허용. SHALL을 「n.mu 아래 기록 **또는** 삽입 전 자기 사유」로 고친다. durable 실패는 무조건 래치하고, exit critical이면 승격한다. census는 메서드 이름·인터페이스 호출까지 센다. a090·a094·a066에 교차 통지한다 |
| 4 | P2 (T) | 「모든 발송자」 문단과 Scenario가 a124 정본의 허용 창(근거 확정과 세대 읽기 사이의 해제는 「앞」으로 봐도 된다)을 싣지 않았다. 문자 그대로는 지킬 수 없는 SHALL이다. a124의 「승격 쓰기 실패 → 무조건 차단」도 동기 발송자에게 옮겨지지 않았다 | 델타 `:64`,`:101-104` · 정본 `openspec/specs/engine-safety/spec.md:1463-1464,1471,1523` · `internal/app/engine/alertdelivery.go:409-418,437-439` · `notifier.go:385-390`(로그만) | 허용 창 문구와 AA1 조항을 옮긴다 |
| 5 | P2 | 사람 완화가 생기면 a124 「늦은 적용 = 제때 적용」의 승격 쪽이 새 인터리빙에서 거짓이 된다: 근거 확정 → 운영자 ack + `mode-release` → 늦은 escalate. 제때였다면 완화가 풀었을 승격이 완화 **뒤** AUTO 행으로 다시 선다(보수 방향, 창은 작다) | `internal/app/engine/alertdelivery.go:366,437,451` · `internal/journal/operating_mode.go:405-409`(방향 +1) | 이름 붙인 잔여로 적거나, 모드 해제 세대를 둔다 |
| 6 | P2 | C13: Notifier는 전송 실패를 오류로 돌려주지 않는다. 그래서 `ErrModeAnnouncementFailed`는 outbox 쓰기 실패 때만 난다. 전송 수단이 죽어 있으면 완화 통지가 소진된다. 그러면 `Notifier.escalate`가 **같은 호출 안에서** 방금 푼 모드를 CRITICAL_ALERT_UNDELIVERED로 다시 강화한다. 명령은 「성공」만 보고한다 | `notifier.go:124-129,223-230,382-383` · `operating_mode.go:478-481` · 델타 `:306` | 명령이 전이 뒤 모드와 게이트 사유를 다시 읽어 보이게 한다. RED를 더한다(C18은 절차로만 다룸) |
| 7 | P2 (T) | C15 (ㄴ)의 근거가 약하다. 같은 `buildGateway`의 앞 두 복원은 원장 읽기 실패에 기동을 거부하므로, 공통 원장 장애면 모드 복원 전에 이미 기동이 거부된다. (ㄴ)이 실제로 사는 경우는 모드 행 고유 오류(시각 파싱)뿐이다. 그 경우 「완화 명령으로 푼다」도 거짓이다 — `mode-release`도 `currentModeTx`에서 같은 오류를 만난다 | `internal/app/engine/gateway.go:261-263,269-271` · `internal/journal/operating_mode.go:398,705-709` | 풀 경로가 원장 수리뿐임을 적는다. 두 처분의 불일치를 이유와 함께 기록한다 |
| 8 | P2 (T) | AC2 창 허용이 정본 risk-management 「journal 영속과 **동시에** EntryGate 투영」과 모순은 아니다(HEAD도 커밋 뒤에 투영한다). 그러나 해석을 적지 않았다. 관측 두절은 자기 사유 래치 없이 승격만 하므로, 그 창의 진입 점검은 통과할 수 있다 | `openspec/specs/risk-management/spec.md:133` · `operating_mode.go:468-476` · `internal/app/engine/exitloop.go:830-852` · 델타 `:295` | 「동시에 = 같은 전이 호출 안」이라는 해석 문장을 넣는다 |
| 9 | P2 (T) | rowid 근거에 VACUUM 조건이 빠졌다. TEXT PK 테이블은 VACUUM/VACUUM INTO에서 rowid가 다시 매겨질 수 있다(순서는 보존). 울타리 초기값(행 0개 = rowid 0, 비교는 「보다 큰」)도 정의가 없다 | `internal/journal/backup.go:76` · `internal/journal/core_domain.go:183-193` | 문서화하고 초기값을 명시한다 |
| 10 | P2 | C8: `runAuxiliary`가 모든 보조 실행자의 정지를 `EventAlertUndelivered`로 기록한다. 주석이 「둘째 보조 실행자는 자기 이벤트 타입을 가져와야 한다」고 적는데 설계는 말이 없다. 종료 배수 때 버퍼에 남은 일반 알림의 버림 기록도 설계에 없다. exit-policy MODIFIED의 무조건 「캡 발생은 알림된다(SHALL)」과 「버퍼가 차면 버림 허용」이 긴장 관계다 | `internal/app/engine/auxiliary.go:99-109` · engine-safety 델타 `:48` · exit-policy 델타 `:28` | 이벤트 타입, 배수 때 기록, 한정어를 더한다 |
| 11 | P2 | `RecordAlert`는 exported다. `n.mu` 밖의 호출자가 쓰면 남 대신 정착 행을 재무장해 셈~해제 창에 PENDING 행을 만든다. `EnqueueAlert`가 0을 넘기는 이유가 바로 이것이다 | `internal/journal/outbox.go:142-146` | 비시험 호출자 census 핀을 둔다(Notifier 어댑터 하나) |
| 12 | P2 (T) | 기록 전용 announcer가 기록에 실패했을 때의 처분과, `n.mu`를 잡는지가 D0.3g 1에 없다. 오늘의 Notifier announcer는 claim 실패 시 래치+승격한다. 투영이 통지보다 앞(`:475`→`:479`)이라 셈~해제 창은 모드 사유가 덮는다 — 그것도 적어야 한다 | `notifier.go:219,262-282` · `operating_mode.go:475-479` · 델타 `:62`,`:72` | 명시한다 |
| 13 | P2 (T) | 좌표 오류 | 아래 판정 8 | 정정한다 |

## 반드시 판정할 것 1~8

**1. 제목 달성 — 조건부 달성.** 22판의 네 주입 지점을 바꾸면 exit goroutine에서 원격 전송을 동기로 기다리는 도달 경로는 0이다. 경로별로 셌다.
- **`Notify` 입구**: `exitloop.go:1710` 한 곳. 모든 ExitObserver 알림이 `o.alert`를 지난다(`exit_quarantine_announce.go` 포함).
- **Announcer 입구**: 관측 두절 `exitloop.go:846-847`.
- **Retrier**: `:747` → `retry.go:357-365` → `:413-414`. 주입은 `exitwiring.go:333`.
- **청산 상한 floor**: `exitwiring.go:207,231`. `gateway.go:350-353`에서 조립되고, 비시험 호출자는 `exitloop.go:1513` 하나다.
- **`n.mu` 대기**: 재구성 뒤 보유자는 로컬 작업뿐이다. `Flush`의 비시험 호출자는 0이다.
- **Submit·Issuer**: Notify·Escalate 호출이 0이다. `EscalateOperatingMode` 비시험 호출자 7곳을 전수했고, `riskguardian.go:644`는 `IssueEntry :436`만 부른다.
- **기록 실패 승격**: nil announcer다(`notifier.go:382-383`).

D0.3g 1의 측정은 참이다. `retry.go:305-328`은 설정값만 가진 필드이고 뮤텍스가 없으며, 메서드는 필드를 읽기만 한다(`:335-446`). 구현 주의점이 둘 있다. `Context.ExitObserver`가 `opts.Retrier`를 무조건 덮어쓰므로(`exitwiring.go:333`) 복사본은 그 안에서 만들어야 한다. `c.exitFloor`는 공유 포인터로 조립된다. a090·a094가 입구를 더하면 전수를 다시 해야 한다.

**2. 기록 경로 안 나 — 유지된다.** 재무장(`claimOwed :382-410`), 중복 제거(`recordAlertTx :289-346`), 임차 무접촉 모두 선다. PENDING 행은 읽기만 하고, 정착 행은 이미 임차가 비어 있다(`outbox.go:454,497`). `n.mu` 아래에서 부르면 셈~해제 배제도 유지된다. 새 원장 표면의 위험은 #11이고, key 결함(#1)이 이 래퍼 위에서도 그대로 남는다.

**3. 원칙 E — 승인 뒤 재잠금은 닫히고, 잘못 여는 경로는 새로 생기지 않는다.** 재잠금은 근거 확정과 세대 읽기 사이의 창을 빼면 닫힌다(#4). `BlockUnlessClearedSince`는 삽입하거나 아무것도 바꾸지 않는다(`retry.go:571-583`). `deliver`를 잠금 밖으로 옮겨도 셈~해제 사이에 끼는 것은 전달 기록뿐이라 수를 줄일 뿐이다. 게이트를 여는 쪽은 직접 기록자다(#3).

**4. 커밋 순서 하나 — 참.** 근거를 하나씩 확인했다.
- `operating_modes` 쓰기는 INSERT 한 곳뿐이다(`operating_mode.go:445`). DELETE·UPDATE는 0이다.
- 전이는 `_txlock=immediate`로 BEGIN IMMEDIATE다(`journal.go:216-225`).
- 원장 연결은 하나다(`journal.go:174`).
- `operating_modes`는 TEXT PK의 rowid 테이블이다(`core_domain.go:183-193`).
- 「현재」를 따로 읽는 곳은 없다. `CurrentOperatingMode`의 비시험 호출자는 복원 하나뿐이다.

AC2 문구는 정본과 모순이 아니고, 해석을 적어야 한다(#8).

**5. archive 정합성 — 결함 있음.** 새 MODIFIED 「배달 실행자의 정지가 …」는 정본과 근거 ①(과 머리 주석)만 다르다. diff로 확인했다. 다만 이 블록의 주인이 되면서 거짓 좌표와 a092 자기 기술을 다시 인증한다(#13). 「등급화된 알림」 MODIFIED에는 지킬 수 없는 SHALL이 둘 있다(#1 통지, #4 창). a124 정본 문장이 거짓이 되는 인터리빙도 생긴다(#5).

**6. Q4 핀과 C8.** Q4 핀의 형태 ①②는 선다. 형태 ③은 오늘의 `parkAlert`와 모순이고(#2), census는 불완전하다 — 인터페이스 경유 호출을 놓친다(#3). C8은 선다. 배달 실행자의 사이클과 분리한 판단은 옳다. 남은 것은 #10이다.

**7. a094와의 교차 — 충돌 있음.**
- #1: a092가 같은 결함 부류(에피소드 신원 없는 key)를 새로 활성화한다.
- #3: a090·a094·a066의 직접 기록자가 a092의 SHALL, Q4 핀, durable 실패 규칙과 충돌한다.
- 계약 차이: a092 어댑터는 remindAfter 1h로 재무장하고, a094·a090은 0과 에피소드 key를 쓴다. a094 D−5.6이 넘긴 `noteDelay`는 a092 어댑터를 지나므로 1h 재무장 계약이 된다. 이것은 a094의 「재시작은 새 에피소드가 아니다」와 다른 계약이라 어느 쪽인지 명시해야 한다.
- `outbox.go:142-146`·`:382-384` 자체는 a094의 서술대로다.

**8. 좌표·AST 인용.**

잘못된 것:
- `cmd/tossctl/engine.go:636` → 실제는 `:639`.
- `position_policy_command.go:94` → retrier 대입은 `:98`.
- `outbox.go:140-146` → 주석은 `:142-146`(`:140`은 `defer`).
- 「B2 안에서」 오독(#2).
- MODIFIED 블록의 `retry.go:498-505` → HEAD에서 그 줄은 `NewEntryGate`이고 `Block`은 `:532-539`.
- MODIFIED 블록의 `a092/specs/engine-safety/spec.md:126-130`「모드 승격 SHALL · 다섯 조항」 → 22판의 `:126-128`은 「일반 등급도 루프를 붙잡지 않는다」다.

일치하는 것:
- `deliver :484/:520/:571` · `BlockUnlessClearedSince :572/:574/:577` · `TransitionOperatingMode` B14 409 · B16 417 · B18 424 · B19 428 · B24 461 · B26 475 · B27 478
- `Retrier.Query` B7 357 · `modeLatestOrder :626` · `:388→:391` · `exitwiring.go:333` · `gateway.go:324,350-353` · `reconcileloop.go:360` · `adoption.go:257-258` · `exitloop.go:1513` · `mode.go:49-74` · `operating_mode.go:479-481`
- `engine_risk_relaxation.go:136,194`(`7cf80832` 기준)
- 파일 sha 6개가 MANIFEST와 같다.

**안전 불변식.** 손절 즉시성은 약해지지 않는다: RISK_REDUCING은 모든 모드에서 허용되고(`operating_mode.go:232-234`), 게이트는 청산을 막지 않는다(`retry.go:458-460`). C15 (ㄴ)은 엔진을 살린다. 토글 OFF 동작은 불변이다 — 게이트 OFF면 엔진이 뜨지 않는다. `mode-release`와 `alerts ack` 모두 `mutating: "true"`다. 설계상 LIVE side effect는 0이다.

**저장소 상태.** `rtk proxy git -C /mnt/D/Axipient/workspace/TossOS status --short -- openspec/changes/a092-an-alert-does-not-hold-the-stop` 출력:
```
?? openspec/changes/a092-an-alert-does-not-hold-the-stop/analysis/review-22/
```
저장소 파일은 만들거나 고치지 않았다. 스크래치는 `/tmp/claude-1000/a092-r22/voiceA-scratch/`에만 썼다.

Recommendation: BLOCK because exit goroutine 전송 대기 0과 손절 즉시성은 서지만, a092가 여는 완화 경로에서 모드 통지가 key 흡수로 조용히 유실되고(P1 #1), Q4 핀 형태 ③이 오늘의 parkAlert와 모순이며(P1 #2), a090·a094·a066의 직접 기록자와 a092의 셈~해제 SHALL·핀·durable 실패 규칙이 충돌해 PENDING critical 행이 있는 채로 게이트가 열리는 경로가 남는다(P1 #3) — 셋 다 구현 전에 문서와 교차 통지로 닫을 수 있다.
