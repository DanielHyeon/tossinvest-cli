# Branch Test Map: `Journal.recordExitJudgementTx`

AST의 모든 분기를 1행씩 덮는다. 조건 열은 `internal/journal/exit_state.go`의 실제 소스 줄이고,
테스트 열과 판정 열은 `go test -covermode=count` 프로파일에서 **측정**한 값이다.
주장이 아니라 측정이므로 이 표는 덮이지 않은 분기를 숨길 수 없다.

> 미검증 분기: B1, B2, B3, B4, B5, B6, B7, B8, B9, B10, B11, B12, B13, B14, B15, B16, B17, B18, B19, B20, B21, B22, B23, B24, B25, B26, B27, B28, B29, B30, B31, B32, B33, B34, B35, B36, B37, B38, B39, B40, B41, B42, B43, B44, B45, B46, B47, B48, B49, B50, B51. 이 change가 바꾼 동작이 아니거나 테스트 하네스가 구성할 수 없는 상태다.
> 관측하지 않은 것을 관측했다고 적지 않는다 — 네 번째 독립 리뷰가 잡은 것이 그 행들이다.

| Branch | Condition | Covering test | Measured |
|---|---|---|---|
| B1 | (395) `if` — if id == "" { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B2 | (398) `if` — if err := judgement.Provenance.validate(); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B3 | (401) `if` — if judgement.Provenance.zero() && strings.TrimSpace(judgement.ArmSuppressedReason) != "" { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B4 | (404) `if` — if judgement.Proposal != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B5 | (405) `if` — if err := validateProposal(*judgement.Proposal); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B6 | (408) `if` — if err := judgement.Proposal.Provenance.validate(); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B7 | (411) `if` — if judgement.Provenance.zero() != judgement.Proposal.Provenance.zero() \|\| | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B8 | (418) `if` — if !judgement.Provenance.zero() { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B9 | (422) `if` — if err := validateJudgementSnapshot(id, judgement, candidate); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B10 | (430) `if` — if err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B11 | (436) `if` — if err != nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B12 | (439) `if` — if current.Completed { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B13 | (443) `if` — if expectedLifecycle == 0 { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B14 | (446) `if` — if expectedLifecycle != current.LifecycleGeneration { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B15 | (455) `if` — if errors.Is(err, sql.ErrNoRows) { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B16 | (457) `else` — } else if err != nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B17 | (457) `if` — } else if err != nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B18 | (460) `if` — if lifecycleStatus != positionpolicy.StatusManaged \|\| lifecycleGeneration != expectedLifecycle { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B19 | (464) `if` — if recomputed != nil && recomputed.Line.PositionGeneration != current.PositionGeneration { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B20 | (468) `if` — if recomputed != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B21 | (472) `if` — if err == nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B22 | (479) `if` — if !errors.Is(err, sql.ErrNoRows) { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B23 | (487) `if` — if recomputed == nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B24 | (488) `if` — if err := notBelow("high water", id, judgement.HighWater, current.HighWater); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B25 | (491) `if` — if err := notBelow("baseline", id, judgement.Baseline, current.Baseline); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B26 | (496) `if` — if level == "" { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B27 | (502) `if` — if current.Effective != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B28 | (506) `if` — if recomputed != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B29 | (508) `if` — if saved != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B30 | (513) `if` — if selectErr != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B31 | (514) `if` — if _, qerr := quarantineExitSnapshotTx(ctx, tx, id, recomputed.Line.PositionGeneration, | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B32 | (518) `if` — if err := tx.Commit(); err != nil { | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
| B33 | (527) `if` — if err := releaseReJudgedQuarantineTx(ctx, tx, judgement.ReJudgingVersion, id, | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B34 | (531) `if` — if source == exitpolicy.RecoverySavedMonotone { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B35 | (534) `else` — } else { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B36 | (544) `if` — if effectiveSource == EffectiveSourceSaved { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B37 | (557) `if` — if effective != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B38 | (559) `if` — if err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B39 | (575) `if` — if _, err := tx.ExecContext(ctx, updateSQL, args...); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B40 | (578) `if` — if err := j.runExitWriteHook("after_state"); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B41 | (584) `if` — if judgement.Proposal != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B42 | (586) `if` — if err := armExitProposalTx(ctx, tx, id, *judgement.Proposal, now); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B43 | (589) `if` — if err := j.runExitWriteHook("after_arm"); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B44 | (600) `if` — if recomputed != nil && effective != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B45 | (603) `if` — if err := appendExitEventTx(ctx, tx, event); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B46 | (606) `if` — if err := j.runExitWriteHook("after_event"); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B47 | (609) `if` — if err := tx.Commit(); err != nil { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B48 | (613) `switch` — switch { | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B49 | (614) `case` — case effectiveSource == EffectiveSourceSaved: | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B50 | (616) `case` — case judgement.Proposal != nil: | 패키지 suite (기존 커버리지) | 측정: 이 change의 테스트는 닿지 않는다. 패키지 전체 suite가 덮는다 |
| B51 | (620) `case` — case judgement.ArmSuppressedReason == ArmSuppressedWorkingOrder, | 없음 | **측정: 어떤 테스트도 이 줄을 실행하지 않는다** |
