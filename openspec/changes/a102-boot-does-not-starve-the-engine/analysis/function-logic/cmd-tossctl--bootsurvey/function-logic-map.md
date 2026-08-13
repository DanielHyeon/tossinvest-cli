# Function Logic Map: `bootSurvey`

> **이 change는 이 함수를 편집하지 않는다.** D7이 `runConfiguredSoakAutostart`·`bootSurvey`
> 무편집을 명시로 못박았다. 이 문서는 **proposal 근거용 분석**이고, 여기 적힌 분기 열거는
> 전부 `ast.json`이 낸 것이다 — 손으로 읽은 것이 아니다.

- Source: `cmd/tossctl/soakautostart.go` (142-151)
- AST evidence: `ast.json` — AST 기준 branches **2** / returns 3 / calls 5
- Risk scan: `risk-pattern-report.md` (매치 없음)
- source SHA-256 (현재 HEAD): `937b0c68762523ff02a39f0554371572c636308c431d9d19e42a7b7f66adab1c`
- 작성 사유: a102 proposal이 "부팅 서베이가 엔진과 rate 예산을 다툰다"를 주장하면서 **이
  함수의 분기**(이미 실행 중이면 두기 / 열거 실패는 시작하지 않음)를 근거로 썼다. 그 주장의
  AST 근거가 이 묶음이다.

## 이 함수가 하는 일

부팅 시점의 "하나 있는지 확인하고 없으면 세운다". 버튼(`restartSoak`)의 "있는 것을 갈아
치운다"와 **다른 행위**다 — a101이 그 차이를 이 함수로 분리했다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `running func() ([]int, error)` | non-nil | `runConsole`의 `soakFindProcesses(soakRecord)` closure | 오류는 **그대로 올린다** — "못 셌다"는 "없다"가 아니다 |
| `start func() (string, error)` | non-nil | `runConsole`의 `restartSoak(root, soakRecord)` closure (**PrepareSpawn 없음**) | 오류는 그대로 올린다 |

> **불변식**: 열거 실패에 시작하지 않는다. 하나의 record에 두 서베이가 append하는 것이
> soakproc이 "사람이 필요한 서베이"보다 나쁘다고 판정한 상태다.

## Branches and early returns

| Branch | 위치 | Condition | Mutation/side effect | Return/이탈 |
|---|---|---|---|---|
| B1 | `:144` | `running()` 오류 | 없음 | `:145` `return "", err` — **start를 부르지 않는다** |
| B2 | `:147` | `len(pids) > 0` | 없음 | `:148` "이미 실행 중이다 (pid …)" + nil |
| — | — | else | `start()` 호출 | `:150` `return start()` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `running()` `:143` | 이 프로필의 서베이 열거 | 오류는 삼키지 않는다 | ast.json |
| `fmt.Sprintf` `:148` | 운영자 문장 | — | ast.json |
| `joinPIDs(pids)` `:148` | pid 전부 나열 | — | ast.json |
| `start()` `:150` | 없을 때만 기동 | 결과를 그대로 돌려준다 | ast.json |

live binding — 유일한 호출자는 `runConsole`의 `bootSurveyIfAbsent` closure
(`console.go:372-377`). **a102는 그 closure를 감쌀 뿐 이 함수를 부르는 방식을 바꾸지 않는다.**

## State mutations and fallbacks

- 이 함수 자체는 프로세스 밖 상태를 바꾸지 않는다. `start`가 바꾼다.
- fallback 없음 — 두 실패 방향 다 호출자에게 그대로 돌아간다.

## Safety conclusion

- Safe edit boundary: **없음 — a102는 이 함수를 편집하지 않는다.**
- High-risk impact: yes(attestation을 갱신하는 프로세스를 세운다). a102는 그 위험을
  **더하지 않는다** — 대기는 이 함수 **앞**에 붙고, 이 함수의 계약은 그대로다.
- a102가 이 함수에 의존하는 사실 하나: **대기 중 운영자가 [soak 재시작]을 누르면**
  버튼 경로가 먼저 서베이를 띄우고, goroutine이 뒤늦게 이 함수에 도달했을 때 B2가
  pid를 보고 물러난다. 그 경합의 안전이 B2에 있다.
