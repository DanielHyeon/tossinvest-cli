# Post-edit Python Branch Test Map: `tools/sdd/sdd_doctor.py`

Source SHA-256: `e038dce83e10cec6df49977a20c6bcb8dc5d873edae6bf81a95e08f2f7f04268`.
The direct AST inventory is [sdd-doctor-external-interpreter-ast.json](sdd-doctor-external-interpreter-ast.json).

| Function / branch | Contract | Executable proof |
|---|---|---|
| `_driver_status` 159 true | Absent `SDD_PYTHON` retains the local `.sdd/.venv/bin/python` probe and setup hint. | `test_absent_override_preserves_local_venv_command`; `test_absent_override_keeps_local_setup_hint` |
| `_external_python` 134/137/153 | Explicit empty, relative, missing, directory and non-executable values fail before a driver probe or local fallback. | `test_explicit_invalid_values_fail_without_local_fallback` |
| `_external_python` 139–152 | Dot segments are collapsed with `Path(os.path.abspath(...))` without resolving the candidate link. The candidate then fails before resolution if it is inside either normalized supplied root or canonical root; resolved targets still must be outside. | `test_both_path_boundaries_and_symlink_directions`; `test_lexical_boundary_normalizes_alias_and_dot_segments_before_resolving_links`; preserved RED/GREEN logs `/tmp/a120-external-boundary-{red,green}.log` with numeric `.exit` files |
| `_resolved_detail` 103 and `_external_python` 147 exception paths | A real two-link symlink loop is reported as an unresolved explicit invalid selection without a driver probe; the local diagnostic path also remains structured rather than raising. | `test_symlink_loop_is_an_invalid_external_selection_and_safe_local_diagnostic`; preserved RED/GREEN logs `/tmp/a120-external-loop-{red,green}.log` with numeric `.exit` files |
| `_typedb_pin` 111–129 | Explicit mode requires one exact `typedb-driver==version` requirement; unreadable/malformed UTF-8, missing, malformed, multiple and annotated/nonexact values fail before probing. | `test_explicit_mode_requires_exact_single_pinned_driver`; `test_explicit_mode_malformed_utf8_requirements_fails_without_driver_probe` |
| `_external_python` `ValueError` path | Direct helper calls containing NUL report a structured invalid selection. This defensive fixture is unreachable through real `os.environ`, which cannot contain NUL. | `test_nul_path_is_defensive_direct_helper_input_not_an_os_environment_case` |
| `_driver_status` 196 | A missing, failed or mismatched selected interpreter result fails with expected and observed values; it does not choose local Python. | `test_explicit_mode_rejects_missing_and_mismatched_driver` |
| external success / adoption | A real external venv with the pinned driver succeeds while a real isolated Git adoption fixture has no `.sdd/.venv`; `validate()` remains real, not mocked. | `test_real_adoption_without_local_venv_accepts_external_doctor_probe` with explicit external `SDD_PYTHON` |

The map is Python-tooling evidence only. It records the new doctor selection
branches and does not reclassify the unchanged adoption validator, source guard,
execution-base selection, services, or advisory `refresh_indexes.py` behavior.
