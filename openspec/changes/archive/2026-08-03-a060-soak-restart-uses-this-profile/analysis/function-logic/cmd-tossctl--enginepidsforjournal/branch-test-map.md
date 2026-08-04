# Branch Test Map: `enginePIDsForJournal`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at line 351 — journal 디렉터리 공백 | `TestTheDefaultProfileMatchesAnEngineStartedWithoutFlags` | n/a — a059 계약, 회귀 고정 | yes |
| B2 | range at line 357 — 후보 순회 | `TestOnlyThisProfilesEngineIsFound` | n/a — a059 계약 | yes |
| B3 | if at line 359 — 파싱 실패 또는 매처 불일치 | `TestUnparsableProcessLinesAreDropped` | n/a — a059 계약 | yes |
| B4 | if at line 363 — `--config-dir` 부재 → 기본값 | `TestTheDefaultProfileMatchesAnEngineStartedWithoutFlags` | n/a — a059 계약 | yes |
| B5 | if at line 366 — 정체 불일치 → 배제 | `TestOnlyThisProfilesEngineIsFound` | n/a — a059 계약 | yes |

이 change는 이 함수의 **분기를 하나도 바꾸지 않는다.** 추출이므로 위 다섯 줄은 전부
회귀 고정이고, 새 RED는 soak 쪽 map에 있다.

## 추출이 추출임을 무엇이 증명하는가

a059가 이 함수를 위해 쓴 테스트가 전부 green으로 남는 것 (tasks 5.4):

- `TestOnlyThisProfilesEngineIsFound`
- `TestTheDefaultProfileMatchesAnEngineStartedWithoutFlags`
- `TestUnparsableProcessLinesAreDropped`
- `TestStoppingDoesNotSignalAnotherProfilesEngine`
- `TestStoppingFindsTheEngineTheConsoleStarted`
- `TestStartingIsRefusedWhenOurOwnEngineIsAlreadyRunning`
- `TestBothButtonsAskAboutThisProfilesJournal`

하나라도 손대야 한다면 그것은 추출이 아니라 동작 변경이고, 그때는 이 map을 다시 써야
한다.

## 변이 검증 (tasks 5.4)

| 변이 | RED가 되어야 하는 테스트 | 되돌림 |
|---|---|---|
| 추출한 헬퍼에서 매처 검사 제거 | `TestUnparsableProcessLinesAreDropped` | 예정 |
| 엔진 wrapper가 `resolve`에 기본값을 안 넘김 | `TestTheDefaultProfileMatchesAnEngineStartedWithoutFlags` | 예정 |
