# Branch Test Map: `Client.doRequest`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 전송 실패는 `ErrTransport`, 예산 기록 이전이다 | `internal/official/client_test.go`의 non-2xx/transport 케이스 + `TestRawReadsClassifyErrorsLikeEveryOtherRead`(orders_raw_test.go) | — (반환 계약 무변경) | yes |
| B2 | 본문 읽기 실패도 `ErrTransport`, status는 그대로 전달 | 동상 | — | yes |
| (무분기 꼬리) | 2xx·4xx·5xx 어느 쪽이든 예산이 기록된다 | `ratebudget_test.go` 전체 (`readRateBudget`·`rateBudgets` 단위) + `TestGetUnwrapsEnvelopeAndRetriesOn401`(401 후 재요청이 여전히 1회) | yes (`readRateBudget`/`record` 부재로 컴파일 실패) | yes |

**정직한 커버리지 기록**: httptest 서버가 429를 돌려줬을 때 `c.RateBudget(path)`가 채워지는지를
**끝에서 끝까지** 확인하는 테스트는 없다. 429에서도 기록된다는 주장은 구조적 근거로 선다 —
`record` 호출이 `classifyStatus`보다 앞이고 status에 대한 분기가 없다(AST branches = B1, B2 뿐).
`ratebudget_test.go`는 `readRateBudget`/`rateBudgets`를 직접 재고,
`orders_raw_test.go:TestRawReadsClassifyErrorsLikeEveryOtherRead`가 429→`ErrRateLimited` 사상이
그대로임을 클라이언트를 통해 잰다.
