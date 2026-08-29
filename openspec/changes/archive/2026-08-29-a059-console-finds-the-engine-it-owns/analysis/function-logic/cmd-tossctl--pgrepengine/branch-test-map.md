# Branch Test Map: `pgrepEngine`

편집 후 갱신. 편집 전 이 함수는 열거·파싱·pid 수집을 한 몸에 갖고 있었고 분기가 넷이었다
(B1 pgrep 오류, B2 exit 1, B3 순회, B4 파싱 실패). 지금은 두 질문이 분리되어 각자의
함수에 산다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at line 312 — 열거 오류 전파 | `TestEnumerationFailureKeepsTheRefusal` (seam 경유) | n/a — a056 계약, 회귀 고정 | yes |
| — | 성공 경로: 열거 → 소유 판정 | `TestStoppingFindsTheEngineTheConsoleStarted`, `TestStoppingDoesNotSignalAnotherProfilesEngine`, `TestStartingIsRefusedWhenOurOwnEngineIsAlreadyRunning` | yes | yes |

옮겨간 분기는 아래에서 계속 고정된다.

| 옮겨간 곳 | 무엇 | 테스트 |
|---|---|---|
| `pgrepEngineLines` | pgrep 실행, exit 1과 실행 실패의 구분 | `TestEnumerationFailureKeepsTheRefusal` |
| `enginePIDsForJournal` | 소유 판정, 증명 불가 항목 폐기 | `TestOnlyThisProfilesEngineIsFound`, `TestTheDefaultProfileMatchesAnEngineStartedWithoutFlags`, `TestUnparsableProcessLinesAreDropped` |
| `splitProcessLine` | `pid command` 한 줄 파싱 | `TestUnparsableProcessLinesAreDropped` |
| `engineCommandConfigDir` | `--config-dir` 두 표기 되뽑기 | `TestOnlyThisProfilesEngineIsFound` |

## 이 change가 뒤집는 칸

| 관측된 명령줄 | 변경 전 | 변경 후 |
|---|---|---|
| `tossctl engine run` (플래그 없음, 기본 프로필) | 발견 | 발견 — 동일 |
| `tossctl --config-dir <ours> … engine run` | **못 찾음** | **발견** |
| `tossctl --config-dir <남의 것> … engine run` | 못 찾음 (패턴 불일치) | 제외 (소유 판정) |
| `tossctl --config-dir <ours> … console …` | 못 찾음 | 못 찾음 — 동일 |
| `tossctl soak run` | 못 찾음 | 못 찾음 — 동일 |
| pgrep 실행 오류 | error | error — 동일 |
| pgrep exit 1 | 빈 목록 | 빈 목록 — 동일 |

두 번째 줄이 결함이고, 세 번째 줄은 그것을 고칠 때 생기는 위험이다. 결과는 같아 보이나
근거가 다르다 — 변경 전에는 패턴이 못 맞아서 우연히 안전했고, 변경 후에는 판정이 있어서
안전하다.

## 변이 검증 (tasks 5.1~5.3, 2026-08-03 실행)

| 변이 | RED가 된 테스트 | 되돌림 |
|---|---|---|
| 패턴을 `tossctl engine run`으로 복원 | `TestTheProcessPatternMatchesWhatTheConsoleSpawns`, `TestStoppingFindsTheEngineTheConsoleStarted`, `TestStartingIsRefusedWhenOurOwnEngineIsAlreadyRunning`, `TestTheEngineProcessPatternMatchesTheAutostartScript` | yes |
| 소유 판정을 약화 (`\|\|` → `&&`, 다른 journal도 채택) | `TestOnlyThisProfilesEngineIsFound`, `TestTheDefaultProfileMatchesAnEngineStartedWithoutFlags`, `TestStoppingDoesNotSignalAnotherProfilesEngine` | yes |
| 패턴에서 토큰 경계 제거 (`tossctl.*engine run`) | `TestTheProcessPatternIgnoresTheOtherSubcommands` | yes |
