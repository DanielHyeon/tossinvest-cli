# Proposed external SDD tooling environment (not executed)

The a120 amendment requires the interpreter to stay outside this worktree.
The following is a proposed, reviewable setup only; this preparation did not
create the directory, install packages, modify global packages, or change any
service/config/lock.

```bash
EXTERNAL_SDD_VENV=/tmp/tossos-a120-sdd-tools/.venv
uv venv --python 3.13 "$EXTERNAL_SDD_VENV"
uv pip install --python "$EXTERNAL_SDD_VENV/bin/python" -r tools/sdd/requirements.txt
SDD_PYTHON="$EXTERNAL_SDD_VENV/bin/python" make sdd-check
SDD_PYTHON="$EXTERNAL_SDD_VENV/bin/python" make gate CHANGE=a120-audit-legacy-execution-baseline
```

Pinned input: `tools/sdd/requirements.txt` contains only
`typedb-driver==3.11.5`. The later implementation must make `SDD_PYTHON` affect
only `sdd_doctor`; `make` forwarding is an environment mechanism, not an
execution-baseline selector. Creation and any final gate run require Manager
freeze and their own actual evidence.
