# Status — a071-wire-kr-us-protection-readiness

- Isolated core: GREEN (`internal/protectionreadiness`)
- Isolated lifecycle: GREEN (`internal/protectionlifecycle`), dormant and authority-free
- Release: KR and US paired, both default `UNWIRED`
- Signature: strict canonical JSON, Ed25519-only explicit allowlist, pinned roots
- Key lifecycle: key ID, revocation, bounded overlap and cross-key monotonic serial
- Time/state: every valid trusted-time observation advances the sealed floor; corrupt preimages are preserved
  with state commit forbidden; accepted serial/time changes form one pure durable transition
- Scope: exact account/profile/market/order/session/quantity/trigger/replace/build/evidence and broker identity/query/dedup contract
- Wiring: exact sealed supervisor binding required in addition to valid attestation
- Decision boundary: Gateway checks the exact market snapshot before journaling refusal and revalidates the
  same generation/identity immediately before broker transport; snapshot drift is `NOT_DISPATCHED`
- Scalar compatibility: `ProfileProtection=UNWIRED` remains reporting-only. The public scalar test override and
  exported WIRED/UNWIRED test pointers were removed; no production bool/config/string can authorize entry
- Reduction continuity: reduce-only SELL, CANCEL and exposure-reducing AMEND do not read the readiness provider;
  stop/emergency/reconciliation/fill loops have no readiness dependency
- File input: exact path/owner/`0600`/regular/non-symlink/size, duplicate and unknown fields rejected
- Authority: immutable readiness snapshot only; external mutation count is always zero; no lane, activation,
  toggle, LIVE approval or broker transport
- Lifecycle identity: exact account/position/market/generation/revision/operation kind; exact scoped lookup and
  broker-ID recovery only
- Lifecycle entry: closed unless exact full-coverage ACTIVE and unlatched; pending/unknown/reconcile/terminal closed
- Recovery: active protection retained, no unattested resubmit, first orphan evidence retained, duplicate fill once
- Market isolation: KR and US lifecycle/recovery are independent; fills and exits continue through a market latch
- Verification: targeted and affected-package tests, affected-package race, vet, strict OpenSpec and diff-check pass
- Analysis: all a071-owned post-edit maps are current; repository analysis checker remains non-green only for
  concurrent a066/a073 modified functions outside this wave
- SDD sync: CodeGraph fingerprint synced 11 changed files; advisory CodeGraphContext update stalled for more than
  two minutes and was interrupted without changing runtime state
- Remaining: production controller/supervisor/official-gateway mutation wiring (3.5), full repository gates and independent final review

No market was enabled, no attestation was installed into production, and no live broker request was made.
