# Status — a068-add-kr-us-reversal-lanes

- Date: 2026-08-03
- Scope: paired KR/US reversal lane core (`internal/reversallane`)
- State: focused implementation and tests complete; global integration gates pending

## Delivered together

- `kr_short_absorption_reversal_v1` and `us_short_dislocation_reversal_v1` share one release
  marker and are both registered default OFF.
- Each market has an independent strict schema/evaluator and cannot consume the peer market's
  evidence, activation, session state, or progress.
- Final-leg authorization requires same-scope bounded sweep-break-reclaim evidence; price decline
  alone cannot authorize it.
- Immutable 2:4:8 ceilings, sealed a066 cap/FX inputs, conservative actual-risk accounting,
  partial/duplicate/cancel handling, overage/unknown latches and stop non-retreat are implemented.
- A066 sealers are private to the package; cap seals bind the exact plan digest and evaluated
  reservation quantity, and zero cap/reservation bases fail closed.
- Fill-state corruption and unidentified/zero fills retain prior held/filled accounting and evidence,
  then latch unknown risk so no further exposure can be admitted before reconciliation.
- Unidentified cancel events use full non-ID preimage digests: exact retries coalesce, distinct raw
  observations stay separate, and no unidentified cancel can release held risk.
- Zero-quantity observations remain non-applied even on an existing positive FillID conflict.
- The package exposes no broker, journal writer, exit authority, activation or toggle writer.

## Focused verification

- `go test -count=1 -race ./internal/reversallane` — PASS
- `go vet ./internal/reversallane` — PASS
- strict OpenSpec validation for a067 continuation and a068 reversal — PASS
- `FuzzTwoFourEightConservesQuantity` (2s) — PASS
- `FuzzFillRetryIsIdempotent` (2s) — PASS

## Pending integration ownership

- Broader evidence/campaign/risk/strategy/exit/scheduler regression suite.
- `make sdd-sync`, `make sdd-check`, and `make gate CHANGE=a068-add-kr-us-reversal-lanes`.
- Runtime/registry wiring remains outside this isolated package change and must preserve both lanes OFF.
