# CodeGraph baseline: trusted-network revision

- Date: 2026-07-31
- Change: `enable-vpn-console-access`
- Base commit: see `base-commit.txt`
- Current evidence: CodeGraph indexed the current worktree before production edits.

## Symbols

- `internal/console.newRemoteRuntime`
- `internal/console.validateRemoteFields`
- `internal/console.remoteAccessEmpty`
- `internal/console.(*Console).URL`
- `internal/console.(*Console).routes`
- `internal/console.(*Console).session0`
- `cmd/tossctl.newConsoleCmd`
- `cmd/tossctl.remoteAccessOptions`

## Hard-evidence conclusion

`Console.routes` wraps every operational route with `session0`; `session0` is the
single application-login gate for both local and remote traffic. `mutating` is a
separate method/origin/CSRF gate and must remain unchanged. Remote transport
validation and peer/Host security are separate from the login/session code and can
remain fail-closed while trusted-network access bypasses only `session0`.
