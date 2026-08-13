# Tasks: a102-boot-does-not-starve-the-engine

역할 분담: Manager(Fable)가 §0·§5·§6을, T1(Opus)이 §1·§2를, T2(Opus)가 §3·§4를 진다.
각 Teammate 뒤에 전담 적대 리뷰가 붙고(§5), 지적은 그 Teammate가 고친다.
Teammate는 여기·design.md에 없는 결정이 필요해지면 **멈추고 Manager에게 묻는다.**

## 0. Manager 선행 산출물

- [x] 0.1 proposal.md (1aa3f0e1) + AST 7건 — 완료, 커밋됨
- [x] 0.2 design.md — D1~D7 확정
- [x] 0.3 spec delta — `reconciliation`, `operator-console`
- [x] 0.4 STORY-TOS-a102 + `_registry.yaml` + feature 역링크
  (sdd-test가 story 없는 active change를 즉시 실패시킨다 — a103 실측)

## 1. T1 — 겹1: recovery가 rate limit에 죽지 않는다 (High-risk: internal/reconcile)

- [x] 1.1 FLM/BTM **편집 전** 작성 — `Collector.Collect`(AST 있음 8/6/17),
  `Recovery.stableSnapshot`(AST 있음 5/4/7), 편집하게 되는 다른 기존 함수 전부
  (`Stabilisation.withDefaults` 포함 예상). `function-logic-map.md`와
  `branch-test-map.md`, 형식은 a098의 것을 따른다.
  → 편집 전 4종(AST+커버리지 측정): `collector.collect` · `recovery.stablesnapshot` ·
  `stabilisation.withdefaults`(AST 신규 생성 3/1/0) · `recovery.run`(3단계 호출부를
  편집하므로 대상). GREEN 뒤 같은 도구로 **재측정**해 좌표·수를 2판으로 갱신했다.
- [x] 1.2 [T] RED — D1b: `%v` wrap 때문에 `errors.Is(…, official.ErrRateLimited)`가
  Collect 오류에서 끊긴다는 실패 테스트. 429를 내는 fake broker로
  Collect 오류를 받아 `errors.Is` 가 false임을 먼저 보인다.
  → `a102_rate_limit_identity_test.go`, RED 6건 실패 실측.
- [x] 1.3 GREEN — snapshot.go 4곳(248·253·263·271)의 원인 wrap을 `%w`로.
  `ErrPartialSnapshot` 소비자 무회귀 테스트 포함.
  → 현재 좌표 `:261`·`:266`·`:276`·`:284`. 무회귀는
  `TestCollectStillReportsAPartialSnapshot` + 기존 4지점.
- [x] 1.4 [T] RED — stableSnapshot: 429 두 번 뒤 안정 스냅샷이 오는 fake에서
  (a) recovery가 완주하고 (b) attempt·taken이 429에 소모되지 않고
  (c) 대기 각 15s이며 (d) `Report.RateLimitWaits/RateLimitWaited`가 실측을
  담는다는 테스트. 지금 구현에서는 (a)부터 실패한다.
  → `a102_recovery_rate_limit_test.go`, RED는 컴파일 실패 11건.
- [x] 1.5 GREEN — D3 구현: `Stabilisation.RateLimitBackoff/MaxRateLimitWait`
  (zero-default 15s/5m, `withDefaults`), 429는 attempt 미소모 + 백오프,
  예산 소진 시 `ErrRecoveryIncomplete`(rate limit과 대기 시간 명시).
  → `internal/reconcile/ratelimit.go`(신규) + `recovery.go`.
- [x] 1.6 [T] 경계 고정 — (a) 429 아닌 오류는 오늘처럼 즉시 실패 (b) 예산 소진
  경로 (c) 백오프 중 ctx 취소가 즉시 통과 (불변식 4).
- [x] 1.7 [T] 뮤테이션 정산 — design §검증계약 (a)(b)(c)를 실행해 각각 어느
  테스트가 죽는지 기록하고 원복을 심볼로 확인한다 (통과는 증거가 아니다).
  → (a)(b)(c) + 자발 (d)(e), 총 8회 가함. **(c)의 네 자리 중 `:266`은 살아남았고**
  그 이유를 `collector.collect/branch-test-map.md`에 적었다.
- [x] 1.8 `go test ./internal/reconcile` + `make lint` rc=0. 결과를 branch-test-map에
  기록.
  → 143건 통과 · coverage 86.6% · `make lint` rc=0 (gofmt 실재 확인).

- [x] 1.9 [T] **A1 리뷰 반영 (FIX-FIRST)** — ① `ratelimit.go:95` `%v`→`%w`
  (F3: 취소 정체가 지워진다 — 이 change의 존재 이유와 같은 결함) + 취소 시
  `errors.Is(err, context.Canceled)` 고정 테스트 ② 취소 후 브로커 호출 수 단언
  (F2 — 생존 뮤테이션 N2를 죽인다) ③ `MaxRateLimitWait = 2×RateLimitBackoff`
  경계 테스트 (F4 — 생존 뮤테이션 N1을 죽인다) ④ 대기 중 `CheckEntry` 단언 (F7)
  ⑤ `withDefaults`에 백오프 하한 클램프 + 테스트 (F6 — spec의 SHALL NOT을 노브
  값과 무관하게 성립시킨다) ⑥ 영구 거부 브로커(refusals 무한) 케이스 (A1 부수
  관찰 — 예산의 존재 이유를 산술이 아니라 무한 거부로도 고정) ⑦ 예산 < 백오프
  1회일 때 메시지 정직성 (F8). 뮤테이션 N1·N2 재가해 죽음 확인.
  → ① `ratelimit.go:108` `%w` + `errors.Is(err, context.Canceled)` 단언 (RED 실측:
  `= false; err = …: waiting out a rate limit: context canceled`).
  ② `orders.calls == 1` 단언 — N2 아래 `reads = 21, want 1`로 죽는다.
  ③ `TestRateLimitBudgetStopsExactlyAtTheBoundary`(30s/15s) — N1 아래 죽는다.
  ④ 대기 중 `CheckEntry` = recovery_incomplete 단언.
  ⑤ `withDefaults` B4를 `< DefaultRateLimitBackoff` 한 분기로 (포섭되는 죽은 분기를
  만들지 않는다) + `TestRateLimitBackoffNeverGoesBelowTheSurveyInterval`
  (RED: `RateLimitWaited = 5s, want the floor 15s`).
  ⑥ `TestPermanentlyRefusingBrokerExhaustsTheBudget` — 무한 거부, 20회 대기/5m/21 호출.
  ⑦ `TestBudgetTooSmallForOneBackoffSaysSo` (RED: 옛 문장이 한 번의 거부를
  "kept being refused"라 적었다). **147건 통과 · `make lint` rc=0 · N1·N2 둘 다 죽음.**

## 2. T1 — 산출물 정리

- [x] 2.1 tasks 체크 + 커밋 (제목에 [a102 §1], RED/GREEN/뮤테이션 증거를 본문에)

## 3. T2 — 겹2: 서베이가 엔진의 준비 신호를 기다린다 (T1 커밋 위에서)

- [x] 3.1 FLM/BTM **편집 전** — `enginelock.Hold`(AST 생성부터), `runEngineRun`
  (a098 FLM은 옛 base — 재기준), `runConsole`(AST 있음 44/21/114), 편집하는
  기존 함수 전부.
  → 편집 **전** 4종(AST + 커버리지 실측): `hold`(신규 4/5/14) · `runenginerun`(18/14/51) ·
  `engineruntime`(신규 6/8/11) · `runconsole`(44/21/114). GREEN 뒤 같은 도구로 **재측정**해
  2판을 병기했다(각각 3/4/10 · 18/14/53 · 6/7/12 · 44/22/119).
  겹2 표면의 AST-only 4종(`bootsurvey`·`runconfiguredsoakautostart`·`runconsole`·
  `startengine`)도 정리했다 — 앞의 셋 중 무편집 셋은 "이 change 무편집·기존 커버리지
  그대로"를 실측과 함께 명시했다. 시그니처가 닿은 **테스트 함수 13개**는 a056·a059 선례
  형식의 기계적 묶음으로 채웠다. `check_analysis.py` **rc=0**.
- [x] 3.2 [T] RED — enginelock: (a) `Ready(now)` 후 marker 파일에 `ready_at`이
  있고 (b) refresh 재작성이 그것을 보존하며 (c) `Ready` 멱등 (d) `Read`가
  ReadyAt을 노출한다는 테스트. Hold의 현재 반환(release func)으로는 컴파일부터
  실패한다 — 그것이 RED다.
  → `internal/enginelock/a102_ready_signal_test.go`. RED 실측: 컴파일 실패
  (`held.Release undefined (type func() has no field or method Release)` 외 10건).
- [x] 3.3 GREEN — D4: `Marker.ReadyAt`, `Hold` 핸들 반환(`Release`/`Ready`),
  refresh 보존. 호출자 engine.go:239 갱신.
  → `Held{mu, marker, live, done, once}` + `Release`/`Ready`/`refresh`(뮤텍스 보호 상태를
  읽는 ticker 본문) + `Status.Ready()`. 실패한 Hold는 **inert**(1판의 no-op release 의미
  보존). 기존 enginelock 11건 전부 통과 유지 → **19건**.
- [x] 3.4 [T] RED→GREEN — D5: `recoverThenReady(run, ready)` — 성공 시에만
  ready, 실패 시 절대 안 부름. runEngineRun 배선.
  → `cmd/tossctl/engineready.go`(신규). 취소(`ctx.Err() != nil`)도 ready를 부르지 않는다.
  `engineRuntime`에 `ready func()` 인자를 더해 6단계 핸들을 7단계 클로저로 내렸다.
- [x] 3.4b [T] **D5b (A1 F1)** — 같은 클로저가 Report를 받아
  `RateLimitWaits > 0`이면 obs 이벤트 한 줄 (몇 번·총 얼마나). 성공·실패 모두.
  → `engineRecoveryObserver(logger)` + `obs.EventRecoveryRateLimited`
  (`reconcile.recovery_rate_limited`, `count`/`duration_ms` 필드). nil logger 견딤.
- [x] 3.5 [T] RED→GREEN — D6: `awaitEngineReady(ctx, observe, clk, cap, poll)` —
  4 verdict(준비/엔진 없음/cap 초과/ctx 취소) 각각 fake clock·observe로.
  cap 120s·poll 2s 상수와 근거 주석.
  → 4 verdict 전부 테스트. 상한 초과는 **정확히 60회 폴링·합계 120s**로 산술 확인.
  엔진 부재는 매 회 확인한다(1초 만에 죽은 엔진을 상한까지 기다리지 않는다).
- [x] 3.6 D7: runConsole의 soak 블록을 goroutine으로, start seam을
  `awaitEngineReady`로 감싼다. `runConfiguredSoakAutostart`·`bootSurvey` 무편집.
  노트가 어느 쪽이었는지 말한다 — 조용한 cap 초과 금지.
  → `runConsole`의 분기 수 **44 그대로**, `go` 문 0 → 1. 두 함수 무편집(AST 해시 불변).
  새 함수는 `soakautostart.go`가 아니라 **같은 패키지의 새 파일**에 뒀다 — 그 파일에 붙이면
  `@@ -174,0 @@` 훅이 무편집 `rememberSoakApproval`을 "수정된 기존 함수"로 끌고 들어온다(실측).
  근거는 `cmd-tossctl--runconsole/function-logic-map.md`에 남겼다.
- [x] 3.7 [T] 뮤테이션 정산 — design §검증계약 (d)(e)(f). 기록 방식은 1.7과 같다.
  → (d)(e)(f) + 자체 (g)(h)(i)(j)(k), **총 8회 가함, 8회 전부 죽음.** 원복은 sha256
  동일성으로 증명. **(g)는 처음에 살아남았고**(파일 경유 Status는 stale이면 Marker가 zero다)
  술어를 네 조합으로 고정하는 테스트를 새로 써서 죽였다. (f)의 첫 형태는 컴파일 오류였으므로
  **뮤테이션이 아니라고 판정**하고 도달 불가 deadline 형태로 바꿔 다시 가했다.
- [x] 3.8 `go test ./internal/enginelock ./cmd/tossctl` + `make lint` rc=0.
  → **571건 통과**(19 + 552) · `-race` 포함 통과 · `internal/reconcile`·`internal/console`·
  `internal/obs`·`internal/app/engine` **1440건 무회귀** · `make lint` rc=0(gofmt 실재 확인).

## 4. T2 — 산출물 정리

- [x] 4.1 tasks 체크 + 커밋 (제목에 [a102 §3])

## 5. 리뷰 (Manager가 발주)

- [ ] 5.1 A1: T1 커밋 적대 리뷰 → 지적은 T1이 고치고 재커밋, 왕복 기록을
  review.md에
- [ ] 5.2 A2: T2 커밋 적대 리뷰 → 동일
- [ ] 5.3 gstack /review (브랜치 전체) → Fix-First 정산 → review.md 종합

## 6. Manager 완료 검증

- [ ] 6.1 뮤테이션 스팟체크 — T1·T2가 기록한 뮤테이션 중 각 1건을 직접 재현
- [ ] 6.2 tasks 전 항목 대조 + design 검증 계약 5조 확인
- [ ] 6.3 PM 갱신 (story measured/deviations/limits 실측 기입)
- [ ] 6.4 `make sdd-sync && make sdd-check && make gate CHANGE=a102-boot-does-not-starve-the-engine`
- [ ] 6.5 완료 보고 — 남은 위험과 not-applicable 목록 포함
