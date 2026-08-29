# Branch Test Map: `stopEngine`

편집 후 갱신. 편집 전 B8(`engineJournalDir`을 마지막에 구하고 오류를 삼키던 if)이
사라지고, 그 해석이 함수 맨 앞의 B1으로 올라왔다. 분기 개수는 아홉으로 같다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at line 236 — journal 디렉터리 해석 오류 → 전파 | 기존 `cmd/tossctl` 커버리지 | n/a — 신규지만 오류 전파뿐 | yes |
| B2 | if at line 240 — 열거 오류 | 기존 `cmd/tossctl` 커버리지 | n/a — 기존 계약 | yes |
| B3 | range at line 244 — pid 순회 | `TestStoppingSignalsAndWaits` | n/a | yes |
| B4 | if at line 245 — 이 콘솔은 건너뛴다 | `TestStoppingNeverSignalsThisProcess` | n/a — 기존 계약, 회귀 고정 | yes |
| B5 | if at line 248 — 시그널 오류 | 기존 `cmd/tossctl` 커버리지 | n/a | yes |
| B6 | if at line 253 — 시그널 대상 0건 | `TestNoEngineToStopIsNotAFailure` | n/a — 기존 계약 | yes |
| B7 | range at line 256 — 종료 대기 | `TestStoppingSignalsAndWaits` | n/a | yes |
| B8 | if at line 257 — 시간 내 미종료 | `TestAnEngineThatWillNotGoIsReportedRatherThanKilled` | n/a — 기존 계약 | yes |
| B9 | if at line 261 — 종료 후에도 마커 fresh | 기존 `cmd/tossctl` 커버리지 | n/a — a056 I1이 정당한 사용으로 판정 | yes |

fall-through(시그널 + 완주 대기 성공)는 `TestStoppingFindsTheEngineTheConsoleStarted`와
`TestStoppingDoesNotSignalAnotherProfilesEngine`이 덮는다.

## 이 change가 뒤집는 칸

| 관측 상태 | 변경 전 (컨테이너 실측) | 변경 후 |
|---|---|---|
| 우리 엔진 1건 실행 중 | `"실행 중인 엔진을 찾지 못했다."` — **엔진은 계속 돈다** | SIGTERM 1건, 완주 대기, 무엇을 세웠는지 안내 |
| 우리 엔진 + 남의 프로필 엔진 | 둘 다 못 찾음 | 우리 것만 SIGTERM |
| 남의 프로필 엔진만 | 못 찾음 | 시그널 없음, "찾지 못했다" |
| 아무것도 없음 | "찾지 못했다" | "찾지 못했다" — 동일 |
| 목록에 이 콘솔의 pid | 건너뜀 | 건너뜀 — 동일 |
| 엔진이 60초 안에 안 죽음 | 보고, kill 없음 | 보고, kill 없음 — 동일 |
| journal 디렉터리 해석 실패 | 시그널은 시도하고 마커 보고만 생략 | 아무것도 시그널하지 않고 오류 반환 |

첫 줄이 결함이다. 셋째 줄은 첫 줄을 고칠 때 반드시 함께 지켜야 하는 칸이고, 이 두 줄이
같은 change에 있는 이유다 — 첫 줄만 고치면 셋째 줄이 "남의 엔진에 SIGTERM"으로 뒤집힌다.

마지막 줄은 의도한 방향 전환이다. 소유를 판정할 수 없으면 시그널 대상을 고를 수 없고,
고르지 못할 때 하는 일은 아무것도 하지 않는 것이다.

## 변이 검증 (tasks 5.2, 2026-08-03 실행)

| 변이 | RED가 된 테스트 | 되돌림 |
|---|---|---|
| 소유 판정을 약화 (`\|\|` → `&&`) | `TestStoppingDoesNotSignalAnotherProfilesEngine` | yes |
| 패턴을 `tossctl engine run`으로 복원 | `TestStoppingFindsTheEngineTheConsoleStarted` | yes |
