# Branch Test Map: `routePathLiteral`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 보통 문자열·raw string·메서드 패턴 세 리터럴 형태 | `TestTheRoutePathIsReadOutOfEitherQuotationForm` | yes — 큰따옴표만 트림하던 형태로 되돌리면 raw string 케이스가 실패한다 | yes |
