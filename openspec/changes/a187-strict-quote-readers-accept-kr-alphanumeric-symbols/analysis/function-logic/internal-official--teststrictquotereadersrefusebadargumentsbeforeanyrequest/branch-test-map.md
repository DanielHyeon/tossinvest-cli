# Branch Test Map: `TestStrictQuoteReadersRefuseBadArgumentsBeforeAnyRequest`

시험 함수라 각 분기는 이 함수 자신이 실행함. 편집은 사례 표만 바꾸므로 분기 수(8)는 그대로임.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 남은 거절 사례 7개 순회 | 자신 | 기존 | 편집 뒤 확인 |
| B2·B3·B4 | 호가 리더 거절·사유 | 자신 | 기존 | 편집 뒤 확인 |
| B5·B6·B7 | 현재가 리더 거절·사유 | 자신 | 기존 | 편집 뒤 확인 |
| B8 | 요청 0회 | 자신 | 기존 | 편집 뒤 확인 |
