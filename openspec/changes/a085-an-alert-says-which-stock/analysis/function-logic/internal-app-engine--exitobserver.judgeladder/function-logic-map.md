# Function Logic Map: `ExitObserver.judgeLadder`

- Source: `internal/app/engine/exitloop.go` (lines 916–980)
- AST evidence: `ast.json` (`source_sha256: 6625c92061d5b05f566ecb0913f5c5f74a7fdde4cc4b5d8e7dfe8e75dd71de00`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk**

## What it does

ladder 정책 평가. `judgeRatchet`와 동일한 한 곳만 바뀌었다 —
`if !snapshot.Changed && !m.reJudge`. 주석이 이유를 담고 있다: 재판정은 라인이 아니라
선택기가 바뀐 것이라 `Changed`가 알 수 없고, 여기서 돌아가면 재시도만 태우고 격리는 남는다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ladder` | immutable registry의 rung 표 | `o.ladderFor(m)` | 오류면 거절 |
| `m.state.PendingLevel` | 미결 rung 이름 또는 빈 문자열 | `exit_states` | `RungIndex` 실패는 `NoRung`으로 남는다 — upstream 동작 |
| `m.reJudge` | bool | `workingSet` → `judge` | false면 upstream 동작과 동일 |
| `m.state.ActiveRung` | 활성 rung index | `exit_states` | 표보다 길면 `checkLadderPolicyStillFits`가 거절 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Required test |
|---|---|---|---|
| B1 | (919) `if` — if err != nil { | 본문 참조 | 아래 Branch Test Map |
| B2 | (923) `if` — if err := o.checkLadderPolicyStillFits(m, ladder); err != nil { | 본문 참조 | 아래 Branch Test Map |
| B3 | (928) `if` — if err != nil { | 본문 참조 | 아래 Branch Test Map |
| B4 | (933) `if` — if m.state.PendingLevel != "" { | 본문 참조 | 아래 Branch Test Map |
| B5 | (934) `if` — if idx, err := exitpolicy.RungIndex(m.state.PendingLevel); err == nil { | 본문 참조 | 아래 Branch Test Map |
| B6 | (962) `if` — if err != nil { | 본문 참조 | 아래 Branch Test Map |
| B7 | (975) `if` — if !snapshot.Changed && !m.reJudge { | 본문 참조 | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `o.ladderFor` / `o.checkLadderPolicyStillFits` | 정책 표 조회와 적합성 | 오류면 거절 — 판정 없음 | AST `calls` L918·L923 |
| `exitpolicy.RungIndex` | 미결 rung 이름 → index | 오류를 삼키고 `NoRung` 유지 | AST `calls` L934 |
| `exitpolicy.EvaluateLadderSnapshot` | 불변 스냅샷 평가 | 오류면 거절 | AST `calls` L961 |
| `o.record` | 판정 기록과 무장 | 오류 전파 | AST `calls` L978 |

## State mutations and fallbacks

- `o.clearRefused(m.position.ID)`.
- 쓰기는 전부 `record` 안에서 일어난다.

## Safety conclusion

- Safe edit boundary: 조기 반환 조건.
- High-risk impact: yes — ladder 손절도 이 경로를 지난다.
