# a205 design — 청산 관측 주기 5s → 2s

분기 · 호출 주장은 `analysis/ast-evidence/claims.md`(C1~C4, base `e2921fc00` AST)를 근거로 씀. 위험도 **High-risk**(손절 · 익절 관측 경로).

## §1 사실 (Manager 실측 2026-11-27 — 읽기 전용, 계좌 · 종목 · 금액 없음. 팀메이트 재측정 안 함)

| # | 사실 | 이 change 에서의 쓰임 |
|---|---|---|
| F1 | US 1종목 `exit_states.last_observed_at` 0.25s 폴링 539초: 간격 106개 중앙 5.083s, 범위 5.067~5.158s, 6s 초과 0 | 주기 = 대기 + 처리. 2s 면 실제 간격 ≈ 2.07~2.16s |
| F2 | KR 거절 반복 구간 처리시간: 10월 p50 0.22 · p90 0.31 · p99 0.68s, 11월 p99 0.49s, 최대 2.6s(8월) | 처리 > 대기 가 되는 드문 주기도 겹치지 않음 — `Run` 은 처리 뒤 잔다(C1). 최대 2.6s 면 그 주기 간격 4.6s |
| F3 | 60s 이상 두절(`exit.observation_outage`) 10~11월 9회, 전부 계정 전체 브로커 읽기 장애(filldetect loop_degraded 동반) | 두절 임계는 시간 기준이라 주기와 무관(D3) |
| F4 | 체결 감지 SLO 양보(B1) 실측 0회(fill_events 지연 p95 KR 4s · US 3s) | 양보 갈래는 주기와 무관하게 그대로 |
| F5 | 가격 읽기는 주기당 배치 1회(`internal/official/market_reads.go:142`, 종목을 콤마로 묶음) → 분당 12 → 30 | §5 예산 |
| F6 | 계좌 정책 상한(a166) 3 req/s, 운영 예산 OFF | §5 예산 |
| F7 | heartbeat 는 포지션마다 매 주기 `exit_states` UPDATE, 거절 반복 구간은 매 주기 `exit_events` 행. journal.db 약 291MB | §5 저널 |
| F8 | KR · US 는 같은 루프 · 같은 가격 배치 — 상수 하나가 두 시장을 함께 바꿈 | 사용자가 고른 범위 |

재측정한 것(팀메이트, 코드만): 생산 배선은 `ExitObserverOptions.Interval` 을 넘기지 않음(C1) · 엔진 패키지 시험 전체가 상수 2s 에서 rc 0(사전 탐침, tasks 2.1).

## D1 — 상수 변경, 설정 아님

`DefaultExitObservationInterval = 2 * time.Second`. 설정으로 만들 이유가 없음: 사용자는 값 하나(2s)를 골랐고 설정 가능을 요구하지 않음(YAGNI).
설정이면 하한 · audit · 콘솔 사본과의 일치 규칙이 따라와야 하는데(안전 불변식 §0.5), 그 비용을 정당화할 요구가 없음.
`ExitObserverOptions.Interval`(시험 seam)은 그대로 둠 — 생산 배선은 0 을 넘겨 상수를 씀(C1).

**값 2s 의 근거.** 처리시간 p99 0.49~0.68s(F2)가 대기 2s 보다 작음 — 주기 대부분이 대기임. 대기가 처리보다 짧아지는 1s 이하는 사용자가 고르지 않음.
손절 · 익절 트리거를 늦게 볼 수 있는 최대 폭(표본 간격)은 약 5.1s → 약 2.1s(F1). 보수 방향(안전 불변식 §0.9 — 판정 규칙 · 손절폭 · 사이징 무변경, 관측을 더 자주 할 뿐).

## D2 — 하한가 갱신기: 읽기 시한은 주기에서 계속 유도한다(코드 변경 0)

> **구현 리뷰 r1(Codex) 뒤 정정.** 초판은 (가)가 a087 의 「겹침 금지」를 「지킨다」, 재시도가 잦아 첫 성공까지의 기대 시간이 「늘지 않는다」고 적었음.
> 둘 다 입증되지 않음(r1 #2 · #3) — 아래가 정정본.

**사실(C2 · C3).** 갱신기는 관측 goroutine 밖(보조 실행자)이고 토큰 보내기는 막히지 않으므로, 갱신기가 무엇을 하든 손절 판정 시간은 그대로다.
갱신 한 번의 **예산**은 `사전조회(interval − read) + 읽기(read)` = interval, `read = min(2s, interval/2)`. 주기 2s 면 읽기 1s · 사전조회 1s, 시도 빈도는 토큰당 1 → 2.5배.

**생산에서는 지금 꺼져 있음.** `Context` 가 판을 만들 때 admission 을 `nil` 로 넘김(`internal/app/engine/engine.go:627`
`newExitLowerLimitBoard(clk, off, jrn, jrn, accountRef, nil)`) → `canRefresh` 거짓 → `exitLowerLimitBoard.Run` B1 이 취소까지 대기, 하한가 GET 0
(a166 admission 이 배선되기 전까지 — 함수 주석 「admission이 없으면 취소까지 대기하여 a166 전 broker 읽기를 0으로 유지함」). 그러므로 아래 위험은
**admission 이 배선된 뒤에만** 생기는 조건부 위험이다. 이 change 는 생산 하한가 동작을 바꾸지 않는다(GET 0 → GET 0).

**겹침 금지(a087 SHALL NOT)와의 관계 — 예산과 비중첩은 다른 것.** 예산 합 = 주기는 a087 이 5s 에서 쓰던 구조를 같은 비율로 유지할 뿐,
시간상 비중첩을 **보장하지 않는다**: 시한은 갱신기가 토큰을 **받은 뒤** 시작하고(`refresh` :191 · :221), 관측자는 토큰을 보내자마자 따로 잔다(`Run` :434 · :435).
토큰 수신이 늦으면 갱신이 다음 `/prices` 와 겹칠 수 있음 — 이 틈은 5s 에도 같은 모양으로 있었고(a087 의 「겹침 0」 논증의 같은 틈), 2s 에서는 여유의 절대값이 줄어
더 잦을 수 있음. 비중첩을 강제하는 절대 마감 · 동기화는 새 가드라 이 change 에서 만들지 않음 → **사람 결정**(남은 위험 R1b).

| 안 | 2s 에서 읽기 시한 | 갱신 한 번의 예산 vs 주기 | 판정 |
|---|---|---|---|
| (가) 공식 그대로 — `cycleComplete(o.Interval())` | 1s | 예산 = 주기(5s 때와 같은 구조) | **채택** — 코드 변경 0, 생산 dormant |
| (나) 주기와 분리 — `cycleComplete(5s)` 등 고정 예산 | 2s (사전조회 3s) | 예산 5s > 주기 2s — 토큰 지연 없이도 다음 `/prices` 와 겹침 | 기각 — a087 SHALL NOT 을 구조적으로 깸 |
| (다) 공식을 고쳐 읽기 몫을 키움(예: 사전조회 0.5s · 읽기 1.5s) | 1.5s | 예산 = 주기 | 기각 — 기존 High-risk 함수 편집, 근거 수치(`/price-limits` 지연) 미측정, 생산 dormant 라 이득 0 |

**(가)의 대가(활성화 뒤에만).**

1. 손절 즉시성: 갱신기는 판정 경로 밖이라 주기와 무관하게 0 지연(C2). 판정 주기는 빨라짐.
2. 읽기 시한 2s → 1s: `/price-limits` 응답이 1s~2s 대에 몰리면 5s 때는 성공하던 읽기가 매번 실패할 수 있음 — 재시도가 잦아도 그 경우 첫 성공은 오지 않음(r1 #3).
   같은 계좌 공식 단건 GET 3일 p99 233~569ms(R7, `verify-execution-capability/analysis/retry-matrix-slo-2026-10-22.md:27`)는 1s 아래지만 `/price-limits` 자체 분포는 **미측정**.
3. 실패의 결과: 그 주기 청산 가격은 「오늘의 LIMIT」(관측가 지정가, `exitFallbackLimit`) — 제출은 계속되지만 하한가 지정가와 **동등한 보호는 아님**
   (a087 order-execution delta :34 가 둘의 차이를 적음). 「약화 아님」이 아니라 「활성화 뒤 `/price-limits` 지연이 1s 를 넘는 비율만큼 a087 의 이득이 줄 수 있음」.
4. 기록은 그날 유효(`recordValidNow`)라 종목당 하루 1 성공이면 끝 — 시도 빈도 2.5배는 기록이 빈 동안에만.
5. admission 배선(a166 활성화) 전에 `/price-limits` 지연을 재고, 1s 가 모자라면 (다)를 별도 change 로 — **사람 결정**(R1).

## D3 — 시간 기준 상수는 그대로, 틀린 주석은 고침

| 상수 | 값 | 판정 |
|---|---|---|
| `DefaultExitObservationOutage` | 60s | 그대로 — 시간 기준(`checkOutage` 가 `o.clk.Now().Sub(since) < o.outageAfter()` 로 잼 — `exitloop.go:961`). 주석 「It is four intervals」은 base 에서 이미 틀림(60/5 = 12) → 「thirty intervals at the default」로 고침(60/2 = 30). 의미(한 번의 실패 · 양보는 사고가 아니고 지속은 사고)는 유지 |
| `execgw.QueryPriceEvidenceDuration` | 15s | 무관 — 시세 사용 기한. 주기가 짧아져 더 여유 |
| `DefaultExitLiquidationDelayBound` | 30s | 무관 — IN_DOUBT 해소 절차 기한 |
| `externalBuyListTimeout`(a197) | 3s | 값 그대로. 주석의 「관측 주기 5 초보다 짧고」만 틀리게 됨 → 근거를 a197 design 의 실제 유도(R7 p99 ≤ 0.6s 의 5배, 관측 max 4.17s 보다 짧음)로 고침. **3s 는 목록 판독만의 시한**이고 단계 전체에는 시한이 없음(같은 파일 :159 — 사전 단건 조회 2s · 취소 15s 는 어댑터 시한, r1 #4 정정). 그 단계가 도는 주기는 다음 주기 시작이 단계 체류만큼 밀림(파일 머리 D4 · D5 공시) — a197 의 기존 대가이고 a205 가 새로 만든 지연이 아님. 주기가 짧아도 그 체류는 그대로 |
| 엔진 감독 `DefaultDegradationThreshold`(연속 5주기) | — | 무관 — exit 루프는 Health 없이 등록됨(`cmd/tossctl/engine.go:735-740` 주석: 60s 두절 계약이 그 사다리) |

그 밖에 「every five seconds」 류 산문(`exitloop.go:57 · :455`, `exit_quarantine_announce.go`, `reconcileloop.go:499`, `exitpolicy/adoption.go:28` 등)은
과거 사건 · 근거 서술이라 고치지 않음. `exitpolicy/adoption.go:28` 의 「두 5초 틱 사이 잡음」은 편입 손절 하한 0.02 의 근거인데, 틱이 짧아지면 그 잡음은
작아지므로 하한은 그대로 보수적임.

## D4 — 콘솔 두 상수: 사본 유지, 값 2s

콘솔은 엔진 패키지를 import 할 수 없음 — `internal/console/static_test.go` `TestTheConsoleHoldsNoBrokerOfItsOwn` 의 금지 목록에
`internal/app/engine` 이 있고(:598), 두 상수의 주석도 같은 경계를 이유로 사본을 둔다고 적음. 그래서 사본 + 값 단언 시험 유지.

- `engineRuntimeInterval` 5s → **2s**: operator-console 정본 「살아 있는 대사 차단 … 읽기의 간격은 엔진의 관측 주기를 넘어서는 안 된다」(SHALL) — 엔진이 2s 면 5s 는 위반.
  비용: runtime 읽기는 DB 를 안 건드림(Unix dial + descriptor 읽기, 주석) — 쓰기 커넥션 경합 없음. lifecycle(30s, 쓰기 커넥션 경합)은 그대로.
- `lineRefreshInterval` 5s → **2s**: 정본 「보호선 화면은 … 엔진 주기를 따라간다」. 전체 재로드 감당 근거 셋(접힘은 URL · 편집할 form 없음 · 브로커 상한은 캐시 TTL)은 그대로 성립.
  브로커 호출은 TTL 이 묶어 늘지 않음(a080 시험 `TestReloadingAtTheEngineCadenceKeepsTheBudgetCeiling` · `TestTheOverviewSpendsNothingHoweverOftenItReloads`). 엔진 읽기는 위 캐시가 묶음.
  대가: 열린 화면이 2초마다 다시 그려짐(브라우저 재로드). 사람이 보는 화면의 깜박임이 거슬리면 별도 change.
- 표류 위험: 엔진을 2s 보다 짧게 바꾸고 콘솔을 안 바꾸면 runtime 간격이 정본을 위반함. 새 가드는 만들지 않음 — 두 콘솔 시험이 2s 를 고정하고 주석이 엔진 상수를 이름으로 가리킴.

## D5 — exit-policy spec 델타

`관측 경로와 fail-safe`: 「주기 기본 5초」 → 2초, 「관측 주기는 … 기본 5초이며」 → 2초. 시나리오 하나 추가(기본 주기 2초).
`flat 관측은 기존 cadence…` 의 시나리오 이름 「5초 cadence의 flat position」은 고치지 않음 — 본문에 주기 값이 없고(호출 수 · 순서만 요구), openspec 1.16 의 MODIFIED 는 시나리오 이름 바꾸기를 「기존 시나리오 누락」으로 거절함(strict validate 실측). 이름표가 옛 값을 담는 것은 남은 위험 R6.
`:687` 「막힌 손절이 5초마다 같은 줄을 쓰지 않는다」는 근거 괄호라 그대로(같은 줄 반복 금지는 주기 값과 무관). operator-console 은 숫자를 쓰지 않으므로 델타 없음.

## D6 — 토글 OFF = upstream

이 change 는 토글이 아니라 기본값 변경임. upstream tossinvest-cli(`upstream/main`)에는 `internal/app/engine` 도 `internal/console` 도 없음
(`git ls-tree -r --name-only upstream/main | grep -c '^internal/app/engine\|^internal/console/'` → 0 · 0). 청산 루프와 콘솔은 둘 다 TossOS 전용이다.
엔진 쪽 변경은 엔진이 도는 동안만 효과가 있고, **콘솔 재로드 주기는 엔진 실행 여부와 무관하게 바뀐다**(r1 ⑦ 정정 — 엔진이 꺼져 있어도 열린 보호선 화면은 2초마다 재로드).
둘 다 upstream 에 없는 표면이라 「OFF = upstream」 규칙의 대상(upstream 동작)이 없음. upstream 상속 시험 영향 없음.
판단: 「사람 결정 필요」 아님.

## §5 부하 (늘어나는 것, 상한 없는 것은 남은 위험)

| 항목 | 5s | 2s | 근거 |
|---|---|---|---|
| 청산 가격 배치 GET | 0.2 req/s | 0.5 req/s | F5 |
| 하한가 GET(KR 보유 · 그날 기록 없음 · 주문 창 안) | ≤0.2 req/s | ≤0.5 req/s | C3, 기록이 차면 0 |
| 체결 감지 OPEN 목록 | 0.33 req/s | 그대로 | retry matrix |
| 막힌 손절의 재제출(브로커 거부가 반복되는 포지션마다) | ≤0.2 POST/s | ≤0.5 POST/s | F7 · a197 found_by(하루 6,206회 거절) — 재제출 감쇠 없음(`rg -i backoff` 0) |
| 계좌 공유 상한 | 3 req/s | 3 req/s | F6 — 운영 예산 OFF 라 프로세스 간 강제는 없음 |

위 표는 **일부 경로의 평균 빈도 추산**이지 계좌 안전 상한이 아님(r1 #5 정정). 정상(막힌 손절 0)이면 이 표의 엔진 몫 ≈ 0.5 + 0.33 (+ 하한가 ≤0.5, 생산 dormant) req/s,
막힌 손절이 K 개면 + 0.5K. 표에 **없는 것**: a197 단계의 목록 판독(최대 5쪽, `a197_external_opposite_buy.go:260`) · 사전 단건 조회 · 취소, 공식 클라이언트의 401 재인증 · 재전송
(`internal/official/client.go:325`), 다른 루프(대사 · 매수 새로고침 등), 같은 계좌를 치는 다른 프로세스(soak · verify · CLI · 이웃 제품), 같은 초에 몰리는 버스트(평균 req/s 로 안 묶임 —
R2 는 간격 없는 7요청/535ms 에서 429). 그러므로 3 req/s 에 닿는 K 의 경계는 이 표로 정할 수 없고 K 보다 작을 수 있음. 남은 위험 R2.

저널: heartbeat UPDATE 포지션당 0.2 → 0.5/s, 거절 반복 구간 `exit_events` 행 2.5배(F7). journal.db 성장률은 그 행 수에 비례 — 미측정, 남은 위험 R3.

## §6 시험 전략

- RED: 새 파일 `internal/app/engine/a205_exit_observation_interval_test.go` — (1) `DefaultExitObservationInterval == 2s`, (2) 생산 배선 기본(옵션 0)인 관측자의 `Interval()` 이 2s,
  (3) 하한가 예산 유도값: `lowerLimitReadTimeout(기본 주기) == 1s` 이고 사전조회 몫 `주기 − 읽기` 가 양수(`refresh` B1 대체 갈래로 가 예산이 2×읽기가 되지 않음).
  이 시험은 **예산 값만** 고정함 — 실제 갱신기를 돌리지 않으며 시간상 비중첩을 증명하지 않음(r1 #2). 콘솔 두 시험 기대값 2s,
  콘솔 `TestTheOverviewSpendsNothingHoweverOftenItReloads` 의 재로드 횟수 10 → 16(창 = 횟수 × 2s 가 TTL 30s 를 넘어야 peek 와 get 을 가름 — r1 #1).
- GREEN: 상수 3개(엔진 1 · 콘솔 2) + 주석.
- 회귀: `go test ./internal/app/engine/... ./internal/console/...`(엔진 패키지 전체는 상수 2s 사전 탐침에서 rc 0), `cmd/tossctl`(엔진 배선 · 콘솔 조립 시험),
  `make lint`, `make test-seams` 중 엔진 · 콘솔 태그 시험, `make test-race` 의 엔진 동시성 목록(관측 루프 · 갱신기 경합 시험 포함). 전체 시험 · `make gate` 는 요청 없음 — 안 돌림.
- 선택 근거: `codegraph affected … --filter '*_test.go'`(v1.6.0)는 import 사슬로 저장소 대부분(시험 파일 1,300여 줄, 100+ 패키지)을 잡아 선택에 못 씀. 상수의 실제 소비자는
  `rg DefaultExitObservationInterval|\.Interval\(\)` 로 `exitloop.go` · `tracer.go`(+ 시험 3곳)뿐이고 콘솔 두 상수는 콘솔 패키지 안에서만 쓰임.

## §7 FLM

바뀌는 기존 함수 본문(생산) 0 — 상수 선언 · 주석만. 바뀌는 기존 **시험** 함수 3개는 gate ⑤ 규칙대로 번들을 만듦:
`internal-console--TestTheRedrawCadenceIsTheEnginesAndNotTheCaches`, `internal-console--TestTheCacheIntervalsAreTheirOwnConstants`,
`internal-console--TestTheOverviewSpendsNothingHoweverOftenItReloads`(구현 리뷰 r1 #1 뒤 — 편집 전 scaffold, 편집 후 AST).
새 시험 파일의 함수는 base 에 없어 「modified existing function」이 아님.

## 남은 위험

- R1 (조건부 — 하한가 admission 배선 뒤) `/price-limits` 지연 분포 미측정 — 읽기 시한 1s 에서 실패율이 2s 시한 때보다 높을 수 있고, 그만큼 그 종목의 청산은 하한가 대신 「오늘의 LIMIT」. 활성화 전 측정 · 필요 시 (다) — 사람 결정.
- R1b (조건부) 토큰 수신 지연 시 하한가 GET 이 다음 `/prices` 와 겹칠 수 있음 — 5s 에도 있던 틈, 2s 에서 여유 축소. 비중첩 강제는 새 가드 — 사람 결정.
- R2 막힌 손절 재제출 빈도 2.5배 + 표 밖 호출 · 버스트(§5) — 계좌 3 req/s(운영 예산 OFF)에 닿는 경계는 미측정. 감쇠 · admission 은 이 change 범위 밖(새 가드 금지) — 사람 결정.
- R3 저널 쓰기 2.5배 — 성장률 미측정. 배포 뒤 journal.db 크기 · WAL checkpoint 를 관찰.
- R4 콘솔 화면 2초 재로드 — 사람이 보기 불편할 수 있음(기능 위험 아님).
- R5 엔진 · 콘솔 사본 표류 — 시험 두 개가 값을 고정할 뿐 서로를 대조하지 않음.
- R6 exit-policy 시나리오 이름 「5초 cadence의 flat position」이 옛 값을 담음(본문 무관, D5).

## 배포(사람 승인 뒤, 이 change 는 하지 않음)

- `make image CHANGE=a205-exit-observation-2s` → 엔진 교체. 엔진 정지 = 손절 없음이므로 교체는 두 시장이 닫힌 창에서 — 미국 정규장 마감 뒤(11월 표준시 KST 06:00) ~ KR NXT 개장(08:00) 전, 또는 KR 마감 뒤 ~ 미국 장 시작 전. 창은 실행 직전 사람이 고름.
- 콘솔도 같은 이미지라 함께 바뀜. 롤백: 직전 이미지 핀으로 교체(스키마 · 저널 변경 0).
