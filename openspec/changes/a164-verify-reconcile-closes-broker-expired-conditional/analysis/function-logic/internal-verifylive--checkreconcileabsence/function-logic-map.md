# Function Logic Map: `checkReconcileAbsence`

- Source: `internal/verifylive/reconcile_check.go` (371-396)
- Qualified function: `checkReconcileAbsence`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 2546999c3f878b39…)
- AST evidence: `ast.json` — AST branches 7, 반환 6(줄 373, 377, 385, 388, 392, 395), 호출 이름 3
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 부재 판정. a164 는 A5 세 판정(OPEN 조건 · OPEN 일반 · 발동 흔적)을 보조 함수 둘로 뽑아 두 모드가 같이 부르게 한다(design D4). 부재 모드 거절 순서(OPEN 조건 → 일반 OPEN → CLOSED target → 발동 흔적)는 불변. Q6 문구만 D8 로 바꿈.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `r reconcileRead` | 첫 읽기 스냅숏(둘째와 multiset 동일 확인 뒤) | `readReconcileSet` | — |
| `target Artifact` | 기록이 고른 후보 | `selectReconcileCandidate` | — |

## Branches and early returns

AST 열거 순서 그대로다. 「Required test」 칸은 손으로 고르지 않았다 — 시험 함수 하나씩 `-coverprofile` 을 떠서 분기 본문
블록이 실행된 시험만 적었다(측정 도구 `analysis/measure/` — `pertest.sh` · `btm_cov.py` · `gen_flm.py`).

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if 줄 372: `if len(r.condOpen) > 0 {` | 없음 | `return refuse(RefuseOpenConditional,` | `TestReconcileRefusesALiveSuccessorInTheOpenConditionalGroup` |
| B2 | if 줄 376: `if len(r.plainOpen) > 0 {` | 없음 | `return refuse(RefuseOpenPlainOrder,` | `TestReconcileRefusesAnyOpenPlainOrderOnTheSymbol` |
| B3 | range 줄 380: `for _, row := range r.closed {` | 없음 | `if row.ID != target.ID {` | `TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt` · `TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt` · `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` · `TestReconcileRefusesClosedStatusOutsideTheAllowlist` 외 2 |
| B4 | if 줄 381: `if row.ID != target.ID {` | 없음 | `continue` | `TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt` · `TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt` · `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` · `TestReconcileRefusesClosedStatusOutsideTheAllowlist` |
| B5 | if 줄 384: `if row.Status == "EXPIRED" {` | 없음 | `return refuse(RefuseClosedTargetExpired, "the artifact is listed CLOSED as EXPIRED: "+reconcileQ6Message+` | `TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage` |
| B6 | range 줄 390: `for _, row := range r.closed {` | 없음 | `if row.TriggeredOrderID != "" \|\| row.Status == "COMPLETED" {` | `TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt` · `TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt` · `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` · `TestReconcileRefusesClosedStatusOutsideTheAllowlist` |
| B7 | if 줄 391: `if row.TriggeredOrderID != "" \|\| row.Status == "COMPLETED" {` | 없음 | `return refuse(RefuseClosedFiredTrace, "the symbol's CLOSED group holds a fired conditional — a successor may have fired"` | `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `refuse` | 거절 코드 생성 | 순수 | AST |
| `fmt.Sprintf` | 문구 | 순수 | AST |

AST 호출 이름 전체: `fmt.Sprintf`, `len`, `refuse`.

## State mutations and fallbacks

- 상태 변경 없음. 대상 행 루프(B3~B5)는 첫 일치에서 반환 — EXPIRED 면 `closed-target-expired`(Q6), 그 밖 `closed-target-present`.
- 발동 흔적 루프(B6·B7)는 대상 행을 빼지 않음 — a164 만료 모드에서 대상 CLOSED 행이 `COMPLETED` 면 `closed-fired-trace`(P3-5).

## Safety conclusion

- Safe edit boundary: B1·B2 를 `checkReconcileNoLiveOrder`, B6·B7 을 `checkReconcileNoFiredTrace` 로 옮기되 이 함수는 그 둘을 같은 순서로 부름. 문구 상수만 D8 로 교체.
- High-risk impact: yes — A5 는 생존 판정의 유일한 자리(design D2 위계). 보조 함수에 대상 제외 인자를 넣지 않음.
