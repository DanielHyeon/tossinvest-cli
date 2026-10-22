# Function Logic Map: `newVerifyReconcileCmd`

- Source: `cmd/tossctl/verify_reconcile.go` (42-70)
- Qualified function: `newVerifyReconcileCmd`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 82787649cc4147d1…)
- AST evidence: `ast.json` — AST branches 0, 반환 2(줄 63, 69), 호출 이름 4
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 명령 표면. a164 는 같은 명령에 `--expired` 플래그 하나를 더함(별도 하위 명령 금지, `mutating: "true"` 상속).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `root *rootOptions` | 전역 플래그 | cobra | — |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 단일 경로 | 없음 | func newVerifyReconcileCmd(root *rootOptions) *cobra.Command { | `TestVerifyReconcileIsRegisteredAsAMutatingOfficialCommand` · `TestVerifyReconcileOffersNoIdentifierSymbolOrBypassFlag` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cmd.Flags().StringVar` | 플래그 | — | AST |
| `runVerifyReconcile` | RunE | 오류 반환 | AST |

AST 호출 이름 전체: `StringVar`, `cmd.Flags`, `runVerifyReconcile`, `strings.TrimSpace`.

## State mutations and fallbacks

- 상태 없음.

## Safety conclusion

- Safe edit boundary: Annotations 불변, 플래그 하나 추가.
- High-risk impact: 낮음 — 표면.
