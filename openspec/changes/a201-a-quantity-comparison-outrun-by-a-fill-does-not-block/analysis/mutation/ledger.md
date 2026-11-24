# a201 변이 원장 (tasks 2.5 · design §3-15)

- 사본: 연결 worktree `/home/daniel/tossos-work/a201-mut-677937`(detached — 시작 `599f957a9`, M14 · M15 재실행은 행동 고정 시험을 커밋한 `77c690c22` · `ea6b2011b`). 한 번에 한 판, 판마다 `git checkout -- <파일>` 원복 후 `git diff --quiet` ∧ `hash-object == HEAD:<파일>` 확인(전부 `reverted-ok`).
- 하네스: `run_mutant.sh`(세 패키지 `-run 'TestA201'`, `-covermode=set -coverpkg` 세 패키지). 원문 `raw/<이름>.txt`(머리에 변이 diff), 커버 프로파일은 크기 때문에 세션 스크래치에만 둠.
- **CAUGHT** = 실패한 시험이 하나 이상. **reached** = 변이 줄이 커버 프로파일에서 실행됨(삭제 변이는 삭제 자리의 다음 줄). 둘을 따로 적음.
- 무변이 대조군: `M00-control`(599f957a9) 33 PASS · FAIL 0 · reached yes, `M00b-control-77c690c22` 34 PASS · FAIL 0 — 대조군이 GREEN 이라 판정을 씀.

| # | 변이(design §3-15) | 자리 | reached | CAUGHT | 잡은 시험 |
|---|---|---|---|---|---|
| M01 | V2 (a) 삭제 — 수량 불일치가 아니라 분쟁 종목 전부(`symbolsInDispute`) | a201_fill_credit.go:56 | yes | yes(1) | `TestA201FillOutranSymbolsNamesOnlyMovedUnconvergedQuantitySymbols` |
| M02 | V2 (b) 삭제 — 워터마크 비교 없음(`>= 0`) | :65 | yes | yes(2) | 위 + `TestA201TheTrailingSymbolWhoseFillLandsAfterObserveStaysStuck` |
| M03 | V2 (c) 삭제 — 수렴한 종목도 부여 | :58 | yes | yes(1) | `TestA201FillOutranSymbolsNamesOnlyMovedUnconvergedQuantitySymbols` |
| M04 | V2 (d) 삭제 — 블록 없는 종목도 부여 | :89 | yes | yes(1) | `TestA201AFillCreditNeedsAStandingBlock` |
| M05 | 표식을 로컬 읽기 뒤(수렴 뒤)에 읽음 | reconcileloop.go:458 | yes | yes(8) | W2 · W3-a · tx 직전 · 앞 크레딧 위 W2 · 상쇄 · 다른 종목 분쟁 · 결속 행 · 구조 시험(W3-b 는 체결이 Refresh 라 이 변이로는 안 갈림) |
| M06 | 종목별이 아닌 전역 워터마크로 판정 | a201_fill_credit.go:61 | yes | yes(2) | `TestA201FillOutranSymbolsNamesOnlyMovedUnconvergedQuantitySymbols` · `TestA201TheTrailingSymbolWhoseFillLandsAfterObserveStaysStuck` |
| M07a | 판독 실패 시 부여 — 「모름」 표식을 0 으로 씀 | :48 | yes | yes(1) | `TestA201AnUnreadableJournalGrantsNoFillCredit` |
| M07b | 판독 실패 시 부여 — 종목 워터마크 오류면 부여 | :63 | yes | yes(1) | `TestA201AnUnreadableJournalGrantsNoFillCredit` |
| M08 | Observe 앞에서 부여 | reconcileloop.go:466 | yes | yes(6) | W2 · W3-a · tx 직전 · W3-b · 다른 종목 분쟁 · 구조 시험 |
| M09 | Observe 오류 주기에도 부여(오류 반환 앞) | reconcileloop.go:472 | yes | yes(2) | `TestA201AnObserveErrorCycleGrantsNoFillCredit` · 구조 시험 |
| M10 | 원인 표지 반대로(체결 credit 해제 → ADJUSTMENT_APPLIED) | mismatch.go:950 | yes | yes(10) | 엔진 해제 시험 전부 · `TestA201AFillCreditIsSpentOnlyByAStrictlyLaterCleanComparison` 외 |
| M11 | 원인 등식을 「filled 에 있음」으로(값 비교 삭제) | mismatch.go:712 | yes | yes(2) | `TestA201ALaterAdjustmentNamesTheRelease` · `TestA201ALeftoverFillMarkDoesNotRenameALaterAdjustmentRelease` |
| M12 | `ValidReconcileReleaseCause` 에서 새 원인 삭제 | reconcile_states.go:114 | yes | yes(14) | `TestA201FillAppliedIsAReleaseCauseThisBuildWrites` · `TestA201TheFillReleaseCauseIsFillApplied` · 엔진 해제 시험(해제 쓰기 오류) |
| M13 | V4b 반박 호출 삭제 | mismatch.go:668 | yes | yes(3) | V4b 조정 판 · 체결 판 · 구조 시험 |
| M14 | V4b 를 오류 갈래 밖(권위 확인 앞, 항상)으로 옮김 | mismatch.go:665 | yes | yes(1 → 2) | 599f957a9: 구조 시험만(행동 시험 0 — 선택 reconcile 173 도 PASS, `raw/M14-…reconcile-selected.txt`). 그래서 `77c690c22` 에 행동 고정 `TestA201WithoutAnAuthorityErrorADisputedCreditStillTakesTheAwaitingBranch` 를 더하고 다시 돌림 → 구조 + 행동 2 |
| M15(목록 밖 추가) | 부여 메서드의 곁 맵 정리(prune) 삭제(`&& false`) | a201_fill_credit.go:82 | yes — 조건 줄 82 실행(첫 판은 하네스 인자를 몸 줄 83 으로 줘 `reached=no` 로 찍힘 — 변이로 죽은 줄; `raw/M15b-…` 는 82 로 잼) | 첫 판(77c690c22) **no — SURVIVED**(34 PASS, `raw/M15-no-filled-prune-77c690c22.txt`) → `ea6b2011b` 에 행동 고정 추가 뒤 yes(1) | `TestA201AGrantPrunesALeftoverFillMarkAfterARestart` |

- 대조군 `M00c-control-ea6b2011b`: 35 PASS · FAIL 0.

## 요약

- 변이 16(design §3-15 의 15 — M07 두 판 포함 — + 목록 밖 M15): 최종 reached 16/16, CAUGHT 16/16, SURVIVED 0.
- 처음 판에서 행동 시험이 못 잡던 것 2:
  - M14 — 구조 시험만 잡음 → 행동 고정 `TestA201WithoutAnAuthorityErrorADisputedCreditStillTakesTheAwaitingBranch`(77c690c22) 추가 뒤 둘 다 잡음.
  - M15 — SURVIVED(곁 맵 정리는 V5b 등식이 원인 판정을 막고 있어 보통 입력에서는 드러나지 않음; 드러나는 것은 design §4 의 「재기동 뒤 같은 글자 as-of」 모호함을 좁히는 효과뿐) → 그 입력의 행동 고정 시험(ea6b2011b) 추가 뒤 CAUGHT.

## 구현 리뷰 r1 처분 뒤 추가(사본 `04d6a003f`, 대조군 `M00d-control-04d6a003f` FAIL 0)

| # | 변이 | 자리 | reached | CAUGHT | 잡은 시험 |
|---|---|---|---|---|---|
| M16(R3 — 재지 않음) | `FillOutranSymbols` 의 종목 워터마크 오류 시 `return nil` → `return out`(앞 종목까지는 줌) | a201_fill_credit.go:63 | — | **SURVIVED(수용)** | 없음 — 둘 다 design V2 안이고 안전 방향 차이 없음(뒤 종목 판독 실패 때 앞 종목에 주느냐뿐). `*journal.Journal` 구체 타입이라 종목별 실패를 만들려면 시험용 seam(새 추상화)이 필요 — 넣지 않음(Manager 결정). 지금 시험의 판독 실패는 모든 종목이 함께 실패하는 닫힌 장부뿐 |
| M17(R2) | Observe 끝 루프 B29 의 `disputed[symbol] \|\|` 삭제(분쟁 종목 credit 을 안 지움) | mismatch.go:853 | yes | yes(2) | `TestA201ALaterDisputeRefutesTheFillCredit`(R2 로 뒤 주기를 계좌 = 로컬로 바꾼 판) · `TestA201WithoutTheAuthorityErrorTheSameInputIsUnchanged` |
| M18(R4) | 표식 읽기를 `stabilise` 앞으로 옮김 | reconcileloop.go:428 | yes | yes(1) | `TestA201TheFillCreditIsGrantedAtOneSiteAfterASuccessfulObserve`(구조 — 「표식 바로 다음 문장이 로컬 읽기」) |
| M19(R5 확인) | Observe B7(QM ∧ adjusted_reconcile 아님 → continue) 무력화 | mismatch.go:676 | yes | yes(2) | `TestA201AFillCreditDoesNotReleaseAnOperatorOnlyBlock`(R5 로 추가) · `TestA201AConvergedSymbolGetsNoFillCredit`(옛 인용도 B7 에 기대고 있었음 — 다만 쓸 수 있는 체결 credit 이 있는 입력은 아님) |

- r1 뒤 합계: 잰 변이 19(M01~M15 · M07 두 판 · M17~M19) 모두 reached · CAUGHT, 재지 않고 수용한 SURVIVED 1(M16).
