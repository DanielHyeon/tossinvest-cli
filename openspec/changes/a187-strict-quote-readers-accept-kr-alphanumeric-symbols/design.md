# Design

## Context

- 한국 종목 문법 검사는 `strictMinuteCheckSymbol` 한 곳에 있음. B1(221행)이 KR 이면 `strictMinuteKRSymbol` 을 고르고 B2(224행)가 불일치면 `StrictReasonSymbol` 로 거절함.
- 호가 쪽: `strictQuoteIdentity` B2(201행)가 `strictMinuteCheckSymbol` 을 불러 오류면 `StrictQuoteReasonSymbol` 로 감싸 돌려줌. `Client.StrictOrderbookTop` B2(143행)와 `Client.StrictLastPrice` B2(174행)는 그 오류를 그대로 돌려주고 각각 146행·177행의 `c.strictQuoteFetch` 에 닿지 않음. 그래서 거절된 종목은 요청이 0회임.
- 분봉 쪽: `Client.StrictMinuteCandles` B3(140행)은 `strictMinuteCheckCandleSymbol` 을 부름. 이 래퍼의 B1(213행)이 KR 이고 `strictMinuteCandleKRSymbol`(`^[A-Z0-9]{6}$`)에 맞으면 통과시키고, 아니면 216행에서 공용 검사로 넘김.
- 위 분기 열거는 `analysis/pre-edit-ast/` 의 AST 산출물에서 옮긴 것임(문서 작성 전에 생성, base `6dbb752f6`).
- 호출자: Proposal 의 Impact 참조. `officialbars.PollQuoteL1` 은 리더를 부르기 전에 `checkSymbol`(108행)로 빈 값·공백만 거르고 종목 모양은 보지 않음(`internal/officialbars/producer.go:486`). 둘째 벽이 아님.
- 주문 쪽 시장 추론 `orderintent.InferMarketFromStockCode` 는 `A` + 숫자 6자리만 `kr` 로 보고 나머지는 `us` 로 돌려줌(B1 277행, B2 279행). 영문 섞인 한국 코드는 여기서 `us` 가 됨. 이 change 는 이 함수를 건드리지 않으며 영문 섞인 코드를 주문 경로로 보내는 호출자를 만들지 않음.
- 주문 의도 정규화 `orderintent.NormalizePlace` 에서 종목 모양을 보는 분기는 B4(101행) 하나임: 시장이 `us` 이고 `looksLikeKRSymbol` 이면 `kr` 로 바꿈. 종목 모양으로 **거절**하는 분기는 없음(분기 17개 열거, `analysis/pre-edit-ast/normalizeplace.ast.json`). 그래서 시장을 `kr` 로 명시하면 영문 섞인 코드도 정규화를 통과함. 이것은 이 change 전부터의 동작이고 CLI 주문의 실행 승인·설정·confirm 가드는 그 뒤에 따로 있음(리뷰 r1 발견 1).

## Goals / Non-Goals

**Goals:** 두 엄격 호가 리더가 분봉 리더와 같은 한국 종목 문법을 씀. 문법 값은 한 곳에만 있음.

**Non-Goals:** Proposal 의 Non-goals 와 같음.

## Decisions

### D1. 공용 정규식을 넓히고 분봉 전용 사본을 없앰

`strictMinuteKRSymbol` 을 `^[A-Z0-9]{6}$` 로 바꾸고 `strictMinuteCandleKRSymbol` 을 지움. 같은 값을 두 변수에 두면 한쪽만 고쳐지는 일이 다시 생김(a143 이 그렇게 갈라 놓은 것이 이번 결함의 자리임).

대안 「호가 전용 래퍼를 하나 더 만든다」는 값의 사본을 셋으로 늘리므로 버림.

### D2. `strictMinuteCheckCandleSymbol` 은 한 줄 위임으로 남김

래퍼의 분기를 지우고 `return strictMinuteCheckSymbol(market, symbol)` 만 남김. 래퍼를 없애려면 `Client.StrictMinuteCandles` 본문(140행)을 고쳐야 하는데, 그 함수는 엔진 봉 수집 경로이고 이 change 의 목적과 무관함. 본문을 고치는 함수는 래퍼 하나뿐임.

`strictMinuteCheckSymbol` 본문은 그대로임(변수 값만 바뀜). 주석 「The shared validator remains numeric-only for strict quote readers」는 사실이 아니게 되므로 고침.

### D3. 거절 목록은 그대로

소문자(`0158s0`), 앞뒤 공백, 길이 5·7, 구두점, 비 ASCII 는 계속 요청 전에 거절함. 대문자로 바꿔 주지 않음. 미국 문법과 시장 값 검사는 건드리지 않음. 정규식은 양끝 고정·ASCII 전용임.

fail-closed 가 거부하는 정상 입력: 공식 명세의 `pattern` 은 소문자도 허용하지만(`^[A-Za-z0-9.\-]+$`) 이 리더는 계속 거절함. 한국 종목 코드에 소문자가 쓰이는 사례는 본 자료에 없음: a182 브랜치의 `analysis/kr-stock-universe.json`(이 base 에는 없는 파일) 391개 코드가 모두 6자리이고, 숫자만이 아닌 19개는 소문자가 0개이며 영문자가 5번째 자리 18개, 6번째 자리 1개(`18064K`)임.

### D4. 주문 정책은 넓히지 않음

`internal/orderintent`(시장 추론 포함)와 a087 주문 탐침 정책은 고치지 않음. 이 change 는 주문 경로를 만들지 않고 호가 리더의 결과를 주문 경로에 잇지 않음. 스펙 개정문에는 「주문 쪽은 바꾸지 않는다」만 적음 — 「주문 검증이 숫자 전용이다」는 사실이 아니므로 적지 않음(Context). 영문 섞인 종목의 자동 주문 여부는 a184 실거래 전환 change 가 정함.

### D5. 실측은 사람 승인 뒤 읽기 전용으로 2회

수정 빌드로 `0158S0` 의 호가 1회, 현재가 1회를 둘째 키 전용 프로필(`--config-dir ~/.config/tossctl-a184`)로 읽음. 정규장 안이어야 호가 양쪽이 차 있음(장 밖이면 `EMPTY_SIDE` 로 거절될 수 있음). 결과(상태 코드, 거절 사유 또는 성공, 본문 digest)만 기록하고 가격은 기록하지 않음. 공식이 거절하면 그 사실을 기록하고 a184 design 의 「받지 않으면 A 에서 탈락」 문장대로 둠.

## Risks / Trade-offs

- 프로브 입력의 조용한 변화: 반영 뒤 영문 섞인 종목의 판단 입력에 호가가 생김. 식별자는 안 바뀜. a128·a183 의 분석은 반영 시각을 경계로 나눠 봐야 함. 완화: 반영 커밋과 시각을 a128 change 의 기록에 남김(tasks 4.2). a184 는 2판 창을 반영 뒤에 시작해 섞이지 않음.
- 주문 쪽 시장 추론이 영문 섞인 한국 코드를 `us` 로 봄(Context). 시장을 생략한 주문은 그 종목을 미국 종목으로 다루고, 시장을 `kr` 로 명시한 주문은 종목 모양 검사 없이 통과함. 둘 다 이 change 전부터의 동작이고 이 change 가 바꾸지 않음. a184 의 실거래 전환(후속 change)이 이 자리를 반드시 다뤄야 함.
- 공식이 실제로는 호가 조회에서 영문 섞인 코드를 거절할 수 있음. 그러면 요청 1회가 4xx 로 끝나고 호출자는 기존 오류 경로를 탐. 틀린 값을 쓰지 않음.
- 응답은 종목 신원을 따로 증명하지 않음(호가 본문에 종목 필드 없음). 이 한계는 a143 과 같고 이 change 가 바꾸지 않음.

## Migration Plan

운영 설정·토글 변경 없음. main 반영이 곧 프로브에 적용됨(다음 프로브 기동부터, `go run`). 엔진 이미지는 이 리더를 운영 경로에서 부르지 않으므로 재빌드가 필요하지 않음. 되돌리려면 커밋을 revert 함.

## Open Questions

- (닫힘, 2026-11-12) 공식 호가·현재가 조회가 영문 섞인 코드를 받는지: `0158S0` 은 받음. 실측은 D5 의 「정규장 안」과 달리 장 밖 1회였음(사용자 결정). 정규장 중 응답과 `K` 로 끝나는 코드는 프로브·신호 도구의 실제 가동에서 확인함.
