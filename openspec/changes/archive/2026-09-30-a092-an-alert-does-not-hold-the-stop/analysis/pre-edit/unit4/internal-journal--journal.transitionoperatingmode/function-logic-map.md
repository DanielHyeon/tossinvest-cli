# Function Logic Map: `Journal.TransitionOperatingMode`

- Source: `internal/journal/operating_mode.go`
- AST evidence: `ast.json` — **편집 전**, :346–485, 분기 28 · 반환 20 · 호출 39, source_sha256 `ed35ea67c26b…`, 추출 HEAD `22db26e7`.
- Risk scan: `risk-pattern-report.md`
- 편집 목적(착지 단위 ④): 22판 D0.3g 5 · 23판 K3 · 24판 M2 — 삽입한 행의 rowid 를 `OperatingModeRecord`(새 필드)에 싣는다(울타리 순서). 방향 판정의 「현재」는 `currentModeTx` 가 rowid 순으로 준다. 커밋 → 투영 → 통지 순서와 사이에 `go`·반환이 없는 모양은 그대로(K3 구조 핀 대상).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 그대로 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch` (:353) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` |
| B2 | `case account == "":` (:354) | — | — | `TestUnusableTransitionRequestsAreRefused` |
| B3 | `case !ValidOperatingMode(mode):` (:357) | — | — | `TestUnusableTransitionRequestsAreRefused` |
| B4 | `case !ValidModeActor(actor):` (:360) | — | — | `TestUnusableTransitionRequestsAreRefused` |
| B5 | `case cause == "":` (:364) | — | — | `TestUnusableTransitionRequestsAreRefused` |
| B6 | `if actor == ModeActorAuto` (:371) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` |
| B7 | `switch mode` (:372) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` |
| B8 | `case ModeHaltAll:` (:373) | — | — | `TestARefusedTransitionIsNotAnnounced`, `TestTheFullTransitionMatrix` |
| B9 | `case ModeNormal:` (:375) | — | — | `TestTheFullTransitionMatrix` |
| B10 | `if !AutomaticTrigger(cause)` (:381) | — | — | (미실행) |
| B11 | `if err != nil` (:392) | — | — | (미실행) |
| B12 | `if err != nil` (:399) | — | — | (미실행) |
| B13 | `if err != nil` (:406) | — | — | (미실행) |
| B14 | `switch` (:409) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` |
| B15 | `case direction == 0:` (:410) | — | — | `TestANoOpTransitionIsNotAnnounced`, `TestConcurrentEscalationsConvergeOnTheStrictestMode` |
| B16 | `case direction < 0 && actor == ModeActorAuto:` (:417) | — | — | `TestConservativePrecedenceKeepsTheStricterMode`, `TestTheFullTransitionMatrix` |
| B17 | `case direction < 0:` (:423) | — | — | `TestTheFullTransitionMatrix`, `TestTransitionIDsDoNotCollideInsideOneSecond` |
| B18 | `if approval == ""` (:424) | — | — | (미실행) |
| B19 | `if req.Auditor == nil` (:428) | — | — | (미실행) |
| B20 | `if err != nil` (:437) | — | — | (미실행) |
| B21 | `if id == ""` (:441) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` |
| B22 | `if _, err := tx.ExecContext(ctx,` (:444) | — | — | (미실행) |
| B23 | `if req.Auditor != nil` (:460) | — | — | `TestConservativePrecedenceKeepsTheStricterMode`, `TestFlattenIsNeverGatedByTheMode` |
| B24 | `if err := req.Auditor.RecordAction(AuditActionOperatingMode, "operating_mode:"+account,` (:461) | — | — | (미실행) |
| B25 | `if err := tx.Commit(); err != nil` (:468) | — | — | (미실행) |
| B26 | `if p := j.modeProjectorRef(); p != nil` (:475) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` |
| B27 | `if req.Announcer != nil` (:478) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` |
| B28 | `if err := req.Announcer.AnnounceOperatingMode(ctx, current.Mode, record); err != nil` (:479) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `AutomaticTrigger`, `AutomaticTriggers`, `UTC`, `ValidModeActor`, `ValidOperatingMode`, `currentModeTx`, `fmt.Errorf`, `formatJournalTime`, `j.clk.Now`, `j.db.BeginTx`, `j.modeProjectorRef`, `modeDirection`, `modeRowCount`, `operatingModeID` | 편집 전 호출 | 편집 전 계약 | AST |

## State mutations and fallbacks

- 편집 전 동작 그대로 — 편집은 위 목적의 범위.

## Safety conclusion

- Safe edit boundary: 위 목적에 적힌 것만.
- High-risk impact: High-risk(원장 · 모드). 분기 B1~B28 의 판정은 그대로, 삽입 결과에서 rowid 를 읽는 오류 갈래가 하나 는다.
