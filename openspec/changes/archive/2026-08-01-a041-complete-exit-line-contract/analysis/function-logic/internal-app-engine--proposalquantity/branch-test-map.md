# Branch Test Map: `proposalQuantity`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | invalid quantity/ratio rejects | existing/new projector unit tests | existing | yes |
| B2 | ratio above one caps to held quantity | existing/new projector unit test | existing | yes |
| B3 | one × partial floors to zero | a041 one-share projection test | no | yes |
| B4 | one × full returns one | a041 one-share projection test | no | yes |
| B5 | zero/negative cannot produce an order | a041 zero-order invariant test | no | yes |
