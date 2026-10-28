# Function Logic Map: `newVerifyReconcileCmd`

- Source: `cmd/tossctl/verify_reconcile.go` (44-79)
- Qualified function: `newVerifyReconcileCmd`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 4008a00b6a32f2da…)
- AST evidence: `ast.json` — AST branches 0, 반환 2(줄 70, 78), 호출 이름 5
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
| B1 | 분기 없음 — 단일 경로 | 없음 | func newVerifyReconcileCmd(root *rootOptions) *cobra.Command { | `TestVerifyReconcileExpiredIsAFlagOfTheSameMutatingCommand` · `TestVerifyReconcileIsRegisteredAsAMutatingOfficialCommand` · `TestVerifyReconcileOffersNoIdentifierSymbolOrBypassFlag` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `cmd.Flags().StringVar` | 플래그 | — | AST |
| `runVerifyReconcile` | RunE | 오류 반환 | AST |

AST 호출 이름 전체: `BoolVar`, `StringVar`, `cmd.Flags`, `runVerifyReconcile`, `strings.TrimSpace`.

## State mutations and fallbacks

- 상태 없음.

## Safety conclusion

- Safe edit boundary: Annotations 불변, 플래그 하나 추가.
- High-risk impact: 낮음 — 표면.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 0 에서 편집 뒤 0 로. `source_sha256` 82787649cc41… → 4008a00b6a32….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: 분기 없음(둘 다 0).
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
