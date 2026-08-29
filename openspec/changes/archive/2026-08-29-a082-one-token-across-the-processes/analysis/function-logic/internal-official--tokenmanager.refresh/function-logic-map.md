# Function Logic Map: `tokenManager.refresh`

- Source: `internal/official/token.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**High-risk이고 이 change의 중심이다.** base에서는 분기가 없다 — 무조건
`exchange()`한다. 그 무조건이 핑퐁의 엔진이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `refused` (신규 인자) | **호출자가 실제로 거부당한 토큰** | `send()`가 요청에 실었던 값 | 빈 문자열이면 파일의 어떤 토큰이든 채택 대상이 된다 |
| `m.cacheFile` 내용 | 공유 파일. 다른 보유자가 이미 새 토큰을 썼을 수 있다 | console·engine·httpapi가 공유 | 읽기·파싱 실패는 교환으로 내려간다 |
| 반환 `adopted` | 채택했으면 true, 발급했으면 false | 이 함수 | `send()`가 재시도 예산을 이것으로 정한다 |
| 불변식 1 | **거부당한 토큰은 절대 반환하지 않는다** | B1의 `ct.AccessToken != refused` | 어기면 죽은 토큰을 영원히 되돌려준다 |
| 불변식 2 | 채택은 네트워크도 파일 쓰기도 하지 않는다 | B1 본문 | 어기면 다른 보유자의 토큰을 죽인다 |

`refused`를 **인자로 받는** 이유: base처럼 `m.cache`에서 추론하면, 형제 goroutine이
그 사이 캐시를 바꾼 창에서 잘못된 답을 낸다. 엔진은 supervised loop마다 goroutine을
띄워 client 하나를 공유한다 (`runtime.go:277-283`).

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 파일 토큰이 유효하고 `refused`와 **다르다** | `m.cache = ct` — 네트워크 0, 파일 쓰기 0 | 파일 토큰, `adopted=true` | `TestARefusedProcessAdoptsTheTokenAnotherProcessAlreadyGot`, `TestARotationThatLandsMidRequestCostsNoToken`, `TestASiblingGoroutineThatAlreadyReplacedTheTokenIsNotOutbought` |
| (fallthrough) | 없거나·무효거나·`refused`와 같다 | `exchange()` — 네트워크 + 파일 쓰기 | 새 토큰, `adopted=false` | `TestTokenRefresh`(기존), `TestARefusedProcessWithNothingToAdoptStillExchanges` |

**B1이 없으면 보유자들이 서로의 토큰을 무한히 죽인다.** 있으면 진 쪽이 이긴 쪽의
토큰을 받아 수렴한다 (design D1).

`!= refused`가 조건인 것이 좁음의 핵심이다. "항상 채택"으로 넓히면 죽은 토큰을
채택해 회복하지 못한다 — 변이 M4가 기존 `TestTokenRefresh`를 포함해 5건을 깬다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `m.mu.Lock` | 같은 프로세스의 동시 갱신 직렬화 | 프로세스 안에서만. **프로세스 간 대기는 없다** (design D3) | AST calls |
| `m.loadCache` | 채택 후보 읽기 | 실패는 교환으로 내려간다 | AST calls |
| `isStillValid` | 채택 후보 만료 판정 | 순수. nil 안전 | AST calls |
| `m.exchange` | OAuth 교환 | 네트워크. `classifyStatus` 분류 | AST calls |

## State mutations and fallbacks

- B1은 `m.cache`만 바꾸고 **파일을 쓰지 않는다**. 다른 보유자의 토큰을 살려 두는
  것이 이 갈래의 목적이므로 여기서 파일을 쓰면 목적이 무너진다.
- fallthrough는 `exchange()`가 `m.cache`와 파일을 함께 바꾼다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: **함수 본문과 시그니처.** 호출자는 `client.go`의 `send()`
  하나이고 함께 고친다. `m.mu` 획득은 유지한다.
- 도입 금지: 프로세스 간 블로킹 대기, 그리고 취소할 수 없는 파일시스템 호출
  (design D2 철회 근거·D3).
- 보존: 단일 프로세스 의미론. 채택할 것이 없으면 base와 똑같이 교환한다.
- High-risk impact: **yes.** 틀리면 401 무한 루프이거나, 거짓 자격증명 실패로
  엔트리 게이트가 잠기고 그 잠금은 재시작으로 안 풀린다.
