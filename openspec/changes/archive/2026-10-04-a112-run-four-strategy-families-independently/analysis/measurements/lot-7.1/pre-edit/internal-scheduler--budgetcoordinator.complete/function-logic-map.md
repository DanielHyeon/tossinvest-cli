# Function Logic Map (편집 전): `Complete`

- Source: `internal/scheduler/budget.go`
- Source SHA-256: `9569cfb6db93df7ec359e92ebafcd7e258b9686c60fb8a7a996e98d998b6ab97`
- Signature: `BudgetCoordinator.Complete(params=2, results=1)`
- Source range: `404:1`–`439:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.1, Manager 판정 Q1~Q4 2026-10-01).

## Inputs and invariants

- 편집 계획: 본문을 내부 함수로 옮기고(범위 인자 하나 추가 — 범위 없는 호출은 0 값) 공개 메서드는 그 함수를 0 범위로 부른다. 범위 없는 경로의 판정은 불변이어야 한다.

## Branches and early returns

- Exact AST return nodes: `406:3, 412:3, 416:3, 420:3, 423:3, 427:3, 430:3, 438:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 405:2 | `if c == nil // token == (CommitmentToken{}) {` |
| B2 | if | 411:2 | `if !ok {` |
| B3 | if | 414:2 | `if token.coordinator != c.coordinatorID // token.keyDigest != sha256.Sum256([]byte(key)) //` |
| B4 | if | 419:2 | `if !ok // record.completed // record.class != token.class // record.generation != token.generation {` |
| B5 | if | 422:2 | `if c.now == nil {` |
| B6 | if | 426:2 | `if completedAt.IsZero() {` |
| B7 | if | 429:2 | `if c.completionSequence == ^uint64(0) {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `c.mu.Lock` | 408:2 |
| `c.mu.Unlock` | 409:8 |
| `sha256.Sum256` | 414:64 |
| `(unnamed)` | 414:78 |
| `c.now` | 425:17 |
| `completedAt.IsZero` | 426:5 |
| `uint64` | 429:30 |

## State mutations and fallbacks

- endpoint 단위 commitment 집합 · 발급 기록 · 완료 순번을 쓴다(뮤텍스 아래).

## Safety conclusion

- High-risk 아님(생산 호출자 0 — a070 처분 감사). 편집은 범위 없는 경로의 판정을 바꾸지 않아야 한다.
