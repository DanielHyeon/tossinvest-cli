# Function Logic Map: `resolveNotificationPublisher`

- Source: `internal/app/engine/notifications.go` (`67`–`106`)
- Qualified: `resolveNotificationPublisher`
- AST evidence: `ast.json` (`source_sha256` 31a1d84d8529cd68…)
- Risk scan: `risk-pattern-report.md`
- 분기 5 · return 4 · 호출 8

**역할.** 설정의 알림 블록에서 전송기를 만든다. 전송기가 없는 사유를 `notificationResolution`에 남긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `cfg.Rejected` | 설정 거부 사유 | `internal/config` | B2 — 거부된 블록은 이미 0으로 만들어져 있다(nil) |
| `cfg.Enabled` | 알림 켜짐 | 설정 | B3 — 거짓이면 nil(파일에 남은 topic으로 다시 켜지 않음) |
| topic | 환경변수 또는 설정 | `envNtfyTopic` · `cfg.Topic` | B5 — 켜졌는데 topic이 없으면 nil + 사유 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:69` `if getenv == nil {` | `strings.TrimSpace` | — | 예 |
| B2 | if | `:77` `if resolution.Refused != "" {` | — | :81 | 예 |
| B3 | if | `:83` `if !cfg.Enabled {` | `getenv`, `strings.TrimSpace` | :87 | 예 |
| B4 | if | `:91` `if topic == "" {` | `strings.TrimSpace` | — | 예 |
| B5 | if | `:94` `if topic == "" {` | `getenv`, `ntfy.UsesPublicService`, `strings.TrimSpace` | :97, :105 | 예 |

## Calls and live bindings

`getenv` · `strings.TrimSpace` · `ntfy.UsesPublicService`.

결과는 `(Publisher, notificationResolution)`다 — 오류 대신 해석 결과의 `Refused`에 사유를 담고 전송기를 nil로 돌려준다(B2 · B5 창).

## State mutations and fallbacks

없다 — 전송기와 해석 결과를 돌려준다.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** 전송기 nil은 **세 경우**에서 나온다: B2 설정 거부 · B3 알림 꺼짐 · B5 **알림 켜짐 + topic 없음**. 그러므로 `Publisher == nil`은 「알림 꺼짐」의 대용이 될 수 없다(r4 R4-2) — Q1의 판정 근거는 설정의 `enabled`(이 함수의 `cfg.Enabled`, B3의 조건)로 한정한다. 켜짐 + topic 없음은 a095에게 「켜짐」이고, 그 critical 행은 a124 정본의 전송 수단 부재 처리를 탄다.
- **High-risk impact**: yes — 알림 켜짐 판정의 근거다.
