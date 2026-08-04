# Branch Test Map: `pgrepSoak`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at line 181 — pgrep 실행 오류 | `TestAFailureToLookForTheSoakIsReportedAndNothingIsStarted` | n/a — 기존 계약, 회귀 고정 | yes |
| B2 | if at line 183 — exit 1(매칭 없음) | `TestRestartingWithNothingRunningJustStartsOne` | n/a — 기존 계약, 회귀 고정 | yes |
| B3 | range at line 190 — 출력 각 줄 해석 | `TestOnlyThisRecordsSoakIsFound` | yes | yes |
| B4 | if at line 192 — 파싱 불가·비양수 pid | `TestOnlyThisRecordsSoakIsFound` | yes | yes |

## 이 change가 뒤집는 칸

| 관측된 명령줄 | 변경 전 | 변경 후 |
|---|---|---|
| `tossctl soak run` (플래그 없음, 기본 프로필) | 발견 | 기본 프로필 콘솔이면 발견, 아니면 제외 |
| `tossctl --config-dir <ours> … soak run` | **못 찾음** | **발견** |
| `tossctl --config-dir <남의 것> … soak run` | 못 찾음 | 제외 (소유 판정) |
| `tossctl --config-dir <ours> … engine run` | 못 찾음 | 못 찾음 — 패턴이 `soak run`을 요구 |
| pgrep 실행 오류 | error | error — 동일 |

첫 줄의 변화가 유일하게 좁아지는 칸이다. 기본 프로필이 아닌 콘솔은 이제 기본 프로필의
서베이를 자기 것으로 보지 않는다. 그것이 의도다 — 그 서베이는 다른 기록에 쓴다.

## 변이 검증 (tasks 5.2, 5.3)

| 변이 | RED가 되어야 하는 테스트 | 되돌림 |
|---|---|---|
| 패턴을 `tossctl soak run`으로 복원 | `TestTheSoakPatternMatchesWhatTheConsoleSpawns` | 예정 |
| 소유 판정 약화 | `TestOnlyThisRecordsSoakIsFound` | 예정 |
| 패턴에서 `soak`을 빼 `engine`도 잡게 | `TestTheSoakPatternIgnoresTheOtherSubcommands` | 예정 |
