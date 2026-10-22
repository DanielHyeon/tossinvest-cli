# Function Logic Map: `ReconcileApproval.WriteText`

- Source: `internal/verifylive/reconcile_check.go` (413-423)
- Qualified function: `ReconcileApproval.WriteText`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 2546999c3f878b39…)
- AST evidence: `ast.json` — AST branches 0, 반환 1(줄 422), 호출 이름 4
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 승인 화면. a164 는 만료 모드에서 머리 문구와 「상세 조회 2회」 를 보이는 갈래를 더함(P3-6). 부재 모드 출력은 바이트 단위로 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 수신자 `a ReconcileApproval` | 스칼라 필드만(봉인 2층) | `Reconcile` | — |
| `w io.Writer` | 단말 | cmd | 쓰기 오류를 돌려줌 → `approval-display` |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 분기 없음 — 단일 경로 | 운영자 화면에 승인 내용 씀 | func (a ReconcileApproval) WriteText(w io.Writer) error { | `TestADeleteNotFoundFromEveryListReadDoesNotReconcile` · `TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt` · `TestReconcileAcceptsExactlyThePageCap` · `TestReconcileAcceptsWellFormedEmptyListsOverHTTP` 외 43 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `fmt.Fprintf`·`io.WriteString` | 화면 문구 | 쓰기 오류 반환 | AST |

AST 호출 이름 전체: `b.String`, `fmt.Fprintf`, `fmt.Fprintln`, `io.WriteString`.

## State mutations and fallbacks

- 쓰기 하나가 부작용. 원문 계좌 없음(마스크만).

## Safety conclusion

- Safe edit boundary: 모드 필드가 비면(부재) 지금 문구 그대로.
- High-risk impact: 낮음 — 표시만, 판정 없음.
