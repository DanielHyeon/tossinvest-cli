# Input profile comparison

`run` defaults to `--input-profile compare`, automatically writing to the separate `shadow-ledger-compare` directory. Explicit `--input-profile current` keeps the existing a128 input and legacy ledger behavior. `initial` projects the original basic field types from the current collection; it is not a reconstruction of the first run or its returns. `compare` judges both profiles from each single collected snapshot. Both use the same questions, model, observation time, price and label horizon; current is requested first, so model timing/order can affect results. Compare uses two model requests per snapshot. Existing screening excludes J2 ≥ 0.5; the comparison report separately counts crossings of the analysis threshold J2 ≥ 0.6.

The standard operator workflow automatically collects both profiles and reads their combined report:

```sh
TYPESAFE_API_KEY=... go run ./tools/a128-jev-shadow-probe run --market KR
go run ./tools/a128-jev-shadow-probe report
```

`initial` and `compare` automatically use separate output directories. An explicit `--out` must remain outside the default operational ledger, including its symlink aliases. Each run receives its own cohort; do not merge experimental output into the preregistered ledger. Starting a run contacts the configured APIs and is a separate operator action. Example only:

```sh
TYPESAFE_API_KEY=... go run ./tools/a128-jev-shadow-probe run --market KR --input-profile compare --out /path/to/separate-experiment
```

Offline reporting of a previously collected experiment:

```sh
go run ./tools/a128-jev-shadow-probe report --compare --ledger /path/to/separate-experiment/ledger.jsonl
python3 tools/a128-jev-shadow-probe/profile_eval.py /path/to/separate-experiment/ledger.jsonl --states /path/to/separate-experiment/states --types /path/to/security-types.json
```

`--types` is optional; absent or unknown types remain unclassified. The evaluator verifies state digests over the existing canonical JSON payload (the stored trailing newline is excluded), uses referenced source snapshots for price bars including snapshots from failed judgments, and reports price gaps. KR uses the existing morning signal window and 10:00→14:10 prices; US uses the existing signal window and 60-minute label, selecting each profile’s first flagged observation (or its first paired observation if none is flagged). Profiles can therefore select different times and returns; the evaluator reports the number of stock-days where that happens. The two profiles always share the label for the same observation. Cost assumptions are 0.50% KR and 2.10% US. Returns are descriptive and carry no profitability PASS verdict. Stop an experiment and use explicit `--input-profile current` to roll back; preserve its separate ledger for audit.
