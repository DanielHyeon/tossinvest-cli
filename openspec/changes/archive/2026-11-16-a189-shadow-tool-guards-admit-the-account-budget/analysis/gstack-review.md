# gstack review 원문 (a189, 2026-11-15)

- 스킬: gstack `review`(preamble 4e82fa4c839f). SESSION_KIND: interactive, REPO_MODE: solo. 리뷰어(부모)는 구현 팀메이트 Opus — 작성자와 같은 계열이므로 바깥 보이스로 Codex(gpt-6-astra)를 함께 돌림.
- 대상: base `7d1e3163a`(merge-base origin/main) .. HEAD `d929433ba`(리뷰 시작 시점, 작업 트리에는 빈 `mutation-after-fixture.txt` 의 완성본만 미커밋). REVIEW_START `cdf598ed-43d8-4586-ad1c-a5f114c862c7`.
- 규모: DIFF_LINES 5370(대부분 증거·문서), Go 59줄. SCOPE_BACKEND·TESTS·CONFIG·DOCS true, STATE_MACHINE false.
- 디스패치: specialists 5 — testing, maintainability, security, performance, simplification(DIFF_LINES>100). Red Team(DIFF_LINES>200). Claude 적대 패스 1(PASS_START `98e17c9b-…`). Codex 적대 패스 1(CODEX_MODE ready, PASS_START `9b155fd5-…`). Codex 구조 리뷰는 DIFF_TOTAL≥200 이라 해당되지만 돌리지 않음(아래 「안 한 것」).
- Plan completion audit: not run (no plan is bound to this branch and no docs/designs/ file matches). 의도 출처는 커밋 메시지와 OpenSpec proposal.
- Banned terms: none. Approval classes: none. PR 없음(봇 코멘트 없음).

## Gate scan 원문

```
GATE_SUMMARY: listed=7 eligible=3 inspected=3 unread=0 tagged={} unmatched=56 coverage=RH-1,RH-3,RH-4,RH-12,RH-13,RH-14,RH-15,RH-16 languages_unlisted=none candidate=7d1e3163a704+d65a8368602e
[80d9fc903a8e] untagged internal/official/a112_mbus_static_test.go @@ -494,7 +494,8 @@ func a112MBUSReferenceAllowed(path string) bool {
[cfe5b07efdd8] untagged tools/a128-jev-shadow-probe/static_test.go @@ -169,12 +186,15 @@ func TestImportsStayInsideTheAllowlist(t *testing.T) {
[976f27364b5d] untagged tools/a140-breakout-signal-shadow/static_test.go @@ -82,10 +82,15 @@ func TestNoMutatingAccountOrWriteSurfaceIsNamed(t *testing.T) {
```

## Specialists 원문

### testing
```
{"severity":"INFORMATIONAL","confidence":8,"path":"tools/a140-breakout-signal-shadow/static_test.go","line":101,"category":"testing","summary":"The new accountbudget/accountbudgetruntime allowlist checks match the selector's receiver by its literal identifier (`allowed[pkg.Name]`, and `allowedBudget[pkg.Name]` in a128 static_test.go:128), not by the import path. TestImportsStayInsideTheAllowlist now accepts both import paths under any local name. So a sibling non-test file (both packages have many, e.g. collect.go and shadow.go) can write `ab \".../internal/accountbudget\"` and then `ab.WithPriority(ctx, ab.Critical)` without being checked. main.go still uses the canonical name, so the 'saw nothing' check stays true and every guard passes. The harness mutant `*-f-alias-runtime` aliases every use inside main.go, so it only proves the all-aliased case and does not support the general claim that aliasing is caught. Raising the shadow tool to Critical priority on the shared account budget is the outcome the Auxiliary-only allowlist exists to prevent (stop-loss immediacy). The same weakness already existed for official/tossconfig. The a112 guard, by contrast, resolves aliases from the import specs.","fix":"Resolve each tracked package's local name from file.Imports (as a112MBUSFindForbiddenRefs does): build alias->path per file, look up the allowlist by import path, and reject dot-imports and non-canonical names for tracked paths. Add a harness mutant that adds an aliased import plus `ab.Critical` to a second non-test file while main.go stays unchanged.","fingerprint":"tools/a140-breakout-signal-shadow/static_test.go:101:testing","specialist":"testing","cause":{"path":"tools/a140-breakout-signal-shadow/static_test.go","line":137}}
(test_stub 생략 — 별칭 import 를 지역 이름으로 고정하는 TestTrackedPackagesAreImportedUnderTheirOwnName 골격)
```

### maintainability
```
NO FINDINGS
```

### security
```
Fixtures and test-corpus files were reviewed in summary mode only (`git diff --stat`). Their raw payload bytes were not pulled into adversarial reasoning. Because a189 changes only test files, I read the four guard Go test diffs in full as the code under review. The other changed files are fixture-like analysis artifacts (FLM ast.json, mutation fixtures, spec). I covered those by stat only.

NO FINDINGS
```

### performance
```
NO FINDINGS
```

### simplification
```
NO FINDINGS
```

## Red Team 원문

```
Fixtures and test files were reviewed in summary mode (diff --stat). The exceptions were the Go static-guard test diffs and the spec delta, which the caller named as the code under review and which I read in full. Targeted guard tests pass on HEAD d929433ba and gofmt -l is clean.

1. [INFORMATIONAL c8] analysis/mutation/mutation-after-fixture.txt:1 evidence-integrity — The after-fixture mutation receipt at HEAD is an empty file (blob e69de29). The real 110-line output exists only as an uncommitted worktree change. Fix: commit the completed receipt, note it was measured at 6e6b29a5.
2. [INFORMATIONAL c7] tools/a128-jev-shadow-probe/static_test.go:82 spec-vs-guard — The forbidden-by-name list was not extended to the method surface a166 brought in. *Factory.Budget() returns *accountbudget.Budget, which has EnginePollKey(accountRef, kind), ObservationSnapshot, PollRemaining, NewObservationFence and Acquire. None is forbidden in either tool. The spec delta scenario says names outside the list and account surface names "still fail the check", which is false for these method names. Fix: add Budget, Acquire, EnginePollKey, PollRemaining, ObservationSnapshot, NewObservationFence and VerifyIdentity to both forbidden lists; add a mutant that plants factory.Budget().
3. [INFORMATIONAL c6] specs/breakout-retest-strategy-lane/spec.md:10 spec-vs-runtime — The spec says budget participation "only sets the send permission and timing of read requests". In ModeRequired, factory.NewClient wraps the client in accountbudget.Transport (OpenObserved); RoundTrip calls obs.Begin/obs.Finish on every request (kind ObservationOther), writing the account-wide observation store the engine-side consumer reads. Not reconciled with "evidence writes only to its own store / no engine journal writes". Fix: name the shared budget/observation state as the one permitted write outside the tool's own store, or add a test that auxiliary non-poll requests cannot change what the engine sees for engine keys.
4. [INFORMATIONAL c5] specs/breakout-retest-strategy-lane/spec.md:10 fail-open-default — With neither TOSSOS_ACCOUNT_BUDGET_MODE nor _MANIFEST set, ParseLaunchConfig returns ModeOff and the tool runs unbudgeted. The auxiliary-priority protection is opt-in; pre-existing a166 behaviour that a189's spec makes official. Fix: state that ModeOff is unbudgeted and the default, or refuse ModeOff when the engine manifest exists; record the launch requirement in docs/operations.md.
```
(Red Team 원문은 JSON 한 줄씩이었고 위는 필드 그대로 번호만 붙여 옮김.)

## Claude 적대 패스 (발견 본문)

```
Coverage. Fixtures were reviewed in summary mode, so this coverage is reduced on purpose. For test and fixture files I used `git diff --stat` only. ... I read in full: the three static-guard diffs; mutation.py, the spec delta and the relevant parts of design.md and proposal.md. To reason about what the guards protect I also read the tools' main.go, internal/accountbudget, internal/accountbudgetruntime/{runtime_linux.go,launch_config.go}, internal/official/{client.go,token.go,account_budget.go} and a166_budget_test.go.

1. The guard admits a priority escalation on the shared account budget without failing. FIXABLE.
   - allowedBudget (a128 static_test.go:102, a140 static_test.go:92) only checks selector names whose X is the package identifier. Three paths pass every static test:
     (a) accountbudget.WithPriority(ctx, 1). The untyped constant 1 is Critical (budget_linux.go: Normal=0, Critical=1, Auxiliary=2), and a literal is not a selector.
     (b) b, _ := factory.Budget(); b.Acquire(ctx, 1). Factory.Budget() (runtime_linux.go:284) hands back the shared *accountbudget.Budget. Neither Budget nor Acquire is in forbidden, and receiver calls on local variables are never tracked.
     (c) holding leases with the same method to starve the queue.
   - The only behavioural backstop is a166AuxTransport in a166_budget_test.go. It checks the priority of the first request only (the tests assert calls == 1).
   - Failure scenario: a later edit to a shadow tool claims Critical or Normal on the shared budget. The engine's stop-loss and emergency-exit requests then queue behind the shadow tool's reads (safety invariant 4), and make test/gate stays green.
   - The spec delta says "그 연결은 예산 시작과 보조 우선순위 표시, 그 예산에 묶인 공식 클라이언트 생성으로 한정되어야 하며 (SHALL)". The guard enforces that only for package-qualified names, so the SHALL is not actually pinned.
   - Fix: AST-pin every accountbudget.WithPriority call so its second argument is exactly the selector accountbudget.Auxiliary. Add Budget, Acquire, NewObservationFence, OpenObserved, OpenContext, Provision, ProvisionOffline, NewTransport, EnginePollKey, PollRemaining, ObservationSnapshot and VerifyIdentity to both forbidden maps (grepped: none appear in the tools' non-test sources). Add matching mutants.
   - Severity: Medium. Blocks: no (current tool code uses accountbudget.Auxiliary correctly).
2. The alias bypass (R1) now also covers the two budget packages, which design.md says it does not. FIXABLE (decision INVESTIGATE). Before a166/a189 the two packages could not be imported at all; now an aliased import skips the name allowlist. The saw-check only catches the all-aliased case; Go allows importing the same path twice, so keeping the canonical use and adding ab.WithPriority(ctx, ab.Critical) passes. Fix: key allowlists on import path, or at minimum correct the R1 sentence in design.md. Severity: Medium. Blocks: no.
3. The M-B0 admission exempts the whole file, not one symbol (acknowledged R2). Today it has exactly one reference (official.A112MBUSCandle, line 463). Fix: allow that path with an exact name set. Severity: Low. Blocks: no.
4. A mutant counts as CAUGHT if any test in the package fails (mutation.py lines 150, 157–161); failing test names are printed but never asserted. f-no-accountbudget was also caught by a behavioural test. Fix: give each mutant its expected --- FAIL: <TestName>. Severity: Low. Blocks: no.
5. The harness runs tests that read the operator's real token cache, with no HOME/XDG isolation (pre-existing a166 tool tests call runCommandWithOptions without --config-dir → ~/.cache/tossctl/openapi-token.json). DialContext keeps it on localhost and the 500 means no cache write; still host-dependent. Fix: set HOME and XDG_CACHE_HOME to temp dirs in go(). INVESTIGATE. Severity: Low. Blocks: no.
6. All harness integrity checks are assert statements (vanish under python -O). Fix: explicit raise. Severity: Low. Blocks: no.

Nothing found would fail a test today, break production, or lose data or secrets. Every finding is a gap in what the guards or the harness prove.

Recommendation: Fix finding 1 (pin accountbudget.WithPriority's argument to accountbudget.Auxiliary and forbid Budget/Acquire at a128 static_test.go:102 and a140 static_test.go:92) before merge because, as written, a shadow tool can claim Critical priority on the shared account budget through an untyped constant or factory.Budget().Acquire, delay engine stop-loss requests, and still pass every guard.
```
(1번과 Recommendation 줄은 원문 문장 그대로이고 목록 들여쓰기만 줄임. 머리말 Coverage 와 2~6번은 사실·위치·수리안을 남기고 문장을 줄여 옮김 — 전체 원문은 이 세션의 하위 에이전트 a512e7a94604cbcbd 보고에 있음.)

## Codex 적대 패스 원문 (gpt-6-astra, OUTSIDE_STATUS completed)

```
CODEX_MODEL: gpt-6-astra (exec; source: /home/daniel/.codex/config.toml model)
NOTICE: gstack outside reviews send the review prompt and code to Codex (provider: openai) using the ChatGPT account Codex is logged in with (see: codex login status). Reviews stay on; this notice shows once per machine. To turn them off: gstack-config set codex_reviews disabled
1. **Unknown mutation IDs silently succeed.** In [mutation.py:112](openspec/changes/a189-shadow-tool-guards-admit-the-account-budget/analysis/mutation/mutation.py:112), requested IDs are filtered without validation. Passing `g2-prefix-wide` instead of `g2-prefix-widen` leaves `selected` empty: no controls or mutations run, yet lines 165–166 report zero issues and exit successfully. Mixing valid and invalid IDs silently omits requested coverage. Reject unknown IDs before creating the temporary checkout. Severity: Low; Blocks: no

No additional in-scope defects identified in the Go changes. Review was static; no tests or mutation harness executed, and no files modified. The two explicitly excluded weaknesses were not counted.

Recommendation: reject unknown mutation IDs because a misspelled ID currently returns success without running any control or mutation test.
VERDICT: clean
FINDINGS: P3
OUTSIDE_USAGE: input=321248 cached_input=243840 cache_write_input=0 output=1754 reasoning_output=468
OUTSIDE_STATUS: completed provider=codex host=claude
rc=0
```
(codex 세션 로그의 무관한 skill 로드 오류 한 줄 `ERROR codex_core … skill-tester … missing YAML frontmatter` 은 뺐음.)
