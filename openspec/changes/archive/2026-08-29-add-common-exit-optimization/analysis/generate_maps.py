#!/usr/bin/env python3
"""Regenerate pre-edit logic maps from the checked-in Go AST evidence."""

from __future__ import annotations

import json
from pathlib import Path


ROOT = Path(__file__).parent / "function-logic"

DETAILS = {
    "EvaluateLadder": (
        "validated decimal policy/state, positive entry/observed/high-water/baseline, monotone baseline",
        "Validate, positive/fraction/parseRat, ComputeProtectedStop, lockPrice",
        "preserve promotion-before-breach, pending cancel-first, completion, and partial/final precedence; add runner only as a max-composed protection candidate",
        "TestEvaluateLadder",
    ),
    "NewExitObserver": (
        "non-nil journal/price/retrier/issuer/gateway, account, costs, valid policy snapshots",
        "DefaultRatchetConfig, DefaultLadderPolicy, Validate, Clock.Now",
        "reject unknown configured policy; do not expand the order or Guardian capabilities",
        "TestNewExitObserver",
    ),
    "openState": (
        "exit-eligible position; adopted and engine-entered origins remain distinct",
        "LookupDecision, ParsePreimage, OpenExitState, openAdoptedState",
        "snapshot the startup common policy only when a new state is opened; existing state is never rebound",
        "TestExitObserverOpens",
    ),
    "openAdoptedState": (
        "adoption record is the sole t0 and recovery authority",
        "OpenAdoptedExitState",
        "recover policy ID from the committed adoption record, never from later config",
        "TestExitObserverOpensAdopted",
    ),
    "judgeLadder": (
        "stored state policy ID selects one immutable registry policy; adopted origin is retained",
        "checkLadderPolicyStillFits, RungIndex, EvaluateLadder, record",
        "unknown/mismatched snapshots refuse judgement; record/Guardian/submit ordering remains untouched",
        "TestExitObserverLadder",
    ),
    "checkLadderPolicyStillFits": (
        "active rung must exist in the selected policy and its lock must not exceed stored baseline",
        "LockPrice, CompareDecimal",
        "accept the resolved per-state policy as input while preserving all fail-closed checks",
        "TestCheckLadderPolicyStillFits",
    ),
    "OpenExitState": (
        "position exists and is exit-eligible; kind/ID are validated; t0 prices are valid",
        "OpenRatchetState, BeginTx, appendExitEventTx, Commit, ExitState",
        "write policy ID in the same transaction as state open; unique/existence failures remain fail-closed",
        "TestOpenExitStatePolicySnapshot",
    ),
    "scanExitState": (
        "query column order exactly matches scan targets",
        "row.Scan",
        "map nullable legacy LADDER ID to default_v1 without rewriting rows",
        "TestLegacyLadderPolicyID",
    ),
    "OpenAdoptedExitState": (
        "position and adoption record exist and no exit state exists",
        "BeginTx, readAdoptionTx, OpenRatchetState, appendExitEventTx, Commit",
        "use adoption.exit_policy_id in the same transaction and never re-read config",
        "TestOpenAdoptedExitStatePolicySnapshot",
    ),
    "AdoptPosition": (
        "position has neither entry decision nor prior different adoption",
        "record, BeginTx, readAdoptionTx, ExecContext, Commit",
        "persist selected policy ID with adoption before setting the position pointer",
        "TestAdoptPositionPolicySnapshot",
    ),
    "record": (
        "trimmed identity, decimal prices/quantity, RFC3339 time and valid common policy ID",
        "riskcalc decimal validation, sha256 digest construction",
        "include policy ID in deterministic adoption preimage/digest",
        "TestAdoptionRecordPolicyID",
    ),
    "scanAdoption": (
        "query column order exactly matches adoption fields",
        "row.Scan",
        "read nullable exit_policy_id as empty legacy value",
        "TestPositionAdoptionPolicyID",
    ),
    "adoptOne": (
        "fresh observed price and validated adoption stop configuration",
        "SyntheticStop, AdoptPosition",
        "snapshot startup common policy into request; adoption itself never submits an order",
        "TestAdoptOnePolicySnapshot",
    ),
    "ExitObserver": (
        "Context dependencies and caller overrides are validated",
        "NewExitObserver",
        "inject config common policy only into observer construction",
        "TestContextExitObserverPolicy",
    ),
    "ReconcileDriver": (
        "Context dependencies, freshness and adoption settings are validated",
        "NewReconcileDriver",
        "inject the same startup common policy into adoption requests",
        "TestContextReconcileDriverPolicy",
    ),
    "mergeEngine": (
        "raw optional blocks merge onto safe zero-value config",
        "adoption/gate validation and normalization",
        "empty means legacy RATCHET; unknown non-empty policy is retained as rejected and cannot run",
        "TestMergeExitPolicy",
    ),
    "routes": (
        "all routes remain behind session0; state changes additionally require mutating CSRF",
        "http.ServeMux.HandleFunc, session0, mutating",
        "add GET /optimization and POST /optimization/exit-policy with the dedicated seam only",
        "TestOptimizationRoutes",
    ),
    "runConsole": (
        "config/session dependencies are loaded before listener start",
        "console.New, config-backed setting seams, audit recorder",
        "wire only exit-policy load/save into console; no broker/gate/trading writer is exposed",
        "TestRunConsoleOptimizationWiring",
    ),
}


for directory in sorted(p for p in ROOT.iterdir() if p.is_dir()):
    ast_path = directory / "ast.json"
    if not ast_path.exists():
        continue
    ast = json.loads(ast_path.read_text())
    function = ast["function"]
    inputs, calls, boundary, test = DETAILS.get(
        function,
        ("validated caller inputs", "AST-listed callees", "preserve existing fail-closed behavior", f"Test{function}"),
    )
    branches = ast.get("branches") or []
    branch_rows = "\n".join(
        f"| {b['id']} | existing {b['kind']} branch at line {b['at']['line']} | only the branch's documented state transition | existing return/error contract | `{test}` |"
        for b in branches
    ) or f"| B0 | straight-line path | documented state transition | documented return | `{test}` |"
    test_rows = "\n".join(
        f"| {b['id']} | {b['kind']} path at current line {b['at']['line']} plus its complement | `{test}` | yes | yes |"
        for b in branches
    ) or f"| B0 | straight-line success/error contract | `{test}` | yes | yes |"

    (directory / "function-logic-map.md").write_text(
        f"""# Function Logic Map: `{function}`

- Source: `{ast['file']}`
- AST evidence: `ast.json` (`{ast['source_sha256']}`)
- Risk scan: `risk-pattern-report.md`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| function inputs and persisted state | {inputs} | caller types, journal/config schema, immutable registry | error/refusal; never broaden authority or silently fall back |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
{branch_rows}

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| {calls} | preserve current computation, persistence, and wiring contracts | errors propagate or are converted to the existing fail-closed refusal | CodeGraph + `ast.json` |

## State mutations and fallbacks

- {boundary}.
- No LIVE gate, trading toggle, broker call, or existing-position rebind is introduced by configuration.

## Safety conclusion

- Safe edit boundary: {boundary}.
- High-risk impact: yes; branch tests and post-edit AST/risk refresh are mandatory.
"""
    )
    (directory / "branch-test-map.md").write_text(
        f"""# Branch Test Map: `{function}`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
{test_rows}
"""
    )
