# Risk Pattern Report: `TestDispatchRejectsCorruptAndFutureIssuedSnapshots`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/protectionreadiness/dispatch_test.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| none | local corrupt-snapshot test | reviewed-safe | no side effect |
