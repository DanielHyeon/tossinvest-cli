# a145 offline shadow analysis

Run against a frozen a128 snapshot containing `ledger.jsonl` and `states/`. The destination must not exist and must be outside both snapshot and repository. No network calls are made.

```sh
python3 tools/a145-jev-offline-analysis/analysis.py /path/to/snapshot /path/to/new-output
python3 -m unittest discover -s tools/a145-jev-offline-analysis -p 'test_*.py'
```

The output has an input manifest, per-state quality sidecar, per-session same-cycle comparisons, quality-stratum counts and a separate r2 status. Returns are 60-minute price changes before costs; observations repeat and overlap. Quality strata overlap and row means are not cycle-weighted comparisons. K=0 and incomplete cycles are excluded and counted. No trading outcome or causal claim follows.

Without `--r2-responses PATH`, the comparison is BLOCKED. A bundle requires explicit `evidence_kind`: `synthetic` produces fixture-only results; `recorded` requires the original ledger to anchor exact original request bytes by `request_sha256` and a bundle row with `baseline_request_json`, `r2_request_json`, `r2_response_json`, matching state digest, and explicit `model_version` both in the original judgment rows and each response row. Requests must retain exactly the same state, model, all non-question options and J1; only J2 changes to the question fixed in the design. Original J1 and J2 question digests and response model must match. Missing or unprovable provenance excludes the entire cycle. The current historical snapshot does not record the original request hash and cannot establish an empirical r2 match. No responses are generated. Request/response checks establish only consistency of supplied records, not that a model call happened or the response is authentic. A later model may have hindsight even when request evidence matches.
