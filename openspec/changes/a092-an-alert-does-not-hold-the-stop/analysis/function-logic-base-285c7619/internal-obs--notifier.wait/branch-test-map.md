# Branch Test Map: `Notifier.wait`

Source: `internal/obs/notifier.go` (410-420). AST 기준 분기 2 / 이탈 1 /
defers 0 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:412` `RetryDelay <= 0` → 2s | **없음** — 아래 열거 참조 | no | **no** |
| B2 | `:416` `Clock == nil` → `clock.System()` | **없음** — `obs_test.go`의 `Notifier` 리터럴은 가짜 시계를 채운다. 실시계 분기는 프로덕션에서만 탄다 | no | no |

> **18라운드 B-P5.** 이 표는 B1을 *"채우지 않는 테스트에서 간접 통과"*로 두고 GREEN
> `yes`를 적고 있었다. 그런 테스트는 없다. `internal/obs`의 `Notifier` 리터럴 12개를
> 전수로 갈랐다: `RetryDelay`를 채우는 것이 9개(`a096:60`·`a096:311`·`a096b:110`·
> `a097:47`·`a097:144`·`escalation:43`·`obs_test:347`·`obs_test:570`·`obs_test:589`),
> 채우지 않는 것이 3개(`measurement_test:168`·`mode_test:91`·`obs_test:603`)인데
> **그 셋은 `Journal`이 없다.** `Journal`이 없으면 `notifyCritical` B1 `:171`에서
> `publishBestEffort`로 빠지므로 `deliver`에도 `wait`에도 도달하지 않는다.
>
> **즉 프로덕션이 실제로 쓰는 기본값 2s는 어떤 테스트도 통과시키지 않는다.**
> "간접 통과"는 관측이 아니라 추정이었다.

## 프로덕션과 테스트가 정확히 반대로 탄다

| | B1 (`RetryDelay` 0) | B2 (`Clock` nil) |
|---|---|---|
| 프로덕션 (`newNotifier`) | **탄다** — 필드 없음 | 안 탄다 — `Clock: clk` 채움 |
| 테스트 (`obs_test.go`) | **안 탄다** | **안 탄다** |

**즉 프로덕션이 쓰는 기본값 2s는 어느 테스트도 단언하지 않는다.**
a092는 `newNotifier`가 `RetryDelay`를 채우게 만들어 B1을 프로덕션에서도 안 타게 한다.

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | `newNotifier`가 만든 `Notifier`의 `RetryDelay` | 0이 아니라 명시된 값 (현행 RED: 0) |
| R2 | B1 (`RetryDelay <= 0`) | 여전히 2s — **회귀 없음**을 단언 |
