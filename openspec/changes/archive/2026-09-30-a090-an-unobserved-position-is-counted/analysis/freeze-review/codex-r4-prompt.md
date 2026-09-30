You are an independent ADVERSARIAL senior engineer doing the FOURTH (NARROW CONFIRMATION) codex proposal-freeze review of an OpenSpec change in a Go repository
that runs a real-money automated trading engine. You are read-only: do NOT edit, create, or delete any file, do NOT run git
commands that write, do NOT run network calls, do NOT run the engine or any command that could place an order. Reading files
and grep/rg are fine. Do not run Go tests.

Repository root: the current directory. It is an export of commit 0b5ea307 (all code and documents are exactly that commit).
Every Go file this change cites is byte-identical to its base `d3bd1843` except `cmd/tossctl/engine.go`, which gained 2 lines at
:121-122 (a066 operator command registration, unrelated). The change's engine.go citations (:634-640, :639, :671) were taken from
this later layout and are already correct in this tree (at the base they are 2 lines lower).

Change under review — `openspec/changes/a090-an-unobserved-position-is-counted/` (FIFTH draft, 5판 2fb2f393, plus design §D14 at 9d84abac).

THIS IS A NARROW CONFIRMATION ROUND. Scope: only whether the codex round-3 findings (R3-1 … R3-8, analysis/freeze-review/codex-r3-output.md)
are resolved by the 5th draft and D14, and whether those changes introduce a new P0/P1. Do not re-litigate decisions recorded as Manager
decisions unless they create a safety defect.
Read: design.md (5판 header, D4, D5, D10, D12, D13, D14, D3's fixture paragraph), specs/exit-policy/spec.md, tasks.md (0.6d–0.6f, 0.9,
1.0, 2.3, 2.3a, 2.3f, 2.3g, 2.15, 2.17, 3.1, 3.4), review.md 「codex 3라운드」, 「5판」, 「a094 7라운드 R7-4 교차」, 「계좌 정보 사실 고정」;
analysis/code-context/codegraph-baseline.md; analysis/function-logic/internal-app-engine--exitobserver.observeonce/function-logic-map.md;
and the a092 text D14 cites (openspec/changes/a092-an-alert-does-not-hold-the-stop/specs/engine-safety/spec.md:44, design.md §D0.3h item 4).

Relevant code: internal/app/engine/exitloop.go, internal/obs/notifier.go (:262-280, :385-396), internal/obs/mode.go, internal/obs/log.go,
internal/journal/outbox.go, internal/journal/operating_mode.go, internal/clock/clock.go, internal/app/engine/a111_flat_exit_observation_test.go
(a111WallElapsedClock), cmd/tossctl/engine.go (engineRuntime).

Produce (concise, evidence with file:line against this tree):
1. For each round-3 finding R3-1 … R3-8: RESOLVED / PARTIAL / NOT RESOLVED, one line why.
2. New findings introduced by the 5th draft / D14 only (table: id | severity P0/P1/P2/P3 | finding | evidence | fix). In particular: is
   every alert/announcement/log value a090 hands to the shared entrance account-free by construction, and does the failure-injection
   canary (tasks 2.17) actually prove that? Is the hard implementation dependency on a092's entrance stated consistently everywhere?
   Is the exit-census coordinate (ancestor path + occurrence ordinal + kind) now unique for every frozen exit?
3. Final line: `VERDICT: PASS` (no P0/P1 left; P2 may be recorded) or `VERDICT: REJECT`, with one sentence why.
