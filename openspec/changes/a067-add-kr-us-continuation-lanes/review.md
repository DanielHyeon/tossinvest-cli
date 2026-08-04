# Review: a067-add-kr-us-continuation-lanes

Date: 2026-08-04 · scope: pure dormant continuation contracts

## Concurrent KR/US outcome

- `kr_short_flow_continuation_v1` and `us_short_participation_continuation_v1` are returned by one registry unit under `a067-kr-us-continuation-v1`.
- Both descriptors remain desired/effective `OFF`; neither evaluator waits for peer readiness, session state, or activation.
- KR uses signed notional flow ppm. US uses share participation and signed price-change ppm. Both use strict market schemas and checked integer arithmetic.
- The pure package has no broker, journal, exit, gateway, operating-toggle, or registry authority dependency.

## RED to GREEN record

The first focused run failed at compile time because the market, frozen-FX, campaign-plan, cap, and evaluation contracts did not exist. The RED test set already contained both KR and US schema boundaries, one-market failure isolation, 8:4:2 allocation, cap/risk admission, fill/cancel replay, invalidation/common-exit separation, property checks, fuzz targets, and static dependency closure.

After implementation, both markets became GREEN in the same package and release. No existing registry, runtime, journal, engine, or dispatch function was edited, so task 1.1's existing-function Logic Map gate is not applicable; the new-function branch maps are executable tests in `internal/continuationlane`.

## Safety review

- Allocation is immutable: `floor(Q*8/14)`, `floor(Q*4/14)`, final remainder; partial fill and cancel never move unused quantity upward.
- Proposed quantity is bounded by both planned remaining and the frozen a066 `q_final`. The private cap seal binds the exact immutable campaign-plan digest, exact plan policy digest, proposed reservation quantity, bucket-set provenance, validity window, and frozen FX snapshot; policy mismatch and same-market/same-policy cross-plan replay are refused.
- KR/US threshold config values are created through sealed constructors. Frozen FX and a066 cap attestations use package-private validated constructors plus private content seals, so external caller literals and post-construction mutation cannot pass evaluation. Same-currency plans reject every FX object and zero-quantity plans are invalid.
- Admission checks `filled + held + proposed <= immutable budget` with checked arithmetic.
- Positive-fill risk is the greater of transferred conservative reservation and exact stop-distance/fee risk. Cross-currency US risk uses the identical official frozen quote, quote-to-account direction, haircut, and final minor-unit ceiling.
- Actual overage or unknown risk applies the risk event and latches all later exposure-raising legs. Duplicate fills and cancels are idempotent; cancel releases held risk only.
- Missing FillID/CancelID uses a length-framed full non-ID preimage digest. The exact raw retry is idempotent, a distinct raw preimage remains separate evidence, and unidentified cancels cannot release held risk.
- Fill and cancel apply paths are bound to the exact plan/risk state, campaign, leg, order, source and RFC3339Nano observation time. Campaign/order/source identities are capped at 256 bytes. Foreign campaign, leg 99, oversized identity and empty/invalid provenance retain prior held/filled accounting, store non-applied evidence and latch unknown risk; only the identical scoped raw retry is a duplicate.
- Quantity-zero fills are retained as non-applied evidence, leave held/filled accounting unchanged and latch unknown risk. The same invariant holds for zero observations that reuse an existing positive FillID.
- Fill accounting computes held and filled successors before assigning either. Transferred-over-held, corrupt held/filled and `filled+risk` overflow preserve both prior values while retaining fill evidence and latching unknown risk; integer text is bounded to 256-bit/78-digit inputs before `big.Int` parsing.
- Stop provenance is created through a sealed constructor binding price, source, policy, version, digest, observed time and fresh-until. Evaluation requires exact `observed_at <= evaluated_at <= fresh_until`; stale but well-formed and post-seal-mutated candidates fail closed.
- Strict KR/US JSON decoding rejects duplicate object keys at every nesting level as well as unknown fields and multiple top-level values.
- Structural/exit invalidation without a non-empty typed code is a typed refusal, and both cancelled and expired legs are terminal.
- Invalidation can suppress an add but never creates an exit decision. `CommonExitIndependent` stays true and the existing common-exit package passes its race test unchanged.
- Registration is descriptive and dormant. There is no LIVE hostname, approval, toggle, broker, journal-writer, or exit-authority mutation.

## Verification

| Check | Result |
|---|---|
| `go test ./internal/continuationlane` | pass |
| `go test -race ./internal/continuationlane ./internal/exitpolicy` | pass |
| allocation/admission property test, 20 repetitions (100,000 generated cases) | pass |
| allocation fuzz after final hardening, 2 seconds (618,747 executions) | pass |
| fill retry/zero-quantity fuzz after final hardening, 2 seconds (276,492 executions) | pass |
| `go vet ./internal/continuationlane` | pass |
| evidence/campaign/risk/strategy/dispatch/exit/scheduler regressions | pass |
| `openspec validate a067-add-kr-us-continuation-lanes --strict` | valid |
| focused statement coverage | 77.5% |

Full-repository `make sdd-sync`, `make sdd-check`, and `make gate CHANGE=a067-add-kr-us-continuation-lanes` remain an integration-stage task. This review does not claim activation or runtime wiring.
