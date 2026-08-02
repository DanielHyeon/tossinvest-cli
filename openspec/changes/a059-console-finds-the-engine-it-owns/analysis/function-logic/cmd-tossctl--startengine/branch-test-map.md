# Branch Test Map: `startEngine`

편집 후 갱신. 분기 구조는 a056 이후 그대로이고, 이 change가 바꾼 것은 B3/B4가 읽는
입력을 만드는 호출의 인자 하나다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at line 179 — `engineJournalDir` 오류 | 기존 `cmd/tossctl` 커버리지 | n/a — 편집하지 않음 | yes |
| B2 | if at line 183 — `binstamp.SelfPath` 오류 | 기존 `cmd/tossctl` 커버리지 | n/a — 편집하지 않음 | yes |
| B3 | if at line 199 — 마커 fresh AND (관측 OR 열거 오류) | `TestAGhostMarkerDoesNotRefuseAStart`, `TestAFreshMarkerWithALiveProcessStillRefuses`, `TestEnumerationFailureKeepsTheRefusal`, `TestStartingIsRefusedWhenOurOwnEngineIsAlreadyRunning` | yes (마지막 하나) | yes |
| B4 | if at line 204 — 프로세스 관측 | `TestStartingIsRefusedWhenAProcessIsAlreadyThere` | n/a — 기존 계약, 회귀 고정 | yes |
| B5 | if at line 210 — spawn 오류 | 기존 `cmd/tossctl` 커버리지 | n/a | yes |
| B6 | select at line 217 — probe 창 안 즉시 종료 | `TestARefusedStartReportsTheEnginesOwnLog` | n/a — 기존 계약 | yes |
| B7 | if at line 220 — 즉시 종료가 오류 없이 | 기존 `cmd/tossctl` 커버리지 | n/a — 기존 계약 | yes |

fall-through(spawn 1회)는 `TestStartingSpawnsTheEngineWithThisProfilesConfigDir`,
`TestAStaleMarkerDoesNotBlockAStart`, `TestAGhostMarkerDoesNotRefuseAStart`가 덮는다.

## 입력이 참이 되면 무엇이 달라지는가

| 상황 (컨테이너) | 변경 전 실제 동작 | 변경 후 |
|---|---|---|
| 유령 마커, 엔진 없음 | 진행 (a056) | 진행 — 동일 |
| 마커 fresh, 우리 엔진이 실제로 돌고 있음 | **진행** (관측 실패로 B3·B4 모두 빠져나감) → flock이 거부 | **거부** — B3에 도달, 안내에 pid·갱신시각 |
| 마커 없음, 우리 엔진이 돌고 있음 | **진행** → flock이 거부 | **거부** — B4에 도달, 안내에 pid |
| 남의 프로필 엔진만 돌고 있음 | 진행 | 진행 — 소유 판정이 걸러서 우리 관측은 비어 있다 |

두 번째·세 번째 줄이 a056 `issues.md` I2가 "거부 분기 도달 불가"라고 기록한 것이다.
결과는 지금도 올발랐다(flock이 받았다). 달라지는 것은 운영자가 받는 **문장**이다.

## 변이 검증 (tasks 5.1~5.3, 2026-08-03 실행)

| 변이 | RED가 된 테스트 | 되돌림 |
|---|---|---|
| `engineFindProcesses`에 journal 디렉터리 대신 빈 문자열 전달 | `TestBothButtonsAskAboutThisProfilesJournal`, `TestStartingIsRefusedWhenOurOwnEngineIsAlreadyRunning` | yes |
| 패턴을 `tossctl engine run`으로 복원 | `TestStartingIsRefusedWhenOurOwnEngineIsAlreadyRunning` | yes |
