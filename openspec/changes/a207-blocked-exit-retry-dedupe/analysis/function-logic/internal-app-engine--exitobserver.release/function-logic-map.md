# Function Logic Map: `ExitObserver.release`

- Source: `internal/app/engine/exitloop.go` (1651-1661)
- Qualified: `ExitObserver.release` · revision: current(편집 뒤, worktree `~/tossos-work/a207` 브랜치 `feat/a207-blocked-exit-retry-dedupe`)
- AST evidence: `ast.json` (source_sha256 0a799bac4f3b…) — AST branches 2 · returns 2
- 편집 전 증거: `analysis/ast-evidence/engine--ExitObserver.release.ast.json`(1645-1650, 1 갈래, sha 7626cf237058)
- Risk scan: `risk-pattern-report.md` (파일 단위)

**역할.** 비수용 발의를 원장 판정 하나로 해제함(submit 의 해제 5 곳 공용). a207 은 원장이 실제로 풀었을 때(released) 메모리 셈을 더함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `intentID` · `how` | 무장된 intent, REFUSED/CANCELLED | submit | — |
| 불변식 | 해제 판정은 원장(ReleaseUnacceptedExitProposal) 하나 | a094 D−2.5 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:1653` 원장 해제 오류 | — | `:1654` wrap | `TestA129OAFailedReleaseStillSurfaces` |
| B2 | **a207** `:1656` released=true | `noteRefusalStreak`(메모리) | — | `TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle`(참) · `TestA207AReleaseTheLedgerRefusesIsNotCounted`(거짓) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.opts.Journal.ReleaseUnacceptedExitProposal` | 해제 판정 · 쓰기 | 오류 → B1 | AST `:1652` |
| `o.noteRefusalStreak` | **a207** 연속 해제 셈 | 오류 없음(메모리) | AST `:1658` |

## State mutations and fallbacks

- 원장: 해제(무변경). 메모리: `refusalStreaks`(a207).

## Safety conclusion

- Safe edit boundary: 원장 호출 뒤의 메모리 갱신 하나. 반환값 · 오류 처리는 같음 — submit 의 다섯 호출 자리는 무편집이고 그 자리의 분기 수도 같음(submit AST 18 → 18).
- High-risk impact: yes(청산 해제 경로) — 하지만 이 갈래는 해제가 끝난 뒤의 셈이라 해제 결과 · 시점에 영향 없음.
