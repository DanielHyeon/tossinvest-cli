# Function Logic Map: `checkReconcileWindow`

- Source: `internal/verifylive/reconcile_check.go` (139-154)
- Qualified function: `checkReconcileWindow`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 4bf4a9feb234771e…)
- AST evidence: `ast.json` — AST branches 3, 반환 3(줄 147, 151, 153), 호출 이름 4
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: Q3 창 + Q1 나이 재검. a164 편집: 모드 인자 `expired` — 참이면 Q1 재검을 건너뜀(B2, design D7).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `start`·`now` | 창 시작(a164 만료 모드는 by-id ① 직전)과 지금 | `Reconcile` | 음수·초과 = `freshness-exceeded` |
| `target.CreatedAt` | Q1 나이 | 기록 줄 | `checkReconcileRetention` 이 판정 |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 141: `if elapsed < 0 \|\| elapsed > reconcileFreshnessBound {` | 없음 | `// 부재 모드 문구는 a121 바이트 그대로(design D1 — D8 예외만), 만료 모드는 창 시작이 by-id ① 직전임을 말함.` | `TestExpiryReconcileWindowStartsBeforeTheFirstByIDRead` · `TestReconcileAbsenceKeepsItsA121Wording` · `TestReconcileFreshnessWindowCoversAppendAdmission` · `TestReconcileRefusesAtTheWriteBoundaryGateItself` 외 1 |
| B2 | if 줄 144: `if expired {` | 없음 | `since = "the first by-id read"` | `TestExpiryReconcileWindowStartsBeforeTheFirstByIDRead` |
| B3 | if 줄 150: `if expired {` | 없음 | `return nil` | `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` · `TestExpiryReconcileAcceptsATerminalChainSibling` · `TestExpiryReconcileAcceptsAnotherIdentifiersExpiredClosedRow` · `TestExpiryReconcileAcceptsTwoStatementsAndAnExpiredClosedRow` 외 11 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `checkReconcileRetention` | Q1 나이 재검 | 거절 | AST — a164: 만료 모드는 부르지 않음 |

AST 호출 이름 전체: `checkReconcileRetention`, `fmt.Sprintf`, `now.Sub`, `refuse`.

## State mutations and fallbacks

- 상태 변경 없음.

## Safety conclusion

- Safe edit boundary: 부재 모드(인자 거짓)는 지금과 같은 두 판정을 같은 순서로.
- High-risk impact: 인접 — Q1 을 만료 모드에서 다시 부르면 수락 경로가 죽음(변이 4.7), 부재 모드에서 건너뛰면 부재 대사가 열림(변이 4.7 둘째).

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 1 에서 편집 뒤 3 로. `source_sha256` 2546999c3f87… → 4bf4a9feb234….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: equal 옛 B1..B1 = 새 B1..B1 · insert 옛 B2..B1(0) → 새 B2..B3(2).
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
