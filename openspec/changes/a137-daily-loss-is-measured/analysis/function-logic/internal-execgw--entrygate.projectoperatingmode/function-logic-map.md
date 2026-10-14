# Function Logic Map: `EntryGate.ProjectOperatingMode`

- Source: `internal/execgw/modegate.go` (35-73, 편집 뒤)
- Qualified: `EntryGate.ProjectOperatingMode`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` 7d7526dd82e19b1e…) · `ast-base.json`(편집 전, base `a000cf29`, `source_sha256` 8cb275362b7d5c25…, 편집 전 분기수 6) — 현재 branches 7 · returns 2
- Risk scan: `risk-pattern-report.md`

**역할.** 운영 모드 행을 진입 관문 latch 로 투영(교체, 커밋 순서 울타리).

**a137 편집.** 모드 latch 를 쓸 때 AUTO · `DAILY_LOSS_LIMIT_REACHED` · ENTRY_BLOCKED 이면 `modeDailyLossDetail = detail`, 그 밖은 ""; 진입을 허용하는 투영은 "" 로 지움. 분기 6 → 7(B6 추가).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `rec` | 원장 운영 모드 행 | journal | Seq 울타리(B3) |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if rec.Actor != "" {` (if 37) | — | 분기 본문 참조 | `go test ./internal/execgw -run Mode` |
| B2 | `if rec.Cause != "" {` (if 40) | — | 분기 본문 참조 | `go test ./internal/execgw -run Mode` |
| B3 | `if rec.Seq <= g.modeSeq {` (if 49) | — | 분기 본문 참조 | `go test ./internal/execgw -run Mode` |
| B4 | `if !rec.BlocksEntry() {` (if 54) | — | 분기 본문 참조 | `TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock`(해제 후 표지 없음) · 기존 모드 투영 시험 |
| B5 | `if had {` (if 56) | — | 분기 본문 참조 | `go test ./internal/execgw -run Mode` |
| B6 | `if rec.Mode == journal.ModeEntryBlocked && rec.Actor == journal.ModeActorAuto && rec.Cause == journal.ModeTriggerDailyLossLimit {` (if 65) | — | 분기 본문 참조 | `TestA137OnlyTheAutomaticDailyLossModeIsAQuietAccountBlock`(사람 전이 · 다른 trigger · HALT_ALL 은 표지 없음) |
| B7 | `if !had {` (if 70) | — | 분기 본문 참조 | `go test ./internal/execgw -run Mode` |

Exact AST return positions: 50 · 60

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `rec.BlocksEntry` | 진입 차단 여부 | — | AST calls |

## State mutations and fallbacks

- g.latches · g.revision(기존) · g.modeDailyLossDetail(a137, g.mu 아래).

## Safety conclusion

- Safe edit boundary: 표지 문자열 하나. latch · 세대 갱신 규칙 무변경.
- High-risk impact: yes(진입 관문) — 표지는 분류에만 쓰이고 차단 여부를 바꾸지 않음.
