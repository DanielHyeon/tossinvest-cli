# Function Logic Map: `pgrepEngine`

- Source: `cmd/tossctl/engineproc.go`
- AST evidence: `ast.json` (revision: current)
- Change: a059-console-finds-the-engine-it-owns
- Risk scan: `risk-pattern-report.md`

이 map은 편집 **전에** 작성했다 (tasks.md 1.1). 이 함수는 엔진 종료 시그널의 대상
목록을 만드는 자리이므로 면제하지 않는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `engineProcessPattern` | `pgrep -f`가 받는 ERE | 이 파일의 상수 + `tools/engine-autostart.sh` | 잘못된 패턴은 **조용히** 빈 결과가 된다 — 지금 고치는 결함 |
| `pgrep` 실행 가능 여부 | PATH에 존재 | OS | exit≠1 오류는 error로 전파 (부재 아님) |
| pgrep exit code | 0=매칭, 1=매칭 없음 | procps-ng / BusyBox | 1은 `nil, nil` (정상적인 "없음") |
| pgrep stdout | pid 목록 (변경 후: `pid cmdline` 줄) | pgrep | 파싱 불가 토큰은 건너뛴다 |
| 콘솔 자신의 journal 디렉터리 | **(추가 입력)** `engineJournalDir(root)`의 결과 | 호출부 | 해석 실패는 호출부가 error로 올린다 |

불변식: 이 함수는 시그널을 보내지 않는다. **목록만** 만든다. 그러나 그 목록이
`stopEngine`의 SIGTERM 대상이 되므로, 여기서의 오탐은 곧 잘못된 프로세스 종료다.

기존 불변식 두 개는 유지된다.

- exit 1(매칭 없음)과 실행 오류는 다른 것이다. 전자는 빈 목록, 후자는 error.
- 이 함수와 `tools/engine-autostart.sh`는 같은 후보 패턴을 쓴다 (drift 테스트).

## Branches and early returns

편집 전 이 함수는 넷을 한 몸에 갖고 있었다 (pgrep 오류, exit 1, 순회, 파싱 실패).
편집 후에는 두 질문이 분리되어 여기에는 열거 오류 전파 하나만 남고 나머지는
`pgrepEngineLines`·`enginePIDsForJournal`·`splitProcessLine`으로 옮겨갔다. 각자의
map이 그 분기를 계속 고정한다.

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `engineListProcesses()` 오류 | 없음 | `error` — 열거 실패, 부재 아님 | `TestEnumerationFailureKeepsTheRefusal` |
| — | 성공 경로 → 소유 판정 | 없음 | 우리 journal의 pid 목록 | `TestStoppingFindsTheEngineTheConsoleStarted`, `TestStoppingDoesNotSignalAnotherProfilesEngine` |
| (옮김) | pgrep 실행과 exit 1 | 없음 | `pgrepEngineLines` | `TestEnumerationFailureKeepsTheRefusal` |
| (옮김) | 줄 파싱과 비양수 pid | 건너뜀 | `splitProcessLine` | `TestUnparsableProcessLinesAreDropped` |
| (옮김) | journal 디렉터리 일치 판정 | 건너뜀 | `enginePIDsForJournal` | `TestOnlyThisProfilesEngineIsFound` |

### 이 change가 바꾸는 것

| | 변경 전 | 변경 후 |
|---|---|---|
| 패턴 | `tossctl engine run` (연속 문자열) | `tossctl( .*)? engine run` (토큰 경계 ERE) |
| pgrep 호출 | `-f` | `-a -f` — pid와 명령줄을 함께 |
| 출력 해석 | `strings.Fields` → 전부 pid | 줄 단위 `pid cmdline` 분리 |
| 소유 판정 | **없음** | journal 디렉터리 일치 시에만 채택 |
| 콘솔이 spawn한 엔진 | 찾지 못함 | 찾음 |
| 다른 프로필의 엔진 | 찾지 못함(패턴이 못 맞아서) | 찾지 않음(소유 판정이 거름) |

마지막 두 줄이 이 change의 전부다. "다른 프로필을 건드리지 않는다"는 결과는 같지만
근거가 우연에서 판정으로 바뀐다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `exec.Command("pgrep", …).Output()` | 프로세스 열거 | timeout 없음, 재시도 없음. exit 1은 정상 | AST |
| `errors.As` / `ExitCode` | "매칭 없음"과 실패 구분 | — | AST |
| `strconv.Atoi` | pid 파싱 | 실패는 건너뜀 | AST |
| **(추가)** `engineJournalDir` | 명령줄의 `--config-dir`를 journal 디렉터리로 해석 | 콘솔이 자기 경로를 구할 때와 **같은 함수** | design D2 |

live config binding 없음. 이 함수는 설정 파일·토글·브로커를 읽지 않는다.

## State mutations and fallbacks

- 도메인 변경 없음. 주문·손절·익절·사이징·Guardian·원장·대사·인증·체결 어느 것도 이
  함수에서 바뀌지 않는다.
- side effect 없음 — 외부 프로세스 `pgrep` 1회 실행이 전부이며 읽기 전용이다.
- fallback: 파싱할 수 없는 줄은 **버린다**. 소유를 증명할 수 없는 프로세스도 버린다
  (design D3). 두 fallback 모두 목록을 좁히는 방향이다.

## Safety conclusion

- Safe edit boundary: 프로세스 **발견**. 시그널·flock·게이트는 이 함수 밖이다.
- 이 change는 목록을 두 방향으로 동시에 움직인다. 넓히는 쪽(패턴)은 콘솔이 자기가 띄운
  엔진을 찾게 하고, 좁히는 쪽(소유 판정)은 남의 엔진을 대상에서 뺀다. 좁히는 쪽이 없으면
  넓히는 쪽 혼자서는 새 위험을 만든다 — 호스트 PID namespace는 컨테이너 프로세스를 본다.
- 손절 즉시성 방향: 다른 프로필의 엔진을 SIGTERM하지 않는 것이 이 change에서 즉시성을
  지키는 조항이다. 우리 엔진을 세우는 것은 운영자의 명시적 요청이므로 약화가 아니다.
