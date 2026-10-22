# Function Logic Map: `ReconcileApproval.WriteText`

- Source: `internal/verifylive/reconcile_check.go` (436-452)
- Qualified function: `ReconcileApproval.WriteText`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 4211d94e33e07b9f…)
- AST evidence: `ast.json` — AST branches 2, 반환 1(줄 451), 호출 이름 4
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 승인 화면. a164 편집: 만료 모드 갈래(B1 — 모드·상세 조회 2회·근거 규칙), 부재 모드(B2 else)는 a121 문구 그대로.

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
| B1 | if 줄 438: `if a.Mode == ReconcileModeExpired {` | 만료 모드 머리 두 줄 | `// a164 P3-6 — 만료 모드는 by-id 상세 조회(계좌 범위 GET) 두 번을 더 하므로 화면이 그것과 모드를 말함.` | `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` · `TestExpiryReconcileAcceptsATerminalChainSibling` · `TestExpiryReconcileAcceptsAnotherIdentifiersExpiredClosedRow` · `TestExpiryReconcileAcceptsTwoStatementsAndAnExpiredClosedRow` 외 13 |
| B2 | else 줄 442: `} else {` | a121 머리 한 줄 | `fmt.Fprintf(&b, "대사 대상 (공식 목록 조회 전용 · 기록에 사건 한 줄 추가 · 주문 변경 없음)\n")` | `TestADeleteNotFoundFromEveryListReadDoesNotReconcile` · `TestExpiryReconcileApprovalNamesTheModeAndTheByIDReads` · `TestReconcileAbsenceGuardsRefuseInTheirContractOrder` · `TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt` 외 46 |

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

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 0 에서 편집 뒤 2 로. `source_sha256` 2546999c3f87… → 4211d94e33e0….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: insert 옛 B1..B0(0) → 새 B1..B2(2).
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
