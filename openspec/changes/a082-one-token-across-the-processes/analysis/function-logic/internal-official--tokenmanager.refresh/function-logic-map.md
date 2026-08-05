# Function Logic Map: `tokenManager.refresh`

- Source: `internal/official/token.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

**High-risk이고 이 change의 중심이다.** 지금 이 함수는 분기가 없다 — 무조건
`exchange()`한다. 그 무조건이 핑퐁의 엔진이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 호출 시점 | `send()`가 401을 받은 직후, 그때뿐 | `client.go:331` — production 호출자는 여기 하나 (HEAD 재검증) | — |
| `m.cache.AccessToken` | **방금 401을 받은 그 토큰** | `token()`의 불변식 1 | nil이면 채택 비교가 성립하지 않아 교환한다 |
| `m.cacheFile` 내용 | 공유 파일. 다른 프로세스가 이미 새 토큰을 썼을 수 있다 | console·engine·httpapi가 공유 | 읽기·파싱 실패는 교환으로 내려간다 |
| 불변식 | **401은 "내 토큰이 낡았다"이지 "새 토큰을 발급받아야 한다"가 아니다** | design D1 | 둘을 같게 취급한 것이 고치는 결함 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (현재) 분기 없음 | — | `exchange()` — 네트워크 + 파일 쓰기 | 새 토큰 또는 오류 | `TestTokenRefresh` |
| B1 (신규) | 디스크 토큰이 유효하고 `m.cache.AccessToken`과 **다르다** | `m.cache = ct` — **네트워크 없음, 파일 쓰기 없음** | 디스크 토큰, nil | `TestARefusedProcessAdoptsTheTokenAnotherProcessAlreadyGot` |
| B2 (신규) | 그 외 (없거나·무효거나·같다) | `exchange()` — 기존 동작 | 새 토큰 또는 오류 | `TestTokenRefresh` (손대지 않음) |

**B1이 없으면 두 프로세스가 서로의 토큰을 무한히 죽인다.** 있으면 진 쪽이 이긴
쪽의 토큰을 받아 수렴한다 (design D1 수렴표).

**"다르다"가 조건인 것이 좁음의 핵심이다.** "항상 채택"으로 바꾸면 진짜 만료된
토큰을 채택해 401을 무한 반복한다 — `TestTokenRefresh`가 그것을 잡는다 (변이 6.4).

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `m.mu.Lock` | 같은 프로세스의 동시 갱신 직렬화 | 프로세스 안에서만. **프로세스 간 대기는 없다** | AST calls |
| `m.exchange` | OAuth 교환 | 네트워크. `classifyStatus` 분류 | AST calls |
| `m.loadCache` | **신규** — 채택 후보 읽기 | 실패는 교환으로 내려간다 | design D1 |
| `isStillValid` | **신규** — 채택 후보 만료 판정 | 순수 함수 | design D1 |

## State mutations and fallbacks

- B1은 `m.cache`만 바꾸고 **파일을 쓰지 않는다**. 다른 프로세스의 토큰을 살려 두는
  것이 이 갈래의 목적이므로, 여기서 파일을 쓰면 목적이 무너진다.
- B2는 기존과 같이 `exchange()`가 `m.cache`와 파일을 함께 바꾼다.
- fallback 없음.

## Safety conclusion

- Safe edit boundary: **함수 본문 전체.** 다만 `m.mu` 획득은 유지하고, 시그니처와
  호출자(`client.go:331`)는 바꾸지 않는다. B1은 **네트워크와 파일 쓰기를 하지
  않아야** 한다.
- 도입 금지: 프로세스 간 블로킹 대기 (design D3, 스펙 SHALL NOT).
- 보존: 단일 프로세스 의미론. 디스크가 자기 토큰과 같으면 지금과 똑같이 교환한다
  (design D5).
- High-risk impact: **yes.** 인증 경로이고, 이 함수가 틀리면 401 무한 루프나
  거짓 자격증명 실패로 엔트리 게이트가 잠긴다.
