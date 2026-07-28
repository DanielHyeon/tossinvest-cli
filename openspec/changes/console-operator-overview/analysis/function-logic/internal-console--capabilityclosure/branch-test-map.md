# Branch Test Map: `capabilityClosure`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 노드 중복 방문 차단 | 고정점 종료 — 순환 타입 변이 | — | yes |
| B2 | 이름 중복 차단 | 같은 위 | — | yes |
| B3 | 이름에서 선언으로 내려감 | 인터페이스 seam 6, func 타입 seam 8 | yes | yes |
| B4 | 빈 시그니처 | 인자·결과 없는 메서드 — 현재 표면에 없음(방어) | — | n/a |
| B5 | 시그니처 필드 순회 | func 타입 seam 8 | — | yes |
| B6 | 큐 소진 | 전 필드 | — | yes |
| B7 | 노드 종류 분기 | 전 필드 | — | yes |
| B8 | 식별자 | 이름 있는 타입 전부 | — | yes |
| B9 | 한정 타입 | `io.Writer`·`context.Context`·`domain.Position` | — | yes |
| B10 | 포인터 | `*Console` 수신자 등 | — | yes |
| B11 | 괄호 | 현재 표면에 없음(방어) | — | n/a |
| B12 | 슬라이스 | `[]string`, `[]domain.Position` | — | yes |
| B13 | 가변 인자 | 현재 표면에 없음(방어) | — | n/a |
| B14 | 맵 | 현재 표면에 없음(방어) | — | n/a |
| B15 | 채널 | 현재 표면에 없음(방어) | — | n/a |
| B16 | 제네릭 1인자 | `type Desk Seam[OrderPlacer]` 변이 | yes — 이 케이스가 없을 때 이름이 하나도 안 나왔다 | yes |
| B17 | 제네릭 다인자 | `Seam[A, B]` 변이 | yes | yes |
| B18 | 인덱스 순회 | 같은 위 | yes | yes |
| B19 | func 타입 | func 타입 seam 8 | — | yes |
| B20 | 인터페이스 | 인터페이스 seam 6 | — | yes |
| B21 | 인터페이스 메서드 순회 | 같은 위 | — | yes |
| B22 | 메서드 이름 수집 | 같은 위 + `AccountHandle`→`OrderHandle` 개명 변이 | yes | yes |
| B23 | 구조체 | `GateLimits` 값 타입 | — | yes |
| B24 | 구조체 필드 순회 | 같은 위 | — | yes |
| B25 | 능력 가능 필드만 이름 검사 | `GateLimits.MaxOrderNotional`이 통과하고 `PlaceOrder func(...)` 필드는 실패하는 두 방향 | yes | yes |
| B26 | 필드 이름 수집 | 같은 위 | yes | yes |
