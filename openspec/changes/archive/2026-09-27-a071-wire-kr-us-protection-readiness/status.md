# Status — a071-wire-kr-us-protection-readiness

- Isolated core: GREEN (`internal/protectionreadiness`)
- Isolated lifecycle: GREEN (`internal/protectionlifecycle`), dormant and authority-free
- Release: KR and US readiness are designed and verified concurrently as independent paired read-only lanes; neither market waits for the other, and both remain `UNWIRED`
- Signature: strict canonical JSON, Ed25519-only explicit allowlist, pinned roots
- Key lifecycle: key ID, revocation, bounded overlap and cross-key monotonic serial
- Time/state: every valid trusted-time observation advances the sealed floor; corrupt preimages are preserved
  with state commit forbidden; accepted serial/time changes form one pure durable transition
- Scope: exact account/profile/market/order/session/quantity/trigger/replace/build/evidence and broker identity/query/dedup contract
- Wiring: production consumes simultaneous KR/US readiness contracts read-only, but both supervisor assemblies are
  deliberately `Wired=false`; no committed-fill lifecycle currently derives exact trigger/expiry authority and no
  official protection gateway/controller is constructed
- Decision boundary: Gateway checks the exact market snapshot before journaling refusal and revalidates the
  same generation/identity immediately before broker transport; snapshot drift is `NOT_DISPATCHED`
- Scalar compatibility: `ProfileProtection=UNWIRED` remains reporting-only. The public scalar test override and
  exported WIRED/UNWIRED test pointers were removed; no production bool/config/string can authorize entry
- Reduction continuity: reduce-only SELL, CANCEL and exposure-reducing AMEND do not read the readiness provider;
  stop/emergency/reconciliation/fill loops have no readiness dependency
- File input: exact path/owner/`0600`/regular/non-symlink/size, duplicate and unknown fields rejected
- Authority: immutable readiness snapshot only; all production evidence remains `UNWIRED` because lifecycle wiring
  is absent, and no lane, autostart, automation, toggle or LIVE approval is changed
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
- Security hardening: every cached market is rechecked independently at expiry/revocation/rotation boundaries,
  including when only its peer artifact changes; a root-dirfd component walk with `openat(O_NOFOLLOW)` and an
  owner-only cross-process state lock protect attestation and monotonic state; only a post-state fsynced bootstrap
  marker makes later missing state corrupt, and peer serial advancement invalidates stale cache
- Invalid manifest safety: both market contracts are constructed in local state and published atomically only
  after the full pair validates; malformed partial evidence leaves entry closed without blocking safety runtime
- Superseded (2026-08-11): task 3.5 is transferred to `a100-wire-fill-to-broker-protection` and is no longer part of
  this change's completion condition. Its precondition — journal-committed fill → exact journal-derived
  position/campaign stop+expiry → durable idempotent Plan/Register lifecycle with KR/US official fixture proof —
  lies outside this change's scope, and a100 measured that 9 of 13 branches in the wiring targets have never
  executed their true outcome, so the work is refusal-path RED tests first, wiring second. Rationale and the
  contracts a100 must not redesign are recorded in `tasks.md` §6. Reversible: if a100 is cancelled or drops
  production assembly, 3.5 returns here.
- Remaining: the change set is now frozen at task 3.4. Tasks 5.1–5.3 (post-edit maps, full test/vet/OpenSpec,
  `make sdd-sync`/`sdd-check`/`gate` plus adversarial independent review) run against that frozen set.

No market was enabled, no attestation was installed into production, no production controller can be minted, and no
live broker request was made.
