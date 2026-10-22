# a087 tasks 3.5 — 시장가 보호 청산의 체결가 확정 경로

- 측정 순간: 2026-10-22, HEAD `0686c0f0a`. AST: `ast/`(`tools/logic-map`, 생성 시 `source_sha256` 일치).
- **판정: 이미 `filldetect` 가 하는 일 — 변경 없음, 시험으로 고정.** 갭 없음.

## 경로 (AST 근거)

1. `filldetect.parseSnapshot`(`internal/filldetect/payload.go:69-131`) — 체결가는 브로커 응답 `execution.averageFilledPrice`
   하나에서만 옴. B3(`:87` `payload.Execution != nil`) 안에서 `avgPrice = trimPtr(payload.Execution.AverageFilledPrice)`(`:91`) —
   분기 없이 문자열 그대로. 주문 유형·intent 가격을 읽는 분기는 0(분기 7 개 전부 파싱·시각·취소 판정).
2. `filldetect.JournalLedger.Apply`(`internal/filldetect/ledger.go:36-110`) — 그 값을 `journal.FillObservation.AveragePrice`(`:67`)로
   넘기고 `Journal.RecordFill` 이 스냅숏(`scoped_fill_snapshots`·`fill_snapshots`)과 체결 이벤트에 씀. 분기 11 개는 계좌 범위·브로커
   시각·RECONCILE 갱신·fail-closed 번역뿐.
3. 체결 경로가 `intents.price` 를 읽지 않음은 3.7 열거가 독립적으로 보임 — 비시험 Go 에서 그 열을 읽는 SQL 은 `intentSelect`·
   `LiveOrdersForSymbol` 둘뿐(`../intents-price-null/README.md` ①).

## 고정 시험

`internal/filldetect/a087_market_exit_fill_pin_test.go` `TestA087MarketSellFillPriceIsTheBrokersAverage` — 실제 저널 + 실제 Detector,
브로커는 시험 더블. 가격 없는 KR 시장가 매도(intent price NULL)가 부분(4 @ 70100) → 전량(10, 누적 평균 70050) 체결될 때 스냅숏
평균가 = 브로커의 70050, 체결 이벤트 두 행의 평균가 = 70100·70050, intent 는 `MARKET`·가격 `""` 그대로(관측가가 체결가로 새지 않음).
변이 M10(평균가 버림) CAUGHT — `../pin-mutations/mutants-2026-10-22.log`.

## 범위 밖

체결 **품질**(체결가가 관측가에 비해 얼마나 나쁜가)은 3.8 질의(`../exit-slippage-3.8/`)이고, 판정 근거로 쓰지 않음(I-R3).
