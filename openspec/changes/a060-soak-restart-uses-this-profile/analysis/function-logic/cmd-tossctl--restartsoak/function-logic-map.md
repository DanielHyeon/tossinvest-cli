# Function Logic Map: `restartSoak`

- Source: `cmd/tossctl/soakproc.go`
- AST evidence: `ast.json` (revision: current)
- Change: a060-soak-restart-uses-this-profile
- Risk scan: `risk-pattern-report.md`

편집 **전에** 작성했다 (tasks 1.1). 콘솔 버튼이 직접 부르는 함수이며 SIGINT를 보낸다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `recordPath` | 콘솔 프로필의 기록 경로 | `resolveSoakRecord(root,"")` — console.go:225 | — |
| **(추가)** `root` | 콘솔 자신의 프로필 | console.go 호출부 | — |
| 서베이 프로세스 목록 | **이 콘솔이 소유한** 서베이 | `soakFindProcesses` | 오류 전파, 시그널 없음 |
| `os.Getpid()` | 이 콘솔 | 런타임 | 목록에 있어도 건너뛴다 |
| `soakStopTimeout` | 30s | 이 파일의 상수 | 초과는 오류, **kill 아님** |
| `prepareSpawn` | token cache 준비 | `openAPISeam.PrepareSpawn` | 실패면 spawn하지 않는다 |

불변식 — 유지해야 하는 것들.

- **kill하지 않는다.** SIGINT 후 기다리고, 안 죽으면 **spawn도 하지 않는다** — 한 기록에
  두 서베이가 append하는 것이 사람 손을 부르는 것보다 나쁘다.
- **이 콘솔 자신에게 시그널하지 않는다.**
- `prepareSpawn` 실패 시 새 서베이를 띄우지 않는다.

**이 change가 더하는 불변식**: 이 콘솔이 소유하지 않은 서베이에는 시그널하지 않는다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `binstamp.SelfPath` 오류 | 없음 | 오류 | 기존 커버 |
| B2 | `soakFindProcesses` 오류 | 없음 | 오류 | `TestAFailureToLookForTheSoakIsReportedAndNothingIsStarted` |
| B3 | pid 목록 순회 | — | — | — |
| B4 | `pid == os.Getpid()` | 건너뜀 | — | `TestTheRestartNeverSignalsThisProcess` |
| B5 | `soakSignalProcess` 오류 | 일부 시그널됨 | 오류 | 기존 커버 |
| B6 | 시그널한 pid 순회 | — | — | — |
| B7 | `waitForExit` 오류 | 시그널은 갔음 | 오류, **spawn 없음** | 기존 커버 |
| B8 | `prepareSpawn` 제공됨 | token cache 준비 | — | `soakproc_openapi_test.go` |
| B9 | `prepareSpawn` 오류 | 없음 | 오류, spawn 없음 | `soakproc_openapi_test.go` |
| B10 | `soakSpawnDetached` 오류 | 없음 | 오류 | 기존 커버 |
| B11~B14 | `len(stopped)` 0 / 1 / 다수 | spawn 완료 | 각각의 안내 문구 | `TestRestartingWithNothingRunningJustStartsOne` 외 |

분기 **구조는 바뀌지 않는다.** 바뀌는 것은 두 호출의 인자다: `soakFindProcesses`가
기록 경로를 받고, `soakSpawnDetached`가 argv를 받는다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `binstamp.SelfPath` | spawn할 실행 파일 | 오류 전파 | AST |
| `soakFindProcesses(recordPath)` | 소유한 서베이 목록 | 오류 전파 | AST, seam |
| `soakSignalProcess` | SIGINT | 오류 전파 | AST, seam |
| `waitForExit` | 기록 정합 close 대기 | 30s 초과는 오류, **kill 없음** | AST |
| `prepareSpawn` | token cache | 실패면 spawn 안 함 | AST |
| `soakSpawnDetached(binary, logPath, soakArgs(root))` | 서베이 기동 | 오류 전파 | AST, seam |

## State mutations and fallbacks

- 프로세스 상태 변경: 대상 pid에 **SIGINT**, 그리고 새 서베이 spawn 1회.
- 계좌·journal·주문 어느 것도 닿지 않는다. `internal/soak`은 구조적으로 mutation이 없다.
- fallback: 소유 미증명 pid는 목록에서 빠져 B11(0건)로 흐른다.

## Safety conclusion

- Safe edit boundary: 시그널 대상 선정과 자식 argv. 시그널 방식(SIGINT·대기·비-kill)과
  "안 죽으면 spawn 안 함" 규칙은 그대로다.
- 조회 전용 경로다. 주문·손절·익절·사이징·Guardian·원장·대사·체결 어느 것도 이 함수에
  없다.
- 지금 이 버튼은 프로덕션에서 아무 효과가 없다(자식이 즉시 죽는다). 이 change는 그것을
  동작하게 만들면서, 동작하게 된 시그널이 남의 서베이로 새지 않도록 함께 막는다.
