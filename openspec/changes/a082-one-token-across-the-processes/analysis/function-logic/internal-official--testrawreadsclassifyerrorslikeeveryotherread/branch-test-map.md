# Branch Test Map: `TestRawReadsClassifyErrorsLikeEveryOtherRead`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 세 오류 종류(429·401·500)를 모두 돈다 | 자기 자신 | no | yes |
| B2 | 오류가 없거나 기대한 sentinel이 아니면 실패. a082가 인증 sentinel을 감싼 뒤 `==`로는 통과할 수 없다 | 자기 자신 | **yes** — `==` 상태로 a082의 wrap을 적용하면 `got official: authentication failed (HTTP 401)`로 실패한다 (실측) | yes |
