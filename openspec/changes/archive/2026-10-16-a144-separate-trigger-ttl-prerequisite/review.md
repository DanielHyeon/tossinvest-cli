# Review — a144-separate-trigger-ttl-prerequisite

## Scope

Reviewed the one-predicate exception in `M0ExactPrerequisites` and its offline
tests. It excludes only `idempotency-ttl-edge` beside the pre-existing trigger
step. No operating configuration, live verification, broker credential, or
order path changed.

## Independent reviews

- Adversarial review: APPROVE —
  `/home/daniel/.codex/.chatgpt-projects/g-p-6a6a11adcde48191ab0a385cd9e910c4/outputs/ttl-prerequisite-implementation-adversarial-review.md`.
- GStack review: PASS, per reviewer handoff. It required the TTL firing-fake,
  all non-TTL verdict matrix, and existing CLI factory-boundary coverage now
  present in the final test set.

## Evidence

- Focused RED recorded before the predicate change: TTL skipped, failed,
  refused, missing, and nonterminal latest records were rejected as M0
  prerequisites.
- Focused GREEN and package regression: `go test ./internal/verifylive
  ./cmd/tossctl -count=1` passed (17.461s and 61.905s).
- `openspec validate a144-separate-trigger-ttl-prerequisite --strict --no-interactive`
  and `git diff --check` passed.
- Final analysis evidence is under `analysis/function-logic/`; its checker
  passed before repository-level SDD checks.

## Remaining gate state

The original dirty worktree could not pass `sdd-check` because an unrelated
unstaged change in `.claude/CLAUDE.md` made the mirrored safety bootstrap drift.
It was preserved. A detached clean worktree at `1f4749e4` ran the required
offline checks instead:

- `go test ./internal/verifylive ./cmd/tossctl -count=1`: PASS.
- `make sdd-sync`: PASS for the hard CodeGraph index.
- `make sdd-check`: PASS after the standard ignored local `.sdd/.venv` setup.
- Initial final `make gate` reached `test-race` but failed only because Go VCS
  stamping ran `git status` from `/tmp` for the detached worktree and exited
  128. A focused race proof then passed with the process-scoped
  `GOFLAGS=-buildvcs=false` setting.
- The final `make gate CHANGE=a144-separate-trigger-ttl-prerequisite` passed
  11/11 with that same process-scoped setting: exit 0 and strict validation
  49 passed, 0 failed. It disables Go build VCS stamping only; no `go env -w`,
  source edit, or operating-setting change was made.

The final gate receipt is
`/tmp/a144-final-gate-20261017/isolated/make-gate-final-buildvcs-false.log`
with its exit file beside it. The sync wrapper returned zero but its GBrain
import reported 12,097 advisory parse skips. This change does not claim that
every advisory index succeeded; the hard CodeGraph freshness check passed.
Raw receipts are retained outside the repository at
`/tmp/a144-final-gate-20261017/isolated/`. The adversarial and GStack reviews
passed after this final receipt, and Manager confirmed the final gate evidence
before authorizing formal archive. This archive preserves the earlier failed
receipt as history; it does not replace it with a claim that the advisory
GBrain import passed.
