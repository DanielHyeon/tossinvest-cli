# Branch Test Map: `Notifier.wait`

Source: `internal/obs/notifier.go` (290-300). AST 기준 분기 2 / 이탈 1 /
defers 0 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:292` `RetryDelay <= 0` → 2s | `obs_test.go`의 재시도 테스트가 `RetryDelay`를 채우므로 이 분기는 **거의 안 탄다**. 채우지 않는 테스트에서 간접 통과 | no | yes |
| B2 | `:296` `Clock == nil` → `clock.System()` | **없음** — `obs_test.go`의 `Notifier` 리터럴은 가짜 시계를 채운다. 실시계 분기는 프로덕션에서만 탄다 | no | no |

## 프로덕션과 테스트가 정확히 반대로 탄다

| | B1 (`RetryDelay` 0) | B2 (`Clock` nil) |
|---|---|---|
| 프로덕션 (`newNotifier`) | **탄다** — 필드 없음 | 안 탄다 — `Clock: clk` 채움 |
| 테스트 (`obs_test.go`) | 대체로 안 탄다 | **안 탄다** |

**즉 프로덕션이 쓰는 기본값 2s는 어느 테스트도 단언하지 않는다.**
a092는 `newNotifier`가 `RetryDelay`를 채우게 만들어 B1을 프로덕션에서도 안 타게 한다.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | `newNotifier`가 만든 `Notifier`의 `RetryDelay` | 0이 아니라 명시된 값 (현행 RED: 0) |
| R2 | B1 (`RetryDelay <= 0`) | 여전히 2s — **회귀 없음**을 단언 |
