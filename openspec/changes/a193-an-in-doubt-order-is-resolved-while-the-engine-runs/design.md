# a193 · 설계

> 분기 인용 `B<n>(kind@line)`·`call@line`·`return@line` 은 전부 `analysis/cited-ast/*.ast.json`(HEAD `791fb6ef1`, `GOFLAGS=-trimpath go run ./tools/logic-map`,
> 40개 전부 rc 0)의 `branches[].id`·`calls[].at.line`·`returns[].at.line` 이다. 요약은 `analysis/cited-ast/summary.txt`. AST 밖(SQL 문자열·상수·복합
> 리터럴·Go 표준 라이브러리)은 「손으로 읽음」으로 표시하고 `analysis/cited-ast/README.md` 표에 모았다. 운영 사실은 Manager 실측(proposal.md Why)만 쓴다.

## 1. 지금의 기전

### 1.1 IN_DOUBT 를 해소하는 자리

| 자리 | 언제 | 근거 |
|---|---|---|
| `reconcile.Recovery.Run` | 기동, 루프 시작 전 한 번 | `Runtime.Run` B1(if@289)·B2(if@294 → return@295 오류면 아무 루프도 안 뜸) 뒤에야 B3(range@304, 감독 루프)·B4(range@322, 보조 실행자). `Recovery.Run` 은 `RecoverPending`(call@253) → `PendingAttempts`(call@266) → B3(range@270) 안 B4(if@271 IN_DOUBT 아님) 갈래 외 → `replay`(call@291) → 미종결이면 `Resolve`(call@300) |
| `flatten.Saga.resolveStep` | 수동 flatten(별도 CLI 프로세스) | `Resolve`(call@428) — 대상은 그 saga 의 취소 step 의 attempt 뿐(`settleStep`; 청산 매도 `Saga.sell` 은 IN_DOUBT 를 step 상태로만 적음 — B10 case@516) |

CodeGraph 1.6.0 `callers Resolve`(이름 단위, 15건)에서 `execgw.Resolver.Resolve` 의 production 호출자는 위 둘(`analysis/code-context/codegraph-baseline.md`), `rg '\.Resolve\(ctx'`
(시험 제외)로 재확인: `internal/flatten/flatten.go:428`, `internal/reconcile/recovery.go:300` 2건.

### 1.2 세션 중 IN_DOUBT 가 생긴 뒤 일어나는 일 (사고의 기전)

1. `ExitObserver.submit` B6(if@1553) 보호면 Critical 우선순위 → `Submit.Place`(call@1556, 요청에 `Baseline` 없음 — 복합 리터럴 L1556~1561, 손으로 읽음).
2. 결과 IN_DOUBT → B10(case@1577) → B11(if@1582 UNRESOLVED 면 타입만 바꿈) → `noteExitOrder`(call@1585) → return@1586 nil. `noteExitOrder` 는 B1(if@199)·B2(if@202
   거절 아님)를 지나 대기열 append(append@219) — **구조화 로그 줄**이고 outbox 행이 아니다.
3. 발의는 무장된 채다. 다음 주기 판정 진입의 `noteHeldProposal`: B1(if@127 무장)·B2(if@130 intent 있음) → `ExitIntentAttempts`(call@135) →
   `exitIntentAttemptsQ` 에서 IN_DOUBT 는 B7(switch@210) 의 B11(case@216 — 소스는 `default:` 절, AST 종류 표기는 case) → `live` → B14(switch@228) B15(case@229) `ExitIntentLive`. 그래서 `noteHeldProposal`
   B4(if@140 Unaccepted) 아님, B6(range@148 Parked) 0회, B7(range@154 Awaiting) 0회, B10(range@166) 은 미체결 매도의 취소 대기만 봄 → return@172 false.
   **IN_DOUBT 를 알리거나 해소하는 갈래가 0이다.**
4. 같은 종목의 새 mutation 은 게이트웨이가 막는다: `Gateway.checkSymbolFree` → `unsettledFor`(call@804) → B2(if@808) → `reject(ReasonSymbolInFlight)`(call@810).
   그래서 손절 재제출도 멈춘다(사고의 23:57:11 뒤 `exit_events` 끊김과 일치 — 원인 추정이며 운영 로그로 확인한 것은 시각뿐).
5. 다음 기동의 `Recovery.Run` 이 그 attempt 를 해소한다 — 사고에서는 36.8시간 뒤.

### 1.3 `Resolver.Resolve` 의 갈래 (PLACE)

`Resolve`(indoubt.go:229-357, 분기 26 · return 18):

| 갈래 | 조건 | 결과 |
|---|---|---|
| B5(if@253) | 이미 종결 | return@257 — 상태 그대로, 오류 없음(멱등) |
| B6(if@259) | IN_DOUBT 아님 | return@260 `ErrNotResolvable` |
| B9(switch@274) B10/B11 | CANCEL/AMEND | 다른 절차(return@276·@278) — 이 change 대상 아님 |
| B13(for@291) → B14(switch@294) | 스캔 루프 | |
| B15(case@295) | 스캔 오류 | B16(if@298) `!waitOrExpire` → `park`(return@299) |
| B17(case@304) | 정확히 1건 | `ResolveConfirmed`(call@311) → B18(if@311) 쓰기 오류면 return@313, 아니면 return@315 CONFIRMED |
| B19(case@317) | 2건 이상 | `park`(return@325) |
| B21(case@329) | 이번엔 없음 | B23(if@336) `stable && longEnough` 면 `absenceCorroborated`(call@337) → B24(if@338) 증명이면 `ResolveFailed`(call@343) → return@346 FAILED_CONFIRMED, 아니면 `park`(return@348) |
| B26(if@350) | 아직 증명 전 | `!waitOrExpire` → `park`(return@351) |

`waitOrExpire`(360-365): B1(if@361) 다음 주기가 기한 넘으면 return@362 false, 아니면 return@364 `clk.Sleep(ctx, …) == nil` — **ctx 취소도 false**.

`absenceCorroborated`(445-500): B1(if@448 계좌 판독기 없음) → B2(if@459) `windowContaminated`(call@459) 오염이면 return@460 → B3(if@463) **`DecodeBaseline` 실패면
return@464 「without a pre-dispatch account baseline」**(사고의 park 사유 원문과 같은 자리) → 보유·매수가능 읽기(call@467·@475) → B7(if@481)·B8(if@492) → return@498 증명.

`park`(372-384): `ResolveUnresolved`(call@376) → **B1(if@376) 쓰기 오류면 return@377 — `Gate.Block` 앞** → B2(if@379) `Gate.Block`(call@380) → return@383.

### 1.4 park 의 쓰기와 ctx (D3 의 근거)

- `Attempt.ResolveUnresolved` 는 분기 0, `transition`(call@78) 하나(resolution.go:77-84).
- `Attempt.transition`(durability.go:548-674) 은 `a.j.db.BeginTx(ctx, nil)`(call@553) → B1(if@554) 오류면 return@555. 상태 읽기 `tx.QueryRowContext`(call@560) …
  끝에 `tx.Commit`. `j.db` 는 `*sql.DB`(journal.go:86, 손으로 읽음).
- **Go 표준 라이브러리(손으로 읽음, go1.26.5 `database/sql/sql.go`)**: `DB.BeginTx` → `DB.conn`(:1316) 이 연결을 얻기 전에 `ctx.Done()` 을 검사해 닫혔으면
  `ctx.Err()` 를 돌려준다(:1322-1328). `Tx.Commit`(:2287) 도 커밋 전에 tx 의 ctx 를 검사해 닫혔으면 커밋하지 않는다(:2291-2298).
- 따라서 **취소된 ctx 로 들어온 park 는 원장을 바꾸지 못하고**(B1 return@377), 진입 게이트도 걸지 않는다(`Gate.Block` 은 B2 — 쓰기 성공 뒤). `Resolve` 는
  `res.State = UNRESOLVED_IN_DOUBT`(메모리 값)와 **오류**를 같이 돌려준다 — 호출자가 오류를 보고 결과를 버리면 원장의 attempt 는 IN_DOUBT 그대로다.
- `park` 를 부르는 네 return 자리(return@299·@325·@348·@351 — `absenceCorroborated` 가 취소로 읽기에 실패해 false 를 돌려주는 경우는 B24 를 지나 @348)가 전부
  **같은 ctx** 를 `park` 에 넘긴다(`Resolve` 안에 ctx 를 바꾸는 호출 없음 — calls 44 에 `context.` 0). 그래서 「취소 때문에 park 갈래로 왔다」면 그 쓰기는 반드시
  같은 취소를 본다. `waitOrExpire` 의 `clk.Sleep` 은 시스템·가짜 시계 둘 다 ctx 가 닫히면 `ctx.Err()` 를 돌려준다(clock.go:87-95, fake.go:44-56 — 손으로 읽음).
- CONFIRMED·FAILED 쪽도 같다: `ResolveConfirmed`(call@311)·`ResolveFailed`(call@343)도 `transition` 을 같은 ctx 로 부르므로 취소 뒤에는 쓰지 못하고 B18 return@313·
  B25 return@344 가 오류를 돌려준다 — attempt 는 IN_DOUBT 로 남고 다음 기동이 다시 판정한다.
- 반대 방향: 취소가 **커밋 진입 검사 뒤**에 오면 park 가 남을 **수 있다**(커밋 성공 여부에 따른다 — 성공해야 `park` B2 의 래치에 닿는다). 남는다면 그 park 는 취소 전에 내린 판정(기한 소진·부재 미증명)이다.

## 2. 설계 질문과 결정

### D1. 세션 중 해소 — 대상과 절차

- 대상: 원장의 미종결 attempt(`Journal.PendingAttempts` — RECORDED·DISPATCH_STARTED·ACKED·IN_DOUBT, recorded_at 순; 분기 B1~B4, SQL 손으로 읽음) 중 **`State == IN_DOUBT`
  ∧ `Kind == PLACE`** 인 것. RECORDED·DISPATCH_STARTED 는 세션 중 진행 중인 전송이라 손대지 않고(정본 SHALL NOT — 재시작 규칙 금지), ACKED 는 기동 확정 경로(a094)
  의 일, CANCEL·AMEND 는 비목표.
- 절차: 기존 `execgw.Resolver.Resolve` 를 **그대로** 부른다. 인스턴스는 엔진 조립의 `Context.Resolver`(gateway.go:298 의 `&execgw.Resolver{Journal, Orders, Order,
  Account, Clock, Gate: entry}` — 복합 리터럴, 손으로 읽음)를 공유한다. `Resolve` 는 수신자 필드에 쓰지 않는다(AST assignments 34줄 — 230~343; 대상 이름이
  전부 지역 변수(`cfg`·`clk`·`rec`·`res`·`attempt`·`matcher`·`absent`…)임은 손으로 읽음, `r.<필드> =` 0줄) — 기동 복구와 같은 인스턴스를 나눠 써도 상태가 섞이지 않는다.
- **재생 없음**: `Recovery.replay`(B1 if@364 Replayer 없으면 false) 같은 단계를 두지 않는다. 정본의 「재생 우선」은 기동 순서에 남긴다(spec delta). 이유: 재생은
  주문 생성 엔드포인트로 같은 본문을 다시 보내는 호출이다(`HTTPReplay.ReplayPlace` — a166 census 대상). 세션 해소는 조회만 하는 경로로 둔다(Manager 결정).
- 증명 규칙 무변경: `newMatcher`·`scanBoth`(B1 range@414 OPEN·CLOSED 둘 다, B6 if@429 dedup)·`absenceCorroborated`·`windowContaminated` 편집 0.

### D2. 실행 자리 — 보조 실행자 `in-doubt-settle`

**왜 보조 실행자인가.** exit 관측 루프 안에서 부르면 `Resolve` 가 최대 5분(+ 마지막 스캔) 폴링하는 동안 다른 포지션의 손절 판정이 멈춘다(불변식 4). 감독
루프(Loops)에 넣으면 그 비정상 반환이 엔진 전체를 내린다(`Runtime.Run` 의 방어적 종료 — B5 if@351 아님 → `ErrLoopFailed` return@378). 보조 실행자는
`Runtime.Run` B4(range@322)에서 **복구(B2) 뒤에** 시작되고, 같은 `wg` 로 기다려 원장보다 먼저 끝나며(call@347 `wg.Wait`), 패닉은 `runAuxiliaryBody` B1(if@132)
이 정지로 바꾸고, 비정상 정지는 `runAuxiliary` 가 B1(if@92 정상 정지 아님) 뒤 로그(call@113)로 남기고 B3(if@117) 뒤 `OnStop`(call@120)을 부른다. 기존 선례 5개
(`AlertDeliverer`·`NormalAlertRelayExecutor`·`AttestationWatch`·`ManualBuyRefresher`·`ExitLowerLimitRefresher` — `engineRuntime` call@704·@748)와 같은 모양이다.

**모양(새 파일 `internal/app/engine/in_doubt_settle.go`, 새 함수만).**

```text
const inDoubtSettleName = "in-doubt-settle"

func (c *Context) InDoubtSettler(clk clock.Clock) AuxiliaryExecutor
    → AuxiliaryExecutor{Name: inDoubtSettleName, Run: s.Run, OnStop: s.onStop, StopEvent: obs.EventOrderInDoubt}
    s 는 c.Journal(PendingAttempts) · c.Resolver · c.Notifier(RecordCritical) · c.Log · clk 를 잡음.

(s *inDoubtSettle) Run(ctx) error
    배선이 비었으면(Journal·Resolver nil) 즉시 오류 반환 → runAuxiliary 가 정지로 기록하고 OnStop 이 critical — 조용히 쉬지 않음.
    loop: ctx 가 이미 끝났으면 return ctx.Err()       // 첫 스윕 **전에** — 아래 주의
          s.sweep(ctx); 오류가 ctx 취소면 return ctx.Err()
          clk.Sleep(ctx, 주기) 오류면 return ctx.Err()

    주의(기존 시험): TestEngineRuntimeConstructionBranchesFailClosedAndAssembleExactSuccess 는 영값 `&journal.Journal{}`(db nil)로 조립한 런타임을
    `wired.Run` 으로 돌리고, ready 가 복구 직후 run ctx 를 취소한다(engine_runtime_branch_test.go:36·:118~126 — 손으로 읽음). 보조 실행자는 그 뒤 시작되므로
    첫 스윕 전 ctx 검사가 없으면 nil db 질의로 패닉 → 정지 기록이 난다(시험은 통과하더라도 잡음). 첫 검사로 그 경로를 닫는다.

(s *inDoubtSettle) sweep(ctx) error
    recs ← Journal.PendingAttempts(ctx)        // 실패: 로그 1줄(같은 문구 연속은 1회) 후 nil — 다음 주기 재시도
    for rec in recs where rec.State==IN_DOUBT && rec.Kind==PLACE (recorded_at 순):
        if ctx.Err() != nil { return ctx.Err() }   // 알림 기록 **전에** — 아래 D3 「알림 기록과 취소」
        s.noteStale(ctx, rec)                  // D4
        res, err ← Resolver.Resolve(ctx, rec.ID)
        if err != nil:
            if ctx.Err() != nil { return ctx.Err() }   // D3 — res 를 버린다(메모리의 UNRESOLVED 는 원장에 없다)
            로그 1줄(attempt id·사유; 같은 (attempt, 문구) 연속은 1회) ; continue
        결과 로그 1줄(attempt id·종목·상태·사유 코드 — 수량·가격·계좌 없음)
    return nil
```

**주기와 동시 실행 상한(정본이 요구한 「실측」 — 측정 가능한 것은 재고, 운영 값은 Manager 읽기 전용 측정으로).**

| 항목 | 값 | 근거 |
|---|---|---|
| 스윕 간격 | 직전 스윕이 **끝난 뒤** **5초** = `DefaultResolveConfig().PollInterval` | indoubt.go:160(상수, 손으로 읽음). 새 설정 없음. **스윕 안에서 해소가 돌면 다음 원장 읽기는 그 해소들이 끝날 때까지 늦어진다**(리뷰 1회차 R1-1) |
| 미결 0건일 때 비용 | 원장 질의 1회/5초, **브로커 호출 0** | `PendingAttempts` 만 |
| 동시 해소 상한 | **1**(직렬, goroutine 하나) | 한 번에 `Resolve` 하나 |
| 해소 하나의 길이 | 증명되면 짧다(부재: 안정 3회·45초 이상). 기한 `MaxDuration` 5분은 **다음 대기를 할지에서만** 검사된다(`waitOrExpire` B1 if@361) — 스캔 하나(`scanBoth` 의 `ScanOrders` OPEN·CLOSED 각 최대 50쪽, 요청당 HTTP 시한 15초)와 부재 교차 확인 읽기는 기한을 받지 않는다. 그래서 **전체 길이에 강제 상한은 없다**: 정상(1쪽·빠른 응답) 경로는 ≈5분 + 마지막 스캔, 병적 경로(여러 쪽·느린 응답)는 스캔 하나만으로 5분을 넘을 수 있다(리뷰 1회차 R1-2 — 사실 정정) | indoubt.go:158-161; `Resolve` B23(if@336)·`waitOrExpire` B1(if@361); `scanBoth` call@415 |
| 해소 하나의 브로커 요청(정상 1페이지) | 스캔마다 목록 2(OPEN·CLOSED), 부재 증명 경로는 + 보유·매수가능 2 | `scanBoth` B1(range@414) `ScanOrders`(call@415); `absenceCorroborated` call@467·@475. 부재 경로 ≈ 10스캔 × 2 + 2 = **22**, 기한 소진 경로 ≤ 61스캔 × 2 = **122**(5분, ≈0.41 req/s) |
| 페이지 상한 | 목록 1회 `MaxPages` 50 | indoubt.go:162 |
| 실측 1(시험) | 위 요청 수를 가짜 브로커로 센다 | tasks 3.9 `TestA193ResolutionRequestsAreBounded` |
| 실측 2(운영, 사람) | 운영 저널 사본에서 기간별 IN_DOUBT PLACE 수·동시 최대 수(같은 순간 IN_DOUBT 인 attempt 수) | tasks 1.10 — Manager 읽기 전용(계좌 필드 미기록). 1단계에서는 **안 잼**(지시상 운영 저널 읽기 금지) |

동시에 IN_DOUBT 인 attempt 의 수는 종목 수로 묶인다 — 게이트웨이가 종목당 미종결 하나만 허용한다(`checkSymbolFree` B2 if@808). 직렬이라 k 번째 attempt 는 앞의
k−1 개가 끝날 때까지 기다린다(정상 경로 최악 ≈(k−1) × 5분, 강제 상한 없음). **D4 의 10분 검사도 같은 스윕 안에서 해소 사이에 도므로 그만큼 늦어진다** —
「10분에 알린다」는 보장이 아니라 「해소가 돌지 않는 동안 10분 뒤 첫 스윕에 알린다」이다(리뷰 1회차 R1-1, Manager 결정 대기 Q-F).

**예산(a166)과 손절 우선.** 해소의 브로커 읽기는 같은 공식 클라이언트(`OfficialOrders`·`OfficialAccount{Client: in.official}`, gateway.go:259-260 — 손으로 읽음)로
나간다. ctx 에 우선순위를 싣지 않으므로 `accountbudget.PriorityFromContext` 의 기본값 **Normal**(transport_linux.go:24-29, 손으로 읽음)이다 — 기동 복구와 같다.
a166 이 enforcing 이면 대기열은 Critical 티켓을 먼저 고른다(budget_linux.go:551 의 선택식 — 손으로 읽음), 보호 청산 제출은 `submit` B6(if@1553)에서 Critical 을 싣는다.
그래서 **함께 대기 중이면** 손절 제출이 해소 읽기보다 먼저 나간다(같은 Normal 끼리는 순번). **이미 wire 임차를 쥔 해소 읽기는 선점하지 못한다** — 임차는 응답 body 가
끝날 때까지 유지되므로(transport_linux.go:459 부근, 손으로 읽음) 손절의 wire 진입이 그 요청 하나의 점유만큼 늦을 수 있다. 점유 길이는 **미측정**이다 — HTTP 시한(15초)과 임차 해제는 같은 경계가 아니고(요청 끝 뒤 후처리
`obs.Finish` 등을 거쳐 해제 — transport_linux.go:400-418, 리뷰 2회차 R2-4 가 짚음; 손으로 읽지 않음) 숫자 상한을 주장하지 않는다(리뷰 1회차 R1-4 「앞지르지 못한다」 과장 정정,
2회차 R2-4 「≤ 15초」 철회). **같은 지연은 오늘도 있다**: 대사 루프·체결 감지(3초 주기)의 조회도 우선순위를 싣지 않아 Normal 이고(`rg WithPriority` — exitloop·flatten·replay 만,
손으로 읽음) 손절과 동시에 돈다. a166 의 사용자 승인 문구는 「보호 요청 대기 허용」이다(a166 proposal.md:3). 그래도 이것이 불변식 4 에 맞는지는 **freeze 차단 결정 Q-G**. a166 이 꺼져 있으면 중재가 없고 해소 읽기는 기존 retry matrix 의 조회 몫이다
(정본 §0.4 — 새 엔드포인트 없음, 빈도만 는다: tasks 4.6 이 matrix 문서에 계상을 적는다). 보조(Auxiliary) 우선순위는 고르지 않았다 — 대기열이 비어 있을
때만 나가므로 부하가 있으면 해소가 굶는다(Manager 결정 대기 Q-C).

**배선 — 바꾸는 기존 함수 하나.** `cmd/tossctl/engine.go` `engineRuntime`(655-752, 분기 6) 의 `Auxiliary` 목록(L748~749, 분기 아님)에 `ectx.InDoubtSettler(clk)` 원소 하나.
분기·return 무변경. a166 census(`internal/accountbudgetruntime/testdata/a166-wiring/shapes.json`)의 키 50개에 `engineRuntime` 은 없다(키 목록 대조 — 손으로 읽음),
고정된 넷(`buildGateway`·`newGatewayReplay`·`NewContext`·`NewOrderPath`)은 무편집이다 — `Context.Resolver` 가 이미 있기 때문이다(engine.go:188). 새 파일은 census 가
세는 선택자(`Do`·`RoundTrip`·`NewClient`·`Budget`·`official.New*`·`net/http` 직접 호출 — a166_wiring_census_test.go:81-94, 손으로 읽음)를 쓰지 않는다.
보조 실행자 집합을 고정하는 기존 시험 `TestProductionRuntimeStartsExactlyOneAlertDeliverer`(B2 if@28 `reflect.DeepEqual`)의 기대 목록에 이름 하나를 더한다 — 시험 함수
본문 편집이라 FLM 번들 대상(선례 a087·a135·a136 이 같은 줄을 고쳤다).

### D3. 종료는 park 가 아니다 — 근본 수리 대 호출자 방어

**정한 것: Resolver 는 고치지 않는다. 호출자(새 실행자)가 오류 + `ctx.Err() != nil` 을 「해소 안 됨, 멈춤」으로 다루고, 그 성질을 시험으로 못 박는다.**

근거(§1.4):

1. 원장에서는 **이미 park 가 아니다.** 취소로 park 갈래에 온 경우 `ResolveUnresolved` 가 같은 취소된 ctx 로 `BeginTx` 를 열다 실패하고 `park` B1 이 `Gate.Block` 앞에서
   돌아간다. 메모리 결과만 UNRESOLVED 이고 오류가 같이 온다.
2. 기동 복구도 같다: 기동 중 취소면 `Recovery.Run` B9(if@301) → return@302 `ErrRecoveryIncomplete` → `Runtime.Run` B2 return@295 — 루프 없이 끝나고 원장의 attempt 는
   IN_DOUBT 그대로(다음 기동이 다시 해소).
3. Resolver 를 고치면(예: `waitOrExpire` 가 취소와 기한을 구별) `Resolve`·`resolveCancel`·`resolveAmend`·`settleAmendFromClosedOriginal` 네 호출자(CodeGraph `callers
   waitOrExpire`) 의 갈래가 바뀌고, 기동·flatten 경로까지 FLM·BTM·변이 대상이 넓어진다. 원장 결과는 같다(1). 바뀌는 것은 오류 메시지뿐이므로 비례 원칙상 고르지 않는다.

**이 안전은 우연에 기대고 있다**(기억 「살아남은 뮤테이션 = 우연이 지키는 안전」): park 의 쓰기가 같은 ctx 를 쓴다는 사실이 지킨다. 누가 park 쓰기를
`context.WithoutCancel` 로 바꾸면 배포 때마다 park 가 난다. 그래서 못을 박는다 — tasks 3.4 `TestA193ACancelledResolutionLeavesTheAttemptInDoubt`(실제 원장 +
가짜 시계로 `waitOrExpire` 의 Sleep 중 취소 → 원장 상태 IN_DOUBT, 게이트 래치 0, critical 0, 실행자 반환 `context.Canceled`)와 그 변이(m3: park 쓰기 ctx 를
`WithoutCancel` 로 → 시험 FAIL 이어야 함, m4: 실행자의 `ctx.Err()` 검사 삭제 → 결과 로그/알림이 나가 FAIL).

**남는 경계**: 취소가 park 의 **커밋 진입 검사 뒤**에 오면(`Tx.Commit` 의 ctx 검사 통과 뒤) park 가 남을 **수 있다** — 커밋 성공 여부에 따른다(리뷰 2회차 R2-5 — 단정 정정). 그 park 의 원인은 취소가 아니다 — 취소가 원인이면
취소는 `waitOrExpire` 의 Sleep·스캔 오류보다 **먼저** 일어났고, 한 번 닫힌 `ctx.Done()` 은 다시 열리지 않으므로 `BeginTx` 검사가 이미 그것을 본다. 그래서 커밋 진입 뒤의 취소와
겹치는 park 는 취소 전에 내린 판정(기한 소진·부재 미증명·2건 일치)이다. 이 「원인 → 검사」 순서는 위 표준 라이브러리 읽기에 기대고 드라이버 원문은 보지 않았다(리뷰 1회차
R1-6 — 경계 서술 보강).

**알림 기록과 취소(리뷰 1회차 R1-6 — 설계 수정)**: 알림기의 단일 입구는 원장 기록이 실패하면 진입 게이트에 `ReasonAlertUndelivered` 래치를 건다(record_only.go 의
`recordCritical` — `RecordAlert` 실패 → `n.Gate.Block`, 손으로 읽음). 취소된 ctx 로 `noteStale` 이 기록하면 그 래치가 선다. 그래서 스윕은 `noteStale` **앞에서** ctx 를
검사한다(위 의사 코드). **이것은 부분 해결이다**(리뷰 2회차 R2-2): 검사 **전**의 취소는 기록하지 않아 래치가 안 서지만, 검사 **뒤** 기록 중의 취소는 `RecordAlert` 실패 →
`ReasonAlertUndelivered` 래치를 세울 수 있다. 같은 입구를 쓰는 모든 기록자(exit 관측 포함)의 기존 성질이고 이 change 가 넓히지 않지만, delta 의 규범은 그것에 맞춰
「검사 의무」로만 쓰고 「취소로 잠그지 않는다」 를 약속하지 않는다(이름 붙인 잔여). 그 래치는 기록 실패라 원장 행이 없으므로 재시작에 복원되지 않는다(`restoreAlertEntryLatch` 는 `UndeliveredCount` — 미전달 **행** 수로만 복원, gateway.go:158-173 손으로 읽음). 「취소와 기한 소진이 같은 갈래」라는 AST 사실은 그대로이고, 원장
쓰기의 ctx 결속이 그 둘을 가른다.

### D4. 오래 남은 IN_DOUBT 알림

- **조건**: 스윕이 읽은 PLACE attempt 가 아직 `IN_DOUBT` 이고, `DispatchStartedAt`(빈 값이면 `RecordedAt` — `windowContaminated` B1 if@516 과 같은 규칙) 부터
  지금까지 **10분 이상**. 10분 = 2 × `DefaultResolveConfig().MaxDuration`(5분): 해소 하나의 전체 예산 + 직렬 대기열에서 앞선 해소 하나. 건강한 실행자에서 단독
  attempt 는 정상 경로(1쪽·빠른 응답)면 약 5분 안에 종결(CONFIRMED·FAILED·park)되므로 이 알림은 **해소가 오류로 끝나 되풀이되거나, 대기열이 밀렸거나, 해소가 시작되지 못한** 경우에만 난다.
  `Resolver.Config.MaxDuration` 이 0 이 아니면 그 값에서 유도한다(오늘 production 은 0 → 기본값).
- **알림**: 기존 critical 타입 `obs.EventOrderInDoubt`(`order.in_doubt` — event.go:55, 등급표 event.go:364 에 이미 critical; 손으로 읽음). 새 타입 없음 → `internal/obs`
  무편집. key = `order.in_doubt|stale:<attempt id>` — 에피소드 = attempt(원장 행, 재시작에도 같은 행; 선례 `acked_boot.go:77` 의 `order.in_doubt|acked:<id>`).
  입구 = `Notifier.RecordCritical(ctx, e, 0)`(a094 D−6.1 · a092 단일 입구, 창 0 — record_only.go:86-101, B1·B2). 기록 성공 때만 프로세스 래치(같은 프로세스에서
  다시 적재하지 않음 — `recordEpisodeCritical` B2·B3 와 같은 규칙); 실패하면 다음 스윕이 다시 시도하고, 진입 잠금은 입구의 몫(a092).
- **본문(한국어, 계좌·수량·가격 없음)**: 「주문 결과 불명(IN_DOUBT)이 N분째 해소되지 않았다 — 원 주문이 살아 있을 수 있어 그 종목의 다른 주문(손절 포함)이 막혀
  있다. 엔진 로그의 in-doubt-settle 줄에서 해소 오류를 확인하라」 + attempt id · 종목 · 종류 · 상태 · 경과 · 마지막 해소 오류의 사유 코드(있으면).
- **실행자 정지**: `OnStop(ctx, err)` 이 key `order.in_doubt|settler-stopped:<실행 id>`(실행 id = 실행자 생성 때 128비트 난수 한 번 — a094 D−7.1 의 연속 id 와 같은 이유:
  벽시계 key 는 역행에 흡수된다)로 critical 하나: 「세션 중 IN_DOUBT 해소가 멈췄다 — 엔진 재시작 전까지 새 IN_DOUBT 는 해소되지 않는다」. 정상 정지(취소)에서는
  `runAuxiliary` B1(if@92)이 먼저 돌아가 `OnStop` 을 안 부른다.
- **중복 없음 — 단, 덮는 범위는 exit 발의뿐(리뷰 1회차 R1-3 — 사실 정정)**: park 로 끝난 attempt 는 종결이라 스윕 대상이 아니다 → 이 알림 0. park 의 알림은 기존
  `recordParkCause`(key `order.unresolved_in_doubt|<position>|attempt:<id>`)가 **무장된 exit 발의의 intent 에 속한 attempt** 에 대해서만 낸다(`noteHeldProposal`
  B1 if@127·B2 if@130 → `ExitIntentAttempts(m.state.PendingIntentID)` call@135 → B6). 전략 진입·수동 flatten 청산 매도의 PLACE 가 세션 중 park 되면 **critical 이 없다**
  — 계정 전역 진입 래치만 선다(기동 복구의 park 도 같다 — D5). 그 통지를 이 change 에 넣을지는 Manager 결정 대기 Q-E. 운영 원장의 intent 는 전부 엔진 매도
  (Manager 실측 13,052/13,052)라 오늘 실제로 이 공백에 드는 모양은 exit 발의가 아닌 엔진 매도(flatten 청산)뿐이다.
- **원장 읽기 연속 실패의 침묵(리뷰 1회차 R1-1)**: `PendingAttempts` 가 계속 실패하면 스윕은 로그만 남기고 실행자는 살아 있다 — `noteStale` 도 `OnStop` 도 안 돈다.
  승격(연속 N회 → critical)은 Manager 결정 대기 Q-F.
- **덮지 못하는 것(이름 붙인 잔여)**: 실행자가 살아 있으나 한 호출 안에서 멈춘 경우(브로커 HTTP 는 `defaultTimeout` 15초 — official/client.go:22, 원장은
  `busy_timeout` — journal.go:224; 둘 다 유한이라 「영원히」는 아니다 — 손으로 읽음). exit 관측 루프가 아닌 실행자가 알리므로 실행자의 생존에 기댄다 — 정지는
  `OnStop` 이 덮는다.

### D5. 세 결과의 뒤

| 결과 | 그 뒤(AST) | 세션 중 누가 | 빠진 것 |
|---|---|---|---|
| **CONFIRMED**(목록에 정확히 1) | `ResolveConfirmed` 가 broker_order_id 를 기록 → 추적 주문 SQL(`TrackedFillOrders` 의 `state = CONFIRMED AND kind IN ('PLACE','AMEND') AND broker_order_id <> ''` — fills.go:1554-1559, 손으로 읽음)에 들어감 → 체결 감지(a191)가 다음 주기부터 그 주문의 체결·종결을 기록. 발의는 해제하지 않음(`exitIntentAttemptsQ` B10 case@214 confirmed → 종결 기록 전 B17 `AwaitingClose`, 뒤 B18 `OrdersClosed`) — 정본 「CONFIRMED 면 해제 금지」 | 체결 감지 + 기존 exit 경로 | 없음 — 기동 복구의 CONFIRMED 와 같은 원장 상태 |
| **FAILED_CONFIRMED**(부재 증명) | 기준선이 있어야만 가능(`absenceCorroborated` B3 return@464). 그 attempt 가 exit 발의의 것이면 다음 판정에서 `noteHeldProposal` B4(if@140) → `ReleaseUnacceptedExitProposal`(call@141) → 해제 → 재발의 | 기존 exit 경로 | **엔진 손절에는 도달 불가** — 손절 요청에 기준선이 없다(§1.2-1). 오늘 기준선을 싣는 production 자리는 전략 진입(`PlaceClaimedStrategy` → `g.place`, call@64) 뿐. 이 change 의 비목표 |
| **UNRESOLVED_IN_DOUBT**(park) | `park` B2 `Gate.Block(ReasonUnresolvedInDoubt)`(call@380) — **계정 전체 신규 진입 차단**(`EntryGate.Block` 은 이유별 래치, 종목 구분 없음 — B1 if@561). 다음 판정의 `noteHeldProposal` B6(range@148) → `recordParkCause`(call@149) critical(창 0, key 에피소드 = attempt). 종목 단위: `checkSymbolFree` B3(if@814)·B5(range@821)·B7(if@826) — 노출을 늘리는 mutation 만 그 종목에서 거절. 발의는 무장 유지(정본) → 손절은 운영자 해동까지 동결 | 실행자(park) + exit 루프(알림 — **무장된 exit 발의의 attempt 일 때만**, 리뷰 R1-3) | 아래 공시; exit 발의가 아닌 PLACE 의 park 통지 없음(Q-E) |

**기동 복구의 `blockedSymbol`·report 가 하던 일**: `blockedSymbol`(B1 if@425)은 ACKED attempt 의 보고용 값만 만든다(부작용 0). `Report.Unresolved`·`Resolutions` 의
production 소비자는 `engineRecoveryObserver` 가 돌려주는 클로저(return@102)인데 그것은 `RateLimitWaits` 만 본다(B1 if@103 ≤ 0 이면 return@104, 아니면 경고 로그
call@106 하나) — **기동 복구도 park 를 알리지 않는다**.
세션에서 이 자리를 대신할 것은 없다(빠진 것 없음). 세션 해소는 결과마다 계좌·수량·가격 없는 구조화 로그 줄 하나를 남긴다(오늘 기동 복구보다 많다).

**공시(정본이 요구한 셋째 — 운영 결과).** 세션 중 park 는:
1. 그 엔진 프로세스가 끝날 때까지 **계정 전체의 신규 진입**을 막는다. `ReasonUnresolvedInDoubt` 래치를 푸는 production 호출자는 0이다(`rg '\.Clear\('` 시험 제외 —
   풀리는 이유는 `AlertUndelivered`·`RecoveryIncomplete`·`FillDetectionSLO`·대사의 둘뿐, 손으로 읽음). 운영자 해동(`attempt-resolve`)은 다른 프로세스의 원장 쓰기라
   이 메모리 래치를 못 푼다. 재시작하면 래치는 사라지고 종목 단위 거절(원장의 park 행)만 남는다.
2. 청산·손절은 이 래치로 막히지 않는다(게이트웨이의 진입 게이트 검사는 노출을 늘리는 mutation 에만 — gateway.go:860, 손으로 읽음).
3. 그 attempt 의 발의는 무장된 채라 **그 포지션의 손절은 해동까지 동결**이다 — 오늘과 같다; 이 change 는 그것을 알 수 있게 하는 시각만 당긴다.
4. 사고와 같은 모양(손절 매도, 장 마감 뒤 5xx)이면 예상 결과는 약 45~50초 뒤 park(`absence … without a pre-dispatch account baseline`)와 park 원인 critical 이다.
   지금은 그 결과가 다음 재시작 때 나온다.

### D6. 동시성 — 같은 attempt·같은 종목을 동시에 만지는 자

| 상대 | 무엇을 만지나 | 겹침 | 근거 |
|---|---|---|---|
| 기동 복구 `Recovery.Run` | 모든 미종결 attempt | **시간이 겹치지 않는다** — 보조 실행자는 `Runtime.Run` B2 통과 뒤 B4 에서 시작 | `Runtime.Run` B1·B2(if@289·@294) → B4(range@322) |
| 수동 flatten `resolveStep` | 자기 saga 의 **취소** attempt | 대상 종류가 겹치지 않음(실행자는 PLACE 만). flatten 의 청산 매도 PLACE 가 IN_DOUBT 면 flatten 은 step 에만 적고(`Saga.sell` B10 case@516) 해소하지 않음 → 실행자가 해소(기동 복구도 같은 집합을 해소함). 별도 프로세스라 메모리 잠금은 공유하지 않음 | `resolveStep` call@428; `Saga.sell` |
| 운영자 해동(`attempt-resolve`) | **UNRESOLVED_IN_DOUBT** attempt | 상태가 겹치지 않음(실행자는 IN_DOUBT 만). 실행자가 park 한 직후 운영자가 해동하는 것은 순서일 뿐 | 정본 「park 가 아닌 attempt 의 운영자 해소는 거절」 |
| 체결 감지(a191) | CONFIRMED 주문의 체결 스냅숏 | attempt 상태를 쓰지 않음. 실행자의 CONFIRMED 뒤에 그 주문을 새로 추적함 | `TrackedFillOrders` SQL(손으로 읽음) |
| 게이트웨이(exit·전략·flatten 제출) | 같은 종목의 새 mutation | IN_DOUBT 가 남아 있는 동안 그 종목의 새 mutation 은 **전부** 거절(`checkSymbolFree` B2) → 해소의 관측 창을 다른 제출이 오염시키지 않음. 다른 프로세스(flatten)도 같은 원장으로 같은 판정 | `checkSymbolFree` B1·B2 |
| 실행자 자신 | 같은 attempt 두 번 | goroutine 하나·직렬. 그래도 겹치면(시험·미래 편집) 원장 전이가 `from` 상태를 tx 안에서 다시 읽어 거절(`transition` call@560 → `checkTransitionAllowed` call@567 → B4 if@567 return@568) | `transition` |

**같은 attempt 를 두 번 해소하지 않는다**는 위 마지막 줄 두 층(직렬 + 전이 가드)으로 선다. 시험 3.6 이 직렬을, 기존 전이 시험이 가드를 잰다.

**프로세스 사이의 종목 배제는 절대적이지 않다(리뷰 1회차 R1-5 — 주장 철회).** `claimSymbol` 은 게이트웨이 인스턴스 안의 잠금이고 flatten 은 자기 게이트웨이를 만든다;
`checkSymbolFree` 의 원장 검사와 attempt 기록(`Prepare`)은 다른 트랜잭션이다(gateway.go:522·529·548, durability.go:383 — 손으로 읽음). 그래서 「엔진 제출이 빈 원장을
확인 → flatten 제출이 IN_DOUBT 기록 → 실행자가 해소 시작 → 앞서 검사를 통과한 엔진 제출이 기록·전송」 순서가 가능하다. 위 표의 「게이트웨이」 행은 **이미 IN_DOUBT 가
보이는 뒤의 새 호출**에 대해서만 맞다. 이 경합은 **기존 제출 경로의 공백**이고(기동 복구의 해소도 같은 순서에 노출된다) 이 change 가 만들거나 넓히지 않는다. 그 경우
해소 쪽 결과: 끼어든 mutation 이 관측 창 안에 원장에 기록되면 `windowContaminated`(B3 if@525)가 부재 증명을 막아 park; 같은 지문의 주문이 목록에 둘이면 park(B19);
**끼어든 주문이 같은 지문으로 하나만 보이면 그것을 이 attempt 의 주문으로 CONFIRMED 할 수 있다**(지문은 종목·방향·수량·가격·제출 시각 창 — 다른 프로세스가 같은 값으로
낼 때만). 프로세스 사이 원자적 배제는 Manager 결정 대기 Q-H.

**앞으로의 제약(이름 붙여 둠)**: a171(진입 접수 결과 불명의 8초 창 재전송 — 아직 착수 전)처럼 **게이트웨이가 IN_DOUBT 를 기록한 뒤 같은 `Place` 안에서 그 attempt 를
다시 만지는 단계**가 생기면, 세션 해소와의 순서(예: 그 창이 끝날 때까지 해소 보류)를 그 change 가 정해야 한다. 오늘은 `journal.Dispatch` 의 IN_DOUBT 전이
(dispatch.go:174·191·206·215, 손으로 읽음) 뒤 같은 attempt 를 다시 쓰는 단계가 없다.

## 3. 바꾸지 않는 것

`execgw` 전체(`Resolver`·`Gateway`·재생), `reconcile.Recovery`, exit 관측 루프 전체(`submit`·`noteHeldProposal`·알림 함수), `journal`(스키마·질의), `internal/obs`(이벤트
열거·등급표), a166 census 네 함수, flatten, 설정·토글. upstream 상속 시험 영향 없음(엔진은 TossOS 계층).

## 4. fail-closed 가 거부할 정상 입력 / 통과시키는 것

| 입력 | 처리 | 정상 입력을 거부하나 |
|---|---|---|
| 배선 누락(Journal·Resolver nil) | Run 즉시 오류 → 정지 기록 + critical | production 은 `Context.Recovery` B1(if@176)이 같은 nil 을 이미 거부해 엔진이 안 뜸 → 0 |
| `PendingAttempts` 실패 | 그 스윕 건너뜀, 로그 | 일시 오류는 다음 스윕(5초) |
| `Resolve` 오류(ctx 아님) | 그 attempt 건너뜀, 다음 스윕 재시도, 10분 뒤 D4 | 재시도 빈도 5초 — 오류가 브로커 이전(`newMatcher` 등)이면 브로커 호출 0 |
| IN_DOUBT 이지만 CANCEL·AMEND | 손대지 않음 | 비목표(기동 복구가 계속 맡음) — 10분 알림도 안 냄(Manager 결정 대기 Q-B) |
| 다른 계좌의 attempt(같은 원장) | **기동 복구와 같이 거르지 않음** | 원장은 config dir 단위 — 기동 복구와 다른 집합을 만들지 않기 위함(정본 「판정을 둘로 두지 않는다」) |

## 5. 변이 계획 (2단계 — 원장 `analysis/mutation/`)

m1 State 필터 삭제(RECORDED 를 Resolve → ErrNotResolvable 로그만 — 시험은 「손대지 않음」을 원장 전이 0 으로 잰다) · m2 Kind 필터 삭제(CANCEL IN_DOUBT 를 해소) ·
m3 park 쓰기 ctx 를 `WithoutCancel`(Resolver 쪽 — 시험 3.4 가 FAIL 이어야) · m4 실행자의 `ctx.Err()` 검사 삭제 · m5 10분을 5분으로 · m6 경과 기준을 `RecordedAt` 고정 ·
m7 창 0 → 1시간 · m8 key 에서 attempt id 삭제 · m9 래치를 기록 실패에도 세움 · m10 OnStop 의 critical 삭제 · m11 `engineRuntime` 에서 원소 삭제(배선 시험) ·
m12 실행자를 Loops 로 옮김(감독 시험) · m13 스윕을 exit 관측 주기에서 동기 호출(비대기 시험) · m14 `ReplayInDoubt` 호출 추가(재생 0 시험) · m15 결과 로그에 수량 필드.
각 변이는 양성 대조군(닿음) 먼저, 사본 하네스는 무변이 대조군 GREEN 과 시작 sha 단언 뒤(기억 「뮤테이션은 대상에 닿아야 한다」).

## 6. Manager 결정 대기 (review.md 에도 같은 목록)

- **Q-A**: 정본 「등급화된 알림」은 IN_DOUBT **발생**을 critical 로 정했는데 exit 제출은 로그 줄만 남긴다(§1.2-2). 선택지 (a) 이 change 밖 — 이름만 붙임(현재 설계),
  (b) 실행자가 IN_DOUBT PLACE 를 **처음 볼 때** critical 한 번(key `order.in_doubt|attempt:<id>`, 창 0) — 10분 알림과 별개. 추천 (b): 비용이 새 함수 하나이고
  정본과 코드의 차이를 닫는다. 범위 확장이라 Manager 결정.
- **Q-B**: CANCEL·AMEND IN_DOUBT 의 10분 알림(해소는 비목표지만 알림은 싸다). 추천: 비목표 유지(별도 Story 에서 해소와 함께).
- **Q-C**: 해소 읽기의 예산 우선순위 Normal(기동 복구와 같음) 대 Auxiliary(대기열이 빌 때만). 추천: Normal — 손절 제출은 Critical 이라 이미 앞서고, Auxiliary 는
  부하 때 해소를 굶겨 손절 동결을 길게 한다.
- **Q-D**: Story 의 Feature — FEAT-TOS-009(a094, exit line truth) 로 둠. 지시의 「a094·a192 가 속한 Feature」가 둘(009·005)이라 하나를 골랐다.
- **Q-E**(리뷰 R1-3): exit 발의가 아닌 PLACE(전략 진입·flatten 청산 매도)의 세션 중 park 는 critical 이 없다(기동도 같다). (a) 이름 붙인 잔여 / (b) 실행자가 park 결과를
  받으면, 그 attempt 의 intent 가 무장된 exit 발의의 intent 가 **아닐 때만** critical 1회(key `order.unresolved_in_doubt|attempt:<id>`, 창 0 — exit 쪽 알림과 겹치지 않게).
  추천 (b): 세션 park 는 계정 전역 진입을 막는 사건인데 통지가 없으면 「공시」가 문서에만 있다. **정합성(리뷰 2회차 R2-3)**: 정본(이 delta 가 보존한) 「해소 불능」 시나리오는
  park 때 운영자 알림 발송을 발의 여부와 무관하게 요구한다 — (a) 를 고르면 기동·세션 둘 다 그 시나리오를 충족하지 못하는 상태를 「잔여」로 기록하는 것이고 충족으로 보고하지 않는다.
  범위 확장이라 결정 필요.
- **Q-F**(리뷰 R1-1): (1) 10분 검사를 직렬 해소에서 떼어 같은 실행자 안의 별도 goroutine(5초 간격, 원장 읽기만)으로 — 「10분에 알린다」를 참으로 만든다;
  (2) `PendingAttempts` 연속 실패 `obs.DefaultCriticalAttempts` 회 → critical 1회(에피소드 = 연속 id, a094 D−7.1). 추천 둘 다 예 — 새 함수만, 브로커 호출 0.
  고르지 않으면 문서의 약속을 「해소가 돌지 않는 동안」으로 좁힌 현재 문장이 남는다.
- **Q-G — freeze 차단**(리뷰 R1-4 · 2회차 R2-1 P0): 해소 읽기가 쥔 wire 임차만큼 손절 wire 진입이 늦을 수 있다(점유 길이 미측정). 리뷰 2회차는 이것을 불변식 4(손절 지연 금지)
  위반으로 보아 「측정 뒤 잔여 공시」로 닫을 수 없다고 판정했다. 작성자 근거: 같은 지연을 오늘 대사·체결 감지의 Normal 조회가 이미 만들고, a166 사용자 승인이 「보호 요청 대기
  허용」이다. 선택지 (a) a166 승인 범위 안의 기존 성질로 보고 측정·공시 / (b) 해소 읽기의 점유를 짧게(시한) / (c) a166 이 운영 enforcing 이 아니면 지금 이 지연은 없고
  enforcing 전환 때 다시 판단. **추천: 먼저 a166 운영 상태 확인 → (c) 면 (a) 로 기록, enforcing 이면 Manager·사용자 판단.** 결정 전 READY 아님.
- **Q-H**(리뷰 R1-5): 엔진·flatten 프로세스 사이 「검사 → 기록」 경합(기존 공백). 추천: 이 change 밖 — 이름 붙인 후속 후보. 세션 해소가 그 결과를 더 나쁘게 만들지 않는다
  (기동 복구와 같은 해소 함수).
- **Q-I**(리뷰 R1-2): Resolver 전체 호출에 강제 시한을 둘지. 추천: 두지 않음 — 단순 `WithTimeout` 은 D3 에 따라 기한 소진 park 대신 IN_DOUBT 잔류를 만든다(해소 예산
  만료와 엔진 종료 취소의 처분을 가르는 Resolver 수리가 필요해지고 그것은 비례 원칙상 이 change 밖).

## 7. 후속 후보 (이 change 밖)

1. **매도용 부재 증거 모델 + 손절 기준선**(정본 「부재 판정의 증거 모델은 매도에 대해 성립해야 한다」) — 이것이 생겨야 손절의 부재가 park 가 아니라 해제·재무장으로 간다.
2. 세션 중 park 의 계정 전역 진입 래치를 운영자 해동과 잇기(지금은 프로세스 수명).
3. CANCEL·AMEND 의 세션 중 해소.
4. 재생(`ReplayInDoubt`)의 세션 사용 — 능력 attestation 실측 뒤.
5. 사고 당일 알림 13시간 지연의 원인 조사.
