# Function Logic Map: `ReconcileDriver.alertUnmanaged`

- Source: `internal/app/engine/adoption.go` (`439`–`469`)
- Qualified: `ReconcileDriver.alertUnmanaged`
- AST evidence: `ast.json` (`source_sha256` 26a0601d9987c7dc…)
- Risk scan: `risk-pattern-report.md`
- 분기 3 · return 2 · 호출 5

**역할.** 엔진이 관리하지 않는 보유를 알린다. 사실(조건 칸)을 `unmanagedFact`로 고르고, 알림 켜짐 ∧ 편입 켜짐 — 시도 실패만 critical(`alertAdoptionFailed`)로, 그 밖은 normal로 보낸다(10판).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `d.opts.NotificationsEnabled` | 로드된 알림 켜짐 | 생산 조립 `Context.ReconcileDriver` | 거짓이면 어떤 사실도 critical 아님 |
| `result` | 후보의 편입 결과 | `adopt` | 시도 실패 / 연기를 가른다 |
| `d.unmanaged[p.ID][fact]` | 이 사실을 이미 알렸나 | 프로세스 메모리 — (포지션, 조건) | normal 만 억제, critical 은 거치지 않음 |

## Branches and early returns

> **표의 유래.** 조건은 소스의 그 줄 원문이다. 「창의 호출/return」은 `ast.json`이 기록한 좌표를 `[분기 줄, 다음 분기 줄)` 창에 넣은 것이며 **분기의 의미가 아니라 위치**다. 「진입 실측」은 `go test ./internal/<pkg>/ -count=1 -covermode=set -coverprofile=…` 를 obs · journal · app/engine · reconcile 네 패키지에 각각 돌린 프로파일에서 **그 줄로 시작하는 블록**의 count가 0보다 큰지다 — 자체 블록이 없는 분기는 `—`다. 생성: `analysis/harness/render_bundles.py`.

| Branch | 종류 | 조건 (원문) | 창의 호출 (AST) | 창의 return | 진입 실측 |
|---|---|---|---|---|---|
| B1 | if | `:441` `if d.opts.NotificationsEnabled && fact == factEnabledFailed {` | `d.alertAdoptionFailed` | :443 | 예 |
| B2 | if | `:445` `if d.unmanaged[p.ID][fact] {` | — | :446 | 예 |
| B3 | if | `:448` `if d.unmanaged[p.ID] == nil {` | `d.alert`, `d.label`, `reconcileUnmanagedKey` | — | 예 |

## Calls and live bindings

`d.unmanagedFact` · `d.alertAdoptionFailed`(critical 갈래) · `reconcileUnmanagedKey` · `d.alert` · `d.label`.

결과값이 없다 — 오류를 돌려주지 않는다. `d.alert`(→ `ReconcileDriver.alert` B2)가 `Notify`의 오류를 로그로만 남긴다.

## State mutations and fallbacks

normal 래치 `d.unmanaged[p.ID][fact] = true`(메모리) · 알림 1건. key 는 `<종류>|reconcile|<조건>|<posID>` — exit 관측 자리와 다르다.

## Safety conclusion

- **Safe edit boundary**: **10판 편집**: 옛 B1 포지션 래치 → (포지션, 조건) 래치이고 critical 은 래치 앞에서 갈라져 거치지 않는다(6판 원칙). why-matrix 의 사유 switch 는 `unmanagedFact`(새 leaf)로 옮겨 조건 칸을 함께 돌려준다 — 순서(설정 거부 → 제외 → 편입 켜짐 → include 지정 → 기본)는 그대로다. 사유 문구는 연기와 시도 실패로 갈렸다(편입 켜짐 · include 지정 각각).
- **High-risk impact**: yes — reconcile 쪽 무보호 보고와 진입 차단 도달의 자리다.
