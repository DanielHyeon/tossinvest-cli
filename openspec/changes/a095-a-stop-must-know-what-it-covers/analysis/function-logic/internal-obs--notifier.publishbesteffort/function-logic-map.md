# Function Logic Map: `Notifier.publishBestEffort`

- Source: `internal/obs/notifier.go` (`161`–`173`)
- Qualified: `Notifier.publishBestEffort`
- AST evidence: `ast.json` (`source_sha256` 0bc75668ff17c3d6…)
- Risk scan: `risk-pattern-report.md`
- 분기 2 · return 1 · 호출 6

**역할.** 일반 등급 알림을 보내고 잊는다. outbox 행도 재시도도 없다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `n.Publisher` | 전송기 | `Notifier` 구성 | nil이면 B1 창의 return |
| `n.Log` | 구조 로그 | `Notifier` 구성 | B2 조건의 `&& n.Log != nil` |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:162` `if n.Publisher == nil {` | — | :163 | 예 |
| B2 | if | `:165` `if err := n.Publisher.Publish(ctx, notificationFor(e, severity)); err != nil && n.Log != nil {` | `err.Error`, `n.Log.Warn`, `n.Publisher.Publish`, `notificationFor`, `string` | — | 예 |

## Calls and live bindings

`n.Publisher.Publish`(B2 조건 안) · `notificationFor` · `n.Log.Warn`(B2 창).

결과값이 없다 — 오류를 돌려주지 않는다. 전송 실패는 로그로만 남긴다(B2 창).

## State mutations and fallbacks

외부 전송뿐. 원장에 쓰지 않고 뮤텍스를 잡지 않는다.

## Safety conclusion

- **Safe edit boundary**: **a095는 이 함수를 바꾸지 않는다.** 2판은 B1을 「로그도 없다」로 적었으나 구조 로그는 `Notify`가 경로 분기(B1) 전에 `logEvent`로 이미 남긴다(보이스 B B-P2-13; 순서는 `SeverityOf` → `logEvent` → B1, 6판 r5 R5-4 정정). 이 경로의 부재는 「durable outbox와 재시도가 없다」이다.
- **High-risk impact**: no — 이 함수 자체는 설계대로다.
