# Function Logic Map: `checkReconcileAbsence`

- Source: `internal/verifylive/reconcile_check.go` (384-399)
- Qualified function: `checkReconcileAbsence`
- Revision: `current` (구현 base `aec9bde67`, `source_sha256` 4bf4a9feb234771e…)
- AST evidence: `ast.json` — AST branches 4, 반환 4(줄 386, 393, 396, 398), 호출 이름 3
- Risk scan: `risk-pattern-report.md`
- 이 change 에서의 역할: 부재 판정. a164 편집: OPEN 조건·일반 OPEN(옛 B1·B2)과 발동 흔적(옛 B6·B7)을 `checkReconcileNoLiveOrder`·`checkReconcileNoFiredTrace` 로 옮김(분기 7 → 4) — 두 모드가 같은 보조 함수를 부름. 부재 모드 거절 순서는 `TestReconcileAbsenceGuardsRefuseInTheirContractOrder` 가 편집 전후 같은 결과로 고정. Q6 문구 D8.

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
| B1 | if 줄 385: `if err := checkReconcileNoLiveOrder(r); err != nil {` | 없음 | `return err` | `TestReconcileAbsenceGuardsRefuseInTheirContractOrder` · `TestReconcileRefusesALiveSuccessorInTheOpenConditionalGroup` · `TestReconcileRefusesAnyOpenPlainOrderOnTheSymbol` |
| B2 | range 줄 388: `for _, row := range r.closed {` | 없음 | `if row.ID != target.ID {` | `TestReconcileAbsenceGuardsRefuseInTheirContractOrder` · `TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt` · `TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt` · `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` 외 3 |
| B3 | if 줄 389: `if row.ID != target.ID {` | 없음 | `continue` | `TestReconcileAbsenceGuardsRefuseInTheirContractOrder` · `TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt` · `TestReconcileRefusesAnOpenFamilyClosedStatusEvenIfTheAllowlistListsIt` · `TestReconcileRefusesAnotherIdentifierThatFiredInClosed` 외 1 |
| B4 | if 줄 392: `if row.Status == "EXPIRED" {` | 없음 | `return refuse(RefuseClosedTargetExpired, "the artifact is listed CLOSED as EXPIRED: "+reconcileQ6Message+` | `TestReconcileAbsenceGuardsRefuseInTheirContractOrder` · `TestReconcileRefusesTheTargetExpiredInClosedWithThePermanentMessage` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `refuse` | 거절 코드 생성 | 순수 | AST |
| `fmt.Sprintf` | 문구 | 순수 | AST |

AST 호출 이름 전체: `checkReconcileNoFiredTrace`, `checkReconcileNoLiveOrder`, `refuse`.

## State mutations and fallbacks

- 상태 변경 없음. 대상 행 루프(B3~B5)는 첫 일치에서 반환 — EXPIRED 면 `closed-target-expired`(Q6), 그 밖 `closed-target-present`.
- 발동 흔적 루프(B6·B7)는 대상 행을 빼지 않음 — a164 만료 모드에서 대상 CLOSED 행이 `COMPLETED` 면 `closed-fired-trace`(P3-5).

## Safety conclusion

- Safe edit boundary: B1·B2 를 `checkReconcileNoLiveOrder`, B6·B7 을 `checkReconcileNoFiredTrace` 로 옮기되 이 함수는 그 둘을 같은 순서로 부름. 문구 상수만 D8 로 교체.
- High-risk impact: yes — A5 는 생존 판정의 유일한 자리(design D2 위계). 보조 함수에 대상 제외 인자를 넣지 않음.

## a164 GREEN 재추출

- 편집 전(구현 base `aec9bde67`) 분기 열 길이 7 에서 편집 뒤 4 로. `source_sha256` 2546999c3f87… → 4bf4a9feb234….
- 옛/새 분기 열(kind + 원문 줄)을 difflib 으로 정렬한 결과: replace 옛 B1..B2(2) → 새 B1..B1(1) · equal 옛 B3..B5 = 새 B2..B4 · delete 옛 B6..B7(2) → 새 B5..B4(0).
- 위 표의 B 번호는 **새** AST 기준이고, 시험 칸은 GREEN 코드에서 시험별 프로필을 다시 떠 측정한 값이다(손 재번호 없음).
