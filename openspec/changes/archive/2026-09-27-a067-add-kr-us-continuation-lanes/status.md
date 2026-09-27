# Status: a067-add-kr-us-continuation-lanes

Implementation status: **KR and US continuation contracts implemented together; integration gate pending**.

- KR lane: implemented, tests passing, desired/effective `OFF`.
- US lane: implemented, tests passing, desired/effective `OFF`.
- Same-release conformance: passing; a one-market-only descriptor set is rejected.
- A066 caps are sealed to one exact campaign plan, plan policy and reservation quantity; policy mismatch and cross-plan replay are refused.
- Fill/cancel apply is exact-scope bound. Missing IDs use full-preimage digest identities, exact retries are idempotent, foreign/incomplete events preserve accounting and latch UNKNOWN, and unidentified cancels never release held risk.
- Zero-quantity fills are non-applied evidence with held/filled invariance. Stop candidates use sealed version/digest/freshness provenance, duplicate JSON keys are rejected, untyped invalidation is refused, and cancelled/expired legs are terminal.
- Fill accounting assigns held/filled atomically only after all parsing, transferred<=held and overflow checks pass; identity provenance is bounded to 256 bytes.
- Runtime/broker/journal/toggle wiring: intentionally absent.
- Remaining: repository-wide SDD sync/check and `make gate CHANGE=a067-add-kr-us-continuation-lanes` in the parent integration wave.
