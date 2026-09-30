# Function Logic Map: `Journal.TransitionOperatingMode`

- Source: `internal/journal/operating_mode.go`
- AST evidence: `ast.json` — **편집 뒤**, :353–500, 분기 29 · 반환 21 · 호출 41, source_sha256 `093341ae5773…`, 추출 커밋 `2714e393`. 편집 전 번들은 `analysis/pre-edit/unit4/`.
- Risk scan: `risk-pattern-report.md`
- 편집(착지 단위 ④ `2714e393`): 삽입 결과에서 rowid(`LastInsertId`)를 읽어 레코드 `Seq` 에 실음 — 읽기 오류 갈래 하나 추가. 판정 분기 불변. 방향 판정의 「현재」는 `currentModeTx`(rowid 순)가 줌.
- 재번호: difflib 정렬(편집 전 `22db26e7` 소스 대비): B1~B21 → B1~B21(같은 분기, 줄 이동); 교체 B22 → B22, B23; B23~B28 → B24~B29(같은 분기, 줄 이동).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 | 편집 전 계약 + 위 편집 | 호출자 | 아래 분기 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `switch` (:360) | — | — | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` |
| B2 | `case account == "":` (:361) | — | — | `TestUnusableTransitionRequestsAreRefused` |
| B3 | `case !ValidOperatingMode(mode):` (:364) | — | — | `TestUnusableTransitionRequestsAreRefused` |
| B4 | `case !ValidModeActor(actor):` (:367) | — | — | `TestUnusableTransitionRequestsAreRefused` |
| B5 | `case cause == "":` (:371) | — | — | `TestUnusableTransitionRequestsAreRefused` |
| B6 | `if actor == ModeActorAuto` (:378) | — | — | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` |
| B7 | `switch mode` (:379) | — | — | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` |
| B8 | `case ModeHaltAll:` (:380) | — | — | `TestARefusedTransitionIsNotAnnounced`, `TestTheFullTransitionMatrix` |
| B9 | `case ModeNormal:` (:382) | — | — | `TestTheFullTransitionMatrix` |
| B10 | `if !AutomaticTrigger(cause)` (:388) | — | — | (미실행) |
| B11 | `if err != nil` (:399) | — | — | (미실행) |
| B12 | `if err != nil` (:406) | — | — | (미실행) |
| B13 | `if err != nil` (:413) | — | — | (미실행) |
| B14 | `switch` (:416) | — | — | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` |
| B15 | `case direction == 0:` (:417) | — | — | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestANoOpTransitionIsNotAnnounced` |
| B16 | `case direction < 0 && actor == ModeActorAuto:` (:424) | — | — | `TestConservativePrecedenceKeepsTheStricterMode`, `TestTheFullTransitionMatrix` |
| B17 | `case direction < 0:` (:430) | — | — | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` |
| B18 | `if approval == ""` (:431) | — | — | (미실행) |
| B19 | `if req.Auditor == nil` (:435) | — | — | (미실행) |
| B20 | `if err != nil` (:444) | — | — | (미실행) |
| B21 | `if id == ""` (:448) | — | — | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` |
| B22 | `if err != nil` (:455) | — | — | (미실행) |
| B23 | `if err != nil` (:462) | — | — | (미실행) |
| B24 | `if req.Auditor != nil` (:475) | — | — | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` |
| B25 | `if err := req.Auditor.RecordAction(AuditActionOperatingMode, "operating_mode:"+a` (:476) | — | — | (미실행) |
| B26 | `if err := tx.Commit(); err != nil` (:483) | — | — | (미실행) |
| B27 | `if p := j.modeProjectorRef(); p != nil` (:490) | — | — | `TestA092CurrentModeIsTheLatestCommitNotTheLatestClock`, `TestA092TransitionsCarryTheirCommitSequence` |
| B28 | `if req.Announcer != nil` (:493) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition`, `TestANoOpTransitionIsNotAnnounced` |
| B29 | `if err := req.Announcer.AnnounceOperatingMode(ctx, current.Mode, record); err !=` (:494) | — | — | `TestAFailedAnnouncementDoesNotUndoTheTransition` |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `AutomaticTrigger`, `AutomaticTriggers`, `UTC`, `ValidModeActor`, `ValidOperatingMode`, `currentModeTx`, `fmt.Errorf`, `formatJournalTime`, `j.clk.Now`, `j.db.BeginTx`, `j.modeProjectorRef`, `modeDirection`, `modeRowCount`, `operatingModeID`, `p.ProjectOperatingMode`, `req.Announcer.AnnounceOperatingMode` | 편집 뒤 호출 | 위 편집 참조 | AST |

## State mutations and fallbacks

- 위 편집 요약이 전부 — 나머지는 편집 전과 같음(편집 전 번들 대조).

## Safety conclusion

- Safe edit boundary: 위 편집만.
- High-risk impact: High-risk(원장 · 모드).
