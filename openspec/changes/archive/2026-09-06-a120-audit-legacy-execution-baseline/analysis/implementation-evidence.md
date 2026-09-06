# Implementation evidence — implementation ready, acceptance pending

This record is intentionally incomplete. It must not be read as an acceptance,
adoption approval, or final gate result.

- Added `tools/logic-map/execution_baseline.py` as the narrowly fixed a063/P/E
  policy helper and integrated its effective-base result into `resolve_base`.
- The helper currently rejects a present record outside the fixed tuple, changed
  a063 planning-base bytes, malformed/duplicate JSON, invalid ancestry or source
  tree, dirty worktrees, S-to-H tracked source drift, bad committed evidence
  bytes/digests, planning history mismatch, and an E-to-S Go-path mismatch.
- `/tmp/a120-red-fixture-exit.txt` records the observed initial failure: the
  first skeleton accepted a deliberately empty E-to-S path ledger. The subsequent
  correction is covered by `/tmp/a120-focused-green-exit.txt` (exit 0, 36 tests).
- Two-mode `go list -deps -test -json` input enumeration and detached-HEAD
  rejection have been added after the first fixture. They still require their
  dedicated Git-fixture proofs.

Implemented: exact S/H/worktree Go source lock, symlink-path rejection, endpoint
inventories, fixed non-overwriting draft outputs, strict review/debt schema, B2
source/build-input fixtures, and checker integration. Acceptance remains pending
independent review, final SDD/gate commands, Manager verification, and archive.

The former B2 source-guard defect is resolved: filesystem traversal covers ignored
and ordinary untracked entries and Git path lists are NUL-delimited/lossless.

## Inventory and draft-generator increment

The immutable-target helper now obtains the ordinary modified-existing function
rows from Git commit objects without switching checkout, and records sorted
`file`, `function`, `revision`, and source SHA-256 fields. The same canonical
serializer supplies P-to-E and E-to-S payloads to the draft generator and the
validator. The draft CLI refuses any pre-existing ledger or record output.

`/tmp/a120-inventory-generator-fixtures-exit.txt` records exit 0 for real
temporary Git repositories covering missing, duplicate, wrong-hash and
wrong-revision rows, merge-parent and reverted-path history, deletion with a
base revision, and draft overwrite refusal. The combined ordinary checker run
is `/tmp/a120-inventory-generator-combined-exit.txt`, exit 0 (40 tests).

## Historical focused evidence (superseded)

- `/tmp/a120-final-helper-green.exit`: exit 0, 27 helper/fixture tests.
- `/tmp/a120-checker-e2e-final.exit`: exit 0, 31 checker tests, including real
  detached valid-adoption selection of E and missing-current-map failure.
- `/tmp/a120-check-analysis-red.exit`: retained exit 1 before the parser test fix.

The `31`-checker count and pending matrix below predate the later G2/G4/reference
and CLI additions; they are retained as historical command evidence only.

## Current final-source fixture matrix

The final Python source snapshot was re-read on 2026-09-06. Its hashes remain
`check_analysis.py` `f7d95a10cee85773189b4ee446a282eed7fb78021e361868029bda4ae5f1ebf2`
and `execution_baseline.py` `e9d8a55c718fe3568d09facb355e507c06c565f5bed915ef1b0f780d0104e1c3`;
the corresponding changed-function AST inventory is retained in
`analysis/post-edit-python/final-function-ast.json`.

`/tmp/a120-g2-final.exit` is `0` for the real ordinary dirty-Go fixture: it
first obtains the P-derived missing-map result, then accepts an actual
current-revision bundle without cleaning the worktree. `/tmp/a120-final-matrix-fixtures.exit`
is `0` for its final override/reference additions: ordinary `SDD_BASE_REF=P`
passes while `HEAD` and a different resolved commit fail, and a reference-only
change accepts the same P then rejects a referenced different base. These are
fixture-level proofs only; the final full SDD and repository command exits are
recorded separately and remain required for acceptance.

| Claim | Current focused evidence | Status |
| --- | --- | --- |
| helper schema, source lock, generator and B2 profiles | `/tmp/a120-final-helper-green.exit` (27 helper fixtures) | verified fixture evidence |
| valid E map, missing/stale current map and deleted base-revision map | real detached Git/Go checker fixtures | verified fixture evidence |
| ordinary P dirty-Go path and override assertion | `/tmp/a120-g2-final.exit`, `/tmp/a120-final-matrix-fixtures.exit` | verified fixture evidence |
| reference-only same-P acceptance and different-P rejection | `/tmp/a120-final-matrix-fixtures.exit` | verified fixture evidence |
| portable fixture branch initialization | `/tmp/a120-gstack-p2-cli.exit` under `init.defaultBranch=main` | verified fixture evidence |
| CLI ordinary/adoption/invalid distinction | `/tmp/a120-gstack-p2-cli.exit`, `/tmp/a120-gstack-cli-adoption.exit` | verified fixture evidence |
| final source branch inventory | `analysis/post-edit-python/final-function-ast.json` and `final-branch-test-map.md` | reviewed inventory; see stated per-branch limits |

No row above is final acceptance, an adoption approval, or a replacement for
the final full SDD, broad test, independent review, Manager, and archive gates.

## Current broad verification results

The final reviewed source completed these commands with exit `0`: `make test`,
`make test-seams`, `make test-race`, `make vet`, `go vet -tags
tossos_testseams ./...`, and post-fix `make sdd-test`. Their durable logs are
`/tmp/a120-broad-<name>.log` with matching `.exit` files, except the final SDD
run at `/tmp/a120-final-sdd-test.log` and `.exit`. Strict a120 and all-change
OpenSpec validation plus PM tracker generate/check also exited `0`. These are
verification results, not a gate/archive result; SDD freshness and Manager
readiness remain separate pending steps.
