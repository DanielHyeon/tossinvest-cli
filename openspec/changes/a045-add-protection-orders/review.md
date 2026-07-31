# Review: a045-add-protection-orders

- Date: 2026-07-31
- Voices: Security, Test Architecture, Maintainability

## Security decision

**ACCEPTED WITH DORMANT SCOPE.** strict attestation schema/parser, pure saga, fake official gateway, migration과 `UNWIRED/OFF` wiring만 승인한다. 실제 conditional mutation, `ProtectionReady=WIRED`, activation 또는 LIVE entry dependency는 사람 attestation 전 금지한다.

## Findings and decisions

1. capability matrix는 account/profile/market/session/type/trigger/quantity/persistence/reservation/idempotency/atomic replace/tool-build/evidence digest를 strict하게 검증한다. legacy/unknown/symlink/owner/mode mismatch는 fail-closed다.
2. first-fill protection은 1초 arm/2초 ACTIVE SLA를 충족하지 못하면 exposure latch와 `PROTECTION_GAP`으로 간다.
3. flatten cancel-confirm deadline은 2초다. 모호하면 `IN_DOUBT`, 최우선 reconcile과 human emergency action으로 전환하고 blind oversell liquidation은 금지한다.
4. older binary가 active protection을 감독할 수 없으므로 binary-only rollback은 금지한다.

## Verification evidence

- OpenSpec strict validation: pass.
- External capability attestation: absent; WIRED/LIVE gate intentionally unavailable.

## Verdict

dormant 범위만 구현할 수 있다. 실제 broker evidence 없이는 이 change의 LIVE capability task와 activation을 완료 처리하지 않는다.

## Dormant implementation record · 2026-07-31

### Pre-Edit Gate

- change/task: `a045-add-protection-orders`, dormant portions only
- target symbols: new strict protection-matrix parser and new pure protection domain/repository
- existing behavior evidence: legacy attest tests/callers, `execgw.ProfileProtection`, official conditional client, engine gateway/reservations, reconciliation and flatten were inspected with current HEAD and refreshed CodeGraph
- upstream inheritance impact: no existing Go function body changed; legacy attestation and UNWIRED entry refusal remain intact
- failing tests first: yes; both new packages failed to compile before implementation and then passed focused/race verification
- safety invariant review: pass for dormant scope; real mutation, WIRED, activation, and operational flatten remain blocked

Function Logic Map: not-applicable — only new Go functions were added. Detailed hard-evidence and branch coverage are recorded in `analysis/dormant-impact.md`.
