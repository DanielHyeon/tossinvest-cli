# Function Logic Map: `EntryGate.Block`

- Source: `internal/execgw/retry.go` (498-505)
- AST evidence: `ast.json` — branches 1, returns 0, calls 2, assignments 0,
  **defers 1, go_statements 0**
- Risk scan: `risk-pattern-report.md`

**증거용.** a092는 이 함수를 편집하지 않는다. 17판의 대가 — *"진입 게이트 래치가
~3.5초에서 ≈16.5초로 늦어진다"* — 가 무엇을 늦추는지, 그리고 왜 그것이 손절을
늦추지 않는지가 이 함수의 크기에 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `reason` | `ReasonCode` 열거 | 호출자 11곳 (`rg`로 확인) | 검증 없음 |
| `detail` | 임의 문자열 | 호출자 | **첫 래치의 것만 남는다** — B1 참조 |
| `g.latches` | map | `EntryGate` | `g.mu`로 보호 |
| `g.revision` | 단조 증가 | 같은 위 | 같은 위 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return |
|---|---|---|---|
| **B1 `:501`** | **이 `reason`이 아직 래치되지 않았다** | `latches[reason] = detail` `:502`, `revision++` `:503` | 없음 (void) |
| — | 이미 래치돼 있다 | **아무것도 안 한다** | 없음 |

**분기 하나가 멱등성의 전부다.** 같은 이유로 두 번 막아도 `detail`은 처음 것이
남고 `revision`도 한 번만 는다. 그래서 17판의 배달 루프가 매 주기 실패할 때마다
`Block`을 불러도 **래치가 요동치지 않는다.**

`return` 문이 0개다(AST returns null). 이탈점이 함수 끝 하나뿐이므로
**이 함수는 실패할 수 없다** — 오류도, 대기도, 기한도 없다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.mu.Lock` `:499` | map 보호 | 다른 게이트 연산과만 경합 | AST calls |
| `g.mu.Unlock` `:500` | **defer** | — | AST defers **1** |

**네트워크 없음. DB 없음. goroutine 없음.** 체류는 map 쓰기 하나다.

## State mutations and fallbacks

| 대상 | 자리 | 성격 |
|---|---|---|
| `g.latches[reason]` | `:502` | 인메모리. **프로세스 재시작으로 사라진다** |
| `g.revision` | `:503` | 인메모리 단조 증가 — 캐시 무효화 신호 |

- fallback 없음.
- **래치는 내구적이지 않다.** 그래서 "알림 미전달"이라는 사실의 내구성은
  outbox 행이 지고 래치는 지지 않는다.

## Safety conclusion

- **Safe edit boundary**: a092는 편집하지 않는다.
- **High-risk impact**: yes — 신규 진입을 막는 유일한 자리다.
  **막는 것은 진입뿐이다**: `notifier.go:255`의 주석이 그것을 명시하고,
  청산·손절 경로는 이 게이트를 거치지 않는다. 그러므로 17판이 래치 시점을
  ≈13초 늦추는 대가는 **"신규 진입이 13초 더 열려 있다"**이지
  **"손절이 13초 늦다"가 아니다.** 이 구별이 17판이 대가를 받아들인 근거다.
- **`ReasonAlertUndelivered`를 푸는 프로덕션 경로가 없다.**
  `Clear(ReasonAlertUndelivered)`의 호출자는 `notifier.go:482`·`:511`
  (둘 다 `Notifier.Acknowledge` 안)뿐이고, `Notifier.Acknowledge`의
  프로덕션 호출자는 0이다(`internal/`·`cmd/` 전체 `rg` 확인).
  **오늘 이 래치가 걸리면 프로세스 재시작 말고는 풀 방법이 없다.**
  a092가 만드는 결함이 아니라 a092가 **발견한** 결함이며,
  17판이 래치를 더 확실히 걸게 만들므로 **더 중요해진다.**
  review.md 17.9와 §6.0 R17-11이 지고 간다.
