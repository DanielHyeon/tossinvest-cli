# Function Logic Map: `restoreAlertEntryLatch`

- Source: `internal/app/engine/gateway.go`
- AST evidence: `ast.json` — **편집 전**, :153–168, 분기 2 · 반환 3 · 호출 4.
  source_sha256 `cf4833845140…`, 추출 HEAD `463cc895` (2026-09-25).
- Risk scan: `risk-pattern-report.md`

**a124 는 이 함수를 편집하지 않는다.** proposal 이 「a092 뒤에 지속 실패로 잠그는 자리는 기동 복원뿐」을
근거로 쓰기 때문에 만들었다. 호출자는 `gateway.go:269` 하나(비테스트, grep).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `j.UndeliveredCount` | ≥0 | 원장 | 오류 → B1, **기동 거부** |
| `gate` | non-nil | 호출자 :269 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 계수 오류 (:155) | 없음 | wrapped error (:156) — 기동이 거부된다 | `a098_restart_does_not_release_the_gate_test.go` |
| B2 | `undelivered <= 0` (:158) | 없음 | `nil` (:159) | 같음 |
| 종단 | 미전달 > 0 | **`gate.Block(ReasonAlertUndelivered, "<n> … not been delivered")`** :164 | `nil` (:167) | 같음 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.UndeliveredCount(ctx)` :154 | 재시작을 넘길 미전달 수 | 오류 → 기동 거부 | AST |
| `gate.Block` :164 | 메모리 래치 복원 | — | AST |

## State mutations and fallbacks

- 기동 시 **한 번** 돈다. 살아 있는 프로세스에서 미전달이 쌓여도 다시 돌지 않는다 — 그래서 a092 뒤에 이것만
  남으면 「지속 실패 → 차단」은 재시작해야 성립한다. a124 R1 이 그 사이를 배달 실행자로 메운다.
- 운영 모드는 여기서 건드리지 않는다(승격은 원장에 남아 있으므로 복원할 것이 없다).
