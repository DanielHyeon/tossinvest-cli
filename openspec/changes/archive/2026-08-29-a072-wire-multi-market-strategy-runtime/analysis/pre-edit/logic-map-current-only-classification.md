# SDD analysis: current-only Go functions are not modified-existing evidence targets

Date: 2026-08-04
Change: `a072-wire-multi-market-strategy-runtime`

## Problem

The frozen a072 base predates several additive KR/US changes. `changed_existing_functions` parsed every
current function intersecting a diff hunk and required a Function Logic Map even when that qualified
function did not exist in the base file. This contradicted `docs/WORKFLOW.md`, which scopes the hard gate
to existing function logic, and produced 73 false targets dominated by newly added tests and helpers.

Advancing `base-commit.txt` would hide cumulative edits and is therefore rejected. Creating placeholder
maps for new functions would satisfy form while weakening the evidence signal.

## Contract

- A current function is a modified-existing target only when the same qualified function exists in the
  frozen base file and its current source range intersects the diff.
- Base functions whose old range intersects a deletion/replacement remain required with `revision=base`.
- Existing high-risk functions remain non-exempt.
- Additive high-risk functions may retain voluntary full maps; they are not deleted merely because the
  checker no longer misclassifies every new leaf/test helper.
- Invalid base, diff failure and base-file load failure remain fail closed.

## RED / GREEN evidence

RED: `test_new_function_in_existing_file_is_not_reported_as_modified_existing` constructed an existing
file with one base function and one newly added current function. Before the fix, `Added` appeared in the
required set.

GREEN: the current scan filters current functions through the base file's qualified-function set. The
test passes while modified-function, invalid-base, diff-failure, placeholder, branch and exemption
regressions continue to pass.

## Safety boundary

This changes only analysis-target classification. It grants no runtime, broker, approval, activation,
configuration or deployment authority. The persisted a072 base commit remains unchanged.
