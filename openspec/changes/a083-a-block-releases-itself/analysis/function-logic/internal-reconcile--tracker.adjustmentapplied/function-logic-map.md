# Function Logic Map: `Tracker.AdjustmentApplied`

- Source: `internal/reconcile/mismatch.go` (lines 354–367)
- AST evidence: `ast.json` (`source_sha256: 0adce8e7229ac24e1ef08d6d8b31b6b5c6b390cfcb706a1de6ad7fd444958113`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **High-risk** — 해제 증거의 유일한 생산자.

## What it does

조정이 한 심볼의 투영을 계좌값으로 수렴시켰다는 사실을, 그 조정이 계산된 비교의 as-of와 함께 기록한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| comparison | 조정이 계산된 diff의 as-of (RFC3339) | `Converger.ConvergeQuantities`의 `asOf` | 빈 문자열·파싱 불가는 영원히 사용 불가 credit이 되어 차단이 유지된다 (fail-closed) |
| symbols | 수렴된 심볼들. 공백·빈 문자열은 무시 | 같은 호출 | 무시 |
| t.adjusted | 심볼 → 비교 as-of. nil이면 지연 초기화 | 이 함수와 `Observe` | 재시작 시 복원하지 않는다 (의도) |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (357) `if` — if t.adjusted == nil | 본문 참조 | — | 아래 Branch Test Map |
| B2 | (360) `range` — for _, symbol := range symbols | 본문 참조 | — | 아래 Branch Test Map |
| B3 | (362) `if` — if key == "" | 본문 참조 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 355, 'column': 2}, 'text': 't.mu.Lock'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 356, 'column': 8}, 'text': 't.mu.Unlock'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 361, 'column': 10}, 'text': 'strings.ToUpper'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 361, 'column': 26}, 'text': 'strings.TrimSpace'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 365, 'column': 21}, 'text': 'strings.TrimSpace'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- `t.adjusted[심볼] = comparison`. 같은 심볼의 기존 credit은 덮어쓴다 — 더 최근 조정이 더 강한 증거다.

## Safety conclusion

- Safe edit boundary: 인자 하나 추가와 map 값 타입 변경. 차단·gate·원장은 건드리지 않는다.
- High-risk impact: yes — 해제 증거를 만드는 함수다. 다만 이 함수 자체는 아무것도 해제하지 않으며, 판정은 전부 `Observe`가 한다.
