# Function Logic Map: `Artifact.terminal`

- Source: `internal/verifylive/record.go` (736-736)
- Qualified function: `Artifact.terminal`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 9a289d21048578d8…)
- AST evidence: `ast.json` — AST branches 0, 반환 1(줄 736), 호출 이름 0
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 종결 술어 한 곳. a164 는 넷째 종결 `ReconciledExpired` 를 더함 — `Outstanding`·`PendingCleanup`·HOLD·노출 상한이 한 번에 따름(design D5).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 수신자 `a Artifact` | 기록 한 줄의 artifact | `record.go` `Artifact` | 순수 술어 |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 단일 경로 | 없음 | func (a Artifact) terminal() bool { return a.Cancelled \|\| a.Filled \|\| a.ReconciledAbsent } | `TestADeleteNotFoundFromEveryListReadDoesNotReconcile` · `TestADeleteNotFoundLeavesTheArtifactOutstanding` · `TestAFilledObjectIsNotOutstanding` · `TestAFullRunProducesAReportWithNoIdempotencyGaps` 외 69 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (없음) | 호출 0 | 해당 없음 | AST |

AST 호출 이름 전체: 없음.

## State mutations and fallbacks

- 상태 변경 없음. 생산 호출 자리(HEAD grep): `record.go` `outstandingLines` 두 곳 · `projection_reconciled.go` · `runner.go` `joinChain`.

## Safety conclusion

- Safe edit boundary: `|| a.ReconciledExpired` 한 항만. 새 필드를 세우는 경로는 만료 대사 하나뿐이어야 함.
- High-risk impact: yes — 이 술어가 정리 대상·노출 상한·HOLD 를 정함.
