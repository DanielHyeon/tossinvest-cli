# Function Logic Map: `reconcileReads.ensureIdentity`

- Source: `cmd/tossctl/verify_reconcile.go` (226-240)
- Qualified function: `reconcileReads.ensureIdentity`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 4008a00b6a32f2da…)
- AST evidence: `ast.json` — AST branches 3, 반환 4(줄 228, 232, 236, 239), 호출 이름 5
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 근거 함수(편집 없음) — 새 by-id 메서드가 기존 두 계좌 범위 메서드처럼 이 함수 뒤에만 위임한다는 설계 주장(P2-4)의 근거.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r.verified` | 첫 계좌 범위 읽기 전 거짓 | 생성자 | 한 번 확인 뒤 참 |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 227: `if r.verified {` | 없음 | `return nil` | `TestExpiryReconcileCommandAppendsWithOnlyGetReadsAndLeavesProofsUnchanged` · `TestReconcileNarrowReaderBindsTheValidatedSequence` · `TestReconcileSucceedsWithOnlyGetRequestsAndTheTokenPost` |
| B2 | if 줄 231: `if err != nil {` | 없음 | `return fmt.Errorf("re-reading the account list before the reads: %w", err)` | 미덮임(시험별 coverage 측정 0) — 재조회 자체 실패 — 기존 공백 |
| B3 | if 줄 234: `if len(accounts) != 1 \|\| strings.TrimSpace(accounts[0].DisplayName) != r.ref \|\|` | 신원 불일치 → 읽지 않음 | `return errReconcileIdentityChanged` | `TestReconcileByIDReadRechecksIdentityBeforeItIsSent` · `TestReconcileReaderRechecksIdentityBeforeItsFirstScopedRead` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `r.client.Accounts` | 신원 재확인 | 오류 반환 | AST |

AST 호출 이름 전체: `fmt.Errorf`, `len`, `r.client.Accounts`, `strconv.Itoa`, `strings.TrimSpace`.

## State mutations and fallbacks

- `r.verified = true` 한 번.

## Safety conclusion

- 편집하지 않음. 새 메서드가 이 호출을 빼면 변이 4.20 둘째가 잡아야 함.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 3 에서 편집 뒤 3 로. `source_sha256` 82787649cc41… → 4008a00b6a32….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: equal 옛 B1..B3 = 새 B1..B3.
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
