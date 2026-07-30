# Branch Test Map: `TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `Options` 필드를 하나도 못 읽음 | positive control | — | yes |
| B2 | 26개 필드 순회 | 이 테스트 자신 | — | yes |
| B3 | 필드 이름 수집 | 같은 위 | — | yes |
| B4 | 열거되지 않은 새 능력 | spec 시나리오 `열거되지 않은 새 능력` + 필드 추가 변이 | yes — 리뷰가 시연한 `type PlaceOrderFunc func(...)` 주입이 이 분기에서 처음 막힌다 | yes |
| B5 | `Options`의 구조체 embed | embed 삽입 변이 | yes | yes |
| B6 | allowlist 순회 | 이 테스트 자신 | — | yes |
| B7 | 필드가 사라진 항목 | 필드 제거 변이 | yes | yes |
| B8 | stale 보고 | 같은 위 | yes | yes |
| B9 | 패키지 전 파일 순회 | 이 테스트 자신 | — | yes |
| B10 | `verifylive.Broker` 이름 | spec 시나리오 `광폭 브로커 인터페이스 주입 차단` | yes | yes |
