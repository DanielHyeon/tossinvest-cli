# Branch Test Map: `TestTheCurrencyConsequenceIsRenderedAndNotFolded`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | consequence 속성 부재 | 이 함수 | RED: current 속성 0 | GREEN: current 1 |
| B2 | notice class 부재 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B3 | paired-FX 필수 표현 순회 | 이 함수 | 기존 보호 | GREEN에서 유지 |
| B4 | paired-FX 표현 부재 | 이 함수 | 기존 보호 | GREEN에서 유지 |
