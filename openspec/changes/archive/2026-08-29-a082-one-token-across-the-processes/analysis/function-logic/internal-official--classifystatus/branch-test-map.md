# Branch Test Map: `classifyStatus`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | switch 진입 — 모든 상태 코드가 정확히 한 갈래로 간다 | 아래 전부 | no | yes |
| B2 | 401·403이 인증 갈래로 간다. 어떤 코드가 인증인지는 안 바뀐다 | `TestAnAuthRefusalCarriesItsStatusCode` | **yes** (코드가 실리지 않았다) | yes |
| B3 | 본문에 `ip` 낱말이면 `ErrIPNotAllowed`, 아니면 `ErrAuth`. 어느 쪽이든 **메시지에 본문이 들어가지 않고 상태 코드는 들어간다** | `TestAnAuthRefusalCarriesItsStatusCode`, `TestAnAuthRefusalDoesNotCarryTheResponseBody` | **yes** | yes |
| B4 | 429 → `ErrRateLimited` | 기존 (손대지 않음) | no | yes |
| B5 | ≥500 → `ErrServer` | 기존 (손대지 않음) | no | yes |
| B6 | 그 외 4xx → `*APIError{Code, Body}` — 이 갈래는 본문을 계속 싣는다 | 기존 (손대지 않음) | no | yes |
