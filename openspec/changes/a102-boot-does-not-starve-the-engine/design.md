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
기존 report 소비 지점에 그대로 실린다 — 새 로거 배관 금지.

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
- `Recovery.Run`의 replay/observation 경로 429 — 관측된 실패 없음, not-applicable.

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
