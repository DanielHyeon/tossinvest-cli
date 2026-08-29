# Branch Test Map: `startEngine`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at line 138 — `engineJournalDir` 오류 | 기존 `cmd/tossctl` 커버리지 | n/a — 이 change가 편집하지 않음 | yes |
| B2 | if at line 142 — `binstamp.SelfPath` 오류 | 기존 `cmd/tossctl` 커버리지 | n/a — 이 change가 편집하지 않음 | yes |
| B3 | if at line 156 — 마커 fresh **AND** (프로세스 관측 OR 열거 오류) → 거부 | `TestAGhostMarkerDoesNotRefuseAStart`, `TestAFreshMarkerWithALiveProcessStillRefuses`, `TestEnumerationFailureKeepsTheRefusal`, `TestNoPathRefusesOnMarkerFreshnessAlone` | yes | yes |
| B4 | if at line 161 — 프로세스 관측 (마커 무관) → 거부 | `TestStartingIsRefusedWhenAProcessIsAlreadyThere` | n/a — 기존 계약, 회귀 고정 | yes |
| B5 | if at line 167 — `engineSpawnDetached` 오류 | 기존 `cmd/tossctl` 커버리지 | n/a — 이 change가 편집하지 않음 | yes |
| B6 | select at line 174 — probe 창 안 즉시 종료 | `TestARefusedStartReportsTheEnginesOwnLog` | n/a — 기존 계약 | yes |
| B7 | if at line 177 — 즉시 종료가 오류 없이 | 기존 `cmd/tossctl` 커버리지 | n/a — 기존 계약 | yes |

B3만 이 change가 편집한 분기다. 나머지는 회귀 고정이다. 어느 거부에도 걸리지 않는
fall-through(spawn 1회)는 `TestStartingSpawnsTheEngineWithThisProfilesConfigDir`,
`TestAStaleMarkerDoesNotBlockAStart`, `TestAGhostMarkerDoesNotRefuseAStart`가 덮는다.

## 이 change가 뒤집는 칸

| 마커 | 프로세스 열거 | 변경 전 | 변경 후 |
|---|---|---|---|
| fresh | 관측됨 | 거부 (마커 문구) | 거부 (마커 문구) — 동일 |
| fresh | 없음 | **거부** | **진행** — 유령 마커 |
| fresh | 열거 오류 | 거부 | 거부 — 동일 |
| stale/없음 | 관측됨 | 거부 (프로세스 문구) | 거부 (프로세스 문구) — 동일 |
| stale/없음 | 없음 | 진행 | 진행 — 동일 |
| stale/없음 | 열거 오류 | 진행 | 진행 — 동일 |

여섯 칸 중 하나만 움직인다. 변경 전 B3의 조건은 `status.Running` 하나였고 `return`으로
끝나서, 바로 아래 B4(실제 프로세스 검사)에 **도달하지 못했다.** 진리표 여섯 칸 자체는
`TestMarkerRefusesStartOnlyWithCorroboration`이 프로세스 spawn 없이 직접 고정한다.

## 변이 검증 (tasks 5.1, 5.2)

| 변이 | RED가 된 테스트 | 되돌림 |
|---|---|---|
| B3 조건을 `status.Running` 단독으로 복원 | `TestAGhostMarkerDoesNotRefuseAStart`, `TestNoPathRefusesOnMarkerFreshnessAlone` | yes |
| `observed`를 항상 false로 | `TestAFreshMarkerWithALiveProcessStillRefuses`, `TestStartingIsRefusedWhenAProcessIsAlreadyThere` | yes |
| 열거 오류를 부재로 취급 (`findErr != nil` → `false`) | `TestEnumerationFailureKeepsTheRefusal` | yes |

## 대체된 기존 계약

`TestStartingIsRefusedWhileTheMarkerIsFresh`가 (마커 fresh, 프로세스 없음) 칸을 거부로
고정하고 있었다. 원래 근거는 테스트 주석에 이렇게 적혀 있었다 — "answer is 'already
running' rather than a spawned process that immediately loses the race for the flock."

그 근거가 사는 경우는 **엔진이 실제로 살아 있는데 pgrep이 못 보는** 경우뿐이다(예: 호스트
native 엔진 + 컨테이너 콘솔, 서로 다른 PID namespace — issues.md I2). 그때도 flock이 두 번째
런타임을 죽이므로 journal 단일 writer는 유지되고, 잃는 것은 안내 문구의 품질이다. 반대편에
있는 것은 2026-08-02에 실측된 8분 감시 공백이며, 그동안 OPEN 포지션 5건과 활성 exit 정책
4건에 exit 루프가 돌지 않았다. 문구를 잃고 감시를 지킨다.
