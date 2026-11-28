# a063 task 4.2 engine-start operational record

Timestamp (KST): 2026-09-06T21:05:49+0900

## Result

**BLOCKED — engine launch was attempted exactly once; no 4.2 mutation occurred.**

The human-authorized one-time detached engine start was issued with the reviewed candidate and explicit target profile. The shell accepted the launch (exit 0), but subsequent executable-bound process classification found zero running instances of that reviewed candidate with `engine run`. No retry was attempted. Because a running engine with an explicit config-dir equal to the target profile could not be proved, the fail-closed profile gate did not pass and the 4.2 activation block was not entered.

No repository file, task checkbox, archive, trading setting, Guardian setting, automation setting, console setting, or order/cancel/amend state was changed. No survey was started. No service was manually started, stopped, or restarted.

## Sanitized preflight and launch evidence

| Check | Classification | Exit |
| --- | --- | --- |
| Profile directory | directory, non-symlink, normalizes | `0 / 1 / 0` |
| User systemd bus | available | `0` |
| Reviewed candidate | regular, non-symlink | `0 / 1` |
| Reviewed service template | regular, non-symlink | `0 / 1` |
| Reviewed timer template | regular, non-symlink | `0 / 1` |
| Candidate + both template SHA-256 bindings | exact | `0` |
| Existing binary/service/timer targets | each regular, non-symlink | `0 / 1` each |
| Both unit FragmentPath values | exact expected targets | `0` each |
| Both unit drop-in classifications | none | `0` each |
| Timer enabled and active | yes / yes | `0 / 0` |
| Attestation service active state | inactive | `0` |
| Initial verified engine count | zero | `0` for zero-count test |
| Detached engine log preparation | succeeded | `0` |
| Detached engine launch submission | succeeded once | `0` |
| Post-launch executable-bound engine count | zero | classification only |
| Post-launch explicit config-dir/profile equality | not provable | classification only |

Verified SHA-256 values:

- candidate: `d9ef7f5ecac62a9e93159ebfc5cf5cf50d1ac1ee93b68c7ff9d48704777f6a6b`
- service template: `0bf50972660fa14ece6953b1d43aa51fe2df407a51783c7ea0c6decb3100de96`
- timer template: `0db4e0cbfada2a59de4756747f17a3d7596d80dc7a0adafb3b3460849611edde`

The one permitted engine command was:

```bash
setsid /tmp/a063-reviewed-20260906/tossctl --config-dir "$HOME/.config/tossctl" engine run
```

Its stdout and stderr were redirected only to `/tmp/a063-task42-engine.log`. That log was not read or copied into this record. Raw process command lines, credentials, accounts, record contents, attestation values, and runtime logs were neither printed nor retained here. The short-lived launcher PID was not retained because it did not establish a running engine identity.

## 4.2 activation result

The approved installation sequence was **not** run: no backup directory was created, `tossos-attest.timer` was not disabled or enabled, no candidate or unit template was installed, no `systemd-analyze --user verify` was run, and no `systemctl --user daemon-reload` was run. This preserves the previously enabled, active timer and inactive service state.

## Survey classification and 4.3 handoff

Executable-bound read-only classification found zero running `soak run` processes from the reviewed candidate and zero same-profile surveys. No shell or console survey start was invoked. Task 4.3 remains a human action through the existing console survey control, only after a separately successful same-profile engine proof; it must not be started from shell.

## Final repository status

`git status --porcelain=v1` exited `0`; the repository remains dirty with 44 pre-existing/unrelated status entries. This operation made no repository writes.

