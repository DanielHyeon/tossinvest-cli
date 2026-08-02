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
11. Post-push UI review found one managed KR row that was also still present in desired include and therefore received both `엔진 관리` and the desired-only `편입 예약됨` fallback. A RED production-shape fixture reproduced the contradiction.
12. `positionRow.PendingDesignation` now centralizes the fallback boundary for the template, row reason, and exit-reference adapter: broker-present, unmanaged, non-released, journal-known, designated, non-excluded, and without an engine projection. Focused normal/race/vet passed; the redeployed managed KR row retains its legacy reference with zero pending/runtime-unknown copy, while both US rows retain their typed price-free adoption plan.
13. Final logic review found that a known `RELEASED` lifecycle could still be revived visually by a stale desired include entry. A RED truth table reproduced both that error and a broker-absent designation label leak.
14. The shared pending predicate now excludes known releases, while `Label` gives release explicit priority and delegates candidate copy to that predicate. The truth table covers desired-only, managed active/waiting/completed, released, excluded, journal-unknown, broker-absent, and typed-projection states. A production-shape released+desired US HTML test additionally binds label, reason, template, and exit-reference behavior. The matching OpenSpec scenario and Function Logic/Branch Test maps are current.
15. Final independent verdicts after the release fix: Security 0 findings approved; Test/Maintainability P0/P1/P2 = 0/0/0 and archive approved; StockOS-informed UI/UX approved with no blocking issue.
16. `GOFLAGS=-buildvcs=false make gate CHANGE=a053-restore-visible-exit-line-references` passed all eight stages. The flag is the documented linked-worktree workaround for Go VCS stamping and does not alter tested code.
17. The rebuilt deployment is healthy. HTTPS HTTP/2 returned 200 for `/positions`, `/position-management`, and `/api/v1/positions`; KR managed rows show validated `LEGACY_RAW` KRW references, US desired rows show price-free `RUNTIME_UNKNOWN` USD references while the runtime seam is unavailable, unsupported-US phrases are absent, and `/positions` has no visible input surface.

## Verdict

Approved. No live order, reconcile mutation, journal write, operating-toggle change, or input surface was added.
