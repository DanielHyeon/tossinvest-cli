# Function Logic Map: `ExitObserver.submit`

- Source: `internal/app/engine/exitloop.go` (1518-1644)
- Qualified: `ExitObserver.submit` · revision: current(편집 뒤, worktree `~/tossos-work/a207` 브랜치 `feat/a207-blocked-exit-retry-dedupe`)
- AST evidence: `ast.json` (source_sha256 0a799bac4f3b…) — AST branches 18 · returns 10
- 편집 전 증거: `analysis/ast-evidence/engine--ExitObserver.submit.ast.json`(1513-1638, 18 갈래, sha 7626cf237058)
- Risk scan: `risk-pattern-report.md` (파일 단위)

**역할.** 무장된 청산 발의 하나를 브로커에 냄. a207 은 B9(확정) 갈래 끝에 해소 대기열 넣기 호출 하나만 더함 — **분기 · return · Place 앞 작업 추가 0**(AST 분기 18 → 18, return 10 → 10).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `m` · `proposal` · `quantity` · `intentID` · `observed` · `provenance` | 무장된 발의 | record | — |
| 불변식 | 브로커 mutation 0~1 회(Place), 해제는 release 하나 | exit_proposal_release.go | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:1523` 하한 산술 실패 | 줄 | `:1527` (해제 없음) | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease` |
| B2 | `:1530` 하한 0 → 해제(REFUSED) | release | `:1537` | `TestAZeroFloorSubmitsNothingAndLeavesTheLevelProposable` |
| B3 | `:1552` Guardian 거절 → 알림 · 해제 | alertProposalRefused · release | `:1556` | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease` |
| B4 | `:1563` intent 부착 실패 | 줄 | `:1565` | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease` |
| B5 | `:1569` 주문 모양 실패 → 해제 | alertRefused · release | `:1573` | 도달 시험 없음(a197 BTM 기록) |
| B6 | `:1576` 보호면 Critical 우선순위 | — | — | 미측정 |
| B7 | `:1585` 보호면 Place 뒤 하한가 시도 표시 | — | — | `TestA087SubmitUsesOneChoiceAndMarksOnlyAfterPlace` |
| B8 | `:1590` 결과 switch | — | — | B9~B18 |
| B9 | `:1591` 확정 — **a207 호출 하나** `noteUnblocked`(`:1599`, 대기열에 넣기만) | 로그 · 줄 · 해소 대기열 | `:1600` nil | `TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle` · `TestA207AnAcceptedStopWithoutRefusalsIsNotAnUnblock` |
| B10 | `:1601` 결과 미상 → 무장 유지 | 줄 | `:1610` | `TestAnInDoubtSubmissionKeepsTheProposalArmed` |
| B11 | `:1606` 미해소 결과 미상 이벤트 | — | — | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease` |
| B12 | `:1611` 종목 진행 중 → 지연 시계 · 해제(CANCELLED) | noteDelay · release | `:1615` | `TestA129OAFailedReleaseStillSurfaces` |
| B13 | `:1616` attempt 있음 · 비수용 미입증 → 무장 유지 | 줄 | `:1624` | `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` |
| B14 | `:1620` B13 안 err 없음 | err 만듦 | — | `TestA094AnUnrecordedOutcomeKeepsTheProposalArmed` |
| B15 | `:1626` 기본(확정 거절 · 미발송) → 알림 · 해제(REFUSED) | alertProposalRefused · release | `:1642` | `TestARefusedProposalReleasesTheLevelAndAlerts` · `TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle` |
| B16 | `:1628` 상세 비고 · 오류 있음 | detail | — | 미측정 |
| B17 | `:1633` 보호 손절 ∧ 확정 거절 ∧ 반대 미체결 → a197 대기열 | noteExternalBuyBlock | — | `TestA197AnExternalBuyIsCancelledAtTheCycleEndAndTheStopFollows` |
| B18 | `:1637` 거부 vs 미발송 이벤트 | — | — | `TestA129DEveryOtherOutcomeLeavesOneLineAfterItsRelease` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.opts.Submit.Place` | 유일한 브로커 호출 | 결과 분류 B9~B18, 재시도 0 | AST `:1580` |
| `o.release` | 비수용 발의 해제(5 자리) | 오류 → wrap | AST `:1535` `:1554` `:1571` `:1613` `:1640` |
| `o.noteUnblocked` | **a207** 확정 뒤 해소 대기열(메모리) | 오류 없음 | AST `:1599` |
| `o.alertProposalRefused` · `o.alertRefused` · `o.noteDelay` · `o.noteExitOrder` · `o.noteExternalBuyBlock` | 알림 · 줄 · 대기열 | 기록 전용 | AST |

## State mutations and fallbacks

- 원장: AttachExitIntent · release · Guardian 결정(무변경). 메모리: a129 줄 · a197 대기열 · **a207 해소 대기열(B9)**.

## Safety conclusion

- Safe edit boundary: B9 의 `return nil` 앞, Place · 로그 · 줄 뒤에 호출 하나. 해제 5 자리와 그 밖의 갈래 · Place 앞은 무편집.
- High-risk impact: yes(주문 제출 경로) — 추가는 Place 가 확정으로 돌아온 뒤의 메모리 작업 하나뿐.
