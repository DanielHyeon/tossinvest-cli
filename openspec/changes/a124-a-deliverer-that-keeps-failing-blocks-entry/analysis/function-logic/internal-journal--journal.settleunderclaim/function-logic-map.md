# Function Logic Map: `Journal.settleUnderClaim`

- Source: `internal/journal/alert_claim.go`
- AST evidence: `ast.json` — **편집 전**(base 논리), :310–345, 분기 8 · 반환 9 · 호출 15.
  source_sha256 `3d0b70b889d8…`, 추출 base `4798d399` (2026-09-26, Teammate — 재freeze 로트).
- Risk scan: `risk-pattern-report.md`

**이 번들은 design 2판 D1 이 이 함수의 분기(B5 적용 경로)를 근거로 쓰기 때문에 문서보다 먼저 만들었다.**
a124 는 이 함수를 편집한다(B5 경로에 같은 트랜잭션 안의 `attempts` 읽기) — 편집 뒤 `revision: current` 재추출(tasks 1.3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `token` | 비어 있지 않음 | 호출자의 임차 | B1 → 오류 |
| `stmt` | id · PENDING · token CAS 로 끝나는 UPDATE | 호출자 3: `MarkAlertDelivered` :449 · `MarkAlertAttemptFailed` :467 · `ReleaseAlertClaim` | — |
| 원장 연결 | 하나 (`journal.go:174`) | — | BeginTx 대기 = busy timeout |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 토큰 공백 (:313) | 없음 | 오류 (:314) | a099 토큰 없는 정산 거절 |
| B2 | `BeginTx` 오류 (:317) | 없음 | 오류 (:318) | 1.4 |
| B3 | `ExecContext` 오류 (:323) | 롤백 | 오류 (:324) — **deliverOne B10 의 원천 하나** | 1.4 / a124 RED 2.10 |
| B4 | `RowsAffected` 오류 (:327) | 롤백 | 오류 (:328) | 1.4 |
| B5 | `n == 1` (:330) — CAS 적중 | 커밋 | `SettleResult{Outcome: SettleApplied}` (:334) — **attempts 없음** | a099 |
| B6 | 커밋 오류 (:331) | 롤백 | 오류 (:332) | 1.4 |
| B7 | `explainSettleTx` 오류 (:338) | 롤백 | 오류 (:339) | 1.4 |
| B8 | 설명 뒤 커밋 오류 (:341) | 롤백 | 오류 (:342) | 1.4 |
| 종단 | CAS 빗나감 | 읽기만 | `LeaseLost` · `AlreadySettled` · `NotFound` (:344, `explainSettleTx` :349-366) | a099 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `j.db.BeginTx` :316 | 정산 트랜잭션 | 오류 → B2 | AST |
| `tx.ExecContext` :322 | CAS UPDATE | 오류 → B3 | AST |
| `tx.Commit` :331 · :341 | 확정 | 오류 → B6 · B8 | AST |
| `explainSettleTx` :337 | 0 행의 사유 | 오류 → B7 | AST |

## State mutations and fallbacks

- 오류 반환은 모두 `SettleResult{}` 이고 **그 `Outcome` 영값은 `SettleApplied`** 다(`alert_claim.go:108`). 호출자는 `err` 를
  먼저 봐야 한다(review F10).
- B5 는 갱신된 행의 어떤 열도 돌려주지 않는다. `MarkAlertAttemptFailed` 의 `attempts = attempts + 1` (outbox.go:471) 결과를
  호출자가 알 방법이 HEAD 에 없다 — review F1, design 2판 D1 이 이 경로에 같은 트랜잭션 읽기를 더한다.

## Safety conclusion

- Safe edit boundary: B5 경로의 커밋 **전**에 `SELECT attempts … WHERE id = ?` 하나를 더하고 새 필드에 담는다. 다른 분기·CAS·
  세 호출자의 SQL 은 그대로. 읽기 실패는 롤백 후 오류(정산하지 않은 것으로 — B3 과 같은 결과).
- High-risk impact: yes — 원장 outbox 정산. 필드는 additive, 스키마 무변경.
