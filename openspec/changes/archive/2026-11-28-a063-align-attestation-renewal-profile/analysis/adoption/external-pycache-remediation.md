# External Python cache remediation

Date: 2026-09-06

## Scope and removal

Only the existing generated cache directory below was removed:

```text
tools/logic-map/__pycache__
```

The directory contained the five observed generated `.pyc` artifacts. No Go source,
tests, OpenSpec maps, ledger, record, review documents, tasks, design, service,
timer, runtime files, or Git history were changed by this remediation.

Removal command:

```bash
find tools/logic-map/__pycache__ -depth -delete
test ! -e tools/logic-map/__pycache__
```

## Scoped checker run with an external cache

The cache prefix was created outside the worktree and passed only to the requested
checker invocation:

```bash
cache_prefix=$(mktemp -d /tmp/a063-external-pycache.XXXXXX)
PYTHONPYCACHEPREFIX="$cache_prefix" \
  python3 tools/logic-map/check_analysis.py \
    --change a063-align-attestation-renewal-profile
```

Resolved external cache prefix:

```text
/tmp/a063-external-pycache.j7RjSn
```

Python wrote project bytecode under that external prefix, including:

```text
/tmp/a063-external-pycache.j7RjSn/tmp/tossos-a063-execution-adoption/tools/logic-map/execution_baseline.cpython-312.pyc
/tmp/a063-external-pycache.j7RjSn/tmp/tossos-a063-execution-adoption/tools/logic-map/role_check.cpython-312.pyc
```

The checker exited `1`, as expected for the pre-bind state, and explicitly stopped
at the unbound execution-baseline record before deriving modified Go functions:

```text
[logic-map] cannot derive modified Go functions: invalid execution-baseline adoption: fatal: path 'openspec/changes/a063-align-attestation-renewal-profile/execution-baseline.json' exists on disk, but not in 'c727ad12a42dcd15c494c1997e92816edaf17b6b'
```

## Repository cache proof

After the run, both checks produced no output:

```bash
find . \( -type d -name __pycache__ -o -type f -name '*.pyc' \) -print | sort
git ls-files ':(glob)**/__pycache__/**' ':(glob)**/*.pyc'
```

Therefore no repository `__pycache__` directory or `.pyc` file remains, and no
tracked Python cache artifact exists. This report is the only file added by this
remediation; it makes no source-change claim beyond that explicit non-change scope.

## S comparison and report whitespace check

Against S `c727ad12a42dcd15c494c1997e92816edaf17b6b`, the following commands
produced no output for non-metadata changes:

```bash
git diff --name-only c727ad12a42dcd15c494c1997e92816edaf17b6b -- \
  . ':(exclude)openspec/**'
git ls-files --others --exclude-standard -- . ':(exclude)openspec/**'
```

The complete S-relative diff consists only of pre-existing OpenSpec metadata
paths plus this untracked OpenSpec evidence report. No non-metadata tracked or
untracked path differs from S.

The report was whitespace-checked with:

```bash
git diff --no-index --check /dev/null \
  openspec/changes/a063-align-attestation-renewal-profile/analysis/adoption/external-pycache-remediation.md
```

It emitted no whitespace diagnostics. Its exit status was `1`, which is the
normal `--no-index` indication that `/dev/null` and the newly added report differ.
