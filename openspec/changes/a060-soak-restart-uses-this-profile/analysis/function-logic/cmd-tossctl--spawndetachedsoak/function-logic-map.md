# Function Logic Map: `spawnDetachedSoak`

- Source: `cmd/tossctl/soakproc.go`
- AST evidence: `ast.json` (revision: current)
- Change: a060-soak-restart-uses-this-profile
- Risk scan: `risk-pattern-report.md`

편집 **전에** 작성했다 (tasks 1.1). 이 함수가 이 change의 결함 지점이다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `binary` | 존재하는 실행 파일 경로 | `binstamp.SelfPath` | spawn 오류 전파 |
| `logPath` | 콘솔 프로필 안의 `soak.log` | `soakLogPath(recordPath)` | 디렉터리/파일 오류 전파 |
| **(추가)** `args` | 콘솔의 `--config-dir`·`--session-file` + `soak run` | `soakArgs(root)` | — |
| 자식의 프로필 | **현재: 언제나 기본 프로필** | `exec.Command(binary,"soak","run")` | 자격증명을 못 찾아 즉시 종료 — 관측된 결함 |

불변식: 이 함수는 브로커·자격증명·토큰을 쥐지 않는다. `soak run`을 띄울 뿐이고 그
패키지는 구조적으로 계좌를 바꿀 수 없다(`internal/soak` import-graph 테스트).

**깨져 있는 불변식**: 로그는 콘솔 프로필에 쓰는데(`logPath`) 자식은 기본 프로필로 뜬다.
한 함수 안에서 두 프로필이 섞인다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `os.MkdirAll` 오류 | 없음 | 오류 | 기존 커버 |
| B2 | `os.OpenFile` 오류 | 없음 | 오류 | 기존 커버 |
| B3 | `cmd.Start()` 오류 | 없음 | 오류 | 기존 커버 |
| — | 성공 | **프로세스 spawn**, 로그 append | nil | `TestTheSoakSpawnCarriesThisProfile` |

분기 **구조는 바뀌지 않는다.** 바뀌는 것은 `exec.Command`에 넘기는 argv뿐이다.

| | 변경 전 | 변경 후 |
|---|---|---|
| argv | `soak run` | `--config-dir … --session-file … soak run` |
| 자식이 읽는 자격증명 | 기본 프로필 (없음 → 즉시 종료) | 콘솔 프로필 (있음) |
| 자식이 쓰는 기록 | 기본 프로필 | 콘솔 프로필 — 화면이 가리키는 그 경로 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `os.MkdirAll` / `os.OpenFile` | 로그 준비 | 오류 전파 | AST |
| `exec.Command` + `detachProcess` + `Start` | 서베이 기동 | 오류 전파, 재시도 없음 | AST |
| **(추가)** `soakArgs` | 콘솔 프로필을 자식에게 | 순수 함수 | design D1 |

live config binding 없음. 기록 경로는 넘기지 않는다 — `--config-dir`에서 자식이
`resolveSoakRecord`로 같은 값을 유도한다 (D1).

## State mutations and fallbacks

- side effect: 프로세스 spawn 1회, 로그 파일 append. 계좌·journal·주문 어느 것도 닿지
  않는다.
- fallback 없음. 실패는 전부 오류로 올라간다.

## Safety conclusion

- Safe edit boundary: 자식에게 넘기는 argv.
- 방향은 교정이다. 지금은 콘솔이 보는 프로필과 자식이 쓰는 프로필이 다르고, 그래서
  버튼이 프로덕션에서 100% 실패한다. 고친 뒤에는 같아진다.
- 조회 전용성은 이 함수 밖에서 보장된다. argv에 붙는 것은 프로필 플래그 두 개뿐이고
  하위 명령은 그대로 `soak run`이다.
