# a102 Design — 부팅이 엔진을 굶기지 않는다

이 문서는 Manager가 결정을 확정한 설계다. Teammate는 여기서 설계하지 않는다 —
구현하고, 여기 없는 결정이 필요해지면 **멈추고 되묻는다**.

## 결정 목록

### D1 — 429 판별은 `errors.Is(err, official.ErrRateLimited)` 다

sentinel은 이미 있다: `internal/official/errors.go:15`
`ErrRateLimited = errors.New("official: rate limited")`.

같은 패턴의 선례 셋: `internal/execgw/classify.go:114`, `internal/execgw/retry.go:62`,
`internal/candidatesrc/candidatesrc.go:253`. 서베이도 같은 오류로 재시도를 건다
(`internal/soak/soak.go` `retryRateLimited`). 문자열 매칭은 금지다.

### D1b — Collect의 원인 wrap을 `%w`로 바꾼다 (전제 조건)

**지금 이대로면 D1은 동작하지 않는다.** `Collector.Collect`는 원인을 `%v`로 감싼다:

```
internal/reconcile/snapshot.go:248  "%w: walking the open-order list: %v"   ← %v가 정체를 지운다
internal/reconcile/snapshot.go:253  "%w: %v"
internal/reconcile/snapshot.go:263  "%w: sweeping the holdings: %v"
internal/reconcile/snapshot.go:271  "%w: reading the %s buying power: %v"
```

`%v`는 wrap이 아니므로 `errors.Is(err, official.ErrRateLimited)`가 이 지점에서 끊긴다.
네 곳 모두 원인 쪽 동사를 `%w`로 바꾼다. Go 1.25(go.mod)이고 다중 `%w`는 1.20부터
지원된다. `ErrPartialSnapshot`의 `errors.Is`는 그대로 성립하므로 기존 소비자
(reconcile driver의 mismatch 경로)는 무변이다 — 보수 방향.

`Collector.Collect`는 기존 함수이므로 **FLM을 먼저 만든다**
(AST는 `analysis/function-logic/internal-reconcile--collector.collect/ast.json`,
branches 8 / returns 6 / calls 17).

### D2 — 겹1의 수정 지점은 `stableSnapshot` 하나다

관측된 사망 지점이다 (2026-08-13 02:03:30.545Z, `engine.log`). 주기 reconcile은
이미 다음 주기가 받는다 — 2026-08-12T12:52:37Z `reconcile.mismatch`
"will be retried next period" 실측. `Recovery.Run`의 replay/observation 경로는
**not-applicable**: 관측된 실패가 없고, High-risk 경로의 최소 보수 편집 원칙.

### D3 — 429는 attempt를 소모하지 않고, 15s 백오프, 총 대기 예산 5m

`internal/reconcile/recovery.go` `stableSnapshot` (AST: branches 5 / returns 4 /
calls 7 — 재시도 루프와 `clk.Sleep` seam이 **이미 있다**):

- `Collect`가 429로 실패하면 **stabilisation attempt를 소모하지 않는다.**
  읽기가 없었으므로 스냅샷이 없고, 안정화 판정의 입력이 아니다. `taken`도 안 센다.
- `RateLimitBackoff`(기본 15s) 기다린 뒤 다시 읽는다. 15s는 서베이의 규율과 같다
  — 보호를 든 쪽이 조회만 하는 쪽보다 먼저 포기하지 않는다는 하한.
- 한 recovery 실행의 429 대기 총합이 `MaxRateLimitWait`(기본 5m)을 넘으면
  기존과 같이 `ErrRecoveryIncomplete`로 fail-closed. 메시지에 rate limit과
  기다린 시간을 적는다 — 게이트는 닫힌 채 운영자가 본다.
- 429가 아닌 오류는 **오늘과 동일하게 즉시 실패한다.** 회귀 테스트로 고정한다.
- 백오프 sleep은 기존 `clk.Sleep(ctx, …)` seam — ctx 취소가 즉시 통과해야 한다
  (불변식 4: 종료 지연 금지). 테스트로 고정한다.

구현 형태: `Stabilisation`에 `RateLimitBackoff time.Duration` /
`MaxRateLimitWait time.Duration` 필드 추가, zero값이면
`DefaultRateLimitBackoff = 15s` / `DefaultMaxRateLimitWait = 5m`
(기존 Interval/Required/MaxAttempts와 같은 zero-default 패턴, `withDefaults()`).

관측: `Report`에 `RateLimitWaits int`와 `RateLimitWaited time.Duration` 추가.

> **A1 정정 (2026-08-13)**: 초판의 "기존 report 소비 지점에 그대로 실린다"는
> 전제가 틀렸다 — **소비 지점이 없다.** `cmd/tossctl/engine.go:402`의 Recover
> 클로저가 `_, rerr := recovery.Run(ctx)`로 Report를 버린다. 이대로면 최대 5분의
> 대기가 로그 한 줄 없는 무음 구간이 된다 (A1 F1). 소비는 그 클로저를 어차피
> 편집하는 **T2의 D5**에 붙인다: `RateLimitWaits > 0`이면 obs 이벤트 한 줄.
> `internal/reconcile`에 로거를 넣지 않는다는 원칙은 유지된다.

**왜 총 대기 예산이고 횟수가 아닌가**: 운영자에게 의미 있는 수는 "보호가 없는
시간"이다. 5m 근거: KRX 아침 429 창(분 단위 예산)을 여유 있게 덮되, 죽은 브로커를
조용히 영원히 기다리지 않는 상한.

### D4 — ready 신호는 `enginelock.Marker`의 `ReadyAt`이다

- `Marker`에 `ReadyAt *time.Time` (JSON `ready_at,omitempty`) 추가.
  **restart recovery가 완주한 뒤에만** 쓴다. 부재 = 준비 안 됨.
- `Hold`의 반환을 핸들로 바꾼다: `Release()`와 `Ready(now time.Time)`를 가진
  구조체. 프로덕션 `Hold` 호출자는 `cmd/tossctl/engine.go:239` **하나**다.
  `Ready`는 멱등이고, refresh 루프는 내부 상태(뮤텍스 보호)에서 ReadyAt을 읽어
  **보존**한다 — refresh가 ReadyAt을 지우면 안 된다는 테스트를 만든다.
- 읽기: `Read`는 관대한 reader다(기존 주석). Marker 필드 추가로 충분한지
  teammate가 확인하고, `Status`에서 ReadyAt이 보이게 한다.
- `Hold`는 기존 함수 — **FLM 먼저**. AST는 teammate가
  `tools/logic-map`으로 생성한다.

**D4b (A2 F1 정정) — 준비 신호는 살아 있는 그 프로세스의 것일 때만 신호다.**
크래시는 `Release`를 부르지 않으므로 `ready_at`이 든 마커가 StaleAfter(5m) 동안
디스크에 남고, 신선도만 보면 "죽은 엔진이 계좌를 지키고 있다"고 읽힌다 — A2가
실행으로 재현했다(02:03 사고의 바로 다음 장면: 콘솔이 새 엔진을 띄우고, 새
엔진이 조립 단계에 있는 동안 전임자의 ready_at이 '준비 확인'으로 읽혀 서베이가
새 엔진의 복구 예산을 때린다). 따라서:

- ready 판정은 `ReadyAt != nil` **그리고 마커의 PID가 지금 살아 있는 엔진
  프로세스 중 하나일 때만** 참이다. `Marker.PID`는 이미 있고, 콘솔 쪽에는
  `engineFindProcesses(dir)` seam이 이미 있다 — cmd 쪽에서 합성한다.
  enginelock 패키지에 프로세스 열거를 넣지 않는다.
- 살아 있는 엔진 프로세스가 있는데 마커 PID가 그 집합에 없으면: 그 마커는
  시체의 것이다 — "준비 안 됨"으로 취급하고 계속 기다린다 (새 엔진이 flock을
  잡으면 자기 PID로 마커를 다시 쓴다).
- 살아 있는 엔진 프로세스가 없으면: 기존 D6대로 즉시 시작이다.

**D4c (A2 F3·F4 정정) — 마커 쓰기의 규율.** 뮤텍스는 메모리만 덮고 파일은
덮지 않았다. A2 실측: ready_at이 디스크에서 지워지는 순서(139/3000), 독자의
찢어진 읽기(3617/12259, `os.WriteFile`의 O_TRUNC 창), 그리고 `Release` 뒤
refresh가 마커를 부활시키는 갈래(결정적 재현 — 편집 전에는 "잘못 그린 상태
줄"이었지만 이제 "죽은 보호가 서 있다"가 된다). 따라서:

- `write`는 뮤텍스 안에서 한다 (1분 1회 — 비용 없음).
- 파일 교체는 tmp+rename으로 원자화한다 (찢어진 읽기 제거).
- `Release`는 뮤텍스 아래에서 live 플래그를 끄고, refresh는 그것을 확인해
  쓰기를 건너뛴다 (부활 차단).

**D4b-2 (gstack 리뷰 P1 정정 — Codex 2패스·Claude 적대 3모델 수렴) — PID는
인스턴스가 아니다.** 컨테이너 recreate는 D4b의 PID 등식을 뚫는다: journal
볼륨은 살아남으므로 전임자의 마커가 신선한 채(`refresh` 1분 주기 < StaleAfter
5분) `ready_at`+PID P를 담고 있고, **재생성된 PID namespace는 PID 배정이
사실상 결정적**이라 교체 엔진이 같은 P를 받을 수 있다. 교체 엔진이 pgrep에
보이는 시점부터 step 6 `Hold`가 마커를 덮어쓰기 전까지 몇 초 동안
`pid==P ∧ ready_at≠nil`이 성립해 서베이가 교체 엔진의 복구 예산을 때린다 —
바로 우리 배포 flow의 시나리오다. 따라서:

- 마커에 **프로세스 인스턴스 토큰**을 추가한다: `/proc/sys/kernel/random/boot_id`와
  `/proc/<pid>/stat`의 starttime ticks를 합친 opaque 문자열
  (`json:"proc_instance,omitempty"`). 시계 환산·스큐 산술 없이 **정확 일치**로
  비교한다 — PID 재사용은 starttime이 다르고, 재부팅은 boot_id가 다르다.
- 토큰 계산은 cmd 쪽 seam이다(enginelock은 OS 열거 무지를 유지한다 — D4b와
  같은 선). 엔진은 `Hold`에 자기 토큰을 넘겨 마커에 싣고, 콘솔은 산 후보 pid의
  토큰을 읽어 마커의 것과 비교한다.
- ready 판정은 이제 세 가지 전부다: `ReadyAt != nil` ∧ 마커 PID가 산 집합에
  있음 ∧ **그 pid의 인스턴스 토큰이 마커의 것과 일치**.
- 토큰 부재(구 마커·비Linux dev)·읽기 실패 → not-yet (보수 방향: 상한까지
  기다린 뒤 시작 — "모름을 준비됨으로 읽지 않는다").

### D5 — 마킹은 cmd의 Recover 클로저에서 한다

런타임은 `opts.Recover`를 루프 시작 전에 부른다(`internal/app/engine/runtime.go:289-294`).
그 클로저를 만드는 곳도, marker를 쥔 곳도 cmd다. 따라서:

```go
// cmd/tossctl — 형태만; 정확한 코드는 teammate 몫
recoverThenReady(run func(ctx context.Context) error, ready func()) func(ctx context.Context) error
```

`run` 성공 시에만 `ready()`. 실패 시 절대 부르지 않는다. 인자 받는 함수로 빼서
테스트가 닿게 한다 (`runEngineRun`은 커버리지가 낮다 — a098 FLM 실측).
`internal/app/engine`은 **무변경**이다.

**D5b (A1 F1 반영)**: 같은 클로저가 지금 버리는 `reconcile.Report`를 받아
`RateLimitWaits > 0`이면 obs 이벤트 한 줄을 남긴다 — 몇 번, 총 얼마나 기다렸는지.
운영자가 "보호가 없던 시간"을 사후에 셀 수 있어야 한다. 성공·실패 경로 모두다.

`runEngineRun`은 기존 함수이고 클로저 배선이 바뀌므로 **FLM 재기준**이 필요하다
(a098의 FLM은 다른 base다).

### D6 — 콘솔 대기는 `awaitEngineReady`, cap 120s · poll 2s

`cmd/tossctl/soakautostart.go`에 새 함수:

```go
awaitEngineReady(ctx, observe func() enginelock.Status, clk, cap, poll) verdict
```

- marker 부재 또는 !Running → **기다리지 않는다** ("엔진이 없다" — 오늘의 동작).
  엔진이 1초 만에 죽은 오늘 같은 경우 120s를 헛기다리지 않는 조건이다.
- Running && ReadyAt 있음 → ready.
- Running && ReadyAt 없음 → poll 간격으로 재관측, cap까지. cap 초과 →
  **그냥 시작하되 verdict가 그렇게 말한다.** 서베이는 선택 기계장치이고
  (a101, soakautostart.go:83-86), cap 초과 후의 경합은 겹1이 받는다.
- ctx 취소(콘솔 종료) → 시작하지 않는다는 verdict.

**cap 120s 근거**: 실측 recovery ~50s(서베이 주기 01:35:02→01:35:52) + 429 여유.
겹1의 5m 예산보다 짧은 것은 의도다 — cap은 조용한 대기의 상한일 뿐, 그 뒤의
경합은 겹1이 생존시킨다. **poll 2s**: marker refresh보다 촘촘하면 낭비, 엔진
준비 지연을 초 단위로 감지하면 충분.

새 상수는 설정 노브가 아니다(YAGNI) — 근거 주석을 단 상수로 둔다.

**D6-2 (gstack 리뷰 3패스 수렴 정정) — cap은 겹1의 예산에서 유도한다.**
120s는 *측정된* 2회 판독(~50s) 기준인데, 그 측정의 판독당 비용(~24s — 2363건
주문·26페이지가 실제 운영 계좌다)으로 5회 판독 아침을 계산하면 스로틀 **없이도**
~128s다. 체결이 착지 중인 개장 계좌가 바로 판독 수가 늘어나는 계좌이고, 상한
초과는 fail-open이라 서베이가 복구가 아직 쓰는 예산을 때린다 — 상한이 평범한
아침에 걸리면 겹2의 목적이 그 꼬리에서 무너진다 (Claude 적대 F1, performance,
security 독립 수렴). 따라서:

- `engineReadyCap = reconcile.DefaultMaxRateLimitWait` (5m) — 상수 하나가 양쪽
  절반을 지배한다. 서베이는 엔진의 429 예산이 소진되기 전에는 조용한 대기를
  포기하지 않는다: "보호를 든 쪽보다 조회만 하는 쪽이 먼저 포기하지 않는다"의
  대기 버전이다.
- 엔진이 죽으면 absent가 **매 회** 확인되므로 큰 상한이 서베이를 불필요하게
  미루지 않는다 (D6의 기존 규칙 그대로). 이제 cap 초과는 평범한 아침이 아니라
  "살아는 있는데 5분 넘게 준비를 말하지 않는 엔진"만을 뜻하고, 노트의 문장도
  그만큼 무거워진다.
- poll 2s는 유지 — 근거 불변.

### D7 — 대기는 goroutine이고 기존 함수는 안 바꾼다

- `runConfiguredSoakAutostart`·`bootSurvey`는 **무편집**. 대기는 start seam을
  감싼다: `start = func() { verdict := awaitEngineReady(…); note := bootSurveyIfAbsent(); … }`.
  autostart가 OFF면 start가 안 불리므로 대기도 없다 — 공짜로 옳다.
- `runConsole`은 soak 블록을 `go func(){ note := …; fmt.Fprintln(stderr, note) }()`로
  바꾼다 — **호출과 print만** (a101 패턴: 판단은 seam 함수에, runConsole은 0.0%).
  콘솔 화면은 어떤 경우에도 대기하지 않는다 (a101: "no operator screen" 금지).
- 노트는 어느 쪽이었는지 말한다: 준비 확인/엔진 없음/cap 초과/콘솔 종료.
  **조용한 cap 초과는 금지다.**
- 버튼 경합: 대기 중 운영자가 soak 재시작을 누르면 button 경로가 먼저 띄우고,
  goroutine의 `bootSurvey`가 pid를 보고 "이미 실행 중"으로 물러난다 — 기존 가드가
  그대로 맞는 동작이다.
- `runConsole`은 기존 함수 — **FLM 재기준** (calls-only delta 기대, a101 선례).

**D7b (A2 F5 정정) — 부팅 경로와 버튼은 직렬화된다.** 편집 전 자동 시작
블록은 리스너보다 앞에서 동기로 끝나 버튼과 겹칠 수 없었다. goroutine화가
최대 120초의 동시 창을 새로 만들었고, 두 경로 모두 같은 record 위의
check-then-act라 잠금 없이는 한 기록에 두 서베이가 붙을 수 있다(soakproc가
스스로 "더 나쁘다"고 적은 결말). 콘솔 프로세스 안의 뮤텍스 하나로 부팅
goroutine의 시작 경로와 버튼의 restartSoak 경로를 직렬화한다. 대기 중 버튼을
누르면 버튼이 먼저 잡고, 부팅 경로는 그 뒤 pid 가드("이미 실행 중")로
물러난다 — 순서가 아니라 잠금이 그것을 보장하게 만든다.

**D7c (A2 F6 정정) — 형태가 아니라 실행으로 고정한다.** 소스 형태 테스트는
"goroutine 안에 join을 넣는" 코드에 뚫린다(A2가 실제로 뚫었다). 부팅 블록의
동기 부분을 이름 있는 함수(예: `startSoakAutostartAsync`)로 빼고, 그 함수가
**영원히 블록하는 start를 줘도 즉시 반환**함을 실행으로 단언한다. runConsole의
형태 검사는 그 함수를 부른다는 것 하나로 줄인다.

**D5c (A2 F2 정정) — 배선 자체가 테스트에 잡혀야 한다.** A2의 생존 뮤테이션
넷(N1 ready no-op · N2 관측 무력화 · N3 cap/poll 맞바꿈 · N5 ready=nil)은 전부
"단위는 100%인데 프로덕션 배선은 0%"의 결과다. stubRuntime이 ready를 버리지
않고 잡아서 호출 여부를 단언하고, `engineReadiness`·cap·poll의 전달을 이름
있는 seam으로 빼서 단위 테스트한다. 최소 N1·N2·N3을 죽인다.

## 겹의 결합 — 왜 둘 다인가

| 시나리오 | 겹1만 | 겹2만 | 둘 다 |
|---|---|---|---|
| 부팅 경합 (오늘 02:03) | 살아남지만 매번 백오프 낭비 | 대부분 회피, cap 초과 시 죽음 | 회피 + 초과해도 생존 |
| 장중 429 (08-12 12:52) | 이미 다음 주기가 받음 + recovery도 생존 | 무방비 | 생존 |
| 엔진 즉사 (마커 부재) | — | 헛기다림 없이 서베이 시작 | 동일 |

## 범위 밖 (proposal과 동일 + 확정 사유)

- `engineStartProbe`의 결과 기반 전환 — 겹2가 서베이를 신호에 묶으면 probe는
  서베이 기동과 무관해진다. probe 주석의 거짓 전제는 그대로 남지만, 그 주석을
  고치는 것은 별도 change다.
- 자동시작 supervisor화 — 이 change는 죽지 않게 하는 쪽이다.
- `Recovery.Run`의 replay/resolution 경로 429 — 관측된 실패 없음, not-applicable.
  **gstack 리뷰가 정밀화** (Claude 적대 F3): 두 호출부(`recovery.go`의 replay·
  resolution wrap)가 `%v`라 429 정체가 지워지고, step 2는 in-doubt attempt가
  있을 때만 돌므로 **dirty-crash 재시작**에서는 a102의 인내가 닿지 않는다 —
  실패 양식은 오늘과 같은 fail-closed다. 여기에 백오프를 넣는 것은 조회 재시도가
  아니라 **취소·인수 의미론이 걸린 경로의 재시도 설계**라 별도 change가 필요하다.
  후속 후보로 등록한다.
- **취소가 `Runtime.Run` 밖으로 나가는 계약** (A2 F9 후반) — recovery가 nil을
  돌려준 채 ctx가 이미 취소돼 있으면 `rt.Run`이 `context.Canceled`를 반환해
  "nil = 정상 배수" 주석과 어긋난다. 고치려면 `internal/app/engine` 편집이
  필요한데 그것은 이 change의 못이다 — 선언된 생략으로 남기고 후속 후보로
  기록한다. cmd 쪽 ctx nil 방어(F9 전반)만 이번에 한다.
- **비상 청산(flatten)의 429** — `internal/flatten/liquidate.go:596`이 같은
  `Collect`를 쓰므로 D1b 이후 429의 정체가 그곳까지 도달하지만, flatten은 그것을
  구분하지 않고 청산 라운드를 접는다 (A1 F5 발견). **이 change에서 고치지 않는다**:
  청산은 High-risk 중에서도 비상 경로라 인내를 넣는 방향조차 별도 승인이 필요하고
  (기다리는 청산이 옳은가 자체가 설계 질문이다), 관측된 실패 사례도 아직 없다.
  후속 change 후보로 등록하며, 침묵한 생략이 아니라 **선언된 생략**이다.

### gstack 리뷰(§5.3)가 추가한 선언된 생략

- **429 재시도의 전체 재수집** (performance·Codex 수렴) — 스로틀된 Collect
  재시도는 1페이지부터 다시 걷는다(26페이지 계좌에서 최악 ~546 페이지 요청/예산).
  15s 간격이 요청 속도를 묶지만, cursor 재개는 `execgw.ScanOrders` 계약 변경이라
  별도 change다.
- **대기 중 관측·대시보드의 Ready 표기** (Claude 적대 F6, Codex 7·8) — obs 한
  줄은 복구가 끝난 뒤에 남고, `Status.Ready()`를 그리는 화면이 없다. 5분 창을
  실시간으로 보려면 콘솔 템플릿 변경이 필요하다 — 후속 후보.
- **취소 중 부분 백오프의 미계상** (Claude 적대 F9) — 첫 백오프 중 취소면
  `RateLimitWaits=0`이라 observer가 침묵한다. 운영자 주도 종료 경로의 관측
  정밀도 문제이고 보호 경로가 아니다 — 기록만 한다.
- **pgrep 실행의 무제한 대기** (Codex 6) — `exec.Command(...).Output()`에
  context·timeout이 없어 procfs가 멈춘 host에서는 관측 goroutine이 멈춘다.
  콘솔 화면은 안 막히고(분리 goroutine) 서베이만 안 뜬다. pgrep seam은 a102
  이전부터 있던 표면이라(다른 소비자 셋 공유) 일괄 timeout 도입은 별도 change다.
- **spawn 게이트의 프로세스-로컬성** (Codex 5) — 겹치는 콘솔 둘은 여전히 이중
  spawn이 가능하다. a102 이전에도 같은 check-then-act가 게이트 없이 있었고,
  D7b는 비동기화가 **새로 연** 프로세스 내 창만 닫았다. 교차 프로세스 배타는
  soakproc의 flock 도입이라 별도 change다.
- **콘솔 stderr 콜백의 미조인 goroutine** (Claude 적대 F11) — 운영은
  `os.Stderr`라 무해하고, cobra의 버퍼 writer는 테스트에서만 쓰인다. 기록만.

**겹1을 겹2 없이 운영에 올리지 않는다.** 겹1 단독이면: marker는 recovery보다
먼저 생기므로 대기 5분 내내 "Running·신선"이고, T2 이전의 콘솔은 그것만 보고
부팅 서베이를 시작해 **recovery가 기다리는 바로 그 rate 예산에 읽기를 더한다.**
"엔진은 살아 있는데 보호는 없는" 창이 즉사(~0초)에서 최대 5분으로 넓어진다.
두 겹은 한 change이므로 같이 배포되는 것이 기본이고, 이 절은 부분 배포(예: §1
커밋만 cherry-pick)를 금지 사유와 함께 기록한다.

## Spec delta

- `reconciliation` — ADDED: restart recovery는 브로커의 rate limit을 영구 실패와
  구분해야 하고(SHALL), 그 대기는 유한해야 하며, 대기 동안과 소진 후 모두
  게이트는 닫혀 있어야 한다.
- `operator-console` — ADDED: 부팅 서베이는 엔진의 준비 신호를 유한 시간
  기다려야 하고(SHALL), 그 대기가 운영자 화면을 지연시켜서는 안 되며(SHALL NOT),
  어느 쪽으로 끝났는지 노트가 말해야 한다.

## Teammate 배정

| | 범위 | 편집 파일 | FLM 대상 (편집 전 필수) |
|---|---|---|---|
| **T1** | 겹1 + D1b | `internal/reconcile/recovery.go`, `internal/reconcile/snapshot.go` | `Recovery.stableSnapshot`, `Collector.Collect` (+`withDefaults` 등 편집하게 되는 기존 함수 전부) |
| **T2** | 겹2 | `internal/enginelock/enginelock.go`, `cmd/tossctl/engine.go`, `cmd/tossctl/soakautostart.go`(추가만), `cmd/tossctl/console.go` | `Hold`, `runEngineRun`, `runConsole` (+편집하는 기존 함수 전부) |

T1이 먼저다(독립적이고 High-risk 코어). T2는 T1의 커밋 위에서 작업한다.
각 teammate 뒤에 전담 적대 리뷰가 붙고, 지적은 그 teammate가 고친다.

## 검증 계약 (내가 이것으로 완료를 판정한다)

1. RED가 실제로 먼저 있었다는 증거 (실패 출력 인용).
2. 뮤테이션 증거 — 최소: 겹1 (a) attempt 미소모를 소모로 바꾸면 실패하는 테스트
   (b) 예산 제거 시 실패하는 테스트 (c) `%w`를 `%v`로 되돌리면 실패하는 테스트.
   겹2 (d) refresh가 ReadyAt을 지우게 만들면 실패 (e) recovery 실패에도 Ready를
   부르게 만들면 실패 (f) cap 없이 무한 대기하게 만들면 실패.
3. `go test ./internal/reconcile ./internal/enginelock ./cmd/tossctl` 통과 +
   `make lint` rc=0.
4. 기존 테스트 무회귀 (특히 reconcile·enginelock 기존 스위트).
5. FLM/BTM이 편집 **전** 생성되었고 gate 5/9가 통과한다.
