# Status — a071-wire-kr-us-protection-readiness

- Isolated core: GREEN (`internal/protectionreadiness`)
- Release: KR and US paired, both default `UNWIRED`
- Signature: strict canonical JSON, Ed25519-only explicit allowlist, pinned roots
- Key lifecycle: key ID, revocation, bounded overlap and cross-key monotonic serial
- Time/state: every valid trusted-time observation advances the sealed floor; corrupt preimages are preserved
  with state commit forbidden; accepted serial/time changes form one pure durable transition
- Scope: exact account/profile/market/order/session/quantity/trigger/replace/build/evidence and broker identity/query/dedup contract
- Wiring: exact sealed supervisor binding required in addition to valid attestation
- File input: exact path/owner/`0600`/regular/non-symlink/size, duplicate and unknown fields rejected
- Authority: immutable readiness snapshot only; external mutation count is always zero; no lane, activation,
  toggle, LIVE approval or broker transport
- Remaining: existing controller/gateway/engine/journal mapping and integration, lifecycle recovery tests, full repository gates and independent final review

No market was enabled, no attestation was installed into production, and no live broker request was made.
