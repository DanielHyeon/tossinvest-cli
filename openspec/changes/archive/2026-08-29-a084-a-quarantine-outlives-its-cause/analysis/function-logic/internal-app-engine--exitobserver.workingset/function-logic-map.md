# Function Logic Map: `ExitObserver.workingSet`

- Source: `internal/app/engine/exitloop.go` (lines 476–585)
- AST evidence: `ast.json` (`source_sha256: 3cf97f20c5eafa4b8f4d57bdbc1bc9d9f639c1f425590e4f212b238e7d0d5c8c`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk** — 어떤 포지션이 손절 판정을 받는지 결정한다.

## What it does

보유 포지션과 exit state를 맞춰 판정 대상 목록을 만든다. a084는 활성 격리 분기 하나를 나눈다 — 현재 선택기가 쓴 격리는 지금처럼 제외하고, 다른(또는 미기록) 선택기가 쓴 격리는 제외하지 않고 판정으로 보낸다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 포지션 투영 | 계좌의 비-CLOSED 인스턴스 | `Journal.Positions` | 읽기 실패는 사이클 오류 |
| exit state | 포지션별 저장 상태 | `Journal.OpenExitStateResults` | corruption은 격리 생성 |
| 활성 격리 | 세대별 최대 1행 | `ActiveExitSnapshotQuarantine` | 읽기 실패는 사이클 오류 |
| `q.SelectorRevision` | 격리를 만든 선택기 개정. 0은 미기록 | 격리 행 | 0은 '다름'으로 취급 — 재판정 대상 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (478) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (482) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (486) `range` — for _, result := range stateResults | 본문 참조 | — | 아래 Branch Test Map |
| B4 | (491) `range` — for _, p := range positions | 본문 참조 | — | 아래 Branch Test Map |
| B5 | (492) `if` — if p.State == journal.PositionClosed || isZeroQuantity(p.Quantity) | 본문 참조 | — | 아래 Branch Test Map |
| B6 | (495) `if` — if !p.ExitEligible() | 본문 참조 | — | 아래 Branch Test Map |
| B7 | (506) `if` — if !ok | 본문 참조 | — | 아래 Branch Test Map |
| B8 | (508) `if` — if err != nil | 본문 참조 | — | 아래 Branch Test Map |
| B9 | (509) `if` — if cycle.Err == nil | 본문 참조 | — | 아래 Branch Test Map |
| B10 | (514) `if` — if opened.PositionID == "" | 본문 참조 | — | 아래 Branch Test Map |
| B11 | (523) `if` — if result.Corruption != nil | 본문 참조 | — | 아래 Branch Test Map |
| B12 | (526) `if` — if qerr != nil | 본문 참조 | — | 아래 Branch Test Map |
| B13 | (527) `if` — if cycle.Err == nil | 본문 참조 | — | 아래 Branch Test Map |
| B14 | (537) `if` — if q, active, qerr := o.opts.Journal.ActiveExitSnapshotQuarantine(ctx, p.ID, p… | 본문 참조 | — | 아래 Branch Test Map |
| B15 | (542) `else` — } else if active && !q.NeedsReJudgement() | 본문 참조 | — | 아래 Branch Test Map |
| B16 | (538) `if` — if cycle.Err == nil | 본문 참조 | — | 아래 Branch Test Map |
| B17 | (542) `if` — } else if active && !q.NeedsReJudgement() | 본문 참조 | — | 아래 Branch Test Map |
| B18 | (546) `else` — } else if active | 본문 참조 | — | 아래 Branch Test Map |
| B19 | (546) `if` — } else if active | 본문 참조 | — | 아래 Branch Test Map |
| B20 | (563) `if` — if identityErr != nil | 본문 참조 | — | 아래 Branch Test Map |
| B21 | (566) `if` — if qerr != nil | 본문 참조 | — | 아래 Branch Test Map |
| B22 | (567) `if` — if cycle.Err == nil | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 477, 'column': 20}, 'text': 'o.opts.Journal.Positions'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 481, 'column': 23}, 'text': 'o.opts.Journal.OpenExitStateResults'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 485, 'column': 16}, 'text': 'make'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 485, 'column': 57}, 'text': 'len'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 492, 'column': 43}, 'text': 'isZeroQuantity'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 495, 'column': 7}, 'text': 'p.ExitEligible'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 501, 'column': 4}, 'text': 'o.alertUnmanaged'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 507, 'column': 19}, 'text': 'o.openState'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 524, 'column': 15}, 'text': 'o.opts.Journal.QuarantineExitSnapshot'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 525, 'column': 32}, 'text': 'result.Corruption.Error'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 532, 'column': 14}, 'text': 'append'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 533, 'column': 18}, 'text': 'fmt.Errorf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 534, 'column': 4}, 'text': 'o.announceQuarantine'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 537, 'column': 25}, 'text': 'o.opts.Journal.ActiveExitSnapshotQuarantine'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- `refused` / `out` 목록 구성. 원장 쓰기는 corruption·identity 격리 생성 두 경로뿐이며 a084가 바꾸지 않는다.

## Safety conclusion

- Safe edit boundary: 활성 격리 분기의 조건 하나와 재판정 진입 시의 로그 한 줄. 격리 생성 경로·정책 identity 해소·순서(격리는 마지막)는 그대로다.
- High-risk impact: yes — 판정 대상 집합을 바꾼다. 방향은 보호가 있는 쪽이다: 지금 제외되는 포지션은 손절도 평가되지 않으며, 재판정은 같은 선택을 다시 돌릴 뿐 어떤 규칙도 완화하지 않는다.
