## 1. Contract and hard evidence

- [ ] 1.1 Reserve `STORY-TOS-a121`, capture the implementation base, and validate this change strictly.
- [ ] 1.2 Produce CodeGraph and Go AST Function Logic/Branch Test Maps for the existing record,
      outstanding, cleanup-plan, and official conditional-read functions that will be edited.
- [ ] 1.3 Complete proposal-freeze adversarial and gstack reviews; record the accepted no-live-mutation boundary.

## 2. RED

- [ ] 2.1 Add failing tests that a DELETE 404 or a generic operator observation cannot reconcile an artifact.
- [ ] 2.2 Add failing tests for complete fresh OPEN+CLOSED pagination, exact opaque identity, profile/account
      binding, stale/partial/ambiguous reads, and idempotent append-only reconciliation.
- [ ] 2.3 Add failing tests that reconciliation removes only its exact outstanding artifact from resume cleanup
      planning while preserving failed cleanup evidence and every verification verdict.
- [ ] 2.4 Add failing tests that the reconciliation event cannot enter successful endpoints, soak attestation,
      or engine-interlock coverage, and that no live mutation is reachable.

## 3. GREEN

- [ ] 3.1 Implement the minimal versioned append-only reconciliation event and safe record projection.
- [ ] 3.2 Implement the bounded official-read reconciliation command with redacted operator output and no
      mutation path.
- [ ] 3.3 Preserve backward-compatible record handling or fail closed before write; document the rollback rule.

## 4. VERIFY and handoff

- [ ] 4.1 Run focused tests, `make test`, `make vet`, `make validate`, `make sdd-sync`, and `make sdd-check`.
- [ ] 4.2 Complete independent adversarial diff/test review followed by gstack review and Manager verification.
- [ ] 4.3 Record a read-only, redacted reconciliation observation only with explicit human approval for the
      selected profile; do not perform a live order mutation.
- [ ] 4.4 Synchronize PM, run `make gate CHANGE=a121-reconcile-stale-verification-artifacts`, and archive only
      after it succeeds. a063 remains independently unarchived until its own tasks complete.
