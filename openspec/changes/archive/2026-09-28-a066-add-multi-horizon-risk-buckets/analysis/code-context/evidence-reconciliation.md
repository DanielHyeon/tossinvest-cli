# Evidence reconciliation — a066 Wave 2A (2026-09-25)

- Base `23794f86` · CodeGraph 1.6.0 · Go AST bundles in `analysis/function-logic/` (26 after this lot) ·
  CodeGraphContext unavailable (see `codegraphcontext-context.md`).

| Fact | CodeGraph | AST / HEAD | Conclusion |
|---|---|---|---|
| `Gateway.submit` callers | `callers submit` = 4 production incl. `record` (`internal/app/engine/exitloop.go:1177`) | `record` calls `exitObserver.submit` (`exitloop.go:1343`), a different method; `Gateway.submit` is called only by `place`, `Cancel`, `Amend` | **Mismatch — name-based resolution merges four `submit` methods.** HEAD is authoritative: 3 callers. No edit depends on it |
| callers count cap | default output says "(20)" for `RecordFill` | `--limit 500 --json` returns 147 (1 production) | default limit truncates; every count in the baseline is uncapped |
| a066 functions a gate must see | `codegraph` does not answer "which base functions did this change edit" | Go parser over each a066 commit (`4aee6853 c60fee07 4a364caf 9bf1a3a6 a37d97f5 ebb87d3d`): **18** functions that existed at base had their body changed | 14 had bundles; **4 had none**: `Journal.RecordFill`, `riskbucket.ApplyFill`, test helpers `recordConfirmedFillOrder`, `recordConfirmedFillOrderScope`. Bundles written in Wave 2A |
| bundle freshness | — | 24 bundles vs HEAD: 16 MATCH (2 of them with a map-only defect: `activeScopeWhere`, `scopeArgs` cited a non-branch `B3`), 4 SHIFT (body identical, lines moved; `TestReEntering…` is a `revision: base` bundle and stays as is), 4 DIFF | see `status.md` Wave 2A table |
| `Gateway.submit` growth | impact 34 | AST 25 → 57 branches; `difflib` alignment keeps all 25 old branches, 32 inserted by `8022f578` | a066 sites unchanged (B28, B36); B41 is the strategy-path copy of the a066 barrier |
| `CommitRiskBucketAdmission` DIFF | — | body change is `8022f578`'s `ref.snapshotWindow()`; 44 → 44 branches, alignment identical | not an a066 edit; not gate-required (new since base); bundle refreshed |
| loss lock | no symbol | `grep -rniE 'loss_?lock'` 0 production hits | 2.7 RED references a test-local seam only (`activateEntryLossLock`), no production symbol |
| affected tests | default 1 (Python), `--filter '*_test.go'` 880 | full `./internal/journal/...`, `./internal/execgw/...`, `./internal/riskbucket/...`, `./internal/officialfx/...` run | the filtered set is the one used; default is the documented blind spot |

## Gate window attribution (task 1.2)

`check_analysis.py --change a066-…` at HEAD `0233d776` (working tree has other sessions' dirty files): window
`23794f86 → working tree`, **391** required functions, 395 commits after base. After Wave 2A the only remaining
findings are **372** `missing evidence` lines. None of the 18 a066-edited functions is among them (script check:
each `file function` pair against the missing list, 0 hits). The 372 belong to later changes' commits
(`internal/console` 50, `internal/app` 49, `internal/verifylive` 32, `internal/reconcile` 32, `internal/journal`
32, …) — the stacked-window artifact the change `a122-the-logic-map-gate-outlives-a-merge` exists to fix.
`--record-landing` is refused while tracked files are dirty (other sessions' edits), so the window cannot be
narrowed from this worktree now.

Mismatches are resolved against HEAD; nothing blocks this lot's edits (test file + evidence only).
