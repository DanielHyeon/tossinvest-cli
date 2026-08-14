# Branch Test Map: `Journal.recordExitJudgementTx`

| Branch | Source condition / scenario | Test/evidence | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | blank position ID | invalid request table | existing | existing |
| B2 | judgement provenance validation fails | provenance identity cases | existing | existing |
| B3 | legacy provenance asserts arm suppression | legacy suppression rejection case | existing | existing |
| B4 | proposal is present | proposal validation table | existing | existing |
| B5 | proposal action/level invalid | state-only/nonorderable arm case | existing | existing |
| B6 | proposal provenance invalid | provenance validation case | existing | existing |
| B7 | proposal and judgement provenance differ | cross-snapshot proposal case | existing | existing |
| B8 | nonlegacy provenance builds recomputed candidate | full snapshot persistence suite | existing | existing |
| B9 | recomputed candidate validation fails | forged/partial tuple cases | existing | existing |
| B10 | begin transaction fails | storage error/rollback seam | existing | existing |
| B11 | current exit progress read fails | missing/corrupt state cases | existing | existing |
| B12 | exit state completed | completed-state rejection case | existing | existing |
| B13 | lifecycle generation omitted | legacy generation compatibility case | existing | existing |
| B14 | expected and current exit lifecycle differ | stale generation case | existing | existing |
| B15 | no lifecycle row, apply legacy managed default | legacy journal case | existing | existing |
| B16 | lifecycle query returned a row | managed lifecycle cases | existing | existing |
| B17 | lifecycle query failed | storage error case | existing | existing |
| B18 | lifecycle status/generation not current managed | release/readopt race case | existing | existing |
| B19 | snapshot and position generations differ | generation mismatch case | existing | existing |
| B20 | recomputed snapshot checks decision uniqueness | duplicate-decision suite | existing | existing |
| B21 | decision already exists | duplicate returns conservative pending | existing | existing |
| B22 | duplicate query failed unexpectedly | storage error case | existing | existing |
| B23 | legacy judgement uses scalar monotonic guards | legacy compatibility suite | existing | existing |
| B24 | legacy high-water descends | descending high-water case | existing | existing |
| B25 | legacy baseline descends | descending baseline case | existing | existing |
| B26 | ratchet level omitted | retain current level case | existing | existing |
| B27 | saved effective snapshot exists | recovery candidate persistence suite | existing | existing |
| B28 | recomputed snapshot exists | recovery selection suite | existing | existing |
| B29 | saved snapshot supplied to selector | saved-vs-recomputed cases | existing | existing |
| B30 | recovery selection is ambiguous | quarantine case | existing | existing |
| B31 | quarantine write fails | quarantine rollback seam | existing | existing |
| B32 | quarantine commit fails | commit failure seam | existing | existing |
| B33 | rejudged quarantine release CAS fails | rejudgement eligibility suite | existing | existing |
| B34 | saved monotone snapshot wins | saved-monotone suppression case | existing | existing |
| B35 | recomputed snapshot wins | full recomputed persistence case | existing | existing |
| B36 | saved source strips proposal/suppression | saved snapshot cannot arm case | existing | existing |
| B37 | effective snapshot exists for complete tuple update | JSON/flattened persistence suite | existing | existing |
| B38 | stored snapshot encoding/validation fails | output-digest/forgery case | existing | existing |
| B39 | state UPDATE fails | staged transaction rollback case | existing | existing |
| B40 | after-state hook fails | staged rollback case | existing | existing |
| B41 | proposal is present for arming | arming-before-submit suite | existing | existing |
| B42 | proposal arm fails | pending/arm rollback case | existing | existing |
| B43 | after-arm hook fails | staged rollback case | existing | existing |
| B44 | complete evaluation tuple attached to event | event tuple integrity case | existing | existing |
| B45 | event append fails | rollback seam | existing | existing |
| B46 | after-event hook fails | staged rollback case | existing | existing |
| B47 | final commit fails | crash/rollback case | existing | existing |
| B48 | project committed arm outcome | result outcome table | existing | existing |
| B49 | saved source chosen | saved-monotone outcome | existing | existing |
| B50 | proposal durably armed | armed proposal authority case | existing | existing |
| B51 | known arm suppression reason | suppressed-working-order outcome | existing | existing |

## A111 sibling-transaction requirements

- N1: exact EVALUATED equality replaces effective JSON and every flattened field atomically, but appends no event and arms no proposal.
- N2: SEED, missing effective evidence, lifecycle/generation/status drift, or operational inequality returns a typed conflict without mutation.
- N3: an injected failure during refresh leaves the entire previous JSON/flattened tuple intact.
- N4: concurrent semantic judgement versus refresh yields one coherent winner; no stale overwrite and no lost event/proposal.
