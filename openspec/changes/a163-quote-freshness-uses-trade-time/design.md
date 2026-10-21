# a163 design — 체결 시각은 보고의 입력이지 판정의 입력이 아니다

근거 문서: `analysis/quote-time-evidence.md`(소비자 C1~C16 · 계약 · 긴장 §4). 분기 좌표는 전부
`analysis/ast/enumerate-output.md`(AST 22 · HEAD `f7d80c34`; freeze 리뷰 반영 로트에서 AST 29 · STALE 0 으로 재생성 — `internal` 은
base 와 diff 0)에서 온다.

## D0. 방향 — 사용자 결정 2026-10-21: A

Manager 경유 사용자 회신 「A」(2026-10-21): **브로커 체결·시세 시각을 받은 시각 옆에 기록하고, 오래된 체결은 보고·알림만 한다.
청산 루프는 체결 나이 때문에 손절·비상 청산을 건너뛰거나 미루지 않는다(안전 4).**

## D1. 고려한 선택지

| 안 | 내용 | 처분 |
| --- | --- | --- |
| **A** | 브로커 시각을 보존하고 체결 나이를 기록·알림. 판정 자격은 받은 시각 기준 그대로 | **채택**(D0) |
| B | 진입·전략 판정만 체결 나이로 거르고 청산은 받은 시각 유지 | 기각 — 사용자 결정 A. 진입 쪽 거름은 임계 실측(H6) 없이 정하면 정상 진입을 조용히 죽인다(fail-closed 가 거부할 정상 입력 미열거) |
| C | 청산이 오래된 체결 관측을 「주의」로 표시하되 시장가 매도는 허용 | 기각 — 사용자 결정 A. 「주의」 표시가 판정 경로 안에 들어가면 손절 경로에 새 분기가 생기고, 그 분기가 나중에 거름으로 자라는 문을 연다 |

## D2. 왜 `FetchedAt` 의 의미를 바꾸지 않나 (새 필드를 더하는 이유)

`FetchedAt` 은 판정 자격 외에도 세 곳에서 **값 그대로** 쓰인다(C6):

1. `observe` B7 `:871-872` 미래·15 s 규칙과 `quoteUsable` B1 `:1184` 미래 규칙 — 의미를 바꾸면 곧 B 안이 된다.
2. `stableObservationID` `:1208-1209` — 관측 **identity** 의 stamp. 체결 시각으로 바꾸면 같은 체결가가 이어지는 주기들이 같은 id 를 갖고,
   exit-policy 「같은 observation이 재전달된다 → 성공 no-op」에 흡수되어 snapshot freshness 가 멈춘다.
3. `record` B3 `:1313` · `refreshObservation` — 원장 `ObservedAt`(exit-line freshness 의 정본). 바꾸면 화면의 「신선함」이
   체결 간격에 묶여 장외·저유동에서 line 이 stale 로 떨어진다.

셋 다 판정·원장 동작을 바꾸므로 A 와 양립하지 않는다. 따라서 **새 필드**를 더하고 `FetchedAt` 은 건드리지 않는다.
이름에 「trade」를 넣지 않는 이유: 계약은 그 값을 「데이터 시각」이라 부를 뿐 마지막 체결 시각이라 명시하지 않는다
(`analysis/quote-time-evidence.md` §2.1).

**필드 형태 — freeze 리뷰 P1-1 반영(2026-10-21): `domain.Quote.SourceTimestamp string` · json `source_timestamp,omitempty`.**
브로커 원문 문자열을 그대로 보존하고, 해독은 청산 루프의 보고 경로(`exit_source_age.go`)에서만 한다.

- 첫 가안(`SourceAt time.Time` · `source_at,omitempty`)은 결함이었다: `omitempty` 는 struct 인 `time.Time` 의 0 값을 생략하지 않는다.
  실측(go1.26.5 스크래치, 0 값 struct 를 `json.Marshal`): `time.Time`+`omitempty` → `{"a":"0001-01-01T00:00:00Z"}` ·
  `time.Time`+`omitzero` → `{}` · `string`+`omitempty` → `{}`. `tossctl quote --json`(`cmd/tossctl/quote.go` → `hybrid.Client.GetQuote` `internal/hybrid/client.go:117-130`
  → `output.WriteQuote` → `writeJSON(quote)` `internal/output/quote.go:27-28`)의 WTS 갈래 생산자 `internal/client/quote.go:115`
  (`quote := domain.Quote{…}`)와 `/stocks` 어댑터 `adaptStocks`(`internal/official/stock_reads.go:81-97`)는 이 필드를 채우지 않으므로
  첫 가안대로면 출력에 `"source_at":"0001-01-01T00:00:00Z"` 가 생겨 WTS(upstream) 출력 계약이 바뀐다(안전 3).
- 두 수리 후보: (a) `time.Time`+`omitzero`(선례 `internal/verifylive/record.go:254` 「영이면 생략(omitzero)해 다른 줄의 바이트를 바꾸지
  않는다」) · (b) 원문 문자열(선례 같은 어댑터 파일의 `adaptTrades` `internal/official/market_reads.go:311` `Time: t.Timestamp`,
  `adaptMarketIndicatorPrices` `:436` `Timestamp: p.Timestamp`, 도메인 `internal/domain/models.go:721`
  `Timestamp string \`json:"timestamp,omitempty"\``).
- **(b) 를 고른 이유**: ① 어댑터에 새 분기가 0 이다 — `adaptPrices` 의 변경은 대입 한 줄이고, 어댑터가 해독 실패로 행을 버리거나 오류를
  내는 경로(손절 관측 실패로 번지는 경로, 아래 D3)가 **구조적으로** 생기지 않는다. (a) 는 High-risk 어댑터 안에 해독 분기를 하나 더한다.
  ② 해독이 보고 경로의 recover 경계(D4) 안에 놓인다. ③ null(빈 문자열)과 해독 실패(비어 있지 않은 원문)를 보고 쪽에서 가를 수 있다(P3-3).
  ④ 같은 파일의 공식 어댑터 둘이 이미 이 형태다. `string`+`omitempty` 는 빈 문자열을 생략하므로 필드를 채우지 않는 생산자의 JSON 바이트가
  편집 전과 같다 — tasks 2.0 이 WTS·`adaptStocks` 두 생산자로 이를 고정한다.
- H7 과의 관계: H7 의 실질(JSON 에 노출 · 값이 없으면 생략 · 호가 어댑터 제외)은 그대로다. 바뀐 것은 D2 가안의 이름·타입뿐이다
  (아래 「사람 결정 기록」 뒤 주석).

## D3. 어댑터 규칙

- `adaptPrices` 는 `SourceTimestamp: p.Timestamp` **한 줄**만 더한다(분기 0 추가 — 편집 전 AST 분기 1 · 반환 1 이 편집 뒤에도 같아야 함,
  `analysis/ast/official--adaptprices.ast.json`). 해독하지 않는다. 실측 형식 `2026-08-28T05:28:30.000+09:00`(두 시장 모두 `+09:00`)은
  보고 경로가 `time.Parse(time.RFC3339Nano, …)` 로 읽는다.
- null·키 없음은 `apiPrice.Timestamp string`(`market_reads.go:137`)의 디코드 규칙상 둘 다 `""` 다(구별 불가 — 계약상 둘 다 「모름」).
  어댑터는 `""`·깨진 문자열 어느 쪽도 **거르지도, 오류를 내지도 않는다** — 이 값은 보고 입력이고, 어댑터가 실패하면 그 실패가 청산 관측
  실패(`observe` B1)로 번져 손절을 미룬다. 행 탈락(`continue`)·오류 반환 변이는 tasks 4.1 ⑦⑧ 이 e2e(2.1~2.3)로 잡는다.
- 엄격 리더의 형식 검사(`strictMinuteInstant`)를 재사용하지 않는다: 그 리더는 거절이 목적이고 이 어댑터는 거절하면 안 된다.
- `adaptOrderbook` 은 H7 결과에 따른다(기본: 손대지 않음).
- 기존 위험(이 change 가 키우지 않음, freeze 리뷰 P3-7): `apiPrice.Timestamp` 는 **이미** `string` 으로 디코드된다(`market_reads.go:137`).
  브로커가 이 필드 타입을 숫자 epoch 등으로 바꾸면 오늘도 `Prices` 전체가 디코드 오류 → `observe` B1 → 그 주기 판정 0 이다. a163 은 이
  표면을 넓히지 않는다(디코드 타입 불변). tasks 5.1 잔여에 적는다.

## D4. 청산 루프 규칙 — 판정 경로에 새 분기 0, 판정 순회 안 새 코드 0

> 사용자 결정 2026-10-22: 이 절의 보고(구간 분류 · 래치 · `SourceAgeLog` 줄 · `source_missing` 알림)는 **기본 OFF 인 보고 토글이 켜졌을 때만**
> 돈다(D8). 「운반」과 「생산 싱크」는 토글 밖이다.

freeze 리뷰(2026-10-21) P1-3·P1-4·P2-3·P2-4·P2-5·P3-3·P3-4 반영판. 분기·defer·recover 주장은 전부 `analysis/ast/enumerate-output.md`
(이 로트에서 `Runtime.Run` · `ExitObserver.Run` · `writeExitOrder` · `snapshotContext` · `requiresFullJudgement` · `noteHeldProposal` ·
`raiseUnobserved` AST 일곱을 더해 재생성, AST 29 · STALE 0)에서 온다.

- **운반**: `observedQuote`(`exitloop.go:829-834`)에 `SourceTimestamp string` 을 더하고 `observe` 의 `next[symbol] = observedQuote{…}`
  한 자리(`:875`)에서 `q.SourceTimestamp` 를 싣는다. `observe` 의 분기 15·반환 6 은 **그대로**다 — 후보 제외(B4·B7)는 여전히
  `FetchedAt` 기준. 판정 함수 고정 수단은 D7.
- **자리**: `ObserveOnce` 의 판정 순회(`:501-523`)가 끝난 뒤, `o.settleUnobserved(ctx, &cycle)`(`:524`) **바로 앞**에 보고 호출 한 줄.
  판정 순회 **안에는 새 코드를 두지 않는다**(첫 가안의 「순회 안에서 `noteJudged` 처럼 적기만」 폐기 — 순회 안 코드는 아래 recover 경계
  밖이다). 어느 포지션이 판정됐는지는 a090 주기 집계(`o.cycleUnobserved`, `owner == &cycle` 일 때)의 `judged`·`held` 를 **읽기만** 해서
  안다. 그 집계를 비우는 것은 `settleUnobserved`(`exit_unobserved.go:142-143`)이므로 보고는 그 앞이어야 한다. 둘 다 판정 뒤라는 D4 의
  첫 조건은 그대로다. `settleUnobserved` 의 임계 알림은 `o.opts.Critical`(outbox, `exit_unobserved.go:211-215`)로 가므로 보고의 일반
  등급 relay 와 버퍼를 나누지 않는다. 무보유 경로(`:481`)와 조기 반환 경로에서는 보고를 부르지 않는다(그 주기에는 판정이 없다).
- **패닉 격리(P1-4, 안전 4)**: 감독자는 패닉을 거두지 않는다 — `Runtime.Run` 의 루프 goroutine(`runtime.go:306-309`)의 defer 는
  `wg.Done` 하나이고 `Run` 전체에 `recover` 호출 0, `ExitObserver.Run`(`exitloop.go:395-405`) 은 defer 0 · `recover` 0
  (enumerate-output.md 의 두 절). 그래서 보고 코드의 패닉은 곧 프로세스 종료이고, protection UNWIRED 인 지금 엔진 정지 = 손절 없음이다.
  보고 진입 함수(`exit_source_age.go`)는 첫 줄에서 a129 `writeExitOrder`(`exit_order_log.go:264`, AST `RECOVER` 행)와 같은
  `defer func() { if recover() != nil { … } }()` 경계를 세우고, 해독·래치 맵·로그·알림을 **전부 그 함수의 호출 트리 안에서** 한다
  (Go 의 recover 는 패닉한 프레임 위의 defer 에서만 받는다 — `exit_order_log.go:258` 주석). 거둔 패닉은 버리고 계좌 없는 고정 문구 한
  줄을 `SourceAgeLog` 에 남긴다. 그 주기의 판정·제출은 이미 끝났고 `settleUnobserved` 는 그대로 돈다.
- **단일 goroutine 소유**: 생산에서 `ObserveOnce` 를 부르는 곳은 `Run`(`exitloop.go:400`) 하나다. `tracer.go:297` 도 부르지만
  `NewTracer`(`tracer.go:239`)의 비시험 호출자는 0 이다(2026-10-21 grep 전수). 래치 맵 `o.sourceAge` 는 `o.unobserved`(`exitloop.go:285`)
  와 같이 그 goroutine 만 읽고 쓰며 잠금을 두지 않는다. **concurrent map write 는 recover 로 거둘 수 없는 fatal** 이므로 이 소유 전제가
  패닉 격리의 일부다 — tasks 2.10 의 시험을 `Makefile` `RACE_ENGINE_TESTS`·`RACE_ENGINE_FILES` 에 넣는다.
- **분류(로그 구간)** — 판정된 포지션마다 하나:
  `missing`(`SourceTimestamp == ""` — null·키 없음·빈 문자열) · `unparseable`(비어 있지 않으나 `time.RFC3339Nano` 해독 실패) ·
  `ahead`(나이 < 0) · `le_15s`(0 ≤ 나이 ≤ `execgw.QueryPriceEvidenceDuration`, `internal/execgw/retry.go:192` 기존 상수) ·
  `le_5m` · `le_1h` · `gt_1h`. `missing` 과 `unparseable` 을 가르는 이유는 P3-3: 계약 형식 변화(`+0900`·공백 구분 등)를 null 과
  구별하려고. 구간 경계는 **보고 해상도**일 뿐 판정·알림에 쓰지 않는다(H1 후속 결정에서 바꿀 수 있다).
- **나이 = `observedQuote.FetchedAt − 해독한 브로커 시각`**(P3-4). `observedQuote.FetchedAt` 은 `observe` 가 넣은 `effectiveAt` 이다 —
  `domain.Quote.FetchedAt` 이 0 이면 `validatedAt` 으로 대체된 값(`exitloop.go:867-870`). `domain.Quote.FetchedAt` 을 직접 쓰지 않는다
  (0 이면 거대 음수가 `ahead` 로 잘못 분류된다).
- **로그 빈도(P2-3)**: 포지션의 구간이 **바뀔 때만**(처음 판정된 주기 포함) 한 줄. 같은 구간이 이어지면 0 줄 — 같은 파일의 정책
  (`exitloop.go:422-423` 「One line every five seconds forever is not observability.」)과 a090 의 전이 줄(`exit_unobserved.go:172-179`
  해제 · `:189-193` 연속 시작)과 같은 모양. H6 (b) 분포는 전이 줄의 시각 차(구간 체류 시간)와 전이 순간의 나이로 모은다 — 구간 **내부**의
  값 분포는 얻지 못한다(한계, H1 후속 결정 때 다시 본다).
- **생산 싱크(P1-3)**: 새 옵션 `ExitObserverOptions.SourceAgeLog *obs.Logger`(a090 `UnobservedLog` `exitloop.go:190-191` · a091
  `ZeroFloorLog` · a129 `ExitOrderLog` 선례). `opts.Log` 는 생산에서 nil 이므로(`exitloop.go:188-191` 주석, `cmd/tossctl/engine.go:671-681`
  은 `UnobservedLog`·`ExitOrderLog` 만 넣음, `o.log` `exitloop.go:1932-1941` 은 nil 이면 무동작) 쓰지 않는다. 알림 경로의 로그 줄도 대신하지
  못한다 — 생산 알림기 `obs.RecordOnly`(`exitwiring.go:333·349`)의 `Notify` 는 `withoutFields`(`internal/obs/record_only.go:51`·
  `:106-108`)로 **필드를 지운 뒤** 로그한다. 배선: `cmd/tossctl/engine.go` `engineRuntime`(`:655`)의 `engine.ExitObserverOptions` 리터럴에
  `SourceAgeLog: logger`. nil 이면 줄 0(단위 시험용 선택) — 그래서 생산 배선 시험(tasks 2.11)이 필요하다.
- **알림(H4)**: 사유 `source_missing` 하나 — 구간이 `missing` 또는 `unparseable` 인 포지션. 일반 등급(새 `EventType` 을
  `criticalEvents` `internal/obs/event.go:357-392` 에 넣지 않음 → `SeverityOf` normal). `o.opts.Alerts` 로 보내고, 생산에서는
  `RecordOnly` → `NormalRelay.Offer`(비차단)다.
- **래치와 정리(P2-4)** — 포지션 id → {구간, 알림함}:
  - 이번 주기에 판정된 포지션: 구간을 갱신한다. 구간이 `missing`/`unparseable` 이고 알림함이 거짓이면 알리고 참으로. 구간이 그 밖으로
    바뀌면 알림함을 거짓으로(재무장).
  - 이번 주기 `held` 에는 있으나 판정되지 않은 포지션(`no_quote`·`quote_expired`·발의 보류 등): 기록을 **바꾸지 않는다**.
  - 이번 주기 `held` 에 없는 id: 지운다 — a090 `settleUnobserved` 의 정리 기준(`exit_unobserved.go:152-157`)과 같다. 집계가 없거나
    `held` 가 비면 맵 전체를 비운다.
  - 재시작하면 잃는다(a090 미관측 기록과 같은 수명) — 재시작 뒤 다시 알린다.
- **주기 상한(P2-5)**: 한 주기의 `source_missing` 알림은 최대 8건(= `obs.DefaultNormalRelayCapacity` 64 의 1/8, `normal_relay.go:18`).
  넘는 포지션은 알림함을 세우지 않으므로 다음 주기에 넘어간다(id 정렬 순서 — 결정적). 로그 줄은 전이 기준이라 따로 유계다. 재시작 직후
  보유 전 종목이 한 주기에 `missing` 이어도 버퍼의 7/8 은 다른 exit 일반 알림(capped·unmanaged — `normal_relay.go:15-17` 주석)에 남는다.
  버퍼가 가득 차면 버림이 이 등급의 정의다(`normal_relay.go:55-56`) — `source_missing` 도 버려질 수 있고, 그때도 `SourceAgeLog` 줄은 남는다.
- **필드**: `position_id` · `market`(`journal.Position.Market`) · `symbol` · `bucket` · `prev_bucket` · `age_seconds`(해독된 경우만) ·
  `source_timestamp`(`unparseable` 일 때만, 원문 앞 64 바이트). 계좌 ref 를 싣지 않는다(안전 8).

## D5. 진입·전략·flatten 경로 (C7~C10)

이 change 에서 **동작 변화 0**. C8·C9(t0 기준가에 나이 검사 0)와 C10(비상 청산 지정가에 나이 검사 0)은 a163 이 측정으로 찾은 잔여이고,
고치려면 각각 별도 Story 다(H3). 특히 C10 은 flatten-all 즉시성(안전 4)에 직접 닿으므로 「보고만」 외의 변경은 이 change 에서 금지.

## D6. 실패 모양과 그것이 거부하는 정상 입력

이 change 는 fail-closed 가드를 **하나도 더하지 않는다**. 더하는 것은 보고뿐이므로 거부되는 정상 입력은 0 이어야 한다 — 그것을 시험으로
센다: 「모름」 시각 · 9h30m 묵은 시각 · 미래 시각 · 정상 시각 네 입력 모두에서 `cycle.Judged` 와 제출 수가 편집 전과 같다.
보고 경로 자체의 실패(패닉 포함)도 같은 규칙이다 — 보고 진입 함수의 recover 경계가 거두고, 그 주기의 판정·제출과 다음 주기의 제출 수는
보고가 없을 때와 같다(tasks 2.10).

## D7. 「판정 경로 불변」 의 강제 수단 (freeze 리뷰 P1-2)

첫 가안(「편집 뒤 `source_sha256` 이 base 와 같아야 함」)은 성립하지 않는다: logic-map 의 `source_sha256` 은 **파일 전체** 해시다
(`exitloop.go` 번들 전부 `ba305154…`; 이 change 는 같은 파일의 `observedQuote`·`observe`·`ObserveOnce` 를 편집하므로 반드시 바뀐다).
그리고 일곱 함수 본문이 그대로여도 그들이 부르는 헬퍼가 브로커 시각을 흘리면 판정이 바뀐다. 그래서 셋으로 고정한다.

1. **함수별 렌더 해시 시험**(커밋된 Go 시험): `internal/app/engine` 의 비시험 파일을 `go/parser`(주석 미적재)로 읽어 수신자 `ExitObserver` 의
   `quoteUsable` · `judge` · `judgeRatchet` · `judgeLadder` · `record` · `refreshObservation` · `stableObservationID` · `snapshotContext` ·
   `requiresFullJudgement` · `noteHeldProposal` 열 개를 `go/printer` 로 렌더한 sha256 을 시험 안에 굳힌 기대값과 비교한다. 기대값의 원천:
   `analysis/render-hash/render-hash-base.txt`(HEAD `60033240a` — `internal` 은 base `f7d80c348` 과 diff 0, 생성기
   `analysis/render-hash/render_hash.go.txt`). **이름 충돌**: `judge`(`alertdelivery.go`) · `record`(`strategy_lane_runtime.go`)가 다른
   수신자로 있으므로 이름만으로 고르면 안 된다. 주석 미적재 렌더라 주석 편집에는 둔감하고 코드 편집에는 민감하다. 선례: go/printer 를 쓰는
   구조 시험 `internal/app/engine/a112_shadow_load_pin_test.go:45`.
2. **금지 census**(커밋된 Go 시험): `internal/app/engine`·`internal/official` 비시험 파일의 `SourceTimestamp` 식별자 **모든 출현**(selector ·
   복합 리터럴 키 · 필드 선언 · method value)을 세어, 허용 자리와 자리별 정확한 수 = {`observedQuote` 필드 선언 1 · `observe` 의
   `observedQuote{…}` 리터럴 2(키 + `q.SourceTimestamp`) · `adaptPrices` 리터럴 키 1 · `exit_source_age.go` 의 임의 수} 밖에서 0 임을
   단언한다. 존재 검사가 아니라 금지 census 다 — 헬퍼가 판정에 브로커 시각을 흘리는 경로 전부를 이 범위가 닫는다.
   같은 시험이 보고 토글 식별자 `SourceAgeReport` 도 센다(허용 자리는 D8 「census 확장」) — 사용자 결정 2026-10-22 로 더한 토글이 판정 함수·
   `ObserveOnce` 에 분기로 들어오는 경로를 닫는다.
3. **어댑터~판정 e2e**(tasks 2.1~2.3): 공식 어댑터를 httptest 로 물려 어댑터 행 탈락·오류 반환 변이를 잡는다.

`observe`·`ObserveOnce`·`adaptPrices`·`engineRuntime` 은 편집 대상이므로 1 의 대상이 아니라 FLM/BTM `revision: base` 대조 대상이다
(편집 뒤 분기·반환 수 불변: `observe` 15·6, `ObserveOnce` 8·5(defer 2), `adaptPrices` 1·1 — 늘어나는 것은 호출·대입뿐).
세 시험은 각각 변이로 반증한다(tasks 4.1 ⑨⑩⑬).

## 사람 결정

| # | 질문 | 선택지 | 권고(근거) |
| --- | --- | --- | --- |
| H1 | 「오래된 체결」 알림 임계 — 시장·세션별 값 | (a) 단일 값 하나 · (b) KR/US 별 값 · (c) 정규장 안에서만 알림(장외는 기록만) · (d) 임계 없이 나이 기록만, 값은 H6 실측 뒤 | **(d) 후 (c)**. 장중 시각 동작이 0건 측정이라 숫자를 지어낼 근거가 없다. 청산 루프에는 세션 입력이 0 이다(a087 0.5) — 장외 알림을 거르려면 `internal/clock` 의 정규장 술어를 **보고 경로에만** 새로 들여야 하며, 그것도 이 결정의 일부다 |
| H2 | 알림 등급·경로 | (a) 일반 등급(로그 + 사이클 밖 유실 허용 전송, outbox 없음) · (b) critical(a074 운영자 도착 경로, outbox) · (c) 콘솔·로그만 | **(a)**. critical 은 outbox 기록 실패 시 진입 게이트 래치·모드 승격을 부른다(exit-policy 「durable 기록이 실패한다」) — 보고 전용 신호가 진입을 막는 경로가 생긴다. 보호 부재 사건(a090·a091)과 등급을 섞지 않는다 |
| H3 | 진입·전략·편입·flatten(C7~C10)도 지금은 보고만 할지 | (a) 이 change 는 청산 루프 보고만, 나머지는 잔여로 기록 · (b) C8·C9 에도 같은 보고를 더함 · (c) C10 에도 보고 | **(a)**. C10 은 비상 청산 경로라 이 change 의 경량 리뷰로 다루지 않는다. C8·C9 보고는 후속 Story 하나로 묶는 것을 권고 |
| H4 | 브로커 시각이 없거나(null·키 없음) 해독 불가일 때 | (a) 판정 그대로 + `source_missing` 기록·알림 · (b) 판정 그대로 + 기록만(알림 없음) · (c) 받은 시각으로 메움 | **(a) 또는 (b)**, **(c) 금지**(결정 35: 우리 시계로 메우면 보고가 언제나 「신선」이라고 거짓말한다). 판정은 어느 경우든 그대로(안전 4) |
| H5 | 브로커 시각을 원장(exit snapshot)에 보존할지 | (a) 로그·알림만 · (b) journal additive-nullable 열 | **(a)**. schema 변경은 순서·rollback 계획과 배포 쌍을 부른다(§0.6). 필요가 측정되면 별도 Story |
| H6 | 임계를 정하기 전 장중 실측 | (a) 사람 승인 읽기 전용 프로브(두 시장 장중, 엄격 리더 재사용) · (b) 배포 뒤 (d) 로그로 분포 수집 · (c) 실측 없이 값 지정 | **(b)** — 브로커 호출 0 추가, 운영 로그가 곧 표본. (a) 는 사람 승인 필요 |
| H7 | `source_at` JSON 노출 · `adaptOrderbook` 포함 | (a) `omitempty` 노출, 호가 제외 · (b) 노출 안 함(`json:"-"`) · (c) 호가 어댑터도 보존 | **(a)**. 호가는 생산 판정 소비자 0(C13·C14 는 엄격 리더) |

### 사람 결정 기록 — 사용자 결정 2026-10-21 「권장대로 착수 승인」

Manager 경유 사용자 회신(2026-10-21): H1~H7 전부 위 표의 권고대로. 아래가 구현의 정본이다.

| # | 결정 | 구현에 주는 뜻 |
| --- | --- | --- |
| H1 | **(d) 후 (c)** — 사용자 결정 2026-10-21 권장대로 | 이 change 는 체결 나이를 **기록만** 한다(로그 필드). 「오래됨」 알림의 임계와 정규장 한정은 H6 분포가 모인 뒤 별도 결정 — 그 전에는 `source_old` 알림을 내지 않는다 |
| H2 | **(a) 일반 등급** — 사용자 결정 2026-10-21 권장대로 | 로그 + 사이클 밖 유실 허용 전송. a074 critical outbox 경로·진입 래치·모드 강화 0 |
| H3 | **(a)** — 사용자 결정 2026-10-21 권장대로 | 청산 루프 보고만. C7~C10 과 a112 잔여 (k) 는 `issues.md` 잔여로 기록(tasks 5.1) |
| H4 | **(a)** — 사용자 결정 2026-10-21 권장대로 | 판정 그대로 + `source_missing` 기록·알림. 받은 시각·로컬 시계로 메우기 금지 |
| H5 | **(a) 로그·알림만** — 사용자 결정 2026-10-21 권장대로 | journal schema 변경 0 |
| H6 | **(b) 배포 뒤 로그로 분포 수집** — 사용자 결정 2026-10-21 권장대로 | 브로커 호출 0 추가, 프로브 없음 |
| H7 | **(a)** — 사용자 결정 2026-10-21 권장대로 | `source_at,omitempty` 노출, `adaptOrderbook` 제외 |

H1 과 H4 의 결합: 이 change 에서 알림을 내는 사유는 `source_missing`(H4) 하나이고 `source_old`·`source_ahead` 는 로그 필드로만 남는다.
`source_old` 알림은 H1 의 후속 결정(임계·정규장 술어)이 정해질 때 별도 Story 로 켠다.

**freeze 리뷰 반영 주석(2026-10-21, 결정 재론 아님)**

- H7: 위 표의 `source_at,omitempty` 는 D2 의 첫 가안 이름·타입을 옮겨 적은 것이다. freeze 리뷰 P1-1 로 그 형태가 0 값을
  `0001-01-01T00:00:00Z` 로 내보내는 결함임이 실측돼 구현 형태를 `SourceTimestamp string` · `source_timestamp,omitempty` 로 바꿨다(D2).
  H7 의 실질 — JSON 에 노출 · 값이 없으면 생략 · `adaptOrderbook` 제외 — 은 그대로다. 사람 결정의 범위를 넓히거나 좁히지 않는다.
- H2 보충 — 보고 끄기 토글: 2026-10-21 판에는 「두지 않는다」(설계 판단, 사용자 확인 대기)였다. **사용자 결정 2026-10-22 「토글 필요」로
  대체됨** — 아래 「사람 결정 기록 — 사용자 결정 2026-10-22」와 D8 이 정본이다.

### 사람 결정 기록 — 사용자 결정 2026-10-22 (설계 동결)

Manager 경유 사용자 회신(2026-10-22). 2026-10-21 처분 로트가 남긴 「사용자 확인 대기 잔여」 둘(`review.md`)과 수치 셋을 닫는다.
이 기록으로 설계를 동결한다(tasks 0.8).

| # | 결정 | 구현에 주는 뜻 |
| --- | --- | --- |
| H2 보충 | **토글 필요** — 사용자 결정 2026-10-22. 범위(Manager 결정): **a163 안에 둔다**(별도 Story 아님), **기본 OFF** | D8. OFF = 오늘 동작(새 로그 줄 0 · `source_missing` 알림 0). 토글은 보고·로그·알림 경로만 막고, 어댑터가 `timestamp` 원문을 `SourceTimestamp` 에 보존하는 것(D3)과 `observedQuote` 운반(D4)은 토글 밖이다. 생산에서 ON 으로 flip 하는 것은 사람의 행동이다(§0.7) |
| H7 형태 | **승인** — 사용자 결정 2026-10-22: `SourceTimestamp string` · json `source_timestamp,omitempty` 가 `source_at` time 을 대체 | D2 의 형태가 정본. 2026-10-21 표 H7 행의 `source_at,omitempty` 문자열은 이 결정으로 대체됨 |
| 수치 | **승인** — 사용자 결정 2026-10-22: 로그 구간 경계 **5분 · 1시간**, 한 주기 알림 **최대 8건** | D4 「분류」의 `le_5m`·`le_1h`·`gt_1h` 경계와 D4 「주기 상한」 8 이 정본(사람 승인 값). 바꾸려면 새 사람 결정 |

## D8. 보고 토글 — 사용자 결정 2026-10-22 「토글 필요」

search-first 결과: 새 설정 메커니즘을 만들지 않는다. 엔진 설정 블록의 bool 키 · 기동 1회 읽기 · 기동 audit 행이라는 **기존 세 패턴**을
그대로 쓴다. 함수 내부 주장은 이 로트에서 만든 AST 셋(`analysis/ast/config--mergeengine.ast.json` ·
`engine--recordgatesettings.ast.json` · `engine--context.exitobserver.ast.json`, `enumerate-output.md` 재생성 — AST 32 · STALE 0)에서 온다.

- **무엇을 막나**: 보고 진입 함수(`exit_source_age.go`)의 전부 — 구간 분류 · 래치 맵 갱신 · `SourceAgeLog` 전이 줄 · `source_missing` 알림.
  **막지 않는 것**: `adaptPrices` 의 `SourceTimestamp` 보존(D3), `observedQuote` 운반(D4 「운반」), `SourceAgeLog` 생산 배선(D4 「생산 싱크」 —
  싱크는 늘 꽂혀 있고 쓰기 여부만 토글이 정함). 그래서 공식 `/prices` 갈래의 `tossctl quote --json` 에 생기는 `source_timestamp` 키는
  토글과 무관하다(H7 사용자 승인 — WTS 갈래·`adaptStocks` 바이트 불변은 tasks 2.0 이 고정).
- **OFF 의 정의**: 새 로그 줄 0 · `source_missing` 알림 0 · 래치 맵 항목 0. 판정·발의·제출·원장·a090 집계·`settleUnobserved` 는 ON 과
  같다(D6·D7 — 토글은 판정 입력이 아니다). 엔진 로그·알림 표면은 a163 이전과 같다.
- **저장 — 기존 패턴**: `config.File` engine 블록의 bool 키 `engine.exit_source_age_report`. 선례 `Engine.Autostart`
  (`internal/config/engine.go:48` 필드 · `:265` raw `*bool` · `:328-330` merge). raw 는 `*bool` 이고 키 없음·명시 `false` 모두 `false`
  (선례 주석 `engine.go:246-248` · `notifications.go:76-79`). 「unset means on」 금지(`engine.go:17-21` 파일 머리 규칙). 태그는
  `json:"exit_source_age_report,omitempty"` — 꺼져 있으면 `tossctl config show`(`config.Status.Engine`, `internal/config/service.go:124·214`)
  출력 바이트가 편집 전과 같다(Autostart 는 omitempty 가 없어 늘 보이지만, 선례 `ExitPolicy` 필드들의 omitempty `engine.go:64-65` 를 따른다).
  bool 하나라 거부할 값이 없으므로 `Rejected` 필드를 두지 않는다.
- **merge 자리**: `mergeEngine`(`engine.go:324-358`) B4 조기 반환(`if raw.AutomationGate == nil { return }` `:342-343`, AST) **앞**.
  뒤에 두면 `automation_gate` 키가 없는 파일에서 토글이 조용히 무시된다 — `mergeNotifications` 를 같은 이유로 앞에 둔 주석 `:337-341` 과 같은
  규칙. 편집 뒤 `mergeEngine` 분기 5→6(B2 와 같은 모양 `if raw.X != nil` 하나), 반환 2 불변.
- **공개 스키마**: `schemas/config.schema.json` 의 engine 블록이 `additionalProperties: false`(`:95-98`)이므로 boolean · `default: false`
  속성을 더한다. `schema_version` 은 5 그대로 — 선례 a136(`c8783ad90`, 키 추가에 버전 유지). `TestPublishedSchemaMatchesSchemaVersion`
  (`internal/config/engine_test.go:186-206`)이 그대로 통과해야 한다.
- **읽는 시점 — 기동 1회(기존 패턴, 주기 재읽기 아님)**: 엔진은 설정을 `engine.go:370`(`config.NewService(paths.ConfigFile).Load`)에서 한 번
  읽고, `internal/app/engine` 비시험 파일에서 주기마다 설정을 다시 읽는 곳은 0 이다(2026-10-22 grep 전수 — `config.NewService`·`.Load(`·`LoadRaw`
  출현은 그 한 줄). 청산 관측자에는 `Context.ExitObserver`(`exitwiring.go:319-362`)가 로드된 설정을 옵션에 **덮어써서** 넘긴다 — 선례
  `opts.NotificationsEnabled = c.Config.Engine.Notifications.Enabled`(`:355`)와 `opts.CommonPolicy = …`(`:342`), AST 상 둘 다 어느 분기 안에도
  없는 무조건 대입. 새 옵션 `ExitObserverOptions.SourceAgeReport bool`(선례 필드 `NotificationsEnabled` `exitloop.go:210-213`)을 같은 자리에서
  `opts.SourceAgeReport = c.Config.Engine.ExitSourceAgeReport` 로 덮는다(호출자 값과 무관 · 분기 0 추가 — `Context.ExitObserver` 분기 6·반환 4
  불변). 주기 재읽기를 고르지 않은 이유: 엔진 선례 0 이고, 설정 파일 I/O 와 그 실패 분기가 청산 goroutine 에 들어온다.
  결과: flip 은 **다음 엔진 기동부터** 효력이다. 엔진 재기동은 그동안 손절 판정이 없으므로 두 시장이 닫힌 창(KST 05:00~09:00)에서만 한다
  (기존 운영 규칙).
- **판단 자리**: 보고 진입 함수 안, recover 경계(D4 「패닉 격리」)를 세운 **바로 뒤 첫 판단** — 꺼져 있으면 즉시 반환한다. `ObserveOnce` 의 보고
  호출 한 줄(D4 「자리」 — 판정 순회 뒤 · `settleUnobserved` 앞)은 토글과 무관하게 무조건이다. 그래서 ① `ObserveOnce` 분기·반환 8·5 불변(D7),
  ② 판정 함수 열 개 렌더 해시 불변(D7 1), ③ 토글은 판정과 주문 제출 **사이에 서지 않는다**(호출 자리가 이미 그 주기의 판정·제출 뒤).
- **census 확장(D7 2)**: `SourceAgeReport` 식별자도 `internal/app/engine` 비시험 파일의 모든 출현을 센다. 허용 = `ExitObserverOptions` 필드 선언 1 ·
  `exitwiring.go` 덮어쓰기 1 · `exit_source_age.go` 임의 수, 그 밖 0 — 토글이 판정 함수·`ObserveOnce` 에 분기로 들어오는 경로를 닫는다.
- **audit — 기존 기동 audit 행**: `recordGateSettings`(`internal/app/engine/interlock.go:739-811`)의 `changes` 목록 **끝**에 한 행
  (선례 a136·a138·a139 행 `:795-803` 「끝에 덧붙임(위 순서 무변경)」). setting `engine.exit_source_age_report` · 값 `strconv.FormatBool` ·
  action 새 상수 `audit.ActionExitSourceAgeReport = "exit.source_age_report"` — 다른 것을 허가하는 설정은 자기 action 을 갖는다는 선례
  (`internal/audit/audit.go:64-101` 주석: adoption·trading·notifications). 이 스위치는 주문을 허가하지 않는 보고 스위치라 gate·notification
  action 에 섞지 않는다. 기록은 `RecordChange`(`audit.go:217-239`)라 값이 바뀐 기동에만 행이 생기고(old→new), 같은 값이면 0 행이다.
  함수 시그니처에 끝 인자 `sourceAgeReport bool` 을 더한다 — 호출자는 생산 1(`engine.go:472`, `cfg.Engine.ExitSourceAgeReport`) · 시험 4
  (`a074_notification_audit_test.go:55·84·105·124`, 기계적 `false`). 실패 모양은 불변: 이 행의 기록 실패도 기존 행과 같은 `B6 → return err`
  (AST `engine--recordgatesettings`) → 기동 거부(`engine.go:472-479`)다 — 같은 파일·같은 쓰기라 새 실패 종류가 아니다.
- **표시**: `tossctl config show` 가 켜졌을 때만 키를 보인다(omitempty). **콘솔 스위치는 두지 않는다**(비목표) — 사람의 flip 은 config 파일
  편집 + 엔진 재기동이고 기동 audit 행이 그 변경을 남긴다(§0.5). 콘솔 토글이 필요하면 별도 Story(콘솔 저장 경로는 자기 audit 행을 쓴다 —
  선례 `cmd/tossctl/engineautostart.go:28-45`). 이 change 는 이 값을 **쓰는** 코드(config 저장 함수)를 더하지 않는다 — 에이전트·코드 flip 경로 0.
- **H6 과의 결합**: 기본 OFF 이므로 H6 (b) 「배포 뒤 로그로 분포 수집」은 사람이 켠 기동부터 시작된다. 켜기 전 기간의 표본은 없다.

## 위험

- **판정 경로에 실수로 닿는 편집**: 보고 코드가 `quoteUsable` 옆에 서면 C 안이 된다. 막는 법 — D7 의 함수별 렌더 해시 시험 ·
  `SourceTimestamp` 금지 census · 어댑터~판정 e2e, 편집 대상 함수는 FLM `revision: base` 대조, 그리고 「보고가 손절을 건너뛰게 한다」
  변이(보고 사유가 있으면 `continue`)가 시험에 걸리는지 변이 원장에 남긴다.
- **보고 코드의 패닉이 손절을 멈추는 편집**: 감독자에 recover 가 없어(D4) 보고 패닉 = 엔진 종료다. 막는 법 — 보고 진입 함수의 recover 경계,
  패닉 주입 시험(2.10), 경계 제거 변이(⑪), 래치 맵 단일 goroutine 소유와 `-race` 등재.
- **시험만 초록이고 생산 로그 0 줄**: 로거를 주입하는 시험과 nil 인 생산 배선의 차이(a129 선례). 막는 법 — `SourceAgeLog` 생산 배선 시험(2.11)과
  배선 제거 변이(⑫).
- **보고 시간이 손절 앞에 서는 편집**: 보고를 순회 안에서 보내면 뒤 포지션의 판정이 늦어진다. 막는 법 — 순서 시험(가짜 알림기가 첫 포지션
  보고 시점에 뒤 포지션 제출 수를 단언).
- **스팸**: 장외 모든 보유 종목이 매 주기 줄·알림을 내는 것 — 로그는 구간 전이 때만, 알림은 `source_missing` 하나에 래치 + 주기당 8건 상한(D4).
- **토글이 판정 앞에 서거나 무시되는 편집**(D8): 토글 판단을 `ObserveOnce` 나 판정 함수에 분기로 넣으면 판정 경로에 새 분기가 생기고, 토글을
  무시하거나 기본값을 ON 으로 뒤집으면 OFF = 오늘 동작이 깨진다. 막는 법 — 판단은 보고 진입 함수 안에만(`SourceAgeReport` census), OFF 시험
  (줄 0 · 알림 0 · 판정·제출이 ON 과 같음), 기본값·merge 자리 시험, 배선 덮어쓰기 시험, audit 행 시험, 그리고 변이 ⑰~㉒(tasks 4.1).
