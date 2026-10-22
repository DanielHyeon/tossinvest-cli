# Function Logic Map: `checkReconcileWindow`

- Source: `internal/verifylive/reconcile_check.go` (139-149)
- Qualified function: `checkReconcileWindow`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 4211d94e33e07b9f…)
- AST evidence: `ast.json` — AST branches 2, 반환 3(줄 142, 146, 148), 호출 이름 4
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
| B1 | if 줄 141: `if elapsed < 0 \|\| elapsed > reconcileFreshnessBound {` | 없음 | `return refuse(RefuseFreshnessExceeded,` | `TestExpiryReconcileWindowStartsBeforeTheFirstByIDRead` · `TestReconcileFreshnessWindowCoversAppendAdmission` · `TestReconcileRefusesAtTheWriteBoundaryGateItself` · `TestReconcileRefusesWhenThePairExceedsTheFreshnessBound` |
| B2 | if 줄 145: `if expired {` | 없음 | `return nil` | `TestAnOlderBinaryStillSeesTheExpiredArtifactAsOutstanding` · `TestExpiryReconcileAcceptsATerminalChainSibling` · `TestExpiryReconcileAcceptsAnotherIdentifiersExpiredClosedRow` · `TestExpiryReconcileAcceptsTwoStatementsAndAnExpiredClosedRow` 외 11 |

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

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 1 에서 편집 뒤 2 로. `source_sha256` 2546999c3f87… → 4211d94e33e0….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: equal 옛 B1..B1 = 새 B1..B1 · insert 옛 B2..B1(0) → 새 B2..B2(1).
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
