# Function Logic Map: `pgrepSoak`

- Source: `cmd/tossctl/soakproc.go`
- AST evidence: `ast.json` (revision: current)
- Change: a060-soak-restart-uses-this-profile
- Risk scan: `risk-pattern-report.md`

편집 **전에** 작성했다 (tasks 1.1). 이 함수가 만드는 목록이 SIGINT 대상이 된다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `soakProcessPattern` | `pgrep -f`가 받는 ERE | 이 파일의 상수 | 잘못된 패턴은 **조용히** 빈 결과 |
| pgrep exit code | 0=매칭, 1=매칭 없음 | procps-ng / BusyBox | 1은 `nil, nil` |
| pgrep stdout | pid 목록 (변경 후: `pid cmdline`) | pgrep | 파싱 불가는 건너뜀 |
| **(추가)** 콘솔의 기록 경로 | `resolveSoakRecord(root,"")` | 호출부 | 해석 실패 시 호출부가 오류를 올린다 |

불변식: exit 1(매칭 없음)과 실행 오류는 다르다. 전자는 빈 목록, 후자는 error.

**이 change가 더하는 불변식**: 이 콘솔이 소유하지 않은 서베이는 목록에 없다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `Output()` 오류 | 없음 | 아래 두 갈래 | — |
| B2 | `*exec.ExitError` **AND** exit 1 | 없음 | `nil, nil` | `TestRestartingWithNothingRunningJustStartsOne` |
| — | 그 밖의 오류 | 없음 | `error` | `TestAFailureToLookForTheSoakIsReportedAndNothingIsStarted` |
| B3 | 출력 순회 | `pids` 누적 | — | `TestOnlyThisRecordsSoakIsFound` |
| B4 | 파싱 불가·비양수 pid | 건너뜀 | — | `TestOnlyThisRecordsSoakIsFound` |

### 이 change가 바꾸는 것

| | 변경 전 | 변경 후 |
|---|---|---|
| 패턴 | `tossctl soak run` | `tossctl( .*)? soak run` |
| pgrep 호출 | `-f` | `-a -f` |
| 소유 판정 | 없음 | 기록 경로 일치 시에만 채택 |
| 콘솔이 spawn한 서베이 | (지금은 뜨지도 못한다) | 찾는다 |
| 남의 프로필 서베이 | 못 찾음(패턴 불일치) | 찾지 않음(소유 판정) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exec.Command("pgrep",…)` | 프로세스 열거 | timeout 없음, 재시도 없음 | AST |
| `errors.As` / `ExitCode` | "없음"과 "실패" 구분 | — | AST |
| `strconv.Atoi` | pid 파싱 | 실패는 건너뜀 | AST |
| **(추가)** `pidsOwnedBy` | 소유 판정 | 순수 함수, 증명 불가는 배제 | design D3 |
| **(추가)** `resolveSoakRecord` | 명령줄의 `--config-dir` → 기록 경로 | 콘솔이 쓰는 **같은 함수** | design D2 |

## State mutations and fallbacks

- 도메인 변경 없음. side effect는 `pgrep` 1회 실행(읽기 전용)뿐이다.
- fallback: 파싱 불가·소유 미증명은 **버린다**. 목록을 좁히는 방향이다.

## Safety conclusion

- Safe edit boundary: 프로세스 발견. 시그널 방식은 이 함수 밖이다.
- 목록이 두 방향으로 움직인다. 넓히는 쪽(패턴)은 콘솔이 자기 서베이를 찾게 하고,
  좁히는 쪽(소유 판정)은 남의 서베이를 뺀다.
- soak은 조회 전용이므로 잘못된 SIGINT의 비용은 엔진보다 낮다. 그래도 같은 판정을
  쓰는 이유는, 두 경로가 갈라지면 그 차이 자체가 다음 버그이기 때문이다 (D3).
