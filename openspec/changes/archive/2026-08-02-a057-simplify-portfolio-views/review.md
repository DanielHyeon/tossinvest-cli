# Design Review — a057 simplify portfolio views

## Scope

Existing-site UI review of `/dashboard` and `/positions`, using the live StockOS pages supplied by the user as the information-hierarchy reference. TossOS visual identity, safety wording, evidence rules, and read-only boundaries remain authoritative.

## Reference findings

- StockOS keeps each holding scan to instrument, quantity, average/current price, a five-line protection stack, market value, and unrealized PnL.
- TossOS already has the data, but its primary row mixes management verdict, provenance, diagnostics, identifiers, and prices; the dashboard has only market aggregates.
- The desired workflow is glance first, inspect second. No typing or settings action belongs in these views.

## Findings and dispositions

| Severity | Finding | Disposition |
|---|---|---|
| P1 | Dashboard and positions can interpret lifecycle/effective evidence differently because enrichment exists only in the positions handler. | Share one read-only enrichment helper before exposing dashboard rows. |
| P1 | Actionable and raw stored lines must not be visually conflated. | Primary line fields consume only canonical projection; raw evidence stays in details with `실효 미확인`. |
| P2 | The current five-column positions row has too much prose and too few direct broker facts. | Replace with the requested primary scan columns and one per-row disclosure. |
| P2 | Seven desktop columns risk mobile overflow. | Reuse the accessible `.data-table` card transformation and require 375 px browser verification. |
| P2 | Management blocked/stale/unknown states could be hidden with the diagnostics. | Keep verdicts and reasons outside disclosure; fold provenance and identities only. |
| P3 | A full StockOS visual copy would weaken TossOS consistency. | Borrow information hierarchy only; retain TossOS typography, colors, navigation, and CSS primitives. |

## Accessibility and safety review

- Native `<details>/<summary>` provides keyboard and no-script disclosure.
- Summary target must be at least 44 CSS pixels with visible focus.
- Table caption, scope headers, mobile `data-label`, color-independent labels, and `—` unknown markers remain mandatory.
- No input, POST form, script, broker mutation, operating-toggle action, or order route is introduced.

## GSTACK REVIEW REPORT

**Runs:** 1 | **Mode:** Existing-site design plan review | **Status:** Approved

Current information hierarchy: 4/10. Approved target: 9/10.

NO UNRESOLVED DECISIONS

## Independent implementation reviews

### Security

Approved with P0/P1/P2/P3 = 0/0/0/0. The diff adds no LIVE order,
configuration save, operating-toggle, route, or broker-refresh capability.
Both markets share the same fail-closed projection; raw and stale evidence is
not promoted into primary actionable lines; no form, input, script, CSP, secret,
or authentication surface changed.

### Maintainability and tests

The first review blocked on three P1 findings. All were fixed and independently
re-reviewed:

- the dashboard now explains broker-only holdings whose readable journal has no
  matching position;
- the already-masked account reference remains available inside each holding's
  disclosure, including the shared dashboard projection;
- current AST, Function Logic Maps, Branch Test Maps, and direct callees agree.

Final verdict: approved with no P0/P1. Non-blocking follow-ups remain a larger
state-parity matrix, potential later decomposition of the shared decorator, and
real-browser 375 px overflow verification in the delivery canary.

### UI/UX and accessibility

The first review blocked on mobile grid auto-placement: cells with multiple
children could alternate content across the label and value columns. The mobile
contract now pins the generated label to column one and every direct content
child to column two. The five-line stack remains internally paired. Line text
was raised to the table's 0.9 rem size and diagnostic details changed from two
narrow columns to one. Focus, 44 px summary target, native disclosure, semantic
headers, and the input-free contract remain intact. Re-review approved with no
P0/P1; deployed 375 px density/overflow remains a canary item.

## Verification record

- RED observed for the missing seven-column hierarchy, dashboard projection,
  raw-evidence boundary, dashboard broker-only explanation, and masked account
  disclosure.
- `go test ./internal/console -count=1`: PASS.
- focused `go test -race`: PASS.
- `go vet ./internal/console`: PASS.
- Function Logic Map checker: PASS.
- `openspec validate a057-simplify-portfolio-views --strict`: PASS.
- `make sdd-check`: PASS after installing the repository-declared SDD Python
  environment and refreshing the worktree-bound CodeGraph index.
- Worktree note: Go 1.25 resolves VCS metadata from `/tmp` for linked worktrees;
  the full gate therefore uses `GOFLAGS=-buildvcs=false`. The same nested real
  binary tests pass with that VCS-stamping-only override.

## Deployed browser canary

Main commit `f700fd70` was rebuilt and deployed through the existing hardened
Docker Compose project. Both `tossos` and `httpapi` reported healthy.

Chromium verified `/dashboard` and `/positions` at 1440×1000 and 375×812:

- document `scrollWidth` equaled viewport width (1440 and 375 respectively);
- all seven requested headers rendered in both screens;
- four deployed holding rows each rendered exactly five line items;
- there were no visible inputs, forms, textareas, selects, or scripts;
- the first native disclosure opened successfully in all four runs;
- KR and US rows shared the same hierarchy and US rendered `US · USD`.

Visual review of the four full-page screenshots confirmed the mobile label/value
flow and readable single-column disclosure. No deployment or UI blocker remains.
