# Function Logic Map: `Stabilisation.withDefaults`

- Source: `internal/reconcile/recovery.go` (84-104)
- AST evidence: `ast.json` — AST 기준 branches **5** / returns 1 / calls 0 / defers 0 / go_statements 0
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `e0d5690ff164c31e3f5cbb0086cdab84b206a0854a84110b120e87fb98aedfb1`

**a102 §1(D3)이 편집한 함수다.** 새 노브 둘의 zero-default를 **여기에** 놓는 것이 설계의
결정이다 — 기존 셋과 같은 패턴이어야 "0이면 기본값"이라는 규율이 한 곳에 남는다.

## 두 판

| | 1판 (편집 전) | **2판 (편집 후, 이 문서)** |
|---|---|---|
| 위치 | `recovery.go:75-86` | **`:84-104`** |
| 분기 | 3 | **5** |
| 이탈 | 1 | **1** (그대로) |
| source SHA-256 | `80ee029c…` | `e0d5690f…` |

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.Interval` | 임의 (≤0이면 대체) | 호출자의 `Options.Stabilise` | B1 — `DefaultStabilisationInterval`(2s, `snapshot.go:362`) |
| `s.Required` | 임의 (≤0이면 대체) | 같음 | B2 — `DefaultStabilisationCount`(2, `snapshot.go:366`) |
| `s.MaxAttempts` | 임의 (≤0이면 대체) | 같음 | B3 — **리터럴 5** (상수 이름이 없다, 편집 전 그대로) |
| `s.RateLimitBackoff` | 임의 (≤0이면 대체) | 같음 | B4 — `DefaultRateLimitBackoff`(15s, `ratelimit.go:51`) |
| `s.MaxRateLimitWait` | 임의 (≤0이면 대체) | 같음 | B5 — `DefaultMaxRateLimitWait`(5m, `ratelimit.go:59`) |

> **불변식**: 값 수신자다(`func (s Stabilisation)`). 호출자의 `Stabilisation`을 바꾸지 않고
> 채워진 복사본을 돌려준다. 호출 지점은 하나 — `New`(`recovery.go:171`)이고, 결과는
> `Recovery.stab`에 한 번 고정된다. **런타임 중 노브가 바뀌는 경로는 없다.**

## Branches and early returns

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:85` | `s.Interval <= 0` | `s.Interval = DefaultStabilisationInterval` | — |
| B2 | `:88` | `s.Required <= 0` | `s.Required = DefaultStabilisationCount` | — |
| B3 | `:91` | `s.MaxAttempts <= 0` | `s.MaxAttempts = 5` | — |
| B4 | `:97` | `s.RateLimitBackoff <= 0` | `s.RateLimitBackoff = DefaultRateLimitBackoff` | — |
| B5 | `:100` | `s.MaxRateLimitWait <= 0` | `s.MaxRateLimitWait = DefaultMaxRateLimitWait` | — |

Return: `:103` `return s` (AST 1개).

> 다섯 분기가 전부 `<= 0`을 쓴다. **음수도 기본값으로 접힌다.** 다른 술어(`== 0`)를 쓰면
> 음수 백오프가 살아남고 `clock.Sleep`은 비양수 duration을 즉시 반환하므로, 429 재시도가
> 이미 throttle 중인 브로커를 향한 spin이 된다. 그래서 새 두 분기도 같은 술어다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | 호출이 **없다** (AST calls 0) | — | `ast.json` |

유일한 호출자: `reconcile.New`(`recovery.go:171`). 읽는 쪽은 `Recovery.stableSnapshot`
(`:368`·`:371`·`:388`·`:391`)과 `Recovery.waitOutRateLimit`(`ratelimit.go:87`·`:88`).

## State mutations and fallbacks

- 수신자 복사본의 다섯 필드만 쓴다. 전역·외부 상태 없음.
- fallback 자체가 이 함수의 목적이다. **fallback이 없는 필드는 없다** — 그래서 zero 값
  `Stabilisation{}`이 여전히 안전한 기본이고, 새 노브를 몰랐던 기존 호출자
  (`cmd/tossctl`, `internal/app/engine`)의 코드가 그대로 15s/5m을 얻는다.

## Safety conclusion

- Safe edit boundary: **분기 추가 두 개.** 기존 세 분기의 술어·값·순서는 불변이다.
- High-risk impact: **yes** (재시작 복구의 설정 경로). 실패 양식은 "기본값이 안 채워진다"
  하나이고, 그 결과는 `stableSnapshot`에서 즉시 드러난다.
- 물려받은 공백: 기존 세 분기의 **본문은 편집 전에도 지금도 커버리지 count=0**이다
  (`branch-test-map.md`). a102는 자기가 더한 두 분기만 메웠다.
