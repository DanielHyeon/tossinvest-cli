# Function Logic Map: `validateJudgementSnapshot`

- Source: `internal/journal/exit_snapshot.go` (lines 277–309)
- AST evidence: `ast.json` (`source_sha256: 0e376d1ca4f6e29b27308c540088d35ad9725304b6fc5cff20c8c7eed9780524`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk**

## What it does

판정이 불변 스냅샷과 일치하는지 쓰기 시점에 검증한다. 개정 2가 바꾼 것은 한 줄:
무장 억제 사유의 허용 목록을 리터럴 비교에서 `knownArmSuppression(reason)` 호출로 바꿨다.
쓰기와 읽기가 같은 목록을 쓰게 하기 위해서다 — 개정 1은 쓰기만 넓혀서, 새 사유로 기록된
행을 `ExitEvents`가 손상으로 판정하고 hard-fail 했다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `judgement` | 기록하려는 판정 | `record` | 스냅샷과 어긋나면 오류 — 기록 없음 |
| `stored.Line` | 불변 스냅샷 | 같은 트랜잭션이 저장 | 필드 하나라도 다르면 거절 |
| `judgement.ArmSuppressedReason` | 빈 문자열 또는 `knownArmSuppression` 통과 값 | `record` | 미지의 값이면 거절 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | (278) `if` — if err := stored.validate(); err != nil { | 본문 참조 | 아래 Branch Test Map |
| B2 | (282) `if` — if line.PositionID != positionID \|\| line.ObservationID != judgement.Provenance.ObservationID \|\| | 본문 참조 | 아래 Branch Test Map |
| B3 | (290) `if` — if expected.Zero() { | 본문 참조 | 아래 Branch Test Map |
| B4 | (291) `if` — if judgement.Proposal != nil \|\| judgement.ArmSuppressedReason != "" { | 본문 참조 | 아래 Branch Test Map |
| B5 | (296) `if` — if judgement.Proposal == nil { | 본문 참조 | 아래 Branch Test Map |
| B6 | (297) `if` — if !knownArmSuppression(judgement.ArmSuppressedReason) { | 본문 참조 | 아래 Branch Test Map |
| B7 | (302) `if` — if judgement.ArmSuppressedReason != "" { | 본문 참조 | 아래 Branch Test Map |
| B8 | (305) `if` — if judgement.Proposal.Action != string(expected.Action) \|\| judgement.Proposal.Level != expected.Level { | 본문 참조 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `stored.validate` | 스냅샷 자체 검증 | 오류 전파 | AST `calls` L278 |
| `line.ExecutableProposal` | 스냅샷이 허용하는 제안 | 순수 | AST `calls` L289 |
| `knownArmSuppression` | 무장 억제 사유 허용 목록 | 읽기 경로와 공유 | AST `calls` L297 |

## State mutations and fallbacks

- 없음. 검증만.

## Safety conclusion

- Safe edit boundary: 허용 목록 호출. 쓰기만 넓히면 읽기가 손상으로 판정한다.
- High-risk impact: yes — 원장 무결성 검증.
