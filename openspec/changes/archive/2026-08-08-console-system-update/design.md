## Context

The console already fingerprints the executable at startup, detects when the
path has been replaced, and can `exec` the path on the same port with a
single-use browser handoff. Installation is the missing half. Development agents
currently create binaries in session-specific `/tmp` paths and ask the operator
to overwrite `~/.local/bin/tossctl` manually.

Installing a binary is not a config edit. It changes the program that will own
the next process and can therefore change all later behavior. The console must
not accept arbitrary paths, shell commands, or uploaded bytes.

## Goals / Non-Goals

**Goals:**

- Give the operator one settings-menu action to inspect, install, and restart
  into a staged development build.
- Restrict the candidate to `<current-executable>.candidate`.
- Validate module identity and platform without executing the candidate.
- Refuse while engine or verification work is active.
- Replace recoverably and retain `<current-executable>.rollback`.
- Keep all behavior behind the existing loopback session and CSRF gates.

**Non-Goals:**

- Building Go source from the browser.
- Executing arbitrary `/tmp` paths or shell commands.
- Replacing Homebrew, release downloads, or `tossctl update`.
- Automatically installing without a human click.
- Restarting or stopping the engine on the operator's behalf.

## Decisions

### D1. Fixed sibling candidate, no path input

The only candidate is `binstamp.SelfPath() + ".candidate"`. The rollback is
`binstamp.SelfPath() + ".rollback"`. `cmd/tossctl` injects an installer already
bound to those paths; the HTTP form contains no path, URL, or command.

Alternative rejected: store the Claude scratch path in config. It expires, is
session-specific, and would turn a config edit into arbitrary executable
selection.

### D2. Inspect and copy from no-follow file descriptors

`internal/localupdate` opens the candidate with no-follow semantics, verifies the
opened descriptor is regular and executable, and copies from that descriptor
into a same-directory prepared file. SHA-256 and `debug/buildinfo.ReadFile` are
computed from the prepared immutable copy, not by reopening the candidate path.
It rejects symlinks, non-regular/non-executable files, the wrong module path,
any Go main-package path other than
`github.com/JungHoonGhae/tossinvest-cli/cmd/tossctl`, and GOOS/GOARCH
mismatches. Checking only `BuildInfo.Main.Path` is insufficient because every
tool built from this repository has the same module identity; `BuildInfo.Path`
is the command identity. It does not run candidate code to ask what it is.

The view reports current/candidate path, size, UTC mtime, hash, Go version,
module version, GOOS, and GOARCH. Unknown or invalid metadata is an explicit
refusal, not an install button.

The install form carries only the displayed candidate SHA-256 as a concurrency
token. At commit the prepared copy must have that exact hash. Path, URL, command,
and uploaded bytes are ignored. If the candidate changed after the page was
rendered, installation is refused and the operator reviews the new facts.

### D3. Recoverable Unix replacement

Installation copies the current executable into a same-directory rollback
temporary file while the current path remains intact, sets mode 0755, syncs and
closes it, atomically publishes it at the rollback path, then syncs the
directory. Only after that durable rollback exists does it atomically rename the
prepared candidate over the current path and sync the directory. There is no
point at which the current path is absent.

If the candidate rename fails, the old current path is still intact. If the
post-replacement directory sync fails, the updater attempts an atomic restore
from the durable rollback copy and syncs the directory again. It reports both
the original and restoration errors and never requests relaunch on an uncertain
commit.

Non-Unix builds expose an unsupported result and no install button because the
running executable may be locked and the existing console re-exec contract is
also Unix-only.

Alternative rejected: call `install -m755` from the handler. A shell child loses
typed error boundaries and cannot provide the rollback transaction the console
promises.

### D4. Installation and restart are one authenticated operation

`POST /settings/system-update/install` is registered as a mutating route behind
session and CSRF. Before installation the handler takes an in-process
update/start mutex and refuses when:

- a verification run exists and is unfinished,
- the real engine exclusion cannot be acquired,
- the external verification marker is fresh or cannot be read reliably,
- no relaunch seam is wired,
- candidate inspection is not installable,
- the installed target is no longer the same regular executable fingerprinted
  when this console started.

The exclusion remains held through candidate preparation, the commit-time
verification-marker check, replacement, and relaunch request. The engine-start
and verification-start handlers use the same in-process mutex; after a
successful commit they remain refused until the old console exits.

The real journal-directory flock is an execution exclusion shared by the engine,
the updater, standalone `tossctl verify run`, and the console verification
starter. Every live verification acquires it before record/account/broker work
and holds it until runner cleanup completes. Therefore an updater that holds the
flock excludes a new external verification all the way through rename, and a
running external verification excludes update even when its advisory marker
cannot be written. The freshness marker remains a fail-closed supplemental
diagnostic at update commit; it is not the mutual-exclusion mechanism.

On success it renders the existing restart interstitial, requests same-port
relaunch, and logs the old/new hashes and rollback path to console output. On
failure it redirects to settings with the reason and keeps serving the old
process.

### D5. Staging is a build workflow, not an operator command

`make stage-local-update` builds `bin/tossctl` and writes it to
`$(TOSSCTL_INSTALL_PATH).candidate`, defaulting to
`$(HOME)/.local/bin/tossctl.candidate`. Agents can run this target after gates;
the human installs only by reviewing the settings panel and clicking.

## Risks / Trade-offs

- [Same-directory rollback consumes extra disk] → retain one file only and
  replace it on the next successful update.
- [Engine and verification markers are advisory] → engine, updater, and both
  verification entry points hold the same crash-released real flock for their
  complete execution; markers remain status/diagnostic signals only.
- [Candidate or installed target changes between display and commit] → bind the
  prepared bytes to the displayed hash and compare the current no-follow
  fingerprint with the console's startup stamp immediately before commit.
- [Crash during replacement] → publish and sync rollback while the current path
  still exists, then use one atomic rename over current; fault-test both rename
  and directory-sync restoration paths.
- [A malicious or wrong-command candidate with valid Go metadata] → require
  both the module root and the exact `cmd/tossctl` main-package path, keep the
  path non-selectable, never execute candidate code during inspection, and
  retain the human-confirmed local action. Release signing is separate scope.

## Migration Plan

Ship the new console and `localupdate` package with no candidate present. The
panel initially reports “staged candidate 없음” and changes nothing. Developers
stage a gated build, review its metadata in the browser, and click install. The
previous binary remains at `.rollback`.

Rollback is to stop the console and rename `.rollback` back to `tossctl`; the
update package itself never modifies config, credentials, account state, or
journal rows.

## Open Questions

None.
