#!/usr/bin/env bash
# a198 보강 2 — D9·D10 증거용 AST 를 작업 트리에서 뽑음(읽기 전용). 사용: bash d9-d10-ast.sh <repo-root>
# 2026-11-24 HEAD f8a5bb898 에서 실행한 것과 같은 목록(analysis/d9-d10-evidence.md 머리 표).
set -euo pipefail
root="$(cd "$1" && pwd)"
out="$root/openspec/changes/a198-a-clearing-cancel-is-issued-as-a-cancel-decision/analysis/cited-ast"
cd "$root"
unset GOFLAGS
ex() { # file func outname
  go run ./tools/logic-map/extract_go_ast.go -file "$1" -func "$2" > "$out/$3.ast.json"
  printf '%s\t%s\t%s\n' "$3" "$1" "$(python3 -c 'import json,sys;d=json.load(open(sys.argv[1]));print(d["source_sha256"][:12], d["start"]["line"], d["end"]["line"], len(d["branches"] or []))' "$out/$3.ast.json")"
}
ex internal/reconcile/acked_boot.go confirmAcked d2_Recovery_confirmAcked
ex internal/reconcile/recovery.go Recovery.Run d2_Recovery_Run
ex internal/app/engine/in_doubt_settle.go inDoubtSettle.sweep d2_inDoubtSettle_sweep
ex internal/app/engine/in_doubt_settle.go inDoubtSettle.resolve d2_inDoubtSettle_resolve
ex internal/execgw/amend_indoubt.go Resolver.resolveCancel d2_Resolver_resolveCancel
ex internal/journal/dispatch.go Attempt.DispatchVerified d2_Attempt_DispatchVerified
ex internal/journal/resolution.go Attempt.ResolveConfirmed d2_Attempt_ResolveConfirmed
ex internal/journal/resolution.go Attempt.ResolveFailed d2_Attempt_ResolveFailed
ex internal/journal/durability.go Attempt.transition d2_Attempt_transition
ex internal/journal/durability.go Attempt.MarkAcked d2_Attempt_MarkAcked
ex internal/journal/lifecycle.go ValidateTransition d2_ValidateTransition
ex internal/journal/recovery.go Journal.PendingAttempts d2_Journal_PendingAttempts
ex internal/execgw/a094_shared_judgements.go Gateway.unsettledFor d2_Gateway_unsettledFor
ex internal/execgw/gateway.go Gateway.checkSymbolFree d2_Gateway_checkSymbolFree
ex internal/journal/exit_proposal_release.go Journal.ConfirmedCancelOf d2_Journal_ConfirmedCancelOf
ex internal/journal/exit_proposal_release.go intentOrdersAwaitingCloseQ d2_intentOrdersAwaitingCloseQ
ex internal/journal/fills.go Journal.LiveOrdersForSymbol d2_Journal_LiveOrdersForSymbol
ex internal/journal/fills.go markFillRefused d2_markFillRefused
ex internal/journal/fills.go updateFillSnapshotRefusal d2_updateFillSnapshotRefusal
ex internal/journal/fills.go upsertFillSnapshot d2_upsertFillSnapshot
ex internal/journal/lineage.go Journal.ResolveCurrentOrderIDScoped d2_Journal_ResolveCurrentOrderIDScoped
ex internal/brokerstate/derive.go Derive d2_brokerstate_Derive
ex internal/brokerstate/derive.go State.IsTerminal d2_brokerstate_State_IsTerminal
ex internal/flatten/flatten.go Saga.settleStep d2_Saga_settleStep
ex internal/flatten/flatten.go Saga.resolveStep d2_Saga_resolveStep
ex internal/flatten/flatten.go classifyCancel d2_flatten_classifyCancel
ex internal/flatten/liquidate.go Saga.sell d2_Saga_sell
ex internal/flatten/liquidate.go Saga.planLiquidation d2_Saga_planLiquidation
ex internal/execgw/gateway.go Gateway.Cancel d2_Gateway_Cancel
ex internal/app/engine/exitloop.go ExitObserver.clearTheSymbol d2_ExitObserver_clearTheSymbol_head
ex internal/reconcile/a094_acked_boot_test.go TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed d2_TestA094AnAckedCancelOrNumberlessPlaceIsOnlyNamed
ex internal/reconcile/a094_acked_boot_test.go TestA094BootAndPostPlaceShareOneConfirmation d2_TestA094BootAndPostPlaceShareOneConfirmation
ex internal/app/engine/a193_in_doubt_settle_test.go TestA193OnlyInDoubtPlaceAttemptsAreTouched d2_TestA193OnlyInDoubtPlaceAttemptsAreTouched
ex internal/app/engine/a193_in_doubt_settle_test.go TestA193TheSettlerNeverReplays d2_TestA193TheSettlerNeverReplays
