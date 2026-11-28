# Branch Test Map: `strictMinuteCheckCandleSymbol`

편집 뒤 본문은 분기가 없음(`ast.json` branches 0) — 아래 B1 은 위임 한 줄의 기본 경로임. 편집 전 B1(213행, KR 영숫자 수락)은 삭제됐고, 그 수락은 공용 검사가 같은 집합으로 지킴(`function-logic-map.md` 「편집 뒤」).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 위임 경로: KR `0101N0` 분봉 수락, KR 거절 모양·US 거절 | `TestStrictMinuteCandlesAcceptsDomesticAlphanumericAndPreservesRequestSymbol`, `TestStrictMinuteCandlesRefusesBadArgumentsBeforeAnyRequest` | 편집 전 통과(red-2.1.txt) | yes(green-2.2.txt) |
