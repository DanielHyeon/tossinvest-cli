# Branch Test Map: `TestClassifyQueryError`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 표를 돈다. 감싼 인증·주소 거부 2행이 새로 들어갔다 | 자기 자신 | **yes** — 대상을 `==`로 바꾸면 새 행이 깨진다 | yes |
| B2 | 기대 class와 다르면 실패 | 자기 자신 (손대지 않음) | no | yes |
