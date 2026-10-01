#!/usr/bin/env python3
"""a112 로트 5.6.2 · 5.2.2 변이 하네스 — a092 `mutate_unit2.py` 의 사본 방식(저장소 추적 파일을 pid 붙은 사본으로 복사 · 무변이
대조군이 GREEN 이 아니면 멈춤 · 변이는 한 번에 하나, 사본에서만 · 판정 CAUGHT/SURVIVED/BUILD-FAIL).

사용: python3 a112_lot_mutate.py --set 5.6.2.1|5.2.2.1|5.2.2.1-fix|6.2-seal|5.2.2.2 <scratch-dir> [로트 파일 경로 …]
(N 으로 시작하는 변이는 동작이 같은 리팩터 — GREEN-AS-EXPECTED 여야 한다.)
"""
from __future__ import annotations

import os
import re
import shutil
import subprocess
import sys
from pathlib import Path

ROOT = Path(subprocess.run(["git", "rev-parse", "--show-toplevel"], capture_output=True, text=True, check=True).stdout.strip())

SUP = "internal/app/engine/strategy_entry_supervisor.go"
FC = "internal/execgw/failclosed.go"
RT = "internal/execgw/retry.go"
SET_5621 = [
    ("E01 blocker claims success without latching", SUP,
     "\tif s == nil || s.entry == nil || worker == nil {\n\t\treturn false\n\t}\n\ts.entry.Block(",
     "\tif s == nil || s.entry == nil || worker == nil {\n\t\treturn false\n\t}\n\tif true {\n\t\treturn true\n\t}\n\ts.entry.Block("),
    ("E02 central fault swallowed again", SUP,
     "if isCentralStrategyIntegrity(err) && !s.blockEntryOnCentralIntegrity(worker) {",
     "if false && isCentralStrategyIntegrity(err) && !s.blockEntryOnCentralIntegrity(worker) {"),
    ("E03 central fault on refresh-only stops the engine", SUP,
     "if isCentralStrategyIntegrity(err) && !s.blockEntryOnCentralIntegrity(worker) {",
     "if isCentralStrategyIntegrity(err) {"),
    ("E04 every refresh-only error blocks entry", SUP,
     "if isCentralStrategyIntegrity(err) && !s.blockEntryOnCentralIntegrity(worker) {",
     "if !s.blockEntryOnCentralIntegrity(worker) || isCentralStrategyIntegrity(err) && false {"),
    ("E05 nil gate swallows instead of escalating", SUP,
     "\tif s == nil || s.entry == nil || worker == nil {\n\t\treturn false\n\t}\n\ts.entry.Block(",
     "\tif s == nil || s.entry == nil || worker == nil {\n\t\treturn true\n\t}\n\ts.entry.Block("),
    ("E06 production supervisor not given the gate", SUP,
     "Workers: workers, EntryGate: c.Entry,", "Workers: workers,"),
    ("E07 production supervisor built without a gate", SUP,
     "\tif c.Entry == nil {\n\t\treturn nil, fmt.Errorf(\"%w: the strategy refresh supervisor needs the entry gate\"",
     "\tif false {\n\t\treturn nil, fmt.Errorf(\"%w: the strategy refresh supervisor needs the entry gate\""),
    ("E08 option dropped by the constructor", SUP,
     "\t\tentry: opts.EntryGate,\n", ""),
    ("E09 wrong reason code", SUP,
     "s.entry.Block(execgw.ReasonStrategyCentralIntegrity,", "s.entry.Block(execgw.ReasonStrategyDispatchFenced,"),
    ("E10 reason not registered in the vocabulary", FC,
     # 2026-09-30 리뷰 보이스 B #3: 옛 앵커 `…,\n\t}` 는 766a8456(a094)이 뒤에 줄을 덧붙인 뒤로 착지 트리에 없었다(NOT-APPLIED).
     "\t\tReasonStrategyCentralIntegrity,\n\n", "\n"),
    ("E11 reason missing from the latch order", RT,
     "\tReasonStrategyCentralIntegrity,\n\t// Appended, per the rule above. The operating mode", "\t// Appended, per the rule above. The operating mode"),
]
SET_5621_TESTS = [
    ["go", "test", "-count=1", "-run",
     "TestARefreshOnly|TestAnOrdinaryRefreshOnlyCycleErrorDoesNotBlockEntry|TestWithoutAnEntryGate|TestTheProductionStrategySupervisor|TestTheCompleteCensus|TestTheOnlyWorkerProduction|TestTheFourEscalations",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "-run", "TestReasonCodeEnumIsStable|TestTheStrategyCentralIntegrityLatch", "./internal/execgw"],
]
HO = "internal/strategyhandoff/handoff.go"
DH = "internal/app/engine/strategy_dispatch_handoff.go"
FL = "internal/app/engine/strategy_account_first_leg_authority.go"
SET_5221 = [
    ("F01 activation gate removed (every market per scope)", DH,
     "\tif !authority.familyActivation().Verified() {\n\t\treturn []strategyhandoff.Handoff{authority.dispatchHandoff()}",
     "\tif false {\n\t\treturn []strategyhandoff.Handoff{authority.dispatchHandoff()}"),
    ("F02 activated market stays market-wide", DH,
     "\tif !authority.familyActivation().Verified() {\n\t\treturn []strategyhandoff.Handoff{authority.dispatchHandoff()}",
     "\tif true {\n\t\treturn []strategyhandoff.Handoff{authority.dispatchHandoff()}"),
    ("F03 duplicate owner scope not refused", HO,
     "if _, duplicate := seen[scope]; duplicate {", "if _, duplicate := seen[scope]; duplicate && false {"),
    ("F04 owner scope ignores position generation", HO,
     "\t\tgeneration: result.Lineage.PositionGeneration,", "\t\tgeneration: 0,"),
    ("F05 owner scope ignores symbol", HO,
     "\t\tsymbol:     strings.ToUpper(strings.TrimSpace(result.Lineage.Symbol)),", "\t\tsymbol:     \"\","),
    ("F06 symbol spelling not normalised", HO,
     "\t\tsymbol:     strings.ToUpper(strings.TrimSpace(result.Lineage.Symbol)),", "\t\tsymbol:     result.Lineage.Symbol,"),
    ("F07 account spelling not normalised", HO,
     "\t\taccount:    strings.TrimSpace(result.Lineage.AccountRef),", "\t\taccount:    result.Lineage.AccountRef,"),
    ("F08 closed market split into scopes", HO,
     "\tif !ready || len(selected) == 0 {\n\t\treturn []Handoff{Admit(ready, selected)}",
     "\tif len(selected) == 0 {\n\t\treturn []Handoff{Admit(ready, selected)}"),
    ("F09 delivery loop stops after the first scope", DH,
     "\tfor _, handoff := range handoffs {\n\t\tif err := handoff.Deliver(body); err != nil {",
     "\tfor _, handoff := range handoffs[:1] {\n\t\tif err := handoff.Deliver(body); err != nil {"),
    ("F10 delivery loop continues past a fault", DH,
     "\t\tif err := handoff.Deliver(body); err != nil {\n\t\t\treturn err\n\t\t}",
     "\t\tif err := handoff.Deliver(body); err != nil {\n\t\t\t_ = err\n\t\t}"),
    ("F11 production cycle reverts to the market-wide handoff", SUP,
     "return deliverEachStrategyHandoff(fresh.proposals.forMarket(market).dispatchHandoffs(),",
     "return deliverEachStrategyHandoff([]strategyhandoff.Handoff{fresh.proposals.forMarket(market).dispatchHandoff()},"),
    ("F12 activated branch mints through the market-wide door", DH,
     "\treturn strategyhandoff.AdmitEachOwnerScope(authority.snapshot.Ready, selected)",
     "\treturn []strategyhandoff.Handoff{strategyhandoff.Admit(authority.snapshot.Ready, selected)}"),
    ("F13 activated branch ignores readiness", DH,
     "\treturn strategyhandoff.AdmitEachOwnerScope(authority.snapshot.Ready, selected)",
     "\treturn strategyhandoff.AdmitEachOwnerScope(true, selected)"),
    ("F14 first-leg count guard removed (today-equivalence pin must see it)", FL,
     "\tif len(proposal.entries) != 1 || !riskAuthority.snapshot.Ready || !fx.snapshot.Ready",
     "\tif !riskAuthority.snapshot.Ready || !fx.snapshot.Ready"),
    ("F15 a second site mints per-scope handoffs", DH,
     "\treturn strategyhandoff.Admit(authority.snapshot.Ready, selected)\n}",
     "\t_ = strategyhandoff.AdmitEachOwnerScope\n\treturn strategyhandoff.Admit(authority.snapshot.Ready, selected)\n}"),
]
SET_5221_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run",
     "TestOnlyAnActivatedMarket|TestTwoOwnerScopesStill|TestEveryAdmittedOwnerScope|TestAScopeFault|TestTheProductionCycleDelivers"
     "|TestExactlyOneProduction|TestAdmitCensus|Handoff|Seam",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "./internal/strategyhandoff"],
]
MD = "internal/app/engine/strategy_market_handoff_delivery.go"
RO = "internal/strategyrouter/router.go"
# 2026-09-30 리뷰 수리 로트(codex · 보이스 A/B/C). G = 잡혀야 하는 변이, N = 동작이 같은 리팩터(초록이어야 함 — 거짓 양성 대조).
SET_5221_FIX = [
    ("G01 owner scope keyed by horizon (B X03)", HO,
     "\t\tsymbol:     strings.ToUpper(strings.TrimSpace(result.Lineage.Symbol)),",
     "\t\tsymbol:     strings.ToUpper(strings.TrimSpace(result.Lineage.Symbol)) + \"|\" + string(result.Lineage.Horizon),"),
    ("G02 owner scope keyed by lane (B X04)", HO,
     "\t\tsymbol:     strings.ToUpper(strings.TrimSpace(result.Lineage.Symbol)),",
     "\t\tsymbol:     strings.ToUpper(strings.TrimSpace(result.Lineage.Symbol)) + \"|\" + result.Lineage.LaneID,"),
    ("G03 owner scope ignores account (codex P2-4)", HO,
     "\t\taccount:    strings.TrimSpace(result.Lineage.AccountRef),", "\t\taccount:    \"\","),
    ("G04 owner scope ignores market (codex P2-4 · B X02)", HO,
     "\t\tmarket:     string(result.Lineage.Market),", "\t\tmarket:     \"\","),
    ("G05 router OwnerKey normalises account differently (C #4a)", RO,
     "\t\tAccountRef:         strings.TrimSpace(account),", "\t\tAccountRef:         strings.ToUpper(strings.TrimSpace(account)),"),
    ("G06 production body drops the second scope onward (B X09)", MD,
     "\treturn deliverEachStrategyHandoff(handoffs, func(delivered strategyhandoff.Delivered) error {\n\t\tlineage := delivered.Result().Lineage",
     "\tcalls := 0\n\treturn deliverEachStrategyHandoff(handoffs, func(delivered strategyhandoff.Delivered) error {\n\t\tcalls++\n\t\tif calls > 1 {\n\t\t\treturn nil\n\t\t}\n\t\tlineage := delivered.Result().Lineage"),
    ("G07 cycle delivers through a same-named method of another type (B X08)", SUP,
     ("\t\"github.com/JungHoonGhae/tossinvest-cli/internal/execgw\"\n)", "fresh.proposals.forMarket(market).dispatchHandoffs())"),
     ("\t\"github.com/JungHoonGhae/tossinvest-cli/internal/execgw\"\n\t\"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff\"\n)", "a112MarketWideOnly{fresh.proposals.forMarket(market)}.dispatchHandoffs())\n}\n\ntype a112MarketWideOnly struct{ a strategyProposalMarketAuthority }\n\nfunc (m a112MarketWideOnly) dispatchHandoffs() []strategyhandoff.Handoff {\n\treturn []strategyhandoff.Handoff{m.a.dispatchHandoff()}")),
    ("G08 local shadow of the delivery function (B X10)", SUP,
     ("\t\"github.com/JungHoonGhae/tossinvest-cli/internal/execgw\"\n)", "\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, fresh.proposals.forMarket(market).dispatchHandoffs())"),
     ("\t\"github.com/JungHoonGhae/tossinvest-cli/internal/execgw\"\n\t\"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff\"\n)", "\tdispatchStrategyMarketHandoffs := func(ctx context.Context, _ strategyCampaignCASReader, d strategyHandoffDispatcher, hs []strategyhandoff.Handoff) error {\n\t\treturn nil\n\t}\n\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, fresh.proposals.forMarket(market).dispatchHandoffs())")),
    ("N01 same-behaviour refactor: handoffs through a local variable (B X11)", SUP,
     "\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, fresh.proposals.forMarket(market).dispatchHandoffs())",
     "\thandoffs := fresh.proposals.forMarket(market).dispatchHandoffs()\n\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, handoffs)"),
    ("G09 cycle passes the market-wide handoff (old F11)", SUP,
     ("\t\"github.com/JungHoonGhae/tossinvest-cli/internal/execgw\"\n)", "fresh.proposals.forMarket(market).dispatchHandoffs())"),
     ("\t\"github.com/JungHoonGhae/tossinvest-cli/internal/execgw\"\n\t\"github.com/JungHoonGhae/tossinvest-cli/internal/strategyhandoff\"\n)", "[]strategyhandoff.Handoff{fresh.proposals.forMarket(market).dispatchHandoff()})")),
    ("G10 B2 always refuses — only the control half should fail (codex P2-3 · F16)", FL,
     "\tif len(proposal.entries) != 1 || !riskAuthority.snapshot.Ready || !fx.snapshot.Ready",
     "\tif len(proposal.entries) >= 1 || !riskAuthority.snapshot.Ready || !fx.snapshot.Ready"),
    ("G11 second minting door via a public var's private type (codex P1-2)", HO,
     "var ErrNoDelivery = errors.New(\"strategyhandoff: delivery body is nil\")",
     "var ErrNoDelivery = deliveryError{errors.New(\"strategyhandoff: delivery body is nil\")}\n\ntype deliveryError struct{ error }\n\nfunc (deliveryError) Mint(r strategyflow.Result) Handoff {\n\treturn Handoff{selected: []strategyflow.Result{r}, pending: 1}\n}"),
    ("G12 engine admits each element separately (C #2 experiment C2)", DH,
     "\treturn strategyhandoff.AdmitEachOwnerScope(authority.snapshot.Ready, selected)",
     "\tadmit := strategyhandoff.AdmitEachOwnerScope\n\tif !authority.snapshot.Ready || len(selected) == 0 {\n\t\treturn admit(authority.snapshot.Ready, selected)\n\t}\n\tvar out []strategyhandoff.Handoff\n\tfor _, one := range selected {\n\t\tout = append(out, admit(true, []strategyflow.Result{one})...)\n\t}\n\treturn out"),
    ("G13 a fourth Single() site with the new spelling (C #3 experiment D)", DH,
     "func deliverEachStrategyHandoff(",
     "func a112Peek(authority strategyProposalMarketAuthority) bool {\n\t_, ok := authority.dispatchHandoffs()[0].Single()\n\treturn ok\n}\n\nfunc deliverEachStrategyHandoff("),
    ("G14 Single() answer discarded via an intermediate binding (C #3 experiment D')", DH,
     "func deliverEachStrategyHandoff(",
     "func a112Peek(authority strategyProposalMarketAuthority) strategyflow.Result {\n\th := authority.dispatchHandoff()\n\tr, _ := h.Single()\n\treturn r\n}\n\nfunc deliverEachStrategyHandoff("),
    ("G15 new exported door the engine uses (C #1 experiment B shape)", HO,
     "// ownerScope 는 포지션 소유 범위의 비교용 표기",
     "// AdmitVar 는 새 문.\nvar AdmitVar = AdmitEachOwnerScope\n\n// ownerScope 는 포지션 소유 범위의 비교용 표기"),
    ("G16 delivery keeps going after a fault (old F10, now on the production body)", DH,
     "\t\tif err := handoff.Deliver(body); err != nil {\n\t\t\treturn err\n\t\t}",
     "\t\tif err := handoff.Deliver(body); err != nil {\n\t\t\t_ = err\n\t\t}"),
    ("G18 B2 count condition removed — both orders must fail (old F14; fixture order by orders, coordinator order by the message)", FL,
     "\tif len(proposal.entries) != 1 || !riskAuthority.snapshot.Ready || !fx.snapshot.Ready",
     "\tif !riskAuthority.snapshot.Ready || !fx.snapshot.Ready"),
    ("G17 claimed scope stops the cycle instead of being skipped", MD,
     "\t\tif cas.Claimed || cas.State != \"FLAT\" && cas.State != \"CLOSED\" {\n\t\t\treturn nil\n\t\t}",
     "\t\tif cas.Claimed || cas.State != \"FLAT\" && cas.State != \"CLOSED\" {\n\t\t\treturn errors.New(\"claimed\")\n\t\t}"),
]
SET_5221_FIX_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run",
     "Handoff|Seam|Admit|OwnerScope|Delivery|Delivers|Dispatch|Classified|TwoOwner|SameOwner|Skipped|FaultInOne|Discards|OnlyAnActivated|EveryAdmitted|AScopeFault",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "-run",
     "Handoff|Seam|Admit|Delivery|Delivers|Dispatch|Classified|Skipped|FaultInOne|Discards",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "./internal/strategyhandoff"],
]
# 3차(codex 재확인 뒤, 2026-10-01): 주기 행동 시험 · 타입 동일성 census 를 죽이는 변이. H01 이 Manager 가 요구한 「n-1 전달」.
SET_5221_FIX3 = [
    ("H01 cycle hands n-1 handoffs (codex recheck #1 — clear the tail at the call site)", SUP,
     "\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, fresh.proposals.forMarket(market).dispatchHandoffs())",
     "\ths := fresh.proposals.forMarket(market).dispatchHandoffs()\n\tif len(hs) > 1 {\n\t\ths = hs[:len(hs)-1]\n\t}\n\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, hs)"),
    ("H02 cycle zeroes the tail handoffs in place (clear)", SUP,
     "\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, fresh.proposals.forMarket(market).dispatchHandoffs())",
     "\ths := fresh.proposals.forMarket(market).dispatchHandoffs()\n\tif len(hs) > 1 {\n\t\tclear(hs[1:])\n\t}\n\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, hs)"),
    ("H03 delivery body skips every scope after the first via a pointer counter (codex recheck #2 shape)", MD,
     "\treturn deliverEachStrategyHandoff(handoffs, func(delivered strategyhandoff.Delivered) error {\n\t\tlineage := delivered.Result().Lineage",
     "\tseen := new(int)\n\treturn deliverEachStrategyHandoff(handoffs, func(delivered strategyhandoff.Delivered) error {\n\t\t(*seen)++\n\t\tif *seen > 1 {\n\t\t\treturn nil\n\t\t}\n\t\tlineage := delivered.Result().Lineage"),
    ("N02 same-behaviour refactor: var-declared handoffs (codex recheck #4)", SUP,
     "\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, fresh.proposals.forMarket(market).dispatchHandoffs())",
     "\tvar hs = fresh.proposals.forMarket(market).dispatchHandoffs()\n\treturn dispatchStrategyMarketHandoffs(ctx, c.Journal, fresh.dispatch, hs)"),
    ("H04 private alias + init reassignment mints a Delivered (codex recheck #3)", HO,
     "// ownerScope 는 포지션 소유 범위의 비교용 표기",
     "type hiddenDelivered = Delivered\n\ntype mintingError struct{ error }\n\nfunc (mintingError) Mint(r strategyflow.Result) hiddenDelivered {\n\treturn hiddenDelivered{result: r}\n}\n\nfunc init() {\n\tErrNoDelivery = mintingError{ErrNoDelivery}\n}\n\n// ownerScope 는 포지션 소유 범위의 비교용 표기"),
    ("H05 private-alias method alone (no reassignment)", HO,
     "// ownerScope 는 포지션 소유 범위의 비교용 표기",
     "type hiddenHandoff = Handoff\n\ntype quietMinter struct{}\n\nfunc (quietMinter) mint(r strategyflow.Result) hiddenHandoff {\n\treturn Admit(true, []strategyflow.Result{r})\n}\n\n// ownerScope 는 포지션 소유 범위의 비교용 표기"),
    ("H06 out-param minter with no results (codex recheck: skipped at :52-53)", HO,
     "// ownerScope 는 포지션 소유 범위의 비교용 표기",
     "func admitInto(dst *[]Handoff, ready bool, selected []strategyflow.Result) {\n\t*dst = AdmitEachOwnerScope(ready, selected)\n}\n\n// ownerScope 는 포지션 소유 범위의 비교용 표기"),
    ("H07 function-valued package var re-exposing a door", HO,
     "// ownerScope 는 포지션 소유 범위의 비교용 표기",
     "var eachScope = AdmitEachOwnerScope\n\n// ownerScope 는 포지션 소유 범위의 비교용 표기"),
    ("H08 a Handoff built with a selection outside Admit, through an alias", HO,
     "// ownerScope 는 포지션 소유 범위의 비교용 표기",
     "type aliasHandoff = Handoff\n\nfunc (h Handoff) widen(r strategyflow.Result) Handoff {\n\treturn aliasHandoff{selected: []strategyflow.Result{r}, pending: 1}\n}\n\n// ownerScope 는 포지션 소유 범위의 비교용 표기"),
]
# 4차(Manager 판정 안 (b) — 소스 digest 동결): 동결을 뚫으려는 편집. digest 를 갱신하지 **않는** 편집은 전부 잡혀야 한다.
SET_5221_FIX4 = [
    ("H09 codex round-3 generic minter + any-returning method + pointer reassignment (digest not re-pinned)", HO,
     "// ownerScope 는 포지션 소유 범위의 비교용 표기",
     "type mintingError struct{ error }\n\nfunc makeSeam[T ~struct{ result strategyflow.Result }](r strategyflow.Result) T {\n\treturn T{result: r}\n}\n\nfunc (mintingError) Mint(r strategyflow.Result) any {\n\treturn makeSeam[Delivered](r)\n}\n\nfunc replaceError(dst *error) {\n\t*dst = mintingError{*dst}\n}\n\nfunc init() {\n\treplaceError(&ErrNoDelivery)\n}\n\n// ownerScope 는 포지션 소유 범위의 비교용 표기"),
    ("H10 twin-struct conversion minter (digest not re-pinned)", HO,
     "// ownerScope 는 포지션 소유 범위의 비교용 표기",
     "type twin struct{ result strategyflow.Result }\n\nfunc (h Handoff) sneak(r strategyflow.Result) any {\n\treturn Delivered(twin{r})\n}\n\n// ownerScope 는 포지션 소유 범위의 비교용 표기"),
    ("H11 comment-only edit to the seam (digest not re-pinned — every edit is a re-pin)", HO,
     "// ownerScope 는 포지션 소유 범위의 비교용 표기",
     "// (주석 한 줄 추가)\n// ownerScope 는 포지션 소유 범위의 비교용 표기"),
    ("H12 a new production file in the seam package (digest not re-pinned)", "internal/strategyhandoff/zz_extra.go",
     None, "package strategyhandoff\n\n// extraFileMarker 는 새 생산 파일이다.\nconst extraFileMarker = 1\n"),
    ("N03 gofmt-normalised whitespace only (digest must not move)", HO,
     "\tpending := len(selected)\n",
     "\tpending  :=  len(selected)\n"),
]
# 6.2 봉인 로트(2026-10-01): 범위 재유도 · A-lite · strategyflow 봉인 census. S = 엔진 봉인, F = strategyflow census, N = 동작 동등.
SEAL = "internal/app/engine/strategy_first_leg_owner_scope.go"
FLA = "internal/app/engine/strategy_account_first_leg_authority.go"
SFT = "internal/strategyflow/types.go"
SFL = "internal/strategyflow/flow.go"
REVIEW = "openspec/changes/a112-run-four-strategy-families-independently/review.md"
SET_62_SEAL = [
    ("S01 re-derivation removed (the guard compares accepted with itself)", FLA,
     "\tresult := proposalAuthority.Proposal()\n", "\tresult := accepted.result\n\t_ = proposalAuthority\n"),
    ("S02 selection by identity instead of owner scope (the self-reference trap)", SEAL,
     "\t\tif key == want {", "\t\tif _ = want; key == key && held.Identity == lineage.Identity {"),
    ("S03 comparison weakened: execution-terms disjunct dropped", FLA,
     "if result.Lineage.Identity != accepted.result.Lineage.Identity || result.ExecutionTerms.Identity() != accepted.result.ExecutionTerms.Identity() {",
     "if result.Lineage.Identity != accepted.result.Lineage.Identity {"),
    ("S04 scope key weakened: symbol ignored on both sides", SEAL,
     ("\twant, err := strategyrouter.NewOwnerKey(lineage.AccountRef, lineage.Market, lineage.Symbol, lineage.PositionGeneration)",
      "\t\tkey, err := strategyrouter.NewOwnerKey(held.AccountRef, held.Market, held.Symbol, held.PositionGeneration)"),
     ("\twant, err := strategyrouter.NewOwnerKey(lineage.AccountRef, lineage.Market, \"ANY\", lineage.PositionGeneration)",
      "\t\tkey, err := strategyrouter.NewOwnerKey(held.AccountRef, held.Market, \"ANY\", held.PositionGeneration)")),
    ("S05 uniqueness dropped (first match wins)", SEAL,
     "\tif matches != 1 {", "\tif matches < 1 {"),
    ("S06 selection by position (entries[0]) — half of the pre-seal shape (with M7)", SEAL,
     "\tif matches != 1 {\n\t\treturn strategyproposal.ProductionAuthority{}, false\n\t}\n\treturn chosen, true",
     "\tif len(authority.entries) == 0 {\n\t\treturn strategyproposal.ProductionAuthority{}, false\n\t}\n\t_ = chosen\n\treturn authority.entries[0].authority, true"),
    ("S07 selection failure ignored", FLA,
     "\tif !scoped {\n\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, errors.New(strategyFirstLegOwnerScopeRefusal)\n\t}",
     "\tif !scoped && false {\n\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, errors.New(strategyFirstLegOwnerScopeRefusal)\n\t}"),
    ("S08 A-lite contract removed", DH,
     "\tready := authority.snapshot.Ready && strategyProposalSetDigest(authority.entries) == authority.snapshot.ProposalSetDigest\n",
     "\tready := authority.snapshot.Ready\n"),
    ("S09 A-lite applied to the unactivated market (toggle OFF must stay upstream)", DH,
     "\tif !authority.familyActivation().Verified() {\n\t\treturn []strategyhandoff.Handoff{authority.dispatchHandoff()}",
     "\tif !authority.familyActivation().Verified() && strategyProposalSetDigest(authority.entries) == authority.snapshot.ProposalSetDigest {\n\t\treturn []strategyhandoff.Handoff{authority.dispatchHandoff()}"),
    ("F01 a second seal writer (FinalizeProposalQuantity re-seals)", SFT,
     "\tfinal.proposalSeal = [32]byte{}\n", "\tfinal.proposalSeal = proposalResultSeal(final)\n"),
    ("F02 twin struct carrying the seal field (conversion channel)", SFT,
     "func sealProposalResult(result Result) Result {",
     "type sealTwin struct{ proposalSeal [32]byte }\n\nvar _ = sealTwin{}\n\nfunc sealProposalResult(result Result) Result {"),
    ("F03 new exported minter in the default build", SFL,
     "func Propose(request Request) Result {",
     "func SealForAnyone(result Result) Result { return sealProposalResult(result) }\n\nfunc Propose(request Request) Result {"),
    ("F04 ValidProposal weakened (quantity clause dropped)", SFT,
     "result.Code == RefusalNone && result.Quantity > 0 &&", "result.Code == RefusalNone &&"),
    ("F05 the review record drops the frozen pins (re-pin/review binding)", REVIEW,
     "  - `sealProposalResult` = `sha256:5e45db69120f540a463f43b2c6d8ba8cd0e7f94d3b070fcb9a80b42629d6bd6a`\n",
     "  - `sealProposalResult` = (removed)\n"),
    ("S10 loader keeps the shared slice (codex #1 — no detach at construction)", FLA,
     "proposals: detachedStrategyProposalPair(proposals), risk: riskAuthority", "proposals: proposals, risk: riskAuthority"),
    ("S11 detach copies only one market", SEAL,
     "\tpair.kr, pair.us = detach(pair.kr), detach(pair.us)", "\tpair.kr = detach(pair.kr)"),
    ("F06 function-value alias of the sealer (codex #2)", SFL,
     "\t\tresult = sealProposalResult(result)", "\t\tmint := sealProposalResult\n\t\tresult = mint(result)"),
    ("F07 seal written through the field's address (codex #2)", SFT,
     "\tfinal.proposalSeal = [32]byte{}\n", "\tseal := &final.proposalSeal\n\t*seal = proposalResultSeal(final)\n"),
    ("M7 count gate before the scope selection (the pre-seal order, with S06)", FLA,
     ("\tproposalAuthority, scoped := proposal.authorityForOwnerScope(accepted.result.Lineage)\n\tif !scoped {",
      "\t// 시장 단위 개수 관문 — 봉인이 아니라 시장당 하나 상한임. 걷어 내는 일(두 소유자 범위 시장의 거래)은 5.2.2.2.\n\tif len(proposal.entries) != 1 {\n\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, errors.New(\"paired production authority is incomplete for market\")\n\t}\n\tresult := proposalAuthority.Proposal()"),
     ("\tif len(proposal.entries) != 1 {\n\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, errors.New(\"paired production authority is incomplete for market\")\n\t}\n\tproposalAuthority, scoped := proposal.authorityForOwnerScope(accepted.result.Lineage)\n\tif !scoped {",
      "\tresult := proposalAuthority.Proposal()")),
    ("M12 selection scans only the first entry", SEAL,
     "\tfor _, entry := range authority.entries {", "\tfor _, entry := range authority.entries[:min(1, len(authority.entries))] {"),
    ("K1 selection key also requires the accepted campaign (same-scope loser no longer selected)", SEAL,
     "\t\tif key == want {", "\t\tif key == want && held.CampaignID == lineage.CampaignID {"),
    ("K2 selection key also requires the accepted lane (gated lane no longer selected)", SEAL,
     "\t\tif key == want {", "\t\tif key == want && held.LaneID == lineage.LaneID {"),
    ("M4 scope key drops the position generation (both sides) — expected to SURVIVE (seam cannot mint generation != 1; named)", SEAL,
     ("lineage.Symbol, lineage.PositionGeneration)", "held.Symbol, held.PositionGeneration)"),
     ("lineage.Symbol, 1)", "held.Symbol, 1)")),
    ("M5 scope key drops the account (both sides)", SEAL,
     ("strategyrouter.NewOwnerKey(lineage.AccountRef,", "strategyrouter.NewOwnerKey(held.AccountRef,"),
     ("strategyrouter.NewOwnerKey(\"acct\",", "strategyrouter.NewOwnerKey(\"acct\",")),
    ("M6 scope key drops the market (both sides)", SEAL,
     ("lineage.AccountRef, lineage.Market,", "held.AccountRef, held.Market,"),
     ("lineage.AccountRef, strategyrouter.MarketKR,", "held.AccountRef, strategyrouter.MarketKR,")),
    ("F08 a !cgo file re-seals in the production image (B#1 — the host's build context skipped it)", "internal/strategyflow/reseal_nocgo.go",
     None, "//go:build !cgo\n\npackage strategyflow\n\n// ResealForAnyone re-seals a result.\nfunc ResealForAnyone(r Result) Result {\n\tseal := proposalResultSeal(r)\n\tcopy(r.proposalSeal[:], seal[:])\n\treturn r\n}\n"),
    ("N04 comment-only edit inside a frozen seal function (digest is comment-free)", SFT,
     "func sealProposalResult(result Result) Result {\n", "func sealProposalResult(result Result) Result {\n\t// (주석 한 줄)\n"),
]
SET_62_SEAL_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run",
     "TestTheFirstLegSeal|TestFirstLegAuthority|TestTheFirstLegBackstop|TestAnActivatedMarketWhose|TestTheProposalSetDigest|TestOnlyAnActivated|TestTwoOwnerScopes|TestTheProductionCycle|TestWithoutAnActivation|TestTheScopeSelector|TestAnInPlaceSwap|TestTheSingleProposalAssumption",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "-run", "TestTheSingleProposalAssumption|TestTheFirstLegBackstop|Handoff|Seam|Admit|Classified", "./internal/app/engine"],
    ["go", "test", "-count=1", "./internal/strategyflow"],
]
SET_5221_FIX3_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run",
     "TheProductionCycleHandsEvery|WithoutAnActivationTheProductionCycle|TheProductionCycleEnds|TheDeliveryBodyHands|TheMarketDelivery|Skipped|FaultInOne|TwoOwner|SameOwner|Classified|ExactlyOneProduction",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "./internal/strategyhandoff"],
]
ADM = "internal/app/engine/strategy_first_leg_admission.go"
DCY = "internal/app/engine/strategy_dispatch_cycle.go"
RSK = "internal/app/engine/strategy_risk_authority.go"
PRA = "internal/app/engine/strategy_proposal_authority.go"
PRJ = "internal/app/engine/strategy_runtime_projection.go"
SET_5222 = [
    ("X01 every delivery fault skipped as a scope refusal (J4 misclassification)", MD,
     "\t\tif errors.As(err, &scope) {", "\t\tif true || errors.As(err, &scope) {"),
    ("X02 a scope refusal stops the cycle (starvation — J3)", MD,
     "\t\tif errors.As(err, &scope) {", "\t\tif false && errors.As(err, &scope) {"),
    ("X03 skipped refusals dropped when a later fault stops (J4 ③)", MD,
     "\t\treturn errors.Join(append(skipped, err)...)", "\t\treturn err"),
    ("X04 skipped refusals dropped at the end (silent skip — J3 ②)", MD,
     "\treturn errors.Join(skipped...)", "\treturn nil"),
    ("X05 admit types every collection failure as a scope refusal", ADM,
     "\t\tif scope := (*strategyScopeRefusal)(nil); errors.As(err, &scope) {\n\t\t\trefusal.scope = scope",
     "\t\tif scope := (*strategyScopeRefusal)(nil); true {\n\t\t\tif !errors.As(err, &scope) {\n\t\t\t\tscope = &strategyScopeRefusal{detail: err.Error()}\n\t\t\t}\n\t\t\trefusal.scope = scope"),
    ("X06 dispatch wraps every admission refusal with the scope type", DCY,
     "\t\tif admitted.scope != nil {", "\t\tif true {"),
    ("X07 identity mismatch returned as a scope refusal", FLA,
     "\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, errors.New(\"production proposal identity changed\")",
     "\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, &strategyScopeRefusal{detail: \"production proposal identity changed\"}"),
    ("X08 risk authority falls back to the market bundle (J3 ① envelope fallback)", FLA,
     "\triskBundle, riskScoped := riskAuthority.forScope(key)", "\triskBundle, riskScoped := riskAuthority.bundle, true"),
    ("X09 account authority falls back to the market authority (J3 ①)", FLA,
     "\taccountAuthority, accountScoped := account.forScope(key)", "\taccountAuthority, accountScoped := account.authority, true"),
    ("X10 currency from the envelope again (A#3)", FLA,
     "\tcurrency, currencyKnown := map[strategyrouter.Market]string{strategyrouter.MarketKR: \"KRW\", strategyrouter.MarketUS: \"USD\"}[result.Lineage.Market]",
     "\tcurrency, currencyKnown := accepted.currency, true"),
    ("X11 market count gate restored", FLA,
     "\tkey, _ := strategyOwnerKeyOf(result.Lineage)\n",
     "\tif len(proposal.entries) != 1 {\n\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, errors.New(\"paired production authority is incomplete for market\")\n\t}\n\tkey, _ := strategyOwnerKeyOf(result.Lineage)\n"),
    ("X12 risk loader loads only the first scope", RSK,
     "\tfor _, scoped := range result.results() {", "\tfor _, scoped := range result.results()[:1] {"),
    ("X13 result authority hands the risk loader only the first scope", PRA,
     "\t\tif len(results) > 1 || value.familyActivation().Verified() {", "\t\tif false {"),
    ("X14 account loader loads the first entry's symbol for every scope (A#5)", FL,
     "AccountCurrency: loader.accountCurrency, Symbol: result.Lineage.Symbol, Market: accountMarket,",
     "AccountCurrency: loader.accountCurrency, Symbol: proposal.entries[0].authority.Proposal().Lineage.Symbol, Market: accountMarket,"),
    ("X15 worker promotion needs every scope (one refusal starves the market)", SUP,
     "\t\tif _, err := gateway.ObserveStrategyEntryGate(ctx, strings.ToLower(string(market)), result.Lineage.Symbol); err != nil {\n\t\t\tcontinue",
     "\t\tif _, err := gateway.ObserveStrategyEntryGate(ctx, strings.ToLower(string(market)), result.Lineage.Symbol); err != nil {\n\t\t\treturn dormant"),
    ("X16 worker promotion reads the market-wide handoff", SUP,
     "\tfor _, handoff := range p.dispatchHandoffs() {", "\tfor _, handoff := range append(p.dispatchHandoffs()[:0], p.dispatchHandoff()) {"),
    ("X17 worker promoted although the entry gate refused every scope", SUP,
     "\t\tif _, err := gateway.ObserveStrategyEntryGate(ctx, strings.ToLower(string(market)), result.Lineage.Symbol); err != nil {",
     "\t\tif _, err := gateway.ObserveStrategyEntryGate(ctx, strings.ToLower(string(market)), result.Lineage.Symbol); err != nil && false {"),
    ("X18 assembly digest drifts from the contract function (A#6)", PRA,
     "ProposalSetDigest: strategyProposalSetDigest(entries),", "ProposalSetDigest: strategyProposalSetDigest(entries[:len(entries)-1]),"),
    ("X19 projection reads the market-wide handoff", PRJ,
     "\t\tfor _, handoff := range assembly.proposals.forMarket(market).dispatchHandoffs() {",
     "\t\tfor _, handoff := range append(assembly.proposals.forMarket(market).dispatchHandoffs()[:0], assembly.proposals.forMarket(market).dispatchHandoff()) {"),
    ("M20 risk generation read from the market bundle — expected to SURVIVE (every scope's bundle comes from the same signed market manifest; generation is per market by construction)", DCY,
     "\t\tif bundle, scoped := cycle.risk.forMarket(market).forScope(key); scoped {",
     "\t\tif bundle, scoped := cycle.risk.forMarket(market).bundle, key.Symbol != \"\"; scoped {"),
    ("X21 the delivery body stops on the first scope refusal of the account (J3, account half)", FLA,
     "\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, &strategyScopeRefusal{scope: key, detail: \"no ready account authority for this owner scope\"}",
     "\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, errors.New(\"no ready account authority for this owner scope\")"),
    ("X22 the risk scope refusal untyped (J3, risk half)", FLA,
     "\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, &strategyScopeRefusal{scope: key, detail: \"no ready risk authority for this owner scope\"}",
     "\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, errors.New(\"no ready risk authority for this owner scope\")"),
]
SET_5222_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run",
     "TestAnActivatedTwoScope|TestTheSecondLeg|TestAScopeWithout|TestTheFirstLegCurrency|TestTheRiskStubBridge|TestAForgedScope|TestTheFirstLegSeal|TestTwoOwnerScopes|TestAScopeFault|TestEveryAdmitted|TestARefusedHandoff|TestTheScopeSelector|TestTheProposalSetDigest|TestAnActivatedMarketWhose|TestOnlyAnActivated|TestTheSameOwnerScope|TestAnInPlaceSwap|TestFirstLegAuthority|TestTheFirstLegBackstop|TestTheProductionCycle|TestWithoutAnActivation",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "-run", "TestTheSingleProposalAssumption|TestTheScopeRefusalType|TestOnlyATypedScopeRefusal|Handoff|Seam|Admit|Classified|Single", "./internal/app/engine"],
]
OWN = "internal/app/engine/strategy_owner_scope_authority.go"
RB = "internal/riskbucket/production_snapshot_authority.go"
SET_5222_FIX = [
    ("Y01 sentinel removed at the out-of-policy symbol site", RB,
     'nil, fmt.Errorf("%w: symbol sector mapping unavailable", ErrProductionRiskScopeRefused)', 'nil, errors.New("symbol sector mapping unavailable")'),
    ("Y02 sentinel removed at the scope latch site", RB,
     'return nil, fmt.Errorf("%w: scope latch present", ErrProductionRiskScopeRefused)', 'return nil, errors.New("risk bucket: scope latch present")'),
    ("Y03 sentinel widened onto the latch read fault", RB,
     'return nil, fmt.Errorf("risk bucket: scope latch unreadable: %w", err)', 'return nil, fmt.Errorf("%w: scope latch unreadable: %w", ErrProductionRiskScopeRefused, err)'),
    ("Y04 sentinel widened onto every ledger-entry fault", RB,
     "\tentries, err := loadProductionRiskEntries(ctx, config, manifest.productionRiskPolicyBody, scope, reserve, limits)\n\tif err != nil {\n\t\treturn RiskSnapshotAuthorityBundle{}, fmt.Errorf(\"%w: %w\", ErrProductionRiskSnapshotUnavailable, err)",
     "\tentries, err := loadProductionRiskEntries(ctx, config, manifest.productionRiskPolicyBody, scope, reserve, limits)\n\tif err != nil {\n\t\treturn RiskSnapshotAuthorityBundle{}, fmt.Errorf(\"%w: %w: %w\", ErrProductionRiskSnapshotUnavailable, ErrProductionRiskScopeRefused, err)"),
    ("Y05 bind cause flattened again (%v)", RB,
     "\tscope, reserve, limits, err := bindProductionRiskInputs(config, manifest.productionRiskPolicyBody, input)\n\tif err != nil {\n\t\treturn RiskSnapshotAuthorityBundle{}, fmt.Errorf(\"%w: %w\", ErrProductionRiskSnapshotUnavailable, err)",
     "\tscope, reserve, limits, err := bindProductionRiskInputs(config, manifest.productionRiskPolicyBody, input)\n\tif err != nil {\n\t\treturn RiskSnapshotAuthorityBundle{}, fmt.Errorf(\"%w: %v\", ErrProductionRiskSnapshotUnavailable, err)"),
    ("Y06 engine risk loader stamps every failure scope-local", RSK,
     "\t\t\t\tentry.cause = err\n", "\t\t\t\tentry.cause = riskbucket.ErrProductionRiskScopeRefused\n"),
    ("Y07 risk classifier calls every cause scope-local", OWN,
     "return errors.Is(scope.cause, riskbucket.ErrProductionRiskScopeRefused), scope.cause", "return true, scope.cause"),
    ("Y08 a scope the risk authority does not hold becomes scope-local", OWN,
     'return false, errors.New("production risk authority holds no entry for this owner scope")', 'return true, errors.New("production risk authority holds no entry for this owner scope")'),
    ("Y09 account ctx fault classified scope-local", OWN, "return !fault, scope.cause", "return !fault || true, scope.cause"),
    ("Y10 every account failure classified as a fault", OWN, "return !fault, scope.cause", "return !fault && false, scope.cause"),
    ("Y11 account loader ignores its own context after a failure", FL,
     "\t\t\t\tif ctxErr := ctx.Err(); ctxErr != nil {", "\t\t\t\tif ctxErr := ctx.Err(); false && ctxErr != nil {"),
    ("Y12 unactivated count gate removed", FL,
     "\tif !proposal.familyActivation().Verified() && len(proposal.entries) != 1 {", "\tif false && !proposal.familyActivation().Verified() && len(proposal.entries) != 1 {"),
    ("Y13 count gate applied to activated markets too", FL,
     "\tif !proposal.familyActivation().Verified() && len(proposal.entries) != 1 {", "\tif len(proposal.entries) != 1 {"),
    ("M14 unkeyed first leg no longer refused — expected to SURVIVE (unreachable: the seal's selection already normalises the key)", FL,
     "\tif !keyed { // 봉인이", "\tif false && !keyed { // 봉인이"),
    ("Y15 worker promotion skips the scope's risk readiness", SUP,
     "\t\tif _, ready := r.forScope(key); !ready {\n\t\t\tcontinue\n\t\t}", "\t\tif _, ready := r.forScope(key); false && !ready {\n\t\t\tcontinue\n\t\t}"),
    ("Y16 worker promotion skips the scope's account readiness", SUP,
     "\t\tif _, ready := a.forScope(key); !ready {\n\t\t\tcontinue\n\t\t}", "\t\tif _, ready := a.forScope(key); false && !ready {\n\t\t\tcontinue\n\t\t}"),
    ("Y17 worker expiry from the first ready scope again", SUP,
     "\texpiresAt := a.earliestFreshUntil()", "\texpiresAt := a.authority.FreshUntil()"),
    ("Y18 account identity not bundled over scopes", FL,
     "\t\tif len(scopes) > 1 {\n\t\t\tready := make([]string, 0, len(scopes))", "\t\tif false {\n\t\t\tready := make([]string, 0, len(scopes))"),
    ("Y19 admit drops the collection cause", ADM, "\t\trefusal.cause = err\n", ""),
    ("Y20 dispatch flattens the collection cause (%v)", DCY,
     'return execgw.Outcome{}, fmt.Errorf("engine: first-leg admission %s: %w", admitted.Code, admitted.cause)',
     'return execgw.Outcome{}, fmt.Errorf("engine: first-leg admission %s: %v", admitted.Code, admitted.cause)'),
    ("Y21 dispatch types a cause-less refusal as a scope refusal (non-nil wrap — review B #7)", DCY,
     'return execgw.Outcome{}, fmt.Errorf("engine: first-leg admission %s: %s", admitted.Code, admitted.Detail)',
     'return execgw.Outcome{}, fmt.Errorf("engine: first-leg admission %s: %w", admitted.Code, &strategyScopeRefusal{detail: admitted.Detail})'),
    ("Y22 delivery skips every fault (J4 misclassification, carried from X01)", MD,
     "\t\tif errors.As(err, &scope) {", "\t\tif true || errors.As(err, &scope) {"),
]
SET_5222_FIX_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run",
     "TestAnActivatedTwoScope|TestTheSecondLeg|TestAScope|TestTheFirstLegCurrency|TestTheRiskStubBridge|TestAForgedScope|TestTheFirstLegSeal|TestARiskScope|TestACorrupt|TestAnAccountLoad|TestTheLeaseRisk|TestAnUnactivated|TestAWorker|TestTheSameOwnerScope|TestEveryAdmitted|TestARefusedHandoff|TestTheProposalSetDigest|TestFirstLegAuthority|TestTheFirstLegBackstop|TestTheProductionCycle",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "-run", "TestTheSingleProposalAssumption|TestTheScopeRefusalType|TestOnlyATypedScopeRefusal|TestNoEngineErrorTypeImplementsAs|Single", "./internal/app/engine"],
    ["go", "test", "-count=1", "./internal/riskbucket"],
]
SET_5222_FIX2 = [
    ("Z01 account failure classified scope-local again (the pre-ruling boundary)", FL,
     "\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, fmt.Errorf(\"production account authority fault for owner scope %s: %w\", key.Symbol,\n\t\t\taccount.accountScopeCause(key))",
     "\t\treturn execgw.QFinalCampaignFirstLegIssuance{}, &strategyScopeRefusal{scope: key, detail: \"no ready account authority for this owner scope\",\n\t\t\tcause: account.accountScopeCause(key)}"),
    ("Z02 account fault cause flattened (%v)", FL,
     "fmt.Errorf(\"production account authority fault for owner scope %s: %w\", key.Symbol,", "fmt.Errorf(\"production account authority fault for owner scope %s: %v\", key.Symbol,"),
    ("Z03 account loader ignores its own context after a failure", FL,
     "\t\t\t\tif ctxErr := ctx.Err(); ctxErr != nil {", "\t\t\t\tif ctxErr := ctx.Err(); false && ctxErr != nil {"),
    ("Z04 risk scope-local classifier widened to every cause (carried Y07)", OWN,
     "return errors.Is(scope.cause, riskbucket.ErrProductionRiskScopeRefused), scope.cause", "return true, scope.cause"),
    ("Z05 delivery skips every fault (carried Y22)", MD,
     "\t\tif errors.As(err, &scope) {", "\t\tif true || errors.As(err, &scope) {"),
]
LRT = "internal/app/engine/strategy_lane_runtime.go"
SET_5622 = [
    ("W01 a latched lane surfaces as a market cycle error", LRT,
     "\truntime.record(observations)\n",
     "\truntime.record(observations)\n\tfor _, observation := range observations {\n\t\tif observation.Health == strategyworker.LaneLatched {\n\t\t\treturn context.DeadlineExceeded\n\t\t}\n\t}\n"),
    ("W02 a latched lane escalates as central integrity", LRT,
     "\truntime.record(observations)\n",
     "\truntime.record(observations)\n\tfor _, observation := range observations {\n\t\tif observation.Health == strategyworker.LaneLatched {\n\t\t\treturn StrategyCentralIntegrityFailure(context.DeadlineExceeded)\n\t\t}\n\t}\n"),
    ("W03 an unrecorded lane latch is dropped silently", LRT,
     "\tif err := runtime.persistMarketLatches(ctx, market, activationGeneration, runtime.clk.Now()); err != nil {\n\t\treturn err\n\t}",
     "\t_ = runtime.persistMarketLatches(ctx, market, activationGeneration, runtime.clk.Now())"),
    ("W04 an unrecorded lane latch escalates as central integrity", LRT,
     "\tif err := runtime.persistMarketLatches(ctx, market, activationGeneration, runtime.clk.Now()); err != nil {\n\t\treturn err\n\t}",
     "\tif err := runtime.persistMarketLatches(ctx, market, activationGeneration, runtime.clk.Now()); err != nil {\n\t\treturn StrategyCentralIntegrityFailure(err)\n\t}"),
    ("W05 production supervisor workers stop refreshing", SUP,
     "\t\t\tMarket: market, PollInterval: DefaultStrategyCycleLimit, RefreshesAuthority: true,",
     "\t\t\tMarket: market, PollInterval: DefaultStrategyCycleLimit, RefreshesAuthority: false,"),
    ("W06 fault stream capacity 2 -> 1", SUP,
     "faults: make(chan StrategyWorkerFault, 2)", "faults: make(chan StrategyWorkerFault, 1)"),
]
SET_5622_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run",
     "TestEightSimultaneousLaneFaults|TestALaneLatchThatCannotBeRecorded|TestTheProductionSupervisorStillHasTwo|TestTheFaultStream|TestEveryWorkerCanHandOff|TestARefreshOnly|TestAnOrdinaryRefreshOnly|TestTheProductionCycleHandsEvery",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "-run", "Census|TestTheFaultStream|TestEveryWorkerCanHandOff", "./internal/app/engine"],
]
SRP = "internal/strategyrouter/production.go"
SET_61 = [
    ("V01 two families sharing a risk id accepted", RB,
     "\t\tif prior, seen := riskFamilies[value.RiskID]; !known || seen && prior != family {",
     "\t\tif prior, seen := riskFamilies[value.RiskID]; !known || seen && prior != family && false {"),
    ("V02 a lane without a family accepted", RB,
     "\t\tif prior, seen := riskFamilies[value.RiskID]; !known || seen && prior != family {",
     "\t\tif prior, seen := riskFamilies[value.RiskID]; (!known && false) || seen && prior != family {"),
    ("V03 family lint removed", RB,
     "\t\triskFamilies[value.RiskID] = family", "\t\t_ = family"),
    ("V04 family resolved from the wrong market", RB,
     "strategyrouter.ProductionLaneFamily(strategyrouter.Market(body.Market), value.LaneID)",
     "strategyrouter.ProductionLaneFamily(strategyrouter.MarketKR, value.LaneID)"),
    ("V05 resolver ignores the market (any market's lane resolves)", SRP,
     "\tdescriptor, ok := productionRouteDescriptors(market)[laneID]\n\treturn descriptor.Family, ok",
     "\tdescriptor, ok := productionRouteDescriptors(market)[laneID]\n\tif !ok {\n\t\tdescriptor, ok = productionRouteDescriptors(MarketKR)[laneID]\n\t}\n\tif !ok {\n\t\tdescriptor, ok = productionRouteDescriptors(MarketUS)[laneID]\n\t}\n\treturn descriptor.Family, ok"),
    ("V06 resolver invents a family for unknown lanes", SRP,
     "\tdescriptor, ok := productionRouteDescriptors(market)[laneID]\n\treturn descriptor.Family, ok",
     "\tdescriptor, ok := productionRouteDescriptors(market)[laneID]\n\tif !ok {\n\t\treturn FamilyContinuation, true\n\t}\n\treturn descriptor.Family, ok"),
    ("V07 weekly horizon squeezed into SHORT", RB,
     "\thorizon := Horizon(lineage.Horizon)\n",
     "\thorizon := Horizon(lineage.Horizon)\n\tif horizon == \"WEEKLY\" {\n\t\thorizon = HorizonShort\n\t}\n"),
]
SET_61_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "./internal/riskbucket"],
    ["go", "test", "-count=1", "-run", "TestProductionLaneFamily|TestPaired|TestProduction", "./internal/strategyrouter"],
]
SET_62 = [
    ("U01 bucket exhaustion no longer a scope refusal (the pre-ruling stop)", ADM,
     "errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalBucketCapExhausted {",
     "errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalBucketCapExhausted && false {"),
    ("U02 every precheck refusal widened into a scope refusal", ADM,
     "\t\tif qFinal := (*execgw.QFinalRefusal)(nil); errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalBucketCapExhausted {",
     "\t\tif qFinal := (*execgw.QFinalRefusal)(nil); true || errors.As(err, &qFinal) && qFinal.Code == riskbucket.RefusalBucketCapExhausted {"),
    ("U03 another q_final code widened into a scope refusal", ADM,
     "qFinal.Code == riskbucket.RefusalBucketCapExhausted {", "qFinal.Code != riskbucket.RefusalZeroQuantity {"),
    ("U04 issuance-stage STALE reclassified as a scope refusal (design 4 protection)", ADM,
     "\t\treturn strategyFirstLegRefusal(StrategyFirstLegAtomicAdmissionFailed, accepted.market, err.Error())",
     "\t\trefusal := strategyFirstLegRefusal(StrategyFirstLegAtomicAdmissionFailed, accepted.market, err.Error())\n\t\tif key, keyed := strategyOwnerKeyOf(authority.Result.Lineage); keyed && strings.Contains(err.Error(), \"BUCKET_USAGE_STALE\") {\n\t\t\trefusal.cause = &strategyScopeRefusal{scope: key, detail: \"stale\", cause: err}\n\t\t}\n\t\treturn refusal"),
    ("U05 the scope refusal drops its cause", ADM,
     "detail: \"this owner scope's risk bucket is exhausted\", cause: err}", "detail: \"this owner scope's risk bucket is exhausted\"}"),
    ("U06 q_final ignores the strategy (family) cap", "internal/riskbucket/admission.go",
     "\t\tqFinal = minUint64(qFinal, capQuantity)\n", "\t\tif bucket.Key.Dimension != DimensionStrategy {\n\t\t\tqFinal = minUint64(qFinal, capQuantity)\n\t\t}\n"),
]
SET_62_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run",
     "TestQFinalIsBound|TestAnExhaustedFamilyBucket|TestASharedDimension|TestTwoFamiliesRacing|TestAPrecheckRefusalOtherThan|TestAnExistingGuardianCap|TestAnActivatedTwoScopeMarketIssues|TestTheScopeRefusalType|TestAScope|TestACorrupt|TestARiskScope",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "-run", "TestTheScopeRefusalType|TestOnlyATypedScopeRefusal|TestNoEngineErrorType", "./internal/app/engine"],
]
BUD = "internal/scheduler/budget.go"
SSC = "internal/scheduler/strategy_scope.go"
SET_71 = [
    ("T01 completion ignores the scope (cross-family replay accepted)", BUD,
     "record.generation != token.generation || record.scope != scope {", "record.generation != token.generation {"),
    ("T02 scoped completion skips the token's own scope check", SSC,
     "\tif !scope.Valid() || token.scope != scope.digest() {", "\tif !scope.Valid() {"),
    ("T03 unscoped API completes scoped capabilities (scoped -> unscoped replay)", BUD,
     "\treturn c.complete(key, token, [sha256.Size]byte{})", "\treturn c.complete(key, token, c.endpoints[key].commitments[token.capability].scope)"),
    ("T04 scoped API completes unscoped capabilities (unscoped -> scoped replay)", SSC,
     "\treturn c.complete(key, token.token, token.scope)", "\treturn c.complete(key, token.token, [sha256.Size]byte{}) || c.complete(key, token.token, token.scope)"),
    ("T05 issuance does not bind the scope", BUD,
     "budgetCommitment{class: class, generation: state.generation, scope: scope}", "budgetCommitment{class: class, generation: state.generation}"),
    ("T06 each family gets its own capacity (capacity multiplied)", BUD,
     "\tif discretionary <= 0 || len(state.commitments) >= discretionary {",
     "\tif discretionary <= 0 || len(state.commitments) >= 4*discretionary {"),
    ("T07 refusal name leaks the internal reason instead of BUDGET_DEFERRED", SSC,
     "\t\tresult.Refusal = StrategyBudgetDeferred\n\t\treturn result", "\t\tresult.Refusal = string(grant.Reason)\n\t\treturn result"),
    ("T08 safety classes accepted by the strategy API", SSC,
     "\tif !isKnownPollClass(class) || isSafetyClass(class) {", "\tif !isKnownPollClass(class) {"),
    ("T09 scope validation removed", SSC,
     "\tif !scope.Valid() {\n\t\treturn StrategyBudgetGrant{", "\tif false && !scope.Valid() {\n\t\treturn StrategyBudgetGrant{"),
    ("T10 a production caller of the strategy budget API appears (Q4 pin)", "internal/app/engine/a112_budget_wire_mutant.go", None,
     "package engine\n\nimport \"github.com/JungHoonGhae/tossinvest-cli/internal/scheduler\"\n\nvar _ = (*scheduler.BudgetCoordinator).TryAcquireStrategy\n"),
]
SET_71_TESTS = [
    ["go", "test", "-count=1", "./internal/scheduler"],
]
LPJ = "internal/app/engine/strategy_lane_projection.go"
LRT = "internal/app/engine/strategy_lane_runtime.go"
RPJ = "internal/app/engine/strategy_runtime_projection.go"
PLN = "internal/strategyprojection/lanes.go"
SET_73 = [
    ("P01 projection offers a trigger to the lane (read-only violated)", LPJ,
     "Pending:      lane.Pending(),", "Pending:      func() int { lane.Offer(); return lane.Pending() }(),"),
    ("P02 projection records a lane failure (read-only violated)", LPJ,
     "ConsecutiveFailures: lane.ConsecutiveFailures(),",
     "ConsecutiveFailures: func() uint64 { lane.Fail(\"projection\", false); return lane.ConsecutiveFailures() }(),"),
    ("P03 Read does not overlay the live lanes", RPJ, "\tif lanes != nil {\n\t\tsnapshot.Lanes", "\tif false && lanes != nil {\n\t\tsnapshot.Lanes"),
    ("P04 record never advances the wave", LRT,
     "\tif runtime.waves[market] < ^uint64(0) {\n\t\truntime.waves[market]++\n\t}", "\tif false {\n\t\truntime.waves[market]++\n\t}"),
    ("P05 one wave counter shared by both markets", LRT,
     ("\t\truntime.waves[market]++", "\t\tobservation.Wave = runtime.waves[market]"),
     ("\t\truntime.waves[StrategyMarketKR]++", "\t\tobservation.Wave = runtime.waves[StrategyMarketKR]")),
    ("P06 lane desired/effective ignore the activation", LRT,
     "Desired: lane.Desired(promotion), Effective: lane.Effective(promotion),",
     "Desired: strategyrouter.StateOff, Effective: strategyrouter.StateOff,"),
    ("P07 selected keeps only the first scope (R4 regression)", LPJ,
     "\t\tvalue.Selected = append(value.Selected, selected)\n",
     "\t\tvalue.Selected = append(value.Selected, selected)\n\t\tbreak\n"),
    ("P08a selected drops the admitted check — EQUIVALENT by construction: a refused Single() returns the zero Result "
     "(pinned by strategyhandoff TestARefusedSingleReturnsTheZeroResult) whose ValidProposal() is false, so the remaining check "
     "excludes exactly the same handoffs; removing both checks is P08b (CAUGHT)", LPJ,
     "\t\tif !admitted || !scoped.ValidProposal() {", "\t\tif !scoped.ValidProposal() {\n\t\t\t_ = admitted"),
    ("P08b selected drops both checks", LPJ, "\t\tif !admitted || !scoped.ValidProposal() {", "\t\tif false {\n\t\t\t_ = admitted"),
    ("P09 coordinators not set (failure branch hides them)", RPJ,
     "\t\tsnapshot.Coordinators[index] = strategyCoordinatorProjection(market, assembly.proposals.forMarket(market))",
     "\t\t_, _ = index, strategyCoordinatorProjection"),
    ("P10 Validate drops the fixed lane order", PLN,
     "\t\t\tlane.Horizon != key.hor {\n\t\t\treturn errors.New(\"lanes out of the fixed production order\")",
     "\t\t\tlane.Horizon != key.hor && false {\n\t\t\treturn errors.New(\"lanes out of the fixed production order\")"),
    ("P11 Validate accepts an invented runtime", PLN,
     "!validState(lane.Effective) || lane.Runtime != LaneRuntimeUnobserved {", "!validState(lane.Effective) {"),
    ("P12 Clone copies lanes shallowly", PLN,
     "\tout := make([]LaneRuntimeProjection, len(lanes))\n", "\treturn append([]LaneRuntimeProjection(nil), lanes...)\n\tout := make([]LaneRuntimeProjection, len(lanes))\n"),
    ("P13 Validate lets an unobserved lane carry facts", PLN,
     "\t\t\treturn errors.New(\"unobserved lane carries inferred facts\")", "\t\t\treturn nil"),
    ("P14 coordinator reason vocabulary misses an engine reason", PLN,
     "\"PROPOSAL_PRODUCTION_FAULT\", \"FAMILY_GATE_CLOSED\"}", "\"PROPOSAL_PRODUCTION_FAULT\"}"),
    ("P15 lane start vocabulary misses a worker value", PLN,
     "\t\tstring(LaneStartTooSoon), string(LaneStartNoTrigger)}", "\t\tstring(LaneStartTooSoon)}"),
    ("P16 a lane JSON name drops out of the contract list", PLN, "return []string{\"abandoned\", \"abnormal\", ", "return []string{\"abandoned\", "),
    ("P17 the default lanes are born ON", PLN,
     "Horizon: key.hor, Desired: StateOff, Effective: StateOff,", "Horizon: key.hor, Desired: StateOn, Effective: StateOff,"),
    ("P18 lane health read from the last observation, not the lane", LPJ,
     "health := strategyprojection.LaneHealth(lane.Health())", "health := strategyprojection.LaneHealth(observation.Health)"),
    ("P19 latch reason not normalized", LPJ,
     "FirstFailure: strategyprojection.NormalizedText(lane.FirstFailure()),", "FirstFailure: projectionOptional(lane.FirstFailure()),"),
    ("P20 coordinator lists null when empty", LPJ,
     "GatedOutcomes: []string{},\n", "GatedOutcomes: nil,\n"),
    # 판정 (A) — first refusal · config · calibration 계보.
    ("P21 lane refusal carried regardless of the outcome (Manager condition)", LPJ,
     "\t\t\tif observation.Outcome == strategyworker.OutcomeRefused {\n\t\t\t\tvalue.Refusal = projectionOptional(string(observation.Refusal))\n\t\t\t}",
     "\t\t\tcode := string(observation.Refusal)\n\t\t\tvalue.Refusal = &code"),
    ("P21b lane refusal outcome gate removed but empty codes still null — EQUIVALENT by the worker contract: only a REFUSED cycle "
     "carries a non-empty Cycle.Refusal (strategyworker FamilyWorker.Run); the outcome gate is the projection's own guard", LPJ,
     "\t\t\tif observation.Outcome == strategyworker.OutcomeRefused {\n\t\t\t\tvalue.Refusal",
     "\t\t\tif true {\n\t\t\t\tvalue.Refusal"),
    ("P22 runLane does not record the cycle refusal", LRT,
     "\tobservation.Refusal = bounded.Cycle.Refusal\n", ""),
    ("P23 selected calibration read from the first entry, not the scope's own (first run BUILD-FAIL: unused loop variable — redefined)", LPJ,
     "\t\tif entry.authority.Proposal().Lineage.Identity == lineage.Identity {\n\t\t\treturn entry.route, true",
     "\t\tif entry.authority.Proposal().Lineage.Identity != \"\" {\n\t\t\treturn entry.route, true"),
    ("P24 selected config digest dropped", LPJ,
     ", ConfigDigest: projectionOptional(lineage.ConfigDigest)}", "}"),
    ("P25 Validate drops the refusal/outcome pairing", PLN,
     "\tif (lane.Refusal != nil) != refused || lane.Refusal != nil && !member(*lane.Refusal, ArbitrationRefusals()) {",
     "\tif lane.Refusal != nil && !member(*lane.Refusal, ArbitrationRefusals()) && refused {"),
]
SET_73_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run",
     "TestAProcessWithoutLanes|TestTheCycleGeneration|TestTheLaneDesiredAndEffective|TestEightLatchedLanes|TestReadingTheLaneProjection|TestTheCoordinatorChild|TestTheCoordinatorReasonVocabulary|TestALatchReasonWithControl|TestAnActivatedTwoScopeMarketPromotesItsWorkerPerScope",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "./internal/strategyprojection"],
    ["go", "test", "-count=1", "-run", "TestTheProjection|TestTheLaneRead", "./internal/strategyworker"],
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run", "TestTheLaneDesiredAndEffectiveFollow", "./internal/strategyworker"],
    ["go", "test", "-count=1", "-run", "OpenAPI|StrategyRuntime", "./internal/httpapi"],
]
SET_74 = [
    ("Q01 a production engine file starts emitting expvar metrics", "internal/app/engine/a112_metric_mutant.go", None,
     "package engine\n\nimport _ \"expvar\"\n"),
    ("Q02 a production cmd file starts emitting OpenTelemetry metrics", "cmd/tossctl/a112_metric_mutant.go", None,
     "package main\n\nimport _ \"go.opentelemetry.io/otel/metric\"\n"),
]
SET_74_TESTS = [["go", "test", "-count=1", "-run", "TestNoProductionCodeEmitsMetrics", "./internal/strategyprojection"]]
LVW = "internal/strategyworker/lane_view.go"
WAV = "internal/app/engine/strategy_refresh_wave.go"
RTM = "internal/app/engine/runtime.go"
SET_75 = [
    ("R01 lanes run sequentially again (Manager: revert-to-sequential mutant)", LRT,
     "\t\tgo func(index int, lane *strategyworker.Lane, input strategyworker.Input) {",
     "\t\tfunc(index int, lane *strategyworker.Lane, input strategyworker.Input) {"),
    ("R02 a lane goroutine panic is swallowed instead of re-raised on the market cycle", LRT,
     "\t\tif recovered != nil {\n\t\t\tpanic(recovered)", "\t\tif recovered != nil {\n\t\t\t_ = recovered"),
    ("R03 the lane goroutine does not recover (a panic would end the process)", LRT,
     "\t\t\tdefer func() { panics[index] = recover() }()\n", ""),
    ("R04 evaluate does not join its lanes", LRT, "\tjoin.Wait()\n", ""),
    ("R05 the projection reads a per-field state accessor again (torn row)", LPJ,
     "ConsecutiveFailures: row.ConsecutiveFailures,", "ConsecutiveFailures: lane.ConsecutiveFailures(),"),
    ("R06 Status judges health differently from Health", LVW, "Health: lane.healthLocked(),", "Health: LaneHealthy,"),
    ("R07 the test-seam laneStepFor ignores its hook (the hang never reaches the lane — reach control; re-anchored after the seam moved behind the build tag)",
     "internal/app/engine/strategy_lane_step_testseam.go",
     "\tif hook, ok := strategyLaneStepHooks.Load(runtime); ok {", "\tif hook, ok := strategyLaneStepHooks.Load(runtime); false && ok {"),
    ("R08 every lane receives every sealed proposal (fan-out duplicated)", LRT,
     "\t\t\tif lane.Owns(candidate.Proposal) {", "\t\t\tif true || lane.Owns(candidate.Proposal) {"),
    ("R09 a market cycle bypasses the one-second wave cache (a wave per cycle)", WAV,
     "\tif c.strategyRefresh != nil && !now.Before(c.strategyRefreshAt) && now.Sub(c.strategyRefreshAt) < time.Second {",
     "\tif false && c.strategyRefresh != nil && !now.Before(c.strategyRefreshAt) && now.Sub(c.strategyRefreshAt) < time.Second {"),
    ("R10 the projection Read waits for an in-flight remote wave", RPJ,
     "\tc.strategyProjectionMu.RLock()\n\tstore, supervisor := c.strategyProjection, c.strategySupervisor",
     "\tc.strategyRefreshMu.Lock()\n\tif wave := c.strategyRefreshWave; wave != nil {\n\t\tc.strategyRefreshMu.Unlock()\n\t\t<-wave.done\n\t} else {\n\t\tc.strategyRefreshMu.Unlock()\n\t}\n"
     "\tc.strategyProjectionMu.RLock()\n\tstore, supervisor := c.strategyProjection, c.strategySupervisor"),
    ("R11 the runtime starts its loops one after another (safety loops wait behind the entry supervisor)", RTM,
     "\t\tgo func(loop SupervisedLoop) {\n\t\t\tdefer wg.Done()", "\t\tfunc(loop SupervisedLoop) {\n\t\t\tdefer wg.Done()"),
    # 판정 (A) 핀 강화의 변이 둘(Manager).
    ("R12 the production laneStepFor consults a hook (a function field reaches the production binary)", "internal/app/engine/strategy_lane_step.go",
     "\treturn strategyFamilyLaneStep(lane, promotion)\n}\n",
     "\tif hook := strategyLaneStepProductionHook; hook != nil {\n\t\treturn hook(lane, promotion)\n\t}\n\treturn strategyFamilyLaneStep(lane, promotion)\n}\n\n"
     "var strategyLaneStepProductionHook func(*strategyworker.Lane, strategyrouter.FamilyActivation) strategyworker.Step\n"),
    ("R13 the seam file's build constraint widens beyond the test-seam build (a debug build would ship the seam)",
     "internal/app/engine/strategy_lane_step_testseam.go", "//go:build tossos_testseams\n", "//go:build tossos_testseams || tossos_debug\n"),
]
SET_75_TESTS = [
    ["go", "test", "-tags", "tossos_testseams", "-count=1", "-run",
     "TestAHungLane|TestAPanicOutsideALaneStep|TestAProjectedLaneRow|TestTheLaneProjectionReadsALaneRow|TestAStalledRemoteWave|TestSafetyLoopsKeepTheirCadence|TestEightLanesShareOneAuthorityWave|TestOnlyThePackageLevelStepEverRunsInsideALane",
     "./internal/app/engine"],
    ["go", "test", "-count=1", "-run", "TestTheLaneStatusIsTheRowItsAccessorsRead", "./internal/strategyworker"],
    ["go", "test", "-count=1", "-run", "TestOnlyThePackageLevelStepEverRunsInsideALane", "./internal/app/engine"],
]
SETS = {"5.6.2.1": (SET_5621, SET_5621_TESTS), "5.2.2.1": (SET_5221, SET_5221_TESTS),
        "5.2.2.1-fix": (SET_5221_FIX, SET_5221_FIX_TESTS), "5.2.2.1-fix3": (SET_5221_FIX3, SET_5221_FIX3_TESTS),
        "5.2.2.1-fix4": (SET_5221_FIX4, [["go", "test", "-count=1", "./internal/strategyhandoff"]]),
        "6.2-seal": (SET_62_SEAL, SET_62_SEAL_TESTS), "5.2.2.2": (SET_5222, SET_5222_TESTS),
        "5.2.2.2-fix": (SET_5222_FIX, SET_5222_FIX_TESTS), "5.2.2.2-fix2": (SET_5222_FIX2, SET_5222_FIX_TESTS),
        "5.6.2.2": (SET_5622, SET_5622_TESTS), "6.1": (SET_61, SET_61_TESTS),
        "6.2": (SET_62, SET_62_TESTS), "7.1": (SET_71, SET_71_TESTS), "7.3": (SET_73, SET_73_TESTS), "7.4": (SET_74, SET_74_TESTS), "7.5": (SET_75, SET_75_TESTS)}
# 리뷰 B #10(5.2.2.2): 대조군의 pass 사건 수를 고정한다 — 0 보다 큼만 보면 시험 일부가 조용히 빠져도 대조군이 GREEN 이다. 집합별 기대치는 그 집합을
# 처음 돌린 대조군의 실측(원장 CONTROL 줄)이고, 시험을 더하면 여기를 같이 바꾼다(바꾸는 편집이 리뷰에 보인다).
EXPECTED_PASSES = {"5.2.2.2-fix": [60, 8, 122]}  # 첫 대조군(2026-10-01) 실측
MUTANTS, TESTS = SET_5621, SET_5621_TESTS

def run_tests(copy: Path, env: dict) -> tuple[str, str]:
    failed, notes, build = [], [], False
    for command in TESTS:
        result = subprocess.run(command, cwd=copy, env=env, capture_output=True, text=True)
        if result.returncode != 0:
            out = result.stdout + result.stderr
            if "[build failed]" in out or "[setup failed]" in out:
                build = True
                notes.append(next((l for l in out.splitlines() if ".go:" in l), "build failed")[:200])
                continue
            names = [line.strip()[len("--- FAIL: "):].split(" ")[0] for line in out.splitlines() if line.strip().startswith("--- FAIL: ")]
            leaves = [n for n in names if not any(o != n and o.startswith(n + "/") for o in names)]
            failed.extend(leaves)
            if not names:
                notes.append(out.strip().splitlines()[-1][:160] if out.strip() else "no output")
    failed = list(dict.fromkeys(failed))  # 같은 시험이 두 명령(태그 · 무태그)에서 세어지지 않게(6.2 리뷰 보이스 B #11)
    why = f"{len(failed)} failing: " + ", ".join(failed[:8]) + (" …" if len(failed) > 8 else "") + (" | " + " | ".join(notes) if notes else "")
    if build:
        return "BUILD-FAIL", why
    if failed or notes:
        return "RED", why
    return "GREEN", ""


def main() -> None:
    args = sys.argv[1:]
    only = None
    if "--only" in args:
        i = args.index("--only")
        only = re.compile(args[i + 1])
        args = args[:i] + args[i + 2:]
    global MUTANTS, TESTS
    expected = None
    if "--set" in args:
        i = args.index("--set")
        MUTANTS, TESTS = SETS[args[i + 1]]
        expected = EXPECTED_PASSES.get(args[i + 1])
        args = args[:i] + args[i + 2:]
    scratch, own = Path(args[0]), args[1:]
    copy = scratch / f"mut-a112-lot-{os.getpid()}"
    copy.mkdir(parents=True)
    # 사본은 HEAD 커밋 트리 + 이 로트가 넘긴 파일(own)뿐이다 — 병행 세션의 미커밋 편집이 사본에 섞이면 대조군부터
    # 깨지고(2026-09-30 실측: 남의 미커밋 journal 편집이 남의 미추적 파일을 참조), 섞인 채 GREEN 이면 무엇을 쟀는지 모른다.
    archive = subprocess.run(["git", "archive", "HEAD", "go.mod", "go.sum", "internal", "cmd", "tools", "openspec"], cwd=ROOT,
                             capture_output=True, check=True).stdout
    subprocess.run(["tar", "-x", "-C", str(copy)], input=archive, check=True)
    for rel in own:
        source = ROOT / rel
        (copy / rel).parent.mkdir(parents=True, exist_ok=True)
        shutil.copy2(source, copy / rel)
    env = dict(os.environ, GOFLAGS="-trimpath")
    head = subprocess.run(["git", "rev-parse", "HEAD"], cwd=ROOT, capture_output=True, text=True, check=True).stdout.strip()
    dirty = []
    ledger = open(copy / "ledger.tsv", "w", encoding="utf-8")
    ledger.write(f"TREE\tHEAD {head} (git archive) + own: {','.join(own) or 'none'}\n")
    verdict, why = run_tests(copy, env)
    # 대조군 GREEN 은 종료 코드다 — 시험이 실제로 돌았는지 pass 사건으로 확인한다(6.2 리뷰 codex #4, 이월 #1 의 자기 적용). init 조기 종료
    # 같은 모양이면 대조군도 변이도 전부 「GREEN」이 되어 원장이 무의미해진다.
    if verdict == "GREEN":
        for command in TESTS:
            json_command = command[:2] + ["-json"] + command[2:]
            completed = subprocess.run(json_command, cwd=copy, env=env, capture_output=True, text=True)
            passes = sum(1 for line in completed.stdout.splitlines() if '"Action":"pass"' in line and '"Test":' in line)
            # 종료 코드도 본다(6.2 봉인 이월 — pass 사건만 세면 한 시험이 빨개도 다른 시험의 pass 로 대조군이 「GREEN」이 된다).
            if completed.returncode != 0:
                verdict, why = "CONTROL-JSON-RED", " ".join(command) + f" -json exited {completed.returncode}"
                break
            if passes == 0:
                verdict, why = "NO-TESTS-RAN", " ".join(command) + " emitted no test pass event"
                break
            index = TESTS.index(command)
            if expected is not None and passes != expected[index]:
                verdict, why = "CONTROL-PASS-COUNT", " ".join(command) + f" pass events {passes} != expected {expected[index]}"
                break
            why = (why + " " if why else "") + f"[{command[-1]} pass events {passes}]"
    ledger.write(f"CONTROL\t{verdict}\t{why}\n")
    ledger.flush()
    if verdict != "GREEN":
        print("control not GREEN — stop:", verdict, why)
        sys.exit(2)
    for ident, rel, old, new in MUTANTS:
        if only and not only.search(ident):
            continue
        target = copy / rel
        if old is None:  # 새 파일을 만드는 변이 — 끝나면 지운다
            if target.exists():
                ledger.write(f"{ident}\tNOT-APPLIED\ttarget exists\n")
                ledger.flush()
                continue
            target.write_text(new, encoding="utf-8")
            verdict, why = run_tests(copy, env)
            target.unlink()
            label = {"RED": "CAUGHT", "GREEN": "SURVIVED"}.get(verdict, verdict)
            ledger.write(f"{ident}\t{label}\t{why}\n")
            ledger.flush()
            continue
        pristine = target.read_text(encoding="utf-8")
        # 한 변이가 같은 파일의 여러 자리를 바꿀 수 있다(old · new 가 튜플이면 짝지어 차례로) — 각 앵커는 정확히 한 번.
        pairs = list(zip(old, new)) if isinstance(old, tuple) else [(old, new)]
        mutated, missing = pristine, [a for a, _ in pairs if pristine.count(a) != 1]
        if missing:
            ledger.write(f"{ident}\tNOT-APPLIED\tanchor count != 1: {missing[0][:60]!r}\n")
            ledger.flush()
            continue
        for a, b in pairs:
            mutated = mutated.replace(a, b, 1)
        target.write_text(mutated, encoding="utf-8")
        verdict, why = run_tests(copy, env)
        label = {"RED": "CAUGHT", "GREEN": "SURVIVED"}.get(verdict, verdict)
        if ident.startswith("N"):  # 동작이 같은 리팩터 — 초록이어야 한다(빨가면 거짓 양성)
            label = {"GREEN": "GREEN-AS-EXPECTED", "RED": "FALSE-POSITIVE"}.get(verdict, verdict)
        ledger.write(f"{ident}\t{label}\t{why}\n")
        ledger.flush()
        target.write_text(pristine, encoding="utf-8")
    ledger.close()
    print((copy / "ledger.tsv").read_text(encoding="utf-8"))


if __name__ == "__main__":
    main()
