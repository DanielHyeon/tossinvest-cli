# Branch Test Map: `ExitObserver.sellIntent`

| 분기 | 시나리오 | 시험 | GREEN 관찰 |
|---|---|---|---|
| B1 | 선택값 빈 문자열 | a087 empty fallback | baseline 선택 유지 |
| B2 | observed·baseline 모두 빈 문자열 | 기존 직접 시험 | 기존 오류 유지 |
| B3 | 수량 파싱 실패 | 기존 직접 시험 | 기존 오류 유지 |
| B4 | 가격 파싱 실패 | 기존 직접 시험 | 기존 오류 유지 |
| 정상 | lower_limit / observed | a087 PlaceIntent·wire golden | SELL LIMIT·가격 바이트 일치 |
