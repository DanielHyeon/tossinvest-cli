# Branch Test Map: `TestTheRawOrderReadReportsThatThePageWasTruncated`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 읽기가 오류 없이 돌아온다 | 자체 실행 | yes (`RawOrderList`에 페이지 경계가 없으면 컴파일 실패) | yes |
| B2 | `HasNext`가 보존된다 | 자체 실행 | yes | yes |
| B3 | `NextCursor`가 보존된다 | 자체 실행 | yes | yes |
