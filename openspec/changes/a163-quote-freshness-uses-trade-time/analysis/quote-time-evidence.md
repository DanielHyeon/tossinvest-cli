# a163 — 시세 시각의 생산자·소비자·브로커 계약 열거 (proposal 근거)

- 측정 순간: 2026-10-21, HEAD `f7d80c348f612cb7d981cea0a958c81b40d8b7c8`(브랜치 `feat/a112-four-family-runtime`).
  대상 production 파일은 워킹트리 무편집(`git status --short internal/official internal/app/engine internal/flatten
  internal/hybrid internal/breakoutlane internal/officialbars` 빈 출력). 각 `ast/*.ast.json` 의 `source_sha256` 을
  `ast/enumerate.py` 가 실행 때 다시 대조한다(STALE 0).
- 도구: `tools/logic-map`(HEAD 에서 `GOFLAGS=-trimpath` 빌드) · **codegraph 1.6.0**(index 2,929 files · 51,218 nodes) ·
  `grep`(이 환경의 `rg` 는 rtk 가 grep 으로 바꾼다).
- 성격: read-only. production 편집 0 · 브로커 호출 0 · `tossctl` 0.
- 이 문서의 분기·early return·부수효과 주장은 전부 `ast/enumerate-output.md`(AST 22개 · 분기 162 · 반환 109, 손편집 금지)의
  좌표에서 온다. AST 를 만들지 않은 함수는 표에 **AST 없음**으로 적고, 그 줄에서는 분기 주장을 하지 않는다.

## 1. 결함 — `FetchedAt` 은 받은 시각이다

`adaptPrices`(`internal/official/market_reads.go:168-180`) AST: 분기 1(B1 `range` `:170`) · 반환 1(`:179`) · 시각 호출 1
(`:175` `time.Now`). 호출 텍스트 전부가 `make`·`len`·`append`·`parseDecimal`·`UTC`·`time.Now` 이고 **`Timestamp` 를 읽는 호출·대입은
없다.** 응답 구조체는 `apiPrice.Timestamp`(`:137`, json `timestamp`)를 디코드만 하고 버린다. `adaptOrderbook`(`:229-261`) 도 같다 —
분기 2(B1·B2 `range`) · `:258` `time.Now` · `apiOrderbook.Timestamp`(`:205`) 미사용.

`Client.Prices`(`:142-150`) AST: 분기 1(B1 `:146` `c.get` 오류) → `return nil, err` / `:149` `return adaptPrices(raw), nil`.
비시험 호출자는 CodeGraph `callers adaptPrices` = `Prices` 하나.

그래서 `domain.Quote.FetchedAt`(`internal/domain/models.go:421`)은 공식 `/prices` 경로에서 **응답을 디코드한 우리 시계**다.
같은 파일의 `adaptTrades`·`adaptMarketIndicatorPrices`·`adaptMarketIndicatorCandles` 는 브로커 시각을 보존한다(`:311`·`:436`·`:489`
— 다른 함수이므로 이 문서의 결론 근거가 아니라 대조 사실).

선례: 엄격 리더(`internal/official/strict_quote_reads.go`)는 이 결함을 이미 알고 따로 만들어졌다 — 파일 머리 주석 「기존
Orderbook/Prices는 … 브로커가 보낸 시각(timestamp)과 통화를 **버리고**」, `strictQuoteInstant` 주석 「우리 시계로 대신하지 않는다
(결정 35) … 그 값이 우리 읽은 시각이면 그 검사는 자기 시계를 재는 셈이 되어 언제나 통과한다」. a163 의 결함은 같은 문장을
`/prices` 일반 경로에 적용한 것이다.

## 2. 브로커 계약 — 시각은 있고, nullable 이다

### 2.1 OpenAPI (`docs/migration/openapi.latest.json`, 마지막 갱신 커밋 `926d67a2f`)

| 스키마 | 필드 | 타입 | 설명(원문) | required |
| --- | --- | --- | --- | --- |
| `PriceResponse` | `timestamp` | `["string","null"]`, `date-time` | 「데이터 시각. 체결 미발생 등으로 시각이 없을 경우 null」, 예 `2026-03-25T09:30:00.123+09:00` | **아니오**(`symbol`·`lastPrice`·`currency` 만) |
| `OrderbookResponse` | `timestamp` | `["string","null"]`, `date-time` | 「데이터 시각. 데이터 미제공 시 null」, 같은 예 | **아니오**(`currency`·`asks`·`bids`) |
| `Trade` | `timestamp` | `string`, `date-time` | 「체결 시각」 | 예 |

읽기: 계약은 `/prices` 의 `timestamp` 를 「데이터 시각」이라 부르고 null 사유로 「체결 미발생」을 든다 — 체결과 묶인 시각이라는
**암시**이지 「마지막 체결 시각」이라는 **명시**는 아니다(호가 갱신 시각일 가능성을 계약이 배제하지 않는다). 키 자체가 빠질 수 있다
(required 아님). → 시각의 의미는 `[계약 모호 — 장중 실측 없음]`.

### 2.2 기록된 실측 (사람 승인 읽기 전용 프로브, a112 결정 46)

출처: `openspec/changes/archive/2026-10-04-a112-run-four-strategy-families-independently/review.md:774-835`
(「2026-08-28 L1c probe runs — both markets CLOSED」). 엄격 리더 두 개로 GET 4회(두 시장 × 두 엔드포인트). 본문·가격은 저장소에
남기지 않았고 digest 와 시각만 남았다. 계좌 정보 없음.

| | US / AAPL | KR / 005930 |
| --- | --- | --- |
| `/orderbook` broker instant | `2026-08-28T05:28:58.000+09:00` | `2026-08-27T20:00:00.000+09:00` |
| `/prices` broker instant | `2026-08-28T05:28:30.000+09:00` | `2026-08-27T19:59:59.000+09:00` |
| 읽은 시각 | 2026-08-28 05:29–05:31 KST | 같음 |
| 「Both `timestamp` fields were non-null in all four bodies」 | 예 | 예 |

여기서 확정되는 것:

1. **두 시장 모두 `/prices`·`/orderbook` 에 시각이 실렸다**(4/4 non-null). 형식은 밀리초 3자리 + 숫자 오프셋.
2. **US 시각도 `+09:00`(KST 오프셋)으로 온다.** 시장 현지 시각이 아니다. RFC 3339 로 해독하면 절대 시각은 같으므로 나이 계산에는
   무해하지만, 오프셋으로 시장을 추정하면 틀린다.
3. **KR 은 9시간 30분 묵은 시각이 성공으로 왔다**(review.md:817 잔여 (k)). KR 본문 시각 `19:59:59`/`20:00:00` 은 정규장 종료
   15:30 이 아니라 20:00 이다 — NXT 애프터마켓 종료와 일치한다는 해석은 `[추정 — 미검증]`.
4. **US 는 정규장(05:00 KST 종료) 뒤의 시각**(05:28)이다 — 애프터마켓 체결이 시각을 밀고 있다는 해석은 `[추정 — 미검증]`.

확정되지 않는 것: **장중 동작은 한 번도 관측되지 않았다**(위 4건은 전부 두 시장이 닫힌 시각). 장중 갱신 주기, 유동성 낮은 종목의
시각 지연, VI·거래정지 중 값, null 이 실제로 오는 조건 — 전부 `[미측정]`. 저장소의 `/prices` 시험 픽스처(`market_reads_test.go:127`)는
`timestamp` 키가 **없는** 합성 본문이고, `strict_quote_reads_test.go:292-295` 는 합성 거절 사례다 — 브로커 실측이 아니다.

## 3. 소비자 열거 — 나이로 무엇을 결정하나

CodeGraph 1.6.0 `callers Prices`(비시험 18) · `callers quoteUsable`(6) · `callers adaptPrices`(1) · `callers GetQuote`(비시험 1:
`cmd/tossctl/quote.go`) · `callers stableObservationID`(비시험 1: `snapshotContext`) 와 `grep` 으로 모은 `/prices`·`/orderbook`
소비 경로. 생산 배선 열은 a087 0.5 (`exitwiring.go:337` `opts.Prices = c.Official`)와 `cmd/tossctl/flatten.go:262`
(`Prices: orderPath.Official`) 원문.

| # | 소비자 (AST) | 읽는 시각 | 결정 | 임계 | 「오래됨」일 때 (AST 좌표) | 생산 |
| --- | --- | --- | --- | --- | --- | --- |
| C1 | `ExitObserver.observe` `exitloop.go:842-913` | `q.FetchedAt`(=받은 시각) vs `validatedAt`(`:855` `clock.LeaseAnchor`) | 종목을 판정 후보에 넣나 | `execgw.QueryPriceEvidenceDuration` = 15 s (`retry.go:192`) | B7 `:871-872` 미래 또는 15 s 초과 → 그 종목 `continue`(후보 제외). 전부 제외면 B8 `:878` → `:879` `NewEvidenceInvalidError`(`ClassifyQueryError` → `ClassPermanent`, 재시도 없음). B5 `:868` `FetchedAt` 0 이면 `cycle:` 대체 시각 | 예 — 청산 루프 |
| C2 | `ExitObserver.quoteUsable` `:1179-1188` | 벽시계 vs `FetchedAt`(B1 `:1184` 미래 거절) · 로컬 단조 경과 `:1187` `clock.LeaseElapsed(leaseAnchor) ≤ 15 s` | 이 시세를 지금 써도 되나(사용 임대) | 15 s | `false` → 호출자 6곳이 각각 판정을 건너뜀(C3) | 예 |
| C3 | `quoteUsable` 의 호출자: `ObserveOnce` B7 `:511` · `judge` B1 `:963` · `judgeRatchet` B5 `:1067` · `judgeLadder` B7 `:1138` · `refreshObservation` B1 `:1161` · `record` B1 `:1291` | — | — | — | `ObserveOnce` B7: `noteUnobservedCause(…, unobservedQuoteExpired)` + `continue`(판정 없음). 나머지 다섯: `return nil`(판정·기록·refresh·제출 없음) | 예 |
| C4 | `ExitObserver.settleUnobserved` `exit_unobserved.go:139-204` | 마지막 판정 이후 로컬 경과 `:198` | 미관측 포지션을 critical 로 올리나 | `outageAfter()` = 60 s | B12 `:199` 미만이면 `continue`, 이상이면 `:202` `raiseUnobserved`(a090: critical 알림 기록 + 모드 강화, 파일 머리 주석) | 예 |
| C5 | `ExitObserver.checkOutage` `:916-958` | `lastObserved`(로컬) | 계좌 두절 | 60 s | B2 `:921` 미만 → return; 넘으면 알림 + `EscalateOperatingMode`(ENTRY_BLOCKED) | 예 |
| C6 | `stableObservationID` `:1202-1226` · `record` B3 `:1313` · `refreshObservation` | `FetchedAt` 을 **신원과 기록 값**으로 | 관측 identity · 원장 `ObservedAt` | — | 결정 아님. B2 `:1209` `FetchedAt` 0 이면 `cycle` 신원. **`FetchedAt` 의 의미를 바꾸면 관측 identity 가 바뀐다** — 같은 체결 시각·같은 가격의 두 주기가 같은 id 가 되어 spec 「같은 observation이 재전달된다」(성공 no-op)로 흡수된다 | 예 |
| C7 | `EntryGate.RecordSuccess` `execgw/retry.go:530-540` | `:533` `g.clk.Now`(로컬) | 진입 게이트의 `QueryPrice` 신선도 | `DefaultStaleness[QueryPrice]` = 15 s | B1 `:536` — 받은 시각으로 갱신. 시세의 브로커 시각은 이 게이트에 닿지 않는다 | 예 — 진입 |
| C8 | `PositionPolicyCommandService.prepare` `position_policy_command.go:303-365` | `:323` `s.clk.Now`(로컬) | 재편입 t0 가격 → 손절가 유도(`:349` `ValidateStopPct`) | 없음 | 나이 분기 0. B10 `:340` `Last > 0` 이면 채택, B11 `:345` 없으면 오류. **오래된 체결가가 그대로 t0** | 예 — 운영자 명령 |
| C9 | `ReconcileDriver.observeCandidates` `adoption.go:271-342` | `:298` `d.clk.Now` → `readAt` | 자동 편입 t0 가격 | 없음 | 나이 분기 0. B8 `:319` `Last <= 0` 만 거름, B11 `:336` 통화. **오래된 체결가가 그대로 t0** | 예 — 편입 |
| C10 | `Saga.aggressivePrice` `flatten/liquidate.go:372-398` | 없음 | 비상 청산 지정가 | 없음 | B2 `:374` 하한가가 있으면 그것(`:377`). 없으면 B5 `:385` `Last <= 0` 만 거르고 `:394` `Last×(1−discount)`. **나이 분기 0 — 오래된 체결가로 지정가를 매긴다** | 예 — flatten-all(High-risk, 불변식 4) |
| C11 | `Tracer.freshenGateInputs` `tracer.go:353-403` | `:379`·`:398` `Clock.Now`(요청~응답 경과) | 사람이 모는 tracer 의 진행 | `Params.Freshness` | B10 `:398` 초과 → `ErrTracerRefused`. 왕복 시간이지 체결 나이 아님 | 사람 진단 경로 |
| C12 | `hybrid.Client.GetQuote` `hybrid/client.go:117-130` → `cmd/tossctl/quote.go` → `internal/output/quote.go:134` | `FetchedAt` 를 바닥글 시각으로 **표시** | 사람이 보는 시세 시각 | — | 결정 없음. 사람에게 받은 시각을 「시세 시각」처럼 보인다 | CLI 표시 |
| C13 | `breakoutlane.validateQuote` `breakoutlane/rules.go:32-59` | `SourceObservedAtMS`(**브로커 시각**, 엄격 리더 경유) | breakout 진입 거절 | `MaxQuoteAgeMS`(a140 도구 기본 `--max-quote-age-ms 5000`) | B2 `:37` 순서 역전 · B3 `:40` 나이 초과 → `RefusalQuoteStale` | **생산 0** — `PollQuoteL1` 의 비시험 호출자는 `tools/a140-breakout-signal-shadow/shadow.go:289` 하나(그림자) |
| C14 | `officialbars.PollQuoteL1` `officialbars/quote.go:98-188` | 브로커 시각 둘(이른 쪽 = observed, 늦은 쪽 = received) | 증거 적재 | 없음 | B10 `:142` `IsZero` → `RefusalQuoteInstant`(**존재만** 검사). 나이 분기 0 = a112 잔여 (k) 그대로 | 그림자만 |
| C15 | `tools/a128-jev-shadow-probe/collect.go:403·822` | **AST 없음** — `p.now()` 로컬 | 그림자 라벨의 기준가·실현가 | — | 분기 주장 안 함. 원문상 `quote.Last` 만 쓰고 시각은 로컬 | 그림자 도구 |
| C16 | `verifylive` `marketFrame`·`marketTop`, `soak` `probePrices`, `ops/read_operations.go:186` | **AST 없음** | 사람 승인 실검증 · 부하 측정 · CLI | — | 이 change 범위 밖으로 둔다(분기 주장 안 함) | 사람 도구 |

요약:

- **생산에서 시세 나이로 무엇을 거르는 곳은 청산 루프(C1~C5)와 진입 게이트(C7) 둘뿐이고, 둘 다 받은 시각을 잰다.**
- 브로커 시각을 쓰는 유일한 판정(C13)은 생산 호출자가 0 이다.
- 나이 검사가 **아예 없는** 생산 경로가 셋이다: 재편입 t0(C8), 자동 편입 t0(C9), 비상 청산 지정가(C10).

## 4. 핵심 긴장 — 체결 나이로 바꾸면 손절이 건너뛰어진다

### 4.1 지금 `quoteUsable` 이 `false` 일 때 (AST)

`ObserveOnce` B7 `:511` → `noteUnobservedCause(unobservedQuoteExpired)` · `continue`: **그 포지션은 이 주기에 판정되지 않는다** —
손절선 비교(`judgeRatchet`/`judgeLadder`)·기록·제출이 전부 없다. `judge`·`judgeRatchet`·`judgeLadder`·`refreshObservation`·`record` 의
`quoteUsable` 거짓은 모두 `return nil`. 그 뒤의 안전망은 시간 기반이다: 포지션 단위 `settleUnobserved` B12(60 s 이상 미판정 →
critical + 모드 강화), 계좌 단위 `checkOutage` B2(60 s 이상 무관측 → ENTRY_BLOCKED). **안전망은 진입을 막고 사람을 부를 뿐
그 포지션을 팔지 않는다.**

`observe` 단계에서 걸러지면(B7) 그 종목은 `quotes` 맵에 없으므로 `ObserveOnce` B6 `:503` → `unobservedNoQuote` 로 같은 결말이다.
전 종목이 걸러지면 B8 → `EvidenceInvalid`(`ClassPermanent`) → `ObserveOnce` B4 `:492` → `checkOutage` 후 `return cycle`.

### 4.2 신선도를 체결 나이(15 s)로 바꾸면 새로 「오래됨」이 되는 입력

계약(2.1)과 실측(2.2)만으로 셀 수 있는 것과 추정을 가른다.

| 입력 | 근거 | 지금 | 체결 나이 판정이면 |
| --- | --- | --- | --- |
| 닫힌 시장의 마지막 체결가 | 실측 — KR 9h30m(2.2-3) | 15 s 안에 받았으면 판정됨 | 판정 0. 다음 체결 전까지 손절 비교 없음 |
| `timestamp` null / 키 없음 | 계약 — 「체결 미발생 등」, required 아님 | `FetchedAt` 이 늘 채워지므로 판정됨 | 해독 불가 → 판정 0(또는 `cycle:` 대체로 회귀 — 설계 선택) |
| 장중 15 s 이상 체결 없는 저유동 종목 | `[추정 — 장중 미측정]` | 판정됨 | 체결 사이마다 판정 0 |
| KR VI(단일가 2분) · 거래정지 · 서킷브레이커 | `[추정 — 미측정]` | 판정됨(묵은 가격으로) | 정지 내내 판정 0 → 재개 첫 체결까지 손절 대기 |
| KR 장전·장후 시간외 · NXT · US 프리/애프터 | `[추정]` — 2.2-3·4 의 시각이 그 구간을 가리킴 | 판정됨 | 체결 간격이 15 s 를 넘을 때마다 판정 0 |
| 시계 차(브로커 시각이 우리 시계보다 미래) | C1 B7 「미래」 규칙을 브로커 시각에 그대로 쓰면 | — | 브로커 시계가 앞서면 정상 시세도 거절 |

**결론.** 체결 나이로 청산 판정을 거르는 순간, 위 여섯 경우마다 `quoteUsable` 이 `false` 가 되고 4.1 의 경로대로 **손절 비교가
건너뛰어진다.** 특히 VI·정지처럼 가격이 가장 크게 움직일 수 있는 구간이 정확히 체결이 멈추는 구간이다. 이것은 안전 불변식
「손절·비상 청산의 즉시성을 약화·지연하지 않는다」(`.claude/CLAUDE.md` 안전 4, `docs/WORKFLOW.md` §0.3) 위반이다. 반대 방향
— 받은 시각을 유지하면 — 묵은 체결가로 판정해 장외 시간에 손절이 발의될 수 있다(a087 0.5 결론 2, D7 처분 대기). 그 위험은
a087 의 몫이고 이 change 가 키우지도 줄이지도 않는다.

## 5. 한계

- 장중 브로커 시각의 동작은 0건 측정이다(2.2). 임계값 숫자는 이 문서에서 나오지 않는다.
- C15·C16 은 AST 를 만들지 않았다 — 표의 그 줄에는 분기 주장이 없다.
- `codegraph affected … --filter '*_test.go'` 는 `market_reads.go` 에 1,197건, `exitloop.go` 에 1,216건을 낸다
  (`codegraph-affected-*.txt`). 전이 의존이 넓어 시험 선택의 근거로는 쓰지 않는다 — 직접 시험 파일 목록은 tasks 1.x 에서 고른다.
- 분류는 AST 좌표의 **첫 줄**만 본다(여러 줄 조건의 둘째 줄은 `calls` 로 따로 잡힌다 — 예 `observe` `:872` `validatedAt.Sub`).
