# Function Logic Map: `Schedule.UntilNextDue`

- Source: `internal/candidate/source.go`
- AST evidence: `ast.json` (revision: current, source_sha256 bound)
- Risk scan: `risk-pattern-report.md`

이 change가 추가한 신규 함수다. 루프의 tick과 원천별 일정이 '우연히 맞아야 하는 두 숫자'인 것을 끝내려고 존재한다 — `engineYieldFactor`가 한쪽만 두 배로 만들자 루프가 격 tick마다 깨어나 아무것도 못 읽던 §5 리뷰 P0의 수리다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `sources` | 이번 시장의 패널 | 호출자 | 빈 패널이면 `(0, false)` — '물어볼 것이 없다'는 '지금 due'와 다르다 |
| `at` | 기준 순간 | 호출자 인자(주입 clock) | 시각은 주입된 clock 또는 호출자 인자에서만 온다. `TestNothingInThisPackageAsksTheWallClockWhatTimeItIs`가 `time` import를 **경로로** 해석해 `Now/Since/Until/After/Tick/NewTimer/NewTicker/AfterFunc/Sleep` 9개 이름과 alias import·dot import 두 형태, 합쳐 11가지 철자를 막는다. |
| `s.lastRun` | 원천별 마지막 실행 | `Ran` | 한 번도 안 돈 원천은 즉시 `(0, true)` — `Due`와 같은 규칙 |
| `s.every(id)` | 유효 간격 = 설정값을 floor로 올리고 엔진 양보 중이면 2배 | `intervals` + `yielding` | 미설정 원천은 `unconfiguredFloor`(15s) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 패널 순회 | 지역 `soonest`/`known` | — | `TestARepeatedScanCannotBeAskedToPollFasterThanItsFloor` |
| B2 | `!ran` — 한 번도 안 읽힌 원천 | 없음 | `(0, true)` 즉시 | `TestTheWatchLoopWaitsOnTheInjectedClock` |
| B3 | `wait < 0` — 이미 지남 | `wait = 0` | — | `TestATickBelowTheSourceIntervalDoesNotEndTheDiscoveryLoop` |
| B4 | `!known || wait < soonest` | `soonest, known` 갱신 | — | `TestTheEngineYieldDoesNotEndTheDiscoveryLoop` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `s.mu.Lock` / `s.mu.Unlock` | `lastRun`·`intervals`·`yielding` 보호 | — | §2-4가 재현한 `concurrent map read and map write` |
| `s.every` | 락을 이미 쥔 상태에서의 유효 간격 | 순수 | 같은 파일 |
| `src.ID` | 원천 식별 | 순수 | `Source` 인터페이스 |
| `last.Add`, `Sub` | 남은 시간 산술 | — | 벽시계를 부르지 않는다 |

## State mutations and fallbacks

- 상태 변경 없음 — 읽기만 한다. 락은 잡지만 아무것도 쓰지 않는다.
- 음수 wait를 0으로 접는다. 얼마나 지났는지는 호출자가 할 수 있는 일이 없고, 음수 duration은 끝나지 않는 sleep 한 걸음 앞이다.
- `known=false`(빈 패널)는 호출자가 자기 tick으로 되돌아가라는 신호다. `(0, true)`와 섞으면 빈 패널에서 바쁜 루프가 된다.

## Safety conclusion

- Safe edit boundary: 두 번째 반환을 없애 `0`으로 합치는 것, 음수 wait를 그대로 내보내는 것은 금지
- High-risk impact: no — 순수 읽기. 주문 경로 무접촉. 잘못되면 발굴이 느려지거나 바빠질 뿐 돈이 걸리지 않는다.
