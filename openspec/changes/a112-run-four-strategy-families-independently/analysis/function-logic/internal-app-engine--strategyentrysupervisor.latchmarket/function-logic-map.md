# Function Logic Map: `StrategyEntrySupervisor.latchMarket`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `66150078e25dfad6d1fec322b955e5f23e3aad77f0525321867a500e0960f58f`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Revision: base — 이 change 는 이 함수를 **고치지 않는다.** 태스크 5.6 이
  "중앙 무결성 처리를 보존한다"를 시험으로 만들면서 이 함수의 실패 갈래를
  영수증으로 인용하기 때문에 만든 번들이다. 손으로 읽은 인용을 근거로 쓰지 않기
  위한 AST 열거다.

## CodeGraph hard evidence

`codegraph_callers`/`codegraph_callees`/`codegraph_impact` 를 그대로 옮긴다.
같은 이름의 심볼이 저장소에 둘 있고(`internal/protectionlifecycle/lifecycle.go:353`),
아래는 `internal/app/engine/strategy_entry_supervisor.go:915` 쪽이다.

| 관계 | 결과 |
|---|---|
| callers | `runMarket` (`strategy_entry_supervisor.go:769`) — **유일하다** |
| callees | `strategyRestartBackoff` (`:969`), `strategyRestartNotBefore` (`:977`), `IsZero`, `strategyMarketRuntime` |
| impact (depth 2) | 엔진 쪽은 `latchMarket:915`, `runMarket:769`, `Run:667`, `StrategyEntrySupervisor:212` 넷뿐 |

즉 이 함수의 실패는 오직 `runMarket` 을 지나 `Run` 의 반환으로만 밖에 나간다.
그것이 5.6 이 재는 경계다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `worker` | non-nil `*strategyMarketRuntime` | `s.workers[market]` — 생성자가 KR/US 둘을 요구한다 (`:577`) | nil 이면 패닉. 생성자가 그것을 막으므로 이 함수는 확인하지 않는다 |
| `failure` | non-nil error | `invokeBoundedStrategyCycle` 의 반환 또는 `ErrStrategyAuthorityExpired` | 문구가 비면 `:918` 이 자리표시자를 넣는다 — 이유 없는 잠금은 남기지 않는다 |
| `abnormal` | bool | 같은 호출의 두 번째 반환값 | `true` 면 refusal 이 `WORKER_ABNORMAL` (`:922`) |
| `s.clk` | non-nil `clock.Clock` | 생성자가 주입, 생성 시 `Now().IsZero()` 를 거부 (`:532`) | 나중에 0 을 돌려주면 B4 가 잠금을 **거부**한다 |
| `s.faults` | 버퍼 있는 채널, 용량 2 (`:584`) | 생성자 리터럴 | 가득 차면 B9 의 `default` 가 잠금을 오류로 되돌린다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 (`933:2`) | 실패 문구가 빈 문자열 | 지역 `reason` 만 | — | `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`(음의 갈래) |
| B2 (`937:2`) | `abnormal` | 지역 `refusal` 만 | — | `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive` |
| B3 (`940:2`) | 실패가 `ErrStrategyAuthorityExpired` | 지역 `refusal` 만 | — | `TestExpiredAuthorityLatchesBeforeEvaluation` |
| B4 (`944:2`) | 관측 시각이 0 | **없음 — 잠금 전에 반환한다** | `errors.New("strategy fault observation time is unavailable")` (`945:3`) | `TestTheFourEscalationsThatStopTheEngine…`/"관측 시각이 없으면…" |
| B5 (`948:2`) | `worker.latchRevision == math.MaxUint64` | 잠금 mutex 를 풀고 **아무것도 바꾸지 않는다** | `errors.New("strategy latch revision exhausted")` (`950:3`) | 같은 시험의 "latch revision 이 소진되면…"·"권한 만료의 잠금도…" |
| B6 (`954:2`) | 첫 refusal 이 비어 있음 | `worker.firstRefusal` | — | `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal` |
| B7 (`957:2`) | 첫 실패 문구가 비어 있음 | `firstFailure`·`firstAbnormal`·`latchID`·`latchRevision++` | — | 같은 시험 |
| B8 (`963:2`) | 재시작 시도 수가 상한 미만 | `restartAttempt++` | — | 같은 시험(포화 갈래 포함) |
| B9-a (`978:2`) | fault 를 스트림에 건넴 | 없음 | `fault.RestartNotBefore, nil` (`979:3`) | `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining` |
| B9-b (`980:2`) | 스트림 포화 | 없음 — **잠금은 이미 일어났다** | `errors.New("strategy fault handoff saturated …")` (`981:3`) | **없음 — 오늘은 도달 불가(아래)** |

## Calls and live bindings

표는 `ast.json` 의 `calls` 를 그 순서 그대로 생성한 것이다. 손으로 고른 목록이 아니다.

| Callee expression | Position | Why called / contract |
|---|---|---|
| `strings.TrimSpace` | 932:12 | 실패 문구 정규화 |
| `failure.Error` | 932:30 | 잠금 이유의 원문 |
| `errors.Is` | 940:5 | 권한 만료를 별도 refusal 로 분류 |
| `s.clk.Now` | 943:16 | **관측 시각.** 주입 시계이지 `time.Now` 가 아니다 |
| `observedAt.IsZero` | 944:5 | B4 — 시각 없이 잠금을 기록하지 않는다 |
| `errors.New` | 945:23 | B4 의 오류 |
| `s.mu.Lock` | 947:2 | 상태 변경 구간 시작 |
| `s.mu.Unlock` | 949:3 | B5 의 조기 반환이 잠금을 푼다 |
| `errors.New` | 950:23 | B5 의 오류 |
| `fmt.Sprintf` | 960:20 | `latchID` — market·generation·revision+1 |
| `strategyRestartBackoff` | 966:18 | 5s 계단, 30s 상한 |
| `strategyRestartNotBefore` | 967:28 | 절대 기한. 9999 년으로 포화 |
| `observedAt.UTC` | 973:119 | fault 의 관측 시각 |
| `s.mu.Unlock` | 976:2 | 상태 변경 구간 끝 |
| `errors.New` | 981:23 | B9-b 의 오류 |

Exact AST return positions: 945:3, 950:3, 979:3, 981:3

## State mutations and fallbacks

- 상태 변경은 전부 `947:2`–`976:2` 의 한 임계 구간 안에 있다: `effective=false`,
  `latched=true`, 첫 refusal·첫 실패·`latchID`·`latchRevision++`,
  `restartAttempt++`, `restartNotBefore`.
- **첫 원인을 보존한다.** B6·B7 은 이미 채워진 값을 덮지 않는다. 두 번째 실패가
  운영자에게 보이는 원인을 바꾸지 못한다.
- **B9-b 의 함정:** 스트림이 포화해도 잠금은 이미 일어났다(`936`–`937`). 그런데
  반환은 오류다. 즉 이 갈래는 "잠갔지만 잠갔다고 말하지 못했다"를 **중앙 무결성
  고장으로 승격**시킨다. 되돌리기(rollback)는 없다.

## Safety conclusion

- Safe edit boundary: 이 change 는 이 함수를 편집하지 않는다. 인용만 한다.
- High-risk impact: yes — 진입 잠금과 프로세스 정지가 둘 다 이 함수의 반환에 달렸다.
- **인용해 가는 사실 하나 (측정):** B9-b 는 오늘 **도달할 수 없다.** 용량이 2 이고,
  잠긴 시장은 `evaluationState`(`873:2`)가 거부하므로 시장당 잠금은 한 번이며,
  시장은 정확히 둘이다. 2 = 2 라서 넘칠 수 없고 프로파일도 `block 964-965 count=0`
  이다. 이 균형은 **우연**이고 어디에도 적혀 있지 않았다. 5.1.2 가 시장 둘을
  lane 여덟으로 바꾸면 세 번째 잠금이 이 갈래를 열고, 그 결과는 진입 정지가
  아니라 **엔진 정지 — 손절 loop 포함**이다. 태스크 5.6 은 그 등식을
  `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch` 로 못 박는다.
