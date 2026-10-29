# a087 설계 — 보호 청산의 가격 (12판: 그날 거래소 하한가 지정가 · 1~7판: 시장가)

> **9판 (Manager 결정 2026-10-29 — 8판 독립 재리뷰 FAIL 처분)**: 정본은 **D11 + D12** 다. D12 가 D11 의 신선도(timestamp · raw decimal) · 갱신 주기(관측 주기 끝 토큰 ·
> 주기당 1건 · 429 뒤 5분 정지) · 되먹임(`CONFIRMED` 아닌 모든 결과 · 재시작 재구성) · 불변식 4 표 · a165 문장을 대체한다.
>
> **10판 (Manager 결정 2026-10-29 — 9판 재리뷰 FAIL 처분, KISS)**: 정본은 **D11 + D12 + D13** 이다. D13 이 되먹임을 「그날 그 종목의 첫 보호 매도 시도에만 하한가」로,
> 429 정지 입력을 `doRequest` 의 HTTP 429 시각 하나로 바꾸고, 불변식 4 표(D13.5)에 release 실패 칸을 더한다. timestamp 뜻은 5.4 gate 로 보류(D13.3).
>
> **11판 (Manager 결정 2026-10-29 — 10판 재리뷰 FAIL 처분)**: 정본은 **D11 ~ D14** 다. D14 가 재구성 술어의 action 을 Go 상수에서 유도하고(`BASELINE_BREACH_EXIT` 오기 정정),
> wire 상한(401 시 토큰당 3) · 엔진 시작 뒤 5분 GET 0 · 예산 기준 계좌 공유 3 req/s 를 적고, 불변식 4 표를 D14.4 로 바꾼다.
>
> **12판 (Manager 결정 2026-10-29 — 11판 재리뷰 FAIL 처분)**: 정본은 **D11 ~ D15** 다. D15 — 하한가 GET 은 a166 계좌 공유 admission 이 주입될 때만(엔진 배선은 그 전까지 nil →
> 생산 동작 변화 0, 활성화 조건), `doRequest` 429 기록은 a166 으로 옮겨 a087 의 `doRequest` 편집 0, 구조 시험은 평가된 `Action` 값 집합 비교, 불변식 4 표 D15.3.

> **8판 (사람 결정 2026-10-28 「3은 3」 — `issues.md` D-1 (다))**: 보호 청산 매도는 시장가가 아니라 **그날 거래소 하한가로 낸 지정가(LIMIT)** 다.
> 2026-10-10 「시장가 방향」 결정을 이것으로 대체한다. **정본은 D11** 이다. 아래 「문제의 형태」·D1~D10 의 시장가 전용 설계는 지우지 않고 절마다
> 「8판에서 대체됨」을 달았다. 8판에서도 그대로 유효한 것: D1 의 `isProtective` 재사용 원칙과 익절 무변경, D2a 「반증」, D3 「왜 진입은 계속 막는가」,
> D4 의 원칙(원장은 보낸 것을 적는다), D5 대조, D8 의 잔존 위험 행, D9, D10. change id(`…-is-a-market-order`)는 경로 안정성 때문에 바꾸지 않는다.

## 문제의 형태

> **8판에서 대체됨(일부)**: 「가격을 없앤다」는 1~7판의 축이다. 8판은 가격을 없애지 않고 **거래소가 공표한 가격(하한가)으로 바꾼다** — 아래 표의
> 「시장가에서」 칸 대부분(그리드 이탈 · 중간가 · override · KOSDAQ 표 · `big.Rat` · 사본 정본화)은 하한가 지정가에서도 같은 이유(관측가를 가격으로 쓰지 않음)로 성립한다.
> 「가격 밴드 clamp — 브로커가 처리」 칸은 8판에서 「하한가는 밴드의 끝값 그 자체」로 읽는다(D11 (a)).

초안은 "제출가가 호가 그리드 위에 있게 하자"였다. proposal-freeze 리뷰가 그 설계를 무너뜨렸고
(`review.md`), 무너진 자리에서 더 단순한 질문이 남았다.

**보호 청산이 왜 가격을 갖는가?**

답은 없었다. 코드가 인용한 근거(`riskcalc`의 LIMIT 전용)는 진입 전용 규칙이었고, 코드가
실제로 적은 이유는 "원장에 적을 가격이 필요해서"였다. 거부되지 않을 주문 형태를 원장 표기와 맞바꾼 것이다.
(3판까지 「체결 보장을 원장 표기와 맞바꿨다」였다 — **정정(재리뷰 P1-2 · 2차 리뷰 C3)**: 시장가는 체결을 보장하지
않는다. 경계 조건은 D8.) 그리고 2차 리뷰가 찾은 **셋째 이유**가 있다 — 이 fork 의 서비스 계층
(`internal/trading` `placeIntentSupported`)이 비분수 주문에 대해 지원하는 유일한 형태가 LIMIT 이다(D3).

가격을 없애면 초안이 풀려던 문제가 **전부 존재하지 않게 된다**:

| 초안이 풀려던 것 | 시장가에서 |
| --- | --- |
| 호가 그리드 이탈 (실측 5회 거부) | 가격이 없다 |
| 시세가 중간가일 가능성 (D1) | 가격을 안 쓴다 |
| ETF·우선주 override 갭 | 표를 안 탄다 |
| KOSDAQ 표 분기 | 〃 |
| `big.Rat` 정밀도 (US 48% 이동) | 〃 |
| 세 사본 정본화·AST 가드 | 이 경로에서는 불필요 |
| 가격 밴드 clamp | 브로커가 처리 |

## D11 — 8판 정본: 보호 매도 = 그날 하한가 지정가 (사람 결정 2026-10-28 「3은 3」, `issues.md` D-1 (다))

> 번호 D11 은 기존 D1~D10 을 인용한 문서를 깨지 않으려고 뒤에 붙였고, 읽는 순서를 위해 앞에 둔다. 아래 D1~D10 의 시장가 전용 설계는
> 지우지 않고 각 절 머리에 「8판에서 대체됨」을 달았다. 분기 근거는 HEAD `1f4145c77`(origin/main `4b67dd6a9` 위 문서 커밋 3개 — 소스 바이트는 main 과 같다)의
> AST 36개(`analysis/evidence/*.ast.json`, `python3 -I …/analysis/evidence/enumerate.py` stale 0, 열거 `enumerate-output.md`)다. 이 판에서 새로 만든 AST 는
> 11개다 — `flatten--saga.aggressiveprice` · `official--client.pricelimits` · `official--adaptpricelimits` · `official--parsedecimal` · `client--client.getpricelimits` ·
> `engine--exitorderlinefor` · `engine--exitorderline.args` · `engine--context.manualbuyrefresher` · `engine--manualbuyrefresh.run` · `engine--exitobserver.run` ·
> `tossctl--engineruntime`.

### (a) 주문 형태 — 유형은 LIMIT 그대로, KR 정규장 보호 매도의 가격만 하한가

| 제안 | 시장 | 조건 | 주문 |
| --- | --- | --- | --- |
| 보호(`isProtective` — `BASELINE_BREACH_EXIT` · `STOP_LOSS_LADDER`) | KR | KRX 정규장 고정 창 `[09:00, 15:30)` KST · 평일(`clock.MarketKR.InRegularSession`) **그리고** 엔진 메모리에 그 종목의 그날 하한가가 유효하게 있다(아래 (b) 조건 ①~⑥) | **LIMIT @ 그날 하한가** |
| 보호 | KR | 그 밖 — 창 밖 · 하한가 없음 · 낡음 · 검사 실패 · 그날 거부됨 | **오늘과 같은 LIMIT**(관측가 → 기준선 → 거부) |
| 보호 | US | 언제나 | 오늘과 같은 LIMIT — 가격제한 제도가 없다((d)) |
| 익절(`LADDER_TAKE_PROFIT`) | 모두 | 언제나 | 오늘과 같은 LIMIT(§0.9 — D1 「익절을 안 바꾸는 이유」 그대로) |

- 유형이 바뀌지 않으므로 바뀌는 것은 **가격 인자 하나**다. `sellIntent`(`exitloop.go:1806-1832`, B1 `:1808` 관측가 빔 → 기준선 · B2 `:1811` 가격 없음 거부 · B3·B4 파싱)는
  **바이트 그대로** 둔다. `submit` 이 `sellIntent` 를 부르는 자리(`:1532` `o.sellIntent(m, submitQuantity, observed)`)에 관측가 대신 아래 (b) 의 결정 가격을 넘긴다.
- **그리드 위라는 근거**: ① 하한가는 거래소가 그날 공표하는 가격제한폭의 끝값이다 — openapi `/api/v1/price-limits` 「종목의 당일 상한가 및 하한가를 조회합니다」
  (`docs/migration/openapi.latest.json:573`), 응답 `lowerLimitPrice`(`:8577`). ② 저장소 선례: `flatten` 이 이미 이 값을 1순위 매도 지정가로 쓴다 —
  `Saga.aggressivePrice` B2(`internal/flatten/liquidate.go:374`, `:377` 반환 출처 문자열 `"exchange lower limit"`), 주석 「maximally aggressive, always a valid tick,
  and never rejected for being outside the band」(`:375-376`). ③ **[미측정]** — 이 저장소에 하한가 지정가 주문의 접수·체결 영수증은 **0 건**이다
  (`verify-execution-capability/measurements.md` 검색 0, flatten-all 실주문 기록 없음). 「호가 단위 위」는 제도와 flatten 주석에 기댄 주장이지 측정이 아니다.
  틀려도 오늘보다 나빠지는 폭은 (b) 의 거부 되먹임이 종목당 하루 1 주기로 묶는다. 실측은 tasks 5.4(사람 결정).
- 2026-08-05 245,750 그리드 불일치 거부(proposal Why)는 KR 정규장 창 안에서 해소된다 — 관측가를 가격으로 쓰지 않는다. 창 밖과 폴백 칸은 오늘과 같으므로 그 칸에서
  그 클래스는 남는다(그리드 정본화는 무소유 Non-goal).

### (b) 하한가 출처 · 캐시 · fail-safe · 기록

**출처 — 이미 있는 것만 쓴다(새 브로커 메서드 · 새 의존성 0)**

- openapi: `GET /api/v1/price-limits`(`openapi.latest.json:567`, `**Rate Limits Group**: MARKET_DATA` `:573`, 쿼리 `symbol` 하나), 스키마 `PriceLimitResponse`
  (`:8547` — `lowerLimitPrice` decimal nullable 「하한가. 미국 주식 등 가격제한이 없는 시장에서는 null」 `:8577`, `timestamp` required 「데이터 시각」).
- Go: `official.Client.PriceLimits`(`internal/official/price_limits_reads.go:31-39`, B1 `:35` GET 오류 → 반환) → `adaptPriceLimits`(`:55-62`, 분기 0) →
  `domain.PriceLimits.LowerLimit float64`(`internal/domain/models.go:116`). null · 빈 값 · 파싱 불가는 `parseDecimal`(`internal/official/market_reads.go:15-24` —
  B1 `:16` 빈 값 → 0, B2 `:20` 파싱 실패 → 0)이 **0** 으로 만든다. 그래서 「0 이하 = 없음」이 유일한 결측 표시다.
- 엔진: 읽기 표면 `OfficialReads.PriceLimits`(`internal/app/engine/reads.go:72`)에 이미 선언돼 있고 비시험 호출자 0 이다. `Context.Official` 이 그 표면이다.
- 응답 `timestamp` 는 `apiPriceLimits` 가 디코드하지만(`price_limits_reads.go:26`) `adaptPriceLimits` 가 버리고 `domain.PriceLimits` 에 칸이 없다. **칸을 더하지 않는다**(9판 정정 → D12.1: 엔진 전용 읽기 형이 timestamp · raw decimal 을 보존한다. CLI 형만 무변경) —
  `domain.PriceLimits` 는 사람 CLI `tossctl quote limits` 출력(`cmd/tossctl/quote.go:200-204` → `internal/output/marketdata.go:57`)에도 쓰여 upstream 출력이 바뀐다(불변식 3).
  ~~신선도는 아래 「언제 읽나」로 대신 세운다.~~ (9판 정정 → D12.1 — 수신 시각만으로는 그날 값을 증명하지 못한다)
- WTS `client.GetPriceLimits` B2(`internal/client/marketdata.go:96-97` 「미국장은 일일 가격제한 제도가 없음」)는 KR 전용이라는 사실의 근거로만 인용한다 — 엔진은 WTS 를 쓰지 않는다.

**캐시 = 엔진 메모리의 하루치 하한가 판(board)** — 새 파일 `internal/app/engine/exit_lower_limit.go`

- 판은 `symbol → {하한가 decimal 문자열, 읽은 날(KST Session.Day), 그날 거부됨}` 의 **불변 스냅숏**이다. 쓰는 쪽(갱신기 · 거부 되먹임)은 잠금 아래 사본을 만들어 원자 포인터를
  바꾸고, 읽는 쪽(가격 결정)은 포인터 한 번 읽기만 한다 — 판정 쪽 잠금 대기 0(7판 「판정 함수의 모양」 요구 승계).
- **만드는 자리 = `NewContext` 한 곳**(반환 리터럴 `engine.go:578` 에 필드 한 줄). `NewOrderPath`(`engine.go:359-406`) · `buildFlattenWiring`(`cmd/tossctl/flatten.go:199-272`) ·
  `internal/app/app.go` · `cmd/tossctl/mcp.go` 에는 0 이다. 0.10 독립 리뷰 P0-1 의 규칙(엔진 전용 상태는 공용 빌더가 아니라 `NewContext` 의 엔진 조립에서만)을 그대로 따른다 —
  8판에서 붙는 것은 `trading` 옵션이 아니라 이 판 하나다. `Context.ExitObserver`(`exitwiring.go:319-364`)가 판을 관측자 옵션에 넣는다 — 호출자 값과 무관하게 덮는
  기존 규칙(`:336-342` 의 `Journal`·`Prices`·`Retrier`·`Issuer`·`Submit` 과 같은 자리).
- **갱신기 = 보조 실행자(AuxiliaryExecutor) 하나.** 선례가 있다 — a136 `Context.ManualBuyRefresher`(`internal/app/engine/manual_buy_refresh.go:62-81`)가 판을 채우는
  보조 실행자를 돌려주고 `engineRuntime` 의 `Auxiliary` 목록(`cmd/tossctl/engine.go:748`)에 들어간다. `manualBuyRefresh.Run` B1(`:85`)처럼 입력(판 · 읽기 · 원장)이 없으면
  아무것도 읽지 않고 취소를 기다린다. 보조 실행자는 감독 루프 밖이다(`cmd/tossctl/engine.go:746-747` 주석 「감독 밖이라 그 결함이 손절 루프를 내리지 않음」) — 갱신기의 결함이
  `exit` 루프(`:735-736` `observer.Run`)를 멈추지 않는다. `ExitObserver.Run`(`exitloop.go:407-417`, B1~B3)은 무변경이다 — 관측 주기에 GET 이 끼지 않는다.
- **언제 읽나 (신선도)**: 갱신기는 `clock.MarketKR.InRegularSession(now)` 이 `(true, nil)` 일 때만 GET 한다 — KRX 정규장 고정 창 안. 정규장이 열린 뒤의 「당일 상한가 및
  하한가」는 그날 값이라는 것이 이 설계의 신선도 근거다(브로커가 정규장 중에 전일 값을 줄 가능성은 `[미측정]` — 되먹임이 묶는다). 항목의 날 = 읽은 시각의 `Session.Day`.
  가격 결정은 **결정 시각의 `Session.Day` 와 같은 날의 항목만** 쓴다 — 날이 바뀌면 판 전체가 저절로 무효다. 장전(NXT 프리마켓 · KRX 장전 동시호가)에 읽지 않는 이유:
  그 시각에 브로커가 오늘 값을 주는지 모르고, 그 구간은 창 밖이라 어차피 쓰지 않는다.
- **무엇을 읽나**: 원장 보유 포지션(`Journal.Positions` — 관측 루프 `exitloop.go:566` 과 같은 로컬 원장 읽기) 중 시장이 KR 이고 그날 항목이 없는 종목. 한 번 채우면 그날 다시 읽지
  않는다(종목당 하루 1 GET + 실패 재시도).
- **얼마나 자주** — **9판 정정 → D12.2**(관측 주기 끝 토큰 · 주기당 최대 1건 · 429 뒤 5분 정지로 대체): ~~주기 30초(상수), 한 주기에 최대 5 종목을 순서대로, 첫 오류에서 그 주기를 멈춘다(429 포함 — 다음 주기에 재시도). `MARKET_DATA` 한도 초당 10회
  (`docs/migration/open-api.md:75`)를 관측 루프 `/prices`(5초에 1회 묶음)와 나눠 쓴다. 상한 5 GET / 30초는 한도의 약 1.7% 다. 다른 `MARKET_DATA` 소비자와 합친 실제 부하는
  `[미측정]`(7판 P3-5 와 같은 처분). 토큰은 같은 `Context.Official` 클라이언트의 캐시를 쓴다 — `--config-dir` 없는 다른 도구가 토큰을 번갈아 무효화한 2026-10-19 사고 경로가 아니다.~~
- 새로 산 포지션(9판: 「30초」 → 「다음 토큰들」): 체결 뒤 최대 30초 + GET 1회 동안 판에 없다 → 그동안의 보호 매도는 오늘과 같은 LIMIT 이다(약화 아님).

**가격 결정 = 총함수** — 같은 새 파일, 모양 `exitLimitPrice(m, proposal, observed, board, now) (price, source string)`

- error 를 돌려주지 않고, 패닉하지 않고, 브로커 · 디스크를 부르지 않는다 — **제출 경로 GET 0**(7판 D7 「즉시성 제약」과 0.10 독립 리뷰 P2-1 의 요구를 그대로 잇는다).
- 하한가를 고르는 조건 — **전부 참일 때만**: ① `isProtective(proposal)`(`exitloop.go:1463-1465`, 술어 무변경 — D1) ② `clock.ParseMarket(m.position.Market)` 이 KR
  ③ `clock.MarketKR.InRegularSession(now)` 이 `(true, nil)`(`clock/market.go:146-152` — B 오류면 거짓) ④ 판에 그 종목의 결정 시각 `Session.Day` 항목이 있고 「그날 거부됨」이 아님
  ⑤ 하한가 > 0 ⑥ 하한가 ≤ 관측가(관측가를 파싱할 수 없으면 거짓). 하나라도 거짓이면 오늘 사다리 — `price = observed`, 비면 `m.state.Baseline`(오늘 `sellIntent` B1 과 같은 규칙),
  `source = observed | baseline`(둘 다 비면 `price = ""` 그대로 넘겨 오늘의 B2 거부를 받는다).
- ⑥ 은 낡은 판을 거르는 값싼 검사다 — 그날 가격제한폭 안의 체결가는 그날 하한가보다 낮을 수 없으므로 하한가 > 관측가면 그 하한가는 그날 것이 아니다. 반대 방향(낡은 하한가가
  그날 하한가보다 낮음 → 범위 밖 거부)은 못 거른다 → 되먹임이 받는다.
- 결정 가격은 `sellIntent` 의 `observed` 인자 자리로 들어간다. 하한가를 고른 경우 그 값은 비지 않고 파싱되는 decimal 문자열이라 `sellIntent` B2·B3·B4 는 오늘과 같은 입력에서만
  닿는다. **그러므로 이 결정이 `submit` B5(`:1533` alert · 해제)에 닿는 새 사유를 만들지 않는다.**
- 결정은 `submit` 머리에서 **한 번** 한다(`applyFloor` 앞 — 수량이 필요 없다). 그 값이 a129 줄과 `sellIntent` 둘 다로 간다(아래 「기록」).

**fail-safe — 결측 · 낡음 · 거부 → 오늘의 LIMIT**

- 판이 없음(기동 직후 · 갱신 실패) · 그날 항목 없음 · 하한가 0(null) · 창 밖 · 시장 판정 불가 · ⑥ 실패 → 오늘의 가격 산출. 결과 `PlaceIntent` 는 오늘과 바이트 같다.
- **거부 되먹임(이 설계의 일부)** — **9판 정정 → D12.3**(B13 만이 아니라 `CONFIRMED` 가 아닌 모든 결과, 재시작 재구성): `submit` 의 기본 거부 갈래 B13(`:1582` default — alert · `ProposalRefused` 해제)에 닿은 주문의 가격 출처가 하한가였으면 그 종목의 그날 항목에
  「그날 거부됨」을 적는다. 해제된 발의는 다음 관측(5초 — `DefaultExitObservationInterval` `exitloop.go:98`)에 다시 발의되고, 그때 ④ 가 거짓이라 오늘의 LIMIT 이 나간다.
  하한가 값이 틀린 모든 경우(낡음 · SOR 경로의 허용 범위 차이 · 그리드 밖 — 전부 `[미측정]`)의 피해가 **종목당 하루 첫 1 주기**로 묶인다. B7(확정) · B8·B9(결과 미상 — 무장 유지) ·
  B10(종목 진행 중 — 취소 해제) · B11(attempt 결과 미기록 — 무장 유지)에서는 적지 않는다 — 거부가 입증되지 않았다.
- 되먹임이 없으면 낡은 하한가가 매 주기 같은 거부를 받아 그날 보호 매도가 0 이 된다(오늘보다 약함). 그래서 되먹임은 빼면 안 되는 갈래이고 RED · 변이 대상이다.

**기록 — 원장 · 로그**

- 원장: `intents.price` 는 **실제로 보낸 가격**(하한가)이다. D4 의 원칙 「원장은 실제로 보낸 것을 적어야 한다」가 그대로 적용된다. schema 변경 없음(`internal/journal/schema.go:6`
  `SchemaVersion = 36` 무변경). 출처 칸을 원장에 더하지 않는다 — schema 를 올려야 하고, 출처는 판정 근거가 아니라 관측이다. `intents.price` 의 Go 소비 4자리(3.7 census —
  `Prepare` 비교 · `workingOrderPrice` · matcher 와일드카드 · `intentNotional`)는 가격이 NULL 이 아닌 보통의 LIMIT 으로 읽는다(오늘과 같은 갈래). 값만 작아진다 — `intentNotional`
  (`execgw/indoubt.go:545-555`)의 매수가능금액 대조 임계(`:490-495`, 0.5×notional)가 낮아져 「부재 확증」을 더 잘 거절하는 쪽(보수)이다.
- 로그: a129 청산 주문 줄(`exitOrderLine` 화이트리스트, `exit_order_log.go:85-100`)에 `price_source` 필드 하나를 더한다(값 `lower_limit` · `observed` · `baseline` —
  `exitOrderTokenPattern` `^[A-Za-z_]{1,64}$` 모양). 줄의 `price` 는 **보낸 가격**으로 맞춘다. 오늘 `exitOrderLineFor`(`:102-114`, B1 `:104`)는 `sellIntent` 의 사다리를 **베껴**
  가격을 다시 계산한다(`:101` 주석 「지정가는 sellIntent 와 같은 규칙(관측가, 비면 기준선)」). 사본을 그대로 두면 줄은 관측가를 적고 브로커는 하한가를 받는다
  ([[transcribed-code-needs-both-sides-pinned]]). 그래서 `exitOrderLineFor` 는 사본 사다리 대신 `submit` 의 결정 값(가격 · 출처)을 받는다(편집 대상). `exitOrderLine.args`
  (`:140-169`, B1 `:161` 체결 줄)에 `"price_source"` 한 쌍을 더하고 `"order_type", "LIMIT"`(`:146`)은 그대로다. `observed_price` · `baseline` 필드는 관측가 · 기준선을 그대로 적는다.
- 생산 배선: a129 줄의 싱크 `ExitOrderLog` 는 생산에서 엔진 로거로 간다(`cmd/tossctl/engine.go:680`).
- 3.8 슬리피지 질의(`analysis/exit-slippage-3.8/exit_slippage.sql` — 관측가 vs 체결 평균가)는 무변경으로 쓴다. 「하한가 지정가는 시장가와 거의 같다」의 착지 뒤 증거가 이 질의다.

### (c) 구간별 거동과 위험 (KR)

사용자에게 한 설명의 확인 결과:

| 설명 | 확인 |
| --- | --- |
| 접속매매에서 하한가 매도 지정가는 상대 최우선 매수 호가부터 체결되므로 시장가와 거의 같다 | **[미측정]** — 가격 우선 체결(매도 지정가는 그 가격 이상의 매수 호가와 높은 호가부터 체결)은 거래소 매매 제도의 일반 원칙이지만 이 저장소에 규정 원문도, 하한가 지정가의 체결 영수증도 없다(M 계열 0 건). flatten 주석(`liquidate.go:375-376`)은 코드 주석이지 실측이 아니다. 확인 수단: 착지 전 tasks 5.4(사람 결정) · 착지 뒤 3.8 질의 |
| 하한가는 「이보다 싸게는 안 판다」는 바닥일 뿐이다 | **부분 확인** — openapi 는 `LIMIT` 을 「지정가」로만 정의한다(`:11178`). 「지정가 매도는 그 가격 미만으로 체결되지 않는다」는 문장은 openapi 에 없다 — 지정가의 정의상 성질이지만 문서 근거는 `[미인용]`. 같은 날 가격제한폭 밖에서는 어떤 주문도 체결될 수 없으므로 이 바닥은 시장가 매도의 바닥과 같다(제도 사실 — 저장소 근거 `[미인용]`) |
| 동시호가 구간에서는 단일가로 체결된다 | **부분 확인** — openapi 달력 스키마가 단일가 구간의 **존재**를 적는다(정규장 안 종가단일가 「KRX 기준」 `:9808`, 프리마켓 시가단일가 `:9773`). 단일가 **체결 방식**은 openapi 에 없다 — `[미측정]` |

| 구간 | 8판 주문 | 거동 · 위험 |
| --- | --- | --- |
| 장전 — NXT 프리마켓 · KRX 장전 동시호가 · 장전 시간외 (09:00 전) | 오늘 LIMIT(창 밖) | 무변화 |
| 정규장 접속매매(09:00~15:20, 평상일) | 하한가 LIMIT | 매수 호가를 위에서부터 쓸어 내린다. 얇은 호가에서는 관측가보다 훨씬 낮게 체결될 수 있다 — 시장가와 같은 성질(D8 「하한가 도달」 행과 별개의 잔존 위험). `[미측정]` |
| 정규장 종가단일가(15:20~15:30) | 하한가 LIMIT | 단일가 경매에 가장 공격적인 매도로 들어간다 — 제도대로라면 경매 가격(≥ 하한가)으로 체결된다(9판 정정: 단정 → 조건형 · `[미측정]`). 오늘의 관측가 LIMIT 은 경매 가격이 관측가보다 낮으면 미체결이므로 8판이 더 확실히 나간다. 큰 수량은 경매 가격을 끌어내릴 수 있다. `[미측정]`. 7판이 이 구간을 LIMIT 으로 뺀 이유(MARKET 접수 미측정)는 8판에 없다 — 유형이 이미 LIMIT 이다 |
| VI 단일가(장중) | 하한가 LIMIT | VI 중에도 가격제한폭 안의 지정가는 받는다는 것은 제도 일반 원칙 — `[미측정]`. 거부되면 되먹임으로 다음 주기 오늘 LIMIT. 7판 P1-1 의 「VI 를 가를 입력이 없다」는 8판에서 판정 입력이 아니다 — 같은 LIMIT 유형이라 VI 여부로 갈래를 나눌 필요가 없다 |
| 개장 지연일(예: 연초 첫 거래일 · 수능일) 09:00~실제 개장 | 하한가 LIMIT(고정 창은 09:00 부터) | KRX 는 장전 동시호가 중이고 NXT 는 열려 있을 수 있다. 하한가가 그날 범위 안이면 접수돼 경매 · 접속에 참여할 수 있다 — `[미측정]`(9판 정정: 단정 → 조건형). 거부되면 되먹임. 개장이 늦어 마감도 늦은 날의 15:30 뒤는 창 밖이라 오늘 LIMIT(이득을 덜 얻을 뿐, 약화 아님) |
| 정규장 후 — NXT 애프터마켓 · KRX 장후 시간외 · 시간외단일가(15:30 뒤) | 오늘 LIMIT(창 밖) | 무변화. **창을 15:30 에서 자르는 이유**: KRX 시간외단일가의 가격 범위는 그날 가격제한폭보다 좁다(제도 — 저장소 근거 `[미인용]`). 거기에 하한가를 내면 openapi 「주문 가격 허용 범위 초과」(`:5800-5805`)로 거부될 수 있다 — 그 칸은 오늘보다 약해진다 |
| 주말 | 오늘 LIMIT(창 밖 — `Session.Contains` 주말 거름 `clock/market.go:114`) | 무변화 |
| 평일 휴장일 | 판에 그날 값이 있으면 하한가 LIMIT | 휴장이면 가격과 무관하게 접수 · 체결될 수 없다(openapi 「현재 해당 주문을 접수할 수 없는 시간입니다」 `:5777-5782` 류) — 오늘과 같은 결과. 갱신기가 휴장일 GET 에 무엇을 받는지 `[미측정]`. 거부되면 되먹임 → 다음 주기 오늘 LIMIT(역시 접수 불가 — 오늘과 같다) |
| 종목 거래정지 · 시장 전체 정지(서킷브레이커 · 사이드카) | 하한가 LIMIT | 가격과 무관한 거부 · 대기 — 오늘과 같다. 거부되면 되먹임 |
| 하한가 도달 · 매수 호가 공백 | 하한가 LIMIT | 체결 안 됨 — 시장가 · 오늘 LIMIT 과 같다(D8). 잔량은 a167 |

**NXT · SOR**: KR 주문은 투자자지시 거래소가 통합(SOR)이어야 한다(openapi 422 예시 「투자자지시 거래소가 통합(SOR) 으로 설정되어 있어야 주문할 수 있습니다」 `:5840-5845`).
창 안의 하한가 지정가가 KRX · NXT 중 어디서 체결되는지, `/price-limits` 의 하한가가 NXT 의 허용 범위와 같은지는 `[미측정]` 이다. 다르면 그 주문은 거부되고 되먹임이 1 주기로 묶는다.

### (d) 적용 범위 — KR 만

`lowerLimitPrice` 는 「미국 주식 등 가격제한이 없는 시장에서는 null」(`openapi.latest.json:8577`, 예시 `usStock` `upperLimitPrice: null`)이고, `apiPriceLimits` 주석도
「US stocks have no daily price limits」(`price_limits_reads.go:19-20`)다. 그래서 US 보호 매도는 오늘 그대로(관측가 LIMIT)이고 갱신기는 US 종목을 읽지 않는다(조건 ②).
7판 Q2(US 정규장 LIMIT)와 결과가 같다.

### (e) 불변식 판정

- **불변식 3 (엔진 전용 · 토글 OFF = upstream)** — **PASS.** ②`checkOrderShape` · ③`placeIntentSupported` · `trading.Service` · `NewOrderPath` · flatten · CLI · ops · MCP 는
  **바이트 그대로**다 — 유형이 LIMIT 이라 관문을 열 일이 없다. 엔진 전용 상태는 하한가 판 하나이고 `NewContext` 에서만 만든다(P0-1 규칙 유지). 갱신기는 `engineRuntime` 의
  보조 실행자로만 산다. 엔진이 꺼지면 `NewContext` 도 `engineRuntime` 도 없으므로 흔적 0.
- **불변식 4 (손절 즉시성 — 오늘보다 약해지지 않음)** — **9판 정정 → D12.7 표가 대체한다**(8판의 「나빠질 수 있는 칸은 하나」는 재리뷰 P0-2 · P0-3 으로 반증됨). 아래는 8판 기록:
  - 제출 경로 GET 0(가격 결정은 메모리 읽기) · 결정 실패 = 오늘 LIMIT(거부 아님) · 새 B5 사유 0 · 갱신기는 관측 주기 밖(보조 실행자, `Run` 무변경).
  - 오늘보다 나빠질 수 있는 칸은 하나다 — **하한가 가격이 브로커에 거부되는 경우**(낡음 · SOR 범위 차이 · 개장 지연 · VI 중 거동 — 전부 `[미측정]`). 피해는 그 종목 그날 첫 보호
    매도가 1 주기(5초) 늦는 것이고 되먹임이 그 뒤를 오늘과 같게 만든다. 이 1 주기를 수락할지는 **사람 확인**(`issues.md` D-2 ①). tasks 5.4 실측이 그 칸의 확률을 잰다.
  - 나아지는 칸: 창 안의 그리드 밖 관측가(245,750 클래스) 거부 소멸, 단일가 구간 미체결 감소.
- **불변식 6 (손절·익절·사이징은 보수 방향만)** — **PASS(근거 명시).** 손절 수준 · 수량 · 발의 조건은 무변경이다. 바뀌는 것은 보호 매도의 지정가를 더 체결 쪽으로 낮추는
  것뿐이고, 보호가 실제로 나가는 확률이 오르는 방향이다. 대가는 얇은 호가에서의 체결가 하락(시장가와 같은 성질)이며 그 축은 2026-10-10 시장가 결정과 2026-10-28 D-1 (다)
  결정이 받아들였다. 익절은 무변경(§0.9).

### (f) 정수 수량(0.10 독립 리뷰 P1-2) — 8판에서 필요 없음

P1-2 의 문제는 「MARKET 자격」이었다 — openapi 는 소수 수량을 US MARKET SELL 에만 허용하고 그 밖(국내 · 지정가 포함)은 `400 invalid-request` 다(`:11197`). 8판은 유형을
바꾸지 않으므로 KR 보호 매도의 수량 경로는 오늘의 LIMIT 과 바이트 같다 — 소수 KR 수량이 있다면 오늘도 400 이고 8판도 400 이다. 새 약화가 아니다. 그래서 7판의 판정 조건 6 ·
②③ 정수 backstop · `canonicalProtectionQuantity` 내보내기는 **하지 않는다**(D3 「0.10 독립 리뷰 반영」 P1-2 는 8판에서 대체됨). KR 포지션 수량의 정수 불변(입구 증명)은
원장 투영 소관이라 이 change 범위 밖으로 남는다(0.10 독립 리뷰 「착수 전 처분」 마지막 문장의 그 부분 — 오늘 동작과 같으므로 차단 아님).

### (g) a165 와의 충돌 재대조 (8판 작성 시점 문서 대조 — a165 3판 `bc56522c7`, origin/main `4b67dd6a9` 에 a165 코드 없음. 9판 정정 → D12.5: a165 는 지금 착지 중이고 재대조는 tasks 1.0)

| 표면 | a087 8판 | a165 3판 | 판정 |
| --- | --- | --- | --- |
| `internal/app/engine/exitloop.go` | `submit` 본문(결정 한 번 · 되먹임) · `ExitObserverOptions` 필드 하나 | `clearTheSymbol` FLM 번들(인용 `:1686` · `:1639`, 본문 편집 문장 없음) | 같은 파일 · 다른 함수. a165 가 이 파일을 바꾸면 좌표 · AST 가 낡는다 |
| `internal/execgw` | **0**(8판에서 `failclosed.go` 편집 소멸) | `gateway.go` dispatch 클로저 · 분류 함수 번들 | 겹침 0 |
| `internal/trading` · `engine.NewOrderPath` | **0** | `trading.Service.Cancel` 번들 | 겹침 0 |
| `cmd/tossctl` | `engine.go` `engineRuntime` 의 `Auxiliary` 목록 한 줄 | `flatten.go` `buildFlattenWiring` | 다른 파일. 7판 P2-3 의 「flatten 경계」 의존은 **소멸** — 8판은 flatten 경로에 아무것도 붙이지 않는다 |
| `internal/flatten` | 0 — 판을 공유하지 않는다(flatten 은 제출 시점 `PriceLimits` GET 을 그대로 쓴다) | saga 취소 409 번들 | 겹침 0 |
| 행동 | 보호 매도 가격 | 청산 전 취소 409 → 다음 주기 재취소 | 독립. 하한가 지정가가 하한가 도달로 미체결로 남으면 다음 발의 전 `clearTheSymbol` 이 그 주문을 취소한다 — a165 의 409 처리가 그 취소를 돕는다 |

a165 착지 뒤 다시 할 일(8판): ① `enumerate.py` stale 확인 → stale 인 AST 재생성 · 좌표 재기록 ② `clearTheSymbol` 시험과 a087 종단 시험(1.6)이 같은 시험 더블을 쓰면 기대값 대조
③ gate ⑤ 창의 형제 함수(Manager 처분 P2-3 — archive 번들 복사). 7판 「a165 착지 뒤 할 일」 6(flatten 경계 재대조)은 8판에서 대상이 없다.

a167(잔량 재가격) 교차: 창 안의 첫 제출이 이미 하한가라 그 잔량을 더 낮게 다시 매길 가격이 없다. a167 의 KR 지정가 구간 규칙(「매수 1호가 / 하한가」)은 이 판을 재사용할 수 있다.
a167 문서 반영은 이 디렉터리 범위 밖 — `issues.md` D-2 ④.

### 착수 시 FLM 대상 (8판 — gate ⑤, 본문을 고치는 기존 함수)

`ExitObserver.submit`(`exitloop.go:1483-1596`) · `exitOrderLineFor`(`exit_order_log.go:102-114`) · `exitOrderLine.args`(`:140-169`) · `NewContext`(`engine.go:437-607`) ·
`Context.ExitObserver`(`exitwiring.go:319-364`) · `engineRuntime`(`cmd/tossctl/engine.go`) · 바뀌는 기존 시험 함수(시험 파일 포함). **9판 추가**: `ExitObserver.Run`(`exitloop.go:407-417` — 주기 끝 토큰 · 429 정지 알림, D12.2). **10판**: `Run` 은 토큰 한 줄만(429 알림 대체), **`official.Client.doRequest`(`client.go:191-207`) 추가**(D13.2 — a166 · a100 과 같은 함수), **(12판: `doRequest` 는 다시 뺀다 — D15.1, a166 소유)** T3 원장 읽기 → 「그날 보호 청산 매도 intent 존재」 읽기(D13.1), T2 갈래 삭제. 새 파일의 새 함수: 엔진 전용 하한가 읽기(`internal/official`) · T3 원장 읽기(`internal/journal`, D12.3). T2 를 고르면 `Gateway.Place` · `Gateway.submit` · `Attempt.MarkDispatchStarted`(또는 `Prepare`)가 더해진다. HEAD AST 는 `analysis/evidence/` 에 있다 —
구현 때 그 HEAD 에서 번들(`ast.json` + `function-logic-map.md` + `branch-test-map.md` + `risk-pattern-report.md`)로 다시 만든다. 새 함수(판 · 갱신기 · `exitLimitPrice`)는 새 파일이다.
**`sellIntent` · `checkOrderShape` · `placeIntentSupported` · `Service.PreviewPlace` · `Service.Place` · `NewOrderPath` · `buildFlattenWiring` 은 8판에서 대상이 아니다**(본문 무변경 —
시험이 읽기만 한다). 7판 목록(`review.md` 「착수 시 FLM 대상」·「착수 시 FLM 대상 정정」)은 8판에서 대체됨.

## D12 — 9판: 8판 독립 재리뷰 blocker 처분 (Manager 결정 2026-10-29)

> 8판 독립 재리뷰(`review.md` 「8판 독립 재리뷰」, `5e691ac4e`)는 FAIL(P0 3 · P1 3 · P2 1)이었고 Manager 가 7건 모두 타당하다고 판정했다. 이 절이 D11 의 해당 문단을
> **대체**한다. D11 본문에는 대체된 자리마다 「9판 정정 → D12.x」를 달았다. 새 분기 근거는 HEAD `5e691ac4e`(소스 바이트는 origin/main `4b67dd6a9` 와 같다)의 AST 42개
> (`analysis/evidence/`, `enumerate.py` stale 0) — 이 판에서 6개를 더 만들었다: `engine--exitobserver.observeonce` · `engine--exitobserver.observe` ·
> `engine--exitobserver.noteheldproposal` · `execgw--retrier.query` · `official--client.dorequest` · `journal--attempt.transition`.

### D12.1 — 응답 `timestamp` 와 raw decimal 을 보존하는 엔진 전용 읽기 (P0-1 · P1-1)

- **새 읽기 형**(새 파일 `internal/official/a087_price_limits_raw.go`): 같은 `GET /api/v1/price-limits` 를 부르되 응답을 `lowerLimitPrice *string`(null 과 빈 값을 가른다) ·
  `timestamp string` 그대로 디코드해 돌려준다. 기존 `Client.PriceLimits` · `adaptPriceLimits` · `domain.PriceLimits` · CLI `tossctl quote limits` 출력은 **바이트 그대로**다
  (D11 (b) 의 「칸을 더하지 않는다」는 CLI 형에 대해서만 참으로 남는다). 엔진은 `Context.Official` 이 이 메서드를 가질 때만 쓴다(좁은 인터페이스 단언 — `OfficialReads`
  인터페이스는 무변경이라 그것을 구현한 시험 더블이 깨지지 않는다. 없으면 갱신기는 아무것도 읽지 않는다 — `manualBuyRefresh.Run` B1 `:85` 와 같은 모양).
- **쓰는 조건(Manager 결정)** — 둘 다 참일 때만 판 항목이 된다(10판 정정 → D13.3: 이 조건은 신선도 증명이 아니고, 「KST 날짜」 일치는 평일 휴장 · 기준가 변경을 증명하지 않는다 — 5.4 gate 로 보류): ① `timestamp` 를 RFC 3339 로 파싱한 시각의 **KST 날짜 = 결정 시각의 KRX 거래일**(KST `Session.Day` —
  창이 평일만 열리므로 평일 날짜) ② 그 시각이 **그날 09:00 KST 이후**.
- **거르는 것 — 사유를 따로 남긴다**(한 사유가 다른 사유를 덮지 않게, P1-1): `lower_limit_null` · `lower_limit_unparsable` · `lower_limit_not_positive`(0 이하) ·
  `lower_limit_not_integral`(KR 가격은 원 단위 정수 — openapi 주문 `price` 「KR: 정수 (원 단위)」 `:11205`) · `timestamp_missing` · `timestamp_unparsable` ·
  `timestamp_other_day` · `timestamp_before_open` · `fetch_failed` · `rate_paused`. 기록은 갱신기의 엔진 로거 줄(`c.Log` — a136 갱신기와 같은 싱크) 한 줄씩, (종목 · 날 · 사유)마다
  한 번만(래치). 계좌 필드는 싣지 않는다.
- **raw decimal 의 정본**: 판에는 검증을 통과한 원 문자열(양의 정수)을 그대로 둔다. 그 문자열은 `sellIntent` B3·B4 의 `floatOf`(`exitloop.go:2014-2020`, B1 은 `ParseFloat`
  실패뿐)를 반드시 지나고, 2^53 미만 정수라 `float64` 왕복이 정확하다(주식 가격 범위). 그래서 하한가 행에서 `submit` B5 는 **도달 불가**다 — 이것을 시험 행(「판 값 → B5 미도달」)으로
  고정한다. 정수가 아니면 위 사유로 거른다(호가 단위 계산을 우리가 하지 않는다).
- **「데이터 시각」의 뜻은 `[미측정]`** 이다(openapi 스키마는 「데이터 시각」 한 마디 — `:8547` 근처 `PriceLimitResponse.timestamp`). 하한가가 효력을 갖는 날짜인지, 응답을 만든 시각인지
  모른다. 응답 시각이라면 조건 ①② 는 「정규장 중에 받았다」만 말하고 「그날 하한가다」를 증명하지 못한다. 그래서 tasks 5.4 에 「timestamp 가 무엇을 가리키는지」를 넣었다(D12.4).
- **전일 값이 들어와도 과매도 경로는 없다 — 논증.** 아래 두 문장의 제도 근거(가격제한폭 밖 호가 불가 · 체결가는 그날 범위 안)는 이 저장소에 규정 원문이 없다 — `[미인용]`.
  - 판 항목은 D11 조건 ⑥(하한가 ≤ 결정 시점 관측가)을 지나야 쓰인다. 관측가는 그날 체결된 가격(장 시작 전이면 전일 종가 = 그날 기준가)이므로 그날 상한가 이하다 → 판 값도
    그날 **상한가 이하**다. 위쪽으로 범위를 벗어나는 경로는 없다.
  - 판 값이 그날 하한가보다 **낮으면**(전일 하한가가 더 낮은 날) 그날 가격 범위 밖이라 브로커가 거부할 것이다(openapi 「주문 가격 허용 범위 초과」 `:5800-5805` — 실제 응답은
    `[미측정]`). 그 거부는 D12.3 의 되먹임이 받아 그 종목은 그날 오늘의 LIMIT 으로 돌아간다.
  - 판 값이 그날 하한가보다 **높으면** 범위 안의 더 높은 바닥이 될 뿐이다. 그 바닥은 관측가 이하라 오늘의 LIMIT(바닥 = 관측가)보다 공격적이거나 같다. 더 싸게 파는 경로가
    생기지 않으므로 과매도로 이어지지 않는다.

### D12.2 — `MARKET_DATA` 예산 직렬화 (P0-2)

**사실(AST · 손 읽기 표시).** 관측 루프의 가격 읽기는 `ObserveOnce` B4(`exitloop.go:504` — `observe` 오류면 `checkOutage` 뒤 그 주기 반환 `:507`)에서 실패하면 그 주기의 판정 · 제출
전체를 건너뛴다. `observe`(`:860-931`)는 `Retrier.Query` 로 감싸고, `Retrier.Query`(`internal/execgw/retry.go:335-389`)는 B10 `ClassRateLimited` 에서 Retry-After 를 기다려 다시
부른다(B11 상한 초과면 반환, B12 시한 초과면 반환). `Client.doRequest`(`internal/official/client.go:191-207`)는 응답 헤더를 기록만 하고(`:199` `c.rates.record` — 손 읽기, 분기 아님)
요청을 예약 · 제한하지 않는다. 그러므로 같은 `Context.Official` 의 다른 `MARKET_DATA` 요청은 `/prices` 의 429 확률을 올리고, 그 429 는 오늘도 있는 「주기 건너뛰기 · 재시도 대기」
경로로 손절 판정을 늦춘다.

**원칙(Manager 결정)**: (1) 하한가 GET 은 관측 주기의 `/prices` 와 **시간상 겹치지 않는다** — 그 주기의 판정 · 제출이 끝난 뒤에만 돈다. (2) **관측 주기당 최대 1건**. (3) 어디서든
429 · rate limited 를 보면 **5분** 동안 갱신을 멈춘다.

**구현 수단 비교 — 가) 채택.**

| 안 | 편집하는 기존 함수 | 지연 | 판정 |
| --- | --- | --- | --- |
| 가) 보조 실행자 + 관측자가 주기 끝에 내주는 토큰 | `ExitObserver.Run`(`exitloop.go:407-417`, 분기 3) 한 줄 — `o.reportCycle(o.ObserveOnce(ctx))` 뒤 · `Sleep` 앞에서 판에 「주기 끝」을 알림(채널 용량 1, 막히지 않는 보내기) | 관측 루프에 0 — GET 은 보조 실행자가 `Sleep` 동안 한다 | **채택** — High-risk 편집이 분기 3개 함수 하나 |
| 나) `ObserveOnce` 끝에서 직접 호출 | `ObserveOnce`(`:469-541`, 분기 8 · 반환 5 — 모든 반환 경로를 덮으려면 defer, 이미 defer 둘 `:471`·`:474`) | GET 시간이 다음 주기 시작을 그만큼 늦춘다(`Run` 은 `ObserveOnce` 반환 뒤에 `Sleep`) — 손절 판정 지연 | 기각 — 불변식 4 에 새 지연을 더한다 |

- **겹침 0 의 근거**: 토큰은 `ObserveOnce` 가 돌아온 뒤에만 나온다(그 주기의 `/prices` · 재시도 대기 · 판정 · 제출 · a129 defer 가 모두 끝남). 갱신기는 토큰 하나에 GET 하나를 하고, 그 GET 에
  **시한 = min(관측 주기/2, 2초)** 를 건다. 다음 `/prices` 는 `Sleep(Interval)`(기본 5초 — `DefaultExitObservationInterval` `:98`) 뒤에 시작하므로, 시한 안에 끝나거나 취소된 GET 은
  다음 `/prices` 와 겹치지 않는다. 토큰이 없으면(tracer 가 `ObserveOnce` 를 직접 돌림 · 관측자 없음) GET 0.
- **주기당 1건**: 채널 용량 1 + 토큰당 GET 1. 밀린 토큰은 쌓이지 않는다(막히지 않는 보내기는 꽉 차 있으면 버린다).
- **5분 정지 — 세 입력**(**10판에서 대체됨 → D13.2**: `doRequest` 의 HTTP 429 시각 하나가 셋을 대신한다): ① 갱신기 자신의 GET 이 `official.ErrRateLimited` ② `Run` 이 받은 주기 결과 `cycle.Err` 가 `errors.Is(…, official.ErrRateLimited)`(관측 `/prices` 의 429 —
  `Retrier.Query` B11·B12 반환은 `%w` 또는 원 오류라 사슬이 남는다는 것은 손 읽기, 구현 시 시험으로 고정) ③ 같은 클라이언트의 `RateBudgets()`(`internal/official/ratebudget.go:220` —
  모든 경로의 마지막 허용량, 429 응답 포함 `:224` 주석)에 `ObservedAt` 이 5분 안이고 `Exhausted()` 인 항목이 있으면(다른 소비자의 429 — 헤더가 없는 429 는 `Reported=false` 라 못 본다:
  `[미측정]` 잔여). 정지 중에는 판에 새 항목이 생기지 않는다 → 오늘의 LIMIT(약화 아님).
- **새 압력의 수치**(**11판 정정 → D14.3**: 기준은 그룹 10/s 가 아니라 계좌 공유 3 req/s — 주기당 ≤ 1 ≈ 6.7%, 401 시 ≤ 3): 관측 주기 5초 · 그룹 한도 초당 10회(`docs/migration/open-api.md:75`)에서 갱신기의 추가는 **주기당 최대 1건 = 초당 0.2건 = 한도의 2%** 이고, 그 1건은 자기 관측
  주기의 `/prices` 와 다른 초에 있다(겹침 0 · 시한 ≤ 2초 < 5초). `/prices` 429 로 주기를 건너뛰는 것은 **오늘도 있는 동작**이다 — 새로 더해지는 것은 「같은 초에 다른 소비자가 이미
  한도 근처일 때 그 1건이 429 를 앞당길 가능성」뿐이고, 그 크기는 다른 소비자 부하에 달려 `[미측정]` 이다. 종목당 하루 1 GET 이라 보유 종목 N 개면 그날 GET 은 N + 재시도이고,
  새로 산 종목은 다음 토큰부터 채워진다(그 사이 오늘 LIMIT).
- **시험이 고정할 것**: 가짜 transport 가 요청 시작 · 끝을 기록 → **관측 `/prices` 가 진행 중인 모든 순간에 하한가 요청 동시 수 = 0** · **한 주기 안의 `MARKET_DATA` 요청 수 ≤ 오늘의 수 + 1**
  (오늘의 수 = 같은 시나리오를 판 없이 돌린 계수) · 429 뒤 5분 동안 하한가 GET 0(가짜 시계) · `/prices` 429 주기 뒤 5분 정지 · 토큰 없으면 GET 0.
- D11 (b) 「얼마나 자주」(30초 · 주기당 5)와 「갱신기 = 보조 실행자」의 독립 주기는 이것으로 **대체**된다. 보조 실행자 등록(`engineRuntime` `Auxiliary`)과 감독 밖 성질은 그대로다.

### D12.3 — 비동기 비수용까지 덮는 되먹임 · 재시작 재구성 (P0-3) — **10판에서 대체됨 → D13.1**(첫 시도 한정 · T2/T3 · D-2 ⑤ 대체)

**규칙(Manager 결정)**: 그 종목에서 그날 하한가로 낸 시도가 **없거나**, 가장 최근의 하한가 시도가 **브로커 수용으로 확정**(`CONFIRMED`)된 경우에만 하한가를 쓴다. 그 밖 — 직접 거부(B13),
IN_DOUBT, 확정 비수용(`FAILED_CONFIRMED`), `NOT_DISPATCHED`, 종목 진행 중(B10), 미상 — 은 전부 **그날 남은 시간 동안 오늘의 LIMIT**.

- **왜 B13 만으로 모자랐나(재리뷰 P0-3, AST)**: `submit` B8(`:1557` IN_DOUBT · UNRESOLVED)은 발의를 무장한 채 두고, 뒤에 원장이 비수용으로 확정하면 `noteHeldProposal` B4
  (`exit_held_proposal.go:140` `facts.Verdict == journal.ExitIntentUnaccepted`)가 발의를 푼다. 8판 되먹임은 그 갈래를 보지 않아 다음 관측이 같은 하한가를 다시 냈다.
- **9판 처리 — 제출 순간에 정한다**: `submit` 이 하한가 출처 주문의 `Place` 결과를 받으면 **결과가 `CONFIRMED` 가 아닌 모든 갈래**(B8 · B10 · B11 · B13)에서 그 종목의 그날 항목을
  「금지」로 바꾼다. IN_DOUBT 는 그 순간 이미 「수용 확정」이 아니므로 금지가 먼저 서고, 그 뒤 비수용 확정 → `noteHeldProposal` B4 해제 → 다음 관측의 재발의는 오늘의 LIMIT 이다.
  그래서 `noteHeldProposal` 은 **편집하지 않는다**. 금지는 그날 풀리지 않는다(뒤에 수용으로 확정돼도 — 오늘 LIMIT 으로 기우는 쪽이라 약화 아님).
- **판정 술어는 하나**: 「최근 하한가 시도의 상태 → 허용/금지」를 함수 하나(`lowerLimitAllowedAfter(state)`)로 두고, 제출 순간의 메모리 갱신과 재시작 재구성이 같은 함수를 부른다
  ([[two-judgements-cover-for-each-other]]). 두 입력 경로를 각각 시험한다.
- **가격 출처의 원장 기록 — Manager 결정(attempt 꼬리표)과 9판 수단.** Manager 는 「attempt 에 꼬리표, schema v36, a165 의 detail 꼬리표 방식」을 정했다. 대조 결과
  `mutation_attempts.detail` 은 **뒤 전이가 비지 않은 detail 을 주면 덮어쓴다**(`Attempt.transition` B16 `internal/journal/durability.go:627` `if o.detail != ""` → `detail = ?`).
  하한가 시도가 IN_DOUBT 로 가거나 거부 사유를 받는 바로 그 전이에서 꼬리표가 사라진다. 같은 목적(원장에 남고, schema 무변경, 재시작에 재구성)을 지키는 수단 둘:

  | 수단 | 쓰는 곳 | 편집 | a165 와 | 판정 |
  | --- | --- | --- | --- | --- |
  | T2 꼬리표 — 첫 전이의 append-only 행 | `attempt_transitions.detail`(`durability.go:650-652`, 행은 갱신 · 삭제 없음 — schema 주석) 의 PLACE 첫 전이에 `[a087 price_source=lower_limit]` | `execgw.PlaceRequest` 필드 · `Gateway.Place`/`submit` 이 그 값을 넘김 · `Attempt.MarkDispatchStarted`(또는 `Prepare`) 서명 — High-risk 3 함수 | **`Gateway.submit` 은 a165 의 FLM 번들**(`internal-execgw--gateway.submit`) — 같은 함수 | 가능, 비쌈 |
  | T3 유도 — 이미 있는 칸으로 | 쓰지 않는다. 읽기: 그날 `intents.trading_day`(`schema.go:206`, 색인 `(symbol, trading_day)` `:220`)의 그 종목 보호 청산 매도 intent 중 **`price` 가 판의 그날 하한가와 같은 decimal** 인 것 → 그 intent 의 최신 PLACE attempt 상태 | 새 파일의 읽기 전용 원장 함수 1개, 기존 함수 편집 0 | 겹침 0 | **9판 문서의 기본** |

  T3 의 우연 일치(그날 관측가 = 하한가인 하한가 도달 상황의 오늘 LIMIT 시도)는 하한가 시도로 세어진다 — 그 시도가 비수용이면 금지가 서고, 돌아갈 오늘 LIMIT 가격이 같은 값이라
  결과가 같다. **Manager 결정(꼬리표)에서 수단을 T3 로 바꾸는 것은 새 Manager 결정**이다(`issues.md` D-2 ⑤). 결정이 T2 면 tasks 2.5 의 편집 목록에 위 3 함수를 더하고 a165 착지 뒤에만 착수한다.
- **재시작 재구성**: 갱신기가 한 종목의 그날 항목을 처음 세울 때(D12.1 검증 통과 직후, 판에 내기 **전**) T3 읽기로 그날 하한가 시도의 최신 상태를 구해 같은 술어로 허용/금지를 정한다.
  재구성이 끝나기 전에는 항목이 없으므로 오늘 LIMIT 이다. 원장 읽기 실패는 항목을 세우지 않는다(사유 `ledger_read_failed`) → 오늘 LIMIT. 그래서 재시작이 거부를 한 번 더 만들지 않는다.
- a129 줄의 `price_source`(D11 (b))는 관측용으로 그대로 남는다 — 판정 입력이 아니다.

### D12.4 — 단정 문장과 5.4 착지 전 gate (P1-2)

- D11 (c) 의 「체결될 것이다」·「접수돼 … 참여할 것이다」는 조건형으로 고쳤다(그 표에 「9판 정정」 표시).
- **tasks 5.4 = 착지(운영 배포) 전 필수 gate.** 재는 것: ① 하한가 지정가 매도 1주의 접수 · 상태 · 체결 평균가 · 같은 시각 매수 1호가 ② **응답 `timestamp` 가 무엇을 가리키는지** —
  같은 종목을 장전(09:00 전)과 장중에 읽어 하한가 값과 timestamp 가 각각 어떻게 바뀌는지 기록(엔진 전용 읽기 형으로 — 읽기 전용, 주문 없음). 처분 후보(사람 결정, `issues.md` D-2 ②):
  **수용 → 진행 · 거부 → 범위를 KR off 로 축소(보호 매도 전부 오늘 LIMIT — 사실상 (나)) · 판정 불가 → 착지 보류.**

### D12.5 — a165 (P1-3)

D11 (g) 의 「a165 코드 0, 미착지」는 **8판 작성 시점(origin/main `4b67dd6a9`)의 사실**이었고, 지금 a165 는 main 에 착지 중이다(`clearTheSymbol` · `checkSymbolFree` ·
`cancel_refusal*.go` — Manager 전달, 이 worktree 는 아직 그 커밋을 보지 않았다). 그래서 D11 (g) 표는 「착지 전 문서 대조」로 낮춘다. **a165 archive · push 가 확인된 sha 에서**
`clearTheSymbol` · `checkSymbolFree` 를 포함해 a087 이 인용 · 편집하는 `exitloop.go` 함수의 AST · FLM · 공유 시험 더블을 다시 대조하는 것을 tasks **1.0 「구현 착수 조건」** 으로 둔다.
T2 를 고르면 `Gateway.submit` 도 그 대조에 들어간다.

### D12.6 — `price_source` 골든 (P2)

a129 골든 helper 는 기대 키만 부분 비교한다(재리뷰 — `a129_exit_order_log_test.go:204-211`, 손 읽기). 그래서 tasks 3.1 의 골든은 `price_source` 를 **필수 키**로 두고 값
`lower_limit` · `observed` · `baseline` 셋을 각각 기대한다. 행: B1(하한 산술 실패 — 결정은 이미 됐으므로 출처가 실림) · B5(`sellIntent` 실패 — 하한가 행에서는 도달 불가이므로
관측가 · 기준선 행만) · B13(직접 거부) · B7(확정) · B8 → 비수용 확정(비동기 결과 — 체결 따라가기 줄까지) · 폴백 행. 소비자 0(재리뷰 검색)은 구현 시 다시 센다.

### D12.7 — 불변식 4 표 (9판) — **10판에서 대체됨 → D13.5**

오늘(모든 보호 매도 = 관측가 LIMIT)보다 나빠질 수 있는 칸 전부와 각 칸의 상한:

| # | 칸 | 무엇이 나빠지나 | 상한 | 근거 · 남은 `[미측정]` |
| --- | --- | --- | --- | --- |
| 1 | 하한가 가격이 브로커에 **직접 거부**(전일 값 < 그날 하한가 · SOR/NXT 범위 차이 · VI · 개장 지연 거동) | 그 종목 그날 첫 보호 매도가 거부 → 해제 → 다음 관측에서 오늘 LIMIT | **종목당 하루 1회 × (거부 응답 시간 + 관측 주기 5초)** | D12.3 금지(B13). 거부가 실제로 나는지 · 응답 시간은 `[미측정]` → 5.4 |
| 2 | 하한가 시도가 **IN_DOUBT → 비수용 확정** | 오늘 LIMIT 시도의 IN_DOUBT 와 같은 해소 대기 + 해제 뒤 1 주기 | **종목당 하루 1회 × (해소 시간 + 5초)** — 해소 시간 자체는 오늘의 IN_DOUBT 와 같은 경로 | D12.3(B8 에서 금지 먼저). 하한가가 IN_DOUBT 확률을 바꾸는지 `[미측정]` |
| 3 | 재시작 | 재구성 전 항목 없음 → 오늘 LIMIT | **추가 거부 0**(재구성은 항목을 내기 전) | D12.3. 원장 읽기 실패 → 오늘 LIMIT |
| 4 | `MARKET_DATA` 압력 → 관측 `/prices` 429 → 주기 건너뛰기 · 재시도 대기(오늘도 있는 동작) | 그 대기가 더 자주 일어날 가능성 | 추가 요청 **주기당 ≤ 1 = 0.2/s(한도 10/s 의 2%)**, 자기 `/prices` 와 겹침 0, 429 를 보면 5분 0건 | D12.2. 다른 소비자와 같은 초에 겹칠 때의 효과 `[미측정]` · 헤더 없는 429 는 ③ 입력에 안 보임 |
| 5 | 판 · 갱신기 고장 · 원장 읽기 실패 · 정지 | — | **0** — 항목이 없으면 오늘 LIMIT | D11 (b) fail-safe |
| 6 | 전일 하한가가 그날 하한가보다 높음(⑥ 통과) | — | **0** — 바닥이 관측가 이하라 오늘보다 공격적이거나 같다 | D12.1 논증(제도 근거 `[미인용]`) |
| 7 | 체결가(얇은 호가 sweep · 종가단일가 대량 가격 충격) | 즉시성이 아니라 가격 | 불변식 4 밖 — 불변식 6 · D-1 (다) 결정이 받은 축 | 3.8 질의로 착지 뒤 기록 |

1·2·4 가 「오늘보다 나빠질 수 있는」 칸이고 각각 위 상한을 갖는다. 1·2 의 수락은 사람 확인(`issues.md` D-2 ①), 4 는 수치 상한으로 판정한다.

## D13 — 10판: 9판 재리뷰 처분 — 첫 시도 한정 · 429 기록 · release 실패 칸 (Manager 결정 2026-10-29, KISS)

> 9판 재리뷰(`review.md` 「9판 재리뷰」, `525b4fb6c`)는 FAIL(P0 1 · P1 3 · P2 1)이었다. Manager 는 **복잡도를 줄이는 쪽**으로 정했다. 이 절이 D12.2 의 429 입력 · D12.3 전체 ·
> D12.7 표를 **대체**한다(그 자리에 「10판에서 대체됨」 표시). D12.1(엔진 전용 읽기) · D12.2 의 토큰 · 주기당 1건 · 겹침 0 · D12.4~D12.6 은 그대로다. 새 AST 는
> `engine--exitobserver.release` 1개다(HEAD `525b4fb6c`, 소스 바이트 = origin/main `4b67dd6a9`, `enumerate.py` 43개 stale 0).

### D13.1 — 하한가는 그날 그 종목의 **첫 보호 매도 시도**에만 (P0 — T3 단조성)

- **규칙**: 원장에 같은 `account_ref` · `market` · `trading_day` · symbol 의 **보호 청산 매도 intent 가 하나라도** 있으면 그날 그 종목은 하한가 금지다. 그 intent 의 가격 · 상태 · 출처는 묻지 않는다.
  「보호 청산 매도 intent」 = `intents`(`internal/journal/schema.go:202-217` — `market` · `trading_day` · `account_ref` · `symbol` · `side` 칸을 가진다) 행 중 `side = 'SELL'` 이고
  `exit_events.proposed_intent_id` 로 그 행을 가리키는 판정의 `action` 이 `BASELINE_BREACH_EXIT` · `STOP_LOSS_LADDER` 인 것(**11판 정정 → D14.1**: 질의는 이 값을 리터럴로 쓰지 않고 `isProtective` 와 같은 Go 상수에서 유도한다 — 10판 원문은 `_EXIT` 가 빠진 오기였다)(`exit_events` `core_domain.go:251-261`, 색인
  `account_views.go:113` — 손 읽기, 구현 때 쿼리 시험으로 고정). symbol 은 `UPPER(TRIM())` 로, market 은 소문자로 맞춘다.
- **왜 단조인가**: `intents` 는 「Immutable record … Never updated」(`schema.go:201` 주석)이고 지우는 경로가 없다 — 행이 한 번 생기면 그날 끝까지 있다. 그래서 재시작 뒤에도 같은 답이 나오고,
  9판의 「최신 PLACE 상태」 같은 비단조 입력(`IN_DOUBT → CONFIRMED` 전이로 금지가 풀림 — 재리뷰 P0)이 없다. 꼬리표도 출처 판정도 필요 없다 — **T2 · T3 · `issues.md` D-2 ⑤ 는 10판에서 대체됨.**
- **원장 읽기 오류 · 파싱 불가 → 금지**(fail-closed — 항목을 세우지 않아 오늘 LIMIT).
- **메모리 쪽**: `submit` 이 보호 청산의 `Place` 를 부르면 그 결과와 무관하게 그 종목의 그날 항목을 「금지」로 바꾼다(제출 순간 — 9판의 결과별 갈래가 필요 없다). 재시작 뒤에는 갱신기가 항목을
  세우기 **전에** 위 원장 질의로 같은 술어(「그날 보호 청산 매도 intent 가 있나」)를 계산한다. 두 경로가 다를 수 있는 한 경우: `Place` 가 `Prepare` 전에 로컬에서 끝나 intent 행이 생기지
  않으면(예: `ReasonSymbolInFlight` B10 류) 메모리는 금지, 재시작 뒤 원장은 허용이다. 그 시도는 브로커에 닿지 않았으므로 재시작 뒤 하한가 1회는 「같은 거부의 반복」이 아니다 —
  보수적 쪽(메모리)이 운영 중 기본이고, 재시작은 많아야 하루 1회 하한가를 더 허용한다.
- **보수적 손실(의도)**: 첫 시도 뒤의 보호 매도 — 부분 체결 뒤 잔량, 첫 시도가 창 밖(08:59 등)이었던 날의 창 안 재발의, 첫 하한가 시도가 수용됐지만 미체결로 끝난 뒤의 재발의 — 는
  전부 오늘의 LIMIT(관측가)으로 나간다. 하한가의 공격성을 그 시도들에서 잃을 뿐, 오늘보다 나빠지지 않는다. 잔량의 재가격은 a167 소관이다.
- 이 규칙으로 8판 · 9판의 되먹임(B13 · 결과별 금지 · T3 재구성)은 **전부 대체**된다. 9판 D12.3 의 「술어 하나」 원칙은 남는다 — 술어는 「그날 보호 청산 매도 intent 존재」 하나다.

### D13.2 — HTTP 429 를 헤더와 무관하게 기록 (P1 — 헤더 없는 429) — **12판에서 대체됨 → D15.1**(429 기록은 a166 으로, a087 `doRequest` 편집 0)

- **사실(재리뷰 · AST)**: 헤더 없는 429 가 `Retrier.Query` 안에서 재시도로 성공하면(B10 대기 → 다음 시도 B5 `ClassOK` `retry.go:351-355`) `cycle.Err` 는 nil 이고, 그 응답의 budget 은
  `Reported=false` 라 `Exhausted()`(`ratebudget.go:177`)도 거짓이다 — 9판의 정지 입력 ②③ 이 둘 다 놓친다.
- **처분**: `official.Client` 에 「마지막 HTTP 429 시각」 원자 값 하나와 읽기 접근자 하나를 둔다. 기록 자리는 모든 응답이 지나는 유일한 `c.hc.Do` 자리인 `Client.doRequest`
  (`internal/official/client.go:191-207`) — `c.hc.Do` 성공 뒤(B1 `:194` 다음), `resp.StatusCode == 429` 이면 헤더와 무관하게 `time.Now()` 를 저장한다(같은 함수가 이미
  `time.Now()` 로 `c.rates.record` 를 한다 `:199`). 반환값 · 오류는 바뀌지 않는다. 갱신기는 토큰을 받을 때 그 시각이 5분 안이면 GET 0 이다. 비교는 같은 벽시계(`time.Since`)로 한다 —
  엔진 시계가 시험 더블인 시험에서는 정지 판정 함수에 두 시각을 넣어 잰다.
- 9판의 정지 입력 ②(`Run` 의 `cycle.Err`)와 ③(`RateBudgets()` `Exhausted()`)은 **10판에서 대체됨** — 이 하나가 셋을 다 덮는다(갱신기 자신의 GET 도 `doRequest` 를 지난다).
  그래서 `ExitObserver.Run` 편집은 토큰 보내기 한 줄로 줄어든다.
- **편집하는 기존 함수 = `Client.doRequest`** — Pre-Edit · FLM 대상에 넣는다. 다른 change 와의 겹침(문서 대조, 이 worktree 기준):
  - **a166**(`a166-account-shared-rate-limit`, proposal 단계 · 코드 0): 삽입 자리 후보가 `official.Client.doRequest` · `send`(a166 proposal `:43` · `:76`). **같은 함수 · 같은 관심사(429)** 다 —
    둘 중 먼저 착지하는 쪽 뒤에 다른 쪽이 그 sha 에서 AST 를 다시 만든다. a166 이 계좌 공유 429 기록을 만들면 a087 은 그것을 읽기만 하는 쪽으로 줄일 수 있다(그때 편집 0).
  - **a100**(HOLD): `internal-official--client.dorequest` FLM 번들이 있다(a100 tasks `:67` — 계측 transport). 착수 순서는 열려 있다 — 같은 처리.
  - **a165**: FLM 번들 목록에 `client.send` · `client.postwithheaders` 는 있고 `client.dorequest` 는 없다(`openspec/changes/a165-safe-order-placement-retry/analysis/function-logic/` 목록).
    `send` 가 `doRequest` 를 부르므로(`client.go:329` · `:354`) 호출자 쪽 좌표는 tasks 1.0 재대조에 넣는다.
- **시험**: httptest 가 헤더 없는 429 를 한 번 주고 다음 시도에 200 → `Retrier.Query` 성공 · `cycle.Err == nil` → 그 뒤 5분 안의 토큰에서 하한가 GET 0. 헤더 있는 429 · 갱신기 자신의 429 행도 같은 결과.
  `doRequest` 의 반환값 · 오류 · `rates.record` 동작이 편집 전과 같다(기존 ratebudget 시험 그대로 green).

### D13.3 — timestamp 의 뜻은 5.4 착지 전 gate 로 **보류** (P1)

- 정확한 상태: 9판 P0-1(전일 하한가)은 **닫히지 않았다** — D12.1 은 raw · timestamp 를 보존하고 조건 ①② 를 세웠을 뿐, timestamp 가 효력 날짜인지 응답 시각인지 모르는 한 신선도 증명이 아니다.
  그 판정을 tasks 5.4(착지 전 필수 gate)로 **보류**한다. 구현 · 배포 완료를 주장하지 않는다.
- **기본 처분(사람 확인 대상 — `issues.md` D-2 ②)**: 5.4 에서 timestamp 가 **응답 시각**으로 드러나면 **KR 하한가를 끈다(off — 보호 매도 전부 오늘 LIMIT)**. 다른 권위 있는 효력일 검증은 별도 change 로 한다.
- 문구 정정: D12.1 의 「timestamp 의 KST 날짜 = 결정 시각의 KRX 거래일」은 **KST 달력 날짜(평일)의 일치**일 뿐이다 — 평일 휴장과 기준가 변경(권리락 등)을 증명하지 않는다. 그 둘은 timestamp 뜻과 함께 5.4 · 별도 change 몫이다.

### D13.4 — 하한가 값 비교의 정본 (P2)

- 첫 시도 규칙으로 T3 의 가격 일치 비교와 그 오탐 문장은 사라졌다. 남은 비교는 D11 조건 ⑥(하한가 ≤ 관측가) 하나다.
- **정본**: 하한가 원 문자열은 앞뒤 공백 없는 양의 정수 decimal(`^[1-9][0-9]*$`)이어야 하고 정수로 파싱한다. 관측가는 decimal 로 파싱해 정수 하한가와 값으로 비교한다(문자열 비교 금지).
  어느 쪽이든 파싱 불가면 하한가를 쓰지 않는다(오늘 LIMIT, 사유 기록). 시험 행: `"50400"` vs 관측가 `"50400"` · `"50401"` · `"50399"` / 하한가 `"050400"` · `"50400.0"` · `" 50400"` · `"5.04e4"` · `""` →
  쓰지 않음 / 관측가 파싱 불가 → 쓰지 않음.

### D13.5 — 불변식 4 표 (10판) — **11판에서 대체됨 → D14.4**

오늘(모든 보호 매도 = 관측가 LIMIT)보다 나빠질 수 있는 칸과 상한:

| # | 칸 | 무엇이 나빠지나 | 상한 | 근거 · 남은 `[미측정]` |
| --- | --- | --- | --- | --- |
| 1 | 첫 하한가 시도가 브로커에 **직접 거부**(전일 값 < 그날 하한가 · SOR/NXT 범위 차이 · VI · 개장 지연 거동) | 거부 → 해제 → 다음 관측에서 오늘 LIMIT | **종목당 하루 1회** × (거부 응답 시간 + 관측 주기 5초) | D13.1 첫 시도 규칙. 거부가 실제로 나는지 `[미측정]` → 5.4 |
| 2 | 첫 하한가 시도가 **IN_DOUBT → 비수용 확정** | 오늘의 IN_DOUBT 와 같은 해소 대기 + 해제 뒤 1 주기 | **종목당 하루 1회** × (오늘과 같은 해소 시간 + 5초) | D13.1. `noteHeldProposal` B4(`exit_held_proposal.go:140`)가 해제. 하한가가 IN_DOUBT 확률을 바꾸는지 `[미측정]` |
| 3 | 위 1 · 2 의 **journal release 실패 · 복구** | `release`(`exitloop.go:1603-1608`, B1 `:1604` `ReleaseUnacceptedExitProposal` 오류 → 반환)가 실패하면 발의가 무장된 채 남고, 다음 주기부터 `noteHeldProposal` B4 → B5(`:142`)가 해제를 다시 시도 | **오늘의 비수용 해제 실패 경로와 같은 상한**(같은 함수 · 같은 재시도) · 발생은 **종목당 하루 1회 이하**(첫 시도 규칙 — 하한가가 새로 만드는 거부는 하루 1번뿐) | 원장 쓰기 실패의 지속 시간 자체는 오늘과 같은 미지수 — a087 이 바꾸지 않는다 |
| 4 | `MARKET_DATA` 압력 → 관측 `/prices` 429 → 주기 건너뛰기 · 재시도 대기(오늘도 있는 동작) | 그 대기가 더 자주 일어날 가능성 | 추가 요청 **주기당 ≤ 1 = 0.2/s(한도 10/s 의 2%)** · 자기 `/prices` 와 겹침 0 · **HTTP 429(헤더 유무 무관)를 본 뒤 5분 GET 0** | D12.2 토큰 + D13.2. 다른 소비자와 같은 초에 겹칠 때의 효과 `[미측정]`. 다른 프로세스(사람 CLI 등)의 429 는 이 클라이언트에 안 보인다 — a166(계좌 공유) 몫 |
| 5 | 재시작 · 판 · 갱신기 고장 · 원장 읽기 실패 · 정지 | — | **0** — 항목이 없거나 금지면 오늘 LIMIT | D13.1 fail-closed |
| 6 | 전일 하한가가 그날 하한가보다 높음(⑥ 통과) | — | **0** — 바닥이 관측가 이하라 오늘보다 공격적이거나 같다 | D12.1 논증(제도 근거 `[미인용]`) |
| 7 | 첫 시도 뒤의 보호 매도(잔량 · 재발의) | — | **0** — 오늘 LIMIT(보수적 손실, D13.1) | — |
| 8 | 체결가(얇은 호가 sweep · 종가단일가 대량 가격 충격) | 즉시성이 아니라 가격 | 불변식 4 밖 — 불변식 6 · D-1 (다) 결정의 축 | 3.8 질의 |

1 · 2 · 3 · 4 가 「오늘보다 나빠질 수 있는」 칸이다. 1~3 의 발생은 모두 종목당 하루 1회 이하로 묶인다. 수락은 사람 확인(`issues.md` D-2 ①).

## D14 — 11판: 10판 재리뷰 처분 — 보호 action 상수 유도 · 시작 정숙 · 3 req/s 기준 (Manager 결정 2026-10-29)

> 10판 재리뷰(`review.md` 「10판 재리뷰」, `4242ce3ed`)는 FAIL(P0 1 · P1 1)이었다. 이 절이 D13.1 의 action 문자열 · D13.2 의 수명 범위 · D12.2 「새 압력의 수치」 · D13.5 표를 **대체**한다.
> 새 AST 는 없다 — 인용은 기존 `engine--isprotective`(`exitloop.go:1463-1465`) · `engine--exitobserver.submit` 과 손 읽기(`client.go:317-365` `send`, 표시함)다. `enumerate.py` 43개 stale 0.

### D14.1 — 재구성 술어의 action 집합은 Go 상수에서 유도 (P0)

- **오기 정정**: 기준선 이탈 action 의 실제 값은 `exitpolicy.ActionBaselineBreach = "BASELINE_BREACH_EXIT"`(`internal/exitpolicy/ratchet.go:97`), 손절 사다리는
  `exitpolicy.ActionLadderStop = "STOP_LOSS_LADDER"`(`:104`)다. 10판 D13.1 · tasks 2.9 의 `BASELINE_BREACH`(접미사 `_EXIT` 없음) 는 존재하지 않는 값이었다 — 그대로면 첫 기준선 이탈 하한가 시도가
  재시작 뒤 질의에 보이지 않아 금지가 풀린다. 이 change 문서의 같은 오기(1~10판 서술 포함)를 전부 `BASELINE_BREACH_EXIT` 로 고쳤다(재검: 이 change 의 design · tasks · proposal · issues · spec 에서 「`_EXIT` 없는 기준선 이탈 값」 0건, 이 문장의 인용과 `review.md` 의 리뷰어 원문 · `analysis/` 기록물은 제외 —
  StockOS 쪽 `STOP_LOSS` 는 StockOS 어휘라 그대로).
- **유도 규칙**: 재구성 질의는 action 을 **리터럴 문자열로 쓰지 않는다.** 질의 인자는 `string(exitpolicy.ActionBaselineBreach)` · `string(exitpolicy.ActionLadderStop)` — `isProtective`
  (`exitloop.go:1463-1465`, `p.Action == exitpolicy.ActionBaselineBreach || p.Action == exitpolicy.ActionLadderStop`)가 참을 내는 **같은 상수**다. `isProtective` 본문은 편집하지 않는다.
  `exit_events.action` 에는 판정의 `Action` 이 `string(...)` 그대로 기록된다(재리뷰 손 읽기 — `exitloop.go:1401-1415`).
- **구조 시험**(**12판 정정 → D15.2**: 이름 열거가 아니라 타입 검사로 평가한 `Action` 값 집합): `internal/exitpolicy` 의 `Action` 상수 선언 전부를 AST 로 열거하고, 각 값 `a` 에 대해 `isProtective(Proposal{Action: a})` 가 참인 집합 = 질의 action 집합임을 단언한다.
  새 보호 action 이 생겨 `isProtective` 가 바뀌면(또는 새 상수가 `isProtective` 에 들어가면) 질의 쪽이 따라오지 않는 한 실패한다. 상수 열거가 0개이거나 `isProtective` 참 집합이 비면 실패한다
  (표본 0 통과 금지).
- **행동 시험**: 기준선 이탈(`ActionBaselineBreach`) 하한가 시도 → 재시작 → 같은 종목 보호 청산 = 관측가 LIMIT. 손절 사다리도 같은 행. 결과별(CONFIRMED · FAILED_CONFIRMED · IN_DOUBT→비수용)
  세 행을 두 action 에 각각.

### D14.2 — wire 요청 상한 (P1)

- **(a) 401 재전송**: 엔진 전용 하한가 읽기(D12.1)는 표준 `Client.get` → `Client.send`(`internal/official/client.go:317-365` — 손 읽기)를 탄다. `send` 는 첫 `doRequest` 가 401 이면 토큰을
  새로 받아 **최대 두 번 더** `doRequest` 를 부른다(`:340` `for attempt := 0; attempt < 2 && code == http.StatusUnauthorized`). 그래서 토큰 하나의 wire GET 상한은 **최대 3** 이다.
  새 전송 경로는 만들지 않는다 — 401 은 드물고(토큰 교대 사고는 `--config-dir` 처분으로 닫힘), 재전송 정책을 두 벌로 두지 않는다. 상한을 표에 그대로 적는다(D14.4).
- **(b) 시작 정숙 구간**: 갱신기는 **엔진이 시작한 뒤 5분 동안 GET 0** 이다(갱신기 `Run` 시작 시각 기준). 마지막 429 시각(D13.2)은 프로세스 메모리라 재시작하면 0 이 되지만,
  시작 정숙 구간이 「429 직후 재시작」 칸을 닫는다 — 429 를 저장하지 않아도 재시작 뒤 첫 5분은 어떤 경우든 GET 0 이다. 그동안의 보호 매도는 오늘 LIMIT(약화 아님).
- **시험**: 401 → 401 → 200 에서 wire GET 3 · 토큰 하나 · 다음 토큰까지 추가 0 / 헤더 없는 429 → 재시도 성공 → 재시작 → 5분 GET 0 / 시작 직후 4분 59초 토큰 GET 0, 5분 뒤 토큰 GET 1.

### D14.3 — 예산 기준 = 계좌 공유 3 req/s (Manager 발견 — 재리뷰 밖) — 기준은 유지, 「a087 은 a166 을 기다리지 않는다」 · 평균 수치는 **12판에서 대체됨 → D15.1**(a166 admission 전 GET 0)

- **근거**: 사용자 결정(2026-10-22) — 「요청 속도 상한 3 req/s 는 계좌 전체 공유다 — 엔진 · soak · verify · 프로브 · CLI 등 같은 계좌를 치는 모든 프로세스의 합」
  (`openspec/changes/verify-execution-capability/issues.md` 「사용자 결정 (2026-10-22, 1.3 확정)」 결정 3, `:578`), 같은 절 Q2 — 3 req/s 는 우리 키 하나의 예산(`:616`). 소유 change 는
  `a166-account-shared-rate-limit`(proposal 단계 · 정의만 — `proposal.md`).
- **정정**: 9판 D12.2 · 10판 D13.5 의 「그룹 한도 10/s 의 2%」는 openapi 의 그룹 한도(`docs/migration/open-api.md:75`)를 기준으로 삼은 잘못이다. 기준은 **계좌 공유 3 req/s** 다 —
  갱신기 추가는 **주기당 ≤ 1(평균 0.2/s ≈ 3 req/s 의 6.7%)**, 401 이면 그 주기에 **최대 3(0.6/s ≈ 20%)**. 그 1~3 건은 자기 관측 `/prices` 와 다른 초에 있다(D12.2 겹침 0).
- **a166 과의 관계**: a166 이 계좌 공유 상한을 착지하면 갱신기 GET 도 그 예산을 탄다(같은 `official.Client` 전송 자리 `doRequest` 를 지나므로). a087 은 a166 을 **기다리지 않는다** —
  a166 전에는 위 수치가 a087 이 더하는 압력의 상한이고, a166 뒤에는 a166 의 배분이 그 위에 선다. D13.2 의 429 기록이 a166 과 겹치는 처리는 D13.2 그대로다.

### D14.4 — 불변식 4 표 (11판) — **12판에서 대체됨 → D15.3**

오늘(모든 보호 매도 = 관측가 LIMIT)보다 나빠질 수 있는 칸과 상한:

| # | 칸 | 무엇이 나빠지나 | 상한 | 근거 · 남은 `[미측정]` |
| --- | --- | --- | --- | --- |
| 1 | 첫 하한가 시도의 **직접 거부** | 거부 → 해제 → 다음 관측에서 오늘 LIMIT | **종목당 하루 1회**(재시작 포함 — D14.1 상수 유도로 두 보호 action 모두 재구성에 보임) × (거부 응답 + 5초) | D13.1 · D14.1. 실제 거부 여부 `[미측정]` → 5.4 |
| 2 | 첫 하한가 시도의 **IN_DOUBT → 비수용 확정** | 오늘의 IN_DOUBT 와 같은 해소 대기 + 해제 뒤 1 주기 | **종목당 하루 1회** × (오늘과 같은 해소 시간 + 5초) | D13.1 · `noteHeldProposal` B4 |
| 3 | 1 · 2 의 **journal release 실패 · 복구** | 발의 무장 유지 → 다음 주기 `noteHeldProposal` B4→B5 재해제 | **오늘의 비수용 해제 실패 경로와 같은 상한**, 발생 **종목당 하루 1회 이하** | `release` B1(`exitloop.go:1604`) |
| 4 | **계좌 공유 3 req/s** 압력 → 관측 `/prices` 429 → 주기 건너뛰기 · 재시도 대기(오늘도 있는 동작) | 그 대기가 더 자주 일어날 가능성 | 추가 wire GET **주기당 ≤ 1(평균 0.2/s ≈ 6.7%), 401 이면 ≤ 3(≈ 20%)** · 자기 `/prices` 와 겹침 0 · **HTTP 429(헤더 무관)를 본 뒤 5분 GET 0** · **엔진 시작 뒤 5분 GET 0**(재시작이 429 기록을 지워도 닫힘) | D12.2 · D13.2 · D14.2 · D14.3. 다른 프로세스(사람 CLI · soak · verify)와 같은 초에 겹칠 때의 효과 `[미측정]` — 그 배분은 a166 몫 |
| 5 | 재시작 · 판 · 갱신기 고장 · 원장 읽기 실패 · 정지 · 시작 정숙 구간 | — | **0** — 항목이 없거나 금지면 오늘 LIMIT | D13.1 · D14.2 (b) |
| 6 | 전일 하한가가 그날 하한가보다 높음(⑥ 통과) | — | **0** — 바닥이 관측가 이하 | D12.1 논증(`[미인용]`) |
| 7 | 첫 시도 뒤의 보호 매도(잔량 · 재발의) | — | **0** — 오늘 LIMIT(보수적 손실) | D13.1 |
| 8 | 체결가(얇은 호가 sweep · 종가단일가 대량 충격) | 즉시성이 아니라 가격 | 불변식 4 밖 — 불변식 6 · D-1 (다) | 3.8 질의 |

1~4 가 「오늘보다 나빠질 수 있는」 칸이다. 1~3 은 종목당 하루 1회 이하, 4 는 위 수치 상한. 수락은 사람 확인(`issues.md` D-2 ①).

## D15 — 12판: 11판 재리뷰 처분 — a166 admission 전 하한가 GET 0 · Action 값 집합 비교 (Manager 결정 2026-10-29)

> 11판 재리뷰(`review.md` 「11판 재리뷰」, `bb0080ecd`)는 FAIL(P0 1 · P1 1)이었다. 이 절이 D13.2(`doRequest` 429 기록) · D14.2 (a) 의 「상한에 그대로 적는다」 · D14.3 의
> 「a087 은 a166 을 기다리지 않는다」 · D14.1 의 구조 시험 방식 · D14.4 표를 **대체**한다. 새 AST 0(문서만). `enumerate.py` 43개 stale 0.

### D15.1 — 하한가 GET 은 계좌 공유 예산의 admission 을 받을 때만 (P0)

- **재리뷰 사실**: 401 → 401 → 200 은 토큰 하나에 wire GET 3 이고, 그 셋 사이에 pacing · admission 이 없다. 토큰은 `ObserveOnce` 반환 직후(`Sleep` 전)에 나오므로 직전 관측 `/prices` 와
  동시 실행은 0 이어도 **같은 1초 창**일 수 있다. 계좌 공유 상한은 **3 req/s**(사용자 결정 2026-10-22 결정 3 — D14.3)라 그 창에서 3 + `/prices` 1 은 상한을 넘는다. 「≤ 3/5초 = 0.6/s」는
  평균일 뿐 한 초 창의 상한이 아니다. 429 기록 · 시작 정숙은 응답을 받은 **뒤** 다음 GET 만 멈춘다.
- **처분 (fail-closed)**: 갱신기는 「계좌 공유 예산의 admission」 함수가 **주입될 때만** 돈다. 모양(제안): `admit(ctx) error` — 요청 하나를 보내기 전에 부르고, 거절이면 그 요청을 보내지 않는다.
  **nil 이면 GET 0** 이고 판은 비어 있으므로 보호 매도는 전부 오늘의 LIMIT 이다. 401 재전송의 둘째 · 셋째 요청도 각각 admission 을 탄다 — 그래서 admission 은 갱신기 쪽 래퍼가 아니라
  **wire 요청마다** 걸려야 하고, 그 자리는 a166 이 소유하는 계좌 공유 전송 관문이다(a166 proposal `:43` · `:76` — 삽입 후보 `official.Client.doRequest` · `send`). a087 은 그 관문의
  admission 결과(허용 · 거절)만 받는다.
- **엔진 배선**: a166 착지 전까지 **nil**. a087 의 `Context.ExitObserver` · 갱신기 생성 자리는 admission 을 받는 칸만 두고 값을 넣지 않는다. 그러므로 **a166 전에 a087 이 착지해도 생산 동작
  변화는 0** 이다 — 판이 늘 비어 가격 결정이 언제나 오늘 사다리를 고른다(D11 (b) fail-safe). 이것은 착수 조건이 아니라 **활성화 조건**이다. 사람 확인 대상: `issues.md` D-2 ⑥.
- **10판 `doRequest` 429 원자 기록 — a166 으로 옮긴다(권장안 채택).** 비용 비교: a087 에 남기면 `doRequest`(High-risk, a166 · a100 과 같은 함수)를 편집하고 a166 착지 때 그 기록과
  a166 의 공유 예산이 두 판정이 된다([[two-judgements-cover-for-each-other]]). 옮기면 a087 의 `doRequest` 편집이 **0** 이 되고 a100 · a166 과의 겹침도 사라진다. 429 뒤 정지는 a166 admission 이
  거절로 표현한다(a166 이 정한다). 그래서 D13.2 · tasks 2.10 은 **12판에서 대체됨**이고, a087 의 Pre-Edit · FLM 목록에서 `official.Client.doRequest` 를 뺀다. a166 문서 반영(429 기록을
  가져갈 것)은 이 디렉터리 범위 밖 — Manager 요청으로 남긴다(`issues.md` D-2 ⑥).
- 남는 것: 토큰(주기 끝 · 주기당 1건 · GET 시한 ≤ 2초 — D12.2), 시작 정숙 5분(D14.2 (b) — 갱신기 로컬, `doRequest` 편집 아님), 401 재전송 3건(표준 `send` — 새 경로 없음).
- **시험(셋)**: ① admission nil → 하한가 GET 0 · 판 비어 있음 · 창 안 보호 매도 = 오늘 LIMIT 골든 ② admission 이 거절 → 그 요청 GET 0(그 주기 추가 wire 0) ③ 401 → 401 → 200 에서 wire 요청 3건이
  **각각** admission 을 한 번씩 탐(호출 계수 3, 거절이 둘째에 오면 셋째 wire 0). ③ 은 시험 더블의 계좌 공유 관문으로 잰다 — 실제 관문은 a166 이 만든다.

### D15.2 — 구조 시험은 평가된 `exitpolicy.Action` 값 전체 집합을 비교 (P1)

- 구조 시험은 상수 **이름 수**나 식별자 prefix 로 세지 않는다. `go/packages`(또는 `go/types` 로 타입 검사한 패키지)로 `internal/exitpolicy` 의 패키지 범위 객체 전체를 훑어
  **타입이 `exitpolicy.Action` 인 모든 상수**(`*types.Const`, `types.Identical(obj.Type(), Action)`)를 모으고, 각 값을 `go/constant` 로 평가한 `string` 값의 **집합**을 만든다.
  `const ActionFuture = Action("...")` · 다른 파일의 상수 · 타입 자리가 다른 선언도 타입 검사로 같은 집합에 든다.
- 같은 값을 가진 두 이름(별칭)은 집합에서 하나가 되므로, 시험은 따로 **값 → 이름 목록** 지도를 만들어 이름이 둘 이상인 값을 보고(실패가 아니라 기록 — 같은 값이면 `isProtective` 판정이
  같으므로 집합 비교는 여전히 옳다).
- 단언: { v ∈ 값 집합 | `isProtective(Proposal{Action: Action(v)})` } = 재구성 질의의 action 값 집합. 값 집합이 비거나 참 집합이 비면 실패(표본 0 통과 금지). 새 보호 값이 생기면 질의 쪽이
  따라오기 전까지 실패한다.

### D15.3 — 불변식 4 표 (12판 — D14.4 를 대체)

| # | 칸 | 상한 — **a166 admission 전(생산 기본)** | 상한 — a166 admission 뒤(활성화 시) | 근거 |
| --- | --- | --- | --- | --- |
| 1 | 첫 하한가 시도의 직접 거부 | **0** — 판이 비어 하한가 시도가 없다 | 종목당 하루 1회 × (거부 응답 + 5초), 재시작 포함 | D13.1 · D14.1 · D15.1 |
| 2 | 첫 하한가 시도의 IN_DOUBT → 비수용 확정 | **0** | 종목당 하루 1회 × (오늘과 같은 해소 시간 + 5초) | D13.1 |
| 3 | 1 · 2 의 journal release 실패 · 복구 | **0** | 오늘의 비수용 해제 실패 경로와 같은 상한, 종목당 하루 1회 이하 | D13.5 #3 |
| 4 | 계좌 공유 3 req/s 압력 → 관측 `/prices` 429 · 지연 | **0** — 하한가 GET 0 | wire 요청마다 a166 admission 이 허용한 것만(401 재전송 포함) · 토큰 주기당 1 · 시작 뒤 5분 0. 한 초 창의 상한은 a166 의 배분이 정한다 | D15.1 · D12.2 · D14.2 (b) |
| 5 | 재시작 · 판 · 갱신기 고장 · 원장 읽기 실패 · admission nil/거절 · 시작 정숙 | 0 | 0 — 오늘 LIMIT | D13.1 · D15.1 |
| 6 | 전일 하한가 > 그날 하한가(⑥ 통과) | 0 | 0 — 바닥이 관측가 이하 | D12.1 |
| 7 | 첫 시도 뒤의 보호 매도 | 0 | 0 — 오늘 LIMIT(보수적 손실) | D13.1 |
| 8 | 체결가(얇은 호가 · 종가단일가) | 0 | 불변식 4 밖 — 불변식 6 · D-1 (다) | 3.8 |

a166 admission 전에는 a087 이 오늘보다 나빠지게 하는 칸이 **없다**(생산 동작 변화 0). 활성화 뒤의 1~4 수락은 사람 확인(`issues.md` D-2 ①).

## D1 — 유형은 제안의 성격이 정한다

> **8판에서 대체됨(코드 블록)**: 유형은 바뀌지 않는다 — 보호 제안이 정하는 것은 **가격**이다(D11 (a)·(b) 조건 ①). 「`isProtective` 를 재사용한다」·
> 「새 술어를 만들지 않는다」·「`isFullExit` 를 쓰지 않는다」·「익절을 안 바꾸는 이유」는 8판에서도 그대로다 — 셋째 소비자는 유형 결정이 아니라 `exitLimitPrice` 다.

`sellIntent`는 지금 유형을 상수로 쓴다. 제안을 인자로 받아 판정한다.

```go
orderType := "limit"
if isProtective(proposal) {   // BASELINE_BREACH_EXIT · STOP_LOSS_LADDER
    orderType = "market"
}
```

**`isProtective`를 재사용하는 것이 핵심**이다. 새 술어를 만들면 "보호란 무엇인가"의 정의가
둘이 되고, 그 둘이 갈라지는 순간 한쪽만 시장가가 된다. [`exitloop.go`](../../../internal/app/engine/exitloop.go)의
`isProtective`(패키지 함수)는 이미 §0.3 근거로 존재하고 주석이 그 이유를 적어 뒀다 — "Nothing may
withhold one of these: §0.3 forbids weakening or delaying the immediacy of a stop".
같은 §0.3이 이 change의 근거이므로 같은 술어여야 한다.

**소비자가 셋이 된다 (재리뷰 P3-2).** base 이후 `isProtective` 는 재판정 보류 예외(`record`)와 a091 floor 의미
(`submit`)에 쓰인다(CodeGraph 1.6.0 비시험 호출자 = `record`·`submit` — `analysis/gate-0.4-codegraph.md`). a087 의
유형 결정이 셋째다. 그러므로 a087 은 **술어 자체를 바꾸지 않는다** — 바꾸면 세 소비자가 함께 움직인다. tasks 2.1 의
risk-pattern-report 가 세 소비자를 적고, 2.2 가 술어 불변을 시험으로 고정한다. (각 소비자 안의 분기 위치는 검증
필요 — AST 미작성, task 2.1 소관.)

`isFullExit`는 쓰지 않는다. 그것은 익절을 **포함**하고(`ActionLadderTakeProfit`), 익절은
지정가로 남는다.

### 익절을 안 바꾸는 이유

§0.9는 보수 방향만 허용한다. 익절의 목적은 즉시성이 아니라 **가격 확보**다. 시장가로
바꾸면 체결가 불확실성이 커지고 그 불확실성은 이익을 깎는 쪽으로도 열린다 — 보수 방향이
아니다. 그리고 익절이 안 나가도 포지션은 여전히 보호받는다(보호 제안이 별도로 존재한다).
보호가 안 나가면 아무것도 남지 않는다. **비대칭이 설계다.**

## D2 — "가격 없음" 거부가 보호를 막지 않게 한다 (2상 — 사용자 결정 2026-09-28)

> a089 처분에서 사용자가 확정한 재작성 축: **D2 를 분리해 선행**한다. 시장가 전환(D1·D3)은
> 사람 실측(§0.7) 뒤에만 착지할 수 있으므로, 그 실측을 기다리는 동안 "가격 없음" 거부가
> 보호를 막는 상태를 두지 않는다.

`sellIntent`(base `102d4e99` 좌표 exitloop.go:1571-1578 — HEAD 에서는 함수 머리 뒤 가격 사다리, 본문 바이트 동일
— 재리뷰 P2-5)의 가격 사다리:

```go
price := strings.TrimSpace(observed)
if price == "" { price = strings.TrimSpace(m.state.Baseline) }
if price == "" {
    return orderintent.PlaceIntent{}, fmt.Errorf(
        "position %s has no price to submit a liquidation at", m.position.ID)
}
```

관측가도 기준선도 없으면 **청산을 거부한다.** 시세를 못 읽은 순간 손절이 막힌다는 뜻이고,
그것은 §0.3이 금지하는 형태다.

### D2a — 선행 (실측 불요): 사다리에 셋째 단 — KR 당일 하한가

보호 제안(`isProtective`)에 한해 사다리를 늘린다: **관측가 → 기준선 → (KR) 당일 하한가 →
거부**. 하한가 LIMIT 매도는 "그 세션에서 가능한 가장 공격적인 유효 지정가"이고, 거래소
공표값이라 호가 그리드 위에 있다.

코드 영수증:

- 값의 출처는 공식 read `PriceLimits`(`internal/official/price_limits_reads.go` —
  `lowerLimitPrice`, nullable). 엔진 reader 인터페이스에 **이미 선언돼 있으나 비시험
  호출자 0** 이다(`internal/app/engine/reads.go:72`) — 배선만 없다.
- **KRX 전용**이다: `internal/client/marketdata.go:97` "미국장은 일일 가격제한 제도가
  없음". 그러므로 이 단은 KR 에만 선다.
- §0.4: 이 읽기는 관측 루프가 아니라 **거부 직전 폴백에서만** 1회 — 사건 한정이다.

남는 거부(정직한 열거 — fail-closed 는 거부하는 정상 입력을 말한다):

- US 포지션에서 관측가·기준선이 모두 빈 경우(하한가 제도 없음), 그리고 KR 하한가 읽기
  자체가 실패·null 인 경우. 이때는 거부하되 **critical 로 알린다**(무음 거부 금지 —
  a090 의 미관측 계수와 합집합, 교차 인용). US 무가격은 D2b 가 착지하면 자연 소멸한다
  (시장가는 가격이 필요 없다).
- 두 값이 모두 비는 모집단이 언제 생기는지(기준선 없는 포지션의 유형)는 phase 1 RED 가
  열거한다.

#### 반증 (2026-09-30, P1.2 ① 모집단 열거) — D2a 는 세우지 않는다

열거 결과 모집단은 **0** 이다(이 절의 줄 좌표는 전부 base `102d4e99` 기준 기록이다). 위 사다리의 폴백(B1 `:1573`)과
거부(B2 `:1576`)는 생산 경로에서 도달할 수 없다.
Manager 가 독립 대조해 확인했다(2026-09-30). 영수증(`issues.md` I-P1, `analysis/`):

- **호출 사슬 (CodeGraph 1.6.0)** — `sellIntent` 의 생산 호출 1곳 = `submit` `:1382`, `submit` ← `record` `:1301`,
  `record` ← `judgeRatchet`·`judgeLadder`. `observed` = `snapshot.ObservedPrice`(`record` `:1190`), 도중 재대입 없음.
- **가드 좌표 (AST, `analysis/p1-population/`)** — 스냅숏은 평가 성공 **뒤에만** 생성된다(`snapshot.go:189`→`:191`,
  `:120`→`:122` 오류 반환). 평가기는 성공 반환 전에 `positive("observed price", …)` 를 무조건 지난다
  (`ratchet.go:347` · `ladder.go:333`, 그 앞 반환은 전부 오류). `positive` 는 파싱 실패·`Sign() <= 0` 거부.
- **커버리지 0** — 엔진 스위트(`go test ./internal/app/engine -coverprofile`) B1 본문 `1573.17-1575.3` = 0,
  B2 본문 `1576.17-1579.3` = 0 (`analysis/p1-population/sellintent-coverage.txt`).

그러므로 **D2a 는 세우지 않는다.** 도달 불가 분기 뒤의 High-risk 코드는 생산 경로로 변이가 닿지 않고, 하한가 자동
제출은 미검증 가상 호출자에게 주문을 내주는 방향이라 보수 방향도 아니다 — 오늘의 B2 거부가 이미 그 가상 경로의
fail-closed 방어다. **B1·B2 는 fail-closed 방어 반환으로 남고, 닫힌 이유는 핀 시험이 고정한다**:
`internal/exitpolicy/a087_observed_price_pin_test.go`(빈·공백·"0"·파싱 불가 관측가 → 두 평가기 모두 거부, 변이 4/4 CAUGHT).
가드가 빠지는 편집은 그 시험을 깨뜨리고, 그때 B1·B2 가 다시 열린 문이 된다는 사실이 드러난다.

**"무가격 → 손절 없음"의 실제 자리는 `sellIntent` 가 아니라 판정 이전이다** — `observe` 가 `Last <= 0` 을 거르고
(`exitloop.go:765`) `ObserveOnce` 가 응답 없는·만료된 종목을 무음 `continue` 한다. 가격이 없으면 이탈을 판정할 수 없어
보호 제안 자체가 생기지 않는다. 그 자리는 **a090**(`a090-an-unobserved-position-is-counted`, freeze 완료) 소관이다.

### D2b — 시장가 전환 (실측 §0.7 뒤): 거부의 구조적 소멸

> **8판에서 대체됨**: 8판은 가격을 읽지 않는 갈래를 만들지 않는다. 창 안의 보호 매도는 판의 하한가를 가격으로 쓰고, 그 밖은 오늘의 사다리다 — `sellIntent` B1·B2 는
> 오늘처럼 남는다(D11 (b) 「가격 결정」). 245,750 사건을 고치는 것은 8판에서 D11 이다.

시장가 경로는 가격을 필요로 하지 않으므로 **보호 제안에서는 이 거부가 도달 불가가 된다.**
초안 리뷰의 H1이 "새 거부 경로를 만들지 말라"고 한 것의 반대편 — 있던 거부 경로 하나가
사라진다.

가격 읽기 자체를 보호 분기에서 **건너뛴다**. 읽어서 버리는 것이 아니라 읽지 않는다.
읽으면 실패할 수 있고, 실패가 거부로 이어지는 경로가 다시 생긴다.

~~D2b 가 착지하면 D2a 의 하한가 단은 보호 경로에서 도달 불가가 된다 — phase 2 GREEN 이
그 단을 **제거**하고 도달 불가 증명을 함께 남긴다.~~ (2026-09-30: D2a 불구현 — 위 「반증」. 제거할 단이 없다.
D2b 가 보호 분기에서 가격 읽기를 건너뛰면 B1·B2 는 보호 경로에서 구조적으로 사라지고, 익절 경로에는 오늘처럼
평가기 가드 뒤의 fail-closed 방어로 남는다.)
245,750 그리드 불일치 사건(옛 표기 「9분 사건」 — 척도는 「6회 결정 중 5회 미제출」, 관측가는 **있었다**)을 고치는 것은 D2a 가 아니라 D2b 다:
D2a 는 가격 부재 거부만 다룬다.

## D3 — 게이트는 축소에 한해서만 연다 (두 관문 — 2026-10-10 개정)

> **8판에서 대체됨 — 관문을 열지 않는다.** 보호 매도가 LIMIT 이므로 ②`checkOrderShape` · ③`placeIntentSupported` · `trading.Service` 옵션 · `NewContext` 의
> 옵션 부착 · KR 정수 backstop 이 전부 필요 없다(D11 (e)·(f)). 이 절의 ②·③ AST 대조와 「0.10 독립 리뷰 반영」은 7판의 기록으로 남는다. 그대로 유효한 것은
> 맨 끝 「왜 진입은 계속 막는가」 — 진입 시장가 거부는 8판에서도 바뀌지 않고, 그 고정 시험은 tasks 1.1 이 잇는다. P0-1 의 규칙(엔진 전용 상태는 `NewContext` 에서만)은
> 8판의 하한가 판 생성 자리에 그대로 적용된다(D11 (b)).

> **개정 (재리뷰 P0-1 · 사용자 결정 2026-10-10)**: 3판의 D3 는 `checkOrderShape`(②) 하나만 다뤘다. 엔진 주문은
> 그 뒤에 `trading.Service.Place` → `placeIntentSupported`(③)를 지나고, ③은 비분수 비지정가를 `ErrPlaceUnsupported`
> 로 거부한다. **②만 열면 보호 청산은 100% 로컬 거부된다.** 사용자 결정으로 ③도 같은 모양으로 연다.
>
> **재개정 (3차 재리뷰 A-P1-1 · 사용자 결정 2차 2026-10-10)**: ③ 개방은 **엔진 인스턴스 한정**이다(아래 ③). CLI·ops·MCP
> 기본 경로는 upstream 과 바이트 동일. 1차 결정의 「사람 CLI upstream 동작 변경 수용」은 철회됐다.

### ② `checkOrderShape` (`internal/execgw/failclosed.go`, 호출자 `CheckPlace` 1) — 8판에서 대체됨(무변경)

```go
if orderType != "limit" {
    return reject(ReasonUnsupportedOrderType,
        "only limit orders (and US fractional market orders) are supported, got %q", …)
}
```

이 거부가 진입과 청산을 구분하지 않는다. 구분을 넣는다.

```text
fractional          → 기존 분기 그대로 (US market only)
sell + market       → 통과. 단 Price != 0 이면 거부
buy  + market       → 거부 (현행 유지)
그 외 non-limit     → 거부 (현행 유지)
```

**새 분기의 위치 (재리뷰 P2-2 — 검증 필요, AST 미작성: task 1.0.1)**: 재리뷰의 손 읽기는 `orderType != "limit"` 거부
**뒤에** KR 통화 검사와 수량 양수 검사가 온다고 적었다. 그 자리에 `sell+market` 의 **조기 통과 반환**을 넣으면 그 검사들을
건너뛴다. 그러므로 새 분기는 **통과를 반환하지 않는다** — 비지정가 거부의 예외 조건으로만 서고, 뒤의 통화·수량 검사는
`sell+market` 에도 그대로 적용돼야 한다(또는 같은 검사를 반복). 실제 순서는 task 1.0.1 의 AST 열거로 확정하고, RED 는
「`sell+market` + 비 KRW 통화(KR)」·「`sell+market` + 수량 0」이 여전히 거부되는 행을 포함한다(tasks 1.1).

**마지막 가격 검사는 limit 전용으로 남는다 (3차 재리뷰 A-P2-2 — 검증 필요, AST 미작성: task 1.0.1)**: 손 읽기로는
`checkOrderShape` 의 맨 끝 거부 「a limit order needs a positive price」(`intent.Price <= 0`)가 수량 검사 뒤에 선다. 이 검사가
`sell+market` 에 적용되면 가격 없는 시장가가 전부 거부된다. 그러므로 `sell+market` 에서는 이 검사 대신 위의 `Price != 0`
거부가 선다 — 「limit 이면 `Price <= 0` 거부, sell+market 이면 `Price != 0` 거부」로 갈래가 나뉘고, 통화·수량 검사는 둘 다에
공통이다. RED 1.1 의 「`sell+market`·가격 없음 → 통과」 행이 이 갈래를 잡지만 설계가 침묵하지 않도록 여기 적는다.

**`Price != 0` 검사를 넣는 이유**: openapi가 `MARKET`에 `price` 전달을 금지하고 전달 시
`400 invalid-request`를 준다. ~~`orderintent`가 이미 정규화하므로 게이트는 두 번째 확인이다~~ — **정정(재리뷰 P2-2)**:
`orderintent` 의 MARKET → `Price = 0` 정규화는 `NormalizePlace` 안에 있고, `sellIntent` 는 `PlaceIntent` 를 **구조체
리터럴로 조립**하므로 그 정규화를 지나지 않는다. 그러므로 `checkOrderShape` 의 `Price != 0` 거부는 「두 번째 확인」이 아니라
**이 경로의 유일한 로컬 확인**이다. (`checkOrderShape` 주석의 「strict subset filter」는 이중 확인의 근거가 아니라
「진짜 검사는 서비스에 있다」는 뜻이었다 — 그것이 ③이다.) 결론(거부를 넣는다)은 같다.

**비분수 MARKET 의 정수 수량 (재리뷰 P2-2)**: openapi 는 소수 수량을 US 시장가 매도에만 허용하고 그 외는 400 이다.
현재 어느 관문이 비분수 MARKET 의 정수성을 검사하는지는 **미확인**(검증 필요 — task 1.0.1·1.5.1 AST). 처분은 그 열거
뒤 tasks 1.1a 에서 정한다 — 로컬 거부를 넣거나, 브로커 400 에 맡기고 그 사실을 시험 이름으로 고정하거나. 어느 쪽이든
침묵한 생략은 없다. **→ 정해짐(0.10 독립 리뷰 P1-2)**: 「브로커 400 위임」은 철회한다. KR 정수 수량은 MARKET 자격 조건이고 두 관문에
backstop 을 둔다 — 아래 「0.10 독립 리뷰 반영」 P1-2.

### ③ `placeIntentSupported` (`internal/trading/service.go`, 비시험 호출자 `PreviewPlace`·`Place`) — 엔진 인스턴스 한정 — 8판에서 대체됨(무변경)

```go
// 비분수 주문 (재리뷰 인용, 검증 필요 — AST 미작성: task 1.5.1)
if intent.OrderType != "limit" { return false }   // Place → ErrPlaceUnsupported
// (손 읽기) 그 뒤 KR → CurrencyMode == "KRW", US → "KRW" | "USD"
```

**누가 이 술어를 지나는가 (3차 재리뷰 A-P1-1)**: `trading.NewService` 로 만든 인스턴스 셋 — 엔진
(`internal/app/engine/engine.go` `OrderPath.Trading`, 소비자 `execgw` 게이트웨이·엔진 `tradingService`), CLI 앱
(`internal/app/app.go` → `tossctl order place`·`tossctl ops`), **MCP 서버**(`cmd/tossctl/mcp.go` — ops `place_order` 가
`internal/mcp/catalog.go` 로 노출되는 에이전트 표면). 셋은 같은 `cfg.Trading` 을 공유하므로 config 토글로는 엔진과 사람·에이전트
경로를 가를 수 없다. 갈라지는 것은 **인스턴스**다.

**처분 (사용자 결정 2차 2026-10-10)**: 개방은 엔진 인스턴스에만 생성자 옵션으로 선다. 모양은 **제안**이고 코드는 tasks 1.5.3
몫이다 — 기존 빌더 관례(`WithLineage`·`WithConditional`)를 따라:

```go
// 제안 (tasks 1.5.3 에서 확정)
func (s *Service) WithProtectiveMarketSell() *Service { s.protectiveMarketSell = true; return s }

// ~~engine.go:401(NewOrderPath) — 엔진 인스턴스에서만 켠다~~ — 철회(0.10 독립 리뷰 P0-1): NewOrderPath 는
// flatten-all 도 쓰는 공용 빌더다. 옵션은 NewContext 안에서만 덧입힌다 — 아래 「0.10 독립 리뷰 반영」 P0-1.
trading.NewService(cfg.Trading, broker).WithConditional(broker) // NewOrderPath — 바이트 그대로

// 술어는 인스턴스 값을 인자로 받는다 — PreviewPlace·Place 두 호출자가 같은 값을 넘긴다
func placeIntentSupported(intent orderintent.PlaceIntent, protectiveMarketSell bool) bool
```

- **기본값(옵션 없음)** — `app.go`·`mcp.go`·시험의 모든 `NewService` 는 옵션을 켜지 않으므로 비분수 non-limit 은 오늘처럼
  `false` → `Place` 가 `ErrPlaceUnsupported`, `PreviewPlace` 의 `LiveReady`·경고 문구도 upstream 과 같다. 이 동일성을 시험으로
  고정한다(tasks 1.5.5).
- **엔진 인스턴스(옵션 켬)** — ②와 **같은 모양**으로만 연다: `sell + market + 가격 없음` → 지원. `buy + market` → 비지원
  (시험 고정). `sell + market + 가격 있음` → 비지원. fractional·limit 분기는 무변화.
- **두 호출자가 같은 값을 써야 한다** — `Place` 는 지원 판정 뒤 `guard(…, s.PreviewPlace(intent), …)` 를 부르고(손 읽기,
  검증 필요 — AST 미작성: task 1.5.1), 엔진 게이트웨이도 `PreviewPlace(i).ConfirmToken` 을 쓴다. 한쪽만 옵션을 보면
  `Place` 는 지원·`PreviewPlace` 는 비지원으로 갈라진다. 그러므로 옵션은 술어의 인자로 들어가고, 1.5.2 RED 가 엔진 인스턴스의
  `PreviewPlace(sell+market).LiveReady` 를 함께 잰다.
- **조기 통과 금지 (3차 재리뷰 A-P2-2 — P2-2 를 ③ 에 적용)**: 손 읽기로는 `OrderType != "limit"` 거부 **뒤에** 시장별 통화
  검사가 온다. `sell+market` 예외를 조기 `return true` 로 넣으면 그 통화 검사를 건너뛴다. 새 분기는 비지정가 거부의 예외 조건으로만
  서고, 뒤의 통화 검사(KR → KRW, US → KRW|USD)는 `sell+market` 에도 그대로 적용된다. 엔진 경로에는 ② 가 앞에 있지만 ③ 의 단독
  시험(1.5.2)은 ② 없이 이 규칙을 잰다 — 통화 행을 포함한다.
- **문구** — `PreviewPlace` 의 「Live place supports … limit orders … and US fractional (market) orders.」는 기본 인스턴스에서
  바이트 그대로 둔다. 엔진 인스턴스에서 그 문장을 바꿀지(또는 한 줄 덧붙일지)는 1.5.4 가 정하되, 기본 인스턴스의 문구 바이트는
  바뀌지 않는다. `checkOrderShape` 주석 「internal/trading is not ours to change — design D1」(execgw 측 설계의 D1)은 이 change 가
  `internal/trading` 에 옵션을 더하므로 「기본 동작은 upstream 그대로, 엔진 인스턴스만 옵션」으로 고친다(1.5.4).
- **불변식 3** — 엔진이 꺼져 있으면 엔진 인스턴스가 만들어지지 않으므로 이 개방의 흔적은 0 이다. CLI·ops·MCP 는 엔진 토글과
  무관하게 upstream 동일. `issues.md` I-R1 종결.
- **엔진 인스턴스 안의 다른 생산자** — ~~엔진 인스턴스를 지나는 매도 조립은 `sellIntent`·`flatten` 둘이고~~ **정정(0.10 독립 리뷰 P0-1)**:
  `tossctl flatten-all` 은 엔진 인스턴스가 아니라 `NewOrderPath` 로 **자기 인스턴스**를 만든다(`cmd/tossctl/flatten.go:200`). 엔진
  인스턴스를 지나는 보호 매도 조립은 `sellIntent` 하나이고(2차 재리뷰 P3-3, `flatten` 은 LIMIT) a112 전략 레인은 매수 전용이다. 그 열거를 1.5.1 risk-pattern-report 가 CodeGraph 로 다시 적는다(엔진
  `tradingService` 소비자 포함). 엔진 인스턴스를 쓰는 사람 조작 표면이 발견되면 그 표면은 엔진 토글 뒤에 있으므로 불변식 3 은
  유지되지만, 그 사실을 1.5.0 Pre-Edit 에 적는다.

**②와 ③의 판정이 갈라질 위험**: 같은 규칙이 두 자리에 산다. 한쪽만 바뀌면 다른 쪽이 그 시험을 대신 통과시킨다
([[two-judgements-cover-for-each-other]]). 그러므로 RED 는 두 관문을 **각각 단독으로** 잰다(한 관문을 우회한 직접 호출
시험 각 1벌) — tasks 1.1·1.5.2.

### ②·③ 손 읽기 주장의 AST 대조 (2026-10-28, HEAD `0c9bfd272` — 0.10 리뷰 증거)

위의 「검증 필요 — AST 미작성」 주장 넷을 HEAD AST(`analysis/evidence/execgw--checkordershape.ast.json` ·
`trading--placeintentsupported.ast.json` · `trading--service.place.ast.json` · `trading--service.previewplace.ast.json`, 열거
`analysis/evidence/enumerate-output.md`)로 대조했다. 이 대조는 구현 전 FLM(1.0.1·1.5.1)을 **대체하지 않는다** — 그 번들은 착수 시 HEAD 에서
다시 만든다. base `102d4e99` 이후 `failclosed.go` 의 변경은 `AllReasonCodes` 의 6줄(a112 `3260f4eb6` · a094 `766a8456c`)뿐이고
`checkOrderShape`·`trading/service.go` 본문은 base 와 같다(`git diff 102d4e99 HEAD`).

- **② 위치**: `checkOrderShape` B10(`:84` `orderType != "limit"` 거부) 뒤에 B11(`:88` KR 통화) · B12(`:91` 수량 `<= 0`) · B13(`:94` 가격 `<= 0`)이
  온다. 손 읽기와 같다 — `sell+market` 예외를 B10 의 조기 `return nil` 로 넣으면 B11·B12 를 건너뛰고, B13 은 `sell+market` 에서 `Price != 0` 거부로
  갈라야 한다.
- **③ 위치**: `placeIntentSupported` B3(`:281` 비지정가 `return false`) 뒤에 B4(`:284` KR → KRW) 와 마지막 반환(`:287` US → KRW|USD)이 온다. 손 읽기와 같다.
  ③ 에는 **수량·가격 검사가 없다**(분기 4개 전부 시장·분수·유형·통화).
- **정수 수량 (1.1a 입력)**: ② B12 는 `<= 0` 만, ③ 은 수량을 보지 않는다 — **어느 관문도 비분수 MARKET 의 정수성을 검사하지 않는다.** 1.1a 의
  처분(로컬 거부 추가 / 브로커 400 위임)은 이 사실에서 출발한다.
- **두 호출자**: `Service.Place` B1(`:188`)이 술어를 부르고 B4(`:200`)에서 `s.guard(…, s.PreviewPlace(intent), …)` 를 부른다. `PreviewPlace`(:109-143,
  분기 6)는 `:115` 에서 술어를 다시 부른다. 한쪽만 옵션을 보면 갈라진다는 위 「두 호출자가 같은 값」 주장과 같다.
- **문자열 정규화 차이**: ② 는 `ToLower`·`TrimSpace` 로 정규화하고(`:54-57`), ③ 은 `"kr"`·`"market"` 을 그대로 비교한다. 엔진 `sellIntent` 는
  `m.position.Market` 을 그대로 싣는다. 오늘 LIMIT 보호 청산이 ③ 을 지나므로 생산 값은 소문자다 — 새 판정(D7 「구간 입력」 조건 1)은 `clock.ParseMarket`
  으로 정규화하고, 실패는 LIMIT 이다.

### 0.10 독립 리뷰 반영 (2026-10-28, HEAD `3e9989f61`) — 엔진 전용 배선(P0-1) · KR 정수 수량(P1-2) — 8판에서 대체됨(P0-1 규칙은 판 생성 자리에 승계, P1-2 는 D11 (f))

근거 AST 는 `analysis/evidence/` 에 새로 만든 다섯(`engine--newcontext` · `tossctl--buildflattenwiring` · `execgw--canonicalprotectionquantity` ·
`official--buildordercreate` · `engine--floatof`)과 기존 `engine--neworderpath` · `execgw--checkordershape` · `engine--exitobserver.submit` ·
`engine--exitobserver.sellintent` 다. 소스 바이트는 `0c9bfd272` 와 같다(`enumerate.py` 재실행 stale 0, 25개). 호출 자리 수는 `rg` 와 AST 로 셌다.

**P0-1 — `NewOrderPath` 호출 자리 (비시험 코드, 호출 자리 단위)**

| # | 자리 | 감싼 함수 | 무엇을 만드나 |
| --- | --- | --- | --- |
| 1 | `internal/app/engine/engine.go:438` | `NewContext` (B1 `:439` 가 그 오류 반환) | 엔진 프로필. `path.Trading` 이 `buildGateway`(`:511`)와 `Context.tradingService`(`:603`)로 간다 |
| 2 | `cmd/tossctl/flatten.go:200` | `buildFlattenWiring` (B1 `:201`) | 사람 명령 `tossctl flatten-all` 의 자기 배선. `orderPath.Trading` 이 자기 게이트웨이(`:235`)로 간다 |

시험 코드에는 호출 자리가 0 이다. 문자열로 이름을 확인하는 시험이 둘 있다(`cmd/tossctl/flatten_test.go:141` — flatten 이 `engine.NewOrderPath(` 로 만든다는 구조
고정, `cmd/tossctl/engine_reconcile_test.go:112` — 금지 토큰 목록). `NewOrderPath` 안의 `trading.NewService` 생성은 `:401` 하나다. 같은 생성자의 다른 비시험 자리는
`internal/app/app.go:158` · `cmd/tossctl/mcp.go:74` 둘이다(1.5.5 가 이미 센 것). **그래서 `:401` 에 옵션을 붙이면 자리 2(사람 CLI)가 시장가 허용 인스턴스를 받는다** — 리뷰 지적이 맞다.

**새 배선**: 옵션은 공용 빌더가 아니라 엔진 조립에서만 덧입힌다.

```go
// NewOrderPath(engine.go:359-406) — 바이트 그대로. 옵션 없음.

// NewContext(engine.go:437-) — B1(:439) 바로 뒤, buildGateway(:509) 앞에 한 줄
path, err := NewOrderPath(opts)
if err != nil { return nil, err }
path.Trading.WithProtectiveMarketSell() // 엔진만 — flatten-all 은 NewOrderPath 를 따로 불러 새 인스턴스를 받는다
```

- `WithX` 는 받은 `*Service` 를 고쳐서 돌려주는 기존 관례다(`WithLineage` `service.go:69` · `WithConditional` `conditional.go:60`). `NewOrderPath` 는 호출마다
  `trading.NewService` 를 새로 부르므로(`:401`) `NewContext` 가 고친 인스턴스는 자리 2 와 공유되지 않는다. 같은 인스턴스가 게이트웨이(`:511`)와
  `tradingService`(`:603`)에 함께 가므로 엔진 안의 두 소비자는 같은 값을 본다.
- 이제 고치는 기존 함수는 `NewOrderPath` 가 아니라 `NewContext` 다(gate ⑤ FLM 대상 교체 — 「착수 시 FLM 대상」 목록 정정은 `review.md` 0.10 독립 리뷰 처분).
- 불변식 3: 엔진이 꺼져 있으면 `NewContext` 가 불리지 않는다. 자리 2·`app.go`·`mcp.go` 는 옵션 0 이다.

**시험 (tasks 1.5.5 개정)**: (c) 구조 시험은 **옵션 호출 자리가 비시험 코드 전체에서 `NewContext` 본문 하나뿐**이고 `NewOrderPath` · `buildFlattenWiring` ·
`app.go` · `mcp.go` 에는 0 임을 함수 단위 AST 로 센다. (d) 행동 시험 — `engine.NewOrderPath`(임시 config dir · 시험 자격 증명, 네트워크 없음 — 그 함수 문서
「It performs no network call」)가 돌려준 `Trading.PreviewPlace(KR sell+market).LiveReady == false`, 그리고 `buildFlattenWiring` 이 게이트웨이에 넘기는
인스턴스도 같은 판정. (e) 엔진 시험 — `NewContext` 로 만든 엔진의 게이트웨이 경로에서 같은 의도가 `LiveReady == true`. 변이: 「`NewOrderPath` 반환식에 옵션」·
「`buildFlattenWiring` 에서 `orderPath.Trading` 에 옵션」(= flatten 경로에 옵션 켜기) 둘은 (c)(d) 가, 「`NewContext` 에서 옵션 제거」는 (e)·1.6 이 잡아야 한다.

**P1-2 — KR MARKET 은 정확한 양의 정수 수량일 때만**

- 사실(AST): `checkOrderShape` B12(`:91`)는 `<= 0` 만, ③ `placeIntentSupported` 는 수량을 안 본다. `buildOrderCreate` B5(`:136`)도 `<= 0` 만 보고
  `formatDecimal` 로 그대로 싣는다. `floatOf`(B1 `:2016`)는 `ParseFloat` 실패만 거른다. 정수 검사는 `canonicalProtectionQuantity`(`protection.go:112-123`,
  B1 소수점·지수·부호 문자, B2 0·2^53-1 초과·왕복 불일치)에 있고 보호 진입 readiness 만 쓴다.
- **판정 조건으로 넣는다(1차 방어).** D7 판정 규칙에 조건 6 「시장이 KR 이면 `sellIntent` 가 싣을 **그 float 수량**이 정확한 양의 정수」를 더한다. 거짓이면
  MARKET 자격이 없고 **오늘의 LIMIT 경로**(가격 사다리 그대로)로 간다 — 거부가 아니다. 소수 KR 수량의 LIMIT 은 오늘도 같은 주문이라 약화가 아니다.
  이 경우 관측 줄에 판정 사유 `kr-quantity-not-integral` 을 남긴다(입구 정수 불변의 위반 신호 — 새 경보 경로는 만들지 않는다. 입구 증명은 원장 투영 소관이라 범위 밖).
- **술어는 하나다.** 판정과 ② backstop 이 같은 함수를 부른다 — `canonicalProtectionQuantity` 를 execgw 에서 내보내 쓴다(엔진은 이미 execgw 를 import 한다,
  `exitloop.go:87`). 같은 규칙을 두 자리에 쓰면 서로의 시험을 대신 통과시킨다([[two-judgements-cover-for-each-other]]). ③(`internal/trading`)은 execgw 를
  import 하지 않으므로 그 자리의 검사는 따로 서고, 1.5.2 가 단독으로 잰다.
- **2차 방어(backstop)**: ② 의 `sell+market`(비분수) 갈래와 ③ 엔진 인스턴스의 `sell+market` 예외에 정수성 검사를 더한다(B10 예외 조건 안 — 조기 통과 금지 유지).
  소수 KR 수량은 **전송 본문에 닿지 않는다**.
- **제출 실패 갈래가 제안을 해제하는 문제**: backstop 이 거부하면 게이트웨이 결과는 `submit` B13(`:1582` default)로 가서 alert · `ProposalRefused` 해제 →
  다음 관측이 다시 발의한다. 판정이 그 수량에 MARKET 을 계속 고르면 「거부 → 해제 → 재발의」가 반복되고 보호 청산은 0 이 된다(오늘보다 약함). 그래서
  backstop 은 **도달 불가여야 한다** — 판정이 같은 술어로 같은 float 값을 먼저 보기 때문이다. 이것을 시험으로 고정한다: 수량 표(`1` · `1.5` · `1e-9` ·
  `2^53+1` · `0` · `-1` · `NaN` · `+Inf` · `"1.0"` · `" 2 "`)에서 「판정 = MARKET ⇒ ②·③ 통과」, 그리고 소수 KR 행의 엔진 종단 결과가 **B13 해제가 아니라 LIMIT
  제출**이다. backstop 이 거부한 경우의 관측 사유는 기존 `ReasonInvalidRequest` 와 다른 문구로 남겨 「판정·관문 불일치(버그)」로 읽히게 한다.
- 변이: 「판정에서 정수 조건 제거」(종단 시험이 B13 해제 또는 wire 소수로 잡음) · 「② backstop 제거」 · 「③ backstop 제거」(각 관문 단독 시험이 잡음) ·
  「판정이 다른(문자열) 수량으로 정수성 판정」(float 왕복 행이 잡음).

### 왜 진입은 계속 막는가

`riskcalc`의 근거가 진입에서는 여전히 옳다 — 시장가 매수는 노출 평가가가 없고, Guardian이
예약할 금액을 계산할 수 없다. 그 규칙을 이 change가 건드리지 않는다는 것을 **별도 테스트로
고정**한다. 안 그러면 다음 사람이 "a087이 시장가를 열었다"로 읽는다.

## D4 — 원장이 무엇을 기록하는가

> **8판에서 대체됨(가격 칸)**: 8판의 `intents.price` 는 비지 않는다 — 보낸 가격(창 안 보호 매도면 하한가, 그 밖이면 오늘의 관측가 · 기준선)이다. 아래 「원장은 실제로
> 보낸 것을 적어야 한다」는 원칙은 그대로이고 그 원칙의 8판 적용이 D11 (b) 「기록」이다. 「화면은 `—`가 아니라 "시장가"」 요구(tasks 3.4)는 생산할 가격 없는 청산이 없어
> 8판에서 대상이 없다.

시장가는 제출 시점에 가격이 없다. 원장의 `intents.price`는 빈 문자열이 된다.

**이것이 정보 손실처럼 보이지만 아니다.** 지금 그 칸에 들어가는 값은 관측가이고, 그것은
체결가가 아니며, 실측 5건에서는 **주문이 되지도 못한 값**이었다. 없는 가격을 적는 것보다
비우는 것이 정직하다.

체결가는 `filldetect`가 체결 조회로 확정한다 — 이미 그렇게 동작한다면 변경 없음을
테스트로 고정하고, 아니면 갭으로 기록한다(tasks 3.5).

초안 리뷰 M3이 "정렬된 제출가를 원장에 남겨라"고 한 것과 방향이 반대로 보이지만 같은
요구다: **원장은 실제로 보낸 것을 적어야 한다.** 초안에서는 정렬된 가격이 실제로 보낸
것이었고, 여기서는 가격이 없는 것이 실제로 보낸 것이다.

화면·알림은 빈 값을 결측으로 그리면 안 된다. `—`가 아니라 "시장가"다.

## D5 — StockOS 대조 (승계와 미승계)

> **8판 정정(첫 행 판정)**: 8판의 a087 1차 보호 매도는 **하한가 지정가**다. StockOS 1차 손절도 LIMIT 이지만 가격 규칙이 다르므로(StockOS 는 0.5% 초과 이탈 · EOD 에서만
> MARKET) 첫 행은 여전히 **이탈**이다 — 근거는 사용자 결정(2026-10-28 D-1 (다))이고 StockOS 는 근거가 아니다. 「US = MARKETABLE_LIMIT」 행 아래의 「양 시장 모두
> MARKET으로 간다」는 7판에서 이미 철회됐고(Q2) 8판에서도 US 는 오늘 LIMIT 이다.

> **정정 (재리뷰 P1-2 · 2차 리뷰 C2)**: 3판의 첫 행은 「KRX 긴급 청산 = MARKET → **승계**」였다. StockOS 의 평범한
> 손절 1차 제출은 LIMIT 이고 MARKET 은 0.5% 초과 이탈·EOD 에스컬레이션에서만 쓰인다(`_apply_exit_order_style`).
> a087 은 1차부터 MARKET 이라 그 행은 승계가 아니라 **이탈**이다. StockOS 는 a087 의 근거가 아니다(proposal 「StockOS 대조」).
> 「a088로」·「a089로」 행은 대상이 없다(재리뷰 P3-1) — **무소유**로 고친다.

| StockOS | 위치 | a087 판정 |
| --- | --- | --- |
| 1차 손절 = LIMIT, 0.5% 초과 이탈·EOD 에서만 KRX MARKET | `auto_exit_execution.py` `_apply_exit_order_style`·`_aggressive_exit_order_type` | **이탈** — a087 은 임계 없이 1차부터 MARKET. 근거는 사용자 결정·§5 실측(D6) |
| 긴급 청산은 지연 게이트 전부 우회 | `auto_exit_execution.py:3130` | **무소유** |
| N회 후 MARKETABLE_LIMIT → true MARKET 승격 | `auto_exit_execution.py:3160` | **무소유** |
| US = MARKETABLE_LIMIT | 같은 파일 | **미승계** — KIS 계약 제약이고 토스에는 없다 |
| 호가 정렬 `Decimal(str(price))` | `auto_exit_execution.py:3001` | **무소유** |
| 정렬 실패 시 fallback, 거부 없음 | `auto_exit_execution.py:3005-3017` | **무소유** |
| `TickSizeProvider` + symbol override | `tick_size_provider.py` | **무소유** |

StockOS가 US에 marketable limit을 쓰는 것은 `protective_order_capability`가 기록한 KIS의
한계 때문이다 — `kis_us_stock_rest_stop_oco_not_supported`. 토스 openapi에는 그 제약이
없고 US 온주 MARKET 매도를 금지하는 문장이 없다. 그래서 양 시장 모두 MARKET으로 간다.
`[미측정 — US MARKET 매도 실주문 없음]` — 이 한계의 처분은 proposal What Changes §1 「US 한계」.

## D6 — 거래소 하한가 지정가 원안 (미채택 기록 — 재리뷰 P1-3, 3차 A-P2-3 정정)

> **8판에서 채택됨 (사람 결정 2026-10-28 「3은 3」 — `issues.md` D-1 (다))**. 아래 「미채택」 처분과 그 사유는 7판까지의 기록이다. 표의 두 칸은 8판 설계가 고쳤다 —
> 「브로커 호출: 보호 청산마다 `PriceLimits` GET 1회」 → **제출 경로 GET 0**, 백그라운드 판(종목당 하루 1 GET — D11 (b)). 「시간외단일가: 가격 범위 밖일 가능성」 →
> **창 밖이라 오늘 LIMIT**(D11 (c)). 원안의 범위는 「KR 정규장 보호 청산의 가격 = 그날 하한가」로 확정됐다(D11 (a)). I-R3(bps 재비교)는 비교 대상(MARKET)이 없어져
> 3.8 질의가 「하한가 지정가 vs 관측가」 체결 품질 기록으로만 남는다.

**원안은 평가된 적이 없다.** 2026-08-06 2차 리뷰 C3·「3차 교정 방향」 2 는 보호 청산의 가격으로 **거래소 하한가
지정가**를 쓰라고 했다(`flatten` 이 이미 그렇게 한다 — `internal/flatten/liquidate.go` 「The exchange's own floor」).
D2a 는 그것을 관측가·기준선 **뒤의 셋째 폴백 단**으로 좁혔고, 그 폴백이 도달 불가로 반증됐다(D2a 「반증」). 원안 —
보호 제안에서 관측가 **대신** 하한가 — 은 그 반증과 무관하게 비교되지 않은 채 남았다(proposal 이 「그리드 불일치 사건(옛 표기 「9분 사건」)은 Phase 1 이
고치지 않는다」고 적은 것이 그 증거다).

재리뷰가 적은 원안의 성질(비교 기록 — 측정하지 않은 값은 표기대로):

| 축 | 하한가 지정가 원안 | 시장가 (a087) |
| --- | --- | --- |
| 세 관문 | 전부 통과(여전히 LIMIT — ③ 변경 불요) | ②·③ 변경 필요(D3) — ③ 은 엔진 인스턴스 옵션, upstream 동작 변화 0(2차 결정) |
| 245,750 사건 | 해소(거래소 공표값 = 온그리드) | 해소(가격 없음) |
| 원장 가격 | 남는다 | 비운다(D4·§3 화면 작업 필요) |
| 시장 | **KR 전용**(`client/marketdata.go` 「미국장은 일일 가격제한 제도가 없음」) | KR 만(7판 — US 는 실측 전까지 LIMIT, D7 Q2) |
| 브로커 호출 | 보호 청산마다 `PriceLimits` GET 1회(§0.4) | 추가 0 |
| 시간외단일가 | 가격 범위 밖일 가능성 `[미측정]` | MARKET 불가 → 오늘과 같은 LIMIT(D7 처분 2026-10-22) |
| 체결 품질 | 하한가 도달·매수 호가 공백 시 미체결 | 같음(D8) |

**처분: 미채택 — 시장가 방향 사용자 결정(2026-10-10)에 따른 것이며, 원안 자체 평가는 미수행.** (정정 — 3차 재리뷰
A-P2-3: 4판은 「기각 (사용자 결정)」이라 적었으나 사용자 결정 기록에 원안은 나오지 않는다. 「(가)를 골랐으니 (다)는 기각」은
추론이었다.) 재리뷰가 P0-1 의 처분으로 (가) ③ 수정 (나) 우회 (다) 이 원안을 제시했고, 사용자는 (가) — **시장가 방향 유지** —
를 택했다(2차 결정으로 ③ 개방은 엔진 인스턴스 한정, D3). 이 표는 그 결정이 무엇을 받아들였는지의 기록이다(원장 가격 대신
③ 의 엔진 전용 옵션·원장/화면 작업, 대신 US 동시 해소). 「세 관문」 행의 비용은 2차 결정으로 줄었다 — upstream 동작 변화가
없어졌고 남는 비용은 `internal/trading` 의 옵션 하나다. 재개 조건: §5.1 이 MARKET 매도를 거부하면(tasks §5 표) 이 원안이
재검토 후보가 된다 — 그 판단은 사람 결정이다. 원안의 문언 기각을 받을지는 사람 판단으로 남긴다(`issues.md` I-R3).

**계기 (재리뷰 P1-3 · 2차 리뷰 「3차 교정」 3·4)**: 현재 `slippagePct` 는 진입 전용이고 청산 슬리피지를 재는 코드가
없다(재리뷰 인용). §5.1 은 접수만 재고 체결 품질을 재지 않는다. 그러므로 시장가가 원안보다 나은지는 착지 뒤 데이터로만
답할 수 있고, 그 데이터를 남기는 계기를 tasks 3.8 에 둔다(보호 청산 체결가 vs 결정 시 관측가, bps). bps 로 두 방식을
다시 비교할지와 그 임계는 이 change 에서 정하지 않는다(`issues.md` I-R3).

## D7 — 세션 교집합: MARKET 불가 ∩ LIMIT 금지 = 제출 불가 (재리뷰 P1-4)

> **8판에서 대체됨 — 세션 교집합 문제는 생기지 않는다.** 8판에는 MARKET 칸이 없어 「MARKET 불가 ∩ LIMIT 금지」 구간이 없다. 이 절의 처분 표 · 「구간 입력」(브로커 달력 +
> 고정 시각 · 판정 조건 1~6 · 스냅숏 형) · 「KRX 단독 판정 입력 조사」는 7판의 기록으로 남는다. 8판에서 쓰는 것은 셋뿐이다 — ① 0.5 의 사실(엔진에 세션 게이트 없음 →
> 시각 입력을 새로 배선해야 한다) ② 「즉시성 제약」의 원칙(제출 경로 GET 0 · 판정 실패 = 오늘 LIMIT · 갱신은 관측 주기 밖) ③ 고정 시각 술어 `clock.MarketKR.InRegularSession`
> (AST `clock--market.inregularsession` 등). **브로커 달력은 8판에서 읽지 않는다** — 하한가 지정가는 휴장 · 조기 폐장 · 개장 지연 · VI 에서도 같은 LIMIT 유형이라 그 날을
> 가를 필요가 없고, 가격이 틀린 경우는 되먹임이 묶는다(D11 (c)·(b)). 그래서 `official.MarketCalendarSession` 형 확장 · `scheduler` 대조 · 달력 갱신은 하지 않는다.
> §5.2 처분(not-applicable)은 그대로다 — 8판에는 정규장 밖 MARKET 이 없다.

spec Requirement 1 은 보호 청산의 지정가를 **금지**한다(SHALL NOT). 브로커가 MARKET 은 받지 않고 LIMIT 은 받는
구간에서는 두 조건의 교집합으로 보호 청산의 **제출 가능성이 0** 이 된다:

| 구간 | MARKET | LIMIT | a087 후 보호 청산 |
| --- | --- | --- | --- |
| KRX 정규장 | §5.1 측정 대상 | 가능(오늘) | MARKET — §5.1 결과에 달림 |
| KRX 시간외단일가 | **없음**(거래소 제도 — proposal 스스로 인정) | 가능 | **제출 불가** |
| KRX 장 마감 후·SOR/확장 세션 | `[미측정]` — §5.2 | `[미측정]` | §5.2 결과에 달림 |
| US 정규장 | `[미측정]` | 가능(오늘) | 미측정 착지(proposal 「US 한계」) |
| US 프리/애프터 | `[미측정]` | `[미측정]` | 미측정 — MARKET 불가이면 **제출 불가** |

**엔진의 세션 게이트 유무: 미해결, §1~§3 착수 전 확인.** 재리뷰는 HEAD 에서 `InRegularSession`(`clock/market.go`)의
비시험 호출자가 `verifylive/hours.go` 하나뿐이고 `internal/app/engine` 에는 0 이라고 적었다(CodeGraph·`rg`). 그러나
엔진이 다른 수단(관측 루프 기동 시간대·시세 만료 판정 등)으로 정규장 밖 판정을 사실상 막는지는 **이 문서가 모른다**
(검증 필요 — 새 분기 주장 아님). 착수 전 확인 방법: 엔진의 청산 판정 상류(`ObserveOnce`·`judge`)를 CodeGraph 로 따라가
세션·시각 입력이 있는지 열거한다(tasks 0.5).

처분은 그 확인과 §5.2 결과 **둘을 입력으로** 사람이 정한다. 문서가 미리 정하지 않는다. 다만 무효화 대상은 지금 명시한다 —
교집합 구간이 실재하면(엔진이 그 구간에 판정하고, 브로커가 MARKET 을 거부하면):

- spec Requirement 1 의 「지정가로 제출해서는 안 된다(SHALL NOT)」는 그 구간에서 **보호 청산을 0 으로 만든다** → 세션별
  유형(정규장 MARKET, 그 밖 LIMIT) 같은 축소가 필요해지고, 그것은 새 설계(세션 입력 배선)라 이 개정이 하지 않는다.
- tasks 2.4 의 「보호 제안일 때 가격을 읽지 않는다」는 그 구간에서 LIMIT 가격이 필요하므로 무효가 된다.
- 엔진이 그 구간에 판정하지 않음이 확인되면 교집합은 공집합이고 위 두 줄은 유지된다.

### 처분 (사용자 결정 2026-10-22) — 시장가를 받지 않는 구간은 오늘처럼 지정가 — **8판에서 대체됨**(8판의 폴백 칸 = 「오늘처럼 지정가」 원칙은 승계)

입력 둘이 나왔다. ① tasks 0.5(2026-10-20, `analysis/gate-0.5-session-inputs.md`): 청산 판정 상류 AST 12개에서 **세션 입력 0 ·
시각 입력 40(전부 경과 시간)** — 엔진에 세션 게이트가 없다. 정규장 밖에도 시세가 오면 판정·제출 시도가 일어난다. 그래서 위
「교집합 공집합」 줄은 성립하지 않는다. ② §5.1(2026-10-19, M50): KRX 정규장 MARKET 매도 1주 접수·체결완료. §5.2 는 미측정이다.

Manager 질문 「시장가가 안 되는 시간대에 손절을 어떻게 낼지. 권장: 그 시간대에는 지금처럼 지정가」에 사용자가 그 권장대로 답했다.
그래서 spec Requirement 1 을 다음처럼 개정했다(requirement 수준 변경 — WORKFLOW 「리뷰 게이트」에 따라 수정분 리뷰를 다시 받는다, tasks 0.10).

> **7판 개정 (사용자 결정 2026-10-28 「권장대로」 — `issues.md` 「사용자 결정 (2026-10-28)」)**: 아래 표는 Q2·Q3·Q4 를 반영한
> 판본이다. 바뀐 칸은 넷이다 — ① **US 정규장 MARKET → LIMIT**(Q2, 실측 전까지), ② KRX 정규장 MARKET 칸을 **브로커 달력이 정상 영업일로
> 보이는 날**로 좁혔다(Q3 — 휴장·조기 폐장·개장 지연일은 종일 LIMIT), ③ 「§5.2 가 접수를 보이기 전까지」 문구를 **LIMIT 확정**으로 바꿨다
> (Q4 — §5.2 not-applicable), ④ 종가단일가(15:20~15:30)를 LIMIT 으로 뺐다(Q1 원칙 적용 — `review.md` 0.10 P2-4). 6판 표의 US 정규장 행 「MARKET — 미측정 착지」는 **철회**됐다. 구간 판정 규칙은 아래 「구간 입력」 절.

| 구간 | 보호 청산 유형 | 근거 |
| --- | --- | --- |
| KRX 정규장 접속매매 — 브로커 달력이 그날을 **정상 영업일**로 보이고(아래 「구간 입력」 조건 1~5) 그 시각이 09:00~15:20 KST 안 | MARKET | §5.1 접수 실측(M50, 14:59 접속매매) |
| KRX 정규장 종가단일가(15:20~15:30) | **LIMIT(오늘과 같음)** | Q1 원칙(접수 미측정 → LIMIT)을 이 하위 구간에 적용 — §5.1 은 이 구간을 재지 않았다. 문서 유도, Manager 확인(`review.md` 0.10 P2-4) |
| KRX 휴장일 · 조기 폐장일 · 개장 지연일 (달력의 그날 정규장이 정상 영업일 모양이 아님) | **LIMIT — 그날 종일** | Q3(조기 폐장일 13:00~15:30 MARKET 누출 방지) + 아래 「구간 입력」 해석 |
| KRX 시간외단일가 | **LIMIT(오늘과 같음)** | 거래소 제도상 시장가 없음 |
| KRX 장 마감 후·장 개시 전 그 밖의 구간 | **LIMIT 확정** | 사용자 결정 2026-10-22 · Q4(§5.2 not-applicable — 이 칸을 옮길 실측을 하지 않는다) |
| US 정규장 | **LIMIT — 실측 전까지** | Q2(2026-10-28). US MARKET 접수 실측 0(M12 는 조건주문). 6판의 「미측정 착지」 철회 |
| US 프리/애프터 | **LIMIT** | 사용자 결정 2026-10-22 · Q2 |
| 구간 판정 입력을 못 읽음 — 달력 오류·결측·낡음·시계 역행·시장 문자열 판정 불가 | **LIMIT** | Q1(불변식 4 유도 수락) · Q3(달력 오류·결측이면 LIMIT) |

LIMIT 의 가격은 오늘 `sellIntent` 사다리 그대로다(관측가 → 기준선 → 거부 — `analysis/function-logic/internal-app-engine--exitobserver.sellintent/`
base `102d4e99` 산출물. HEAD `0c9bfd272` AST `analysis/evidence/engine--exitobserver.sellintent.ast.json` 도 분기 4·반환 4 로 같다. 구현 때 HEAD 에서 FLM 으로 다시 만든다 — tasks 2.1).

**불변식 4 근거 (약화 없음).** 기준선은 오늘의 동작(보호 청산은 언제나 LIMIT)이다. 이 처분은 오늘보다 나빠지는 칸을 만들지 않는다.

- MARKET 칸은 접수가 실측된 구간 하나 — 정상 영업일의 KRX 정규장 접속매매(09:00~15:20) — 뿐이다(7판: US 정규장 칸 삭제 Q2 · 종가단일가 제외 P2-4).
- 그 밖의 칸은 LIMIT 이라 오늘과 바이트 같은 주문을 낸다.
- 접수가 실측되지 않은 구간과 판정 입력이 없는 경우를 LIMIT 으로 둔 것은 이 문서의 유도다. 사용자 문언은 「시장가가 안 되는 시간대」다.
  모르는 구간에 MARKET 을 내면 거부될 수 있고, 같은 구간에 LIMIT 은 오늘 접수될 수 있다. 그러면 그 칸은 오늘보다 약해진다(§0.3 위반).
  그래서 모름은 LIMIT 으로 기운다. **이 유도는 사용자가 수락했다(2026-10-28 Q1).**
- 반대 방향(실제로는 MARKET 을 받는 구간인데 LIMIT 을 냄)은 오늘과 같은 동작이다. 그러므로 약화가 아니다. 이득을 덜 얻을 뿐이다.

**무효화된 줄.**

- 위 「spec Requirement 1 … 보호 청산을 0 으로 만든다」 — 개정으로 해소. 교집합 구간에서 보호 청산은 LIMIT 으로 나간다.
- tasks 2.4 「보호 제안일 때 가격을 읽지 않는다」 — **MARKET 구간에서만** 참이다. LIMIT 구간에서는 오늘처럼 가격을 읽는다. tasks 2.4·2.5 를 그에 맞춰 고쳤다.

**새 설계 의무 (구현 task 2.6).** 엔진은 오늘 세션 입력을 하나도 읽지 않는다(0.5). 그래서 구현은 MARKET/LIMIT 을 가를 **구간·접수
입력**을 새로 배선해야 한다. ~~이 입력의 출처는 이 문서가 정하지 않는다.~~ 출처는 7판에서 정해졌다(아래 「구간 입력」, Q3).
`sellIntent` 시그니처에 시각 또는 구간 인자가 더해지면 2.1 FLM 의 분기 집합이 바뀐다. 그래서 2.1·2.6 은 같은 로트에서 재생성한다.

### 구간 입력 (7판 — 사용자 결정 2026-10-28 Q3) — **8판에서 대체됨**(D7 머리말 · 브로커 달력을 읽지 않는다)

**출처 = 브로커 시장 달력 + 고정 시각.** 달력 읽기는 `official.Client.TypedMarketCalendar`(`internal/official/typed_calendar_reads.go:37`)이고,
엔진 진입 쪽 `strategyScheduleAuthorityLoader.prepare`(`strategy_schedule_authority.go:220`)가 이미 같은 함수를 부른다. **달력 오류·결측이면 LIMIT.**
분기 근거는 손 읽기가 아니라 HEAD `0c9bfd272` 의 AST 다 — `analysis/evidence/*.ast.json`, 열거 `analysis/evidence/enumerate-output.md`
(재생성 `python3 -I …/analysis/evidence/enumerate.py`, `source_sha256` 대조 stale 0).

**`clock` 술어만으로는 안 되는 이유 (AST).** `Market.RegularSession`(`clock/market.go:122-143`)의 분기는 B1(`Location` 오류)·B2(시장 표 없음)
둘뿐이고 성공 반환(:136)은 고정 표 `regularHours`(:55-58, KR 09:00~15:30)로 창을 만든다. `Session.Contains`(:113-118)는 B1 에서 주말·0 값만 거른다.
휴장일·조기 폐장·개장 지연을 가르는 분기는 **없다** — 파일 주석(:49-54)도 휴일은 브로커 달력 몫이라 적는다. 그래서 `InRegularSession` 만 쓰면
조기 폐장일 13:00~15:30 에 MARKET 이 나간다(Q3 의 누출).

**달력 함수의 분기 (AST).**

- `TypedMarketCalendar`: B1 나라 문자열(KR·US 외 오류) · B2 날짜 검증 · B4 GET 오류 → 오류 반환. 응답을 해석하지 않고 그대로 돌려준다.
- `scheduler.adaptCalendarDay`(`scheduler/calendar.go:103-134`): KR 은 B4·B5 에서 `Integrated.RegularMarket` 만 읽는다. B8 — 세션이 없으면
  `Regular == nil` 인 **정상 반환**(오류 아님 — 휴장일 모양). B9 — 시작·끝 0 값·역순 오류, B10 — 거래소 현지 날짜 밖 오류.
- `scheduler.AdaptOfficialCalendar`(`:56-101`): B6 이전 < 오늘 < 다음 영업일 순서, B7 이전·다음 영업일의 정규장 존재를 요구한다. **오늘의
  `Regular == nil` 은 통과한다**(휴장일). `EarlyClose`(:132)는 길이가 6h30m 미만인지만 본다 — 개장 지연(길이 같고 시각이 밀림)은 못 가른다.
- `CalendarSnapshot.ValidityAt`(`:138-149`): B1 시계 역행 · B2 6시간 신선도 · **B3 `REFRESH_TOO_LATE` — 정규장 중에 장 개시 뒤 읽은 달력을 거부**.

**openapi 의 KR 정규장 의미 (`docs/migration/openapi.latest.json` 스키마 `RegularMarketSession`·`IntegratedHour`)**: 달력의 KR 정규장은
「**KRX·NXT 정규장의 합집합**(가장 이른 시작 ~ 가장 늦은 종료). 종가단일가 구간을 포함」이다. 같은 스키마의 `singlePriceAuctionStartTime` 은
「정규장 내 종가단일가 구간 시작 (**KRX 기준**). KRX 휴장이면 null」이다. 즉 합집합 창은 **KRX 자신의 조기 폐장을 숨길 수 있다**(NXT 가 정상
운영하면 합집합은 정상 모양). 그리고 현재 Go 형 `official.MarketCalendarSession`(`typed_calendar_reads.go:11-14`)은 `startTime`·`endTime` 둘만
읽는다 — KRX 기준 필드가 없다. openapi 예시에 조기 폐장일 표본은 없다(예시 요약: 영업일 · 휴장일 · 부분 휴장 — `[미측정]`).

> **보류 (0.10 독립 리뷰 P1-1, 2026-10-28)**: 아래 규칙은 「정상 영업일 KRX 접속매매」를 증명하지 못한다는 판정을 받았다 — KRX 단독 개장 시각·VI·
> 거래정지를 가르는 입력이 없다. 조사 결과와 선택지는 아래 「KRX 단독 판정 입력 조사」 절이고, **사람 결정 전까지 이 규칙으로 착수하지 않는다**.
> 조건 6(정수 수량, P1-2)은 결정과 무관하게 MARKET 이 남는 모든 선택지에 붙는다.

**판정 규칙 (MARKET 은 여섯 조건이 모두 참일 때만, 하나라도 거짓·불명이면 LIMIT).**

1. 시장이 KR 이다 — `clock.ParseMarket` 으로 정규화. 실패 → LIMIT. US 는 Q2 로 언제나 LIMIT 이므로 US 달력은 읽지 않는다.
2. 엔진 메모리에 KR 달력 응답이 있고, 그 응답이 `AdaptOfficialCalendar` 검증을 통과한다(위 B1~B8 — 오류면 LIMIT).
3. 응답의 신선도: 읽은 시각이 `now` 보다 미래가 아니고 읽은 지 6시간 미만(`ValidityAt` B1·B2 와 같은 값·같은 의미). **`ValidityAt` 을 그대로
   쓰지 않는다** — B3 `REFRESH_TOO_LATE` 는 진입 활성화의 규칙이고, 청산에 쓰면 6시간 상한과 겹쳐 KRX 정규장 끝 구간이 매일 LIMIT 으로 강등된다
   (장 개시 전에 읽은 달력은 늦어도 장 개시 6시간 뒤 낡는데 정규장은 6시간 30분이다). 강등은 안전 약화가 아니라 이득 손실이지만, 규칙을 가져오는
   이유가 청산에 없다.
4. 응답의 오늘 날짜가 `now` 의 KST 날짜와 같고, 그날 KR 정규장이 **정상 영업일 모양**이다 — 시작 09:00 · 끝 15:30 KST(합집합) **그리고**
   KRX 기준 종가단일가 시작 15:20 KST. 이 셋 중 하나라도 다르거나 없으면(휴장 `integrated` null · 조기 폐장 · 개장 지연 · KRX 휴장) **그날 종일 LIMIT**.
   KRX 기준 필드를 읽으려면 `official.MarketCalendarSession` 에 `singlePriceAuctionStartTime` 을 더해야 한다(타입 확장 — 2.6.4).
5. 고정 시각: `now` 가 `[09:00, 15:20)` KST 안이고 주말이 아니다. 시작과 주말은 `clock.MarketKR.RegularSession(now)`(`Open` · `Weekend`)를 쓰고, 끝 15:20 은
   종가단일가 시작 상수다 — `Session.Contains` 의 끝(15:30)은 종가단일가를 포함하므로 그대로 쓰지 않는다(위 표 종가단일가 행).
6. (0.10 독립 리뷰 P1-2) 시장이 KR 이면 `sellIntent` 가 싣을 float 수량이 정확한 양의 정수다 — execgw 의 정수 술어(`canonicalProtectionQuantity` 를 내보낸 것)
   하나로 판정한다. 거짓이면 LIMIT(D3 「0.10 독립 리뷰 반영」 P1-2).

**판정 함수의 모양 (0.10 독립 리뷰 P2-1 · 보완 조건 「B5 와 캐시 미준비 구분」)**: 판정은 `(market bool, reason string)` 를 돌려주는 **총함수**다 — error 를 돌려주지
않고, 패닉하지 않고, 브로커·디스크를 부르지 않는다(GET 0 · 비차단). 입력은 엔진 메모리의 **불변 스냅숏 한 개**(원자 포인터 교체로 갱신 — 읽는 쪽은 잠금 대기 없음)와
`now` · 수량뿐이다. 그래서 두 실패가 갈린다.

| 상황 | 판정 | 그다음 | `submit` 갈래 |
| --- | --- | --- | --- |
| 달력 캐시 미준비(기동 직후 · 갱신 실패 · 낡음 · 미래 `fetchedAt` · 날짜 불일치) | `false`, `calendar-not-ready` 등 | 오늘의 LIMIT 사다리 | 오늘과 같다(가격이 있으면 Place, B5 는 오늘의 사유로만) |
| `sellIntent` 자체 실패(가격 없음 · 수량/가격 파싱 불가 — `sellIntent` B2·B3·B4) | 판정과 무관 | — | B5(`:1533`) → alert · 해제 — **오늘과 같은 사유만** |

판정이 B5 에 닿는 새 사유를 만들면 안 된다 — 시험은 「캐시 미준비 행에서 B5 미도달 · 결과 intent = 오늘 LIMIT 골든」과 「GET 호출 계수 0」을 함께 잰다.
갱신 goroutine 과 판정의 경합은 `-race` 아래 시험한다(스냅숏 교체 중 판정이 반쪽 스냅숏을 보지 않음 · 스냅숏 nil 이면 LIMIT).

**스냅숏 형 (보완 조건 「raw → 검증 snapshot → 판정 보존」)**: KRX 기준 필드는 `official.MarketCalendarSession` 에 `SinglePriceAuctionStartTime *time.Time` 을 더해
디코드한다(디코드 전용 — 그 형을 그대로 출력하는 비시험 자리는 `rg` 로 0). 그러나 `scheduler.CalendarDay` 에는 **더하지 않는다** — `AdaptOfficialCalendar` 는
`CalendarDay` JSON 의 sha256 을 `Version` 으로 쓰고(`calendar.go:85-98`), 그 값이 콘솔 `CalendarVersion`(`marketschedule.go:105`)과 a112 전략 권한 수집에 쓰인다.
필드를 더하면 같은 달력의 Version 이 바뀐다. 그래서 a087 판정 스냅숏은 엔진 새 파일의 자기 형이다 — 같은 응답을 `AdaptOfficialCalendar` 로 검증(오류면 스냅숏 없음)한 뒤
그 결과의 오늘 정규장 창 · `FetchedAt` 과 원 응답의 `Today.Integrated.RegularMarket.SinglePriceAuctionStartTime` 을 한 값으로 묶는다. 시험: openapi 예시 응답으로
만든 `Version` 이 필드 추가 전후 바이트 같음.

「고정 시각」의 이 해석(정상 영업일이 아니면 종일 LIMIT, 정상일이면 고정 창)은 이 문서의 유도다 — 사용자 문언은 「브로커 달력 + 고정 시각」이다.
대안(달력 창과 고정 창의 교집합)은 조기 폐장일 09:00~13:00 에 MARKET 이득을 더 얻지만, 합집합 의미 때문에 KRX 조기 폐장을 NXT 창이 가리면
교집합이 13:00~15:30 을 다시 연다. 이 해석은 MARKET 칸을 대안보다 좁히기만 하므로 불변식 4 에 대해 보수 방향이다. Manager 확인 항목으로 남긴다
(`review.md` 0.10 P2-2).

**즉시성 제약 (불변식 4 — 이 절의 가장 무거운 요구).**

- **제출 경로에서 브로커를 부르지 않는다.** 판정은 메모리의 마지막 달력 응답만 읽는다. 응답이 없거나 낡았으면 그 자리에서 LIMIT 이다 —
  달력 GET 을 기다리지 않는다. 오늘의 LIMIT 보호 청산이 겪지 않는 대기를 더하면 그것이 지연이다.
- **판정 실패는 거부가 아니라 LIMIT 이다.** 판정 함수는 오류를 반환해 `sellIntent`·`submit` 의 거부 갈래(`submit` B5 `:1533` → 해제)로 보내지
  않는다. 달력 문제로 보호 청산이 **안 나가는** 갈래가 생기면 오늘보다 약해진다.
- **달력 갱신은 청산 관측 주기 밖에서 한다.** 관측 루프(`Run` B3 `o.clk.Sleep(ctx, o.Interval())` — 0.5 AST)가 달력 GET 에 묶이면 판정이 늦어진다.
  갱신 주기는 6시간 상한보다 짧아야 하고 KR 만 읽는다(§0.4 — 하루 수 회의 GET). 구조(별도 루프·기존 엔진 루프 목록 등록 등)는 2.6.4 가 정한다.
- 엔진이 꺼져 있으면 이 갱신·판정이 없다(불변식 3 — CLI·ops·MCP 경로는 달력을 읽지 않는다).

**잔존 `[미측정]`**: ① 브로커 달력이 KRX 조기 폐장·개장 지연을 위 필드로 실제로 보이는지(openapi 예시 없음). 거짓으로 정상 모양을 보이면 그날
조기 폐장 뒤 MARKET 이 나갈 수 있다 — 그 시각 MARKET 이 거부되면 오늘보다 약하다. ② 정규장 칸 안의 종가단일가(15:20~15:30)와 VI 단일가는 §5.1
(14:59 접속매매) 밖이다 — 종가단일가는 LIMIT 으로 뺐고(위 표), VI 단일가는 예측 입력이 없어 D8 잔존 행이다.

### KRX 단독 판정 입력 조사 (0.10 독립 리뷰 P1-1, 2026-10-28) — **사람 결정됨 2026-10-28 (다) → 8판에서 대체됨**

> 이 조사가 D-1 의 입력이었다. 결정 (다)로 KRX 단독 「정상 접속매매」를 증명할 필요가 사라졌다 — 8판의 창 안 보호 매도는 KRX 상태와 무관하게 같은 LIMIT 유형이다(D11 (c)).

조사 범위: `docs/migration/openapi.latest.json`(경로 목록 전수 · 스키마 전수에서 KRX·NXT·정지·VI·status 검색)과 저장소 Go 형. 브로커 호출 0.

| 필요한 사실 | openapi 에 있나 | Go 가 읽나 | 제출 경로 GET 0 아래에서 「지금」을 말하나 |
| --- | --- | --- | --- |
| KRX 휴장 | 있음 — `RegularMarketSession.singlePriceAuctionStartTime` 「KRX 기준. KRX 휴장이면 null」(`:9808`) | 아니다 — `MarketCalendarSession` 은 시작·끝만(`typed_calendar_reads.go:11-14`) | 예 — 하루 단위 값이라 백그라운드 캐시로 충분 |
| KRX 조기 폐장(종가단일가 시작이 당겨짐) | 같은 필드(KRX 기준) | 아니다(같음) | 예(같음). 실제 조기 폐장일 응답 표본은 0 — `[미측정]` |
| **KRX 개장 시각 · 개장 지연** | **없다** — 정규장 `startTime` 은 「가장 이른 KRX/NXT 정규장 시작」(`:9804`). 프리마켓은 NXT(`:9719` `IntegratedHour`) | — | **아니다** — NXT 가 제때 열면 합집합 시작이 KRX 지연을 가린다 |
| 종목별 KRX 거래정지 | 있음 — `StockInfo.koreanMarketDetail.krxTradingSuspended`(`KrMarketDetail` `:8797`, `:8815`; 경로 `/api/v1/stocks` `:849`) | 아니다 — `apiStockInfo` 에 `koreanMarketDetail` 필드 없음(`stock_reads.go:30-40`, 주석 `:29` 에만 있음) | **아니다** — 종목별 GET 이 필요하고, 장중 정지는 캐시 사이에 생긴다 |
| 종목별 VI(정적·동적) | 있음 — `/api/v1/stocks/{symbol}/warnings` `VI_STATIC`·`VI_DYNAMIC`(`StockWarning` `:8830`, 「VI 발동/해제는 수 초 내 반영」 `:1268`), `exchange` 로 KRX 구분 | 예 — `apiStockWarning`(`warnings_reads.go:20-25`), 다만 엔진은 안 부른다 | **아니다** — VI 는 수 분짜리 장중 사건이라 캐시가 「지금」을 못 말한다. 제출 직전 GET 은 즉시성 제약 위반 |
| 시장 전체 매매 정지(서킷브레이커·사이드카) | **없다**(검색 0) | — | 아니다 |

**결론**: KRX 전용 입력은 **일부만 있다.** KRX 휴장·조기 폐장은 달력의 KRX 기준 필드로 가를 수 있고(형 확장 = 위 「스냅숏 형」), 종목별 정지·VI 는 openapi 에 있으나
종목별 GET 이라 제출 경로 GET 0 아래에서는 「지금」의 증거가 되지 못한다. **KRX 단독 개장 지연과 시장 전체 정지는 입력 자체가 없다.** 그래서 리뷰가 요구한
「정상 KRX 접속매매의 증명」은 이 입력으로 닫히지 않고, Q1 원칙(불명 → LIMIT)을 그대로 적용하면 **KR MARKET 칸도 비고 a087 Phase 2 가 바꾸는 주문이 0** 이 된다.
이것은 사용자 결정(2026-10-10 시장가 방향 · 2026-10-28 Q1~Q3)의 전제를 바꾸므로 문서가 고르지 않는다. 선택지(권장 포함)는 `issues.md` 「사람 결정 필요 (2026-10-28,
0.10 독립 리뷰)」 D-1 · `review.md` 0.10 독립 리뷰 처분 절. 결정 전까지 D7 판정 규칙·2.6.x 는 착수 금지이고, 이 절의 표와 「스냅숏 형」은 결정 입력이다.

### §5.2 처분 (7판 — 사용자 결정 2026-10-28 Q4): not-applicable

§5.2(정규장 밖 KR MARKET 매도 1주 실측)는 하지 않는다. 사유: D7 이 모름을 LIMIT 으로 고정했으므로 실측의 이득은 「나중에 어떤 칸을 MARKET 으로
옮길 근거」뿐이다. 그 칸은 오늘 LIMIT 이고 바뀌지 않는다. 반대로 실측은 실주문이 다음 세션에 체결될 위험이 있다(tasks §5 표 「세션 밖인데 200」 행).
이득 0 · 위험 1 이다. 그래서 tasks 5.2 는 not-applicable 로, 5.3 은 5.1 분(M50)만으로 닫는다. LIMIT 칸을 MARKET 으로 옮기려면 새 실측 · spec 개정 ·
리뷰를 거친다(이 change 밖). `issues.md` I-R2(`order-type-not-allowed` 422 분류)는 「§5.2 결과 뒤 판정」이었으므로 **판정할 입력이 없는 채 무소유로 남는다**
— 정규장 밖 MARKET 을 내지 않으므로 그 코드를 받을 생산 경로가 없다.

## D8 — 시장가가 지정가보다 낫지 않거나 나빠지는 경계 조건 (재리뷰 P2-3)

> **8판 읽기**: 아래 표의 「MARKET」을 「하한가 지정가」로 읽는다 — 하한가 도달 · 매수 호가 공백 · 갭다운 · 거래정지 행은 같은 처분이다. VI · 종가단일가 행은 8판에서
> D11 (c) 구간 표가 대신한다(유형이 LIMIT 이라 접수 미측정 사유가 없다). 1억 이상 행: 지정가 금액이 하한가 × 수량이라 관측가 기준보다 작다 — 브로커가 그 금액으로
> `confirm-high-value-required` 를 판정하는지는 `[미측정]`(그렇다면 오늘보다 덜 거부된다). 부분체결 잔량은 a167(D10).

| 경계 조건 | 처분 | 사유 |
| --- | --- | --- |
| 하한가 도달·매수 호가 공백 | **범위 밖 — 잔존 위험** | MARKET 매도도 체결되지 않는다. 하한가 지정가 대비 이득 0. 체결 보장 주장 철회(proposal §0.3)가 이 행의 기록 |
| VI(변동성 완화장치) 단일가 구간 | **범위 밖 — `[미측정]`** | 단일가 매매 중 MARKET 접수·체결 거동 실측 없음. §5 측정 항목도 아니다. `issues.md` I-R4 |
| KRX 종가단일가(15:20~15:30 — 7판) | **다룬다 — LIMIT(오늘과 같음)** | openapi `RegularMarketSession` 이 정규장을 「종가단일가 구간을 포함」으로 정의한다. §5.1 은 14:59 접속매매 한 건이라 이 구간을 재지 않았다. D7 Q1 원칙(접수 미측정 → LIMIT)을 적용해 MARKET 칸에서 뺐다 — Manager 확인(`review.md` 0.10 P2-4) |
| 부분체결 잔량 (1차 리뷰 A14 의 행방) | **다룬다 — tasks 3.6** → 갭 I-R9 는 **새 change a167 로 이연**(사용자 결정 2026-10-22: 경보만이 아니라 자동 재가격·재주문) | MARKET 잔량이 남는지·`filldetect` 가 그 잔량을 어떻게 닫는지 확인하고, 갭이면 issues 에 기록. 3.6 결과: 잔량이 비종결로 남는 동안 시간 상한·경보 없음(`analysis/partial-fill-3.6/`). 처분은 아래 D10 |
| 1억 이상 주문 | **범위 밖 — 유형 무관 기존 위험** | `orders_write.go` `ConfirmHighValueOrder: false` → `400 confirm-high-value-required`(재리뷰 인용)는 LIMIT 에도 같다. MARKET 의 금액 평가 기준 `[미측정]` — `issues.md` I-R4 |
| 갭다운·거래정지 | **범위 밖 — 잔존 위험** | 접수 ≠ 체결. 거래정지 중 제출 응답 `[미측정]` |

## D9 — a100 D8 이중 매도 계약과의 상호작용 (재리뷰 P2-4 — a100 으로 명시 이연)

> **8판 읽기**: 창 안의 하한가 지정가는 즉시 체결 가능한(시장성 있는) 매도라 아래 「인프로세스 보호 매도가 MARKET 이 되면 즉시 체결」의 의미가 그대로 적용된다. 처분(a100 이연)은 같다.

a100 design D8 은 계약 1 을 「a087 Phase 2 가 착지하면 창은 줄어든다」로 개정해 a087 을 전제로 삼았고, 계약 2 는 「상주
조건주문 취소 확인 실패는 인프로세스 매도를 막지 않는다」이다. 인프로세스 보호 매도가 MARKET 이 되면 즉시 체결되고,
취소가 확인되지 않은 상주 SINGLE+MARKET 이 같이 발동하는 창(M13)의 의미가 바뀐다. 브로커 매도가능수량 예약(M29)이 둘째
주문을 막는지는 `[미측정]`.

**처분: a100 으로 이연한다.** a100 은 HOLD(`f3a061e5`)라 착지 순서가 열려 있고, 이중 매도 계약의 소유자는 a100 이다.
a087 은 이 상호작용을 바꾸는 코드를 갖지 않는다(조건주문 경로 무변화 — tasks 4.3). 다만 a087 이 먼저 착지하면 a100 의
계약 1 전제가 참이 되는 순간이 a100 착지보다 앞서므로, **a100 착수 시 이 절을 재대조**해야 한다. a100 문서 쪽 반영은
이 개정의 디렉터리 범위 밖이다 — `issues.md` I-R5 에 Manager 반영 요청으로 남긴다.

## D10 — 비종결 잔량의 자동 재가격·재주문은 새 change a167 이 소유한다 (사용자 결정 2026-10-22, I-R9)

> **8판 읽기**: 아래 「a167 의 재제출 유형은 D7 의 구간 표를 따라야 한다」는 8판에서 「D11 (a) 표를 따른다」로 읽는다. a167 의 「구간 입력(tasks 2.6) 뒤 착지」 선행 조건은
> 8판에서 「하한가 판(tasks 2.3·2.4) 뒤」가 된다 — a167 문서 반영은 범위 밖(`issues.md` D-2 ④).

Manager 질문 「남은 잔량 경보. 권장: 가격을 자동으로 다시 정해 재주문」에 사용자가 그 권장대로 답했다. 고른 것은 **경보만이 아니라
자동 재가격·재주문**이다. 이것은 청산 경로에 새 주문 행동(취소 → 재제출 또는 정정)을 더하는 High-risk 변경이다. a087 의 범위
(보호 청산 1차 제출의 유형)에 없다. 그래서 a087 은 이것을 구현하지 않는다. 소유자는 새 change
`a167-a-protective-remainder-is-re-priced`(`STORY-TOS-a167`, proposal 수준 정의만)다.

a087 이 남기는 것:

- 갭의 기록 — `issues.md` I-R9 와 `analysis/partial-fill-3.6/`(AST 근거, HEAD `0686c0f0a`).
- 이연 사실 — tasks 「후속 change」 표의 a167 행.
- 상호작용 하나 — a167 의 재제출 유형은 이 문서 D7 의 구간 표를 따라야 한다. LIMIT 구간이면 재제출도 LIMIT 이다. 그래서 a167 은
  a087 Phase 2 의 구간 입력(tasks 2.6) 뒤에 착지한다. a167 proposal 「선행」에 적었다.

## 건드리지 않는 것

- **(8판) 실행 관문 ②·③ · `internal/trading` · `engine.NewOrderPath` · `buildFlattenWiring` · CLI · ops · MCP.** 유형이 LIMIT 이라 열 것이 없다(D11 (e))
- **(8판) `sellIntent` 본문.** 결정 가격이 인자로 들어갈 뿐 사다리 · 거부 갈래는 바이트 그대로다(D11 (a))
- **(8판) 브로커 달력 · `official.MarketCalendarSession` · `scheduler.CalendarDay`.** 7판의 구간 입력은 대체됐다(D7 머리말)
- **(8판) 원장 schema.** `intents.price` 는 보낸 가격 그대로, 출처는 a129 줄에만(D11 (b))
- **진입 주문 유형.** riskcalc 규칙이 실제로 적용되는 곳
- **`flatten`.** 이미 거래소 하한가를 1순위로 쓴다(`liquidate.go` 「The exchange's own floor」). 무변경. 하한가를 보호 청산에
  쓰는 원안은 D6(미채택 기록)
- **`verifylive`.** "체결되면 안 되는 지정가"가 그 도구의 안전 성질이다
- **조건주문.** 2c 범위
- **호가 그리드 정본화.** **무소유**(옛 표기 「a088」은 존재한 적 없다 — 3차 재리뷰 A-P3-1, proposal Non-goals)

## 검증

### 8판 검증 (정본)

> **9판**: 아래에 D12.1(timestamp · raw decimal 사유별 거름) · D12.2(겹침 0 · 주기당 ≤ 오늘 + 1 · 429 뒤 5분 정지) · D12.3(`CONFIRMED` 아닌 모든 결과 금지 · `IN_DOUBT → FAILED_CONFIRMED → 해제 → 다음 주기 관측가 LIMIT` 종단 · 재시작 재구성) · D12.6(골든) 시험이 더해진다 — tasks 2.2~2.7 · 3.1.

- **가격 결정 표**(`exitLimitPrice` 단독, 시계 · 판은 시험 더블): 창 안 보호 2종 + 유효 판 → 하한가 · `lower_limit` / 익절 → 관측가(판 무시) / US → 관측가(판에 값이 있어도) /
  창 밖(08:59:59 · 15:30:00 · 주말 · 시간외) → 오늘 사다리 / 판 nil · 그날 항목 없음 · 전날 항목 · 하한가 0 · 하한가 > 관측가 · 관측가 파싱 불가 · 그날 거부됨 · 시장 문자열 판정 불가 →
  오늘 사다리 / 관측가 빔 → 기준선 · `baseline` / 둘 다 빔 → `""`(오늘의 B2 거부를 받음)
- **`PlaceIntent` 골든**: 폴백 행 전부의 `PlaceIntent` 가 편집 전 HEAD 의 `sellIntent` 출력과 같다 — 골든은 RED 커밋에 literal fixture 로 먼저 커밋(0.10 독립 리뷰 P2-2 승계,
  바뀐 함수끼리 비교 금지). 하한가 행은 유형 `limit` · 가격 = 판 값 · 수량 · 통화가 오늘과 같다
- **제출 경로 GET 0**(호출 계수 더블) · **결정이 `submit` B5 에 닿지 않음**(판 미준비 행) · 판 교체와 결정의 경합(`-race`)
- **거부 되먹임**: 하한가 주문이 B13 으로 거부 → 다음 주기 발의가 오늘 LIMIT 으로 나감(엔진 종단) · B7/B8/B10/B11 에서는 적지 않음
- **갱신기**: 창 밖 GET 0 · KR 만 · 그날 항목 있으면 다시 안 읽음 · 한 주기 상한 5 · 첫 오류에서 멈춤 · 입력 없으면 취소 대기 · 감독 밖(결함이 `exit` 루프를 멈추지 않음)
- **배선**: 판 생성 자리가 비시험 코드에서 `NewContext` 하나뿐(함수 단위 AST 계수 — `NewOrderPath` · `buildFlattenWiring` · `app.go` · `mcp.go` 0) · `Context.ExitObserver` 가 덮음 ·
  `engineRuntime` 의 `Auxiliary` 에 갱신기
- **기록**: a129 줄 `price` = 보낸 가격 · `price_source` 3값 · `observed_price` 는 관측가 그대로 / `intents.price` = 보낸 가격
- **관문 무변경 고정**: 하한가 LIMIT 보호 매도가 ②·③ 의 limit 갈래로 통과해 전송 계층에 닿음(종단) · 비분수 `sell+market` · `buy+market` 은 엔진 인스턴스에서도 계속 거부
- `flatten` · `verifylive` · 조건주문 무변화(4.1~4.3 기존 핀) · upstream 상속 시험 green · 전체 `go test ./... -race` 는 사용자 명시 요청 시에만
- **실계좌**: 하한가 지정가 매도 1주 접수 · 체결가 기록(tasks 5.4 — 사람 결정, 사람 터미널)

### 7판까지의 검증 (8판에서 대체됨 — 기록)

- `sellIntent` 분기: 보호 2종 → market·익절 → limit·시장가에 가격 없음·**보호가 가격
  없음으로 거부되지 않음**(MARKET 구간) · **D7 LIMIT 구간·구간 판정 불가에서 보호 2종 → 오늘과 같은 limit**(7판: US 정규장 포함 ·
  휴장·조기 폐장·개장 지연일 종일 · 달력 오류·결측·낡음 — `PlaceIntent` 전체가 오늘 출력과 같음) · **판정 실패가 거부가 아니라 LIMIT** ·
  **제출 경로의 브로커 호출 0**(D7 「구간 입력」 즉시성 제약)
- `checkOrderShape` 표: `sell+market` 통과 / `buy+market` 거부 / 가격 실린 시장가 거부 /
  `sell+market` 의 통화·수량 검사 생존 / fractional·limit 기존 분기 무변화
- `placeIntentSupported` 표(③, 단독 시험, **엔진 인스턴스 옵션 켬**): `sell+market`·가격 없음 지원 / `buy+market` 비지원 /
  가격 실린 시장가 비지원 / `sell+market` + 시장별 통화 위반 비지원 / fractional·limit 무변화 / `PreviewPlace` `LiveReady` 일치
- **upstream 동일성 (기본 인스턴스)**: 옵션 없는 `NewService` 가 비분수 market 매도를 `ErrPlaceUnsupported` 로 계속 거부,
  `PreviewPlace` 경고 문구 바이트 동일, ops `place_order`(MCP 노출 표면) 가 market 매도를 계속 거부 — 사람·에이전트 경로 무변화
- 엔진 경로 종단 시험: 보호 청산 MARKET 이 ①②③ 을 **모두** 지나 전송 계층에 닿는다(P0-1 의 회귀 시험)
- 진입 시장가 거부가 **살아 있음**을 별도 테스트로 고정(②·③ 각각)
- `isProtective` 술어 불변 + 세 소비자(D1)
- 원장: `price` 빈 값, `order_type` market, 관측가·기준선이 제출가로 새지 않음, 청산 슬리피지 계측(D6)
- 전송 표류 가드: KR market sell 골든 행(`decision_test.go` — 현재 US 만)
- `flatten`·`verifylive`·조건주문 무변화
- 전체 `go test ./... -race` 회귀 0, upstream 650 green 유지
- **실계좌 1회**(사용자 승인, `tools/a087-market-sell-probe/`): KR MARKET 매도 접수(5.1 — 2026-10-19 M50 수용) — 세션 경계(5.2)는
  not-applicable(7판 Q4, D7 「§5.2 처분」)
