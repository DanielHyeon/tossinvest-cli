# Function Logic Map: `hydrateExitEventEvidence`

- Source: `internal/journal/exit_snapshot.go` (397-426)
- Qualified: `hydrateExitEventEvidence` · revision: current(편집 뒤, worktree `~/tossos-work/a207` 브랜치 `feat/a207-blocked-exit-retry-dedupe`)
- AST evidence: `ast.json` (source_sha256 09af632ae584…) — AST branches 7 · returns 8
- 편집 전 증거: `analysis/ast-evidence/journal--hydrateExitEventEvidence.ast.json`(392-417, 6 갈래, sha 19ebe0eeb987)
- Risk scan: `risk-pattern-report.md` (파일 단위)

**역할.** exit_events 행의 판정 증거를 뷰로 읽음(리더 셋 공용). a207 은 압축 표지 갈래 하나(B2)를 더함.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `evidence` | 20 열의 NULL/값 | 스캔 결과 | 부분 모양 → 손상 |
| 불변식 | 표지 없는 부분 증거는 손상(a206 C2) | design D4 | — |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:398` 증거 열이 하나도 없음 | 세 뷰 legacy | `:402` nil | `TestLegacyEventRequiresEveryV10ColumnToBeNull` |
| B2 | **a207** `:404` effective_source = `recomputed_repeat` | `hydrateCompactExitEvent` 가 모양 검사 · 반복 사유 | `:406` 그 결과 | `TestA207ReadersReadACompactRowAsARepeat` · `TestA207AMarkedRowWithTheWrongShapeIsCorrupt` |
| B3 | `:408` 전체 튜플 아님 | — | `:409` 손상 | `TestA207ReadersReadACompactRowAsARepeat`(표지 지운 부분 행) |
| B4 | `:413` saved 디코드 실패 | — | `:414` | `TestExitSnapshotCorruptionIsPerPositionAndNeverRecomputed` |
| B5 | `:416` recomputed 디코드 실패 | — | `:417` | 기존(무변경) |
| B6 | `:419` effective 디코드 실패 | — | `:420` | 기존(무변경) |
| B7 | `:422` 평탄화 ↔ 스냅숏 불일치 | — | `:423` | 기존(무변경) |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `evidence.any` · `evidence.full` | 모양 판정 | — | AST |
| `hydrateCompactExitEvent` | **a207** 압축 모양 검사 | 모양 어긋남 → `ErrExitSnapshotCorrupt` | AST `:406` |
| `decodeExitEventSnapshot` · `validateExitEventFlattenedEvidence` · `validateExitEventArmSuppression` | 전체 증거 검사 | 오류 → 손상 | AST |

## State mutations and fallbacks

- 쓰기 없음(읽기 함수). 압축 행은 세 뷰 UnknownReason=`repeat_of_refused_proposal`, EffectiveSource=표지.

## Safety conclusion

- Safe edit boundary: B1 뒤 · B3 앞에 표지 갈래 하나. 표지가 아닌 행의 경로는 무변경.
- High-risk impact: 원장 리더(콘솔 계좌 이력 · orders, 생산 호출자 0 인 Journal.ExitEvents). 엔진은 이 함수를 쓰지 않음(a206 결론).
