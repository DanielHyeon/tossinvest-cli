# Review: fix-adoption-console-truth

## Proposal freeze round 1 — 2026-07-31

- Voice: separate read-only Codex session, product/scope + UI/accessibility +
  adversarial engineering/security + operator/DX.
- Verdict: **REJECT**.
- Confirmed production facts: console/engine journal path split, v8
  `exit_states.policy_id` query failure, CSP-blocked adoption slider, and
  fraction-at-next-start configuration semantics.
- Blocking findings:
  1. Active predecessor still required a mouse slider and prohibited typing.
  2. NaN/infinity, full half-percent grid, deterministic formatting, legacy
     off-grid values, and real config round-trip were unspecified.
  3. Successful journal scenario claimed unimplemented schema display; selected
     v8 column compatibility and two-database/no-fallback tests were absent.
  4. "No client-side script" was broader than the adoption control.
  5. Docker recreation would stop and possibly autostart the engine, violating
     the implementation change's lifecycle boundary.

## Manager disposition

- Archived completed predecessor `console-adoption-controls` after syncing its
  exit-policy and operator-console deltas to main specs. STORY-TOS-025 now
  points to that archive.
- Replaced the slider rule through a full `MODIFIED` 편입 설정 requirement.
- Added finite percentage tick semantics, deterministic display, legacy
  off-grid behavior, full-grid and real config round-trip tests.
- Added pre-query v9 required-column validation with typed too-old reporting;
  no migration, fallback, or query downgrade.
- Narrowed CSP scope to the adoption control.
- Limited this change to image build; production recreation is a separate
  human-authorized lifecycle operation.
- Corrected rollback wording: data-safe, but old image reintroduces both defects.

## Proposal freeze round 2 — 2026-07-31

- Voice: separate read-only Codex session, restricted to Story/OpenSpec
  artifacts.
- Verdict: **APPROVE**.
- Disposition: all five round-1 blockers resolved; Story↔OpenSpec 1:1 is sound;
  implementation may start.

## Pre-Edit Gate

- change id / task id: `fix-adoption-console-truth` / 3.1, 3.2
- 대상 심볼:
  - `cmd/tossctl.runConsole`
  - `internal/console.(*Console).handleSettingsSave`
  - `internal/console.settingsPage.StopPctSlider`
  - `internal/console.settingsPage.StopPctPercent`
  - `internal/journal.(*ReadOnly).checkSchema`
- CodeGraph definition/callers/callees/impact:
  - `runConsole` is called only by `newConsoleCmd`; impact is confined to
    console assembly. Its current callees include `journal.DefaultPath`,
    `engineJournalDir`, and `console.ListenAndServe`.
  - `handleSettingsSave` is reached only by `ANY /settings/save`; impact is the
    adoption settings route and injected `AdoptionSettings` seam.
  - `checkSchema` is called only by `OpenReadOnly`; impact reaches read-only
    console positions/history and read-only tests, never the writable Journal.
- CodeGraphContext 후보와 evidence reconciliation: the advisory report emitted
  no additional candidate beyond the hard-evidence path/settings/schema
  symbols. See `analysis/code-context/`.
- 기존 동작 파악 근거: base commit `2dd1fe5`, current HEAD, CodeGraph source,
  existing `console_test.go`, `settings_test.go`, `readonly_test.go`, and
  production observations recorded in proposal/design.
- Function Logic Map / Branch Test Map: five symbol directories under
  `analysis/function-logic/`; new leaf helpers are covered by their callers'
  maps and direct tests.
- upstream 상속 테스트 영향: yes. Default profile path, settings field
  preservation/CSRF, read-only no-write/no-migration, and schema-direction
  tests remain mandatory.
- 실패 테스트 선행 작성: yes; task 2.1/2.2 before production edits.
- 설정·DB·journal 변경과 rollback: isolated tests write temporary config/DB
  only. Production code changes path selection and read-only validation; no
  schema migration/write. Config representation remains fractional. Image
  rollback is data-safe but reintroduces both UI defects.
- 안전 불변식 §0 위반 여부 검토: 통과. No engine/soak lifecycle call, toggle
  flip, broker call, account mutation, or LIVE order is permitted in tests.

## Implementation review round 1 — 2026-07-31

- Voice: separate read-only Codex session using the code-reviewer rubric.
- Verdict: **REJECT**.
- Blocking findings:
  1. The console resolver trimmed whitespace while the engine treats every
     non-empty `configDir` as explicit, so the two profile rules were not
     identical.
  2. Tests did not yet prove selected-journal behavior with two real databases,
     the remote `/settings` CSP response, or the real `7.5 → 0.075` HTTP/config
     round trip; one invalid-save test still posted the obsolete field.
  3. Legacy off-grid detection used a floating tolerance inconsistent with the
     exact server parser.
  4. `git diff --check` found an extra trailing blank line in the synced main
     exit-policy spec.

## Implementation review round 1 remediation

- Made `consoleJournalPath` follow the engine's exact non-empty profile rule,
  including whitespace-only values.
- Added real temporary journals proving selected-profile identity, no fallback,
  and positions rendering from only the selected journal.
- Added an authenticated remote `/settings` response test that preserves the
  script-free CSP and numeric percentage control.
- Replaced the synthetic config assertion with a real authenticated HTTP save
  through the production config seam and corrected the invalid-save field.
- Reused the exact percentage parser for legacy correction detection and added
  near-grid coverage.
- Removed the trailing blank line and reran focused regressions plus the
  logic-map completeness checker successfully.

## Implementation review round 2 — 2026-07-31

- Voice: separate read-only Codex session after round-1 remediation.
- Verdict: **APPROVE**.
- Disposition: all prior blockers are resolved. The resolver now matches the
  engine's exact profile rule; real-journal tests prove identity and no
  fallback; remote CSP remains deny-by-default and script-free for the control;
  the authenticated production HTTP/config seam proves `7.5 → 0.075`; and
  legacy correction reuses the exact parser.
- Residual risk: future drift in journal selection, CSP wiring, or parser reuse.
  The focused integration tests pin all three contracts.

## Implementation review round 3 — 2026-07-31

- Voice: separate read-only Codex session explicitly rechecking every round-1
  blocker end to end.
- Verdict: **REJECT**; this supersedes round 2's incomplete approval.
- Blocking finding: `consoleJournalPath` preserved a whitespace-only explicit
  `configDir`, but `Console.openJournal` later applied `strings.TrimSpace` to
  the selected path. A relative `" "` profile therefore changed from
  `" /journal.db"` to `"/journal.db"` before the read-only open. The resolver
  test covered only the first half and did not expose the downstream identity
  change.
- Other contracts were accepted: selected-only/no-fallback behavior for normal
  paths, v8 pre-query classification, authenticated remote CSP, real HTTP
  `7.5 → 0.075`, exact off-grid refusal, and lifecycle safety.
