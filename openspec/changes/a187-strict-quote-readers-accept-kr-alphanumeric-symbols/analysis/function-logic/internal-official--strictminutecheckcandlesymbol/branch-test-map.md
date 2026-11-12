# Branch Test Map: `strictMinuteCheckCandleSymbol`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 (213행, 편집 전) | KR `0101N0` 분봉 수락 | `TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol` | 기존(편집 전 통과) | 편집 뒤 확인 |
| fallthrough (216행) | KR 거절 모양·US | `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` | 기존 | 편집 뒤 확인 |
