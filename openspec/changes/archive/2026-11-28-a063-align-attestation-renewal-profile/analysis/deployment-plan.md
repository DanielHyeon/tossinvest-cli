# Proposed Linux deployment plan

**Status: reviewed proposal, not executed; explicit human approval and
same-profile evidence pending.** It is not runtime or operational readiness.

## Reviewed candidate

Built on 2026-09-06 with exit 0:

```bash
go build -o /tmp/a063-reviewed-20260906/tossctl ./cmd/tossctl
```

| Artifact | SHA-256 |
|---|---|
| `/tmp/a063-reviewed-20260906/tossctl` | `d9ef7f5ecac62a9e93159ebfc5cf5cf50d1ac1ee93b68c7ff9d48704777f6a6b` |
| `deploy/systemd/tossos-attest.service` | `0bf50972660fa14ece6953b1d43aa51fe2df407a51783c7ea0c6decb3100de96` |
| `deploy/systemd/tossos-attest.timer` | `0db4e0cbfada2a59de4756747f17a3d7596d80dc7a0adafb3b3460849611edde` |

Read-only `candidate soak attest --help` confirmed `--record-renewal-status`.
The intended profile is `$HOME/.config/tossctl`; its configured attestation
resolver remains authoritative, with status at `<resolved-attestation>.renewal-status.json`.

## Proposed human-run procedure

This is an operational activation procedure. Run it only after explicit human
approval of the install window. It does not authorize an engine restart, a
survey reset, or an automation-toggle change.

The repository templates are currently executable (`0755`), so
`systemd-analyze --user verify` reports that warning while returning exit 0.
The installed copies below are deliberately mode `0644`; verify those copied
files before enabling the timer.

```bash
set -eu
candidate=/tmp/a063-reviewed-20260906/tossctl
unit_dir="$HOME/.config/systemd/user"
backup="$unit_dir/a063-backup-$(date -u +%Y%m%dT%H%M%SZ)"
mkdir -p "$HOME/.local/bin" "$unit_dir" "$backup"

# Bind the reviewed candidate and the exact repository templates immediately
# before installation. Stop here if any digest differs.
printf '%s  %s\n' \
  'd9ef7f5ecac62a9e93159ebfc5cf5cf50d1ac1ee93b68c7ff9d48704777f6a6b' "$candidate" \
  '0bf50972660fa14ece6953b1d43aa51fe2df407a51783c7ea0c6decb3100de96' 'deploy/systemd/tossos-attest.service' \
  '0db4e0cbfada2a59de4756747f17a3d7596d80dc7a0adafb3b3460849611edde' 'deploy/systemd/tossos-attest.timer' |
  sha256sum -c -

# This narrow procedure supports only existing regular files and an enabled,
# active timer. A symlink, absent target, masked/disabled/inactive timer, or a
# drop-in means stop and make a separate reviewed migration/rollback plan.
for path in "$HOME/.local/bin/tossctl" "$unit_dir/tossos-attest.service" "$unit_dir/tossos-attest.timer"; do
  [ -f "$path" ] && [ ! -L "$path" ] || { echo "unsupported source: $path" >&2; exit 1; }
done
systemctl --user is-enabled --quiet tossos-attest.timer
systemctl --user is-active --quiet tossos-attest.timer
for unit in tossos-attest.service tossos-attest.timer; do
  dropins=$(systemctl --user show -p DropInPaths --value "$unit")
  [ -z "$dropins" ] || { echo "unapproved drop-in for $unit: $dropins" >&2; exit 1; }
  systemctl --user show -p FragmentPath -p DropInPaths "$unit"
done
systemctl --user cat tossos-attest.service tossos-attest.timer

# Copy the preceding `show` and `cat` output to an approved, redacted evidence
# record and compare the effective unit before proceeding. `--help` only proves
# command support; it does not prove profile, resolved paths, or account.
"$candidate" --config-dir "$HOME/.config/tossctl" soak attest --help | grep -- --record-renewal-status
echo "STOP: attach approved redacted console-profile/record/attestation/account evidence before install" >&2
exit 1
```

After the separate evidence gate has been approved, run this activation block.
It repeats the digest binding because the candidate and templates can change
between preflight and installation.

```bash
set -eu
candidate=/tmp/a063-reviewed-20260906/tossctl
unit_dir="$HOME/.config/systemd/user"
backup="$unit_dir/a063-backup-REPLACE_WITH_RECORDED_TIMESTAMP"
[ -d "$backup" ] || { echo "missing approved backup directory" >&2; exit 1; }
printf '%s  %s\n' \
  'd9ef7f5ecac62a9e93159ebfc5cf5cf50d1ac1ee93b68c7ff9d48704777f6a6b' "$candidate" \
  '0bf50972660fa14ece6953b1d43aa51fe2df407a51783c7ea0c6decb3100de96' 'deploy/systemd/tossos-attest.service' \
  '0db4e0cbfada2a59de4756747f17a3d7596d80dc7a0adafb3b3460849611edde' 'deploy/systemd/tossos-attest.timer' |
  sha256sum -c -
# Revalidate every mutable precondition immediately before the snapshot and
# mutation.  Only the exact user-unit paths, no drop-ins, and the documented
# enabled/active timer state are supported by this draft.
for path in "$HOME/.local/bin/tossctl" "$unit_dir/tossos-attest.service" "$unit_dir/tossos-attest.timer"; do
  [ -f "$path" ] && [ ! -L "$path" ] || { echo "unsupported source: $path" >&2; exit 1; }
done
for unit in tossos-attest.service tossos-attest.timer; do
  fragment=$(systemctl --user show -p FragmentPath --value "$unit")
  [ "$fragment" = "$unit_dir/$unit" ] || { echo "unexpected fragment: $fragment" >&2; exit 1; }
  dropins=$(systemctl --user show -p DropInPaths --value "$unit")
  [ -z "$dropins" ] || { echo "unapproved drop-in: $dropins" >&2; exit 1; }
done
systemctl --user is-enabled --quiet tossos-attest.timer
systemctl --user is-active --quiet tossos-attest.timer
service_state=$(systemctl --user show -p ActiveState --value tossos-attest.service)
[ "$service_state" = inactive ] || { echo "renewal service state: $service_state" >&2; exit 1; }

# The complete, regular-file snapshot exists before the timer is stopped.
cp -a -- "$HOME/.local/bin/tossctl" "$unit_dir/tossos-attest.service" "$unit_dir/tossos-attest.timer" "$backup/"
for path in "$HOME/.local/bin/tossctl" "$unit_dir/tossos-attest.service" "$unit_dir/tossos-attest.timer"; do
  name=$(basename "$path")
  [ -f "$backup/$name" ] && [ ! -L "$backup/$name" ] && cmp -s "$path" "$backup/$name" || {
    echo "backup validation failed: $name" >&2; exit 1;
  }
done
systemctl --user disable --now tossos-attest.timer
# Any subsequent failure requires running the documented rollback block with
# this backup before attempting another activation.
service_state=$(systemctl --user show -p ActiveState --value tossos-attest.service)
[ "$service_state" = inactive ] || { echo "post-stop failure; run Proposed rollback with $backup" >&2; exit 1; }
install -m 0755 "$candidate" "$HOME/.local/bin/tossctl"
install -m 0644 deploy/systemd/tossos-attest.service "$unit_dir/tossos-attest.service"
install -m 0644 deploy/systemd/tossos-attest.timer "$unit_dir/tossos-attest.timer"
systemd-analyze --user verify "$unit_dir/tossos-attest.service" "$unit_dir/tossos-attest.timer"
systemctl --user daemon-reload
systemctl --user enable --now tossos-attest.timer  # activation: human-approved only
systemctl --user status tossos-attest.timer --no-pager
```

Do not restart the engine, reset the survey, or change an automation toggle.
After an approved qualifying window, retain three consecutive qualifying survey
days and inspect the resolved attestation/status using read-only commands.

## Proposed rollback

Use the exact `backup` directory printed or recorded during the approved
installation. First stop the a063 timer so it cannot execute between restore
steps. This procedure applies only to the preflight's known regular files and
enabled, active timer state, and restores precisely that approved state.

```bash
set -eu
unit_dir="$HOME/.config/systemd/user"
backup="$unit_dir/a063-backup-REPLACE_WITH_RECORDED_TIMESTAMP"
[ -d "$backup" ] || { echo "missing a063 backup: $backup" >&2; exit 1; }
systemctl --user disable --now tossos-attest.timer
for path in "$HOME/.local/bin/tossctl" "$unit_dir/tossos-attest.service" "$unit_dir/tossos-attest.timer"; do
  name=$(basename "$path")
  [ -f "$backup/$name" ] && [ ! -L "$backup/$name" ] || { echo "invalid backup: $name" >&2; exit 1; }
  install -m "$(stat -c %a "$backup/$name")" "$backup/$name" "$path"
done
systemctl --user daemon-reload
systemctl --user enable --now tossos-attest.timer
```

Console deployment is separate: deploy a console build that contains this UI
only through its existing approved release procedure. Its running launch
arguments are not established here; do not kill a process to replace it.
