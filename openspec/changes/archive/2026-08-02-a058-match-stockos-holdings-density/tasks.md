## 1. Contract and Evidence

- [x] 1.1 Record measured StockOS/TossOS desktop typography, row height, padding, and column widths
- [x] 1.2 Add a RED rendering regression test for explicit seven-column layout, compact type, numeric alignment, and folded verbose diagnostics
- [x] 1.3 Record `Function Logic Map: not-applicable` because production changes are template/CSS constants only and no existing Go function body changes

## 2. Compact Holdings UI

- [x] 2.1 Add one explicit seven-column `colgroup` shared by dashboard and positions
- [x] 2.2 Match StockOS-like header/body typography, padding, line height, and numeric axes within the holdings table only
- [x] 2.3 Keep concise management/evidence verdicts visible and move verbose management/exit explanations into the existing details disclosure
- [x] 2.4 Preserve KR/US parity, input-free markup, native disclosure, and 375 px card flow

## 3. Verification and Delivery

- [x] 3.1 Run focused console tests, OpenSpec strict validation, `make sdd-sync`, and `make sdd-check`
- [x] 3.2 Run independent design/code/security review and `make gate CHANGE=a058-match-stockos-holdings-density`
- [x] 3.3 Browser-verify dashboard/positions desktop column metrics, row density, 375 px overflow, and details behavior
- [x] 3.4 Commit the worktree, merge to local main, push origin main, deploy, and verify the deployed routes
- [x] 3.5 Archive the OpenSpec change, refresh PM trackers/memory, and push the archive commit

## 4. Trailing Detail Popup Follow-up

- [x] 4.1 Add a RED regression for a trailing detail column, tighter current/line scan, and script-free popup
- [x] 4.2 Move detail out of PnL into the final column and render its content as a centered native disclosure popup
- [x] 4.3 Verify desktop spacing, popup open/close, 375 px overflow, console errors, and dashboard/positions parity
- [x] 4.4 Run independent UX/security/test review, SDD sync/check, and the full a058 gate
- [x] 4.5 Commit, merge to local main, push, redeploy, verify, and archive the change
