# Supporting context

Date: 2026-07-31

## StockOS methodology reference

`/mnt/D/project/axipient/stockos/.claude/CLAUDE.md` defines the shared Full SDD
sequence and PM contract:

1. memory recall
2. Story + approved OpenSpec and READY
3. CodeGraph hard evidence
4. CodeGraphContext supporting context
5. evidence reconciliation
6. Function Logic Map when existing function logic changes
7. Pre-Edit declaration
8. RED → GREEN → REFACTOR → VERIFY
9. gstack/review/security/QA and project gates
10. archive + PM sync
11. verified episodic then canonical memory promotion

It also requires one Delivery Story per active OpenSpec change and derives delivery
state from repository evidence rather than treating generated trackers as authority.

## TossOS context that must remain local

- Go AST Function Logic Map tooling and Go test/race/vet gates
- Toss official Open API mutation boundary and WTS read-only/best-effort boundary
- upstream OFF behavior and inherited regression suite
- journal single-writer, Guardian, reconciliation, and authenticated operator controls
- TossOS graph/memory namespaces and Docker/Make deployment entry points

StockOS portfolio IDs, KIS/Python commands, and runtime facts are reference material,
not TossOS evidence.
