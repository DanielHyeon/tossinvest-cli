# Risk Pattern Report: `scopeMatches`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/protectionreadiness/attestation.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| long boolean authority predicate | `scopeMatches` | reviewed-safe | table matrix mutates every conjunct |
