## 1. 고정과 증거

- [x] 1.1 `capture_change_base.py` 로 base commit 을 고정한다.
- [x] 1.2 편집할 함수 `strictMinuteCheckCandleSymbol` 의 편집 전 AST 로 Function Logic Map·Branch Test Map 번들을 만든다. `analysis/pre-edit-ast/` 의 열거가 design Context 와 다르면 멈추고 보고한다.
- [x] 1.3 CodeGraph 로 `strictMinuteCheckSymbol`·`strictQuoteIdentity`·두 호가 리더의 호출자를 다시 열거해 proposal Impact 의 `grep` 전수와 대조한다(`codegraph affected` 는 `--filter '*_test.go'`, 수치에 버전을 적는다).

## 2. 구현 (RED → GREEN)

- [x] 2.1 실패 시험을 먼저 쓴다: 두 호가 리더가 KR `0101N0`·`00593A` 로 요청 1회와 정확한 query 를 보내고 그 종목을 돌려줌, 소문자·공백·길이 5·7·구두점은 요청 0회로 거절, KR `005930`·US `AAPL` 의 query 와 결과 불변, 분봉 리더의 기존 수락·거절 사례 불변, 주문 쪽 `InferMarketFromStockCode`·`looksLikeKRSymbol` 의 `0101N0` 결과 불변.
- [x] 2.2 정규식 변수를 하나로 합치고 `strictMinuteCheckCandleSymbol` 을 한 줄 위임으로 줄인다. 주석을 사실에 맞게 고친다. 다른 함수 본문은 고치지 않는다.
- [x] 2.3 편집 뒤 Function Logic Map 을 재생성하고 `check_analysis.py --change a187-strict-quote-readers-accept-kr-alphanumeric-symbols` 를 통과시킨다.

## 3. 검증

- [x] 3.1 사본 변이 시험(무변이 대조군 GREEN 먼저): 정규식을 숫자 전용으로 되돌림, 길이 고정 제거, 양끝 고정 제거, 소문자 허용, 래퍼가 검사를 건너뜀. 각 변이가 어느 시험에 잡히는지 원문을 `analysis/` 에 남긴다.
- [x] 3.2 영향 패키지(`internal/official`, `internal/officialbars`, `internal/orderintent`, `tools/a128-jev-shadow-probe`, `tools/a140-breakout-signal-shadow`, `tools/a112-l1c-quote-probe`) 시험과 `go vet`·gofmt 를 돌리고 원문을 기록한다. 전체 gate 는 사용자가 요청할 때만 실행한다.
- [x] 3.3 다른 계열 모델의 리뷰와 gstack 리뷰를 받고 처리 결과를 `review.md` 에 기록한다.
  - 다른 계열 리뷰 r1(gpt-6-astra)·gstack 리뷰 처리 완료(review.md 「구현 리뷰 r1 처리」·「gstack 리뷰 처리」).
- [x] 3.4 Manager 가 diff, 시험 원문, 변이 결과를 독립 재검증한다.

## 4. 실측과 인계

- [x] 4.1 (2026-11-12 23:44 KST 실행, 정규장 밖 — 사용자 결정, 기록 `analysis/measurement/l1c-0158S0-2026-11-12.md`) 사람 승인 뒤 정규장 안에서 수정 빌드로 `0158S0` 의 호가 1회·현재가 1회를 둘째 키 전용 프로필로 읽는다(GET 2회). 상태 코드와 성공·거절 사유, 본문 digest 만 기록한다.
- [x] 4.2 main 반영 커밋·시각과 그 뒤 첫 프로브 기동 시각을 a128 change 기록에 남긴다(프로브 입력은 그 기동부터 달라짐). a184 브랜치에 가져올 수 있게 알린다(a184 tasks 6.8).
- [ ] 4.3 PM check, strict validate, archive 와 Story 경로 동기화. 검증된 교훈만 retain 한다.
