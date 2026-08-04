# Risk Pattern Report: `readinessFixture.marketInputForBody`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/protectionreadiness/attestation_test.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| cryptographic test helper | fixture construction | reviewed-safe | test-only key and files |
