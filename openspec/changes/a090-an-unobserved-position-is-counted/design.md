# a090 · design (2판)

> **2판(2026-09-29) — 1라운드 적대 보이스 1(REJECT, P1 4) 반영.** Manager 판정(2026-09-29): F1~F4 전부 반영, F2 는 enqueue-only 로, F4 는 범위 안/
> 명명 잔여를 가른 표로. 1판 대비 바뀐 곳은 각 절 머리의 「2판」 표시. 1판 원문은 git 이력(`101f1d29`·`ceb7801f`)에 있다.
>
> 분기 주장은 전부 `analysis/function-logic/` 의 AST 두 번들(`ObserveOnce` 8 분기 · `workingSet` 22 분기, 이 문서보다 먼저)에서 온다.
> 진입 실측은 `analysis/harness/observeonce.blocks`(commit `eac13df1`). 줄 번호는 base(`d3bd1843`) 기준, `exitloop.go` sha256 `522d5d81…`.
> 코드에 없는 값·측정되지 않은 사실은 **[미측정]** 으로 적는다.

## D1. 무엇을 세는가 — 「보유·대상 포지션인데 이번 주기에 판정에 닿지 않음」 (2판: F4)

1판은 `ObserveOnce` 의 B6(`:453`)·B7(`:459`) 두 무음 `continue` 만 셌다. **거짓이었다** — `workingSet` 이 보유 포지션을 순회 **전에** 떨군다
(`workingSet` FLM: 탈락 다섯 자리 B8 · B10 · B12 · B14 · B21, 전부 시험 진입 0). 보유 포지션이 전부 떨어지면 `states` 가 비어 `ObserveOnce` B3
(`:431-439`)이 "Nothing is held" 로 **계정 두절 시계까지 리셋**한다.

그래서 2판은 **세는 단위를 자리가 아니라 집합으로** 바꾼다: 한 주기에 `workingSet` 이 본 **보유·exit 대상 포지션**(B5·B6 을 통과한 포지션)
중 그 주기에 **판정에 닿지 않은** 것이 미관측이다. 자리를 하나씩 계측하지 않으므로 새 탈락 자리가 생겨도 빠지지 않는다.

| 자리(보유 포지션이 판정에 못 닿는 곳) | 오늘 | a090 2판 | 근거 |
|---|---|---|---|
| `ObserveOnce` B6 `:453` 응답에 없음(0가격·NaN·Inf·오래된·미래 시각 포함, `observe` `:765-775`) | 무음 | **센다** — 원인 `no_quote` | 1판과 같음 |
| `ObserveOnce` B7 `:459` 사용 임대 만료 | 무음 | **센다** — 원인 `quote_expired` | 1판과 같음 |
| `workingSet` B8 `:527-531` exit state 열기 실패(진입 결정에 손절 없음 `:667-671` 이면 매 주기) | `cycle.Err` 로그뿐 | **센다** — 원인 `not_in_working_set` | 보유·대상인데 판정 안 됨 |
| `workingSet` B12 `:545-549` · B14 `:556-560` · B21 `:592-596` 격리 쓰기·읽기 실패 | `cycle.Err` 로그뿐 | **센다** — 같은 원인 | 같음 |
| `workingSet` B6 `:512-519` exit 대상 아님(진입 결정 없음) | 미관리 경보(`EventExitPositionUnmanaged`, 포지션 래치) | **범위 밖 — 명명 잔여** | exit 정책의 대상이 아니다(정본 「exit 대상이 아니며 발견 시 알림」). 그 경보가 normal 등급인 것은 정책 질문 |
| `workingSet` B10 `:533-537` 보유 중인데 exit state 가 완료 | 무음 | **범위 밖 — 명명 잔여** | 관측 실패가 아니라 정책 수명 문제(완료된 정책은 더 판정하지 않는다, `:534-536`). 후속 후보 |
| 격리 포지션(B11 · B17 · B20 → `refused` → `judge` 가 `alertRefused`) | 자기 critical | **관측됨으로 친다**(F10) | 판정 경로에 닿아 자기 critical(격리 공지·판정 거절)을 낸다 |
| `ObserveOnce` B1 양보 · B2 작업 집합 오류 · B4 전 종목 미응답 | 계정 사다리 | **무변화** — 그 주기에 포지션 단위 판정을 하지 않는다(보유 집합을 모르거나 계정 사다리가 본다) | 시계는 **마지막 판정 시각**에서 재므로(D3) 그 주기들도 뒤에 가서 세어진다 |
| 판정 진입 뒤 하류 임대 재검사 5자리(`:859` `:956` `:1027` `:1050` `:1180`) | 무음 | **범위 밖 — 명명 잔여**(Q3 확정) | 판정 진입을 "관측됨" 으로 본다. 그 포지션 **자신의** 판정이 매 주기 15초를 넘길 때만 무음 — 극단 edge `unsupported` |

**B7 의 현실 원인(F15).** 공식 클라이언트 시한이 15초(`internal/official/client.go:20` `defaultTimeout`)이고 시세 사용 임대도 15초(`execgw/retry.go:192`)
다 — 앞 포지션의 멈춘 요청 하나가 그 주기의 뒤 포지션 전부를 B7 로 보낸다. 그래서 B7 은 이론이 아니라 확정적으로 재현되는 과도다.

**tracer(F14).** `ObserveOnce` 의 둘째 생산 호출자 tracer(`tracer.go:297`)는 비시험 생성자 호출이 0 이고 단일 종목·무보유 계정을 요구해 B6·B7 에
닿지 않는다(한 종목 미스는 B4). 이 change 의 동작은 tracer 실행에서 관찰되지 않는다.

## D2. 상태 — 관측자 필드, 포지션 id 단위 (2판: F1 · F3 · F7)

포지션 id → `{seenAt, judgedAt, streakStart, cause, alerted}`. `seenAt` 은 `workingSet` 이 그 포지션을 처음 보유·대상으로 본 시각, `judgedAt` 은
마지막으로 판정에 닿은 시각, `streakStart` 는 현재 미관측 연속의 기점(D3). 지연 초기화(`quarantineAnnounced` `:245-250` 과 같이) — `NewExitObserver`
편집 없음.

- **정리는 보유 집합 기준(F4)**: 순회 뒤, 그 주기에 `workingSet` 이 보유·대상으로 표시하지 않은 포지션의 기록은 지운다(보유 종료·대상 제외). 경보 없음.
- **B3 조기 반환(F7)**: 보유·대상 표시가 있는데 `states` 가 비었으면(전부 탈락) 그 주기도 **순회 뒤 처리와 같은 판정**을 받는다(판정에 닿은 것 0).
  표시가 없으면(정말 무보유) 기록을 전부 지운다. 계정 시계 리셋(`:437-438`)은 **바꾸지 않는다** — 그 리셋이 보유 중 탈락을 가리는 문제는 포지션 단위
  판정이 대신 드러낸다(계정 사다리의 의미는 무변화, R5).
- **재시작하면 잃는다**(파일 머리 계약 `:67-68`). 재시작 뒤에는 `seenAt` 이 새로 찍혀 기점이 늦어진다 — 재시작 직전까지의 무관측 시간은 이 기록에서
  사라진다. 그 창은 계정 시계(`startedAt` 기점)가 덮는 범위만큼만 덮인다. **정직한 한계**로 적는다(원장에 기점을 저장하면 닫히지만 스키마 변경 — 범위 밖).
  원장의 `exit_states.last_observed_at` 은 기점으로 쓸 수 **없다** — 판정이 선을 움직이지 않으면 쓰이지 않는 값이라 "관측 안 됨" 이 아니라 "선이 안 변함"
  을 뜻한다(`internal/journal/exit_snapshot_integrity.go:9-12`, "an untouched stamp means the line did not change — not that nobody looked").

## D3. 임계 — **마지막 판정 시각**부터, 계정 사다리와 같은 60초 (2판: F1)

1판은 연속 기점을 **첫 미스**로 잡아, 그 앞의 양보(B1)·전 종목 실패(B4) 주기를 빼먹었다 — 무음 창이 최대 약 2×`outageAfter()` 였다(보이스 1 F1).

2판: **연속 기점 = 그 포지션의 `judgedAt`(없으면 `seenAt`).** 순회 뒤 판정 시점에 `now − 기점 ≥ o.outageAfter()` 이면 경보한다. 그래서 B1·B2·B4 주기가
기점 뒤에 끼어도 전부 시간에 들어간다 — 정본 시나리오 "보유 포지션의 가격 관측이 60초 이상 실패하면"(`exit-policy/spec.md:65`)의 문자대로다.

영수증(1판과 같음): 정본 `spec.md:62`(두절 60초 → critical + ENTRY_BLOCKED) · `DefaultExitObservationOutage` 60초(`exitloop.go:105`, "four intervals").
`obs.DefaultCriticalAttempts`(전달 재시도 3회)는 쓰지 않는다 — 다른 영역이고, 주기 수는 양보 주기를 세지 못한다.

## D4. 알림 — enqueue-only · 순회 뒤 · 에피소드 key (2판: F2 · F5, a094 D−5.2·D−5.3 과 같은 형태)

**루프 안에서는 기록만 한다. 임계 판정·알림·모드 강화는 순회가 끝난 뒤 한 번.** 이유(F2): 루프 안의 동기 알림은 경보당 최악 54초
(`obs.DefaultAlertDeliveryBound`, `internal/obs/alert_lease.go:57` "With today's defaults it is 54s", `n.mu` 아래)를 뒤 포지션의 손절 판정 앞에 세우고,
그 지연이 15초 임대를 태워 **B7 을 스스로 만든다.** 같은 파일이 이미 같은 규칙을 둔다 — 격리 행을 맨 뒤로(`exitloop.go:608-610` "every valid
position—including an emergency breach—is recorded/armed/submitted before alert delivery can wait").

- **전송 경로: enqueue-only** — `o.opts.Journal.EnqueueAlert`(`internal/journal/outbox.go:131`, "Use it when the alert only has to be *recorded*"). 전송은
  a098 전달 실행자가, 계속 실패 시 진입 차단은 a124 판정이 한다. 관측 루프는 전송을 기다리지 않는다. a094 D−4.6 과 같은 근거·같은 형태.
- **이벤트**: 기존 `obs.EventExitObservationOutage`(critical, `event.go:332`) — 새 타입 없음.
- **key = 에피소드 신원**: `type|account|positionID|<연속 기점 RFC3339>`. 에피소드 = **미관측 연속**, 신원 = 그 연속의 기점(`judgedAt` 또는 `seenAt`) —
  한 연속에 하나뿐인 원장 밖 사실이지만 **유한**하다(연속은 "판정됨 → 판정 안 됨" 전이마다 하나). a094 는 park attempt id 를 에피소드로 쓴다(a094
  D−5.2 — 교차 인용). outbox 의미론 무변경: `EnqueueAlert` 는 재알림 창 0 이라(`outbox.go:142-146`) 같은 key 를 다시 적재하면 옛 행을 재사용하고
  보내지 않는다 — **같은 연속은 한 번**, 새 연속은 새 행(F5 해소). 재시작 뒤 같은 포지션은 `seenAt` 이 새로 찍혀 **새 에피소드**가 된다(D2 한계와 짝) — a094 의 park attempt(원장 행이라 재시작에도 같은 에피소드)와 **다른 점**이며, 원장에 판정 시각의 믿을 만한 기록이 없어서다(D2).
- **적재 실패(a094 D−5.3 과 같은 형태)**: 적재 전에 `ClearEpoch(ReasonAlertUndelivered)` 를 읽고, `EnqueueAlert` 가 실패하면
  `o.opts.Retrier.Gate.BlockUnlessClearedSince(ReasonAlertUndelivered, epoch, detail)`(`execgw/retry.go:571`; `Retrier.Gate` 필드 `:310` — 새 배선 없음)
  로 진입을 잠그고 내용 없는 로그 한 줄. 관측 루프·청산 무영향. 범위(계정 단위)는 a094 Q7-1 과 같이 따른다.
- **필드**: `account`, `symbol`, `position_id`, `unobserved_seconds`, `cause`. 본문이 **포지션을 명명**해 계정 두절 경보와 구별된다(F6). 계좌번호·잔고·가격 없음.
- **연속당 1회**. 판정에 닿으면 연속이 끝난다.

## D5. 모드 강화 — 정본 준수(Q1 확정), 순회 뒤 (2판: F2 · F6)

`EscalateOperatingMode(ctx, account, ModeTriggerExitObservationOutage, announcer)`(`:846-847` 과 같은 호출·트리거)를 **순회 뒤**, 경보를 적재한
연속마다 1회. 이미 그 모드면 행도 공지도 없다(`operating_mode.go:409-421` — "Already there. Not an error, and not a row"). 그래서 공지(동기)는 **모드가
실제로 바뀔 때 한 번**이고, 그 공지는 다음 주기 시작을 늦출 수 있으나 같은 주기의 다른 포지션 판정 앞에는 서지 않는다(순회 뒤).

**Q1 의 비용 보충(F6).** 운영자가 완화한 뒤에도 같은 연속이 이어지면 다시 강화하지 않는다(연속당 1회). 그러나 **재시작마다** 새 에피소드라(D2·D4)
정지 종목을 보유하면 재시작할 때마다 진입이 다시 막힌다. 모드 이력의 트리거는 계정 두절과 같은 `EXIT_OBSERVATION_OUTAGE` 라 이력만으로는 둘을 못
가른다 — 경보 본문(포지션 명명)으로 가른다. 이 비용은 Q1 의 정본 준수 판단과 같이 기록한다.

## D6. fail-closed 가 거부할 정상 입력 — 열거 (2판: F8 · F15)

| 입력 | 오늘 | a090 뒤 | 근거 |
| --- | --- | --- | --- |
| 장 마감 시장의 보유 종목 | 관측됨 | **변화 없음** | **생산 증거**: 비거래일(토) `exit_states.last_observed_at = 2026-08-08T01:15:19Z`(a096 `proposal.md:173`) — 마감 중에도 exit 배치가 관측했다. 보조: a112 결정 46(2026-08-28, 단일 종목 · L1c 엄격 리더, 배치 경로 아님). 거래소 어댑터는 `FetchedAt` 을 읽은 시각으로 채운다(`official/market_reads.go:175`). **혼합 시장 배치(한 시장만 마감)는 [미측정]** |
| 거래정지·관리·상장폐지 절차 종목 | [미측정 · 사전 승인된 실측 대기] | 응답에서 빠지거나 0가격이면 60초 뒤 경보(+강화), 연속당 1회 | 코드 주석은 정지 종목이 답하지 않는다고 믿는다(`exitloop.go:869-870`) — 측정 아님. Q2 |
| `Last = 0` 인 보유 종목 | 무음 | 60초 뒤 경보 | [미측정 · 사전 승인된 실측 대기] — Q2 |
| B7 과도 — 앞 포지션의 멈춘 요청(클라이언트 시한 15초 = 임대 15초) | 무음 | 한두 주기는 경보 아님. 60초 연속이면 경보 | `official/client.go:20` · `execgw/retry.go:192` · 시험 :833(16초 제출) |
| exit state 열기가 매 주기 실패하는 보유 포지션(진입 결정에 손절 없음) | 로그뿐, 판정 영영 없음 | 60초 뒤 경보 | `workingSet` B8 · `openState` `:667-671` — 경보가 옳다 |
| 응답 종목 표기 불일치 | 무음(영구) | 60초 뒤 경보 | `:452`·`:760`·`:764` — 경보가 옳다 |

**생산 빈도는 [미측정]이다** — 이 change 가 그 흔적을 처음 만든다(D7).

## D7. 임계 아래의 가시성 (2판: F3)

1판은 임계 아래 미관측을 메모리에만 두어 어디에도 흔적이 없었다 — a089 원문("종목 단위로 **세고**")·a092 R1("미관측으로 **계수된다**")·delta("조용히
건너뛰어서는 안 된다") 미충족.

- `ExitCycle.Unobserved int` — 그 주기의 미관측 포지션 수(상태 표면·시험이 읽는다). `ExitCycle` 은 구조체 선언이다(함수 편집 아님).
- **연속 시작**과 **연속 해제** 때 normal 구조화 로그 한 줄씩: `position_id` · `symbol` · `cause` · (해제 시) `unobserved_seconds`. 매 주기 반복하지 않는다
  (`reportCycle` 의 "One line every five seconds forever is not observability", `exitloop.go:380-381` 와 같은 판단).
- 이벤트 타입: 기존 `EventExitObservationOutage` 를 **로그**로 쓴다(`o.log`, critical 등급은 알림 경로에서만 적용 — 로그는 등급 없음). 새 타입 없음.

## D8. 무엇을 편집하는가 (2판)

| 편집 | 성격 |
| --- | --- |
| `workingSet`(기존) | B6 통과 뒤(`:520`) **"보유·대상으로 표시" 호출 1개.** 탈락 다섯 자리는 무편집. 분기 조건·이탈 무변화 |
| `ObserveOnce`(기존) | B6·B7 의 `continue` 직전 원인 기록 1개씩 · 판정 진입(`:464`) 표시 1개 · B3 조기 반환 앞 순회-뒤 처리 호출 1개 · 순회 뒤(`:469` 앞) 처리 호출 1개. **분기 조건·이탈 무변화** |
| `ExitCycle`(구조체) | `Unobserved int` 필드 |
| `ExitObserver`(구조체) | 기록 맵 필드(지연 초기화) |
| 새 파일 `internal/app/engine/exit_unobserved.go` | 표시·기록·순회 뒤 판정·적재(에피소드 key)·적재 실패 잠금·강화·정리·로그 |

새 브로커 호출 0 · 원장 스키마 0 · 토글 0 · 새 이벤트 타입 0 · 새 reason/trigger 0 · 새 배선 0(`Retrier.Gate` 사용).

## D9. 잔여 — 이름 붙임

- **동기 경보 경로의 기존 알림**(`o.alert` 호출자 7 — `noteDelay` 등)은 그대로다(a092 소관). a090 이 더하는 경보만 enqueue-only.
- **재시작 창**(D2): 재시작 직전의 무관측 시간은 이 기록에서 사라진다.
- **workingSet B6(미관리)·B10(완료 정책)** — 범위 밖, 후속 후보(D1 표).
- **하류 임대 재검사 5자리** — Q3.

## Q — 결정 기록 (Manager, 2026-09-29 — 셋 다 사용자행 아님)

- **Q1 = (a) 확정 — 포지션 단위 두절에 critical + ENTRY_BLOCKED.** 근거: 정본이 이미 요구한다 — `openspec/specs/exit-policy/spec.md:62`
  「관측 경로와 fail-safe」 "관측 두절이 staleness 임계(기본 60초)를 넘으면 critical 알림 + ENTRY_BLOCKED 자동 강화가 발동한다(SHALL …)" 와
  `:65` 시나리오 "**보유 포지션의** 가격 관측이 60초 이상 실패하면". **정본 준수이지 신규 정책이 아니다.** 대가(D6: 정지 종목 하나가 사람이 완화할
  때까지 신규 진입을 막음 · D5: 재시작마다 재강화)는 진입이지 청산이 아니며, 자동 경로는 조이기만 한다는 승인 원칙에 맞다. spec delta 는 그대로(MODIFIED 불요).
- **Q2 = 구현 로트로 이연 + 사전 승인.** 정지·0가격 종목의 `/prices` 응답은 **[미측정 · 사전 승인된 실측 대기]** 다. 승인 범위: 읽기 전용 시세
  GET 1회(정지 종목 포함, 쓰기 0), 구현 로트가 **장중에** 1회 실행하고 그 결과로 D6 의 두 [미측정] 행을 확정한다. 측정은 경보 **빈도**를 알려 줄 뿐
  판정 규칙은 같다 — 결과가 무엇이든 구현이 멈추지 않는다.
- **Q3 = 초안 그대로 — 관측점은 판정 진입(`:464`).** 근거: 헌장(`ObserveOnce` B5/B6 최소 편집)과 일치하고 a094 의 `record` 편집과 겹치지 않는다.
  **하류 무음 5자리는 명명된 잔여**다 — `judge` `:859` · `judgeRatchet` `:956` · `judgeLadder` `:1027` · `refreshObservation` `:1050` ·
  `record` `:1180`(전부 `!o.quoteUsable(quote)` → `return nil`). **후속 change 후보** — 이 change 는 D1 의 `unsupported` 로 닫는다.
