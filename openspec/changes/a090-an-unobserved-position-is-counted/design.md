# a090 · design

> 분기 주장은 전부 `analysis/function-logic/internal-app-engine--exitobserver.observeonce/`(AST 8 분기, 이 문서보다 먼저)에서 온다.
> 줄 번호는 base(`base-commit.txt`) 기준이며 `exitloop.go` 의 `source_sha256` 은 `522d5d81…`.
> 코드에 없는 값·측정되지 않은 사실은 **[미측정]** 으로 적고 질문으로 남긴다(Q 절).

## D1. 자리 — `ObserveOnce` B6·B7 만

미관측이 일어나는 자리는 정확히 둘이다(AST `branches`): **B6** `:453` `if !ok`(응답에 없음 — `observe` 가 0가격·NaN·Inf·요청 밖·
오래된·미래 시각을 버린 것까지 포함, `:765-775`)와 **B7** `:459` `if !o.quoteUsable(quote)`(사용 임대 만료, a111). 둘 다 무음 `continue`.

- 이 둘이 **보유 포지션을 주기마다 정확히 한 번 보는 유일한 자리**다(B5 range). a089 3라운드 ⑤ 와 같은 결론 — 하류를 개별 계측하지
  않는다.
- B4(전 종목 미응답)는 계정 사다리가 이미 본다 — **무변화**(R5). B1(양보)도 무변화(R6) — 양보 주기는 `states` 를 읽지 않으므로
  포지션 단위 기록을 만들 수 없고, 만들 필요도 없다: 계정 시계가 계속 돈다(`:418-423`).
- **하류의 임대 재검사 다섯 자리도 무음이다** — `quoteUsable` 의 비시험 호출은 B7 `:459` 외에 `judge` `:859` · `judgeRatchet` `:956` ·
  `judgeLadder` `:1027` · `refreshObservation` `:1050` · `record` `:1180` 다섯이고(grep `-F 'quoteUsable('`, CodeGraph 1.6.0
  `callers ExitObserver.quoteUsable` 6 과 일치), 전부 `return nil` 이다. a111 이 기록 경계에서 임대를 다시 보게 한 것이다
  (`TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary`). a090 은 **B8 직전(`:464`, 판정 진입)을 "관측됨" 으로 본다** — 그래서 판정에
  들어간 뒤 **그 포지션 자신의 판정 처리**가 15초 임대를 넘겨 기록 없이 끝나는 경우는 세지 않는다. 그 경우가 **매 주기** 반복돼야
  무기한 무음이 된다(한 포지션의 원장 판정이 매번 15초 이상). **극단 edge 로 `unsupported` 로 닫는다**(비례 원칙) — 판정 함수들은
  편집하지 않는다. 관측점을 기록 경계로 옮기려면 다섯 함수의 FLM 과 a094 가 편집하는 `record` 와의 충돌이 따른다 — Q3 확정: 명명된 잔여·후속 후보.

## D2. 상태 — 관측자 필드, 포지션 id 단위

`unobservedSince map[string]time.Time`(포지션 id → 연속 미관측이 시작된 시각)과 `unobservedAlerted map[string]bool`(그 연속에 대해 알림을
냈는가). 선례는 같은 파일의 `delayedSince`·`delayAlerted`(`exitloop.go:251-254`, 청산 지연 경보)와 같은 짝이다.

- **지연 초기화** — `quarantineAnnounced`(`:245-250`, "initialised lazily on first use; a nil map already reads as nothing announced")와
  같이. 그래서 **`NewExitObserver` 는 편집하지 않는다**(기존 함수 편집은 `ObserveOnce` 하나).
- **포지션 id 단위**인 이유: `states` 가 포지션 단위이고(`managed.position.ID`), 같은 종목의 인스턴스가 둘일 수 있다. 지연 경보 key 도
  포지션 단위다(`type|positionID`, `exitloop.go:1688`).
- **재시작하면 잃는다.** 파일 머리 계약(`:67-68`): "the only fields on the observer are alert latches and timers, whose loss on
  restart re-raises the alert rather than losing it." 재시작 뒤 같은 포지션이 계속 미관측이면 임계를 다시 채우고 다시 알린다.
  원장 스키마 변경 없음.

## D3. 임계 — 계정 사다리와 같은 시간 임계, 주기 수가 아니다

**미관측이 시작된 뒤 `o.outageAfter()`(기본 `DefaultExitObservationOutage` = 60초, `:105`; `ExitObserverOptions.OutageAfter` 로 설정)
이상 지나고도 그 포지션이 판정에 닿지 않았으면** 판정한다.

영수증:

- 정본 exit-policy 「관측 경로와 fail-safe」: "관측 두절이 staleness 임계(기본 60초)를 넘으면 critical 알림 + ENTRY_BLOCKED 자동
  강화가 발동한다(SHALL — "보류"가 무기한 무손절이 되어서는 안 된다 …)". 시나리오 「관측 장기 두절」: "**보유 포지션의** 가격 관측이
  60초 이상 실패하면". 문장의 주어는 포지션이다 — 코드가 계정 단위로만 구현했다.
- `DefaultExitObservationOutage` 의 주석(`:103-104`): "four intervals, which is what makes a single failed read — or a single deferred
  cycle — not an incident while a persistent one is."

**`obs.DefaultCriticalAttempts`(=3, `internal/obs/notifier.go:45`)를 쓰지 않는 이유.** 그 상수는 **전달 재시도** 횟수다
(`notifier.go:39-44` — "the failures worth retrying through are transient (a DNS blip …)"). 관측과 다른 영역이고, 주기 수로 세면
① 5초 주기에서 15초 — 사용 임대(15초)와 같은 크기라 아래 D6 의 정상 과도(B7 한두 번)에 걸린다 ② 양보 주기(B1)는 `states` 를 읽지
않으므로 **세어지지 않아** 양보가 길수록 늦게 울린다 — 계정 시계는 양보 중에도 흐르는데(`:418-420`) 포지션 계수만 멈추는 비대칭이
생긴다. 시간 임계는 둘 다 없다. (a094 D−3.8 이 같은 상수 차용을 "관례 일관성이지 증거가 아니다" 로 기록한 것과 같은 판단.)

## D4. 알림 — 기존 critical 이벤트를 포지션 key 로

`o.alert(ctx, obs.Event{Type: obs.EventExitObservationOutage, Key: "<type>|<account>|<positionID>", …})`.

- **새 이벤트 타입을 만들지 않는다.** `EventExitObservationOutage` 는 이미 critical 등급표에 있고(`internal/obs/event.go:332`), 뜻이
  정확히 같다(`event.go:175-178` — "with no broker-resident stop, an unobserved position is an unprotected one"). 등급표·
  engine-safety 의 critical 목록 시험(`TestGradingMatchesTheSpec`)은 무변화.
- **key 는 계정 key(`type|account`, `exitloop.go:832`)와 겹치지 않는다** — 세그먼트가 하나 더 있다. 그래서 a096 재알림 창·outbox
  행이 포지션마다 따로다. 이 이벤트의 소비자는 생산 코드에 `exitloop.go:831-832` 하나뿐이다(grep — 키 접두 파싱 없음).
- **연속 하나에 1회**(`unobservedAlerted`). 판정에 닿으면 지워지고, 다시 빠지면 새 연속이다.
- 필드: `account`, `symbol`, `position_id`, `unobserved_seconds`, `cause` = `no_quote`(B6) | `quote_expired`(B7). 계좌 번호·잔고·
  가격은 싣지 않는다(§8).
- 전송은 오늘의 `ExitAlerter.Notify` 경로 그대로 — **새 전송 경로 없음**.

## D5. 모드 강화 — 정본 준수(Q1 확정)

정본 문장은 두절에 **critical + ENTRY_BLOCKED** 를 요구한다(D3 인용). 이 change 의 delta 는 그 문장을 **그대로 포지션 단위에 적용**해
`EscalateOperatingMode(ctx, account, journal.ModeTriggerExitObservationOutage, announcer)`(`:846-847` 과 같은 호출·같은 트리거,
`internal/journal/operating_mode.go:84`)를 연속당 1회 부르도록 쓴다. 자동 경로는 조이기만 하고(`EscalateOperatingMode` 는 이미 그 모드면
no-op), 완화는 사람이다.

이 강화가 거부할 정상 입력은 D6 에 열거했다. Manager 가 정본 준수로 확정했다(Q1 기록).

## D6. fail-closed 가 거부할 정상 입력 — 열거

포지션 단위 판정이 경보(+Q1 에 따라 진입 차단)로 바꿀, **위험이 아닐 수 있는** 입력:

| 입력 | 오늘 | a090 뒤 | 근거 |
| --- | --- | --- | --- |
| 장 마감 시장의 보유 종목(주간의 US, 야간의 KR) | 관측됨 | **변화 없음** | a112 결정 46 실측(2026-08-28 05:29-05:31 KST): 두 시장 마감 중 `/prices` 가 HTTP 200·행 1개. 거래소 어댑터는 `FetchedAt` 을 읽은 시각으로 채운다(`internal/official/market_reads.go:175`) → `observe` 의 나이 검사에 안 걸린다. **n=1/시장** |
| 거래정지·관리·상장폐지 절차 종목 | **[미측정 · 사전 승인된 실측 대기]** | 응답에서 빠지거나 0가격이면 60초 뒤 경보(+Q1 강화) — 정지가 며칠이면 그 연속 내내 1회 | 코드 주석은 정지 종목이 답하지 않는다고 **믿는다**(`exitloop.go:869-870` "a halted symbol, a suspended one") — 측정 아님. Q2 |
| `Last = 0` 인 보유 종목(첫 체결 전 신규상장 등) | 무음 | 60초 뒤 경보 | [미측정 · 사전 승인된 실측 대기] — Q2 |
| B7 과도 — 앞 포지션 처리가 사용 임대(15초)를 넘김 | 무음 | 한 번은 경보 아님. **60초 연속**이면 경보 | 시험 :833 이 16초 제출로 재현. 60초 연속이면 매 주기 앞 포지션이 15초 넘게 걸린다는 뜻 — 그 자체가 뒤 포지션의 무보호다 |
| 응답 종목 표기가 보유 종목 표기와 다른 경우 | 무음(영구) | 60초 뒤 경보 | 조회는 `ToUpper(TrimSpace)` 일치(`:452`·`:760`·`:764`). 불일치가 실재하면 오늘 그 포지션은 **한 번도** 판정되지 않고 있다 — 경보가 옳다 |

**생산 빈도는 [미측정]이다.** B6·B7 은 흔적을 남기지 않으므로 원장으로 셀 수 없다(이 change 가 그 흔적을 처음 만든다).

## D7. 전달 비용 — 이름 붙인 잔여(a094 D−3.7 과 같은 사실)

새 critical 은 `ExitAlerter.Notify` 로 간다. 오늘 그 경로는 관측 루프 안에서 동기로 불리고(`exitloop.go:1710`), 전달은 `n.mu` 아래
동기 발행·재시도다(`internal/obs/notifier.go:254-255`·`:309`). 그래서 전달이 막힌 동안 **같은 주기 다른 포지션의 판정이 늦어질 수
있다.** a090 의 경보는 연속당 1회라 빈도는 유계지만 지연 자체는 없어지지 않는다. a094 5라운드 R5-4 에 대해 Manager 가 "enqueue-only
요구" 로 처분했다 — **a090 도 같은 요구를 따른다**(구현 시 그 형태가 정해지면 같은 호출을 쓴다). 동기 전달 제거 자체는 a092 소관.

## D8. 무엇을 편집하는가

| 편집 | 성격 |
| --- | --- |
| `ObserveOnce`(기존) | B6·B7 의 `continue` 직전 호출 1개씩, `cycle.Judged++`(`:464`) 직전 해제 1개, 순회 뒤 정리 1개. **분기 조건·이탈 무변화** |
| 새 파일 `internal/app/engine/exit_unobserved.go` | 기록·판정·알림·(Q1)강화·정리 메서드와 두 필드의 지연 초기화 |
| `ExitObserver` 구조체 | 필드 2개 추가(선언만) |

새 브로커 호출 0 · 원장 스키마 0 · 토글 0 · 새 이벤트 타입 0 · 새 reason/trigger 0.

## Q — 결정 기록 (Manager, 2026-09-29 — 셋 다 사용자행 아님)

- **Q1 = (a) 확정 — 포지션 단위 두절에 critical + ENTRY_BLOCKED.** 근거: 정본이 이미 요구한다 — `openspec/specs/exit-policy/spec.md:62`
  「관측 경로와 fail-safe」 "관측 두절이 staleness 임계(기본 60초)를 넘으면 critical 알림 + ENTRY_BLOCKED 자동 강화가 발동한다(SHALL …)" 와
  `:65` 시나리오 "**보유 포지션의** 가격 관측이 60초 이상 실패하면". **정본 준수이지 신규 정책이 아니다.** 대가(D6: 정지 종목 하나가 사람이 완화할
  때까지 신규 진입을 막음)는 진입이지 청산이 아니며, 자동 경로는 조이기만 한다는 승인 원칙에 맞다. spec delta 는 그대로(MODIFIED 불요).
- **Q2 = 구현 로트로 이연 + 사전 승인.** 정지·0가격 종목의 `/prices` 응답은 **[미측정 · 사전 승인된 실측 대기]** 다. 승인 범위: 읽기 전용 시세
  GET 1회(정지 종목 포함, 쓰기 0), 구현 로트가 **장중에** 1회 실행하고 그 결과로 D6 의 두 [미측정] 행을 확정한다. 측정은 경보 **빈도**를 알려 줄 뿐
  판정 규칙은 같다 — 결과가 무엇이든 구현이 멈추지 않는다.
- **Q3 = 초안 그대로 — 관측점은 판정 진입(`:464`).** 근거: 헌장(`ObserveOnce` B5/B6 최소 편집)과 일치하고 a094 의 `record` 편집과 겹치지 않는다.
  **하류 무음 5자리는 명명된 잔여**다 — `judge` `:859` · `judgeRatchet` `:956` · `judgeLadder` `:1027` · `refreshObservation` `:1050` ·
  `record` `:1180`(전부 `!o.quoteUsable(quote)` → `return nil`). 판정에 들어간 포지션 **자신의** 처리가 매 주기 15초 임대를 넘기는 경우만 여기서
  무음이 된다. **후속 change 후보**(관측점을 기록 경계로 옮기거나 다섯 자리에 같은 기록 호출) — 이 change 는 D1 의 `unsupported` 로 닫는다.
