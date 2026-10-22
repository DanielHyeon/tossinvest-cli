# Branch Test Map: `Saga.settleStep`

- Source: `internal/flatten/flatten.go` (301-416)
- AST branches 11 — 실행 칸은 편집 전 패키지 스위트 coverprofile 측정(시험 이름 귀속은 GREEN 재추출 때)

| Branch | Coordinate | Scenario | Baseline executed | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `switch at 316:2` | switch step.State { | 측정 불가(본문 블록 없음) | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B2 | `case at 317:2` | case journal.FlattenStepDone, journal.FlattenStepFailed: | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B3 | `case at 321:2` | case journal.FlattenStepInDoubt: | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B4 | `if at 327:3` | if err := s.Journal.UpdateFlattenStep(ctx, step.ID, outcome.State, | 측정 불가(본문 블록 없음) | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B5 | `if at 334:2` | if s.DryRun { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B6 | `if at 348:2` | if target, ok := byOrder[step.TargetOrderID]; ok { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B7 | `if at 368:2` | if err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B8 | `if at 375:3` | if err := s.Journal.UpdateFlattenStep(ctx, step.ID, outcome.State, | 측정 불가(본문 블록 없음) | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B9 | `if at 395:2` | if state == journal.FlattenStepInDoubt { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B10 | `if at 400:2` | if outcome.State == journal.FlattenStepInDoubt { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B11 | `if at 411:2` | if err := s.Journal.UpdateFlattenStep(ctx, step.ID, outcome.State, | 측정 불가(본문 블록 없음) | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
