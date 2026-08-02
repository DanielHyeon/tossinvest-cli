# Function Logic Map: `positionRow.Label`

- Source: `internal/console/portfolio.go`
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| Typed management projection | present/absent | position-policy runtime projection | present projection label wins |
| Lifecycle | managed/released/unknown/legacy | position journal + lifecycle state | unknown remains explicit; released never becomes a pending candidate |
| Desired lists | designated/excluded | adoption configuration | exclude wins; designation is only a pending request |
| Exit state | absent/active/completed | journal exit projection | used only after management ownership is established |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | typed management projection exists | none | projection label | management projection tests |
| B2 | enter ordered status switch | none | one of B3-B10 | truth-table test |
| B3 | journal truth is unknown | none | `관리 여부 불명` | truth-table test |
| B4 | lifecycle is known `RELEASED` | none | `관리 외(운영자 해제)` | truth-table test |
| B5 | unmanaged and explicitly excluded | none | `관리 제외` | truth-table test |
| B6 | shared pending-designation predicate is true | none | `관리 편입` | truth-table test |
| B7 | remaining unmanaged state | none | `관리 외(미편입)` | truth-table test |
| B8 | managed exit completed | none | `관리 종료` | existing portfolio tests |
| B9 | managed exit active | none | `엔진 관리` | managed/designated collision test |
| B10 | managed without exit state | none | `엔진 관리(대기)` | existing portfolio tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `HasManagementProjection` | preserve typed runtime verdict | pure predicate | AST + projection tests |
| `Unknown` | keep unreadable journal explicit | pure predicate | AST + truth-table test |
| `Managed` | select only established engine ownership | pure predicate | AST + portfolio tests |
| `PendingDesignation` | centralize desired-only candidate eligibility | pure predicate | AST + truth-table test |

## State mutations and fallbacks

- Pure display projection; no mutation, order, reconcile, or configuration write.
- Priority is projection, unknown, released, excluded, pending candidate, unmanaged, then managed exit detail.

## Safety conclusion

- Safe edit boundary: status copy only, while lifecycle priority must remain aligned with `ProjectManagement`.
- High-risk impact: no direct trading side effect; incorrect ordering can misrepresent protection state, so branch coverage is required.
