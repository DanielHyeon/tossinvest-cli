# Branch Test Map: `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest`

시험 함수라 각 분기는 이 함수 자신이 실행함. 편집은 사례 표 두 줄 삭제뿐이고 분기 수(8)는 그대로임. 행 번호는 편집 뒤 `ast.json` 기준.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 75행 사례 표 순회(남은 7개) | `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest` | 통과(red-2.1.txt) | yes(green-2.2.txt) |
| B2 | 91행 호가 리더가 받아들이면 실패 | `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest` | 통과 | yes |
| B3 | 93행 호가 리더 거절(else) | `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest` | 통과 | yes |
| B4 | 93행 기대 사유와 다르면 실패 | `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest` | 통과 | yes |
| B5 | 96행 현재가 리더가 받아들이면 실패 | `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest` | 통과 | yes |
| B6 | 98행 현재가 리더 거절(else) | `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest` | 통과 | yes |
| B7 | 98행 기대 사유와 다르면 실패 | `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest` | 통과 | yes |
| B8 | 101행 요청이 0회가 아니면 실패 | `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest` | 통과 | yes |
