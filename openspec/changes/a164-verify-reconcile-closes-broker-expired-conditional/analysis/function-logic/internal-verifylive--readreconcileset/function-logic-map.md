# Function Logic Map: `readReconcileSet`

- Source: `internal/verifylive/reconcile_check.go` (190-203)
- Qualified function: `readReconcileSet`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 4211d94e33e07b9f…)
- AST evidence: `ast.json` — AST branches 3, 반환 4(줄 194, 197, 200, 202), 호출 이름 2
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 근거 함수(편집 없음) — 만료 모드도 이 함수를 그대로 두 번 부른다는 설계 주장(A1 · P2-5 ⑤)의 근거.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r ReconcileReader`·`symbol` | 봉인된 리더 · artifact 줄 심볼 | `Reconcile` | — |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 193: `if out.condOpen, err = readReconcileConditionalGroup(ctx, r, "OPEN", ReconcileGroupConditionalOpen, symbol); err != nil {` | 없음 | `return reconcileRead{}, err` | `TestADeleteNotFoundFromEveryListReadDoesNotReconcile` · `TestReconcileRefusesMalformedListResponsesEvenWhenBothReadsAgree` · `TestReconcileRefusesPaginationFaultsInEitherRead` |
| B2 | if 줄 196: `if out.plainOpen, err = readReconcilePlainGroup(ctx, r, symbol); err != nil {` | 없음 | `return reconcileRead{}, err` | `TestReconcileRefusesMalformedListResponsesEvenWhenBothReadsAgree` · `TestReconcileRefusesPaginationFaultsInEitherRead` |
| B3 | if 줄 199: `if out.closed, err = readReconcileConditionalGroup(ctx, r, "CLOSED", ReconcileGroupConditionalClosed, symbol); err != nil {` | 없음 | `return reconcileRead{}, err` | `TestReconcileRefusesADuplicateRowWithinOneRead` · `TestReconcileRefusesPaginationFaultsInEitherRead` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `readReconcileConditionalGroup` | OPEN · CLOSED 그룹 | 페이지 오류 = 거절 | AST |
| `readReconcilePlainGroup` | 일반 OPEN | 같음 | AST |

AST 호출 이름 전체: `readReconcileConditionalGroup`, `readReconcilePlainGroup`.

## State mutations and fallbacks

- 상태 변경 없음. 고정 순서: 조건 OPEN → 일반 OPEN → CLOSED.

## Safety conclusion

- 편집하지 않음. a164 만료 모드가 다른 읽기 집합을 만들면 변이 4.17 과 시험 2.4 의 호출 수 단언이 잡음.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 3 에서 편집 뒤 3 로. `source_sha256` 2546999c3f87… → 4211d94e33e0….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: equal 옛 B1..B3 = 새 B1..B3.
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
