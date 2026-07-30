# Branch Test Map: `TestARankOfZeroIsNotAFirstSighting`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 승격 | 자체 실행 | yes (컴파일) | yes |
| B2 | 불량 위치 5종 | 자체 실행 | yes | yes |
| B3 | 전부 거부 | 자체 실행 | yes | yes |
| B4 | instant 없는 위치 거부 | 자체 실행 | yes | yes |
| B5 | source 없는 위치 거부 | 자체 실행 | yes | yes |
| B6 | 없는 후보는 `ErrNoCandidate` | 자체 실행 | yes | yes |
| B7 | 없는 후보에서 아무것도 읽히지 않는다 | 자체 실행 | yes | yes |
| (음수 요청 수) | 이 표에 **없다** | `TestANegativeRequestedCountIsRefusedByTheFirstRankWrite`(별도 파일) | yes | yes |

음수 요청 행 수는 이 함수의 여섯 번째 거부이고 이 테스트의 표에 들어 있지 않다.
evidence 생산 시점에 그 거부를 넘기는 테스트가 **하나도 없었고**, 그래서
`negative_request_test.go`를 추가했다.
