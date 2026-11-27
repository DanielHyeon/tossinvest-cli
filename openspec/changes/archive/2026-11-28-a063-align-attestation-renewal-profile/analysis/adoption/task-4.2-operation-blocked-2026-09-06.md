# a063 task 4.2 operational record

Timestamp (KST): 2026-09-06T18:40:43+0900

## Result

**BLOCKED — no mutation performed.**

The required running-engine proof failed closed: zero running `tossctl engine`
processes were found (the direct probe exit was `1`). Therefore no explicit
engine `--config-dir` value exists to normalize and compare with the exact
console target. The approved target was normalized read-only to
`/home/daniel/.config/tossctl`; no profile was inferred or substituted.

The requested activation block was not entered. No backup was made, no file
was installed, no unit was disabled/enabled, no daemon reload occurred, and no
service was manually started, stopped, or restarted. No engine, console launch
argument, trading control, make command, order command, repository file, task
checkbox, or archive was changed.

## Preflight evidence (redacted)

| Check | Result | Exit |
| --- | --- | --- |
| Console target directory exists | pass | 0 |
| Console target is a symlink | no | 1 |
| User systemd bus (`systemctl --user show-environment`) | available | 0 |
| Candidate exists / is symlink | regular / no | 0 / 1 |
| Reviewed service template exists / is symlink | regular / no | 0 / 1 |
| Reviewed timer template exists / is symlink | regular / no | 0 / 1 |
| All three reviewed SHA-256 bindings | pass | 0 |
| Existing `tossctl`, service, timer target files | each regular, non-symlink | 0 / 0 / 0 |
| Service/timer `FragmentPath` | each exact expected target | 0 / 0 |
| Service/timer drop-ins | none | 0 / 0 |
| Timer enabled / active | yes / yes | 0 / 0 |
| Attestation service active state | inactive | 0 |
| Running engine profile proof | blocked: count 0 | direct process probe 1 |

Verified SHA-256 values:

- candidate: `d9ef7f5ecac62a9e93159ebfc5cf5cf50d1ac1ee93b68c7ff9d48704777f6a6b`
- service template: `0bf50972660fa14ece6953b1d43aa51fe2df407a51783c7ea0c6decb3100de96`
- timer template: `0db4e0cbfada2a59de4756747f17a3d7596d80dc7a0adafb3b3460849611edde`

No raw process command line, account data, session filename, credential,
record reference, or attestation value was emitted or retained. Process command
lines were read only transiently to classify the presence of an explicit
`--config-dir`; because no engine was present, none was retained.

## Survey inspection (read-only)

The source command surface inspection exited `0`. The `tossctl soak run`
process classification found count `0`; the direct process probe exited `1`.
Thus no same-profile survey is running. No shell command started a survey;
`engine run`, `--record`, record reset, and duplicate survey actions were not
used.

Task 4.3 needs a human action through the existing console survey control if a
survey is to be started, after a separate successful same-profile engine proof.
Its approved three-consecutive-future-local-calendar-day window begins only at
that successful actual installation timestamp; no historical result was used.

## Exact commands run and exit details

1. Read-only packet/workflow/memory and initial status inspection:

   ```bash
   rg -n -i -C 2 'a063|attestation|task 4\\.2|tossos-attest' /home/daniel/.codex/memories/MEMORY.md
   sed -n '1,240p' .claude/CLAUDE.md
   sed -n '1,260p' docs/WORKFLOW.md
   sed -n '1,260p' openspec/changes/a063-align-attestation-renewal-profile/analysis/adoption/task-4.2-4.4-operational-evidence-template.md
   sed -n '1,320p' openspec/changes/a063-align-attestation-renewal-profile/analysis/deployment-plan.md
   git status --short
   git status --porcelain=v1
   ```

   Exits: memory search `1` (no matching local memory entry); packet/workflow
   reads `0`; initial git status command `0`.

2. A read-only shell preflight ran these checks (all stdout was redacted to the
   result table above): `readlink -f -- "$HOME/.config/tossctl"`; `test -d`
   and `test -L` for the profile; `systemctl --user show-environment`; `test
   -f` and `test -L` for the reviewed candidate/templates and exact existing
   targets; the three-line `sha256sum -c -` binding shown in the approved
   deployment plan; `systemctl --user show` for `FragmentPath` and
   `DropInPaths`; `systemctl --user is-enabled --quiet`; `systemctl --user
   is-active --quiet`; and `systemctl --user show -p ActiveState --value`.
   A transient `/proc/<pid>/cmdline` scan classified `pgrep -f
   '[t]ossctl engine'` candidates without outputting their command lines.
   Shell wrapper exit: `0`; individual exits are in the table.

3. Read-only survey inspection:

   ```bash
   rg -n -i -C 3 'soak run|survey|record-renewal-status|already.*running|running.*survey' cmd/tossctl internal/console internal/soak > /tmp/a063-task42-source-inspection.log 2>&1
   pgrep -f '[t]ossctl soak run' >/dev/null
   ```

   Source inspection exit: `0`; direct survey probe exit: `1`; classification
   wrapper exit: `0`. The temporary source-inspection log contains repository
   source matches only, not runtime/process/account output.

4. Final direct probes and repository status:

   ```bash
   pgrep -f '[t]ossctl engine' >/dev/null
   pgrep -f '[t]ossctl soak run' >/dev/null
   git status --porcelain=v1
   ```

   Exits: engine probe `1`; survey probe `1`; git status `0`.

## Final repository status

`git status --porcelain=v1` exited `0` and remains dirty with pre-existing
changes. This operation did not write inside the repository; the only files
created by this operation are this report and the read-only source-inspection
log under `/tmp`.

Observed porcelain paths:

```text
 M cmd/tossctl/soak.go
 M cmd/tossctl/soak_test.go
 M docs/WORKFLOW.md
 M docs/operations.md
 M docs/pm/generated/00-master-tracker.md
 M docs/pm/generated/01-active-change-map.md
 M docs/pm/generated/02-release-readiness.md
 M docs/pm/portfolio/_registry.yaml
 M docs/pm/portfolio/features/FEAT-TOS-001.yaml
 M docs/pm/portfolio/stories/STORY-TOS-a063.yaml
 M internal/console/console_test.go
 M internal/console/data.go
 M internal/console/templates.go
 M internal/soak/attest.go
 M internal/soak/attest_test.go
 M openspec/changes/a063-align-attestation-renewal-profile/design.md
 M openspec/changes/a063-align-attestation-renewal-profile/proposal.md
 M openspec/changes/a063-align-attestation-renewal-profile/review.md
 M openspec/changes/a063-align-attestation-renewal-profile/specs/engine-safety/spec.md
 M openspec/changes/a063-align-attestation-renewal-profile/tasks.md
 M openspec/specs/sdd-workflow/spec.md
 M tools/logic-map/README.md
 M tools/logic-map/check_analysis.py
 M tools/logic-map/test_check_analysis.py
 M tools/sdd/sdd_doctor.py
 M tools/sdd/test_sdd_doctor.py
?? deploy/systemd/
?? docs/pm/portfolio/stories/STORY-TOS-a120.yaml
?? internal/soak/renewal_status.go
?? internal/soak/renewal_status_other.go
?? internal/soak/renewal_status_test.go
?? internal/soak/renewal_status_unix.go
?? internal/soak/renewal_status_unix_test.go
?? openspec/changes/a063-align-attestation-renewal-profile/analysis/
?? openspec/changes/a063-align-attestation-renewal-profile/execution-baseline.json
?? openspec/changes/a063-align-attestation-renewal-profile/issues.md
?? openspec/changes/a119-codex-session-handoff-and-gbrain-startup/analysis/
?? openspec/changes/a119-codex-session-handoff-and-gbrain-startup/design.md
?? openspec/changes/a119-codex-session-handoff-and-gbrain-startup/review.md
?? openspec/changes/a119-codex-session-handoff-and-gbrain-startup/specs/
?? openspec/changes/a119-codex-session-handoff-and-gbrain-startup/tasks.md
?? openspec/changes/archive/2026-09-06-a120-audit-legacy-execution-baseline/
?? tools/logic-map/execution_baseline.py
?? tools/logic-map/test_execution_baseline.py
```

## Independent review and Manager verification

- Adversarial review: **CLEAR for the block correctness**, retained as
  [`task-4.2-adversarial-review.md`](task-4.2-adversarial-review.md), SHA-256
  `205873dc8adf4a3201bcf379e2c44d807e6c53757ad480edc6d7be6c61f646dc`.
- Subsequent gstack review: **CLEAR for the block correctness**, retained as
  [`task-4.2-gstack-review.md`](task-4.2-gstack-review.md), SHA-256
  `ecaac5a21090a3b4234afb96c2faf1e49a35c133b9738b03be7fb8df8e2d684f`.
- Manager independently compared the retained operation report and reviews, confirmed that the
  running-engine probe is absent (`exit 1`) and that no mutable command entered the activation block.

Task 4.2 remains unchecked. Its approval cannot be consumed while the required same-profile engine
proof is absent; task 4.3 has not begun and no historical survey date may be reused.
