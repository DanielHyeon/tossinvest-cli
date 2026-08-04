# Function Logic Map: `evaluate`

- Source: `internal/continuationlane/evaluator.go`
- AST evidence: `ast.json`

## Inputs and invariants

Consumes sealed plan/evidence/risk state and explicit caller-provided entry/target terms. Admitted
evaluation additionally consumes a sealed a066 cap. Proposal evaluation deliberately has no cap and
returns the immutable current-leg remaining as `q_candidate`. No target inference is permitted.

## Branches and early returns

B1-B6 cover plan, invalidation, OFF, scope and signal refusals. B7 rejects a public saved-stop scalar without a valid private authority. B8-B13 cover stop, leg, latch and zero planned remaining. B14 distinguishes proposal from admitted evaluation: proposal keeps planned remaining as `q_candidate`; admitted evaluation requires a sealed cap and applies `min(remaining,q_final)`. B15-B20 cover arithmetic and campaign risk budget refusals. Proposal checks existing filled+held only; admitted evaluation also includes the exact cap reservation. B21 requires sealed exact execution terms and canonical `stop < entry < target` before decision. The selected saved price comes from private authority, so public scalar mutation cannot retreat it.

## Calls and live bindings

Private saved-stop price selection, pure stop composition, planned-remaining allocation and risk arithmetic only.

## State mutations and fallbacks

None. Proposal authority is not admission authority and carries no cap/snapshot/reservation. Accepted
outcome preserves canonical terms. A valid private saved authority overrides the public compatibility
scalar; an absent authority permits only an empty public scalar.

## Safety conclusion

Safe edit boundary: one shared evaluator with an explicit proposal/admitted mode. Proposal mode may
skip only cap-specific validation/reservation and must preserve every other refusal. High-risk impact:
malformed prices must never reach accepted output and proposal quantity must never exceed planned remaining.
