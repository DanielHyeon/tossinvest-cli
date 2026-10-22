# Function Logic Map: `checkReconcileWindow`

- Source: `internal/verifylive/reconcile_check.go` (136-143)
- Qualified function: `checkReconcileWindow`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 2546999c3f878b39…)
- AST evidence: `ast.json` — AST branches 1, 반환 2(줄 139, 142), 호출 이름 4
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: Q3 창 + Q1 나이 재검. a164 는 만료 모드에서 Q1 재검을 건너뛰는 모드 인자를 더함(design D7 — 만료 모드는 Q1 을 부르지 않음).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `start`·`now` | 창 시작(a164 만료 모드는 by-id ① 직전)과 지금 | `Reconcile` | 음수·초과 = `freshness-exceeded` |
| `target.CreatedAt` | Q1 나이 | 기록 줄 | `checkReconcileRetention` 이 판정 |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 138: `if elapsed < 0 \|\| elapsed > reconcileFreshnessBound {` | 없음 | `return refuse(RefuseFreshnessExceeded,` | `TestReconcileFreshnessWindowCoversAppendAdmission` · `TestReconcileRefusesWhenThePairExceedsTheFreshnessBound` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `checkReconcileRetention` | Q1 나이 재검 | 거절 | AST — a164: 만료 모드는 부르지 않음 |

AST 호출 이름 전체: `checkReconcileRetention`, `fmt.Sprintf`, `now.Sub`, `refuse`.

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- Safe edit boundary: 부재 모드(인자 거짓)는 지금과 같은 두 판정을 같은 순서로.
- High-risk impact: 인접 — Q1 을 만료 모드에서 다시 부르면 수락 경로가 죽음(변이 4.7), 부재 모드에서 건너뛰면 부재 대사가 열림(변이 4.7 둘째).
