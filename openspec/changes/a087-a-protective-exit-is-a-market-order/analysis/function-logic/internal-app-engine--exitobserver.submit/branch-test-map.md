# Branch Test Map: `ExitObserver.submit`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | floor 산술 실패 | a087 floor line case | 작성 전 | 기존 해제 규칙 |
| B2 | floor 0 | a087 floor zero case | 작성 전 | 기존 해제 규칙 |
| B3 | Guardian 거절 | a087 guardian case | 작성 전 | 가격 선택 불간섭 |
| B4 | intent attach 오류 | a087 attach case | 작성 전 | broker GET 0 |
| B5 | intent 조립 오류 | a087 fallback golden | 작성 전 | 기존 B5 |
| B6 | Place 결과 분기 | a087 end-to-end | 작성 전 | LIMIT 전달 |
| B7 | CONFIRMED | a087 end-to-end | 작성 전 | lower-limit price |
| B8 | IN_DOUBT | a087 first-attempt | 작성 전 | latch 유지 |
| B9 | unresolved event | a087 first-attempt | 작성 전 | latch 유지 |
| B10 | symbol busy | a087 first-attempt | 작성 전 | no latch |
| B11 | unsettled attempt | a087 first-attempt | 작성 전 | latch 유지 |
| B12 | missing error | a087 first-attempt | 작성 전 | latch 유지 |
| B13 | refusal default | a087 rejection fallback | 작성 전 | source 보존 |
| B14 | failed confirmed | a087 rejection fallback | 작성 전 | source 보존 |
| B15 | empty detail | a087 rejection fallback | 작성 전 | whitelist 유지 |
| B16 | default refusal | a087 rejection fallback | 작성 전 | 기존 release 유지 |
