#!/usr/bin/env python3
"""a199 인용 AST 추출기 — tools/logic-map 으로 대상 함수의 AST 를 뽑아 이 폴더에 쓰고 요약을 출력함.

사용(저장소 루트에서):
  python3 <이 파일> [--bin <logic-map 바이너리>] [--compare <사전 증거 폴더>]
--bin 이 없으면 `go run ./tools/logic-map` 을 씀. --compare 를 주면 사전 증거의 같은 함수 AST 와
시작·끝 줄, 분기 (id, kind, 줄), source_sha256 을 대조함.
"""

from __future__ import annotations

import argparse
import json
import subprocess
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent

# (소스 파일, 함수) — 사전 증거 12 + 1단계에서 더한 것
TARGETS: list[tuple[str, str]] = [
    ("internal/app/engine/interlock.go", "verifyGate"),
    ("internal/app/engine/interlock.go", "runInterlock"),
    ("internal/app/engine/interlock.go", "refuseStartup"),
    ("internal/app/engine/interlock.go", "newAutomationStatus"),
    ("internal/app/engine/interlock.go", "logGateDecision"),
    ("internal/app/engine/engine.go", "NewContext"),
    ("cmd/tossctl/engine.go", "runEngineRun"),
    ("cmd/tossctl/engine.go", "engineRuntime"),
    ("internal/app/engine/runtime_wiring.go", "UnmetInterlockClauses"),
    ("internal/attest/attest.go", "Load"),
    ("internal/attest/attest.go", "Attestation.Verify"),
    ("internal/attest/attest.go", "Attestation.MissingEndpoints"),
    ("internal/app/engine/attestation_watch.go", "judgeAttestation"),
    ("internal/execgw/protection.go", "Gateway.checkProtection"),
    ("internal/app/engine/exitwiring.go", "Context.ExitObserver"),
    ("internal/app/engine/reconcileloop.go", "Context.ReconcileDriver"),
    # 1.4 호출 자리 수를 세기 위한 호출자 함수
    ("cmd/tossctl/engine_assembly.go", "assembleEngine"),
    ("internal/app/engine/engine.go", "New"),
    ("internal/app/engine/attestation_watch.go", "attestationWatch.prepare"),
    ("internal/app/engine/attestation_watch.go", "attestationWatch.eventFor"),
    ("internal/app/engine/attestation_watch.go", "Context.AttestationWatch"),
    ("internal/app/engine/attestation_watch.go", "newAttestationWatch"),
    ("internal/app/engine/strategy_entry_supervisor.go", "Context.NewPairedStrategyEntryProductionAssembly"),
    ("internal/app/engine/strategy_entry_supervisor.go", "Context.productionStrategyWorker"),
    # 1.6 · 1.7
    ("internal/attest/attest.go", "SameAccountMasked"),
    ("internal/execgw/failclosed.go", "AllReasonCodes"),
    ("internal/execgw/retry.go", "EntryGate.checkAccountEntryLocked"),
    ("internal/execgw/gateway.go", "Gateway.checkEntry"),
    ("internal/app/engine/gateway.go", "restoreAlertEntryLatch"),
]

# 사전 증거 파일 이름(scratch a199-evidence) ↔ 대상
PRIOR_NAMES = {
    ("internal/app/engine/interlock.go", "verifyGate"): "ast-interlock-verifyGate.json",
    ("internal/app/engine/interlock.go", "runInterlock"): "ast-interlock-runInterlock.json",
    ("internal/app/engine/interlock.go", "refuseStartup"): "ast-interlock-refuseStartup.json",
    ("internal/app/engine/engine.go", "NewContext"): "ast-engine-NewContext.json",
    ("cmd/tossctl/engine.go", "runEngineRun"): "ast-engine-runEngineRun.json",
    ("cmd/tossctl/engine.go", "engineRuntime"): "ast-engine-engineRuntime.json",
    ("internal/app/engine/runtime_wiring.go", "UnmetInterlockClauses"): "ast-runtime_wiring-UnmetInterlockClauses.json",
    ("internal/attest/attest.go", "Load"): "ast-attest-Load.json",
    ("internal/attest/attest.go", "Attestation.Verify"): "ast-attest-Attestation_Verify.json",
    ("internal/attest/attest.go", "Attestation.MissingEndpoints"): "ast-attest-Attestation_MissingEndpoints.json",
    ("internal/app/engine/attestation_watch.go", "judgeAttestation"): "ast-attestation_watch-judgeAttestation.json",
    ("internal/execgw/protection.go", "Gateway.checkProtection"): "ast-protection-Gateway_checkProtection.json",
}


def out_name(src: str, func: str) -> str:
    pkg = str(Path(src).parent).replace("/", "-")
    return f"{pkg}--{func.lower()}.ast.json"


def extract(bin_path: str | None, src: str, func: str) -> dict:
    cmd = [bin_path] if bin_path else ["go", "run", "./tools/logic-map"]
    proc = subprocess.run(cmd + ["--file", src, "--func", func], capture_output=True, text=True)
    if proc.returncode != 0:
        raise SystemExit(f"logic-map 실패 rc={proc.returncode} {src} {func}: {proc.stderr.strip()}")
    (HERE / out_name(src, func)).write_text(proc.stdout, encoding="utf-8")
    return json.loads(proc.stdout)


def shape(d: dict) -> tuple:
    return (
        d["start"]["line"],
        d["end"]["line"],
        tuple((b["id"], b["kind"], b["at"]["line"]) for b in (d["branches"] or [])),
    )


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--bin")
    ap.add_argument("--compare")
    args = ap.parse_args()
    for src, func in TARGETS:
        d = extract(args.bin, src, func)
        br = " ".join(f"{b['id']}:{b['kind']}@{b['at']['line']}" for b in (d["branches"] or []))
        print(f"{src} {func} {d['start']['line']}-{d['end']['line']} branches={len(d['branches'] or [])} sha={d['source_sha256'][:12]}")
        print(f"  {br}")
        if args.compare and (src, func) in PRIOR_NAMES:
            prior = json.loads((Path(args.compare) / PRIOR_NAMES[(src, func)]).read_text(encoding="utf-8"))
            same_sha = prior["source_sha256"] == d["source_sha256"]
            same_shape = shape(prior) == shape(d)
            print(f"  대조: source_sha256 {'같음' if same_sha else '다름'} · 줄/분기 {'같음' if same_shape else '다름'}")
            if not same_shape:
                print(f"    사전 {shape(prior)}")
                print(f"    현재 {shape(d)}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
