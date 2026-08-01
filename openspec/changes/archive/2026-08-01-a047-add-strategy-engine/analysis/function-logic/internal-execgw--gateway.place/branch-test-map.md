# Branch Test Map: `Gateway.Place`

| Branch | Scenario | Test | RED observed | GREEN observed |
| --- | --- | --- | --- | --- |
| B1 | optional preflight exists, so `Place` attaches its `CheckPlace` callback to the mutation plan | `TestGatewayAsksTheSymbolQuestion`, protection preflight tests | baseline | baseline |
| Scenario | no preflight is configured; the plan leaves its callback nil | ordinary `Gateway.Place` decision/round-trip tests | baseline | baseline |
| Scenario | decision, mode, reservation, journal, official-call and in-doubt behavior belongs to `g.submit`; none is an AST branch in `Gateway.Place` | decision, modegate, reservation-gate, roundtrip and in-doubt suites; see `Gateway.submit` maps | baseline | baseline |
| A047 | strategy has no alternate submitter/import path | dormant wiring and official adapter static tests | missing | pass |
