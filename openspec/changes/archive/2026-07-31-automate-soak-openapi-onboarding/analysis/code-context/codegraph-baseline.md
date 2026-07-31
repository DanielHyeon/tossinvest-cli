# CodeGraph baseline

Base commit: `fa57d98`

CodeGraph was synchronized before production editing. Hard evidence:

- `internal/console.(*Console).routes` registers `/soak/restart` as
  `session0(mutating(handleSoakRestart))`.
- `internal/console.(*Console).mutating` checks POST and exact remote origin,
  then calls `ParseForm`, then checks CSRF.
- `internal/console.(*Console).handleSoakRestart` currently calls
  `Options.RestartSoak` directly and redirects the returned note.
- `cmd/tossctl.runConsole` resolves the soak record and supplies a narrow
  `RestartSoak` closure.
- `cmd/tossctl.validateOpenAPICredentials` builds the official client and calls
  the read-only `Accounts` probe.
- `internal/official.LoadCredentials` gives a complete environment pair
  precedence over the credential file.
- `internal/official.tokenManager.token` reuses a still-valid disk token before
  exchanging credentials, so replacement validation needs an isolated cache.
- `cmd/tossctl.restartSoak` signals prior soak processes, waits for exit, then
  spawns one detached child and returns its operator note.

Affected existing production functions are `routes`, `handleSoakRestart`, and
`runConsole`. New credential orchestration helpers remain narrow leaves.
