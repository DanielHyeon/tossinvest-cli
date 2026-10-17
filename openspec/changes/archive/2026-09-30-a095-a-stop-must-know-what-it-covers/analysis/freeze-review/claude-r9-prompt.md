You are an independent ADVERSARIAL senior engineer doing a NARROW re-verification of revision 9 (9판) of an OpenSpec
change in a Go repository that runs a real-money automated trading engine. Read-only: do NOT edit, create, or delete any
file, do NOT run git or Go tests or any network or engine command. Reading files and grep/rg are fine.

Repository root: the current directory — an export of commit 5d2f1b6e with the change directory overlaid (identical to
that commit). No .git directory. Change: openspec/changes/a095-a-stop-must-know-what-it-covers/. Do NOT open anything
under its analysis/freeze-review/ directory.

SCOPE — deliberately narrow (manager's instruction). Revision 8 was re-verified and rejected on one P1 (N1): an ordering
condition "a095 is implemented after a092's all-lock-holders requirement lands, or in the same window; P1 if broken"
contradicted the change's own dependency section, which says a095 is independent of a092 (user decision (1): "묶지
않는다" — do not bind a095's scope to a092), and gave no action if the premise broke. The manager ruled: decision (1) is
about SCOPE and is untouched; implementation ORDER is a scheduling matter under the manager's authority; the condition is
rewritten as "구현 착수 순서는 Manager 스케줄링 사항이다. a092 「모든 보유자」 착지 전에 a095 를 구현하는 경우, 그 창의 Q7 첫째 면
증폭을 알고 수용한다는 Manager 승인 기록이 착수 전에 있어야 하며(깨질 때의 행동 = 그 기록 없이 착수 금지), tasks 7.3 공시에 증폭
수용을 명기한다." and the old task 7.0 moves to section 0 as 0.7. A second item, N3: the sentence "a refused adoption engine
is one that asked for protection" was false for one config shape.

Review ONLY these three things:
1. tasks.md 「선후 관계」 section: is it now consistent with the rewritten condition — does it still say a095 is
   independent of a092 (scope) while not contradicting the scheduling precondition? Is there any remaining sentence
   anywhere in proposal.md, design.md, tasks.md, issues.md or either spec delta that still states the condition as a SPEC
   precondition (e.g. "전제로 한다", "깨지면 P1") or contradicts it?
2. tasks.md task 0.7 (and 7.3): is the rewritten condition stated exactly as quoted above, is the "action when broken"
   explicit and checkable before implementation starts, and is it placed where an in-order implementer meets it first?
   Also check proposal.md's transfer-record row and design.md's 「남는 경합」 text say the same thing. Check review.md §3.23's
   recorded reconciliation cites a real authority (docs/WORKFLOW.md 「역할 분리」).
3. The N3 correction: in design.md, proposal.md (Q2(a)), tasks.md 2.6a and the function map
   analysis/function-logic/internal-config--mergeadoption/function-logic-map.md and
   analysis/function-logic/internal-config--adoption.validate/function-logic-map.md — are the sentences true of
   internal/config/engine.go Adoption.validate (lines ~157–168) and mergeAdoption (~268–284)? Verify branch coordinates
   against the ast.json in those two bundle directories. Does task 2.6a now assert only observable facts (fact cell / key)
   and no grade?

Output: a table for items 1–3 (RESOLVED / NOT RESOLVED, one line why, file:line), a new-findings table limited to this
scope (id | severity P0/P1/P2/P3 | finding | evidence | fix), and a final line `VERDICT: APPROVE` (no P0/P1) or
`VERDICT: REJECT` with one sentence why.
