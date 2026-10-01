# Function Logic Map (편집 전): `TryAcquire`

- Source: `internal/scheduler/budget.go`
- Source SHA-256: `9569cfb6db93df7ec359e92ebafcd7e258b9686c60fb8a7a996e98d998b6ab97`
- Signature: `BudgetCoordinator.TryAcquire(params=3, results=1)`
- Source range: `310:1`–`396:2`
- AST evidence: `ast.json` — 편집 **전**(a112 7.1, Manager 판정 Q1~Q4 2026-10-01).

## Inputs and invariants

- 편집 계획: 본문을 내부 함수로 옮기고(범위 인자 하나 추가 — 범위 없는 호출은 0 값) 공개 메서드는 그 함수를 0 범위로 부른다. 범위 없는 경로의 판정은 불변이어야 한다.

## Branches and early returns

- Exact AST return nodes: `312:3, 315:3, 318:3, 324:3, 327:3, 337:3, 341:3, 345:3, 349:3, 353:3, 359:3, 366:4, 371:3, 377:3, 381:3, 395:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 311:2 | `if !isKnownPollClass(class) {` |
| B2 | if | 314:2 | `if isSafetyClass(class) {` |
| B3 | if | 317:2 | `if c == nil {` |
| B4 | if | 323:2 | `if ok && state.provenanceConflict {` |
| B5 | if | 326:2 | `if !ok // !state.observation.Reported // state.observation.Reset.IsZero() // state.observation.ObservedAt.IsZe` |
| B6 | if | 331:2 | `if state.hasTrustedReset {` |
| B7 | if | 335:2 | `if !validResetSemantics(obs) {` |
| B8 | if | 339:2 | `if now.Before(obs.ObservedAt) {` |
| B9 | if | 343:2 | `if !now.Before(effectiveReset) {` |
| B10 | if | 347:2 | `if now.Sub(obs.ObservedAt) >= budgetObservationMaxAge {` |
| B11 | if | 351:2 | `if obs.Limit < 0 // obs.Remaining < 0 // obs.Remaining > obs.Limit {` |
| B12 | if | 357:2 | `if discretionary <= 0 // len(state.commitments) >= discretionary {` |
| B13 | if | 362:2 | `if class == PollAnalytics {` |
| B14 | if | 364:3 | `if countCommitments(state.commitments, PollAnalytics) >= analyticsLimit {` |
| B15 | if | 369:2 | `if !c.entropyOK // state.generationExhausted // len(state.issued) >= maxIssuedCommitmentsPerGeneration {` |
| B16 | if | 374:2 | `if _, err := io.ReadFull(c.entropy, capability[:]); err != nil {` |
| B17 | if | 379:2 | `if _, collision := state.issued[capability]; collision {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `isKnownPollClass` | 311:6 |
| `isSafetyClass` | 314:5 |
| `c.mu.Lock` | 320:2 |
| `c.mu.Unlock` | 321:8 |
| `state.observation.Reset.IsZero` | 326:43 |
| `state.observation.ObservedAt.IsZero` | 326:79 |
| `validResetSemantics` | 335:6 |
| `now.Before` | 339:5 |
| `now.Before` | 343:6 |
| `now.Sub` | 347:5 |
| `SafetyReserve` | 355:18 |
| `len` | 357:27 |
| `len` | 361:36 |
| `countCommitments` | 364:6 |
| `len` | 369:50 |
| `io.ReadFull` | 374:15 |
| `sha256.Sum256` | 391:16 |
| `(unnamed)` | 391:30 |

## State mutations and fallbacks

- endpoint 단위 commitment 집합 · 발급 기록 · 완료 순번을 쓴다(뮤텍스 아래).

## Safety conclusion

- High-risk 아님(생산 호출자 0 — a070 처분 감사). 편집은 범위 없는 경로의 판정을 바꾸지 않아야 한다.
