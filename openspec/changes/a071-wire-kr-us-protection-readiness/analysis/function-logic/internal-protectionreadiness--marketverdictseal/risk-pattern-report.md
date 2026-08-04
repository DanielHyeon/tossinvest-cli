# Risk Pattern Report: `marketVerdictSeal`

Run:

```bash
ast-grep scan -c tools/logic-map/sgconfig.yml internal/protectionreadiness/dispatch.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| incomplete authorization preimage | seal omitted signed operational scope fields | defect | include all authority-bearing scope and capability digest in verdict/snapshot seals |
