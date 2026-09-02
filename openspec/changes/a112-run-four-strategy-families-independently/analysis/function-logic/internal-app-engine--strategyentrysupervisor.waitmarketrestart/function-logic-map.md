# Function Logic Map: `StrategyEntrySupervisor.waitMarketRestart`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `66150078e25dfad6d1fec322b955e5f23e3aad77f0525321867a500e0960f58f`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Revision: base — 이 change 는 이 함수를 **고치지 않는다.** 태스크 5.6 이
  runMarket 의 두 비취소 갈래(`795-796`, `829-830`)를 채우려면 이 함수가 **왜**
  오류를 낼 수 있는지를 열거해야 해서 만든 번들이다.

## CodeGraph hard evidence

| 관계 | 결과 |
|---|---|
| callers | `runMarket` (`:769`) — 두 자리에서 부른다(만료 잠금 뒤 `:791`, 보통 잠금 뒤 `:825`) |
| callees | `s.clk.Now`, `s.clk.Sleep`, `notBefore.Sub`, `IsZero`, `errors.New` |

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | non-nil | runMarket 의 child context | 취소되면 `clk.Sleep` 이 `ctx.Err()` 를 돌려주고, 호출자가 그것을 취소로 분류한다 (`808:6`, `842:5`) |
| `notBefore` | 0 이 아닌 **절대** 시각 | `latchMarket` 의 반환값 (`979:3`) | 0 이면 B1 이 거절 — 다만 `latchMarket` 은 성공 시 0 을 돌려주지 않는다 |
| `s.clk` | non-nil, 단조 | 주입 시계 | `Now()` 가 0 이면 B2, 시계가 **뒤로 가면** B3 |
| `MaximumStrategyRestartBackoff` | 30s 상수 (`:35`) | 같은 파일의 const 블록 | 남은 시간이 이 값을 넘으면 기다리지 않고 거절 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`853:2`) | 기한이 0 | 없음 | `errors.New("strategy market restart deadline is unavailable")` (`854:3`) | **없음 — `latchMarket` 이 성공 시 0 을 주지 않는다(구조적으로 도달 불가)** |
| B2 (`857:2`) | 현재 시각이 0 | 없음 | `errors.New("strategy market restart clock is unavailable")` (`858:3`) | **없음 — B3 가 같은 확대 경로를 이미 연다(아래)** |
| B3 (`861:2`) | 남은 시간 > 30s | 없음 | `errors.New("strategy market restart delay is outside the bounded contract")` (`862:3`) | `TestTheFourEscalationsThatStopTheEngine…`/"재시작 기한이 계약 밖이면…"·"만료 뒤의 재시작 대기도…" |
| B4 (`864:2`) | 남은 시간 ≤ 0 | 없음 | `nil` (`865:3`) | `TestPairedMarketRestartHonorsPublishedAbsoluteDeadlineAfterHandoffRace` |
| 본문 (`867:2`) | 그 외 | 없음 | `s.clk.Sleep(ctx, delay)` | `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal` |

## Calls and live bindings

표는 `ast.json` 의 `calls` 를 그 순서 그대로 생성한 것이다.

| Callee expression | Position | Why called / contract |
|---|---|---|
| `notBefore.IsZero` | 853:5 | B1 |
| `errors.New` | 854:10 | B1 의 오류 |
| `s.clk.Now` | 856:9 | **지금**을 다시 읽는다 — 잠금 시각이 아니라 |
| `now.IsZero` | 857:5 | B2 |
| `errors.New` | 858:10 | B2 의 오류 |
| `notBefore.Sub` | 860:11 | 절대 기한 − 지금 = 남은 시간 |
| `errors.New` | 862:10 | B3 의 오류 |
| `s.clk.Sleep` | 867:9 | 취소를 존중하는 유일한 대기 |

Exact AST return positions: 854:3, 858:3, 862:3, 865:3, 867:2

## State mutations and fallbacks

- 상태 변경 없음. `ast.json` 의 `assignments` 둘은 지역 변수다.
- **fallback 없음.** 세 실패는 전부 오류이고, 호출자는 그 셋을 구별하지 않는다 —
  `ctx.Err()` 가 아니면 중앙 무결성으로 올린다(`808:6`/`842:5`가 취소만 걸러낸다).

## Safety conclusion

- Safe edit boundary: 편집 없음. 인용만.
- High-risk impact: yes — 이 함수의 오류는 프로세스를 세운다.
- **인용해 가는 사실:** 기한은 **절대 시각**이고 대기는 그것을 다시 읽은 지금과의
  차로 계산한다. 그래서 시계가 뒤로 가면 남은 시간이 상한을 넘고, 함수는
  "더 기다리기"가 아니라 **거절**을 고른다. 보수적인 선택이지만 그 거절이
  프로세스 정지로 이어진다는 점은 5.6 이 처음으로 값으로 확인했다.
