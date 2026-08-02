# Function Logic Map: `startEngine`

- Source: `cmd/tossctl/engineproc.go`
- AST evidence: `ast.json` (revision: current)
- Change: a059-console-finds-the-engine-it-owns
- Risk scan: `risk-pattern-report.md`

이 map은 편집 **전에** 작성했다 (tasks.md 1.1). 엔진 기동 경로이므로 면제하지 않는다.
a056이 같은 함수의 map을 남겼고, 이 change는 그 위에 **입력 하나**만 바꾼다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root.configDir` | journal 디렉터리로 해석 가능한 경로 | `engineJournalDir` | 오류 반환, 기동 없음 |
| 자기 실행 파일 경로 | 존재하는 실행 파일 | `binstamp.SelfPath` | 오류 반환, 기동 없음 |
| 엔진 활성 마커 | mtime이 `enginelock.StaleAfter`(5분) 안이면 fresh | `enginelock.Read` | **자문 신호.** 단독으로 거부 근거가 될 수 없다 (a056) |
| 엔진 프로세스 목록 | **이 콘솔이 소유한** 엔진의 PID | `engineFindProcesses(dir)` | 오류면 부재를 주장할 수 없음 → 기존 거부 유지 |
| journal flock | 단일 writer | spawn된 `engine run`이 첫 동작으로 획득 | **정본 배타** |

불변식은 a056에서 그대로다. 이 함수는 브로커·자격증명·토큰을 쥐지 않고, 게이트·인터록·
락 판정은 전부 spawn된 `engine run`이 자기 안에서 한다.

**이 change가 바꾸는 입력은 넷째 줄 하나다.** `engineFindProcesses`가 지금까지
컨테이너에서 언제나 빈 목록을 돌려주었으므로 `observed`는 언제나 false였다. 즉 a056이
만든 결합 판정에 **거짓 입력이 들어가고 있었다.** 이 change는 그 입력을 참으로 만든다.

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `engineJournalDir` 오류 | 없음 | 오류 | 기존 커버 |
| B2 | `binstamp.SelfPath` 오류 | 없음 | 오류 | 기존 커버 |
| B3 | 마커 fresh **AND** (프로세스 관측 **OR** 열거 오류) | 없음 | "엔진이 이미 실행 중이다 (pid, 갱신시각)" | `TestAGhostMarkerDoesNotRefuseAStart`, `TestAFreshMarkerWithALiveProcessStillRefuses`, `TestEnumerationFailureKeepsTheRefusal` |
| B4 | 프로세스 관측 (마커 무관) | 없음 | "엔진 프로세스가 이미 있다 (pids)" | `TestStartingIsRefusedWhenAProcessIsAlreadyThere` |
| B5 | `engineSpawnDetached` 오류 | 없음 | 오류 | 기존 커버 |
| B6 | probe 창 안에 즉시 종료 | spawn됨 | 로그 tail + 오류 | `TestARefusedStartReportsTheEnginesOwnLog` |
| B7 | 즉시 종료가 오류 없이 | spawn됨 | 로그 tail + "오류 없이" | 기존 커버 |
| — | 위 어느 것도 아님 | **spawn** | "엔진을 시작했다 — 로그 <path>" | `TestStartingSpawnsTheEngineWithThisProfilesConfigDir`, `TestAStaleMarkerDoesNotBlockAStart` |

분기 **구조는 바뀌지 않는다.** 바뀌는 것은 `engineFindProcesses` 호출이 journal
디렉터리를 인자로 받는다는 점 하나이며, `dir`은 B1 직후에 이미 손에 있다.

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineJournalDir` | journal 디렉터리 결정 | 오류 전파, 재시도 없음 | AST |
| `binstamp.SelfPath` | spawn할 실행 파일 | 오류 전파 | AST |
| `enginelock.Read` | 마커 신선도와 PID·갱신시각 | 파일 없음/파손은 `Running=false` | AST + `internal/enginelock` |
| `engineFindProcesses(dir)` | **이 콘솔이 소유한** 엔진 관측 | 오류는 **부재 아님** (a056 D3) | AST, seam |
| `markerRefusesStart` | a056의 결합 판정 | 순수 함수 | AST |
| `engineSpawnDetached` | 엔진 기동 | 오류 전파 | AST, seam |
| `readLogTail` | 즉시 종료 시 엔진 자신의 사유 | 없음 | AST |

호출 **계약**은 바뀌지 않는다. 시그니처 하나가 인자를 받는다.

## State mutations and fallbacks

- 도메인 변경 없음.
- 유일한 side effect는 `engine run` 프로세스 spawn이며, 그 프로세스가 자기 게이트·
  인터록·flock을 스스로 검사한다.
- fallback: 열거 실패는 부재로 읽지 않는다 (a056 D3, 변경 없음).

## Safety conclusion

- Safe edit boundary: 기동 **전 안내** 판정. 배타 자체는 건드리지 않는다.
- 이 change는 a056의 거부 분기를 **컨테이너에서 처음으로 도달 가능하게** 만든다.
  방향은 거부가 늘어나는 쪽이지만, 늘어나는 거부는 전부 "엔진이 실제로 돌고 있을 때"이며
  이는 `엔진 런타임 수명주기`의 `두 번째 인스턴스 기동` scenario가 이미 정한 결과다.
- 유령 마커(a056의 대상)는 여전히 기동을 막지 못한다. 소유 판정이 유령을 되살리지
  않는지가 `TestAGhostMarkerDoesNotRefuseAStart` 회귀 고정의 요점이다.
