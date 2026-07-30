## ADDED Requirements

### Requirement: Settings exposes one fixed staged system update
The authenticated settings page SHALL display the running executable and the
single staged candidate at `<running-path>.candidate`. The console SHALL accept
no path, URL, command, or uploaded executable from an HTTP request. It SHALL
display file size, UTC modification time, SHA-256, Go/module metadata, and
GOOS/GOARCH for an inspectable candidate, and SHALL explain why a missing or
invalid candidate cannot be installed. The displayed SHA-256 SHALL bind the
operator's click to the bytes later prepared for installation.

#### Scenario: Candidate is absent
- **WHEN** no sibling `.candidate` file exists
- **THEN** settings reports that no update is staged and renders no enabled install action

#### Scenario: Valid candidate is staged
- **WHEN** the fixed sibling candidate is a valid tossctl executable for the current platform
- **THEN** settings shows its identity and an authenticated install action

#### Scenario: Request attempts to select another path
- **WHEN** a client adds a path, URL, or command field to the install request
- **THEN** the console ignores it and the installer can reach only its pre-bound sibling candidate

#### Scenario: Candidate changes after review
- **WHEN** the candidate bytes no longer match the SHA-256 displayed in the submitted settings form
- **THEN** installation is refused and neither current nor rollback bytes change

### Requirement: Candidate validation does not execute candidate code
Before installation the updater SHALL open with no-follow semantics, copy from
that one descriptor into a same-directory prepared file, and inspect file
metadata, SHA-256, and Go build information on the prepared bytes without
executing the candidate. It SHALL reject symlinks, non-regular files, files
without executable bits, wrong module identities, main-package identities other
than `github.com/JungHoonGhae/tossinvest-cli/cmd/tossctl`, GOOS/GOARCH
mismatches, and a prepared hash that differs from the operator-reviewed hash.

#### Scenario: Symlink candidate
- **WHEN** the sibling candidate is a symlink
- **THEN** inspection refuses it and no installed or rollback byte changes

#### Scenario: Wrong module or platform
- **WHEN** build information names another module, GOOS, or GOARCH
- **THEN** inspection refuses it and no candidate code runs

#### Scenario: Different command from the same module
- **WHEN** build information names this repository as the module but names a main package other than `cmd/tossctl`
- **THEN** inspection refuses it and the running executable cannot be replaced by another repository tool

### Requirement: Install is idle-only, authenticated, and recoverable
`POST /settings/system-update/install` SHALL require the normal console session
and CSRF token. It SHALL refuse while a verification run is unfinished, while
the real engine exclusion cannot be held, while external verification evidence
is fresh or unreadable, or when same-port relaunch is unavailable. The
engine-start, verification-start, and update handlers SHALL share an in-process
exclusion, and a successful commit SHALL refuse new starts until the old console
exits. Immediately before commit, the installed target SHALL still be the same
regular executable fingerprinted at console startup.

The engine, updater, standalone verification command, and console verification
starter SHALL also share the same kernel-enforced journal-directory flock.
Verification SHALL acquire it before record/account/broker work and hold it
through complete runner cleanup. The updater SHALL hold it through executable
replacement and relaunch request. Advisory verification evidence SHALL NOT be
the sole exclusion.

A successful install SHALL prepare and sync a same-directory temporary file,
atomically publish and sync a rollback copy while the current path remains
intact, atomically replace the running path, sync the directory, record old/new
hashes and rollback path in console output, and then request the existing
authenticated same-port relaunch. A failed replacement SHALL leave or restore
the previous executable and keep the current console serving.

#### Scenario: Engine is running
- **WHEN** the engine holds its journal-directory exclusion and the operator posts a valid install request
- **THEN** installation is refused, the engine is not stopped, and executable files are unchanged

#### Scenario: Verification is running
- **WHEN** a verification run is unfinished and the operator posts a valid install request
- **THEN** installation is refused and executable files are unchanged

#### Scenario: External verification races with replacement
- **WHEN** update holds the real execution flock and an external or console verification attempts to start before replacement finishes
- **THEN** verification refuses before account resolution or any order-capable broker construction

#### Scenario: Advisory verification marker cannot be written
- **WHEN** a live verification owns the real execution flock but its advisory marker is missing or unwritable
- **THEN** installation is still refused by the real flock

#### Scenario: Replacement succeeds
- **WHEN** the console is idle, the candidate is valid, and replacement completes
- **THEN** the old executable is the rollback file, the candidate bytes occupy the running path, and the browser returns through the existing same-port handoff

#### Scenario: Final replacement fails
- **WHEN** candidate rename fails or the post-replacement directory sync fails
- **THEN** the current path is never absent, the updater leaves or restores the previous executable, reports restoration status, and does not request relaunch

#### Scenario: Installed target drifted
- **WHEN** the current executable path or bytes differ from the console startup fingerprint at commit time
- **THEN** installation is refused and the operator is told to restart the console before reviewing the update again

### Requirement: Development staging has a stable target
The repository SHALL provide a staging target that builds the current source and
writes the result to `<install-path>.candidate` with executable mode. It SHALL
not overwrite the installed executable or trigger a restart.

#### Scenario: Agent stages a gated build
- **WHEN** a developer or agent runs the documented staging target
- **THEN** only the fixed candidate is written and the operator can inspect and install it from settings
