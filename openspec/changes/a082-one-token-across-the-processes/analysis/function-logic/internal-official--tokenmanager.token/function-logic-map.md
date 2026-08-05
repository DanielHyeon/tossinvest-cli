# Function Logic Map: `tokenManager.token`

- Source: `internal/official/token.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**High-risk.** 이 함수가 주는 값이 모든 브로커 읽기·쓰기의 Authorization 헤더가
된다. 손절·비상 청산이 그 위에 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m.cache` | nil이거나 `cachedToken` | 이 프로세스가 마지막으로 교환했거나 디스크에서 읽은 값 | nil이면 디스크로 내려간다 |
| `m.cacheFile` 내용 | `{access_token, expires_at}` JSON | **여러 프로세스가 공유**한다 (console·engine·httpapi) | 읽기 실패·파싱 실패는 교환으로 내려간다 |
| `now` | `time.Now()` | 프로세스 시계 | 시계가 뒤로 가면 유효 판정이 보수적이 된다 |
| `isStillValid` skew | 60초 | token.go:84 | 만료 60초 전부터 교환한다 |
| 불변식 1 | **반환 직전 `m.cache.AccessToken`은 반환값과 같다** | 세 갈래 전부 | `refresh()`가 "방금 실패한 토큰"을 알아내는 근거 |
| 불변식 2 | 이 함수는 `m.mu`를 잡는다 | token.go:61-62 | 같은 프로세스 안의 동시 호출은 직렬화된다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 메모리 캐시가 유효 (`isStillValid(m.cache, now)`) | 없음 | 캐시 토큰, nil | `TestTokenExchangeAndCache` (교환 1회) |
| B2 | 디스크 캐시가 유효 | `m.cache = ct` | 디스크 토큰, nil | `TestTokenColdLoadFromDiskCache` |
| (fallthrough) | 둘 다 무효 | `exchange()` — 네트워크, `m.cache` 갱신, **파일 쓰기** | 새 토큰 또는 오류 | `TestTokenExchangeAndCache` |

**B1이 이 change의 편집 지점이다.** 지금 B1은 디스크를 보지 않으므로, 다른
프로세스가 파일을 바꿔 놓아도 이 프로세스는 최대 24시간 옛 토큰을 계속 제시한다.
편집 후 B1은 "메모리가 유효하고 **그 뒤로 파일이 안 바뀌었으면**"이 된다.

B2와 fallthrough는 손대지 않는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `isStillValid` | 만료 판정 (60초 skew) | 순수 함수 | AST calls |
| `m.loadCache` | 디스크 캐시 읽기 | 실패는 무시하고 교환으로 내려간다 (err != nil이면 조건 불성립) | AST calls |
| `m.exchange` | OAuth 교환 | 네트워크. `classifyStatus`로 분류된 오류 | AST calls |
| `os.Stat` | **신규** — 파일 변경 감지 | 실패는 "바뀌었다"로 읽어 디스크를 다시 읽는다 (안전 방향) | design D2 |

## State mutations and fallbacks

- `m.cache`는 B2와 `exchange()`에서만 바뀐다. B1은 읽기만 한다.
- `exchange()`가 캐시 파일을 쓴다 — **이 함수의 유일한 프로세스 밖 side effect**.
- fallback 없음. 오류는 그대로 올라간다.

## Safety conclusion

- Safe edit boundary: **B1의 조건식과 그 앞의 상태 판독뿐.** B2의 디스크 읽기,
  fallthrough의 교환, `m.mu` 획득, 반환값 셋 어느 것도 바꾸지 않는다. 특히
  **불변식 1을 깨면 안 된다** — `refresh()`의 채택 판정이 그것에 기대고, 깨지면
  이 change가 고치려는 핑퐁이 다른 형태로 돌아온다.
- 도입 금지: **프로세스 간 블로킹 대기.** 이 함수는 exit 루프의 모든 읽기가
  지나가므로 여기서 다른 프로세스를 기다리면 손절 판정 간격에 그 시간이 더해진다
  (안전 불변식 3, 스펙 SHALL NOT, design D3).
- High-risk impact: **yes.** 인증 경로다.
