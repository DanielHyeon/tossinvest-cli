# a120 adversarial implementation review — provisional

Date: 2026-09-06
Reviewer: independent adversarial review
Status: **WIP only — no implementation verdict or completion approval.**

The current WIP commit is intentionally incomplete while the implementation
teammate adds the inventory/generator work. This record covers only concrete
behavior already present in `tools/logic-map/execution_baseline.py`; it does not
restate not-yet-implemented planned work as a new defect. The disclosed
skeleton-before-RED sequencing deviation remains a process fact: the later
fixture is useful functional proof, but it is not evidence that implementation
started RED-first.

## Implemented-behavior findings

### I1 — two review inputs need not be distinct or change-local

`validate()` reads `adversarial_review_path` and `gstack_review_path`
independently but never requires different paths. A record can point both fields
to the same committed file and SHA, satisfying the current code while violating
the adopted requirement for two distinct review passes. It also accepts ledger
and review paths anywhere in the repository, although the contract places the
ledger under the a063 change analysis directory and intends adoption evidence to
be bound to that change.

Required regression: a record that reuses one review file for both roles fails;
ledger/review paths outside the fixed a063 change (and ledger outside its
`analysis/` directory) fail; valid distinct a063-local files pass.

### I2 — inherited-history debt disposition is currently unchecked

`inherited_history_disposition` is required as a key but no accepted value is
checked. Therefore a record can contain `"completed"`, `"waived"`, or a false
pre-edit claim and still select E if its other current checks pass. That defeats
the policy's explicit honesty boundary.

Required regression: only the exact approved debt disposition is accepted;
`completed`, `waived`, empty, and arbitrary values fail before E is returned.

### I3 — evidence parent directories can escape through a symlink

`_regular_committed()` uses `lstat()` only on the final path. A path such as
`openspec/changes/a063.../analysis/ledger.json` can traverse an intermediate
`analysis` symlink to an external regular file; the final `lstat` then succeeds.
This violates the repository-contained/no-symlink evidence requirement even
though final-file symlinks are rejected.

Required regression: an intermediate symlink in the record, ledger, or review
path fails. Implement component-by-component `lstat` from repository root and
require the resolved path to remain under root before reading either worktree or
H-object bytes.

### I4 — `S..H` filenames are parsed lossily in the implemented Go-path guard

The source guard derives `expected_paths` with `git diff --name-only` and
`splitlines()`. Git permits a tracked filename containing a newline. Such a
filename is split into synthetic paths, so the ledger equality and fixed path
allowlist are no longer evaluated over Git's real path set. The immutable
history code correctly uses `-z`; the E..S guard should use the same NUL-delimited
approach and reject non-UTF-8 filenames explicitly.

Required regression: a newline-containing E..S Go path fails deterministically
rather than being split; ordinary valid paths continue to enumerate exactly.

### I5 — the metadata scan does not establish coverage of ordinary untracked files

The helper currently relies on one `git ls-files --others --ignored
--exclude-standard -z` call. That invocation's ignored-file selection does not
prove enumeration of ordinary untracked files, so an untracked C/assembly/embed
asset can escape the metadata allowlist and subsequent candidate/input overlap
test. The frozen contract requires all untracked **and** ignored files to be
considered.

Required regression: add one ordinary untracked `.c` or embed asset and one
ignored counterpart; both must fail. Implement separate Git queries with
documented semantics or a repository filesystem scan (excluding `.git`) that
classifies every non-tracked entry before the metadata allowlist is applied.

### I6 — a broken record symlink is treated as an absent record

`record_path.exists()` follows symlinks. A broken
`execution-baseline.json` symlink consequently returns false and causes
`validate()` to return `None`, selecting ordinary P behavior instead of
fail-closing a present-but-invalid exception artifact. Use `lstat` to distinguish
absence from any final-path symlink or other non-regular entry before the
ordinary-path return.

Required regression: final record symlink, including a broken symlink, fails;
only a genuinely absent path preserves ordinary P behavior. Also add a tracked
Go symlink-at-S regression: source entries accepted as the snapshot must be
regular files, not merely unchanged symlinks.

### I7 — the current RED fixture can pass for the wrong reason

`test_red_ledger_without_execution_to_source_inventory_is_rejected` creates no
`go.mod`, uses `kept.go` outside the fixed E..S allowlist, and accepts any
`AdoptionError`. Once `go list` and the source-path guard run, either earlier
failure can satisfy the assertion even if the intended missing inventory guard
is removed. It is therefore not proof of that guard.

Required regression discipline: construct one valid detached positive fixture
with a module, an allowlisted a063 source path, valid record/ledger/reviews, and
passing source guard. Each negative test must mutate exactly one condition from
that fixture and assert its specific failure category. Apply this to source
guard, inventory, history, review binding, and ordinary-path tests; generic
`assertRaises(AdoptionError)` is insufficient for the key policy guards.

## Review continuation

After the implementation teammate declares all planned fixture, inventory,
generator, and integration work ready, independently run the actual focused
suite and inspect no-record ordinary behavior, malformed-adoption failure,
source guard, and E-based complete map derivation before replacing this
provisional status with a final verdict.

## Frozen-contract implementation matrix — bounded read-only pass

This pass deliberately does not re-review the B2 source-guard implementation
while its fixture work is in progress. “Present” means visible in the current
WIP code and tests, not independently executed proof.

| Frozen requirement | Current observation | Status / required proof |
| --- | --- | --- |
| Exact one a063/P/E envelope and P base-file binding | `validate()` has fixed constants, fixed record location, and checks the committed base file before E selection. | Present; require isolated wrong-change/P/E/base-file mutation fixtures at final review. |
| Strict record and ledger schema plus honest inherited debt | Record exact-key/type validation is present. Manager already identified bool-as-`1`, ledger unknown-key, and disposition validation gaps. | Open under Manager findings; no independent duplicate finding here. |
| Complete P..E and E..S immutable inventories | `range_payload()` now derives endpoint inventories through immutable-target extraction and validation compares range payloads. | Present in code; final review needs a valid full adoption fixture with one omitted, duplicate, hash, revision, and base-deletion mutation each. |
| Non-overwriting, repository-contained draft generator | `draft()` checks existing output then writes canonical drafts, but accepts caller-selected paths and writes before proving they are inside the repository/current a063 locations. A malicious absolute or `..` output can therefore be mutated before `relative_to(root)` fails. | **Open G1.** Validate both output paths before any mkdir/write: regular, non-symlink, repository-contained, fixed a063 record and analysis-ledger locations (or remove override paths). Add outside-root and parent-symlink no-write fixtures. The separately known broken-link overwrite finding remains open too. |
| Ordinary no-record checker behavior | Helper-only test asserts `validate()` returns `None` for an absent record. | **Open G2.** Add a `check_analysis.resolve_base`/`check` fixture proving an ordinary change selects its persisted P, accepts its existing local/reference semantics, and remains usable on a dirty worktree. |
| Valid adoption selects E and requires ordinary full maps | `resolve_base()` receives `effective_base` from the helper, but current visible tests exercise helper validation and immutable inventory functions rather than `check_analysis.check()` with an E-based complete bundle set. | **Open G3.** Add end-to-end fixture: valid detached a063 adoption plus complete E..S maps passes checker; remove one map or use stale current/base revision and checker fails. Verify P..E rows do not silently become required maps. |
| `SDD_BASE_REF` remains validation-only | Code compares the override to selected `effective` base. | **Open G4.** Test ordinary P accepts only P; valid adoption accepts only E; P/HEAD/another SHA fails under adoption; an invalid record cannot use `SDD_BASE_REF=E` to select E. |
| I1/I2/I3 claimed fixes | Current code requires distinct review paths, constrains evidence under current change analysis, checks exact debt string, and walks evidence components with `lstat`. | Code present; final fixture pass must cover each specific failure, including intermediate—not only final—symlink. |
| I7 valid-positive mutation discipline | `_valid_adoption()` now has a module, an allowlisted source file, detached H, and a specific E..S inventory mutation assertion. Manager reports the implementation-side valid-pair, binding, and interim-B2 commands exited 0; `/tmp/a120-valid-inventory-pair-exit.txt` records one such exit. The old generic RED fixture still exists but is no longer the sole stated proof. | Implementation evidence is distinct from this review. Independently execute and inspect the valid-positive path before crediting it in the final verdict. |
| Legacy generic RED fixture truthfulness | `test_red_ledger_without_execution_to_source_inventory_is_rejected` still has the old outside-allowlist/no-module construction and generic `assertRaises`. With the current stricter debt-string validation it can fail before the named E..S inventory condition. | **Open G5.** Remove it, rename it to the actual earlier rejection it demonstrates, or replace it with a one-field mutation of `_valid_adoption()` that asserts the inventory-specific error. Do not retain its current RED/docstring claim. |

No completion conclusion follows from this matrix. The final implementation
review remains blocked on G1--G5, Manager's schema/draft findings, the in-flight
B2 source-guard fixture pass, and actual independent command results.

## B2 source-guard focused review — 2026-09-06

Scope: only `untracked_filesystem_entries`, `nul_paths`, `go_inputs`, source-S
symlink handling, and the focused scan fixture. This is not a whole-suite or
final adoption verdict; schema/envelope work was moving separately.

### Independent command evidence

Run from `tools/logic-map`:

```text
python3 test_execution_baseline.py \
  ExecutionBaselineUnitTests.test_filesystem_scan_metadata_and_source_mutations
exit: 0

.
----------------------------------------------------------------------
Ran 1 test in 2.040s

OK
```

The command proves the current fixture's `.codegraph/cache.json` positive case
and its ordinary untracked C file, metadata symlink, executable cache file, and
nested `__pycache__` negative cases. It does not prove the full frozen B2
contract.

### B2 findings and required regressions

| Area | Implemented behavior | Review result |
| --- | --- | --- |
| Untracked enumeration | `untracked_filesystem_entries()` scans the filesystem without following directory links and excludes `.git`; it catches ordinary untracked files, closing I5 in code. | Focused fixture proves one ordinary C file only. Add an ignored counterpart and an ordinary untracked embed asset; each must fail through the same guard. |
| NUL filenames | `nul_paths()` uses `-z` and strict UTF-8 decoding for Git-derived paths. | Correct implementation direction, but no fixture covers newline or invalid-byte E..S paths. Add both, including rejection without line splitting. |
| Allowed generated state | `_allowed_metadata()` limits direct `.pyc` and rejects executable/source-object suffixes. | Only `.codegraph/cache.json` has a positive fixture. Add positives for `.sdd/index-state.json`, GBrain state, each permitted history location, direct `tools/.../__pycache__/*.pyc`, and direct `auth-helper/.../__pycache__/*.pyc`; retain arbitrary/nested cache failures. |
| Actual Go inputs | `go_inputs()` runs untagged and `tossos_testseams` `go list -deps -test -json`, including the installed source/embed fields. | No focused assertion proves both invocations occur, an input overlap fails, or either enumeration failure fail-closes. Add an embed/input overlap fixture, a seams-only input fixture, and per-mode command/JSON failure fixtures. |
| S source symlinks | `ls-tree -r -z S` rejects a Go symlink. | No fixture yet. Add a tracked `.go` symlink in S and assert the source-snapshot error. |
| Tracked Go working-tree mode/bytes | The implementation relies on clean `git diff` plus the S..H path restriction; it does not directly compare each tracked S Go entry's Git mode/blob with the working-tree `lstat` mode/bytes. | **Concrete B2 defect.** With `core.filemode=false`, changing `internal/soak/attest.go` from `0644` to `0755` left `git diff --quiet` at exit 0 and `validate()` accepted the fixture (independent reproduction on 2026-09-06). This violates the explicit S mode lock. Enumerate all S tracked Go entries with NUL-safe Git tree data; require the same path and Git mode at H, then require each working-tree entry to be a regular non-symlink file with the matching Git executable mode and blob bytes. Add a `core.filemode=false` chmod regression and a content/substitution regression. |

The first five rows identify missing test evidence for reviewed behavior; the
tracked-mode/bytes row is an actual bypass. None is a final approval, and all
must be resolved by the successor implementation teammate before the
independent final run.

## Helper re-review — 2026-09-06

This pass covers only the stable `execution_baseline.py` helper and its direct
test file. It does not clear the moving `check_analysis` ordinary/adoption
integration, documentation, or final gate.

### Independent execution

```text
cd tools/logic-map && python3 test_execution_baseline.py
exit: 0
Ran 26 tests in 19.305s
```

I also repeated the prior mode bypass directly in a fresh valid detached
fixture: with `core.filemode=false`, the `0644 -> 0755` chmod left
`git diff --quiet` at exit 0, but validation now rejected it with
`source snapshot Go worktree mode mismatch: internal/soak/attest.go`.

### Helper results

`go_tree_entries()` and `verify_source_go_lock()` now compare the S and H Go
trees and compare each worktree Go file's regular-file state, executable bit,
and immutable H blob. This closes the concrete mode bypass. The helper tests
also now cover exact integer schemas (including JSON `true`), unknown ledger
keys, fixed debt disposition, duplicate keys, record/evidence digest
mutations, source/base substitutions, parent and final evidence symlinks,
fixed non-overwriting generator outputs and rollback of its own partial
ledger. The obsolete generic ledger RED fixture is gone.

The remaining helper-level gaps are bounded regression coverage rather than a
demonstrated bypass:

1. Task 2.1a names untracked/ignored Go, C, assembly, header, and embed inputs.
   The direct fixture covers C and assembly plus ignored C/assembly/embed, but
   has no untracked or ignored `.go` and header mutation.
2. The allowed-metadata positive fixture covers `.codegraph`, index state,
   direct event JSONL, and both direct Python cache roots. It does not exercise
   GBrain state, checkpoints, refresh queue, or index-refresh lock, despite
   each being a separate fixed allowance.
3. `go_inputs()` is unit-called with and without `tossos_testseams`, and the
   helper unions both calls, but no `validate()` fixture makes the two profiles
   return distinct inputs or proves either profile's failure blocks the full
   adoption path.

These must be closed before final helper credit. They do not change the
separate G2--G4 end-to-end checker requirements.

## Final code-pass matrix — 2026-09-06 (pending E2E additions)

Independent execution in this worktree passed `27` helper tests and `31`
checker tests; `git diff --check` over the changed logic-map files also passed.
Those counts do not substitute for an absent scenario.

| Item | Disposition after code pass | Evidence / remaining requirement |
| --- | --- | --- |
| I1, I2, I3 | Clear in helper scope | Exact review-path/debt checks and component `lstat` are present; the valid-fixture mutations and symlink cases are covered. |
| I4, I5, I6 | Clear in helper scope | NUL parsing, filesystem enumeration, source-tree/worktree lock, source Go symlink rejection, and the direct `core.filemode=false` regression all ran independently. |
| I7 / G5 | Clear | The generic invalid ledger fixture was removed; its replacement starts with a valid detached fixture and asserts the inventory-specific error. |
| B2 source guard | Clear in helper scope | The 27-test run covers ordinary and ignored Go/C/assembly/header/embed files; every fixed metadata root; both Go enumeration profiles and their failures. It also covers source bytes, mode and symlink substitution. |
| G1 generator safety | Clear in helper scope | Fixed output paths are validated before writes; outside and parent/broken target paths fail, and a record-write failure removes only its own partial ledger. |
| G3 valid E selection / missing current bundle | Partial | A real detached adoption fixture passes with an E-based map and fails after removing it. **Still open:** real checker fixtures for stale E-based current evidence and a deleted E function requiring a base-revision bundle. |
| G2 ordinary no-record | **Open / blocking** | `validate()` returning `None` and mocked `check()` tests do not prove real persisted-P ordinary checking, dirty-worktree usability, or unchanged reference semantics. Add a real Git checker fixture. |
| G4 `SDD_BASE_REF` | **Open / blocking** | The only environment test mocks `subprocess.run`. Add real valid-adoption cases where only E succeeds and P/HEAD/other fail; add real ordinary P acceptance/rejection; prove an invalid record with `SDD_BASE_REF=E` fails rather than falling back. |
| Reference conflict | **Open / blocking** | No real fixture exercises `function-logic-reference.txt` against adoption E or ordinary P. Add same-base acceptance and conflicting-base rejection through `check()`. |

The former helper coverage gaps are now closed by the current 27-test suite.
The G2/G4/reference and remaining G3 cases are required task-2/3 evidence;
there is no final implementation clearance until they are real, non-mocked
checker scenarios and the separate documentation/gate review is complete.

## Final-fixture regression found — 2026-09-06

The earlier checker green result predated the expanded E2E fixture. Independent
execution of the current files gave helper `27/27`, but checker `34/36`:
`test_valid_adoption_uses_e_and_requires_complete_current_bundle` and
`test_real_adoption_sdd_base_ref_accepts_only_e_and_invalid_record_never_falls_back`
both failed with a stale AST hash. `_adoption_with_complete_bundle()` extracted
the AST at E (source version 2) before committing S/H at version 3, then used
that E AST as a current-revision bundle. The implementation must extract the
current AST after S for non-deleted functions; the deleted case alone needs the
E/base-revision AST.

The new ordinary dirty-worktree fixture is also insufficient for G2: it dirties
only a non-Go note and uses an exemption with no P-based required function. Add
a real no-record fixture with a Go function modified *uncommitted* after P:
first prove the P-based missing-map error, then create a current-revision bundle
and prove `check()` succeeds while the Go worktree remains dirty. Until both
changes are independently green, this report remains blocked.

## Post-fix final pass — 2026-09-06

Independent current execution now passes helper `27/27` and checker `36/36`.
The ordinary fixture is a real Git P-to-uncommitted-Go path: it first observes
the missing P-based map obligation and then accepts a current-revision bundle
without cleaning the Go worktree. The adoption fixture now extracts the S AST
for current evidence, while its deleted-function variant retains E/base
evidence. Valid E selection, missing map, stale hash, invalid-record no-fallback
and local-reference coexistence rejection all execute through `check()`.

Two final scenario claims still lack direct real-fixture proof:

1. The ordinary fixture never sets `SDD_BASE_REF`. It must accept the exact P
   override and reject HEAD/another commit, rather than relying on the prior
   mocked resolver test.
2. The reference fixture has both local maps and a reference, so it correctly
   fails before reading the referenced change. It does not prove an empty-local
   reference accepts a target with the same P base and rejects a target with a
   different base at the exact-base comparison.

These are narrow remaining G4/reference coverage blockers, not a demonstrated
implementation bypass. Do not mark this adversarial review clear until the
two real fixtures are independently green.

## Code-level final verdict — CLEAR (2026-09-06)

The two remaining scenarios were added and independently exercised. Current
commands and results:

```text
cd tools/logic-map && python3 test_execution_baseline.py
27 tests: OK

cd tools/logic-map && python3 test_check_analysis.py
37 tests: OK

python3 -m unittest \
  test_check_analysis.CheckAnalysisTests.test_ordinary_no_record_uses_p_and_allows_dirty_worktree \
  test_check_analysis.CheckAnalysisTests.test_real_reference_requires_the_same_p_base
2 tests: OK
```

The ordinary real-Git path has an uncommitted Go edit after P, proves the
missing-map failure, accepts the matching current bundle with
`SDD_BASE_REF=P`, then rejects `HEAD` and the resolved distinct commit with the
exact persisted-base error. The reference-only path has no local map, accepts
the same P target, then rejects a different target base with the exact-base
conflict. The adoption path independently covers E-only overrides, invalid
record no-fallback, stale current evidence, missing current evidence, and
deleted E function base-revision evidence.

All I1--I7 and G1--G5 findings are clear for the implementation code and its
focused helper/checker fixtures. This is **not** a claim that broad Go tests,
OpenSpec validation, documentation/PM review, `make sdd-check`, `make gate`,
or the requested separate gstack review have run or passed; those are separate
completion evidence owned outside this code-level adversarial pass.
