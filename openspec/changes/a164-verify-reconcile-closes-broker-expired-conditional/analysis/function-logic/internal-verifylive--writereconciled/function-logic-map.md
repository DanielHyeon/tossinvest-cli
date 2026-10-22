# Function Logic Map: `writeReconciled`

- Source: `internal/verifylive/report.go` (299-307)
- Qualified function: `writeReconciled`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` e96b2103cb0c4d39…)
- AST evidence: `ast.json` — AST branches 2, 반환 1(줄 301), 호출 이름 2
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: report·status 텍스트. a164 는 artifact 마다 「reconciled absent」/「reconciled expired」 를 갈라 씀(task 2.8).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `arts []Artifact` | `ReconciledArtifacts` 결과 | report.go | — |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 300: `if len(arts) == 0 {` | 없음 | `return` | `TestProgressPointsAtTheRestartWhenOneIsPending` · `TestReportTextNamesTheUnverifiedProperties` |
| B2 | range 줄 304: `for _, a := range arts {` | 텍스트 한 줄씩 씀 | `fmt.Fprintf(w, "  %s %s (%s) reconciled absent\n", a.Kind, a.ID, a.Symbol)` | `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Fprintf` | 텍스트 | 오류 무시(기존 관례) | AST |

AST 호출 이름 전체: `fmt.Fprintf`, `len`.

## State mutations and fallbacks

- 쓰기만.

## Safety conclusion

- Safe edit boundary: 부재 artifact 의 줄 문구는 그대로.
- High-risk impact: 낮음.
