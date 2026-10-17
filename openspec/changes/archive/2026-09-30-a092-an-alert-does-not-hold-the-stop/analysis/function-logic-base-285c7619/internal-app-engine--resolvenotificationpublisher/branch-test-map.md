# Branch Test Map: `resolveNotificationPublisher`

Source: `internal/app/engine/notifications.go` (67-106). AST 기준 분기 5 / 이탈 4 /
defers 0 / go_statements 0.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:69` `getenv == nil` → `os.Getenv`로 채움, **반환 없음** | `a074_notification_wiring_test.go` (주입 없이 부르는 경로) | no | yes |
| B2 | `:77` `Rejected`가 차 있음 → nil + 사유 | 같은 파일 (거부 블록) | no | yes |
| B3 | `:83` `Enabled == false` → nil, 오류 아님 | 같은 파일 — **기본 경로** | no | yes |
| B4 | `:91` 환경이 비어 파일 topic 사용 | 같은 파일 | no | yes |
| B5 | `:94` 양쪽 다 비었음 → nil + 사유 | 같은 파일 | no | yes |

**분기 커버리지는 좋다.** 빈 것은 분기가 아니라 `:101` 리터럴이 **비우는 필드**에
대한 단언이다.

## 없는 것

`a074_notification_wiring_test.go`는 `obs.Ntfy`가 만들어지는지, `BaseURL`·`Topic`이
옳은지를 본다. **`Timeout`을 보는 테스트는 없다** — 그래서 10초 기본값은 테스트에
나타나지 않고, `internal-obs--ntfy.publish` branch-test-map이 이미 적은 대로
`Ntfy.Publish` B3(`timeout <= 0`)는 **모든 테스트에서 통과한다**(테스트도 `Timeout`을
비우므로).

## 필요한 RED

| # | Scenario | 기대 |
|---|---|---|
| R1 | `resolveNotificationPublisher`가 돌려준 `*obs.Ntfy`의 `Timeout` | 명시된 값 (현행 RED: 0) |
| R2 | B2·B3·B5 세 이탈 (`:81`·`:87`·`:97`) | **여전히 nil** — 회귀 없음 |
| R3 | `cmd/tossctl/notificationsettings.go:151`의 시험 발송용 `Ntfy` | `Timeout` 무변화(10s 기본) — 범위 밖임을 고정 |
