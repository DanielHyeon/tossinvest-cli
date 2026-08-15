# Risk Pattern Report: `newVerifyRunCmd`

- Source: `cmd/tossctl/verify.go`
- Range: `131-217`
- Command: `ast-grep scan -c tools/logic-map/sgconfig.yml cmd/tossctl/verify.go`

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| Mutation/retry/error patterns | `cmd/tossctl/verify.go:131-217` | reviewed-safe only with recorded fail-closed invariants | `function-logic-map.md` |
| Causal authority loss | `cmd/tossctl/verify.go:131-217` | defect if execution continues | `branch-test-map.md` |

No automated match is promoted without source and named-test confirmation.
