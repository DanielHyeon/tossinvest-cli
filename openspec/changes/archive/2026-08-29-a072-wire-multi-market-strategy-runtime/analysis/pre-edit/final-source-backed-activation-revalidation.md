# Final source-backed activation revalidation — pre-edit map

## Function Logic Map

1. Production activation verification returns the signed manifest generation and expiry together with the opaque activation.
2. Dispatch lease evidence records that signed generation, never the mutable desired revision as a substitute.
3. Lease TTL is bounded by the signed activation expiry.
4. Immediately before broker bytes can be sent, Gateway invokes the strategy-only source revalidator.
5. The revalidator re-reads paired desired/calendar/signed manifests and requires exact market, generation, digest, revision, and calendar identity.
6. Gateway then re-checks entry gate, q_final reservations, and only then creates the send tracker.

## Branch Test Map

| Branch | Expected |
|---|---|
| Exact still-current signed authority | send may proceed |
| Desired OFF/revision changed | NOT_SENT; lease/refusal settled |
| Manifest expired/revoked/replaced | NOT_SENT |
| KR authority changed during US dispatch | US comparison remains market-scoped; paired load still fails safely on central corruption |
| Final revalidator absent | strategy request rejected before transport |
