# Function Logic Map: `ReadOnly.BrokerOrderExitLinks`

- Source: `internal/journal/account_views.go` (373-593)
- Qualified: `ReadOnly.BrokerOrderExitLinks` · revision: current(편집 뒤, worktree `~/tossos-work/a209` 브랜치 `a209-console-exit-evidence-counts-arming-rows-only`)
- AST evidence: `ast.json` (source_sha256 8351a432b348…) — AST branches 28 · returns 12
- 편집 전 증거: `analysis/ast-evidence/journal--ReadOnly.BrokerOrderExitLinks.base.ast.json`(370-588, 28 갈래 · 12 return, sha 26032b49b93b)
- Risk scan: `risk-pattern-report.md` (파일 단위, 일치 0)

**역할.** 콘솔 주문 화면이 보여 줄 브로커 주문마다 broker order → PLACE attempt → intent → exit event 사슬을 하나의 재귀 SQL 로 읽어,
연결 하나 · 모호 · 연결 없음 중 하나로 답함. a209 는 Go 갈래를 바꾸지 않고 SQL 서브쿼리 둘(:453-454 행 수, :471-473 단일 행 연결)에
「무장 행만」 조건을 더함 — 갈래 수 28 → 28, return 12 → 12 그대로.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `scopes` | 1..MaxBrokerOrderEvidenceScopes, market 파싱 가능, 주문 · 계좌 비지 않음, 거래일 YYYY-MM-DD | 호출자(console) | `ErrBrokerOrderEvidenceScope` |
| `exit_events` 행(intent 별) | 무장 행 0..1 이 정상, 훅 행(a207HookActions) 0..n | 원장 | 무장 행 ≥2 → 모호 |
| 불변식 | 무장 행이 둘 이상이면 근거를 고르지 않음(fail-closed) | spec operator-console a209 Requirement | `ambiguous_exit_evidence` |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:376` scope 0 | — | `:377` nil, nil | `TestBrokerOrderExitLinksSkipsEmptyAndBoundsScopeBeforeSQL` |
| B2 | `:379` scope 상한 초과 | — | `:380` 오류 | 같음 |
| B3 | `:384` scope 정규화 루프 | `clean[i]` | — | 모든 시험 |
| B4 | `:386` market 파싱 실패 | — | `:387` 오류 | `TestBrokerOrderExitLinksSkipsEmptyAndBoundsScopeBeforeSQL` |
| B5 | `:392` 주문 · 계좌 빈 값 | — | `:393` 오류 | 같음 |
| B6 | `:395` 거래일 형식 | — | `:396` 오류 | 같음 |
| B7 | `:400` scope JSON 인코딩 실패 | — | `:401` 오류 | 구조상 도달 불가(평범한 struct) |
| B8 | `:476` 질의 실패 | — | `:477` 오류 | 기존(무변경) |
| B9 | `:493` 결과 틀 초기화 | `out` · `aggregates` | — | 모든 시험 |
| B10 | `:505` 행 루프 | — | — | 모든 시험 |
| B11 | `:507` 행 상한 도달 | — | `:508` 오류 | `TestBrokerOrderExitLinksRowSentinelIsStructurallyAboveTheQueryMaximum` |
| B12 | `:518` Scan 실패 | — | `:529` 오류 | 기존(무변경) |
| B13 | `:531` 요청 신원 불일치 | — | `:534` 오류 | 기존(무변경) |
| B14 | `:536` 엔진 주문 | `Engine=true` | — | `TestReadOnlyLinksBrokerOrdersToExitSnapshotsOnlyThroughTheAttemptIntent` |
| B15 | `:539` 안전 사유 switch | `unsafeReason` | — | 아래 case 들 |
| B16 | `:540` 계보 순환 | `lineage_cycle` | — | `TestBrokerOrderExitLinksFailsClosedForPureSingleParentCycle` |
| B17 | `:542` 부모 둘 이상 | `lineage_ambiguous` | — | `TestBrokerOrderExitLinksDoesNotExpandWideBranchingLineage` |
| B18 | `:544` 범위 밖 부모 | `lineage_scope_mismatch` | — | `TestBrokerOrderExitLinksRefusesIncompleteAmendEvidence` |
| B19 | `:546` 깊이 상한 | `lineage_depth_exceeded` | — | `TestBrokerOrderExitLinksFailsClosedOnlyWhenLineageExceedsDepthBound` |
| B20 | `:548` PLACE 둘 이상 | `ambiguous_exit_evidence` | — | `TestBrokerOrderExitLinksAttributesOnlyConfirmedPlaceAttempts` |
| B21 | `:550` **무장 행** 수 > 1(a209: SQL `event_count` 가 훅 행을 뺌) | `ambiguous_exit_evidence` | — | `TestA209TwoArmingRowsStayAmbiguousWithAHookRow` · `TestReadOnlyMarksDuplicateExitEvidenceAmbiguous`; 안 걸리는 쪽 `TestA209AFilledExitOrderLinksItsArmingRowDespiteTheFillHook` · `TestA209AReleasedExitOrderLinksItsArmingRowDespiteTheReleaseHook` |
| B22 | `:553` 연결 행 NULL(a209: 무장 행이 정확히 하나가 아니면 NULL) | — | continue | `TestA209TwoArmingRowsStayAmbiguousWithAHookRow`(NULL) · `TestA209A…`(값 있음) |
| B23 | `:560` 증거 hydrate 실패 | `invalid_event_evidence` | — | 기존(무변경) |
| B24 | `:572` rows.Err | — | `:573` 오류 | 기존(무변경) |
| B25 | `:575` 결과 조립 루프 | — | — | 모든 시험 |
| B26 | `:577` 모호(사유 · 다중 사건 · 다중 attempt) | `Ambiguous=true` | continue | `TestA209TwoArmingRowsStayAmbiguousWithAHookRow` |
| B27 | `:580` 사유 빈 값 | `ambiguous_exit_evidence` | — | `TestReadOnlyMarksDuplicateExitEvidenceAmbiguous` 계열 |
| B28 | `:585` 단일 사건 채우기 | `Event` · `AttemptID` · `IntentID` · `UnknownReason` | `:592` out | `TestA209AFilledExitOrderLinksItsArmingRowDespiteTheFillHook` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `marketclock.ParseMarket` | scope 정규화 | 오류 → scope 오류 | AST calls |
| `json.Marshal` | scope 를 SQL json_each 로 | 오류 반환 | AST calls |
| `r.db.QueryContext` | 재귀 SQL 한 번(a207HookActions 상수를 문자열로 이어 붙임 — 바인드 인자 아님, 상수라 주입 면 없음) | ctx 취소 · 오류 반환, 재시도 없음 | AST calls |
| `hydrateExitEventEvidence` | 무장 행 증거 → 세 뷰 | 실패 → `invalid_event_evidence`(모호 아님) | AST calls |

## State mutations and fallbacks

- 원장 쓰기 0(ReadOnly 핸들). 함수 지역 `out` · `aggregates` 만 바꿈.
- fallback: 연결 없음은 `exit_evidence_unlinked`, 모호는 fail-closed.

## Safety conclusion

- Safe edit boundary: SQL 서브쿼리 두 곳의 WHERE 조건만. Go 갈래 · return · 결과 타입 무변경(AST 갈래 28 · return 12 동일).
- High-risk impact: no — 표시 전용(호출자 `Console.ledgerOrderEvidence` 하나).
