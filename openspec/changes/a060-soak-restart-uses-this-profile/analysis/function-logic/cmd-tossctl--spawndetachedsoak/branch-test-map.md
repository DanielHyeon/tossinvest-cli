# Branch Test Map: `spawnDetachedSoak`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at line 224 — 로그 디렉터리 생성 실패 | 기존 `cmd/tossctl` 커버리지 | n/a — 편집하지 않음 | yes |
| B2 | if at line 228 — 로그 파일 열기 실패 | 기존 `cmd/tossctl` 커버리지 | n/a — 편집하지 않음 | yes |
| B3 | if at line 239 — `cmd.Start()` 실패 | 기존 `cmd/tossctl` 커버리지 | n/a — 편집하지 않음 | yes |
| — | 성공 경로의 argv | `TestTheSoakSpawnCarriesThisProfile`, `TestTheSoakPatternMatchesWhatTheConsoleSpawns` | yes | yes |

분기는 하나도 편집하지 않는다. 편집하는 것은 성공 경로가 `exec.Command`에 넘기는 argv다.

## 이 change가 뒤집는 칸

| 콘솔의 프로필 | 변경 전 자식 | 변경 후 자식 |
|---|---|---|
| `--config-dir /var/lib/tossos/config` | 기본 프로필 → **자격증명 없음, 즉시 종료** | 같은 프로필 → 정상 기동 |
| 플래그 없음 | 기본 프로필 | 기본 프로필 — 동일 |
| `--session-file`만 | 기본 프로필 | 세션 파일 상속 |

첫 줄이 프로덕션에서 관측된 상태다. `config/soak.log`가 같은 오류만 반복한다.

## 변이 검증 (tasks 5.1)

| 변이 | RED가 되어야 하는 테스트 | 되돌림 |
|---|---|---|
| argv를 `"soak","run"`으로 복원 | `TestTheSoakSpawnCarriesThisProfile` | 예정 |
