# Planned execution-baseline adoption — 2026-09-06

Status: adoption preparation authorized; not yet adopted. The user authorized
the separate common SDD repair after the original a063 analysis gate was shown
to include intervening committed history. Tooling change a120 passed its actual
gate, was independently accepted and officially archived at
`openspec/changes/archive/2026-09-06-a120-audit-legacy-execution-baseline`.
Its final isolated commit is `71d1017223b536f797041c94e20b6c9febdb02b2`.
The complete reviewed range from E is integrated into the shared checkout;
strict/PM and post-sync SDD checks passed, and all 107 pre-existing dirty or
untracked paths retained their bytes and physical modes. Shared HEAD/index
remain unchanged. This closes the tooling prerequisite, not this adoption.

## Fixed provenance

- Original planning base P remains `da80ce31b6a1ab5d443016768f970a82bab102db`.
- The proposed one-time execution base E is
  `e65e394bf84b3c6e4559a219e816af96d341d75d`.
- This is an explicit `execution-baseline adoption exception` with retrospective
  provenance. It does not claim the original pre-edit evidence was complete.
- Historical P-based analysis and its 327 missing-row measurement remain visible;
  a new ledger must enumerate the intervening commits, parents, paths and net
  function inventory rather than relabeling that history as completed FLM.

## Required sequence

1. Accept, gate and archive a120 through its own independent review sequence.
2. Create a fresh detached a063 worktree descended from E and the accepted tools.
   Copy only the reviewed a063 implementation/evidence; retain the root checkout.
3. Verify the fifteen recorded implementation digests before committing source S.
   Re-derive every E-based modified-existing function. Preserve P-based maps as
   historical artifacts and generate E-based maps explicitly as retrospective
   analysis, including required base-revision deletions.
4. Generate the complete immutable history ledger with the accepted helper.
   Obtain separate adversarial adoption review followed by gstack adoption review
   of the actual source/path/function inventory and inherited-history accounting.
5. Generate the record binding S, the ledger and those real review documents;
   commit the evidence at a strict descendant H. Validate at clean detached H
   with the accepted source/build-input checks and full E-based function maps.
6. Record the actual result and its limits. A successful adoption analysis does
   not complete the service installation, same-profile proof, actual qualifying
      days, final gate or a063 archive. Those conditions remain in tasks 4.1–4.5.

## Implementation and review assignments

Terra owns the isolated source/evidence preparation and all executable checks.
Use a fresh detached worktree based on accepted commit `71d1017...`, verify the
fifteen-file `analysis/gstack/implementation-reviewed-digests.json` manifest,
and commit source S before creating adoption records. No product-code changes
are authorized by this step. Preserve the original P-based bundles in a clearly
historical location; derive the exact E-to-S target set rather than relabeling
the earlier eleven-bundle set. All new maps explicitly retain retrospective
provenance.

Use the already-tested external environment via
`SDD_PYTHON=/tmp/tossos-a120-external-sdd-venv/bin/python`; never create a local
`.sdd/.venv` or relax the source guard. The generator initially leaves review
bindings empty. A dedicated independent Terra adversary reviews the real
ledger/inventory/maps and source snapshot, then a separate gstack review occurs.
Only then bind those actual documents, commit metadata-only descendant H and
perform final clean-detached validation. Manager verifies those results.

All verification processes also inherit an external `PYTHONPYCACHEPREFIX`
directory under `/tmp`. The dedicated Terra preflight verified that explicit
`compileall` writes its caches there; `PYTHONDONTWRITEBYTECODE=1` alone does not
prevent compileall from creating forbidden `scripts/__pycache__` entries. This
keeps the accepted source guard unchanged. Retain the external cache throughout
the run and never treat it as evidence stored in the source snapshot.

## Operating boundary

This plan does not install a binary or unit, start a survey, restart an engine,
alter any trading control, or assert that existing operating evidence proves the
reviewed deployment. The separately reviewed deployment proposal and explicit
human operating approval remain necessary.
