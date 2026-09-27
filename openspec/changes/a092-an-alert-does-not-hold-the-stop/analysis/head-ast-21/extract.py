#!/usr/bin/env python3
"""a092 21판 — 21판 문서가 분기를 근거로 쓰는 함수를 현재 HEAD 에서 AST 로 추출함.

20판 census.py 와 달리 번들과의 SAME/SHIFT 분류는 하지 않음 — 20라운드 B-7 이
그 분류(줄 좌표를 지운 비교)가 호출 이동을 가린다고 짚었음. 여기서는 추출물과
파일 sha·HEAD 커밋만 남김. 문서는 이 추출물의 좌표를 인용함.

사용: python3 extract.py            (출력은 이 파일 옆 <이름>.json · MANIFEST.txt)
저장소 루트는 이 파일 위치에서 유도함.
"""
import hashlib
import json
import subprocess
import sys
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[4]

TARGETS = [
    ("internal/obs/notifier.go", "Notifier.Notify"),
    ("internal/obs/notifier.go", "Notifier.notifyCritical"),
    ("internal/obs/notifier.go", "Notifier.claimAndDeliver"),
    ("internal/obs/notifier.go", "Notifier.Acknowledge"),
    ("internal/obs/notifier.go", "Notifier.logClaimHeld"),
    ("internal/app/engine/exitloop.go", "ExitObserver.alert"),
    ("internal/app/engine/runtime.go", "Runtime.Run"),
    ("internal/app/engine/runtime.go", "Runtime.alert"),
    ("internal/app/engine/alertdelivery.go", "alertDeliverer.cycle"),
    ("internal/app/engine/alertdelivery.go", "alertDeliverer.deliverOne"),
    ("internal/app/engine/alertdelivery.go", "alertDeliverer.judge"),
    ("internal/app/engine/gateway.go", "restoreAlertEntryLatch"),
    ("internal/execgw/modegate.go", "EntryGate.ProjectOperatingMode"),
    ("internal/journal/operating_mode.go", "Journal.SetModeProjector"),
    ("internal/journal/operating_mode.go", "Journal.TransitionOperatingMode"),
    ("internal/journal/operating_mode.go", "Journal.RestoreOperatingModeProjection"),
    ("internal/journal/alert_claim.go", "Journal.ReleaseAlertClaim"),
    ("internal/journal/outbox.go", "Journal.ClaimAlertForDelivery"),
]


def main() -> int:
    head = subprocess.run(["git", "rev-parse", "HEAD"], cwd=ROOT, check=True,
                          capture_output=True, text=True).stdout.strip()
    rows = [f"HEAD {head}"]
    with tempfile.TemporaryDirectory() as tmp:
        tool = Path(tmp) / "logicmap"
        subprocess.run(["go", "build", "-trimpath", "-o", str(tool), "./tools/logic-map"],
                       cwd=ROOT, check=True)
        for file, func in TARGETS:
            raw = subprocess.run([str(tool), "--file", file, "--func", func], cwd=ROOT,
                                 check=True, capture_output=True).stdout
            j = json.loads(raw)
            name = file.replace("/", "-").removesuffix(".go") + "--" + func.lower()
            (HERE / f"{name}.json").write_bytes(raw)
            cur = hashlib.sha256((ROOT / file).read_bytes()).hexdigest()
            if cur != j["source_sha256"]:
                print(f"sha 불일치: {file}", file=sys.stderr)
                return 1
            n = {k: len(j.get(k) or []) for k in ("branches", "returns", "calls", "defers")}
            rows.append(f"{name} {file}:{j['start']['line']}-{j['end']['line']} "
                        f"sha={j['source_sha256'][:12]} branches={n['branches']} "
                        f"returns={n['returns']} calls={n['calls']} defers={n['defers']}")
    (HERE / "MANIFEST.txt").write_text("\n".join(rows) + "\n", encoding="utf-8")
    print("\n".join(rows))
    return 0


if __name__ == "__main__":
    sys.exit(main())
