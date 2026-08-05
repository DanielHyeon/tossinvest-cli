# Branch Test Map: `tokenManager.exchange`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 요청 생성 실패 | 기존 커버리지 (손대지 않음) | no | yes |
| B2 | 전송 실패 → `ErrTransport` | `TestTokenExchangeError` (손대지 않음) | no | yes |
| B3 | 본문 읽기 실패 | 기존 커버리지 (손대지 않음) | no | yes |
| B4 | 비-200 → `classifyStatus`. 401/403이면 이제 상태 코드가 실린다 | `TestAnAuthRefusalCarriesItsStatusCode` (교환 엔드포인트도 같은 분류를 쓴다) | **yes** (M3) | yes |
| B5 | JSON 파싱 실패 → `ErrServer` | 기존 커버리지 (손대지 않음) | no | yes |
| B6 | 빈 `access_token` → `ErrServer` | `TestTokenEmptyAccessTokenErrors` (손대지 않음) | no | yes |
| (성공) | 파일을 쓰고 그 mtime을 기억한다 — 자기가 쓴 파일을 "바뀌었다"로 읽지 않는다 | `TestTokenExchangeAndCache` (교환 1회를 단언, 손대지 않음) | **yes** — stamp를 안 하면 매 요청 디스크를 다시 읽는다 | yes |
