# Branch Test Map: `buildGateway`

- Source: `internal/app/engine/gateway.go` (245-370)
- AST branches 6 — 실행 칸은 편집 전 패키지 스위트 coverprofile 측정(시험 이름 귀속은 GREEN 재추출 때)

| Branch | Coordinate | Scenario | Baseline executed | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if at 250:2` | if err := checkProjectionWired(in.journal); err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B2 | `if at 272:2` | if err := tracker.Restore(ctx); err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B3 | `if at 280:2` | if err := restoreAlertEntryLatch(ctx, in.journal, entry); err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B4 | `if at 284:2` | if err := bindOperatingModeProjection(ctx, in.journal, entry, in.accountRef, in.logger); err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B5 | `if at 308:2` | if err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B6 | `if at 332:2` | if err != nil { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
