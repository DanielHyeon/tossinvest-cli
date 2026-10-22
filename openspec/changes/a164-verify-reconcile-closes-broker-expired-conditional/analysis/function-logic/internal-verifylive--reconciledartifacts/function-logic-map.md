# Function Logic Map: `ReconciledArtifacts`

- Source: `internal/verifylive/projection_reconciled.go` (8-31)
- Qualified function: `ReconciledArtifacts`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 55b48e9c512efa40…)
- AST evidence: `ast.json` — AST branches 6, 반환 1(줄 30), 호출 이름 3
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: status·report 투영 필터(`projection_reconciled.go:26`). a164 는 넷째 종결도 「대사로 종결」 로 보임(P2-5 ⑦).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `entries []Entry` | 기록 전체 | `LoadEntries` | — |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | range 줄 11: `for _, e := range entries {` | 없음 | `for _, a := range e.Artifacts {` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestProgressPointsAtTheRestartWhenOneIsPending` · `TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged` · `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` 외 4 |
| B2 | range 줄 12: `for _, a := range e.Artifacts {` | 없음 | `key := a.Kind + "\x00" + a.ID` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestProgressPointsAtTheRestartWhenOneIsPending` · `TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged` · `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` 외 2 |
| B3 | if 줄 15: `if !seen {` | 없음 | `order = append(order, key)` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestProgressPointsAtTheRestartWhenOneIsPending` · `TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged` · `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` 외 2 |
| B4 | if 줄 18: `if seen && prev.terminal() && !a.terminal() {` | 없음 | `continue` | `TestReconciledProjectionIsMonotoneLikeOutstanding` |
| B5 | range 줄 25: `for _, key := range order {` | 없음 | `if a := latest[key]; a.ReconciledAbsent {` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestProgressPointsAtTheRestartWhenOneIsPending` · `TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged` · `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` 외 2 |
| B6 | if 줄 26: `if a := latest[key]; a.ReconciledAbsent {` | 없음 | `out = append(out, a)` | `TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged` · `TestReconcileTextLabelsReconciledAbsentInReportAndStatus` · `TestReconciledProjectionIsMonotoneLikeOutstanding` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `Artifact.terminal` | 단조성 | 순수 | AST |

AST 호출 이름 전체: `a.terminal`, `append`, `prev.terminal`.

## State mutations and fallbacks

- 상태 변경 없음. 종결 뒤 비-종결 줄은 되살리지 않음(B4).

## Safety conclusion

- Safe edit boundary: B6 조건에 `|| a.ReconciledExpired` 만.
- High-risk impact: 낮음 — 표시 투영.
