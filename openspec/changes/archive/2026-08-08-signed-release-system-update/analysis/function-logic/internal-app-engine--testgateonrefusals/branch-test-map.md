# Branch Test Map: `TestGateOnRefusals`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | all refusal cases execute | `TestGateOnRefusals` | baseline | pass |
| B2 | optional trading mutation | same | baseline | pass |
| B3 | optional attestation write | same | baseline | pass |
| B4 | no-Guardian row selects sealed disable seam | same | row auto-constructed a Guardian | pass |
| B5 | ordinary rows use normal assembly | same | baseline | pass |
| B6 | any successful refusal case fails | same | baseline | pass |
| B7 | partial context is forbidden | same | baseline | pass |
| B8 | outer automation refusal sentinel required | same | baseline | pass |
| B9 | specific cause sentinel required | same | baseline | pass |
| B10 | operator-readable text required | same | baseline | pass |
