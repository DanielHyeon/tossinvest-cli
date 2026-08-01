# Branch Test Map: `Tracer.submitEntry`

| Branch | Scenario | Test | RED observed | GREEN observed |
| --- | --- | --- | --- | --- |
| B1 | unsupported market refuses | tracer parameter validation tests | baseline | baseline |
| B2 | Guardian refusal causes zero submit calls | tracer issuer/refusal tests | baseline | baseline |
| B3 | invalid quantity/limit conversion refuses before gateway | tracer numeric tests | baseline | baseline |
| B4 | invalid limit conversion refuses before gateway | tracer numeric tests | baseline | baseline |
| B5 | failed/in-doubt/non-confirmed gateway outcome is not success | tracer outcome tests | baseline | baseline |
| B6 | missing outcome detail falls back to the gateway error without treating it as success | tracer failed-outcome tests | baseline | baseline |
| Confirmed | confirmed entry records exact IDs | `runTracerWithFills`-based tracer tests | baseline | baseline |
| A047 | strategy orchestrator revalidates manifest around planning/dispatch without using tracer | dispatch TOCTOU/lease tests + dormant wiring guard | missing | pass |
