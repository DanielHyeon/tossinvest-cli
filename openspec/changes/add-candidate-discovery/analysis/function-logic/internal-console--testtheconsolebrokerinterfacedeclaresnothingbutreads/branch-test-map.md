# Branch Test Map: `TestTheConsoleBrokerInterfaceDeclaresNothingButReads`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `holdings.go` 부재 | 파일 삭제 변이 | — | n/a(HEAD에서 삭제된 함수) |
| B2 | 다른 TypeSpec 건너뛰기 | 구조 분기 | — | n/a |
| B3 | 인터페이스가 아닌 `HoldingsReader` | 타입 교체 변이 | — | n/a |
| B4 | 메서드 목록 순회 | 선언 1개 | — | n/a |
| B5 | 메서드 이름 순회 | 같은 위 | — | n/a |
| B6 | 인터페이스 embed | embed 삽입 변이 | — | n/a |
| B7 | 메서드를 하나도 못 읽음 | positive control | — | n/a |
| B8 | 선언 메서드 순회 | 선언 1개 | — | n/a |
| B9 | 허용 밖 메서드 | 두 번째 메서드 추가 변이 | — | n/a |
| B10 | 금지 동사 순회 | 같은 위 | — | n/a |
| B11 | 금지 동사 철자 | `PlaceOrder` 추가 변이 | — | n/a |
| B12 | 패키지 전 파일 순회 | 이 절반만 전체를 봤다 | — | n/a |
| B13 | `verifylive.Broker` 이름 사용 | 대체 가드가 같은 검사를 그대로 승계했다(`TestEveryCapability…` B9·B10) | yes | yes(대체 가드에서) |
