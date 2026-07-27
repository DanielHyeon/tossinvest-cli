# Counterfactual entry-geometry boundary map

## What this output can and cannot give

| Downstream requirement | This output | Why |
|---|---|---|
| ① gross/net boundary map | **yes** | the synthetic grid below |
| ② stop-width distribution (the basis for `k`) | **only from the real-trade population** | the grid's widths are chosen, so deriving `k` from them is circular |
| ③ measured cost ratios per market | **no** | all seven rates are `[미검증]`; today's values restate the placeholders |
| ④ declared target vs realised exit | **no** | needs closed positions; this change only records the target |

- Cost model fingerprint: `costs/71d81b5150330fd2`
- Cost scope: `FEE_TAX_ONLY` — commission and tax on both legs. Slippage is **not** included, so the metric is 수수료·세금 차감 후 RR.
- This is a boundary map, not a distribution. The grid's density is chosen by whoever wrote the spec, so the *counts* below measure that choice; only the boundary values (the largest ratio a candidate refuses and the smallest it keeps) are properties of the chain and the cost model.
- Left-truncated. The stop-contract rung (5) already refuses any target below the cost-inclusive break-even, so no chain-allowed point can carry a net ratio at or below zero. The absence of such points is a property of the existing chain, not of entries.
- The minimum stop-width constant `k` is NOT settled by this output. The grid's stop widths are values the spec's author chose, so deriving a floor from them is circular. The only non-synthetic source is the real-trade population; without it, `k` remains open.

## Market `kr`

Grid points: 70. Chain-allowed: 30.

Refusals, by rung — reported apart because they are different facts:

- `stop_contract` (target below break-even): 6
- `min_reward_risk` (gross ratio under 2.0): 34

### Candidate net thresholds

| Candidate | would refuse | would keep | largest refused | smallest kept |
|---|---|---|---|---|
| 1.3 | 12 | 18 | 1.0382559305 | 1.3294353491 |
| 1.5 | 18 | 12 | 1.3980738363 | 1.6631016043 |
| 2.0 | 24 | 6 | 1.797752809 | 2.1806569343 |

## Market `us`

> ⚠️ US limits are fabricated. checkOrderSize runs before min_reward_risk and refuses a cross-currency intent as INPUT_UNAVAILABLE, and risk.DefaultPolicy() is KRW in every field — so a US grid point cannot be evaluated at all without inventing a USD limit set. These numbers have no provenance and must not be cited as policy.

Grid points: 70. Chain-allowed: 16.

Refusals, by rung — reported apart because they are different facts:

- `stop_contract` (target below break-even): 40
- `min_reward_risk` (gross ratio under 2.0): 14

### Candidate net thresholds

| Candidate | would refuse | would keep | largest refused | smallest kept |
|---|---|---|---|---|
| 1.3 | 12 | 4 | 1.105748758 | 1.4567068843 |
| 1.5 | 14 | 2 | 1.4567068843 | 1.8076650106 |
| 2.0 | 16 | 0 | 1.8076650106 | — |

## Real-trade population

- Source: StockOS 058 post-mortem, transcribed. Eight entries, zero wins, one strategy, one session — a transcription, not a sample. Geometry normalised to entry 100 because the document tabulates ratios rather than instruments. (`document`)
- Rows: **8**. Every number below is over 8 rows and is not a rate.

### Observed stop widths

This is the one axis the synthetic grid cannot supply — the grid's widths are chosen, these were traded.

- widths: `0.007`, `0.007`, `0.007`, `0.007`, `0.007`, `0.007`, `0.007`, `0.007`

### Entries through today's chain

| label | market | entry | stop | target | width | gross | net | verdict |
|---|---|---|---|---|---|---|---|---|
| 058-1 | kr | 100 | 99.30 | 101.05 | 0.007 | 1.5 | 0.4558970932 | min_reward_risk/MIN_RR_NOT_MET |
| 058-2 | kr | 100 | 99.30 | 101.05 | 0.007 | 1.5 | 0.4558970932 | min_reward_risk/MIN_RR_NOT_MET |
| 058-3 | kr | 100 | 99.30 | 101.05 | 0.007 | 1.5 | 0.4558970932 | min_reward_risk/MIN_RR_NOT_MET |
| 058-4 | kr | 100 | 99.30 | 101.05 | 0.007 | 1.5 | 0.4558970932 | min_reward_risk/MIN_RR_NOT_MET |
| 058-5 | kr | 100 | 99.30 | 101.05 | 0.007 | 1.5 | 0.4558970932 | min_reward_risk/MIN_RR_NOT_MET |
| 058-6 | kr | 100 | 99.30 | 101.05 | 0.007 | 1.5 | 0.4558970932 | min_reward_risk/MIN_RR_NOT_MET |
| 058-7 | kr | 100 | 99.30 | 101.05 | 0.007 | 1.5 | 0.4558970932 | min_reward_risk/MIN_RR_NOT_MET |
| 058-8 | kr | 100 | 99.30 | 101.05 | 0.007 | 1.5 | 0.4558970932 | min_reward_risk/MIN_RR_NOT_MET |

### Candidate net thresholds against these entries

| candidate | would refuse | would keep | largest refused | smallest kept |
|---|---|---|---|---|
| 1.3 | 0 | 0 | — | — |
| 1.5 | 0 | 0 | — | — |
| 2.0 | 0 | 0 | — | — |

The minimum stop-width constant `k` is NOT settled by this output. The grid's stop widths are values the spec's author chose, so deriving a floor from them is circular. The only non-synthetic source is the real-trade population; without it, `k` remains open.
