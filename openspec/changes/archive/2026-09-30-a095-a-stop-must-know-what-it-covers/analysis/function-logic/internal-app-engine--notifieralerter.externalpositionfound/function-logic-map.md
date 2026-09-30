# Function Logic Map: `notifierAlerter.ExternalPositionFound`

- Source: `internal/app/engine/exitwiring.go` (`98`–`121`)
- Qualified: `notifierAlerter.ExternalPositionFound`
- AST evidence: `ast.json` (`source_sha256` 2b0696f0b0cccdf0…)
- Risk scan: `risk-pattern-report.md`
- 분기 1 · return 2 · 호출 4

**역할.** reconcile 패키지의 fold 알림을 `Notifier`로 옮긴다. 주석이 등급을 normal로 선언한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `a.notifier` | 알림기 | 배선 | nil이면 B1 창의 return |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:100` `if a.notifier == nil {` | `a.names.Label`, `a.notifier.Notify`, `string`, `strings.TrimSpace` | :101, :103 | 아니오 |

## Calls and live bindings

`a.notifier.Notify` · `a.names.Label` · `strings.TrimSpace`.

결과는 `error`다 — `Notify`의 오류를 그대로 돌려준다(nil 알림기는 B1에서 nil).

## State mutations and fallbacks

알림 1건. 본문은 `exit_eligible`과 무관하게 「손절·익절이 자동으로 걸려 있지 않다」를 쓴다.

## Safety conclusion

- **Safe edit boundary**: **10판 재추출 주석** — 분기 표 · 좌표 · 진입 실측은 현재 소스(아래 `source_sha256`)에서 기계로 다시 그렸다. 아래 산문은 3판(base `02716357`)의 판단이며, 그 뒤 a092 가 이 소스를 바꿨다(특히 `claimAndDeliver`는 이제 claim만 `n.mu` 아래에서 하고 전송은 잠금 밖 — `fbc6df5f`). a095 는 이 함수를 편집하지 않는다. 산문의 잠금 · 좌표 서술과 현재 소스가 어긋나면 표와 `review.md` §4.2 정정이 우선한다. **a095는 이 함수와 그 등급을 바꾸지 않는다.** 생산 배선에서 이 함수의 유일한 호출 자리(`Ingestor.IngestExternalPositions` B13)는 B12(`in.Alert == nil`)에 막힌다 — `ReconcileDriver`가 `d.ingest.Alert = nil`로 복사하기 때문이다(`reconcileloop.go:338`). 2판 tasks 6.2a의 「오류가 대사를 실패시킨다」는 이 자리의 등급이 바뀔 때만 성립하므로 3판에서는 성립하지 않는다.
- **High-risk impact**: no — 생산에서 도달하지 않고 3판은 등급을 바꾸지 않는다.
