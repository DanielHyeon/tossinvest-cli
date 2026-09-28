# Function Logic Map: `mergeNotifications`

- Source: `internal/config/notifications.go` (`94`–`110`)
- Qualified: `mergeNotifications`
- AST evidence: `ast.json` (`source_sha256` 9b4124f002ca761c…)
- Risk scan: `risk-pattern-report.md`
- 분기 3 · return 2 · 호출 3

**역할.** 설정 파일의 알림 블록을 엔진 설정에 옮긴다. 검증에 실패한 블록은 전체를 0으로 만들고 사유를 남긴다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `raw` | 파일의 알림 블록 | config.json | B1 — 없으면 기본값(꺼짐) 유지 |
| `next.validate()` | 블록 검증 | `Notifications.validate` | B3 — 실패면 `Notifications{Rejected: why}` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:95` `if raw == nil {` | `strings.TrimSpace` | :96 | 예 |
| B2 | if | `:102` `if raw.Enabled != nil {` | — | — | 예 |
| B3 | if | `:105` `if why := next.validate(); why != "" {` | `next.validate` | :107 | 예 |

## Calls and live bindings

`strings.TrimSpace` · `next.validate`(B3 조건).

결과값이 없다 — 오류를 돌려주지 않는다. 검증 실패는 블록 전체를 0으로 만들고 `Rejected`에 사유를 담는다(B3 창 return) — 그래서 거부된 블록의 `Enabled`는 거짓이다.

## State mutations and fallbacks

`cfg.Notifications` — 통과한 블록 또는 `Rejected`만 가진 0 블록.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** B3 창이 거부된 블록을 `Notifications{Rejected: why}`로 만든다 — 로드된 `Enabled`는 거짓이다. 그러므로 파일에 `enabled: true`를 적었어도 블록이 거부되면 a095에게는 「꺼짐」이다(8판, r7 F3 · Manager 처분: 「거부 = 꺼짐」을 명시하고 시험). 이 방향은 R4-2(전송 수단 부재를 꺼짐으로 읽지 않는다)와 반대다 — 거부는 전송 수단 부재가 아니라 설정 자체가 무효라는 사실이고, 결과가 critical을 만들지 않는 **안전 방향**이다.
- **High-risk impact**: yes — 알림 켜짐 판정의 원천이다.
