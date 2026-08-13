# Function Logic Map: `Stabilisation.withDefaults`

- Source: `internal/reconcile/recovery.go` (86-109)
- AST evidence: `ast.json` — AST 기준 branches **5** / returns 1 / calls 0 / defers 0 / go_statements 0
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256: `f32ab95497925c87fbd750dcd60772f75f30a39190dc8be03a1a7c8704622dc5`

**a102 §1(D3)이 편집한 함수다.** 새 노브 둘의 zero-default를 **여기에** 놓는 것이 설계의
결정이다 — 기존 셋과 같은 패턴이어야 "0이면 기본값"이라는 규율이 한 곳에 남는다.

## 세 판

| | 1판 (편집 전) | 2판 (`1c76a580`) | **3판 (§1.9, 이 문서)** |
|---|---|---|---|
| 위치 | `recovery.go:75-86` | `:84-104` | **`:86-109`** |
| 분기 | 3 | 5 | **5** |
| 이탈 | 1 | 1 | **1** (그대로) |
| source SHA-256 | `80ee029c…` | `e0d5690f…` | `f32ab954…` |

**3판이 B4의 의미를 바꿨다.** 2판의 B4는 `RateLimitBackoff <= 0`(zero-default)이었고,
3판은 `< DefaultRateLimitBackoff`(하한)이다. 분기 수는 그대로다 — 두 규칙을 두 분기로
쓰면 앞의 것이 뒤의 것에 **포섭되어** 지워도 동작이 같은 분기가 생기고, 그것은
반증할 수 없는 분기다. 하나로 합쳐 그 구멍을 없앴다(§1.9 F6).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `s.Interval` | 임의 (≤0이면 대체) | 호출자의 `Options.Stabilise` | B1 — `DefaultStabilisationInterval`(2s, `snapshot.go:362`) |
| `s.Required` | 임의 (≤0이면 대체) | 같음 | B2 — `DefaultStabilisationCount`(2, `snapshot.go:366`) |
| `s.MaxAttempts` | 임의 (≤0이면 대체) | 같음 | B3 — **리터럴 5** (상수 이름이 없다, 편집 전 그대로) |
| `s.RateLimitBackoff` | 임의 — **15s 미만이면 전부 15s로** | 같음 | B4 — `DefaultRateLimitBackoff`(15s, `ratelimit.go:51`)는 기본값이자 **하한** |
| `s.MaxRateLimitWait` | 임의 (≤0이면 대체) | 같음 | B5 — `DefaultMaxRateLimitWait`(5m, `ratelimit.go:59`) |

> **불변식**: 값 수신자다(`func (s Stabilisation)`). 호출자의 `Stabilisation`을 바꾸지 않고
> 채워진 복사본을 돌려준다. 호출 지점은 하나 — `New`(`recovery.go:171`)이고, 결과는
> `Recovery.stab`에 한 번 고정된다. **런타임 중 노브가 바뀌는 경로는 없다.**

## Branches and early returns

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:87` | `s.Interval <= 0` | `s.Interval = DefaultStabilisationInterval` | — |
| B2 | `:90` | `s.Required <= 0` | `s.Required = DefaultStabilisationCount` | — |
| B3 | `:93` | `s.MaxAttempts <= 0` | `s.MaxAttempts = 5` | — |
| B4 | `:102` | `s.RateLimitBackoff < DefaultRateLimitBackoff` | `s.RateLimitBackoff = DefaultRateLimitBackoff` | — |
| B5 | `:105` | `s.MaxRateLimitWait <= 0` | `s.MaxRateLimitWait = DefaultMaxRateLimitWait` | — |

Return: `:108` `return s` (AST 1개).

> B1~B3·B5는 `<= 0`이고 **B4만 `< 15s`다.** 이유는 요구가 다르기 때문이다: 앞의 넷은
> "안 정했으면 기본값"이지만 백오프는 reconciliation 델타가 **하한**을 못박는다
> ("서베이의 재시도 간격보다 짧아서는 안 된다 SHALL NOT"). `<= 0`만 쓰면 5초짜리 노브가
> 그대로 살아 그 SHALL NOT이 struct를 채운 사람 손에 달린다. `< 15s`는 0·음수·너무 짧은
> 값을 **한 분기로** 덮는다 — 음수 백오프가 `clock.Sleep`을 즉시 통과시켜 throttle 중인
> 브로커를 향한 spin이 되는 것도 같은 한 줄이 막는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| — | 호출이 **없다** (AST calls 0) | — | `ast.json` |

유일한 호출자: `reconcile.New`(`recovery.go:176`). 읽는 쪽은 `Recovery.stableSnapshot`
(`:373`·`:376`·`:393`·`:396`)과 `Recovery.waitOutRateLimit`(`ratelimit.go:87`·`:88`).

## State mutations and fallbacks

- 수신자 복사본의 다섯 필드만 쓴다. 전역·외부 상태 없음.
- fallback 자체가 이 함수의 목적이다. **fallback이 없는 필드는 없다** — 그래서 zero 값
  `Stabilisation{}`이 여전히 안전한 기본이고, 새 노브를 몰랐던 기존 호출자
  (`cmd/tossctl`, `internal/app/engine`)의 코드가 그대로 15s/5m을 얻는다.

## Safety conclusion

- Safe edit boundary: **분기 추가 두 개**(§1)와 **B4 술어 교체 하나**(§1.9). 기존 세 분기의 술어·값·순서는 불변이다.
- High-risk impact: **yes** (재시작 복구의 설정 경로). 실패 양식은 "기본값이 안 채워진다"
  하나이고, 그 결과는 `stableSnapshot`에서 즉시 드러난다.
- 물려받은 공백: 기존 세 분기의 **본문은 편집 전에도 지금도 커버리지 count=0**이다
  (`branch-test-map.md`). a102는 자기가 더한 두 분기만 메웠다.
- 보수 방향 확인: B4의 교체는 백오프를 **늘리는 쪽으로만** 작동한다. 15s 이상인 값은
  건드리지 않으므로 기존 테스트·호출자와 호환된다(전부 정확히 15s를 쓴다).
