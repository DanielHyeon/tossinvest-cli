# Branch Test Map: `Context.ExitObserver`

- Source: `internal/app/engine/exitwiring.go` (319-356); **편집 뒤** 측정 — `branch_coverage.py`, 연결 워크트리 `c6e2e3ac`, `./internal/app/engine`의
  시험 11개(`ExitObserver|TestA092|TestProductionGuardian|Tracer`)를 하나씩(`analysis/harness/coverage-post-exitobserver.json`). 편집 전 표는 `analysis/pre-edit/unit2/`.
- 재번호(편집 전 → 뒤): B1~B3 그대로 · B4 :338 → :343 · B5 :341 → :346 · **B6 :349 새 분기**(Announcer 기본값) · 편집 전 B6(Floor :344) → B7 :352.

| Branch | AST anchor | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | if at 320:2 | nil Context | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B2 | if at 323:2 | 게이트 미검증 | `TestTheExitObserverIsUnavailableWithoutAVerifiedGate` | 해당 없음(분기 불변) | 블록 323.28-325.3, PASS |
| B3 | if at 327:2 | Guardian이 발의 불가 | (미실행) | 해당 없음(분기 불변) | 측정 표본의 시험 0개 |
| B4 | if at 343:2 | Names 기본값 | `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver` · `TestA092ExitObserverGetsRecordOnlyAlertPaths` | 해당 없음(분기 불변) | 블록 343.23-345.3, PASS |
| B5 | if at 346:2 | Alerts 기본값 = 기록 전용 | `TestA092ExitObserverGetsRecordOnlyAlertPaths` | 편집 전: `Alerts = *obs.Notifier` 로 FAIL | 블록 346.45-348.3, PASS · 변이 U20 CAUGHT |
| B6 | if at 349:2 | Announcer 기본값 = 기록 전용 | `TestA092ExitObserverGetsRecordOnlyAlertPaths` | 편집 전: `Announcer = <nil>` 로 FAIL | 블록 349.48-351.3, PASS · 변이 U21 CAUGHT |
| B7 | if at 352:2 | Floor = exit retrier 복사본 | `TestA092ExitObserverGetsRecordOnlyAlertPaths` | 편집 전: 공유 Retrier 로 FAIL | 블록 352.23-354.3, PASS · 변이 U18 · U19 CAUGHT |
