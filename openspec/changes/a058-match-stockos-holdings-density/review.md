# Review — a058 match StockOS holdings density

## Proposal freeze — 2026-08-03

Voices: Manager self-review, rendered design comparison, safety review.

### Evidence

- StockOS desktop measurement: 10 px headers, 12 px cells, 15/18 px line-height, 8×12 px padding, widths 355/72/122/130/201/143/129 px, 96 px holding row.
- TossOS desktop measurement: 14.4 px headers/cells, 23.04 px line-height, 12×10.4 px padding, seven equal 159 px columns, first holding row 268 px.
- Root cause: fixed-layout widths are assigned only to body `.position-row` cells while the header is the first layout row.

### Findings and decisions

1. **High — header/value axes do not match the reference.** Accepted: use a seven-column `colgroup` so header and body share width definitions.
2. **High — verbose primary diagnostics destroy scan density.** Accepted: keep concise safety verdicts visible; fold verbose reason/explanation text into the existing native detail disclosure.
3. **Medium — type and padding are materially larger than StockOS.** Accepted: scope compact 10/12 px typography and 8×12 px padding to `.positions-table` only.
4. **Safety — missing/stale evidence must remain fail-closed.** Accepted unchanged: no line inference, concise visible verdict plus `—`, raw evidence in details.
5. **Security — no new inputs, script, route, broker call, mutation, or CSP change.** Accepted.

Function Logic Map: not-applicable — production edits are limited to inline CSS/template constants; no existing Go function body is modified.

Status: approved for implementation.

## Implementation verification — 2026-08-03

### Independent reviews

- Independent code review found that a broker-only excluded holding could lose
  its short policy explanation when no detail disclosure was rendered. The
  explanation now remains beside the compact verdict; the existing excluded-row
  regression and the focused a058 tests pass.
- Independent maintainability review found that fallback `PendingDesignation`
  could show only `관리 편입`, obscuring that protection was not yet effective.
  The primary row now keeps the compact `편입 예약 · 실행 미확인` verdict while
  the longer warning remains inside details, with a direct primary/detail boundary
  regression.
- Independent security review reported P0/P1/P2/P3 = 0/0/0/0. The change adds no
  order or settings mutation, route, script, input, external asset, or CSP change;
  primary actionable lines still consume only the fail-closed `ExitLine` projection.
- The review's browser-test gap is covered by direct rendered-browser geometry
  evidence below and the repository's existing status truth-table regressions.

### Browser evidence

- Desktop `/positions` and `/dashboard` both render the same 1115 px table with
  columns 346/67/123/123/201/134/123 px, 10/12 px header/body text, 15/18 px
  line-height, 8×12 px cell padding, and a 104 px first holding row.
- At 375 CSS px, `documentElement.scrollWidth === innerWidth === 375`; the native
  summary is 44 px high, verbose evidence is absent while closed and present
  after opening, and the browser console reports no errors.
- The five primary line values remain visible without opening details, while the
  verbose reconciliation and exit-reason prose stays folded.

### Automated evidence

- `go test ./internal/console -run '^(TestTheStatusColumnHeaderSaysAdoption|TestAnExcludedRowIsLabelledAndOffersRelease|TestA058.*)$' -count=1`
- `go test ./internal/console -count=1`
- `openspec validate a058-match-stockos-holdings-density --strict --no-interactive`
- `git diff --check`

Independent review status: approved — security and maintainability reviewers
reported P0/P1/P2/P3 = 0/0/0/0 after the excluded and pending primary-verdict
regressions were corrected.

## Trailing detail popup follow-up — 2026-08-03

### Decisions

- Move the disclosure from the PnL cell into a narrow final `상세` column.
- Change the line stack to `space-between` so labels start beside the current-price
  column while prices retain a stable right numeric axis.
- Render the native details disclosure as a centered, content-sized, non-modal
  popup. Keep it input-free and script-free, cap its height to the viewport, and
  keep a sticky 44 px summary available as the close action.
- Scope every open-state rule through `.detail-cell`; order trace disclosures must
  remain ordinary inline details.

### Independent review

- Security initially found a P1 click-through risk in a modal-looking backdrop.
  The backdrop was removed and the contract made explicitly non-modal. Final
  security result: P0/P1/P2/P3 = 0/0/0/0.
- UX initially found a 34 px open close target and a forced 672 px popup height.
  The target is now 44 px and the popup uses content height with only a max-height.
  Long desktop detail measures 610.14 px; short IONQ/TSLA detail measures 339.14 px.
- Test review found popup selectors leaked to `/orders`; all new selectors are now
  scoped to `.detail-cell`. The order disclosure regression and full console suite pass.
- Nonblocking follow-up: independent non-modal `details` can be opened concurrently.
  Exclusive grouping can be considered with a future markup-contract revision.

### Browser and automated evidence

- Both `/positions` and `/dashboard`: eight headers/cells, final `상세` column,
  24 px current-value to first-line-label gap, 95.75 px row height closed/open.
- 375 px: document scroll width 375 px, row height 596.44 px closed/open, popup
  within viewport, 44 px sticky close action, no console/page error.
- 375×400: popup content scrolls to the end while the sticky close action remains visible.
- Holdings rows contain no script or input; actionable lines still use the unchanged
  fail-closed `HasExit`/`ExitLine` projection.
- Focused a057/a058 tests, full `internal/console`, race, vet, operatorview,
  positionpolicy, OpenSpec strict validation, and `git diff --check` pass.

Independent review status: approved; no P0/P1 deployment blockers.
