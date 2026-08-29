# Function Logic Map: `Snapshot.Digest`

- Source: `internal/reconcile/snapshot.go`
- AST evidence: `ast.json` — lines 174-204, branches 3
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| orders, holdings, balances | one complete account snapshot | Collector output | canonical deterministic rendering; timestamps excluded |
| holding quantity | blank/unreadable evidence or exact finite decimal | raw holdings boundary | blank/unreadable remains distinct from exact zero |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | each sorted open order | append canonical order identity and decimal fields | continue | digest order-independence tests |
| B2 | each sorted holding | append symbol, evidence-preserving quantity, canonical average price | continue | `TestA110SnapshotDigestAndStabiliserDistinguishBlankHoldingFromExactZero` |
| B3 | each sorted balance | append currency and canonical buying power | continue | digest order-independence tests |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `brokerOrderIdentity` | deterministic order sorting/rendering | unchanged identity vocabulary | AST B1 |
| `canonicalHoldingQuantity` | keep blank/unreadable holding quantity distinct from exact zero | exact finite values still canonicalize | AST B2; F10 RED/GREEN |
| `canonicalDecimal` | preserve established optional-decimal blank-as-zero vocabulary | order, average price and balance only | AST B1-B3 |

## State mutations and fallbacks

- Pure rendering; it sorts copies and does not mutate the snapshot.
- Holding quantity is deliberately narrower than other optional decimals because the stabiliser treats digest equality as corroborating account evidence.

## Safety conclusion

- Safe edit boundary: use the evidence helper only for holding quantity; leave order, average-price and balance vocabulary unchanged.
- High-risk impact: yes. Aliasing blank and zero can make two disagreeing account reads stable and authorize reconciliation/adoption work.
