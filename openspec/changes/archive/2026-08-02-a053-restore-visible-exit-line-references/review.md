# Review: a053-restore-visible-exit-line-references

- Date: 2026-08-02
- Voices: Security, Test/Maintainability, UI/UX, Manager

## Proposal-freeze findings

1. Security classified any broker average/current-price derived candidate stop or target as P0 forbidden. Candidate rows may show only `runtime.Effective.DefaultStopPct`, never desired/default fallback or a synthetic price.
2. Legacy baseline and initial stop may be visible only as `저장된 원장 기준선 · 현재 실효 미확인`; they must not populate current protection, next target/protection, rung, quantity, or order authority.
3. Stale canonical behavior stays fail-closed and is not widened to a primary raw-price reference in a053.
4. Test/maintainability review found a lifecycle-generation join gap. Known mismatches must suppress canonical/raw prices and identities in both console and API.
5. StockOS `OrdersPanel` informed an always-visible vertical `익절 / 손절 / 기준` stack, market/currency badge, tabular values, and text-bearing semantic pills. StockOS's hide-on-missing behavior and 10px labels are not copied.
6. Accessibility uses text plus colour, explains every dash, keeps core facts outside `<details>`, preserves the 375px card layout, and adds no `aria-live` to the meta-refreshing page.
7. Proposal-freeze verdict: approved after the above scope corrections. No order, reconcile, journal, or operating-toggle mutation is authorized.

## Verification evidence

1. RED tests failed on undefined shared reference types/API fields and the existing primary-row dashes before production code was added.
2. Focused normal tests passed for `internal/operatorview`, `internal/positionpolicy`, `internal/httpapi`, `internal/console`, and the a053 `cmd/tossctl` reader cases.
3. Focused race tests and `go vet` passed. `GOFLAGS=-buildvcs=false go test ./...` passed; the flag only works around Go VCS stamping incorrectly running `git status` from `/tmp` in a linked worktree.
4. Corrupt/partial snapshot, lifecycle-generation mismatch, required-but-unverified lifecycle, runtime unavailable, valid/corrupt released evidence, KR/US legacy, and US pending/blocked cases are explicitly covered.
5. `make sdd-sync` completed. `make sdd-check` passed after installing the repository-declared `typedb-driver` into `.sdd/.venv` with `make sdd-infra`.
6. Function Logic Map checker, strict OpenSpec validation, PM tracker check, JSON validation, and `git diff --check` passed.
7. Final independent verdicts: Security P0/P1/P2 = 0/0/0 approved; Test/Maintainability approved; StockOS-informed UI/UX approved.
8. Deployment canary reproduced a production-composition gap: the HTTP API normalized an unavailable runtime to `RUNTIME_UNKNOWN`, while the console's absent combined commander left the row status empty. A RED fixture now covers `PositionPolicies=nil` plus a desired-included US holding; the console normalizes only that non-excluded desired row to typed unknown without using the desired percentage.
9. A follow-up UI review removed the contradictory legacy detail text from both typed pending/blocked and commander-unavailable desired rows. The page now distinguishes a stored inclusion request, runtime reflection, and effective protection.
10. Docker services were rebuilt from the integrated local `main`. HTTPS HTTP/2 canaries returned 200 for `/positions`, `/position-management`, and `/api/v1/positions`; two KR rows exposed validated `LEGACY_RAW` KRW references, two US rows exposed price-free `RUNTIME_UNKNOWN` USD references, unsupported-US phrases were absent, and `/positions` contained no form/input/select/textarea/button/contenteditable surface.

## Verdict

Approved. No live order, reconcile mutation, journal write, operating-toggle change, or input surface was added.
