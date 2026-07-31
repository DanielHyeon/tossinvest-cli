# Risk Pattern Report: `ReadOnly.BrokerOrderIDs`

Run:

```bash
python3 tools/logic-map/risk_pattern_report.py internal/journal/account_views.go
```

## Findings

| Rule | Location | Classification | Function Logic Map link |
|---|---|---|---|
| none | `ReadOnly.BrokerOrderIDs` | reviewed-safe | no configured panic/go/defer/mutation risk pattern matched; see safety conclusion |
