# Branch Test Map: `classifyMutation`

- Source: `internal/execgw/classify.go` (21-99)
- AST branches 10 — 실행 칸은 편집 전 패키지 스위트 coverprofile 측정(시험 이름 귀속은 GREEN 재추출 때)

| Branch | Coordinate | Scenario | Baseline executed | RED observed | GREEN observed |
|---|---|---|---|---|---|
| B1 | `if at 22:2` | if err == nil { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B2 | `if at 32:2` | if reason, refused := policyRefusal(err); refused { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B3 | `switch at 44:2` | switch reason, verdict := classifyRefusalCode(err); verdict { | 측정 불가(본문 블록 없음) | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B4 | `case at 45:2` | case refusalCodeDefinitive: | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B5 | `case at 52:2` | case refusalCodeContradictory: | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B6 | `if at 66:2` | if reason, refused := ClassifyBrokerRefusal(err); refused { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B7 | `if at 69:3` | if errors.As(err, &branch) && branch.Source == trading.BranchSourcePostPrepareConfirmation { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B8 | `if at 84:2` | if status, known := statusOf(err); known { | 실행됨 — 패키지 스위트 count=1 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B9 | `if at 87:3` | if outcome.Detail == "" { | 미실행 — 패키지 스위트 count=0 | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
| B10 | `else at 89:10` | } else { | 측정 불가(본문 블록 없음) | 해당 없음 — 편집 전 기준선(0.4) | 해당 없음 — 편집 전 기준선(0.4) |
