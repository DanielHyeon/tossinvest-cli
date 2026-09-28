# Function Logic Map: `SeverityOf`

- Source: `internal/obs/event.go` (`347`–`352`)
- Qualified: `SeverityOf`
- AST evidence: `ast.json` (`source_sha256` 7732f564d6e5b496…)
- Risk scan: `risk-pattern-report.md`
- 분기 1 · return 2 · 호출 0

**역할.** 이벤트 **종류** 하나를 받아 등급을 답한다. `criticalEvents` map만 본다 — 분기가 하나다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` | `EventType` | 호출자 | 없음 — 순수 |
| `criticalEvents` | 등급 표(18종) | 소스의 리터럴 map | 미등재는 `SeverityNormal` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:348` `if criticalEvents[t] {` | — | :349, :351 | 예 |

## Calls and live bindings

없다.

브로커·원장에 닿는 호출의 오류·타임아웃 계약은 각 호출자의 것이며, 이 함수는 그것을 되던진다(위 표의 return 열이 그 자리다).

## State mutations and fallbacks

없다.

## Safety conclusion

- **Safe edit boundary**: **이 함수의 편집 경계는 Q1의 답에 달렸다(4판, r3 N4).** 결정 (2)는 등급을 「이벤트 종류가 아니라 사실로」 가르라고 한다. B1은 종류만 보므로 같은 종류의 두 발신 자리(exit 관측 · reconcile 대사)에 다른 등급을 줄 수 없다. Q1이 (a) 새 이벤트 종류 등재면 이 함수 **본문은 불변**이고 map만 늘어난다. (b) 등급을 `Event`에 싣고 이 함수의 계약을 바꾸면 **경계를 다시 선언**해야 한다 — 번들 재생성과 재리뷰가 필요하다.
- **High-risk impact**: yes — 이 답이 알림의 durable 여부와 진입 차단 도달 여부를 정한다.
