## 1. Contract and Evidence

- [x] 1.1 Record StockOS/TossOS desktop and mobile reference findings and approve the input-free information hierarchy
- [x] 1.2 Add RED rendering tests for shared dashboard/positions primary fields, canonical unknown handling, disclosures, US parity, and responsive/accessibility contracts
- [x] 1.3 Create Function Logic Maps and Branch Test Maps for every modified existing Go function

## 2. Shared Projection

- [x] 2.1 Extract position settings/runtime/lifecycle/exit enrichment into one shared read-only helper
- [x] 2.2 Retain joined position rows in the dashboard account panel without adding broker calls
- [x] 2.3 Verify dashboard and positions render identical management and exit-line facts from the same inputs

## 3. Simple Holdings UI

- [x] 3.1 Create one reusable holdings template with the requested primary fields and line semantics
- [x] 3.2 Move provenance, raw journal evidence, identities, and secondary diagnostics into accessible per-row details
- [x] 3.3 Render the shared holdings table on dashboard and positions for both KR and US
- [x] 3.4 Tune desktop column density and 375 px card layout while preserving focus, touch-target, CSP, and input-free contracts

## 4. Verification and Delivery

- [x] 4.1 Run focused console tests, function-logic checks, and full `make sdd-check`
- [x] 4.2 Run OpenSpec strict validation, independent review, and `make gate CHANGE=a057-simplify-portfolio-views`
- [x] 4.3 Browser-QA desktop/mobile dashboard and positions, including no horizontal overflow and disclosure behavior
- [x] 4.4 Commit the worktree, merge to local main, push origin main, deploy, and verify the deployed pages
- [x] 4.5 Archive the OpenSpec change, refresh PM trackers/memory, and push the archive commit
