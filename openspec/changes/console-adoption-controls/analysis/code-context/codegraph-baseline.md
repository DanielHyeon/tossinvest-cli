# CodeGraph baseline — external-position automatic-management menu revision

- Date: 2026-07-31
- Frozen base: `1dbef864038c81f0a2982a03e7d9549369d21669`
- Query: external/manual position adoption, console settings, reconcile adoption,
  and policy handoff.
- Existing implementation:
  - `config.Adoption.Enabled`, `DefaultStopPct`, `ExcludeSymbols`, and
    `IncludeSymbols` are the persisted control surface.
  - `ReconcileDriver.judgeHoldings → adopt → adoptOne` is the sole automatic
    adoption execution path.
  - `adoptOne` snapshots the current common policy ID, records the adoption, and
    opens the adopted exit state without issuing an order as part of adoption.
  - `Console.handleSettings` and the existing `/settings` routes expose only the
    config seam; they do not obtain journal-adoption or broker-mutation authority.
- Target impact:
  - `settingsTemplates` affects only its template source.
  - `baseTemplates` affects `pageTemplates` and its template source.
- Decision: change only template constants and rendered-string tests. No config,
  reconcile, exit-policy, journal, engine, route, or order symbol is in edit scope.
