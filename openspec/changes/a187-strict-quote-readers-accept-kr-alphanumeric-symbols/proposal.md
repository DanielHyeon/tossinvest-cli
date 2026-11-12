# Proposal

## Why

엄격 호가 조회와 엄격 현재가 조회가 영문자가 섞인 한국 종목 코드를 요청 전에 거절함. 2026-11-12 a184 신호 도구의 A 안에서 `0158S0` 이 낙폭 조건을 통과하고도 `ask_unavailable`(`parse`)로 탈락함. 그날 낙폭 통과 후보는 2건이었는데 예산 원장의 호가 조회는 1회뿐이었음(요청을 보내지 않음).

거절하는 자리는 공용 종목 문법 검사임. `strictMinuteKRSymbol` 이 `^[0-9]{6}$` 이고(`internal/official/strict_minute_candles.go:51`), `strictQuoteIdentity` 가 이 검사를 불러 두 호가 리더가 함께 걸림(`internal/official/strict_quote_reads.go:201`). 분봉 리더만 a143 에서 `^[A-Z0-9]{6}$` 로 넓혔고, 호가 리더는 숫자 전용으로 남겼음(a143 design 결정 2, 정본 스펙 `official-read-compatibility` 「Minute symbol compatibility does not widen quote or order policy」).

공식 명세(`docs/migration/openapi.latest.json`, 판본 1.2.24)는 호가 조회 `symbol` 을 「KRX: 6자리 종목코드 (숫자 또는 영문·숫자 조합; 예: 005930, 0101N0)」로 적음. 숫자 전용은 명세가 아니라 우리 쪽 제한임.

영향: a184 2판 사전등록(커밋 `7bfaf5ee2ac4e78ca71ce4d752b78f181d23ef91`)이 이 change 의 main 반영을 창 시작의 선행 조건으로 둠. 사용자 결정 2026-11-12(「호가 규칙을 넓히는 별도 체인지 먼저 해」).

## What Changes

- 엄격 호가 조회(`Client.StrictOrderbookTop`)와 엄격 현재가 조회(`Client.StrictLastPrice`)가 한국 종목 코드로 ASCII 대문자·숫자 6자리(`^[A-Z0-9]{6}$`)를 받음. 분봉 리더가 이미 받는 모양과 같음.
- 한국 종목 문법을 한 값으로 합침. `strictMinuteKRSymbol` 을 넓히고 분봉 전용 사본 `strictMinuteCandleKRSymbol` 을 없앰.
- 그 밖의 거절은 그대로임: 소문자, 앞뒤 공백, 길이 5·7, 구두점, 비 ASCII, 미국 종목 문법, 시장 값. 종목 코드를 정규화하지 않음.
- **BREAKING(스펙)**: 정본 스펙의 「두 엄격 호가 리더는 숫자 전용을 유지해야 한다」 요구사항을 개정함.

## Capabilities

### New Capabilities

없음

### Modified Capabilities

- `official-read-compatibility`: 한국 영문 섞인 6자리 코드를 분봉뿐 아니라 엄격 호가·현재가 조회도 받음. 주문 정책은 넓히지 않음.

## Non-goals

- 주문 경로. `internal/orderintent` 와 a087 주문 탐침 정책은 고치지 않음. 이 change 는 주문 경로를 새로 만들지도, 호가 리더의 결과를 주문 경로에 잇지도 않음. 주의: 주문 의도 정규화가 숫자 전용인 것은 **시장 추론**뿐임(`krSymbolPattern`, `looksLikeKRSymbol`). 시장을 `kr` 로 명시한 주문은 이 change 전부터 종목 모양으로 거절되지 않음(design Context). 그 경로를 막는 것은 이 change 의 범위가 아님.
- 전략 유니버스·후보 선정의 변경. 어떤 종목이 리더에 넘어오는지는 호출자가 정하며 이 change 는 바꾸지 않음.
- `RawMinuteCandles` 의 `krxRawCandleSymbol`(`internal/official/candle_raw.go:22`). 운영 코드 호출자가 없음(시험 제외 `grep` 0건).
- 응답 쪽 검증의 변경. 호가 본문 해독, 통화 검사, 현재가의 종목 메아리 검사는 그대로임.
- `Client.StrictMinuteCandles` 본문의 편집. 분봉 리더의 동작은 바뀌지 않음.
- a128 프로브의 입력 판본(`state_rev`) 올리기. 아래 Impact 의 프로브 항목 참조.

## Impact

- 편집: `internal/official/strict_minute_candles.go`(정규식 변수 2개 → 1개, `strictMinuteCheckCandleSymbol` 본문, 주석), 해당 시험(`strict_quote_reads_test.go` 의 거절 사례 2건이 수락으로 바뀜).
- 기존 Go 함수 본문을 고치므로 그 함수의 Function Logic Map 번들이 필요함. 엔진 증거로 쓰이는 리더의 문법 검사라 High-risk 로 다룸(a143 design 결정 2 의 분류를 따름): 변이 시험과 다른 계열 리뷰를 함.
- 넓어진 검사를 타는 호출자(시험 제외, `grep` 전수):
  - `internal/officialbars/quote.go:113·117` `PollQuoteL1`. 이 함수를 부르는 운영 코드는 `tools/a140-breakout-signal-shadow/shadow.go:289` 하나임. 엔진 패키지 호출자는 없음.
  - `tools/a128-jev-shadow-probe/collect.go:453`(프로브), `tools/a184-jev-warning-signal/main.go:287`(a184 브랜치), `tools/a140-breakout-signal-shadow/source.go:148·154`, `tools/a112-l1c-quote-probe/main.go:87·91`.
- 프로브 입력이 달라짐: 프로브는 `go run` 으로 main 체크아웃에서 돎(`run-probe.sh`). 반영 즉시 영문 섞인 종목의 입력 상태에 호가가 들어감. 지금은 그 종목에 `top_of_book` 결손이 붙고 호가 없이 판단함. 입력 판본 `a128-s3` 은 그대로라 원장만으로는 반영 전후를 가를 수 없음. 반영 시각을 기록으로 남겨 가름. `initial` 입력도 호가를 복사하므로(`profile.go` `initialState`) compare 원장의 `current`·`initial` 두 입력군이 모두 달라짐. 그 종목의 판단값, J1/J2 선별 결과, 입력 토큰 수, 수집 시간도 달라질 수 있음.
- 공식 조회가 영문 섞인 코드를 실제로 받는지는 호가·현재가 경로로 보낸 적이 없어 모름. 일반 봉 조회는 2026-11-12 에 `0158S0` 으로 200 을 받았음(a184 예산 원장). 사람 승인 뒤 읽기 전용 실측으로 확인함.
- 호출량: 지금까지 요청 전에 거절되던 종목이 실제 GET 을 보내게 됨. 재시도가 없으면 영문 섞인 종목 N개당 실제 GET 이 N회 늘어남(프로브는 회차마다, a184 신호 도구는 A 후보마다). 프로브의 조회 계수(`broker_reads`)는 리더를 부르기 전에 세므로(`collect.go:189`) 거절되던 호출도 이미 세고 있었음 — 계수는 그대로이고 실제 HTTP 시도만 늘어남. 실제 요청이 되면 그 호출도 429 재시도와 401 토큰 갱신 경로를 탐.
