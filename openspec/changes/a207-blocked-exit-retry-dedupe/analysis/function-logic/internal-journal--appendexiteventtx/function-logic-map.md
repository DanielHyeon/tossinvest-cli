# Function Logic Map: `appendExitEventTx`

- Source: `internal/journal/exit_state.go` (792-859)
- Qualified: `appendExitEventTx` · revision: current(편집 뒤, worktree `~/tossos-work/a207` 브랜치 `feat/a207-blocked-exit-retry-dedupe`)
- AST evidence: `ast.json` (source_sha256 7305a44b9cbd…) — AST branches 9 · returns 8
- 편집 전 증거: `analysis/ast-evidence/journal--appendExitEventTx.ast.json`(792-851, 7 갈래, sha ea9f939d35d3)
- Risk scan: `risk-pattern-report.md` (파일 단위)

**역할.** exit_events 행 하나를 씀(판정 · 무장 행과 훅 행 공용). a207 은 B2 안에 압축 갈래 둘(B7 · B8)을 더함 — 같은 조건의 반복 무장이면 JSON 셋을 비우고 표지를 씀.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `tx` | 판정 · 해제 트랜잭션(BEGIN IMMEDIATE) | 호출자 | — |
| `row.Evaluation` | nil(훅 · OPENED 등) 또는 recomputed · effective 가 찬 평가 | recordExitJudgementTx | B2 거짓이면 증거 20 열 NULL |
| `row.Action` · `row.ProposedIntentID` | 무장 행이면 둘 다 참, 판정 전용 행이면 action 빈 값 | recordExitJudgementTx | B7 거짓 → 압축 안 함 |
| 불변식 | 행 하나만 INSERT, 압축은 JSON 셋 · effective_source 만 바꿈 | design D3 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:794` 세대 조회 실패 | — | `:796` wrap | 기존 경로(무변경) |
| B2 | `:801` 평가가 있고 effective · recomputed 스냅숏 있음 | JSON · 평탄화 값 준비 | — | `TestA207AStopRefusedForNCyclesIsSubmittedOnTheFirstClearCycle` |
| B3 | `:803` encode 안: 스냅숏 없음 | — | `:804` NULL | `TestExitSnapshotPersistsWholeRecoveryCandidates` |
| B4 | `:809` saved encode 실패 | — | `:810` | 도달 시험 없음(정상 스냅숏은 encode 됨) |
| B5 | `:812` recomputed encode 실패 | — | `:813` | 도달 시험 없음 |
| B6 | `:815` effective encode 실패 | — | `:816` | 도달 시험 없음 |
| B7 | **a207** `:830` 출처 recomputed ∧ action · intent 있음 ∧ arm suppression 없음 | — | — | `TestA207RowsOutsideARefusalRunAreUnchanged`(거짓 갈래) · `TestA207ARepeatedRefusedArmIsWrittenCompact`(참) |
| B8 | **a207** `:832` 술어 조회 성공 ∧ 반복 | JSON 셋 NULL · effective_source=`recomputed_repeat` | — | `TestA207ARepeatedRefusedArmIsWrittenCompact` · `TestA207ALookupFaultWritesTheFullRowAndStillArms`(오류 → 전체) |
| B9 | `:855` INSERT 실패 | — | `:856` wrap | 기존 경로(무변경) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `tx.QueryRowContext`(세대) | 행의 세대 묶기 | 오류 → B1 | AST `:794` |
| `encodeStoredSnapshot`(encode) | JSON 직렬화 | 오류 → B4~B6 | AST `:806` |
| `repeatsARefusedArmTx` | **a207** 같은 조건 반복 판정(한 SQL, 같은 tx) | 오류는 무시하고 전체 증거(D1b) — 판정 오류로 돌려주지 않음 | AST `:832` |
| `tx.ExecContext` | INSERT | 오류 → B9 | AST `:838` |

## State mutations and fallbacks

- 원장: exit_events INSERT 하나(무변경). 압축이면 그 행의 JSON 셋 NULL · effective_source 표지.
- 폴백: 술어 조회 오류 → 전체 증거 행(D1b). tx 가 되돌려진 오류면 B9 가 지금처럼 실패함.

## Safety conclusion

- Safe edit boundary: B2 안, 평탄화 값 계산 뒤. B1 · B3~B6 · B9 와 INSERT 열 목록은 무변경 — 호출자 다섯 중 훅 · OPENED · 재편입 · 조정 닫힘은 B2 거짓이라 바이트 불변.
- High-risk impact: yes — 원장(판정 증거). 무장 · 제출 · 해제 결정은 이 함수 밖(recordExitJudgementTx)이고 바뀌지 않음.
