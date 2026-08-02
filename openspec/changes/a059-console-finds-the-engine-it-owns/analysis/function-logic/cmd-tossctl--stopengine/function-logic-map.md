# Function Logic Map: `stopEngine`

- Source: `cmd/tossctl/engineproc.go`
- AST evidence: `ast.json` (revision: current)
- Change: a059-console-finds-the-engine-it-owns
- Risk scan: `risk-pattern-report.md`

이 map은 편집 **전에** 작성했다 (tasks.md 1.1). 이 함수는 **실행 중인 엔진에 SIGTERM을
보내는 자리**다. 이 change에서 가장 위험한 함수이며 면제하지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 엔진 프로세스 목록 | **이 콘솔이 소유한** 엔진의 PID | `engineFindProcesses(dir)` | 오류 전파, 아무것도 시그널하지 않음 |
| `os.Getpid()` | 이 콘솔 | 런타임 | 목록에 있어도 건너뛴다 |
| `engineStopTimeout` | 60s | 이 파일의 상수 | 초과는 보고, **kill 아님** |
| 엔진 활성 마커 | fresh/stale | `enginelock.Read` | 종료 **후 보고**에만 쓴다 (a056 I1) |
| `root.configDir` | journal 디렉터리로 해석 가능 | `engineJournalDir` | **(변경)** 지금은 마지막에만 쓰고 오류를 삼킨다; 앞으로 앞에서 쓰고 오류를 전파한다 |

불변식 — 이 change가 유지해야 하는 것들.

- **kill하지 않는다.** SIGTERM 후 최대 `engineStopTimeout`까지 기다리고, 안 죽으면
  보고한다. 강제 종료는 journal이 정합하게 닫히지 않는다는 뜻이다.
- **이 콘솔 자신에게 시그널하지 않는다.**
- 마커는 종료 뒤 상태를 *설명*하는 데만 쓴다. 아무것도 거부하지 않는다.

**이 change가 더하는 불변식**: 이 콘솔이 소유하지 않은 엔진에는 시그널하지 않는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `engineFindProcesses` 오류 | 없음 | 오류 | 기존 커버 |
| B2 | pid 목록 순회 | — | — | — |
| B3 | `pid == os.Getpid()` | 건너뜀 | — | `TestStoppingNeverSignalsThisProcess` |
| B4 | `engineSignalProcess` 오류 | 일부는 이미 시그널됨 | 오류 | 기존 커버 |
| B5 | `len(stopped) == 0` | 없음 | "실행 중인 엔진을 찾지 못했다." | `TestNoEngineToStopIsNotAFailure` |
| B6 | 시그널한 pid 순회 | — | — | — |
| B7 | `waitForEngineExit` 오류 | 시그널은 갔음 | 오류 (kill 아님) | `TestAnEngineThatWillNotGoIsReportedRatherThanKilled` |
| B8 | `engineJournalDir` 오류 | — | 마커 보고 생략 | 기존 커버 |
| B9 | 마커가 아직 fresh | 없음 | "종료시켰지만 활성 마커가 아직 신선하다" | 기존 커버 |
| — | 위 어느 것도 아님 | 시그널 + 대기 완료 | "…종료 시그널을 보내 루프 완주·journal 정합 close까지 기다렸다" | `TestStoppingSignalsAndWaits` |

### 이 change가 바꾸는 것

1. **B1 앞에 journal 디렉터리 해석이 온다.** 지금은 B8에서 늦게, 오류를 삼키며 구한다.
   앞으로 앞에서 구하고 오류를 전파한다 — 소유를 판정할 수 없으면 시그널 대상을 고를 수
   없기 때문이다. 그 결과 B8의 중복 호출이 사라진다.
2. **B5의 의미가 정확해진다.** 지금 컨테이너에서 이 분기는 "엔진이 없다"가 아니라
   "패턴이 못 맞았다"를 뜻한다. 실제로는 엔진이 돌고 있다. 이것이 고치는 결함이다.

분기의 **개수와 순서는 바뀌지 않는다.** B8이 사라지고 B1 앞에 하나가 생긴다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineJournalDir` | 소유 판정 기준 | **(변경)** 오류 전파. 지금은 삼킴 | AST |
| `engineFindProcesses(dir)` | 소유한 엔진 목록 | 오류 전파, 재시도 없음 | AST, seam |
| `engineSignalProcess` | SIGTERM | 오류 전파. 두 번째 시그널 없음 | AST, seam |
| `waitForEngineExit` | 루프 완주·journal close 대기 | 60s 초과는 오류 보고, **kill 없음** | AST |
| `enginelock.Read` | 종료 후 마커 상태 보고 | 거부 아님 (a056 I1) | AST |

live config binding 없음.

## State mutations and fallbacks

- 프로세스 상태 변경: 대상 pid에 **SIGTERM**. 이것이 이 함수의 실제 side effect이고,
  소유 판정이 지키는 대상이다.
- journal 자체는 건드리지 않는다 — 엔진이 스스로 정합하게 닫는다.
- fallback: 소유를 증명할 수 없는 pid는 목록에서 빠진다 → B5로 흐른다. "찾지 못했다"가
  잘못된 시그널보다 낫다 (design D3).
- 도메인 변경 없음. 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 로직은 이 함수에
  없다.

## Safety conclusion

- Safe edit boundary: **시그널 대상 선정**. 시그널 방식(SIGTERM·대기·비-kill)은 그대로다.
- 이 change는 시그널을 **보내게** 만드는 동시에 **덜 보내게** 만든다. 지금은 컨테이너에서
  0건을 보낸다(우리 엔진을 못 찾아서). 앞으로는 우리 엔진 1건을 보내고 남의 엔진 0건을
  유지한다.
- 손절 즉시성: 이 함수는 exit 루프를 **멈추는** 경로다. 그래서 위험 방향은 "너무 많이
  멈추는 것"이고, 그 방향을 막는 것이 소유 판정이다. 운영자가 자기 엔진을 세우는 것은
  명시적 요청이며 이 change가 새로 만드는 능력이 아니다 — 원래 있어야 했던 능력이다.
- 잘못된 방향으로 틀렸을 때의 결과: 소유 판정이 우리 엔진을 놓치면 B5로 흘러 "찾지
  못했다"가 된다. 이는 **현재 상태와 같으며** 더 나빠지지 않는다.
