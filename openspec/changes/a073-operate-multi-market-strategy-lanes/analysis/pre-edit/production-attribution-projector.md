# Production lane-attribution projector — pre-edit map

## Gap

The immutable `performance.db` schema, exact attribution builder and
`journal.ReadOnly` adapter exist, but production has no owner that invokes
`AttributionRebuild` and `PersistAttributionRebuild`. Console and HTTP readers
therefore can only observe fixture- or externally-created data.

## Boundary

- add a separate `tossctl performance project-attribution` operational lane;
- read the trading journal only through `journal.ReadOnly`;
- write only the rebuildable derived `performance.db`;
- discover exact account identities from persisted journal rows;
- use a bounded closed-trade window and deterministic composite lineage;
- preserve current missing fill/cost/FX facts as `link_missing` and
  `not_measured`, never as numeric zero;
- expose no broker, order, toggle, activation, protection or LIVE capability;
- support a one-shot run and an explicitly configured periodic loop without
  coupling the projector to console/httpapi reader processes or safety loops.

## Branch test map

| Branch | Expected proof |
|---|---|
| missing journal | typed refusal; no performance database created |
| no account rows | successful zero-account projection; no invented account |
| account rows | one atomic rebuild per exact account identity |
| journal evidence gap | unavailable row persists without numeric/FX defaults |
| periodic interval invalid | command refuses before opening either database |
| capability surface | no official client, Gateway, broker or operating writer |

