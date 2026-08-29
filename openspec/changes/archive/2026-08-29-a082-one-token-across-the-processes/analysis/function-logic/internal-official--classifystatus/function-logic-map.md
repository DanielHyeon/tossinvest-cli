# Function Logic Map: `classifyStatus`

- Source: `internal/official/errors.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**High-risk.** 이 함수의 반환값이 `execgw`의 재시도·게이트 차단 판정을 결정한다.
`ClassAuthFatal`은 엔트리 게이트를 잠근다 (`retry.go:334`).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `code` | HTTP 상태 코드 | 브로커 응답 | 분류되지 않는 4xx는 `*APIError` passthrough |
| `body` | 응답 본문 | 브로커 응답. **계좌 식별자·자격증명 파생값이 들어올 수 있다** | 빈 본문도 유효 입력 |
| `reIPWord` | `\bip\b` 정규식 | errors.go:33, 한 번만 컴파일 | — |
| 불변식 1 | `errors.Is(err, ErrAuth)`가 판정 기준이다 | `execgw/retry.go:60`, `classify.go:111`, `failclosed.go:210`, `cmd/tossctl/soak.go:613`, `openapi.go:56` — 전부 `errors.Is`/`errors.As`. **문자열 비교로 판정하는 곳은 없다** | 문자열 판정이 생기면 이 함수의 메시지 변경이 분류를 바꾼다 |
| 불변식 2 | 본문은 로그로 나가지 않는다 | `ErrAuth`/`ErrIPNotAllowed`는 본문을 안 싣는다 | 싣게 되면 계좌 정보가 로그로 샌다 |

**호출자는 둘이다** (CodeGraph는 하나만 보고했고 HEAD 재검증에서 하나를 더 찾았다):
`token.go:123` (OAuth 교환 자체)와 `client.go:345` (일반 요청).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch` 진입 | 없음 (순수 함수) | — | 아래 전부 |
| B2 | `code == 401 \|\| code == 403` | 없음 | B3로 | `TestClassifyStatus*` |
| B3 | 본문에 `ip`가 낱말로 있다 | 없음 | `ErrIPNotAllowed` / 아니면 `ErrAuth` | `TestClassifyStatus*` |
| B4 | `code == 429` | 없음 | `ErrRateLimited` | 기존 |
| B5 | `code >= 500` | 없음 | `ErrServer` | 기존 |
| B6 | 그 외 | 없음 | `*APIError{Code, Body}` | 기존 |

**이 change가 바꾸는 것은 B3가 돌려주는 값의 *메시지*뿐이다.** 어떤 코드가 어느
갈래로 가는지, `ip` 낱말 검사, 반환하는 sentinel의 정체 — 전부 그대로다.
`fmt.Errorf("%w (HTTP %d)", ErrAuth, code)`이므로 `errors.Is`는 계속 성립한다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `reIPWord.Match` | IP 거부와 자격증명 거부 구분 | 순수 | AST calls |
| `bytes.ToLower` | 대소문자 무시 | 순수 | AST calls |
| `fmt.Errorf` | **신규** — sentinel을 감싸 코드를 싣는다 | 순수. `%w`라 `errors.Is` 보존 | design D4 |

## State mutations and fallbacks

- 상태를 바꾸지 않는다. 순수 함수다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: **B2·B3가 돌려주는 오류를 감싸는 것뿐.** 조건식(401/403,
  429, 500, `ip` 낱말)과 `*APIError` 갈래는 글자 그대로 보존한다.
- **본문을 메시지에 싣지 않는다** (불변식 2). 코드만 싣는다.
- 불변식 1이 깨지면 이 편집이 분류를 바꾼다. 문자열로 판정하는 곳이 없다는 것을
  테스트로 고정한다 (task 4.3).
- High-risk impact: **yes.** 이 반환값이 엔트리 게이트 차단으로 이어진다.
