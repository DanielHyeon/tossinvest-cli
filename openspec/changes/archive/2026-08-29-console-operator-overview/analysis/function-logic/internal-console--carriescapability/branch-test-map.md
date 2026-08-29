# Branch Test Map: `carriesCapability`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 타입 종류 분기 | 구조체를 가진 필드 전부 | — | yes |
| B2 | func·인터페이스·제네릭은 능력 가능 | `PlaceOrder func(...)` 필드 변이 | yes | yes |
| B3 | 다른 패키지 타입은 능력 가정 | `domain.Position` 필드 | — | yes |
| B4 | 선언 이름 | 패키지 선언 타입 필드 | — | yes |
| B5 | 순환 | 순환 타입 변이 | — | yes |
| B6 | 선언으로 재귀 | 패키지 선언 타입 필드 | — | yes |
| B7 | 포인터 | 포인터 필드 | — | yes |
| B8 | 괄호 | 현재 표면에 없음(방어) | — | n/a |
| B9 | 슬라이스 | `[]string` 필드 | — | yes |
| B10 | 가변 인자 | 현재 표면에 없음(방어) | — | n/a |
| B11 | 맵 | 현재 표면에 없음(방어) | — | n/a |
| B12 | 채널 | 현재 표면에 없음(방어) | — | n/a |
| B13 | 중첩 구조체 | `GateLimits` 값 타입 | — | yes |
| B14 | 중첩 필드 순회 | 같은 위 | — | yes |
| B15 | 중첩 필드 하나라도 능력 가능 | `PlaceOrder func(...)`를 중첩 구조체에 두는 변이 | yes | yes |
