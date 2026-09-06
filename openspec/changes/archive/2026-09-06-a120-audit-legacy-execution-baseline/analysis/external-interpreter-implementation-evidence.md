# External interpreter amendment implementation evidence

Date: 2026-09-06. This record is limited to task 2.5 implementation and does
not replace the pending independent implementation/adversarial reviews or final
gate.

Pre-edit verification: `tools/sdd/sdd_doctor.py` SHA-256 was
`eb6c4ecb4d1f80c17a1514bd93ba7683b37f40d5c8e508aec5c89e95a6a7c26d`, matching
the frozen pre-edit `report` artifact before this edit. The previous post-edit
source hash was `9e09b3b781b377221f3cc66a41634c871098d6338eda9cc0d2bf48221c346b04`;
the lexical-boundary amendment hash is
`e038dce83e10cec6df49977a20c6bcb8dc5d873edae6bf81a95e08f2f7f04268`; its
AST inventory and branch-to-test map are in `analysis/post-edit-python/`.

Implementation facts:

- `SDD_PYTHON` is selected by environment membership. Absence retains the
  pre-existing local command and setup hint; every explicit invalid selection
  produces a required failure and skips the local probe.
- Explicit paths are checked against both the lexical checkout path and their
  strict resolved target. The dependency process receives an argument array
  beginning with exactly the selected interpreter and uses the existing bounded
  `command_status` timeout.
- Explicit mode reads one exact `typedb-driver==...` pin from
  `tools/sdd/requirements.txt` and compares the metadata-probed version.
- The doctor status contains mode, raw path, resolved target and dependency
  result. Adoption source validation, execution-base selection, services and
  `tools/sdd-history/refresh_indexes.py` were not changed.
- The strict-resolution diagnostic handlers catch `RuntimeError` and narrowly
  catch `ValueError` for defensive direct-helper inputs only. A real two-link symlink loop
  is a required `mode=external` invalid selection and does not execute a driver
  command; local diagnostic resolution is equally structured. The original
  report pre-edit artifact remains intact; the supplemental helper pre-fix AST
  evidence is `analysis/pre-edit-external-resolution-helpers.md`.

Boundary amendment facts:

- `_normalized_path` uses `Path(os.path.abspath(...))` to collapse dot
  segments without resolving candidate symlinks. Before candidate resolution,
  `_external_python` rejects a path within either the normalized supplied root
  or canonical `root.resolve()`, while retaining the separate
  resolved-target-outside check.
- Real symlink-alias root, supplied-root `..`, and candidate `..` regressions
  fail as invalid external selections with no typedb-driver probe. A wholly
  external interpreter control remains accepted.
- `_typedb_pin` treats `UnicodeDecodeError` as the existing structured
  cannot-read failure and skips the driver probe.
- `/tmp/a120-external-boundary-red.{log,exit}` retains RED exit `1`; the
  corresponding green files retain exit `0`.

Executable proof uses `/tmp/tossos-a120-external-sdd-venv`, created with `uv`
outside the checkout and installed from the unchanged pinned requirements. Its
metadata reports `typedb-driver 3.11.5`. The durable command logs and exit files
are `/tmp/a120-external-*.log` and `/tmp/a120-external-*.exit`.

After the test commands and AST/map updates, the CLI teammate attempted an
extra nested reviewer using a different model. Manager stopped that extra
process and the implementation context. No nested verdict is used as
acceptance evidence, and the implementation process exit alone is not treated
as completion. The assigned Terra adversary and subsequent separate gstack
reviewer remain the required review sequence. NUL handling above is defensive
helper coverage only; operating-system environments cannot contain NUL.

The subsequent dedicated gstack pass requested two test-only closures: a
successful dependency process returning the wrong version, and the same real
external-doctor/adoption fixture with a forbidden local virtualenv input. A
fresh bounded Terra context added both without production changes. Its final
focused doctor suite (15 tests), real integration case (1), and external-mode
full SDD suite (242 Python tests plus Go logic-map tests) each exited 0; exact
commands/logs/exits are `/tmp/a120-final-test-closures-*`. The initial integration
assertion had an over-escaped regex, which was corrected; that fixture error is
retained in the CLI event log and is not presented as a production RED result.
The dedicated adversary independently re-executed both added cases, exit 0.

Final test SHA-256 values are
`43fca70980b7bd24c8cc9931eff04bea2c2d78788dd301f84dcd22b274c2369a`
for `tools/sdd/test_sdd_doctor.py` and
`85e0de6a38ebbc5e68ea00781c314bc01db416fe44742a30d5df74c312b73430`
for `tools/logic-map/test_check_analysis.py`. Doctor source is unchanged from
the boundary amendment hash above.
