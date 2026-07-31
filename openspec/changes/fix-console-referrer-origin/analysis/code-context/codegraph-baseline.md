# CodeGraph Baseline

- Change: `fix-console-referrer-origin`
- Base commit: `a8d43a5e826bbaa8898707c0411a18e84fdbcb28`
- Query: browser form `POST /restart` reaches the remote origin refusal when
  the response policy is `no-referrer`.

## Hard evidence

| Symbol | Definition | Evidence |
|---|---|---|
| `remoteRuntime.security` | `internal/console/remote.go:276` | Sets all remote response security headers before peer and Host gates. `Referrer-Policy` is exactly `no-referrer`. |
| `Console.render` | `internal/console/pages.go:376` | Sets rendered HTML response headers after template execution and currently overwrites `Referrer-Policy` with `no-referrer`. |
| `pageTemplates` | `internal/console/templates.go:9` | The shared `head` and standalone `restart` templates each declare `<meta name="referrer" content="no-referrer">`. |
| `Console.mutating` | `internal/console/console.go:796` | Enforces POST → remote origin → form parse → CSRF → handler. The reported Korean text is emitted only by the origin branch. |
| `remoteRuntime.sameOriginForMutation` | `internal/console/remote.go:267` | Explicit `Origin` or `Referer` is final evidence; only total header absence can use direct TLS+Host fallback. |

`remoteRuntime.security` has one caller, `Console.routes`. CodeGraph impact
reports 49 reachable symbols because the wrapper fronts all console routes.
`Console.render` has two direct indexed impacts (the method and its file), but
it is shared by normal console pages. The edit boundary is therefore the two
response-header values and the two document meta values; peer, Host, health,
origin, CSRF, audit, handler branches, and all other security headers must
remain byte-for-byte unchanged.

## Runtime reproduction

A real headless Chrome form submission with an intentionally invalid CSRF value
sent:

```text
Origin: null
Referer: <absent>
```

The response was the origin-refusal page. Adding a document-level
`same-origin` referrer policy in the browser-only probe changed the headers to:

```text
Origin: https://127.0.0.1:37085
Referer: https://127.0.0.1:37085/
```

The response then reached the CSRF-refusal page. The invalid token proves the
restart handler was not invoked in either probe.
