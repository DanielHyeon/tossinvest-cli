# Risk Pattern Report: `runVerifyRun`

- Source: `cmd/tossctl/verify.go`
- Range: `272-400`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml cmd/tossctl/verify.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `cmd/tossctl/verify.go:272-400` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `cmd/tossctl/verify.go:272-400` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
