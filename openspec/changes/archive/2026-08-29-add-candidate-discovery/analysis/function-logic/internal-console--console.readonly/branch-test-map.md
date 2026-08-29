# Branch Test Map: `Console.readOnly`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 유효 세션으로 `/orders`에 POST | `TestAPostToTheOrdersScreenIsRefusedByTheReadOnlyWrapper` (405 + Allow에 GET) | yes — 래퍼를 떼면 200으로 서빙된다(issues.md I-2의 개명 후 변이 확인) | yes |
