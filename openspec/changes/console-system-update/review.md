# Proposal-freeze review

- Change: `console-system-update`
- Base: `676f8aef0f45140bd57ab16d435b5b3c2178aa02`
- Review date: 2026-07-30
- Reviewer: independent adversarial Eng review (`proposal_review`)
- Initial verdict: **HOLD**

## Findings and decisions

### P1 — Inspection and installation were vulnerable to candidate TOCTOU

Accepted. Candidate bytes are copied from one no-follow descriptor into a
prepared file, inspected there, and compared with the SHA-256 the operator
reviewed. A path replacement cannot select different installed bytes.

### P1 — Moving current to rollback created a missing-executable crash window

Accepted. Rollback is a durable copy published while current remains intact.
Candidate installation is one atomic rename over current. Rename and directory
sync failures have explicit restoration behavior and fault tests.

### P1 — Idle checks were check-then-act and failed open on unknown state

Accepted. Update and in-process engine/verification starts share an exclusion;
the updater holds the real engine flock through commit and checks strict external
verification evidence at commit time. Missing means idle; unreadable or
otherwise indeterminate evidence refuses.

### P1 — Installed-target provenance was not guarded

Accepted. Immediately before commit the updater no-follow inspects and hashes
the current target and requires equality with the console startup fingerprint.
Drift tells the operator to restart the console.

## Freeze verdict

**GO after strict validation of these accepted changes.** No P0 finding remains.

## Implementation evidence

- `go test ./internal/localupdate -count=1`: pass, including no-follow
  inspection, reviewed-hash binding, current fingerprint drift, ordered
  rollback publication, rename failure and directory-sync restoration
- `go test ./internal/console -count=1`: pass, including authenticated install,
  same-port relaunch, no path/command input, refusal paths and concurrent start
  serialization
- `go test -race ./internal/localupdate ./internal/console ./cmd/tossctl`: pass
- `make test`, `make lint`, `make validate`: pass
- `make -n stage-local-update
  TOSSCTL_INSTALL_PATH=/tmp/tossos-stage-test/tossctl`: stages only the fixed
  `.candidate`; it does not overwrite current or restart
- both post-edit Function Logic Map checks: pass

## Independent implementation review — round 1

- Review date: 2026-07-30
- Reviewer: independent `proposal_review` context
- Verdict: **NO-GO**
- P0 findings: none
- P1 findings: two

### P1 — external verification exclusion remained TOCTOU/fail-open

The updater checked the advisory verification marker before commit, but a
standalone verification could begin after that check and before executable
rename. Marker creation was intentionally nonfatal to verification, so it could
not be the hard exclusion. The reviewer required live verification and update to
hold the same real cross-process lock for their complete critical sections.

Resolution: standalone `runVerifyRun` and `consoleVerifyStarter` now acquire the
same journal-directory kernel flock as engine/update before record, account, or
broker work and defer release through runner cleanup. The updater already holds
that flock through install and relaunch request. A RED regression reached the
broker under contention; GREEN refuses both entry points with zero broker builds.
The advisory marker remains supplemental and is no longer claimed as exclusion.

### P1 — same-module non-tossctl executable was accepted

The validator checked only `BuildInfo.Main.Path`, which is the repository module
root shared by `tossctl`, `tools/boundarymap`, and other commands.

Resolution: validation now also requires
`BuildInfo.Path == github.com/JungHoonGhae/tossinvest-cli/cmd/tossctl` and exposes
the command path in settings. The regression builds a real `cmd/tossctl` current
binary and a real same-module `tools/boundarymap` candidate; `New` accepts the
former and inspection refuses the latter without executing it.

Post-resolution evidence:

- both new RED→GREEN regressions pass
- `go test ./internal/localupdate ./internal/console ./cmd/tossctl -count=1`: pass
- `go test -race -count=1 ./internal/app/engine ./internal/localupdate
  ./internal/console ./cmd/tossctl`: pass
- tagged actual CLI assembly race regression: pass
- journal crash suite: pass
- both Function Logic Map checks and strict OpenSpec validations: pass

## Independent implementation review — round 2

- Review date: 2026-07-30
- Reviewer: independent `proposal_review` context
- Verdict: **GO**
- P0/P1 findings: none

The reviewer confirmed that both verification entry points acquire the same
`engineJournalDir(root)` flock before record/account/broker work and that defer
ordering retains it until runner, advisory-marker, and recorder cleanup finish.
The updater retains the same flock through install and relaunch request, while
the console's `activityMu` serialization remains intact.

The reviewer also confirmed that production validation requires the exact
`cmd/tossctl` `BuildInfo.Path`, and that the regression accepts a real tossctl
current binary while refusing a real same-module `tools/boundarymap` candidate.
Independent targeted, uncached race, `enginelock` race, strict OpenSpec, and
`git diff --check` reruns all passed.
