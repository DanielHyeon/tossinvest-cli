# Function Logic Map: `tokenManager.exchange`

- Source: `internal/official/token.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**High-risk.** OAuth 교환 자체다. 이 change의 편집은 **한 줄**이다 — 파일을 쓴 뒤
그 mtime을 기억한다. 교환 로직은 손대지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m.creds` | API key/secret | 자격증명 저장소 | 잘못되면 브로커가 401/403 |
| `m.base` | API base URL | 설정 | — |
| `tr.ExpiresIn` | 초 단위 수명 (운영 실측 86400) | 브로커 응답 | 0이면 즉시 만료로 취급된다 |
| 불변식 | **호출 시 `m.mu`를 이미 쥐고 있어야 한다** | 함수 주석, `token()`·`refresh()` 둘 다 잠그고 부른다 | 어기면 `m.cache` 경쟁 |
| 불변식 (신규) | 파일을 쓴 뒤 `m.stamp`가 그 파일의 mtime과 같다 | `stampCacheFile()` | 어기면 자기가 쓴 파일을 "바뀌었다"로 읽어 매 요청 디스크를 다시 읽는다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 요청 생성 실패 | 없음 | `ErrTransport` | 기존 |
| B2 | 전송 실패 | 없음 | `ErrTransport` | 기존 |
| B3 | 본문 읽기 실패 | 없음 | `ErrTransport` | 기존 |
| B4 | 상태 코드 ≠ 200 | 없음 | `classifyStatus(...)` — **이 change가 그 메시지에 코드를 싣는다** | `TestTokenExchangeError` (손대지 않음) |
| B5 | JSON 파싱 실패 | 없음 | `ErrServer` | 기존 |
| B6 | `access_token`이 빈 문자열 | 없음 | `ErrServer` | `TestTokenEmptyAccessTokenErrors` (손대지 않음) |
| (성공) | — | `m.cache` 갱신, **파일 쓰기**, `m.stamp` 갱신(신규) | 새 토큰 | `TestTokenExchangeAndCache` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `m.hc.Do` | OAuth 교환 | ctx 취소 존중. 재시도 없음 | AST calls |
| `classifyStatus` | 비-200 분류 | 이 change가 인증 거부에 상태 코드를 싣는다 | AST calls |
| `m.saveCache` | 공유 파일에 쓴다 | **best-effort** — 실패해도 호출은 성공한다 | AST calls |
| `m.stampCacheFile` | **신규** — 방금 쓴 파일의 mtime을 기억 | stat 실패는 zero stamp로, 다음 `token()`이 디스크를 다시 읽는다 (안전 방향) | design D2 |

## State mutations and fallbacks

- `m.cache = ct` — 항상.
- `saveCache` — 프로세스 밖 side effect. **실패해도 진행한다** (기존 계약).
  실패하면 `stampCacheFile`이 옛 mtime을 잡으므로 다음 `token()`이 디스크를 다시
  읽고, 거기 있는 옛 토큰이 유효하면 그것을 쓴다. 방금 산 토큰을 잃지만 유효한
  토큰을 쓰는 것이므로 안전 방향이다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: **`saveCache` 바로 뒤에 `stampCacheFile()` 한 줄.** 요청
  생성·전송·파싱·분류·`m.cache` 대입 어느 것도 바꾸지 않는다.
- `stampCacheFile`은 `m.mu`를 다시 잡지 않는다 — 이미 쥔 채로 불린다.
- High-risk impact: **yes.** 인증 경로다.
