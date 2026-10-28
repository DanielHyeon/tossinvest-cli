# Branch Test Map: `adaptPrices`



| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | range 170 — 행마다 Quote 하나, 브로커 원문 보존 · 받은 시각 채움 | `TestA163AdaptPricesKeepsTheBrokerTimeVerbatim` · `TestA163AMissingOrMalformedBrokerTimeJudgesTheSame` · `TestAdaptPricesUnit` | yes(2.5 — 골격에서 SourceTimestamp 빈 값으로 실패) | pass |
