# Function Logic Map: `ReconciledArtifacts`

- Source: `internal/verifylive/projection_reconciled.go` (9-32)
- Qualified function: `ReconciledArtifacts`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` e7ff440063e72b55…)
- AST evidence: `ast.json` — AST branches 6, 반환 1(줄 31), 호출 이름 3
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
| B1 | range 줄 12: `for _, e := range entries {` | 없음 | `for _, a := range e.Artifacts {` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestExpiryReconcileEndsTheArtifactForEveryOutstandingReader` · `TestExpiryReconcileTextLabelsReconciledExpired` · `TestProgressPointsAtTheRestartWhenOneIsPending` 외 7 |
| B2 | range 줄 13: `for _, a := range e.Artifacts {` | 없음 | `key := a.Kind + "\x00" + a.ID` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestExpiryReconcileEndsTheArtifactForEveryOutstandingReader` · `TestExpiryReconcileTextLabelsReconciledExpired` · `TestProgressPointsAtTheRestartWhenOneIsPending` 외 5 |
| B3 | if 줄 16: `if !seen {` | 없음 | `order = append(order, key)` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestExpiryReconcileEndsTheArtifactForEveryOutstandingReader` · `TestExpiryReconcileTextLabelsReconciledExpired` · `TestProgressPointsAtTheRestartWhenOneIsPending` 외 5 |
| B4 | if 줄 19: `if seen && prev.terminal() && !a.terminal() {` | 없음 | `continue` | `TestReconciledProjectionIsMonotoneLikeOutstanding` |
| B5 | range 줄 26: `for _, key := range order {` | 없음 | `if a := latest[key]; a.ReconciledAbsent \|\| a.ReconciledExpired {` | `TestAFullRunProducesAReportWithNoIdempotencyGaps` · `TestExpiryReconcileEndsTheArtifactForEveryOutstandingReader` · `TestExpiryReconcileTextLabelsReconciledExpired` · `TestProgressPointsAtTheRestartWhenOneIsPending` 외 5 |
| B6 | if 줄 27: `if a := latest[key]; a.ReconciledAbsent \|\| a.ReconciledExpired {` | 없음 | `out = append(out, a)` | `TestExpiryReconcileEndsTheArtifactForEveryOutstandingReader` · `TestExpiryReconcileTextLabelsReconciledExpired` · `TestReconcileAbsenceKeepsItsA121Wording` · `TestReconcileLeavesTheReportAttributesAndVerdictsUnchanged` 외 2 |

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

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 6 에서 편집 뒤 6 로. `source_sha256` 55b48e9c512e… → e7ff440063e7….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: equal 옛 B1..B5 = 새 B1..B5 · replace 옛 B6..B6(1) → 새 B6..B6(1).
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
