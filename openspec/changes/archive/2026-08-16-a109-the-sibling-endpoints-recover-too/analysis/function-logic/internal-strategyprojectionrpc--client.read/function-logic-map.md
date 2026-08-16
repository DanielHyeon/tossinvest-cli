# Function Logic Map: `Client.Read`

- Source: `internal/strategyprojectionrpc/transport.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**이 함수는 a109가 편집하지 않는다.** 증거 묶음을 만드는 이유는 a109 §2b.3 G5가 같은
파일에 `Client.Close`를 **바로 뒤에** 더하기 때문이다 — 인접 편집이 이 함수의 개정을
건드리는지 AST로 확인한다. 분기 8개·반환 9개는 편집 전후로 같다.

읽기 계약이 Close의 안전성 논증의 전제이기도 하다: `Read`는 요청 하나를 보내고 응답을
**끝까지 읽은 뒤 닫는다**(defer Body.Close). 그래서 `CloseIdleConnections`가 닫을 수 있는
것은 이미 끝난 요청의 연결뿐이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 수신자 | non-nil · `http` non-nil · 토큰 ≥ 32자 | `Dial` | "invalid client"(B1) |
| `ctx` | 호출자 수명 | 소비자 | `NewRequestWithContext` error(B2) · `Do` error(B3) |
| 응답 크기 | ≤ `MaxProjectionBytes` | endpoint | "unreadable or oversized"(B5) |
| 응답 상태 | 200 | endpoint | "HTTP %d"(B6) |
| 응답 본문 | 알려진 필드만 있는 스냅샷 1개 | endpoint | "invalid response"(B7) · "must contain one value"(B8) |
| 스냅샷 내용 | `strategyprojection.Validate` 통과 | 도메인 | 그 error(B4 뒤 종단 검사) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `c == nil \|\| c.http == nil \|\| len(token) < 32` | 없음 | "invalid client" | a108 소유(직접 커버 없음 — 방어) |
| B2 | 요청 생성 실패 | 없음 | 그 error | 커버 없음 |
| B3 | `c.http.Do` error | 연결 1회 시도 | 그 error | a109 재부착 시나리오(죽은 endpoint) |
| B4 | `io.ReadAll` error | 본문 소비 | "unreadable or oversized" | a108 oversize 핀 |
| B5 | 본문 > 상한 | 본문 소비 | 같음 | a108 oversize 핀 |
| B6 | status != 200 | — | "HTTP %d" | a108 인증·라우트 핀 |
| B7 | 디코딩 실패(알 수 없는 필드 포함) | — | "invalid response" | `TestUnixClientRejectsUnknownSchemaFieldsAndOversizedResponse` |
| B8 | 값이 둘 이상 | — | "must contain one value" | 같음 |
| 종단 | `Validate` 실패/성공 | 없음 | error 또는 `Clone(result)` | a108 소유 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `http.NewRequestWithContext` | 요청 생성 | error 전파 | AST calls |
| `c.http.Do` | 전송 | Timeout 5s(Dial이 설정) · 재시도 없음 | AST calls |
| `response.Body.Close` | defer 해제 | — | AST defers |
| `io.LimitReader`·`io.ReadAll` | 상한 있는 소비 | 상한 초과는 오류 | AST calls |
| `strategyprojection.Validate`·`Clone` | 도메인 검증·복사 | error 전파 | AST calls |

## State mutations and fallbacks

- 프로세스 상태를 바꾸지 않는다. 연결은 transport가 관리하고, a109 이후 그 transport는
  keep-alive를 쓰지 않으므로 요청이 끝나면 연결도 끝난다.

## Safety conclusion

- Safe edit boundary: **편집하지 않는다.** a109의 변경은 같은 파일의 새 메서드 하나다.
- High-risk impact: no — 조회 전용 경로.
