# Branch Test Map: `restartSoak`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at line 89 — `binstamp.SelfPath` 오류 | 기존 `cmd/tossctl` 커버리지 | n/a — 편집하지 않음 | yes |
| B2 | if at line 95 — 열거 오류 | `TestAFailureToLookForTheSoakIsReportedAndNothingIsStarted` | n/a — 기존 계약 | yes |
| B3 | range at line 100 — pid 순회 | `TestRestartingTheSoakInterruptsItThenStartsItAgain` | n/a | yes |
| B4 | if at line 101 — 이 콘솔은 건너뛴다 | `TestTheRestartNeverSignalsThisProcess` | n/a — 기존 계약, 회귀 고정 | yes |
| B5 | if at line 104 — 시그널 오류 | 기존 `cmd/tossctl` 커버리지 | n/a | yes |
| B6 | range at line 110 — 종료 대기 | `TestRestartingTheSoakInterruptsItThenStartsItAgain` | n/a | yes |
| B7 | if at line 111 — 시간 내 미종료 → spawn 없음 | 기존 `cmd/tossctl` 커버리지 | n/a — 기존 계약 | yes |
| B8 | if at line 116 — `prepareSpawn` 제공 | `soakproc_openapi_test.go` | n/a | yes |
| B9 | if at line 117 — `prepareSpawn` 오류 | `soakproc_openapi_test.go` | n/a | yes |
| B10 | if at line 121 — spawn 오류 | 기존 `cmd/tossctl` 커버리지 | n/a | yes |
| B11 | switch at line 125 — 안내 문구 분기 | `TestRestartingWithNothingRunningJustStartsOne` | n/a | yes |
| B12 | case at line 126 — 0건 | `TestRestartingWithNothingRunningJustStartsOne`, `TestTheRestartDoesNotSignalAnotherRecordsSoak` | yes (후자) | yes |
| B13 | case at line 128 — 1건 | `TestRestartingTheSoakInterruptsItThenStartsItAgain` | n/a | yes |
| B14 | case at line 131 — 다수 | 기존 `cmd/tossctl` 커버리지 | n/a | yes |

## 이 change가 뒤집는 칸

| 관측 상태 | 변경 전 (프로덕션 실측) | 변경 후 |
|---|---|---|
| 콘솔이 격리 프로필, 서베이 없음 | 새 서베이 spawn → **자격증명 못 찾고 즉시 사망** | 프로필을 물려받아 정상 기동 |
| 콘솔이 격리 프로필, 자기 서베이 실행 중 | 못 찾음 → 세우지 않고 두 번째를 띄운다(그것도 즉사) | 찾아서 SIGINT → 완주 대기 → 재기동 |
| 남의 기록 서베이만 실행 중 | 못 찾음 | 시그널 없음 |
| 목록에 이 콘솔의 pid | 건너뜀 | 건너뜀 — 동일 |
| 서베이가 30초 안에 안 죽음 | 오류, spawn 없음 | 오류, spawn 없음 — 동일 |

## 변이 검증 (tasks 5.1~5.3)

| 변이 | RED가 되어야 하는 테스트 | 되돌림 |
|---|---|---|
| argv를 `"soak","run"`으로 복원 | `TestTheSoakSpawnCarriesThisProfile` | 예정 |
| 소유 판정 약화 | `TestTheRestartDoesNotSignalAnotherRecordsSoak` | 예정 |
