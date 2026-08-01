# Risk Pattern Report: `Console.routes`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/console/console.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| `go-panic` | `internal/console/console.go:908`, outside `Console.routes` | not-applicable | finding belongs to random-session-key construction, not target function or a043 diff |
