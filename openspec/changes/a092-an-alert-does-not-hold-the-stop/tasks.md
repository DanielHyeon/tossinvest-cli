# a092 tasks (11판)

base: `ec29dc72c0fd589daa2069ccf26bad26baeb2a04` (`base-commit.txt`)

> **9판이 다르게 하는 한 가지**: **고친 줄마다 그 줄을 우회하는 케이스를 같은
> 커밋에 넣는다.** 8판은 방어를 돌려 보는 자체 시험을 만들고도 **자기가 고친 부분에는
> 그 규칙을 안 지켰다** — 케이스 8건이 전부 7라운드 지적의 재생이었고, 8판이 새로
> 넣은 세 구성은 한 건도 안 덮여서 그대로 8라운드 차단 B2·B3·B4가 됐다.
> 9판의 자체 시험은 **19건**이고 그중 11건이 8라운드가 실증한 우회다.
>
> 네 번째 차단(B1)은 형태가 더 오래됐다 — **로그 열거 정정이 요구 문단에만 착지하고
> 같은 파일 Scenario에는 안 갔다.** 9판은 그것을 고치고 **기계로 막는다**
> (`check_dwell_enumeration`): 체류 구성을 세는 줄은 다섯 항을 전부 담아야 한다.
>
> **8판이 다르게 한 것은 그대로 유지된다**: 만든 방어를 돌려 본다.
> 7라운드 차단 5건 중 **4건이 7판이 새로 만든 두 산출물**(`tools/check_values.py`·
> task 7.1.1) 안에 있었다. 형태가 하나였다 — 방어를 쓰고 실행하지 않았다.
> 스크립트는 자기가 잡는다고 적은 `28.9`를 못 잡았고, 프로브는 컴파일조차 안 됐다.
> 그래서 8판은 검사를 고치는 것과 **그 검사를 우회해 보는 것**을 같은 task로 묶는다
> (`tools/check_values_selftest.py`, 7.4.0). 7.1.1의 프로브는 **실제로 돌려서**
> 표의 여섯 칸을 이 기계에서 재현했다.
>
> 다섯 번째 차단은 성격이 달랐다 — **처음으로 산문이 아니라 기제가 틀렸다.**
> 컴파일 단언이 `reserve == 0`을 허용했고, 그것은 두 delta가 `SHALL NOT`으로
> 금지한 바로 그 상태다. 8.2가 단언을 **여섯으로** 만든다 — `- 1` 넷이 0과 음수를,
> `/time.Millisecond` 둘이 **단위 누락**을 잡는다(9라운드 H-1, 10판이 더하고
> 11판이 8.2로 내렸다).
>
> **7판이 다르게 한 것은 그대로 유지된다**: 정정의 단위를 "리뷰어가 준 file:line"에서
> **"값"**으로 바꾼다. 여섯 판이 각각 지목된 인스턴스만 고쳤고, 그래서 매번 아직
> 지목되지 않은 문서에서 같은 오류가 살아남았다(6라운드의 진단). 그 단위를 강제하는
> 것이 `tools/check_values.py`이고 **7.4.0이 그것을 돌린다.**
>
> **손으로 박아도 되는 값과 안 되는 값을 가른다.** 사람이 고른 상수(T·D·시도 수·예산)는
> 유도할 수 없으므로 스크립트 안에 **한 번** 선언하고 모든 문서를 거기 맞춘다.
> 셀 수 있는 것(산출물 수·표 행 수·비표준 절 수·등식의 값)은 **절대 선언하지 않고**
> 파일에서 유도해 서로 대조한다. 6판의 7.4.2가 기대값 16을 손으로 박았다가
> 자기가 절을 하나 더하는 순간 거짓이 된 것이 그 구별을 안 했기 때문이다.

## 1. 기억·계약 (완료)

- [x] 1.1 memory recall — `tossos-*` 기억 + `.claude/CLAUDE.md` 안전 불변식
- [x] 1.2 OpenSpec 진입 — change 디렉터리·`base-commit.txt`·PM 등록
- [x] 1.3 `openspec validate --strict`

## 2. Hard evidence (완료 — 4판에서 재측정)

- [x] 2.1 `Notify` 도달 경로 전수 → `analysis/notify-reach.md` (직접 7 + Announcer 4),
      **각 경로의 루프 주기와 래치 포함**
- [x] 2.2 **측정 창 정정** — 배선 **커밋 날짜**가 아니라 **프로세스 재시작**이다.
      engine.log line 6837(08-05 00:11)이 `no notification publisher is configured`
      (`notifier.go:253`)를 쓴다 ⇒ 그때까지 **돌던 프로세스**에 publisher가 없었다.
      창 = **line 6866 이후**(2026-08-05T00:36:18Z)
- [x] 2.3 **`.Notify(`에 닿는 이벤트 16종 열거, 프로덕션 publish 12종.**
      3판은 14종, 4판 초안은 15종이었다 — `exit_quarantine_announce.go:71`과
      `liquidate.go:238`(`flatten.complete`)이 차례로 빠졌다.
      그리고 **flatten 4종은 프로덕션에서 `Notifier`가 nil**이라 로그 전용이다
      (`cmd/tossctl/flatten.go:247`, `Saga.event`의 nil 가지) — `notify-reach.md` P7이
      이미 적어 둔 사실이다. 로그 전용 생산자를 가진 **4종**에 판별자 확정
      (`from_state` / `stopped_with` / INFO / WARN)
- [x] 2.4 **냉연결 비율 재산출** — 창(line 6866~8353) 안 publish 유발 줄
      **39개 중 냉 18개(46%)**. **창의 끝을 고정해야 재현된다**(4라운드 M5).
      3판의 47/24(51%)는 창과 열거가 둘 다 틀렸다. **표본 6건은 값이 그대로**이고
      전부 창 안이다. 냉 표본 **n=1(0.754s)**
- [x] 2.5 조립 지점 확정 — `exitwiring.go`의 `newNotifier`,
      `notifications.go`의 `resolveNotificationPublisher` 안 `&obs.Ntfy{...}` 리터럴,
      범위 밖 `cmd/tossctl/notificationsettings.go:151`
      (`publishNotificationTest`의 **함수 지역 변수** — 값으로 도달 불가)
- [x] 2.6 운영 원장 읽기 전용 — `alert_outbox` PENDING 9 / DELIVERED 3,
      `operating_modes` **1행**(`ENTRY_BLOCKED` 2026-07-31, 미완화)
- [x] 2.7 `AnnounceOperatingMode` 발동 이력 — **전체 로그에 1회**(line 372, 2026-07-31).
      창 안 0회 ⇒ P1a는 오늘 발동하지 않는다
- [x] 2.8 **채택 상수 국소 실측** — `go test -overlay`(저장소 무편집).
      **회차·조건·수치의 정본은 `analysis/delivery-latency.md` §7.2**이고 여기에
      옮겨 적지 않는다. 6판은 이 자리에 후보 #2 회차의 조건("10회 · `GOMAXPROCS=2`")과
      어떤 측정 표에도 없는 `28.9ms`를 적었다(6라운드 차단 B2) — **복제가 원인이므로 <!-- rejected-value -->
      복제를 없앤다.** 값이 §7.2와 같은지는 7.4.0이 기계로 본다
- [x] 2.9 **`n.mu` 배수의 대조군 실측** — engine-safety `:28`의 SHALL이 인용하는
      **8.458초**를 만든 절차다(6라운드 M-5: SHALL의 근거가 tasks에 재현 절차 없이
      존재했다). 프로브는 2.8과 같은 모양에 goroutine 둘이 같은 `*obs.Notifier`로
      critical을 올리는 것을 더한다. **채택 상수와 후보 #2를 둘 다 돌린다** — 후자가
      대조군이고, 4판이 실었던 9.231초가 채택하지 않은 상수의 값임을 그것이 보인다. <!-- rejected-value -->
      결과는 `analysis/delivery-latency.md` §7.4가 갖는다

## 3. Function Logic Map (완료 — 문서보다 먼저)

- [x] 3.1 함수 **23개** AST + FLM + Branch Test Map + risk report
- [x] 3.2 `python3 tools/logic-map/check_analysis.py --change a092-...` PASS
- [x] 3.3 1판 FLM의 "6종" 오류 정정 (`exitobserver.alert`)
- [x] 3.4 `notifier.flush` 산출물 제거 — a093으로 이관
- [x] 3.5 **3라운드 H2(b) — 분기를 근거로 쓰면서 산출물이 없던 함수 7개 추가**:
      `Runtime.CheckHealth` · `Runtime.escalate` · `Runtime.takeLatch` ·
      `ExitObserver.alertProposalRefused` · `ExitObserver.alertRefused` ·
      `ExitObserver.announceQuarantine` · `Journal.TransitionOperatingMode`.
      **`alertProposalRefused`의 `"branches": null`·`"returns": null`이 "억제가 없다"는
      부재 주장을 처음으로 열거로 증명한다**
- [x] 3.6 **4라운드 H4 — `TargetModeForTrigger` 산출물 신설**(23번째).
      3라운드 H2(b)와 같은 형태가 재발한 자리다: proposal H1의 첫 고리가
      "여섯 트리거가 전부 같은 목표 모드로 간다"는 **전칭**인데 산출물이 없었다.
      AST는 branches 3(B1 switch `:538` · B2 case `:539` · B3 default `:546`) ·
      returns 2(`:545`·`:547`)이고, **값을 돌려주는 case가 B2 하나**여서
      전칭이 분기 구조로 증명된다.
- [x] 3.7 **5라운드 H2 — `resolveNotificationPublisher` 산출물 정정.**
      `function-logic-map.md`와 `branch-test-map.md`가 nil 이탈을 **B1·B3·B5**로
      적고 B1 `:69`를 `cfg.Rejected != ""`라고 썼다. `ast.json`(`B1@69`·`B2@77`,
      returns `81/87/97/105`)과 소스(`:69`는 `getenv == nil`, 반환 없음)가 둘 다 반증한다.
      **4라운드 M1의 정정이 design·tasks에만 착지하고 산출물에는 안 착지한 것**이고,
      그래서 7.4의 기계적 검사를 신설했다.
- [x] 3.8 **5라운드 H3 — `publishBestEffort`·`ntfy.Publish` 산출물 정정.**
      두 산출물이 줄 간격 표본(최대 **1.836초**, 채택 상한 1.3초 초과)을 "왕복 실측"이라
      불렀는데, 같은 change의 `delivery-latency.md` §5.3이 그 방법을 이미 무효로
      판정하고 있었다(조립 자리 4곳 · 셋이 대사 루프). 6판이 산출물 안에서 무효임을
      표시하고, **1.3초를 넘는다는 사실 자체는 design D2가 정면으로 다룬다**.

## 4. 문서 (완료 — 4판)

- [x] 4.1 `proposal.md` — 34s 헤드라인, 68s는 NORMAL 계정 상한, 창 정정, 39/18,
      `alertProposalRefused` AST 근거, 새 결합(빌드 결합·`o.Interval()` drift) 기록
- [x] 4.2 `design.md` — D2에 **reserve 개념과 실측 10회**, `alertRetryDelay` 유도(3라운드 M1),
      D3에 **컴파일 단언 + "테스트로는 등식을 지킬 수 없다"**, D6에 SHALL 범위 한정
      (4판의 단언은 **한쪽 1줄**이었고 그것이 7라운드 B1이다 — 8판이 넷으로 만든다: 4.10)
- [x] 4.3 `specs/exit-policy/spec.md` — 체류에 **비전송 작업 포함** 명시,
      예산 < 주기를 SHALL NOT로
- [x] 4.4 `specs/engine-safety/spec.md` — 조립 SHALL **범위 한정**(H1),
      감독자 면제 근거를 **래치 하나로**(H2(a)), "동기 체류의 합" → **알림 하나**(C2)
- [x] 4.5 컴파일 타임 단언 실측 — reserve 800ms 통과, 음수면
      `invalid array length alertOverheadReserve` 로 **상수 이름을 말하며** 실패.
      **이 실측은 음수만 봤고 `0`은 안 봤다** — 그것이 7라운드 B1이다.
      8판의 전수 실측은 4.10
- [x] 4.6 `openspec validate a092-... --strict`
- [x] 4.7 **`tools/check_values.py` 신설 (7판의 첫 산출물)** — 값 단위 교차문서 검사.
      6라운드가 확정한 형태("정정이 지목된 문서에만 착지한다")를 기계로 막는다.
      실행 자리는 7.4.0. 첫 실행이 6라운드 차단 2건을 재현했고 **리뷰가 못 찾은 것을
      둘 더 냈다** — `design.md`의 `9.2s`(실제 9.3s, 게다가 그 경우 reserve는 음수라  <!-- not-a-measurement: 이력 서술 — 6판이 적었던 오기와 그 정정값이며 측정이 아니다 -->
      문장이 자기모순이었다)와 **`exitobserver.alert` FLM에 있던 B3의 두 번째 사본**
- [x] 4.8 **6라운드 12개 좌표 수정 + 값 단위 훑기** — 판정과 조치의 정본은 `review.md`
      「a092 7판」절이다. **proposal에서 수치 복제를 없앤 것이 이 판의 구조 변경**이고,
      차단 2건이 그 복제에서 나왔다

### 8판이 더한 것 — 방어를 돌려 본다

- [x] 4.9 **reserve 재측정 — 로거를 달고 (7라운드 H2).** `newNotifier`의 4번째 인자에
      `nil`을 넘겨 소진 경로의 구조화 로그 세 줄을 **구조적으로 제거한 채** 쟀던 것을
      다시 잰다. 로거를 유일한 변수로 두고 셀 5개 × 5회, `-race`·`GOMAXPROCS=60`.
      결과는 `analysis/delivery-latency.md` §7.5 — **로그 쓰기는 측정 한계 아래이고**
      (로거 없는 셀의 최악 319.3 ms > 로거 단 셀의 최악 209.6 ms), 그러므로
      **800ms는 안 바뀌고 바뀌는 것은 열거다.** D3 주석과 두 delta의 SHALL 열거에
      로그 쓰기를 더했다
- [x] 4.10 **컴파일 단언 재설계 + 실측 (7라운드 B1·H1).**
      단언을 넷으로 만들고 전부 `- 1`을 붙인다.  <!-- 8판 시점의 기록이다. 10판이 차원 단언 둘을 더해 현재는 여섯이고, 정본은 8.2다 -->
      **10판이 여기에 둘을 더했다**(9라운드 H-1 — 나노초를 세는 넷은 1300ms와
      1300ns를 구별하지 못한다). 현재 정본은 **8.2의 여섯 줄**이다. `go build -overlay`로 이 패키지에
      design D3 표의 구성들을 넣어 돌렸고, 채택 #3·`reserve = 1ns`·M-4 대안은 BUILD OK,
      후보 #1(reserve 0)·음수·세 상수의 0은 전부 FAIL이며 **메시지가 각각 자기 상수
      이름을 말한다.** 표는 design D3
- [x] 4.11 **M-4에 측정으로 답한다.** D2 후보 표에 #7(D=50ms, reserve 1000ms)을 더하고,
      §7.5 셀 E로 **초과분이 D에 의존하지 않음**을 보인 뒤 #3을 유지한다.
      기각이 아니라 **대기**이고 재검토 조건(초과분 400ms 초과 관측)을 명시했다
- [x] 4.12 **`tools/check_values_selftest.py` 신설 — 8판의 첫 산출물.**
      7라운드 차단 4건의 형태가 "방어를 쓰고 돌려 보지 않았다"였다. 이 시험은 change
      디렉터리를 임시 사본으로 복제하고 **7라운드가 실증한 우회 8가지를 각각 주입해
      검사가 실패하는 것을 본다.** 실행 자리는 7.4.0.
      첫 실행이 **`proposal.md`에 살아 있던 `28.9` 인용을 즉시 잡았다** —
      7판의 스크립트가 exit 0으로 통과시키던 자리다
- [x] 4.13 **7.1.1 프로브를 실제로 돌렸다 (7라운드 B3·B4·B5·H3).**
      프로브 본체를 새 파일로 옮겨 import 충돌을 없애고, 모듈 경로를 고치고, `-v`를 붙이고,
      마커에 함수 이름을 붙였다. 이 기계에서 표의 여섯 칸을 재현했다 —
      `68/80/82`(태그 없음) · `72/86/88`(태그판).
      **8라운드 보이스 B가 이 명령을 문서 그대로 복사해 같은 여섯 칸을 재현했다**

### 9판이 더한 것 — 고친 줄마다 우회 케이스를 같은 커밋에

- [x] 4.14 **`check_values.py`의 8판 신규 구성 세 개를 고친다**(8라운드 B2·B3·B4).
      (a) `MEASURE`의 부정 룩비하인드를 없앤다 — 그것은 면제가 아니라 **첫 자리
      삭제**였다(`≥ 28.9 ms` → `8.9`). 왼쪽 앵커는 `(?<![\d.])`이고 부등호 판정은 <!-- rejected-value -->
      매치 **뒤에** 한다. (b) `COUNTERFACTUAL` 낱말 목록을 **명시 마커**로 바꾼다 —
      `후보·기각·만약`은 D2 산문에 상시 등장해 면제가 사실상 기본값이었다.
      (c) 기각값 조회를 `LATENCY_WORDS`/`OUT_OF_SCOPE` 게이트 **밖으로** 빼
      별도 검사로 만든다 — `review.md` 8.1이 "코퍼스와 무관하게 항상 잡고"라고 쓴
      문장이 그 시점에 거짓이었다
- [x] 4.15 **정수 측정치와 절 번호를 본다**(8라운드 M-7). `MEASURE_INT`는 ms 계열
      정수를 잡고, 코퍼스는 `§7.5`와 **`§` 없는 마크다운 절 제목**의 번호를 뺀다.
      **첫 실행이 `tasks.md:511`의 `521ms`를 잡았다** — 3라운드가 잰 값인데
      `analysis/` 어디에도 없던 **여덟 판을 살아남은 고아**다. 측정 정본 §7.1에 실었다
- [x] 4.16 **명명값 네 개 추가**(8라운드 B4·M-5·M-8). §7.5의 세 최악 초과분과
      냉 publish 측정 0.754s. 8판까지 **대표 산출물의 수치가 어느 검사에도 안 걸려**
      design과 analysis가 서로 다른 값을 말해도 통과했다. 산포 배수는
      `11`·`11.2`·`10` 세 표기가 살아 있었고 검사가 `11`만 읽었다 — **한 표기로 통일**
- [x] 4.17 **후보 표의 값은 선언하지 않고 유도한다**(`candidate_tokens()`).
      정수 검사를 켜니 후보 #2의 reserve `400ms`가 고아로 잡혔는데, 그것은 D2 표에서
      **읽을 수 있는 값**이다. 읽을 수 있는 것을 선언하면 표가 바뀔 때 거짓이 된다 —
      이 스크립트 머리말의 구별 그대로
- [x] 4.18 **`check_dwell_enumeration` 신설**(8라운드 B1). 8판은 두 delta의 요구
      문단에 "구조화 로그 줄"을 더하고 **"이 열거는 완전해야 한다(SHALL)"까지 새로
      걸어 놓고** 같은 파일의 Scenario는 안 고쳤다. 열거를 의미로 검사할 수는 없으므로
      **구조**를 본다 — `outbox`와 `승격`을 함께 말하는 줄은 다섯 항을 전부 담아야 한다
- [x] 4.19 **자체 시험 8건 → 19건.** 새 11건은 전부 **8라운드가 실증한 우회**이고,
      추가 전에 **RED를 확인했다**(9/9 통과 = 검사가 못 잡음, 그중 1건은 "실패는 했으나
      이유가 다르다"로 유령 숫자까지 드러났다). 고친 뒤 19/19
- [x] 4.20 **§7.5 프로브를 change 안에 싣는다**(8라운드 H3). 8판은 이 측정으로 두
      delta의 SHALL 열거를 바꿔 놓고 프로브를 안 실어 **아무도 재검증할 수 없었다.**
      §7.5.1에 전문과 명령을 넣고, "로그 줄 3"이 회차마다 3인 이유
      (`openTestJournal`가 회차마다 새 DB → `changed` 항상 참)와 **프로덕션에서는 2줄**
      이라는 것을 함께 적었다
- [x] 4.21 **§7.5 결론 2의 추론을 데이터가 지는 만큼으로 줄인다**(8라운드 H4).
      n=5의 max 순서 역전은 "효과가 없다"가 아니라 **"이 설계로는 안 잡힌다"**이다.
      세 줄의 크기는 **미측정으로 남긴다.** 하중을 지는 것은 결론 하나(319.3 < 356.1
      이므로 유도 규칙의 입력이 안 바뀐다)뿐이고 **그것은 옳다**
- [x] 4.22 **D2가 사이클 총합을 transport로 곱한 것을 고친다**(8라운드 H2).
      `4.2s × N` → 알림당 상한은 `alertBudget`이다. 같은 절 첫머리가 "예산이 걸리는
      것은 전송이 아니라 **호출**"이라고 3판을 기각한 근거 그대로이고,
      `exit-policy` 델타는 옳게 "관측 주기"라고 쓴다 — **유도 정본이 spec보다
      낙관적이면 유도 정본이 틀린 것이다**
- [x] 4.23 **인용 정정 5건.** D3 Go 주석의 0-폴백 두 건(`notifier.go:291`→`:292`,
      `ntfy.go:95`→`:96` — 선언 줄 vs 분기 줄, **프로덕션 소스에 실릴 주석**) ·
      `notifier.go:187`(`escalate` 호출)과 `:217`(journal 호출) 구분 ·
      `TargetModeForTrigger` `:537-548`→**`:537-549`**(**4개 문서**) ·
      `Run:353-362`→**`:353-363`**. 넷 다 `ast.json`의 `end`로 확인했다
- [x] 4.24 **D3 메시지 표의 행 라벨과 타입을 맞춘다**(8라운드 H5).
      `alertRetryDelay = 0`은 **무타입**이라 `untyped int constant -1`이고
      `0 * time.Millisecond`라야 `"time".Duration`이다. 두 행으로 갈라 싣고
      잘라낸 두 행(`…`)도 전문으로 복원했다. **어느 쪽으로 쓰든 FAIL이므로 방어는
      그대로 성립한다** — 어긋난 것은 문자열뿐이다
- [x] 4.25 **M-4 판정 규칙을 원칙으로 선언한다**(8라운드 M-11). "규칙이 만족되면
      더 옮기지 않는다"가 "하드한 쪽에 여유를 준다"보다 **우선한다**는 순서를 D2에
      적었다. 후자가 앞서면 규칙이 값을 고정하지 못한다 — 언제든 "더 하드한 쪽"을
      찾을 수 있기 때문이다

### 11판이 고친 것 (10라운드 차단 4건 + H1 + M)

> **순서를 뒤집었다.** 10판까지는 "고치고 나서 무엇을 고쳤는지 적었다". 11판은
> **축 1을 먼저 고쳐 무엇이 안 시험되고 있는지 기계에게 물었고**, 그것이 이름을
> 부른 자리에 케이스를 넣었다. 아래 4.26의 미커버 3건은 사람이 고른 것이 아니다.

- [x] 4.26 **축 1의 계수 단위를 줄에서 *경로*로 바꿨다**(10라운드 차단 4).
      10판의 shim은 `stack()[1].lineno`, 즉 **직전** 호출자만 봤다. 래퍼를 거치는
      호출은 `fail()` 줄이 같으므로 호출자들이 라벨 하나로 붕괴했다. 래퍼와 그
      호출자는 **AST로 열거한다** — 래퍼인지 아닌지는 호출 그래프의 성질이지 줄의
      성질이 아니고, 10판의 줄 훑기로는 원리적으로 볼 수 없었다.
      **고치자마자 미커버 3건이 이름을 갖고 나왔다**:
      `claims_adopted`(`fail@:318`)의 호출자 `:366`(뺄셈 등식)와 `:386`(한 항 등식),
      그리고 `_dwell_fail`(`fail@:873`)의 호출자 `:895`(요구 문단).
      **좌표는 11라운드에 재열거해 고쳤다** — 처음 적은 값은 그 게이트 실행
      시점의 것이고 그 뒤 같은 파일을 계속 고쳐 밀려 있었다(11라운드 C2).
      셋 다 자체 시험 케이스로 닫았다.

      함께 고친 둘: **`want` 비교**를 더해 "잡기를 그만둔 케이스"를 말할 수 있을
      때만 죽은 것으로 센다(10판은 `rc != 0`이면 크래시도 커버리지로 셌다).
      그리고 **라벨을 원본 좌표로** 낸다 — shim이 정의를 늘려 사본의 줄이 밀리므로
      10판의 진단은 실제 파일에 **없는** 줄을 가리켰다.
- [x] 4.27 **`coverage_gate.py`를 task로 배선했다**(10라운드 차단 2).
      7.4.0.1(구현 전)과 9.6(구현 후) 두 자리다. `Makefile`·`docs/WORKFLOW.md`에는
      **넣지 않는다** — change-local 도구는 archive와 함께 사라진다. 그 사유를
      7.4.0.1에 적었다(not-applicable).
- [x] 4.28 **10라운드 우회 36건 중 11건을 닫고 나머지를 계열로 선언했다**(차단 3).
      닫은 것: 요구 열거의 자리를 **앵커로 선언**(B14·B16 — 트리거는 발견용이고
      발견은 축소에 약하다), `target_files()`를 목록에서 **범위**로(B27·B28),
      단위 철자 확장(B08·B22), 보이지 않는 문자·합자·소수점 쉼표 접기(B09~B12),
      `not-a-measurement` 빈 사유 금지(B24), 비율의 산술 평가(B26).
      **B26을 고치자 10판이 한 번도 안 본 두 번째 비율**(`47개 중 24개(51%)`)이
      드러났다 — 정규식에 `46%`를 리터럴로 박아 두면 46%가 아닌 비율은 존재하지
      않는 것과 같다.

      닫지 **못하는** 것은 `coverage_gate.py`의 `UNSEEABLE`에 계열 U1~U7로 적었고
      **매 실행마다 출력된다**. 두 축이 세는 것은 값의 일치이고 문장의 참이 아니다 —
      통과 줄만 읽고 "전부 덮었다"로 오독하는 것이 7·8·9판의 형태였다.
- [x] 4.29 **차원 단언 두 줄을 8.2와 proposal Impact로 내렸다**(10라운드 차단 1).
      10판은 `design.md`에만 넣었고, 그래서 8.2를 적힌 대로 실행하면 그 판이 스스로
      "넷보다 나쁘다"고 선언한 실패 모드가 살아남았다. **값의 사본을 전부 고쳤다** —
      8.2 본문·`tasks.md` 머리말·4.10의 역사 기록·`proposal.md` Impact.
- [x] 4.30 **9라운드 H-1의 네 번째 구성에 답했다**(10라운드 H1).
      「`alertBudget`을 주기에서 분리」는 **컴파일 타임에 잡을 수 없다** — 배열 단언은
      상수의 값을 보지 그 값이 어디서 왔는지를 보지 않고, 실측으로 확인했다(여섯
      단언을 전부 걸어도 리터럴 `5 * time.Second`는 주기 1초에서도 BUILD OK).
      그래서 9.7에 grep 검사로 두고 **not-applicable 사유를 적었다.**
      이것이 proposal의 "구조적 보증"의 정확한 경계다: 값의 **관계**는 컴파일러가,
      값의 **출처**는 사람이 지킨다.
- [x] 4.31 **손으로 박은 수와 목록을 없앴다**(10라운드 M3·M7·M8).
      `EXEMPT` 사유 하나가 거짓이고 하나가 축 오귀속이었다(축 2에 「냉 측정 변경」
      칸은 없다 — 그 커버리지는 축 1에 있다). 자체 시험 주석의 「14건」·「134칸」과
      7.4.0의 「8가지」를 **수를 적지 않는 문장**으로 바꿨다 — 두 도구가 자기 수를
      세어 출력한다. 9.2의 줄 갱신 목록은 손으로 센 세 항목에서 **grep 명령**으로
      바꿨다(10판의 목록은 `tasks.md` 자신을 빠뜨렸고, 그래서 그 절이 지시한
      7.4.1 재검사가 반드시 실패했다).

      > **`19건`은 고치지 않았다.** 보이스 D가 현재 주장으로 읽었지만 원문은
      > 「**9판의** 자체 시험은 19건」으로 시점이 한정된 역사 기록이라 참이다.
      > not-applicable — 고칠 것이 없다.

## 5. proposal-freeze 재리뷰 (게이트)

- [ ] 5.1 적대적 Eng 리뷰 (High-risk 필수) — 두 보이스 독립.
      **교차 모델로 돌린다** — 6라운드는 codex 사용 한도로 두 보이스가 모두
      Claude였고(문맥은 독립, 모델 편향은 겹침) 그것을 그 라운드의 약점으로 기록했다.

      **7라운드도 못 지켰다 — 한도가 아직 안 풀렸다.** 보이스를 셋으로 늘리고
      세 번째만 다른 Claude 모델(Sonnet)에 걸었다. **이것은 완화이지 충족이 아니다.**
      8라운드에서도 못 지키면 그 사실을 판정문에 **또** 적는다.

      **8라운드도 못 지켰다 — 이번에는 한도가 아니라 사용자의 명시적 지시다**
      ("코덱스 쓰지말고 Claude 보이스로 돌려"). 이유가 달라도 결손은 같으므로
      **세 라운드 연속 미충족으로 기록한다.** 9라운드에서도 못 지키면 또 적는다.

      **9라운드는 codex 한도가 풀린 뒤에 도는 것이 이 요구를 실제로 지키는 길이다.**

      **9라운드도 못 지켰다 — 8라운드와 같은 이유(사용자의 명시적 지시,
      "Claude 보이스")다. 네 라운드 연속 미충족으로 기록한다.** 보이스 셋을
      서로 다른 렌즈(재현·적대·인용정합)로 돌렸고 셋 다 FAIL을 냈으며 서로
      다른 계열을 찾았다 — 차단 1은 두 보이스가 독립 발견했다. **완화이지
      충족이 아니다.** 10라운드에서도 못 지키면 또 적는다.

      **10라운드도 못 지켰다 — 8·9라운드와 같은 이유(사용자의 명시적 지시,
      "10라운드를 지금 Claude 보이스로 돌리세요")다. 다섯 라운드 연속
      미충족으로 기록한다.** 보이스를 넷으로 늘려 서로 다른 렌즈(계수기 감사·
      적대적 우회·소스 대조·구현 가능성)로 돌렸고 **넷 다 FAIL**을 냈으며,
      중심 차단은 **세 경로가 독립적으로** 도달했다. **완화이지 충족이 아니다** —
      같은 모델의 사각이 이번에 실제로 드러났다: 보이스 A가 검사 소스를 읽고
      낸 차단 하나(`EXEMPT` 사유 셋이 거짓)가 소스 대조에서 절반 틀려
      MEDIUM으로 내려갔다(`review.md` 10.10 M3). 11라운드에서도 못 지키면 또 적는다.

      > **10라운드 판정 FAIL — 차단 4건**(`review.md` 10.8). ① 차원 단언 2줄이
      > `design.md`에만 있고 `tasks.md` §8.2·`proposal.md`에 안 내려왔다 —
      > §8.2를 적힌 대로 실행하면 `alertPublishTimeout = 1300`이 **BUILD OK**로
      > 통과하고 `1.3µs`가 된다. ② `coverage_gate.py`를 실행하는 task가
      > **하나도 없다**(`tasks.md`·`Makefile`·`docs/WORKFLOW.md` 전부 0매치).
      > ③ 새 우회 36건이 세 스크립트를 **전부** 통과한다 — 9판의 34건은 실제로
      > 다 막히지만, 이번 절반은 검사가 **정의상 안 보는 자리**에 있다.
      > ④ 축 1이 줄 번호로 죽여서 `_dwell_fail`의 두 호출자가 라벨 하나로
      > 붕괴한다. **10판의 진전(세는 도구를 먼저 만든 것)과 그 산술의 정직성은
      > 확인됐다** — 실패한 것은 계수 단위와 배선이다.

      **11라운드도 못 지켰다 — 8·9·10라운드와 같은 이유(사용자의 명시적 지시,
      "11라운드를 Claude 보이스")다. 여섯 라운드 연속 미충족으로 기록한다.**
      보이스 넷(계수기 감사·적대적 우회·문서 좌표 정합·게이트 배선)이 전부 FAIL을
      냈고, 계수기 자신의 구멍에 **세 보이스가 서로 다른 방법으로 수렴**했다
      (A는 원리, C는 코드 읽기, B는 무기화). 문맥 독립성은 작동했다. 그러나
      **모델이 구조적으로 못 보는 것은 여섯 라운드째 한 번도 시험되지 않았다** —
      완화이지 충족이 아니다. 12라운드에서도 못 지키면 또 적는다.

      > **11라운드 판정 FAIL — 차단 6건**(`review.md` 11.8~11.13).
      > ① 안전에 직접 닿는 위조 11건을 한 문서에 동시에 적용해도 세 도구가
      > **전부 rc=0**이다(`alertPublishTimeout` 2.0s · 시도 2회 · 관측 주기 10초 ·
      > 체류 상한을 critical 전용으로 반전 · 기각값 9,231s와 28.9ms를 reserve의 <!-- rejected-value -->
      > 근거로 부활 · 체류 열거를 5항에서 2항으로).
      > ② `DWELL_ANCHORS`의 앵커는 문서 안의 문자열이라 미끼 문단으로 옮기면
      > 무효이고, **앵커 실패 진단이 시키는 정정을 그대로 따르면 초록이 된다** —
      > 방어가 아니라 유지보수 지시로 위장한 우회 경로다.
      > ③ 채택 상수를 등식 밖 산문으로 거짓 기재하면 검사가 없다. 통과 여부가
      > **후보 표에 그 수가 있느냐**에 달려 있어, 기각된 구성의 상수로 채택값을
      > 다시 쓰면 항상 통과한다.
      > ④ `skip = {"review.md"}`가 파일 **이름**을 보므로 `analysis/review.md`는
      > 모든 검사 밖이면서 동시에 측정 코퍼스다(`__pycache__`도 같다).
      > ⑤ `coverage_gate.py`가 자기 밖을 못 본다 — `FAILURES.append` 직행은
      > `fail_paths()`의 AST 열거에 안 잡히고(A-차단1 = B-G2), `blob_at`의 5줄 창은
      > 남의 면제 사유를 물려받는다(C3 = B-G1, 실측으로 지금 새고 있음을 확인).
      > ⑥ 면제 사유 `:1216`이 지목한 자체 시험 케이스가 그 자리를 지나지 않는다 —
      > 축 1이 같은 실행에서 `미커버 lost=0`을 낸다. 면제표가 스스로 반증된다.
      > **11판의 진전(계수 단위를 줄에서 경로로 바꾼 것, 그 게이트가 11판 자신을
      > 세 번 되돌려 보낸 것)은 확인됐다** — 실패한 것은 계수기의 경계와
      > 검사가 정의상 안 보는 자리다. **프로덕션 Go는 열한 판째 0줄이다.**

      > **9라운드 판정 FAIL — 차단 3계열**(`review.md` 9.5). 보이스 셋이 독립
      > FAIL이었고 차단 1(우회 34건이 전부 통과)은 두 보이스가 독립 발견했다.

      > **8라운드 판정 FAIL — 차단 4건**(`review.md` 8라운드 절). 차단 **3건이
      > `tools/check_values.py`의 8판 신규 구성**이고 자체 시험 8건 중 그것을 덮는
      > 케이스가 하나도 없었다. 나머지 1건은 **H2 정정이 요구 문단에만 착지하고
      > 같은 파일 Scenario에는 안 간 것**이다. 판정 세부는 보이스 B(구현 가능성)와
      > C(소스 대조)가 PASS이고, **설계·기제·실행 사슬은 세 보이스가 독립 재현했다** —
      > 실패한 것은 검사와 열거이지 설계가 아니다.

      > **7라운드 판정 FAIL — 차단 5건**(`review.md` 7라운드 절). 차단 4건이
      > **7판이 새로 만든 두 산출물**(`tools/check_values.py`·task 7.1.1) 안에 있었다.
      > 방어를 만들면서 그 방어를 돌려 보지 않은 것이 이 라운드의 형태다.
- [ ] 5.2 `review.md` 갱신 — 판정 · **냉 표본 n=1이라는 한계** · 손실 ② · reserve의 로컬성
- [ ] 5.3 gstack 리뷰 원장에 판정 기록

      **리뷰 원장 (라운드별 판정 · `review.md`가 정본)**

      | 라운드 | 보이스 | 모델 | 판정 | 차단 |
      |---|---|---|---|---|
      | 6 | 2 | Claude (codex 한도) | FAIL | — |
      | 7 | 3 | Claude ×2 + Sonnet | FAIL | 5 (4건이 7판 신규 산출물 안) |
      | 8 | 3 | Claude (사용자 지시) | FAIL | 4 (3건이 8판 신규 검사 구성) |
      | 9 | 3 | Claude (사용자 지시) | FAIL | 3계열 (우회 34건 전부 통과) |
      | 10 | 4 | Claude (사용자 지시) | **FAIL** | **4** (차원 단언 미하달 · `coverage_gate` 미배선 · 새 우회 36건 · 축 1 계수 단위) |
      | 11 | 4 | Claude (사용자 지시) | **FAIL** | **6** (위조 11건 번들 통과 · 앵커가 우회 경로 · 산문 상수 미검사 · `analysis/review.md` 세탁 · 계수기가 자기 밖 못 봄 · 면제 사유 자기반증) |

      10라운드 렌즈: 계수기 감사(A) · 적대적 우회(B) · 소스 대조·실측 재현(C) ·
      구현 가능성(D). **넷 다 독립 FAIL.** 중심 차단(차원 단언)은 D·C·판정자가
      서로 다른 세 방법으로 도달했다. 보이스 A의 차단 1건은 검증에서
      MEDIUM으로 내려갔다(`review.md` 10.10 M3) — 승격 전 재검증이 실제로 작동한 사례다.

      11라운드 렌즈: 계수기 감사(A) · 적대적 우회(B) · 문서 좌표 정합(C) ·
      게이트 배선·실행(D). **넷 다 독립 FAIL.** 중심 차단(계수기가 자기 밖을 못
      본다)에 **세 보이스가 서로 다른 방법으로 도달**했다 — A는 원리로, C는 코드
      읽기로, B는 무기화로. Manager가 직접 측정해 A-H1을 **차단으로 올렸고**(면제
      사유가 축 1 실행에 반증됨), A-차단2는 **MEDIUM으로 내렸다**(U5는 거짓이
      아니라 부정확 — 새 값을 실은 행은 고아 규칙이 잡는다). 승격과 강등이 둘 다
      측정으로 일어났다.

      **교차 모델 다섯 라운드 연속 미충족** — 상세는 5.1과 `review.md` 10.11.
- [ ] **5.4 통과 전에는 6·8·9·10절을 시작하지 않는다**

      > **"6절 이후"라고 쓰면 §7을 포함해 읽힌다**(7라운드 NIT). §7은 4라운드 B1에
      > 따라 **§6보다 먼저** 돌고, 7.4는 문서 검사라 freeze 전에 도는 것이 정상이다.
      > 이 게이트가 막는 것은 **구현과 그 검증**(§6 RED · §8 GREEN · §9 VERIFY ·
      > §10 gate)이지 §7의 기준선·훑기가 아니다.

## 7. 폭발 반경 확인 — **깨끗한 나무에서, 6절보다 먼저**

> **4라운드 B1: 이 절이 6절 뒤에 있으면 실행 자체가 불가능하다.** 6.2~6.6이 참조하는
> 상수는 8.2 전까지 존재하지 않고, `package engine`의 `_test.go`에 미정의 식별자가
> 하나라도 있으면 **패키지 테스트 바이너리 전체가 안 만들어진다**(`[build failed]`).
> 그 시점에는 회귀도 못 세고 기준선도 못 잡는다. **3라운드 R4와 같은 형태가 절 단위로
> 올라온 것이다** — 자기 실패가 자기 실행을 막는 구조.
>
> 그래서 7.1·7.2는 **6.1보다 먼저** 깨끗한 나무에서 돌고, 7.3만 8절 뒤에 돈다.

- [ ] 7.1 **base에서 회귀 기준선을 잡는다.**

      ```bash
      go test -count=1 -timeout 30m ./...
      go test -count=1 -timeout 30m -tags tossos_testseams ./...
      ```

      **`-timeout 30m`이 빠지면 이 단계가 간헐적으로 죽는다.** Go의 패키지당 기본
      기한은 **10분**인데 `internal/journal`이 그것을 넘는다 — 5라운드가 태그판에서
      **999.031초**를 쟀고, `Makefile:28-35`가 같은 사실을 적어 두고 `make test`에  <!-- not-a-measurement: 테스트 스위트 벽시계 시간 — 알림 지연 계열이 아니다 -->
      `-timeout 30m`을 붙여 놓았다. 조용한 기계에서는 통과하고 바쁜 기계에서는
      죽는다. **기준선을 잡는 단계가 부하에 따라 중단되는 것이 가장 나쁜 실패 방식이다.**

      **`cmd/tossctl`의 조립 경로는 build tag 뒤에 있다**: `engine_assembly_test.go`와
      `engine_account_seq_recovery_test.go`가 `//go:build tossos_testseams`이고
      `make test`도 `go test ./...`도 그 태그를 안 켠다. **태그 없이 돌면 CLI 조립
      경로가 통째로 빠진다** — 4라운드가 실측으로 잡아낸 함정이다.
      그러므로 **두 번 돈다**: 태그 없이, 그리고 `-tags tossos_testseams`로.
- [ ] 7.1.1 **도달 테스트 수는 이 두 명령이 만들어내지 못한다 — 별도 프로브다.**
      위 두 줄은 `ok`/`FAIL`만 낸다. 아래 표의 68/72·80/86·82/88은 **overlay 프로브**의
      산출이고, 그것을 여기 적어 두지 않으면 이 박스는 절 제목의 절반만 채우고
      체크된다(6라운드 H-3 — 침묵한 생략).

      > **7판이 여기 적은 프로브는 돌지 않았다 — 7라운드 차단 B3·B4·B5.** 셋이
      > 겹쳐서 **조용히 `0`을 출력하고 exit 0**이 됐다. 침묵한 생략을 고치려고 만든
      > task가 침묵한 0 생성기였다. 셋을 각각 이렇게 고쳤다.
      >
      > | # | 7판의 결함 | 8판의 수정 |
      > |---|---|---|
      > | B3 | 두 파일에 `"os"`를 무조건 주입 → `notifications.go`가 이미 import해서 `os redeclared` | **프로브 본체를 새 파일**(`a092probe.go`)로 넣는다. 기존 파일의 import 블록을 아예 안 건드린다 |
      > | B4 | grep 패턴이 `tossos/` — 모듈 경로는 `github.com/JungHoonGhae/tossinvest-cli` | 프레임을 **모듈 경로로** 찾는다 |
      > | B5 | `-v`가 없어 통과한 패키지의 테스트 바이너리 출력을 cmd/go가 버린다 | **`-v`를 붙인다** |
      > | H3 | 두 프로브가 같은 마커를 써서 6칸 표에 넣을 수가 2개뿐 | 마커에 **함수 이름을 붙여** 함수별로 세고 합집합까지 낸다 |

      **먼저 전수임을 확인한다** — `internal/app/engine`을 import하는 패키지가
      `cmd/tossctl` 하나뿐이어야 두 패키지가 전부다.

      ```bash
      go list -f '{{.ImportPath}}{{range .Imports}} {{.}}{{end}}' ./... |
        grep 'internal/app/engine' |
        grep -v '^github.com/JungHoonGhae/tossinvest-cli/internal/app/engine ' |
        cut -d' ' -f1
      # 기대: github.com/JungHoonGhae/tossinvest-cli/cmd/tossctl  (이 한 줄뿐)
      ```

      그다음 프로브. **저장소를 편집하지 않는다** — `-overlay`만 쓴다.

      ```bash
      P=$(mktemp -d)
      python3 - "$P" <<'PY'
      import json, pathlib, sys
      out = pathlib.Path(sys.argv[1]); pkg = pathlib.Path('internal/app/engine')
      # 프로브 본체는 새 파일이다 — 기존 import 블록을 건드리지 않는다(B3).
      helper = ('package engine\n\nimport (\n\t"os"\n\t"runtime"\n\t"strings"\n)\n\n'
                'func a092reach(tag string) {\n'
                '\tif os.Getenv("A092_REACH") == "" {\n\t\treturn\n\t}\n'
                '\tbuf := make([]byte, 1<<16)\n'
                '\tdump := string(buf[:runtime.Stack(buf, false)])\n'
                '\tos.Stderr.WriteString("A092REACH " + tag + " " +\n'
                '\t\tstrings.ReplaceAll(dump, "\\n", "|") + "\\n")\n}\n')
      (out / 'a092probe.go').write_text(helper)
      overlay = {"Replace": {str(pkg / 'a092probe.go'): str(out / 'a092probe.go')}}
      for name, fn, tag in (('exitwiring.go', 'func newNotifier(', 'newNotifier'),
                            ('notifications.go', 'func resolveNotificationPublisher(',
                             'resolvePublisher')):
          src = (pkg / name).read_text()
          j = src.index('{\n', src.index(fn)) + 2
          (out / name).write_text(src[:j] + f'\ta092reach("{tag}")\n' + src[j:])
          overlay["Replace"][str(pkg / name)] = str(out / name)
      (out / 'overlay.json').write_text(json.dumps(overlay))
      PY
      cat > "$P/count.py" <<'PY'
      import re, sys
      MOD = r"github\.com/JungHoonGhae/tossinvest-cli/[^ (]*?\.(Test[A-Za-z0-9_]*)"
      seen = {}
      for line in sys.stdin:
          if "A092REACH " not in line:
              continue
          tag, _, dump = line.split("A092REACH ", 1)[1].partition(" ")
          hits = re.findall(MOD, dump)
          if hits:                       # 스택은 안쪽부터 — 바깥쪽 Test가 마지막
              seen.setdefault(tag, set()).add(hits[-1])
      a, b = seen.get("newNotifier", set()), seen.get("resolvePublisher", set())
      print(f"newNotifier={len(a)} resolvePublisher={len(b)} union={len(a | b)}")
      PY
      for TAGS in "" "-tags tossos_testseams"; do
        printf '%-28s ' "${TAGS:-(no tag)}"
        A092_REACH=1 go test -overlay "$P/overlay.json" -count=1 -timeout 30m -v \
          $TAGS ./internal/app/engine ./cmd/tossctl 2>&1 | python3 "$P/count.py"
      done
      ```

      **이 명령은 8판에서 실제로 돌렸고 아래 표를 이 기계에서 재현했다**(go1.26.5,
      코어 20):

      ```text
      (no tag)                     newNotifier=68 resolvePublisher=80 union=82
      -tags tossos_testseams       newNotifier=72 resolvePublisher=86 union=88
      ```

      > **복사할 때 6칸 들여쓰기를 지운다**(8라운드 NIT). 이 파일의 코드펜스는 전부
      > 리스트 항목 안에 있어서 6칸이 붙는데, 그대로 붙여넣으면 heredoc 종료 표지가
      > 안 맞아 `here-document at line 2 delimited by end-of-file (wanted 'PY')`가 난다.
      > dedent 후에는 `bash -n`이 조용하다.

      **`0`이 나오면 그것은 "도달이 없다"가 아니라 "프로브가 안 돌았다"이다.**
      먼저 `-v`가 붙어 있는지, 그다음 모듈 경로가 맞는지 본다. 얻은 수가 아래 표와
      다르면 **표를 이 기계의 값으로 갱신한다.**
- [ ] 7.2 기준선을 **이 기계에서** 기록한다. 4라운드 실측이 보여준 대로 실행 간 편차가
      6.6이 더하는 ~4.3초보다 크므로(같은 기계에서 engine 74.5s와 118.7s), **뺄셈으로  <!-- not-a-measurement: 테스트 패키지 벽시계 시간 — 알림 지연 계열이 아니다 -->
      비용을 말하지 않는다** — R5의 비용은 그 테스트 자신의 `t.Logf`로 보고한다.
- [ ] 7.3 **8절 뒤에 7.1의 두 명령을 그대로 다시 돈다**(`-timeout 30m` 포함, 태그판 포함).
      Impact의 "기존 테스트 재작성 0건"을 여기서 검증한다. 깨지는 것이 있으면
      **목록을 Impact에 옮겨 적고** 재작성 범위를 다시 산정한다.
      **1판이 여기서 16건 틀렸다**

      **⚠ `make gate`는 태그판을 절대 돌리지 않는다.** 게이트 6단계는 `make test`
      (= `go test -timeout 30m ./...`, 태그 없음)이므로 **태그로만 도는 테스트는
      게이트에서 한 번도 실행되지 않는다.** 그 수는 세 개이고 서로 다르다 —
      6판은 그중 가장 작은 것만 적어서 게이트가 놓치는 전부처럼 읽혔다(6라운드 M-3).

      | 범위 | 태그로만 도는 테스트 |
      |---|---|
      | 나무 전체 | **68건** |
      | 폭발 반경 5개 패키지 | **43건** |
      | 그중 `cmd/tossctl`의 **조립** 테스트 | **6건** ← a092가 편집하는 리터럴에 닿는 것 |

      초록 게이트를 "조립 경로까지 검증됐다"로 읽으면 안 되고, 그 검증은
      **7.3의 태그판이 유일한 자리**다. 7.3을 건너뛰면 a092가 편집하는 두 리터럴 중
      CLI 쪽 도달 경로는 아무도 안 돌린 채로 통과한다.

### 4라운드가 실측한 도달 수 (참고 — **7.1.1**이 이 기계에서 재확인한다)

방법: `newNotifier`와 `resolveNotificationPublisher`에 overlay로 `runtime.Stack` 프로브를
넣어 바깥쪽 `Test*` 프레임을 기록. `go list`로 `internal/app/engine`을 import하는 패키지가
`cmd/tossctl` **하나뿐**임을 확인했으므로 두 패키지가 전수다.

| 실행 | `newNotifier` 도달 | `resolveNotificationPublisher` 도달 | 합집합 |
|---|---|---|---|
| `go test ./...` (태그 없음) | 68 | 80 | **82** |
| `-tags tossos_testseams` | 72 | 86 | **88** |

태그로만 도는 6건: `TestAssembleEngineWiresOneProductionGuardianToTheEngineJournalAndExitObserver`,
`TestActualEngineRecovery{ReusesTheStartupAccountSequence, AcceptsAMatchingExplicitAccountSequence,
StillFailsClosedOnASnapshot429}`, `TestEngineRefuses{AnIncompleteFirstAccountRecordBeforeOpeningTheJournal,
AnExplicitSequenceThatDoesNotMatchTheFirstRecord}`.
**2라운드가 쓴 `cmd/tossctl` "7건"은 틀렸다 — 6건이다.**

**8.1~8.4 편집으로 깨지는 기존 테스트: 0건**(4라운드가 overlay로 두 패키지 태그 유무
모두 + `internal/obs`·`internal/console`·`internal/execgw`까지 확인).

### 7.4 훑기의 기계적 검사 — 표를 사람이 믿지 않게 만든다

- [ ] 7.4.0 **값 단위 교차문서 검사를 먼저 돌린다.**

      ```bash
      python3 openspec/changes/a092-an-alert-does-not-hold-the-stop/tools/check_values.py
      ```

      **이것이 7판의 첫 산출물이고, 이 절의 나머지보다 먼저 돈다.** 7.4.1~7.4.3은
      *산출물과 표*가 같은 지도를 가리키는지 보고, 7.4.0은 *문서들끼리* 같은 값을
      말하는지 본다. 6라운드 차단 3건 중 둘이 후자였고 전자로는 잡히지 않았다.

      스크립트가 하는 것: 등식을 **평가**하고(6라운드 B1은 상수가 틀리기 전에 산술이
      안 닫혔다), 채택량의 등식이 채택 상수만 쓰는지 보고, design D2 후보 표의 각 행이
      자기 산술을 지키는지 보고, 인용된 지연 측정치가 `analysis/`에 있는지 보고
      (6라운드 B2의 `28.9ms`가 여기 걸린다), 명명된 양이 문서마다 하나의 값인지 보고, <!-- rejected-value -->
      판 번호가 전 문서에서 같은지 본다.

      **못 하는 것도 적는다 — 스크립트 머리말의 셋을 그대로 옮긴다.**
      7판은 **셋 중 둘만** 옮겼고, 빠진 것이 하필 나머지 전체를 한정하는 것이었다
      (7라운드 M-6).

      1. **문장의 참을 검사하지 않는다.** `design.md`의 Go 주석이 "그 다섯이 전부
         journal 커넥션에 줄을 선다"고 거짓을 적은 것(6라운드 H2, `EntryGate.Block`은
         뮤텍스뿐)은 소스를 읽어야 나온다. 7라운드 H4의 기각된 창 정의도 같은 종류다 —
         **값은 하나도 안 틀리고 문장만 틀렸다.**
      2. **`review.md`는 대상이 아니다.** 기각된 값을 일부러 인용하는 파일이다.
      3. **정규식은 열거이지 전칭이 아니다.** 새 양을 문서에 도입하면 스크립트의
         `NAMED_QUANTITIES`에 손으로 추가해야 하고, **그것을 강제하는 기계는 없다.**
         이 항목이 앞의 둘을 한정한다 — 초록은 "검사한 것이 일치한다"이지
         "검사할 것이 다 있다"가 아니다.

      **그리고 8판은 이 스크립트를 돌려 본다.**

      ```bash
      python3 openspec/changes/a092-an-alert-does-not-hold-the-stop/tools/check_values_selftest.py
      ```

      자체 시험은 change 디렉터리를 임시 사본으로 복제하고 **각 라운드가 실증한
      우회를 하나씩 주입해 검사가 실패하는 것을 본다.** 7라운드 차단 4건의 형태가
      "방어를 쓰고 돌려 보지 않았다"였으므로, 검사를 고치는 것과 그 검사를
      우회해 보는 것은 **같은 task다.** 우회 하나라도 통과하면 자체 시험이 실패한다.

      > **케이스 수를 여기 적지 않는다.** 8판은 「8가지」, 10판은 「19건」이라
      > 적었고 둘 다 그 시점에 이미 틀렸다(10라운드 M7). 자체 시험이 마지막 줄에
      > 자기 케이스 수를 세어 출력한다 — 읽을 곳은 거기다.

- [ ] 7.4.0.1 **커버리지 게이트를 돌린다 — 검사가 실제로 시험되고 있는지를 센다.**

      ```bash
      python3 openspec/changes/a092-an-alert-does-not-hold-the-stop/tools/coverage_gate.py
      ```

      > **10판이 이 도구를 만들고 어떤 task에도 안 넣었다 — 10라운드 차단 2.**
      > `tasks.md`·`Makefile`·`docs/WORKFLOW.md` 어디에도 `coverage_gate` 문자열이
      > 없었고, 오직 `review.md`만 그것을 "10판의 첫 산출물"이라 불렀다.
      > **완료 게이트가 돌리지 않는 방어는 방어가 아니라 기록이다.**

      7.4.0이 "검사가 통과하는가"를 묻는다면 이 task는 **"그 검사에 시험되지 않는
      자리가 있는가"**를 묻는다. 두 축으로 센다 — 축 1은 `fail()` 이 도달할 수 있는
      **경로**를 하나씩 죽여 자체 시험이 그것을 알아채는지 보고, 축 2는 알려진 위반을
      기계적으로 변형해 전 위치에 심는다. 어느 자리든 죽여도 자체 시험이 그대로
      통과하면 실패다.

      **미커버는 케이스를 더해서 닫고, 빠져나간 섭동은 검사를 고친 뒤 그 섭동을
      케이스로 넣는다.** 사람이 "덮었다"고 판단하지 않는다 — 세 판이 그렇게
      판단했고 세 판 다 틀렸다.

      > **이 게이트가 못 세는 것도 매 실행마다 출력된다**(계열 U1~U7).
      > 두 축이 세는 것은 **값의 일치**이고 문장의 참이 아니다. 통과 줄만 읽고
      > "전부 덮었다"로 오독하는 것이 7·8·9판의 형태였으므로, 못 세는 자리를
      > 게이트 자신이 말한다. **U 계열은 사람이 봐야 하는 자리다.**
      >
      > `Makefile`과 `docs/WORKFLOW.md`에는 넣지 않는다 — change-local 도구는
      > archive와 함께 사라지므로 저장소 수준 target에 넣으면 그때 깨진다.
      > `check_values.py`도 같은 이유로 task에만 있다. **not-applicable 사유를
      > 여기 적는 것이 침묵한 생략과 다른 점이다.**
- [ ] 7.4.1 **분기 주장은 `ast.json`으로 대조한다.** 표의 어느 행이든 `B<n>`이나 반환
      줄 번호를 말하면, 그 산출물의 `ast.json`에서 같은 값을 읽어 확인한다.
      이 검사가 5라운드 H2를 **정확히** 잡는다 — 5판 표는 nil 이탈을 B2·B3·B5로
      옳게 적었지만 산출물의 branch 표는 B1·B3·B5였고, `ast.json`은 `B1@69`·
      `returns 81/87/97/105`로 어느 쪽이 참인지 혼자 판정한다.

      ```bash
      python3 - <<'PY'
      import json, pathlib
      base = pathlib.Path('openspec/changes/a092-an-alert-does-not-hold-the-stop/analysis/function-logic')
      for d in sorted(base.iterdir()):
          a = d / 'ast.json'
          if not a.exists(): continue
          j = json.loads(a.read_text())
          bs = [(b['id'], b['at']['line']) for b in (j.get('branches') or [])]
          rs = [r['at']['line'] for r in (j.get('returns') or [])]
          print(f"{d.name}: branches={bs} returns={rs}")
      PY
      ```

- [ ] 7.4.2 **산출물의 비표준 절은 전부 표에 대응이 있어야 한다.** FLM의 다섯 표준 절
      (`Inputs and invariants` / `Branches and early returns` / `Calls and live bindings` /
      `State mutations and fallbacks` / `Safety conclusion`) 밖의 절은 **그 산출물이
      스키마 밖의 사실을 적어 둔 자리**다. 그것이 표에 없으면 H3다.
      **개수를 여기 적지 않는다.** 6판은 `현재 16개`라고 박아 넣었고, **6판 자신이
      5라운드 H3에 대응해 절을 하나 더하는 순간 그 숫자가 거짓이 됐다**(실제 17,
      6라운드 M1/H-4). **기대값을 박은 검사는 검사가 아니라 또 하나의 주장이다.**
      개수도 대응 여부도 **7.4.0이 유도한다** — 훑기 표의 각 행에서 그 산출물의
      비표준 절 제목을 **8자 이상 그대로** 찾고, 하나라도 없으면 실패한다.
      아래 명령은 그 목록을 사람이 읽으려고 뽑는 것이지 판정이 아니다.

      ```bash
      python3 - <<'PY'
      import pathlib, re
      base = pathlib.Path('openspec/changes/a092-an-alert-does-not-hold-the-stop/analysis/function-logic')
      boiler = {'Inputs and invariants','Branches and early returns','Calls and live bindings',
                'State mutations and fallbacks','Safety conclusion'}
      n = 0
      for d in sorted(base.iterdir()):
          f = d / 'function-logic-map.md'
          if not f.exists(): continue
          for h in [l.strip() for l in f.read_text().splitlines() if re.match(r'^#{2,4} ', l)]:
              if h.lstrip('# ').strip() not in boiler:
                  n += 1
                  print(f"{d.name}\t{h}")
      print(f"-- {n}개 (유도값 — 문서에 박지 않는다)")
      PY
      ```

      **그래서 훑기 표를 채우는 규칙이 하나 늘었다**: 산출물이 비표준 절을 가지면
      표의 그 행이 **절 제목을 그대로 인용한다.** 그래야 "표가 그 절을 봤다"가
      사람의 판단이 아니라 문자열 대조가 된다.

- [ ] 7.4.3 **`## Safety conclusion`의 굵은 주장도 같은 대상이다.** `ntfy.publish`의
      "왕복 0.2~1.8초"는 절이 아니라 그 안의 불릿이었고, 그래서 7.4.2만으로는 안 잡힌다.
      23개 산출물의 `Safety conclusion`을 **읽어서** 표와 대조한다 — 이것 하나는
      기계가 아니라 사람이 하고, **읽었다는 증거는 표의 해당 행이 그 문장을 인용하는 것**이다.

**규칙**: 표의 각 행은 **산출물을 열어서** 채운다. 다른 문서에서 옮겨 적으면
그 행은 산출물을 본 적이 없는 것이고, 5라운드 H2가 정확히 그렇게 생겼다.

## 6. RED — `internal/app/engine/a092_alert_budget_test.go` (`package engine`)

`package engine`이어야 하는 이유: `newNotifier`·`alertBudget`·`alertPublishTimeout`이
전부 unexported다. 선례 — `exit_wiring_internal_test.go:97`가 `newNotifier`를 부르고
`interlock_internal_test.go:75`에 `openTestJournal`이 있다.

- [ ] 6.1 파일 신설
- [ ] 6.2 R1 — `newNotifier(...).Attempts == alertPublishAttempts` (RED: `[build failed]`)
- [ ] 6.3 R2 — `newNotifier(...).RetryDelay == alertRetryDelay` (RED: `[build failed]`)
- [ ] 6.4 R3 — `resolveNotificationPublisher(...)`가 돌려준 `obs.Publisher`를
      `*obs.Ntfy`로 타입 단언 후 `Timeout == alertPublishTimeout` (RED: `[build failed]`).

      > **RED는 "0건 통과"가 아니라 `[build failed]`다**(7라운드 NIT). 세 상수가
      > 아직 없으므로 `package engine`의 테스트 바이너리가 만들어지지 않는다 —
      > 7판이 "(RED: 0)"이라고 적은 것은 **실행됐는데 실패했다**로 읽히고,
      > 그것은 이 절이 §7보다 뒤에 있으면 왜 실행 자체가 불가능한지(4라운드 B1)와
      > 어긋난다. 7라운드 보이스 B가 실제로 돌려 `undefined: alertPublishAttempts …
      > [build failed]`를 확인했다.
      반환형이 인터페이스이므로 단언이 필요하다 — 선례 `a074_notification_resolve_test.go:42`
- [ ] 6.5 **R4 — 상수 값 고정 (등식 증명이 아니다).**
      `alertPublishTimeout == 1300ms` · `alertRetryDelay == 150ms` ·
      `alertPublishAttempts == 3` · `alertOverheadReserve == 800ms`를 리터럴로 단언한다.
      **이 테스트는 예산 부등식을 지키지 못한다** — 단언이 같은 패키지의 컴파일 타임
      배열이므로 부등식이 깨지면 **테스트 바이너리가 만들어지지 않고 이 테스트는
      실행조차 되지 않는다**(3라운드 지적). 이것이 잡는 것은 **합법적이지만 의도하지
      않은 재배분**이다 — 예: 1.3s→1.2s, 150ms→300ms는 여전히 4.2s라 컴파일된다.  <!-- not-a-measurement: 합법적 재배분 예시의 유도값 — 3×1.2s+2×300ms=4.2s의 입력이다 -->
      주석에 **무엇을 지키고 무엇을 못 지키는지** 적는다
- [ ] 6.6 **R5 — 알림 하나의 실시계 체류.** 3라운드에서 이 과제의 리터럴이
      **실행 불가**였다(아래 ⚠). 4판은 실측한 모양 그대로 적는다.

      필요한 import: `context` · `net` · `os` · `path/filepath` · `testing` · `time` +
      `internal/clock` · `internal/config` · `internal/execgw` · `internal/obs`.
      **`internal/config`가 빠지면 컴파일이 안 된다** — 6.4가 `config.Notifications`를
      쓴다(6라운드 NIT). `os`·`path/filepath`는 아래의 실제 로거가 쓴다.

      ```go
      ln, err := net.Listen("tcp", "127.0.0.1:0")  // listen만, accept 하지 않는다
      if err != nil { t.Fatalf("listen: %v", err) } // 무시하면 아래에서 nil 역참조다
      t.Cleanup(func() { _ = ln.Close() })
      j := openTestJournal(t)
      clk := clock.System()                        // ⚠ 가짜 시계 금지 (아래)
      gate := execgw.NewEntryGate(clk, nil)
      pub := &obs.Ntfy{
          BaseURL: "http://" + ln.Addr().String(), // ⚠ 없으면 다이얼 전에 반환
          Topic:   "a092-probe",                   // ⚠ 없으면 ErrNtfyNotConfigured
          Timeout: alertPublishTimeout,
      }
      // ⚠ 4번째 인자에 nil을 넘기지 않는다 — 아래 ⚠ 참조.
      lf, err := os.Create(filepath.Join(t.TempDir(), "engine.log"))
      if err != nil { t.Fatalf("log file: %v", err) }
      t.Cleanup(func() { _ = lf.Close() })
      lg := obs.NewLogger(obs.LogOptions{Writer: lf, JSON: true, Clock: clk})
      n := newNotifier(j, gate, "acct-probe", lg, pub, clk)
      start := time.Now()
      nerr := n.Notify(context.Background(), obs.Event{Type: obs.EventExitJudgementRefused})
      elapsed := time.Since(start)
      // 하한: 세 시도가 실제로 기한까지 막혔다
      // 상한: spec이 약속한 것 그대로 — 관측 주기를 넘지 않는다
      if elapsed < alertTransportBudget || elapsed >= alertBudget { t.Errorf(...) }

      // 그리고 **결과가 보존됐는지** 단언한다 — 두 delta의 "상한을 소진해도
      // 기록은 남는다" 시나리오가 여기 걸린다 (아래 ⚠).
      if nerr != nil {
          t.Fatalf("outbox 쓰기가 실패하면 이 실행은 아무것도 안 잰다: %v", nerr)
      }
      rows, qerr := j.PendingAlerts(context.Background(), 0)  // journal/outbox.go:209
      if qerr != nil || len(rows) != 1 {
          t.Errorf("소진 뒤 outbox 행이 PENDING으로 남아야 한다: rows=%d err=%v", len(rows), qerr)
      }
      if _, blocked := gate.Blocks()[execgw.ReasonAlertUndelivered]; !blocked {
          t.Error("소진 뒤 진입 게이트가 래치돼야 한다")   // execgw/retry.go:569
      }
      ```

      > **⚠ 9라운드 차단 1 — 이 네 줄이 세 군데서 틀렸었다.**
      >
      > 1. `_ = n.Notify(...)`가 **반환값을 버렸다.** 8라운드 M-12가 그것을 짚었는데
      >    9판의 정정이 ⚠ 산문에만 착지하고 리터럴에는 안 갔다.
      > 2. 그 아래 `if err != nil`의 `err`는 **`os.Create`의 것**이고 바로 위에서
      >    이미 `t.Fatalf`로 걸렀다 — 그 지점에서 증명적으로 nil인 **죽은 검사**였다.
      >    이제 `nerr`로 이름을 갈라 `Notify`의 값을 본다.
      > 3. `j.PendingAlerts(context.Background())`는 **컴파일되지 않았다.**
      >    `journal/outbox.go:209`의 시그니처는 `(ctx context.Context, limit int)`이고
      >    저장소의 호출자 8곳이 전부 `(ctx, 0)`을 넘긴다
      >    (`notifier.go:311`·`:355`, `obs_test.go:407`·`:480`·`:516`·`:579`,
      >    `execgw/replay_test.go:638`, `journal/outbox_test.go:49`).
      >
      > **`nerr == nil`이 outbox 쓰기 성공의 증거인 것은 참이다** — 다만 8판이
      > 그렇게 쓴 이유 때문이 아니라 `Notify`의 오류 경로가 **하나뿐**이기 때문이다:
      > `notifyCritical`은 `EnqueueAlert` 실패에만 오류를 돌려주고(`notifier.go:177`),
      > 전송이 시도를 전부 소진해도 `escalate` 뒤 **nil을 돌려준다**(`:187` 다음
      > `return nil`). 그 설계 의도는 `Notify`의 선언 주석이 직접 적어 놓았다
      > (`notifier.go:105-107` — 게이트 래치로 여기서 이미 처리했으므로 호출자마다
      > 그 판단을 다시 구현하게 하지 않는다). 그래서 이 프로브에서 `nerr != nil`은
      > **원장이 죽었다**는 뜻이고, 그때는 체류 측정 자체가 무의미하므로 `Fatalf`다.
      >
      > 이 리터럴은 `-overlay`로 `internal/app/engine`에 넣어 `go vet`을 통과시킨
      > 뒤에 여기 적었다. 9판은 그 단계를 건너뛰었고 리뷰가 `go vet` 한 번으로 잡았다.

      **⚠ `Topic`과 `BaseURL`이 없으면 이 테스트는 자기 하한을 떨어뜨린다.**
      `Ntfy.Publish`는 `Topic`을 **먼저** 트림하고 비었으면 `ErrNtfyNotConfigured`로
      **다이얼 전에** 반환한다(`ntfy.go:86-89`). 3판의 리터럴대로면 체류가 521ms다.
      **추론이 아니라 3라운드가 실행해서 잰 값이다.**

      **⚠ `newNotifier`에 반드시 `clock.System()`을 넘긴다.** `newNotifier` FLM이
      `Clock: clk`를 채운다고 적었고 `Notifier.wait`는 그 시계로 잔다
      (`notifier.go:295-299`). 가짜 시계를 넘기면 `Advance`가 없어 **재시도 대기에서
      영원히 막힌다**(`clock/fake.go:42`). 3라운드가 12초 워치독을 넘겨 막히는 것을
      실제로 확인했다. 두 FLM을 겹쳐 읽어야 나오는 함정이다.

      **비용**: 실시간 약 **4.3초**. `-race`는 체류를 의미 있게 늘리지 않는다 —
      3판의 "`-race`에서 더 걸린다"는 틀렸다. 늘리는 것은 **원장 경합**이다.

      **⚠ `Notify`의 반환값을 버리면 6.11의 두 행이 근거를 잃는다**(8라운드 M-12).
      8판의 리터럴은 `_ = n.Notify(...)`였는데 6.11은 ES:68·EP:46("상한을 소진해도
      기록은 남는다")의 §6 대응으로 **R5를 지목**하면서 그 근거를 *"`Notify`가 오류
      없이 반환하는 것이 outbox 쓰기 성공의 증거"*라고 적었다. 버린 값은 증거가 될 수
      없다. 위 네 줄이 그것을 실제 단언으로 만든다 — 8라운드 보이스 B가 같은 세팅에서
      `PendingAlerts = 1`, `Blocks() = map[critical_alert_undelivered:…]`를 실측했다.

      **⚠ `log`에 `nil`을 넘기면 이 테스트가 재는 경로가 프로덕션 경로가 아니다.**
      `Notifier`의 로그 쓰기는 전부 `n.Log != nil` 뒤에 있으므로, `nil`을 넘기면
      소진 경로의 구조화 로그 **세 줄**(`notifier.go:131`·`:279`·`:228`)이 통째로
      빠진다. 7판까지의 프로브가 정확히 그랬고, 그 프로브가 잰 값이 reserve 800ms의
      근거였다(7라운드 H2). **재측정 결과 그 세 줄은 주변 부하 아래라 값은 안 바뀌지만**
      (`delivery-latency.md` §7.5), 그것은 **재고 나서 알게 된 것**이지 nil을 넘겨도
      되는 이유가 아니다. 이 테스트는 프로덕션 조립과 같은 모양을 쓴다.

      **⚠ 상한 단언은 실시계 hard 단언이고, 여유는 기계·주변 부하에 종속이다.**
      비전송 초과분의 전 회차 관측 범위는 **31.9 ~ 356.1ms**이며(`delivery-latency.md`
      §7.2.1) 그 문서가 "산포 배수"라 이름 붙인 값이 **11.2배**다. 9.4가 이 테스트를
      **다른 패키지들과 함께** 돌리므로 주변 부하가 가장 높은 순간에 걸릴 수 있다.

      > **7판은 여기서 `go test ./... -count=1 -race`를 근거로 들었는데 9.4는 그것을
      > 더는 돌리지 않는다**(7라운드 M-5). 5라운드 B1 이후 9.4는 (a) 태그 유무 전체
      > 나무(`-race` 없음)와 (b) 폭발 반경 5개 패키지(`-race`) 두 갈래이고,
      > 이 테스트는 (b)에 든다. **같은 문서 안에서 한쪽만 정정된 자리**이고
      > 6라운드 H-1과 같은 형태다.

      **실패했을 때의 판정 절차 — 완화하지 않는다.** 상한 위반은 spec 위반 신호이므로
      임계를 올리는 것은 답이 아니다. 순서는 이렇다.

      1. **단독 재실행** — `go test -run <이 테스트> -count=5`. 5회 전부 상한 아래면
         부하 아티팩트다. 그 사실과 두 실행의 `elapsed`를 9.4의 기록에 남긴다.
      2. **한 번이라도 상한 이상이면 초과분을 분해한다** — `elapsed - alertTransportBudget`이
         `alertOverheadReserve`(800ms)를 넘었는지 본다. 넘었다면 **상수 유도가 깨진 것**이고
         design D2의 규칙(reserve ≥ 전 회차 최악의 2배)으로 **재유도**한다.
      3. **`elapsed`가 transport 예산보다 작으면** 그것은 상한이 아니라 **하한** 실패이고,
         전송이 실제로 막히지 않았다는 뜻이다 — `Topic`·`BaseURL`·`Clock` 함정(위 ⚠ 둘)을
         먼저 본다.

      **임계를 올려서 통과시킨 실행은 통과가 아니다.** 그렇게 하면 6.6은 spec을
      검사하지 않고 자기 자신을 검사하게 된다 — 4라운드 C2가 그 형태였다.
      **이 배수는 기계마다 다르다**(4라운드 N1: 같은 상수에서 다른 기계가 60.7~82.9ms).
      7.2가 이 기계의 값을 잡는다.
- [ ] 6.7 **ObserveOnce 사이클 단위 테스트는 만들지 않는다** (`not-applicable`).
      계약이 알림당이고, 사이클 총합은 알림 수 × 예산 + `n.mu` 경합이라 a092가 지지 않는다.
      사이클 총합의 RED는 **a093**이 진다
- [ ] 6.8 **회귀 핀(태어날 때 GREEN, RED 아님)** — `resolveNotificationPublisher`의
      세 nil 이탈(**B2 `:77`·B3 `:83`·B5 `:94`**) 무변화.
      **B1 `:69`는 nil 이탈이 아니다** — `getenv == nil`일 때 `os.Getenv`를 넣을 뿐
      반환하지 않는다(4라운드 M1: 3·4판이 틀린 지도 위에 이 과제를 썼다).

      **⚠ 이 핀의 born-GREEN은 여기서 관측할 수 없다** (5라운드 B3). 6.8은
      `package engine`이고, 6.2~6.6이 같은 패키지에 미정의 상수를 남겨 둔다 —
      **패키지 테스트 바이너리가 안 만들어지므로 이 핀은 실행 자체가 안 된다.**
      5라운드가 실측했다: 핀만 있으면 `--- PASS: TestA092NilExitsUnchanged (0.00s)`,  <!-- not-a-measurement: go test 출력의 경과 시간 — 알림 지연 계열이 아니다 -->
      6.2~6.6과 함께면 `undefined: alertPublishAttempts …` · `[build failed]`.
      **4라운드 B1이 절 단위로 잡은 것과 같은 형태가 절 안에서 한 번 더 나온 것이다.**

      셋 중 하나를 고른다.

      1. **6.1보다 먼저 쓰고 돌려서 GREEN을 찍는다** — 7절 직후 **그리고 5.4 통과 뒤**,
         깨끗한 나무에서. (8라운드 NIT: 8판은 "7절 직후"만 적어 5.4의 게이트와
         어긋나 보였다. 6.8은 §6이므로 freeze 아래다.)
         그 로그가 born-GREEN의 증거다.
      2. **8.2 뒤로 미룬다** — 그러면 born-GREEN이 아니라 "GREEN 유지" 확인이 되고,
         6.10에 그렇게 적는다.
      3. **생략한다** — 이미 통과하는 4건이 세 이탈을 전부 덮는다(5라운드가 실재 확인):
         `TestARefusedConfigBlockStaysRefused`(B2),
         `TestNoPublisherWhenNotificationsAreOff` ·
         `TestADisabledBlockWithAChannelStaysOff`(B3),
         `TestAnEnabledBlockWithNoChannelIsRefused`(B5).
         생략하면 **사유를 여기 남긴다** — 침묵한 생략은 금지다.
- [ ] 6.9 **`cmd/tossctl/a092_testsend_source_test.go` — 태어날 때 GREEN**(RED 아님).
      `publishNotificationTest`의 `&obs.Ntfy{...}`(`notificationsettings.go:151`)에
      `Timeout` 필드가 **없음**을 고정한다 — design D6의 결정이 조용히 바뀌지 않게.
      **소스 스캔인 이유는 값으로 도달할 수 없기 때문이다**: 그 리터럴은 함수 지역
      변수이고 즉시 `Publish`에 쓰이며 반환도 저장도 되지 않는다.
      (`package main`이라서가 **아니다** — tossctl의 테스트도 `package main`이다.)
      **`strings.Contains`가 아니라 `go/parser`로 AST를 훑는다** — 선례
      `vetothresholds_source_test.go`(그것도 `package main`이고 `go/parser`를 쓴다).
      **주의: 이 테스트는 상대 경로로 소스를 읽으므로 `-overlay`로 시험할 수 없다**
      (4라운드 N6). 음성 대조는 in-memory 소스로 파서를 직접 돌려서 한다.
- [ ] 6.10 RED 실행 로그 보존. **여기서 관측 가능한 born-GREEN은 6.9 하나다** —
      `cmd/tossctl`은 다른 패키지라 `internal/app/engine`이 `[build failed]`인 동안에도
      돈다(5라운드가 실행해서 확인). **6.8은 같은 패키지라 관측되지 않으므로**
      위 셋 중 무엇을 골랐는지 여기 적는다(5라운드 B3).
      6.5(R4)도 born-GREEN이 **아니다**(4라운드 N4) — 고정하려는 네 상수가 base에
      없으므로 6.2~6.4와 같은 빌드 실패를 공유한다.
      즉 6.2~6.6의 RED 증거는 **빌드 실패 그 자체**이고, 그 로그를 보존한다
- [ ] 6.11 **델타 시나리오 14건 전부에 대해 §6 테스트 또는 `not-applicable` 사유를 적는다.**

      7판은 이 대응을 아무 데도 안 적었다(7라운드 M-9). 침묵한 생략 금지는
      "테스트가 없다"가 아니라 **"없는 이유를 안 적었다"**를 금지한다.

      | # | 시나리오 | §6 대응 |
      |---|---|---|
      | ES:40 | critical 알림 전달 실패 지속 | **not-applicable** — a092 이전부터의 요구이고 코드가 안 바뀐다. `internal/obs`의 기존 outbox·게이트 테스트가 정본 |
      | ES:44 | 조립부에 예산이 없다 | **R1·R2·R3** — 세 필드가 조립부에서 채워짐을 각각 단언한다 |
      | ES:48 | 한 알림의 체류가 exit 관측 주기를 넘지 않는다 | **R5** — 실시계 체류가 `[alertTransportBudget, alertBudget)` 안 |
      | ES:52 | 전송기를 다른 루프가 쥐고 있다 | **not-applicable** — THEN이 하한 주장("그 합은 주기를 넘는다")이라 단언은 **가능하다**. 그런데 그 하한은 **두 goroutine의 스케줄링에 종속**이라 테스트로 고정하면 간헐 실패를 산다. 그리고 task 2.9가 **이미 프로브로 쟀다**(`delivery-latency.md` §7.4, 뒤쪽 8.458초) — 재는 것으로 충분한 것을 테스트로 옮기지 않는다. 8라운드 M-14가 8판의 사유("단언할 상한이 없다")를 거짓으로 짚었다 |
      | ES:56 | 사이클 총 체류는 약속하지 않는다 | **not-applicable** — 같은 종류의 비약속. 근거는 `alertproposalrefused`의 `ast.json`(branches 0 = 억제 없음)이지 테스트가 아니다 |
      | ES:60 | 상한을 읽지 않는 transport | **not-applicable** — `Timeout`을 무시하는 가짜 `Publisher`를 만들어 재면 **그 가짜를 시험하는 것**이지 a092의 코드가 아니다. 이 시나리오가 적는 것은 설계의 한계다 |
      | ES:64 | 받았는데 실패로 기록된다 | **not-applicable** — 손실 ②(design D4). `deliver`의 기존 동작이고 a092는 그것을 **더 있음직하게** 만들 뿐 코드를 안 바꾼다. 재현 자체는 싸다(`httptest`는 표준 라이브러리이고 `internal/app/engine`만 해도 여섯 파일이 이미 쓴다 — 8라운드 M-13이 8판의 "서버가 필요하고"를 과장으로 짚었다). **사유는 비용이 아니라 대상이다**: 지연 서버를 세워 재면 그 테스트가 재는 것은 a092의 상한이 아니라 **그 서버의 지연**이다 |
      | ES:68 | 예산을 줄여도 기록은 그대로다 | **R5** — 프로브가 outbox 행 기록과 게이트 래치를 **직접 단언한다**: `PendingAlerts(ctx, 0)`가 1행, `gate.Blocks()`에 `ReasonAlertUndelivered`. `nerr == nil`은 그 둘의 전제(원장이 살아 있었다)를 확인하는 것이지 그 자체가 증거는 아니다 — `Notify`의 오류 경로는 `EnqueueAlert` 하나뿐이고(`notifier.go:177`) 전송 소진은 nil을 돌려준다(`:105-107`의 선언 주석) |
      | EP:26 | 관측 장기 두절 | **not-applicable** — a092 이전 요구, 무편집 |
      | EP:30 | 확정 하한 캡 | **not-applicable** — a092 이전 요구, 무편집 |
      | EP:34 | 응답하지 않는 알림 전송 중의 관측 | **R5** (ES:48과 같은 자리) |
      | EP:38 | 다른 루프가 전송기를 쥐고 있다 | **not-applicable** — ES:52와 같음 |
      | EP:42 | 한 사이클이 알림을 여러 번 올린다 | **not-applicable** — ES:56과 같음 |
      | EP:46 | 상한을 소진해도 기록은 남는다 | **R5** (ES:68과 같은 자리) |

      **`not-applicable` 9건의 형태는 셋뿐이다**: (a) a092가 코드를 안 바꾸는 기존
      요구, (b) **약속하지 않음을 적는 시나리오** — 단언은 **가능하나 그 값이
      스케줄링 종속이라 테스트가 재현을 보장하지 못한다**, (c) 재현에
      가짜를 세워야 해서 **가짜를 시험하게 되는 것**. 셋 다 "안 봤다"가 아니라
      "테스트가 이것의 증거가 되지 못한다"이고, (b)의 둘은 **테스트 대신 측정과
      AST 열거로** 근거를 갖는다.

      > **(b)의 정의가 9판에서 13줄 위의 ES:52 행과 모순이었다**(9라운드 H-6).
      > 8라운드 M-14가 8판의 사유("단언할 상한이 없다")를 **거짓으로 판정**해서
      > ES:52 행은 "단언은 가능하다"로 고쳤는데, 같은 task의 요약은 거짓으로
      > 판정된 그 문구를 (b)의 **정의**로 다시 쓰고 있었다. `not-applicable` 사유는
      > 「침묵한 생략 금지」가 요구하는 산출물이므로 **분류표가 거짓이면 사유 아홉 건이
      > 전부 근거를 잃는다.** 정정이 표 행에만 착지하고 그 행을 분류하는 정의에는
      > 안 간 것 — 6라운드가 이름 붙인 형태와 같다.

## 8. GREEN

- [ ] 8.1 `notifications.go`에 **`import "time"` 추가** (현재 `os`·`strings`·`config`·`obs`)
- [ ] 8.2 `notifications.go`에 상수 **6개** + 유도 주석 + **컴파일 타임 단언 6줄**

      ```go
      var _ [alertPublishAttempts - 1]struct{}
      var _ [alertPublishTimeout - 1]struct{}
      var _ [alertRetryDelay - 1]struct{}
      var _ [alertOverheadReserve - 1]struct{}

      var _ [alertPublishTimeout/time.Millisecond - 100]struct{}
      var _ [alertRetryDelay/time.Millisecond - 10]struct{}
      ```

      > **앞의 넷 — 7라운드 B1·H1.** 7판까지의 단언은
      > `var _ [alertOverheadReserve]struct{}` 하나였다. `[0]struct{}`가 합법이므로
      > 그 한 줄은 `transport ≤ budget`만 강제하는데, 두 delta는 `SHALL NOT`으로
      > **`<`** 를 요구한다. 그리고 세 상수 어느 하나가 `0`이면 피호출자 폴백이
      > 살아나 **컴파일 reserve는 양수인데 실제 transport가 7.9초**가 된다.  <!-- not-a-measurement: alertRetryDelay=0 반증의 유도값 — 3×1.3s+2×2s의 결과다 -->
      > `- 1` 넷이 그 두 구멍을 닫는다.
      >
      > **뒤의 둘 — 9라운드 H-1. 10판이 더했고 11판이 여기로 내렸다.**
      > 앞의 넷은 **나노초만 세므로 1300ms와 1300ns를 구별하지 못한다.** 단위를
      > 빠뜨린 `alertPublishTimeout = 1300`은 넷을 전부 통과해 **초록 빌드**를
      > 만들고, 그 빌드의 publish 기한은 1300ns라 **다이얼이 끝나기 전에 만료된다** —
      > 매 알림이 소진되고 매 알림이 진입 게이트를 래치하며 운영 모드를 승격시킨다.
      > **넷이 잡는 `0`보다 나쁘다.** 마지막 두 배열이 단위가 지고 있던 부분,
      > 즉 **자릿수**를 본다.
      >
      > **10판은 이 두 줄을 `design.md`에만 넣고 여기에 안 내렸다 — 10라운드 차단 1.**
      > 그래서 이 절을 적힌 대로 실행하면 그 판이 스스로 "넷보다 나쁘다"고 선언한
      > 실패 모드가 살아남았다. 10라운드가 그것을 실행해서 확인했다:
      > 네 줄만 넣고 `alertPublishTimeout = 1300`이면 **BUILD OK**이고
      > `alertOverheadReserve`는 `4.6999961s`로 **건강해 보인다.**  <!-- not-a-measurement: 5.0s − (3 × 1.3µs + 2 × 150ms)의 컴파일 상수 — 측정이 아니라 산술이다 -->
      >
      > 실측 결과와 컴파일러 메시지는 design D3의 표에 있다 — 채택 #3·`reserve = 1ns`·
      > M-4 대안은 BUILD OK, 후보 #1(reserve 0)·음수·세 상수의 0·**단위 누락 둘·단위
      > 오타 둘**은 전부 FAIL이며 **메시지가 각각 자기 상수 이름을 말한다.**
- [ ] 8.3 `resolveNotificationPublisher` 안의 `&obs.Ntfy{...}` 리터럴에
      `Timeout: alertPublishTimeout` — **줄 번호로 지시하지 않는다.** 8.1·8.2가
      import와 상수를 넣으면 그 리터럴의 줄 번호가 내려간다
- [ ] 8.4 `newNotifier`가 만드는 `&obs.Notifier{...}` 리터럴에 `Attempts`·`RetryDelay`
      — 같은 이유로 식별자로 지시한다
- [ ] 8.5 `git diff --stat -- internal/obs` 가 **비어 있음**을 확인하고 그 출력을
      완료 보고에 싣는다 (design D7의 "안 건드리는 것").
      **`cmd/tossctl`은 `git diff`로 보면 안 된다** — 6.9가 그 아래 파일을 **신설**하고
      `git diff`는 미추적 파일을 무시하므로 빈 출력이 거짓 증거가 된다(4라운드 N5).
      `git status --short -- cmd/tossctl`을 함께 싣고, 나와야 하는 것은
      `?? cmd/tossctl/a092_testsend_source_test.go` **하나뿐**임을 확인한다

## 9. REFACTOR / VERIFY

- [ ] 9.1 편집 후 AST 재생성 — `newNotifier` branches 0/returns 1/calls 0,
      `resolveNotificationPublisher` branches 5/returns 4 **무변화 확인**
- [ ] 9.2 **FLM 재생성은 2건이다** — `internal-app-engine--newnotifier`,
      `internal-app-engine--resolvenotificationpublisher`. `source_sha256`가 파일 전체
      해시이므로 **편집한 두 파일에 묶인 산출물만** stale이 된다.
      `check_analysis`는 base에 **존재했던** 함수만 요구하므로 **신설 테스트 파일은
      요구 대상이 아니다.** 나머지 20개는 손대지 않은 파일이라 유효하다.
      **진짜 위험은 반대다** — 기존 `_test.go`를 편집하면 그때 그 파일의 함수가
      요구 대상이 된다. 6절 계획은 신설 파일만 쓰므로 해당 없다

      **⚠ 재생성으로는 부족하다 — FLM의 *산문* 줄 번호가 조용히 낡는다** (5라운드).
      8.1·8.2가 `resolveNotificationPublisher` **위에** 상수 블록을 넣으므로 그 아래
      모든 줄이 밀린다. **이동 폭을 여기 적지 않는다** — 8.2를 적용한 뒤 아래를
      돌려서 나온 수를 쓴다.

      ```bash
      go run ./tools/logic-map -file internal/app/engine/notifications.go \
        -func resolveNotificationPublisher
      ```

      base의 `67-106` · B1..B5 `:69/:77/:83/:91/:94` · 반환 `:81/:87/:97/:105`가
      전부 그 출력의 값으로 간다.

      > **8판은 여기에 "약 +46, 67-106 → 113-152 근처"라고 **박았고 실측은 +68·135-175였다**
      > (8라운드 H1 — 보이스 B가 D3 상수 블록을 verbatim 적용해 위 명령으로 확인).
      > 48% 틀린 수다. **7.4.2 자신이 세운 원칙** — *"기대값을 박은 검사는 검사가
      > 아니라 또 하나의 주장이다"* — 을 같은 파일의 이 절이 어겼다.
      > **유도 가능한 수는 박지 않는다.** 명령이 있고 그 출력이 답이다.

      **`check_analysis`는 이것을 못 잡는다** — 소스 파일을 해시하고 분기 **ID**를
      맞출 뿐 산문 속 `:NN`을 보지 않는다. 그러므로 9.2는 `ast.json` 재생성으로
      끝나지 않고 **그 줄 번호를 지닌 모든 문서를 갱신한다.**

      > **갱신 대상을 손으로 열거하지 않는다 — 10라운드 M8.** 10판은 "두 FLM과
      > 두 branch-test-map"이라고 적었는데 같은 값이 **`tasks.md` 자신에도** 있다
      > (이 절의 `67-106`과 `:69/:77/:83/:91/:94`가 그것이다). 손으로 센 목록이
      > 빠뜨린 문서는 조용히 낡고, **그 낡음을 이 절이 지시한 7.4.1 재검사가
      > 그대로 밟는다.** 목록을 세는 대신 찾는다:
      >
      > ```bash
      > cd openspec/changes/a092-an-alert-does-not-hold-the-stop
      > grep -rln -F -e '67-106' -e ':69/:77/:83/:91/:94' . --include='*.md'
      > ```
      >
      > 나오는 파일 전부가 갱신 대상이다. `review.md`는 **라운드별 기록**이라
      > 그 시점의 값을 그대로 두되, 인용이 현재 값을 주장하는 자리는 갱신한다 —
      > 어느 쪽인지는 사람이 읽어야 한다(커버리지 게이트의 미탐 계열 U6).

      갱신 후 7.4.1의 `ast.json` 대조를 다시 돌려 표와 산출물이 같은 지도를
      가리키는지 확인한다 — **5라운드 H2가 정확히 이 종류의 어긋남이었다.**
- [ ] 9.3 `python3 tools/logic-map/check_analysis.py --change a092-...` PASS

      > **`check_analysis`는 PATH에 없다**(8라운드 M-15) — `which check_analysis`가
      > 빈손이다. 실행 형태는 저장소 안의 스크립트를 직접 부르는 것이고,
      > 3.2·9.3 두 자리에 같은 오기가 있었다.
- [ ] 9.4 **경합 검사는 폭발 반경 안에서만 돈다.** 나무 전체 `-race`는 **완주하지 못한다**
      — 5라운드 실측: `go test ./... -count=1 -race`가 `internal/journal`에서
      `panic: test timed out after 30m0s`로 죽고, 유휴 기계에서 `-timeout 60m`으로
      다시 돌려도 **43분 24초에 아직 진행 중**이었다. Go의 패키지당 기본 기한은
      **10분**이고, `Makefile:28-35`가 바로 이 패키지에 대해 그 사실을 적어 둔다
      (schema v30에서 600초를 넘겼다 · `internal/journal/schema.go:6`은 지금도 v30).
      `make test`의 `-timeout 30m`조차 `-race`에서는 부족하다.

      그래서 두 갈래로 돈다.

      ```bash
      # (a) 나무 전체 — 회귀 확인. race 없이, 저장소가 스스로 정한 기한으로.
      make test                       # = go test -timeout 30m ./...

      # (b) 폭발 반경 5개 패키지 — 경합 확인. 여기가 편집이 닿는 전부다.
      go test -race -count=1 -timeout 30m \
        ./internal/app/engine ./cmd/tossctl ./internal/obs \
        ./internal/console ./internal/execgw
      ```

      **`internal/journal`은 이 change가 편집하지 않는다.** 그 패키지에 `-race`를
      씌우는 비용은 a092가 만든 것이 아니고, 완주하지 못하는 명령은 검증이 아니다.
      나무 전체에 `-race`를 돌려야 할 이유가 생기면 그것은 **별도 과제**이지
      이 change의 VERIFY 단계가 아니다. **design의 검증 목록도 7판에서 이 두 갈래로
      바꿨다** — 5라운드 B1의 정정이 이 절에만 착지하고 `design.md`에는 안 간 것이
      6라운드 H-1이다.

      **(b)의 5개 패키지는 필요한 것보다 넓고, 그것이 의도다.** `internal/app/engine`을
      import하는 패키지는 태그 유무 모두 `cmd/tossctl` 하나뿐이므로 두 개면 충분하다.
      `internal/obs`·`internal/console`·`internal/execgw`를 더 넣는 이유는 **편집이 닿는
      타입의 소유자**(`obs.Notifier`·`obs.Ntfy`·`execgw.EntryGate`)를 경합 검사 안에
      두려는 것이고, 셋 다 빨라서 비용이 없다. **보수적으로 넓힌 것이지 도달 분석이
      그것을 요구한 것이 아니다**(6라운드 NIT).
- [ ] 9.5 `make lint` (= `gofmt -l` + `go vet ./...`, 저장소에 별도 lint 설정은 없다).
      **`gofmt -l`의 범위는 나무 전체가 아니라 세 경로다** — `./cmd ./internal
      ./tools/logic-map`(`Makefile:95`). a092가 편집하는 두 파일은 `./internal` 아래라
      범위 안이지만, **`make lint` 초록을 "저장소 전체가 포맷됐다"로 읽으면 안 된다**
      (7라운드 NIT). `go vet`만 `./...`이다.
      **`make fmt`가 아니다** — 그것은 `gofmt -w`이고 VERIFY 단계에서 파일을 쓴다
      (10절의 fingerprint 순서를 깨뜨린다). base에서 `gofmt -l`은 비어 있으므로
      `make fmt`는 무해한 no-op이지만, **쓰기 명령을 검증 단계에 두지 않는다**.
- [ ] 9.6 **값 검사와 커버리지 게이트를 다시 돌린다 — 구현이 문서를 낡게 만든 뒤에.**

      ```bash
      python3 openspec/changes/a092-an-alert-does-not-hold-the-stop/tools/check_values.py
      python3 openspec/changes/a092-an-alert-does-not-hold-the-stop/tools/check_values_selftest.py
      python3 openspec/changes/a092-an-alert-does-not-hold-the-stop/tools/coverage_gate.py
      ```

      7.4.0·7.4.0.1은 **구현 전** 상태를 본다. 8절이 상수를 넣고 9.2가 줄 번호를
      갱신하면 문서가 바뀌므로, **그 뒤에 한 번 더 돌지 않으면 갱신이 만든 새
      불일치를 아무도 안 본다.** 9.2의 grep 목록이 빠뜨린 문서가 있으면 여기서
      소리가 난다.

      **커버리지 게이트의 미탐 계열 U1~U7 출력을 완료 보고에 싣는다.** 통과했다는
      줄만 싣고 못 세는 자리를 빼면, 그 보고는 7·8·9판이 세 번 한 것과 같은
      주장이 된다.
- [ ] 9.7 **`alertBudget`이 관측 주기와 묶여 있는지 확인한다** (9라운드 H-1 네 번째 구성).

      ```bash
      grep -n "alertBudget" internal/app/engine/notifications.go
      # 기대: DefaultExitObservationInterval 을 통해 유도된다. 리터럴 `5 * time.Second` 금지
      ```

      > **10판이 이 구성에 답하지 않았다 — 10라운드 H1.** 9라운드 H-1은 네 구성을
      > 열거했고 10판은 세 계열(0·음수·단위)만 답했다. 네 번째는
      > 「`alertBudget`을 주기에서 분리」였고, change 전체에서 `review.md`의
      > 한 줄에만 존재했다. 실측으로 확인됐다: **여섯 단언을 전부 걸어도**
      > `alertBudget = 5 * time.Second`로 리터럴화하면 주기를 1초까지 내려도
      > BUILD OK다. 배열 단언은 상수의 **값**을 보지 그 값이 **어디서 왔는지**를
      > 보지 않는다.
      >
      > **컴파일 타임에 이것을 잡을 방법은 없다** — Go의 배열 길이 단언은 유도의
      > 출처를 표현하지 못한다. 그러므로 이 자리는 **기계가 아니라 grep과 리뷰**로
      > 지킨다. 그 사실을 여기 적는 것이 not-applicable 사유이고,
      > proposal이 "구조적 보증"이라 부른 것의 **정확한 경계**다:
      > 값의 관계는 컴파일러가 지키고, 값의 **출처**는 사람이 지킨다.

## 10. 게이트

**순서가 중요하다.** `make sdd-check`의 `check_index_freshness.py`는 **추적 파일 전부와
비무시 미추적 파일 전부**를 해시한다(`:20-40`). 그러므로 `make sdd-sync`는
**worktree에 쓰는 모든 작업이 끝난 뒤**여야 하고, 여기에는 `review.md`·PM·`issues.md`뿐
아니라 **이 파일의 체크박스를 채우는 것까지** 포함된다.
3판은 sync를 10.1에 두었고 그 순서로는 통과할 수 없다.

- [ ] 10.1 gstack 독립 리뷰 (구현 후)
- [ ] 10.2 `review.md` 갱신 — 구현 후 판정
- [ ] 10.3 PM 동기화 — `STORY-TOS-a092`. **YAML을 고치면 생성기를 반드시 다시 돌린다**:
      `docs/pm/generated/*.md`는 **추적 파일**이고 `make sdd-check`가
      `tools/pm/generate_master_tracker.py --check`로 신선도를 검사한다
      (`Makefile:73` → `generate_master_tracker.py:340-345`, 내용이 다르면
      `generated tracker stale`). 재생성을 빼면 **10절 5단계에서 죽고 3단계로 돌아간다**.

      **⚠ 이 스토리의 수용 기준 8개는 a092가 아니라 a093을 서술한다** (5라운드 H4).
      1·2·3·6·7이 비동기 배달을 전제한다 — *"does not wait for the remote send, its
      retries, or the waits between them"* · *"may be dropped when the send queue is
      full"* · *"retried when the transport recovers"*. **a092는 기다림을 없애지 않고
      유계로 만든다.** a092가 남기는 것은 4·5·8(원장 기록 · 게이트 래치와 승격 유지 ·
      판정과 수량과 시각 무변화)뿐이다.
      그러므로 **10.3은 이 스토리를 done으로 넘기지 않는다.** 할 일은 둘 중 하나다 —
      스토리를 a093으로 옮기고 a092용 스토리를 따로 세우거나, 수용 기준을 a092의
      계약("한 알림이 자기 차례에 쓰는 동기 시간이 관측 주기를 넘지 않는다")으로
      다시 쓰고 나머지를 a093 스토리로 넘긴다. **어느 쪽을 골랐는지 10.6에 적는다.**
- [ ] 10.4 `issues.md` — `flatten.Saga.Notifier` nil 지뢰, `Notifier`에 `Close`/`Stop`이
      없는 점(a093 선행 조건), `alertDeliveryBound` 주석과 `runtime.go:415`의 무기한
      대기(a093에서 정정), `o.Interval()` drift
- [ ] 10.5 memory retain — 검증된 것만
- [ ] 10.6 완료 보고 초안 — **냉 표본 n=1 · 실패 관측 0건 · 손실 ①② · reserve의 로컬성**
      (초과분 관측 범위 31.9~356.1ms, **산포 배수 11.2배** — `delivery-latency` §7.2.1) ·
      **`STORY-TOS-a092`의 수용 기준 8개 중 5개는 a093이 채운다**(10.3에서 고른 처리) ·
      **`make gate`가 태그판을 안 돌린다**(7.3이 유일한 조립 경로 검증) ·
      **`-race`는 폭발 반경 5개 패키지에서만 돌았다**(9.4 — 나무 전체는 완주 불가)
- [ ] 10.7 **이 파일의 미체크 박스를 전부 체크한다 — "여기까지"가 아니라 파일 전체다**
      (이것도 worktree 쓰기다). 10.8~10.11도 포함이다. `tools/gate.sh:107`이
      `grep -c '^- \[ \]'`로 **파일 전체**를 세기 때문이고, 6판의 "여기까지"라는 문구는
      자기 뒤의 4건(10.8~10.11)을 제외해서 아래 blockquote 2단계와 어긋나 있었다
      (6라운드 M2/M-1)
- [ ] 10.8 **`make sdd-sync`** (fingerprint 재고정) — 위 쓰기가 전부 끝난 **뒤**
- [ ] 10.9 `codegraph status` · `openspec validate --all` 로 게이트 전제 확인
- [ ] 10.10 `make sdd-check`
- [ ] 10.11 **`make gate CHANGE=a092-an-alert-does-not-hold-the-stop`** —
      게이트는 미체크 박스가 0일 때만 통과하므로 **마지막이다**

**`openspec archive`는 여기서 하지 않는다.** `docs/WORKFLOW.md:80`이 게이트 뒤에
두지만, a092의 델타가 정하는 계약 중 **`n.mu` 배수**는 a093이 없으면 참이 되지 않고
`STORY-TOS-a092`의 수용 기준 5개도 a093 몫이다(10.3). archive는 **a093이 landed된 뒤**
두 change를 함께 정리하는 것이 맞고, 그렇게 하지 않을 거라면 그 결정을 10.6에 적는다.
**침묵한 생략은 금지다.**

> **4라운드 B2: 위 순서로도 안 된다.** "10.8~10.11을 실행하면서 마지막에 체크"하면
> 그 체크가 또 fingerprint를 바꾸고, `make gate`는 미체크 0을 요구하는 동시에
> `make sdd-check`를 돌린다. 4라운드가 실증했다 — 박스 하나만 뒤집어도
> fingerprint가 바뀐다(13,248개 파일 해시, `check_index_freshness.py:20-41`).
>
> **실제로 통과하는 순서는 이것이다** (각 단계가 fingerprint 중립임을 4라운드가 확인):
>
> 1. 10.1~10.6의 **작업**을 한다.
> 2. **파일에 남은 미체크 박스를 전부 체크한다** — 아직 안 돈 단계의 박스도 포함해서.
>    `tools/gate.sh:106-116`은 `grep -c '^- \[ \]'`로 **파일 전체**를 세므로
>    "10절까지"가 아니라 "파일 전체"가 기준이다(5라운드 B4 — 7.4가 §10 뒤에 있던
>    동안 이 문장이 3건을 빠뜨렸다).
>
>    > **다만 이 면제는 10.8~10.11에만 필요하고, 7판은 그것을 7.4.1~7.4.3까지
>    > 번지게 썼다**(7라운드 M-7). 7판에서 **§7.4는 §6보다 먼저 돌므로**, 여기
>    > 도착한 시점에 7.4.1~7.4.3은 **이미 돌았고 이미 체크돼 있어야 한다.**
>    > 그 셋이 이 시점에 미체크라면 그것은 gate.sh를 달래야 할 상황이 아니라
>    > **§7.4를 안 돌린 것**이고, 미리 체크하면 그 사실이 지워진다.
>    > 미리 체크해도 되는 것은 **이 순서 때문에 물리적으로 나중일 수밖에 없는**
>    > 10.8~10.11뿐이다.
>
>    확인:
>
>    ```bash
>    grep -c '^- \[ \]' openspec/changes/a092-an-alert-does-not-hold-the-stop/tasks.md   # → 0
>    ```
>
>    **이것이 마지막 worktree 쓰기다.**
> 3. `make sdd-sync` — `.sdd/index-state.json`·`.codegraph/`·`.sdd/gbrain-home/`만
>    쓰고 전부 gitignore다(`.gitignore:72-79`). `record_index_state`가 쓰기 **전에**
>    fingerprint를 찍는다(`check_index_freshness.py:61` → `:79`).
> 4. `codegraph status` · `openspec validate --all` (읽기 전용).
> 5. `make sdd-check` — `check_agent_config_sync.py`는 `--generate`일 때만 쓰고,
>    `generate_master_tracker.py --check`는 읽기 전용, `compileall`은 `*.pyc`만(gitignore).
> 6. `make gate CHANGE=…` — **중간에 아무것도 쓰지 않는다.**
>    실패해서 파일을 고치면 **3단계로 돌아간다.**
>
> 병행 세션이 커밋하면 base-commit과 fingerprint가 둘 다 깨지므로 재고정 후 연속 실행한다
> (기억: `tossos-parallel-session-gate-contention`, `stacked-changes-break-the-gate`).
>
> **이 순서는 `docs/WORKFLOW.md:77`과 충돌하고, 7판이 그 충돌을 여기 기록한다**
> (6라운드 M-6 — 4라운드 B2가 승인한 우회인데 근거가 어디에도 안 적혀 있었다).
> WORKFLOW 7단계는 *"각 task 체크는 산출물 커밋과 같은 커밋에서 수행한다"*고 쓴다.
> 위 2단계는 **아직 돌지 않은 10.8~10.11의 박스를 미리 체크**하므로 그 규칙을 어긴다.
>
> 어길 수밖에 없는 이유는 `tools/gate.sh:107`과 `check_index_freshness.py`가
> **동시에 만족될 수 없는 두 조건**을 걸기 때문이다 — 게이트는 미체크 0을 요구하고,
> 같은 게이트가 돌리는 `make sdd-check`는 마지막 쓰기 이후 fingerprint가 그대로일 것을
> 요구한다. 박스를 게이트 **뒤**에 체크하면 첫 조건이 깨지고, 게이트 **중간**에
> 체크하면 둘째가 깨진다. 4라운드가 실증했다.
>
> **그래서 대가를 적는다**: 10.8~10.11의 체크는 "그 단계를 했다"가 아니라
> "그 단계를 이 순서로 할 것이다"라는 뜻이다. 실제 수행 증거는 박스가 아니라
> **10.6 완료 보고에 실린 각 명령의 출력**이고, 게이트가 실패해서 3단계로 돌아가면
> 그 보고가 갱신된다. 박스를 증거로 읽으면 안 된다.

## 산출물 역방향 훑기 — **23개 산출물 전부가 행을 갖는다**

2라운드가 지적한 반복 형태는 **"산출물에 이미 적힌 사실을 다음 문서가 안 쓴다"**였다.
3·4라운드가 같은 형태를 또 찾았다.

**4판은 이 표를 만들었는데도 놓쳤다.** 놓친 것이 정확히 **표에 행이 없던 산출물**
(`notifycritical` — `n.mu`와 `n.escalate`를 적어 두었다)과 **표가 훑지 않은 방향**
(`notify-reach` → `delivery-latency`, "flatten `Notifier`는 nil")이었다.
**행이 없으면 그 산출물을 본 적이 없는 것이다.** 그래서 5판은 23개 전부를 행으로 놓고,
쓰지 않은 것은 **사유와 함께** 빈칸으로 두지 않는다.

| # | 산출물 | 그 산출물이 말하는 사실 | 어느 문서가 쓰는가 |
|---|---|---|---|
| 1 | `exitobserver.alert` | 호출자 7곳 · 동기 · 기한 없음 · defers 0. 절「상한 34초의 유도 — 전부 AST 산출물에서」. **7판이 이 산출물의 "그 `ctx`는 journal 트랜잭션과 공유된다"를 정정했다** — announce는 커밋 밖이고, 호출자 기한의 결과는 잘림이 아니라 `BeginTx` 전 만료다 | proposal 34초 유도, notify-reach P4, design D1 안 B |
| 2 | `exitobserver.alertproposalrefused` | **`"branches": null` · `"returns": null`** — 억제 부재의 열거 증명. 절「이 산출물이 손으로 읽은 증거와 다른 점」·「사이클 총 체류가 유계가 아닌 이유」 | proposal「사이클은 여러 번」, design D2, exit-policy 델타 |
| 3 | `exitobserver.alertrefused` | B1 래치 — 포지션당 1회 | notify-reach P4 표, design D2(래치 있는 쪽과의 대조) |
| 4 | `exitobserver.announcequarantine` | B2 래치 · **`o.alert` 7번째 호출자(다른 파일)** | delivery-latency §1.1의 종류 열거 |
| 5 | `exitobserver.checkoutage` | 한 사이클에 `Notify` 2회 도달 · announce는 커밋 뒤. 절「두절 사이클의 최악 동기 체류」(2 × 34s = 68s — 단 NORMAL 계정에서만). **7판이 「Safety conclusion」의 "짧은 기한을 씌우면 원장 트랜잭션이 잘린다"를 정정했다** — 6라운드 차단 B3, 2라운드 M5의 정정이 네 판째 안 온 자리 | proposal H1, design D1 안 B 기각 |
| 6 | `exitobserver.observeonce` | B5 `range states` · **B6 가격 없는 심볼을 조용히 `continue`**. 절「B6가 이 함수의 결함이다」 | B5는 proposal「사이클은 여러 번」. **B6은 안 쓴다 — a090 대상**(의도적) |
| 7 | `exitobserver.run` | AST calls 순서 `:358`→`:359`. 절「사이클 체류는 주기에 더해진다」 | proposal Why, design D2 예산 기준 |
| 8 | `newnotifier` | branches 0/returns 1/calls 0 · **`Clock: clk`를 채운다**. 절「채우는 필드와 비우는 필드」 | design D7 무변화 증명, **tasks 6.6의 가짜 시계 함정** |
| 9 | `reconciledriver.alert` | defers 0 · 60s 주기 | notify-reach P2, design D5 |
| 10 | `resolvenotificationpublisher` | branches 5/returns 4 · **nil 이탈은 B2 `:77`·B3 `:83`·B5 `:94`** (B1 `:69`는 `getenv == nil`이고 반환하지 않는다). 절「`:101`이 비우는 필드」 — 8.3이 채우는 그 자리다 | design D7, **tasks 6.8**. **5판까지 산출물 자신이 이것을 B1·B3·B5로 잘못 적고 있었다** — 4라운드 M1의 정정이 design·tasks에만 착지하고 FLM·branch-test-map에는 안 착지했다(5라운드 H2). 6판이 `ast.json`을 정본으로 두 파일을 고쳤다 |
| 11 | `runtime.alert` | `alertDeliveryBound = 30s` · **30 < 34, 주석이 거짓**. 절「30초는 어디서 왔는가」 | proposal, design D5 |
| 12 | `runtime.checkhealth` | 주기 1초 · returns 0 · **B5 `takeLatch` 래치** | engine-safety 기준 루프 근거, design D2 |
| 13 | `runtime.escalate` | **`:415`에 기한이 없다** — 면제 근거 (a)의 반증. 절「`alertDeliveryBound` 주석이 거짓이다」 | engine-safety 면제 근거 정정, proposal, design D5 |
| 14 | `runtime.takelatch` | 두절당 1회 — **감독자 면제의 유일한 근거** | engine-safety, design D2 |
| 15 | `journal.enqueuealert` | B5 중복 키가 **기존 행 id를 돌려준다** | delivery-latency §2 — 측정이 성립하는 이유 |
| 16 | `journal.transitionoperatingmode` | B15 → `:415` · 커밋 `:468` · announce `:479`. 절「H1 — 오늘의 두절 사이클이 34초인 이유」·「design D1 — 호출자 기한이 원장을 자르지 못하는 이유」 | proposal H1, design D1. **후자가 5·1번 산출물의 거짓을 반증하는 자리다** |
| 17 | `targetmodefortrigger` | **값을 돌려주는 case가 B2 하나**, 6개 트리거를 전부 담는다. 절「전칭 주장이 성립하는 이유는 case가 하나이기 때문이다」 | proposal H1의 첫 고리, notify-reach P1 (**5판 신설 — 4라운드 H4**) |
| 18 | `notifier.deliver` | **`n.mu`를 재시도 전체(대기 포함) 보유** · `detail`은 `Gate.Block`에만 간다. 절「최악 예산의 산술」 | **`n.mu`는 5판 engine-safety의 배수 SHALL·design D2**(4라운드 C1). `detail`은 안 쓴다 — a093 대상(의도적) |
| 19 | `notifier.notify` | `logEvent`가 **`n.mu` 잡기 전에** 돈다 | delivery-latency §5 한계 4(인접 짝짓기의 goroutine 취약성) |
| 20 | `notifier.notifycritical` | `n.deliver`는 `n.mu` 보유 · **`:187`의 `n.escalate`가 원장 트랜잭션을 돌린다**. 절「기록과 발송의 비대칭이 절단선이다」 — 원장 기록은 먼저 끝나고 발송만 예산 안에 있다 | **5판 — 비전송 작업 열거(design D2·D3 주석·proposal·양 spec·delivery-latency §7.2)**(4라운드 H2) |
| 21 | `notifier.publishbesteffort` | **`n.mu`를 잡지 않는다** — 직렬화되는 것은 critical뿐. 절「이 함수의 체류는 **측정된 적이 없다** — 아래 숫자는 체류가 아니다」: 줄 간격 표본 9건(최대 1.836초, +2.499초 사이클) | `n.mu` 사실은 notify-reach 직렬화 절·delivery-latency §1. **줄 간격 표본은 §5.3이 무효로 판정한 방법이고 6판이 산출물 안에서 그렇게 표시했다**(5라운드 H3) — 유도에 쓰지 않는다 |
| 22 | `notifier.wait` | 프로덕션은 B1을 탄다(`RetryDelay` 0 → 2초). 절「`deliver`가 이 함수를 부르는 횟수」 — 시도 사이에만 도므로 `Attempts-1`회다 | proposal의 34초 유도 표 |
| 23 | `ntfy.publish` | **`Topic` 빈 값이면 다이얼 전 반환**(`:86-89`) · B7 `client == nil` → 전역 풀. **그리고 이 산출물은 "왕복 0.2~1.8초, 표본 13"을 실측이라 불렀는데 그 방법은 무효다** | 앞의 둘은 **tasks 6.6의 실행 가능성**·delivery-latency §3의 전역 풀 논거. **"왕복 1.8초"는 6판이 정정했고, 1.3초를 넘는다는 사실을 design D2가 정면으로 다룬다**(5라운드 H3) |

**안 쓰는 것은 두 개이고 둘 다 사유가 있다**: 6번의 B6(a090), 18번의 `detail`(a093).
그 외 21개 산출물의 사실은 전부 어느 문서엔가 쓰인다.

### 5판의 표도 놓쳤다 — 그리고 방향이 뒤집혔다 (5라운드 H2·H3)

23행을 만든 것으로 부족했다. 5라운드가 세 가지를 찾았다.

| 형태 | 5판에서 일어난 일 |
|---|---|
| **행이 산출물의 내용을 잘못 보고했다** | 10번 행이 `resolvenotificationpublisher`의 nil 이탈을 옳게 적었는데, **그 산출물 자신이 틀린 채로 남아 있었다.** 4라운드 M1의 정정이 design·tasks에만 착지했다. 표를 채울 때 **문서 쪽 값을 옮겨 적었기 때문**이다 — 산출물을 열지 않았다 |
| **행이 산출물의 반증 데이터를 안 적었다** | 21·23번 행이 `publishBestEffort`·`ntfy.Publish`가 담은 "왕복 1.8초"를 명명하지 않았다. 그 값은 채택 상한 1.3초를 넘고, 같은 change의 §5.3이 그 방법을 무효로 판정했다 |
| **정정이 문서에만 착지했다** | `alertproposalrefused`의 `[]`→`null`도 같은 형태 (5라운드 M4) |

**이번엔 문서가 참이고 산출물이 거짓이었다.** 2~4라운드와 방향이 반대다.
그래서 훑기의 규칙에 "표를 만든다"로는 부족하고 **표를 무엇으로 채우는가**가 들어간다.

## a093으로 이관한 것 (여기서 하지 않는다)

| 항목 | 근거 |
|---|---|
| 배달 루프 신설, 기록/전송 분리 | 1라운드 C1 (landed 16건 재작성) |
| **사이클 총 체류의 상한** | 2라운드 H2 — 알림당 예산으로 도달 불가 |
| **`n.mu` 교차 루프 경합** | 2라운드 C3 |
| `Runtime.Run` 종료 순서와 루프 실패 알림 | 1라운드 C2 |
| 종료 시 backlog 배수와 그 ctx | 1라운드 C3·H4 |
| `Fields` 복원 계약 | 1라운드 H2 |
| `PendingAlerts` 순서와 backlog 깊이 | 1라운드 H5 |
| 정상 등급 큐 용량·drop 정책 | 1라운드 H6 |
| `alertDeliveryBound` 주석 문구 정정 | 1라운드 C3 |
| **`runtime.go:415` `EscalateOperatingMode`의 무기한 대기** | 3라운드 H2(a) |
| base spec 루프 열거의 `strategy-entry` drift | — |
