# CodeGraph baseline

- Change: `fix-console-origin-fallback`
- Base commit: `c06192799fa9faa744bf194c8d9636bf9441195c`
- Index status: up to date, 981 files / 17,210 nodes / 51,928 edges
- Query: all console POSTs rejected by the remote same-origin gate when the
  browser submits from the canonical HTTPS URL without Origin/Referer.

## Definitions and flow

| Symbol | Definition | Evidence |
|---|---|---|
| `remoteRuntime.sameOrigin` | `internal/console/remote.go:242` | Reads `Origin`; if empty, parses `Referer`; compares only `scheme://host` with `rr.origin` |
| `Console.mutating` | `internal/console/console.go:796` | POST → remote same-origin → form parse → CSRF → handler |
| `remoteRuntime.security` | `internal/console/remote.go:255` | Before the mux, enforces peer CIDR and case-insensitive exact `Host` against the configured public URL |
| `newRemoteRuntime` | `internal/console/remote.go:90` | Stores canonical origin as `publicURL.Scheme + "://" + publicURL.Host` |

## Callers and impact

- Direct callers of `sameOrigin`: `Console.mutating`, `remoteRuntime.loginPost`.
- Depth-3 impact: 11 symbols, limited to `internal/console/remote.go` and
  `internal/console/console.go`.
- All console state-changing routes are wrapped by `mutating`; authenticated
  compatibility login POST also uses the same function.
- No broker, order, Guardian, journal, exit-policy, or engine runtime symbol is
  in the impact set.

## Root cause evidence

`sameOrigin` has no branch for the case where both privacy-sensitive headers are
absent. `remoteRuntime.security` simultaneously emits
`Referrer-Policy: no-referrer`. The user reproduced the error at
`https://127.0.0.1:37085/restart`, proving the browser reached the exact
configured HTTPS scheme, host, and port and that `/restart` was only the POST
target path. The request is rejected before form parsing, so no handler or
state-changing seam runs.
