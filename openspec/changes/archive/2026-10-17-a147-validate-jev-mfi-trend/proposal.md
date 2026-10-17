## Why

Exploratory analysis of stored Jev decisions found a possible association between 1-minute MFI trend and 60-minute price changes, but repeated observations, missing labels and a small number of sessions prevent generalization. The separate discovery report is `outputs/mfi-jev-20261018/report.md` in the ChatGPT task workspace; it is not proof of an improvement.

## What Changes

- Define a separately registered, prospective shadow evaluation for KR and US sessions using closed one-minute bars and 60-minute labels; the exploratory target is five new regular sessions per market, not a statistical sufficiency claim.
- Append source- and offset-preserving bar evidence and analysis sidecars without rewriting r1 questions, screening, decisions, historical evidence, orders or operational settings.
- Compare MFI trend and outcome within market/session and Jev selected/nonselected groups; document missingness, dependence and inconclusive results.
- This authorization covers design only. Implementation, collecting new data, any model or external API call, scheduling and deployment require separate authorization.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `semantic-judgment-shadow`: add a separate prospective MFI observational evaluation contract; retain the existing shadow and offline-quality contracts.

## Impact

Future work may add an offline analysis path and append-only evidence artifacts, with a separately authorized collection path that preserves precise source timestamps. No trading, engine, r1, selection or historical ledger behavior changes are proposed.
