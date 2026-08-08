## Why

Local development builds are currently installed by copying an arbitrary
session-specific `/tmp/.../tossctl-menu` file over `~/.local/bin/tossctl`.
That path expires, gives the console no provenance or rollback evidence, and
forces the operator back to a terminal even though the console already knows how
to detect and restart into a replaced binary.

The console needs a human-clicked, fail-closed system update action for a
well-known staged candidate. It must validate the candidate, refuse while engine
or verification work is active, replace the executable recoverably, and restart
only after a successful installation.

## What Changes

- Add a system-update section to `/settings` showing the running executable and
  the well-known staged candidate with size, timestamp, build metadata, and
  SHA-256.
- Add a session+CSRF protected install action. It accepts no caller-supplied path
  and can install only the candidate path injected by `cmd/tossctl`.
- Validate that the candidate is a regular, non-symlink executable for this
  `tossctl` module and the current GOOS/GOARCH before changing the installed
  binary.
- Refuse installation while a live verification is in flight or the engine
  marker is fresh.
- Install through a same-directory temporary file, preserve one rollback backup,
  sync the file and directory, and restore the previous executable if the final
  replacement fails.
- On success, use the existing authenticated handoff and same-port relaunch
  path. On failure, keep the current console serving and render the reason.
- Add a stable development staging target so agents and developers can produce a
  candidate without inventing `/tmp` paths; the operator's only remaining action
  is the settings-menu click.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `operator-console`: The authenticated settings surface gains a controlled
  binary-install-and-relaunch action with explicit validation, refusal,
  recovery, and audit-visible status.

## Impact

- `internal/console`: update view, handler, route, templates, CSRF/static guards,
  and behavioral tests.
- `internal/localupdate` (new package): candidate inspection and recoverable
  replacement without network or account access.
- `cmd/tossctl`: injection of current executable, stable candidate path, engine
  activity reader, and relaunch seam.
- `Makefile` and developer documentation: stable candidate staging command.
- No account, journal trading rows, automation gate, or risk limits are mutated
  by this capability.
