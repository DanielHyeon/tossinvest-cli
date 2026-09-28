# a090 · design (5판)

> **5판(2026-09-29) — codex 3라운드(REJECT, P0 0 · P1 4) 반영.** Manager 처분: R3-4 a092 입구를 **구현 하드 의존**으로(입구 밖 대안 삭제) · R3-2·R3-3
> 정화는 a090 자체에(a092 재개방 없음) · R3-1 tasks 수리 · P2 셋. **설계 freeze 는 a092 와 독립, 구현은 a092 `RecordAlert` 착지 뒤(하드 조건).**
>
> **4판(2026-09-29) — codex 2라운드(REJECT, P0 1 · P1 2) + a092 22라운드 교차 반영.** Manager 처분: R2-1 계좌 필드 무탑재 · **좁은 전용 로거**(관측자
> 전체 `Log` 배선 철회) · R2-2 공지 적재 별도 상태 · R2-3 `go/parser` 이탈 전수 핀 · P2 셋 · a092 K6 단일 입구 · 해제 세대 읽기 시점. 새 절 D12(계좌 정보).
>
> **3판(2026-09-29) — codex 1라운드(REJECT, P0 2 · P1 4) 반영.** Manager 처분: N1 B10 표시 해제(편집 2자리 + 구조 핀) · N2 강화 공지만 enqueue-only ·
> N3 최소 로거 배선 편입 · N4~N6 · P2 넷. 바뀐 곳은 「3판」 표시, 새 절 D10(실패 전이)·D11(보장의 조건).
>
> **2판(2026-09-29) — 1라운드 적대 보이스 1(REJECT, P1 4) 반영.** Manager 판정(2026-09-29): F1~F4 전부 반영, F2 는 enqueue-only 로, F4 는 범위 안/
> 명명 잔여를 가른 표로. 1판 대비 바뀐 곳은 각 절 머리의 「2판」 표시. 1판 원문은 git 이력(`101f1d29`·`ceb7801f`)에 있다.
>
> 분기 주장은 전부 `analysis/function-logic/` 의 AST 번들(`ObserveOnce` 8 분기 · `workingSet` 22 분기 · `Notifier.AnnounceOperatingMode` 2 분기, 이 문서보다 먼저)에서 온다.
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
| `workingSet` B10 `:533-537` 보유 중인데 exit state 가 완료 | 무음 | **범위 밖 — 명명 잔여. 3판: B10 진입의 첫 문장에서 표시를 해제한다**(N1) | 관측 실패가 아니라 정책 수명 문제(완료된 정책은 더 판정하지 않는다, `:534-536`). 표시는 B6 통과 뒤(`:520`)라 B10 포지션도 표시된다 — 해제하지 않으면 미관측으로 세어져 R14 와 모순. 후속 후보 |
| 격리 포지션(B11 · B17 · B20 → `refused` → `judge` 가 `alertRefused`) | 자기 critical | **관측됨으로 친다**(F10) | 판정 경로에 닿아 자기 critical(격리 공지·판정 거절)을 낸다 |
| `ObserveOnce` B1 양보 · B4 전 종목 미응답 | 계정 사다리 | **무변화** — 그 주기에 포지션 단위 판정을 하지 않는다(계정 사다리가 본다) | 시계는 **마지막 판정 시각**에서 재므로(D3) 그 주기들도 뒤에 가서 세어진다 |
| `ObserveOnce` B2 작업 집합 오류 | **어느 사다리도 재지 않는다**(`checkOutage` 를 부르지 않는다, `:427-430`) | **무변화 — 명명 구멍(D11-2)** | 4판 정정(codex 2라운드 R2-6): 2·3판은 B2 를 계정 사다리가 덮는다고 적었다 — 거짓 |
| 판정 진입 뒤 하류 임대 재검사 5자리(`:859` `:956` `:1027` `:1050` `:1180`) | 무음 | **범위 밖 — 명명 잔여**(Q3 확정) | 판정 진입을 "관측됨" 으로 본다. 판정 진입 뒤 즉시 끝나는 경우와 처리 중 임대 만료를 모두 포함한다(D9 「관측됨」 정의) — 4판 정정: "자기 판정이 15초를 넘길 때만" 은 과장이었다 |

**표시 자리 두 곳 — 표시 `:520`, 해제 B10 첫 문장(3판, N1).** 표시를 B10 **뒤**로 옮기는 안은 버렸다: 그 자리는 B8(열기 실패, `:527-531`)을 지난 뒤라
열기 실패 포지션이 표시되지 않아 **범위 안의 탈락(B8)을 잃는다.** 그래서 표시는 exit 대상이 확정된 첫 자리(B6 통과 직후)에 두고, 범위 밖인 완료 정책만
B10 진입에서 명시적으로 해제한다. 이 쌍은 **구조 핀**으로 고정한다(tasks 2.15): 해제는 B10 블록의 **첫 문장**이고, 표시 `:520` 과 해제 사이에 새
`continue`·`return` 이 끼면 빨강 — 새 탈락 자리가 표시 앞에 생기면 조용히 안 세어지는 모양(루프 머리 `continue` 교훈)의 쌍 판이다.

**B7 의 현실 원인(F15).** 공식 클라이언트 시한이 15초(`internal/official/client.go:20` `defaultTimeout`)이고 시세 사용 임대도 15초(`execgw/retry.go:192`)
다 — 앞 포지션의 멈춘 요청 하나가 그 주기의 뒤 포지션 전부를 B7 로 보낸다. 그래서 B7 은 이론이 아니라 확정적으로 재현되는 과도다.

**tracer(F14 · 4판 정정 · 5판 — tasks·FLM 도 맞춤).** `ObserveOnce` 의 둘째 호출자 tracer(`tracer.go:297`)는 비시험 생성자 호출이 0 이다. 단일 종목이라 B6 에는 닿지 않지만(한 종목 미스는
B4) **B7 에는 닿을 수 있다** — 가격 읽기 뒤의 원장 작업(대체 관측 순번 복구 `exitloop.go:793`)이 사용 임대를 태울 수 있다(codex 2라운드 R2-6). 2·3판의
"B6·B7 도달 불가" 는 틀렸다.

## D2. 상태 — 관측자 필드, 포지션 id 단위 (2판: F1 · F3 · F7)

포지션 id → `{seen, judged, streak, cause, state}`. `seen` 은 `workingSet` 이 그 포지션을 처음 보유·대상으로 본 순간, `judged` 는 마지막으로 판정에
닿은 순간, `streak` 는 현재 미관측 연속(기점 · 에피소드 id), `state` 는 D10 의 실패 전이 상태. **3판(N4)**: 각 순간은 **단조 앵커**(`clock.LeaseAnchor(o.clk)`,
`exitloop.go:756` 과 같은 헬퍼)와 표시용 벽시계 UTC 를 **따로** 가진다 — 경과는 앵커로만 잰다. 지연 초기화(`quarantineAnnounced` `:245-250` 과 같이) — `NewExitObserver`
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

2판: **연속 기점 = 그 포지션의 `judged`(없으면 `seen`).** 순회 뒤 판정 시점에 `clock.LeaseElapsed(o.clk, 기점 앵커) ≥ o.outageAfter()` 이면 경보한다
(**3판 N4** — 벽시계 `now − 기점` 이 아니다. 생산 `Clock.Now()` 는 단조 성분을 벗긴다(`internal/clock/clock.go:77`·`:81`)라 벽시계 역행이 창을 늘린다;
`LeaseElapsed` 는 사용 임대가 이미 역행을 막으려고 쓰는 헬퍼다, `exitloop.go:1076`).
**시험 픽스처 계약(4판 R2-5)**: 이 헬퍼는 **주입 시계를 단조로 만들지 않는다** — 주입 시계에서는 `Now`/`Since` 로 떨어진다(`internal/clock/clock.go:57-69`).
일반 `clock.Fake` 를 되감으면 경과도 되감긴다. **5판(R3-7)**: a111 분리 픽스처(`a111WallElapsedClock`)도 그대로는 안 된다 — `wallOffset` 을 뒤로 돌린 뒤 만든
앵커는 `Since` 가 즉시 그 오프셋만큼을 낸다(`a111_flat_exit_observation_test.go:135-146`). 그래서 역행 시험은 **앵커를 인식하는 픽스처**(앵커 발급 시점의 경과
기준을 기억해 생성 직후 경과 0)를 새로 쓴다. 참고로 옛 문장: 역행 시험(R3f)은 a111 의 벽시계/경과 분리 픽스처(`a111_flat_exit_observation_test.go:127-146`
`a111WallElapsedClock` — `Now` 와 `Since` 를 따로 움직인다)를 쓰고, 앵커를 역행 **전과 후** 모두에서 만든다. 계정 사다리(`checkOutage` `:822`)는 여전히
벽시계다 — 이 change 가 바꾸지 않는다. 그래서 B1·B2·B4 주기가
기점 뒤에 끼어도 전부 시간에 들어간다 — 정본 시나리오 "보유 포지션의 가격 관측이 60초 이상 실패하면"(`exit-policy/spec.md:65`)의 문자대로다.

영수증(1판과 같음): 정본 `spec.md:62`(두절 60초 → critical + ENTRY_BLOCKED) · `DefaultExitObservationOutage` 60초(`exitloop.go:105`, "four intervals").
`obs.DefaultCriticalAttempts`(전달 재시도 3회)는 쓰지 않는다 — 다른 영역이고, 주기 수는 양보 주기를 세지 못한다.

## D4. 알림 — enqueue-only · 순회 뒤 · 에피소드 key (2판: F2 · F5, a094 D−5.2·D−5.3 과 같은 형태)

**루프 안에서는 기록만 한다. 임계 판정·알림·모드 강화는 순회가 끝난 뒤 한 번.** 이유(F2): 루프 안의 동기 알림은 경보당 최악 54초
(`obs.DefaultAlertDeliveryBound`, `internal/obs/alert_lease.go:57` "With today's defaults it is 54s", `n.mu` 아래)를 뒤 포지션의 손절 판정 앞에 세우고,
그 지연이 15초 임대를 태워 **B7 을 스스로 만든다.** 같은 파일이 이미 같은 규칙을 둔다 — 격리 행을 맨 뒤로(`exitloop.go:608-610` "every valid
position—including an emergency breach—is recorded/armed/submitted before alert delivery can wait").

- **전송 경로: enqueue-only, 알림기의 단일 입구로(4판 — a092 K6)** — a092 가 정의하는 critical 기록 단일 입구(알림기의 기록 전용 입구가 `n.mu` 아래에서
  `Journal.RecordAlert(ctx, a, remindAfter)`, 기록자별 창, 0 허용 — a092 design D0.3h 4)로 **창 0** 기록한다. 직접 `Journal.EnqueueAlert` 를 부르지 않는다
  (a092 census 핀). 전송은 a098 전달 실행자가, 계속 실패 시 진입 차단은 a124 판정이 한다. 관측 루프는 전송을 기다리지 않는다. a094 D−6.1 과 같은 형태.
  **구현 조건(5판 R3-4 — 하드)**: 구현은 a092 의 `RecordAlert` 입구가 착지한 **뒤**에만 한다. 입구 밖 대안(4판의 "먼저 구현하면 입구 밖 + 먼저-잠금")은
  **삭제**했다 — 가지 않을 경로의 수명주기(잠금 사유·해제 소유·census 편입)를 설계하는 것은 비례 원칙에 어긋난다. 설계 freeze 는 a092 와 독립이다.
- **이벤트**: 기존 `obs.EventExitObservationOutage`(critical, `event.go:332`) — 새 타입 없음.
- **key = 에피소드 신원(3판 N4 · 4판 R2-1)**: `type|positionID|<연속 id>`(계좌 ref 를 빼도 유일하다 — 포지션 id 는 인스턴스마다 유일, `internal/journal/position_adjustments.go:289`). 연속 id 는 연속이 **시작될 때 한 번** `o.opts.NewID()`(128비트 난수, 의도 id 와 같은
  생성기, `exitloop.go:310-312`)로 만든다 — 벽시계 기점을 key 에 쓰면 역행·재시작에서 옛 settled 행과 겹칠 수 있다(codex N4). **유한**(연속당 하나)하고
  **사실에 결속**(그 연속의 기록이 들고 있고, 알림 필드에 기점 벽시계를 싣는다). a094 는 park attempt id 를 에피소드로 쓴다(a094
  D−5.2 — 교차 인용). outbox 의미론 무변경: `EnqueueAlert` 는 재알림 창 0 이라(`outbox.go:142-146`) 같은 key 를 다시 적재하면 옛 행을 재사용하고
  보내지 않는다 — **같은 연속은 한 번**, 새 연속은 새 행(F5 해소). 재시작 뒤 같은 포지션은 새 연속 id 를 받아 **새 에피소드**가 된다(D2 한계와 짝) — a094 의 park attempt(원장 행이라 재시작에도 같은 에피소드)와 **다른 점**이며, 원장에 판정 시각의 믿을 만한 기록이 없어서다(D2).
- ~~**적재 실패(a094 D−5.3 과 같은 형태)**~~ **4판: 입구의 몫으로 대체** — 기록 실패의 진입 잠금은 입구의 생산자 래치(`internal/obs/notifier.go:262-280`)가 한다.
  아래 3판 문장의 "적재 **전에** 세대를 읽는다" 는 engine-safety 정본 위반이었다(`spec.md:1468-1470` — 확정 순간 = 오류가 돌아온 순간; a094 D−6.2 와 같음).
  세대 읽기는 입구(a092)의 몫이며 정본대로 기록 오류가 **돌아온 직후**다(5판 — 입구 밖 형태는 없다). (3판 원문, 기록용:) 적재 전에 `ClearEpoch(ReasonAlertUndelivered)` 를 읽고, `EnqueueAlert` 가 실패하면
  `o.opts.Retrier.Gate.BlockUnlessClearedSince(ReasonAlertUndelivered, epoch, detail)`(`execgw/retry.go:571`; `Retrier.Gate` 필드 `:310` — 새 배선 없음)
  로 진입을 잠그고 내용 없는 로그 한 줄. 관측 루프·청산 무영향. 범위는 **계정 단위**다 — 적재 실패는 원장의 알림 쓰기 자체의 실패라 고장 범위가 계정(저널)이다(a094 Q7-1 Manager 승인과 같은 근거).
- **필드(4판 R2-1)**: `symbol`, `position_id`, `unobserved_seconds`, `cause`. **`account` 는 싣지 않는다** — `AccountRef` 는 실제 계좌번호다
  (`internal/app/engine/interlock.go:680-686` "DisplayName carries accountNo"). 2·3판의 "계좌번호 없음" 은 `account` 필드를 넣으면서 쓴 **거짓**이었다.
  본문이 **포지션을 명명**해 계정 두절 경보와 구별된다(F6). 잔고·가격 없음. D12.
- **연속당 1회**. 판정에 닿으면 연속이 끝난다.

## D5. 모드 강화 — 정본 준수(Q1 확정), 순회 뒤, **공지만 enqueue-only**(3판: N2)

`EscalateOperatingMode(ctx, account, ModeTriggerExitObservationOutage, announcer)`(`:846-847` 과 같은 호출·트리거)를 **순회 뒤**, 경보를 적재한 연속마다.
이미 그 모드면 행도 공지도 없다(`operating_mode.go:409-421`). **모드 커밋은 동기다**(그 호출 안에서 원장에 쓰인다 — 조이기는 늦추지 않는다).

**3판 N2 — 공지자 교체.** 생산 관측자의 `Announcer` 는 Notifier 이고(`cmd/tossctl/engine.go:639`) 그 공지는 `n.Notify` 로 **동기 전송**된다(AnnounceOperatingMode
FLM, 최악 54초). 순회 뒤라도 **다음 주기**의 손절 관측을 늦춘다 — 2판이 스스로 "다음 주기 시작을 늦출 수 있다" 고 적은 바로 그것이다. 그래서 a090 의
강화 호출에는 관측자 소유의 **enqueue-only 공지자**를 넘긴다: `journal.ModeAnnouncer` 를 구현하고 `Journal.EnqueueAlert` 로 적재만 한다. 내용은
`Notifier.AnnounceOperatingMode` 의 `Event` 를 순수 함수 `obs.OperatingModeEvent(previous, rec)` 로 **추출**해 둘이 같이 쓴다(내용을 두 곳에 두지
않는다 — 함수의 분기·key·`Notify` 는 무변화). **key 와 필드가 다르다(4판 R2-1)**: key 는 `operating_mode:<mode>:<rec.ID>` — 전이 id 를 넣고 **계좌 ref 를 뺀다**(전이 id 가 유일). 추출된 `Event` 의
  `FieldAccount` 는 이 공지자에서 **지운다**(Notifier 의 기존 공지는 그대로 — 기존 관행은 D12). Notifier 는 재알림 창으로
같은 key 를 다시 무장하지만(`:58-60` 이 id 를 뺀 이유) `EnqueueAlert` 는 창 0 이라(`outbox.go:142-146`) id 없는 key 는 **두 번째 강화부터 옛 settled 행에
흡수돼 안 나간다**(a094 R6-2 와 같은 기전). 공지는 `changed=true` 일 때만 불리므로(`operating_mode.go:411-416`) 전이당 한 행이다. a094 D−5.2·D−5.3 과
같은 모양(enqueue-only · 에피소드 key · 적재 실패 잠금) — 교차 인용. **5판(R3-2)**: 이 공지자는 **a090 전용 정화 어댑터**이며 a092 입구 **위에서** 기록한다.
  a092 의 기록 전용 announcer 를 **재사용하지 않는다** — a092 는 기존 Notifier 와 같은 이벤트 필드를 요구하고 그 필드에 원문 `FieldAccount` 가 있다(a092
  design D0.3h, codex 3라운드 R3-2). a090 어댑터는 **구성상 무계좌**다: key `operating_mode:<mode>:<전이 id>`, `FieldAccount` 없는 필드, 창 0. a092 는 재개방하지 않는다.

- 계정 두절 경로(`checkOutage` `:846-847`)의 공지는 **바꾸지 않는다**(범위 밖, a092 소관).

**Q1 의 비용 보충(F6).** 운영자가 완화한 뒤에도 같은 연속이 이어지면 다시 강화하지 않는다(D10 의 상태). 그러나 **재시작마다** 새 에피소드라 정지 종목을
보유하면 재시작할 때마다 진입이 다시 막힌다. 모드 이력의 트리거는 계정 두절과 같은 `EXIT_OBSERVATION_OUTAGE` 라 이력만으로는 둘을 못 가른다 —
경보 본문(포지션 명명)으로 가른다.

## D6. fail-closed 가 거부할 정상 입력 — 열거 (2판: F8 · F15)

| 입력 | 오늘 | a090 뒤 | 근거 |
| --- | --- | --- | --- |
| 장 마감 시장의 보유 종목 | **이 사례에서** 관측됨 | 관측되면 변화 없음 · 마감 중 행이 빠지거나 0가격이면 60초 뒤 경보(보수 쪽 거짓 양성 가능) | **생산 증거(한 사례)**: 비거래일(토) `exit_states.last_observed_at = 2026-08-08T01:15:19Z`(a096 `proposal.md:173`) — 마감 중에도 exit 배치가 그 종목을 관측했다. 전 보유·전 시장에 대한 완전성 증거는 아니다. 보조: a112 결정 46(2026-08-28, 단일 종목 · L1c 엄격 리더, 배치 경로 아님). 거래소 어댑터는 `FetchedAt` 을 읽은 시각으로 채운다(`official/market_reads.go:175`). **혼합 시장 배치(한 시장만 마감)는 [미측정]** |
| 거래정지·관리·상장폐지 절차 종목 | [미측정 · 사전 승인된 실측 대기] | 응답에서 빠지거나 0가격이면 60초 뒤 경보(+강화), 연속당 1회 | 코드 주석은 정지 종목이 답하지 않는다고 믿는다(`exitloop.go:869-870`) — 측정 아님. Q2 |
| `Last = 0` 인 보유 종목 | 무음 | 60초 뒤 경보 | [미측정 · 사전 승인된 실측 대기] — Q2 |
| B7 과도 — 앞 포지션의 멈춘 요청(클라이언트 시한 15초 = 임대 15초) | 무음 | 한두 주기는 경보 아님. 60초 연속이면 경보 | `official/client.go:20` · `execgw/retry.go:192` · 시험 :833(16초 제출) |
| exit state 열기가 매 주기 실패하는 보유 포지션(진입 결정에 손절 없음) | 로그뿐, 판정 영영 없음 | 60초 뒤 경보 | `workingSet` B8 · `openState` `:667-671` — 경보가 옳다 |
| 응답 종목 표기 불일치 | 무음(영구) | 60초 뒤 경보 | `:452`·`:760`·`:764` — 경보가 옳다 |

**생산 빈도는 [미측정]이다** — 이 change 가 그 흔적을 처음 만든다(D7).

## D7. 임계 아래의 가시성 (2판: F3 · 3판: N3)

- `ExitCycle.Unobserved int` — 그 주기의 미관측 포지션 수. **3판 N3**: 생산의 `Run` 은 성공 주기의 `ExitCycle` 을 버리므로(`exitloop.go:359` ·
  `reportCycle` `:382-388`) 이 필드는 **시험·tracer 표면**이지 생산 증거가 아니다 — 생산 증거는 아래 로그다.
- **연속 시작**과 **연속 해제** 때 구조화 로그 한 줄씩: `position_id` · `symbol` · `cause` · (해제 시) `unobserved_seconds`. 매 주기 반복하지 않는다.
- **3판 N3 — 로그 이벤트 타입.** 2판의 "기존 `EventExitObservationOutage` 를 로그로 쓰고, 로그는 등급 없음" 은 **거짓**이었다 — 로거는 모든 줄에
  `SeverityOf(type)` 를 싣는다(`internal/obs/log.go:197`) — 그 타입이면 연속 시작 한 줄이 **critical** 표지를 단다. 그래서 새 **normal** 타입
  `obs.EventExitPositionUnobserved = "exit.position_unobserved"` 를 둔다. 등급표(`criticalEvents`)에 넣지 않으므로 normal 이다(`event.go:348` — 없는 타입은
  normal). 이벤트 타입 등록부·골든은 없다(`AllEventTypes` 류 부재 — grep). critical 은 임계를 넘은 **알림**(`EventExitObservationOutage`)뿐이다.
- ~~**3판 N3 — 최소 로거 배선**~~ **4판 R2-1 — 좁은 전용 로거(Manager 승인).** 3판의 관측자 전체 `Log` 배선은 **철회**한다 — 관측자의 기존 `o.log` 줄이
  생산에 나가기 시작하는데(3판이 기록한 부수 효과) 그 줄들이 계좌번호(`exitloop.go:1730` `obs.FieldAccount, o.opts.AccountRef`)·기준선·수량을 싣는다 —
  §8 격리 문제다(codex 2라운드 R2-1). 대신 **이 change 의 새 normal 줄만** 내보내는 전용 로거를 둔다: `ExitObserverOptions.UnobservedLog *obs.Logger`
  (새 필드, `exit_unobserved.go` 만 쓴다), `engineRuntime` 에서 같은 `logger` 로 배선한다. 관측자의 `Log` 는 생산에서 **여전히 nil** — 기존 줄의 생산
  출력은 **무변화**. 새 줄은 `position_id` · `symbol` · `cause` · `unobserved_seconds` 만 싣는다(계좌·가격·수량·오류 문자열 없음). 배선 수준 시험 R17 에
  **카나리 단언**을 더한다(계좌 ref 문자열이 로그 출력·적재된 알림 어디에도 없음).

## D8. 무엇을 편집하는가 (2판)

| 편집 | 성격 |
| --- | --- |
| `workingSet`(기존) | B6 통과 뒤(`:520`) **표시 1개** + B10 진입 첫 문장 **해제 1개**(3판 N1). 다른 탈락 자리는 무편집. 분기 조건·이탈 무변화 |
| `ObserveOnce`(기존) | B6·B7 의 `continue` 직전 원인 기록 1개씩 · 판정 진입(`:464`) 표시 1개 · B3 조기 반환 앞 순회-뒤 처리 호출 1개 · 순회 뒤(`:469` 앞) 처리 호출 1개. **분기 조건·이탈 무변화** |
| `ExitCycle`(구조체) | `Unobserved int` 필드 |
| `ExitObserver`(구조체) | 기록 맵 필드(지연 초기화) |
| 새 파일 `internal/app/engine/exit_unobserved.go` | 표시·해제·기록·순회 뒤 판정·적재(에피소드 key)·적재 실패 잠금·강화(enqueue-only 공지자)·실패 전이(D10)·정리·로그 |
| `Notifier.AnnounceOperatingMode`(기존, `internal/obs/mode.go`) | `Event` 구성을 순수 함수 `OperatingModeEvent` 로 추출(3판 N2) — 동작 무변화 |
| `internal/obs/event.go` | 상수 `EventExitPositionUnobserved`(normal) 추가(3판 N3) |
| `cmd/tossctl/engine.go` `engineRuntime`(기존) | 관측자 옵션에 **`UnobservedLog: logger`** 한 줄(4판 R2-1 — 전체 `Log` 아님) |
| `ExitObserverOptions`(구조체) | 필드 `UnobservedLog` |

새 브로커 호출 0 · 원장 스키마 0 · 토글 0 · 새 reason/trigger 0 · 새 **normal** 이벤트 타입 1(로그용) · 배선 1줄(전용 로거) · 기록은 a092 단일 입구(**구현 하드 의존**) —
a090 은 게이트를 직접 잠그지 않는다(적재 실패 잠금은 입구의 생산자 래치).

## D9. 잔여 — 이름 붙임

- **동기 경보 경로의 기존 알림**(`o.alert` 호출자 7 — `noteDelay` 등)은 그대로다(a092 소관). a090 이 더하는 경보만 enqueue-only.
- **재시작 창**(D2): 재시작 직전의 무관측 시간은 이 기록에서 사라진다.
- **workingSet B6(미관리)·B10(완료 정책)** — 범위 밖, 후속 후보(D1 표). B10 은 표시 해제로 명시 제외.
- **지속 B2**(작업 집합 오류가 계속) — 포지션 단위도 계정 사다리도 재지 않는다(D11-2). 후속 후보.
- **"관측됨" 의 뜻(3판 N10)**: **판정 진입 도달성**이다 — 손절이 평가됐다는 증명이 아니다. 판정에 들어가 즉시 끝나는 경우(격리 → `alertRefused` ·
  재판정 선택자 스탬프 실패 `:876-879` → `cycle.Err` · 정책 신원 오류 → `alertRefused` · 하류 임대 재검사 5자리)도 관측됨으로 친다. 앞 둘은 자기 신호
  (critical 또는 주기 실패 로그)가 있다. tasks 2.16 이 열거를 고정한다.
- **하류 임대 재검사 5자리** — Q3.

## D12. 계좌 정보 — 이 change 의 범위와 기존 관행 (4판: R2-1)

- **이 change**: 새 경보·새 모드 공지·새 로그는 계좌 ref 를 **싣지 않는다**(D4 · D5 · D7). 관측자 전체 로거 배선은 철회했다(D7). 시험이 카나리로 고정한다(R17).
- **공유 실패 로그(5판 D13)**: a090 데이터는 무계좌라 카나리로 고정. 공유 경로 자체의 계좌·오류 노출(`notifier.go:274-280`·`:385-396`)은 아래 기존 관행 항목에 귀속.
- **기존 관행(범위 밖 — 사용자 결정 대기, Manager 가 사용자 큐에 올림)**: 실제 계좌번호가 `obs.FieldAccount` 로 **비시험 코드 19 자리**에 원문으로 실린다(20 자리 중 1 자리만
  `attest.Mask` 로 가린다 — `interlock.go:441`; 계정 두절 경보 `exitloop.go:838` · 모드 공지 `internal/obs/mode.go:67` 등, review 「계좌 정보 사실 고정」).
  그중 critical 경보는 ntfy **외부 전송**을 탄다. obs 에는 가림 함수가 0 이고 저장소 선례는 `attest.Mask` 하나다. 외부 전송 여부와 가림 설계는 사용자 결정이다 — a090 은 그 관행을 **넓히지 않을** 뿐 고치지 않는다.

## D14. a092 재무장 요구의 대상 (a094 7라운드 R7-4 — 정합 해석, Manager 2026-09-29)

a092 델타는 "exit 관측 goroutine 의 기록이 재알림 창이 지난 정착 행을 다시 무장하는 동작은 그대로여야 한다(SHALL)" 고 적고, 같은 문단에서 "이 재무장 요구는 exit
관측 goroutine 의 기록에 대한 것이다 — 다른 기록자의 재알림 창은 그 기록자가 정하며 0(재무장 안 함)일 수 있다" 고 적는다
(`openspec/changes/a092-an-alert-does-not-hold-the-stop/specs/engine-safety/spec.md:44`). a092 design D0.3h 4 의 표·설명은 "a094 · a090 계획의 exit 루프 기록자" 를
**입구 사용**으로, "다른 기록자는 0 을 넘겨 재무장하지 않을 수 있다 — a094 · a090 의 enqueue-only 계약과 양립한다" 로 적는다.

**a090 의 해석(재개방 아님).** a092 재무장 요구의 대상은 **a092 D0.3h 가 정의한 창 기반 exit 기록자**다. a090 의 **에피소드 key · 창 0 기록자**(미관측 알림 ·
강화 공지)는 **대상 밖**이다 — a090 기록자는 exit 관측 goroutine 에서 돌지만, freeze 된 a092 문서 자체의 표가 창 0 양립을 말한다. a092 델타 문언의 명확화는 a092
구현 로트의 erratum 후보(Manager 등록). a094 D−7.3 과 같은 문장이다(교차 인용).

## D13. 공유 실패 경로의 로그 (5판: R3-3)

a090 이 촉발한 기록·승격이 실패하면 그 오류는 a092 입구의 생산자 경로를 지나고, 그 경로의 로그는 원문 오류와 `n.AccountRef` 를 싣는다
(`internal/obs/notifier.go:274-280` 기록 실패 · `:385-396` 승격 로그). 전용 로거(`UnobservedLog`)는 이 줄들을 덮지 못한다.

- **a090 의 몫(정화는 a090 자체에 — Manager)**: a090 이 입구에 넘기는 이벤트·key·필드는 **구성상 무계좌**다(D4 · D5). 그래서 공유 실패 로그를 지나도
  **a090 데이터에서 오는 계좌 정보는 0** 이다 — 실패 주입 카나리로 단언한다(tasks 2.17: 알림 기록 실패 · 모드 기록 실패 · 공지 기록 실패 셋에서 a090 이 넘긴
  값들이 계좌 문자열을 담지 않음).
- **a090 밖**: 공유 경로가 **자기 필드**로 붙이는 `n.AccountRef`(승격 로그 `:387`·`:394`)와 원문 오류 문자열은 기존 이벤트 전부의 문제다 — **기존 20자리 사용자
  큐 항목**(D12)에 귀속한다. 이 두 좌표를 그 항목에 더했다(review 「계좌 정보 사실 고정」).

## D10. 실패 전이 (3판: N5)

연속마다 상태 셋을 **따로** 둔다(한 `alerted` 깃발이 아니다):

| 상태 | 들어가는 조건 | 다음 주기(연속이 이어질 때) |
|---|---|---|
| `enqueued` | `EnqueueAlert` 성공(같은 key 재적재는 옛 행 반환 — 멱등) | 다시 적재하지 않는다 |
| `enqueue_failed` | `EnqueueAlert` 실패 → 진입 잠금(`BlockUnlessClearedSince`, 해제 세대 비교) + 로그 | **재시도**한다(같은 key). 성공하면 `enqueued` |
| `tightened` | `EscalateOperatingMode` 가 오류 없이 반환(`changed` 무관 — 이미 그 모드면 조인 것으로 본다) · **또는** `ErrModeAnnouncementFailed`(전이는 커밋됨, `operating_mode.go:144-148`) | 다시 강화하지 않는다 — 운영자가 그 사이 완화했어도 **같은 연속에서는 재강화 없음**(완화는 사람의 결정이다) |
| `tighten_failed` | 그 밖의 오류(커밋 실패) → 로그 | **재시도**한다 |

- `ErrModeAnnouncementFailed` 는 공지 **적재**가 실패했다는 뜻이다(공지자가 enqueue-only 라서). **4판 R2-2 — 공지 적재는 별도 상태다**: 전이는 커밋됐으므로
  `tightened` 로 가되(재강화 없음), 반환된 `record`(전이 id 포함)를 관측자의 **미적재 공지 대기열**에 둔다. 다음 **처리 주기**마다(순회 뒤 처리가 도는 주기 — B1·B2·B4 주기는 건너뛴다, 5판 R3-5) 같은 전이 id key 로 공지 적재를
  재시도한다 — 전이는 다시 하지 않는다. **연속이 끝나도 대기열은 남는다**(공지는 전이의 사실이지 연속의 사실이 아니다). 성공하면 뺀다. 재시작하면 대기열을
  잃는다 — 모드 행은 원장에 남고 콘솔·모드 이력에 보인다(**이름 붙인 잔여**). 그동안 진입 잠금은 입구의 생산자 래치가 한다(오늘 no-op 강화는 공지하지 않으므로
  — `operating_mode.go:411-416` — 이 대기열이 없으면 공지는 영영 안 나간다, codex 2라운드 R2-2).
- **해제 보호는 시도 단위다(4판 R2-4)**: 적재 실패 → 운영자 해제 → 재시도 → 다시 실패는 **새 증거**로 다시 잠근다(원칙 E 위반 아님 — 해제 뒤의 새 실패).
  운영 모드 완화는 해제 세대와 무관하다(모드는 사람의 완화로만 풀리고, 같은 연속에서 재강화하지 않는 것이 그것을 지킨다).
- 강화는 **알림 적재가 끝난 연속**(`enqueued`)에서만 시도한다 — 알림 없는 조임을 만들지 않는다. 알림 적재가 계속 실패하는 동안에도 진입은 잠겨 있다(위).
- 연속이 끝나면(판정 도달) 상태를 버린다. 새 연속은 새 상태에서 시작한다.

## D11. 보장의 조건 (3판: N7)

"마지막 판정 뒤 60초면 경보" 는 **무조건 상한이 아니다.** 성립 조건:

1. 관측자가 그 60초 동안 **재시작 없이** 살아 있어야 한다(D2 — 재시작하면 기점이 다시 찍힌다). 임계 전에 재시작이 반복되면 탐지가 무기한 미뤄진다.
2. 그 사이에 **순회 뒤 처리가 도는 주기**가 있어야 한다 — 양보(B1) · 작업 집합 오류(B2) · 전 종목 미응답(B4) 주기는 포지션 단위 처리를 하지 않는다.
   B1·B4 는 계정 사다리가 시간을 재고, **지속 B2 는 어느 쪽도 재지 않는다**(`checkOutage` 를 부르지 않는다, `exitloop.go:427-430`) — 이 change 의 범위 밖 구멍으로 명명한다.
3. B4 주기(3판 N6): 보유 대상 표시는 되었으나 가격이 하나도 안 온 주기다. 포지션 단위 **계수·로그는 하지 않고**(계정 사다리가 그 사실의 주인이다, R5 무변화)
   기록도 지우지 않는다 — 기점이 그대로라 다음 처리 주기에 B4 시간이 경과에 든다(R3b).

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
